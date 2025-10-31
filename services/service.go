package services

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/services/handlers"
	"github.com/Lands-Horizon-Corp/e-coop-server/services/horizon"
	"github.com/go-playground/validator/v10"
	"github.com/rotisserie/eris"
	"go.uber.org/zap"
)

// HorizonService manages all application services and their lifecycle.
type HorizonService struct {
	Environment horizon.EnvironmentService
	Database    horizon.SQLDatabaseService

	// LogDatabase horizon.SQLDatabaseService
	Storage   horizon.StorageService
	Cache     horizon.CacheService
	Broker    horizon.MessageBrokerService
	Cron      horizon.SchedulerService
	Security  horizon.SecurityService
	OTP       horizon.OTPService
	SMS       horizon.SMSService
	SMTP      horizon.SMTPService
	Request   horizon.APIService
	QR        horizon.QRService
	Validator *validator.Validate
	Logger    *zap.Logger
}

// HorizonServiceConfig contains configuration options for initializing HorizonService.
type HorizonServiceConfig struct {
	EnvironmentConfig *EnvironmentServiceConfig
	SQLConfig         *SQLServiceConfig

	// SQLLogConfig         *SQLLogsServiceConfig
	StorageConfig        *StorageServiceConfig
	CacheConfig          *CacheServiceConfig
	BrokerConfig         *BrokerServiceConfig
	SecurityConfig       *SecurityServiceConfig
	OTPServiceConfig     *OTPServiceConfig
	SMSServiceConfig     *SMSServiceConfig
	SMTPServiceConfig    *SMTPServiceConfig
	RequestServiceConfig *RequestServiceConfig
}

// NewHorizonService creates and initializes a new HorizonService instance.
func NewHorizonService(cfg HorizonServiceConfig) *HorizonService {
	service := &HorizonService{}
	service.Validator = validator.New()

	logger, err := zap.NewProduction()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to initialize zap logger: %v\n", err)
		service.Logger = zap.NewNop()
	} else {
		service.Logger = logger
	}
	env := ".env"
	if cfg.EnvironmentConfig != nil {
		env = cfg.EnvironmentConfig.Path
	}
	service.Environment = horizon.NewEnvironmentService(env)
	isStaging := service.Environment.GetString("APP_ENV", "development") == "staging"

	if cfg.BrokerConfig != nil {
		service.Broker = horizon.NewHorizonMessageBroker(
			cfg.BrokerConfig.Host,
			cfg.BrokerConfig.Port,
			cfg.BrokerConfig.ClientID,
			cfg.BrokerConfig.Username,
			cfg.BrokerConfig.Password,
		)
	} else {
		service.Broker = horizon.NewHorizonMessageBroker(
			service.Environment.GetString("NATS_HOST", "127.0.0.1"),
			service.Environment.GetInt("NATS_CLIENT_PORT", 4222),
			service.Environment.GetString("NATS_CLIENT", ""),
			service.Environment.GetString("NATS_USERNAME", ""),
			service.Environment.GetString("NATS_PASSWORD", ""),
		)
	}
	if cfg.RequestServiceConfig != nil {
		service.Request = horizon.NewHorizonAPIService(
			cfg.RequestServiceConfig.AppPort,
			cfg.RequestServiceConfig.MetricsPort,
			cfg.RequestServiceConfig.ClientURL,
			cfg.RequestServiceConfig.ClientName,
			isStaging,
		)
	} else {
		service.Request = horizon.NewHorizonAPIService(
			service.Environment.GetInt("APP_PORT", 8000),
			service.Environment.GetInt("APP_METRICS_PORT", 8001),
			service.Environment.GetString("APP_CLIENT_URL", "http://127.0.0.1:3000"),
			service.Environment.GetString("APP_CLIENT_NAME", "horizon"),
			isStaging,
		)
	}
	if cfg.SecurityConfig != nil {
		service.Security = horizon.NewSecurityService(
			cfg.SecurityConfig.Memory,
			cfg.SecurityConfig.Iterations,
			cfg.SecurityConfig.Parallelism,
			cfg.SecurityConfig.SaltLength,
			cfg.SecurityConfig.KeyLength,
			cfg.SecurityConfig.Secret,
		)
	} else {
		service.Security = horizon.NewSecurityService(
			service.Environment.GetUint32("PASSWORD_MEMORY", 65536),
			service.Environment.GetUint32("PASSWORD_ITERATIONS", 3),
			service.Environment.GetUint8("PASSWORD_PARALLELISM", 2),
			service.Environment.GetUint32("PASSWORD_SALT_LENGTH", 16),
			service.Environment.GetUint32("PASSWORD_KEY_LENGTH", 32),
			service.Environment.GetByteSlice("PASSWORD_SECRET", "Nh4Qq5niSFmK8Cjmp9zfbYQGWLvqRc"),
		)
	}

	if cfg.EnvironmentConfig != nil {
		service.Environment = horizon.NewEnvironmentService(
			cfg.EnvironmentConfig.Path,
		)
	}

	if cfg.StorageConfig != nil {
		service.Storage = horizon.NewStorageImplService(
			cfg.StorageConfig.AccessKey,
			cfg.StorageConfig.SecretKey,
			cfg.StorageConfig.Endpoint,
			cfg.StorageConfig.Bucket,
			cfg.StorageConfig.Region,
			cfg.StorageConfig.Driver,
			cfg.StorageConfig.MaxFilezize,
			isStaging,
		)
	} else {
		service.Storage = horizon.NewStorageImplService(
			service.Environment.GetString("STORAGE_ACCESS_KEY", ""),
			service.Environment.GetString("STORAGE_SECRET_KEY", ""),
			service.Environment.GetString("STORAGE_URL", ""),
			service.Environment.GetString("STORAGE_BUCKET", ""),
			service.Environment.GetString("STORAGE_REGION", ""),
			service.Environment.GetString("STORAGE_DRIVER", ""),
			service.Environment.GetInt64("STORAGE_MAX_SIZE", 1001024*1024*10),
			isStaging,
		)
	}

	if cfg.SQLConfig != nil {
		service.Database = horizon.NewGormDatabase(
			cfg.SQLConfig.DSN,
			cfg.SQLConfig.MaxIdleConn,
			cfg.SQLConfig.MaxOpenConn,
			cfg.SQLConfig.MaxLifetime,
		)
	} else {
		service.Database = horizon.NewGormDatabase(
			service.Environment.GetString("DATABASE_URL", ""),
			service.Environment.GetInt("DB_MAX_IDLE_CONN", 10),
			service.Environment.GetInt("DB_MAX_OPEN_CONN", 100),
			service.Environment.GetDuration("DB_MAX_LIFETIME", 0),
		)
	}

	if cfg.CacheConfig != nil {
		service.Cache = horizon.NewHorizonCache(
			cfg.CacheConfig.Host,
			cfg.CacheConfig.Password,
			cfg.CacheConfig.Username,
			cfg.CacheConfig.Port,
		)
	} else {
		service.Cache = horizon.NewHorizonCache(
			service.Environment.GetString("REDIS_HOST", "127.0.0.1"),
			service.Environment.GetString("REDIS_PASSWORD", "password"),
			service.Environment.GetString("REDIS_USERNAME", "default"),
			service.Environment.GetInt("REDIS_PORT", 6379),
		)
	}

	if cfg.OTPServiceConfig != nil {
		service.OTP = horizon.NewHorizonOTP(
			cfg.OTPServiceConfig.Secret,
			service.Cache,
			service.Security,
		)
	} else {
		service.OTP = horizon.NewHorizonOTP(
			service.Environment.GetByteSlice("OTP_SECRET", "6D90qhBCfeDhVPewzED22XCqhtUJKR"),
			service.Cache,
			service.Security,
		)
	}
	if cfg.SMSServiceConfig != nil {
		service.SMS = horizon.NewSMS(
			cfg.SMSServiceConfig.AccountSID,
			cfg.SMSServiceConfig.AuthToken,
			cfg.SMSServiceConfig.Sender,
			cfg.SMSServiceConfig.MaxChars,
		)
	} else {
		service.SMS = horizon.NewSMS(
			service.Environment.GetString("TWILIO_ACCOUNT_SID", ""),
			service.Environment.GetString("TWILIO_AUTH_TOKEN", ""),
			service.Environment.GetString("TWILIO_SENDER", ""),
			service.Environment.GetInt32("TWILIO_MAX_CHARACTERS", 160),
		)
	}
	if cfg.SMTPServiceConfig != nil {
		service.SMTP = horizon.NewSMTP(
			cfg.SMTPServiceConfig.Host,
			cfg.SMTPServiceConfig.Port,
			cfg.SMTPServiceConfig.Username,
			cfg.SMTPServiceConfig.Password,
			cfg.SMTPServiceConfig.From,
		)
	} else {
		service.SMTP = horizon.NewSMTP(
			service.Environment.GetString("SMTP_HOST", "127.0.0.1"),
			service.Environment.GetInt("SMTP_PORT", 1025),
			service.Environment.GetString("SMTP_USERNAME", ""),
			service.Environment.GetString("SMTP_PASSWORD", ""),
			service.Environment.GetString("SMTP_FROM", "dev@local.test"),
		)
	}

	service.Cron = horizon.NewSchedule()
	service.QR = horizon.NewHorizonQRService(service.Security)
	return service
}

func (h *HorizonService) printStatus(service string, status string) {
	switch status {
	case "init":
		h.Logger.Info("Initializing service", zap.String("service", service))
		_ = os.Stdout.Sync()
	case "ok":
		h.Logger.Info("Service initialized successfully", zap.String("service", service))
		_ = os.Stdout.Sync()
	case "fail":
		h.Logger.Error("Failed to initialize service", zap.String("service", service))
		_ = os.Stdout.Sync()
	}
}

// Run starts all configured services in the correct order.
func (h *HorizonService) Run(ctx context.Context) error {
	fmt.Println("≿━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━༺❀༻━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━≾")
	handlers.PrintASCIIArt()
	h.Logger.Info("Horizon App is starting...")
	delay := 3 * time.Second
	retry := 5

	if h.Broker != nil {
		h.printStatus("Broker", "init")
		if err := handlers.Retry(ctx, retry, delay, func() error {
			return h.Broker.Run(ctx)
		}); err != nil {
			h.printStatus("Broker", "fail")
			h.Logger.Error("Broker error", zap.Error(err))
			return err
		}
		h.printStatus("Broker", "ok")
	}
	if h.Cron != nil {
		h.printStatus("Cron", "init")
		if err := handlers.Retry(ctx, retry, delay, func() error {
			return h.Cron.Run(ctx)
		}); err != nil {
			h.printStatus("Cron", "fail")
			h.Logger.Error("Cron error", zap.Error(err))
			return err
		}
		h.printStatus("Cron", "ok")
	}

	if h.Cache != nil {
		h.printStatus("Cache", "init")
		if err := handlers.Retry(ctx, retry, delay, func() error {
			return h.Cache.Run(ctx)
		}); err != nil {
			h.printStatus("Cache", "fail")
			h.Logger.Error("Cache error", zap.Error(err))
			return err
		}
		h.printStatus("Cache", "ok")
	}

	if h.Storage != nil {
		h.printStatus("Storage", "init")
		if err := handlers.Retry(ctx, retry, delay, func() error {
			return h.Storage.Run(ctx)
		}); err != nil {
			h.printStatus("Storage", "fail")
			h.Logger.Error("Storage error", zap.Error(err))
			return err
		}
		h.printStatus("Storage", "ok")
	}

	if h.Database != nil {
		h.printStatus("Database", "init")
		if err := handlers.Retry(ctx, retry, delay, func() error {
			return h.Database.Run(ctx)
		}); err != nil {
			h.printStatus("Database", "fail")
			h.Logger.Error("Database error", zap.Error(err))
			return err
		}
		h.printStatus("Database", "ok")
	}

	if h.OTP != nil {
		if h.Cache == nil {
			h.Logger.Error("OTP service requires a cache service")
			return eris.New("OTP service requires a cache service")
		}
		if h.Security == nil {
			h.Logger.Error("OTP service requires a security service")
			return eris.New("OTP service requires a security service")
		}
	}

	if h.SMS != nil {
		h.printStatus("SMS", "init")
		if err := handlers.Retry(ctx, retry, delay, func() error {
			return h.SMS.Run(ctx)
		}); err != nil {
			h.printStatus("SMS", "fail")
			h.Logger.Error("SMS error", zap.Error(err))
			return err
		}
		h.printStatus("SMS", "ok")
	}

	if h.SMTP != nil {
		h.printStatus("SMTP", "init")
		if err := handlers.Retry(ctx, retry, delay, func() error {
			return h.SMTP.Run(ctx)
		}); err != nil {
			h.printStatus("SMTP", "fail")
			h.Logger.Error("SMTP error", zap.Error(err))
			return err
		}
		h.printStatus("SMTP", "ok")
	}

	if h.Request != nil {
		h.printStatus("Request", "init")
		if err := handlers.Retry(ctx, retry, delay, func() error {
			return h.Request.Run(ctx)
		}); err != nil {
			h.printStatus("Request", "fail")
			h.Logger.Error("Request error", zap.Error(err))
			return err
		}
		h.printStatus("Request", "ok")
	}

	h.Logger.Info("Horizon App Started")
	fmt.Println("≿━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━༺❀༻━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━≾")
	return nil
}

// Stop gracefully shuts down all running services in reverse order.
func (h *HorizonService) Stop(ctx context.Context) error {
	if h.Request != nil {
		if err := h.Request.Stop(ctx); err != nil {
			return err
		}
	}
	if h.SMTP != nil {
		if err := h.SMTP.Stop(ctx); err != nil {
			return err
		}
	}
	if h.SMS != nil {
		if err := h.SMS.Stop(ctx); err != nil {
			return err
		}
	}

	if h.Cron != nil {
		if err := h.Cron.Stop(ctx); err != nil {
			return err
		}
	}
	if h.Broker != nil {
		if err := h.Broker.Stop(ctx); err != nil {
			return err
		}
	}
	if h.Cache != nil {
		if err := h.Cache.Stop(ctx); err != nil {
			return err
		}
	}
	if h.Storage != nil {
		if err := h.Storage.Stop(ctx); err != nil {
			return err
		}
	}
	if h.Database != nil {
		if err := h.Database.Stop(ctx); err != nil {
			return err
		}
	}
	if h.Broker != nil {
		if err := h.Broker.Stop(ctx); err != nil {
			return err
		}
	}

	return nil
}

// RunDatabase starts the database service.
func (h *HorizonService) RunDatabase(ctx context.Context) error {
	h.Logger.Info("Starting Database Service...")
	delay := 3 * time.Second
	retry := 5

	if h.Database != nil {
		if err := handlers.Retry(ctx, retry, delay, func() error {
			return h.Database.Run(ctx)
		}); err != nil {
			h.Logger.Error("Failed to start Database Service", zap.Error(err))
			return err
		}
	}

	h.Logger.Info("Database Service Started Successfully")
	return nil
}

// StopDatabase stops the database service.
func (h *HorizonService) StopDatabase(ctx context.Context) error {
	h.Logger.Info("Stopping Database Service...")

	if h.Database != nil {
		if err := h.Database.Stop(ctx); err != nil {
			h.Logger.Error("Failed to stop Database Service", zap.Error(err))
			return err
		}
	}

	h.Logger.Info("Database Service Stopped Successfully")
	return nil
}

// RunCache starts the cache service.
func (h *HorizonService) RunCache(ctx context.Context) error {
	h.Logger.Info("Starting Cache Service...")
	delay := 3 * time.Second
	retry := 5

	if h.Cache != nil {
		if err := handlers.Retry(ctx, retry, delay, func() error {
			return h.Cache.Run(ctx)
		}); err != nil {
			h.Logger.Error("Failed to start Cache Service", zap.Error(err))
			return err
		}
	}

	h.Logger.Info("Cache Service Started Successfully")
	return nil
}

// StopCache stops the cache service.
func (h *HorizonService) StopCache(ctx context.Context) error {
	h.Logger.Info("Stopping Cache Service...")

	if h.Cache != nil {
		if err := h.Cache.Stop(ctx); err != nil {
			h.Logger.Error("Failed to stop Cache Service", zap.Error(err))
			return err
		}
	}

	h.Logger.Info("Cache Service Stopped Successfully")
	return nil
}

// RunStorage starts the storage service.
func (h *HorizonService) RunStorage(ctx context.Context) error {
	h.Logger.Info("Starting Storage Service...")
	delay := 3 * time.Second
	retry := 5

	if h.Storage != nil {
		if err := handlers.Retry(ctx, retry, delay, func() error {
			return h.Storage.Run(ctx)
		}); err != nil {
			h.Logger.Error("Failed to start Storage Service", zap.Error(err))
			return err
		}
	}

	h.Logger.Info("Storage Service Started Successfully")
	return nil
}

// StopStorage stops the storage service.
func (h *HorizonService) StopStorage(ctx context.Context) error {
	h.Logger.Info("Stopping Storage Service...")

	if h.Storage != nil {
		if err := h.Storage.Stop(ctx); err != nil {
			h.Logger.Error("Failed to stop Storage Service", zap.Error(err))
			return err
		}
	}

	h.Logger.Info("Storage Service Stopped Successfully")
	return nil
}

// RunBroker starts the message broker service.
func (h *HorizonService) RunBroker(ctx context.Context) error {
	h.Logger.Info("Starting Broker Service...")
	delay := 3 * time.Second
	retry := 5

	if h.Broker != nil {
		if err := handlers.Retry(ctx, retry, delay, func() error {
			return h.Broker.Run(ctx)
		}); err != nil {
			h.Logger.Error("Failed to start Broker Service", zap.Error(err))
			return err
		}
	}

	h.Logger.Info("Broker Service Started Successfully")
	return nil
}
