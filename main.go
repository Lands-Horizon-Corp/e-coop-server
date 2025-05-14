package main

import (
	"context"
	"log"

	"go.uber.org/fx"

	"horizon.com/server/horizon"
	"horizon.com/server/server"
	"horizon.com/server/server/seeders"

	_ "github.com/swaggo/echo-swagger/example/docs"
)

func start(
	app *horizon.HorizonApp,
	lc fx.Lifecycle,
	db *horizon.HorizonDatabase,
	req *horizon.HorizonRequest,
	coop *server.CoopServer,
	seeders *seeders.DatabaseSeeder,
) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			if err := db.Client().AutoMigrate(
				coop.Migrations...,
			); err != nil {
				return err
			}
			err := seeders.Run()
			if err != nil {
				log.Fatal(err)
				return err
			}
			return req.Run(coop.Routes...)
		},
	})
}

func main() {
	app := horizon.Horizon(
		start,
		server.Modules...,
	)
	app.Run()
}
