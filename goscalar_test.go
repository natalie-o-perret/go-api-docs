package goscalar_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/nopereta/goscalar"
)

func TestNew_defaults(t *testing.T) {
	h, err := goscalar.New()
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	if h == nil {
		t.Fatal("expected non-nil Handler")
	}
}

func TestNew_pageContainsConfig(t *testing.T) {
	h, err := goscalar.New(
		goscalar.WithSources(goscalar.Source{URL: "/openapi.json", Title: "Current", Default: true}),
		goscalar.WithTheme(goscalar.ThemeNone),
		goscalar.WithDarkMode(),
		goscalar.WithDisableAgent(),
		goscalar.WithDisableMCP(),
		goscalar.WithHideClientButton(),
	)
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	h.ServeHTTP(rec, req)

	body := rec.Body.String()
	for _, want := range []string{
		"scalar-mount",
		"createApiReference",
		`"theme":"none"`,
		`"darkMode":true`,
		`"hideClientButton":true`,
		`/openapi.json`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("response body missing %q", want)
		}
	}
}

func TestHandler_servesJS(t *testing.T) {
	h, err := goscalar.New()
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/scalar.js", nil)
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	ct := rec.Header().Get("Content-Type")
	if !strings.HasPrefix(ct, "application/javascript") {
		t.Errorf("unexpected Content-Type: %q", ct)
	}
	if rec.Body.Len() == 0 {
		t.Error("expected non-empty JS body")
	}
}

func TestHandler_branding(t *testing.T) {
	h, err := goscalar.New(
		goscalar.WithBranding(goscalar.Branding{
			LogoURL:  "/logo.svg",
			Title:    "Acme",
			Subtitle: "Platform API",
		}),
		goscalar.WithEnvBadge("preprod"),
	)
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	h.ServeHTTP(rec, req)

	body := rec.Body.String()
	for _, want := range []string{"Acme", "Platform API", "preprod", "/logo.svg"} {
		if !strings.Contains(body, want) {
			t.Errorf("response body missing %q", want)
		}
	}
}

func TestWithOption_escapeHatch(t *testing.T) {
	h, err := goscalar.New(
		goscalar.WithOption("tagsSorter", "alpha"),
	)
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	h.ServeHTTP(rec, req)

	if !strings.Contains(rec.Body.String(), `"tagsSorter":"alpha"`) {
		t.Error("expected tagsSorter in config JSON")
	}
}

