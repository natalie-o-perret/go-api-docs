// Package elements provides a Stoplight Elements API docs handler for Go.
// The JS bundle and CSS are vendored via go:embed but you can swap them with
// WithJSPath / WithCSSPath.
//
// Quick start:
//
//	h, err := elements.New(elements.WithSpecURL("/openapi.json"))
//	if err != nil {
//	    log.Fatal(err)
//	}
//	http.Handle("/docs/", h)
package elements

import (
	"bytes"
	"fmt"
	"html/template"
	"net/http"
	"sort"

	"github.com/nopereta/go-api-docs/ui/elements/assets"
)

// Handler serves the Stoplight Elements UI and its assets.
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
		_, _ = w.Write(assets.ElementsJS)
	case h.cfg.cssPath:
		w.Header().Set("Content-Type", "text/css; charset=utf-8")
		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		_, _ = w.Write(assets.ElementsCSS)
	default:
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Cache-Control", "no-cache")
		_, _ = w.Write(h.pageHTML)
	}
}

type config struct {
	specURL   string
	pageTitle string
	envBadge  string
	branding  *Branding
	jsPath    string
	cssPath   string
	darkMode  bool
	attrs     map[string]string // <elements-api> attributes
	tmpl      *template.Template
}

func defaultConfig() *config {
	return &config{
		pageTitle: "API Reference",
		jsPath:    "/elements.js",
		cssPath:   "/elements.css",
		darkMode:  true,
		attrs: map[string]string{
			"router": string(RouterHash),
			"layout": string(LayoutSidebar),
		},
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
	// ElementTag is the pre-rendered <elements-api ...></elements-api> element.
	// Built in Go so html/template never sees raw attribute injection inside a tag.
	ElementTag template.HTML
}

func renderPage(cfg *config) ([]byte, error) {
	// Merge all attributes: cfg.attrs + spec URL.
	merged := make(map[string]string, len(cfg.attrs)+1)
	for k, v := range cfg.attrs {
		merged[k] = v
	}
	merged["api-description-url"] = cfg.specURL

	keys := make([]string, 0, len(merged))
	for k := range merged {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Build the full <elements-api ...></elements-api> tag in Go so that
	// html/template never sees raw attribute injection inside a tag context
	// (which triggers ZgotmplZ sanitisation).
	var elemBuf bytes.Buffer
	elemBuf.WriteString("<elements-api")
	for _, k := range keys {
		elemBuf.WriteByte(' ')
		elemBuf.WriteString(k)
		elemBuf.WriteString(`="`)
		elemBuf.WriteString(template.HTMLEscapeString(merged[k]))
		elemBuf.WriteByte('"')
	}
	elemBuf.WriteString("></elements-api>")

	tmpl := cfg.tmpl
	if tmpl == nil {
		tmpl = defaultTemplate()
	}

	data := PageData{
		PageTitle:  cfg.pageTitle,
		DarkMode:   cfg.darkMode,
		Branding:   cfg.branding,
		EnvBadge:   cfg.envBadge,
		JSPath:     cfg.jsPath,
		CSSPath:    cfg.cssPath,
		ElementTag: template.HTML(elemBuf.String()), // #nosec G203 -- HTML-escaped above
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

