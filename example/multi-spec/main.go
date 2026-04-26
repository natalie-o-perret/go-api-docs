// Example: multi-spec — browse v1 (Swagger 2.0) and v2 (OpenAPI 3.1) from a single UI.
//
// Run from this directory:
//
//	go run .
//	open http://localhost:8082
package main

import (
	_ "embed"
	"log"
	"net/http"

	scalar "github.com/nopereta/go-scalar"
)

//go:embed openapi.json
var specV2 []byte

//go:embed swagger.json
var specV1 []byte

func main() {
	h, err := scalar.New(
		scalar.WithSources(
			scalar.Source{URL: "/v2/openapi.json", Title: "v2 (OpenAPI 3.1)", Default: true},
			scalar.Source{URL: "/v1/swagger.json", Title: "v1 (Swagger 2.0)"},
		),
		scalar.WithTheme(scalar.ThemeMoon),
		scalar.WithBranding(scalar.Branding{
			Title:    "Acme Corp",
			Subtitle: "Tasks API",
		}),
		scalar.With(scalar.DisableAgent, scalar.DisableMCP, scalar.DarkMode),
		scalar.WithBaseServerURL("https://api.acme.example.com"),
		scalar.WithPageTitle("Tasks API — all versions"),
	)
	if err != nil {
		log.Fatal(err)
	}

	mux := http.NewServeMux()
	mux.Handle("/", h)

	mux.HandleFunc("GET /v2/openapi.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(specV2)
	})
	mux.HandleFunc("GET /v1/swagger.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(specV1)
	})

	log.Println("listening on http://localhost:8082")
	log.Fatal(http.ListenAndServe(":8082", mux))
}

