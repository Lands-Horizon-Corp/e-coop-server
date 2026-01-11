package helpers

// Suspicious path patterns
var suspiciousPaths = []string{
	"env",
	"config",
	"secret",
	"password",
	"apikey",
	".git",
	"node_modules",
	"server.go",
	"credentials",
	"database",
	"wp-admin",
	"etc/passwd",
	"docker",
	"ssh",
	"private",
	"backup",
	"phpmyadmin",
	"config.yaml",
	"config.json",
	"secrets.yaml",
}
