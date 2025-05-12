package server

import (
	"github.com/labstack/echo/v4"
	"horizon.com/server/server/handler"
	"horizon.com/server/server/model"
)

type CoopServer struct {
	Routes     []func(*echo.Echo)
	Migrations []any
}

func NewCoopServer(
	handler *handler.Handler,
) (*CoopServer, error) {
	return &CoopServer{
		Routes: []func(*echo.Echo){
			handler.Routes,
		},
		Migrations: []any{
			&model.User{},                   // ✅
			&model.SubscriptionPlan{},       // ✅
			&model.Category{},               // ✅
			&model.ContactUs{},              // ✅
			&model.Media{},                  // ✅
			&model.Organization{},           // ✅
			&model.OrganizationCategory{},   // ✅
			&model.Branch{},                 // ✅
			&model.PermissionTemplate{},     // ✅
			&model.InvitationCode{},         // ✅
			&model.Feedback{},               // ✅
			&model.Footstep{},               // ✅
			&model.UserOrganization{},       // ✅
			&model.OrganizationDailyUsage{}, // ✅
			&model.Notification{},           // ✅
		},
	}, nil
}

var Modules = []any{
	NewCoopServer,

	// All model validation and database model
	model.NewModel,

	// Collections of repository
	model.NewBranchCollection,
	model.NewCategoryCollection,
	model.NewContactUsCollection,
	model.NewFeedbackCollection,
	model.NewFootstepCollection,
	model.NewGeneratedReportCollection,
	model.NewInvitationCodeCollection,
	model.NewMediaCollection,
	model.NewNotificationCollection,
	model.NewOrganizationCategoryCollection,
	model.NewOrganizationDailyUsageCollection,
	model.NewOrganizationCollection,
	model.NewPermissionTemplateCollection,
	model.NewSubscriptionPlanCollection,
	model.NewUserOrganizationCollection,
	model.NewUserCollection,

	// handler.NewHandler,

	// Provider
	// provider.NewProvider,
}
