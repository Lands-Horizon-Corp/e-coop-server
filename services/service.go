package horizon_services

import (
	"context"

	"github.com/go-playground/validator/v10"
	"github.com/lands-horizon/horizon-server/services/horizon"
	"github.com/rotisserie/eris"
)

type HorizonService struct {
	Environment horizon.EnvironmentService
	Database    horizon.SQLDatabaseService
	Storage     horizon.StorageService
	Cache       horizon.CacheService
	Broker      horizon.MessageBrokerService
	Cron        horizon.SchedulerService
	Security    horizon.SecurityService
	OTP         horizon.OTPService
	SMS         horizon.SMSService
	SMTP        horizon.SMTPService
	Request     horizon.APIService
	QR          horizon.QRService
	Validator   *validator.Validate
}

type HorizonServiceConfig struct {
	EnvironmentConfig    *EnvironmentServiceConfig
	SQLConfig            *SQLServiceConfig
	StorageConfig        *StorageServiceConfig
	CacheConfig          *CacheServiceConfig
	BrokerConfig         *BrokerServiceConfig
	SecurityConfig       *SecurityServiceConfig
	OTPServiceConfig     *OTPServiceConfig
	SMSServiceConfig     *SMSServiceConfig
	SMTPServiceConfig    *SMTPServiceConfig
	RequestServiceConfig *RequestServiceConfig
}

func NewHorizonService(cfg HorizonServiceConfig) *HorizonService {
	service := &HorizonService{}
	service.Validator = validator.New()
	env := "./.env"
	if cfg.EnvironmentConfig != nil {
		env = cfg.EnvironmentConfig.Path
	}

	service.Environment = horizon.NewEnvironmentService(env)
	if cfg.RequestServiceConfig != nil {
		service.Request = horizon.NewHorizonAPIService(
			cfg.RequestServiceConfig.AppPort,
			cfg.RequestServiceConfig.MetricsPort,
			cfg.RequestServiceConfig.ClientURL,
			cfg.RequestServiceConfig.ClientName,
		)
	} else {
		service.Request = horizon.NewHorizonAPIService(
			service.Environment.GetInt("APP_PORT", 8000),
			service.Environment.GetInt("APP_METRICS_PORT", 8001),
			service.Environment.GetString("APP_CLIENT_URL", "http://localhost:3000"),
			service.Environment.GetString("APP_CLIENT_NAME", "test-client"),
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
			service.Environment.GetUint32("PASSWORD_ITERATIONS", 4),
			service.Environment.GetUint8("PASSWORD_PARALLELISM", 4),
			service.Environment.GetUint32("PASSWORD_SALT_LENGTH", 32),
			service.Environment.GetUint32("PASSWORD_KEY_LENGTH", 32),
			service.Environment.GetByteSlice("PASSWORD_SECRET", "secret"),
		)
	}

	if cfg.EnvironmentConfig != nil {
		service.Environment = horizon.NewEnvironmentService(
			cfg.EnvironmentConfig.Path,
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

	if cfg.StorageConfig != nil {
		service.Storage = horizon.NewHorizonStorageService(
			cfg.StorageConfig.AccessKey,
			cfg.StorageConfig.SecretKey,
			cfg.StorageConfig.Prefix,
			cfg.StorageConfig.Bucket,
			cfg.StorageConfig.MaxFilezize,
		)
	} else {
		service.Storage = horizon.NewHorizonStorageService(
			service.Environment.GetString("STORAGE_ACCESS_KEY", ""),
			service.Environment.GetString("STORAGE_SECRET_KEY", ""),
			service.Environment.GetString("STORAGE_PREFIX", ""),
			service.Environment.GetString("STORAGE_BUCKET", ""),
			service.Environment.GetInt64("STORAGE_MAX_SIZE", 0),
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
			service.Environment.GetString("REDIS_HOST", ""),
			service.Environment.GetString("REDIS_PASSWORD", ""),
			service.Environment.GetString("REDIS_USERNAME", ""),
			service.Environment.GetInt("REDIS_PORT", 6379),
		)
	}

	if cfg.BrokerConfig != nil {
		service.Broker = horizon.NewHorizonMessageBroker(
			cfg.BrokerConfig.Host,
			cfg.BrokerConfig.Port,
		)
	} else {
		service.Broker = horizon.NewHorizonMessageBroker(
			service.Environment.GetString("NATS_HOST", "localhost"),
			service.Environment.GetInt("NATS_CLIENT_PORT", 4222),
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
			service.Environment.GetByteSlice("OTP_SECRET", "secret-otp"),
			service.Cache,
			service.Security,
		)
	}
	if cfg.SMSServiceConfig != nil {
		service.SMS = horizon.NewHorizonSMS(
			cfg.SMSServiceConfig.AccountSID,
			cfg.SMSServiceConfig.AuthToken,
			cfg.SMSServiceConfig.Sender,
			cfg.SMSServiceConfig.MaxChars,
		)
	} else {
		service.SMS = horizon.NewHorizonSMS(
			service.Environment.GetString("TWILIO_ACCOUNT_SID", ""),
			service.Environment.GetString("TWILIO_AUTH_TOKEN", ""),
			service.Environment.GetString("TWILIO_SENDER", ""),
			service.Environment.GetInt32("TWILIO_MAX_CHARACTERS", 160),
		)
	}
	if cfg.SMTPServiceConfig != nil {
		service.SMTP = horizon.NewHorizonSMTP(
			cfg.SMTPServiceConfig.Host,
			cfg.SMTPServiceConfig.Port,
			cfg.SMTPServiceConfig.Username,
			cfg.SMTPServiceConfig.Password,
			cfg.SMTPServiceConfig.From,
		)
	} else {
		service.SMTP = horizon.NewHorizonSMTP(
			service.Environment.GetString("SMTP_HOST", ""),
			service.Environment.GetInt("SMTP_PORT", 587),
			service.Environment.GetString("SMTP_USERNAME", ""),
			service.Environment.GetString("SMTP_PASSWORD", ""),
			service.Environment.GetString("SMTP_FROM", ""),
		)
	}

	service.Cron = horizon.NewHorizonSchedule()
	service.QR = horizon.NewHorizonQRService(service.Security)
	return service
}

func (h *HorizonService) Run(ctx context.Context) error {

	if h.Cron != nil {
		if err := h.Cron.Run(ctx); err != nil {
			return err
		}
	}
	if h.Broker != nil {
		if err := h.Broker.Run(ctx); err != nil {
			return err
		}
	}
	if h.Cache != nil {
		if err := h.Cache.Run(ctx); err != nil {
			return err
		}
		if err := h.Cache.Ping(ctx); err != nil {
			return err
		}
	}
	if h.Storage != nil {
		if err := h.Storage.Run(ctx); err != nil {
			return err
		}
	}
	if h.Database != nil {
		if err := h.Database.Run(ctx); err != nil {
			return err
		}
		if err := h.Database.Ping(ctx); err != nil {
			return err
		}
	}
	if h.OTP != nil {
		if h.Cache == nil {
			return eris.New("OTP service requires a cache service")
		}
		if h.Security == nil {
			return eris.New("OTP service requires a security service")
		}
	}
	if h.SMS != nil {
		if err := h.SMS.Run(ctx); err != nil {
			return err
		}
	}
	if h.SMTP != nil {
		if err := h.SMTP.Run(ctx); err != nil {
			return err
		}
	}
	if h.Request != nil {
		if err := h.Request.Run(ctx); err != nil {
			return err
		}
	}
	return nil
}

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

	return nil
}
