package server

import (
	"context"

	"go.uber.org/fx"
	"horizon.com/server/horizon"
	"horizon.com/server/server/broadcast"
	"horizon.com/server/server/collection"
	"horizon.com/server/server/repository"
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
			// Migration here
			err = database.Client().AutoMigrate(
				&collection.Feedback{},
			)
			if err != nil {
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
	"server",
	fx.Provide(
		broadcast.NewFeedbackBroadcast,
		repository.NewFeedbackRepository,
	),

	fx.Invoke(NewModel),
)
