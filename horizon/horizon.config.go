package horizon

import (
	"fmt"

	"github.com/caarlos0/env/v9"
)

type ConfigImpl struct {
	// App
	AppPort        int    `env:"APP_PORT" envDefault:"8000"`
	AppMetricsPort int    `env:"APP_METRICS_PORT" envDefault:"8001"`
	AppEnv         string `env:"APP_ENV" envDefault:"development"`
	AppClientURL   string `env:"APP_CLIENT_URL" envDefault:"http://127.0.0.1:3000"`
	AppClientName  string `env:"APP_CLIENT_NAME" envDefault:"horizon"`
	AppToken       string `env:"APP_TOKEN" envDefault:"default-token"`
	AppName        string `env:"APP_NAME" envDefault:"myapp"`

	// NATS
	NatsHost          string `env:"NATS_HOST" envDefault:"127.0.0.1"`
	NatsClientPort    int    `env:"NATS_CLIENT_PORT" envDefault:"4222"`
	NatsMonitorPort   int    `env:"NATS_MONITOR_PORT" envDefault:"8222"`
	NatsWebsocketPort int    `env:"NATS_WEBSOCKET_PORT" envDefault:"8080"`
	NatsClient        string `env:"NATS_CLIENT"`
	NatsUsername      string `env:"NATS_USERNAME"`
	NatsPassword      string `env:"NATS_PASSWORD"`

	// Postgres
	PostgresUser     string `env:"POSTGRES_USER" envDefault:"dev"`
	PostgresPassword string `env:"POSTGRES_PASSWORD" envDefault:"devpass"`
	PostgresDB       string `env:"POSTGRES_DB" envDefault:"devdb"`
	PostgresPort     int    `env:"POSTGRES_PORT" envDefault:"5432"`
	PostgresHost     string `env:"POSTGRES_HOST" envDefault:"127.0.0.1"`
	DatabaseURL      string `env:"DATABASE_URL"`

	// Redis
	RedisHost     string `env:"REDIS_HOST" envDefault:"127.0.0.1"`
	RedisPort     int    `env:"REDIS_PORT" envDefault:"6379"`
	RedisPassword string `env:"REDIS_PASSWORD" envDefault:"password"`
	RedisUsername string `env:"REDIS_USERNAME" envDefault:"default"`

	// Storage
	StorageDriver    string `env:"STORAGE_DRIVER" envDefault:"minio"`
	StoragePort      int    `env:"STORAGE_API_PORT" envDefault:"9000"`
	StorageAccessKey string `env:"STORAGE_ACCESS_KEY" envDefault:"accesskey"`
	StorageSecretKey string `env:"STORAGE_SECRET_KEY" envDefault:"secretkey"`
	StorageBucket    string `env:"STORAGE_BUCKET" envDefault:"bucket"`
	StorageURL       string `env:"STORAGE_URL" envDefault:"127.0.0.1:9000"`
	StorageRegion    string `env:"STORAGE_REGION" envDefault:"us-east-1"`
	StorageMaxSize   int64  `env:"STORAGE_MAX_SIZE" envDefault:"10485760"`

	// Security / Password
	PasswordMemory     uint32 `env:"PASSWORD_MEMORY" envDefault:"65536"`
	PasswordIterations uint32 `env:"PASSWORD_ITERATIONS" envDefault:"3"`
	PasswordParallel   uint8  `env:"PASSWORD_PARALLELISM" envDefault:"2"`
	PasswordSaltLength uint32 `env:"PASSWORD_SALT_LENTH" envDefault:"16"`
	PasswordKeyLength  uint32 `env:"PASSWORD_KEY_LENGTH" envDefault:"32"`
	PasswordSecret     string `env:"PASSWORD_SECRET" envDefault:"changeme"`
	OTPSecret          string `env:"OTP_SECRET" envDefault:"otpsecret"`
	QRSecret           string `env:"QR_SECRET" envDefault:"qrsecret"`

	// SMTP
	SMTPHost         string `env:"SMTP_HOST" envDefault:"127.0.0.1"`
	SMTPPort         int    `env:"SMTP_PORT" envDefault:"1025"`
	SMTPUsername     string `env:"SMTP_USERNAME"`
	SMTPPassword     string `env:"SMTP_PASSWORD"`
	SMTPFrom         string `env:"SMTP_FROM" envDefault:"dev@local.test"`
	SMTPTestReceiver string `env:"SMTP_TEST_RECIEVER"`

	// Twilio
	TwilioAccountSID string `env:"TWILIO_ACCOUNT_SID"`
	TwilioAuthToken  string `env:"TWILIO_AUTH_TOKEN"`
	TwilioSender     string `env:"TWILIO_SENDER"`
	TwilioTestRecv   string `env:"TWILIO_TEST_RECIEVER"`

	// Linode deploy
	LinodeHost          string `env:"LINODE_HOST"`
	LinodeUser          string `env:"LINODE_USER"`
	LinodeKnownHost     string `env:"LINODE_KNOWN_HOST"`
	LinodeSSHPrivateKey string `env:"LINODE_SSH_PRIVATE_KEY"`

	// UI Ports
	PgAdminPort        int    `env:"PGADMN_DEFAULT_PORT" envDefault:"8002"`
	PgAdminEmail       string `env:"PGADMIN_DEFAULT_EMAIL" envDefault:"admin@127.0.0.1.com"`
	PgAdminPassword    string `env:"PGADMIN_DEFAULT_PASSWORD" envDefault:"adminpass"`
	PgAdminHost        string `env:"PGADMIN_HOST" envDefault:"127.0.0.1"`
	RedisInsightHost   string `env:"REDISINSIGHT_HOST" envDefault:"127.0.0.1"`
	RedisInsightPort   int    `env:"REDISINSIGHT_PORT" envDefault:"5540"`
	StorageConsolePort int    `env:"STORAGE_CONSOLE_PORT" envDefault:"9001"`
	MailpitUIHost      string `env:"MAILPIT_UI_HOST" envDefault:"127.0.0.1"`
	MailpitUIPort      int    `env:"MAILPIT_UI_PORT" envDefault:"8025"`
}

func NewConfigImpl() (*ConfigImpl, error) {
	cfg := &ConfigImpl{}
	if err := env.Parse(cfg); err != nil {
		return nil, err
	}
	if cfg.DatabaseURL == "" {
		cfg.DatabaseURL = fmt.Sprintf(
			"postgres://%s:%s@%s:%d/%s?sslmode=disable",
			cfg.PostgresUser, cfg.PostgresPassword, cfg.PostgresHost, cfg.PostgresPort, cfg.PostgresDB,
		)
	}
	return cfg, nil
}
