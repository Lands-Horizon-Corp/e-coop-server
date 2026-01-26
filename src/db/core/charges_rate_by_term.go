package core

import (
	"context"
	"fmt"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/horizon"
	"github.com/Lands-Horizon-Corp/e-coop-server/pkg/registry"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/types"
	"github.com/google/uuid"
)

func ChargesRateByTermManager(service *horizon.HorizonService) *registry.Registry[
	types.ChargesRateByTerm, types.ChargesRateByTermResponse, types.ChargesRateByTermRequest] {
	return registry.NewRegistry(registry.RegistryParams[
		types.ChargesRateByTerm, types.ChargesRateByTermResponse, types.ChargesRateByTermRequest,
	]{
		Preloads: []string{
			"CreatedBy", "UpdatedBy", "ChargesRateScheme",
		},
		Database: service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return service.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *types.ChargesRateByTerm) *types.ChargesRateByTermResponse {
			if data == nil {
				return nil
			}
			return &types.ChargesRateByTermResponse{
				ID:                  data.ID,
				CreatedAt:           data.CreatedAt.Format(time.RFC3339),
				CreatedByID:         data.CreatedByID,
				CreatedBy:           UserManager(service).ToModel(data.CreatedBy),
				UpdatedAt:           data.UpdatedAt.Format(time.RFC3339),
				UpdatedByID:         data.UpdatedByID,
				UpdatedBy:           UserManager(service).ToModel(data.UpdatedBy),
				OrganizationID:      data.OrganizationID,
				Organization:        OrganizationManager(service).ToModel(data.Organization),
				BranchID:            data.BranchID,
				Branch:              BranchManager(service).ToModel(data.Branch),
				ChargesRateSchemeID: data.ChargesRateSchemeID,
				ChargesRateScheme:   ChargesRateSchemeManager(service).ToModel(data.ChargesRateScheme),
				Name:                data.Name,
				Description:         data.Description,
				ModeOfPayment:       data.ModeOfPayment,
				Rate1:               data.Rate1,
				Rate2:               data.Rate2,
				Rate3:               data.Rate3,
				Rate4:               data.Rate4,
				Rate5:               data.Rate5,
				Rate6:               data.Rate6,
				Rate7:               data.Rate7,
				Rate8:               data.Rate8,
				Rate9:               data.Rate9,
				Rate10:              data.Rate10,
				Rate11:              data.Rate11,
				Rate12:              data.Rate12,
				Rate13:              data.Rate13,
				Rate14:              data.Rate14,
				Rate15:              data.Rate15,
				Rate16:              data.Rate16,
				Rate17:              data.Rate17,
				Rate18:              data.Rate18,
				Rate19:              data.Rate19,
				Rate20:              data.Rate20,
				Rate21:              data.Rate21,
				Rate22:              data.Rate22,
			}
		},
		Created: func(data *types.ChargesRateByTerm) registry.Topics {
			return []string{
				"charges_rate_by_term.create",
				fmt.Sprintf("charges_rate_by_term.create.%s", data.ID),
				fmt.Sprintf("charges_rate_by_term.create.branch.%s", data.BranchID),
				fmt.Sprintf("charges_rate_by_term.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *types.ChargesRateByTerm) registry.Topics {
			return []string{
				"charges_rate_by_term.update",
				fmt.Sprintf("charges_rate_by_term.update.%s", data.ID),
				fmt.Sprintf("charges_rate_by_term.update.branch.%s", data.BranchID),
				fmt.Sprintf("charges_rate_by_term.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *types.ChargesRateByTerm) registry.Topics {
			return []string{
				"charges_rate_by_term.delete",
				fmt.Sprintf("charges_rate_by_term.delete.%s", data.ID),
				fmt.Sprintf("charges_rate_by_term.delete.branch.%s", data.BranchID),
				fmt.Sprintf("charges_rate_by_term.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func ChargesRateByTermCurrentBranch(context context.Context,
	service *horizon.HorizonService, organizationID uuid.UUID, branchID uuid.UUID) ([]*types.ChargesRateByTerm, error) {
	return ChargesRateByTermManager(service).Find(context, &types.ChargesRateByTerm{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}
