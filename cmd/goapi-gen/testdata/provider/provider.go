package provider

import (
	"net/http"

	"github.com/natalie-o-perret/go-api-docs/openapi"
)

type Result struct {
	Value string `json:"value"`
}

func (Result) OpenAPISchema() openapi.Schema {
	return openapi.Schema{Type: "string"}
}

var router = openapi.New(openapi.Info{Title: "Provider", Version: "1.0.0"})

func init() {
	openapi.GET[**Result](router, "/result", func(*http.Request) (***Result, error) { return nil, nil })
}
