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
			&collection.Branch{},
			&collection.Category{},
			&collection.ContactUs{},
			&collection.Feedback{},
			&collection.Footstep{},
			&collection.InvitationCode{},
			&collection.Media{},
			&collection.Notification{},
			&collection.OrganizationCategory{},
			&collection.OrganizationDailyUsage{},
			&collection.Organization{},
			&collection.PermissionTemplate{},
			&collection.SubscriptionPlan{},
			&collection.UserOrganization{},
			&collection.User{},
		},
	}, nil
}

var Modules = []any{
	NewCoopServer,

	// Branch
	collection.NewBranchCollection,

	// Category
	collection.NewCategoryCollection,

	// Contact Us
	collection.NewContactUsCollection,
	repository.NewContactUsRepository,
	controller.NewContactUsController,
	broadcast.NewContactUsBroadcast,

	// Feedback
	collection.NewFeedbackCollection,
	repository.NewFeedbackRepository,
	controller.NewFeedbackController,
	broadcast.NewFeedbackBroadcast,

	// Footstep
	collection.NewFootstepCollection,
	repository.NewFootstepRepository,
	controller.NewFootstepController,
	broadcast.NewFootstepBroadcast,

	// Generate report
	collection.NewGeneratedReportCollection,

	// Invitation Code
	collection.NewInvitationCodeCollection,

	// Media
	collection.NewMediaCollection,
	repository.NewMediaRepository,
	controller.NewMediaController,
	broadcast.NewMediaBroadcast,

	// Notification
	collection.NewNotificationCollection,

	// Organization Category
	collection.NewOrganizationCategoryCollection,

	// Organization Daily Usage
	collection.NewOrganizationDailyUsageCollection,

	// Organization
	collection.NewOrganizationCollection,

	// Permission Templates
	collection.NewPermissionTemplateCollection,

	// Subscription Plan
	collection.NewSubscriptionPlanCollection,

	// User organization
	collection.NewUserOrganizationCollection,

	// User
	collection.NewUserCollection,
	repository.NewUserRepository,
	controller.NewUserController,
	broadcast.NewUserBroadcast,
	provider.NewUserProvider,
}
