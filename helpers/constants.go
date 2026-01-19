package helpers

// Suspicious path patterns
var suspiciousPaths = []string{
	// "env",
	// "config",
	// "secret",
	// "apikey",
	// ".git",
	// "node_modules",
	// "server.go",
	// "credentials",
	// "database",
	// "wp-admin",
	// "etc/passwd",
	// "docker",
	// "ssh",
	// "private",
	// "backup",
	// "phpmyadmin",
	// "config.yaml",
	// "config.json",
	// "secrets.yaml",
}

const (
	Green  = "\033[32m"
	Blue   = "\033[34m"
	Yellow = "\033[33m"
	Red    = "\033[31m"
	Reset  = "\033[0m"
	Cyan   = "\033[36m"
)
