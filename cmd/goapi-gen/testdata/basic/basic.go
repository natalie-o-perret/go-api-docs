package basic

import (
	"net/http"

	"github.com/natalie-o-perret/go-api-docs/openapi"
)

type ID string
type Count uint64

type Result struct {
	Maybe *string `json:"maybe"`
}

var router = openapi.New(openapi.Info{Title: "Basic", Version: "1.0.0"})

func init() {
	openapi.GET[ID](router, "/id", func(*http.Request) (*ID, error) { return nil, nil })
	openapi.GET[Count](router, "/count", func(*http.Request) (*Count, error) { return nil, nil })
	openapi.GET[Result](router, "/result", func(*http.Request) (*Result, error) { return nil, nil })
}
