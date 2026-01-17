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

func ChargesRateByRangeOrMinimumAmountManager(service *horizon.HorizonService) *registry.Registry[
	types.ChargesRateByRangeOrMinimumAmount, types.ChargesRateByRangeOrMinimumAmountResponse, types.ChargesRateByRangeOrMinimumAmountRequest] {
	return registry.NewRegistry(registry.RegistryParams[
		types.ChargesRateByRangeOrMinimumAmount, types.ChargesRateByRangeOrMinimumAmountResponse, types.ChargesRateByRangeOrMinimumAmountRequest,
	]{
		Preloads: []string{
			"CreatedBy", "UpdatedBy", "ChargesRateScheme",
		},
		Database: service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return service.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *types.ChargesRateByRangeOrMinimumAmount) *types.ChargesRateByRangeOrMinimumAmountResponse {
			if data == nil {
				return nil
			}
			return &types.ChargesRateByRangeOrMinimumAmountResponse{
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
				From:                data.From,
				To:                  data.To,
				Charge:              data.Charge,
				Amount:              data.Amount,
				MinimumAmount:       data.MinimumAmount,
			}
		},
		Created: func(data *types.ChargesRateByRangeOrMinimumAmount) registry.Topics {
			return []string{
				"charges_rate_by_range_or_minimum_amount.create",
				fmt.Sprintf("charges_rate_by_range_or_minimum_amount.create.%s", data.ID),
				fmt.Sprintf("charges_rate_by_range_or_minimum_amount.create.branch.%s", data.BranchID),
				fmt.Sprintf("charges_rate_by_range_or_minimum_amount.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *types.ChargesRateByRangeOrMinimumAmount) registry.Topics {
			return []string{
				"charges_rate_by_range_or_minimum_amount.update",
				fmt.Sprintf("charges_rate_by_range_or_minimum_amount.update.%s", data.ID),
				fmt.Sprintf("charges_rate_by_range_or_minimum_amount.update.branch.%s", data.BranchID),
				fmt.Sprintf("charges_rate_by_range_or_minimum_amount.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *types.ChargesRateByRangeOrMinimumAmount) registry.Topics {
			return []string{
				"charges_rate_by_range_or_minimum_amount.delete",
				fmt.Sprintf("charges_rate_by_range_or_minimum_amount.delete.%s", data.ID),
				fmt.Sprintf("charges_rate_by_range_or_minimum_amount.delete.branch.%s", data.BranchID),
				fmt.Sprintf("charges_rate_by_range_or_minimum_amount.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func ChargesRateByRangeOrMinimumAmountCurrentBranch(context context.Context, service *horizon.HorizonService,
	organizationID uuid.UUID, branchID uuid.UUID) ([]*types.ChargesRateByRangeOrMinimumAmount, error) {
	return ChargesRateByRangeOrMinimumAmountManager(service).Find(context, &types.ChargesRateByRangeOrMinimumAmount{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}
