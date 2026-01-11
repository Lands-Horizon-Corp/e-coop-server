package main

import (
	"context"

	"github.com/Lands-Horizon-Corp/e-coop-server/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/horizon"
)

func main() {
	service := horizon.NewHorizonService()
	if err := service.Run(context.Background()); err != nil {
		panic(err)
	}

	core := core.NewCore(service)
	if err := core.Start(); err != nil {
		panic(err)
	}
}
