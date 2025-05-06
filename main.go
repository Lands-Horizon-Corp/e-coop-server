package main

import (
	"go.uber.org/fx"
	"horizon.com/server/horizon"
	"horizon.com/server/models"
)

func main() {
	app := fx.New(
		horizon.Modules,
		models.Modules,
	)
	app.Run()
	select {}
}
