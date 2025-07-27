package horizon

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
	"time"
	"unicode"

	"github.com/google/uuid"
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

func structTypeToTSInline(t reflect.Type, visited map[reflect.Type]bool, typeQueue map[reflect.Type]bool) string {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
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
		line += fmt.Sprintf("%s: %s;", name, tsType)
		fields = append(fields, line)
	}
	return fmt.Sprintf("{\n%s\n}", strings.Join(fields, "\n"))
}

func isNamedStruct(t reflect.Type) bool {
	return t.Kind() == reflect.Struct && t.Name() != ""
}

// PATCH: Support uuid.UUID as "uuid" and *uuid.UUID as "uuid | null"
func goTypeToTSType(field reflect.StructField, visited map[reflect.Type]bool, typeQueue map[reflect.Type]bool) string {
	t := field.Type
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
		// PATCH: Handle pointer to uuid.UUID
		if t.Elem() == reflect.TypeOf(uuid.UUID{}) {
			return "uuid | null"
		}
		return goTypeToTSType(reflect.StructField{Type: t.Elem(), Tag: field.Tag}, visited, typeQueue) + " | null"
	case reflect.Slice, reflect.Array:
		return goTypeToTSType(reflect.StructField{Type: t.Elem(), Tag: field.Tag}, visited, typeQueue) + "[]"
	case reflect.Map:
		if t.Key().Kind() == reflect.String {
			return "{ [key: string]: " + goTypeToTSType(reflect.StructField{Type: t.Elem(), Tag: field.Tag}, visited, typeQueue) + " }"
		}
		return "any"
	case reflect.Struct:
		// PATCH: Handle uuid.UUID
		if t == reflect.TypeOf(uuid.UUID{}) {
			return "uuid"
		}
		if t == reflect.TypeOf(time.Time{}) {
			return "string"
		}
		if t.Name() == "" { // anonymous struct: inline fields
			return structTypeToTSInline(t, visited, typeQueue)
		}
		return t.Name() // reference by name, do not add to typeQueue!
	default:
		return "any"
	}
}

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

// Helper to generate all TypeScript interfaces as a single Markdown string
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
			if isNamedStruct(t) && t != rootType {
				handled[t] = true
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

type NestedChild struct {
	ID    int      `json:"id" description:"Child ID" validate:"min=1,max=9999"`
	Tags  []string `json:"tags" description:"Tag list"`
	Flags []bool   `json:"flags" description:"Flags"`
}

type NestedParent struct {
	Name        string        `json:"name" description:"Parent name" validate:"required,min=3,max=50"`
	CreatedAt   time.Time     `json:"created_at" description:"Creation date"`
	Description *string       `json:"description" description:"Optional description"`
	Children    []NestedChild `json:"children" description:"Children list"`
	Extras      []any         `json:"extras" description:"Extra stuff"`
}

type FullComplexObject struct {
	Title       string           `json:"title" description:"Object title" validate:"required,min=5,max=100"`
	Count       int              `json:"count" description:"Count" validate:"min=0,max=1000"`
	Price       float64          `json:"price" description:"Price" validate:"min=0"`
	Active      bool             `json:"active" description:"Is Active"`
	Metadata    map[string]any   `json:"metadata" description:"Metadata map"`
	Nested      NestedParent     `json:"nested" description:"Nested parent object"`
	MixedArray  []any            `json:"mixed_array" description:"Mixed array"`
	ObjectArray []map[string]any `json:"object_array" description:"Array of objects"`
	ExtraNested [][]NestedChild  `json:"extra_nested" description:"2D array of NestedChild"`
}

// Circular example
type Friend struct {
	Name string `json:"name" description:"Friend name" enum:"best,close,acquaintance" validate:"oneof=best close acquaintance"`
	User *User  `json:"user" description:"User reference"`
}
type User struct {
	Name      string   `json:"name" description:"User name"`
	Friends   []Friend `json:"friends" description:"List of friends"`
	InlineObj struct {
		Tags []string `json:"tags"`
	} `json:"inline_obj"`
}

// Enum/oneof example
type FruitBasket struct {
	Fruit string `json:"fruit" description:"Type of fruit" enum:"apple,banana,orange" validate:"oneof=apple banana orange"`
	Size  string `json:"size" description:"Basket size" validate:"oneof=small medium large"`
}

// func main() {
// 	printAllTSInterfacesMarkdown(reflect.TypeOf(User{}))
// }

func ExtractTSInterfaceName(ts string) string {
	lines := strings.Split(ts, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		// Accept lines that start with "export interface" or just "interface"
		if strings.HasPrefix(line, "export interface ") {
			after := strings.TrimPrefix(line, "export interface ")
			parts := strings.Fields(after)
			if len(parts) > 0 {
				return parts[0]
			}
		} else if strings.HasPrefix(line, "interface ") {
			after := strings.TrimPrefix(line, "interface ")
			parts := strings.Fields(after)
			if len(parts) > 0 {
				return parts[0]
			}
		}
	}
	return ""
}

func GetAllRequestInterfaces(routes []Route) []APIInterfaces {
	seen := make(map[string]struct{})
	result := []APIInterfaces{}
	for _, rt := range routes {
		if rt.Request == "" || rt.Request == "None" {
			continue
		}
		name := ExtractTSInterfaceName(rt.Request)
		if name == "" {
			continue
		}
		if _, already := seen[name]; already {
			continue
		}
		seen[name] = struct{}{}
		result = append(result, APIInterfaces{
			Key:   name,
			Value: rt.Request,
		})
	}
	return result
}

func GetAllResponseInterfaces(routes []Route) []APIInterfaces {
	seen := make(map[string]struct{})
	result := []APIInterfaces{}
	for _, rt := range routes {
		if rt.Response == "" || rt.Response == "None" {
			continue
		}
		name := ExtractTSInterfaceName(rt.Response)
		if name == "" {
			continue
		}
		if _, already := seen[name]; already {
			continue
		}
		seen[name] = struct{}{}
		result = append(result, APIInterfaces{
			Key:   name,
			Value: rt.Response,
		})
	}
	return result
}

func (h *HorizonAPIService) GroupedRoutes() API {
	grouped := make(map[string][]Route)
	interfacesMap := make(map[string]map[string]struct{})
	for _, rt := range h.routesList {
		trimmed := strings.TrimPrefix(rt.Route, "/")
		segments := strings.Split(trimmed, "/")
		key := "/"
		if len(segments) > 0 && segments[0] != "" {
			key = segments[0]
		}
		grouped[key] = append(grouped[key], rt)
		if interfacesMap[key] == nil {
			interfacesMap[key] = make(map[string]struct{})
		}
		// Add request/response interface NAMES
		if rt.Request != "" {
			name := ExtractTSInterfaceName(rt.Request)
			if name != "" {
				interfacesMap[key][name] = struct{}{}
			}
		}
		if rt.Response != "" {
			name := ExtractTSInterfaceName(rt.Response)
			if name != "" {
				interfacesMap[key][name] = struct{}{}
			}
		}
	}
	routePaths := make([]string, 0, len(grouped))
	for route := range grouped {
		routePaths = append(routePaths, route)
	}
	sort.Strings(routePaths)
	result := make([]GroupedRoute, 0, len(routePaths))
	for _, route := range routePaths {
		methodGroup := grouped[route]
		sort.Slice(methodGroup, func(i, j int) bool {
			return methodGroup[i].Method < methodGroup[j].Method
		})
		interfaces := make([]string, 0, len(interfacesMap[route]))
		for iface := range interfacesMap[route] {
			interfaces = append(interfaces, iface)
		}
		sort.Strings(interfaces)

		result = append(result, GroupedRoute{
			Key:    route,
			Routes: methodGroup,
		})
	}
	return API{
		GroupedRoutes: result,
		Requests:      GetAllRequestInterfaces(h.routesList),
		Responses:     GetAllResponseInterfaces(h.routesList),
	}
}
