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

func LoanTagManager(service *horizon.HorizonService) *registry.Registry[types.LoanTag, types.LoanTagResponse, types.LoanTagRequest] {
	return registry.NewRegistry(registry.RegistryParams[
		types.LoanTag, types.LoanTagResponse, types.LoanTagRequest,
	]{
		Preloads: []string{
			"CreatedBy", "UpdatedBy", "LoanTransaction",
		},
		Database: service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return service.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *types.LoanTag) *types.LoanTagResponse {
			if data == nil {
				return nil
			}
			return &types.LoanTagResponse{
				ID:                data.ID,
				CreatedAt:         data.CreatedAt.Format(time.RFC3339),
				CreatedByID:       data.CreatedByID,
				CreatedBy:         UserManager(service).ToModel(data.CreatedBy),
				UpdatedAt:         data.UpdatedAt.Format(time.RFC3339),
				UpdatedByID:       data.UpdatedByID,
				UpdatedBy:         UserManager(service).ToModel(data.UpdatedBy),
				OrganizationID:    data.OrganizationID,
				Organization:      OrganizationManager(service).ToModel(data.Organization),
				BranchID:          data.BranchID,
				Branch:            BranchManager(service).ToModel(data.Branch),
				LoanTransactionID: data.LoanTransactionID,
				LoanTransaction:   LoanTransactionManager(service).ToModel(data.LoanTransaction),
				Name:              data.Name,
				Description:       data.Description,
				Category:          data.Category,
				Color:             data.Color,
				Icon:              data.Icon,
			}
		},

		Created: func(data *types.LoanTag) registry.Topics {
			return []string{
				"loan_tag.create",
				fmt.Sprintf("loan_tag.create.%s", data.ID),
				fmt.Sprintf("loan_tag.create.branch.%s", data.BranchID),
				fmt.Sprintf("loan_tag.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *types.LoanTag) registry.Topics {
			return []string{
				"loan_tag.update",
				fmt.Sprintf("loan_tag.update.%s", data.ID),
				fmt.Sprintf("loan_tag.update.branch.%s", data.BranchID),
				fmt.Sprintf("loan_tag.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *types.LoanTag) registry.Topics {
			return []string{
				"loan_tag.delete",
				fmt.Sprintf("loan_tag.delete.%s", data.ID),
				fmt.Sprintf("loan_tag.delete.branch.%s", data.BranchID),
				fmt.Sprintf("loan_tag.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func LoanTagCurrentBranch(context context.Context, service *horizon.HorizonService,
	organizationID uuid.UUID, branchID uuid.UUID) ([]*types.LoanTag, error) {
	return LoanTagManager(service).Find(context, &types.LoanTag{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}
