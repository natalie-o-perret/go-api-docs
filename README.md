# go-api-docs

[![CI](https://github.com/natalie-o-perret/go-api-docs/actions/workflows/ci.yml/badge.svg)](https://github.com/natalie-o-perret/go-api-docs/actions/workflows/ci.yml)
[![Lint](https://github.com/natalie-o-perret/go-api-docs/actions/workflows/lint.yml/badge.svg)](https://github.com/natalie-o-perret/go-api-docs/actions/workflows/lint.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/natalie-o-perret/go-api-docs.svg)](https://pkg.go.dev/github.com/natalie-o-perret/go-api-docs)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

Typed HTTP routing and OpenAPI 3.1 generation from Go structs.

Route registrations define the HTTP operation. Go types and struct tags define
its parameters and JSON schemas. The runtime packages use only the standard
library.

| Package      | Purpose                                                  |
|--------------|----------------------------------------------------------|
| `openapi`    | Typed `net/http` router and runtime OpenAPI generation   |
| `ui/scalar`  | Scalar handler with an embedded or CDN-hosted bundle     |
| `ui/swagger` | Swagger UI handler with embedded or CDN-hosted assets    |
| `goapi-gen`  | Optional source analyser for writing a static spec       |

## Install

```bash
go get github.com/natalie-o-perret/go-api-docs
```

## Quick Start

```go
package main

import (
    "net/http"

    "github.com/natalie-o-perret/go-api-docs/openapi"
)

type CreateTaskInput struct {
    Title string `json:"title" doc:"Task title" minLength:"1"`
}

type Task struct {
    ID    string `json:"id" readOnly:"true"`
    Title string `json:"title"`
}

func main() {
    api := openapi.New(openapi.Info{Title: "Tasks API", Version: "1.0.0"})
    openapi.POST[CreateTaskInput, Task](api, "/tasks", createTask,
        openapi.Summary("Create a task"),
        openapi.Tags("tasks"),
    )
    http.ListenAndServe(":8080", api)
}

func createTask(_ *http.Request, in *CreateTaskInput) (*Task, error) {
    return &Task{ID: "task_42", Title: in.Title}, nil
}
```

The router serves the generated document at `GET /openapi.json`.

## Inputs

A single input struct can contain path, query, header, and JSON body fields:

```go
type UpdateTaskInput struct {
    ID      string  `path:"id"`
    TraceID string  `header:"X-Trace-ID" required:"true"`
    Notify  bool    `query:"notify"`
    Title   *string `json:"title,omitempty" maxLength:"200"`
}

openapi.PATCH[UpdateTaskInput, Task](api, "/tasks/{id}", updateTask)
```

| Tag                  | Meaning                                      |
|----------------------|----------------------------------------------|
| `path:"name"`        | Required URL path parameter                  |
| `query:"name"`       | Query parameter                              |
| `header:"name"`      | Header parameter                             |
| `required:"true"`    | Require a query or header parameter          |
| `doc:"..."`          | Schema or parameter description              |
| `example:"..."`      | Example value in the generated schema        |
| `enum:"a,b"`         | Allowed values in the generated schema       |
| `format:"uuid"`      | JSON Schema format                           |
| `pattern:"..."`      | JSON Schema string pattern                   |
| `minLength:"1"`      | JSON Schema minimum string length            |
| `maxLength:"255"`    | JSON Schema maximum string length            |
| `min:"0"`, `max:"9"` | JSON Schema numeric bounds                 |
| `minItems:"1"`       | JSON Schema minimum array length             |
| `maxItems:"50"`      | JSON Schema maximum array length             |
| `readOnly:"true"`    | Mark a schema property as read-only          |
| `writeOnly:"true"`   | Mark a schema property as write-only         |

Path parameters and parameters tagged `required:"true"` are enforced during
decoding. Constraint tags describe the OpenAPI schema. Use `Validator` when the
server must enforce those constraints or business rules:

```go
func (in *CreateTaskInput) Validate() error {
    if in.Title == "" {
        return openapi.ErrBadRequest("title is required")
    }
    return nil
}
```

Request bodies use the standard `encoding/json` decoder. Body-bearing input is
required for `POST`, `PUT`, and `PATCH`. `GET` and `DELETE` do not decode JSON
bodies.

## Routes

```go
openapi.GET[Output](api, path, handler, options...)
openapi.GETWithInput[Input, Output](api, path, handler, options...)
openapi.POST[Input, Output](api, path, handler, options...)
openapi.PUT[Input, Output](api, path, handler, options...)
openapi.PATCH[Input, Output](api, path, handler, options...)
openapi.DELETE[Input](api, path, handler, options...)
openapi.Handle[Input, Output](api, method, path, handler, options...)
```

Available route options are `Summary`, `Description`, `Tags`, `OperationID`,
`Security`, `Deprecated`, and `Responses`.

`POST` returns `201`, `DELETE` and `struct{}` outputs return `204`, and other
successful handlers return `200`. A nil non-empty output is treated as an
internal server error.

## Errors

```go
func getTask(_ *http.Request, in *TaskIDParam) (*Task, error) {
    task, ok := findTask(in.ID)
    if !ok {
        return nil, openapi.ErrNotFound("task not found")
    }
    return task, nil
}
```

Helpers include `ErrNotFound`, `ErrBadRequest`, `ErrUnauthorized`, and `Err`.
Unexpected errors return a generic `500` response. Log sensitive details in the
application before returning the error.

## Router Options

```go
api := openapi.New(
    openapi.Info{Title: "Tasks API", Version: "1.0.0"},
    openapi.WithServer("https://api.example.com", "Production"),
    openapi.WithSecurityScheme("BearerAuth", openapi.BearerAuth),
    openapi.WithTag("tasks", "Task operations"),
)
```

Security helpers include `BearerAuth`, `BasicAuth`, `APIKeyHeader`, and
`APIKeyQuery`.

`Router` implements `http.Handler`, so an outer framework can mount it as a
standard handler. `WithPathValueFn` supports routers that store path parameters
outside `http.Request.PathValue`.

## Custom Runtime Schemas

Implement `SchemaProvider` when reflection cannot describe a runtime schema:

```go
func (Task) OpenAPISchema() openapi.Schema {
    return openapi.Schema{
        Type: "object",
        Properties: map[string]openapi.Schema{
            "id":    {Type: "string", ReadOnly: true},
            "title": {Type: "string"},
        },
        Required: []string{"id", "title"},
    }
}
```

`SchemaProvider` executes application code and is therefore supported only by
runtime generation. `goapi-gen` reports an error when it encounters one.

## Static Generation

Install the optional source analyser:

```bash
go install github.com/natalie-o-perret/go-api-docs/cmd/goapi-gen@latest
goapi-gen -out openapi.json ./cmd/api
```

It recognises direct calls to `openapi.New`, route registration functions, and
the options listed above. Paths, methods, metadata, and option values must be
compile-time constants. One generated document may contain one `openapi.New`
call.

The command fails on package-loading errors, multiple routers, and
`SchemaProvider` types rather than writing a partial or misleading document.
Keep a generated snapshot checked in when consumers need a static contract:

```bash
goapi-gen -out openapi.json ./cmd/api
git diff --exit-code -- openapi.json
```

## Scalar

```go
docs, err := scalar.New(
    scalar.WithSpecURL("/openapi.json"),
    scalar.WithTheme(scalar.ThemeDefault),
    scalar.With(scalar.DisableAgent, scalar.DisableMCP),
)
if err != nil {
    log.Fatal(err)
}
http.Handle("/", docs)
```

The package embeds `@scalar/api-reference`. Use `WithJSPath` to load a CDN or
self-hosted bundle instead.

```bash
make vendor-js SCALAR_VERSION=1.53.0
```

## Swagger UI

```go
docs, err := swagger.New(
    swagger.WithSpecURL("/openapi.json"),
    swagger.WithPersistAuthorization(),
    swagger.WithFilter(""),
)
if err != nil {
    log.Fatal(err)
}
http.Handle("/", docs)
```

```bash
make vendor-swagger-ui SWAGGER_VERSION=5.19.0
```

Both UI handlers use root-relative asset paths by default. Override their asset
paths when mounting the handler below a URL prefix.

## Examples

Runnable examples live under:

- `examples/openapi` for `net/http`, chi, Echo, Fiber, Gin, Gorilla, and httprouter
- `examples/ui/scalar` for Scalar configurations
- `examples/ui/swagger` for Swagger UI configurations

See [COMPARISON.md](COMPARISON.md) for a short guide to alternative approaches.

## Licence

The Go code is MIT licensed. Vendored browser assets retain their upstream
licences. See [THIRD_PARTY_NOTICES.md](THIRD_PARTY_NOTICES.md).
