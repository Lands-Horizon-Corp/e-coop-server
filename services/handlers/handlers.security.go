package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

var (
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

		"nosql_injection":    regexp.MustCompile(`(?i)(\$where|\$regex|\$ne|\$gt|\$lt|\$in|\$nin|\$or|\$and|\$project|\$group|\$match|\$skip|\$limit|\$unwind)`),
		"ssrf_protocols":     regexp.MustCompile(`(?i)(file://|ftp://|gopher://|dict://|jar://|ldap://|tftp://|http://0\.0\.0\.0|http://127\.0\.0\.1)`),
		"log4shell":          regexp.MustCompile(`(?i)(\$\{jndi:(ldap|rmi|dns|iiop|corba|nds):[^{}]+\})`),
		"graphql_injection":  regexp.MustCompile(`(?i)(__schema|__type|__typename|query\{|\*/|\*/)`),
		"deserialization":    regexp.MustCompile(`(?i)(java\.util\.HashMap|ysoserial|commons-collections|Serializable\.class)`),
		"template_injection": regexp.MustCompile(`(?i)(\{\{|\}\}|\$\%\{|\#\{|\$\{|\{\%|\}\%|\{\{7\*7\}\}|\{\%7\*7\%\})`),
	}

	dangerousExtensions = []string{
		".exe", ".bat", ".sh", ".cmd", ".php", ".asp", ".aspx", ".jsp", ".cgi", ".pl", ".py", ".rb",
		".js", ".vbs", ".ps1", ".env", ".yaml", ".yml", ".ini", ".config", ".conf", ".xml", ".json",
		".git", ".ssh", ".key", ".pem", ".cert", ".crt", ".backup", ".dump", ".db", ".sqlite",
	}

	systemPaths = []string{
		"etc/passwd", "etc/shadow", "windows/system32", ".ssh/", ".aws/", ".env",
		"web.config", "application.properties", "database.yml", "credentials", "secrets",
	}

	traversalPatterns = []string{
		"../", "..\\", "%2e%2e%2f", "%2e%2e%5c", "..%2f", "..%5c",
		"\\..\\/", "....//", "..;/", "..%00/", "/../", "/..", "%2f..",
	}
)

func IsSuspiciousPath(path string) bool {
	if path == "" || len(path) > 4096 {
		return len(path) > 4096
	}
	isAPI := strings.HasPrefix(path, "/web/api/") || strings.HasPrefix(path, "/mobile/api/")
	if hasInjectionPattern(path) {
		return true
	}
	if isAPI {
		return hasPathTraversal(path) || hasSystemAccess(path)
	}

	return filepath.IsAbs(path) || hasPathTraversal(path) ||
		hasDangerousContent(path) || hasSystemAccess(path)
}

func hasInjectionPattern(path string) bool {
	injectionTypes := []string{
		"script_injection", "sql_injection", "command_injection",
		"ldap_injection", "xpath_injection", "xml_injection",
		"nosql_injection", "ssrf_protocols", "log4shell",
		"graphql_injection", "deserialization", "template_injection",
		"null_byte", "control_chars",
	}

	for _, pattern := range injectionTypes {
		if regex := securityRegexes[pattern]; regex != nil && regex.MatchString(path) {
			return true
		}
	}
	return false
}

func hasPathTraversal(path string) bool {
	if securityRegexes["path_traversal"].MatchString(path) ||
		securityRegexes["unicode_dot"].MatchString(path) ||
		securityRegexes["hex_slash"].MatchString(path) ||
		securityRegexes["excessive_dots"].MatchString(path) {
		return true
	}

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

func hasDangerousContent(path string) bool {
	lower := strings.ToLower(path)
	for _, ext := range dangerousExtensions {
		if strings.HasSuffix(lower, ext) {
			return true
		}
	}
	return false
}

func getPathVariants(path string) []string {
	variants := []string{strings.ToLower(path)}
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

func IsSecurePath(path string) bool {
	return path != "" &&
		len(path) <= 1024 &&
		!IsSuspiciousPath(path) &&
		!filepath.IsAbs(path) &&
		!hasTraversalElements(path)
}

func hasTraversalElements(path string) bool {
	clean := filepath.Clean(path)
	return strings.HasPrefix(clean, "../") ||
		clean == ".." ||
		strings.Contains(clean, "/../")
}

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

type ThreatInfo struct {
	IsThreat    bool
	ThreatType  string
	Severity    int
	Pattern     string
	Description string
}

func IsSecurityThreat(input string) (bool, string) {
	threats := map[string]string{
		"script_injection":   "SCRIPT_INJECTION",
		"sql_injection":      "SQL_INJECTION",
		"command_injection":  "COMMAND_INJECTION",
		"ldap_injection":     "LDAP_INJECTION",
		"xpath_injection":    "XPATH_INJECTION",
		"xml_injection":      "XML_INJECTION",
		"nosql_injection":    "NOSQL_INJECTION",
		"ssrf_protocols":     "SSRF_ATTEMPT",
		"log4shell":          "LOG4SHELL_EXPLOIT",
		"graphql_injection":  "GRAPHQL_INJECTION",
		"deserialization":    "DESERIALIZATION_ATTACK",
		"template_injection": "TEMPLATE_INJECTION",
		"path_traversal":     "PATH_TRAVERSAL",
	}

	for pattern, threatType := range threats {
		if regex := securityRegexes[pattern]; regex != nil && regex.MatchString(input) {
			return true, threatType
		}
	}
	return false, ""
}

func AnalyzeThreat(input string) ThreatInfo {
	threatSeverity := map[string]int{
		"LOG4SHELL_EXPLOIT":      5, // Critical
		"COMMAND_INJECTION":      5, // Critical
		"DESERIALIZATION_ATTACK": 5, // Critical
		"SQL_INJECTION":          4, // High
		"SCRIPT_INJECTION":       4, // High
		"SSRF_ATTEMPT":           4, // High
		"NOSQL_INJECTION":        4, // High
		"PATH_TRAVERSAL":         3, // Medium
		"XML_INJECTION":          3, // Medium
		"TEMPLATE_INJECTION":     3, // Medium
		"GRAPHQL_INJECTION":      2, // Low
		"LDAP_INJECTION":         2, // Low
		"XPATH_INJECTION":        2, // Low
	}

	threatDescriptions := map[string]string{
		"LOG4SHELL_EXPLOIT":      "Critical Log4j JNDI lookup vulnerability",
		"COMMAND_INJECTION":      "Operating system command execution attempt",
		"DESERIALIZATION_ATTACK": "Malicious object deserialization attempt",
		"SQL_INJECTION":          "Database query manipulation attempt",
		"SCRIPT_INJECTION":       "Client-side script injection attempt",
		"SSRF_ATTEMPT":           "Server-side request forgery attempt",
		"NOSQL_INJECTION":        "NoSQL database injection attempt",
		"PATH_TRAVERSAL":         "Directory traversal attempt",
		"XML_INJECTION":          "XML external entity injection attempt",
		"TEMPLATE_INJECTION":     "Server-side template injection attempt",
		"GRAPHQL_INJECTION":      "GraphQL introspection or injection attempt",
		"LDAP_INJECTION":         "LDAP query injection attempt",
		"XPATH_INJECTION":        "XPath query injection attempt",
	}

	isThreat, threatType := IsSecurityThreat(input)
	if !isThreat {
		return ThreatInfo{IsThreat: false}
	}

	return ThreatInfo{
		IsThreat:    true,
		ThreatType:  threatType,
		Severity:    threatSeverity[threatType],
		Description: threatDescriptions[threatType],
	}
}

var (
	PathTraversalRegex     = securityRegexes["path_traversal"]
	ScriptInjectionRegex   = securityRegexes["script_injection"]
	SqlInjectionRegex      = securityRegexes["sql_injection"]
	CommandInjectionRegex  = securityRegexes["command_injection"]
	NoSqlInjectionRegex    = securityRegexes["nosql_injection"]
	SSRFProtocolsRegex     = securityRegexes["ssrf_protocols"]
	Log4ShellRegex         = securityRegexes["log4shell"]
	GraphQLInjectionRegex  = securityRegexes["graphql_injection"]
	DeserializationRegex   = securityRegexes["deserialization"]
	TemplateInjectionRegex = securityRegexes["template_injection"]
)

func ValidateRequest(method, path, queryString, userAgent, body string) (bool, []ThreatInfo) {
	var threats []ThreatInfo

	if threat := AnalyzeThreat(path); threat.IsThreat {
		threats = append(threats, threat)
	}

	if queryString != "" {
		if threat := AnalyzeThreat(queryString); threat.IsThreat {
			threats = append(threats, threat)
		}
	}

	if userAgent != "" {
		if threat := AnalyzeThreat(userAgent); threat.IsThreat {
			threats = append(threats, threat)
		}
	}

	if body != "" && len(body) < 1048576 { // Only check bodies under 1MB
		bodyThreats := validateRequestBody(body)
		threats = append(threats, bodyThreats...)
	}

	return len(threats) == 0, threats
}

func ValidateRequestWithHeaders(method, path, queryString, userAgent, body string, headers map[string][]string) (bool, []ThreatInfo) {
	var threats []ThreatInfo

	_, requestThreats := ValidateRequest(method, path, queryString, userAgent, body)
	threats = append(threats, requestThreats...)

	headerThreats := validateHeaders(headers)
	threats = append(threats, headerThreats...)

	return len(threats) == 0, threats
}

func validateRequestBody(body string) []ThreatInfo {
	var threats []ThreatInfo

	if threat := AnalyzeThreat(body); threat.IsThreat {
		threats = append(threats, threat)
	}

	if strings.HasPrefix(strings.TrimSpace(body), "{") {
		var jsonData map[string]any
		if err := json.Unmarshal([]byte(body), &jsonData); err == nil {
			jsonThreats := validateJSONRecursive(jsonData, "")
			threats = append(threats, jsonThreats...)
		}
	}

	return threats
}

func validateJSONRecursive(data any, path string) []ThreatInfo {
	var threats []ThreatInfo

	switch v := data.(type) {
	case string:
		if threat := AnalyzeThreat(v); threat.IsThreat {
			threats = append(threats, threat)
		}
	case map[string]any:
		for key, value := range v {
			newPath := path + "." + key
			if threat := AnalyzeThreat(key); threat.IsThreat {
				threats = append(threats, threat)
			}
			nested := validateJSONRecursive(value, newPath)
			threats = append(threats, nested...)
		}
	case []any:
		for i, item := range v {
			newPath := fmt.Sprintf("%s[%d]", path, i)
			nested := validateJSONRecursive(item, newPath)
			threats = append(threats, nested...)
		}
	}

	return threats
}

func validateHeaders(headers map[string][]string) []ThreatInfo {
	var threats []ThreatInfo

	criticalHeaders := []string{
		"Authorization",
		"X-Forwarded-For",
		"X-Real-IP",
		"X-Forwarded-Host",
		"Host",
		"Referer",
		"Origin",
		"Cookie",
	}

	for _, headerName := range criticalHeaders {
		if values, exists := headers[headerName]; exists {
			for _, value := range values {
				if threat := AnalyzeThreat(value); threat.IsThreat {
					threats = append(threats, threat)
				}
			}
		}
	}

	return threats
}

func IsWhitelistedPath(path string) bool {
	whitelistedPaths := []string{
		"/health",
		"/metrics",
		"/api/routes",
		"/favicon.ico",
		"/.well-known/",
	}

	for _, whitelist := range whitelistedPaths {
		if strings.HasPrefix(path, whitelist) {
			return true
		}
	}

	if strings.HasPrefix(path, "/api/") {
		validAPIPattern := regexp.MustCompile(`^/api(/v\d+)?/[a-zA-Z0-9_-]+(/[a-zA-Z0-9_-]+)*(/\d+)?/?$`)
		return validAPIPattern.MatchString(path)
	}

	return false
}

func GetThreatScore(threats []ThreatInfo) int {
	totalScore := 0
	for _, threat := range threats {
		totalScore += threat.Severity
	}
	return totalScore
}

func ShouldBlockRequest(threats []ThreatInfo) bool {
	for _, threat := range threats {
		if threat.Severity >= 5 {
			return true
		}
	}

	return GetThreatScore(threats) >= 8
}

func ValidateRequestWithTimeout(method, path, queryString, userAgent, body string, headers map[string][]string, timeout time.Duration) (bool, []ThreatInfo, error) {
	done := make(chan struct{})
	var isSecure bool
	var threats []ThreatInfo

	go func() {
		defer close(done)
		isSecure, threats = ValidateRequestWithHeaders(method, path, queryString, userAgent, body, headers)
	}()

	select {
	case <-done:
		return isSecure, threats, nil
	case <-time.After(timeout):
		return false, []ThreatInfo{{IsThreat: true, ThreatType: "VALIDATION_TIMEOUT", Severity: 3, Description: "Security validation timed out"}}, errors.New("validation timeout")
	}
}

func IsContextuallyAllowed(threat ThreatInfo, isAuthenticated bool, userRole string, path string) bool {
	if isAuthenticated && userRole == "admin" {
		if threat.ThreatType == "NOSQL_INJECTION" || threat.ThreatType == "GRAPHQL_INJECTION" {
			if strings.HasPrefix(path, "/api/admin/") {
				return true
			}
		}
	}

	if threat.Severity >= 5 {
		return false
	}

	return false
}
