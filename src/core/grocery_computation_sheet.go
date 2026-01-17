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

func GroceryComputationSheetManager(service *horizon.HorizonService) *registry.Registry[
	types.GroceryComputationSheet, types.GroceryComputationSheetResponse, types.GroceryComputationSheetRequest] {
	return registry.NewRegistry(registry.RegistryParams[
		types.GroceryComputationSheet, types.GroceryComputationSheetResponse, types.GroceryComputationSheetRequest,
	]{
		Preloads: []string{
			"CreatedBy", "UpdatedBy",
		},
		Database: service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return service.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *types.GroceryComputationSheet) *types.GroceryComputationSheetResponse {
			if data == nil {
				return nil
			}
			return &types.GroceryComputationSheetResponse{
				ID:             data.ID,
				CreatedAt:      data.CreatedAt.Format(time.RFC3339),
				CreatedByID:    data.CreatedByID,
				CreatedBy:      UserManager(service).ToModel(data.CreatedBy),
				UpdatedAt:      data.UpdatedAt.Format(time.RFC3339),
				UpdatedByID:    data.UpdatedByID,
				UpdatedBy:      UserManager(service).ToModel(data.UpdatedBy),
				OrganizationID: data.OrganizationID,
				Organization:   OrganizationManager(service).ToModel(data.Organization),
				BranchID:       data.BranchID,
				Branch:         BranchManager(service).ToModel(data.Branch),
				SchemeNumber:   data.SchemeNumber,
				Description:    data.Description,
			}
		},
		Created: func(data *types.GroceryComputationSheet) registry.Topics {
			return []string{
				"grocery_computation_sheet.create",
				fmt.Sprintf("grocery_computation_sheet.create.%s", data.ID),
				fmt.Sprintf("grocery_computation_sheet.create.branch.%s", data.BranchID),
				fmt.Sprintf("grocery_computation_sheet.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *types.GroceryComputationSheet) registry.Topics {
			return []string{
				"grocery_computation_sheet.update",
				fmt.Sprintf("grocery_computation_sheet.update.%s", data.ID),
				fmt.Sprintf("grocery_computation_sheet.update.branch.%s", data.BranchID),
				fmt.Sprintf("grocery_computation_sheet.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *types.GroceryComputationSheet) registry.Topics {
			return []string{
				"grocery_computation_sheet.delete",
				fmt.Sprintf("grocery_computation_sheet.delete.%s", data.ID),
				fmt.Sprintf("grocery_computation_sheet.delete.branch.%s", data.BranchID),
				fmt.Sprintf("grocery_computation_sheet.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func GroceryComputationSheetCurrentBranch(context context.Context, service *horizon.HorizonService,
	organizationID uuid.UUID, branchID uuid.UUID) ([]*types.GroceryComputationSheet, error) {
	return GroceryComputationSheetManager(service).Find(context, &types.GroceryComputationSheet{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}
