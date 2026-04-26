// Example: swagger/swag — spec auto-generated from Go code via swaggo/swag.
//
// The OpenAPI spec is derived entirely from annotation comments on the handler
// functions and the Go types below — no manual JSON or map[string]any required.
//
// Regenerate the spec after touching annotations:
//
//	make gen-swagger-swag
//	# or directly:
//	swag init --generalInfo main.go \
//	          --dir       example/swagger/swag \
//	          --output    example/swagger/swag/docs
//
// Run from this directory:
//
//	go run .
//	open http://localhost:9082
package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/nopereta/go-api-docs/swagger"
	// Blank import registers the generated spec with swag's global registry.
	_ "github.com/nopereta/go-api-docs/example/swagger/swag/docs"
	swagfiles "github.com/swaggo/files"
	httpSwagger "github.com/swaggo/http-swagger"
)

//	@title		Tasks API
//	@version	2.0.0
//	@description	A simple task-management service — swagger/swag example.
//	@host		localhost:9082
//	@BasePath	/
//	@securityDefinitions.apikey	BearerAuth
//	@in					header
//	@name					Authorization

// ── domain types (used by swag to emit $ref schemas) ─────────────────────────

// Task is a to-do item.
type Task struct {
	ID        string `json:"id"                  example:"task_1"`
	Title     string `json:"title"               example:"Design onboarding"`
	Status    string `json:"status"              example:"open"       enums:"open,in_progress,done"`
	CreatedAt string `json:"createdAt,omitempty" example:"2026-04-25T10:00:00Z" format:"date-time"`
}

// TaskInput is the request body for creating a task.
type TaskInput struct {
	Title  string `json:"title"           example:"My new task" binding:"required"`
	Status string `json:"status,omitempty" example:"open"         enums:"open,in_progress,done"`
}

// ErrorResponse is returned on 4xx/5xx responses.
type ErrorResponse struct {
	Code    string `json:"code"    example:"not_found"`
	Message string `json:"message" example:"task not found"`
}

// ── seed data ─────────────────────────────────────────────────────────────────

var now = time.Now().UTC().Truncate(time.Second)

var tasks = []Task{
	{ID: "task_1", Title: "Design new onboarding flow", Status: "in_progress", CreatedAt: now.Add(-72 * time.Hour).Format(time.RFC3339)},
	{ID: "task_2", Title: "Fix CSV export encoding", Status: "open", CreatedAt: now.Add(-24 * time.Hour).Format(time.RFC3339)},
	{ID: "task_3", Title: "Upgrade Go toolchain", Status: "done", CreatedAt: now.Add(-120 * time.Hour).Format(time.RFC3339)},
}

// ── helpers ───────────────────────────────────────────────────────────────────

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func requireBearer(w http.ResponseWriter, r *http.Request) bool {
	if !strings.HasPrefix(r.Header.Get("Authorization"), "Bearer ") {
		writeJSON(w, http.StatusUnauthorized, ErrorResponse{Code: "unauthorized", Message: "missing or invalid bearer token"})
		return false
	}
	return true
}

// ── handlers ──────────────────────────────────────────────────────────────────

// handleListTasks godoc
//
//	@Summary	List tasks
//	@Tags		tasks
//	@Security	BearerAuth
//	@Param		status	query		string		false	"Filter by status"	Enums(open, in_progress, done)
//	@Success	200		{array}		Task
//	@Failure	401		{object}	ErrorResponse
//	@Router		/tasks [get]
func handleListTasks(w http.ResponseWriter, r *http.Request) {
	if !requireBearer(w, r) {
		return
	}
	statusFilter := r.URL.Query().Get("status")
	result := []Task{}
	for _, t := range tasks {
		if statusFilter == "" || t.Status == statusFilter {
			result = append(result, t)
		}
	}
	writeJSON(w, http.StatusOK, result)
}

// handleCreateTask godoc
//
//	@Summary		Create a task
//	@Tags			tasks
//	@Security		BearerAuth
//	@Accept			json
//	@Produce		json
//	@Param			body	body		TaskInput	true	"Task to create"
//	@Success		201		{object}	Task
//	@Failure		400		{object}	ErrorResponse
//	@Failure		401		{object}	ErrorResponse
//	@Router			/tasks [post]
func handleCreateTask(w http.ResponseWriter, r *http.Request) {
	if !requireBearer(w, r) {
		return
	}
	var body TaskInput
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Title == "" {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Code: "invalid_body", Message: "title is required"})
		return
	}
	writeJSON(w, http.StatusCreated, Task{
		ID:        "task_new",
		Title:     body.Title,
		Status:    "open",
		CreatedAt: now.Format(time.RFC3339),
	})
}

// handleGetTask godoc
//
//	@Summary	Get a task
//	@Tags		tasks
//	@Security	BearerAuth
//	@Param		id	path		string	true	"Task ID"
//	@Success	200	{object}	Task
//	@Failure	404	{object}	ErrorResponse
//	@Router		/tasks/{id} [get]
func handleGetTask(w http.ResponseWriter, r *http.Request) {
	if !requireBearer(w, r) {
		return
	}
	id := r.PathValue("id")
	for _, t := range tasks {
		if t.ID == id {
			writeJSON(w, http.StatusOK, t)
			return
		}
	}
	writeJSON(w, http.StatusNotFound, ErrorResponse{Code: "not_found", Message: "task not found"})
}

// handleDeleteTask godoc
//
//	@Summary	Delete a task
//	@Tags		tasks
//	@Security	BearerAuth
//	@Param		id	path	string	true	"Task ID"
//	@Success	204
//	@Failure	404	{object}	ErrorResponse
//	@Router		/tasks/{id} [delete]
func handleDeleteTask(w http.ResponseWriter, r *http.Request) {
	if !requireBearer(w, r) {
		return
	}
	id := r.PathValue("id")
	for _, t := range tasks {
		if t.ID == id {
			w.WriteHeader(http.StatusNoContent)
			return
		}
	}
	writeJSON(w, http.StatusNotFound, ErrorResponse{Code: "not_found", Message: "task not found"})
}

// ── main ──────────────────────────────────────────────────────────────────────

func main() {
	h, err := swagger.New(
		swagger.WithSpecURL("/openapi.json"),
		swagger.WithPageTitle("Tasks API — auto-generated spec"),
		swagger.WithBranding(swagger.Branding{
			Title:    "Acme Corp",
			Subtitle: "Tasks API (swag)",
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

	mux := http.NewServeMux()
	mux.Handle("/", h)

	// Serve the auto-generated OpenAPI JSON — swag registers it in its global
	// registry; httpSwagger.WrapHandler / swagfiles makes it available at /swagger/
	// while we expose /openapi.json for our swagger.Handler.
	mux.Handle("/swagger/", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
	))
	mux.Handle("/swagger/", http.StripPrefix("/swagger/", http.FileServer(http.FS(swagfiles.FS))))

	// Expose the raw JSON for our Swagger UI handler.
	mux.HandleFunc("GET /openapi.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		http.ServeFile(w, r, "docs/swagger.json")
	})

	mux.HandleFunc("GET /tasks", handleListTasks)
	mux.HandleFunc("POST /tasks", handleCreateTask)
	mux.HandleFunc("GET /tasks/{id}", handleGetTask)
	mux.HandleFunc("DELETE /tasks/{id}", handleDeleteTask)

	log.Println("listening on http://localhost:9082")
	log.Println("hint: use any string starting with 'Bearer ' as the auth token")
	log.Fatal(http.ListenAndServe(":9082", mux))
}

