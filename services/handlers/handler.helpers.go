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

func (t *TemplateRenderer) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
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

// GetFreePort finds available TCP port
func GetFreePort() int {
	l, err := net.Listen("tcp", ":0")
	if err != nil {
		return 8123
	}
	defer l.Close()
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
		return "", fmt.Errorf("token generation failed: %w", err)
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
	return fmt.Errorf("after %d attempts: %w", maxAttempts, err)
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

	for _, line := range strings.Split(art, "\n") {
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
