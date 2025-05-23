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
			model.NewModel,

			cooperative_tokens.NewUserToken,
			cooperative_tokens.NewTransactionBatchToken,
			cooperative_tokens.NewUserOrganizatonToken,
		),
		fx.Invoke(func(lc fx.Lifecycle, controller *controller.Controller, model *model.Model, provider *src.Provider) error {
			lc.Append(fx.Hook{
				OnStart: func(ctx context.Context) error {
					if err := controller.Start(); err != nil {
						return err
					}
					if err := provider.Service.Run(ctx); err != nil {
						return err
					}
					if err := model.Start(); err != nil {
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
