package elements_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/nopereta/go-api-docs/ui/elements"
)

func TestNew_defaults(t *testing.T) {
	h, err := elements.New()
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	if h == nil {
		t.Fatal("expected non-nil Handler")
	}
}

func TestNew_pageContainsElements(t *testing.T) {
	h, err := elements.New(elements.WithSpecURL("/openapi.json"))
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	body := pageBody(t, h)
	for _, want := range []string{
		"elements-api",
		"/openapi.json",
		"elements.js",
		"elements.css",
	} {
		if !strings.Contains(body, want) {
			t.Errorf("response body missing %q", want)
		}
	}
}

func TestHandler_htmlCacheHeader(t *testing.T) {
	h, err := elements.New()
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if cc := rec.Header().Get("Cache-Control"); cc != "no-cache" {
		t.Errorf("expected Cache-Control: no-cache, got %q", cc)
	}
}

func TestHandler_jsAsset(t *testing.T) {
	h, err := elements.New()
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/elements.js", nil))
	if ct := rec.Header().Get("Content-Type"); !strings.HasPrefix(ct, "application/javascript") {
		t.Errorf("expected application/javascript, got %q", ct)
	}
	if cc := rec.Header().Get("Cache-Control"); !strings.Contains(cc, "immutable") {
		t.Errorf("expected immutable cache header, got %q", cc)
	}
}

func TestHandler_cssAsset(t *testing.T) {
	h, err := elements.New()
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/elements.css", nil))
	if ct := rec.Header().Get("Content-Type"); !strings.HasPrefix(ct, "text/css") {
		t.Errorf("expected text/css, got %q", ct)
	}
	if cc := rec.Header().Get("Cache-Control"); !strings.Contains(cc, "immutable") {
		t.Errorf("expected immutable cache header, got %q", cc)
	}
}

func TestWithPageTitle(t *testing.T) {
	h, err := elements.New(elements.WithPageTitle("My API Docs"))
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	if !strings.Contains(pageBody(t, h), "My API Docs") {
		t.Error("expected custom page title in page")
	}
}

func TestWithBranding(t *testing.T) {
	h, err := elements.New(
		elements.WithBranding(elements.Branding{
			LogoURL:  "/logo.svg",
			Title:    "Acme",
			Subtitle: "Platform API",
		}),
		elements.WithEnvBadge("staging"),
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

func TestWithLayout(t *testing.T) {
	h, err := elements.New(elements.WithLayout(elements.LayoutStacked))
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	if !strings.Contains(pageBody(t, h), `layout="stacked"`) {
		t.Error("expected layout=stacked in element attributes")
	}
}

func TestWithRouter(t *testing.T) {
	h, err := elements.New(elements.WithRouter(elements.RouterHistory))
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	if !strings.Contains(pageBody(t, h), `router="history"`) {
		t.Error("expected router=history in element attributes")
	}
}

func TestWithHideTryIt(t *testing.T) {
	h, err := elements.New(elements.WithHideTryIt())
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	if !strings.Contains(pageBody(t, h), `hide-try-it="true"`) {
		t.Error("expected hide-try-it in element attributes")
	}
}

func TestWithJSPath(t *testing.T) {
	h, err := elements.New(elements.WithJSPath("https://cdn.example.com/elements.js"))
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	if !strings.Contains(pageBody(t, h), "cdn.example.com/elements.js") {
		t.Error("expected custom JS path in page")
	}
}

func TestWithLightMode(t *testing.T) {
	h, err := elements.New(elements.WithLightMode())
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	if !strings.Contains(pageBody(t, h), "#ffffff") {
		t.Error("expected light background in page CSS")
	}
}

func pageBody(t *testing.T, h *elements.Handler) string {
	t.Helper()
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	return rec.Body.String()
}

