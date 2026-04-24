// Package goscalar provides a self-contained Scalar API documentation handler
// for Go HTTP servers.
//
// The Scalar JS bundle is vendored via go:embed (see [assets]) so the handler
// works fully offline / in air-gapped environments.
//
// # Quick start
//
//	h, err := goscalar.New(
//	    goscalar.WithSources(
//	        goscalar.Source{URL: "/openapi.json", Title: "Current", Default: true},
//	    ),
//	    goscalar.WithTheme(goscalar.ThemeNone),
//	    goscalar.WithDarkMode(),
//	    goscalar.WithBranding(goscalar.Branding{
//	        LogoURL:    "/logo.svg",
//	        Title:      "Acme Corp",
//	        Subtitle:   "Platform API",
//	        FaviconURL: "/favicon.svg",
//	    }),
//	    goscalar.WithEnvBadge("preprod"),
//	    goscalar.WithDisableAgent(),
//	    goscalar.WithDisableMCP(),
//	    goscalar.WithHideClientButton(),
//	)
//	if err != nil { ... }
//
//	// Mount on your router (chi, stdlib mux, …)
//	mux.Handle("/docs/", h)
//
// The handler responds to:
//
//	GET /          → HTML page
//	GET /scalar.js → vendored Scalar JS bundle
package goscalar

import (
	"bytes"
	"encoding/json"
	"html/template"
	"net/http"

	"github.com/nopereta/goscalar/assets"
)

// Handler is an [http.Handler] that serves the Scalar UI and its assets.
// Create one with [New].
type Handler struct {
	cfg      *config
	pageHTML []byte
}

// New creates a Handler applying the provided options.
func New(opts ...Option) (*Handler, error) {
	cfg := defaultConfig()
	for _, o := range opts {
		o(cfg)
	}

	pageHTML, err := renderPage(cfg)
	if err != nil {
		return nil, err
	}

	return &Handler{cfg: cfg, pageHTML: pageHTML}, nil
}

// ServeHTTP dispatches requests to the page or the JS asset.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case h.cfg.jsPath:
		w.Header().Set("Content-Type", "application/javascript")
		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		_, _ = w.Write(assets.ScalarJS)
	default:
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write(h.pageHTML)
	}
}

// ---------------------------------------------------------------------------
// internal config
// ---------------------------------------------------------------------------

type config struct {
	// Scalar JS API configuration object (marshalled to JSON).
	scalarConfig map[string]any
	// sources is written into scalarConfig["sources"] when not nil.
	sources []Source

	// Page-level options.
	pageTitle string
	envBadge  string
	branding  *Branding
	jsPath    string
	tmpl      *template.Template
}

func defaultConfig() *config {
	return &config{
		scalarConfig: map[string]any{
			"darkMode":    true,
			"theme":       ThemeDefault,
			"showToolbar": ShowToolbarNever,
		},
		pageTitle: "API Reference",
		jsPath:    "/scalar.js",
	}
}

// ---------------------------------------------------------------------------
// Rendering
// ---------------------------------------------------------------------------

// PageData is passed to the HTML template.
type PageData struct {
	// PageTitle is the HTML <title> value.
	PageTitle string
	// FaviconURL is the href for <link rel="icon"> (empty → omit).
	FaviconURL string
	// Branding holds optional header branding (nil → no header).
	Branding *Branding
	// EnvBadge is an optional environment label displayed in the header.
	EnvBadge string
	// JSPath is the URL from which the Scalar JS bundle is loaded.
	JSPath string
	// ScalarConfig is the JSON-encoded Scalar configuration object passed to
	// Scalar.createApiReference().
	ScalarConfig template.JS
}

func renderPage(cfg *config) ([]byte, error) {
	// Build the Scalar JS config map.
	sc := make(map[string]any, len(cfg.scalarConfig)+1)
	for k, v := range cfg.scalarConfig {
		sc[k] = v
	}
	if len(cfg.sources) > 0 {
		sc["sources"] = cfg.sources
	}

	configJSON, err := json.Marshal(sc)
	if err != nil {
		return nil, err
	}

	// Resolve template.
	tmpl := cfg.tmpl
	if tmpl == nil {
		tmpl = defaultTemplate()
	}

	data := PageData{
		PageTitle:    cfg.pageTitle,
		Branding:     cfg.branding,
		EnvBadge:     cfg.envBadge,
		JSPath:       cfg.jsPath,
		ScalarConfig: template.JS(configJSON), // #nosec G203 – JSON-safe
	}
	if cfg.branding != nil {
		data.FaviconURL = cfg.branding.FaviconURL
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

