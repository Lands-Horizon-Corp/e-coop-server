package horizon_services

import (
	"time"
)

type EnvironmentServiceConfig struct {
	Path string `env:"APP_ENV"`
}

type SQLServiceConfig struct {
	DSN         string        `env:"DATABASE_URL"`
	MaxIdleConn int           `env:"DB_MAX_IDLE_CONN"`
	MaxOpenConn int           `env:"DB_MAX_OPEN_CONN"`
	MaxLifetime time.Duration `env:"DB_MAX_LIFETIME"`
}

type StorageServiceConfig struct {
	AccessKey   string `env:"STORAGE_ACCESS_KEY"`
	SecretKey   string `env:"STORAGE_SECRET_KEY"`
	Bucket      string `env:"STORAGE_BUCKET"`
	Prefix      string `env:"STORAGE_PREFIX"`
	MaxFilezize int64  `env:"STORAGE_MAX_SIZE"`
}

type CacheServiceConfig struct {
	Host     string `env:"REDIS_HOST"`
	Password string `env:"REDIS_PASSWORD"`
	Username string `env:"REDIS_USERNAME"`
	Port     int    `env:"REDIS_PORT"`
}

type BrokerServiceConfig struct {
	Host string `env:"NATS_HOST"`
	Port int    `env:"NATS_CLIENT_PORT"`
}

type SecurityServiceConfig struct {
	Memory      uint32 `env:"PASSWORD_MEMORY"`
	Iterations  uint32 `env:"PASSWORD_ITERATIONS"`
	Parallelism uint8  `env:"PASSWORD_PARALLELISM"`
	SaltLength  uint32 `env:"PASSWORD_SALT_LENTH"`
	KeyLength   uint32 `env:"PASSWORD_KEY_LENGTH"`
	Secret      []byte `env:"PASSWORD_SECRET"`
}

type OTPServiceConfig struct {
	Secret []byte `env:"OTP_SECRET"`
}

type SMSServiceConfig struct {
	AccountSID string `env:"TWILIO_ACCOUNT_SID"`
	AuthToken  string `env:"TWILIO_AUTH_TOKEN"`
	Sender     string `env:"TWILIO_SENDER"`
	MaxChars   int32  `env:"TWILIO_MAX_CHARACTERS"`
}
type SMTPServiceConfig struct {
	Host     string `env:"SMTP_HOST"`
	Port     int    `env:"SMTP_PORT"`
	Username string `env:"SMTP_USERNAME"`
	Password string `env:"SMTP_PASSWORD"`
	From     string `env:"SMTP_FROM"`
}

type RequestServiceConfig struct {
	AppPort     int    `env:"APP_PORT"`
	MetricsPort int    `env:"APP_METRICS_PORT"`
	ClientURL   string `env:"APP_CLIENT_URL"`
	ClientName  string `env:"APP_CLIENT_NAME"`
}
