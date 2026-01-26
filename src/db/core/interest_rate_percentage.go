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

func InterestRatePercentageManager(service *horizon.HorizonService) *registry.Registry[
	types.InterestRatePercentage, types.InterestRatePercentageResponse, types.InterestRatePercentageRequest] {
	return registry.NewRegistry(registry.RegistryParams[
		types.InterestRatePercentage, types.InterestRatePercentageResponse, types.InterestRatePercentageRequest,
	]{
		Preloads: []string{
			"CreatedBy", "UpdatedBy", "MemberClassificationInterestRate",
		},
		Database: service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return service.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *types.InterestRatePercentage) *types.InterestRatePercentageResponse {
			if data == nil {
				return nil
			}
			return &types.InterestRatePercentageResponse{
				ID:                                 data.ID,
				CreatedAt:                          data.CreatedAt.Format(time.RFC3339),
				CreatedByID:                        data.CreatedByID,
				CreatedBy:                          UserManager(service).ToModel(data.CreatedBy),
				UpdatedAt:                          data.UpdatedAt.Format(time.RFC3339),
				UpdatedByID:                        data.UpdatedByID,
				UpdatedBy:                          UserManager(service).ToModel(data.UpdatedBy),
				OrganizationID:                     data.OrganizationID,
				Organization:                       OrganizationManager(service).ToModel(data.Organization),
				BranchID:                           data.BranchID,
				Branch:                             BranchManager(service).ToModel(data.Branch),
				Name:                               data.Name,
				Description:                        data.Description,
				Months:                             data.Months,
				InterestRate:                       data.InterestRate,
				MemberClassificationInterestRateID: data.MemberClassificationInterestRateID,
				MemberClassificationInterestRate:   MemberClassificationInterestRateManager(service).ToModel(data.MemberClassificationInterestRate),
			}
		},
		Created: func(data *types.InterestRatePercentage) registry.Topics {
			return []string{
				"interest_rate_percentage.create",
				fmt.Sprintf("interest_rate_percentage.create.%s", data.ID),
				fmt.Sprintf("interest_rate_percentage.create.branch.%s", data.BranchID),
				fmt.Sprintf("interest_rate_percentage.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *types.InterestRatePercentage) registry.Topics {
			return []string{
				"interest_rate_percentage.update",
				fmt.Sprintf("interest_rate_percentage.update.%s", data.ID),
				fmt.Sprintf("interest_rate_percentage.update.branch.%s", data.BranchID),
				fmt.Sprintf("interest_rate_percentage.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *types.InterestRatePercentage) registry.Topics {
			return []string{
				"interest_rate_percentage.delete",
				fmt.Sprintf("interest_rate_percentage.delete.%s", data.ID),
				fmt.Sprintf("interest_rate_percentage.delete.branch.%s", data.BranchID),
				fmt.Sprintf("interest_rate_percentage.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func InterestRatePercentageCurrentBranch(context context.Context, service *horizon.HorizonService, organizationID uuid.UUID,
	branchID uuid.UUID) ([]*types.InterestRatePercentage, error) {
	return InterestRatePercentageManager(service).Find(context, &types.InterestRatePercentage{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}
