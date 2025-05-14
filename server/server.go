package server

import (
	"github.com/labstack/echo/v4"
	"horizon.com/server/server/controllers"
	"horizon.com/server/server/model"
	"horizon.com/server/server/providers"
	"horizon.com/server/server/seeders"
)

type CoopServer struct {
	Routes     []func(*echo.Echo)
	Migrations []any
}

func NewCoopServer(
	con *controllers.Controller,

) (*CoopServer, error) {
	return &CoopServer{
		Routes: []func(*echo.Echo){
			con.Routes,
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

			// Maintenantce Table member
			&model.MemberCenter{},
			&model.MemberClassification{},
			&model.MemberGender{},
			&model.MemberGroup{},
			&model.MemberOccupation{},
			&model.MemberType{},
			// End Maintenantce table member
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
	model.NewUserRatingCollection,
	// Maintenantce Table member
	model.NewMemberCenterCollection,
	model.NewMemberClassificationCollection,
	model.NewMemberGenderCollection,
	model.NewMemberGroupCollection,
	model.NewMemberOccupationCollection,
	model.NewMemberTypeCollection,
	// End Maintenantce table member
	controllers.NewController,
	seeders.NewDatabaseSeeder,

	// Provider
	providers.NewProviders,
}
