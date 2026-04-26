// Package assets holds the vendored Stoplight Elements bundles.
// Update them with: make vendor-elements VERSION=8.x.y
package assets

import _ "embed"

// ElementsJS is the Stoplight Elements web-components browser bundle.
//
//go:embed elements.min.js
var ElementsJS []byte

// ElementsCSS is the Stoplight Elements stylesheet.
//
//go:embed elements.min.css
var ElementsCSS []byte

