// Example: openpotato — multi-spec UI for the two live OpenPotato public APIs.
//
//   - OpenHolidays API  https://openholidaysapi.org  – public & school holidays
//   - OpenPLZ API       https://openplzapi.org       – postal-code / street data
//
// Specs are fetched at startup; the servers block in each spec is rewritten so
// that Scalar's "Try it" requests hit /proxy/<api>/... on this server, which
// transparently forwards them to the real upstream.  No CORS, no auth tokens,
// everything flows through localhost:8085.
//
// Run from this directory:
//
//	go run .
//	open http://localhost:8085
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"

	scalar "github.com/nopereta/go-api-docs/ui/scalar"
)

// ── remote spec + proxy ───────────────────────────────────────────────────────

type remoteSpec struct {
	name     string
	upstream string // real API base URL, e.g. "https://openholidaysapi.org"
	data     []byte // spec JSON, served locally and rewritten
}

func (s *remoteSpec) fetch() error {
	specURL := s.upstream + "/swagger/v1/swagger.json"
	log.Printf("fetching %s spec from %s …", s.name, specURL)
	resp, err := http.Get(specURL) //nolint:noctx
	if err != nil {
		return fmt.Errorf("fetch %s: %w", s.name, err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("fetch %s: unexpected status %d", s.name, resp.StatusCode)
	}
	s.data, err = io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read %s: %w", s.name, err)
	}
	log.Printf("  ✓ %s  (%d bytes)", s.name, len(s.data))
	return nil
}

// rewriteServers replaces the spec's servers block so Scalar's "Try it" panel
// sends requests to localBase (the local proxy) instead of the real upstream.
func (s *remoteSpec) rewriteServers(localBase string) error {
	var doc map[string]any
	if err := json.Unmarshal(s.data, &doc); err != nil {
		return fmt.Errorf("parse %s spec: %w", s.name, err)
	}
	doc["servers"] = []map[string]any{{
		"url":         localBase,
		"description": fmt.Sprintf("%s (proxied via localhost)", s.name),
	}}
	var err error
	s.data, err = json.MarshalIndent(doc, "", "  ")
	return err
}

// proxy returns an http.Handler that strips prefix and forwards to upstream.
func (s *remoteSpec) proxy(prefix string) http.Handler {
	target, _ := url.Parse(s.upstream)
	rp := httputil.NewSingleHostReverseProxy(target)
	// Ensure Host header matches the upstream (required by some servers).
	rp.Rewrite = func(pr *httputil.ProxyRequest) {
		pr.SetURL(target)
		pr.Out.Host = target.Host
	}
	return http.StripPrefix(prefix, rp)
}

// ── theme ─────────────────────────────────────────────────────────────────────

const potatoCSS = `
.light-mode, .dark-mode {
  --scalar-color-1:              #e05c00;
  --scalar-color-accent:         #e05c00;
  --scalar-button-1:             #e05c00;
  --scalar-button-1-hover:       #bf4d00;
  --scalar-sidebar-color-active: #e05c00;
  --scalar-sidebar-background-active: rgba(224,92,0,.12);
}
.gs-header {
  background: linear-gradient(135deg, #1a0d00 0%, #2e1500 100%) !important;
  border-bottom: 1px solid rgba(224,92,0,.35) !important;
}
.gs-brand-title    { color: #e05c00 !important; }
.gs-brand-subtitle { color: #ffb380 !important; }
.gs-env-badge {
  background: rgba(224,92,0,.18) !important;
  color: #e05c00 !important;
  border-color: rgba(224,92,0,.45) !important;
}
`

// ── main ──────────────────────────────────────────────────────────────────────

const listenAddr = ":8085"
const baseURL = "http://localhost:8085"

func main() {
	specs := []*remoteSpec{
		{name: "OpenHolidays", upstream: "https://openholidaysapi.org"},
		{name: "OpenPLZ", upstream: "https://openplzapi.org"},
	}

	for _, s := range specs {
		if err := s.fetch(); err != nil {
			log.Fatal(err)
		}
	}

	// Rewrite each spec's servers block to point at our local proxy endpoints.
	if err := specs[0].rewriteServers(baseURL + "/proxy/openholidays"); err != nil {
		log.Fatal(err)
	}
	if err := specs[1].rewriteServers(baseURL + "/proxy/openplz"); err != nil {
		log.Fatal(err)
	}

	h, err := scalar.New(
		scalar.WithSources(
			scalar.Source{URL: "/specs/openholidays.json", Title: "OpenHolidays API", Default: true},
			scalar.Source{URL: "/specs/openplz.json", Title: "OpenPLZ API"},
		),
		scalar.WithTheme(scalar.ThemeNone),
		scalar.WithCustomCSS(potatoCSS),
		scalar.WithBranding(scalar.Branding{
			LogoURL:  "https://avatars.githubusercontent.com/u/108528374?s=32",
			LogoAlt:  "OpenPotato",
			Title:    "OpenPotato",
			Subtitle: "Free open-data APIs",
		}),
		scalar.WithEnvBadge("production"),
		scalar.With(scalar.DisableAgent, scalar.DarkMode),
		scalar.WithShowOperationID(),
		scalar.WithLayout(scalar.LayoutModern),
		scalar.WithPageTitle("OpenPotato APIs"),
	)
	if err != nil {
		log.Fatal(err)
	}

	mux := http.NewServeMux()
	mux.Handle("/", h)

	// Serve the (rewritten) specs.
	mux.HandleFunc("GET /specs/openholidays.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(specs[0].data)
	})
	mux.HandleFunc("GET /specs/openplz.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(specs[1].data)
	})

	// Transparent reverse proxies — "Try it" hits these instead of the real upstream.
	mux.Handle("/proxy/openholidays/", specs[0].proxy("/proxy/openholidays"))
	mux.Handle("/proxy/openplz/", specs[1].proxy("/proxy/openplz"))

	log.Printf("listening on %s", baseURL)
	log.Printf("  proxy: /proxy/openholidays  →  %s", specs[0].upstream)
	log.Printf("  proxy: /proxy/openplz       →  %s", specs[1].upstream)
	log.Fatal(http.ListenAndServe(listenAddr, mux))
}
