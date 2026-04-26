// Example: full — everything turned on, pink-accented theme.
//
// Showcases: branding header, env badge, WithCustomTheme (pink accents),
// WithBaseServerURL, all boolean flags — plus a live mock API.
//
// The OpenAPI spec is built programmatically in buildSpec() using the same
// domain constants used by the handlers — no separate JSON blob to maintain.
//
// Run from this directory:
//
// go run .
// open http://localhost:8084
package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	scalar "github.com/nopereta/go-api-docs/ui/scalar"
)

// ── domain enums — shared between spec builder and handlers ──────────────────
var (
	taskStatuses = []string{"open", "in_progress", "done"}
	priorities   = []string{"low", "medium", "high"}
)

// ── spec (generated) ─────────────────────────────────────────────────────────
var specJSON = buildSpec()

func buildSpec() []byte {
	// ── micro-helpers ────────────────────────────────────────────────────────
	ref := func(name string) map[string]any {
		return map[string]any{"$ref": "#/components/schemas/" + name}
	}
	respRef := func(name string) map[string]any {
		return map[string]any{"$ref": "#/components/responses/" + name}
	}
	// prop builds a property schema; extra maps are merged in.
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
	reqBody := func(schema map[string]any) map[string]any {
		return map[string]any{
			"required": true,
			"content":  map[string]any{"application/json": map[string]any{"schema": schema}},
		}
	}
	resp := func(desc string, schema map[string]any) map[string]any {
		return map[string]any{
			"description": desc,
			"content":     map[string]any{"application/json": map[string]any{"schema": schema}},
		}
	}
	bearer := []map[string]any{{"bearerAuth": []string{}}}
	pathID := []map[string]any{{
		"name": "id", "in": "path", "required": true,
		"schema": map[string]any{"type": "string", "format": "uuid"},
	}}
	spec := map[string]any{
		"openapi": "3.1.0",
		"info": map[string]any{
			"title":       "Tasks API",
			"version":     "2.0.0",
			"description": "A simple task-management API — used as the go-scalar full example.",
			"contact":     map[string]any{"name": "Acme Platform", "url": "https://acme.example.com"},
			"license":     map[string]any{"name": "MIT"},
		},
		"servers": []map[string]any{
			{"url": "http://localhost:8084", "description": "Local dev"},
			{"url": "https://api.acme.example.com/v2", "description": "Production"},
		},
		"tags": []map[string]any{
			{"name": "tasks", "description": "Task operations"},
			{"name": "users", "description": "User operations"},
		},
		"paths": map[string]any{
			"/tasks": map[string]any{
				"get": map[string]any{
					"operationId": "listTasks", "summary": "List tasks",
					"tags": []string{"tasks"}, "security": bearer,
					"parameters": []map[string]any{
						{"name": "status", "in": "query", "schema": prop("string", enum(taskStatuses))},
						{"name": "assignee", "in": "query", "schema": prop("string", map[string]any{"format": "uuid"})},
						{"name": "limit", "in": "query", "schema": prop("integer", map[string]any{"default": 20, "maximum": 100})},
						{"name": "cursor", "in": "query", "description": "Opaque pagination cursor.", "schema": prop("string")},
					},
					"responses": map[string]any{
						"200": resp("Paginated list of tasks", ref("TaskPage")),
						"401": respRef("Unauthorized"),
					},
				},
				"post": map[string]any{
					"operationId": "createTask", "summary": "Create a task",
					"tags": []string{"tasks"}, "security": bearer,
					"requestBody": reqBody(ref("TaskInput")),
					"responses": map[string]any{
						"201": resp("Task created", ref("Task")),
						"400": respRef("BadRequest"), "401": respRef("Unauthorized"),
					},
				},
			},
			"/tasks/{id}": map[string]any{
				"parameters": pathID,
				"get": map[string]any{
					"operationId": "getTask", "summary": "Get a task",
					"tags": []string{"tasks"}, "security": bearer,
					"responses": map[string]any{
						"200": resp("Task found", ref("Task")),
						"404": respRef("NotFound"),
					},
				},
				"patch": map[string]any{
					"operationId": "updateTask", "summary": "Update a task",
					"tags": []string{"tasks"}, "security": bearer,
					"requestBody": reqBody(ref("TaskPatch")),
					"responses": map[string]any{
						"200": resp("Task updated", ref("Task")),
						"404": respRef("NotFound"),
					},
				},
				"delete": map[string]any{
					"operationId": "deleteTask", "summary": "Delete a task",
					"tags": []string{"tasks"}, "security": bearer,
					"responses": map[string]any{
						"204": map[string]any{"description": "Task deleted"},
						"404": respRef("NotFound"),
					},
				},
			},
			"/tasks/{id}/comments": map[string]any{
				"parameters": pathID,
				"get": map[string]any{
					"operationId": "listComments", "summary": "List comments on a task",
					"tags": []string{"tasks"}, "security": bearer,
					"responses": map[string]any{
						"200": resp("List of comments", map[string]any{"type": "array", "items": ref("Comment")}),
					},
				},
				"post": map[string]any{
					"operationId": "addComment", "summary": "Add a comment to a task",
					"tags": []string{"tasks"}, "security": bearer,
					"requestBody": reqBody(ref("CommentInput")),
					"responses": map[string]any{
						"201": resp("Comment added", ref("Comment")),
					},
				},
			},
			"/users/me": map[string]any{
				"get": map[string]any{
					"operationId": "getMe", "summary": "Get current user",
					"tags": []string{"users"}, "security": bearer,
					"responses": map[string]any{
						"200": resp("Current authenticated user", ref("User")),
						"401": respRef("Unauthorized"),
					},
				},
			},
			"/users/{id}": map[string]any{
				"parameters": pathID,
				"get": map[string]any{
					"operationId": "getUser", "summary": "Get a user by ID",
					"tags": []string{"users"}, "security": bearer,
					"responses": map[string]any{
						"200": resp("User found", ref("User")),
						"404": respRef("NotFound"),
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
					"type": "object", "required": []string{"id", "title", "status", "createdAt"},
					"properties": map[string]any{
						"id":          prop("string", map[string]any{"format": "uuid", "readOnly": true}),
						"title":       prop("string", map[string]any{"minLength": 1, "maxLength": 255}),
						"description": prop("string"),
						"status":      prop("string", enum(taskStatuses)),
						"priority":    prop("string", enum(priorities), map[string]any{"default": "medium"}),
						"assignee":    ref("UserRef"),
						"dueAt":       prop("string", map[string]any{"format": "date-time", "nullable": true}),
						"createdAt":   prop("string", map[string]any{"format": "date-time", "readOnly": true}),
						"updatedAt":   prop("string", map[string]any{"format": "date-time", "readOnly": true}),
					},
				},
				"TaskInput": map[string]any{
					"type": "object", "required": []string{"title"},
					"properties": map[string]any{
						"title":       prop("string", map[string]any{"minLength": 1, "maxLength": 255}),
						"description": prop("string"),
						"priority":    prop("string", enum(priorities), map[string]any{"default": "medium"}),
						"assigneeId":  prop("string", map[string]any{"format": "uuid"}),
						"dueAt":       prop("string", map[string]any{"format": "date-time", "nullable": true}),
					},
				},
				"TaskPatch": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"title":       prop("string", map[string]any{"minLength": 1, "maxLength": 255}),
						"description": prop("string"),
						"status":      prop("string", enum(taskStatuses)),
						"priority":    prop("string", enum(priorities)),
						"assigneeId":  prop("string", map[string]any{"format": "uuid", "nullable": true}),
						"dueAt":       prop("string", map[string]any{"format": "date-time", "nullable": true}),
					},
				},
				"TaskPage": map[string]any{
					"type": "object", "required": []string{"items"},
					"properties": map[string]any{
						"items":      map[string]any{"type": "array", "items": ref("Task")},
						"nextCursor": prop("string", map[string]any{"nullable": true}),
					},
				},
				"Comment": map[string]any{
					"type": "object", "required": []string{"id", "body", "author", "createdAt"},
					"properties": map[string]any{
						"id":        prop("string", map[string]any{"format": "uuid", "readOnly": true}),
						"body":      prop("string", map[string]any{"minLength": 1}),
						"author":    ref("UserRef"),
						"createdAt": prop("string", map[string]any{"format": "date-time", "readOnly": true}),
					},
				},
				"CommentInput": map[string]any{
					"type": "object", "required": []string{"body"},
					"properties": map[string]any{
						"body": prop("string", map[string]any{"minLength": 1}),
					},
				},
				"User": map[string]any{
					"type": "object", "required": []string{"id", "name", "email"},
					"properties": map[string]any{
						"id":        prop("string", map[string]any{"format": "uuid", "readOnly": true}),
						"name":      prop("string"),
						"email":     prop("string", map[string]any{"format": "email"}),
						"avatarURL": prop("string", map[string]any{"format": "uri", "nullable": true}),
						"createdAt": prop("string", map[string]any{"format": "date-time", "readOnly": true}),
					},
				},
				"UserRef": map[string]any{
					"type": "object", "required": []string{"id", "name"},
					"properties": map[string]any{
						"id":   prop("string", map[string]any{"format": "uuid"}),
						"name": prop("string"),
					},
				},
				"Error": map[string]any{
					"type": "object", "required": []string{"code", "message"},
					"properties": map[string]any{
						"code":    prop("string"),
						"message": prop("string"),
						"details": map[string]any{"type": "object", "additionalProperties": true},
					},
				},
			},
			"responses": map[string]any{
				"BadRequest":   resp("Invalid request", ref("Error")),
				"Unauthorized": resp("Missing or invalid credentials", ref("Error")),
				"NotFound":     resp("Resource not found", ref("Error")),
			},
		},
	}
	b, err := json.MarshalIndent(spec, "", "  ")
	if err != nil {
		panic("buildSpec: " + err.Error())
	}
	return b
}

// ── helpers ───────────────────────────────────────────────────────────────────
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

// ── seed data ─────────────────────────────────────────────────────────────────
var (
	now   = time.Now().UTC().Truncate(time.Second)
	later = now.Add(48 * time.Hour)
	alice = map[string]any{"id": "usr_alice", "name": "Alice"}
	bob   = map[string]any{"id": "usr_bob", "name": "Bob"}
	tasks = []map[string]any{
		{
			"id": "task_1", "title": "Design new onboarding flow",
			"description": "Revamp the sign-up screens based on last quarter's feedback.",
			"status":      taskStatuses[1], "priority": priorities[2],
			"assignee":  alice,
			"dueAt":     later.Format(time.RFC3339),
			"createdAt": now.Add(-72 * time.Hour).Format(time.RFC3339),
			"updatedAt": now.Add(-1 * time.Hour).Format(time.RFC3339),
		},
		{
			"id": "task_2", "title": "Fix CSV export encoding",
			"description": "Non-ASCII characters are corrupted on Windows.",
			"status":      taskStatuses[0], "priority": priorities[1],
			"assignee":  bob,
			"dueAt":     nil,
			"createdAt": now.Add(-24 * time.Hour).Format(time.RFC3339),
			"updatedAt": now.Add(-24 * time.Hour).Format(time.RFC3339),
		},
		{
			"id": "task_3", "title": "Upgrade Go toolchain to 1.26",
			"description": "Bump go.mod and run full test suite.",
			"status":      taskStatuses[2], "priority": priorities[0],
			"assignee":  alice,
			"dueAt":     nil,
			"createdAt": now.Add(-120 * time.Hour).Format(time.RFC3339),
			"updatedAt": now.Add(-6 * time.Hour).Format(time.RFC3339),
		},
	}
	comments = map[string][]map[string]any{
		"task_1": {
			{"id": "cmt_1", "body": "Mockups are ready in Figma.", "author": alice, "createdAt": now.Add(-2 * time.Hour).Format(time.RFC3339)},
			{"id": "cmt_2", "body": "LGTM, going ahead.", "author": bob, "createdAt": now.Add(-30 * time.Minute).Format(time.RFC3339)},
		},
	}
	users = map[string]map[string]any{
		"usr_alice": {"id": "usr_alice", "name": "Alice", "email": "alice@acme.example.com", "avatarURL": nil, "createdAt": now.Add(-8760 * time.Hour).Format(time.RFC3339)},
		"usr_bob":   {"id": "usr_bob", "name": "Bob", "email": "bob@acme.example.com", "avatarURL": nil, "createdAt": now.Add(-4380 * time.Hour).Format(time.RFC3339)},
	}
)

// ── handlers ──────────────────────────────────────────────────────────────────
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
	writeJSON(w, http.StatusOK, map[string]any{"items": result, "nextCursor": nil})
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
		"description": body["description"],
		"status":      taskStatuses[0], "priority": priorities[1],
		"assignee": nil, "dueAt": nil,
		"createdAt": now.Format(time.RFC3339),
		"updatedAt": now.Format(time.RFC3339),
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
func handleUpdateTask(w http.ResponseWriter, r *http.Request) {
	if !requireBearer(w, r) {
		return
	}
	id := r.PathValue("id")
	for _, t := range tasks {
		if t["id"] == id {
			var patch map[string]any
			_ = json.NewDecoder(r.Body).Decode(&patch)
			merged := map[string]any{}
			for k, v := range t {
				merged[k] = v
			}
			for k, v := range patch {
				merged[k] = v
			}
			merged["updatedAt"] = now.Format(time.RFC3339)
			writeJSON(w, http.StatusOK, merged)
			return
		}
	}
	writeJSON(w, http.StatusNotFound, map[string]any{"code": "not_found", "message": "task not found"})
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
func handleListComments(w http.ResponseWriter, r *http.Request) {
	if !requireBearer(w, r) {
		return
	}
	id := r.PathValue("id")
	c := comments[id]
	if c == nil {
		c = []map[string]any{}
	}
	writeJSON(w, http.StatusOK, c)
}
func handleAddComment(w http.ResponseWriter, r *http.Request) {
	if !requireBearer(w, r) {
		return
	}
	var body map[string]any
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body["body"] == nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"code": "invalid_body", "message": "body is required"})
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{
		"id": "cmt_new", "body": body["body"],
		"author":    alice,
		"createdAt": now.Format(time.RFC3339),
	})
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

// ── main ──────────────────────────────────────────────────────────────────────
func main() {
	h, err := scalar.New(
		scalar.WithSpecURL("/openapi.json"),
		// Pink-accented dark theme — Scalar CSS variables via the builder,
		// plus raw CSS for header elements that need class-specific rules.
		scalar.WithTheme(scalar.ThemeNone),
		scalar.WithCustomTheme(
			scalar.NewCustomTheme().
				Text("#ff2d78").
				Accent("#ff2d78").
				SidebarAccent("#ff2d78").
				SidebarItemActiveBackground("rgba(255,45,120,.12)"),
		),
		scalar.WithCustomCSS(`
/* branded header: hot-pink gradient */
.gs-header {
  background: linear-gradient(135deg, #1a0010 0%, #2d0020 100%) !important;
  border-bottom: 1px solid rgba(255,45,120,.35) !important;
}
.gs-brand-title    { color: #ff2d78 !important; }
.gs-brand-subtitle { color: #ffaacb !important; }
.gs-env-badge {
  background: rgba(255,45,120,.18) !important;
  color: #ff2d78 !important;
  border-color: rgba(255,45,120,.45) !important;
}`),
		// Branded header
		scalar.WithBranding(scalar.Branding{
			LogoURL:     "https://fakeimg.pl/32x32/ff2d78/ffffff?text=A&font=lobster",
			LogoAlt:     "Acme",
			Title:       "Acme Corp",
			Subtitle:    "Platform API",
			FaviconURL:  "https://fakeimg.pl/32x32/ff2d78/ffffff?text=A",
			FaviconType: "image/png",
		}),
		scalar.WithEnvBadge("staging"),
		// Features
		scalar.WithBaseServerURL("http://localhost:8084"),
		scalar.WithLayout(scalar.LayoutModern),
		scalar.WithShowOperationID(),
		scalar.With(scalar.DisableAgent, scalar.DisableMCP, scalar.DarkMode),
		scalar.WithShowDeveloperTools(scalar.ShowToolbarLocalhost),
		scalar.WithPageTitle("Acme Platform API"),
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
	mux.HandleFunc("GET /tasks", handleListTasks)
	mux.HandleFunc("POST /tasks", handleCreateTask)
	mux.HandleFunc("GET /tasks/{id}", handleGetTask)
	mux.HandleFunc("PATCH /tasks/{id}", handleUpdateTask)
	mux.HandleFunc("DELETE /tasks/{id}", handleDeleteTask)
	mux.HandleFunc("GET /tasks/{id}/comments", handleListComments)
	mux.HandleFunc("POST /tasks/{id}/comments", handleAddComment)
	mux.HandleFunc("GET /users/me", handleGetMe)
	mux.HandleFunc("GET /users/{id}", handleGetUser)
	log.Println("listening on http://localhost:8084")
	log.Println("hint: use any string starting with 'Bearer ' as the auth token")
	log.Fatal(http.ListenAndServe(":8084", mux))
}
