package models

import (
	"context"

	"go.uber.org/fx"
	"horizon.com/server/horizon"
	"horizon.com/server/models/broadcast"
	"horizon.com/server/models/collection"
	"horizon.com/server/models/repository"
)

func NewModel(
	lc fx.Lifecycle,
	database *horizon.HorizonDatabase,
	feedback *repository.FeedbackRepository,
) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			err := database.Ping()
			if err != nil {
				return err
			}
			err = database.Client().AutoMigrate(
				&collection.Feedback{},
			)
			if err != nil {
				return err
			}
			sample := &collection.Feedback{
				Email:        "test@example.com",
				Description:  "This is a sample feedback entry created at startup.",
				FeedbackType: "general",
			}
			if err := feedback.Create(sample); err != nil {
				return err
			}

			return nil
		},
		OnStop: func(ctx context.Context) error {
			return nil
		},
	})
}

var Modules = fx.Module(
	"models",
	fx.Provide(
		broadcast.NewFeedbackBroadcast,
		repository.NewFeedbackRepository,
	),

	fx.Invoke(NewModel),
)
