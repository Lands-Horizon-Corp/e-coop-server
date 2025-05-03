package horizon

import (
	"context"

	"go.uber.org/fx"
)

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
	qr *HorizonQR,
) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			log.run()
			schedule.run()
			cache.run()
			request.run()
			smtp.run()
			sms.run()
			return nil
		},
		OnStop: func(ctx context.Context) error {
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
