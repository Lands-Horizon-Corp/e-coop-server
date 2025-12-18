package handlers

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"html/template"
	"io"
	"math"
	"math/big"
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
	"github.com/microcosm-cc/bluemonday"
	"github.com/rotisserie/eris"
)

const (
	Green  = "\033[32m"
	Blue   = "\033[34m"
	Yellow = "\033[33m"
	Red    = "\033[31m"
	Reset  = "\033[0m"
	Cyan   = "\033[36m"
)

func Sanitize(input string) string {
	return bluemonday.UGCPolicy().Sanitize(strings.TrimSpace(input))
}

func IsValidFilePath(p string) error {
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

func FileExists(filename string) bool {
	if filename = strings.TrimSpace(filename); filename == "" {
		return false
	}
	info, err := os.Stat(filename)
	return err == nil && !info.IsDir()
}

func LoadTemplatesIfExists(e *echo.Echo, pattern string) {
	if matches, _ := filepath.Glob(pattern); len(matches) > 0 {
		e.Renderer = &TemplateRenderer{
			templates: template.Must(template.ParseGlob(pattern)),
		}
	}
}

type TemplateRenderer struct {
	templates *template.Template
}

func (t *TemplateRenderer) Render(w io.Writer, name string, data any, _ echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func Validate[T any](ctx echo.Context, v *validator.Validate) (*T, error) {
	var req T
	if err := ctx.Bind(&req); err != nil {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "invalid request format")
	}
	if err := v.Struct(req); err != nil {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "validation failed: "+err.Error())
	}
	return &req, nil
}

func EngineUUIDParam(ctx echo.Context, param string) (*uuid.UUID, error) {
	id, err := uuid.Parse(ctx.Param(param))
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "invalid UUID format")
	}
	return &id, nil
}

func Capitalize(s string) string {
	if s == "" {
		return s
	}
	r, size := utf8.DecodeRuneInString(s)
	return string(unicode.ToUpper(r)) + s[size:]
}

func StringFormat(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func UniqueStrings(input []string) []string {
	seen := make(map[string]struct{}, len(input))
	result := make([]string, 0, len(input))

	for _, v := range input {
		if _, exists := seen[v]; !exists {
			seen[v] = struct{}{}
			result = append(result, v)
		}
	}
	return result
}

func MergeStrings(defaults, overrides []string) []string {
	seen := make(map[string]struct{}, len(defaults)+len(overrides))
	result := make([]string, 0, len(defaults)+len(overrides))
	for _, s := range [][]string{defaults, overrides} {
		for _, v := range s {
			capV := Capitalize(v)
			if capV == "" {
				continue
			}
			if _, exists := seen[capV]; !exists {
				seen[capV] = struct{}{}
				result = append(result, capV)
			}
		}
	}
	return result
}

func IsZero[T comparable](v T) bool {
	return v == *new(T)
}

func GetFreePort() int {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 8123
	}
	defer func() { _ = l.Close() }()
	return l.Addr().(*net.TCPAddr).Port
}

func GenerateToken() (string, error) {
	id, err := uuid.NewRandom()
	if err != nil {
		return "", eris.Wrap(err, "token generation failed")
	}
	return id.String(), nil
}

func GenerateRandomBytes(n uint32) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	return b, err
}

func Create32ByteKey(key []byte) []byte {
	if len(key) >= 32 {
		return key[:32]
	}
	padded := make([]byte, 32)
	copy(padded, key)
	return padded
}

var phoneRegex = regexp.MustCompile(`^\+?(?:\d{1,4})?\d{7,14}$`)

func IsValidEmail(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}

func IsValidPhoneNumber(phone string) bool {
	return phoneRegex.MatchString(phone)
}

func IsValidURL(rawURL string) bool {
	if rawURL == "" {
		return false
	}
	u, err := url.ParseRequestURI(rawURL)
	return err == nil &&
		(u.Scheme == "http" || u.Scheme == "https") &&
		u.Host != "" &&
		!strings.ContainsAny(rawURL, `<>"`)
}

func ParseCoordinate(value string) float64 {
	if value == "" {
		return 0.0
	}
	coord, _ := strconv.ParseFloat(value, 64)
	return coord
}

func ParseUUID(s *string) uuid.UUID {
	if s == nil || *s == "" {
		return uuid.Nil
	}
	id, err := uuid.Parse(*s)
	if err != nil {
		return uuid.Nil
	}
	return id
}

func GenerateRandomDigits(size int) (int, error) {
	switch {
	case size > 8:
		return 0, errors.New("digit size exceeds maximum (8)")
	case size <= 0:
		return 0, errors.New("digit size must be positive")
	}

	minVal := big.NewInt(0).Exp(big.NewInt(10), big.NewInt(int64(size-1)), nil)
	maxVal := big.NewInt(0).Exp(big.NewInt(10), big.NewInt(int64(size)), nil)
	maxVal.Sub(maxVal, big.NewInt(1))

	n, err := rand.Int(rand.Reader, maxVal.Sub(maxVal, minVal))
	if err != nil {
		return 0, err
	}
	return int(n.Add(n, minVal).Int64()), nil
}

func Retry(ctx context.Context, maxAttempts int, delay time.Duration, op func() error) error {
	var err error
	for range maxAttempts {
		if err = op(); err == nil {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
		}
	}
	return eris.Wrapf(err, "after %d attempts", maxAttempts)
}

func PrintASCIIArt() {
	art := `
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

	for line := range strings.SplitSeq(art, "\n") {
		var b strings.Builder
		for _, r := range line {
			switch r {
			case '@':
				b.WriteString(Blue + string(r) + Reset)
			case ',', '.':
				b.WriteString(Green + string(r) + Reset)
			default:
				b.WriteRune(r)
			}
		}
		fmt.Println(b.String())
	}
}

func Ptr[T any](v T) *T {
	return &v
}

func UUIDPtrEqual(a, b *uuid.UUID) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}

func Zip[T, U any](slice1 []T, slice2 []U) []struct {
	First  T
	Second U
} {
	minLen := min(len(slice2), len(slice1))

	result := make([]struct {
		First  T
		Second U
	}, minLen)
	for i := range minLen {
		result[i] = struct {
			First  T
			Second U
		}{slice1[i], slice2[i]}
	}
	return result
}

func HasFileExtension(filename string) bool {
	return strings.Contains(filename, ".") && !strings.HasSuffix(filename, ".")
}

func GetExtensionFromContentType(contentType string) string {
	contentTypeMap := map[string]string{
		"audio/aac":    ".aac",
		"audio/midi":   ".mid",
		"audio/x-midi": ".midi",
		"audio/mpeg":   ".mp3",
		"audio/ogg":    ".oga",
		"audio/wav":    ".wav",
		"audio/webm":   ".weba",
		"audio/3gpp":   ".3gp",
		"audio/3gpp2":  ".3g2",
		"audio/flac":   ".flac", // Added: Free Lossless Audio Codec
		"audio/x-aiff": ".aiff", // Added: Audio Interchange File Format
		"audio/mp4":    ".m4a",  // Added: MPEG-4 Audio

		"video/x-msvideo":  ".avi",
		"video/mp4":        ".mp4",
		"video/mpeg":       ".mpeg",
		"video/ogg":        ".ogv",
		"video/mp2t":       ".ts",
		"video/webm":       ".webm",
		"video/3gpp":       ".3gp",
		"video/3gpp2":      ".3g2",
		"video/quicktime":  ".mov", // Added: QuickTime Movie
		"video/x-matroska": ".mkv", // Added: Matroska Video
		"video/x-flv":      ".flv", // Added: Flash Video

		"image/apng":               ".apng",
		"image/avif":               ".avif",
		"image/bmp":                ".bmp",
		"image/gif":                ".gif",
		"image/jpeg":               ".jpg",
		"image/png":                ".png",
		"image/svg+xml":            ".svg",
		"image/tiff":               ".tiff",
		"image/webp":               ".webp",
		"image/vnd.microsoft.icon": ".ico",
		"image/x-icon":             ".ico",  // Added: Alternative for icons
		"image/heic":               ".heic", // Added: High Efficiency Image Container
		"image/heif":               ".heif", // Added: High Efficiency Image Format

		"font/otf":        ".otf",
		"font/ttf":        ".ttf",
		"font/woff":       ".woff",
		"font/woff2":      ".woff2",
		"font/collection": ".ttc", // Added: TrueType Collection
		"font/sfnt":       ".ttf", // Added: Generic SFNT font (often TrueType)

		"application/x-abiword":        ".abw",
		"application/x-freearc":        ".arc",
		"application/vnd.amazon.ebook": ".azw",
		"application/octet-stream":     ".bin",
		"application/x-bzip":           ".bz",
		"application/x-bzip2":          ".bz2",
		"application/x-cdf":            ".cda",
		"application/x-csh":            ".csh",
		"application/msword":           ".doc",
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document": ".docx",
		"application/vnd.ms-fontobject":                                           ".eot",
		"application/epub+zip":                                                    ".epub",
		"application/gzip":                                                        ".gz",
		"application/x-gzip":                                                      ".gz", // Non-standard but common on Windows/macOS
		"application/java-archive":                                                ".jar",
		"application/json":                                                        ".json",
		"application/ld+json":                                                     ".jsonld",
		"application/vnd.apple.installer+xml":                                     ".mpkg",
		"application/vnd.oasis.opendocument.presentation":                         ".odp",
		"application/vnd.oasis.opendocument.spreadsheet":                          ".ods",
		"application/vnd.oasis.opendocument.text":                                 ".odt",
		"application/ogg":                                                         ".ogx",
		"application/pdf":                                                         ".pdf",
		"application/x-httpd-php":                                                 ".php",
		"application/vnd.ms-powerpoint":                                           ".ppt",
		"application/vnd.openxmlformats-officedocument.presentationml.presentation": ".pptx",
		"application/vnd.rar":       ".rar",
		"application/rtf":           ".rtf",
		"application/x-sh":          ".sh",
		"application/x-tar":         ".tar",
		"application/vnd.visio":     ".vsd",
		"application/manifest+json": ".webmanifest",
		"application/xhtml+xml":     ".xhtml",
		"application/vnd.ms-excel":  ".xls",
		"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet": ".xlsx",
		"application/xml":                         ".xml",
		"application/vnd.mozilla.xul+xml":         ".xul",
		"application/zip":                         ".zip",
		"application/x-zip-compressed":            ".zip", // Non-standard but common on Windows
		"application/x-7z-compressed":             ".7z",
		"application/x-rar-compressed":            ".rar",        // Added: Alternative for RAR
		"application/x-www-form-urlencoded":       ".urlencoded", // Added: Form data (not a file, but common)
		"application/vnd.android.package-archive": ".apk",        // Added: Android Package
		"application/x-apple-diskimage":           ".dmg",        // Added: Apple Disk Image
		"application/x-debian-package":            ".deb",        // Added: Debian Package
		"application/x-redhat-package-manager":    ".rpm",        // Added: RPM Package

		"text/css":           ".css",
		"text/csv":           ".csv",
		"text/html":          ".html",
		"text/javascript":    ".js",
		"text/calendar":      ".ics",
		"text/markdown":      ".md",
		"text/plain":         ".txt",
		"text/xml":           ".xml",
		"text/x-python":      ".py",   // Added: Python script
		"text/x-shellscript": ".sh",   // Added: Shell script (alternative)
		"text/vcard":         ".vcf",  // Added: vCard
		"text/yaml":          ".yaml", // Added: YAML
		"text/x-yaml":        ".yml",  // Added: YAML alternative

		"model/gltf+json":   ".gltf", // Added: glTF JSON
		"model/gltf-binary": ".glb",  // Added: glTF Binary
		"model/obj":         ".obj",  // Added: Wavefront OBJ
		"model/stl":         ".stl",  // Added: Stereolithography

		"text/vtt":             ".vtt", // Added: WebVTT
		"application/x-subrip": ".srt", // Added: SubRip

		"application/x-msdownload":      ".exe", // Added: Windows Executable
		"application/x-shockwave-flash": ".swf", // Added: Shockwave Flash

		"application/sql": ".sql", // Added: SQL script

		"application/rss+xml":  ".rss",  // Added: RSS Feed
		"application/atom+xml": ".atom", // Added: Atom Feed
		"application/wasm":     ".wasm", // Added: WebAssembly
	}
	cleanContentType := strings.Split(contentType, ";")[0]
	cleanContentType = strings.TrimSpace(cleanContentType)

	if ext, exists := contentTypeMap[cleanContentType]; exists {
		return ext
	}
	return ""
}

func GetID[T any](entity *T) (uuid.UUID, error) {
	v := reflect.ValueOf(entity).Elem()
	idField := v.FieldByName("ID")
	if !idField.IsValid() {
		return uuid.Nil, eris.New("entity missing ID field")
	}

	id, ok := idField.Interface().(uuid.UUID)
	if !ok {
		return uuid.Nil, eris.New("ID field is not UUID type")
	}
	return id, nil
}

func SetID[T any](entity *T, id uuid.UUID) error {
	v := reflect.ValueOf(entity).Elem()
	idField := v.FieldByName("ID")
	if !idField.IsValid() {
		return eris.New("entity missing ID field")
	}
	if !idField.CanSet() {
		return eris.New("ID field cannot be set")
	}

	idField.Set(reflect.ValueOf(id))
	return nil
}

func UnitToInches(value float64, unit string) (float64, error) {
	unit = strings.ToLower(strings.TrimSpace(unit))
	switch unit {
	case "mm":
		return value / 25.4, nil
	case "cm":
		return value / 2.54, nil
	case "m", "meter", "meters":
		return value * 39.3701, nil
	case "km", "kilometer", "kilometers":
		return value * 39370.1, nil
	case "in", "inch", "inches":
		return value, nil
	case "ft", "foot", "feet":
		return value * 12, nil
	case "yd", "yard", "yards":
		return value * 36, nil
	case "mi", "mile", "miles":
		return value * 63360, nil
	case "px", "pixel", "pixels":
		return value / 96.0, nil
	case "pt", "point", "points":
		return value / 72.0, nil
	case "pc", "pica", "picas":
		return value / 6.0, nil
	case "dp", "dip", "density-independent-pixel":
		return value / 96.0, nil
	case "twip", "twips":
		return value / 1440.0, nil
	case "em":
		return (value * 16.0) / 96.0, nil
	case "rem":
		return (value * 16.0) / 96.0, nil
	case "f", "ftin", "feet-inches", "foot-inch":
		feet := math.Floor(value)
		inches := (value - feet) * 12
		return feet*12 + inches, nil
	default:
		return 0, errors.New("unsupported unit: " + unit)
	}
}
func ToReadableDate(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	if t.Hour() == 0 && t.Minute() == 0 && t.Second() == 0 {
		return t.Format("January 2, 2006")
	}
	return t.Format("January 2, 2006 3:04 PM")
}

func AddMonthsPreserveDay(t time.Time, months int) time.Time {
	year := t.Year()
	month := int(t.Month())
	day := t.Day()

	month += months
	year += (month - 1) / 12
	month = (month-1)%12 + 1

	loc := t.Location()
	firstOfTarget := time.Date(year, time.Month(month), 1, t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), loc)
	lastOfTarget := firstOfTarget.AddDate(0, 1, -1).Day()
	if day > lastOfTarget {
		day = lastOfTarget
	}
	return time.Date(year, time.Month(month), day, t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), loc)
}

func ToPascalCase(s string) string {
	if len(s) == 0 {
		return s
	}
	parts := strings.Split(s, "_")
	for i, part := range parts {
		if len(part) > 0 {
			parts[i] = strings.ToUpper(part[:1]) + part[1:]
		}
	}
	return strings.Join(parts, "")
}

func ToSnakeCase(str string) string {
	var result []rune
	for i, r := range str {
		if unicode.IsUpper(r) {
			if i > 0 {
				result = append(result, '_')
			}
			result = append(result, unicode.ToLower(r))
		} else {
			result = append(result, r)
		}
	}
	return string(result)
}
