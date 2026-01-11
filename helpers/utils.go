package helpers

import (
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"net/mail"
	"os"
	"reflect"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/microcosm-cc/bluemonday"
	"github.com/rotisserie/eris"
)

func GenerateToken() (string, error) {
	id, err := uuid.NewRandom()
	if err != nil {
		return "", eris.Wrap(err, "token generation failed")
	}
	return id.String(), nil
}

func GenerateDigitCode(digits int) (string, error) {
	if digits <= 0 {
		return "", eris.New("digits must be greater than 0")
	}
	n, err := rand.Int(rand.Reader, new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(digits)), nil))
	if err != nil {
		return "", eris.Wrap(err, "digit code generation failed")
	}
	return fmt.Sprintf(fmt.Sprintf("%%0%dd", digits), n.Int64()), nil
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

func IsValidEmail(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}

func Sanitize(input string) string {
	return bluemonday.UGCPolicy().Sanitize(strings.TrimSpace(input))
}

func IsValidPhoneNumber(phone string) bool {
	return regexp.MustCompile(`^\+?(?:\d{1,4})?\d{7,14}$`).MatchString(phone)
}

func GetHost(c echo.Context) string {
	if host := c.Request().Host; host != "" {
		return strings.TrimSpace(host)
	}
	if host := c.Request().Header.Get("X-Forwarded-Host"); host != "" {
		hosts := strings.Split(host, ",")
		if len(hosts) > 0 {
			return strings.TrimSpace(hosts[0])
		}
	}
	if host := c.Request().Header.Get("X-Original-Host"); host != "" {
		return strings.TrimSpace(host)
	}
	if origin := c.Request().Header.Get("Origin"); origin != "" {
		hostname := strings.TrimPrefix(origin, "https://")
		hostname = strings.TrimPrefix(hostname, "http://")
		return strings.TrimSpace(hostname)
	}
	if referer := c.Request().Header.Get("Referer"); referer != "" {
		hostname := strings.TrimPrefix(referer, "https://")
		hostname = strings.TrimPrefix(hostname, "http://")
		if idx := strings.Index(hostname, "/"); idx != -1 {
			hostname = hostname[:idx]
		}
		return strings.TrimSpace(hostname)
	}
	return ""
}

func GetUserAgent(c echo.Context) string {
	userAgent := c.Request().UserAgent()
	if len(userAgent) > 512 {
		userAgent = userAgent[:512]
	}
	userAgent = strings.ReplaceAll(userAgent, "\x00", "")
	userAgent = strings.Map(func(r rune) rune {
		if r < 32 || r == 127 {
			return -1
		}
		return r
	}, userAgent)

	return strings.TrimSpace(userAgent)
}

func GetPath(c echo.Context) string {
	path := c.Request().URL.Path
	if len(path) > 2048 {
		path = path[:2048]
	}
	path = strings.ReplaceAll(path, "\x00", "")
	path = strings.Map(func(r rune) rune {
		if r < 32 || r == 127 {
			return -1
		}
		return r
	}, path)
	if path == "" || path[0] != '/' {
		path = "/" + path
	}
	for strings.Contains(path, "//") {
		path = strings.ReplaceAll(path, "//", "/")
	}
	return strings.TrimSpace(path)
}

func GetClientIP(c echo.Context) string {
	if ip := c.Request().Header.Get("Fly-Client-IP"); ip != "" {
		return strings.TrimSpace(ip)
	}
	if ip := c.Request().Header.Get("X-Forwarded-For"); ip != "" {
		ips := strings.Split(ip, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}
	if ip := c.Request().Header.Get("X-Real-IP"); ip != "" {
		return strings.TrimSpace(ip)
	}
	if ip := c.Request().Header.Get("CF-Connecting-IP"); ip != "" {
		return strings.TrimSpace(ip)
	}
	return c.RealIP()
}

func IsSuspicious(path string) bool {
	path = strings.ToLower(path)
	for _, p := range suspiciousPaths {
		if strings.Contains(path, strings.ToLower(p)) {
			return true
		}
	}
	return false
}

// ExtractInterfaceName returns the type name of any value passed in.
func ExtractInterfaceName(i any) string {
	if i == nil {
		return "<nil>"
	}
	t := reflect.TypeOf(i)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t.Name()
}
