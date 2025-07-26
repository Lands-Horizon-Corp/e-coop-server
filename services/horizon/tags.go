package horizon

import (
	"fmt"
	"reflect"
	"strings"
	"time"
	"unicode"
)

// Helper: camelCase
func toCamelCase(s string) string {
	if s == "" {
		return s
	}
	runes := []rune(s)
	runes[0] = unicode.ToLower(runes[0])
	return string(runes)
}

// Format Go validation tags and description fields as TypeScript comments
func formatGoTagComment(field reflect.StructField) string {
	var comments []string
	desc := field.Tag.Get("description")
	if desc != "" {
		comments = append(comments, desc)
	}
	validate := field.Tag.Get("validate")
	if validate != "" {
		comments = append(comments, "Validation: "+validate)
	}
	enum := field.Tag.Get("enum")
	if enum != "" {
		comments = append(comments, "Enum: "+enum)
	}
	if len(comments) == 0 {
		return ""
	}
	return "/** " + strings.Join(comments, " | ") + " */"
}

// Converts Go type to TypeScript type string, managing circular/visited types, returns referenced types for top-level export.
func goTypeToTSType(field reflect.StructField, visited map[reflect.Type]bool, typeQueue map[reflect.Type]bool) string {
	t := field.Type
	switch t.Kind() {
	case reflect.String:
		// Prefer enum for union type, otherwise use oneof, otherwise string
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
		return goTypeToTSType(reflect.StructField{Type: t.Elem(), Tag: field.Tag}, visited, typeQueue) + " | null"
	case reflect.Slice, reflect.Array:
		return goTypeToTSType(reflect.StructField{Type: t.Elem(), Tag: field.Tag}, visited, typeQueue) + "[]"
	case reflect.Map:
		if t.Key().Kind() == reflect.String {
			return "{ [key: string]: " + goTypeToTSType(reflect.StructField{Type: t.Elem(), Tag: field.Tag}, visited, typeQueue) + " }"
		}
		return "any"
	case reflect.Struct:
		if t == reflect.TypeOf(time.Time{}) {
			return "string"
		}
		if visited[t] {
			return t.Name()
		}
		typeQueue[t] = true
		return t.Name()
	default:
		return "any"
	}
}

// Generates the TypeScript interface for a struct type, with Markdown comments
func structTypeToTSMarkdown(t reflect.Type, visited map[reflect.Type]bool, typeQueue map[reflect.Type]bool) string {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	visited[t] = true
	defer func() { visited[t] = false }()

	var fields []string
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if field.PkgPath != "" {
			continue
		}
		comment := formatGoTagComment(field)
		jsonTag := field.Tag.Get("json")
		name := strings.Split(jsonTag, ",")[0]
		if name == "" {
			name = field.Name
		}
		name = toCamelCase(name)
		tsType := goTypeToTSType(field, visited, typeQueue)
		line := ""
		if comment != "" {
			line += comment + "\n"
		}
		line += fmt.Sprintf("  %s: %s;", name, tsType)
		fields = append(fields, line)
	}
	return fmt.Sprintf("export interface %s {\n%s\n}", t.Name(), strings.Join(fields, "\n"))
}

// Recursively print all referenced types as top-level TypeScript interfaces, in Markdown code blocks
func printAllTSInterfacesMarkdown(rootType reflect.Type) {
	visited := map[reflect.Type]bool{}
	typeQueue := map[reflect.Type]bool{rootType: true}
	handled := map[reflect.Type]bool{}
	for len(typeQueue) > 0 {
		var keys []reflect.Type
		for k := range typeQueue {
			keys = append(keys, k)
		}
		typeQueue = map[reflect.Type]bool{}
		for _, t := range keys {
			if handled[t] {
				continue
			}
			fmt.Println("```typescript")
			fmt.Println(structTypeToTSMarkdown(t, visited, typeQueue))
			fmt.Println("```")
			handled[t] = true
		}
	}
}
func allTSInterfacesMarkdown(rootType reflect.Type) string {
	visited := map[reflect.Type]bool{}
	typeQueue := map[reflect.Type]bool{rootType: true}
	handled := map[reflect.Type]bool{}
	var sb strings.Builder
	for len(typeQueue) > 0 {
		var keys []reflect.Type
		for k := range typeQueue {
			keys = append(keys, k)
		}
		typeQueue = map[reflect.Type]bool{}
		for _, t := range keys {
			if handled[t] {
				continue
			}
			sb.WriteString("```typescript\n")
			sb.WriteString(structTypeToTSMarkdown(t, visited, typeQueue))
			sb.WriteString("\n```\n")
			handled[t] = true
		}
	}
	return sb.String()
}
func TagFormat(model any) string {
	if model == nil {
		return "None"
	}
	return allTSInterfacesMarkdown(reflect.TypeOf(model))
}

// === EXAMPLES ===

// type NestedChild struct {
// 	ID    int      `json:"id" description:"Child ID" validate:"min=1,max=9999"`
// 	Tags  []string `json:"tags" description:"Tag list"`
// 	Flags []bool   `json:"flags" description:"Flags"`
// }

// type NestedParent struct {
// 	Name        string        `json:"name" description:"Parent name" validate:"required,min=3,max=50"`
// 	CreatedAt   time.Time     `json:"created_at" description:"Creation date"`
// 	Description *string       `json:"description" description:"Optional description"`
// 	Children    []NestedChild `json:"children" description:"Children list"`
// 	Extras      []any         `json:"extras" description:"Extra stuff"`
// }

// type FullComplexObject struct {
// 	Title       string           `json:"title" description:"Object title" validate:"required,min=5,max=100"`
// 	Count       int              `json:"count" description:"Count" validate:"min=0,max=1000"`
// 	Price       float64          `json:"price" description:"Price" validate:"min=0"`
// 	Active      bool             `json:"active" description:"Is Active"`
// 	Metadata    map[string]any   `json:"metadata" description:"Metadata map"`
// 	Nested      NestedParent     `json:"nested" description:"Nested parent object"`
// 	MixedArray  []any            `json:"mixed_array" description:"Mixed array"`
// 	ObjectArray []map[string]any `json:"object_array" description:"Array of objects"`
// 	ExtraNested [][]NestedChild  `json:"extra_nested" description:"2D array of NestedChild"`
// }

// // Circular example
// type Friend struct {
// 	Name string `json:"name" description:"Friend name" enum:"best,close,acquaintance" validate:"oneof=best close acquaintance"`
// 	User *User  `json:"user" description:"User reference"`
// }
// type User struct {
// 	Name    string   `json:"name" description:"User name"`
// 	Friends []Friend `json:"friends" description:"List of friends"`
// }

// // Enum/oneof example
// type FruitBasket struct {
// 	Fruit string `json:"fruit" description:"Type of fruit" enum:"apple,banana,orange" validate:"oneof=apple banana orange"`
// 	Size  string `json:"size" description:"Basket size" validate:"oneof=small medium large"`
// }

// func main() {
// 	fmt.Println("# TypeScript Interfaces Generated from Go Structs\n")
// 	fmt.Println("## FullComplexObject Example")
// 	printAllTSInterfacesMarkdown(reflect.TypeOf(FullComplexObject{}))
// 	fmt.Println("\n## User and Friend (Circular Example)")
// 	printAllTSInterfacesMarkdown(reflect.TypeOf(User{}))
// 	fmt.Println("\n## FruitBasket (Enum/Oneof Example)")
// 	printAllTSInterfacesMarkdown(reflect.TypeOf(FruitBasket{}))
// }
