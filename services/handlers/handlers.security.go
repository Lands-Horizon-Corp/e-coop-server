package handlers

import (
	"errors"
	"fmt"
	"net/url"
	"path/filepath"
	"regexp"
	"strings"
)

// Security patterns - compiled once for performance
var (
	// Core security regex patterns
	securityRegexes = map[string]*regexp.Regexp{
		"path_traversal":    regexp.MustCompile(`(?i)(\.\./|\.\.\\|%2e%2e%2f|%2e%2e%5c|\.\.%2f|\.\.%5c)`),
		"null_byte":         regexp.MustCompile(`(?i)(%00|\\0|\x00)`),
		"control_chars":     regexp.MustCompile(`[\x00-\x1f\x7f-\x9f]`),
		"unicode_dot":       regexp.MustCompile(`(?i)(\\u002e|\\u2024|\\u2025|\\u2026)`),
		"hex_slash":         regexp.MustCompile(`(?i)(\\x2f|\\x5c|%2f|%5c)`),
		"excessive_dots":    regexp.MustCompile(`\.{3,}`),
		"script_injection":  regexp.MustCompile(`(?i)(<script|javascript:|vbscript:|data:|about:|onload=|onerror=|onclick=|onmouseover=|eval\(|expression\()`),
		"sql_injection":     regexp.MustCompile(`(?i)(union\s+select|drop\s+table|delete\s+from|insert\s+into|update\s+set|\bor\s+1\s*=\s*1|\band\s+1\s*=\s*1|information_schema|concat\(|char\(|0x[0-9a-f]+|benchmark\(|sleep\(|waitfor\s+delay)`),
		"command_injection": regexp.MustCompile(`(?i)(\||\&\&|\|\||;|` + "`" + `|\$\(|\$\{|<\(|>\(|nc\s|netcat|curl\s|wget\s|/bin/|/usr/bin/|cmd\.exe|powershell|bash\s|sh\s)`),
		"ldap_injection":    regexp.MustCompile(`(?i)(\*\)|\)\(|\|\(|\&\(|objectclass=|cn=|uid=|ou=|dc=)`),
		"xpath_injection":   regexp.MustCompile(`(?i)(or\s+1\s*=\s*1|and\s+1\s*=\s*1|\[|\]|/\*|\*/|count\(|string\(|substring\()`),
		"xml_injection":     regexp.MustCompile(`(?i)(<!DOCTYPE|<!ENTITY|&[a-z]+;|SYSTEM\s|PUBLIC\s|file://|ftp://|gopher://|dict://)`),
	}

	// Dangerous file extensions
	dangerousExtensions = []string{
		".exe", ".bat", ".sh", ".cmd", ".php", ".asp", ".aspx", ".jsp", ".cgi", ".pl", ".py", ".rb",
		".js", ".vbs", ".ps1", ".env", ".yaml", ".yml", ".ini", ".config", ".conf", ".xml", ".json",
		".git", ".ssh", ".key", ".pem", ".cert", ".crt", ".backup", ".dump", ".db", ".sqlite",
	}

	// Critical system paths
	systemPaths = []string{
		"etc/passwd", "etc/shadow", "windows/system32", ".ssh/", ".aws/", ".env",
		"web.config", "application.properties", "database.yml", "credentials", "secrets",
	}

	// Path traversal patterns
	traversalPatterns = []string{
		"../", "..\\", "%2e%2e%2f", "%2e%2e%5c", "..%2f", "..%5c",
		"\\..\\/", "....//", "..;/", "..%00/", "/../", "/..", "%2f..",
	}
)

// IsSuspiciousPath validates path for security threats with API-aware logic
func IsSuspiciousPath(path string) bool {
	if path == "" || len(path) > 4096 {
		return len(path) > 4096 // Return true only if too long
	}

	isAPI := strings.HasPrefix(path, "/api/")

	// Always check injection patterns
	if hasInjectionPattern(path) {
		return true
	}

	// API paths: check traversal only
	if isAPI {
		return hasPathTraversal(path) || hasSystemAccess(path)
	}

	return filepath.IsAbs(path) || hasPathTraversal(path) ||
		hasDangerousContent(path) || hasSystemAccess(path)
}

// hasInjectionPattern checks for code injection patterns
func hasInjectionPattern(path string) bool {
	injectionTypes := []string{"script_injection", "sql_injection", "command_injection",
		"ldap_injection", "xpath_injection", "xml_injection",
		"null_byte", "control_chars"}

	for _, pattern := range injectionTypes {
		if regex := securityRegexes[pattern]; regex != nil && regex.MatchString(path) {
			return true
		}
	}
	return false
}

// hasPathTraversal checks for directory traversal attempts
func hasPathTraversal(path string) bool {
	// Quick regex check
	if securityRegexes["path_traversal"].MatchString(path) ||
		securityRegexes["unicode_dot"].MatchString(path) ||
		securityRegexes["hex_slash"].MatchString(path) ||
		securityRegexes["excessive_dots"].MatchString(path) {
		return true
	}

	// Check decoded variants
	variants := getPathVariants(path)
	for _, variant := range variants {
		for _, pattern := range traversalPatterns {
			if strings.Contains(variant, pattern) {
				return true
			}
		}
	}
	return false
}

// hasSystemAccess checks for system file access attempts
func hasSystemAccess(path string) bool {
	variants := getPathVariants(path)
	for _, variant := range variants {
		for _, sysPath := range systemPaths {
			if strings.Contains(variant, sysPath) {
				return true
			}
		}
	}
	return false
}

// hasDangerousContent checks for dangerous extensions and content
func hasDangerousContent(path string) bool {
	lower := strings.ToLower(path)
	for _, ext := range dangerousExtensions {
		if strings.HasSuffix(lower, ext) {
			return true
		}
	}
	return false
}

// getPathVariants returns normalized path variants for checking
func getPathVariants(path string) []string {
	variants := []string{strings.ToLower(path)}

	// URL decode up to 3 times
	decoded := path
	for i := 0; i < 3 && len(decoded) <= 4096; i++ {
		if newDecoded, err := url.PathUnescape(decoded); err == nil && newDecoded != decoded {
			decoded = newDecoded
			variants = append(variants, strings.ToLower(decoded))
		} else {
			break
		}
	}
	return variants
}

// IsSecurePath validates if path is safe for file operations
func IsSecurePath(path string) bool {
	return path != "" &&
		len(path) <= 1024 &&
		!IsSuspiciousPath(path) &&
		!filepath.IsAbs(path) &&
		!hasTraversalElements(path)
}

// hasTraversalElements checks for directory traversal in clean path
func hasTraversalElements(path string) bool {
	clean := filepath.Clean(path)
	return strings.HasPrefix(clean, "../") ||
		clean == ".." ||
		strings.Contains(clean, "/../")
}

// SafePathJoin securely joins path components
func SafePathJoin(base string, elements ...string) (string, error) {
	if base == "" {
		return "", errors.New("base path cannot be empty")
	}
	for i, elem := range elements {
		if !IsSecurePath(elem) {
			return "", fmt.Errorf("invalid path element at index %d: %s", i, elem)
		}
	}
	result := filepath.Join(append([]string{base}, elements...)...)
	if absBase, err := filepath.Abs(base); err == nil {
		if absResult, err := filepath.Abs(result); err == nil {
			if !strings.HasPrefix(absResult, absBase) {
				return "", errors.New("path escapes base directory")
			}
		}
	}
	return result, nil
}

func IsSecurityThreat(input string) (bool, string) {
	threats := map[string]string{
		"script_injection":  "SCRIPT_INJECTION",
		"sql_injection":     "SQL_INJECTION",
		"command_injection": "COMMAND_INJECTION",
		"ldap_injection":    "LDAP_INJECTION",
		"xpath_injection":   "XPATH_INJECTION",
		"xml_injection":     "XML_INJECTION",
		"path_traversal":    "PATH_TRAVERSAL",
	}
	for pattern, threatType := range threats {
		if regex := securityRegexes[pattern]; regex != nil && regex.MatchString(input) {
			return true, threatType
		}
	}
	return false, ""
}

var (
	PathTraversalRegex    = securityRegexes["path_traversal"]
	ScriptInjectionRegex  = securityRegexes["script_injection"]
	SqlInjectionRegex     = securityRegexes["sql_injection"]
	CommandInjectionRegex = securityRegexes["command_injection"]
)
