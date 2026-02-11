package core_admin

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/horizon"
	"github.com/Lands-Horizon-Corp/e-coop-server/pkg/registry"
	"github.com/Lands-Horizon-Corp/e-coop-server/pkg/ui"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/types"
	"github.com/rotisserie/eris"
)

func LicenseManager(service *horizon.HorizonService) *registry.Registry[
	types.License, types.LicenseResponse, types.LicenseRequest] {

	return registry.GetRegistry(registry.RegistryParams[types.License, types.LicenseResponse, types.LicenseRequest]{
		Preloads: []string{},
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
	baseName := "License Key"
	for i := 1; i <= 60; i++ {
		number := fmt.Sprintf("%03d", i)
		name := fmt.Sprintf("%s %s", baseName, number)
		license := &types.License{
			Name:        name,
			Description: fmt.Sprintf("%s number %s", baseName, number),
			LicenseKey:  fmt.Sprintf("STARTER-LICENSE-%s", number),
			CreatedAt:   now,
			UpdatedAt:   now,
		}

		if err := LicenseManager(service).Create(ctx, license); err != nil {
			return eris.Wrapf(err, "failed to seed license %s", name)
		}

		log.Println(ui.RenderSection(ui.DefaultTheme(), ui.SectionFrom("🔑 License", license)))
	}
	return nil
}
