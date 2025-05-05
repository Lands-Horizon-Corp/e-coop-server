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
	storage *HorizonStorage,
) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			log.run()
			schedule.run()
			cache.run()
			request.run()
			smtp.run()
			sms.run()
			storage.run()
			go func() {
				value, err := storage.UploadLocalFile("logs/docker-desktop-amd64 (1).deb", func(read, total int64) {
					percent := float64(read) / float64(total) * 100
					fmt.Printf("Upload progress: %.2f%%\n", percent)
				})
				fmt.Println("---")
				fmt.Println(value)
				fmt.Println(err)
				fmt.Println("---")
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			storage.stop()
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
