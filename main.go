package main

import (
	"context"

	"go.uber.org/fx"
	"horizon.com/server/horizon"
	"horizon.com/server/server/broadcast"
	"horizon.com/server/server/collection"
	"horizon.com/server/server/controller"
	"horizon.com/server/server/repository"
)

func NewHorizon(
	lc fx.Lifecycle,
	config *horizon.HorizonConfig,
	log *horizon.HorizonLog,
	schedule *horizon.HorizonSchedule,
	cache *horizon.HorizonCache,
	request *horizon.HorizonRequest,
	otp *horizon.HorizonOTP,
	smtp *horizon.HorizonSMTP,
	sms *horizon.HorizonSMS,
	auth *horizon.HorizonAuthentication,
	qr *horizon.HorizonQR,
	storage *horizon.HorizonStorage,
	report *horizon.HorizonReport,
	broadcast *horizon.HorizonBroadcast,
	database *horizon.HorizonDatabase,
	feedback *controller.FeedbackController,
) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			if err := log.Run(); err != nil {
				return err
			}
			if err := schedule.Run(); err != nil {
				return err
			}
			if err := cache.Run(); err != nil {
				return err
			}
			if err := smtp.Run(); err != nil {
				return err
			}
			if err := sms.Run(); err != nil {
				return err
			}
			if err := storage.Run(); err != nil {
				return err
			}
			if err := broadcast.Run(); err != nil {
				return err
			}
			if err := database.Run(); err != nil {
				return err
			}
			if err := database.Ping(); err != nil {
				return err
			}
			if err := database.Client().AutoMigrate(&collection.Feedback{}); err != nil {
				return err
			}
			return request.Run(feedback.APIRoutes)
		},
		OnStop: func(ctx context.Context) error {
			request.Stop()
			database.Stop()
			broadcast.Stop()
			storage.Stop()
			sms.Stop()
			smtp.Stop()
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
		horizon.NewHorizonConfig,
		horizon.NewHorizonLog,
		horizon.NewHorizonSecurity,

		horizon.NewHorizonAuthentication,
		horizon.NewHorizonBroadcast,
		horizon.NewHorizonCache,
		horizon.NewHorizonDatabase,

		horizon.NewHorizonPrettyJSONEncoder,
		horizon.NewHorizonOTP,
		horizon.NewHorizonQR,
		horizon.NewHorizonRequest,
		horizon.NewHorizonSchedule,

		horizon.NewHorizonSMS,
		horizon.NewHorizonSMTP,

		horizon.NewHorizonStorage,
		horizon.NewHorizonReport,

		// Feedback
		collection.NewFeedbackCollection,
		broadcast.NewFeedbackBroadcast,
		repository.NewFeedbackRepository,
		controller.NewFeedbackController,
	),

	fx.Invoke(NewHorizon),
)

func main() {
	app := fx.New(Modules)
	app.Run()
}
