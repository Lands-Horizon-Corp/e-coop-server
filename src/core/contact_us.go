package core

import (
	"fmt"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/horizon"
	"github.com/Lands-Horizon-Corp/e-coop-server/pkg/registry"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/types"
)

func ContactUsManager(service *horizon.HorizonService) *registry.Registry[types.ContactUs, types.ContactUsResponse, types.ContactUsRequest] {
	return registry.NewRegistry(registry.RegistryParams[types.ContactUs, types.ContactUsResponse, types.ContactUsRequest]{
		Preloads: nil,
		Database: service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return service.Broker.Dispatch(topics, payload)
		},
		Resource: func(cu *types.ContactUs) *types.ContactUsResponse {
			if cu == nil {
				return nil
			}
			return &types.ContactUsResponse{
				ID:            cu.ID,
				FirstName:     cu.FirstName,
				LastName:      cu.LastName,
				Email:         cu.Email,
				ContactNumber: cu.ContactNumber,
				Description:   cu.Description,
				CreatedAt:     cu.CreatedAt.Format(time.RFC3339),
				UpdatedAt:     cu.UpdatedAt.Format(time.RFC3339),
			}
		},
		Created: func(data *types.ContactUs) registry.Topics {
			return []string{
				"contact_us.create",
				fmt.Sprintf("feedback.create.%s", data.ID),
			}
		},
		Deleted: func(data *types.ContactUs) registry.Topics {
			return []string{
				"contact_us.delete",
				fmt.Sprintf("feedback.delete.%s", data.ID),
			}
		},
		Updated: func(data *types.ContactUs) registry.Topics {
			return []string{
				"contact_us.update",
				fmt.Sprintf("feedback.update.%s", data.ID),
			}
		},
	})
}
