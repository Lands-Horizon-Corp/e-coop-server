package horizon

import (
	"context"

	"go.uber.org/fx"
)

type Product struct {
	ID    int     `json:"id"`
	Name  string  `json:"name"`
	Price float64 `json:"price"`
}

func NewHorizon(
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
) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			if err := log.run(); err != nil {
				return err
			}
			if err := schedule.run(); err != nil {
				return err
			}

			if err := cache.run(); err != nil {
				return err
			}
			if err := smtp.run(); err != nil {
				return err
			}
			if err := sms.run(); err != nil {
				return err
			}
			if err := storage.run(); err != nil {
				return err
			}
			if err := broadcast.run(); err != nil {
				return err
			}
			if err := database.run(); err != nil {
				return err
			}
			return nil
		},
		OnStop: func(ctx context.Context) error {
			database.stop()
			broadcast.stop()
			storage.stop()
			sms.stop()
			smtp.stop()
			request.stop()
			cache.stop()
			schedule.stop()
			log.stop()
			return nil
		},
	})
}

var Modules = fx.Module(
	"horizon",
	fx.Provide(
		NewHorizonConfig,
		NewHorizonLog,
		NewHorizonSecurity,

		NewHorizonAuthentication,
		NewHorizonBroadcast,
		NewHorizonCache,
		NewHorizonDatabase,

		NewHorizonPrettyJSONEncoder,
		NewHorizonOTP,
		NewHorizonQR,
		NewHorizonRequest,
		NewHorizonSchedule,

		NewHorizonSMS,
		NewHorizonSMTP,

		NewHorizonStorage,
		NewHorizonReport,
	),

	fx.Invoke(NewHorizon),
)
