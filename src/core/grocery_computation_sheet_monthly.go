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

func GroceryComputationSheetMonthlyManager(service *horizon.HorizonService) *registry.Registry[
	types.GroceryComputationSheetMonthly, types.GroceryComputationSheetMonthlyResponse, types.GroceryComputationSheetMonthlyRequest] {
	return registry.NewRegistry(registry.RegistryParams[
		types.GroceryComputationSheetMonthly, types.GroceryComputationSheetMonthlyResponse, types.GroceryComputationSheetMonthlyRequest,
	]{
		Preloads: []string{
			"CreatedBy", "UpdatedBy", "GroceryComputationSheet",
		},
		Database: service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return service.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *types.GroceryComputationSheetMonthly) *types.GroceryComputationSheetMonthlyResponse {
			if data == nil {
				return nil
			}
			return &types.GroceryComputationSheetMonthlyResponse{
				ID:                        data.ID,
				CreatedAt:                 data.CreatedAt.Format(time.RFC3339),
				CreatedByID:               data.CreatedByID,
				CreatedBy:                 UserManager(service).ToModel(data.CreatedBy),
				UpdatedAt:                 data.UpdatedAt.Format(time.RFC3339),
				UpdatedByID:               data.UpdatedByID,
				UpdatedBy:                 UserManager(service).ToModel(data.UpdatedBy),
				OrganizationID:            data.OrganizationID,
				Organization:              OrganizationManager(service).ToModel(data.Organization),
				BranchID:                  data.BranchID,
				Branch:                    BranchManager(service).ToModel(data.Branch),
				GroceryComputationSheetID: data.GroceryComputationSheetID,
				GroceryComputationSheet:   GroceryComputationSheetManager(service).ToModel(data.GroceryComputationSheet),
				Months:                    data.Months,
				InterestRate:              data.InterestRate,
				LoanGuaranteedFundRate:    data.LoanGuaranteedFundRate,
			}
		},
		Created: func(data *types.GroceryComputationSheetMonthly) registry.Topics {
			return []string{
				"grocery_computation_sheet_monthly.create",
				fmt.Sprintf("grocery_computation_sheet_monthly.create.%s", data.ID),
				fmt.Sprintf("grocery_computation_sheet_monthly.create.branch.%s", data.BranchID),
				fmt.Sprintf("grocery_computation_sheet_monthly.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *types.GroceryComputationSheetMonthly) registry.Topics {
			return []string{
				"grocery_computation_sheet_monthly.update",
				fmt.Sprintf("grocery_computation_sheet_monthly.update.%s", data.ID),
				fmt.Sprintf("grocery_computation_sheet_monthly.update.branch.%s", data.BranchID),
				fmt.Sprintf("grocery_computation_sheet_monthly.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *types.GroceryComputationSheetMonthly) registry.Topics {
			return []string{
				"grocery_computation_sheet_monthly.delete",
				fmt.Sprintf("grocery_computation_sheet_monthly.delete.%s", data.ID),
				fmt.Sprintf("grocery_computation_sheet_monthly.delete.branch.%s", data.BranchID),
				fmt.Sprintf("grocery_computation_sheet_monthly.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func GroceryComputationSheetMonthlyCurrentBranch(context context.Context, service *horizon.HorizonService, organizationID uuid.UUID, branchID uuid.UUID) ([]*types.GroceryComputationSheetMonthly, error) {
	return GroceryComputationSheetMonthlyManager(service).Find(context, &types.GroceryComputationSheetMonthly{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}
