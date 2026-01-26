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

func FinesMaturityManager(service *horizon.HorizonService) *registry.Registry[types.FinesMaturity, types.FinesMaturityResponse, types.FinesMaturityRequest] {
	return registry.NewRegistry(registry.RegistryParams[
		types.FinesMaturity, types.FinesMaturityResponse, types.FinesMaturityRequest,
	]{
		Preloads: []string{
			"CreatedBy", "UpdatedBy", "Account",
		},
		Database: service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return service.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *types.FinesMaturity) *types.FinesMaturityResponse {
			if data == nil {
				return nil
			}
			return &types.FinesMaturityResponse{
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
				From:           data.From,
				To:             data.To,
				Rate:           data.Rate,
			}
		},

		Created: func(data *types.FinesMaturity) registry.Topics {
			return []string{
				"fines_maturity.create",
				fmt.Sprintf("fines_maturity.create.%s", data.ID),
				fmt.Sprintf("fines_maturity.create.branch.%s", data.BranchID),
				fmt.Sprintf("fines_maturity.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *types.FinesMaturity) registry.Topics {
			return []string{
				"fines_maturity.update",
				fmt.Sprintf("fines_maturity.update.%s", data.ID),
				fmt.Sprintf("fines_maturity.update.branch.%s", data.BranchID),
				fmt.Sprintf("fines_maturity.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *types.FinesMaturity) registry.Topics {
			return []string{
				"fines_maturity.delete",
				fmt.Sprintf("fines_maturity.delete.%s", data.ID),
				fmt.Sprintf("fines_maturity.delete.branch.%s", data.BranchID),
				fmt.Sprintf("fines_maturity.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func FinesMaturityCurrentBranch(context context.Context,
	service *horizon.HorizonService, organizationID uuid.UUID, branchID uuid.UUID) ([]*types.FinesMaturity, error) {
	return FinesMaturityManager(service).Find(context, &types.FinesMaturity{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}
