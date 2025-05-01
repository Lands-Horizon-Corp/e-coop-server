package horizon

import "go.uber.org/fx"

var Modules = fx.Module(
	"horizon",
	fx.Provide(
		NewHorizonAuthentication,
		NewHorizonBroadcast,
		NewHorizonCache,
		NewHorizonDatabase,
		NewHorizonLog,
		NewHorizonOtp,
		NewHorizonQr,
		NewHorizonRequest,
		NewHorizonSchedule,
		NewHorizonSecurity,
		NewHorizonSMS,
		NewHorizonSMTP,
		NewHorizonStorage,
	),
)
