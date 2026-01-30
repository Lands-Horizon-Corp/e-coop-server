package core_admin

import (
	"context"
	"fmt"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/helpers"
	"github.com/Lands-Horizon-Corp/e-coop-server/horizon"
	"github.com/Lands-Horizon-Corp/e-coop-server/pkg/registry"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/types"
	"github.com/rotisserie/eris"
)

func LicenseManager(service *horizon.HorizonService) *registry.Registry[
	types.License, types.LicenseResponse, types.LicenseRequest] {

	return registry.GetRegistry(registry.RegistryParams[types.License, types.LicenseResponse, types.LicenseRequest]{
		Preloads: []string{"CreatedBy", "UpdatedBy", "UsedBy"},
		Database: service.AdminDatabase.Client(),

		Dispatch: func(topics registry.Topics, payload any) error {
			return service.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *types.License) *types.LicenseResponse {
			if data == nil {
				return nil
			}
			var expiresAt *string
			if data.ExpiresAt != nil {
				s := data.ExpiresAt.Format(time.RFC3339)
				expiresAt = &s
			}
			var usedAt *string
			if data.UsedAt != nil {
				s := data.UsedAt.Format(time.RFC3339)
				usedAt = &s
			}
			return &types.LicenseResponse{
				ID:          data.ID,
				Name:        data.Name,
				Description: data.Description,
				LicenseKey:  data.LicenseKey,
				ExpiresAt:   expiresAt,
				IsUsed:      data.IsUsed,
				UsedAt:      usedAt,
				IsRevoked:   data.IsRevoked,
				CreatedAt:   data.CreatedAt.Format(time.RFC3339),
				UpdatedAt:   data.UpdatedAt.Format(time.RFC3339),
			}
		},
		Created: func(data *types.License) registry.Topics {
			return []string{
				"license.create",
				fmt.Sprintf("license.create.%s", data.ID),
			}
		},
		Updated: func(data *types.License) registry.Topics {
			return []string{
				"license.update",
				fmt.Sprintf("license.update.%s", data.ID),
			}
		},
		Deleted: func(data *types.License) registry.Topics {
			return []string{
				"license.delete",
				fmt.Sprintf("license.delete.%s", data.ID),
			}
		},
	})
}

func licenseSeed(ctx context.Context, service *horizon.HorizonService) error {
	now := time.Now().UTC()
	licenses := []*types.License{
		{
			Name:        "Starter License",
			Description: "Starter license for testing purposes.",
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			Name:        "Pro License",
			Description: "Pro license with full features.",
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			Name:        "Enterprise License",
			Description: "Enterprise license for large organizations.",
			CreatedAt:   now,
			UpdatedAt:   now,
		},
	}

	for _, license := range licenses {
		key, err := helpers.GenerateLicenseKey()
		if err != nil {
			return eris.Wrapf(err, "failed to generate license key for %s", license.Name)
		}
		license.LicenseKey = key
		if err := LicenseManager(service).Create(ctx, license); err != nil {
			return eris.Wrapf(err, "failed to seed license %s", license.Name)
		}
	}
	return nil
}
