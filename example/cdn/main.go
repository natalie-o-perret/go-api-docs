// Example: cdn — use the Scalar CDN bundle instead of the vendored one.
//
// Run from this directory:
//
//	go run .
//	open http://localhost:8083
package main

import (
	_ "embed"
	"log"
	"net/http"

	scalar "github.com/nopereta/go-scalar"
)

//go:embed openapi.json
var specJSON []byte

const scalarCDN = "https://cdn.jsdelivr.net/npm/@scalar/api-reference@latest/dist/browser/standalone.js"

func main() {
	h, err := scalar.New(
		scalar.WithSpecURL("/openapi.json"),
		scalar.WithJSPath(scalarCDN),
		scalar.WithTheme(scalar.ThemePurple),
		scalar.WithBranding(scalar.Branding{
			Title:    "Acme Corp",
			Subtitle: "Tasks API — CDN build",
		}),
		scalar.With(scalar.DisableAgent, scalar.DisableMCP, scalar.DarkMode),
		scalar.WithPageTitle("Tasks API (CDN)"),
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

	log.Println("listening on http://localhost:8083")
	log.Fatal(http.ListenAndServe(":8083", mux))
}

