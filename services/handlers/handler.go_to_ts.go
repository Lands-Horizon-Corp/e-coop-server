package handlers

import (
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
	// Always treat uuid.UUID as "uuid"
	if t == reflect.TypeOf(uuid.UUID{}) {
		return "uuid"
	}
	// Always treat *uuid.UUID as "uuid | null"
	if t.Kind() == reflect.Ptr && t.Elem() == reflect.TypeOf(uuid.UUID{}) {
		return "uuid | null"
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
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		return "number"
	case reflect.Bool:
		return "boolean"
	case reflect.Interface:
		return "any"
	case reflect.Ptr:
		if t.Elem() == reflect.TypeOf(uuid.UUID{}) {
			return "uuid | null"
		}
		if t.Elem() == reflect.TypeOf(time.Time{}) {
			return "Date | null"
		}
		return g.GoTypeToTSType(reflect.StructField{Type: t.Elem(), Tag: field.Tag}) + " | null"
	case reflect.Slice, reflect.Array:
		return g.GoTypeToTSType(reflect.StructField{Type: t.Elem(), Tag: field.Tag}) + "[]"
	case reflect.Map:
		if t.Key().Kind() == reflect.String {
			return "{ [key: string]: " + g.GoTypeToTSType(reflect.StructField{Type: t.Elem(), Tag: field.Tag}) + " }"
		}
		return "any"
	case reflect.Struct:
		if t == reflect.TypeOf(uuid.UUID{}) {
			return "uuid"
		}
		if t == reflect.TypeOf(time.Time{}) {
			return "Date"
		}
		if t.Name() == "" { // Anonymous struct
			return g.StructTypeToTSInline(t)
		}
		// Named struct: add to queue if not handled
		// if !g.handled[t] {
		// 	g.typeQueue[t] = true
		// }
		return t.Name()
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
