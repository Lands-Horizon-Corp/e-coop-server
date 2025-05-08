package horizon

import (
	"strings"
	"time"
)

const (
	storageDuration   = time.Hour * 4
	authExpiration    = 16 * time.Hour
	tokenLinkValidity = 10 * time.Minute
	maxSMSCharacters  = 2_000
)

// HorizonConfig holds all configuration values.
type HorizonConfig struct {
	AppPort        int
	AppMetricsPort int
	AppName        string
	AppToken       string
	AppLog         string
	AppEnvironment string
	AppMainLog     string
	AppClientURL   string
	AppTokenName   string

	PostgresUser           string
	PostgresPassword       string
	PostgresDB             string
	PostgresHost           string
	PostgresPort           int
	PgAdminDefaultEmail    string
	PgAdminDefaultPassword string
	PgAdminHost            string
	PgAdminPort            int

	RedisPort        int
	RedisHost        string
	RedisPassword    string
	RedisUsername    string
	RedisInsightHost string
	RedisInsightPort int

	SMTPHost      string
	SMTPPort      int
	SMTPUsername  string
	SMTPPassword  string
	SMTPFrom      string
	MailPitUIHost string
	MailPitUIPort int

	StorageDriver      string
	StorageAccessKey   string
	StorageSecretKey   string
	StorageHost        string
	StorageRegion      string
	StorageBucket      string
	StorageApiPort     int
	StorageConsolePort int

	NATSHost        string
	NATSClientPort  int
	NATSMonitorPort int

	NATSClientWSPort int
}

func NewHorizonConfig() (*HorizonConfig, error) {
	return &HorizonConfig{
		AppPort:        GetInt("APP_PORT", 8000),
		AppMetricsPort: GetInt("APP_METRICS_PORT", 8001),
		AppName:        GetString("APP_NAME", "ENGINE"),
		AppToken:       GetString("APP_TOKEN", "oYrsXzg7eu7Yt5So4e62r7LDVH2hj"),
		AppLog:         GetString("APP_LOG", "./logs/"),
		AppEnvironment: GetString("APP_ENV", "production"),
		AppMainLog:     GetString("APP_MAIN_LOG", "./logs/main.log"),
		AppClientURL:   GetString("APP_CLIENT_URL", "http://localhost:3000"),
		AppTokenName:   GetString("APP_TOKEN_NAME", "0aWalpi9q4b6Adm811hIJDzoh"),

		PostgresUser:           GetString("POSTGRES_USER", "dev"),
		PostgresPassword:       GetString("POSTGRES_PASSWORD", "devpass"),
		PostgresDB:             GetString("POSTGRES_DB", "devdb"),
		PostgresHost:           GetString("POSTGRES_HOST", "postgres"),
		PostgresPort:           GetInt("POSTGRES_PORT", 5432),
		PgAdminDefaultEmail:    GetString("PGADMIN_DEFAULT_EMAIL", "admin@localhost.com"),
		PgAdminDefaultPassword: GetString("PGADMIN_DEFAULT_PASSWORD", "adminpass"),
		PgAdminHost:            GetString("PGADMIN_HOST", "pgadmin"),
		PgAdminPort:            GetInt("PGADMIN_PORT", 5050),

		RedisPort:        GetInt("REDIS_PORT", 6379),
		RedisHost:        GetString("REDIS_HOST", "redis"),
		RedisPassword:    GetString("REDIS_PASSWORD", "password"),
		RedisUsername:    GetString("REDIS_USERNAME", "default"),
		RedisInsightHost: GetString("REDISINSIGHT_HOST", "redisinsight"),
		RedisInsightPort: GetInt("REDISINSIGHT_PORT", 8001),

		SMTPHost:      GetString("SMTP_HOST", "mailpit"),
		SMTPPort:      GetInt("SMTP_PORT", 1025),
		SMTPUsername:  GetString("SMTP_USERNAME", ""),
		SMTPPassword:  GetString("SMTP_PASSWORD", ""),
		SMTPFrom:      GetString("SMTP_FROM", "landshorizon@gmail.com"),
		MailPitUIPort: GetInt("MAILPIT_UI_PORT", 8025),
		MailPitUIHost: GetString("MAILPIT_UI_HOST", "mailpit"),

		StorageDriver:      GetString("STORAGE_DRIVER", "minio"),
		StorageAccessKey:   GetString("STORAGE_ACCESS_KEY", "minioadmin"),
		StorageSecretKey:   GetString("STORAGE_SECRET_KEY", "minioadmin"),
		StorageHost:        GetString("STORAGE_HOST", "minio"),
		StorageRegion:      GetString("STORAGE_REGION", "us-east-1"),
		StorageBucket:      GetString("STORAGE_BUCKET", "my-bucket"),
		StorageApiPort:     GetInt("STORAGE_API_PORT", 9000),
		StorageConsolePort: GetInt("STORAGE_CONSOLE_PORT", 9001),

		NATSHost:         GetString("NATS_HOST", "nats"),
		NATSClientPort:   GetInt("NATS_CLIENT_PORT", 4222),
		NATSMonitorPort:  GetInt("NATS_MONITOR_PORT", 8222),
		NATSClientWSPort: GetInt("NAT_CLIENT_WS_PORT", 8080),
	}, nil
}
func (hc *HorizonConfig) CanDebug() bool {
	env := strings.TrimSpace(strings.ToLower(hc.AppEnvironment))
	switch env {
	case "dev", "development", "developer", "test", "testing", "debug", "debugging":
		return true
	default:
		return false
	}
}
