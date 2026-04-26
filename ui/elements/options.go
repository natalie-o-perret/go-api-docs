package elements

import "html/template"

// Option is a function that tweaks the Handler config.
type Option func(*config)

// Router controls how Stoplight Elements handles navigation.
type Router string

const (
	RouterHash    Router = "hash"    // URL hash-based navigation (default, works everywhere)
	RouterMemory  Router = "memory"  // in-memory, no URL changes
	RouterHistory Router = "history" // HTML5 history API
)

// Layout controls the page layout.
type Layout string

const (
	LayoutSidebar    Layout = "sidebar"    // three-panel with left sidebar (default)
	LayoutStacked    Layout = "stacked"    // single scrollable column
	LayoutResponsive Layout = "responsive" // sidebar on wide screens, stacked on narrow
)

// WithSpecURL sets the URL the browser will fetch the OpenAPI spec from.
func WithSpecURL(url string) Option {
	return func(c *config) { c.specURL = url }
}

// WithPageTitle sets the HTML <title> of the page.
func WithPageTitle(title string) Option {
	return func(c *config) { c.pageTitle = title }
}

// WithDarkMode enables dark page chrome and sets the Elements color scheme.
func WithDarkMode() Option {
	return func(c *config) { c.darkMode = true }
}

// WithLightMode disables dark page chrome.
func WithLightMode() Option {
	return func(c *config) { c.darkMode = false }
}

// Branding controls the optional branded header rendered above Elements.
type Branding struct {
	LogoURL     string
	LogoAlt     string
	Title       string
	Subtitle    string
	FaviconURL  string
	FaviconType string // MIME type (default: "image/svg+xml")
}

// WithBranding sets a branded header above Elements.
func WithBranding(b Branding) Option {
	return func(c *config) { c.branding = &b }
}

// WithEnvBadge adds an environment label next to the brand title (e.g. "dev", "staging").
func WithEnvBadge(env string) Option {
	return func(c *config) { c.envBadge = env }
}

// WithRouter sets the navigation router (default: hash).
func WithRouter(r Router) Option {
	return func(c *config) { c.attrs["router"] = string(r) }
}

// WithLayout sets the page layout (default: sidebar).
func WithLayout(l Layout) Option {
	return func(c *config) { c.attrs["layout"] = string(l) }
}

// WithHideInternal hides operations marked with x-internal.
func WithHideInternal() Option {
	return func(c *config) { c.attrs["hide-internal"] = "true" }
}

// WithHideTryIt hides the "Try it" request panel.
func WithHideTryIt() Option {
	return func(c *config) { c.attrs["hide-try-it"] = "true" }
}

// WithTryItCORS sets the CORS policy for try-it requests.
// Valid values: "omit", "same-origin", "include".
func WithTryItCORS(policy string) Option {
	return func(c *config) { c.attrs["try-it-credentials-policy"] = policy }
}

// WithBasePath sets a base path to prepend to the server URL for try-it requests.
func WithBasePath(path string) Option {
	return func(c *config) { c.attrs["base-path"] = path }
}

// WithJSPath overrides the URL for the JS bundle (default: "/elements.js").
func WithJSPath(path string) Option {
	return func(c *config) { c.jsPath = path }
}

// WithCSSPath overrides the URL for the CSS file (default: "/elements.css").
func WithCSSPath(path string) Option {
	return func(c *config) { c.cssPath = path }
}

// WithTemplate replaces the default HTML page template.
// The template receives a [PageData] value.
func WithTemplate(tmpl *template.Template) Option {
	return func(c *config) { c.tmpl = tmpl }
}

// WithAttr sets any <elements-api> attribute not covered by the typed options.
func WithAttr(key, value string) Option {
	return func(c *config) { c.attrs[key] = value }
}

