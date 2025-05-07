package horizon

import "go.uber.org/fx"

func Horizon(callback any, ctors ...any) *fx.App {
	core := []any{
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
	return fx.New(
		fx.Provide(allProviders...),
		fx.Invoke(NewHorizonTerminal),
		fx.Invoke(callback),
	)
}
