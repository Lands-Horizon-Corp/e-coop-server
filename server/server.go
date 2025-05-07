package server

import (
	"github.com/labstack/echo/v4"
	"horizon.com/server/server/broadcast"
	"horizon.com/server/server/collection"
	"horizon.com/server/server/controller"
	"horizon.com/server/server/repository"
)

type CoopServer struct {
	Routes     []func(*echo.Echo)
	Migrations []any
}

func NewCoopServer(
	feedback *controller.FeedbackController,
	media *controller.MediaController,
) (*CoopServer, error) {
	return &CoopServer{

		Routes: []func(*echo.Echo){
			feedback.APIRoutes,
			media.APIRoutes,
		},

		Migrations: []any{
			&collection.Feedback{},
			&collection.Media{},
		},
	}, nil
}

var Modules = []any{
	NewCoopServer,

	collection.NewFeedbackCollection,
	repository.NewFeedbackRepository,
	controller.NewFeedbackController,
	broadcast.NewFeedbackBroadcast,

	collection.NewMediaCollection,
	repository.NewMediaRepository,
	controller.NewMediaController,
	broadcast.NewMediaBroadcast,
}
