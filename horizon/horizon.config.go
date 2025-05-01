package horizon

// HorizonConfig holds all configuration values.
type HorizonConfig struct {
	AppPort                int
	AppName                string
	AppTOKEN               string
	AppLog                 string
	PostgresUser           string
	PostgresPassword       string
	PostgresDB             string
	PostgresHost           string
	PostgresPort           int
	PgAdminDefaultEmail    string
	PgAdminDefaultPassword string
	PgAdminHost            string
	PgAdminPort            int
	RedisPort              int
	RedisHost              string
	RedisPassword          string
	RedisUsername          string
	RedisInsightHost       string
	RedisInsightPort       int
	MailPitSMTPHost        string
	MailPitSMTPPort        int
	MailPitUIHost          string
	MailPitUIPort          int
	MailPitEmail           string
	StorageDriver          string
	StorageAccessKey       string
	StorageSecretKey       string
	StorageEndpoint        string
	StorageRegion          string
	StorageBucket          string
	StorageAPI_Port        int
	StorageConsolePort     int
	NATSClientPort         int
	NATSMonitorPort        int
}

func NewHorizonConfig() (*HorizonConfig, error) {
	return &HorizonConfig{
		AppPort:                GetInt("APP_PORT", 8000),
		AppName:                GetString("APP_NAME", "ENGINE"),
		AppTOKEN:               GetString("APP_TOKEN", ""),
		AppLog:                 GetString("APP_LOG", "./logs/"),
		PostgresUser:           GetString("POSTGRES_USER", "dev"),
		PostgresPassword:       GetString("POSTGRES_PASSWORD", "devpass"),
		PostgresDB:             GetString("POSTGRES_DB", "devdb"),
		PostgresHost:           GetString("POSTGRES_HOST", "postgres"),
		PostgresPort:           GetInt("POSTGRES_PORT", 5432),
		PgAdminDefaultEmail:    GetString("PGADMIN_DEFAULT_EMAIL", "admin@localhost.com"),
		PgAdminDefaultPassword: GetString("PGADMIN_DEFAULT_PASSWORD", "adminpass"),
		PgAdminHost:            GetString("PGADMIN_HOST", "pgadmin"),
		PgAdminPort:            GetInt("PGADMIN_PORT", 5050),
		RedisPort:              GetInt("REDIS_PORT", 6379),
		RedisHost:              GetString("REDIS_HOST", "redis"),
		RedisPassword:          GetString("REDIS_PASSWORD", "password"),
		RedisUsername:          GetString("REDIS_USERNAME", "default"),
		RedisInsightHost:       GetString("REDISINSIGHT_HOST", "redisinsight"),
		RedisInsightPort:       GetInt("REDISINSIGHT_PORT", 8001),
		MailPitSMTPHost:        GetString("MAILPIT_SMTP_HOST", "mailpit"),
		MailPitSMTPPort:        GetInt("MAILPIT_SMTP_PORT", 1025),
		MailPitUIHost:          GetString("MAILPIT_UI_HOST", "mailpit"),
		MailPitUIPort:          GetInt("MAILPIT_UI_PORT", 8025),
		MailPitEmail:           GetString("MAILPIT_EMAIL", "landshorizon@gmail.com"),
		StorageDriver:          GetString("STORAGE_DRIVER", "minio"),
		StorageAccessKey:       GetString("STORAGE_ACCESS_KEY", "minioadmin"),
		StorageSecretKey:       GetString("STORAGE_SECRET_KEY", "minioadmin"),
		StorageEndpoint:        GetString("STORAGE_ENDPOINT", "http://localhost:9000"),
		StorageRegion:          GetString("STORAGE_REGION", "us-east-1"),
		StorageBucket:          GetString("STORAGE_BUCKET", "my-bucket"),
		StorageAPI_Port:        GetInt("STORAGE_API_PORT", 9000),
		StorageConsolePort:     GetInt("STORAGE_CONSOLE_PORT", 9001),
		NATSClientPort:         GetInt("NATS_CLIENT_PORT", 4222),
		NATSMonitorPort:        GetInt("NATS_MONITOR_PORT", 8222),
	}, nil
}
