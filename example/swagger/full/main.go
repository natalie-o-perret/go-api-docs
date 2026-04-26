// Example: swagger/full — Swagger UI with branding, env badge, dark mode,
// persistent auth, request-duration display, and a live mock Tasks API.
//
// The OpenAPI 3.1 spec is built in-process (no external JSON file) using the
// same domain constants shared with the mock handlers.
//
// Run from this directory:
//
//	go run .
//	open http://localhost:9081
package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/nopereta/go-api-docs/ui/swagger"
)

// ── spec (generated) ─────────────────────────────────────────────────────────

var specJSON = buildSpec()

func buildSpec() []byte {
	ref := func(name string) map[string]any {
		return map[string]any{"$ref": "#/components/schemas/" + name}
	}
	prop := func(typ string, extras ...map[string]any) map[string]any {
		m := map[string]any{"type": typ}
		for _, e := range extras {
			for k, v := range e {
				m[k] = v
			}
		}
		return m
	}
	enum := func(vals []string) map[string]any { return map[string]any{"enum": vals} }
	bearer := []map[string]any{{"bearerAuth": []string{}}}

	spec := map[string]any{
		"openapi": "3.1.0",
		"info": map[string]any{
			"title":       "Tasks API",
			"version":     "2.0.0",
			"description": "A simple task-management API — swagger/full example.",
		},
		"servers": []map[string]any{
			{"url": "http://localhost:9081", "description": "Local dev"},
		},
		"paths": map[string]any{
			"/tasks": map[string]any{
				"get": map[string]any{
					"operationId": "listTasks", "summary": "List tasks",
					"security": bearer,
					"parameters": []map[string]any{
						{"name": "status", "in": "query", "schema": prop("string", enum([]string{"open", "in_progress", "done"}))},
					},
					"responses": map[string]any{
						"200": map[string]any{
							"description": "List of tasks",
							"content":     map[string]any{"application/json": map[string]any{"schema": map[string]any{"type": "array", "items": ref("Task")}}},
						},
						"401": map[string]any{"description": "Unauthorized"},
					},
				},
				"post": map[string]any{
					"operationId": "createTask", "summary": "Create a task",
					"security": bearer,
					"requestBody": map[string]any{
						"required": true,
						"content":  map[string]any{"application/json": map[string]any{"schema": ref("TaskInput")}},
					},
					"responses": map[string]any{
						"201": map[string]any{
							"description": "Created",
							"content":     map[string]any{"application/json": map[string]any{"schema": ref("Task")}},
						},
					},
				},
			},
			"/tasks/{id}": map[string]any{
				"parameters": []map[string]any{{"name": "id", "in": "path", "required": true, "schema": prop("string")}},
				"get": map[string]any{
					"operationId": "getTask", "summary": "Get a task",
					"security": bearer,
					"responses": map[string]any{
						"200": map[string]any{"description": "Task", "content": map[string]any{"application/json": map[string]any{"schema": ref("Task")}}},
						"404": map[string]any{"description": "Not found"},
					},
				},
				"delete": map[string]any{
					"operationId": "deleteTask", "summary": "Delete a task",
					"security": bearer,
					"responses": map[string]any{
						"204": map[string]any{"description": "Deleted"},
						"404": map[string]any{"description": "Not found"},
					},
				},
			},
		},
		"components": map[string]any{
			"securitySchemes": map[string]any{
				"bearerAuth": map[string]any{"type": "http", "scheme": "bearer", "bearerFormat": "JWT"},
			},
			"schemas": map[string]any{
				"Task": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"id":        prop("string"),
						"title":     prop("string"),
						"status":    prop("string", enum([]string{"open", "in_progress", "done"})),
						"createdAt": prop("string", map[string]any{"format": "date-time"}),
					},
				},
				"TaskInput": map[string]any{
					"type":     "object",
					"required": []string{"title"},
					"properties": map[string]any{
						"title":  prop("string"),
						"status": prop("string", enum([]string{"open", "in_progress", "done"})),
					},
				},
			},
		},
	}
	b, _ := json.MarshalIndent(spec, "", "  ")
	return b
}

// ── helpers ───────────────────────────────────────────────────────────────────

var now = time.Now().UTC().Truncate(time.Second)

var tasks = []map[string]any{
	{"id": "task_1", "title": "Design new onboarding flow", "status": "in_progress", "createdAt": now.Add(-72 * time.Hour).Format(time.RFC3339)},
	{"id": "task_2", "title": "Fix CSV export encoding", "status": "open", "createdAt": now.Add(-24 * time.Hour).Format(time.RFC3339)},
	{"id": "task_3", "title": "Upgrade Go toolchain", "status": "done", "createdAt": now.Add(-120 * time.Hour).Format(time.RFC3339)},
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func requireBearer(w http.ResponseWriter, r *http.Request) bool {
	if !strings.HasPrefix(r.Header.Get("Authorization"), "Bearer ") {
		writeJSON(w, http.StatusUnauthorized, map[string]any{"code": "unauthorized", "message": "missing or invalid bearer token"})
		return false
	}
	return true
}

// ── main ──────────────────────────────────────────────────────────────────────

func main() {
	h, err := swagger.New(
		swagger.WithSpecURL("/openapi.json"),
		swagger.WithPageTitle("Acme Tasks API — Swagger UI"),
		swagger.WithBranding(swagger.Branding{
			Title:    "Acme Corp",
			Subtitle: "Tasks API v2",
		}),
		swagger.WithEnvBadge("staging"),
		swagger.WithDarkMode(),
		swagger.WithPersistAuthorization(),
		swagger.WithDisplayRequestDuration(),
		swagger.WithTryItOutEnabled(),
		swagger.WithFilter(""),
		swagger.WithDocExpansion(swagger.DocExpansionList),
		swagger.WithDisplayOperationID(),
	)
	if err != nil {
		log.Fatal(err)
	}

	mux := http.NewServeMux()
	mux.Handle("/", h)

	mux.HandleFunc("GET /openapi.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(specJSON)
	})

	mux.HandleFunc("GET /tasks", func(w http.ResponseWriter, r *http.Request) {
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
		writeJSON(w, http.StatusOK, result)
	})

	mux.HandleFunc("POST /tasks", func(w http.ResponseWriter, r *http.Request) {
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
			"status": "open", "createdAt": now.Format(time.RFC3339),
		})
	})

	mux.HandleFunc("GET /tasks/{id}", func(w http.ResponseWriter, r *http.Request) {
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
	})

	mux.HandleFunc("DELETE /tasks/{id}", func(w http.ResponseWriter, r *http.Request) {
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
	})

	log.Println("listening on http://localhost:9081")
	log.Println("hint: use any string starting with 'Bearer ' as the auth token")
	log.Fatal(http.ListenAndServe(":9081", mux))
}
