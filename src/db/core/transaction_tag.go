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

func TransactionTagManager(service *horizon.HorizonService) *registry.Registry[
	types.TransactionTag, types.TransactionTagResponse, types.TransactionTagRequest] {
	return registry.NewRegistry(registry.RegistryParams[
		types.TransactionTag, types.TransactionTagResponse, types.TransactionTagRequest,
	]{
		Preloads: []string{
			"CreatedBy", "UpdatedBy", "Transaction",
		},
		Database: service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return service.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *types.TransactionTag) *types.TransactionTagResponse {
			if data == nil {
				return nil
			}
			return &types.TransactionTagResponse{
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
				TransactionID:  data.TransactionID,
				Transaction:    TransactionManager(service).ToModel(data.Transaction),
				Name:           data.Name,
				Description:    data.Description,
				Category:       data.Category,
				Color:          data.Color,
				Icon:           data.Icon,
			}
		},

		Created: func(data *types.TransactionTag) registry.Topics {
			return []string{
				"transaction_tag.create",
				fmt.Sprintf("transaction_tag.create.%s", data.ID),
				fmt.Sprintf("transaction_tag.create.branch.%s", data.BranchID),
				fmt.Sprintf("transaction_tag.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *types.TransactionTag) registry.Topics {
			return []string{
				"transaction_tag.update",
				fmt.Sprintf("transaction_tag.update.%s", data.ID),
				fmt.Sprintf("transaction_tag.update.branch.%s", data.BranchID),
				fmt.Sprintf("transaction_tag.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *types.TransactionTag) registry.Topics {
			return []string{
				"transaction_tag.delete",
				fmt.Sprintf("transaction_tag.delete.%s", data.ID),
				fmt.Sprintf("transaction_tag.delete.branch.%s", data.BranchID),
				fmt.Sprintf("transaction_tag.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func TransactionTagCurrentBranch(context context.Context, service *horizon.HorizonService,
	organizationID uuid.UUID, branchID uuid.UUID) ([]*types.TransactionTag, error) {
	return TransactionTagManager(service).Find(context, &types.TransactionTag{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}
