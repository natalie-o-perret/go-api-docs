SCALAR_VERSION ?= 1.52.6

.PHONY: vendor-js test lint

## vendor-js: download a specific @scalar/api-reference standalone bundle.
##   make vendor-js VERSION=1.53.0
vendor-js:
	curl -sL "https://cdn.jsdelivr.net/npm/@scalar/api-reference@$(SCALAR_VERSION)/dist/browser/standalone.js" \
	  -o assets/scalar.min.js
	@echo "vendored scalar v$(SCALAR_VERSION) → $$(wc -c < assets/scalar.min.js) bytes"

## test: run all tests
test:
	go test ./...

## lint: run golangci-lint
lint:
	golangci-lint run ./...

