// Example: swagger/openapi — spec generated automatically from Go types and
// handler signatures, zero manual JSON required.
//
// Uses github.com/nopereta/go-api-docs/openapi for typed route registration
// and spec generation, then serves the result via the swagger package.
//
// Run from this directory:
//
//	go run .
//	open http://localhost:9083
package main

import (
	"log"
	"net/http"
	"time"

	"github.com/nopereta/go-api-docs/openapi"
	"github.com/nopereta/go-api-docs/ui/swagger"
)

// ── domain types ─────────────────────────────────────────────────────────────

// Task is a to-do item.
type Task struct {
	CreatedAt time.Time `json:"createdAt,omitempty"`
	ID        string    `json:"id"                  example:"task_1"     readOnly:"true"`
	Title     string    `json:"title"               example:"Design flow"`
	Status    string    `json:"status"              example:"open"        enum:"open,in_progress,done"`
	Priority  string    `json:"priority,omitempty"  example:"medium"      enum:"low,medium,high"`
}

// TaskInput is the request body for creating or updating a task.
type TaskInput struct {
	Title    string `json:"title"             example:"My new task"`
	Status   string `json:"status,omitempty"  example:"open"         enum:"open,in_progress,done"`
	Priority string `json:"priority,omitempty" example:"medium"       enum:"low,medium,high"`
}

// TaskPatch is the request body for partial updates.
type TaskPatch struct {
	Title    *string `json:"title,omitempty"`
	Status   *string `json:"status,omitempty"  enum:"open,in_progress,done"`
	Priority *string `json:"priority,omitempty" enum:"low,medium,high"`
}

// TaskIDParam carries the path parameter for single-task endpoints.
type TaskIDParam struct {
	ID string `path:"id" doc:"Task ID"`
}

// ListTasksInput carries query params for listing tasks.
type ListTasksInput struct {
	Status   string `query:"status"   doc:"Filter by status"   enum:"open,in_progress,done"`
	Priority string `query:"priority" doc:"Filter by priority" enum:"low,medium,high"`
}

// PatchTaskInput combines the path param with a partial body.
type PatchTaskInput struct {
	Title    *string `json:"title,omitempty"`
	Status   *string `json:"status,omitempty"   enum:"open,in_progress,done"`
	Priority *string `json:"priority,omitempty" enum:"low,medium,high"`
	ID       string  `path:"id"              doc:"Task ID"`
}

// ── seed data ─────────────────────────────────────────────────────────────────

var now = time.Now().UTC().Truncate(time.Second)

var tasks = []Task{
	{ID: "task_1", Title: "Design new onboarding flow", Status: "in_progress", Priority: "high", CreatedAt: now.Add(-72 * time.Hour)},
	{ID: "task_2", Title: "Fix CSV export encoding", Status: "open", Priority: "medium", CreatedAt: now.Add(-24 * time.Hour)},
	{ID: "task_3", Title: "Upgrade Go toolchain", Status: "done", Priority: "low", CreatedAt: now.Add(-120 * time.Hour)},
}

// ── handlers ──────────────────────────────────────────────────────────────────

func listTasks(_ *http.Request, in *ListTasksInput) (*[]Task, error) {
	result := []Task{}
	for _, t := range tasks {
		if (in.Status == "" || t.Status == in.Status) &&
			(in.Priority == "" || t.Priority == in.Priority) {
			result = append(result, t)
		}
	}
	return &result, nil
}

func createTask(_ *http.Request, in *TaskInput) (*Task, error) {
	if in.Title == "" {
		return nil, openapi.ErrBadRequest("title is required")
	}
	t := Task{
		ID:        "task_new",
		Title:     in.Title,
		Status:    "open",
		Priority:  in.Priority,
		CreatedAt: now,
	}
	if t.Priority == "" {
		t.Priority = "medium"
	}
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

func patchTask(_ *http.Request, in *PatchTaskInput) (*Task, error) {
	for i, t := range tasks {
		if t.ID == in.ID {
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
	}
	return nil, openapi.ErrNotFound("task not found")
}

func deleteTask(_ *http.Request, in *TaskIDParam) error {
	for _, t := range tasks {
		if t.ID == in.ID {
			return nil
		}
	}
	return openapi.ErrNotFound("task not found")
}

// ── main ──────────────────────────────────────────────────────────────────────

func main() {
	// ── 1. typed router — spec is built here at startup ───────────────────────
	r := openapi.New(
		openapi.Info{
			Title:       "Tasks API",
			Version:     "2.0.0",
			Description: "A simple task-management API — spec auto-generated from Go types.",
		},
		openapi.WithServer("http://localhost:9083", "Local dev"),
		openapi.WithSecurityScheme("BearerAuth", openapi.BearerAuth),
	)

	openapi.GETWithInput[ListTasksInput, []Task](r, "/tasks", listTasks,
		openapi.Summary("List tasks"),
		openapi.Tags("tasks"),
		openapi.Security("BearerAuth"),
	)
	openapi.POST[TaskInput, Task](r, "/tasks", createTask,
		openapi.Summary("Create a task"),
		openapi.Tags("tasks"),
		openapi.Security("BearerAuth"),
	)
	openapi.GETWithInput[TaskIDParam, Task](r, "/tasks/{id}", getTask,
		openapi.Summary("Get a task"),
		openapi.Tags("tasks"),
		openapi.Security("BearerAuth"),
	)
	openapi.PATCH[PatchTaskInput, Task](r, "/tasks/{id}", patchTask,
		openapi.Summary("Update a task"),
		openapi.Tags("tasks"),
		openapi.Security("BearerAuth"),
	)
	openapi.DELETE[TaskIDParam](r, "/tasks/{id}", deleteTask,
		openapi.Summary("Delete a task"),
		openapi.Tags("tasks"),
		openapi.Security("BearerAuth"),
	)

	// ── 2. swagger UI — points at the /openapi.json served by the router ─────
	ui, err := swagger.New(
		swagger.WithSpecURL("/openapi.json"),
		swagger.WithPageTitle("Tasks API — auto-generated spec"),
		swagger.WithBranding(swagger.Branding{
			Title:    "Acme Corp",
			Subtitle: "Tasks API",
		}),
		swagger.WithEnvBadge("dev"),
		swagger.WithDarkMode(),
		swagger.WithPersistAuthorization(),
		swagger.WithDisplayRequestDuration(),
		swagger.WithTryItOutEnabled(),
	)
	if err != nil {
		log.Fatal(err)
	}

	// ── 3. mount: UI at root, API + spec on the same mux ─────────────────────
	mux := http.NewServeMux()

	// Swagger UI assets + HTML.
	mux.Handle("/swagger-ui-bundle.js", ui)
	mux.Handle("/swagger-ui-standalone-preset.js", ui)
	mux.Handle("/swagger-ui.css", ui)
	mux.Handle("/docs", ui)

	// The typed router handles /openapi.json + all API routes.
	mux.Handle("/", r)

	log.Println("listening on http://localhost:9083")
	log.Println("  docs:    http://localhost:9083/docs")
	log.Println("  spec:    http://localhost:9083/openapi.json")
	log.Fatal(http.ListenAndServe(":9083", mux))
}
