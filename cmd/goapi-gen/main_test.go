package main

import (
	"encoding/json"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/natalie-o-perret/go-api-docs/openapi"
	"golang.org/x/tools/go/packages"
)

func TestGenerator_namedScalarsAndNullableTypes(t *testing.T) {
	g, err := generateFixture(t, "basic")
	if err != nil {
		t.Fatal(err)
	}

	idSchema := g.paths["/id"].Get.Responses["200"].Content["application/json"].Schema
	if idSchema.Type != "string" || idSchema.Ref != "" {
		t.Fatalf("named string schema = %#v", idSchema)
	}

	maybe := g.components["Result"].Properties["maybe"]
	if !reflect.DeepEqual(maybe.Type, []string{"string", "null"}) {
		t.Fatalf("nullable string type = %#v", maybe.Type)
	}
	countSchema := g.paths["/count"].Get.Responses["200"].Content["application/json"].Schema
	if countSchema.Type != "integer" || countSchema.Format != "" {
		t.Fatalf("named unsigned schema = %#v", countSchema)
	}
	raw, err := json.Marshal(g.doc)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(raw), `"nullable"`) {
		t.Fatalf("OpenAPI 3.1 output contains nullable: %s", raw)
	}
}

func TestGenerator_rejectsSchemaProvider(t *testing.T) {
	_, err := generateFixture(t, "provider")
	if err == nil || !strings.Contains(err.Error(), "cannot evaluate without running code") {
		t.Fatalf("expected SchemaProvider error, got %v", err)
	}
}

func TestGenerator_rejectsMultipleRouters(t *testing.T) {
	_, err := generateFixture(t, "multiple")
	if err == nil || !strings.Contains(err.Error(), "multiple openapi.New calls") {
		t.Fatalf("expected multiple-router error, got %v", err)
	}
}

func generateFixture(t *testing.T, name string) (*generator, error) {
	t.Helper()
	dir, err := filepath.Abs(filepath.Join("testdata", name))
	if err != nil {
		t.Fatal(err)
	}
	pkgs, err := packages.Load(&packages.Config{
		Dir: dir,
		Mode: packages.NeedName |
			packages.NeedSyntax |
			packages.NeedTypes |
			packages.NeedTypesInfo |
			packages.NeedImports,
	}, ".")
	if err != nil {
		t.Fatal(err)
	}
	for _, pkg := range pkgs {
		if len(pkg.Errors) > 0 {
			t.Fatalf("load fixture: %v", pkg.Errors)
		}
	}
	g := &generator{
		components: map[string]openapi.Schema{},
		paths:      map[string]*openapi.PathItem{},
		doc: openapi.Document{
			OpenAPI: "3.1.0",
			Info:    openapi.Info{Title: "API", Version: "0.0.0"},
		},
	}
	err = g.run(pkgs)
	return g, err
}
