// Example: swagger — Swagger 2.0 spec + live mock API.
//
// Run from this directory:
//
//	go run .
//	open http://localhost:8081
package main

import (
	_ "embed"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	scalar "github.com/nopereta/go-scalar"
)

//go:embed swagger.json
var specJSON []byte

// ── helpers ──────────────────────────────────────────────────────────────────

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func requireBearer(w http.ResponseWriter, r *http.Request) bool {
	if !strings.HasPrefix(r.Header.Get("Authorization"), "Bearer ") {
		writeJSON(w, http.StatusUnauthorized, map[string]any{
			"code":    "unauthorized",
			"message": "missing or invalid bearer token",
		})
		return false
	}
	return true
}

// ── seed data ────────────────────────────────────────────────────────────────

var (
	now = time.Now().UTC().Truncate(time.Second)

	tasks = []map[string]any{
		{
			"id": "task_1", "title": "Migrate DB to Postgres 16",
			"description": "Upgrade from Postgres 14; run migration scripts in staging first.",
			"status": "open", "priority": "high",
			"assigneeId": "usr_alice",
			"dueAt":      now.Add(24 * time.Hour).Format(time.RFC3339),
			"createdAt":  now.Add(-48 * time.Hour).Format(time.RFC3339),
			"updatedAt":  now.Add(-48 * time.Hour).Format(time.RFC3339),
		},
		{
			"id": "task_2", "title": "Write API changelog for v1→v2",
			"description": "Document breaking changes between Swagger v1 and OpenAPI v2.",
			"status": "in_progress", "priority": "medium",
			"assigneeId": "usr_bob",
			"dueAt":      nil,
			"createdAt":  now.Add(-10 * time.Hour).Format(time.RFC3339),
			"updatedAt":  now.Add(-2 * time.Hour).Format(time.RFC3339),
		},
	}

	users = map[string]map[string]any{
		"usr_alice": {"id": "usr_alice", "name": "Alice", "email": "alice@acme.example.com", "createdAt": now.Add(-8760 * time.Hour).Format(time.RFC3339)},
		"usr_bob":   {"id": "usr_bob", "name": "Bob", "email": "bob@acme.example.com", "createdAt": now.Add(-4380 * time.Hour).Format(time.RFC3339)},
	}
)

// ── handlers ─────────────────────────────────────────────────────────────────

func handleListTasks(w http.ResponseWriter, r *http.Request) {
	if !requireBearer(w, r) {
		return
	}
	statusFilter := r.URL.Query().Get("status")
	result := []map[string]any{}
	for _, t := range tasks {
		if statusFilter == "" || t["status"] == statusFilter {
			result = append(result, t)
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"items": result, "total": len(result), "page": 1, "limit": 20,
	})
}

func handleCreateTask(w http.ResponseWriter, r *http.Request) {
	if !requireBearer(w, r) {
		return
	}
	var body map[string]any
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body["title"] == nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"code": "invalid_body", "message": "title is required"})
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{
		"id": "task_new", "title": body["title"],
		"status": "open", "priority": "medium",
		"createdAt": now.Format(time.RFC3339), "updatedAt": now.Format(time.RFC3339),
	})
}

func handleGetTask(w http.ResponseWriter, r *http.Request) {
	if !requireBearer(w, r) {
		return
	}
	id := r.PathValue("id")
	for _, t := range tasks {
		if t["id"] == id {
			writeJSON(w, http.StatusOK, t)
			return
		}
	}
	writeJSON(w, http.StatusNotFound, map[string]any{"code": "not_found", "message": "task not found"})
}

func handleReplaceTask(w http.ResponseWriter, r *http.Request) {
	if !requireBearer(w, r) {
		return
	}
	id := r.PathValue("id")
	var body map[string]any
	_ = json.NewDecoder(r.Body).Decode(&body)
	body["id"] = id
	body["updatedAt"] = now.Format(time.RFC3339)
	if body["createdAt"] == nil {
		body["createdAt"] = now.Format(time.RFC3339)
	}
	writeJSON(w, http.StatusOK, body)
}

func handleDeleteTask(w http.ResponseWriter, r *http.Request) {
	if !requireBearer(w, r) {
		return
	}
	id := r.PathValue("id")
	for _, t := range tasks {
		if t["id"] == id {
			w.WriteHeader(http.StatusNoContent)
			return
		}
	}
	writeJSON(w, http.StatusNotFound, map[string]any{"code": "not_found", "message": "task not found"})
}

func handleGetMe(w http.ResponseWriter, r *http.Request) {
	if !requireBearer(w, r) {
		return
	}
	writeJSON(w, http.StatusOK, users["usr_alice"])
}

func handleGetUser(w http.ResponseWriter, r *http.Request) {
	if !requireBearer(w, r) {
		return
	}
	id := r.PathValue("id")
	if u, ok := users[id]; ok {
		writeJSON(w, http.StatusOK, u)
		return
	}
	writeJSON(w, http.StatusNotFound, map[string]any{"code": "not_found", "message": "user not found"})
}

// ── main ─────────────────────────────────────────────────────────────────────

func main() {
	h, err := scalar.New(
		scalar.WithSpecURL("/swagger.json"),
		scalar.WithTheme(scalar.ThemeSolarized),
		scalar.WithBranding(scalar.Branding{
			Title:    "Acme Corp",
			Subtitle: "Tasks API v1 — Swagger 2.0",
		}),
		scalar.WithEnvBadge("legacy"),
		scalar.WithBaseServerURL("http://localhost:8081"),
		scalar.With(scalar.DisableAgent, scalar.DisableMCP, scalar.DarkMode),
		scalar.WithPageTitle("Tasks API v1 (Swagger 2.0)"),
	)
	if err != nil {
		log.Fatal(err)
	}

	mux := http.NewServeMux()
	mux.Handle("/", h)

	mux.HandleFunc("GET /swagger.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(specJSON)
	})

	mux.HandleFunc("GET /v1/tasks", handleListTasks)
	mux.HandleFunc("POST /v1/tasks", handleCreateTask)
	mux.HandleFunc("GET /v1/tasks/{id}", handleGetTask)
	mux.HandleFunc("PUT /v1/tasks/{id}", handleReplaceTask)
	mux.HandleFunc("DELETE /v1/tasks/{id}", handleDeleteTask)
	mux.HandleFunc("GET /v1/users/me", handleGetMe)
	mux.HandleFunc("GET /v1/users/{id}", handleGetUser)

	log.Println("listening on http://localhost:8081")
	log.Println("hint: use any string starting with 'Bearer ' as the auth token")
	log.Fatal(http.ListenAndServe(":8081", mux))
}

