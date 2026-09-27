package openapi

import (
	"reflect"
	"strconv"
	"strings"
	"time"
)

// SchemaProvider is an optional interface types can implement to provide their
// own OpenAPI schema instead of having one derived via reflection.
//
// This is the compile-time escape hatch: implement it on your type and the
// reflection path is bypassed entirely for that type.
//
//	func (Task) OpenAPISchema() openapi.Schema {
//	    return openapi.Schema{
//	        Type: "object",
//	        Properties: map[string]openapi.Schema{
//	            "id":    {Type: "string", ReadOnly: true},
//	            "title": {Type: "string"},
//	        },
//	        Required: []string{"id", "title"},
//	    }
//	}
type SchemaProvider interface {
	OpenAPISchema() Schema
}

var (
	timeType           = reflect.TypeOf(time.Time{})
	schemaProviderType = reflect.TypeOf((*SchemaProvider)(nil)).Elem()
)

// schemaForType derives a JSON Schema from a Go reflect.Type.
// Struct types are inlined; named struct types are also registered in the
// provided components map so they can be $ref'd.
func schemaForType(t reflect.Type, components map[string]Schema) Schema {
	// Unwrap pointer — mark nullable.
	nullable := false
	for t.Kind() == reflect.Ptr {
		nullable = true
		t = t.Elem()
	}

	// If the type (or its pointer) implements SchemaProvider, use that schema
	// directly — no reflection needed for opted-in types.
	if t.Implements(schemaProviderType) {
		s := reflect.Zero(t).Interface().(SchemaProvider).OpenAPISchema()
		if nullable {
			s = nullableSchema(s)
		}
		return s
	}
	if reflect.PointerTo(t).Implements(schemaProviderType) {
		s := reflect.New(t).Interface().(SchemaProvider).OpenAPISchema()
		if nullable {
			s = nullableSchema(s)
		}
		return s
	}

	s := deriveSchema(t, components)
	if nullable {
		s = nullableSchema(s)
	}
	return s
}

func nullableSchema(s Schema) Schema {
	if typ, ok := s.Type.(string); ok && typ != "" {
		s.Type = []string{typ, "null"}
		return s
	}
	return Schema{AnyOf: []Schema{s, {Type: "null"}}}
}

func deriveSchema(t reflect.Type, components map[string]Schema) Schema {
	if t == timeType {
		return Schema{Type: "string", Format: "date-time"}
	}

	switch t.Kind() {
	case reflect.String:
		return Schema{Type: "string"}
	case reflect.Bool:
		return Schema{Type: "boolean"}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32:
		return Schema{Type: "integer", Format: "int32"}
	case reflect.Int64:
		return Schema{Type: "integer", Format: "int64"}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return Schema{Type: "integer"}
	case reflect.Float32:
		return Schema{Type: "number", Format: "float"}
	case reflect.Float64:
		return Schema{Type: "number", Format: "double"}
	case reflect.Slice:
		if t.Elem().Kind() == reflect.Uint8 {
			// []byte → base64 string
			return Schema{Type: "string", Format: "byte"}
		}
		items := schemaForType(t.Elem(), components)
		return Schema{Type: "array", Items: &items}
	case reflect.Map:
		return Schema{Type: "object"}
	case reflect.Struct:
		return schemaForStruct(t, components)
	case reflect.Interface:
		return Schema{} // any
	default:
		return Schema{Type: "string"}
	}
}

// schemaForStruct builds an object schema from struct fields.
// Named types (t.Name() != "") are registered in components and returned as $ref.
func schemaForStruct(t reflect.Type, components map[string]Schema) Schema {
	name := t.Name()
	if name == "" {
		return buildObjectSchema(t, components)
	}

	if _, ok := components[name]; ok {
		return Schema{Ref: "#/components/schemas/" + name}
	}

	// Register a placeholder first to break recursive cycles.
	components[name] = Schema{}
	s := buildObjectSchema(t, components)
	components[name] = s
	return Schema{Ref: "#/components/schemas/" + name}
}

func buildObjectSchema(t reflect.Type, components map[string]Schema) Schema {
	props := map[string]Schema{}
	var required []string

	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if !f.IsExported() {
			continue
		}

		// Skip fields that carry parameter metadata — they're not body fields.
		if f.Tag.Get("path") != "" || f.Tag.Get("query") != "" || f.Tag.Get("header") != "" {
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
		if f.Tag.Get("readOnly") == "true" {
			s.ReadOnly = true
		}
		if f.Tag.Get("writeOnly") == "true" {
			s.WriteOnly = true
		}
		applyConstraintTags(&s, f.Tag)

		props[name] = s

		// A field is required unless it's a pointer, has omitempty, or is tagged required:"false".
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

// jsonFieldName returns the JSON key for a struct field.
func jsonFieldName(f reflect.StructField) string {
	tag := f.Tag.Get("json")
	if tag == "" {
		return f.Name
	}
	parts := strings.SplitN(tag, ",", 2)
	if parts[0] == "" {
		return f.Name
	}
	return parts[0]
}

// applyConstraintTags reads validation/format struct tags and sets the
// corresponding Schema fields.  Supported tags:
//
//	format:"uuid"        — override the schema format
//	pattern:"^[a-z]+$"  — regexp for strings
//	minLength:"1"        — minimum string length
//	maxLength:"255"      — maximum string length
//	min:"0"              — minimum numeric value
//	max:"100"            — maximum numeric value
//	minItems:"1"         — minimum array length
//	maxItems:"50"        — maximum array length
func applyConstraintTags(s *Schema, tag reflect.StructTag) {
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
