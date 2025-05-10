package server

import (
	"github.com/labstack/echo/v4"
	"horizon.com/server/server/handler"
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
	handler *handler.Handler,
) (*CoopServer, error) {
	return &CoopServer{
		Routes: []func(*echo.Echo){
			handler.Routes,
		},
		Migrations: []any{
			&model.User{},                 // ✅
			&model.SubscriptionPlan{},     // ✅
			&model.Category{},             // ✅
			&model.ContactUs{},            // ✅
			&model.Media{},                // ✅
			&model.Organization{},         // ✅
			&model.OrganizationCategory{}, // ✅
			&model.Branch{},               // ✅
			&model.PermissionTemplate{},
			&model.InvitationCode{},   // ✅
			&model.Feedback{},         // ✅
			&model.Footstep{},         // ✅
			&model.UserOrganization{}, // ✅
			&model.OrganizationDailyUsage{},
			&model.Notification{}, // ✅
		},
	}, nil
}

var Modules = []any{
	NewCoopServer,

	model.NewModel,
	publisher.NewPublisher,
	repository.NewRepository,
	handler.NewHandler,

	// Provider
	provider.NewProvider,
}
