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

func InterestMaturityManager(service *horizon.HorizonService) *registry.Registry[
	types.InterestMaturity, types.InterestMaturityResponse, types.InterestMaturityRequest] {
	return registry.NewRegistry(registry.RegistryParams[
		types.InterestMaturity, types.InterestMaturityResponse, types.InterestMaturityRequest,
	]{
		Preloads: []string{
			"CreatedBy", "UpdatedBy", "Account",
		},
		Database: service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return service.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *types.InterestMaturity) *types.InterestMaturityResponse {
			if data == nil {
				return nil
			}
			return &types.InterestMaturityResponse{
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
		Created: func(data *types.InterestMaturity) registry.Topics {
			return []string{
				"interest_maturity.create",
				fmt.Sprintf("interest_maturity.create.%s", data.ID),
				fmt.Sprintf("interest_maturity.create.branch.%s", data.BranchID),
				fmt.Sprintf("interest_maturity.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *types.InterestMaturity) registry.Topics {
			return []string{
				"interest_maturity.update",
				fmt.Sprintf("interest_maturity.update.%s", data.ID),
				fmt.Sprintf("interest_maturity.update.branch.%s", data.BranchID),
				fmt.Sprintf("interest_maturity.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *types.InterestMaturity) registry.Topics {
			return []string{
				"interest_maturity.delete",
				fmt.Sprintf("interest_maturity.delete.%s", data.ID),
				fmt.Sprintf("interest_maturity.delete.branch.%s", data.BranchID),
				fmt.Sprintf("interest_maturity.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func InterestMaturityCurrentBranch(context context.Context, service *horizon.HorizonService,
	organizationID uuid.UUID, branchID uuid.UUID) ([]*types.InterestMaturity, error) {
	return InterestMaturityManager(service).Find(context, &types.InterestMaturity{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}
