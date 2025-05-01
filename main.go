package main

import (
	"go.uber.org/fx"
	"horizon.com/server/horizon"
)

func main() {
	app := fx.New(
		horizon.Modules,
	)

	app.Run()
}
