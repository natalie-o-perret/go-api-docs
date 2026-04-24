# goscalar

A self-contained [Scalar](https://scalar.com) API documentation handler for Go — no CDN dependency, no spec parsing.

## Why not [scalar-go](https://github.com/bdpiprava/scalar-go)?

| Feature | scalar-go | **goscalar** |
|---|---|---|
| Vendored JS (offline) | ❌ CDN only | ✅ `go:embed` |
| Custom branded header | ❌ fixed template | ✅ full HTML slot |
| Favicon / logo serving | ❌ | ✅ |
| `agent`, `mcp`, `hideClientButton` | ❌ not mapped | ✅ typed options |
| Spec loading | ⚠️ Go roundtrip (field loss risk) | ✅ URL only (Scalar fetches) |
| Multi-sources (hot-reload) | ⚠️ inlined at boot | ✅ URL-based, live |
| Escape hatch for any config key | ⚠️ awkward | ✅ `WithOption(key, val)` |

## Install

```bash
go get github.com/nopereta/goscalar
```

> **Note :** the `assets/scalar.min.js` bundle is **vendored** in this repo.
> To update it: `make vendor-js VERSION=x.y.z`

## Quick start

```go
h, err := goscalar.New(
    goscalar.WithSources(
        goscalar.Source{URL: "/openapi.json", Title: "Current", Default: true},
        goscalar.Source{URL: "/openapi.previous.json", Title: "Previous"},
    ),
    goscalar.WithTheme(goscalar.ThemeNone),
    goscalar.WithDarkMode(),
    goscalar.WithBranding(goscalar.Branding{
        LogoURL:    "/logo.svg",
        LogoAlt:    "Acme",
        Title:      "Acme Corp",
        Subtitle:   "Platform API",
        FaviconURL: "/favicon.svg",
    }),
    goscalar.WithEnvBadge("preprod"),
    goscalar.WithDisableAgent(),
    goscalar.WithDisableMCP(),
    goscalar.WithHideClientButton(),
    goscalar.WithShowDeveloperTools("never"),
)
if err != nil { log.Fatal(err) }

// The handler serves:
//   GET /           → HTML page
//   GET /scalar.js  → vendored JS bundle
mux.Handle("/", h)
```

## Custom HTML template

Replace the default template entirely:

```go
tmpl := template.Must(template.ParseFiles("my_page.html"))
h, _ := goscalar.New(
    goscalar.WithTemplate(tmpl),
    // ...
)
```

The template receives a [`goscalar.PageData`](goscalar.go) value.

## Escape hatch

Pass any Scalar config key not covered by a typed option:

```go
goscalar.WithOption("tagsSorter", "alpha")
goscalar.WithOption("operationsSorter", "method")
```

## Vendoring the JS bundle

```bash
make vendor-js                     # uses default version (1.52.6)
make vendor-js VERSION=1.53.0      # pin a specific version
```

