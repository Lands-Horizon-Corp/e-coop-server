package horizon

import (
	"context"

	"go.uber.org/fx"
)

func NewHorizon(
	lc fx.Lifecycle,
	log *HorizonLog,
	schedule *HorizonSchedule,
	cache *HorizonCache,
	request *HorizonRequest,
	otp *HorizonOTP,
	qr *HorizonQR,
) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			log.Run()
			schedule.Run()
			cache.Run()
			request.Run()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			request.Stop()
			cache.Stop()
			schedule.Stop()
			log.Stop()
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
