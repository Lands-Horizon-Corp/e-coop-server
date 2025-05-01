package main

import (
	"go.uber.org/fx"
	"horizon.com/server/horizon"
)

func main() {
	app := fx.New(
		// reusable modules
		horizon.Modules,

		// app creation
		// app creat
	)
	app.Run()
}
