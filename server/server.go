package server

import (
	"github.com/labstack/echo/v4"
	"horizon.com/server/server/controllers"
	"horizon.com/server/server/model"
)

type CoopServer struct {
	Routes     []func(*echo.Echo)
	Migrations []any
}

func NewCoopServer(
	conttroller *controllers.Controller,
) (*CoopServer, error) {
	return &CoopServer{
		Routes: []func(*echo.Echo){
			conttroller.Routes,
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
	controllers.NewController,

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
