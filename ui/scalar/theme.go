package scalar

import (
	"strings"
)

// CustomTheme is a typed builder for Scalar CSS custom properties.
// Create one with [NewCustomTheme], chain the setter methods, then pass it to
// [WithCustomTheme].
//
//	scalar.New(
//	    scalar.WithCustomTheme(
//	        scalar.NewCustomTheme().
//	            Accent("#2563eb").
//	            Background("#0f172a").
//	            Text("#f1f5f9"),
//	    ),
//	)
//
// Only properties that have been explicitly set are emitted into the CSS.
// Unset properties fall back to the active [Theme]'s defaults.
type CustomTheme struct {
	vars map[string]string
}

// NewCustomTheme returns a new, empty CustomTheme ready for chaining.
func NewCustomTheme() *CustomTheme {
	return &CustomTheme{vars: make(map[string]string)}
}

func (t *CustomTheme) set(key, value string) *CustomTheme {
	t.vars[key] = value
	return t
}

// ── Text colours ─────────────────────────────────────────────────────────────

// Text sets the primary text colour (--scalar-color-1).
func (t *CustomTheme) Text(v string) *CustomTheme { return t.set("--scalar-color-1", v) }

// TextSecondary sets the secondary text colour (--scalar-color-2).
func (t *CustomTheme) TextSecondary(v string) *CustomTheme { return t.set("--scalar-color-2", v) }

// TextMuted sets the muted / tertiary text colour (--scalar-color-3).
func (t *CustomTheme) TextMuted(v string) *CustomTheme { return t.set("--scalar-color-3", v) }

// TextDisabled sets the disabled-state text colour (--scalar-color-disabled).
func (t *CustomTheme) TextDisabled(v string) *CustomTheme {
	return t.set("--scalar-color-disabled", v)
}

// TextGhost sets the ghost / placeholder text colour (--scalar-color-ghost).
func (t *CustomTheme) TextGhost(v string) *CustomTheme { return t.set("--scalar-color-ghost", v) }

// ── Accent ───────────────────────────────────────────────────────────────────

// Accent sets the accent / link / highlight colour (--scalar-color-accent).
func (t *CustomTheme) Accent(v string) *CustomTheme { return t.set("--scalar-color-accent", v) }

// AccentBackground sets the accent background colour (--scalar-background-accent).
func (t *CustomTheme) AccentBackground(v string) *CustomTheme {
	return t.set("--scalar-background-accent", v)
}

// ── Backgrounds ──────────────────────────────────────────────────────────────

// Background sets the primary page background (--scalar-background-1).
func (t *CustomTheme) Background(v string) *CustomTheme { return t.set("--scalar-background-1", v) }

// BackgroundSecondary sets the secondary surface background, e.g. the sidebar
// (--scalar-background-2).
func (t *CustomTheme) BackgroundSecondary(v string) *CustomTheme {
	return t.set("--scalar-background-2", v)
}

// BackgroundTertiary sets the tertiary surface background, e.g. inputs
// (--scalar-background-3).
func (t *CustomTheme) BackgroundTertiary(v string) *CustomTheme {
	return t.set("--scalar-background-3", v)
}

// ── Borders ──────────────────────────────────────────────────────────────────

// Border sets the default border colour (--scalar-border-color).
func (t *CustomTheme) Border(v string) *CustomTheme { return t.set("--scalar-border-color", v) }

// ── Status / badge colours ───────────────────────────────────────────────────

// ColorGreen sets the success / green badge colour (--scalar-color-green).
func (t *CustomTheme) ColorGreen(v string) *CustomTheme { return t.set("--scalar-color-green", v) }

// ColorRed sets the error / red badge colour (--scalar-color-red).
func (t *CustomTheme) ColorRed(v string) *CustomTheme { return t.set("--scalar-color-red", v) }

// ColorYellow sets the warning / yellow badge colour (--scalar-color-yellow).
func (t *CustomTheme) ColorYellow(v string) *CustomTheme { return t.set("--scalar-color-yellow", v) }

// ColorBlue sets the info / blue badge colour (--scalar-color-blue).
func (t *CustomTheme) ColorBlue(v string) *CustomTheme { return t.set("--scalar-color-blue", v) }

// ColorOrange sets the orange badge colour (--scalar-color-orange).
func (t *CustomTheme) ColorOrange(v string) *CustomTheme { return t.set("--scalar-color-orange", v) }

// ColorPurple sets the purple badge colour (--scalar-color-purple).
func (t *CustomTheme) ColorPurple(v string) *CustomTheme { return t.set("--scalar-color-purple", v) }

// ── Typography ───────────────────────────────────────────────────────────────

// Font sets the body font stack (--scalar-font).
func (t *CustomTheme) Font(v string) *CustomTheme { return t.set("--scalar-font", v) }

// FontCode sets the monospace font stack (--scalar-font-code).
func (t *CustomTheme) FontCode(v string) *CustomTheme { return t.set("--scalar-font-code", v) }

// ── Radius ───────────────────────────────────────────────────────────────────

// Radius sets the base border-radius (--scalar-radius).
func (t *CustomTheme) Radius(v string) *CustomTheme { return t.set("--scalar-radius", v) }

// RadiusLg sets the large border-radius (--scalar-radius-lg).
func (t *CustomTheme) RadiusLg(v string) *CustomTheme { return t.set("--scalar-radius-lg", v) }

// RadiusXl sets the extra-large border-radius (--scalar-radius-xl).
func (t *CustomTheme) RadiusXl(v string) *CustomTheme { return t.set("--scalar-radius-xl", v) }

// ── Sidebar ──────────────────────────────────────────────────────────────────

// SidebarBackground sets the sidebar background (--scalar-sidebar-background-1).
func (t *CustomTheme) SidebarBackground(v string) *CustomTheme {
	return t.set("--scalar-sidebar-background-1", v)
}

// SidebarText sets the primary sidebar text colour (--scalar-sidebar-color-1).
func (t *CustomTheme) SidebarText(v string) *CustomTheme { return t.set("--scalar-sidebar-color-1", v) }

// SidebarTextSecondary sets the secondary sidebar text colour (--scalar-sidebar-color-2).
func (t *CustomTheme) SidebarTextSecondary(v string) *CustomTheme {
	return t.set("--scalar-sidebar-color-2", v)
}

// SidebarAccent sets the sidebar accent colour (--scalar-sidebar-color-accent).
func (t *CustomTheme) SidebarAccent(v string) *CustomTheme {
	return t.set("--scalar-sidebar-color-accent", v)
}

// SidebarBorder sets the sidebar border colour (--scalar-sidebar-border-color).
func (t *CustomTheme) SidebarBorder(v string) *CustomTheme {
	return t.set("--scalar-sidebar-border-color", v)
}

// SidebarItemHoverBackground sets the sidebar item hover background
// (--scalar-sidebar-item-hover-background).
func (t *CustomTheme) SidebarItemHoverBackground(v string) *CustomTheme {
	return t.set("--scalar-sidebar-item-hover-background", v)
}

// SidebarItemHoverColor sets the sidebar item hover text colour
// (--scalar-sidebar-item-hover-color).
func (t *CustomTheme) SidebarItemHoverColor(v string) *CustomTheme {
	return t.set("--scalar-sidebar-item-hover-color", v)
}

// SidebarItemActiveBackground sets the sidebar active item background
// (--scalar-sidebar-item-active-background).
func (t *CustomTheme) SidebarItemActiveBackground(v string) *CustomTheme {
	return t.set("--scalar-sidebar-item-active-background", v)
}

// SidebarSearchBackground sets the sidebar search input background
// (--scalar-sidebar-search-background).
func (t *CustomTheme) SidebarSearchBackground(v string) *CustomTheme {
	return t.set("--scalar-sidebar-search-background", v)
}

// SidebarSearchBorder sets the sidebar search input border colour
// (--scalar-sidebar-search-border-color).
func (t *CustomTheme) SidebarSearchBorder(v string) *CustomTheme {
	return t.set("--scalar-sidebar-search-border-color", v)
}

// SidebarSearchColor sets the sidebar search input text colour
// (--scalar-sidebar-search-color).
func (t *CustomTheme) SidebarSearchColor(v string) *CustomTheme {
	return t.set("--scalar-sidebar-search-color", v)
}

// ── Raw escape-hatch ─────────────────────────────────────────────────────────

// Var sets an arbitrary CSS custom property. Use this for any Scalar variable
// not yet covered by a dedicated builder method.
//
//	theme.Var("--scalar-color-1", "#fff")
func (t *CustomTheme) Var(property, value string) *CustomTheme { return t.set(property, value) }

// ── CSS generation ───────────────────────────────────────────────────────────

// CSS renders the CustomTheme as a CSS snippet that can be embedded in a page
// or passed to [WithCustomCSS].  Returns an empty string when no properties are set.
func (t *CustomTheme) CSS() string {
	if len(t.vars) == 0 {
		return ""
	}
	var sb strings.Builder
	sb.WriteString(".light-mode, .dark-mode {\n")
	for k, v := range t.vars {
		sb.WriteString("  ")
		sb.WriteString(k)
		sb.WriteString(": ")
		sb.WriteString(v)
		sb.WriteString(";\n")
	}
	sb.WriteString("}")
	return sb.String()
}

// ── Option ───────────────────────────────────────────────────────────────────

// WithCustomTheme applies a [CustomTheme] as the Scalar customCss configuration.
// It can be combined with [WithCustomCSS] — call [WithCustomCSS] after
// [WithCustomTheme] to append extra rules, or before to prepend them.
func WithCustomTheme(theme *CustomTheme) Option {
	return func(c *config) {
		css := theme.CSS()
		if css == "" {
			return
		}
		existing, _ := c.scalarConfig["customCss"].(string)
		if existing != "" {
			css = existing + "\n" + css
		}
		c.scalarConfig["customCss"] = css
	}
}
