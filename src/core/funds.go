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

func FundsManager(service *horizon.HorizonService) *registry.Registry[types.Funds, types.FundsResponse, types.FundsRequest] {
	return registry.NewRegistry(registry.RegistryParams[types.Funds, types.FundsResponse, types.FundsRequest]{
		Preloads: []string{"CreatedBy", "UpdatedBy", "Account"},
		Database: service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return service.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *types.Funds) *types.FundsResponse {
			if data == nil {
				return nil
			}
			return &types.FundsResponse{
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
				Type:           data.Type,
				Description:    data.Description,
				Icon:           data.Icon,
				GLBooks:        data.GLBooks,
			}
		},
		Created: func(data *types.Funds) registry.Topics {
			return []string{
				"funds.create",
				fmt.Sprintf("funds.create.%s", data.ID),
				fmt.Sprintf("funds.create.branch.%s", data.BranchID),
				fmt.Sprintf("funds.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *types.Funds) registry.Topics {
			return []string{
				"funds.update",
				fmt.Sprintf("funds.update.%s", data.ID),
				fmt.Sprintf("funds.update.branch.%s", data.BranchID),
				fmt.Sprintf("funds.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *types.Funds) registry.Topics {
			return []string{
				"funds.delete",
				fmt.Sprintf("funds.delete.%s", data.ID),
				fmt.Sprintf("funds.delete.branch.%s", data.BranchID),
				fmt.Sprintf("funds.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func FundsCurrentBranch(context context.Context, service *horizon.HorizonService,
	organizationID uuid.UUID, branchID uuid.UUID) ([]*types.Funds, error) {
	return FundsManager(service).Find(context, &types.Funds{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}
