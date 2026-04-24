package goscalar

import "html/template"

// Option configures a [Handler].
type Option func(*config)

// ---------------------------------------------------------------------------
// Scalar config options
// ---------------------------------------------------------------------------

// WithSpecURL sets a single OpenAPI spec loaded by URL at render time (no Go-side
// parsing – Scalar fetches it directly from the browser).
func WithSpecURL(url string) Option {
	return func(c *config) {
		c.scalarConfig["url"] = url
		c.sources = nil
	}
}

// Source is one entry in the multi-spec sources dropdown.
type Source struct {
	URL     string `json:"url,omitempty"`
	Title   string `json:"title,omitempty"`
	Slug    string `json:"slug,omitempty"`
	Default bool   `json:"default,omitempty"`
}

// WithSources configures the multi-spec dropdown (replaces WithSpecURL).
// Scalar loads each spec by URL at render time.
func WithSources(sources ...Source) Option {
	return func(c *config) {
		delete(c.scalarConfig, "url")
		c.sources = sources
	}
}

// Theme mirrors Scalar's built-in theme names.
type Theme string

const (
	ThemeDefault    Theme = "default"
	ThemeAlternate  Theme = "alternate"
	ThemeMoon       Theme = "moon"
	ThemePurple     Theme = "purple"
	ThemeSolarized  Theme = "solarized"
	ThemeBluePlanet Theme = "bluePlanet"
	ThemeDeepSpace  Theme = "deepSpace"
	ThemeSaturn     Theme = "saturn"
	ThemeKepler     Theme = "kepler"
	ThemeMars       Theme = "mars"
	ThemeNone       Theme = "none"
)

// WithTheme sets Scalar's built-in colour theme.
func WithTheme(t Theme) Option {
	return func(c *config) { c.scalarConfig["theme"] = t }
}

// Layout mirrors Scalar's layout names.
type Layout string

const (
	LayoutModern  Layout = "modern"
	LayoutClassic Layout = "classic"
)

// WithLayout sets the sidebar layout.
func WithLayout(l Layout) Option {
	return func(c *config) { c.scalarConfig["layout"] = l }
}

// WithDarkMode forces dark mode on.
func WithDarkMode() Option {
	return func(c *config) { c.scalarConfig["darkMode"] = true }
}

// WithCustomCSS injects CSS into the Scalar configuration object (passed as
// customCss to the JS API).
func WithCustomCSS(css string) Option {
	return func(c *config) { c.scalarConfig["customCss"] = css }
}

// ShowToolbar controls when the developer toolbar appears.
type ShowToolbar string

const (
	ShowToolbarAlways    ShowToolbar = "always"
	ShowToolbarLocalhost ShowToolbar = "localhost"
	ShowToolbarNever     ShowToolbar = "never"
)

// WithShowToolbar sets toolbar visibility (default: never).
func WithShowToolbar(v ShowToolbar) Option {
	return func(c *config) { c.scalarConfig["showToolbar"] = v }
}

// WithHideClientButton hides the "client" button in the UI.
func WithHideClientButton() Option {
	return func(c *config) { c.scalarConfig["hideClientButton"] = true }
}

// WithDisableAgent disables the Scalar agent.
func WithDisableAgent() Option {
	return func(c *config) { c.scalarConfig["agent"] = map[string]any{"disabled": true} }
}

// WithDisableMCP disables MCP integration.
func WithDisableMCP() Option {
	return func(c *config) { c.scalarConfig["mcp"] = map[string]any{"disabled": true} }
}

// WithShowDeveloperTools controls the developer tools panel.
// Typical values: "never", "always", "localhost".
func WithShowDeveloperTools(v string) Option {
	return func(c *config) { c.scalarConfig["showDeveloperTools"] = v }
}

// WithProxy sets an HTTP proxy for Scalar's try-it feature.
func WithProxy(url string) Option {
	return func(c *config) { c.scalarConfig["proxy"] = url }
}

// WithHideModels hides the Models section.
func WithHideModels() Option {
	return func(c *config) { c.scalarConfig["hideModels"] = true }
}

// WithHideDownloadButton hides the "download spec" button.
func WithHideDownloadButton() Option {
	return func(c *config) { c.scalarConfig["hideDownloadButton"] = true }
}

// WithHideSearch hides the search box.
func WithHideSearch() Option {
	return func(c *config) { c.scalarConfig["hideSearch"] = true }
}

// WithShowOperationID shows operation IDs in the sidebar.
func WithShowOperationID() Option {
	return func(c *config) { c.scalarConfig["showOperationId"] = true }
}

// WithOption is the escape hatch for any Scalar config key not yet covered by
// a typed option above.
//
//	goscalar.WithOption("tagsSorter", "alpha")
func WithOption(key string, value any) Option {
	return func(c *config) { c.scalarConfig[key] = value }
}

// ---------------------------------------------------------------------------
// Page / branding options
// ---------------------------------------------------------------------------

// WithPageTitle sets the HTML <title>.
func WithPageTitle(title string) Option {
	return func(c *config) { c.pageTitle = title }
}

// Branding controls the branded header rendered above the Scalar UI.
// Set LogoURL and/or FaviconURL to empty string to omit them.
type Branding struct {
	// LogoURL is the src of the logo image shown in the header (e.g. "/logo.svg").
	LogoURL string
	// LogoAlt is the alt text for the logo image.
	LogoAlt string
	// Title is the company/product name shown next to the logo.
	Title string
	// Subtitle is the secondary line shown under the title.
	Subtitle string
	// FaviconURL is used for the page <link rel="icon"> (e.g. "/favicon.svg").
	FaviconURL string
}

// WithBranding sets the branded header shown above the Scalar UI.
func WithBranding(b Branding) Option {
	return func(c *config) { c.branding = &b }
}

// WithEnvBadge adds a small environment badge next to the brand title
// (e.g. "preprod", "dev").  Pass an empty string to hide it.
func WithEnvBadge(env string) Option {
	return func(c *config) { c.envBadge = env }
}

// ---------------------------------------------------------------------------
// Asset path options
// ---------------------------------------------------------------------------

// WithJSPath overrides the URL path at which the vendored Scalar JS bundle is
// served (default: "/scalar.js").
func WithJSPath(path string) Option {
	return func(c *config) { c.jsPath = path }
}

// ---------------------------------------------------------------------------
// Template options
// ---------------------------------------------------------------------------

// WithTemplate replaces the default HTML page template.  The template receives
// a [PageData] value.
func WithTemplate(tmpl *template.Template) Option {
	return func(c *config) { c.tmpl = tmpl }
}

