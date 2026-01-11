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

	// // host password username port
	// service.Cache = NewCacheImpl(service.Config.)
	// service.API = NewAPIImpl(service)
	return service
}
