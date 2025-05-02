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
	fx.Invoke(func(
		hr *HorizonRequest,
		hl *HorizonLog,
	) {
		hl.Run()
		hr.Run()
	}),
)
