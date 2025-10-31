// Package services defines service-layer configuration interfaces and types used
// across the application. These types describe configuration for external
// services (storage, SMTP, cache, broker, etc.) and are intentionally
// lightweight so they can be populated from environment variables or
// configuration files. Keeping these contracts centralized makes it easier to
// mock or replace concrete service implementations in tests and different
// deployment environments.
package services

import (
	"time"
)

// EnvironmentServiceConfig represents the configuration for environment service
type EnvironmentServiceConfig struct {
	Path string `env:"APP_ENV"`
}

// SQLServiceConfig represents the configuration for SQL service
type SQLServiceConfig struct {
	DSN         string        `env:"DATABASE_URL"`
	MaxIdleConn int           `env:"DB_MAX_IDLE_CONN"`
	MaxOpenConn int           `env:"DB_MAX_OPEN_CONN"`
	MaxLifetime time.Duration `env:"DB_MAX_LIFETIME"`
}

// SQLLogsServiceConfig represents the configuration for SQL logs service
type SQLLogsServiceConfig struct {
	DSN         string        `env:"DATABASE_LOG_URL"`
	MaxIdleConn int           `env:"DB_MAX_IDLE_CONN"`
	MaxOpenConn int           `env:"DB_MAX_OPEN_CONN"`
	MaxLifetime time.Duration `env:"DB_MAX_LIFETIME"`
}

// StorageServiceConfig represents the configuration for storage service
type StorageServiceConfig struct {
	AccessKey   string `env:"STORAGE_ACCESS_KEY"`
	SecretKey   string `env:"STORAGE_SECRET_KEY"`
	Bucket      string `env:"STORAGE_BUCKET"`
	Endpoint    string `env:"STORAGE_URL"`
	Region      string `env:"STORAGE_REGION"`
	MaxFilezize int64  `env:"STORAGE_MAX_SIZE"`
	Driver      string `env:"STORAGE_DRIVER"`
}

// CacheServiceConfig represents the configuration for cache service
type CacheServiceConfig struct {
	Host     string `env:"REDIS_HOST"`
	Password string `env:"REDIS_PASSWORD"`
	Username string `env:"REDIS_USERNAME"`
	Port     int    `env:"REDIS_PORT"`
}

// BrokerServiceConfig represents the configuration for broker service
type BrokerServiceConfig struct {
	Host     string `env:"NATS_HOST"`
	Port     int    `env:"NATS_CLIENT_PORT"`
	ClientID string `env:"NATS_CLIENT"`
	Username string `env:"NATS_USERNAME"`
	Password string `env:"NATS_PASSWORD"`
}

// SecurityServiceConfig represents the configuration for security service
type SecurityServiceConfig struct {
	Memory      uint32 `env:"PASSWORD_MEMORY"`
	Iterations  uint32 `env:"PASSWORD_ITERATIONS"`
	Parallelism uint8  `env:"PASSWORD_PARALLELISM"`
	SaltLength  uint32 `env:"PASSWORD_SALT_LENTH"`
	KeyLength   uint32 `env:"PASSWORD_KEY_LENGTH"`
	Secret      []byte `env:"PASSWORD_SECRET"`
}

// OTPServiceConfig represents the configuration for OTP service
type OTPServiceConfig struct {
	Secret []byte `env:"OTP_SECRET"`
}

// SMSServiceConfig represents the configuration for SMS service
type SMSServiceConfig struct {
	AccountSID string `env:"TWILIO_ACCOUNT_SID"`
	AuthToken  string `env:"TWILIO_AUTH_TOKEN"`
	Sender     string `env:"TWILIO_SENDER"`
	MaxChars   int32  `env:"TWILIO_MAX_CHARACTERS"`
}

// SMTPServiceConfig represents the configuration for SMTP service
type SMTPServiceConfig struct {
	Host     string `env:"SMTP_HOST"`
	Port     int    `env:"SMTP_PORT"`
	Username string `env:"SMTP_USERNAME"`
	Password string `env:"SMTP_PASSWORD"`
	From     string `env:"SMTP_FROM"`
}

// RequestServiceConfig represents the configuration for request service
type RequestServiceConfig struct {
	AppPort     int    `env:"APP_PORT"`
	MetricsPort int    `env:"APP_METRICS_PORT"`
	ClientURL   string `env:"APP_CLIENT_URL"`
	ClientName  string `env:"APP_CLIENT_NAME"`
}
