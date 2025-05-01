package horizon

import "go.uber.org/fx"

var Modules = fx.Module(
	"horizon",
	fx.Provide(
		NewHorizonConfig,
		NewHorizonAuthentication,
		NewHorizonBroadcast,
		NewHorizonCache,
		NewHorizonDatabase,
		NewHorizonLog,
		NewHorizonOTP,
		NewHorizonQR,
		NewHorizonRequest,
		NewHorizonSchedule,
		NewHorizonSecurity,
		NewHorizonSMS,
		NewHorizonSMTP,
		NewHorizonStorage,
	),
)
