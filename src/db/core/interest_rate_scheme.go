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

func InterestRateSchemeManager(service *horizon.HorizonService) *registry.Registry[
	types.InterestRateScheme, types.InterestRateSchemeResponse, types.InterestRateSchemeRequest] {
	return registry.NewRegistry(registry.RegistryParams[
		types.InterestRateScheme, types.InterestRateSchemeResponse, types.InterestRateSchemeRequest,
	]{
		Preloads: []string{
			"CreatedBy", "UpdatedBy",
		},
		Database: service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return service.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *types.InterestRateScheme) *types.InterestRateSchemeResponse {
			if data == nil {
				return nil
			}
			return &types.InterestRateSchemeResponse{
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
				Name:           data.Name,
				Description:    data.Description,
			}
		},
		Created: func(data *types.InterestRateScheme) registry.Topics {
			return []string{
				"interest_rate_scheme.create",
				fmt.Sprintf("interest_rate_scheme.create.%s", data.ID),
				fmt.Sprintf("interest_rate_scheme.create.branch.%s", data.BranchID),
				fmt.Sprintf("interest_rate_scheme.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *types.InterestRateScheme) registry.Topics {
			return []string{
				"interest_rate_scheme.update",
				fmt.Sprintf("interest_rate_scheme.update.%s", data.ID),
				fmt.Sprintf("interest_rate_scheme.update.branch.%s", data.BranchID),
				fmt.Sprintf("interest_rate_scheme.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *types.InterestRateScheme) registry.Topics {
			return []string{
				"interest_rate_scheme.delete",
				fmt.Sprintf("interest_rate_scheme.delete.%s", data.ID),
				fmt.Sprintf("interest_rate_scheme.delete.branch.%s", data.BranchID),
				fmt.Sprintf("interest_rate_scheme.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func InterestRateSchemeCurrentBranch(context context.Context, service *horizon.HorizonService,
	organizationID uuid.UUID, branchID uuid.UUID) ([]*types.InterestRateScheme, error) {
	return InterestRateSchemeManager(service).Find(context, &types.InterestRateScheme{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}
