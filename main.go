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

func main() {
	app := horizon.Horizon(
		func(lc fx.Lifecycle, db *horizon.HorizonDatabase,
			req *horizon.HorizonRequest,
			fb *controller.FeedbackController,
			md *controller.MediaController,
		) {
			lc.Append(fx.Hook{
				OnStart: func(ctx context.Context) error {
					if err := db.Client().AutoMigrate(
						&collection.Feedback{},
						&collection.Media{},
					); err != nil {
						return err
					}
					return req.Run(
						fb.APIRoutes,
						md.APIRoutes,
					)
				},
			})
		},
		// Feedback
		collection.NewFeedbackCollection,
		broadcast.NewFeedbackBroadcast,
		repository.NewFeedbackRepository,
		controller.NewFeedbackController,

		// Media
		collection.NewMediaCollection,
		broadcast.NewMediaBroadcast,
		repository.NewMediaRepository,
		controller.NewMediaController,
	)
	app.Run()
}
