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
	qr *HorizonQR,
) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			err := log.Run()
			if err != nil {
				return err
			}
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
	),
	fx.Invoke(NewHorizon),
)
