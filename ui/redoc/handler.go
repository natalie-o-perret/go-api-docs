// Package redoc provides a Redoc API docs handler for Go.
// The JS bundle is vendored via go:embed but you can swap it with WithJSPath.
//
// Quick start:
//
//	h, err := redoc.New(redoc.WithSpecURL("/openapi.json"))
//	if err != nil {
//	    log.Fatal(err)
//	}
//	http.Handle("/docs/", h)
package redoc

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"

	"github.com/nopereta/go-api-docs/ui/redoc/assets"
)

// Handler serves the Redoc UI and its JS asset.
// Create one with [New].
type Handler struct {
	cfg      *config
	pageHTML []byte
}

// New builds a Handler from the given options.
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

// ServeHTTP implements http.Handler.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case h.cfg.jsPath:
		w.Header().Set("Content-Type", "application/javascript")
		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		_, _ = w.Write(assets.RedocJS)
	default:
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Cache-Control", "no-cache")
		_, _ = w.Write(h.pageHTML)
	}
}

type config struct {
	specURL      string
	pageTitle    string
	envBadge     string
	branding     *Branding
	jsPath       string
	darkMode     bool
	redocOptions map[string]any
	tmpl         *template.Template
}

func defaultConfig() *config {
	return &config{
		pageTitle:    "API Reference",
		jsPath:       "/redoc.js",
		darkMode:     true,
		redocOptions: map[string]any{},
	}
}

// PageData is passed to the HTML template.
type PageData struct {
	PageTitle   string
	FaviconURL  string
	FaviconType string
	FaviconLink template.HTML
	DarkMode    bool
	Branding    *Branding
	EnvBadge    string
	JSPath      string
	SpecURL     template.JS // JSON-encoded spec URL string
	RedocConfig template.JS // JSON-encoded options object
}

func renderPage(cfg *config) ([]byte, error) {
	opts := make(map[string]any, len(cfg.redocOptions))
	for k, v := range cfg.redocOptions {
		opts[k] = v
	}

	specURLJSON, err := json.Marshal(cfg.specURL)
	if err != nil {
		return nil, err
	}

	configJSON, err := json.Marshal(opts)
	if err != nil {
		return nil, err
	}

	tmpl := cfg.tmpl
	if tmpl == nil {
		tmpl = defaultTemplate()
	}

	data := PageData{
		PageTitle:   cfg.pageTitle,
		DarkMode:    cfg.darkMode,
		Branding:    cfg.branding,
		EnvBadge:    cfg.envBadge,
		JSPath:      cfg.jsPath,
		SpecURL:     template.JS(specURLJSON),  // #nosec G203 -- JSON-safe
		RedocConfig: template.JS(configJSON),   // #nosec G203 -- JSON-safe
	}
	if cfg.branding != nil {
		data.FaviconURL = cfg.branding.FaviconURL
		data.FaviconType = cfg.branding.FaviconType
	}
	if data.FaviconType == "" {
		data.FaviconType = "image/svg+xml"
	}
	if data.FaviconURL != "" {
		data.FaviconLink = template.HTML(fmt.Sprintf(
			`<link rel="icon" type="%s" href="%s"/>`,
			template.HTMLEscapeString(data.FaviconType),
			template.HTMLEscapeString(data.FaviconURL),
		))
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

