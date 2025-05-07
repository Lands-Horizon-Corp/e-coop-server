package server

import (
	"github.com/labstack/echo/v4"
	"horizon.com/server/server/broadcast"
	"horizon.com/server/server/collection"
	"horizon.com/server/server/controller"
	"horizon.com/server/server/repository"
)

// CoopServer groups API routes and data migrations
// for all application modules.
type CoopServer struct {
	Routes     []func(*echo.Echo)
	Migrations []any
}

func NewCoopServer(
	feedback *controller.FeedbackController,
	media *controller.MediaController,
) (*CoopServer, error) {
	return &CoopServer{
		// API route setup functions
		Routes: []func(*echo.Echo){
			feedback.APIRoutes,
			media.APIRoutes,
		},

		// Domain models for schema migration
		Migrations: []any{
			&collection.Feedback{},
			&collection.Media{},
		},
	}, nil
}

// Modules exports DI constructors for all server modules,
// including the CoopServer factory.
var Modules = []any{
	// Core server orchestrator
	NewCoopServer,

	// Feedback module: collection, repo, controller, broadcast
	collection.NewFeedbackCollection,
	repository.NewFeedbackRepository,
	controller.NewFeedbackController,
	broadcast.NewFeedbackBroadcast,

	// Media module: collection, repo, controller, broadcast
	collection.NewMediaCollection,
	repository.NewMediaRepository,
	controller.NewMediaController,
	broadcast.NewMediaBroadcast,
}
