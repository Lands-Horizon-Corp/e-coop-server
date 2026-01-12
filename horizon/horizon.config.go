package horizon

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

type ConfigImpl struct {
	// App
	AppPort        int
	AppMetricsPort int
	AppEnv         string
	AppClientURL   string
	AppClientName  string
	AppToken       string
	AppName        string

	// NATS
	NatsHost          string
	NatsClientPort    int
	NatsMonitorPort   int
	NatsWebsocketPort int
	NatsClient        string
	NatsUsername      string
	NatsPassword      string

	// Postgres
	PostgresUser         string
	PostgresPassword     string
	PostgresDB           string
	PostgresPort         int
	PostgresHost         string
	DatabaseURL          string
	DBMaxIdleConn        int
	DBMaxOpenConn        int
	DBMaxLifetimeSeconds int64

	// Redis
	RedisHost     string
	RedisPort     int
	RedisPassword string
	RedisUsername string

	// Storage
	StorageDriver    string
	StoragePort      int
	StorageAccessKey string
	StorageSecretKey string
	StorageBucket    string
	StorageURL       string
	StorageRegion    string
	StorageMaxSize   int64

	// Security / Password
	PasswordMemory     uint32
	PasswordIterations uint32
	PasswordParallel   uint8
	PasswordSaltLength uint32
	PasswordKeyLength  uint32
	PasswordSecret     string
	OTPSecret          string
	QRSecret           string

	// SMTP
	SMTPHost         string
	SMTPPort         int
	SMTPUsername     string
	SMTPPassword     string
	SMTPFrom         string
	SMTPTestReceiver string

	// Twilio
	TwilioAccountSID    string
	TwilioAuthToken     string
	TwilioSender        string
	TwilioTestRecv      string
	TwilioMaxCharacters int32

	// Linode deploy
	LinodeHost          string
	LinodeUser          string
	LinodeKnownHost     string
	LinodeSSHPrivateKey string

	// UI Ports
	PgAdminPort        int
	PgAdminEmail       string
	PgAdminPassword    string
	PgAdminHost        string
	RedisInsightHost   string
	RedisInsightPort   int
	StorageConsolePort int
	MailpitUIHost      string
	MailpitUIPort      int
}

func NewConfigImpl() (*ConfigImpl, error) {
	v := viper.New()

	// Load .env file if present
	v.SetConfigFile(".env")
	_ = v.ReadInConfig() // ignore if missing

	// Automatically read ENV variables
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Set defaults (same as your previous envDefault)
	v.SetDefault("APP_PORT", 8000)
	v.SetDefault("APP_METRICS_PORT", 8001)
	v.SetDefault("APP_ENV", "development")
	v.SetDefault("APP_CLIENT_URL", "http://127.0.0.1:3000")
	v.SetDefault("APP_CLIENT_NAME", "horizon")
	v.SetDefault("APP_TOKEN", "default-token")
	v.SetDefault("APP_NAME", "myapp")

	v.SetDefault("NATS_HOST", "127.0.0.1")
	v.SetDefault("NATS_CLIENT_PORT", 4222)
	v.SetDefault("NATS_MONITOR_PORT", 8222)
	v.SetDefault("NATS_WEBSOCKET_PORT", 8080)

	v.SetDefault("POSTGRES_USER", "dev")
	v.SetDefault("POSTGRES_PASSWORD", "devpass")
	v.SetDefault("POSTGRES_DB", "devdb")
	v.SetDefault("POSTGRES_PORT", 5432)
	v.SetDefault("POSTGRES_HOST", "127.0.0.1")
	v.SetDefault("DATABASE_URL", "")
	v.SetDefault("POSTGRES_HOST", "127.0.0.1")
	v.SetDefault("DATABASE_MAX_IDLE_CONN", 10)
	v.SetDefault("DATABASE_MAX_OPEN_CONN", 100)
	v.SetDefault("DATABASE_MAX_LIFETIME", 0)

	v.SetDefault("REDIS_HOST", "127.0.0.1")
	v.SetDefault("REDIS_PORT", 6379)
	v.SetDefault("REDIS_PASSWORD", "password")
	v.SetDefault("REDIS_USERNAME", "default")

	v.SetDefault("STORAGE_DRIVER", "minio")
	v.SetDefault("STORAGE_API_PORT", 9000)
	v.SetDefault("STORAGE_ACCESS_KEY", "accesskey")
	v.SetDefault("STORAGE_SECRET_KEY", "secretkey")
	v.SetDefault("STORAGE_BUCKET", "bucket")
	v.SetDefault("STORAGE_URL", "127.0.0.1:9000")
	v.SetDefault("STORAGE_REGION", "us-east-1")
	v.SetDefault("STORAGE_MAX_SIZE", 10485760)

	v.SetDefault("PASSWORD_MEMORY", 65536)
	v.SetDefault("PASSWORD_ITERATIONS", 3)
	v.SetDefault("PASSWORD_PARALLELISM", 2)
	v.SetDefault("PASSWORD_SALT_LENTH", 16)
	v.SetDefault("PASSWORD_KEY_LENGTH", 32)
	v.SetDefault("PASSWORD_SECRET", "changeme")
	v.SetDefault("OTP_SECRET", "otpsecret")
	v.SetDefault("QR_SECRET", "qrsecret")

	v.SetDefault("SMTP_HOST", "127.0.0.1")
	v.SetDefault("SMTP_PORT", 1025)
	v.SetDefault("SMTP_FROM", "dev@local.test")

	v.SetDefault("TWILIO_MAX_CHARACTERS", 100)

	v.SetDefault("PGADMN_DEFAULT_PORT", 8002)
	v.SetDefault("PGADMIN_DEFAULT_EMAIL", "admin@127.0.0.1.com")
	v.SetDefault("PGADMIN_DEFAULT_PASSWORD", "adminpass")
	v.SetDefault("PGADMIN_HOST", "127.0.0.1")
	v.SetDefault("REDISINSIGHT_HOST", "127.0.0.1")
	v.SetDefault("REDISINSIGHT_PORT", 5540)
	v.SetDefault("STORAGE_CONSOLE_PORT", 9001)
	v.SetDefault("MAILPIT_UI_HOST", "127.0.0.1")
	v.SetDefault("MAILPIT_UI_PORT", 8025)

	cfg := &ConfigImpl{
		AppPort:        v.GetInt("APP_PORT"),
		AppMetricsPort: v.GetInt("APP_METRICS_PORT"),
		AppEnv:         v.GetString("APP_ENV"),
		AppClientURL:   v.GetString("APP_CLIENT_URL"),
		AppClientName:  v.GetString("APP_CLIENT_NAME"),
		AppToken:       v.GetString("APP_TOKEN"),
		AppName:        v.GetString("APP_NAME"),

		NatsHost:          v.GetString("NATS_HOST"),
		NatsClientPort:    v.GetInt("NATS_CLIENT_PORT"),
		NatsMonitorPort:   v.GetInt("NATS_MONITOR_PORT"),
		NatsWebsocketPort: v.GetInt("NATS_WEBSOCKET_PORT"),
		NatsClient:        v.GetString("NATS_CLIENT"),
		NatsUsername:      v.GetString("NATS_USERNAME"),
		NatsPassword:      v.GetString("NATS_PASSWORD"),

		PostgresUser:         v.GetString("POSTGRES_USER"),
		PostgresPassword:     v.GetString("POSTGRES_PASSWORD"),
		PostgresDB:           v.GetString("POSTGRES_DB"),
		PostgresPort:         v.GetInt("POSTGRES_PORT"),
		PostgresHost:         v.GetString("POSTGRES_HOST"),
		DatabaseURL:          v.GetString("DATABASE_URL"),
		DBMaxIdleConn:        v.GetInt("DATABASE_MAX_IDLE_CONN"),
		DBMaxOpenConn:        v.GetInt("DATABASE_MAX_OPEN_CONN"),
		DBMaxLifetimeSeconds: v.GetInt64("DATABASE_MAX_LIFETIME"),

		RedisHost:     v.GetString("REDIS_HOST"),
		RedisPort:     v.GetInt("REDIS_PORT"),
		RedisPassword: v.GetString("REDIS_PASSWORD"),
		RedisUsername: v.GetString("REDIS_USERNAME"),

		StorageDriver:    v.GetString("STORAGE_DRIVER"),
		StoragePort:      v.GetInt("STORAGE_API_PORT"),
		StorageAccessKey: v.GetString("STORAGE_ACCESS_KEY"),
		StorageSecretKey: v.GetString("STORAGE_SECRET_KEY"),
		StorageBucket:    v.GetString("STORAGE_BUCKET"),
		StorageURL:       v.GetString("STORAGE_URL"),
		StorageRegion:    v.GetString("STORAGE_REGION"),
		StorageMaxSize:   v.GetInt64("STORAGE_MAX_SIZE"),

		PasswordMemory:     uint32(v.GetInt("PASSWORD_MEMORY")),
		PasswordIterations: uint32(v.GetInt("PASSWORD_ITERATIONS")),
		PasswordParallel:   uint8(v.GetInt("PASSWORD_PARALLELISM")),
		PasswordSaltLength: uint32(v.GetInt("PASSWORD_SALT_LENTH")),
		PasswordKeyLength:  uint32(v.GetInt("PASSWORD_KEY_LENGTH")),
		PasswordSecret:     v.GetString("PASSWORD_SECRET"),
		OTPSecret:          v.GetString("OTP_SECRET"),
		QRSecret:           v.GetString("QR_SECRET"),

		SMTPHost:         v.GetString("SMTP_HOST"),
		SMTPPort:         v.GetInt("SMTP_PORT"),
		SMTPUsername:     v.GetString("SMTP_USERNAME"),
		SMTPPassword:     v.GetString("SMTP_PASSWORD"),
		SMTPFrom:         v.GetString("SMTP_FROM"),
		SMTPTestReceiver: v.GetString("SMTP_TEST_RECIEVER"),

		TwilioAccountSID:    v.GetString("TWILIO_ACCOUNT_SID"),
		TwilioAuthToken:     v.GetString("TWILIO_AUTH_TOKEN"),
		TwilioSender:        v.GetString("TWILIO_SENDER"),
		TwilioTestRecv:      v.GetString("TWILIO_TEST_RECIEVER"),
		TwilioMaxCharacters: int32(v.GetInt("TWILIO_MAX_CHARACTERS")),

		LinodeHost:          v.GetString("LINODE_HOST"),
		LinodeUser:          v.GetString("LINODE_USER"),
		LinodeKnownHost:     v.GetString("LINODE_KNOWN_HOST"),
		LinodeSSHPrivateKey: v.GetString("LINODE_SSH_PRIVATE_KEY"),

		PgAdminPort:        v.GetInt("PGADMN_DEFAULT_PORT"),
		PgAdminEmail:       v.GetString("PGADMIN_DEFAULT_EMAIL"),
		PgAdminPassword:    v.GetString("PGADMIN_DEFAULT_PASSWORD"),
		PgAdminHost:        v.GetString("PGADMIN_HOST"),
		RedisInsightHost:   v.GetString("REDISINSIGHT_HOST"),
		RedisInsightPort:   v.GetInt("REDISINSIGHT_PORT"),
		StorageConsolePort: v.GetInt("STORAGE_CONSOLE_PORT"),
		MailpitUIHost:      v.GetString("MAILPIT_UI_HOST"),
		MailpitUIPort:      v.GetInt("MAILPIT_UI_PORT"),
	}

	// Build DatabaseURL if not set
	if cfg.DatabaseURL == "" {
		cfg.DatabaseURL = fmt.Sprintf(
			"postgres://%s:%s@%s:%d/%s?sslmode=disable",
			cfg.PostgresUser,
			cfg.PostgresPassword,
			cfg.PostgresHost,
			cfg.PostgresPort,
			cfg.PostgresDB,
		)
	}

	return cfg, nil
}
