// Example: openapi/chi - openapi.Router mounted inside a chi router.
//
// chi handles global middleware; the openapi.Router handles routing,
// input decoding and spec generation internally. No WithPathValueFn needed.
//
// Run from this directory:
//
//	go run .
//	open http://localhost:9091
package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
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
	ID string `path:"id" doc:"Task ID"`
}
type ListInput struct {
	Status string `query:"status" doc:"Filter by status" enum:"open,done"`
}
type CreateInput struct {
	Title string `json:"title" example:"Buy milk"`
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
	t := Task{ID: fmt.Sprintf("%d", len(tasks)+1), Title: in.Title, Status: "open", CreatedAt: time.Now()}
	tasks = append(tasks, t)
	return &t, nil
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
		openapi.Info{Title: "Tasks API (chi)", Version: "1.0.0"},
		openapi.WithServer("http://localhost:9091", "Local"),
		openapi.WithSecurityScheme("BearerAuth", openapi.BearerAuth),
	)
	openapi.GETWithInput[ListInput, []Task](api, "/tasks", listTasks,
		openapi.Summary("List tasks"), openapi.Tags("tasks"))
	openapi.POST[CreateInput, Task](api, "/tasks", createTask,
		openapi.Summary("Create a task"), openapi.Tags("tasks"))
	openapi.GETWithInput[TaskIDParam, Task](api, "/tasks/{id}", getTask,
		openapi.Summary("Get a task"), openapi.Tags("tasks"))
	openapi.DELETE[TaskIDParam](api, "/tasks/{id}", deleteTask,
		openapi.Summary("Delete a task"), openapi.Tags("tasks"))

	ui, err := scalar.New(scalar.WithSpecURL("/openapi.json"))
	if err != nil {
		log.Fatal(err)
	}

	// chi router wrapping the openapi.Router.
	// chi middleware (Logger, Recoverer) applies to all requests.
	// r.Mount("/", api) delegates everything to the openapi.Router,
	// which uses its own stdlib mux for dispatch and path params.
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Handle("/docs", ui)
	r.Handle("/scalar.js", ui)
	r.Mount("/", api)

	log.Println("listening on http://localhost:9091  (docs:/docs  spec:/openapi.json)")
	log.Fatal(http.ListenAndServe(":9091", r))
}
