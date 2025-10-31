package handlers

import (
	"database/sql/driver"
	"fmt"
	"reflect"
	"strings"
	"unicode"

	"time"

	"github.com/google/uuid"
)

// toCamelCase converts a string to camelCase by lowercasing the first character.
func toCamelCase(s string) string {
	if s == "" {
		return s
	}
	runes := []rune(s)
	runes[0] = unicode.ToLower(runes[0])
	return string(runes)
}

// TypeScriptGenerator encapsulates the logic for converting Go structs to TypeScript interfaces.
type TypeScriptGenerator struct {
	typeQueue map[reflect.Type]bool // Queue of types to process
	handled   map[reflect.Type]bool // Tracks processed types
}

//nolint:modernize // reflect.TypeOf used intentionally for caching type values
var (
	uuidType         = reflect.TypeOf((*uuid.UUID)(nil)).Elem()
	timeType         = reflect.TypeOf((*time.Time)(nil)).Elem()
	driverValuerType = reflect.TypeOf((*driver.Valuer)(nil)).Elem()
	errorInterface   = reflect.TypeOf((*error)(nil)).Elem()
)

// NewTypeScriptGenerator initializes a new generator with empty maps.
func NewTypeScriptGenerator() *TypeScriptGenerator {
	return &TypeScriptGenerator{
		typeQueue: make(map[reflect.Type]bool),
		handled:   make(map[reflect.Type]bool),
	}
}

// FormatGoTagComment generates a TypeScript comment from struct field tags.
func (g *TypeScriptGenerator) FormatGoTagComment(field reflect.StructField) string {
	var comments []string
	if desc := field.Tag.Get("description"); desc != "" {
		comments = append(comments, desc)
	}
	if validate := field.Tag.Get("validate"); validate != "" {
		comments = append(comments, "Validation: "+validate)
	}
	if enum := field.Tag.Get("enum"); enum != "" {
		comments = append(comments, "Enum: "+enum)
	}
	if len(comments) == 0 {
		return ""
	}
	return "/** " + strings.Join(comments, " | ") + " */"
}

// StructTypeToTSInline generates an inline TypeScript type for anonymous structs.
func (g *TypeScriptGenerator) StructTypeToTSInline(t reflect.Type) string {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	var fields []string
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if field.PkgPath != "" { // Skip unexported fields
			continue
		}
		comment := g.FormatGoTagComment(field)
		jsonTag := field.Tag.Get("json")
		name := strings.Split(jsonTag, ",")[0]
		if name == "" {
			name = field.Name
		}
		name = toCamelCase(name)
		tsType := g.GoTypeToTSType(field)
		line := ""
		if comment != "" {
			line += comment + "\n"
		}
		line += fmt.Sprintf("%s: %s;", name, tsType)
		fields = append(fields, line)
	}
	return fmt.Sprintf("{\n%s\n}", strings.Join(fields, "\n"))
}

// GoTypeToTSType converts a Go field type to its TypeScript equivalent.
func (g *TypeScriptGenerator) GoTypeToTSType(field reflect.StructField) string {
	t := field.Type

	// Handle special types first
	if t == uuidType {
		return "uuid"
	}
	if t.Kind() == reflect.Ptr && t.Elem() == uuidType {
		return "uuid | null"
	}
	if t == timeType {
		return "string" // ISO date string format
	}
	if t.Kind() == reflect.Ptr && t.Elem() == timeType {
		return "string | null"
	}

	// Handle gorm.DeletedAt specifically
	if t.String() == "gorm.DeletedAt" {
		return "string | null"
	}

	// Handle json.RawMessage
	if t.String() == "json.RawMessage" {
		return "any"
	}

	// Handle sql null types
	switch t.String() {
	case "sql.NullString":
		return "string | null"
	case "sql.NullInt64", "sql.NullInt32", "sql.NullInt16":
		return "number | null"
	case "sql.NullFloat64":
		return "number | null"
	case "sql.NullBool":
		return "boolean | null"
	case "sql.NullTime":
		return "string | null"
	}

	// Handle driver.Valuer interface (database types)
	if t.Implements(driverValuerType) {
		return "any"
	}

	switch t.Kind() {
	case reflect.String:
		enum := field.Tag.Get("enum")
		val := field.Tag.Get("validate")
		var union []string

		if enum != "" {
			parts := strings.Split(enum, ",")
			for i := range parts {
				parts[i] = fmt.Sprintf("\"%s\"", strings.TrimSpace(parts[i]))
			}
			union = parts
		}

		if strings.Contains(val, "oneof=") && len(union) == 0 {
			oneof := strings.Split(val, "oneof=")[1]
			opts := strings.Fields(oneof)
			for i := range opts {
				opts[i] = fmt.Sprintf("\"%s\"", opts[i])
			}
			union = opts
		}

		if len(union) > 0 {
			return strings.Join(union, " | ")
		}
		return "string"

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return "number"

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return "number"

	case reflect.Float32, reflect.Float64:
		return "number"

	case reflect.Complex64, reflect.Complex128:
		return "{ real: number; imag: number }" // Complex numbers as objects

	case reflect.Bool:
		return "boolean"

	case reflect.Interface:
		// Check if it's error interface
		if t.Implements(errorInterface) {
			return "string" // Errors are typically serialized as strings
		}
		return "any"

	case reflect.Ptr:
		// Handle specific pointer types
		if t.Elem() == uuidType {
			return "uuid | null"
		}
		if t.Elem() == timeType {
			return "string | null"
		}
		return g.GoTypeToTSType(reflect.StructField{Type: t.Elem(), Tag: field.Tag}) + " | null"

	case reflect.Slice, reflect.Array:
		// Handle byte slices specifically (common for binary data)
		if t.Elem().Kind() == reflect.Uint8 {
			return "string" // Base64 encoded string or similar
		}
		return g.GoTypeToTSType(reflect.StructField{Type: t.Elem(), Tag: field.Tag}) + "[]"

	case reflect.Map:
		keyType := t.Key()
		valueType := t.Elem()

		switch keyType.Kind() {
		case reflect.String:
			return "{ [key: string]: " + g.GoTypeToTSType(reflect.StructField{Type: valueType, Tag: field.Tag}) + " }"
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return "{ [key: number]: " + g.GoTypeToTSType(reflect.StructField{Type: valueType, Tag: field.Tag}) + " }"
		default:
			return "Record<string, " + g.GoTypeToTSType(reflect.StructField{Type: valueType, Tag: field.Tag}) + ">"
		}

	case reflect.Struct:
		if t == uuidType {
			return "uuid"
		}
		if t == timeType {
			return "string"
		}

		// Handle embedded structs or anonymous structs
		if t.Name() == "" {
			return g.StructTypeToTSInline(t)
		}

		// Check for common struct types
		switch t.String() {
		case "time.Duration":
			return "number" // Duration as milliseconds or nanoseconds
		case "big.Int":
			return "string" // Big integers as strings
		case "big.Float":
			return "string" // Big floats as strings
		case "decimal.Decimal": // shopspring/decimal
			return "string"
		case "url.URL":
			return "string"
		}

		// Named struct: add to queue for processing
		// if !g.handled[t] {
		// 	g.typeQueue[t] = true
		// }
		return t.Name()

	case reflect.Chan:
		return "never" // Channels don't translate to TypeScript

	case reflect.Func:
		// Try to infer function signature if possible
		numIn := t.NumIn()
		numOut := t.NumOut()

		if numIn == 0 && numOut == 0 {
			return "() => void"
		}
		if numIn == 0 && numOut == 1 {
			return "() => any"
		}
		return "Function" // Generic function type

	case reflect.UnsafePointer:
		return "any" // Unsafe pointers are opaque

	case reflect.Uintptr:
		return "number"

	default:
		return "any"
	}
}

// StructTypeToTSMarkdown generates a TypeScript interface definition for a struct.
func (g *TypeScriptGenerator) StructTypeToTSMarkdown(t reflect.Type) string {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	var fields []string
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if field.PkgPath != "" { // Skip unexported fields
			continue
		}
		comment := g.FormatGoTagComment(field)
		jsonTag := field.Tag.Get("json")
		name := strings.Split(jsonTag, ",")[0]
		if name == "" {
			name = field.Name
		}
		name = toCamelCase(name)
		tsType := g.GoTypeToTSType(field)
		line := ""
		if comment != "" {
			line += comment + "\n"
		}
		line += fmt.Sprintf("  %s: %s;", name, tsType)
		fields = append(fields, line)
	}
	return fmt.Sprintf("export interface %s {\n%s\n}", t.Name(), strings.Join(fields, "\n"))
}

// GenerateAllInterfacesMarkdown generates Markdown with all TypeScript interfaces.
func (g *TypeScriptGenerator) GenerateAllInterfacesMarkdown(rootType reflect.Type) string {
	g.typeQueue[rootType] = true
	var sb strings.Builder
	for len(g.typeQueue) > 0 {
		var keys []reflect.Type
		for k := range g.typeQueue {
			keys = append(keys, k)
		}
		g.typeQueue = make(map[reflect.Type]bool) // Reset queue
		for _, t := range keys {
			if g.handled[t] {
				continue
			}
			sb.WriteString("```typescript\n")
			sb.WriteString(g.StructTypeToTSMarkdown(t))
			sb.WriteString("\n```\n")
			g.handled[t] = true
		}
	}
	return sb.String()
}

// TagFormat converts a Go model to TypeScript interfaces in Markdown.
func TagFormat(model any) string {
	if model == nil {
		return "None"
	}
	generator := NewTypeScriptGenerator()
	return generator.GenerateAllInterfacesMarkdown(reflect.TypeOf(model))
}
