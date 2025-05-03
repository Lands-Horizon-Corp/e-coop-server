package horizon

import (
	"context"
	"fmt"

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

			key := "zalvendayao@gmail.com"
			val, err := otp.GenerateOTP(key)
			if err != nil {
				fmt.Println("Error generating OTP", err)
				return nil
			}
			fmt.Println("OTP", val)

			check1, err := otp.VerifyOTP(key, "23")
			if err != nil {
				fmt.Println("Check 1", err)
			}
			fmt.Println("must be error 1", check1)

			check2, err := otp.VerifyOTP(key, "no")
			if err != nil {
				fmt.Println("Check 2", err)
			}
			fmt.Println("must be error 2", check2)

			check3, err := otp.VerifyOTP(key, "no")
			if err != nil {
				fmt.Println("Check 3", err)
			}
			fmt.Println("must be error 3", check3)

			check4, err := otp.VerifyOTP(key, val)
			if err != nil {
				fmt.Println("Check 4", err)
			}
			fmt.Println("must be correct 4", check4)

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
