package openapi

// ── OpenAPI 3.1 document types ───────────────────────────────────────────────

// Document is the root OpenAPI 3.1 object.
type Document struct {
	Components Components           `json:"components,omitempty"`
	Paths      map[string]*PathItem `json:"paths"`
	Info       Info                 `json:"info"`
	OpenAPI    string               `json:"openapi"`
	Servers    []Server             `json:"servers,omitempty"`
	Tags       []Tag                `json:"tags,omitempty"`
}

// Tag adds a description to a tag used on operations.
type Tag struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

// Info holds general API metadata.
type Info struct {
	Title       string `json:"title"`
	Version     string `json:"version"`
	Description string `json:"description,omitempty"`
}

// Server describes a target host.
type Server struct {
	URL         string `json:"url"`
	Description string `json:"description,omitempty"`
}

// Components holds reusable schema definitions.
type Components struct {
	Schemas         map[string]Schema         `json:"schemas,omitempty"`
	SecuritySchemes map[string]SecurityScheme `json:"securitySchemes,omitempty"`
}

// SecurityScheme describes an auth mechanism.
type SecurityScheme struct {
	Type         string `json:"type"`
	Scheme       string `json:"scheme,omitempty"`
	BearerFormat string `json:"bearerFormat,omitempty"`
	In           string `json:"in,omitempty"`
	Name         string `json:"name,omitempty"`
}

// PathItem holds the operations for a single path.
type PathItem struct {
	Get    *Operation `json:"get,omitempty"`
	Post   *Operation `json:"post,omitempty"`
	Put    *Operation `json:"put,omitempty"`
	Patch  *Operation `json:"patch,omitempty"`
	Delete *Operation `json:"delete,omitempty"`
}

// Operation represents a single API operation on a path.
type Operation struct {
	RequestBody *RequestBody          `json:"requestBody,omitempty"`
	Responses   map[string]Response   `json:"responses"`
	OperationID string                `json:"operationId,omitempty"`
	Summary     string                `json:"summary,omitempty"`
	Description string                `json:"description,omitempty"`
	Tags        []string              `json:"tags,omitempty"`
	Security    []map[string][]string `json:"security,omitempty"`
	Parameters  []Parameter           `json:"parameters,omitempty"`
	Deprecated  bool                  `json:"deprecated,omitempty"`
}

// Parameter describes a path, query or header parameter.
type Parameter struct {
	Name        string `json:"name"`
	In          string `json:"in"`
	Description string `json:"description,omitempty"`
	Schema      Schema `json:"schema"`
	Required    bool   `json:"required"`
}

// RequestBody describes the request payload.
type RequestBody struct {
	Content  map[string]MediaType `json:"content"`
	Required bool                 `json:"required"`
}

// Response describes a single response.
type Response struct {
	Content     map[string]MediaType `json:"content,omitempty"`
	Description string               `json:"description"`
}

// MediaType holds the schema for a content type.
type MediaType struct {
	Schema Schema `json:"schema"`
}

// Schema is a JSON Schema / OpenAPI schema object.
type Schema struct {
	Example     any               `json:"example,omitempty"`
	MinLength   *int              `json:"minLength,omitempty"`
	MaxLength   *int              `json:"maxLength,omitempty"`
	MaxItems    *int              `json:"maxItems,omitempty"`
	MinItems    *int              `json:"minItems,omitempty"`
	Maximum     *float64          `json:"maximum,omitempty"`
	Items       *Schema           `json:"items,omitempty"`
	Properties  map[string]Schema `json:"properties,omitempty"`
	Minimum     *float64          `json:"minimum,omitempty"`
	Pattern     string            `json:"pattern,omitempty"`
	Ref         string            `json:"$ref,omitempty"`
	Format      string            `json:"format,omitempty"`
	Type        string            `json:"type,omitempty"`
	Description string            `json:"description,omitempty"`
	Required    []string          `json:"required,omitempty"`
	Enum        []any             `json:"enum,omitempty"`
	WriteOnly   bool              `json:"writeOnly,omitempty"`
	ReadOnly    bool              `json:"readOnly,omitempty"`
	Nullable    bool              `json:"nullable,omitempty"`
}
