package main

import (
	"cube/task"
	"cube/worker"
	"fmt"
	"time"

	"github.com/golang-collections/collections/queue"
	"github.com/google/uuid"
)

func main() {
	// create worker
	db := make(map[uuid.UUID]*task.Task)
	w := worker.Worker{
		Queue: *queue.New(),
		Db:    db,
	}

	// create task
	t := task.Task{
		ID:    uuid.New(),
		Name:  "test-container-1",
		State: task.Scheduled,
		Image: "strm/helloworld-http",
	}

	// start task
	fmt.Println("starting task")
	w.AddTask(t)
	result := w.RunTask()
	if result.Error != nil {
		panic(result.Error)
	}

	t.ContainerID = result.ContainerId
	fmt.Printf("task %s is running in container %s\n", t.ID, t.ContainerID)
	fmt.Printf("Sleepy time")
	time.Sleep(time.Second * 30)

	// stop task
	fmt.Printf("stopping task %s\n", t.ID)
	t.State = task.Completed
	w.AddTask(t)
	result = w.RunTask()
	if result.Error != nil {
		panic(result.Error)
	}
}
