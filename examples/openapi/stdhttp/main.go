// Example: openapi/stdhttp - openapi.Router used directly with net/http.
//
// This is the zero-dependency baseline: the openapi.Router IS an http.Handler
// and registers itself on the default stdlib mux. No framework adapter needed.
//
// It also demonstrates the compile-time path: running `go generate` will invoke
// goapi-gen and produce openapi.json from the Go source without executing any
// code.
//
// Run from this directory:
//
//	go run .
//	open http://localhost:9097
//
//go:generate goapi-gen -out openapi.json .
package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/nopereta/go-api-docs/openapi"
	"github.com/nopereta/go-api-docs/ui/scalar"
)

type Task struct {
	ID        string    `json:"id"        example:"task_1" readOnly:"true"`
	Title     string    `json:"title"     example:"Buy milk"`
	Status    string    `json:"status"    example:"open"   enum:"open,done"`
	CreatedAt time.Time `json:"createdAt"`
}

type TaskIDParam struct {
	ID string `path:"id" doc:"Task ID" example:"task_1"`
}
type ListInput struct {
	Status string `query:"status" doc:"Filter by status" enum:"open,done"`
}
type CreateInput struct {
	Title string `json:"title" doc:"Task title" example:"Buy milk" minLength:"1" maxLength:"200"`
}
type UpdateInput struct {
	ID     string  `path:"id"               doc:"Task ID"`
	Title  *string `json:"title,omitempty"  doc:"New title"  maxLength:"200"`
	Status *string `json:"status,omitempty" doc:"New status" enum:"open,done"`
}

var tasks = []Task{
	{ID: "1", Title: "Buy milk", Status: "open", CreatedAt: time.Now().Add(-1 * time.Hour)},
	{ID: "2", Title: "Write tests", Status: "done", CreatedAt: time.Now().Add(-2 * time.Hour)},
}

func listTasks(_ *http.Request, in *ListInput) (*[]Task, error) {
	var out []Task
	for _, t := range tasks {
		if in.Status == "" || t.Status == in.Status {
			out = append(out, t)
		}
	}
	return &out, nil
}

func getTask(_ *http.Request, in *TaskIDParam) (*Task, error) {
	for _, t := range tasks {
		if t.ID == in.ID {
			return &t, nil
		}
	}
	return nil, openapi.ErrNotFound("task not found")
}

func createTask(_ *http.Request, in *CreateInput) (*Task, error) {
	if in.Title == "" {
		return nil, openapi.ErrBadRequest("title is required")
	}
	t := Task{
		ID:        fmt.Sprintf("%d", len(tasks)+1),
		Title:     in.Title,
		Status:    "open",
		CreatedAt: time.Now(),
	}
	tasks = append(tasks, t)
	return &t, nil
}

func updateTask(_ *http.Request, in *UpdateInput) (*Task, error) {
	for i, t := range tasks {
		if t.ID != in.ID {
			continue
		}
		if in.Title != nil {
			tasks[i].Title = *in.Title
		}
		if in.Status != nil {
			tasks[i].Status = *in.Status
		}
		return &tasks[i], nil
	}
	return nil, openapi.ErrNotFound("task not found")
}

func deleteTask(_ *http.Request, in *TaskIDParam) error {
	for i, t := range tasks {
		if t.ID == in.ID {
			tasks = append(tasks[:i], tasks[i+1:]...)
			return nil
		}
	}
	return openapi.ErrNotFound("task not found")
}

func main() {
	api := openapi.New(
		openapi.Info{
			Title:       "Tasks API (net/http)",
			Version:     "1.0.0",
			Description: "Plain net/http — no framework adapter needed.",
		},
		openapi.WithServer("http://localhost:9097", "Local dev"),
		openapi.WithSecurityScheme("BearerAuth", openapi.BearerAuth),
		openapi.WithTag("tasks", "Task management"),
	)

	openapi.GETWithInput[ListInput, []Task](api, "/tasks", listTasks,
		openapi.Summary("List tasks"),
		openapi.Tags("tasks"),
		openapi.Security("BearerAuth"),
	)
	openapi.POST[CreateInput, Task](api, "/tasks", createTask,
		openapi.Summary("Create a task"),
		openapi.Tags("tasks"),
		openapi.Security("BearerAuth"),
	)
	openapi.GETWithInput[TaskIDParam, Task](api, "/tasks/{id}", getTask,
		openapi.Summary("Get a task"),
		openapi.Tags("tasks"),
		openapi.Security("BearerAuth"),
		openapi.Responses(map[string]openapi.Response{
			"404": {Description: "Task not found"},
		}),
	)
	openapi.PATCH[UpdateInput, Task](api, "/tasks/{id}", updateTask,
		openapi.Summary("Update a task"),
		openapi.Tags("tasks"),
		openapi.Security("BearerAuth"),
		openapi.Responses(map[string]openapi.Response{
			"404": {Description: "Task not found"},
		}),
	)
	openapi.DELETE[TaskIDParam](api, "/tasks/{id}", deleteTask,
		openapi.Summary("Delete a task"),
		openapi.Tags("tasks"),
		openapi.Security("BearerAuth"),
		openapi.Responses(map[string]openapi.Response{
			"404": {Description: "Task not found"},
		}),
	)

	ui, err := scalar.New(scalar.WithSpecURL("/openapi.json"))
	if err != nil {
		log.Fatal(err)
	}

	// Mount the Scalar UI at /docs; everything else (including /openapi.json)
	// is served by the openapi.Router directly on the default mux.
	mux := http.NewServeMux()
	mux.Handle("/docs", ui)
	mux.Handle("/scalar.js", ui)
	mux.Handle("/", api) // api serves GET /openapi.json + all /tasks routes

	log.Println("listening on http://localhost:9097  (docs:/docs  spec:/openapi.json)")
	log.Fatal(http.ListenAndServe(":9097", mux))
}
