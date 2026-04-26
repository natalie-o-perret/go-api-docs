// Example: basic — Swagger Petstore proxied through localhost.
//
// The Petstore spec is fetched at startup and its servers block is rewritten to
// point at the local /proxy/ prefix.  "Try it" requests flow through this
// server rather than hitting petstore3.swagger.io directly — no CORS issues.
//
// Run from this directory:
//
//	go run .
//	open http://localhost:8080
package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"

	scalar "github.com/nopereta/go-api-docs/ui/scalar"
)

const (
	listenAddr   = ":8080"
	baseURL      = "http://localhost:8080"
	upstreamBase = "https://petstore3.swagger.io"
	specEndpoint = upstreamBase + "/api/v3/openapi.json"
	proxyPrefix  = "/proxy"
)

func main() {
	// ── 1. fetch spec ────────────────────────────────────────────────────────
	log.Printf("fetching spec from %s …", specEndpoint)
	resp, err := http.Get(specEndpoint) //nolint:noctx
	if err != nil {
		log.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("  ✓ %d bytes", len(raw))

	// ── 2. rewrite servers block → local proxy ────────────────────────────────
	var doc map[string]any
	if err := json.Unmarshal(raw, &doc); err != nil {
		log.Fatal("parse spec:", err)
	}
	doc["servers"] = []map[string]any{{
		"url":         baseURL + proxyPrefix + "/api/v3",
		"description": "Petstore (proxied via localhost)",
	}}
	specJSON, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		log.Fatal(err)
	}

	// ── 3. scalar handler ────────────────────────────────────────────────────
	h, err := scalar.New(
		scalar.WithSpecURL("/openapi.json"),
		scalar.WithTheme(scalar.ThemeDefault),
		scalar.With(scalar.DisableAgent, scalar.DisableMCP, scalar.DarkMode),
		scalar.WithPageTitle("Petstore API"),
	)
	if err != nil {
		log.Fatal(err)
	}

	// ── 4. reverse proxy → petstore3.swagger.io ───────────────────────────────
	target, _ := url.Parse(upstreamBase)
	rp := httputil.NewSingleHostReverseProxy(target)
	rp.Rewrite = func(pr *httputil.ProxyRequest) {
		pr.SetURL(target)
		pr.Out.Host = target.Host
	}

	// ── 5. routes ────────────────────────────────────────────────────────────
	mux := http.NewServeMux()
	mux.Handle("/", h)
	mux.HandleFunc("GET /openapi.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(specJSON)
	})
	mux.Handle(proxyPrefix+"/", http.StripPrefix(proxyPrefix, rp))

	log.Printf("listening on %s", baseURL)
	log.Printf("  proxy: %s/  →  %s", proxyPrefix, upstreamBase)
	log.Fatal(http.ListenAndServe(listenAddr, mux))
}
