package main

import (
	"go.uber.org/fx"
	"horizon.com/server/horizon"
	"horizon.com/server/server"
)

func main() {
	app := fx.New(
		horizon.Modules,
		server.Modules,
	)
	app.Run()
	select {}
}
