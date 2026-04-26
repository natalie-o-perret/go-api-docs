package swagger_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/nopereta/go-api-docs/ui/swagger"
)

func TestNew_defaults(t *testing.T) {
	h, err := swagger.New()
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	if h == nil {
		t.Fatal("expected non-nil Handler")
	}
}

func TestNew_pageContainsSwaggerUI(t *testing.T) {
	h, err := swagger.New(swagger.WithSpecURL("/openapi.json"))
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	body := pageBody(t, h)
	for _, want := range []string{
		"swagger-ui",
		"SwaggerUIBundle",
		`/openapi.json`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("response body missing %q", want)
		}
	}
}

func TestHandler_htmlCacheHeader(t *testing.T) {
	h, err := swagger.New()
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
	h, err := swagger.New()
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

func TestWithPageTitle(t *testing.T) {
	h, err := swagger.New(swagger.WithPageTitle("My API Docs"))
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	if !strings.Contains(pageBody(t, h), "My API Docs") {
		t.Error("expected custom page title in page")
	}
}

func TestWithBranding(t *testing.T) {
	h, err := swagger.New(
		swagger.WithBranding(swagger.Branding{
			LogoURL:  "/logo.svg",
			Title:    "Acme",
			Subtitle: "Platform API",
		}),
		swagger.WithEnvBadge("staging"),
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

func TestWithBranding_faviconType(t *testing.T) {
	h, err := swagger.New(swagger.WithBranding(swagger.Branding{
		FaviconURL:  "/favicon.ico",
		FaviconType: "image/x-icon",
	}))
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	if !strings.Contains(pageBody(t, h), `type="image/x-icon"`) {
		t.Error("expected custom favicon MIME type in page")
	}
}

func TestWithLightMode(t *testing.T) {
	h, err := swagger.New(swagger.WithLightMode())
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	body := pageBody(t, h)
	if !strings.Contains(body, "#ffffff") {
		t.Error("expected light background in page CSS")
	}
}

func TestWithDarkMode(t *testing.T) {
	h, err := swagger.New(swagger.WithDarkMode())
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	body := pageBody(t, h)
	if !strings.Contains(body, "#09090c") {
		t.Error("expected dark background in page CSS")
	}
}

func TestWithJSPath(t *testing.T) {
	h, err := swagger.New(swagger.WithJSPath("https://cdn.example.com/swagger-ui.js"))
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	if !strings.Contains(pageBody(t, h), "cdn.example.com/swagger-ui.js") {
		t.Error("expected custom JS path in page")
	}
}

func TestWithCSSPath(t *testing.T) {
	h, err := swagger.New(swagger.WithCSSPath("https://cdn.example.com/swagger-ui.css"))
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	if !strings.Contains(pageBody(t, h), "cdn.example.com/swagger-ui.css") {
		t.Error("expected custom CSS path in page")
	}
}

func TestWithDocExpansion(t *testing.T) {
	h, err := swagger.New(swagger.WithDocExpansion(swagger.DocExpansionFull))
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	if !strings.Contains(pageBody(t, h), `"docExpansion":"full"`) {
		t.Error("expected docExpansion:full in UIConfig")
	}
}

func TestWithFilter(t *testing.T) {
	h, err := swagger.New(swagger.WithFilter(""))
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	if !strings.Contains(pageBody(t, h), `"filter":true`) {
		t.Error("expected filter:true in UIConfig")
	}
}

func TestWithPersistAuthorization(t *testing.T) {
	h, err := swagger.New(swagger.WithPersistAuthorization())
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	if !strings.Contains(pageBody(t, h), `"persistAuthorization":true`) {
		t.Error("expected persistAuthorization:true in UIConfig")
	}
}

func TestWithDisplayRequestDuration(t *testing.T) {
	h, err := swagger.New(swagger.WithDisplayRequestDuration())
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	if !strings.Contains(pageBody(t, h), `"displayRequestDuration":true`) {
		t.Error("expected displayRequestDuration:true in UIConfig")
	}
}

func TestWithTryItOutEnabled(t *testing.T) {
	h, err := swagger.New(swagger.WithTryItOutEnabled())
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	if !strings.Contains(pageBody(t, h), `"tryItOutEnabled":true`) {
		t.Error("expected tryItOutEnabled:true in UIConfig")
	}
}

func TestWithOption_escapeHatch(t *testing.T) {
	h, err := swagger.New(swagger.WithOption("tagsSorter", "alpha"))
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	if !strings.Contains(pageBody(t, h), `"tagsSorter":"alpha"`) {
		t.Error("expected tagsSorter in UIConfig")
	}
}

func TestWithDeepLinking_false(t *testing.T) {
	h, err := swagger.New(swagger.WithDeepLinking(false))
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	if !strings.Contains(pageBody(t, h), `"deepLinking":false`) {
		t.Error("expected deepLinking:false in UIConfig")
	}
}

// pageBody is a test helper that GETs "/" and returns the response body.
func pageBody(t *testing.T, h *swagger.Handler) string {
	t.Helper()
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	return rec.Body.String()
}
