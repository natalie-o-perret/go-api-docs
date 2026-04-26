package openapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strings"
)

// Router is an http.ServeMux wrapper that also builds an OpenAPI 3.1 spec
// from the routes registered through it.
type Router struct {
	mux        *http.ServeMux
	doc        Document
	components map[string]Schema
	pathValue  PathValueFn // how to extract path params (defaults to r.PathValue)
}

// PathValueFn extracts a named URL path parameter from a request.
// Swap this to integrate with any router that carries path params differently.
//
//	// net/http 1.22+ (default):
//	func(r *http.Request, name string) string { return r.PathValue(name) }
//
//	// chi:
//	func(r *http.Request, name string) string { return chi.URLParam(r, name) }
//
//	// gorilla/mux:
//	func(r *http.Request, name string) string { return mux.Vars(r)[name] }
type PathValueFn func(r *http.Request, name string) string

// stdPathValue is the default extractor for net/http 1.22+.
func stdPathValue(r *http.Request, name string) string { return r.PathValue(name) }

// New creates a Router with the given API info.
func New(info Info, opts ...RouterOption) *Router {
	r := &Router{
		mux: http.NewServeMux(),
		doc: Document{
			OpenAPI: "3.1.0",
			Info:    info,
			Paths:   map[string]*PathItem{},
		},
		components: map[string]Schema{},
		pathValue:  stdPathValue,
	}
	for _, o := range opts {
		o(r)
	}
	// Serve the spec itself at /openapi.json.
	r.mux.HandleFunc("GET /openapi.json", r.serveSpec)
	return r
}

// ServeHTTP implements http.Handler.
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.mux.ServeHTTP(w, req)
}

// OpenAPI returns the generated spec as JSON bytes.
func (r *Router) OpenAPI() []byte {
	// Preserve SecuritySchemes set via WithSecurityScheme — read before overwrite.
	r.doc.Components = Components{
		Schemas:         r.components,
		SecuritySchemes: r.doc.Components.SecuritySchemes,
	}
	b, _ := json.MarshalIndent(r.doc, "", "  ")
	return b
}

func (r *Router) serveSpec(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(r.OpenAPI())
}

// pathItem returns (or creates) the PathItem for the given path.
func (r *Router) pathItem(path string) *PathItem {
	if r.doc.Paths[path] == nil {
		r.doc.Paths[path] = &PathItem{}
	}
	return r.doc.Paths[path]
}

// setOperation sets the operation on the correct method slot.
func setOperation(item *PathItem, method string, op *Operation) {
	switch strings.ToUpper(method) {
	case http.MethodGet:
		item.Get = op
	case http.MethodPost:
		item.Post = op
	case http.MethodPut:
		item.Put = op
	case http.MethodPatch:
		item.Patch = op
	case http.MethodDelete:
		item.Delete = op
	}
}

// ── Interfaces ────────────────────────────────────────────────────────────────

// Validator is an optional interface input structs can implement to perform
// custom validation after the standard decoding step.
//
// If the input type implements Validate(), it is called automatically after
// all path/query/header params and the JSON body have been decoded.
// Return a non-nil error to reject the request with a 400 Bad Request.
//
//	func (in *CreateTaskInput) Validate() error {
//	    if in.Title == "" {
//	        return errors.New("title is required")
//	    }
//	    return nil
//	}
type Validator interface {
	Validate() error
}

// ── Route registration ────────────────────────────────────────────────────────

// RouteOption customises a registered route's spec entry.
type RouteOption func(*Operation)

// Summary sets the operation summary.
func Summary(s string) RouteOption { return func(o *Operation) { o.Summary = s } }

// Description sets the operation description.
func Description(s string) RouteOption { return func(o *Operation) { o.Description = s } }

// Tags attaches tag names to the operation.
func Tags(tags ...string) RouteOption { return func(o *Operation) { o.Tags = tags } }

// OperationID sets a custom operationId.
func OperationID(id string) RouteOption { return func(o *Operation) { o.OperationID = id } }

// Deprecated marks the operation as deprecated.
func Deprecated() RouteOption { return func(o *Operation) { o.Deprecated = true } }

// Security adds a security requirement (e.g. Security("BearerAuth")).
func Security(schemes ...string) RouteOption {
	return func(o *Operation) {
		req := map[string][]string{}
		for _, s := range schemes {
			req[s] = []string{}
		}
		o.Security = append(o.Security, req)
	}
}

// Responses merges extra status-code responses into the operation's spec entry.
// Use this to document error codes beyond the auto-generated success response.
//
//	openapi.Responses(map[string]openapi.Response{
//	    "404": {Description: "Task not found"},
//	    "422": {Description: "Validation error"},
//	})
func Responses(extra map[string]Response) RouteOption {
	return func(o *Operation) {
		for code, resp := range extra {
			o.Responses[code] = resp
		}
	}
}

// RouterOption tweaks the Router itself.
type RouterOption func(*Router)

// WithPathValueFn sets a custom path-parameter extractor.
// Use this to integrate with routers other than the standard net/http ServeMux.
// See [PathValueFn] for examples.
func WithPathValueFn(fn PathValueFn) RouterOption {
	return func(r *Router) { r.pathValue = fn }
}

// WithServer adds a server entry to the spec.
func WithServer(url, description string) RouterOption {
	return func(r *Router) {
		r.doc.Servers = append(r.doc.Servers, Server{URL: url, Description: description})
	}
}

// WithSecurityScheme adds a security scheme to the spec components.
func WithSecurityScheme(name string, scheme SecurityScheme) RouterOption {
	return func(r *Router) {
		if r.doc.Components.SecuritySchemes == nil {
			r.doc.Components.SecuritySchemes = map[string]SecurityScheme{}
		}
		r.doc.Components.SecuritySchemes[name] = scheme
	}
}

// WithTag adds a tag with a description to the root spec.
// Tags appear in the Scalar/Swagger UI sidebar and group related operations.
//
//	openapi.WithTag("tasks", "CRUD operations for the task resource")
func WithTag(name, description string) RouterOption {
	return func(r *Router) {
		r.doc.Tags = append(r.doc.Tags, Tag{Name: name, Description: description})
	}
}

// BearerAuth is a pre-built HTTP Bearer security scheme.
var BearerAuth = SecurityScheme{Type: "http", Scheme: "bearer", BearerFormat: "JWT"}

// BasicAuth is a pre-built HTTP Basic security scheme.
var BasicAuth = SecurityScheme{Type: "http", Scheme: "basic"}

// APIKeyHeader returns a pre-built API-key-in-header security scheme.
//
//	openapi.WithSecurityScheme("X-API-Key", openapi.APIKeyHeader("X-API-Key"))
func APIKeyHeader(name string) SecurityScheme {
	return SecurityScheme{Type: "apiKey", In: "header", Name: name}
}

// APIKeyQuery returns a pre-built API-key-in-query security scheme.
func APIKeyQuery(name string) SecurityScheme {
	return SecurityScheme{Type: "apiKey", In: "query", Name: name}
}

// ── Typed registration ────────────────────────────────────────────────────────

// Handle registers a typed handler for the given method and path.
//
// In is a struct whose fields drive both parameter extraction and request body
// decoding:
//
//   - `path:"name"`   → extracted from the URL path segment (required)
//   - `query:"name"`  → extracted from the query string
//   - `header:"name"` → extracted from the request header
//   - Remaining exported fields → decoded from JSON request body (for POST/PUT/PATCH)
//
// Out is JSON-encoded and written as the success response (200, or 201 for POST).
//
// Use struct{} for In when there is no input, and struct{} for Out when the
// handler returns no body (204).
func Handle[In, Out any](r *Router, method, path string, fn func(*http.Request, *In) (*Out, error), opts ...RouteOption) {
	op := buildOperation[In, Out](method, r.components, opts...)
	setOperation(r.pathItem(path), method, op)

	r.mux.HandleFunc(method+" "+path, func(w http.ResponseWriter, req *http.Request) {
		var in In
		if err := decodeInput(req, &in, r.pathValue); err != nil {
			writeErr(w, http.StatusBadRequest, "invalid_input", err.Error())
			return
		}

		out, err := fn(req, &in)
		if err != nil {
			var he *HTTPError
			if errors.As(err, &he) {
				writeErr(w, he.Status, he.Code, he.Message)
				return
			}
			writeErr(w, http.StatusInternalServerError, "internal_error", err.Error())
			return
		}

		// 204 when Out is struct{} or out is nil.
		if out == nil || isEmptyStruct(reflect.TypeOf(out).Elem()) {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		status := http.StatusOK
		if strings.ToUpper(method) == http.MethodPost {
			status = http.StatusCreated
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_ = json.NewEncoder(w).Encode(out)
	})
}

// ── Convenience wrappers ──────────────────────────────────────────────────────

// GET registers a GET handler. Use Handle[In, Out] directly when the handler
// needs path/query parameters via an input struct.
func GET[Out any](r *Router, path string, fn func(*http.Request) (*Out, error), opts ...RouteOption) {
	Handle[struct{}, Out](r, http.MethodGet, path, func(req *http.Request, _ *struct{}) (*Out, error) {
		return fn(req)
	}, opts...)
}

// GETWithInput registers a GET handler that receives path/query params via In.
func GETWithInput[In, Out any](r *Router, path string, fn func(*http.Request, *In) (*Out, error), opts ...RouteOption) {
	Handle[In, Out](r, http.MethodGet, path, fn, opts...)
}

// POST registers a POST handler.
func POST[In, Out any](r *Router, path string, fn func(*http.Request, *In) (*Out, error), opts ...RouteOption) {
	Handle[In, Out](r, http.MethodPost, path, fn, opts...)
}

// PUT registers a PUT handler.
func PUT[In, Out any](r *Router, path string, fn func(*http.Request, *In) (*Out, error), opts ...RouteOption) {
	Handle[In, Out](r, http.MethodPut, path, fn, opts...)
}

// PATCH registers a PATCH handler.
func PATCH[In, Out any](r *Router, path string, fn func(*http.Request, *In) (*Out, error), opts ...RouteOption) {
	Handle[In, Out](r, http.MethodPatch, path, fn, opts...)
}

// DELETE registers a DELETE handler with no request body.
func DELETE[In any](r *Router, path string, fn func(*http.Request, *In) error, opts ...RouteOption) {
	Handle[In, struct{}](r, http.MethodDelete, path, func(req *http.Request, in *In) (*struct{}, error) {
		return nil, fn(req, in)
	}, opts...)
}

// ── Operation builder ─────────────────────────────────────────────────────────

func buildOperation[In, Out any](method string, components map[string]Schema, opts ...RouteOption) *Operation {
	op := &Operation{
		Responses: map[string]Response{},
	}
	for _, o := range opts {
		o(op)
	}

	inType := reflect.TypeOf((*In)(nil)).Elem()
	outType := reflect.TypeOf((*Out)(nil)).Elem()

	// ── parameters from In struct tags ────────────────────────────────────────
	if inType.Kind() == reflect.Struct {
		for i := 0; i < inType.NumField(); i++ {
			f := inType.Field(i)
			if !f.IsExported() {
				continue
			}
			paramIn, paramName := paramLocation(f)
			if paramIn == "" {
				continue
			}
			p := Parameter{
				Name:        paramName,
				In:          paramIn,
				Required:    paramIn == "path" || f.Tag.Get("required") == "true",
				Description: f.Tag.Get("doc"),
				Schema:      schemaForType(f.Type, components),
			}
			if enum := f.Tag.Get("enum"); enum != "" {
				var enums []any
				for _, v := range strings.Split(enum, ",") {
					enums = append(enums, strings.TrimSpace(v))
				}
				p.Schema.Enum = enums
			}
			applyConstraintTags(&p.Schema, f.Tag)
			op.Parameters = append(op.Parameters, p)
		}
	}

	// ── request body from remaining In fields (for mutating methods) ─────────
	if hasBodyFields(inType) && isBodyMethod(method) {
		bodySchema := bodySchemaForType(inType, components)
		op.RequestBody = &RequestBody{
			Required: true,
			Content: map[string]MediaType{
				"application/json": {Schema: bodySchema},
			},
		}
	}

	// ── success response ──────────────────────────────────────────────────────
	statusCode := "200"
	if strings.ToUpper(method) == http.MethodPost {
		statusCode = "201"
	}
	if isEmptyStruct(outType) || strings.ToUpper(method) == http.MethodDelete {
		op.Responses["204"] = Response{Description: "No content"}
		if strings.ToUpper(method) == http.MethodDelete {
			return op
		}
	} else {
		outSchema := schemaForType(outType, components)
		op.Responses[statusCode] = Response{
			Description: "Success",
			Content: map[string]MediaType{
				"application/json": {Schema: outSchema},
			},
		}
	}

	// Always include a generic error response.
	op.Responses["default"] = Response{
		Description: "Error",
		Content: map[string]MediaType{
			"application/json": {Schema: schemaForType(reflect.TypeOf(ErrorBody{}), components)},
		},
	}

	return op
}

// ── Input decoding ────────────────────────────────────────────────────────────

func decodeInput(r *http.Request, dst any, pathValue PathValueFn) error {
	v := reflect.ValueOf(dst).Elem()
	t := v.Type()
	if t.Kind() != reflect.Struct {
		return nil
	}

	hasBody := false
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if !f.IsExported() {
			continue
		}
		paramIn, paramName := paramLocation(f)
		switch paramIn {
		case "path":
			raw := pathValue(r, paramName)
			if err := setField(v.Field(i), raw); err != nil {
				return fmt.Errorf("path param %q: %w", paramName, err)
			}
		case "query":
			raw := r.URL.Query().Get(paramName)
			if raw != "" {
				if err := setField(v.Field(i), raw); err != nil {
					return fmt.Errorf("query param %q: %w", paramName, err)
				}
			}
		case "header":
			raw := r.Header.Get(paramName)
			if raw != "" {
				if err := setField(v.Field(i), raw); err != nil {
					return fmt.Errorf("header %q: %w", paramName, err)
				}
			}
		default:
			hasBody = true
		}
	}

	if hasBody && r.Body != nil && r.ContentLength != 0 {
		// Decode only the body-tagged fields by building a temporary map.
		var raw map[string]json.RawMessage
		if err := json.NewDecoder(r.Body).Decode(&raw); err != nil {
			return fmt.Errorf("decode body: %w", err)
		}
		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)
			if !f.IsExported() {
				continue
			}
			paramIn, _ := paramLocation(f)
			if paramIn != "" {
				continue // skip params
			}
			key := jsonFieldName(f)
			data, ok := raw[key]
			if !ok {
				continue
			}
			fv := v.Field(i)
			dest := reflect.New(f.Type)
			if err := json.Unmarshal(data, dest.Interface()); err != nil {
				return fmt.Errorf("body field %q: %w", key, err)
			}
			fv.Set(dest.Elem())
		}
	}

	// If the input type implements Validator, run custom validation after decode.
	if val, ok := dst.(Validator); ok {
		if err := val.Validate(); err != nil {
			return err
		}
	}
	return nil
}

// setField sets a reflect.Value from a string (for path/query/header params).
func setField(v reflect.Value, s string) error {
	t := v.Type()
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
		nv := reflect.New(t)
		v.Set(nv)
		v = nv.Elem()
	}
	switch t.Kind() {
	case reflect.String:
		v.SetString(s)
	case reflect.Bool:
		v.SetBool(s == "true" || s == "1")
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		var n int64
		if _, err := fmt.Sscanf(s, "%d", &n); err != nil {
			return err
		}
		v.SetInt(n)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		var n uint64
		if _, err := fmt.Sscanf(s, "%d", &n); err != nil {
			return err
		}
		v.SetUint(n)
	case reflect.Float32, reflect.Float64:
		var f float64
		if _, err := fmt.Sscanf(s, "%f", &f); err != nil {
			return err
		}
		v.SetFloat(f)
	default:
		return fmt.Errorf("unsupported field type %s", t)
	}
	return nil
}

// ── Helpers ───────────────────────────────────────────────────────────────────

// paramLocation returns ("path"|"query"|"header", name) from a struct field's tags,
// or ("", "") if the field is a body field.
func paramLocation(f reflect.StructField) (in, name string) {
	if v := f.Tag.Get("path"); v != "" {
		return "path", v
	}
	if v := f.Tag.Get("query"); v != "" {
		return "query", v
	}
	if v := f.Tag.Get("header"); v != "" {
		return "header", v
	}
	return "", ""
}

// hasBodyFields reports whether the struct has any non-param exported fields.
func hasBodyFields(t reflect.Type) bool {
	if t.Kind() != reflect.Struct {
		return false
	}
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if !f.IsExported() {
			continue
		}
		paramIn, _ := paramLocation(f)
		if paramIn == "" && jsonFieldName(f) != "-" {
			return true
		}
	}
	return false
}

// bodySchemaForType builds a schema from only the body fields of a struct.
func bodySchemaForType(t reflect.Type, components map[string]Schema) Schema {
	if t.Kind() != reflect.Struct {
		return schemaForType(t, components)
	}
	// Build an anonymous object from only the non-param fields.
	props := map[string]Schema{}
	var required []string
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if !f.IsExported() {
			continue
		}
		paramIn, _ := paramLocation(f)
		if paramIn != "" {
			continue
		}
		name := jsonFieldName(f)
		if name == "-" {
			continue
		}
		s := schemaForType(f.Type, components)
		if doc := f.Tag.Get("doc"); doc != "" {
			s.Description = doc
		}
		if ex := f.Tag.Get("example"); ex != "" {
			s.Example = ex
		}
		if enum := f.Tag.Get("enum"); enum != "" {
			for _, v := range strings.Split(enum, ",") {
				s.Enum = append(s.Enum, strings.TrimSpace(v))
			}
		}
		applyConstraintTags(&s, f.Tag)
		props[name] = s
		omitempty := strings.Contains(f.Tag.Get("json"), "omitempty")
		if f.Type.Kind() != reflect.Ptr && !omitempty && f.Tag.Get("required") != "false" {
			required = append(required, name)
		}
	}
	s := Schema{Type: "object", Properties: props}
	if len(required) > 0 {
		s.Required = required
	}
	return s
}

func isBodyMethod(method string) bool {
	m := strings.ToUpper(method)
	return m == http.MethodPost || m == http.MethodPut || m == http.MethodPatch
}

func isEmptyStruct(t reflect.Type) bool {
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t.Kind() == reflect.Struct && t.NumField() == 0
}

// ── Error helpers ─────────────────────────────────────────────────────────────

// ErrorBody is the standard error response body.
type ErrorBody struct {
	Code    string `json:"code"    example:"not_found"`
	Message string `json:"message" example:"resource not found"`
}

// HTTPError is a sentinel error carrying an HTTP status.
type HTTPError struct {
	Status  int
	Code    string
	Message string
}

func (e *HTTPError) Error() string { return e.Message }

// Err returns an *HTTPError that the router translates to the given status code.
func Err(status int, code, message string) error {
	return &HTTPError{Status: status, Code: code, Message: message}
}

// ErrNotFound returns a 404 HTTPError.
func ErrNotFound(msg string) error { return Err(http.StatusNotFound, "not_found", msg) }

// ErrBadRequest returns a 400 HTTPError.
func ErrBadRequest(msg string) error { return Err(http.StatusBadRequest, "bad_request", msg) }

// ErrUnauthorized returns a 401 HTTPError.
func ErrUnauthorized(msg string) error { return Err(http.StatusUnauthorized, "unauthorized", msg) }

func writeErr(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(ErrorBody{Code: code, Message: message})
}
