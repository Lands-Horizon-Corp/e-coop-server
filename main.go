package main

import (
	"context"
	"time"

	"github.com/lands-horizon/horizon-server/src"
	"github.com/lands-horizon/horizon-server/src/controller"
	"github.com/lands-horizon/horizon-server/src/cooperative_tokens"
	"github.com/lands-horizon/horizon-server/src/model"
	"go.uber.org/fx"
)

func main() {
	app := fx.New(
		fx.StartTimeout(10*time.Minute),
		fx.Provide(
			src.NewProvider,
			src.NewValidator,
			controller.NewController,

			cooperative_tokens.NewUserToken,
			cooperative_tokens.NewTransactionBatchToken,
			cooperative_tokens.NewUserOrganizatonToken,

			// Models
			model.NewMediaCollection,
			model.NewFeedbackCollection,
		),
		fx.Invoke(func(lc fx.Lifecycle, controller *controller.Controller, provider *src.Provider) error {
			lc.Append(fx.Hook{
				OnStart: func(ctx context.Context) error {
					controller.Routes()
					if err := provider.Service.Run(ctx); err != nil {
						return err
					}
					if err := provider.Service.Database.Client().AutoMigrate(

						&model.Feedback{},
						&model.Media{},
					); err != nil {
						return err
					}
					return nil
				},
				OnStop: func(ctx context.Context) error {
					if err := provider.Service.Stop(ctx); err != nil {
						return err
					}
					return nil
				},
			})
			return nil
		}),
	)
	app.Run()
}
