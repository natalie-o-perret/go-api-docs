package swagger

import "html/template"

// Default local paths where the embedded Swagger UI assets are served.
// Override any of them with WithJSPath / WithPresetPath / WithCSSPath to point
// at a CDN or a self-hosted copy instead.
const (
	defaultJSPath     = "/swagger-ui-bundle.js"
	defaultPresetPath = "/swagger-ui-standalone-preset.js"
	defaultCSSPath    = "/swagger-ui.css"
)

// Option is a function that tweaks the Handler config.
type Option func(*config)

// WithSpecURL sets the URL the browser will fetch the OpenAPI/Swagger spec from.
func WithSpecURL(url string) Option {
	return func(c *config) { c.specURL = url }
}

// WithPageTitle sets the HTML <title> of the page.
func WithPageTitle(title string) Option {
	return func(c *config) { c.pageTitle = title }
}

// WithDarkMode enables the dark-mode page chrome (body background + header).
// Swagger UI itself does not ship a first-party dark mode; this controls only
// the surrounding page styles rendered by this library.
func WithDarkMode() Option {
	return func(c *config) { c.darkMode = true }
}

// WithLightMode disables dark mode (the default page chrome becomes light).
func WithLightMode() Option {
	return func(c *config) { c.darkMode = false }
}

// Branding controls the optional branded header rendered above Swagger UI.
type Branding struct {
	LogoURL     string // src for the header logo image
	LogoAlt     string // alt text for the logo
	Title       string // company or product name
	Subtitle    string // secondary line under the title
	FaviconURL  string // href for the page favicon
	FaviconType string // MIME type for the favicon (default: "image/svg+xml")
}

// WithBranding sets a branded header above the Swagger UI.
func WithBranding(b Branding) Option {
	return func(c *config) { c.branding = &b }
}

// WithEnvBadge adds an environment label next to the brand title (e.g. "dev", "staging").
func WithEnvBadge(env string) Option {
	return func(c *config) { c.envBadge = env }
}

// WithJSPath overrides the URL for the swagger-ui-bundle.js script.
// Defaults to "/swagger-ui-bundle.js" (served from the embedded bundle).
// Point at a CDN or self-hosted URL to skip the embedded asset.
func WithJSPath(path string) Option {
	return func(c *config) { c.jsPath = path }
}

// WithPresetPath overrides the URL for the swagger-ui-standalone-preset.js script.
// Defaults to "/swagger-ui-standalone-preset.js" (served from the embedded bundle).
func WithPresetPath(path string) Option {
	return func(c *config) { c.presetPath = path }
}

// WithCSSPath overrides the URL for the swagger-ui.css stylesheet.
// Defaults to "/swagger-ui.css" (served from the embedded bundle).
// Point at a CDN or self-hosted URL to skip the embedded asset.
func WithCSSPath(path string) Option {
	return func(c *config) { c.cssPath = path }
}

// WithTemplate replaces the default HTML page template.
// The template receives a [PageData] value.
func WithTemplate(tmpl *template.Template) Option {
	return func(c *config) { c.tmpl = tmpl }
}

// DocExpansion controls how operations are expanded in the UI.
type DocExpansion string

const (
	DocExpansionList DocExpansion = "list" // each operation collapsed, but tag groups open
	DocExpansionFull DocExpansion = "full" // all operations and their details expanded
	DocExpansionNone DocExpansion = "none" // everything collapsed
)

// WithDocExpansion sets how the operations are expanded on load.
func WithDocExpansion(e DocExpansion) Option {
	return func(c *config) { c.uiConfig["docExpansion"] = e }
}

// WithDeepLinking enables/disables the deep-linking feature (enabled by default).
func WithDeepLinking(enabled bool) Option {
	return func(c *config) { c.uiConfig["deepLinking"] = enabled }
}

// WithFilter enables the filter bar above the operations list.
// Pass an empty string to show the bar with no pre-filled text,
// or a non-empty string to pre-fill a tag filter.
func WithFilter(filterText string) Option {
	return func(c *config) {
		if filterText == "" {
			c.uiConfig["filter"] = true
		} else {
			c.uiConfig["filter"] = filterText
		}
	}
}

// WithPersistAuthorization keeps authorisation data in the browser between
// page reloads (stored in localStorage).
func WithPersistAuthorization() Option {
	return func(c *config) { c.uiConfig["persistAuthorization"] = true }
}

// WithDisplayRequestDuration shows the request duration in the "Try it out" panel.
func WithDisplayRequestDuration() Option {
	return func(c *config) { c.uiConfig["displayRequestDuration"] = true }
}

// WithTryItOutEnabled expands the "Try it out" panel by default for all operations.
func WithTryItOutEnabled() Option {
	return func(c *config) { c.uiConfig["tryItOutEnabled"] = true }
}

// WithDefaultModelsExpandDepth sets the default expansion depth for models
// (-1 = hidden, 0 = collapsed, 1+ = expanded to that depth).
func WithDefaultModelsExpandDepth(depth int) Option {
	return func(c *config) { c.uiConfig["defaultModelsExpandDepth"] = depth }
}

// WithDefaultModelExpandDepth sets the default expansion depth for the model
// example section.
func WithDefaultModelExpandDepth(depth int) Option {
	return func(c *config) { c.uiConfig["defaultModelExpandDepth"] = depth }
}

// WithDisplayOperationID shows the operationId for every operation.
func WithDisplayOperationID() Option {
	return func(c *config) { c.uiConfig["displayOperationId"] = true }
}

// WithShowExtensions shows vendor-extension fields (x-*) on operations, models
// and parameters.
func WithShowExtensions() Option {
	return func(c *config) { c.uiConfig["showExtensions"] = true }
}

// WithShowCommonExtensions shows common vendor extensions (x-nullable,
// x-discriminator, x-internal, x-logo, x-order, x-required-properties).
func WithShowCommonExtensions() Option {
	return func(c *config) { c.uiConfig["showCommonExtensions"] = true }
}

// WithRequestInterceptor injects a raw JavaScript expression that will be used
// as the requestInterceptor function body. Use this sparingly; prefer proper
// auth configuration via Swagger UI security schemes instead.
//
//	swagger.WithRequestInterceptor("request.headers['X-My-Header'] = 'value'; return request;")
func WithRequestInterceptor(jsBody string) Option {
	return func(c *config) {
		c.uiConfig["requestInterceptor"] = template.JS("function(request){" + jsBody + "}")
	}
}

// WithOption sets any Swagger UI configuration key not covered by the typed
// options above. The value must be JSON-serialisable (or a [html/template.JS]
// for raw JS expressions).
//
//	swagger.WithOption("syntaxHighlight.theme", "monokai")
func WithOption(key string, value any) Option {
	return func(c *config) { c.uiConfig[key] = value }
}
