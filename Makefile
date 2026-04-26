SCALAR_VERSION  ?= 1.52.6
SWAGGER_VERSION ?= 5.18.2

.PHONY: vendor-js vendor-swagger-ui gen-swagger-swag install-goapi-gen test lint

## install-goapi-gen: install the static spec generator into $GOPATH/bin.
##   go generate ./...   will then produce openapi.json with zero runtime.
install-goapi-gen:
	go install github.com/nopereta/go-api-docs/cmd/goapi-gen@latest
	@echo "goapi-gen installed → $$(which goapi-gen)"

## vendor-js: download a specific @scalar/api-reference standalone bundle.
##   make vendor-js VERSION=1.53.0
vendor-js:
	curl -sL "https://cdn.jsdelivr.net/npm/@scalar/api-reference@$(SCALAR_VERSION)/dist/browser/standalone.js" \
	  -o assets/scalar.min.js
	@echo "vendored scalar v$(SCALAR_VERSION) → $$(wc -c < assets/scalar.min.js) bytes"

## vendor-swagger-ui: download a specific swagger-ui-dist release.
##   make vendor-swagger-ui VERSION=5.19.0
vendor-swagger-ui:
	curl -sL "https://cdn.jsdelivr.net/npm/swagger-ui-dist@$(SWAGGER_VERSION)/swagger-ui-bundle.js" \
	  -o swagger/assets/swagger-ui-bundle.min.js
	curl -sL "https://cdn.jsdelivr.net/npm/swagger-ui-dist@$(SWAGGER_VERSION)/swagger-ui-standalone-preset.js" \
	  -o swagger/assets/swagger-ui-standalone-preset.min.js
	curl -sL "https://cdn.jsdelivr.net/npm/swagger-ui-dist@$(SWAGGER_VERSION)/swagger-ui.css" \
	  -o swagger/assets/swagger-ui.min.css
	@echo "vendored swagger-ui v$(SWAGGER_VERSION)"
	@echo "  bundle.js → $$(wc -c < swagger/assets/swagger-ui-bundle.min.js) bytes"
	@echo "  preset.js → $$(wc -c < swagger/assets/swagger-ui-standalone-preset.min.js) bytes"
	@echo "  ui.css    → $$(wc -c < swagger/assets/swagger-ui.min.css) bytes"

## gen-swagger-swag: regenerate the OpenAPI spec for example/swagger/swag from annotations.
gen-swagger-swag:
	swag init \
	  --generalInfo main.go \
	  --dir         example/swagger/swag \
	  --output      example/swagger/swag/docs
	@echo "spec written to example/swagger/swag/docs/"

## test: run all tests
test:
	go test ./...

## lint: run golangci-lint
lint:
	golangci-lint run ./...

