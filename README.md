# go-api-docs

[![CI](https://github.com/natalie-o-perret/go-api-docs/actions/workflows/ci.yml/badge.svg)](https://github.com/natalie-o-perret/go-api-docs/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/nopereta/go-api-docs.svg)](https://pkg.go.dev/github.com/nopereta/go-api-docs)
[![Go Report Card](https://goreportcard.com/badge/github.com/nopereta/go-api-docs)](https://goreportcard.com/report/github.com/nopereta/go-api-docs)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

**The only Go library where the spec IS the code.**  
No annotations. No code generation. No CLI tools. No drift.

> [!NOTE]
> Unapologetically vibe-coded with Claude Opus 4.5.
>
> Every other Go OpenAPI solution forces you to maintain two sources of truth — your
> Go types and your spec. Rename a field, forget to update the YAML, and your API
> contract silently lies to every client. This library makes that impossible.

Three independent packages, zero mandatory dependencies beyond the stdlib:

| Package      | Import path                                  | What it does                                         |
|--------------|----------------------------------------------|------------------------------------------------------|
| `openapi`    | `github.com/nopereta/go-api-docs/openapi`    | Typed router that auto-generates an OpenAPI 3.1 spec |
| `ui/scalar`  | `github.com/nopereta/go-api-docs/ui/scalar`  | Scalar UI handler (vendored JS or CDN)               |
| `ui/swagger` | `github.com/nopereta/go-api-docs/ui/swagger` | Swagger UI handler (vendored bundle or CDN)          |

---

## Why

Every other Go OpenAPI solution has a fundamental problem:

| Approach         | The catch                                                                                                                                                                                                                                                              |
|------------------|------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| **swaggo/swag**  | Comments diverge from code silently. `swag init` must be re-run. Zero compile-time guarantees.                                                                                                                                                                         |
| **oapi-codegen** | You write YAML first. The generated code diverges between regenerations.                                                                                                                                                                                               |
| **huma v2**      | "Zero deps" but pulls in a framework adapter package. Context-based handlers feel alien.                                                                                                                                                                               |
| **Manual JSON**  | 100% accurate on day 1, 0% accurate on day 90. Every rename is a lie.                                                                                                                                                                                                  |
| **go-api-docs**  | Your Go types **are** the spec. Generics enforce handler signatures at compile time. Implement `SchemaProvider` for zero-reflection schemas. `Validator` for compile-time-safe request validation. `goapi-gen` for a fully static spec with no runtime at all. Always. |

---

## Quick start

### Option A — reflection (runtime, zero setup)

The router introspects your types **once at startup** via `reflect` and serves
`GET /openapi.json` automatically. Nothing to run, nothing to generate.

```go
import "github.com/nopereta/go-api-docs/openapi"

r := openapi.New(openapi.Info{Title: "Tasks API", Version: "1.0.0"})
openapi.POST[CreateTaskInput, Task](r, "/tasks", createTask)
http.Handle("/", r) // GET /openapi.json is served automatically
```

### Option B — `goapi-gen` (compile-time, static analysis)

`goapi-gen` walks your Go source files using `go/ast` + `go/types` and
produces `openapi.json` **without executing a single line of your code**.
It reads the same struct tags, the same generic type parameters, the same
route-option calls — statically, at "compile time".

```bash
# install once
go install github.com/nopereta/go-api-docs/cmd/goapi-gen@latest

# generate (run manually or via go generate)
goapi-gen -out openapi.json ./...
```

Add to any `.go` file and `go generate ./...` does the rest:

```go
//go:generate goapi-gen -out openapi.json .
```

`goapi-gen` recognises:

- `openapi.New(Info{…}, WithServer(…), WithTag(…), WithSecurityScheme(…))`
- `GET`, `GETWithInput`, `POST`, `PUT`, `PATCH`, `DELETE`, `Handle`
- All route options: `Summary`, `Description`, `Tags`, `OperationID`, `Security`, `Deprecated`, `Responses`
- All struct tags: `path`, `query`, `header`, `json`, `doc`, `example`, `enum`, `readOnly`, `writeOnly`, `format`,
  `pattern`, `min/maxLength`, `min/max`, `min/maxItems`
- `time.Time` → `string/date-time`, pointer types → nullable, slices → arrays, `$ref` dedup

> [!TIP]
> Both options produce **identical output**. Use the reflection approach during
> development (zero friction) and optionally add `go generate` to your CI
> pipeline to snapshot the spec as a file.

---

## Full example

```go
import "github.com/nopereta/go-api-docs/openapi"

r := openapi.New(
openapi.Info{Title: "Tasks API", Version: "1.0.0"},
openapi.WithServer("https://api.acme.com", "Production"),
openapi.WithSecurityScheme("BearerAuth", openapi.BearerAuth),
openapi.WithTag("tasks", "Task management operations"),
)

// Each struct field drives both HTTP decoding AND the OpenAPI spec.
// Change the type → the spec changes. Delete a field → the spec reflects it.
// There is no second source of truth.
type CreateTaskInput struct {
Title    string  `json:"title"              doc:"Task title"  minLength:"1" maxLength:"200"`
Status   string  `json:"status,omitempty"   enum:"open,done"`
Priority *string `json:"priority,omitempty" enum:"low,medium,high"`
}
type Task struct {
ID     string `json:"id"     readOnly:"true" example:"task_42"`
Title  string `json:"title"  example:"Write more tests"`
Status string `json:"status" enum:"open,done"`
}
type TaskIDParam struct {
ID string `path:"id" doc:"Task ID" example:"task_42"`
}
type ListInput struct {
Status string `query:"status" enum:"open,done"`
Limit  int    `query:"limit"  min:"1" max:"100"`
}

openapi.GETWithInput[ListInput, []Task](r, "/tasks", listTasks,
openapi.Summary("List tasks"),
openapi.Tags("tasks"),
openapi.Security("BearerAuth"),
openapi.Responses(map[string]openapi.Response{
"401": {Description: "Unauthorized"},
}),
)
openapi.POST[CreateTaskInput, Task](r, "/tasks", createTask,
openapi.Summary("Create a task"),
openapi.Tags("tasks"),
openapi.Security("BearerAuth"),
)
openapi.GETWithInput[TaskIDParam, Task](r, "/tasks/{id}", getTask,
openapi.Summary("Get a task"),
openapi.Tags("tasks"),
openapi.Security("BearerAuth"),
openapi.Responses(map[string]openapi.Response{
"404": {Description: "Task not found"},
}),
)

// GET /openapi.json served automatically. Mount r anywhere.
http.Handle("/", r)
```

### Serve Scalar UI

```go
import "github.com/nopereta/go-api-docs/ui/scalar"

h, _ := scalar.New(
scalar.WithSpecURL("/openapi.json"),
scalar.WithTheme(scalar.ThemeDefault),
scalar.With(scalar.DarkMode, scalar.DisableAgent),
scalar.WithBranding(scalar.Branding{Title: "Acme Corp", Subtitle: "Platform API"}),
)
http.Handle("/", h)
```

### Serve Swagger UI

```go
import "github.com/nopereta/go-api-docs/ui/swagger"

h, _ := swagger.New(
swagger.WithSpecURL("/openapi.json"),
swagger.WithDarkMode(),
swagger.WithPersistAuthorization(),
swagger.WithTryItOutEnabled(),
)
http.Handle("/", h)
```

---

## `openapi` — typed router + spec generator

### Input struct tags

| Tag                  | Location               | Notes                                            |
|----------------------|------------------------|--------------------------------------------------|
| `path:"name"`        | URL segment            | Always required in spec                          |
| `query:"name"`       | Query string           | Optional; add `required:"true"` to mark required |
| `header:"name"`      | HTTP header            | Optional; add `required:"true"` to mark required |
| *(no tag)*           | JSON body              | Decoded for POST/PUT/PATCH                       |
| `doc:"…"`            | any                    | Description in spec                              |
| `example:"…"`        | any                    | Example value                                    |
| `enum:"a,b,c"`       | string                 | Restricted values — validated in spec            |
| `readOnly:"true"`    | any                    | Read-only in response schema                     |
| `writeOnly:"true"`   | any                    | Write-only in request schema                     |
| `required:"false"`   | non-pointer body field | Mark body field as optional                      |
| `required:"true"`    | query/header           | Mark param as required in spec                   |
| `format:"uuid"`      | string                 | Override schema format                           |
| `pattern:"^[a-z]+$"` | string                 | Regexp constraint                                |
| `minLength:"1"`      | string                 | Minimum length                                   |
| `maxLength:"255"`    | string                 | Maximum length                                   |
| `min:"0"`            | number                 | Minimum value                                    |
| `max:"100"`          | number                 | Maximum value                                    |
| `minItems:"1"`       | array                  | Minimum item count                               |
| `maxItems:"50"`      | array                  | Maximum item count                               |

Mix path params, query params and body fields in the same struct — they're dispatched automatically:

```go
type UpdateInput struct {
ID     string  `path:"id"               doc:"Task ID"`
Title  *string `json:"title,omitempty"  maxLength:"200"`
Status *string `json:"status,omitempty" enum:"open,in_progress,done"`
}
openapi.PATCH[UpdateInput, Task](r, "/tasks/{id}", updateTask)
```

For cross-field or business-rule validation, implement `Validate()` on the input type — it is called
automatically after decode (see [Compile-time escape hatches](#compile-time-escape-hatches)).

### Registration functions

```go
openapi.GET[Out](r, path, fn, opts...) // no params
openapi.GETWithInput[In, Out](r, path, fn, opts...) // with path/query/header params
openapi.POST[In, Out](r, path, fn, opts...)
openapi.PUT[In, Out](r, path, fn, opts...)
openapi.PATCH[In, Out](r, path, fn, opts...)
openapi.DELETE[In](r, path, fn, opts...) // returns 204
openapi.Handle[In, Out](r, method, path, fn, opts...) // escape hatch
```

### Route options

```go
openapi.Summary("Get a task")
openapi.Description("Returns a single task by its ID.")
openapi.Tags("tasks")
openapi.OperationID("getTask")
openapi.Security("BearerAuth")
openapi.Deprecated()

// Document additional response codes beyond the auto-generated success response:
openapi.Responses(map[string]openapi.Response{
"404": {Description: "Task not found"},
"422": {Description: "Validation error"},
})
```

### Router options

```go
openapi.New(info,
openapi.WithServer("https://api.acme.com", "Production"),
openapi.WithSecurityScheme("BearerAuth", openapi.BearerAuth), // JWT Bearer
openapi.WithSecurityScheme("ApiKey", openapi.APIKeyHeader("X-API-Key")),
openapi.WithSecurityScheme("Basic", openapi.BasicAuth),
openapi.WithTag("tasks", "Task management operations"),
openapi.WithTag("billing", "Subscription and payment endpoints"),
openapi.WithPathValueFn(chi.URLParam), // only needed for httprouter-style routers
)
```

### Compile-time escape hatches

#### `SchemaProvider` — declare your own schema

By default the schema is derived automatically via reflection. If you want
full compile-time control over a type's schema — or want zero reflection for
that type — implement `SchemaProvider`:

```go
func (Task) OpenAPISchema() openapi.Schema {
return openapi.Schema{
Type: "object",
Properties: map[string]openapi.Schema{
"id":     {Type: "string", ReadOnly: true, Example: "task_42"},
"title":  {Type: "string"},
"status": {Type: "string", Enum: []any{"open", "done"}},
},
Required: []string{"id", "title", "status"},
}
}
```

The reflection path is bypassed entirely for opted-in types. The schema is guaranteed by the
compiler — rename a field and forget to update `OpenAPISchema`, the method still compiles,
but your IDE/linter will track the change. This is the closest Go gets to a true compile-time spec.

#### `Validator` — custom validation after decode

Implement `Validate() error` on any input struct and it will be called automatically after all
path/query/header params and the JSON body have been decoded. A non-nil error returns 400 Bad Request.

```go
func (in *CreateTaskInput) Validate() error {
if in.Title == "" {
return errors.New("title is required")
}
if len(in.Title) > 200 {
return errors.New("title exceeds 200 characters")
}
return nil
}
```

This is a compile-time contract: adding `Validate()` to a type automatically enables it —
no registration, no wiring, no annotations needed.

### Error handling

```go
func getTask(_ *http.Request, in *TaskIDParam) (*Task, error) {
task, ok := db.Find(in.ID)
if !ok {
return nil, openapi.ErrNotFound("task not found") // → 404
}
return task, nil
}
```

Helpers: `ErrNotFound` (404), `ErrBadRequest` (400), `ErrUnauthorized` (401),
`Err(status, code, message)` for anything else.

### Security scheme helpers

```go
openapi.BearerAuth // HTTP Bearer JWT
openapi.BasicAuth  // HTTP Basic
openapi.APIKeyHeader("X-API-Key") // API key in header
openapi.APIKeyQuery("api_key") // API key in query string
```

---

## `scalar` — Scalar UI handler

The handler serves the HTML page and the **vendored** `@scalar/api-reference` JS bundle from memory.
Point at a CDN instead with `WithJSPath`.

```go
h, err := scalar.New(
scalar.WithSources(
scalar.Source{URL: "/openapi.json", Title: "v2", Default: true},
scalar.Source{URL: "/openapi.v1.json", Title: "v1"},
),
scalar.WithTheme(scalar.ThemePurple),
scalar.WithBranding(scalar.Branding{
LogoURL:    "/logo.svg",
Title:      "Acme Corp",
Subtitle:   "Platform API",
FaviconURL: "/favicon.svg",
}),
scalar.WithEnvBadge("staging"),
scalar.WithBaseServerURL("https://staging.api.acme.com"),
scalar.With(scalar.DisableAgent, scalar.DisableMCP, scalar.DarkMode),
scalar.WithCustomTheme(
scalar.NewCustomTheme().
Accent("#2563eb").
Background("#0f172a").
Font("'Inter', sans-serif"),
),
)
// Serves:
//   GET /           → HTML page          (Cache-Control: no-cache)
//   GET /scalar.js  → vendored JS bundle (Cache-Control: immutable)
http.Handle("/", h)
```

Available themes: `ThemeDefault`, `ThemeAlternate`, `ThemeMoon`, `ThemePurple`,
`ThemeSolarized`, `ThemeBluePlanet`, `ThemeDeepSpace`, `ThemeSaturn`, `ThemeKepler`,
`ThemeMars`, `ThemeNone`.

Update the vendored bundle:

```bash
make vendor-js                     # uses pinned version (1.52.6)
make vendor-js VERSION=1.53.0      # pin a specific version
```

---

## `swagger` — Swagger UI handler

```go
h, err := swagger.New(
swagger.WithSpecURL("/openapi.json"),
swagger.WithDarkMode(),
swagger.WithBranding(swagger.Branding{Title: "Acme"}),
swagger.WithEnvBadge("dev"),
swagger.WithPersistAuthorization(),
swagger.WithDisplayRequestDuration(),
swagger.WithTryItOutEnabled(),
swagger.WithFilter(""),
swagger.WithDocExpansion(swagger.DocExpansionList),
)
// Serves:
//   GET /                                → HTML page    (Cache-Control: no-cache)
//   GET /swagger-ui-bundle.js            → vendored JS  (Cache-Control: immutable)
//   GET /swagger-ui-standalone-preset.js → vendored JS  (Cache-Control: immutable)
//   GET /swagger-ui.css                  → vendored CSS (Cache-Control: immutable)
http.Handle("/", h)
```

Update the vendored bundles:

```bash
make vendor-swagger-ui                       # uses pinned version (5.18.2)
make vendor-swagger-ui VERSION=5.19.0
```

---

## Framework adapters

`openapi.Router` implements `http.Handler` and mounts into any framework that
can wrap a standard handler — no code changes needed.

```go
// chi
r := chi.NewRouter()
r.Use(middleware.Logger)
r.Mount("/", api) // chi middleware wraps the whole thing

// gorilla/mux
r := mux.NewRouter()
r.PathPrefix("/").Handler(api)

// gin
g := gin.Default()
g.Any("/*path", gin.WrapH(api)) // gin.WrapH converts http.Handler

// echo
e := echo.New()
e.Any("/*", echo.WrapHandler(api)) // echo.WrapHandler does the same

// fiber (requires github.com/gofiber/adaptor/v2)
app := fiber.New()
app.All("/*", adaptor.HTTPHandler(api))

// httprouter — needs WithPathValueFn because its handler signature differs
type paramsKey struct{}
api := openapi.New(info, openapi.WithPathValueFn(func (r *http.Request, name string) string {
ps, _ := r.Context().Value(paramsKey{}).(httprouter.Params)
return ps.ByName(name)
}))
router := httprouter.New()
adapt := func (h http.Handler) httprouter.Handle {
return func (w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
h.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), paramsKey{}, ps)))
}
}
router.GET("/tasks", adapt(api))
router.POST("/tasks", adapt(api))
router.GET("/tasks/:id", adapt(api))
router.NotFound = api // /openapi.json + anything else
```

Because `openapi.Router` uses its own stdlib mux internally, path params are always set via
`r.PathValue()` — no `WithPathValueFn` needed for the mount-at-root pattern. Each framework's
middleware (auth, CORS, rate-limit, …) wraps the whole thing transparently.
See [`COMPARISON.md`](COMPARISON.md) for full details.

---

## Examples

| Path                         | Stack                 | What it shows                                                           |
|------------------------------|-----------------------|-------------------------------------------------------------------------|
| `example/scalar/basic`       | scalar                | Petstore proxied via localhost                                          |
| `example/scalar/cdn`         | scalar                | Load Scalar from jsDelivr CDN                                           |
| `example/scalar/full`        | scalar                | Pink theme, branding, env badge, live Tasks API                         |
| `example/scalar/multi-spec`  | scalar                | v1 + v2 dropdown, Swagger 2.0 + OpenAPI 3.1                             |
| `example/scalar/openpotato`  | scalar                | Two live public APIs, reverse-proxy                                     |
| `example/scalar/swagger2`    | scalar                | Scalar rendering a **Swagger 2.0** spec                                 |
| `example/swagger/basic`      | swagger               | Petstore via Swagger UI                                                 |
| `example/swagger/full`       | swagger               | Branding, dark mode, env badge, live Tasks API                          |
| `example/swagger/openapi`    | openapi + swagger     | **Spec auto-generated** from Go types, Tasks API                        |
| `example/openapi/stdhttp`    | openapi only          | Plain `net/http` — zero framework, also shows `//go:generate goapi-gen` |
| `example/openapi/chi`        | openapi + chi         | Typed router mounted inside chi                                         |
| `example/openapi/gorilla`    | openapi + gorilla/mux | Typed router mounted inside gorilla/mux                                 |
| `example/openapi/gin`        | openapi + gin         | Typed router wrapped with `gin.WrapH`                                   |
| `example/openapi/echo`       | openapi + echo        | Typed router wrapped with `echo.WrapHandler`                            |
| `example/openapi/fiber`      | openapi + fiber       | Typed router wrapped with `adaptor.HTTPHandler`                         |
| `example/openapi/httprouter` | openapi + httprouter  | `WithPathValueFn` + context adapter                                     |

Each example is a self-contained Go module:

```bash
cd example/openapi/stdhttp    && go run . && open http://localhost:9097
cd example/openapi/chi        && go run . && open http://localhost:9091
cd example/openapi/gorilla    && go run . && open http://localhost:9092
cd example/openapi/gin        && go run . && open http://localhost:9093
cd example/openapi/echo       && go run . && open http://localhost:9094
cd example/openapi/fiber      && go run . && open http://localhost:9095
cd example/openapi/httprouter && go run . && open http://localhost:9096
cd example/swagger/openapi    && go run . && open http://localhost:9083
```

---

## License

[MIT](LICENSE)
