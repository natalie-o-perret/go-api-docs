// Package assets holds the vendored Redoc standalone bundle.
// Update it with: make vendor-redoc VERSION=2.x.y
package assets

import _ "embed"

// RedocJS is the Redoc standalone browser bundle.
//
//go:embed redoc.standalone.min.js
var RedocJS []byte

