package horizon

import "go.uber.org/fx"

func Horizon(callback interface{}, ctors ...interface{}) *fx.App {
	core := []interface{}{
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
	}

	allProviders := append(core, ctors...)
	allProviders = append(allProviders, NewHorizonApp)

	return fx.New(
		fx.Provide(allProviders...),
		fx.Invoke(callback),
	)
}
