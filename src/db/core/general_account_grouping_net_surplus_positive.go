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

func GeneralAccountGroupingNetSurplusPositiveManager(service *horizon.HorizonService) *registry.Registry[
	types.GeneralAccountGroupingNetSurplusPositive, types.GeneralAccountGroupingNetSurplusPositiveResponse, types.GeneralAccountGroupingNetSurplusPositiveRequest] {
	return registry.NewRegistry(registry.RegistryParams[
		types.GeneralAccountGroupingNetSurplusPositive,
		types.GeneralAccountGroupingNetSurplusPositiveResponse,
		types.GeneralAccountGroupingNetSurplusPositiveRequest]{
		Preloads: []string{
			"CreatedBy", "UpdatedBy", "Account",
		},
		Database: service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return service.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *types.GeneralAccountGroupingNetSurplusPositive) *types.GeneralAccountGroupingNetSurplusPositiveResponse {
			if data == nil {
				return nil
			}
			return &types.GeneralAccountGroupingNetSurplusPositiveResponse{
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
				AccountID:      data.AccountID,
				Account:        AccountManager(service).ToModel(data.Account),
				Name:           data.Name,
				Description:    data.Description,
				Percentage1:    data.Percentage1,
				Percentage2:    data.Percentage2,
			}
		},
		Created: func(data *types.GeneralAccountGroupingNetSurplusPositive) registry.Topics {
			return []string{
				"general_account_grouping_net_surplus_positive.create",
				fmt.Sprintf("general_account_grouping_net_surplus_positive.create.%s", data.ID),
				fmt.Sprintf("general_account_grouping_net_surplus_positive.create.branch.%s", data.BranchID),
				fmt.Sprintf("general_account_grouping_net_surplus_positive.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *types.GeneralAccountGroupingNetSurplusPositive) registry.Topics {
			return []string{
				"general_account_grouping_net_surplus_positive.update",
				fmt.Sprintf("general_account_grouping_net_surplus_positive.update.%s", data.ID),
				fmt.Sprintf("general_account_grouping_net_surplus_positive.update.branch.%s", data.BranchID),
				fmt.Sprintf("general_account_grouping_net_surplus_positive.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *types.GeneralAccountGroupingNetSurplusPositive) registry.Topics {
			return []string{
				"general_account_grouping_net_surplus_positive.delete",
				fmt.Sprintf("general_account_grouping_net_surplus_positive.delete.%s", data.ID),
				fmt.Sprintf("general_account_grouping_net_surplus_positive.delete.branch.%s", data.BranchID),
				fmt.Sprintf("general_account_grouping_net_surplus_positive.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func GeneralAccountGroupingNetSurplusPositiveCurrentBranch(context context.Context,
	service *horizon.HorizonService, organizationID uuid.UUID, branchID uuid.UUID) ([]*types.GeneralAccountGroupingNetSurplusPositive, error) {
	return GeneralAccountGroupingNetSurplusPositiveManager(service).Find(context, &types.GeneralAccountGroupingNetSurplusPositive{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}
