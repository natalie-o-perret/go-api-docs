// Package scalar is a Scalar API docs handler for Go.
// The JS bundle is vendored via go:embed but you can swap it with WithJSPath.
package scalar

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"

	"github.com/nopereta/go-scalar/assets"
)

// Handler serves the Scalar UI and its JS asset.
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

// ServeHTTP serves the HTML page or the JS bundle depending on the path.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case h.cfg.jsPath:
		w.Header().Set("Content-Type", "application/javascript")
		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		_, _ = w.Write(assets.ScalarJS)
	default:
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Cache-Control", "no-cache")
		_, _ = w.Write(h.pageHTML)
	}
}

type config struct {
	scalarConfig map[string]any
	sources      []Source
	pageTitle    string
	envBadge     string
	branding     *Branding
	jsPath       string
	tmpl         *template.Template
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

// PageData is passed to the HTML template.
type PageData struct {
	PageTitle    string
	FaviconURL   string
	FaviconType  string        // MIME type for the favicon, e.g. "image/svg+xml"
	FaviconLink  template.HTML // pre-rendered <link> tag; used by the default template
	DarkMode     bool          // reflects the darkMode scalar config value
	Branding     *Branding
	EnvBadge     string
	JSPath       string
	ScalarConfig template.JS
}

func renderPage(cfg *config) ([]byte, error) {
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

	tmpl := cfg.tmpl
	if tmpl == nil {
		tmpl = defaultTemplate()
	}

	darkMode, _ := cfg.scalarConfig["darkMode"].(bool)

	data := PageData{
		PageTitle:    cfg.pageTitle,
		DarkMode:     darkMode,
		Branding:     cfg.branding,
		EnvBadge:     cfg.envBadge,
		JSPath:       cfg.jsPath,
		ScalarConfig: template.JS(configJSON), // #nosec G203 -- JSON-safe
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
