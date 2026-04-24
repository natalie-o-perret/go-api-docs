// Package assets holds the vendored Scalar JS bundle embedded at compile time.
//
// To update the bundle, run:
//
//	make vendor-js VERSION=1.x.y
//
// or manually:
//
//	curl -sL https://cdn.jsdelivr.net/npm/@scalar/api-reference@<VERSION>/dist/browser/standalone.js \
//	  -o assets/scalar.min.js
package assets

import _ "embed"

// ScalarJS is the vendored @scalar/api-reference standalone browser bundle.
// It is served at the path configured by [goscalar.WithJSPath] (default: /scalar.js).
//
//go:embed scalar.min.js
var ScalarJS []byte

