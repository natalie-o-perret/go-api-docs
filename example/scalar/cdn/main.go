// Example: cdn — use the Scalar CDN bundle instead of the vendored one.
//
// The OpenAPI spec is the public Swagger Petstore hosted at petstore3.swagger.io,
// so no spec needs to be defined in this file.
//
// Run from this directory:
//
//	go run .
//	open http://localhost:8083
package main

import (
	"log"
	"net/http"

	scalar "github.com/nopereta/go-api-docs/ui/scalar"
)

const scalarCDN = "https://cdn.jsdelivr.net/npm/@scalar/api-reference@latest/dist/browser/standalone.js"

func main() {
	h, err := scalar.New(
		scalar.WithSpecURL("https://petstore3.swagger.io/api/v3/openapi.json"),
		scalar.WithJSPath(scalarCDN),
		scalar.WithTheme(scalar.ThemePurple),
		scalar.WithBranding(scalar.Branding{
			Title:    "Acme Corp",
			Subtitle: "Tasks API — CDN build",
		}),
		scalar.With(scalar.DisableAgent, scalar.DisableMCP, scalar.DarkMode),
		scalar.WithPageTitle("Tasks API (CDN)"),
		// Custom theme: override individual CSS variables on top of ThemePurple.
		scalar.WithCustomTheme(
			scalar.NewCustomTheme().
				Accent("#7c3aed").
				AccentBackground("rgba(124,58,237,.12)").
				Background("#0d0d14").
				BackgroundSecondary("#13131f").
				Border("rgba(255,255,255,.07)").
				Font("'Inter', system-ui, sans-serif").
				Radius("8px"),
		),
	)
	if err != nil {
		log.Fatal(err)
	}

	mux := http.NewServeMux()
	mux.Handle("/", h)

	log.Println("listening on http://localhost:8083")
	log.Fatal(http.ListenAndServe(":8083", mux))
}
