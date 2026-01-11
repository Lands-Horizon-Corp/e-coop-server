package horizon

import (
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
	return &HorizonService{}
}
