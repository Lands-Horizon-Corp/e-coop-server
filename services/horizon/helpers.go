package horizon

import (
	"context"
	"crypto/rand"
	"errors"
	"math/big"

	"fmt"
	"html/template"
	"net"
	"net/http"
	"net/mail"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/go-playground/validator"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

const (
	Green  = "\033[32m"
	Blue   = "\033[34m"
	Yellow = "\033[33m"
	Red    = "\033[31m"
	Reset  = "\033[0m"
	Cyan   = "\033[36m"
)

func isValidFilePath(p string) error {
	info, err := os.Stat(p)
	if os.IsNotExist(err) {
		return errors.New("not exist")
	}
	if err != nil {
		return err
	}
	if info.IsDir() {
		return errors.New("is dir not file")
	}
	return nil
}

// IsValidEmail checks if the provided string is a valid email address format
func IsValidEmail(email string) bool {
	if email == "" {
		return false
	}
	_, err := mail.ParseAddress(email)
	return err == nil
}

func IsValidPhoneNumber(phoneNumber string) bool {
	re := regexp.MustCompile(`^\+?(?:\d{1,4})?\d{7,14}$`)
	return re.MatchString(phoneNumber)
}

func GenerateRandomBytes(n uint32) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}

	return b, nil
}

func Create32ByteKey(key []byte) string {
	if len(key) > 32 {
		return string(key[:32])
	}
	padded := make([]byte, 32)
	copy(padded, key)
	return string(padded)
}

func IsValidURL(rawURL string) bool {
	if rawURL == "" {
		return false
	}
	u, err := url.ParseRequestURI(rawURL)
	if err != nil {
		return false
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return false
	}
	if u.Host == "" {
		return false
	}
	if strings.ContainsAny(rawURL, " <>\"") {
		return false
	}

	return true
}

func GenerateRandomDigits(size int) (int, error) {
	if size > 8 {
		return 0, errors.New("size must not exceed 8 digits")
	}
	if size <= 0 {
		return 0, errors.New("size must be a positive integer")
	}

	min := intPow(10, size-1)
	max := intPow(10, size) - 1
	rangeSize := max - min + 1

	nBig, err := rand.Int(rand.Reader, big.NewInt(int64(rangeSize)))
	if err != nil {
		return 0, err
	}
	return int(nBig.Int64()) + min, nil
}

func intPow(a, b int) int {
	result := 1
	for i := 0; i < b; i++ {
		result *= a
	}
	return result
}

func Capitalize(s string) string {
	if s == "" {
		return s
	}
	r, size := utf8.DecodeRuneInString(s)
	return string(unicode.ToUpper(r)) + s[size:]
}

func MergeString(defaults, overrides []string) []string {
	totalCap := len(defaults) + len(overrides)
	seen := make(map[string]struct{}, totalCap)
	out := make([]string, 0, totalCap)
	for _, slice := range [][]string{defaults, overrides} {
		for _, p := range slice {
			cp := Capitalize(p)
			if cp == "" {
				continue
			}
			if _, exists := seen[cp]; !exists {
				seen[cp] = struct{}{}
				out = append(out, cp)
			}
		}
	}
	return out
}

func EngineUUIDParam(ctx echo.Context, idParam string) (*uuid.UUID, error) {
	param := ctx.Param(idParam)
	id, err := uuid.Parse(param)
	if err != nil {
		return nil, ctx.JSON(http.StatusBadRequest, map[string]string{"error": "invalid ID"})
	}
	return &id, nil
}

func ParseUUID(s *string) uuid.UUID {
	if s == nil || strings.TrimSpace(*s) == "" {
		return uuid.Nil
	}
	if id, err := uuid.Parse(*s); err == nil {
		return id
	}
	return uuid.Nil
}

func Retry(ctx context.Context, maxAttempts int, delay time.Duration, operation func() error) error {
	var err error
	for i := range maxAttempts {
		err = operation()
		if err == nil {
			return nil
		}
		if i < maxAttempts-1 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(delay):
				// wait before retrying
			}
		}
	}
	return err
}

func GetFreePort() int {
	l, err := net.Listen("tcp", ":0")
	if err != nil {
		panic(err)
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port
}

func FileExists(filename string) bool {
	filename = strings.TrimSpace(filename)
	if filename == "" {
		return false
	}
	info, err := os.Stat(filename)
	return err == nil && !info.IsDir()
}

func GenerateToken() (string, error) {
	id, err := uuid.NewRandom()
	if err != nil {
		return "", err
	}
	return id.String(), nil
}

func IsZero[T any](v T) bool {
	return reflect.ValueOf(v).IsZero()
}

func PrintASCIIArt() {
	asciiArt := `
	           ..............                            
            .,,,,,,,,,,,,,,,,,,,                             
        ,,,,,,,,,,,,,,,,,,,,,,,,,,                          
      ,,,,,,,,,,,,,,  .,,,,,,,,,,,,,                        
    ,,,,,,,,,,           ,,,,,,,,,,,                     
    ,,,,,,,          .,,,,,,,,,,,                          
  @@,,,,,,          ,,,,,,,,,,,,                             
@@@,,,,.@@      .,,,,,,,,,,,                                
@,,,,,,,@@    ,,,,,,,,,,,                                   
  ,,,,@@@       ,,,,,,                                      
    @@@@@@@                                          
    @@@@@@@@@@           @@@@@@@@                          
      @@@@@@@@@@@@@@  @@@@@@@@@@@@                          
        @@@@@@@@@@@@@@@@@@@@@@@@@@                          
            @@@@@@@@@@@@@@@@@@@@                             
                  @@@@@@@@
	`

	lines := strings.SplitSeq(asciiArt, "\n")

	for line := range lines {
		coloredLine := ""
		for _, char := range line {
			switch char {
			case '@':
				coloredLine += Blue + "@" + Reset
			case ',', '.':
				coloredLine += Green + string(char) + Reset
			default:
				coloredLine += string(char)
			}
		}
		fmt.Println(coloredLine)
	}
}

func Validate[T any](ctx echo.Context, v *validator.Validate) (*T, error) {
	var req T
	if err := ctx.Bind(&req); err != nil {
		return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := v.Struct(req); err != nil {
		return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return &req, nil
}

func StringFormat(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func ParseCoordinate(value string) float64 {
	if value == "" {
		return 0.0
	}
	coord, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0.0
	}
	return coord
}

// LoadTemplatesIfExists sets the renderer if templates are found.
func LoadTemplatesIfExists(e *echo.Echo, pattern string) {
	matches, err := filepath.Glob(pattern)
	if err != nil || len(matches) == 0 {
		return
	}
	e.Renderer = &TemplateRenderer{
		templates: template.Must(template.ParseGlob(pattern)),
	}
}

// isSuspiciousPath checks if a path is forbidden.

func IsSuspiciousPath(path string) bool {
	lower := strings.ToLower(path)
	decoded, _ := url.PathUnescape(lower)
	if strings.ContainsAny(lower, "\\/") {
		if strings.Contains(lower, "../") || strings.Contains(decoded, "../") ||
			strings.Contains(lower, "..\\") || strings.Contains(decoded, "..\\") {
			return true
		}
	}
	if strings.Contains(lower, "%2e%2e%2f") || strings.Contains(lower, "%2e%2e%5c") ||
		strings.Contains(decoded, "%2e%2e%2f") || strings.Contains(decoded, "%2e%2e%5c") {
		return true
	}
	var extMap = map[string]struct{}{}
	for _, ext := range forbiddenExtensions {
		extMap[ext] = struct{}{}
	}
	for ext := range extMap {
		if strings.HasSuffix(lower, ext) || strings.HasSuffix(decoded, ext) {
			return true
		}
	}
	var substrMap = map[string]struct{}{}
	for _, substr := range forbiddenSubstrings {
		substrMap[substr] = struct{}{}
	}
	for substr := range substrMap {
		if strings.Contains(lower, substr) || strings.Contains(decoded, substr) {
			return true
		}
	}

	return false
}

// Converts a string to camelCase
func toCamelCase(s string) string {
	if s == "" {
		return s
	}
	runes := []rune(s)
	runes[0] = unicode.ToLower(runes[0])
	return string(runes)
}

// Indents lines by spaces
func indent(s string, spaces int) string {
	pad := strings.Repeat(" ", spaces)
	lines := strings.Split(s, "\n")
	for i := range lines {
		if lines[i] != "" {
			lines[i] = pad + lines[i]
		}
	}
	return strings.Join(lines, "\n")
}

// Parses and formats go-playground/validator tags for documentation
func formatValidateTag(validate string) string {
	if validate == "" {
		return ""
	}
	var parts []string
	for _, rule := range strings.Split(validate, ",") {
		rule = strings.TrimSpace(rule)
		switch {
		case rule == "required":
			parts = append(parts, "required")
		case rule == "email":
			parts = append(parts, "format: email")
		case strings.HasPrefix(rule, "min="):
			parts = append(parts, fmt.Sprintf("min: %s", strings.TrimPrefix(rule, "min=")))
		case strings.HasPrefix(rule, "max="):
			parts = append(parts, fmt.Sprintf("max: %s", strings.TrimPrefix(rule, "max=")))
		case strings.HasPrefix(rule, "oneof="):
			options := strings.TrimPrefix(rule, "oneof=")
			parts = append(parts, fmt.Sprintf("one of: %s", options))
		case strings.HasPrefix(rule, "eqfield="):
			parts = append(parts, fmt.Sprintf("equals field: %s", strings.TrimPrefix(rule, "eqfield=")))
		case strings.HasPrefix(rule, "len="):
			parts = append(parts, fmt.Sprintf("length: %s", strings.TrimPrefix(rule, "len=")))
		default:
			parts = append(parts, rule)
		}
	}
	return "Validation: " + strings.Join(parts, ", ")
}

// Converts a Go struct to a TypeScript interface as Markdown code block
func StructToTypeScriptMarkdown(i any, interfaceName string) string {
	return ToMarkdownCodeBlock(fmt.Sprintf("export interface %s %s", interfaceName, structToTypeScriptTSComment(reflect.TypeOf(i), 0)), "typescript")
}

// Converts a Go struct type to a TypeScript interface
func structToTypeScriptTSComment(t reflect.Type, indentLevel int) string {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return "// Not a struct"
	}

	var fields []string
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if field.PkgPath != "" {
			continue // skip unexported fields
		}

		jsonTag := field.Tag.Get("json")
		if jsonTag == "-" {
			continue
		}
		name := strings.Split(jsonTag, ",")[0]
		if name == "" {
			name = field.Name
		}
		name = toCamelCase(name)

		isOmitEmpty := strings.Contains(jsonTag, "omitempty")
		validateTag := field.Tag.Get("validate")
		description := field.Tag.Get("description")
		enumTag := field.Tag.Get("enum")
		defaultTag := field.Tag.Get("default")
		minTag := field.Tag.Get("min")
		maxTag := field.Tag.Get("max")

		isOptional := isOmitEmpty || (!strings.Contains(validateTag, "required") && field.Type.Kind() != reflect.Struct)

		tsType := mapGoTypeToTypeScriptTSComment(field.Type, indentLevel+2, enumTag, validateTag)

		// Compose doc lines
		var docLines []string
		if description != "" {
			docLines = append(docLines, description)
		}
		if validateTag != "" {
			docLines = append(docLines, formatValidateTag(validateTag))
		}
		if defaultTag != "" {
			docLines = append(docLines, fmt.Sprintf("Default: %s", defaultTag))
		}
		if enumTag != "" {
			docLines = append(docLines, fmt.Sprintf("Enum: %s", enumTag))
		}
		if minTag != "" {
			docLines = append(docLines, fmt.Sprintf("Min: %s", minTag))
		}
		if maxTag != "" {
			docLines = append(docLines, fmt.Sprintf("Max: %s", maxTag))
		}
		comment := ""
		if len(docLines) > 0 {
			comment = fmt.Sprintf("/**\n%s\n*/\n", indent(strings.Join(docLines, "\n"), indentLevel+2))
		}

		fieldLine := fmt.Sprintf("%s%s%s: %s;", strings.Repeat(" ", indentLevel+2), name, ternary(isOptional, "?", ""), tsType)
		fields = append(fields, comment+fieldLine)
	}

	var lines []string
	lines = append(lines, "{")
	lines = append(lines, fields...)
	lines = append(lines, strings.Repeat(" ", indentLevel)+"}")
	return strings.Join(lines, "\n")
}

// Type mapping, including enums and parsing oneof validation for union types
func mapGoTypeToTypeScriptTSComment(t reflect.Type, indentLevel int, enumTag string, validateTag string) string {
	switch t.Kind() {
	case reflect.Bool:
		return "boolean"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		return "number"
	case reflect.String:
		// oneof in validateTag
		if strings.Contains(validateTag, "oneof=") {
			oneof := ""
			for rule := range strings.SplitSeq(validateTag, ",") {
				if strings.HasPrefix(rule, "oneof=") {
					oneof = strings.TrimPrefix(rule, "oneof=")
					break
				}
			}
			vals := strings.Fields(oneof)
			for i := range vals {
				vals[i] = fmt.Sprintf("\"%s\"", vals[i])
			}
			return strings.Join(vals, " | ")
		}
		if enumTag != "" {
			values := strings.Split(enumTag, ",")
			for i := range values {
				values[i] = fmt.Sprintf("\"%s\"", strings.TrimSpace(values[i]))
			}
			return strings.Join(values, " | ")
		}
		return "string"
	case reflect.Slice, reflect.Array:
		elemType := mapGoTypeToTypeScriptTSComment(t.Elem(), indentLevel, "", "")
		return elemType + "[]"
	case reflect.Map:
		keyType := mapGoTypeToTypeScriptTSComment(t.Key(), indentLevel, "", "")
		valType := mapGoTypeToTypeScriptTSComment(t.Elem(), indentLevel, "", "")
		return fmt.Sprintf("{ [key: %s]: %s }", keyType, valType)
	case reflect.Struct:
		switch t.String() {
		case "time.Time", "uuid.UUID":
			return "string"
		}
		return indent(structToTypeScriptTSComment(t, indentLevel), 0)
	case reflect.Ptr:
		return mapGoTypeToTypeScriptTSComment(t.Elem(), indentLevel, enumTag, validateTag)
	case reflect.Interface:
		return "any"
	default:
		return "any"
	}
}

func ternary(cond bool, a, b string) string {
	if cond {
		return a
	}
	return b
}

// Output as Markdown code block
func ToMarkdownCodeBlock(ts string, lang string) string {
	return fmt.Sprintf("```%s\n%s\n```\n", lang, ts)
}
