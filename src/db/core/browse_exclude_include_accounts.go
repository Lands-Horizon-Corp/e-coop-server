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

func BrowseExcludeIncludeAccountsManager(service *horizon.HorizonService) *registry.Registry[
	types.BrowseExcludeIncludeAccounts, types.BrowseExcludeIncludeAccountsResponse, types.BrowseExcludeIncludeAccountsRequest] {
	return registry.NewRegistry(registry.RegistryParams[
		types.BrowseExcludeIncludeAccounts, types.BrowseExcludeIncludeAccountsResponse, types.BrowseExcludeIncludeAccountsRequest,
	]{
		Preloads: []string{
			"CreatedBy", "UpdatedBy",
			"ComputationSheet",
			"FinesAccount", "ComakerAccount", "InterestAccount", "DeliquentAccount", "IncludeExistingLoanAccount",
		},
		Database: service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return service.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *types.BrowseExcludeIncludeAccounts) *types.BrowseExcludeIncludeAccountsResponse {
			if data == nil {
				return nil
			}
			return &types.BrowseExcludeIncludeAccountsResponse{
				ID:                           data.ID,
				CreatedAt:                    data.CreatedAt.Format(time.RFC3339),
				CreatedByID:                  data.CreatedByID,
				CreatedBy:                    UserManager(service).ToModel(data.CreatedBy),
				UpdatedAt:                    data.UpdatedAt.Format(time.RFC3339),
				UpdatedByID:                  data.UpdatedByID,
				UpdatedBy:                    UserManager(service).ToModel(data.UpdatedBy),
				OrganizationID:               data.OrganizationID,
				Organization:                 OrganizationManager(service).ToModel(data.Organization),
				BranchID:                     data.BranchID,
				Branch:                       BranchManager(service).ToModel(data.Branch),
				ComputationSheetID:           data.ComputationSheetID,
				ComputationSheet:             ComputationSheetManager(service).ToModel(data.ComputationSheet),
				FinesAccountID:               data.FinesAccountID,
				FinesAccount:                 AccountManager(service).ToModel(data.FinesAccount),
				ComakerAccountID:             data.ComakerAccountID,
				ComakerAccount:               AccountManager(service).ToModel(data.ComakerAccount),
				InterestAccountID:            data.InterestAccountID,
				InterestAccount:              AccountManager(service).ToModel(data.InterestAccount),
				DeliquentAccountID:           data.DeliquentAccountID,
				DeliquentAccount:             AccountManager(service).ToModel(data.DeliquentAccount),
				IncludeExistingLoanAccountID: data.IncludeExistingLoanAccountID,
				IncludeExistingLoanAccount:   AccountManager(service).ToModel(data.IncludeExistingLoanAccount),
			}
		},
		Created: func(data *types.BrowseExcludeIncludeAccounts) registry.Topics {
			return []string{
				"browse_exclude_include_accounts.create",
				fmt.Sprintf("browse_exclude_include_accounts.create.%s", data.ID),
				fmt.Sprintf("browse_exclude_include_accounts.create.branch.%s", data.BranchID),
				fmt.Sprintf("browse_exclude_include_accounts.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *types.BrowseExcludeIncludeAccounts) registry.Topics {
			return []string{
				"browse_exclude_include_accounts.update",
				fmt.Sprintf("browse_exclude_include_accounts.update.%s", data.ID),
				fmt.Sprintf("browse_exclude_include_accounts.update.branch.%s", data.BranchID),
				fmt.Sprintf("browse_exclude_include_accounts.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *types.BrowseExcludeIncludeAccounts) registry.Topics {
			return []string{
				"browse_exclude_include_accounts.delete",
				fmt.Sprintf("browse_exclude_include_accounts.delete.%s", data.ID),
				fmt.Sprintf("browse_exclude_include_accounts.delete.branch.%s", data.BranchID),
				fmt.Sprintf("browse_exclude_include_accounts.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func BrowseExcludeIncludeAccountsCurrentBranch(context context.Context, service *horizon.HorizonService,
	organizationID uuid.UUID, branchID uuid.UUID) ([]*types.BrowseExcludeIncludeAccounts, error) {
	return BrowseExcludeIncludeAccountsManager(service).Find(context, &types.BrowseExcludeIncludeAccounts{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}
