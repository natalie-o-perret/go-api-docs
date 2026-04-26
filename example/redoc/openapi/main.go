// Example: redoc/openapi — spec generated automatically from Go types and
// handler signatures, zero manual JSON required.
//
// Uses github.com/nopereta/go-api-docs/openapi for typed route registration
// and spec generation, then serves the result via the redoc package.
//
// Run from this directory:
//
//	go run .
//	open http://localhost:9084
package main

import (
	"log"
	"net/http"
	"time"

	"github.com/nopereta/go-api-docs/openapi"
	"github.com/nopereta/go-api-docs/ui/redoc"
)

// ── domain types ──────────────────────────────────────────────────────────────

type Task struct {
	ID        string    `json:"id"                  example:"task_1"      readOnly:"true"`
	Title     string    `json:"title"               example:"Design flow"`
	Status    string    `json:"status"              example:"open"         enum:"open,in_progress,done"`
	Priority  string    `json:"priority,omitempty"  example:"medium"       enum:"low,medium,high"`
	CreatedAt time.Time `json:"createdAt,omitempty"`
}

type TaskInput struct {
	Title    string `json:"title"              example:"My new task"`
	Status   string `json:"status,omitempty"   example:"open"    enum:"open,in_progress,done"`
	Priority string `json:"priority,omitempty" example:"medium"  enum:"low,medium,high"`
}

type PatchTaskInput struct {
	ID       string  `path:"id"                  doc:"Task ID"`
	Title    *string `json:"title,omitempty"`
	Status   *string `json:"status,omitempty"    enum:"open,in_progress,done"`
	Priority *string `json:"priority,omitempty"  enum:"low,medium,high"`
}

type TaskIDParam struct {
	ID string `path:"id" doc:"Task ID"`
}

type ListTasksInput struct {
	Status   string `query:"status"   doc:"Filter by status"   enum:"open,in_progress,done"`
	Priority string `query:"priority" doc:"Filter by priority" enum:"low,medium,high"`
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
	t := Task{ID: "task_new", Title: in.Title, Status: "open", Priority: in.Priority, CreatedAt: now}
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
	r := openapi.New(
		openapi.Info{
			Title:       "Tasks API",
			Version:     "2.0.0",
			Description: "A simple task-management API - spec auto-generated from Go types.",
		},
		openapi.WithServer("http://localhost:9084", "Local dev"),
		openapi.WithSecurityScheme("BearerAuth", openapi.BearerAuth),
		openapi.WithTag("tasks", "Task management"),
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
		openapi.Responses(map[string]openapi.Response{"404": {Description: "Not found"}}),
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

	ui, err := redoc.New(
		redoc.WithSpecURL("/openapi.json"),
		redoc.WithPageTitle("Tasks API - Redoc"),
		redoc.WithBranding(redoc.Branding{
			Title:    "Acme Corp",
			Subtitle: "Tasks API",
		}),
		redoc.WithEnvBadge("dev"),
		redoc.WithRequiredPropsFirst(),
		redoc.WithExpandResponses("200,201"),
		redoc.WithLazyRendering(),
	)
	if err != nil {
		log.Fatal(err)
	}

	mux := http.NewServeMux()
	mux.Handle("/redoc.js", ui)
	mux.Handle("/docs", ui)
	mux.Handle("/", r)

	log.Println("listening on http://localhost:9084")
	log.Println("  docs:  http://localhost:9084/docs")
	log.Println("  spec:  http://localhost:9084/openapi.json")
	log.Fatal(http.ListenAndServe(":9084", mux))
}
