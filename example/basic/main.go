// Example shows a minimal goscalar integration with a stdlib HTTP server.
package main

import (
	"log"
	"net/http"

	"github.com/nopereta/goscalar"
)

func main() {
	h, err := goscalar.New(
		// Spec loaded by URL at render time – no Go-side parsing.
		goscalar.WithSources(
			goscalar.Source{URL: "/openapi.json", Title: "Current", Default: true},
			goscalar.Source{URL: "/openapi.previous.json", Title: "Previous"},
		),

		// Visual
		goscalar.WithTheme(goscalar.ThemeNone),
		goscalar.WithDarkMode(),

		// Branded header
		goscalar.WithBranding(goscalar.Branding{
			LogoURL:    "/logo.svg",
			LogoAlt:    "Acme",
			Title:      "Acme Corp",
			Subtitle:   "Platform API",
			FaviconURL: "/favicon.svg",
		}),
		goscalar.WithEnvBadge("dev"),

		// Disable noisy Scalar features
		goscalar.WithDisableAgent(),
		goscalar.WithDisableMCP(),
		goscalar.WithHideClientButton(),
		goscalar.WithShowDeveloperTools("never"),
		goscalar.WithShowToolbar(goscalar.ShowToolbarNever),

		goscalar.WithPageTitle("Acme – Platform API"),
	)
	if err != nil {
		log.Fatal(err)
	}

	mux := http.NewServeMux()

	// The handler serves GET / (HTML) and GET /scalar.js (vendored bundle).
	mux.Handle("/", h)

	// Serve your real OpenAPI specs.
	mux.HandleFunc("/openapi.json", func(w http.ResponseWriter, _ *http.Request) {
		http.ServeFile(w, nil, "openapi.json")
	})

	log.Println("listening on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}

