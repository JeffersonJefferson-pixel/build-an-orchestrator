package worker

import (
	"cube/task"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/golang-collections/collections/queue"
	"github.com/google/uuid"
)

type Worker struct {
	Name      string
	Queue     queue.Queue
	Db        map[uuid.UUID]*task.Task
	TaskCount int
}

func (w *Worker) GetTasks() []*task.Task {
	tasks := []*task.Task{}
	for _, t := range w.Db {
		tasks = append(tasks, t)
	}
	return tasks
}

func (w *Worker) GetTask(taskID uuid.UUID) (*task.Task, error) {
	task, ok := w.Db[taskID]
	if !ok {
		return nil, errors.New(fmt.Sprintf("No task with ID %v found", taskID))
	}
	return task, nil
}

func (w *Worker) CollectStats() {
	fmt.Println("I will collect stats")
}

func (w *Worker) RunTask() task.DockerResult {
	// pull task off the queue
	t := w.Queue.Dequeue()
	if t == nil {
		log.Printf("No tasks in the queue")
		return task.DockerResult{Error: nil}
	}

	// convert to task type
	taskQueued := t.(task.Task)
	fmt.Printf("Found task in queue: %v:\n", taskQueued)

	// retrieve task from db
	taskPersisted := w.Db[taskQueued.ID]
	if taskPersisted == nil {
		taskPersisted = &taskQueued
		w.Db[taskQueued.ID] = &taskQueued
	}

	var result task.DockerResult
	// check whether state transitio is valid
	if task.ValidStateTransition(taskPersisted.State, taskQueued.State) {
		switch taskQueued.State {
		case task.Scheduled:
			// start task
			result = w.StartTask(taskQueued)

		case task.Completed:
			// stop task
			result = w.StopTask(taskQueued)

		default:
			result.Error = errors.New("We should not get here")
		}
	} else {
		err := fmt.Errorf("Invalid transition from %v to %v", taskPersisted.State, taskQueued.State)
		result.Error = err
	}

	return result
}

func (w *Worker) AddTask(t task.Task) {
	w.Queue.Enqueue(t)
}

func (w *Worker) StartTask(t task.Task) task.DockerResult {
	t.StartTime = time.Now().UTC()
	// create docker
	config := task.NewConfig(&t)
	d := task.NewDocker(config)

	// run container
	result := d.Run()
	if result.Error != nil {
		log.Printf("Err running task %v: %v\n", t.ID, result.Error)
		t.State = task.Failed
		w.Db[t.ID] = &t
		return result
	}

	// update task
	t.ContainerID = result.ContainerId
	t.State = task.Running
	w.Db[t.ID] = &t

	return result
}

func (w *Worker) StopTask(t task.Task) task.DockerResult {
	// create docker
	config := task.NewConfig(&t)
	d := task.NewDocker(config)

	// stop container
	result := d.Stop(t.ContainerID)
	if result.Error != nil {
		log.Printf("Error stopping container %v: %v\n", t.ContainerID, result.Error)
	}

	// update task
	t.FinishTime = time.Now().UTC()
	t.State = task.Completed
	w.Db[t.ID] = &t
	log.Printf("Stopped and removed container %v for task %v\n", t.ContainerID, t.ID)

	return result
}
