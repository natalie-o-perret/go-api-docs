package scalar

import "html/template"

// Option is a function that tweaks the Handler config.
type Option func(*config)

// WithSpecURL sets a single spec URL (Scalar fetches it in the browser).
// Mutually exclusive with WithSources — the last one called wins.
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

// WithSources sets multiple specs in the dropdown.
// Mutually exclusive with WithSpecURL — the last one called wins.
func WithSources(sources ...Source) Option {
	return func(c *config) {
		delete(c.scalarConfig, "url")
		c.sources = sources
	}
}

// Theme is one of Scalar's built-in theme names.
type Theme string

// Theme values.
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

// WithTheme sets the colour theme.
func WithTheme(t Theme) Option {
	return func(c *config) { c.scalarConfig["theme"] = t }
}

// Layout is one of Scalar's layout names.
type Layout string

// Layout values.
const (
	LayoutModern  Layout = "modern"
	LayoutClassic Layout = "classic"
)

// WithLayout sets the sidebar layout.
func WithLayout(l Layout) Option {
	return func(c *config) { c.scalarConfig["layout"] = l }
}

// WithDarkMode turns dark mode on.
func WithDarkMode() Option {
	return func(c *config) { c.scalarConfig["darkMode"] = true }
}

// WithLightMode turns dark mode off.
func WithLightMode() Option {
	return func(c *config) { c.scalarConfig["darkMode"] = false }
}

// WithCustomCSS injects custom CSS via the Scalar config.
// Multiple calls accumulate: each new block is appended after the previous one.
func WithCustomCSS(css string) Option {
	return func(c *config) {
		existing, _ := c.scalarConfig["customCss"].(string)
		if existing != "" {
			css = existing + "\n" + css
		}
		c.scalarConfig["customCss"] = css
	}
}

// Visibility controls a tri-state show/hide value used by several Scalar options.
// The ShowToolbar* constants are aliases for convenience.
type Visibility = ShowToolbar

// ShowToolbar controls when the developer toolbar is visible.
type ShowToolbar string

// ShowToolbar values.
const (
	ShowToolbarAlways    ShowToolbar = "always"
	ShowToolbarLocalhost ShowToolbar = "localhost"
	ShowToolbarNever     ShowToolbar = "never"
)

// WithShowToolbar sets toolbar visibility.
func WithShowToolbar(v ShowToolbar) Option {
	return func(c *config) { c.scalarConfig["showToolbar"] = v }
}

// WithHideClientButton hides the client button.
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

// WithShowDeveloperTools sets developer tools visibility.
func WithShowDeveloperTools(v ShowToolbar) Option {
	return func(c *config) { c.scalarConfig["showDeveloperTools"] = v }
}

// WithProxy sets a proxy URL for Scalar's try-it feature.
func WithProxy(url string) Option {
	return func(c *config) { c.scalarConfig["proxy"] = url }
}

// WithBaseServerURL overrides the server URL used in the "Try it" panel.
func WithBaseServerURL(url string) Option {
	return func(c *config) {
		c.scalarConfig["servers"] = []map[string]any{{"url": url}}
	}
}

// WithHideModels hides the Models section.
func WithHideModels() Option {
	return func(c *config) { c.scalarConfig["hideModels"] = true }
}

// WithHideDownloadButton hides the download spec button.
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

// Flag is a bitset of boolean Scalar configuration toggles.
// Combine multiple flags with | or pass them as separate arguments to [With].
type Flag uint

// Flag values for use with [With].
const (
	DisableAgent       Flag = 1 << iota // agent.disabled = true
	DisableMCP                          // mcp.disabled = true
	HideClientButton                    // hideClientButton = true
	HideModels                          // hideModels = true
	HideDownloadButton                  // hideDownloadButton = true
	HideSearch                          // hideSearch = true
	ShowOperationID                     // showOperationId = true
	DarkMode                            // darkMode = true
	LightMode                           // darkMode = false (overrides DarkMode if both set)
)

// With applies one or more boolean feature flags in a single option.
// Flags may be combined with | or passed as separate arguments — equivalent either way.
//
//	scalar.With(scalar.DisableAgent, scalar.DisableMCP, scalar.HideClientButton)
//	scalar.With(scalar.DisableAgent | scalar.DisableMCP | scalar.HideClientButton)
func With(flags ...Flag) Option {
	return func(c *config) {
		var combined Flag
		for _, f := range flags {
			combined |= f
		}
		if combined&DisableAgent != 0 {
			c.scalarConfig["agent"] = map[string]any{"disabled": true}
		}
		if combined&DisableMCP != 0 {
			c.scalarConfig["mcp"] = map[string]any{"disabled": true}
		}
		if combined&HideClientButton != 0 {
			c.scalarConfig["hideClientButton"] = true
		}
		if combined&HideModels != 0 {
			c.scalarConfig["hideModels"] = true
		}
		if combined&HideDownloadButton != 0 {
			c.scalarConfig["hideDownloadButton"] = true
		}
		if combined&HideSearch != 0 {
			c.scalarConfig["hideSearch"] = true
		}
		if combined&ShowOperationID != 0 {
			c.scalarConfig["showOperationId"] = true
		}
		if combined&DarkMode != 0 {
			c.scalarConfig["darkMode"] = true
		}
		// LightMode wins over DarkMode when both are set.
		if combined&LightMode != 0 {
			c.scalarConfig["darkMode"] = false
		}
	}
}

// WithOption sets any Scalar config key not covered by the typed options.
func WithOption(key string, value any) Option {
	return func(c *config) { c.scalarConfig[key] = value }
}

// WithPageTitle sets the HTML page title.
func WithPageTitle(title string) Option {
	return func(c *config) { c.pageTitle = title }
}

// Branding controls the optional header rendered above the Scalar UI.
type Branding struct {
	LogoURL     string // src for the header logo image
	LogoAlt     string // alt text for the logo
	Title       string // company or product name
	Subtitle    string // secondary line under the title
	FaviconURL  string // href for the page favicon
	FaviconType string // MIME type for the favicon (default: "image/svg+xml")
}

// WithBranding sets the branded header.
func WithBranding(b Branding) Option {
	return func(c *config) { c.branding = &b }
}

// WithEnvBadge adds an environment label next to the brand title (e.g. "dev", "preprod").
func WithEnvBadge(env string) Option {
	return func(c *config) { c.envBadge = env }
}

// WithJSPath sets the URL path where the JS bundle is served (default: "/scalar.js").
// Point it at a CDN or any URL you like; in that case the embedded bundle is not served.
func WithJSPath(path string) Option {
	return func(c *config) { c.jsPath = path }
}

// WithTemplate replaces the default HTML page template.
func WithTemplate(tmpl *template.Template) Option {
	return func(c *config) { c.tmpl = tmpl }
}
