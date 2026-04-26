package swagger

import "html/template"

// defaultTemplate returns the built-in Swagger UI page template.
func defaultTemplate() *template.Template {
	return template.Must(template.New("swagger").Parse(defaultPageHTML))
}

// defaultPageHTML is the built-in Swagger UI page template, receives PageData.
const defaultPageHTML = `<!DOCTYPE html>
<html>
<head>
  <title>{{.PageTitle}}</title>
  <meta charset="utf-8"/>
  <meta name="viewport" content="width=device-width, initial-scale=1"/>
  {{- if .FaviconLink}}
  {{.FaviconLink}}
  {{- end}}
  <link rel="stylesheet" href="{{.CSSPath}}"/>
  <style>
    *, *::before, *::after { box-sizing: border-box; }
    body {
      margin: 0;
      font-family: Inter, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif;
      {{- if .DarkMode}}
      background: #09090c;
      color: #f3f3f4;
      {{- else}}
      background: #ffffff;
      color: #09090c;
      {{- end}}
    }

    /* override swagger-ui top bar to match our chrome */
    .swagger-ui .topbar { display: none; }

    /* branded header */
    .gs-header {
      display: flex;
      align-items: center;
      gap: 14px;
      padding: 14px 24px;
      {{- if .DarkMode}}
      background: #0b0b10;
      border-bottom: 1px solid rgba(255,255,255,.08);
      {{- else}}
      background: #f6f6f7;
      border-bottom: 1px solid rgba(0,0,0,.08);
      {{- end}}
    }
    .gs-brand {
      display: flex;
      align-items: center;
      gap: 12px;
      margin-right: auto;
    }
    .gs-brand img {
      width: 32px;
      height: 32px;
      flex-shrink: 0;
    }
    .gs-brand-title {
      font-size: .9rem;
      font-weight: 700;
      letter-spacing: .12em;
      text-transform: uppercase;
      {{- if .DarkMode}}
      color: #fff;
      {{- else}}
      color: #09090c;
      {{- end}}
    }
    .gs-brand-subtitle {
      font-size: .8rem;
      color: #8b8b99;
      margin-top: 1px;
    }
    .gs-env-badge {
      font-size: .68rem;
      font-weight: 600;
      letter-spacing: .08em;
      text-transform: uppercase;
      padding: 3px 9px;
      border-radius: 4px;
      background: rgba(218,41,28,.18);
      color: #ff7060;
      border: 1px solid rgba(218,41,28,.35);
    }

    {{- if .DarkMode}}
    /* minimal dark-mode overrides for swagger-ui */
    .swagger-ui,
    .swagger-ui .wrapper,
    .swagger-ui .opblock-tag,
    .swagger-ui section.models {
      background: #09090c;
    }
    .swagger-ui .opblock-tag,
    .swagger-ui .info .title,
    .swagger-ui .info p,
    .swagger-ui .info li,
    .swagger-ui table thead tr th,
    .swagger-ui .response-col_status,
    .swagger-ui .parameter__name,
    .swagger-ui label,
    .swagger-ui select,
    .swagger-ui .model-title,
    .swagger-ui .model,
    .swagger-ui .prop-type {
      color: #e0e0e8 !important;
    }
    .swagger-ui .opblock .opblock-summary-description,
    .swagger-ui .opblock .opblock-summary-path {
      color: #c0c0cc !important;
    }
    .swagger-ui .opblock,
    .swagger-ui .model-box,
    .swagger-ui section.models .model-container {
      background: #13131a !important;
      border-color: rgba(255,255,255,.08) !important;
    }
    .swagger-ui input[type=text],
    .swagger-ui textarea,
    .swagger-ui select {
      background: #1e1e2a !important;
      border-color: rgba(255,255,255,.15) !important;
      color: #e0e0e8 !important;
    }
    {{- end}}
  </style>
</head>
<body>
{{- if .Branding}}
<header class="gs-header">
  <div class="gs-brand">
    {{- if .Branding.LogoURL}}
    <img src="{{.Branding.LogoURL}}" alt="{{.Branding.LogoAlt}}"/>
    {{- end}}
    {{- if or .Branding.Title .Branding.Subtitle}}
    <div>
      {{- if .Branding.Title}}<div class="gs-brand-title">{{.Branding.Title}}</div>{{end}}
      {{- if .Branding.Subtitle}}<div class="gs-brand-subtitle">{{.Branding.Subtitle}}</div>{{end}}
    </div>
    {{- end}}
  </div>
  {{- if .EnvBadge}}<span class="gs-env-badge">{{.EnvBadge}}</span>{{end}}
</header>
{{- end}}
<div id="swagger-ui"></div>
<script src="{{.JSPath}}"></script>
<script src="{{.PresetPath}}"></script>
<script>
  window.onload = function() {
    SwaggerUIBundle({{.UIConfig}});
  };
</script>
</body>
</html>
`
