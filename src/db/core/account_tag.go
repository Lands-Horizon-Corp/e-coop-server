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

func AccountTagManager(service *horizon.HorizonService) *registry.Registry[types.AccountTag, types.AccountTagResponse, types.AccountTagRequest] {
	return registry.NewRegistry(registry.RegistryParams[
		types.AccountTag, types.AccountTagResponse, types.AccountTagRequest,
	]{
		Preloads: []string{"CreatedBy", "UpdatedBy", "Account"},
		Database: service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return service.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *types.AccountTag) *types.AccountTagResponse {
			if data == nil {
				return nil
			}
			return &types.AccountTagResponse{
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
				Category:       data.Category,
				Color:          data.Color,
				Icon:           data.Icon,
			}
		},
		Created: func(data *types.AccountTag) registry.Topics {
			return []string{
				"account_tag.create",
				fmt.Sprintf("account_tag.create.%s", data.ID),
				fmt.Sprintf("account_tag.create.branch.%s", data.BranchID),
				fmt.Sprintf("account_tag.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *types.AccountTag) registry.Topics {
			return []string{
				"account_tag.update",
				fmt.Sprintf("account_tag.update.%s", data.ID),
				fmt.Sprintf("account_tag.update.branch.%s", data.BranchID),
				fmt.Sprintf("account_tag.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *types.AccountTag) registry.Topics {
			return []string{
				"account_tag.delete",
				fmt.Sprintf("account_tag.delete.%s", data.ID),
				fmt.Sprintf("account_tag.delete.branch.%s", data.BranchID),
				fmt.Sprintf("account_tag.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func AccountTagCurrentBranch(context context.Context, service *horizon.HorizonService, organizationID uuid.UUID, branchID uuid.UUID) ([]*types.AccountTag, error) {
	return AccountTagManager(service).Find(context, &types.AccountTag{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}
