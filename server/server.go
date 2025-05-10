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
	repository.NewRepository,
	handler.NewHandler,

	// Provider
	provider.NewProvider,
}
