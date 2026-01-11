package horizon

import (
	"fmt"
	"os"

	"github.com/go-playground/validator/v10"
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
}

func NewHorizonService() *HorizonService {
	service := &HorizonService{}
	service.Validator = validator.New()
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
	service.Broker = NewMessageBrokerImpl(service.Config.NatsHost, service.Config.NatsClientPort, service.Config.NatsClient, service.Config.NatsUsername, service.Config.NatsPassword)
	service.Cache = NewCacheImpl(service.Config.RedisHost, service.Config.RedisPassword, service.Config.RedisHost, service.Config.RedisPort)
	service.Security = NewSecurityImpl(service.Config.PasswordMemory, service.Config.PasswordIterations, service.Config.PasswordParallel, service.Config.PasswordSaltLength, service.Config.PasswordKeyLength, []byte(service.Config.PasswordSecret), service.Cache)
	return service
}
