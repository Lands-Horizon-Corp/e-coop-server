package horizon

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/helpers"
	"github.com/Lands-Horizon-Corp/e-coop-server/pkg/ui"
	"github.com/go-playground/validator/v10"
	"github.com/jaswdr/faker"
	"github.com/rotisserie/eris"
	"go.uber.org/zap"
)

const delay = 3 * time.Second
const retry = 5

type HorizonService struct {
	API           *APIImpl
	Broker        *MessageBrokerImpl
	Cache         *CacheImpl
	Config        *ConfigImpl
	Database      *DatabaseImpl
	AdminDatabase *DatabaseImpl
	OTP           *OTPImpl
	QR            *QRImpl
	Schedule      *ScheduleImpl
	Security      *SecurityImpl
	SMS           *SMSImpl
	SMTP          *SMTPImpl
	Storage       *StorageImpl

	Validator *validator.Validate
	Logger    *zap.Logger
	Faker     faker.Faker
	secured   bool
}

func NewHorizonService(lifetime bool) *HorizonService {
	service := &HorizonService{}
	service.Validator = validator.New()
	service.Faker = faker.New()
	logger, err := zap.NewProduction()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to initialize zap logger: %v\n", err)
		service.Logger = zap.NewNop()
	} else {
		service.Logger = logger
	}
	service.Logger.Info("Starting HorizonService initialization...")
	service.Config, err = NewConfigImpl()
	if err != nil {
		service.Logger.Fatal("failed to load configuration", zap.Error(err))
	}
	service.secured = helpers.CleanString(service.Config.AppEnv) == "staging"
	service.Broker = NewSoketiPublisherImpl(
		service.Config.SoketiURL,
		service.Config.SoketiAppKey,
		service.Config.SoketiAppSecret,
		service.Config.SoketiAppClient,
	)

	service.Cache = NewCacheImpl(
		service.Config.RedisHost,
		service.Config.RedisPassword,
		service.Config.RedisUsername,
		service.Config.RedisPort)

	service.Security = NewSecurityImpl(
		service.Config.PasswordMemory,
		service.Config.PasswordIterations,
		service.Config.PasswordParallel,
		service.Config.PasswordSaltLength,
		service.Config.PasswordKeyLength,
		[]byte(service.Config.PasswordSecret),
		service.Cache)
	service.QR = NewQRImpl(service.Security)
	service.Storage = NewStorageImpl(
		service.Config.StorageAccessKey,
		service.Config.StorageSecretKey,
		service.Config.StorageURL,
		service.Config.StorageBucket,
		service.Config.StorageRegion,
		service.Config.StorageDriver,
		service.Config.StorageMaxSize,
		service.secured)

	service.Database = NewDatabaseImpl(
		service.Config.DatabaseURL,
		service.Config.DBMaxIdleConn,
		service.Config.DBMaxOpenConn,
		time.Duration(service.Config.DBMaxLifetimeSeconds)*time.Second)

	service.AdminDatabase = NewDatabaseImpl(
		service.Config.AdminDatabaseURL,
		service.Config.DBMaxIdleConn,
		service.Config.DBMaxOpenConn,
		time.Duration(service.Config.DBMaxLifetimeSeconds)*time.Second)

	service.OTP = NewOTPImpl(
		[]byte(service.Config.OTPSecret),
		service.Cache,
		service.Security,
		service.secured)

	service.SMS = NewSMSImpl(
		service.Config.TwilioAccountSID,
		service.Config.TwilioAuthToken,
		service.Config.TwilioSender,
		service.Config.TwilioMaxCharacters,
		service.secured)

	service.SMTP = NewSMTPImpl(
		service.Config.SMTPHost,
		service.Config.SMTPPort,
		service.Config.SMTPUsername,
		service.Config.SMTPPassword,
		service.Config.SMTPFrom,
		service.Config.AppClientName,
		service.secured)
	if lifetime {
		service.API = NewAPIImpl(
			service.Cache,
			service.Config.AppPort,
			service.secured)
	}
	return service
}

func (h *HorizonService) Run(ctx context.Context) error {
	ui.Separator()
	ui.Logo()
	h.printUIEndpoints()
	h.Logger.Info("Horizon App is starting...")
	if h.Schedule != nil {
		h.printStatusUI("Cron", "init")
		if err := helpers.Retry(ctx, retry, delay, func() error {
			return h.Schedule.Run()
		}); err != nil {
			h.printStatusUI("Cron", "fail")
			h.Logger.Error("Cron error", zap.Error(err))
			return err
		}
		h.printStatusUI("Cron", "ok")
	}

	if h.Cache != nil {
		h.printStatusUI("Cache", "init")
		if err := helpers.Retry(ctx, retry, delay, func() error {
			return h.Cache.Run(ctx)
		}); err != nil {
			h.printStatusUI("Cache", "fail")
			h.Logger.Error("Cache error", zap.Error(err))
			return err
		}
		h.printStatusUI("Cache", "ok")
	}

	if h.Storage != nil {
		h.printStatusUI("Storage", "init")
		if err := helpers.Retry(ctx, retry, delay, func() error {
			return h.Storage.Run(ctx)
		}); err != nil {
			h.printStatusUI("Storage", "fail")
			h.Logger.Error("Storage error", zap.Error(err))
			return err
		}
		h.printStatusUI("Storage", "ok")
	}

	if h.Database != nil {
		h.printStatusUI("Database", "init")
		if err := helpers.Retry(ctx, retry, delay, func() error {
			return h.Database.Run(ctx)
		}); err != nil {
			h.printStatusUI("Database", "fail")
			h.Logger.Error("Database error", zap.Error(err))
			return err
		}
		h.printStatusUI("Database", "ok")
	}

	if h.AdminDatabase != nil {
		h.printStatusUI("AdminDatabase", "init")
		if err := helpers.Retry(ctx, retry, delay, func() error {
			return h.AdminDatabase.Run(ctx)
		}); err != nil {
			h.printStatusUI("AdminDatabase", "fail")
			h.Logger.Error("AdminDatabase error", zap.Error(err))
			return err
		}
		h.printStatusUI("AdminDatabase", "ok")
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

	if h.API != nil {
		h.printStatusUI("Request", "init")
		if err := helpers.Retry(ctx, retry, delay, func() error {
			return h.API.Init()
		}); err != nil {
			h.printStatusUI("Request", "fail")
			h.Logger.Error("Request error init", zap.Error(err))
			return err
		}
		h.printStatusUI("Request", "ok")
	}

	ui.Separator()
	return nil
}

func (h *HorizonService) RunLifeTime(ctx context.Context) error {
	if h.API != nil {
		if err := helpers.Retry(ctx, retry, delay, func() error {
			return h.API.Run()
		}); err != nil {
			h.printStatusUI("Request", "fail")
			h.Logger.Error("Request error server", zap.Error(err))
			return err
		}
		h.printStatusUI("Request", "ok")
	}
	if h.Broker != nil {
		go func() {
			ticker := time.NewTicker(3 * time.Second)
			defer ticker.Stop()

			for {
				select {
				case <-ctx.Done():
					h.Logger.Info("Live mode stopped")
					return
				case <-ticker.C:
					if err := h.Broker.Publish("horizon", map[string]any{"status": "ok", "timestamp": time.Now().UTC()}); err != nil {
						h.Logger.Error("Failed to publish live-mode event", zap.Error(err))
					}
				}
			}
		}()
	}
	h.Logger.Info("Horizon App Started")
	return nil
}
func (h *HorizonService) Stop(ctx context.Context) error {
	if h.API != nil {
		if err := h.API.Stop(ctx); err != nil {
			return err
		}
	}
	if h.Schedule != nil {
		if err := h.Schedule.Stop(); err != nil {
			return err
		}
	}
	if h.Cache != nil {
		if err := h.Cache.Stop(ctx); err != nil {
			return err
		}
	}
	if h.Storage != nil {
		if err := h.Storage.Stop(); err != nil {
			return err
		}
	}
	if h.Database != nil {
		if err := h.Database.Stop(); err != nil {
			return err
		}
	}
	if h.AdminDatabase != nil {
		if err := h.AdminDatabase.Stop(); err != nil {
			return err
		}
	}
	return nil
}

func (h *HorizonService) printUIEndpoints() {
	if h.secured {
		return
	}
	ui.PrintEndpoints("🌐 Local UI Endpoints", map[string]string{
		"Mailpit":         fmt.Sprintf("http://%s:%d", h.Config.MailpitUIHost, h.Config.MailpitUIPort),
		"RedisInsight":    fmt.Sprintf("http://%s:%d", h.Config.RedisInsightHost, h.Config.RedisInsightPort),
		"PgAdmin":         fmt.Sprintf("http://%s:%d", h.Config.PgAdminHost, h.Config.PgAdminPort),
		"Storage Console": fmt.Sprintf("http://127.0.0.1:%d", h.Config.StorageConsolePort),
	})
}

func (h *HorizonService) printStatusUI(service, status string) {
	theme := ui.DefaultTheme()
	section := ui.Section{
		Title: "⚙️ Service Status",
		Rows: []ui.Row{
			{Label: "Service", Value: service},
			{Label: "Status", Value: status},
		},
	}
	log.Println("\n", ui.RenderSection(theme, section))
}
