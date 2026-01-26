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

func GeneralAccountGroupingNetSurplusNegativeManager(service *horizon.HorizonService) *registry.Registry[
	types.GeneralAccountGroupingNetSurplusNegative, types.GeneralAccountGroupingNetSurplusNegativeResponse, types.GeneralAccountGroupingNetSurplusNegativeRequest] {
	return registry.NewRegistry(registry.RegistryParams[
		types.GeneralAccountGroupingNetSurplusNegative,
		types.GeneralAccountGroupingNetSurplusNegativeResponse,
		types.GeneralAccountGroupingNetSurplusNegativeRequest,
	]{
		Preloads: []string{
			"CreatedBy", "UpdatedBy", "Account",
		},
		Database: service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return service.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *types.GeneralAccountGroupingNetSurplusNegative) *types.GeneralAccountGroupingNetSurplusNegativeResponse {
			if data == nil {
				return nil
			}
			return &types.GeneralAccountGroupingNetSurplusNegativeResponse{
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
		Created: func(data *types.GeneralAccountGroupingNetSurplusNegative) registry.Topics {
			return []string{
				"general_account_grouping_net_surplus_negative.create",
				fmt.Sprintf("general_account_grouping_net_surplus_negative.create.%s", data.ID),
				fmt.Sprintf("general_account_grouping_net_surplus_negative.create.branch.%s", data.BranchID),
				fmt.Sprintf("general_account_grouping_net_surplus_negative.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *types.GeneralAccountGroupingNetSurplusNegative) registry.Topics {
			return []string{
				"general_account_grouping_net_surplus_negative.update",
				fmt.Sprintf("general_account_grouping_net_surplus_negative.update.%s", data.ID),
				fmt.Sprintf("general_account_grouping_net_surplus_negative.update.branch.%s", data.BranchID),
				fmt.Sprintf("general_account_grouping_net_surplus_negative.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *types.GeneralAccountGroupingNetSurplusNegative) registry.Topics {
			return []string{
				"general_account_grouping_net_surplus_negative.delete",
				fmt.Sprintf("general_account_grouping_net_surplus_negative.delete.%s", data.ID),
				fmt.Sprintf("general_account_grouping_net_surplus_negative.delete.branch.%s", data.BranchID),
				fmt.Sprintf("general_account_grouping_net_surplus_negative.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func GeneralAccountGroupingNetSurplusNegativeCurrentBranch(context context.Context, service *horizon.HorizonService,
	organizationID uuid.UUID, branchID uuid.UUID) ([]*types.GeneralAccountGroupingNetSurplusNegative, error) {
	return GeneralAccountGroupingNetSurplusNegativeManager(service).Find(context, &types.GeneralAccountGroupingNetSurplusNegative{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}
