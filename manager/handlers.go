package manager

import (
	"cube/task"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// handle create task request.
func (a *Api) StartTaskHandler(w http.ResponseWriter, r *http.Request) {
	// decode request body.
	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields()

	// convert to task event.
	te := task.TaskEvent{}
	err := d.Decode(&te)

	if err != nil {
		msg := fmt.Sprintf("Error unmarshalling body: %v\n", err)
		log.Printf(msg)
		w.WriteHeader(400)
		e := ErrResponse{
			HttpStatusCode: 400,
			Message:        msg,
		}
		json.NewEncoder(w).Encode(e)
		return
	}

	// add task to manager.
	a.Manager.AddTask(te)
	log.Printf("Added task %v\n", te.Task.ID)
	w.WriteHeader(201)
	json.NewEncoder(w).Encode(te.Task)
}

// handle get tasks request.
func (a *Api) GetTasksHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	json.NewEncoder(w).Encode(a.Manager.GetTasks())
}

// handle stop task request.
func (a *Api) StopTaskHandler(w http.ResponseWriter, r *http.Request) {
	// get task id from url param.
	taskID := chi.URLParam(r, "taskID")
	if taskID == "" {
		log.Printf("No taskID passed in request.\n")
		w.WriteHeader(400)
	}

	// get task from manager taskdb.
	tID, _ := uuid.Parse(taskID)
	taskToStop, err := a.Manager.TaskDb.Get(tID.String())
	if err != nil {
		log.Printf("No task with ID %v found", tID)
		w.WriteHeader(404)
	}

	// create completed task event.
	te := task.TaskEvent{
		ID:        uuid.New(),
		State:     task.Completed,
		Timestamp: time.Now(),
	}

	taskCopy := taskToStop.(*task.Task)
	taskCopy.State = task.Completed
	te.Task = *taskCopy

	a.Manager.AddTask(te)

	log.Printf("Added task event %v top stop task %v", te.ID, taskCopy.ID)
	w.WriteHeader(204)
}
