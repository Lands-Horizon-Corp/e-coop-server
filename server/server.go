package server

import (
	"github.com/labstack/echo/v4"
	"horizon.com/server/server/broadcast"
	"horizon.com/server/server/collection"
	"horizon.com/server/server/controller"
	"horizon.com/server/server/provider"
	"horizon.com/server/server/repository"
)

type CoopServer struct {
	Routes     []func(*echo.Echo)
	Migrations []any
}

func NewCoopServer(
	feedback *controller.FeedbackController,
	media *controller.MediaController,
	user *controller.UserController,
) (*CoopServer, error) {
	return &CoopServer{

		Routes: []func(*echo.Echo){
			feedback.APIRoutes,
			media.APIRoutes,
			user.APIRoutes,
		},

		Migrations: []any{
			&collection.Branch{},
			&collection.Category{},
			&collection.ContactUs{},
			&collection.Feedback{},
			&collection.Footstep{},
			&collection.Media{},
			&collection.Notification{},
			&collection.OrganizationCategory{},
			&collection.OrganizationDailyUsage{},
			&collection.Organization{},
			&collection.SubscriptionPlan{},
			&collection.UserOrganization{},
			&collection.User{},
		},
	}, nil
}

var Modules = []any{
	NewCoopServer,

	// Feedback
	collection.NewFeedbackCollection,
	repository.NewFeedbackRepository,
	controller.NewFeedbackController,
	broadcast.NewFeedbackBroadcast,

	// Media
	collection.NewMediaCollection,
	repository.NewMediaRepository,
	controller.NewMediaController,
	broadcast.NewMediaBroadcast,

	// User
	collection.NewUserCollection,
	repository.NewUserRepository,
	controller.NewUserController,
	broadcast.NewUserBroadcast,
	provider.NewUserProvider,
}
