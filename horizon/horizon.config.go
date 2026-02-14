package horizon

import (
	"fmt"
	"strings"

	"github.com/Lands-Horizon-Corp/e-coop-server/helpers"
	"github.com/spf13/viper"
)

type ConfigImpl struct {
	AppPort        int
	AppMetricsPort int
	AppEnv         string
	AppClientURL   string
	AppClientName  string
	AppToken       string
	AppName        string

	SoketiHost      string
	SoketiPort      int
	SoketiAppID     string
	SoketiAppKey    string
	SoketiAppSecret string
	SoketiAppClient string
	SoketiURL       string

	PostgresUser         string
	PostgresPassword     string
	PostgresDB           string
	PostgresPort         int
	PostgresHost         string
	DatabaseURL          string
	DBMaxIdleConn        int
	DBMaxOpenConn        int
	DBMaxLifetimeSeconds int64

	AdminPostgresUser     string
	AdminPostgresPassword string
	AdminPostgresDB       string
	AdminPostgresPort     int
	AdminPostgresHost     string
	AdminDatabaseURL      string

	RedisHost     string
	RedisPort     int
	RedisPassword string
	RedisUsername string
	RedisURL      string

	StorageDriver    string
	StoragePort      int
	StorageAccessKey string
	StorageSecretKey string
	StorageBucket    string
	StorageURL       string
	StorageRegion    string
	StorageMaxSize   int64

	PasswordMemory     uint32
	PasswordIterations uint32
	PasswordParallel   uint8
	PasswordSaltLength uint32
	PasswordKeyLength  uint32
	PasswordSecret     string
	OTPSecret          string
	QRSecret           string

	SMTPHost         string
	SMTPPort         int
	SMTPUsername     string
	SMTPPassword     string
	SMTPFrom         string
	SMTPTestReceiver string

	TwilioAccountSID    string
	TwilioAuthToken     string
	TwilioSender        string
	TwilioTestRecv      string
	TwilioMaxCharacters int32

	LinodeHost          string
	LinodeUser          string
	LinodeKnownHost     string
	LinodeSSHPrivateKey string

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

	v.SetConfigFile(".env")
	_ = v.ReadInConfig()

	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	v.SetDefault("APP_PORT", 8000)
	v.SetDefault("APP_METRICS_PORT", 8001)
	v.SetDefault("APP_ENV", "development")
	v.SetDefault("APP_CLIENT_URL", "http://127.0.0.1:3000")
	v.SetDefault("APP_CLIENT_NAME", "horizon")
	v.SetDefault("APP_TOKEN", "default-token")
	v.SetDefault("APP_NAME", "myapp")

	v.SetDefault("SOKETI_HOST", "127.0.0.1")
	v.SetDefault("SOKETI_PORT", 6001)
	v.SetDefault("SOKETI_APP_ID", "")
	v.SetDefault("SOKETI_APP_KEY", "")
	v.SetDefault("SOKETI_APP_SECRET", "")
	v.SetDefault("SOKETI_APP_CLIENT", "")

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

	v.SetDefault("ADMIN_POSTGRES_USER", v.GetString("POSTGRES_USER"))
	v.SetDefault("ADMIN_POSTGRES_PASSWORD", v.GetString("POSTGRES_PASSWORD"))
	v.SetDefault("ADMIN_POSTGRES_DB", v.GetString("POSTGRES_DB"))
	v.SetDefault("ADMIN_POSTGRES_PORT", v.GetInt("POSTGRES_PORT"))
	v.SetDefault("ADMIN_POSTGRES_HOST", v.GetString("POSTGRES_HOST"))
	v.SetDefault("ADMIN_DATABASE_URL", v.GetString("DATABASE_URL"))

	v.SetDefault("REDIS_HOST", "127.0.0.1")
	v.SetDefault("REDIS_PORT", 6379)
	v.SetDefault("REDIS_PASSWORD", "password")
	v.SetDefault("REDIS_USERNAME", "default")
	v.SetDefault("REDIS_URL", "")

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

	v.SetDefault("TWILIO_MAX_CHARACTERS", 255)

	v.SetDefault("PGADMN_DEFAULT_PORT", 8002)
	v.SetDefault("PGADMIN_DEFAULT_EMAIL", "admin@127.0.0.1.com")
	v.SetDefault("PGADMIN_DEFAULT_PASSWORD", "adminpass")
	v.SetDefault("PGADMIN_HOST", "127.0.0.1")
	v.SetDefault("REDISINSIGHT_HOST", "127.0.0.1")
	v.SetDefault("REDISINSIGHT_PORT", 5540)
	v.SetDefault("STORAGE_CONSOLE_PORT", 9001)
	v.SetDefault("MAILPIT_UI_HOST", "127.0.0.1")
	v.SetDefault("MAILPIT_UI_PORT", 8025)

	mem, err := helpers.Int64ToUint32(v.GetInt64("PASSWORD_MEMORY"), "PASSWORD_MEMORY")
	if err != nil {
		return nil, err
	}

	iter, err := helpers.Int64ToUint32(v.GetInt64("PASSWORD_ITERATIONS"), "PASSWORD_ITERATIONS")
	if err != nil {
		return nil, err
	}

	par, err := helpers.Int64ToUint8(v.GetInt64("PASSWORD_PARALLELISM"), "PASSWORD_PARALLELISM")
	if err != nil {
		return nil, err
	}

	saltLen, err := helpers.Int64ToUint32(v.GetInt64("PASSWORD_SALT_LENGTH"), "PASSWORD_SALT_LENGTH")
	if err != nil {
		return nil, err
	}

	keyLen, err := helpers.Int64ToUint32(v.GetInt64("PASSWORD_KEY_LENGTH"), "PASSWORD_KEY_LENGTH")
	if err != nil {
		return nil, err
	}

	cfg := &ConfigImpl{
		AppPort:        v.GetInt("APP_PORT"),
		AppMetricsPort: v.GetInt("APP_METRICS_PORT"),
		AppEnv:         v.GetString("APP_ENV"),
		AppClientURL:   v.GetString("APP_CLIENT_URL"),
		AppClientName:  v.GetString("APP_CLIENT_NAME"),
		AppToken:       v.GetString("APP_TOKEN"),
		AppName:        v.GetString("APP_NAME"),

		SoketiHost:      v.GetString("SOKETI_HOST"),
		SoketiPort:      v.GetInt("SOKETI_PORT"),
		SoketiAppID:     v.GetString("SOKETI_APP_ID"),
		SoketiAppKey:    v.GetString("SOKETI_APP_KEY"),
		SoketiAppSecret: v.GetString("SOKETI_APP_SECRET"),
		SoketiAppClient: v.GetString("SOKETI_APP_CLIENT"),
		SoketiURL:       v.GetString("SOKETI_URL"),

		PostgresUser:         v.GetString("POSTGRES_USER"),
		PostgresPassword:     v.GetString("POSTGRES_PASSWORD"),
		PostgresDB:           v.GetString("POSTGRES_DB"),
		PostgresPort:         v.GetInt("POSTGRES_PORT"),
		PostgresHost:         v.GetString("POSTGRES_HOST"),
		DatabaseURL:          v.GetString("DATABASE_URL"),
		DBMaxIdleConn:        v.GetInt("DATABASE_MAX_IDLE_CONN"),
		DBMaxOpenConn:        v.GetInt("DATABASE_MAX_OPEN_CONN"),
		DBMaxLifetimeSeconds: v.GetInt64("DATABASE_MAX_LIFETIME"),

		AdminPostgresUser:     v.GetString("ADMIN_POSTGRES_USER"),
		AdminPostgresPassword: v.GetString("ADMIN_POSTGRES_PASSWORD"),
		AdminPostgresDB:       v.GetString("ADMIN_POSTGRES_DB"),
		AdminPostgresPort:     v.GetInt("ADMIN_POSTGRES_PORT"),
		AdminPostgresHost:     v.GetString("ADMIN_POSTGRES_HOST"),
		AdminDatabaseURL:      v.GetString("ADMIN_DATABASE_URL"),

		RedisHost:     v.GetString("REDIS_HOST"),
		RedisPort:     v.GetInt("REDIS_PORT"),
		RedisPassword: v.GetString("REDIS_PASSWORD"),
		RedisUsername: v.GetString("REDIS_USERNAME"),
		RedisURL:      v.GetString("REDIS_URL"),

		StorageDriver:    v.GetString("STORAGE_DRIVER"),
		StoragePort:      v.GetInt("STORAGE_API_PORT"),
		StorageAccessKey: v.GetString("STORAGE_ACCESS_KEY"),
		StorageSecretKey: v.GetString("STORAGE_SECRET_KEY"),
		StorageBucket:    v.GetString("STORAGE_BUCKET"),
		StorageURL:       v.GetString("STORAGE_URL"),
		StorageRegion:    v.GetString("STORAGE_REGION"),
		StorageMaxSize:   v.GetInt64("STORAGE_MAX_SIZE"),

		PasswordMemory:     mem,
		PasswordIterations: iter,
		PasswordParallel:   par,
		PasswordSaltLength: saltLen,
		PasswordKeyLength:  keyLen,
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
		TwilioMaxCharacters: int32(v.GetInt32("TWILIO_MAX_CHARACTERS")),

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
	if cfg.AdminDatabaseURL == "" {
		cfg.AdminDatabaseURL = fmt.Sprintf(
			"postgres://%s:%s@%s:%d/%s?sslmode=disable",
			cfg.AdminPostgresUser,
			cfg.AdminPostgresPassword,
			cfg.AdminPostgresHost,
			cfg.AdminPostgresPort,
			cfg.AdminPostgresDB,
		)
	}
	if cfg.AdminDatabaseURL == "" {
		cfg.AdminDatabaseURL = cfg.DatabaseURL
	}

	if cfg.SoketiURL == "" {
		cfg.SoketiURL = fmt.Sprintf(
			"http://%s:%d/apps/%s/events",
			cfg.SoketiHost,
			cfg.SoketiPort,
			cfg.SoketiAppID,
		)
	}
	if cfg.RedisURL == "" {
		if cfg.RedisUsername != "" && cfg.RedisPassword != "" {
			cfg.RedisURL = fmt.Sprintf(
				"redis://%s:%s@%s:%d",
				cfg.RedisUsername,
				cfg.RedisPassword,
				cfg.RedisHost,
				cfg.RedisPort,
			)
		} else {
			cfg.RedisURL = fmt.Sprintf(
				"redis://%s:%d",
				cfg.RedisHost,
				cfg.RedisPort,
			)
		}
	}

	return cfg, nil
}
