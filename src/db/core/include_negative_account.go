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

func IncludeNegativeAccountManager(service *horizon.HorizonService) *registry.Registry[
	types.IncludeNegativeAccount, types.IncludeNegativeAccountResponse, types.IncludeNegativeAccountRequest] {
	return registry.NewRegistry(registry.RegistryParams[
		types.IncludeNegativeAccount, types.IncludeNegativeAccountResponse, types.IncludeNegativeAccountRequest,
	]{
		Preloads: []string{
			"CreatedBy", "UpdatedBy",
			"ComputationSheet", "Account",
		},
		Database: service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return service.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *types.IncludeNegativeAccount) *types.IncludeNegativeAccountResponse {
			if data == nil {
				return nil
			}
			return &types.IncludeNegativeAccountResponse{
				ID:                 data.ID,
				CreatedAt:          data.CreatedAt.Format(time.RFC3339),
				CreatedByID:        data.CreatedByID,
				CreatedBy:          UserManager(service).ToModel(data.CreatedBy),
				UpdatedAt:          data.UpdatedAt.Format(time.RFC3339),
				UpdatedByID:        data.UpdatedByID,
				UpdatedBy:          UserManager(service).ToModel(data.UpdatedBy),
				OrganizationID:     data.OrganizationID,
				Organization:       OrganizationManager(service).ToModel(data.Organization),
				BranchID:           data.BranchID,
				Branch:             BranchManager(service).ToModel(data.Branch),
				ComputationSheetID: data.ComputationSheetID,
				ComputationSheet:   ComputationSheetManager(service).ToModel(data.ComputationSheet),
				AccountID:          data.AccountID,
				Account:            AccountManager(service).ToModel(data.Account),
				Description:        data.Description,
			}
		},
		Created: func(data *types.IncludeNegativeAccount) registry.Topics {
			return []string{
				"include_negative_account.create",
				fmt.Sprintf("include_negative_account.create.%s", data.ID),
				fmt.Sprintf("include_negative_account.create.branch.%s", data.BranchID),
				fmt.Sprintf("include_negative_account.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *types.IncludeNegativeAccount) registry.Topics {
			return []string{
				"include_negative_account.update",
				fmt.Sprintf("include_negative_account.update.%s", data.ID),
				fmt.Sprintf("include_negative_account.update.branch.%s", data.BranchID),
				fmt.Sprintf("include_negative_account.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *types.IncludeNegativeAccount) registry.Topics {
			return []string{
				"include_negative_account.delete",
				fmt.Sprintf("include_negative_account.delete.%s", data.ID),
				fmt.Sprintf("include_negative_account.delete.branch.%s", data.BranchID),
				fmt.Sprintf("include_negative_account.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func IncludeNegativeAccountCurrentBranch(context context.Context, service *horizon.HorizonService,
	organizationID uuid.UUID, branchID uuid.UUID) ([]*types.IncludeNegativeAccount, error) {
	return IncludeNegativeAccountManager(service).Find(context, &types.IncludeNegativeAccount{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}
