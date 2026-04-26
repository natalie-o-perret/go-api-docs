# go-scalar

A Scalar API docs handler for Go. Ships with a vendored JS bundle by default, or point it at any URL you like.

## Why not the others?

| Feature                     | [scalar-go](https://github.com/bdpiprava/scalar-go) | [go-scalar-api-reference](https://github.com/MarceloPetrucio/go-scalar-api-reference) | [gofiber-scalar](https://github.com/yokeTH/gofiber-scalar) | [gin-openapi](https://github.com/PeterTakahashi/gin-openapi) | **go-scalar**                               |
|-----------------------------|-----------------------------------------------------|---------------------------------------------------------------------------------------|------------------------------------------------------------|--------------------------------------------------------------|---------------------------------------------|
| Framework                   | any                                                 | any                                                                                   | Fiber only                                                 | Gin only                                                     | any                                         |
| Returns `http.Handler`      | ✅                                                   | ❌ raw HTML string                                                                     | ❌ fiber.Handler                                            | ❌ gin.HandlerFunc                                            | ✅                                           |
| JS source                   | CDN only                                            | CDN only                                                                              | CDN only                                                   | CDN only                                                     | vendored by default, swap with `WithJSPath` |
| Spec loading                | Go roundtrip                                        | Go roundtrip                                                                          | swaggo parse                                               | swaggo parse                                                 | URL only (browser fetches)                  |
| Custom branded header       | ❌                                                   | ❌                                                                                     | ❌                                                          | ❌                                                            | ✅                                           |
| Typed Scalar config options | partial                                             | partial                                                                               | partial                                                    | partial                                                      | ✅ full + `WithOption` escape hatch          |
| Multi-spec sources          | ✅                                                   | ❌                                                                                     | ❌                                                          | ❌                                                            | ✅                                           |

## Install

```bash
go get github.com/nopereta/go-scalar
```

> **Note :** `assets/scalar.min.js` is vendored in this repo and used by default.
> Swap it for any URL (CDN or self-hosted) with `scalar.WithJSPath("https://...")`.
> To update the vendored bundle: `make vendor-js VERSION=x.y.z`

## Quick start

```go
import scalar "github.com/nopereta/go-scalar"

h, err := scalar.New(
    scalar.WithSources(
        scalar.Source{URL: "/openapi.json", Title: "Current", Default: true},
        scalar.Source{URL: "/openapi.previous.json", Title: "Previous"},
    ),
    scalar.WithTheme(scalar.ThemeNone),
    scalar.WithBranding(scalar.Branding{
        LogoURL:    "/logo.svg",
        LogoAlt:    "Acme",
        Title:      "Acme Corp",
        Subtitle:   "Platform API",
        FaviconURL: "/favicon.svg",
        // FaviconType defaults to "image/svg+xml"; set explicitly for .ico/.png
    }),
    scalar.WithEnvBadge("preprod"),
    scalar.WithBaseServerURL("https://api.acme.com"),
    // Boolean flags — combine as many as you like in one call
    scalar.With(scalar.DisableAgent, scalar.DisableMCP, scalar.HideClientButton, scalar.DarkMode),
    scalar.WithShowDeveloperTools(scalar.ShowToolbarNever),
)
if err != nil { log.Fatal(err) }

// The handler serves:
//   GET /           → HTML page  (Cache-Control: no-cache)
//   GET /scalar.js  → vendored JS bundle  (Cache-Control: immutable)
mux.Handle("/", h)
```

## Boolean feature flags

`With` accepts one or more `Flag` constants and can be combined with `|`:

```go
scalar.With(scalar.DisableAgent, scalar.DisableMCP, scalar.HideClientButton)
// equivalent:
scalar.With(scalar.DisableAgent | scalar.DisableMCP | scalar.HideClientButton)
```

Available flags: `DisableAgent`, `DisableMCP`, `HideClientButton`, `HideModels`,
`HideDownloadButton`, `HideSearch`, `ShowOperationID`, `DarkMode`, `LightMode`.

## Light mode

```go
h, _ := scalar.New(scalar.WithLightMode())
// or via flags:
h, _ := scalar.New(scalar.With(scalar.LightMode))
```

The page chrome (branded header + body background) adapts automatically.

## Server URL override

Point Scalar's "Try it" panel at a specific environment without touching the spec:

```go
scalar.WithBaseServerURL("https://staging.api.acme.com")
```

## Custom HTML template

Replace the default template entirely:

```go
tmpl := template.Must(template.ParseFiles("my_page.html"))
h, _ := scalar.New(
    scalar.WithTemplate(tmpl),
    // ...
)
```

The template receives a [`scalar.PageData`](gos.go) value.

## Escape hatch

Pass any Scalar config key not covered by a typed option:

```go
scalar.WithOption("tagsSorter", "alpha")
scalar.WithOption("operationsSorter", "method")
```

## Examples

| Example | Spec format | What it shows |
|---|---|---|
| [`example/basic`](example/basic/) | OpenAPI 3.1 | Minimal working server with a real spec on disk |
| [`example/swagger`](example/swagger/) | Swagger 2.0 | Same setup, legacy spec format |
| [`example/multi-spec`](example/multi-spec/) | OpenAPI 3.1 + Swagger 2.0 | Version picker dropdown, two specs in one UI |
| [`example/cdn`](example/cdn/) | OpenAPI 3.1 | No vendored JS — loads Scalar from jsDelivr CDN |

Each example is a self-contained runnable server (`go run .` from its directory).

> **Spec format support:** Scalar renders **OpenAPI 3.x** and **Swagger 2.0** specs (JSON or YAML).
> AsyncAPI, GraphQL SDL, and RAML are not supported by the Scalar JS library.

## Vendoring the JS bundle

```bash
make vendor-js                     # uses default version (1.52.6)
make vendor-js VERSION=1.53.0      # pin a specific version
```
