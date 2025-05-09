package main

import (
	"context"

	"go.uber.org/fx"

	"horizon.com/server/horizon"
	"horizon.com/server/server"

	_ "github.com/swaggo/echo-swagger/example/docs"
)

func start(
	app *horizon.HorizonApp,
	lc fx.Lifecycle,
	db *horizon.HorizonDatabase,
	req *horizon.HorizonRequest,
	coop *server.CoopServer,
) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			if err := db.Client().AutoMigrate(
				coop.Migrations...,
			); err != nil {
				return err
			}
			return req.Run(coop.Routes...)
		},
	})
}

// @title Swagger Example API
// @version 1.0
// @description This is a sample server Petstore sersssssver.
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host petstore.swagger.io
// @BasePath /v2
func main() {
	app := horizon.Horizon(
		start,
		server.Modules...,
	)
	app.Run()
}
