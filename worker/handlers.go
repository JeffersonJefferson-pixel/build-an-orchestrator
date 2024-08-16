package worker

import (
	"cube/task"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func (a *Api) StartTaskHandler(w http.ResponseWriter, r *http.Request) {
	// create json decoder
	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields()

	// decode body to task event
	te := task.TaskEvent{}
	err := d.Decode(&te)
	if err != nil {
		msg := fmt.Sprintf("Error unmarshalling body: %v\n", err)
		log.Print(msg)
		w.WriteHeader(400)
		e := ErrResponse{
			HttpStatusCode: 400,
			Message:        msg,
		}
		json.NewEncoder(w).Encode(e)
		return
	}

	// add task to worker's queue
	a.Worker.AddTask(te.Task)
	log.Printf("Added task %v\n", te.Task.ID)
	w.WriteHeader(201)
	json.NewEncoder(w).Encode(te.Task)
}

func (a *Api) GetTasksHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	json.NewEncoder(w).Encode(a.Worker.GetTasks())
}

func (a *Api) StopTaskHandler(w http.ResponseWriter, r *http.Request) {
	// extract task id from request path
	taskID := chi.URLParam(r, "taskID")
	if taskID == "" {
		log.Printf("No TaskID passed in rqeuest.\n")
		w.WriteHeader(400)
	}

	// convert task id to uuid
	tID, err := uuid.Parse(taskID)
	if err != nil {
		log.Printf("Unable to parse task id: %v\n", taskID)
	}

	// get task from worker's db
	taskToStop, err := a.Worker.GetTask(tID)
	if err != nil {
		log.Printf("%v", err)
		w.WriteHeader(404)
	}

	// add completed task to worker's queue
	taskCopy := *taskToStop
	taskCopy.State = task.Completed
	a.Worker.AddTask(taskCopy)

	log.Printf("Added task %v to stop container %v\n", taskToStop.ID, taskToStop.ContainerID)
	w.WriteHeader(204)
}
