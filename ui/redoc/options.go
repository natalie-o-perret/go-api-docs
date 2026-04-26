package redoc

import "html/template"

// Option is a function that tweaks the Handler config.
type Option func(*config)

// WithSpecURL sets the URL the browser will fetch the OpenAPI spec from.
func WithSpecURL(url string) Option {
	return func(c *config) { c.specURL = url }
}

// WithPageTitle sets the HTML <title> of the page.
func WithPageTitle(title string) Option {
	return func(c *config) { c.pageTitle = title }
}

// WithDarkMode enables dark page chrome.
func WithDarkMode() Option {
	return func(c *config) { c.darkMode = true }
}

// WithLightMode disables dark page chrome.
func WithLightMode() Option {
	return func(c *config) { c.darkMode = false }
}

// Branding controls the optional branded header rendered above Redoc.
type Branding struct {
	LogoURL     string
	LogoAlt     string
	Title       string
	Subtitle    string
	FaviconURL  string
	FaviconType string // MIME type (default: "image/svg+xml")
}

// WithBranding sets a branded header above the Redoc UI.
func WithBranding(b Branding) Option {
	return func(c *config) { c.branding = &b }
}

// WithEnvBadge adds an environment label next to the brand title (e.g. "dev", "staging").
func WithEnvBadge(env string) Option {
	return func(c *config) { c.envBadge = env }
}

// WithJSPath sets the URL path where the JS bundle is served (default: "/redoc.js").
// Point at a CDN to skip the embedded bundle.
func WithJSPath(path string) Option {
	return func(c *config) { c.jsPath = path }
}

// WithTemplate replaces the default HTML page template.
// The template receives a [PageData] value.
func WithTemplate(tmpl *template.Template) Option {
	return func(c *config) { c.tmpl = tmpl }
}

// WithHideDownloadButton hides the download spec button.
func WithHideDownloadButton() Option {
	return func(c *config) { c.redocOptions["hideDownloadButton"] = true }
}

// WithDisableSearch disables the search box.
func WithDisableSearch() Option {
	return func(c *config) { c.redocOptions["disableSearch"] = true }
}

// WithExpandResponses sets which HTTP response codes expand by default.
// Pass a comma-separated list (e.g. "200,201") or "all".
func WithExpandResponses(codes string) Option {
	return func(c *config) { c.redocOptions["expandResponses"] = codes }
}

// WithRequiredPropsFirst shows required properties before optional ones.
func WithRequiredPropsFirst() Option {
	return func(c *config) { c.redocOptions["requiredPropsFirst"] = true }
}

// WithSortPropsAlphabetically sorts object properties alphabetically.
func WithSortPropsAlphabetically() Option {
	return func(c *config) { c.redocOptions["sortPropsAlphabetically"] = true }
}

// WithNoAutoAuth disables automatic auth detection.
func WithNoAutoAuth() Option {
	return func(c *config) { c.redocOptions["noAutoAuth"] = true }
}

// WithPathInMiddlePanel moves the path to the middle panel.
func WithPathInMiddlePanel() Option {
	return func(c *config) { c.redocOptions["pathInMiddlePanel"] = true }
}

// WithHideHostname hides the hostname in operation paths.
func WithHideHostname() Option {
	return func(c *config) { c.redocOptions["hideHostname"] = true }
}

// WithExpandSingleSchemaField automatically expands single-field schema objects.
func WithExpandSingleSchemaField() Option {
	return func(c *config) { c.redocOptions["expandSingleSchemaField"] = true }
}

// WithSchemaExpansionLevel sets how many levels deep schemas expand by default.
// Pass 0 to collapse all, or "all" to expand everything.
func WithSchemaExpansionLevel(level any) Option {
	return func(c *config) { c.redocOptions["schemaExpansionLevel"] = level }
}

// WithLazyRendering defers rendering of off-screen operations for faster load.
func WithLazyRendering() Option {
	return func(c *config) { c.redocOptions["lazyRendering"] = true }
}

// WithShowObjectSchemaExamples shows examples in schema objects.
func WithShowObjectSchemaExamples() Option {
	return func(c *config) { c.redocOptions["showObjectSchemaExamples"] = true }
}

// WithHideServerSelection hides the server selection dropdown.
func WithHideServerSelection() Option {
	return func(c *config) { c.redocOptions["hideServerSelection"] = true }
}

// WithScrollYOffset sets a fixed scroll offset (e.g. for sticky headers).
func WithScrollYOffset(px int) Option {
	return func(c *config) { c.redocOptions["scrollYOffset"] = px }
}

// WithTheme sets Redoc's built-in theming object. The value is passed verbatim
// to the Redoc.init options as the "theme" key.
//
//	redoc.WithTheme(map[string]any{
//	    "colors": map[string]any{"primary": map[string]any{"main": "#2563eb"}},
//	})
func WithTheme(theme map[string]any) Option {
	return func(c *config) { c.redocOptions["theme"] = theme }
}

// WithOption sets any Redoc option key not covered by the typed options above.
func WithOption(key string, value any) Option {
	return func(c *config) { c.redocOptions[key] = value }
}

