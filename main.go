package main

import (
	"context"
	"time"

	"github.com/lands-horizon/horizon-server/seeder"
	"github.com/lands-horizon/horizon-server/src"
	"github.com/lands-horizon/horizon-server/src/controller"
	"github.com/lands-horizon/horizon-server/src/cooperative_tokens"
	"github.com/lands-horizon/horizon-server/src/event"
	"github.com/lands-horizon/horizon-server/src/model"

	"go.uber.org/fx"
)

func main() {
	app := fx.New(
		// Set extended startup timeout
		fx.StartTimeout(30*time.Minute),

		// Provide application dependencies
		fx.Provide(
			src.NewProvider,
			src.NewValidator,
			model.NewModel,
			controller.NewController,
			event.NewEvent,
			seeder.NewSeeder,

			cooperative_tokens.NewUserToken,
			cooperative_tokens.NewTransactionBatchToken,
			cooperative_tokens.NewUserOrganizatonToken,
		),

		// Invoke the startup sequence
		fx.Invoke(func(
			lc fx.Lifecycle,
			ctrl *controller.Controller,
			mod *model.Model,
			prov *src.Provider,
			seed *seeder.Seeder,
		) error {
			// Register lifecycle hooks
			lc.Append(fx.Hook{
				OnStart: func(ctx context.Context) error {
					// Start each component in order
					if err := ctrl.Start(); err != nil {
						return err
					}
					if err := prov.Service.Run(ctx); err != nil {
						return err
					}
					if err := mod.Start(); err != nil {
						return err
					}
					if err := seed.Run(ctx); err != nil {
						return err
					}
					return nil
				},

				OnStop: func(ctx context.Context) error {
					// Gracefully stop the service
					return prov.Service.Stop(ctx)
				},
			})

			return nil
		}),
	)

	app.Run()
}
