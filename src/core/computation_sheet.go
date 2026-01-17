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

func ComputationSheetManager(service *horizon.HorizonService) *registry.Registry[
	types.ComputationSheet, types.ComputationSheetResponse, types.ComputationSheetRequest] {
	return registry.NewRegistry(registry.RegistryParams[
		types.ComputationSheet, types.ComputationSheetResponse, types.ComputationSheetRequest,
	]{
		Preloads: []string{
			"CreatedBy", "UpdatedBy", "Currency",
		},
		Database: service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return service.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *types.ComputationSheet) *types.ComputationSheetResponse {
			if data == nil {
				return nil
			}
			return &types.ComputationSheetResponse{
				ID:                data.ID,
				CreatedAt:         data.CreatedAt.Format(time.RFC3339),
				CreatedByID:       data.CreatedByID,
				CreatedBy:         UserManager(service).ToModel(data.CreatedBy),
				UpdatedAt:         data.UpdatedAt.Format(time.RFC3339),
				UpdatedByID:       data.UpdatedByID,
				UpdatedBy:         UserManager(service).ToModel(data.UpdatedBy),
				OrganizationID:    data.OrganizationID,
				Organization:      OrganizationManager(service).ToModel(data.Organization),
				BranchID:          data.BranchID,
				Branch:            BranchManager(service).ToModel(data.Branch),
				Name:              data.Name,
				Description:       data.Description,
				DeliquentAccount:  data.DeliquentAccount,
				FinesAccount:      data.FinesAccount,
				InterestAccountID: data.InterestAccountID,
				ComakerAccount:    data.ComakerAccount,
				ExistAccount:      data.ExistAccount,
				CurrencyID:        data.CurrencyID,
				Currency:          CurrencyManager(service).ToModel(data.Currency),
			}
		},
		Created: func(data *types.ComputationSheet) registry.Topics {
			return []string{
				"computation_sheet.create",
				fmt.Sprintf("computation_sheet.create.%s", data.ID),
				fmt.Sprintf("computation_sheet.create.branch.%s", data.BranchID),
				fmt.Sprintf("computation_sheet.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *types.ComputationSheet) registry.Topics {
			return []string{
				"computation_sheet.update",
				fmt.Sprintf("computation_sheet.update.%s", data.ID),
				fmt.Sprintf("computation_sheet.update.branch.%s", data.BranchID),
				fmt.Sprintf("computation_sheet.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *types.ComputationSheet) registry.Topics {
			return []string{
				"computation_sheet.delete",
				fmt.Sprintf("computation_sheet.delete.%s", data.ID),
				fmt.Sprintf("computation_sheet.delete.branch.%s", data.BranchID),
				fmt.Sprintf("computation_sheet.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func ComputationSheetCurrentBranch(context context.Context, service *horizon.HorizonService,
	organizationID uuid.UUID, branchID uuid.UUID) ([]*types.ComputationSheet, error) {
	return ComputationSheetManager(service).Find(context, &types.ComputationSheet{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}
