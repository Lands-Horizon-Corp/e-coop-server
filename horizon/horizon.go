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
	smtp *HorizonSMTP,
	qr *HorizonQR,
) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			log.Run()
			schedule.Run()
			cache.Run()
			request.Run()

			err := smtp.Send(&SMTPRequest{
				To:      "recipient@example.com",
				Subject: "Test Email",
				Body:    "<h1>Hello from Mailpit!</h1><p>This is a test.</p>",
			})
			if err != nil {
				fmt.Printf("❌ Failed to send email: %v\n", err)

			} else {
				fmt.Println("✅ Email sent successfully!")
			}
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
