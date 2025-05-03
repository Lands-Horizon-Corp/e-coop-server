package horizon

import (
	"context"

	"go.uber.org/fx"
)

func NewHorizon(
	lc fx.Lifecycle,
	log *HorizonLog,
	schedule *HorizonSchedule,
	request *HorizonRequest,
) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			log.Run()
			schedule.Run()
			request.Run()
			return nil
		},
		OnStop: func(ctx context.Context) error {

			request.Stop()
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
		NewHorizonAuthentication,
		NewHorizonBroadcast,
		NewHorizonCache,
		NewHorizonDatabase,
		NewHorizonLog,
		NewHorizonPrettyJSONEncoder,
		NewHorizonOTP,
		NewHorizonQR,
		NewHorizonRequest,
		NewHorizonSchedule,
		NewHorizonSecurity,
		NewHorizonSMS,
		NewHorizonSMTP,
		NewHorizonStorage,
	),
	fx.Invoke(NewHorizon),
)
