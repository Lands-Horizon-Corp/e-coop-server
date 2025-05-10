package server

import (
	"github.com/labstack/echo/v4"
	"horizon.com/server/server/controller"
	"horizon.com/server/server/model"
	"horizon.com/server/server/provider"
	"horizon.com/server/server/publisher"
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
	contactUs *controller.ContactUsController,
) (*CoopServer, error) {
	return &CoopServer{
		Routes: []func(*echo.Echo){
			contactUs.APIRoutes,
			feedback.APIRoutes,
			media.APIRoutes,
			user.APIRoutes,
		},
		Migrations: []any{
			&model.Branch{},
			&model.Category{},
			&model.ContactUs{},
			&model.Feedback{},
			&model.Footstep{},
			&model.InvitationCode{},
			&model.Media{},
			&model.Notification{},
			&model.OrganizationCategory{},
			&model.OrganizationDailyUsage{},
			&model.Organization{},
			&model.PermissionTemplate{},
			&model.SubscriptionPlan{},
			&model.UserOrganization{},
			&model.User{},
		},
	}, nil
}

var Modules = []any{
	NewCoopServer,

	model.NewModel,
	publisher.NewPublisher,

	repository.NewContactUsRepository,
	controller.NewContactUsController,

	repository.NewFeedbackRepository,
	controller.NewFeedbackController,

	repository.NewFootstepRepository,
	controller.NewFootstepController,

	repository.NewMediaRepository,
	controller.NewMediaController,

	repository.NewUserRepository,
	controller.NewUserController,

	// Provider
	provider.NewUserProvider,
}
