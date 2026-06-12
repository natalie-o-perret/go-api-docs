// goapi-gen statically analyses Go source files and emits an OpenAPI 3.1 JSON
// spec without running any code at all.
//
// It recognises every route-registration call from
// github.com/nopereta/go-api-docs/openapi (GET, GETWithInput, POST, PUT,
// PATCH, DELETE, Handle) and the router constructor (New) and extracts:
//   - the HTTP method and path
//   - In / Out type parameters → full JSON Schema (struct tags included)
//   - Summary, Description, Tags, OperationID, Security, Deprecated, Responses
//   - openapi.Info, WithServer, WithTag, WithSecurityScheme from New(…)
//
// Usage:
//
//	goapi-gen [flags] [packages]
//	goapi-gen .                          # current package
//	goapi-gen ./...                      # current module
//	goapi-gen -out openapi.json ./cmd/api
//
// Flags:
//
//	-out  string   output file path, use - for stdout (default "openapi.json")
//
// Integration — add to your main.go or a doc.go:
//
//	//go:generate goapi-gen -out ../../openapi.json .
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"go/ast"
	"go/constant"
	"go/types"
	"log"
	"os"
	"reflect"
	"strconv"
	"strings"

	"github.com/nopereta/go-api-docs/openapi"
	"golang.org/x/tools/go/packages"
)

const openapiPkgPath = "github.com/nopereta/go-api-docs/openapi"

func main() {
	out := flag.String("out", "openapi.json", `output file path; use "-" for stdout`)
	flag.Usage = func() {
		_, _ = fmt.Fprintf(os.Stderr, "Usage: goapi-gen [flags] [packages]\n\nFlags:\n")
		flag.PrintDefaults()
	}
	flag.Parse()
	patterns := flag.Args()
	if len(patterns) == 0 {
		patterns = []string{"."}
	}

	cfg := &packages.Config{
		Mode: packages.NeedName |
			packages.NeedSyntax |
			packages.NeedTypes |
			packages.NeedTypesInfo |
			packages.NeedImports,
	}
	pkgs, err := packages.Load(cfg, patterns...)
	if err != nil {
		log.Fatalf("load packages: %v", err)
	}
	for _, p := range pkgs {
		for _, e := range p.Errors {
			log.Printf("warning: %v", e)
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
	g.run(pkgs)

	b, err := json.MarshalIndent(g.doc, "", "  ")
	if err != nil {
		log.Fatalf("marshal: %v", err)
	}

	if *out == "-" {
		_, _ = os.Stdout.Write(b)
		fmt.Println()
		return
	}
	if err := os.WriteFile(*out, append(b, '\n'), 0o644); err != nil {
		log.Fatalf("write %s: %v", *out, err)
	}
	fmt.Printf("goapi-gen: wrote %s\n", *out)
}

// ── Generator ─────────────────────────────────────────────────────────────────

type generator struct {
	components map[string]openapi.Schema
	paths      map[string]*openapi.PathItem
	doc        openapi.Document
}

func (g *generator) run(pkgs []*packages.Package) {
	for _, pkg := range pkgs {
		for _, file := range pkg.Syntax {
			ast.Inspect(file, func(n ast.Node) bool {
				call, ok := n.(*ast.CallExpr)
				if !ok {
					return true
				}
				g.handleCall(pkg, call)
				return true
			})
		}
	}
	g.doc.Paths = g.paths
	g.doc.Components = openapi.Components{
		Schemas:         g.components,
		SecuritySchemes: g.doc.Components.SecuritySchemes,
	}
}

func (g *generator) handleCall(pkg *packages.Package, call *ast.CallExpr) {
	sel, typeArgs, ok := splitGenericCall(call)
	if !ok {
		return
	}
	// Resolve the function's package.
	obj := pkg.TypesInfo.Uses[sel.Sel]
	if obj == nil || obj.Pkg() == nil || obj.Pkg().Path() != openapiPkgPath {
		return
	}

	name := sel.Sel.Name

	if name == "New" {
		g.handleNew(pkg, call)
		return
	}

	method, pathIdx, inIdx, outIdx := routeSig(name)
	if method == "" {
		return
	}

	args := call.Args
	if len(args) <= pathIdx {
		return
	}

	path := constStr(pkg, args[pathIdx])
	if path == "" {
		return
	}

	// Handle[In,Out](r, method, path, fn, opts...) — method is args[1]
	if name == "Handle" {
		method = strings.ToUpper(constStr(pkg, args[1]))
	}

	// Resolve type arguments.
	var inType, outType types.Type
	if inIdx >= 0 && inIdx < len(typeArgs) {
		if tv, ok := pkg.TypesInfo.Types[typeArgs[inIdx]]; ok {
			inType = tv.Type
		}
	}
	if outIdx >= 0 && outIdx < len(typeArgs) {
		if tv, ok := pkg.TypesInfo.Types[typeArgs[outIdx]]; ok {
			outType = tv.Type
		}
	}

	// Route options start after (r, [method,] path, fn).
	optsStart := pathIdx + 2
	opts := g.extractRouteOpts(pkg, args[optsStart:])

	op := g.buildOp(method, inType, outType, opts)

	if g.paths[path] == nil {
		g.paths[path] = &openapi.PathItem{}
	}
	placeOp(g.paths[path], method, op)
}

// handleNew parses openapi.New(Info{…}, opts…) and populates doc metadata.
func (g *generator) handleNew(pkg *packages.Package, call *ast.CallExpr) {
	if len(call.Args) == 0 {
		return
	}
	if lit, ok := call.Args[0].(*ast.CompositeLit); ok {
		info := extractInfoLit(pkg, lit)
		if info.Title != "" {
			g.doc.Info.Title = info.Title
		}
		if info.Version != "" {
			g.doc.Info.Version = info.Version
		}
		if info.Description != "" {
			g.doc.Info.Description = info.Description
		}
	}
	for _, arg := range call.Args[1:] {
		optCall, ok := arg.(*ast.CallExpr)
		if !ok {
			continue
		}
		optSel, _, ok := splitGenericCall(optCall)
		if !ok {
			continue
		}
		optObj := pkg.TypesInfo.Uses[optSel.Sel]
		if optObj == nil || optObj.Pkg() == nil || optObj.Pkg().Path() != openapiPkgPath {
			continue
		}
		switch optSel.Sel.Name {
		case "WithServer":
			if len(optCall.Args) >= 2 {
				g.doc.Servers = append(g.doc.Servers, openapi.Server{
					URL:         constStr(pkg, optCall.Args[0]),
					Description: constStr(pkg, optCall.Args[1]),
				})
			}
		case "WithTag":
			if len(optCall.Args) >= 2 {
				g.doc.Tags = append(g.doc.Tags, openapi.Tag{
					Name:        constStr(pkg, optCall.Args[0]),
					Description: constStr(pkg, optCall.Args[1]),
				})
			}
		case "WithSecurityScheme":
			if len(optCall.Args) >= 2 {
				sname := constStr(pkg, optCall.Args[0])
				scheme := extractSecurityScheme(pkg, optCall.Args[1])
				if g.doc.Components.SecuritySchemes == nil {
					g.doc.Components.SecuritySchemes = map[string]openapi.SecurityScheme{}
				}
				g.doc.Components.SecuritySchemes[sname] = scheme
			}
		}
	}
}

// ── Route-option extraction ────────────────────────────────────────────────────

type routeOpts struct {
	extraResps map[string]openapi.Response
	summary    string
	desc       string
	opID       string
	tags       []string
	security   []string
	deprecated bool
}

func (g *generator) extractRouteOpts(pkg *packages.Package, args []ast.Expr) routeOpts {
	var o routeOpts
	for _, arg := range args {
		c, ok := arg.(*ast.CallExpr)
		if !ok {
			continue
		}
		sel, _, ok := splitGenericCall(c)
		if !ok {
			continue
		}
		obj := pkg.TypesInfo.Uses[sel.Sel]
		if obj == nil || obj.Pkg() == nil || obj.Pkg().Path() != openapiPkgPath {
			continue
		}
		switch sel.Sel.Name {
		case "Summary":
			if len(c.Args) >= 1 {
				o.summary = constStr(pkg, c.Args[0])
			}
		case "Description":
			if len(c.Args) >= 1 {
				o.desc = constStr(pkg, c.Args[0])
			}
		case "OperationID":
			if len(c.Args) >= 1 {
				o.opID = constStr(pkg, c.Args[0])
			}
		case "Tags":
			for _, a := range c.Args {
				if s := constStr(pkg, a); s != "" {
					o.tags = append(o.tags, s)
				}
			}
		case "Security":
			for _, a := range c.Args {
				if s := constStr(pkg, a); s != "" {
					o.security = append(o.security, s)
				}
			}
		case "Deprecated":
			o.deprecated = true
		case "Responses":
			if len(c.Args) >= 1 {
				o.extraResps = extractResponseMap(pkg, c.Args[0])
			}
		}
	}
	return o
}

// ── Operation builder ─────────────────────────────────────────────────────────

func (g *generator) buildOp(method string, inType, outType types.Type, opts routeOpts) *openapi.Operation {
	op := &openapi.Operation{
		OperationID: opts.opID,
		Summary:     opts.summary,
		Description: opts.desc,
		Tags:        opts.tags,
		Deprecated:  opts.deprecated,
		Responses:   map[string]openapi.Response{},
	}
	if len(opts.security) > 0 {
		req := map[string][]string{}
		for _, s := range opts.security {
			req[s] = []string{}
		}
		op.Security = []map[string][]string{req}
	}

	// Parameters (path / query / header) from In struct fields.
	if inType != nil {
		if st := underlyingStruct(inType); st != nil {
			for i := range st.NumFields() {
				f := st.Field(i)
				if !f.Exported() {
					continue
				}
				tag := reflect.StructTag(st.Tag(i))
				in, pname := paramLocFromTag(tag)
				if in == "" {
					continue
				}
				p := openapi.Parameter{
					Name:        pname,
					In:          in,
					Required:    in == "path" || tag.Get("required") == "true",
					Description: tag.Get("doc"),
					Schema:      g.schemaFor(f.Type()),
				}
				if enum := tag.Get("enum"); enum != "" {
					for _, v := range strings.Split(enum, ",") {
						p.Schema.Enum = append(p.Schema.Enum, strings.TrimSpace(v))
					}
				}
				applyTagConstraints(&p.Schema, tag)
				op.Parameters = append(op.Parameters, p)
			}
		}
	}

	// Request body from body fields in In (POST / PUT / PATCH only).
	if inType != nil && isBodyMethod(method) {
		body := g.bodySchema(inType)
		if len(body.Properties) > 0 {
			op.RequestBody = &openapi.RequestBody{
				Required: true,
				Content:  map[string]openapi.MediaType{"application/json": {Schema: body}},
			}
		}
	}

	// Success response.
	statusCode := "200"
	if strings.EqualFold(method, "POST") {
		statusCode = "201"
	}
	isDelete := strings.EqualFold(method, "DELETE")
	noOut := outType == nil || isEmptyStructType(outType)

	if noOut || isDelete {
		op.Responses["204"] = openapi.Response{Description: "No content"}
		if isDelete {
			for k, v := range opts.extraResps {
				op.Responses[k] = v
			}
			return op
		}
	} else {
		op.Responses[statusCode] = openapi.Response{
			Description: "Success",
			Content:     map[string]openapi.MediaType{"application/json": {Schema: g.schemaFor(outType)}},
		}
	}

	// Generic error envelope — ensure ErrorBody is in components.
	g.ensureErrorBody()
	op.Responses["default"] = openapi.Response{
		Description: "Error",
		Content:     map[string]openapi.MediaType{"application/json": {Schema: openapi.Schema{Ref: "#/components/schemas/ErrorBody"}}},
	}

	for k, v := range opts.extraResps {
		op.Responses[k] = v
	}
	return op
}

func (g *generator) ensureErrorBody() {
	if _, ok := g.components["ErrorBody"]; ok {
		return
	}
	g.components["ErrorBody"] = openapi.Schema{
		Type: "object",
		Properties: map[string]openapi.Schema{
			"code":    {Type: "string", Example: "not_found"},
			"message": {Type: "string", Example: "resource not found"},
		},
		Required: []string{"code", "message"},
	}
}

// ── Schema from go/types ──────────────────────────────────────────────────────

func (g *generator) schemaFor(t types.Type) openapi.Schema {
	nullable := false
	for {
		p, ok := t.(*types.Pointer)
		if !ok {
			break
		}
		nullable = true
		t = p.Elem()
	}
	s := g.derive(t)
	if nullable {
		s.Nullable = true
	}
	return s
}

func (g *generator) derive(t types.Type) openapi.Schema {
	// Named type: time.Time or a named struct ($ref).
	if named, ok := t.(*types.Named); ok {
		obj := named.Obj()
		if obj.Pkg() != nil && obj.Pkg().Path() == "time" && obj.Name() == "Time" {
			return openapi.Schema{Type: "string", Format: "date-time"}
		}
		n := obj.Name()
		if _, exists := g.components[n]; !exists {
			g.components[n] = openapi.Schema{} // placeholder against cycles
			if st, ok := named.Underlying().(*types.Struct); ok {
				g.components[n] = g.objectSchema(st)
			}
		}
		return openapi.Schema{Ref: "#/components/schemas/" + n}
	}

	switch typ := t.(type) {
	case *types.Basic:
		return basicSchema(typ)
	case *types.Slice:
		if isBytes(typ) {
			return openapi.Schema{Type: "string", Format: "byte"}
		}
		items := g.schemaFor(typ.Elem())
		return openapi.Schema{Type: "array", Items: &items}
	case *types.Map:
		return openapi.Schema{Type: "object"}
	case *types.Struct:
		return g.objectSchema(typ)
	case *types.Interface:
		return openapi.Schema{} // any
	default:
		return openapi.Schema{Type: "string"}
	}
}

// objectSchema builds a full object schema from a *types.Struct (all fields).
func (g *generator) objectSchema(st *types.Struct) openapi.Schema {
	props := map[string]openapi.Schema{}
	var required []string
	for i := range st.NumFields() {
		f := st.Field(i)
		if !f.Exported() {
			continue
		}
		tag := reflect.StructTag(st.Tag(i))
		if tag.Get("path") != "" || tag.Get("query") != "" || tag.Get("header") != "" {
			continue
		}
		fname := jsonFieldName(f.Name(), tag)
		if fname == "-" {
			continue
		}
		s := g.schemaFor(f.Type())
		applyFieldTags(&s, tag)
		props[fname] = s

		_, isPtr := f.Type().(*types.Pointer)
		omit := strings.Contains(tag.Get("json"), "omitempty")
		if !isPtr && !omit && tag.Get("required") != "false" {
			required = append(required, fname)
		}
	}
	s := openapi.Schema{Type: "object", Properties: props}
	if len(required) > 0 {
		s.Required = required
	}
	return s
}

// bodySchema builds a schema from only the body fields (no path/query/header tags).
func (g *generator) bodySchema(t types.Type) openapi.Schema {
	for {
		if p, ok := t.(*types.Pointer); ok {
			t = p.Elem()
			continue
		}
		break
	}
	st := underlyingStruct(t)
	if st == nil {
		return g.schemaFor(t)
	}
	props := map[string]openapi.Schema{}
	var required []string
	for i := range st.NumFields() {
		f := st.Field(i)
		if !f.Exported() {
			continue
		}
		tag := reflect.StructTag(st.Tag(i))
		if tag.Get("path") != "" || tag.Get("query") != "" || tag.Get("header") != "" {
			continue
		}
		fname := jsonFieldName(f.Name(), tag)
		if fname == "-" {
			continue
		}
		s := g.schemaFor(f.Type())
		applyFieldTags(&s, tag)
		props[fname] = s

		_, isPtr := f.Type().(*types.Pointer)
		omit := strings.Contains(tag.Get("json"), "omitempty")
		if !isPtr && !omit && tag.Get("required") != "false" {
			required = append(required, fname)
		}
	}
	s := openapi.Schema{Type: "object", Properties: props}
	if len(required) > 0 {
		s.Required = required
	}
	return s
}

// ── Helpers ───────────────────────────────────────────────────────────────────

// splitGenericCall breaks a CallExpr's Fun into (selector, typeArgs).
// Handles three forms:
//
//	pkg.Func(args)           → *SelectorExpr,      nil
//	pkg.Func[T](args)        → *IndexExpr.X,       [T]
//	pkg.Func[T1,T2](args)    → *IndexListExpr.X,   [T1,T2]
func splitGenericCall(call *ast.CallExpr) (*ast.SelectorExpr, []ast.Expr, bool) {
	switch fn := call.Fun.(type) {
	case *ast.SelectorExpr:
		return fn, nil, true
	case *ast.IndexExpr:
		sel, ok := fn.X.(*ast.SelectorExpr)
		if !ok {
			return nil, nil, false
		}
		return sel, []ast.Expr{fn.Index}, true
	case *ast.IndexListExpr:
		sel, ok := fn.X.(*ast.SelectorExpr)
		if !ok {
			return nil, nil, false
		}
		return sel, fn.Indices, true
	}
	return nil, nil, false
}

// routeSig returns the HTTP method, index of the path arg, and indices (in
// typeArgs slice) for In and Out. -1 means "not present / use struct{}".
func routeSig(fn string) (method string, pathIdx, inIdx, outIdx int) {
	switch fn {
	case "GET":
		return "GET", 1, -1, 0
	case "GETWithInput":
		return "GET", 1, 0, 1
	case "POST":
		return "POST", 1, 0, 1
	case "PUT":
		return "PUT", 1, 0, 1
	case "PATCH":
		return "PATCH", 1, 0, 1
	case "DELETE":
		return "DELETE", 1, 0, -1
	case "Handle":
		// args: r, method, path, fn, opts...
		return "HANDLE", 2, 0, 1
	}
	return "", 0, 0, 0
}

func constStr(pkg *packages.Package, expr ast.Expr) string {
	tv, ok := pkg.TypesInfo.Types[expr]
	if !ok || tv.Value == nil || tv.Value.Kind() != constant.String {
		return ""
	}
	s, _ := strconv.Unquote(tv.Value.String())
	return s
}

func extractInfoLit(pkg *packages.Package, lit *ast.CompositeLit) openapi.Info {
	var info openapi.Info
	for _, elt := range lit.Elts {
		kv, ok := elt.(*ast.KeyValueExpr)
		if !ok {
			continue
		}
		key, ok := kv.Key.(*ast.Ident)
		if !ok {
			continue
		}
		v := constStr(pkg, kv.Value)
		switch key.Name {
		case "Title":
			info.Title = v
		case "Version":
			info.Version = v
		case "Description":
			info.Description = v
		}
	}
	return info
}

func extractSecurityScheme(pkg *packages.Package, expr ast.Expr) openapi.SecurityScheme {
	switch e := expr.(type) {
	case *ast.SelectorExpr:
		switch e.Sel.Name {
		case "BearerAuth":
			return openapi.SecurityScheme{Type: "http", Scheme: "bearer", BearerFormat: "JWT"}
		case "BasicAuth":
			return openapi.SecurityScheme{Type: "http", Scheme: "basic"}
		}
	case *ast.CallExpr:
		sel, _, ok := splitGenericCall(e)
		if !ok {
			break
		}
		switch sel.Sel.Name {
		case "APIKeyHeader":
			if len(e.Args) >= 1 {
				return openapi.SecurityScheme{Type: "apiKey", In: "header", Name: constStr(pkg, e.Args[0])}
			}
		case "APIKeyQuery":
			if len(e.Args) >= 1 {
				return openapi.SecurityScheme{Type: "apiKey", In: "query", Name: constStr(pkg, e.Args[0])}
			}
		}
	}
	return openapi.SecurityScheme{}
}

func extractResponseMap(pkg *packages.Package, expr ast.Expr) map[string]openapi.Response {
	lit, ok := expr.(*ast.CompositeLit)
	if !ok {
		return nil
	}
	result := map[string]openapi.Response{}
	for _, elt := range lit.Elts {
		kv, ok := elt.(*ast.KeyValueExpr)
		if !ok {
			continue
		}
		code := constStr(pkg, kv.Key)
		if code == "" {
			continue
		}
		if rl, ok := kv.Value.(*ast.CompositeLit); ok {
			var resp openapi.Response
			for _, field := range rl.Elts {
				fkv, ok := field.(*ast.KeyValueExpr)
				if !ok {
					continue
				}
				if id, ok := fkv.Key.(*ast.Ident); ok && id.Name == "Description" {
					resp.Description = constStr(pkg, fkv.Value)
				}
			}
			result[code] = resp
		}
	}
	return result
}

func placeOp(item *openapi.PathItem, method string, op *openapi.Operation) {
	switch strings.ToUpper(method) {
	case "GET":
		item.Get = op
	case "POST":
		item.Post = op
	case "PUT":
		item.Put = op
	case "PATCH":
		item.Patch = op
	case "DELETE":
		item.Delete = op
	}
}

func underlyingStruct(t types.Type) *types.Struct {
	for {
		switch t2 := t.(type) {
		case *types.Named:
			t = t2.Underlying()
		case *types.Pointer:
			t = t2.Elem()
		case *types.Struct:
			return t2
		default:
			return nil
		}
	}
}

func isEmptyStructType(t types.Type) bool {
	for {
		switch t2 := t.(type) {
		case *types.Pointer:
			t = t2.Elem()
		case *types.Named:
			t = t2.Underlying()
		case *types.Struct:
			return t2.NumFields() == 0
		default:
			return false
		}
	}
}

func isBodyMethod(method string) bool {
	m := strings.ToUpper(method)
	return m == "POST" || m == "PUT" || m == "PATCH"
}

func isBytes(s *types.Slice) bool {
	b, ok := s.Elem().(*types.Basic)
	return ok && b.Kind() == types.Byte
}

func basicSchema(t *types.Basic) openapi.Schema {
	switch t.Kind() {
	case types.Bool:
		return openapi.Schema{Type: "boolean"}
	case types.String:
		return openapi.Schema{Type: "string"}
	case types.Int, types.Int8, types.Int16, types.Int32,
		types.Uint, types.Uint8, types.Uint16, types.Uint32:
		return openapi.Schema{Type: "integer", Format: "int32"}
	case types.Int64, types.Uint64:
		return openapi.Schema{Type: "integer", Format: "int64"}
	case types.Float32:
		return openapi.Schema{Type: "number", Format: "float"}
	case types.Float64:
		return openapi.Schema{Type: "number", Format: "double"}
	default:
		return openapi.Schema{Type: "string"}
	}
}

func jsonFieldName(fieldName string, tag reflect.StructTag) string {
	t := tag.Get("json")
	if t == "" {
		return lcFirst(fieldName)
	}
	parts := strings.SplitN(t, ",", 2)
	if parts[0] == "" {
		return lcFirst(fieldName)
	}
	return parts[0]
}

func lcFirst(s string) string {
	if s == "" {
		return s
	}
	return strings.ToLower(s[:1]) + s[1:]
}

func paramLocFromTag(tag reflect.StructTag) (in, name string) {
	if v := tag.Get("path"); v != "" {
		return "path", v
	}
	if v := tag.Get("query"); v != "" {
		return "query", v
	}
	if v := tag.Get("header"); v != "" {
		return "header", v
	}
	return "", ""
}

// applyFieldTags copies doc/example/enum/readOnly/writeOnly/constraint tags
// onto a Schema — identical logic to the runtime applyConstraintTags.
func applyFieldTags(s *openapi.Schema, tag reflect.StructTag) {
	if v := tag.Get("doc"); v != "" {
		s.Description = v
	}
	if v := tag.Get("example"); v != "" {
		s.Example = v
	}
	if v := tag.Get("enum"); v != "" {
		s.Enum = nil
		for _, part := range strings.Split(v, ",") {
			s.Enum = append(s.Enum, strings.TrimSpace(part))
		}
	}
	if tag.Get("readOnly") == "true" {
		s.ReadOnly = true
	}
	if tag.Get("writeOnly") == "true" {
		s.WriteOnly = true
	}
	applyTagConstraints(s, tag)
}

func applyTagConstraints(s *openapi.Schema, tag reflect.StructTag) {
	if v := tag.Get("format"); v != "" {
		s.Format = v
	}
	if v := tag.Get("pattern"); v != "" {
		s.Pattern = v
	}
	if v := tag.Get("minLength"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			s.MinLength = &n
		}
	}
	if v := tag.Get("maxLength"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			s.MaxLength = &n
		}
	}
	if v := tag.Get("min"); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			s.Minimum = &f
		}
	}
	if v := tag.Get("max"); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			s.Maximum = &f
		}
	}
	if v := tag.Get("minItems"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			s.MinItems = &n
		}
	}
	if v := tag.Get("maxItems"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			s.MaxItems = &n
		}
	}
}

// (spec types are imported from github.com/nopereta/go-api-docs/openapi)
