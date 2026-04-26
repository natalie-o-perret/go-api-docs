# Comparison: go-api-docs vs. existing OpenAPI spec solutions for Go

## TL;DR

|                                            | **go-api-docs/openapi**     | huma v2      | swaggo/swag           | oapi-codegen    | Manual       |
|--------------------------------------------|-----------------------------|--------------|-----------------------|-----------------|--------------|
| Code generation step                       | ✗                           | ✗            | ✓ (`swag init`)       | ✓ (spec → code) | —            |
| Annotations required                       | ✗                           | ✗            | ✓ (`// @Summary ...`) | ✗               | —            |
| Spec direction                             | code → spec                 | code → spec  | code → spec           | spec → code     | hand-written |
| Runtime spec build                         | ✓                           | ✓            | ✗ (file)              | ✗ (file)        | —            |
| Go generics (type-safe handlers)           | ✓                           | ✓            | ✗                     | ✓               | —            |
| Zero external deps for spec gen            | ✓                           | ✗            | ✗                     | ✗               | —            |
| Framework agnostic                         | ✓ (any `http.Handler` host) | ✓ (adapters) | ✓                     | ✓               | —            |
| Built-in UI serving (Scalar + Swagger)     | ✓                           | ✗            | ✗                     | ✗               | —            |
| Auto param decode (path/query/header/body) | ✓                           | ✓            | ✗                     | ✓               | —            |
| Constraint tags (min/max/pattern/format)   | ✓                           | ✓            | ✗                     | ✓               | —            |
| Extra response codes in spec               | ✓ (`Responses(...)`)        | ✓            | ✓                     | ✓               | —            |
| Tag descriptions in spec                   | ✓ (`WithTag`)               | ✓            | ✓                     | ✓               | —            |
| required query/header params               | ✓ (`required:"true"`)       | ✓            | ✓                     | ✓               | —            |
| Pluggable path-param extractor             | ✓                           | n/a          | n/a                   | n/a             | —            |

---

## go-api-docs/openapi

```go
type TaskIDParam struct {
ID string `path:"id" doc:"Task ID"`
}
type Task struct {
ID    string `json:"id"    readOnly:"true"`
Title string `json:"title" example:"Write tests"`
Status string `json:"status" enum:"open,in_progress,done"`
}
r := openapi.New(openapi.Info{Title: "Tasks API", Version: "1.0.0"})
openapi.GETWithInput[TaskIDParam, Task](r, "/tasks/{id}", getTask,
openapi.Summary("Get a task"),
openapi.Tags("tasks"),
openapi.Security("BearerAuth"),
)
// GET /openapi.json served automatically.
// Mount r in any net/http-compatible framework.
```

**Strengths:**

- Zero dependencies for spec generation (pure stdlib + reflection)
- No code generation, no build step, no CLI tools
- Spec is always in sync with the code — it IS the code
- Single struct drives HTTP decoding, validation hints, and the full spec
- Constraint tags: `min`, `max`, `minLength`, `maxLength`, `pattern`, `format`, `minItems`, `maxItems`
- Required query/header params via `required:"true"`
- Extra response codes via `Responses(map[string]Response{...})`
- Tag descriptions via `WithTag(name, description)`
- Pre-built security schemes: `BearerAuth`, `BasicAuth`, `APIKeyHeader`, `APIKeyQuery`
- Pluggable path-param extraction — works with **any** Go HTTP framework
- Ships together with Scalar and Swagger UI handlers in the same module

**Limitations:**

- No request validation at runtime (struct tags drive spec only; add a validator if needed)
- No response headers in spec (planned)
- No `oneOf`/`anyOf`/`allOf` (set schema fields directly on the `Response` / `RequestBody` as escape hatch)

---

## huma v2 (`github.com/danielgtaylor/huma/v2`)

The closest alternative. Same generics-first, code-first philosophy.

```go
huma.Register(api, huma.Operation{
OperationID: "get-task",
Method:      http.MethodGet,
Path:        "/tasks/{id}",
Summary:     "Get a task",
}, func (ctx context.Context, input *struct {
ID string `path:"id" doc:"Task ID"`
}) (*struct{ Body Task }, error) {
...
})
```

**vs. go-api-docs/openapi:**

- huma requires importing `github.com/danielgtaylor/huma/v2` **and** a framework adapter
  (`humachi`, `humaecho`, `humagin`, …) — multiple packages
- go-api-docs/openapi has zero runtime dependencies and uses a one-line adapter for any framework
- huma has more features: response headers, content negotiation, `oneOf`, transforms
- go-api-docs/openapi is lighter and ships with UI serving in the same module
- huma uses `Context` in handler signature; go-api-docs/openapi uses `*http.Request` (familiar)

---

## swaggo/swag

Spec is generated from **comment annotations** above each handler function.
Requires running `swag init` (a code-generation step) to produce `docs/docs.go`.

```go
// @Summary  Get a task
// @Tags     tasks
// @Security BearerAuth
// @Param    id   path     string  true  "Task ID"
// @Success  200  {object} Task
// @Failure  404  {object} ErrorResponse
// @Router   /tasks/{id} [get]
func getTask(w http.ResponseWriter, r *http.Request) {
id := r.PathValue("id")
...
}
```

**vs. go-api-docs/openapi:**

- Comments drift silently out of sync with the actual handler types
- `swag init` must be re-run every time annotations change (CI step or forgotten locally)
- No type checking on annotations — wrong type names compile fine, break the spec silently
- `go-api-docs/openapi` is checked at compile time; if the type changes, the spec changes too
- swag is very widely adopted and has more ecosystem tooling

---

## oapi-codegen (`github.com/oapi-codegen/oapi-codegen`)

**Spec-first**: you write the OpenAPI YAML file; oapi-codegen generates Go interfaces
and models from it. You then implement the interface.

```yaml
# openapi.yaml
paths:
  /tasks/{id}:
    get:
      operationId: getTask
      parameters:
        - name: id
          in: path
          required: true
          schema: { type: string }
```

```bash
oapi-codegen -generate types,server -package api openapi.yaml > api/api.gen.go
```

```go
// You implement the generated interface:
func (s *Server) GetTask(w http.ResponseWriter, r *http.Request, id string) { ... }
```

**vs. go-api-docs/openapi:**

- Spec-first is the right approach when you have an existing YAML spec, multiple
  client implementations, or strict API-contract governance between teams
- go-api-docs/openapi is code-first — ideal for new projects where the code IS the source of truth
- With oapi-codegen the spec can diverge from the implementation between regenerations;
  with go-api-docs/openapi it cannot (spec derives from types at startup)
- oapi-codegen offers strict type-safe server stubs; go-api-docs/openapi gives you more
  flexibility in handler structure

---

## Manual (`map[string]any` or hand-crafted JSON)

Building the spec by hand (as in `example/swagger/full/buildSpec()`) gives complete
control but has obvious problems:

- Any rename, type change, or new field in a handler that isn't also done in `buildSpec()`
  is silently wrong — the spec lies
- Significant up-front effort for any non-trivial API
- No discoverability (you must remember to add every new endpoint)
  go-api-docs/openapi eliminates this entirely: if you add a route without registering it,
  it simply doesn't exist in the spec (same as a 404), not silently wrong.

---

## Ecosystem compatibility — does it work with every Go HTTP framework?

**Short answer: yes.**

`openapi.Router` implements `http.Handler`. Any framework that can wrap a
standard `http.Handler` can host it with zero changes to your handler code.
The router uses Go 1.22 stdlib mux internally so path parameters are always
set correctly via `r.PathValue()` — no extra configuration needed for the
mount-at-root pattern.

### Mount patterns

| Framework             | Wiring snippet                                         | Path params                     |
|-----------------------|--------------------------------------------------------|---------------------------------|
| **net/http** (stdlib) | `http.Handle("/", api)`                                | `r.PathValue` ✓ native          |
| **chi**               | `r.Mount("/", api)`                                    | `r.PathValue` ✓ internal mux    |
| **gorilla/mux**       | `r.PathPrefix("/").Handler(api)`                       | `r.PathValue` ✓ internal mux    |
| **gin**               | `g.Any("/*path", gin.WrapH(api))`                      | `r.PathValue` ✓ internal mux    |
| **echo**              | `e.Any("/*", echo.WrapHandler(api))`                   | `r.PathValue` ✓ internal mux    |
| **fiber**             | `app.All("/*", adaptor.HTTPHandler(api))`              | `r.PathValue` ✓ internal mux    |
| **httprouter**        | `adapt(api)` per route + `WithPathValueFn`             | `httprouter.Params` via context |
| **Beego**             | `beego.Handler("/api/*", api)`                         | `r.PathValue` ✓ internal mux    |
| **Iris**              | `app.Any("/{p:path}", iris.FromStd(api))`              | `r.PathValue` ✓ internal mux    |
| **Buffalo**           | `app.ANY("/api/{p:path}", buffalo.WrapHandlerFunc(…))` | `r.PathValue` ✓ internal mux    |

**Most frameworks (chi, gorilla, gin, echo, fiber, Beego, Iris, Buffalo)** fall
into the easy category: they accept an `http.Handler` via a catch-all or
`PathPrefix` mount, and the `openapi.Router` handles all internal dispatch with
its own stdlib mux — path params included.

**httprouter** is the notable exception: it uses a non-standard handler
signature `func(w, r, Params)` and has no generic `http.Handler` mount.
The solution is a tiny 4-line `adapt()` helper that puts `httprouter.Params`
into the request context and `WithPathValueFn` to read them back:

```go
func adapt(h http.Handler) httprouter.Handle {
return func (w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
h.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), paramsKey{}, ps)))
}
}

api := openapi.New(info, openapi.WithPathValueFn(func (r *http.Request, name string) string {
ps, _ := r.Context().Value(paramsKey{}).(httprouter.Params)
return ps.ByName(name)
}))

router.GET("/tasks/:id", adapt(api))
```

See `examples/openapi/httprouter/` for the full runnable example.

### How path params work without any config (the mount-at-root pattern)

```
[outer framework middleware]
        ↓
openapi.Router.ServeHTTP          ← receives the request
        ↓
internal stdlib mux               ← matches GET /tasks/{id}
        ↓                            calls r.SetPathValue("id", "123")
your typed handler                ← decodeInput calls r.PathValue("id") ✓
```

### Runnable examples (self-contained Go modules)

#### `go-api-docs/openapi` — framework integrations

```
examples/openapi/chi/        — chi + chi middleware
examples/openapi/gorilla/    — gorilla/mux
examples/openapi/gin/        — gin (gin.WrapH)
examples/openapi/echo/       — echo (echo.WrapHandler)
examples/openapi/fiber/      — fiber (gofiber/adaptor)
examples/openapi/httprouter/ — httprouter (WithPathValueFn + context)
```

```bash
cd examples/openapi/chi        && go run .  # :9091
cd examples/openapi/gorilla    && go run .  # :9092
cd examples/openapi/gin        && go run .  # :9093
cd examples/openapi/echo       && go run .  # :9094
cd examples/openapi/fiber      && go run .  # :9095
cd examples/openapi/httprouter && go run .  # :9096
```

#### `go-api-docs/scalar` — Scalar UI serving

```
example/scalar/basic/       — minimal Scalar UI from a local spec
example/scalar/cdn/         — Scalar UI loaded from CDN (no embedded assets)
example/scalar/full/        — full options: theme, custom CSS, layout
example/scalar/multi-spec/  — multiple specs on one page
example/scalar/openpotato/  — live public spec (OpenPotato)
example/scalar/swagger2/    — Swagger 2.0 spec served in Scalar
```

```bash
cd example/scalar/basic      && go run .
cd example/scalar/cdn        && go run .
cd example/scalar/full       && go run .
cd example/scalar/multi-spec && go run .
cd example/scalar/openpotato && go run .
cd example/scalar/swagger2   && go run .
```

#### `go-api-docs/swagger` — Swagger UI serving

```
example/swagger/basic/      — minimal Swagger UI from a local spec
example/swagger/full/       — full options: custom title, layout
example/swagger/openapi/    — spec generated by go-api-docs/openapi served in Swagger UI
example/swagger/swag/       — spec generated by swaggo/swag served in Swagger UI
```

```bash
cd example/swagger/basic   && go run .
cd example/swagger/full    && go run .
cd example/swagger/openapi && go run .
cd example/swagger/swag    && go run .
```

---

## When to use what

| Situation                                                   | Recommended                   |
|-------------------------------------------------------------|-------------------------------|
| New Go project, want spec + UI from code                    | **go-api-docs/openapi**       |
| Existing project with hand-written spec                     | **oapi-codegen** (spec-first) |
| Large codebase, already using swag, need annotation tooling | **swaggo/swag**               |
| Full-featured framework, OK with more deps                  | **huma v2**                   |
| Just serve an existing spec file in Scalar                  | **go-api-docs/scalar**        |
| Just serve an existing spec file in Swagger UI              | **go-api-docs/swagger**       |
