// Package assets holds the vendored Scalar JS bundle.
// Update it with: make vendor-js VERSION=1.x.y
package assets

import _ "embed"

// ScalarJS is the @scalar/api-reference standalone browser bundle.
//
//go:embed scalar.min.js
var ScalarJS []byte
