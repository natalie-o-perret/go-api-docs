// Example: openapi/httprouter - openapi.Router used alongside httprouter.
//
// httprouter has no generic http.Handler mount mechanism, so we use two
// complementary strategies:
//
//  1. Named routes are registered individually with httprouter so httprouter
//     can do its own fast radix-tree lookup. The openapi.Router with
//     WithPathValueFn(httprouterParam) is used solely as an http.Handler
//     adapter per route — path params flow from httprouter.Params.
//
//  2. Catch-all for /openapi.json and any docs: delegated to the openapi.Router
//     via router.NotFound.
//
// Because WithPathValueFn is set, decodeInput reads path params from
// httprouter.Params (stored in context) instead of r.PathValue.
//
// Run from this directory:
//
//	go run .
//	open http://localhost:9096
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/nopereta/go-api-docs/openapi"
	"github.com/nopereta/go-api-docs/ui/scalar"
)

// contextKey is a private type for storing httprouter.Params in context.
type contextKey struct{}

// withParams stores httprouter.Params in the request context so that
// the WithPathValueFn callback can retrieve them.
func withParams(r *http.Request, ps httprouter.Params) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), contextKey{}, ps))
}

// httprouterParam is the WithPathValueFn adapter:
// it reads path params from the httprouter.Params stored in context.
func httprouterParam(r *http.Request, name string) string {
	if ps, ok := r.Context().Value(contextKey{}).(httprouter.Params); ok {
		return ps.ByName(name)
	}
	return ""
}

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

// adapt wraps an http.Handler as an httprouter.Handle, injecting
// path params into the request context so that WithPathValueFn can
// read them.
func adapt(h http.Handler) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		h.ServeHTTP(w, withParams(r, ps))
	}
}

func main() {
	// Build the openapi.Router with a custom path-param extractor that
	// reads from httprouter.Params rather than r.PathValue.
	api := openapi.New(
		openapi.Info{Title: "Tasks API (httprouter)", Version: "1.0.0"},
		openapi.WithServer("http://localhost:9096", "Local"),
		openapi.WithSecurityScheme("BearerAuth", openapi.BearerAuth),
		openapi.WithPathValueFn(httprouterParam),
	)

	// Register routes on the openapi.Router for spec generation.
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

	// httprouter: register each route explicitly.
	// adapt() injects httprouter.Params into the request context.
	router := httprouter.New()

	router.GET("/tasks", adapt(api))
	router.POST("/tasks", adapt(api))
	router.GET("/tasks/:id", adapt(api))
	router.DELETE("/tasks/:id", adapt(api))

	// Serve the generated spec and Scalar UI via NotFound fallback.
	router.NotFound = api // handles /openapi.json
	router.GET("/docs", adapt(ui))
	router.GET("/scalar.js", adapt(ui))

	log.Println("listening on http://localhost:9096  (docs:/docs  spec:/openapi.json)")
	log.Fatal(http.ListenAndServe(":9096", router))
}
