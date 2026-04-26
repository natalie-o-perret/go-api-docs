// Example: openapi/goapi-gen — compile-time spec generation.
//
// goapi-gen analyses this file statically (go/ast + go/types) and writes
// openapi.json WITHOUT starting the server or running any code.
// The same spec is also served at runtime by the openapi.Router.
//
// Workflow:
//
//  1. Edit your routes / types here.
//  2. Run `go generate .` — goapi-gen rewrites openapi.json.
//  3. Commit openapi.json alongside your code (it's your contract).
//  4. CI can diff it to catch accidental spec changes.
//
// Run from this directory:
//
//	go generate .            # regenerate openapi.json from source
//	go run .                 # start server; spec also served at /openapi.json
//	open http://localhost:9098
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

// ── Domain types ──────────────────────────────────────────────────────────────

type Task struct {
	ID        string    `json:"id"                  readOnly:"true"  example:"task_1"`
	Title     string    `json:"title"               doc:"Task title" example:"Buy milk"  minLength:"1" maxLength:"200"`
	Status    string    `json:"status"                               example:"open"     enum:"open,in_progress,done"`
	Priority  string    `json:"priority,omitempty"                  example:"medium"   enum:"low,medium,high"`
	CreatedAt time.Time `json:"createdAt"           readOnly:"true"`
}

type CreateTaskInput struct {
	Title    string `json:"title"              doc:"Task title"    minLength:"1" maxLength:"200"`
	Priority string `json:"priority,omitempty" doc:"Task priority" enum:"low,medium,high"`
}

type UpdateTaskInput struct {
	ID       string  `path:"id"                   doc:"Task ID"`
	Title    *string `json:"title,omitempty"      doc:"New title"    maxLength:"200"`
	Status   *string `json:"status,omitempty"     doc:"New status"   enum:"open,in_progress,done"`
	Priority *string `json:"priority,omitempty"   doc:"New priority" enum:"low,medium,high"`
}

type TaskIDParam struct {
	ID string `path:"id" doc:"Task ID" example:"task_1"`
}

type ListTasksInput struct {
	Status   string `query:"status"   doc:"Filter by status"   enum:"open,in_progress,done"`
	Priority string `query:"priority" doc:"Filter by priority" enum:"low,medium,high"`
	Limit    int    `query:"limit"    doc:"Max results"        min:"1" max:"100"`
}

// ── In-memory store ───────────────────────────────────────────────────────────

var tasks = []Task{
	{ID: "1", Title: "Buy milk", Status: "open", Priority: "low", CreatedAt: time.Now().Add(-2 * time.Hour)},
	{ID: "2", Title: "Write tests", Status: "in_progress", Priority: "high", CreatedAt: time.Now().Add(-1 * time.Hour)},
	{ID: "3", Title: "Deploy to prod", Status: "done", Priority: "high", CreatedAt: time.Now()},
}

func nextID() string { return fmt.Sprintf("%d", len(tasks)+1) }

// ── Handlers ──────────────────────────────────────────────────────────────────

func listTasks(_ *http.Request, in *ListTasksInput) (*[]Task, error) {
	var out []Task
	for _, t := range tasks {
		if in.Status != "" && t.Status != in.Status {
			continue
		}
		if in.Priority != "" && t.Priority != in.Priority {
			continue
		}
		out = append(out, t)
		if in.Limit > 0 && len(out) >= in.Limit {
			break
		}
	}
	return &out, nil
}

func createTask(_ *http.Request, in *CreateTaskInput) (*Task, error) {
	if in.Title == "" {
		return nil, openapi.ErrBadRequest("title is required")
	}
	t := Task{
		ID:        nextID(),
		Title:     in.Title,
		Priority:  in.Priority,
		Status:    "open",
		CreatedAt: time.Now(),
	}
	tasks = append(tasks, t)
	return &t, nil
}

func getTask(_ *http.Request, in *TaskIDParam) (*Task, error) {
	for _, t := range tasks {
		if t.ID == in.ID {
			return &t, nil
		}
	}
	return nil, openapi.ErrNotFound("task not found")
}

func updateTask(_ *http.Request, in *UpdateTaskInput) (*Task, error) {
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
		if in.Priority != nil {
			tasks[i].Priority = *in.Priority
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

// ── Main ──────────────────────────────────────────────────────────────────────

func main() {
	api := openapi.New(
		openapi.Info{
			Title:       "Tasks API (goapi-gen)",
			Version:     "1.0.0",
			Description: "Spec generated at compile time by goapi-gen AND served at runtime. Both outputs are identical.",
		},
		openapi.WithServer("http://localhost:9098", "Local dev"),
		openapi.WithSecurityScheme("BearerAuth", openapi.BearerAuth),
		openapi.WithTag("tasks", "Task management operations"),
	)

	openapi.GETWithInput[ListTasksInput, []Task](api, "/tasks", listTasks,
		openapi.Summary("List tasks"),
		openapi.Tags("tasks"),
		openapi.Security("BearerAuth"),
	)
	openapi.POST[CreateTaskInput, Task](api, "/tasks", createTask,
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
	openapi.PATCH[UpdateTaskInput, Task](api, "/tasks/{id}", updateTask,
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

	mux := http.NewServeMux()
	mux.Handle("/docs", ui)
	mux.Handle("/scalar.js", ui)
	mux.Handle("/", api)

	log.Println("listening on http://localhost:9098  (docs:/docs  spec:/openapi.json)")
	log.Println("tip: run `go generate .` to regenerate openapi.json from source")
	log.Fatal(http.ListenAndServe(":9098", mux))
}
