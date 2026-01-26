package admin

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/horizon"
	"github.com/Lands-Horizon-Corp/e-coop-server/pkg/registry"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/types"
	"github.com/rotisserie/eris"
)

func AdminManager(service *horizon.HorizonService) *registry.Registry[
	types.Admin,
	types.AdminResponse,
	types.AdminRegisterRequest,
] {
	return registry.NewRegistry(registry.RegistryParams[types.Admin, types.AdminResponse, types.AdminRegisterRequest]{
		Preloads: []string{"Media"},
		Database: service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return service.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *types.Admin) *types.AdminResponse {
			if data == nil {
				return nil
			}
			return &types.AdminResponse{
				ID:              data.ID,
				Username:        data.Username,
				Email:           data.Email,
				IsEmailVerified: data.IsEmailVerified,
				FirstName:       data.FirstName,
				MiddleName:      data.MiddleName,
				LastName:        data.LastName,
				FullName:        data.FullName,
				Suffix:          data.Suffix,
				Description:     data.Description,
				IsActive:        data.IsActive,
				LastLoginAt:     data.LastLoginAt,
				CreatedAt:       data.CreatedAt.Format(time.RFC3339),
				UpdatedAt:       data.UpdatedAt.Format(time.RFC3339),
			}
		},

		Created: func(data *types.Admin) registry.Topics {
			return []string{
				"admin.create",
				fmt.Sprintf("admin.create.%s", data.ID),
			}
		},
		Updated: func(data *types.Admin) registry.Topics {
			return []string{
				"admin.update",
				fmt.Sprintf("admin.update.%s", data.ID),
			}
		},
		Deleted: func(data *types.Admin) registry.Topics {
			return []string{
				"admin.delete",
				fmt.Sprintf("admin.delete.%s", data.ID),
			}
		},
	})
}

func GetAdminByEmail(
	ctx context.Context,
	service *horizon.HorizonService,
	email string,
) (*types.Admin, error) {
	return AdminManager(service).FindOne(ctx, &types.Admin{Email: email})
}

func GetAdminByUsername(
	ctx context.Context,
	service *horizon.HorizonService,
	username string,
) (*types.Admin, error) {
	return AdminManager(service).FindOne(ctx, &types.Admin{Username: username})
}

func GetAdminByIdentifier(
	ctx context.Context,
	service *horizon.HorizonService,
	identifier string,
) (*types.Admin, error) {
	if strings.Contains(identifier, "@") {
		if a, err := GetAdminByEmail(ctx, service, identifier); err == nil {
			return a, nil
		}
	}
	if a, err := GetAdminByUsername(ctx, service, identifier); err == nil {
		return a, nil
	}
	return nil, eris.New("admin not found by email or username")
}
