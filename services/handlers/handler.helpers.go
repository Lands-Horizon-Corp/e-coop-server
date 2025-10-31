package handlers

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"html/template"
	"io"
	"math/big"
	"net"
	"net/http"
	"net/mail"
	"net/url"
	"os"
	"path/filepath"
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

// ANSI color escape codes
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

// File operations ------------------------------------------------------------

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

// FileExists checks if a file exists and is not a directory
func FileExists(filename string) bool {
	if filename = strings.TrimSpace(filename); filename == "" {
		return false
	}
	info, err := os.Stat(filename)
	return err == nil && !info.IsDir()
}

// LoadTemplatesIfExists initializes Echo renderer if templates match pattern
func LoadTemplatesIfExists(e *echo.Echo, pattern string) {
	if matches, _ := filepath.Glob(pattern); len(matches) > 0 {
		e.Renderer = &TemplateRenderer{
			templates: template.Must(template.ParseGlob(pattern)),
		}
	}
}

// TemplateRenderer implements echo.Renderer interface
type TemplateRenderer struct {
	templates *template.Template
}

func (t *TemplateRenderer) Render(w io.Writer, name string, data any, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

// Validation helpers ---------------------------------------------------------

// Validate binds and validates request using validator
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

// EngineUUIDParam extracts UUID from Echo path parameter
func EngineUUIDParam(ctx echo.Context, param string) (*uuid.UUID, error) {
	id, err := uuid.Parse(ctx.Param(param))
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "invalid UUID format")
	}
	return &id, nil
}

// String operations ----------------------------------------------------------

// Capitalize first rune in string
func Capitalize(s string) string {
	if s == "" {
		return s
	}
	r, size := utf8.DecodeRuneInString(s)
	return string(unicode.ToUpper(r)) + s[size:]
}

// StringFormat safely dereferences string pointer
func StringFormat(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// UniqueStrings returns deduplicated slice
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

// MergeStrings combines and deduplicates slices with capitalization
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

// IsZero checks if value is type's zero-value
func IsZero[T comparable](v T) bool {
	return v == *new(T)
}

// Network utilities ----------------------------------------------------------

// GetFreePort finds available TCP port bound to localhost (avoids binding all interfaces)
func GetFreePort() int {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 8123
	}
	defer func() { _ = l.Close() }()
	return l.Addr().(*net.TCPAddr).Port
}

// Security helpers -----------------------------------------------------------

var (
	forbiddenExtensions = []string{
		".exe", ".bat", ".sh", ".php", ".asp", ".aspx", ".jsp", ".cgi", ".go",
		".env", ".yaml", ".yml", ".ini", ".config", ".conf", ".xml", ".git",
		".htaccess", ".htpasswd", ".backup", ".secret", ".credential", ".password",
		".private", ".key", ".token", ".dump", ".database", ".db", ".logs", ".debug",
		".pem", ".crt", ".cert", ".pfx", ".p12", ".bak", ".swp", ".tmp", ".cache",
		".session", ".sqlite", ".sqlite3", ".mdf", ".ldf", ".rdb", ".ldb", ".log",
		".old", ".orig", ".example", ".sample", ".test", ".spec", ".out", ".core",
	}
	forbiddenSubstrings = []string{
		"etc/passwd",
		"boot.ini",
		"win.ini",
		"web.config",
		"dockerfile",
		"credentials",
		"secrets",
		"backup",
		"hidden",
	}
)

// IsSuspiciousPath detects potential path traversal attacks
func IsSuspiciousPath(path string) bool {
	lower := strings.ToLower(path)
	decoded, _ := url.PathUnescape(lower)

	// Check for directory traversal
	if strings.Contains(lower, "../") || strings.Contains(decoded, "../") ||
		strings.Contains(lower, "..\\") || strings.Contains(decoded, "..\\") {
		return true
	}

	// Check for encoded traversal
	if strings.Contains(lower, "%2e%2e%2f") || strings.Contains(lower, "%2e%2e%5c") ||
		strings.Contains(decoded, "%2e%2e%2f") || strings.Contains(decoded, "%2e%2e%5c") {
		return true
	}

	// Check dangerous extensions
	for _, ext := range forbiddenExtensions {
		if strings.HasSuffix(lower, ext) || strings.HasSuffix(decoded, ext) {
			return true
		}
	}

	// Check dangerous substrings
	for _, substr := range forbiddenSubstrings {
		if strings.Contains(lower, substr) || strings.Contains(decoded, substr) {
			return true
		}
	}

	return false
}

// GenerateToken creates random UUID token
func GenerateToken() (string, error) {
	id, err := uuid.NewRandom()
	if err != nil {
		return "", eris.Wrap(err, "token generation failed")
	}
	return id.String(), nil
}

// GenerateRandomBytes produces cryptographic random data
func GenerateRandomBytes(n uint32) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	return b, err
}

// Create32ByteKey ensures 32-byte key length
func Create32ByteKey(key []byte) []byte {
	if len(key) >= 32 {
		return key[:32]
	}
	padded := make([]byte, 32)
	copy(padded, key)
	return padded
}

// Data validation ------------------------------------------------------------

var phoneRegex = regexp.MustCompile(`^\+?(?:\d{1,4})?\d{7,14}$`)

// IsValidEmail validates email format
func IsValidEmail(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}

// IsValidPhoneNumber validates international phone format
func IsValidPhoneNumber(phone string) bool {
	return phoneRegex.MatchString(phone)
}

// IsValidURL validates http/https URLs
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

// ParseCoordinate safely parses float coordinates
func ParseCoordinate(value string) float64 {
	if value == "" {
		return 0.0
	}
	coord, _ := strconv.ParseFloat(value, 64)
	return coord
}

// UUID handling --------------------------------------------------------------

// ParseUUID safely parses UUID from string pointer
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

// Math utilities -------------------------------------------------------------

// GenerateRandomDigits creates n-digit random number
func GenerateRandomDigits(size int) (int, error) {
	switch {
	case size > 8:
		return 0, errors.New("digit size exceeds maximum (8)")
	case size <= 0:
		return 0, errors.New("digit size must be positive")
	}

	min := big.NewInt(0).Exp(big.NewInt(10), big.NewInt(int64(size-1)), nil)
	max := big.NewInt(0).Exp(big.NewInt(10), big.NewInt(int64(size)), nil)
	max.Sub(max, big.NewInt(1))

	n, err := rand.Int(rand.Reader, max.Sub(max, min))
	if err != nil {
		return 0, err
	}
	return int(n.Add(n, min).Int64()), nil
}

// Concurrency helpers --------------------------------------------------------

// Retry executes operation with backoff
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

// UI helpers -----------------------------------------------------------------

// PrintASCIIArt renders colored horizon logo
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

func UuidPtrEqual(a, b *uuid.UUID) bool {
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
	minLen := len(slice1)
	if len(slice2) < minLen {
		minLen = len(slice2)
	}

	result := make([]struct {
		First  T
		Second U
	}, minLen)
	for i := 0; i < minLen; i++ {
		result[i] = struct {
			First  T
			Second U
		}{slice1[i], slice2[i]}
	}
	return result
}

// Helper function to check if filename has an extension
func HasFileExtension(filename string) bool {
	return strings.Contains(filename, ".") && !strings.HasSuffix(filename, ".")
}

// Helper function to get file extension from content type
func GetExtensionFromContentType(contentType string) string {
	// Comprehensive content type to extension mappings based on MDN Web Docs
	contentTypeMap := map[string]string{
		// Audio formats (original + additions)
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

		// Video formats (original + additions)
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

		// Image formats (original + additions)
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

		// Font formats (original + additions)
		"font/otf":        ".otf",
		"font/ttf":        ".ttf",
		"font/woff":       ".woff",
		"font/woff2":      ".woff2",
		"font/collection": ".ttc", // Added: TrueType Collection
		"font/sfnt":       ".ttf", // Added: Generic SFNT font (often TrueType)

		// Application formats (original + additions)
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

		// Text formats (original + additions)
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

		// Additional categories
		// 3D models and graphics
		"model/gltf+json":   ".gltf", // Added: glTF JSON
		"model/gltf-binary": ".glb",  // Added: glTF Binary
		"model/obj":         ".obj",  // Added: Wavefront OBJ
		"model/stl":         ".stl",  // Added: Stereolithography

		// Subtitles and captions
		"text/vtt":             ".vtt", // Added: WebVTT
		"application/x-subrip": ".srt", // Added: SubRip

		// Executables and binaries
		"application/x-msdownload":      ".exe", // Added: Windows Executable
		"application/x-shockwave-flash": ".swf", // Added: Shockwave Flash

		// Database and data
		"application/sql": ".sql", // Added: SQL script

		// Web-related
		"application/rss+xml":  ".rss",  // Added: RSS Feed
		"application/atom+xml": ".atom", // Added: Atom Feed
		"application/wasm":     ".wasm", // Added: WebAssembly
	}
	// Clean up content type (remove charset, etc.)
	cleanContentType := strings.Split(contentType, ";")[0]
	cleanContentType = strings.TrimSpace(cleanContentType)

	if ext, exists := contentTypeMap[cleanContentType]; exists {
		return ext
	}
	return ""
}
