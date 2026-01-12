package horizon

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/helpers"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/go-playground/validator/v10"
	"github.com/jaswdr/faker"
	"github.com/rotisserie/eris"
	"go.uber.org/zap"
)

type HorizonService struct {
	API      *APIImpl
	Broker   *MessageBrokerImpl
	Cache    *CacheImpl
	Config   *ConfigImpl
	Database *DatabaseImpl
	OTP      *OTPImpl
	QR       *QRImpl
	Schedule *ScheduleImpl
	Security *SecurityImpl
	SMS      *SMSImpl
	SMTP     *SMTPImpl
	Storage  *StorageImpl

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
	service.Broker = NewMessageBrokerImpl(
		service.Config.NatsHost,
		service.Config.NatsClientPort,
		service.Config.NatsClient,
		service.Config.NatsUsername,
		service.Config.NatsPassword)

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
	fmt.Println("≿━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━༺❀༻━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━≾")
	helpers.PrintASCIIArt()
	h.printUIEndpoints()
	h.Logger.Info("Horizon App is starting...")
	delay := 3 * time.Second
	retry := 5

	if h.Broker != nil {
		h.printStatus("Broker", "init")
		if err := helpers.Retry(ctx, retry, delay, func() error {
			return h.Broker.Run()
		}); err != nil {
			h.printStatus("Broker", "fail")
			h.Logger.Error("Broker error", zap.Error(err))
			return err
		}
		h.printStatus("Broker", "ok")
	}

	if h.Schedule != nil {
		h.printStatus("Cron", "init")
		if err := helpers.Retry(ctx, retry, delay, func() error {
			return h.Schedule.Run()
		}); err != nil {
			h.printStatus("Cron", "fail")
			h.Logger.Error("Cron error", zap.Error(err))
			return err
		}
		h.printStatus("Cron", "ok")
	}

	if h.Cache != nil {
		h.printStatus("Cache", "init")
		if err := helpers.Retry(ctx, retry, delay, func() error {
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
		if err := helpers.Retry(ctx, retry, delay, func() error {
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
		if err := helpers.Retry(ctx, retry, delay, func() error {
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

	if h.API != nil {
		h.printStatus("Request", "init")
		if err := helpers.Retry(ctx, retry, delay, func() error {
			return h.API.Run()
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
func (h *HorizonService) printUIEndpoints() {
	if h.secured {
		return
	}
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("212")).
		Align(lipgloss.Center)
	cellStyle := lipgloss.NewStyle().
		Padding(0, 1)
	urlStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("39"))
	t := table.New().
		Headers("Service", "URL").
		Border(lipgloss.RoundedBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("240"))).
		StyleFunc(func(row, col int) lipgloss.Style {
			if row == 0 {
				return headerStyle
			}
			if col == 1 {
				return urlStyle
			}
			return cellStyle
		}).
		Rows(
			[]string{"Mailpit", fmt.Sprintf("http://%s:%d", h.Config.MailpitUIHost, h.Config.MailpitUIPort)},
			[]string{"RedisInsight", fmt.Sprintf("http://%s:%d", h.Config.RedisInsightHost, h.Config.RedisInsightPort)},
			[]string{"PgAdmin", fmt.Sprintf("http://%s:%d", h.Config.PgAdminHost, h.Config.PgAdminPort)},
			[]string{"Storage Console", fmt.Sprintf("http://127.0.0.1:%d", h.Config.StorageConsolePort)},
			[]string{"NATS Monitor", fmt.Sprintf("http://%s:%d", h.Config.NatsHost, h.Config.NatsMonitorPort)},
		)
	fmt.Println(t)
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
	if h.Broker != nil {
		if err := h.Broker.Stop(); err != nil {
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
	return nil
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
