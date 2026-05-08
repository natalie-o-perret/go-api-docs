package openapi_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/nopereta/go-api-docs/openapi"
)

// ── domain types used across tests ───────────────────────────────────────────

type Task struct {
	CreatedAt time.Time `json:"createdAt,omitempty"`
	ID        string    `json:"id"                  example:"task_1"              readOnly:"true"`
	Title     string    `json:"title"               example:"Design onboarding"`
	Status    string    `json:"status"              example:"open"                enum:"open,in_progress,done"`
}

type TaskInput struct {
	Title  string `json:"title"            example:"My new task"`
	Status string `json:"status,omitempty" example:"open"          enum:"open,in_progress,done"`
}

type GetTaskInput struct {
	ID string `path:"id" doc:"Task ID"`
}

type ListTasksInput struct {
	Status string `query:"status" doc:"Filter by status" enum:"open,in_progress,done"`
}

type CreateTaskInput struct {
	// path/query params + body fields in the same struct
	ProjectID string `path:"projectId" doc:"Project ID"`
	Title     string `json:"title"     example:"My task"`
}

// ── helper ────────────────────────────────────────────────────────────────────

func newRouter() *openapi.Router {
	return openapi.New(openapi.Info{Title: "Test API", Version: "1.0.0"},
		openapi.WithServer("http://localhost:8080", "local"),
		openapi.WithSecurityScheme("BearerAuth", openapi.BearerAuth),
	)
}

func specFrom(r *openapi.Router) map[string]any {
	var m map[string]any
	_ = json.Unmarshal(r.OpenAPI(), &m)
	return m
}

// ── tests ─────────────────────────────────────────────────────────────────────

func TestRouter_serveOpenAPIJSON(t *testing.T) {
	r := newRouter()
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/openapi.json", http.NoBody))

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	ct := rec.Header().Get("Content-Type")
	if !strings.HasPrefix(ct, "application/json") {
		t.Errorf("unexpected Content-Type: %q", ct)
	}
	var doc map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &doc); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if doc["openapi"] != "3.1.0" {
		t.Errorf("expected openapi 3.1.0, got %v", doc["openapi"])
	}
}

func TestGET_noParams(t *testing.T) {
	r := newRouter()
	openapi.GET[[]Task](r, "/tasks", func(_ *http.Request) (*[]Task, error) {
		tasks := []Task{{ID: "1", Title: "t1", Status: "open"}}
		return &tasks, nil
	}, openapi.Summary("List tasks"), openapi.Tags("tasks"))

	// ── handler works ─────────────────────────────────────────────────────────
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/tasks", http.NoBody))
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	// ── spec ──────────────────────────────────────────────────────────────────
	spec := specFrom(r)
	paths := spec["paths"].(map[string]any)
	if _, ok := paths["/tasks"]; !ok {
		t.Error("expected /tasks in spec paths")
	}
}

func TestPOST_bodyDecoding(t *testing.T) {
	r := newRouter()
	openapi.POST[TaskInput, Task](r, "/tasks", func(_ *http.Request, in *TaskInput) (*Task, error) {
		return &Task{ID: "new", Title: in.Title, Status: "open"}, nil
	}, openapi.Summary("Create task"))

	body, _ := json.Marshal(TaskInput{Title: "write tests"})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/tasks", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.ContentLength = int64(len(body))
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", rec.Code)
	}
	var got Task
	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if got.Title != "write tests" {
		t.Errorf("expected title 'write tests', got %q", got.Title)
	}
}

func TestHandle_pathParam(t *testing.T) {
	r := newRouter()
	openapi.Handle[GetTaskInput, Task](r, http.MethodGet, "/tasks/{id}",
		func(_ *http.Request, in *GetTaskInput) (*Task, error) {
			return &Task{ID: in.ID, Title: "found", Status: "open"}, nil
		},
		openapi.Summary("Get task"),
	)

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/tasks/task_42", http.NoBody))

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var got Task
	_ = json.NewDecoder(rec.Body).Decode(&got)
	if got.ID != "task_42" {
		t.Errorf("expected id task_42, got %q", got.ID)
	}
}

func TestHandle_queryParam(t *testing.T) {
	r := newRouter()
	openapi.Handle[ListTasksInput, []Task](r, http.MethodGet, "/tasks",
		func(_ *http.Request, in *ListTasksInput) (*[]Task, error) {
			tasks := []Task{{ID: "1", Status: in.Status}}
			return &tasks, nil
		},
	)

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/tasks?status=done", http.NoBody))

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var got []Task
	_ = json.NewDecoder(rec.Body).Decode(&got)
	if len(got) == 0 || got[0].Status != "done" {
		t.Errorf("expected status 'done', got %+v", got)
	}
}

func TestDELETE_returns204(t *testing.T) {
	r := newRouter()
	openapi.DELETE[GetTaskInput](r, "/tasks/{id}",
		func(_ *http.Request, in *GetTaskInput) error {
			if in.ID == "missing" {
				return openapi.ErrNotFound("task not found")
			}
			return nil
		},
	)

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodDelete, "/tasks/task_1", http.NoBody))
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}
}

func TestDELETE_errNotFound(t *testing.T) {
	r := newRouter()
	openapi.DELETE[GetTaskInput](r, "/tasks/{id}",
		func(_ *http.Request, _ *GetTaskInput) error {
			return openapi.ErrNotFound("task not found")
		},
	)

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodDelete, "/tasks/missing", http.NoBody))
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestSpec_containsPathParam(t *testing.T) {
	r := newRouter()
	openapi.Handle[GetTaskInput, Task](r, http.MethodGet, "/tasks/{id}", func(_ *http.Request, _ *GetTaskInput) (*Task, error) {
		return nil, nil
	}, openapi.Summary("Get task"), openapi.Tags("tasks"), openapi.Security("BearerAuth"))

	raw := r.OpenAPI()
	if !strings.Contains(string(raw), `"path"`) {
		t.Error("expected path parameter in spec")
	}
	if !strings.Contains(string(raw), `"BearerAuth"`) {
		t.Error("expected security requirement in spec")
	}
}

func TestSpec_containsRequestBody(t *testing.T) {
	r := newRouter()
	openapi.POST[TaskInput, Task](r, "/tasks", func(_ *http.Request, _ *TaskInput) (*Task, error) {
		return nil, nil
	})

	raw := r.OpenAPI()
	if !strings.Contains(string(raw), `"requestBody"`) {
		t.Error("expected requestBody in spec")
	}
}

func TestSpec_schemaRegisteredInComponents(t *testing.T) {
	r := newRouter()
	openapi.GET[Task](r, "/tasks/{id}", func(_ *http.Request) (*Task, error) {
		return nil, nil
	})

	spec := specFrom(r)
	components, _ := spec["components"].(map[string]any)
	schemas, _ := components["schemas"].(map[string]any)
	if _, ok := schemas["Task"]; !ok {
		t.Error("expected Task registered in components/schemas")
	}
}

func TestSpec_mixedParamAndBody(t *testing.T) {
	r := newRouter()
	openapi.POST[CreateTaskInput, Task](r, "/projects/{projectId}/tasks",
		func(_ *http.Request, in *CreateTaskInput) (*Task, error) {
			return &Task{ID: "new", Title: in.Title, Status: "open"}, nil
		},
		openapi.Summary("Create task in project"),
	)

	raw := r.OpenAPI()
	if !strings.Contains(string(raw), `"projectId"`) {
		t.Error("expected projectId path param in spec")
	}
	if !strings.Contains(string(raw), `"requestBody"`) {
		t.Error("expected requestBody in spec (body fields from mixed struct)")
	}

	// Decode actual request to confirm path param + body both decoded.
	body, _ := json.Marshal(map[string]string{"title": "my task"})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/projects/proj_1/tasks", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.ContentLength = int64(len(body))
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
}

// ── additional coverage ───────────────────────────────────────────────────────

type UpdateInput struct {
	Title *string `json:"title,omitempty"`
	ID    string  `path:"id"`
}

func TestPUT_replaces(t *testing.T) {
	r := newRouter()
	openapi.PUT[TaskInput, Task](r, "/tasks/{id}", func(_ *http.Request, in *TaskInput) (*Task, error) {
		return &Task{ID: "1", Title: in.Title, Status: "open"}, nil
	})
	body, _ := json.Marshal(TaskInput{Title: "updated"})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/tasks/1", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.ContentLength = int64(len(body))
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestPATCH_partialUpdate(t *testing.T) {
	r := newRouter()
	openapi.PATCH[UpdateInput, Task](r, "/tasks/{id}", func(_ *http.Request, in *UpdateInput) (*Task, error) {
		title := ""
		if in.Title != nil {
			title = *in.Title
		}
		return &Task{ID: in.ID, Title: title, Status: "open"}, nil
	})
	title := "patched"
	body, _ := json.Marshal(map[string]string{"title": title})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPatch, "/tasks/42", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.ContentLength = int64(len(body))
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestGETWithInput_queryFiltered(t *testing.T) {
	r := newRouter()
	openapi.GETWithInput[ListTasksInput, []Task](r, "/items", func(_ *http.Request, in *ListTasksInput) (*[]Task, error) {
		out := []Task{{ID: "1", Status: in.Status}}
		return &out, nil
	})
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/items?status=open", http.NoBody))
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var got []Task
	_ = json.NewDecoder(rec.Body).Decode(&got)
	if len(got) == 0 || got[0].Status != "open" {
		t.Errorf("unexpected body: %+v", got)
	}
}

func TestSpec_routeOptions(t *testing.T) {
	r := openapi.New(
		openapi.Info{Title: "Options API", Version: "2.0.0"},
		openapi.WithTag("items", "item ops"),
		openapi.WithSecurityScheme("ApiKey", openapi.APIKeyHeader("X-API-Key")),
		openapi.WithSecurityScheme("Basic", openapi.BasicAuth),
		openapi.WithSecurityScheme("ApiKeyQ", openapi.APIKeyQuery("api_key")),
	)
	openapi.GET[Task](r, "/items/{id}", func(_ *http.Request) (*Task, error) { return nil, nil },
		openapi.Summary("Get item"),
		openapi.Description("Returns one item by ID."),
		openapi.OperationID("getItem"),
		openapi.Tags("items"),
		openapi.Security("ApiKey"),
		openapi.Deprecated(),
		openapi.Responses(map[string]openapi.Response{
			"404": {Description: "Not found"},
		}),
	)

	raw := string(r.OpenAPI())
	for _, want := range []string{
		`"operationId"`, `"getItem"`,
		`"description"`,
		`"deprecated"`,
		`"404"`,
		`"ApiKey"`,
		`"ApiKeyQ"`,
		`"Basic"`,
		`"items"`,
	} {
		if !strings.Contains(raw, want) {
			t.Errorf("spec missing %q", want)
		}
	}
}

func TestSpec_withPathValueFn(t *testing.T) {
	r := openapi.New(
		openapi.Info{Title: "Custom", Version: "1.0.0"},
		openapi.WithPathValueFn(func(r *http.Request, name string) string {
			return r.Header.Get("X-Path-" + name)
		}),
	)
	openapi.GETWithInput[GetTaskInput, Task](r, "/tasks/{id}", func(_ *http.Request, in *GetTaskInput) (*Task, error) {
		return &Task{ID: in.ID, Title: "t", Status: "open"}, nil
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/tasks/anything", http.NoBody)
	req.Header.Set("X-Path-id", "injected_id")
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var got Task
	_ = json.NewDecoder(rec.Body).Decode(&got)
	if got.ID != "injected_id" {
		t.Errorf("expected injected_id, got %q", got.ID)
	}
}

func TestErrors_helpers(t *testing.T) {
	r := newRouter()
	openapi.GET[Task](r, "/unauth", func(_ *http.Request) (*Task, error) {
		return nil, openapi.ErrUnauthorized("no token")
	})
	openapi.GET[Task](r, "/bad", func(_ *http.Request) (*Task, error) {
		return nil, openapi.ErrBadRequest("bad input")
	})
	openapi.GET[Task](r, "/custom", func(_ *http.Request) (*Task, error) {
		return nil, openapi.Err(http.StatusTeapot, "teapot", "I'm a teapot")
	})

	for _, tc := range []struct {
		path string
		code int
	}{
		{"/unauth", http.StatusUnauthorized},
		{"/bad", http.StatusBadRequest},
		{"/custom", http.StatusTeapot},
	} {
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, tc.path, http.NoBody))
		if rec.Code != tc.code {
			t.Errorf("%s: expected %d, got %d", tc.path, tc.code, rec.Code)
		}
	}
}

// validatingInput uses the Validator interface to reject empty titles.
type validatingInput struct {
	Title string `json:"title"`
}

func (v *validatingInput) Validate() error {
	if v.Title == "" {
		return openapi.ErrBadRequest("title required")
	}
	return nil
}

func TestValidator_interface(t *testing.T) {
	r := newRouter()
	openapi.POST[validatingInput, Task](r, "/validated", func(_ *http.Request, in *validatingInput) (*Task, error) {
		return &Task{ID: "1", Title: in.Title, Status: "open"}, nil
	})

	// Empty title → Validate() should return 400.
	body, _ := json.Marshal(map[string]string{"title": ""})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/validated", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.ContentLength = int64(len(body))
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400 from Validate(), got %d", rec.Code)
	}

	// Valid title → 201.
	body, _ = json.Marshal(map[string]string{"title": "hello"})
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/validated", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.ContentLength = int64(len(body))
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}
}

// schemaProviderTask uses SchemaProvider to bypass reflection.
type schemaProviderTask struct {
	ID string
}

func (schemaProviderTask) OpenAPISchema() openapi.Schema {
	return openapi.Schema{
		Type:       "object",
		Properties: map[string]openapi.Schema{"id": {Type: "string"}},
		Required:   []string{"id"},
	}
}

func TestSchemaProvider_bypassesReflection(t *testing.T) {
	r := newRouter()
	openapi.GET[schemaProviderTask](r, "/sp", func(_ *http.Request) (*schemaProviderTask, error) {
		return &schemaProviderTask{ID: "x"}, nil
	})
	raw := string(r.OpenAPI())

	// SchemaProvider bypasses reflection — the schema is inlined in the response,
	// NOT registered as a $ref in components/schemas.
	// The field name comes from OpenAPISchema() ("id"), not from reflect ("ID").
	if !strings.Contains(raw, `"id"`) {
		t.Error("expected 'id' property from SchemaProvider.OpenAPISchema()")
	}
	spec := specFrom(r)
	components, _ := spec["components"].(map[string]any)
	schemas, _ := components["schemas"].(map[string]any)
	if _, ok := schemas["schemaProviderTask"]; ok {
		t.Error("SchemaProvider types must be inlined, not registered as $ref in components/schemas")
	}
}
