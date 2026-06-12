package scalar_test

import (
	"html/template"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/nopereta/go-api-docs/ui/scalar"
)

func TestNew_defaults(t *testing.T) {
	h, err := scalar.New()
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	if h == nil {
		t.Fatal("expected non-nil Handler")
	}
}

func TestNew_pageContainsConfig(t *testing.T) {
	h, err := scalar.New(
		scalar.WithSources(scalar.Source{URL: "/openapi.json", Title: "Current", Default: true}),
		scalar.WithTheme(scalar.ThemeNone),
		scalar.WithDarkMode(),
		scalar.WithDisableAgent(),
		scalar.WithDisableMCP(),
		scalar.WithHideClientButton(),
	)
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
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
	h, err := scalar.New()
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/scalar.js", http.NoBody)
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
	h, err := scalar.New(
		scalar.WithBranding(scalar.Branding{
			LogoURL:  "/logo.svg",
			Title:    "Acme",
			Subtitle: "Platform API",
		}),
		scalar.WithEnvBadge("preprod"),
	)
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	h.ServeHTTP(rec, req)

	body := rec.Body.String()
	for _, want := range []string{"Acme", "Platform API", "preprod", "/logo.svg"} {
		if !strings.Contains(body, want) {
			t.Errorf("response body missing %q", want)
		}
	}
}

func TestWithOption_escapeHatch(t *testing.T) {
	h, err := scalar.New(
		scalar.WithOption("tagsSorter", "alpha"),
	)
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	h.ServeHTTP(rec, req)

	if !strings.Contains(rec.Body.String(), `"tagsSorter":"alpha"`) {
		t.Error("expected tagsSorter in config JSON")
	}
}

func TestWithSpecURL(t *testing.T) {
	h, err := scalar.New(scalar.WithSpecURL("/api/openapi.yaml"))
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", http.NoBody))
	if !strings.Contains(rec.Body.String(), `/api/openapi.yaml`) {
		t.Error("expected spec URL in page")
	}
}

func TestWithSources_mutualExclusion(t *testing.T) {
	h, err := scalar.New(
		scalar.WithSpecURL("/old.json"),
		scalar.WithSources(scalar.Source{URL: "/new.json", Title: "New"}),
	)
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	body := pageBody(t, h)
	if strings.Contains(body, `/old.json`) {
		t.Error("WithSources should have cleared the WithSpecURL value")
	}
	if !strings.Contains(body, `/new.json`) {
		t.Error("expected new source URL in page")
	}
}

func TestWithLayout(t *testing.T) {
	h, err := scalar.New(scalar.WithLayout(scalar.LayoutClassic))
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	if !strings.Contains(pageBody(t, h), `"layout":"classic"`) {
		t.Error("expected layout in config JSON")
	}
}

func TestWithLightMode(t *testing.T) {
	h, err := scalar.New(scalar.WithLightMode())
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	body := pageBody(t, h)
	if !strings.Contains(body, `"darkMode":false`) {
		t.Error("expected darkMode:false in config JSON")
	}
	if !strings.Contains(body, "#ffffff") {
		t.Error("expected light background in page CSS")
	}
}

func TestWithCustomCSS(t *testing.T) {
	h, err := scalar.New(scalar.WithCustomCSS("body { font-size: 18px; }"))
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	if !strings.Contains(pageBody(t, h), "font-size") {
		t.Error("expected customCss in config JSON")
	}
}

func TestWithProxy(t *testing.T) {
	h, err := scalar.New(scalar.WithProxy("https://proxy.example.com"))
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	if !strings.Contains(pageBody(t, h), `"proxy":"https://proxy.example.com"`) {
		t.Error("expected proxy in config JSON")
	}
}

func TestWithBaseServerURL(t *testing.T) {
	h, err := scalar.New(scalar.WithBaseServerURL("https://api.example.com"))
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	if !strings.Contains(pageBody(t, h), `https://api.example.com`) {
		t.Error("expected servers URL in config JSON")
	}
}

func TestWithHideModels(t *testing.T) {
	h, err := scalar.New(scalar.WithHideModels())
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	if !strings.Contains(pageBody(t, h), `"hideModels":true`) {
		t.Error("expected hideModels in config JSON")
	}
}

func TestWithHideDownloadButton(t *testing.T) {
	h, err := scalar.New(scalar.WithHideDownloadButton())
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	if !strings.Contains(pageBody(t, h), `"hideDownloadButton":true`) {
		t.Error("expected hideDownloadButton in config JSON")
	}
}

func TestWithHideSearch(t *testing.T) {
	h, err := scalar.New(scalar.WithHideSearch())
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	if !strings.Contains(pageBody(t, h), `"hideSearch":true`) {
		t.Error("expected hideSearch in config JSON")
	}
}

func TestWithShowOperationID(t *testing.T) {
	h, err := scalar.New(scalar.WithShowOperationID())
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	if !strings.Contains(pageBody(t, h), `"showOperationId":true`) {
		t.Error("expected showOperationId in config JSON")
	}
}

func TestWithShowDeveloperTools(t *testing.T) {
	h, err := scalar.New(scalar.WithShowDeveloperTools(scalar.ShowToolbarAlways))
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	if !strings.Contains(pageBody(t, h), `"showDeveloperTools":"always"`) {
		t.Error("expected showDeveloperTools in config JSON")
	}
}

func TestWithJSPath_custom(t *testing.T) {
	h, err := scalar.New(scalar.WithJSPath("https://cdn.jsdelivr.net/npm/@scalar/api-reference/dist/browser/standalone.js"))
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	if !strings.Contains(pageBody(t, h), "cdn.jsdelivr.net") {
		t.Error("expected custom JS path in page src")
	}
}

func TestWithTemplate_custom(t *testing.T) {
	tmpl := template.Must(template.New("custom").Parse(`<html><body>CUSTOM {{.PageTitle}}</body></html>`))
	h, err := scalar.New(
		scalar.WithTemplate(tmpl),
		scalar.WithPageTitle("My API"),
	)
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	body := pageBody(t, h)
	if !strings.Contains(body, "CUSTOM My API") {
		t.Errorf("expected custom template output, got: %s", body)
	}
}

func TestWithBranding_faviconType(t *testing.T) {
	h, err := scalar.New(scalar.WithBranding(scalar.Branding{
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

func TestWithBranding_faviconType_defaultsSVG(t *testing.T) {
	h, err := scalar.New(scalar.WithBranding(scalar.Branding{
		FaviconURL: "/favicon.svg",
	}))
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	if !strings.Contains(pageBody(t, h), `type="image/svg+xml"`) {
		t.Error("expected default favicon type image/svg+xml")
	}
}

func TestHandler_htmlCacheHeader(t *testing.T) {
	h, err := scalar.New()
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", http.NoBody))
	if cc := rec.Header().Get("Cache-Control"); cc != "no-cache" {
		t.Errorf("expected Cache-Control: no-cache on HTML, got %q", cc)
	}
}

func TestHandler_jsCacheHeader(t *testing.T) {
	h, err := scalar.New()
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/scalar.js", http.NoBody))
	cc := rec.Header().Get("Cache-Control")
	if !strings.Contains(cc, "immutable") {
		t.Errorf("expected immutable cache for JS, got %q", cc)
	}
}

// ── CustomTheme / WithCustomTheme tests ──────────────────────────────────────

func TestNewCustomTheme_empty(t *testing.T) {
	ct := scalar.NewCustomTheme()
	if got := ct.CSS(); got != "" {
		t.Errorf("expected empty CSS for zero-value theme, got %q", got)
	}
}

func TestCustomTheme_CSS_containsVars(t *testing.T) {
	css := scalar.NewCustomTheme().
		Accent("#2563eb").
		Background("#0f172a").
		Text("#f1f5f9").
		Font("'Inter', sans-serif").
		Radius("6px").
		CSS()

	for _, want := range []string{
		"--scalar-color-accent: #2563eb",
		"--scalar-background-1: #0f172a",
		"--scalar-color-1: #f1f5f9",
		"--scalar-font: 'Inter', sans-serif",
		"--scalar-radius: 6px",
	} {
		if !strings.Contains(css, want) {
			t.Errorf("CSS() missing %q\ngot:\n%s", want, css)
		}
	}
}

func TestWithCustomTheme_appearsInPage(t *testing.T) {
	h, err := scalar.New(
		scalar.WithCustomTheme(
			scalar.NewCustomTheme().
				Accent("#2563eb").
				Background("#0f172a"),
		),
	)
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	body := pageBody(t, h)
	for _, want := range []string{"--scalar-color-accent", "--scalar-background-1"} {
		if !strings.Contains(body, want) {
			t.Errorf("page missing CSS variable %q", want)
		}
	}
}

func TestWithCustomTheme_composesWithCustomCSS(t *testing.T) {
	h, err := scalar.New(
		scalar.WithCustomTheme(scalar.NewCustomTheme().Accent("#2563eb")),
		scalar.WithCustomCSS("body { font-size: 18px; }"),
	)
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	body := pageBody(t, h)
	if !strings.Contains(body, "--scalar-color-accent") {
		t.Error("page missing theme variable after composing with WithCustomCSS")
	}
	if !strings.Contains(body, "font-size: 18px") {
		t.Error("page missing extra CSS after composing with WithCustomCSS")
	}
}

func TestCustomTheme_Var_escapeHatch(t *testing.T) {
	css := scalar.NewCustomTheme().
		Var("--scalar-custom-prop", "42px").
		CSS()
	if !strings.Contains(css, "--scalar-custom-prop: 42px") {
		t.Errorf("Var() escape hatch not rendered, got:\n%s", css)
	}
}

func TestCustomTheme_sidebarMethods(t *testing.T) {
	css := scalar.NewCustomTheme().
		SidebarBackground("#1e293b").
		SidebarText("#cbd5e1").
		SidebarAccent("#38bdf8").
		CSS()

	for _, want := range []string{
		"--scalar-sidebar-background-1",
		"--scalar-sidebar-color-1",
		"--scalar-sidebar-color-accent",
	} {
		if !strings.Contains(css, want) {
			t.Errorf("CSS() missing sidebar variable %q", want)
		}
	}
}

// pageBody is a test helper that GETs "/" and returns the response body.
func pageBody(t *testing.T, h *scalar.Handler) string {
	t.Helper()
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", http.NoBody))
	return rec.Body.String()
}

func TestWith_variadicFlags(t *testing.T) {
	h, err := scalar.New(scalar.With(
		scalar.DisableAgent,
		scalar.DisableMCP,
		scalar.HideClientButton,
		scalar.HideModels,
		scalar.HideDownloadButton,
		scalar.HideSearch,
		scalar.ShowOperationID,
		scalar.DarkMode,
	))
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	body := pageBody(t, h)
	for _, want := range []string{
		`"hideClientButton":true`,
		`"hideModels":true`,
		`"hideDownloadButton":true`,
		`"hideSearch":true`,
		`"showOperationId":true`,
		`"darkMode":true`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("missing %q in config JSON", want)
		}
	}
}

func TestWith_combinedWithPipe(t *testing.T) {
	h, err := scalar.New(scalar.With(scalar.DisableAgent | scalar.HideClientButton))
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	if !strings.Contains(pageBody(t, h), `"hideClientButton":true`) {
		t.Error("expected hideClientButton from combined flag")
	}
}

func TestWith_lightModeWinsOverDarkMode(t *testing.T) {
	h, err := scalar.New(scalar.With(scalar.DarkMode | scalar.LightMode))
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	if !strings.Contains(pageBody(t, h), `"darkMode":false`) {
		t.Error("LightMode should override DarkMode when both set")
	}
}
