// Example: full — everything turned on, pink-accented theme.
//
// Showcases: branding header, env badge, custom CSS (pink accents),
// WithBaseServerURL, and all boolean flags — plus a live mock API.
//
// Run from this directory:
//
//	go run .
//	open http://localhost:8084
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

//go:embed openapi.json
var specJSON []byte

// ── pink custom CSS ───────────────────────────────────────────────────────────

const pinkCSS = `
/* ── scalar variable overrides ── */
.light-mode, .dark-mode {
  --scalar-color-1:            #ff2d78;
  --scalar-color-accent:       #ff2d78;
  --scalar-button-1:           #ff2d78;
  --scalar-button-1-hover:     #e0005f;
  --scalar-sidebar-color-active: #ff2d78;
  --scalar-sidebar-background-active: rgba(255,45,120,.12);
}

/* ── branded header: hot-pink gradient ── */
.gs-header {
  background: linear-gradient(135deg, #1a0010 0%, #2d0020 100%) !important;
  border-bottom: 1px solid rgba(255,45,120,.35) !important;
}
.gs-brand-title  { color: #ff2d78 !important; }
.gs-brand-subtitle { color: #ffaacb !important; }
.gs-env-badge {
  background: rgba(255,45,120,.18) !important;
  color: #ff2d78 !important;
  border-color: rgba(255,45,120,.45) !important;
}
`

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
			"status": "in_progress", "priority": "high",
			"assignee":  alice,
			"dueAt":     later.Format(time.RFC3339),
			"createdAt": now.Add(-72 * time.Hour).Format(time.RFC3339),
			"updatedAt": now.Add(-1 * time.Hour).Format(time.RFC3339),
		},
		{
			"id": "task_2", "title": "Fix CSV export encoding",
			"description": "Non-ASCII characters are corrupted on Windows.",
			"status": "open", "priority": "medium",
			"assignee":  bob,
			"dueAt":     nil,
			"createdAt": now.Add(-24 * time.Hour).Format(time.RFC3339),
			"updatedAt": now.Add(-24 * time.Hour).Format(time.RFC3339),
		},
		{
			"id": "task_3", "title": "Upgrade Go toolchain to 1.26",
			"description": "Bump go.mod and run full test suite.",
			"status": "done", "priority": "low",
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
		"status": "open", "priority": "medium",
		"assignee":  nil, "dueAt": nil,
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

		// Pink-accented dark theme
		scalar.WithTheme(scalar.ThemeNone),
		scalar.WithCustomCSS(pinkCSS),

		// Branded header
		scalar.WithBranding(scalar.Branding{
			LogoURL:    "https://fakeimg.pl/32x32/ff2d78/ffffff?text=A&font=lobster",
			LogoAlt:    "Acme",
			Title:      "Acme Corp",
			Subtitle:   "Platform API",
			FaviconURL: "https://fakeimg.pl/32x32/ff2d78/ffffff?text=A",
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

