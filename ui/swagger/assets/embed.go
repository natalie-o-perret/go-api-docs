// Package assets holds the vendored Swagger UI distribution files.
// Update them with: make vendor-swagger-ui VERSION=5.x.y
package assets

import _ "embed"

// BundleJS is the swagger-ui-bundle standalone browser bundle.
//
//go:embed swagger-ui-bundle.min.js
var BundleJS []byte

// StandalonePresetJS is the swagger-ui-standalone-preset bundle.
//
//go:embed swagger-ui-standalone-preset.min.js
var StandalonePresetJS []byte

// CSS is the swagger-ui stylesheet.
//
//go:embed swagger-ui.min.css
var CSS []byte
