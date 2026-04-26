package redoc_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/nopereta/go-api-docs/ui/redoc"
)

func TestNew_defaults(t *testing.T) {
	h, err := redoc.New()
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	if h == nil {
		t.Fatal("expected non-nil Handler")
	}
}

func TestNew_pageContainsRedoc(t *testing.T) {
	h, err := redoc.New(redoc.WithSpecURL("/openapi.json"))
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	body := pageBody(t, h)
	for _, want := range []string{
		"Redoc.init",
		`"/openapi.json"`,
		"redoc-container",
	} {
		if !strings.Contains(body, want) {
			t.Errorf("response body missing %q", want)
		}
	}
}

func TestHandler_htmlCacheHeader(t *testing.T) {
	h, err := redoc.New()
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if cc := rec.Header().Get("Cache-Control"); cc != "no-cache" {
		t.Errorf("expected Cache-Control: no-cache, got %q", cc)
	}
}

func TestHandler_contentType(t *testing.T) {
	h, err := redoc.New()
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	ct := rec.Header().Get("Content-Type")
	if !strings.HasPrefix(ct, "text/html") {
		t.Errorf("expected text/html Content-Type, got %q", ct)
	}
}

func TestHandler_jsAsset(t *testing.T) {
	h, err := redoc.New()
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/redoc.js", nil))
	ct := rec.Header().Get("Content-Type")
	if !strings.HasPrefix(ct, "application/javascript") {
		t.Errorf("expected application/javascript, got %q", ct)
	}
	if cc := rec.Header().Get("Cache-Control"); !strings.Contains(cc, "immutable") {
		t.Errorf("expected immutable cache header for JS, got %q", cc)
	}
}

func TestWithPageTitle(t *testing.T) {
	h, err := redoc.New(redoc.WithPageTitle("My API Docs"))
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	if !strings.Contains(pageBody(t, h), "My API Docs") {
		t.Error("expected custom page title in page")
	}
}

func TestWithBranding(t *testing.T) {
	h, err := redoc.New(
		redoc.WithBranding(redoc.Branding{
			LogoURL:  "/logo.svg",
			Title:    "Acme",
			Subtitle: "Platform API",
		}),
		redoc.WithEnvBadge("staging"),
	)
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	body := pageBody(t, h)
	for _, want := range []string{"Acme", "Platform API", "staging", "/logo.svg"} {
		if !strings.Contains(body, want) {
			t.Errorf("response body missing %q", want)
		}
	}
}

func TestWithLightMode(t *testing.T) {
	h, err := redoc.New(redoc.WithLightMode())
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	if !strings.Contains(pageBody(t, h), "#ffffff") {
		t.Error("expected light background in page CSS")
	}
}

func TestWithDarkMode(t *testing.T) {
	h, err := redoc.New(redoc.WithDarkMode())
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	if !strings.Contains(pageBody(t, h), "#1a1a2e") {
		t.Error("expected dark background in page CSS")
	}
}

func TestWithJSPath(t *testing.T) {
	h, err := redoc.New(redoc.WithJSPath("https://cdn.example.com/redoc.js"))
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	if !strings.Contains(pageBody(t, h), "cdn.example.com/redoc.js") {
		t.Error("expected custom JS path in page")
	}
}

func TestWithHideDownloadButton(t *testing.T) {
	h, err := redoc.New(redoc.WithHideDownloadButton())
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	if !strings.Contains(pageBody(t, h), `"hideDownloadButton":true`) {
		t.Error("expected hideDownloadButton:true in config")
	}
}

func TestWithExpandResponses(t *testing.T) {
	h, err := redoc.New(redoc.WithExpandResponses("all"))
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	if !strings.Contains(pageBody(t, h), `"expandResponses":"all"`) {
		t.Error("expected expandResponses:all in config")
	}
}

func TestWithOption_escapeHatch(t *testing.T) {
	h, err := redoc.New(redoc.WithOption("requiredPropsFirst", true))
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	if !strings.Contains(pageBody(t, h), `"requiredPropsFirst":true`) {
		t.Error("expected requiredPropsFirst:true in config")
	}
}

func pageBody(t *testing.T, h *redoc.Handler) string {
	t.Helper()
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	return rec.Body.String()
}

