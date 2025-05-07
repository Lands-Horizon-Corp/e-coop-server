package horizon

import (
	"context"

	"go.uber.org/fx"
)

type HorizonApp struct{}

func NewHorizonApp(
	lc fx.Lifecycle,

	config *HorizonConfig,
	log *HorizonLog,
	schedule *HorizonSchedule,
	cache *HorizonCache,
	request *HorizonRequest,
	otp *HorizonOTP,
	smtp *HorizonSMTP,
	sms *HorizonSMS,
	auth *HorizonAuthentication,
	qr *HorizonQR,
	storage *HorizonStorage,
	report *HorizonReport,
	broadcast *HorizonBroadcast,
	database *HorizonDatabase,
) (*HorizonApp, error) {
	// lifecycle hooks for startup and shutdown
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			if err := log.Run(); err != nil {
				return err
			}
			if err := schedule.Run(); err != nil {
				return err
			}
			if err := cache.Run(); err != nil {
				return err
			}
			if err := smtp.Run(); err != nil {
				return err
			}
			if err := sms.Run(); err != nil {
				return err
			}
			if err := storage.Run(); err != nil {
				return err
			}
			if err := broadcast.Run(); err != nil {
				return err
			}
			if err := database.Run(); err != nil {
				return err
			}
			if err := database.Ping(); err != nil {
				return err
			}

			return nil
		},
		OnStop: func(ctx context.Context) error {
			// shutdown in reverse order
			request.Stop()
			database.Stop()
			broadcast.Stop()
			storage.Stop()
			sms.Stop()
			smtp.Stop()
			cache.Stop()
			schedule.Stop()
			log.Stop()
			return nil
		},
	})

	// return the fully configured app instance
	return &HorizonApp{}, nil
}
