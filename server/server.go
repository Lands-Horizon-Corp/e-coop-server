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
			// 1. Base user table - referenced by nearly all other tables
			&model.User{}, // ✅

			// 2. Independent tables
			&model.SubscriptionPlan{}, // ✅
			&model.Category{},         // ✅
			&model.ContactUs{},        // ✅

			// 3. Media table - referenced by multiple subsequent tables
			&model.Media{}, // ✅

			// 4. Organization and its direct dependencies
			&model.Organization{},
			&model.OrganizationCategory{},

			// 5. Branch-related tables
			&model.Branch{},
			&model.PermissionTemplate{},
			&model.InvitationCode{},

			// 6. Tables needing both User and Organization/Branch
			&model.Feedback{}, // ✅
			&model.Footstep{}, // ✅
			&model.UserOrganization{},
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
