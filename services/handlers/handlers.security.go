package handlers

import (
	"errors"
	"fmt"
	"net/url"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	forbiddenExtensions = []string{
		// Executable files
		".exe", ".bat", ".sh", ".cmd", ".com", ".scr", ".msi", ".deb", ".rpm",
		".dmg", ".pkg", ".app", ".run", ".bin", ".elf", ".so", ".dll", ".dylib",

		// Script files
		".php", ".asp", ".aspx", ".jsp", ".cgi", ".pl", ".py", ".rb", ".js", ".vbs",
		".ps1", ".psm1", ".psd1", ".ps1xml", ".clixml", ".wsf", ".wsc", ".hta",

		// Configuration and sensitive files
		".env", ".yaml", ".yml", ".ini", ".config", ".conf", ".xml", ".json",
		".toml", ".properties", ".settings", ".plist", ".reg", ".desktop",

		// Development files
		".go", ".c", ".cpp", ".h", ".hpp", ".cs", ".java", ".kt", ".swift",
		".rs", ".asm", ".s", ".makefile", ".cmake", ".gradle", ".maven",

		// Version control and Git
		".git", ".gitignore", ".gitconfig", ".gitmodules", ".gitattributes",
		".svn", ".hg", ".bzr", ".cvs",

		// Web server files
		".htaccess", ".htpasswd", ".htgroups", ".htdigest", ".apache", ".nginx",
		".iis", ".web", ".sitemap", ".robots", ".crossdomain",

		// Security and credentials
		".backup", ".secret", ".credential", ".password", ".private", ".key",
		".token", ".cert", ".crt", ".pem", ".p12", ".pfx", ".jks", ".keystore",
		".truststore", ".gpg", ".pgp", ".asc", ".sig", ".ssh", ".ppk",

		// Database files
		".dump", ".database", ".db", ".sqlite", ".sqlite3", ".mdf", ".ldf",
		".rdb", ".ldb", ".accdb", ".mdb", ".dbf", ".frm", ".ibd", ".opt",
		".cnf", ".my", ".ora", ".trc", ".aud",

		// Log and temporary files
		".log", ".logs", ".debug", ".trace", ".audit", ".tmp", ".temp",
		".cache", ".session", ".pid", ".lock", ".swp", ".swo", ".backup",
		".old", ".orig", ".bak", ".~", ".autosave",

		// Documentation that might contain sensitive info
		".example", ".sample", ".test", ".spec", ".readme", ".todo", ".notes",

		// System files
		".out", ".core", ".crash", ".minidump", ".stackdump", ".heapdump",

		// Container and deployment
		".dockerfile", ".dockerignore", ".docker", ".k8s", ".kubernetes",
		".helm", ".terraform", ".vagrant", ".ansible",

		// IDE and editor files
		".vscode", ".idea", ".eclipse", ".sublime", ".atom", ".vim", ".emacs",
		".project", ".classpath", ".factorypath", ".prefs",

		// Package managers
		".npm", ".node_modules", ".yarn", ".bower", ".composer", ".pip",
		".cargo", ".maven", ".gradle", ".sbt",
	}
	forbiddenSubstrings = []string{
		// Unix/Linux system files
		"etc/passwd", "etc/shadow", "etc/group", "etc/sudoers", "etc/hosts",
		"etc/hostname", "etc/resolv.conf", "etc/crontab", "etc/fstab",
		"etc/ssh/", "etc/ssl/", "etc/pki/", "etc/security/", "etc/audit/",
		"proc/", "sys/", "dev/", "run/", "var/log/", "var/spool/",
		"boot/", "usr/bin/", "usr/sbin/", "bin/", "sbin/",

		// Windows system files
		"boot.ini", "win.ini", "system.ini", "autoexec.bat", "config.sys",
		"ntuser.dat", "sam", "security", "software", "system",
		"windows/system32/", "windows/syswow64/", "windows/temp/",
		"users/", "documents and settings/", "program files/", "programdata/",
		"appdata/", "temp/", "tmp/",

		// Web server configurations
		"web.config", "httpd.conf", "apache2.conf", "nginx.conf",
		".htaccess", ".htpasswd", "server.xml", "context.xml",
		"tomcat-users.xml", "catalina.properties",

		// Application configs
		"application.properties", "application.yml", "application.yaml",
		"config.json", "settings.json", "app.config", "appsettings.json",

		// Container files
		"dockerfile", "docker-compose", ".dockerignore", "k8s/", "kubernetes/",

		// Security-related
		"credentials", "secrets", "passwords", "tokens", "keys", "certs",
		"certificates", "private", "confidential", "restricted",
		"authorized_keys", "known_hosts", "identity", "id_rsa", "id_dsa",
		"id_ecdsa", "id_ed25519", ".ssh/", ".gnupg/", ".gpg/",

		// Backup and temporary
		"backup", "backups", "dump", "dumps", "archive", "archives",
		"temp", "tmp", "cache", "logs", "log", "trace", "debug",
		"crash", "core", "minidump", "heapdump",

		// Hidden directories
		".git/", ".svn/", ".hg/", ".bzr/", ".cvs/",
		".vscode/", ".idea/", ".eclipse/", ".settings/",
		".env", ".aws/", ".azure/", ".gcp/", ".config/",
		".docker/", ".kube/", ".helm/", ".terraform/",

		// Database-related
		"database", "databases", "db", "sql", "mysql", "postgresql",
		"oracle", "mongodb", "redis", "elasticsearch", "cassandra",
		"sqlite", "mariadb", "mssql", "sqlserver",

		// Development artifacts
		"node_modules/", "vendor/", "packages/", "lib/", "libs/",
		"target/", "build/", "dist/", "out/", "bin/", "obj/",
		".next/", ".nuxt/", ".output/", "coverage/",

		// Version control internals
		"objects/", "refs/", "hooks/", "info/", "logs/",
		"branches/", "remotes/", "tags/",

		// System processes and runtime
		"history", "bashrc", "profile", "zshrc", "vimrc", "tmux",
		"screenrc", "inputrc", "xinitrc", "xsession",

		// Application-specific sensitive areas
		"admin/", "administrator/", "root/", "superuser/", "sa/",
		"test/", "testing/", "debug/", "staging/", "development/",
		"internal/", "private/", "confidential/", "restricted/",
		"secure/", "protected/", "hidden/", "secret/",

		// Cloud and infrastructure
		"metadata", "userdata", "cloud-init", "terraform.tfstate",
		"ansible/", "puppet/", "chef/", "saltstack/",
		".aws/credentials", ".aws/config", "gcloud/", "azure/",

		// Monitoring and metrics
		"prometheus/", "grafana/", "elasticsearch/", "kibana/",
		"logstash/", "fluentd/", "splunk/", "datadog/",

		// Security tools
		"nessus/", "nmap/", "metasploit/", "burp/", "owasp/",
		"nikto/", "sqlmap/", "hydra/", "john/", "hashcat/",

		// Session and authentication
		"session", "sessions", "auth", "oauth", "saml", "jwt",
		"cookies", "tokens", "tickets", "principals",

		// Email and communication
		"mail/", "postfix/", "sendmail/", "dovecot/", "courier/",
		"exchange/", "outlook/", "thunderbird/",

		// Virtualization
		"vmware/", "virtualbox/", "qemu/", "kvm/", "xen/",
		"hyper-v/", "docker/", "lxc/", "podman/",
	}

	// Export regex patterns for use in other packages
	PathTraversalRegex   = regexp.MustCompile(`(?i)(\.\./|\.\.\\|%2e%2e%2f|%2e%2e%5c|\.\.%2f|\.\.%5c)`)
	NullByteRegex        = regexp.MustCompile(`(?i)(%00|\\0|\x00)`)
	ControlCharsRegex    = regexp.MustCompile(`[\x00-\x1f\x7f-\x9f]`)
	UnicodeDotRegex      = regexp.MustCompile(`(?i)(\\u002e|\\u2024|\\u2025|\\u2026)`)
	HexEncodedSlashRegex = regexp.MustCompile(`(?i)(\\x2f|\\x5c|%2f|%5c)`)
	ExcessiveDotsRegex   = regexp.MustCompile(`\.{3,}`)

	// Security injection patterns (now properly exported and used)
	ScriptInjectionRegex  = regexp.MustCompile(`(?i)(<script|javascript:|vbscript:|data:|about:|onload=|onerror=|onclick=|onmouseover=|eval\(|expression\()`)
	SqlInjectionRegex     = regexp.MustCompile(`(?i)(union\s+select|drop\s+table|delete\s+from|insert\s+into|update\s+set|\bor\s+1\s*=\s*1|\band\s+1\s*=\s*1|information_schema|concat\(|char\(|0x[0-9a-f]+|benchmark\(|sleep\(|waitfor\s+delay)`)
	CommandInjectionRegex = regexp.MustCompile(`(?i)(\||\&\&|\|\||;|` + "`" + `|\$\(|\$\{|<\(|>\(|nc\s|netcat|curl\s|wget\s|/bin/|/usr/bin/|cmd\.exe|powershell|bash\s|sh\s)`)

	// Additional security patterns
	LdapInjectionRegex  = regexp.MustCompile(`(?i)(\*\)|\)\(|\|\(|\&\(|objectclass=|cn=|uid=|ou=|dc=)`)
	XpathInjectionRegex = regexp.MustCompile(`(?i)(or\s+1\s*=\s*1|and\s+1\s*=\s*1|\[|\]|/\*|\*/|count\(|string\(|substring\()`)
	XmlInjectionRegex   = regexp.MustCompile(`(?i)(<!DOCTYPE|<!ENTITY|&[a-z]+;|SYSTEM\s|PUBLIC\s|file://|ftp://|gopher://|dict://)`)
)

// Enhanced IsSuspiciousPath that uses all the security patterns
func IsSuspiciousPath(path string) bool {
	if path == "" {
		return false
	}

	// Input length validation (prevent DOS)
	if len(path) > 4096 {
		return true
	}

	// Normalize and clean the path first
	cleanPath := filepath.Clean(path)

	// Reject absolute paths immediately
	if filepath.IsAbs(cleanPath) {
		return true
	}

	// Use all the regex patterns for comprehensive security checking
	securityPatterns := []*regexp.Regexp{
		PathTraversalRegex,
		NullByteRegex,
		ControlCharsRegex,
		UnicodeDotRegex,
		HexEncodedSlashRegex,
		ExcessiveDotsRegex,
		ScriptInjectionRegex,
		SqlInjectionRegex,
		CommandInjectionRegex,
		LdapInjectionRegex,
		XpathInjectionRegex,
		XmlInjectionRegex,
	}

	for _, pattern := range securityPatterns {
		if pattern.MatchString(path) {
			return true
		}
	}

	// Convert to lowercase for case-insensitive checks
	lower := strings.ToLower(path)

	// Safe URL decoding with iteration limit
	decoded := path
	for i := 0; i < 5; i++ { // Increased to 5 rounds but still limited
		if newDecoded, err := url.PathUnescape(decoded); err == nil && newDecoded != decoded {
			decoded = newDecoded
			// Validate decoded length
			if len(decoded) > 4096 {
				return true
			}
		} else {
			break
		}
	}
	decodedLower := strings.ToLower(decoded)

	// Also check the cleaned/normalized versions
	cleanLower := strings.ToLower(cleanPath)
	cleanDecoded, _ := url.PathUnescape(cleanLower)
	if len(cleanDecoded) > 4096 {
		return true
	}

	// All variants to check (limited set for performance)
	variants := []string{lower, decodedLower, cleanLower, cleanDecoded}

	// Enhanced directory traversal patterns (using Contains instead of regex)
	traversalPatterns := []string{
		"../", "..\\",
		"%2e%2e%2f", "%2e%2e%5c",
		"%252e%252e%252f", "%252e%252e%255c",
		"..%2f", "..%5c",
		"..%252f", "..%255c",
		"\\..\\/", "/\\..\\/",
		"....//", "....\\\\",
		"..;/", "..;\\",
		"..%00/", "..%00\\",
		"..\\x2f", "..\\x5c",
		"\u002e\u002e/", "\u002e\u002e\\",
		// Additional dangerous patterns
		"/../", "\\..\\",
		"/..", "\\..",
		"%2f..", "%5c..",
	}

	// Check all variants against all patterns
	for _, variant := range variants {
		// Early exit if variant is too long after processing
		if len(variant) > 4096 {
			return true
		}

		for _, pattern := range traversalPatterns {
			if strings.Contains(variant, pattern) {
				return true
			}
		}

		// Check if path tries to escape current directory
		if strings.HasPrefix(variant, "/") || strings.HasPrefix(variant, "\\") {
			return true
		}

		// Check for dangerous system paths (case-insensitive)
		dangerousPaths := []string{
			"/etc/", "/proc/", "/sys/", "/dev/", "/var/", "/tmp/",
			"/bin/", "/sbin/", "/usr/", "/lib/", "/root/", "/home/",
			"c:\\windows", "c:\\system32", "c:\\users", "c:\\program",
			"\\windows\\", "\\system32\\", "\\users\\", "\\program",
			".ssh/", ".aws/", ".config/", ".docker/", ".kube/",
		}

		for _, dangerous := range dangerousPaths {
			if strings.Contains(variant, dangerous) {
				return true
			}
		}
	}

	// Check dangerous extensions with all variants
	for _, ext := range forbiddenExtensions {
		for _, variant := range variants {
			if strings.HasSuffix(variant, ext) {
				return true
			}
		}
	}

	// Check dangerous substrings with all variants
	for _, substr := range forbiddenSubstrings {
		for _, variant := range variants {
			if strings.Contains(variant, substr) {
				return true
			}
		}
	}

	return false
}

// Additional helper function for strict path validation
func IsSecurePath(path string) bool {
	// Input validation
	if path == "" || len(path) > 1024 { // Stricter length limit
		return false
	}

	// Must not be suspicious
	if IsSuspiciousPath(path) {
		return false
	}

	// Must be relative
	if filepath.IsAbs(path) {
		return false
	}

	// Clean path should not escape current directory
	clean := filepath.Clean(path)
	if strings.HasPrefix(clean, "../") || clean == ".." || strings.Contains(clean, "/../") {
		return false
	}

	// Additional checks for Windows drive letters
	if len(clean) > 1 && clean[1] == ':' {
		return false
	}

	// Check for hidden files (Unix-style)
	parts := strings.SplitSeq(clean, string(filepath.Separator))
	for part := range parts {
		if strings.HasPrefix(part, ".") && part != "." && part != ".." {
			// Allow certain safe dotfiles
			safeDotFiles := []string{".gitkeep", ".htaccess"}
			isSafe := false
			for _, safe := range safeDotFiles {
				if part == safe {
					isSafe = true
					break
				}
			}
			if !isSafe {
				return false
			}
		}
	}

	// Ensure path doesn't contain invalid characters for filesystem
	invalidChars := []string{"|", "?", "*", "<", ">", ":", "\""}
	for _, char := range invalidChars {
		if strings.Contains(clean, char) {
			return false
		}
	}

	return true
}

// SafePathJoin safely joins path components and validates the result
func SafePathJoin(base string, elements ...string) (string, error) {
	if base == "" {
		return "", errors.New("base path cannot be empty")
	}

	// Validate all elements first
	for i, elem := range elements {
		if !IsSecurePath(elem) {
			return "", fmt.Errorf("invalid path element at index %d: %s", i, elem)
		}
	}

	// Join paths
	result := filepath.Join(append([]string{base}, elements...)...)

	// Ensure result is still within base directory
	absBase, err := filepath.Abs(base)
	if err != nil {
		return "", err
	}

	absResult, err := filepath.Abs(result)
	if err != nil {
		return "", err
	}

	// Check if result is within base directory
	if !strings.HasPrefix(absResult, absBase) {
		return "", errors.New("path escapes base directory")
	}

	return result, nil
}

// New comprehensive security check function
func IsSecurityThreat(input string) (bool, string) {
	if ScriptInjectionRegex.MatchString(input) {
		return true, "SCRIPT_INJECTION"
	}
	if SqlInjectionRegex.MatchString(input) {
		return true, "SQL_INJECTION"
	}
	if CommandInjectionRegex.MatchString(input) {
		return true, "COMMAND_INJECTION"
	}
	if LdapInjectionRegex.MatchString(input) {
		return true, "LDAP_INJECTION"
	}
	if XpathInjectionRegex.MatchString(input) {
		return true, "XPATH_INJECTION"
	}
	if XmlInjectionRegex.MatchString(input) {
		return true, "XML_INJECTION"
	}
	if PathTraversalRegex.MatchString(input) {
		return true, "PATH_TRAVERSAL"
	}
	return false, ""
}
