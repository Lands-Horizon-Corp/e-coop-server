package horizon

import (
	"context"
	"fmt"

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
	auth *HorizonAuthentication,
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
			// ─── MANUAL SMTP LINK GENERATION & VALIDATION ────────────────────────
			// Pick a dummy claim for demonstration:
			demo := Claim{
				ID:            "demo-id-123",
				Email:         "demo@user.com",
				ContactNumber: "+639171234567",
			}

			baseURL := "https://example.com/verify"
			link, err := auth.GenerateSMSLink(baseURL, demo)
			if err != nil {
				fmt.Println("❌ GenerateSMTPLink error:", err)
			} else {
				fmt.Println("✅ Generated SMTP link:", link)
			}

			parsed, err := auth.ValidateSMTPLink(link)
			if err != nil {
				fmt.Println("❌ ValidateSMTPLink error:", err)
			} else {
				fmt.Printf("✅ Parsed SMTP claim: ID=%s, Email=%s, Contact=%s\n",
					parsed.ID, parsed.Email, parsed.ContactNumber)
			}
			// password := "password sample 2344"
			// hashed, err := auth.Password(password)
			// if err != nil {
			// 	fmt.Printf("error %v", err)
			// }
			// fmt.Println(hashed)
			// fmt.Println("-----")
			// fmt.Println(" confirm %t", auth.VerifyPassword(hashed, "password sample 234"))
			// fmt.Println("-----")
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
