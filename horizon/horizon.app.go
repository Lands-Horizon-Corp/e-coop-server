package horizon

import (
	"context"
	"os"

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

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			steps := []struct {
				name string
				fn   func() error
			}{
				{"Log", log.Run},
				{"Schedule", schedule.Run},
				{"Cache", cache.Run},
				{"SMTP", smtp.Run},
				{"SMS", sms.Run},
				{"Storage", storage.Run},
				{"Broadcast", broadcast.Run},
				{"Database", database.Run},
				{"Database Ping", database.Ping},
			}

			for _, step := range steps {
				if err := step.fn(); err != nil {
					os.Exit(1)
					return err
				}
			}
			return nil
		},

		OnStop: func(ctx context.Context) error {
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

	return &HorizonApp{}, nil
}
