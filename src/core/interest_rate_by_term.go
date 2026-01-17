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

func InterestRateByTermManager(service *horizon.HorizonService) *registry.Registry[
	types.InterestRateByTerm, types.InterestRateByTermResponse, types.InterestRateByTermRequest] {
	return registry.NewRegistry(registry.RegistryParams[
		types.InterestRateByTerm, types.InterestRateByTermResponse, types.InterestRateByTermRequest,
	]{
		Preloads: []string{
			"CreatedBy", "UpdatedBy", "MemberClassificationInterestRate",
		},
		Database: service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return service.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *types.InterestRateByTerm) *types.InterestRateByTermResponse {
			if data == nil {
				return nil
			}
			return &types.InterestRateByTermResponse{
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
				Descrition:                         data.Descrition,
				MemberClassificationInterestRateID: data.MemberClassificationInterestRateID,
				MemberClassificationInterestRate:   MemberClassificationInterestRateManager(service).ToModel(data.MemberClassificationInterestRate),
			}
		},
		Created: func(data *types.InterestRateByTerm) registry.Topics {
			return []string{
				"interest_rate_by_term.create",
				fmt.Sprintf("interest_rate_by_term.create.%s", data.ID),
				fmt.Sprintf("interest_rate_by_term.create.branch.%s", data.BranchID),
				fmt.Sprintf("interest_rate_by_term.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *types.InterestRateByTerm) registry.Topics {
			return []string{
				"interest_rate_by_term.update",
				fmt.Sprintf("interest_rate_by_term.update.%s", data.ID),
				fmt.Sprintf("interest_rate_by_term.update.branch.%s", data.BranchID),
				fmt.Sprintf("interest_rate_by_term.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *types.InterestRateByTerm) registry.Topics {
			return []string{
				"interest_rate_by_term.delete",
				fmt.Sprintf("interest_rate_by_term.delete.%s", data.ID),
				fmt.Sprintf("interest_rate_by_term.delete.branch.%s", data.BranchID),
				fmt.Sprintf("interest_rate_by_term.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func InterestRateByTermCurrentBranch(context context.Context, service *horizon.HorizonService,
	organizationID uuid.UUID, branchID uuid.UUID) ([]*types.InterestRateByTerm, error) {
	return InterestRateByTermManager(service).Find(context, &types.InterestRateByTerm{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}
