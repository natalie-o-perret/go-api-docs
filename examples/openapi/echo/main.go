// Example: openapi/echo - openapi.Router wrapped inside an Echo server.
//
// echo.WrapHandler converts the openapi.Router (http.Handler) into an
// echo.HandlerFunc. A wildcard route forwards every request to the
// openapi.Router, which internally dispatches via its own stdlib mux.
//
// Run from this directory:
//
//	go run .
//	open http://localhost:9094
package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/labstack/echo/v4/middleware"
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
		openapi.Info{Title: "Tasks API (echo)", Version: "1.0.0"},
		openapi.WithServer("http://localhost:9094", "Local"),
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

	// Echo server wrapping the openapi.Router.
	// echo.WrapHandler converts http.Handler to echo.HandlerFunc.
	// The openapi.Router internal stdlib mux handles dispatch and path params.
	e := echo.New()
	e.HideBanner = true
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.GET("/docs", echo.WrapHandler(ui))
	e.GET("/scalar.js", echo.WrapHandler(ui))
	e.Any("/*", echo.WrapHandler(api))

	log.Println("listening on http://localhost:9094  (docs:/docs  spec:/openapi.json)")
	log.Fatal(e.Start(":9094"))
}
