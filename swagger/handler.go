// Package swagger provides a Swagger UI HTTP handler for Go.
// It renders the official Swagger UI (served via CDN by default) and allows
// customisation through a functional-options API.
//
// Quick start:
//
//	h, err := swagger.New(swagger.WithSpecURL("/openapi.json"))
//	if err != nil {
//	    log.Fatal(err)
//	}
//	http.Handle("/docs/", h)
package swagger

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
)

// Handler serves the Swagger UI page.
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

// ServeHTTP implements http.Handler. All requests are served the Swagger UI
// HTML page (the JS and CSS are loaded from the configured CDN/paths by the
// browser at runtime).
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")
	_, _ = w.Write(h.pageHTML)
}

// config holds the resolved configuration for a Handler.
type config struct {
	specURL   string
	pageTitle string
	envBadge  string
	branding  *Branding
	jsPath    string
	cssPath   string
	darkMode  bool
	uiConfig  map[string]any // forwarded verbatim to SwaggerUIBundle(...)
	tmpl      *template.Template
}

func defaultConfig() *config {
	return &config{
		pageTitle: "API Reference",
		jsPath:    defaultJSCDN,
		cssPath:   defaultCSSCDN,
		darkMode:  true,
		uiConfig:  map[string]any{},
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
	CSSPath     string
	UIConfig    template.JS // JSON-encoded SwaggerUIBundle config object
}

func renderPage(cfg *config) ([]byte, error) {
	uc := make(map[string]any, len(cfg.uiConfig)+3)
	for k, v := range cfg.uiConfig {
		uc[k] = v
	}
	// Ensure essential keys are always set.
	if cfg.specURL != "" {
		uc["url"] = cfg.specURL
	}
	uc["dom_id"] = "#swagger-ui"
	if _, ok := uc["deepLinking"]; !ok {
		uc["deepLinking"] = true
	}
	if _, ok := uc["presets"]; !ok {
		uc["presets"] = template.JS("[SwaggerUIBundle.presets.apis, SwaggerUIStandalonePreset]")
	}
	if _, ok := uc["plugins"]; !ok {
		uc["plugins"] = template.JS("[SwaggerUIBundle.plugins.DownloadUrl]")
	}
	if _, ok := uc["layout"]; !ok {
		uc["layout"] = "StandaloneLayout"
	}

	configJSON, err := marshalUIConfig(uc)
	if err != nil {
		return nil, err
	}

	tmpl := cfg.tmpl
	if tmpl == nil {
		tmpl = defaultTemplate()
	}

	data := PageData{
		PageTitle: cfg.pageTitle,
		DarkMode:  cfg.darkMode,
		Branding:  cfg.branding,
		EnvBadge:  cfg.envBadge,
		JSPath:    cfg.jsPath,
		CSSPath:   cfg.cssPath,
		UIConfig:  template.JS(configJSON), // #nosec G203 -- JSON-safe
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

// marshalUIConfig serialises the UI config map to JSON, correctly handling
// template.JS values (which are already-rendered JS expressions and must not
// be double-quoted).
func marshalUIConfig(m map[string]any) (string, error) {
	var sb bytes.Buffer
	sb.WriteByte('{')
	first := true
	for k, v := range m {
		if !first {
			sb.WriteByte(',')
		}
		first = false

		key, err := json.Marshal(k)
		if err != nil {
			return "", err
		}
		sb.Write(key)
		sb.WriteByte(':')

		switch val := v.(type) {
		case template.JS:
			// Raw JS expression — embed as-is without quoting.
			sb.WriteString(string(val))
		default:
			enc, err := json.Marshal(val)
			if err != nil {
				return "", err
			}
			sb.Write(enc)
		}
	}
	sb.WriteByte('}')
	return sb.String(), nil
}

