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

func AdjustmentTagManager(service *horizon.HorizonService) *registry.Registry[
	types.AdjustmentTag, types.AdjustmentTagResponse, types.AdjustmentTagRequest] {
	return registry.NewRegistry(registry.RegistryParams[
		types.AdjustmentTag, types.AdjustmentTagResponse, types.AdjustmentTagRequest,
	]{
		Preloads: []string{
			"CreatedBy", "UpdatedBy", "AdjustmentEntry",
		},
		Database: service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return service.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *types.AdjustmentTag) *types.AdjustmentTagResponse {
			if data == nil {
				return nil
			}
			return &types.AdjustmentTagResponse{
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
				AdjustmentEntryID: data.AdjustmentEntryID,
				AdjustmentEntry:   AdjustmentEntryManager(service).ToModel(data.AdjustmentEntry),
				Name:              data.Name,
				Description:       data.Description,
				Category:          data.Category,
				Color:             data.Color,
				Icon:              data.Icon,
			}
		},
		Created: func(data *types.AdjustmentTag) registry.Topics {
			return []string{
				"adjustment_entry_tag.create",
				fmt.Sprintf("adjustment_entry_tag.create.%s", data.ID),
				fmt.Sprintf("adjustment_entry_tag.create.branch.%s", data.BranchID),
				fmt.Sprintf("adjustment_entry_tag.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *types.AdjustmentTag) registry.Topics {
			return []string{
				"adjustment_entry_tag.update",
				fmt.Sprintf("adjustment_entry_tag.update.%s", data.ID),
				fmt.Sprintf("adjustment_entry_tag.update.branch.%s", data.BranchID),
				fmt.Sprintf("adjustment_entry_tag.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *types.AdjustmentTag) registry.Topics {
			return []string{
				"adjustment_entry_tag.delete",
				fmt.Sprintf("adjustment_entry_tag.delete.%s", data.ID),
				fmt.Sprintf("adjustment_entry_tag.delete.branch.%s", data.BranchID),
				fmt.Sprintf("adjustment_entry_tag.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func AdjustmentTagCurrentBranch(context context.Context, service *horizon.HorizonService,
	organizationID uuid.UUID, branchID uuid.UUID) ([]*types.AdjustmentTag, error) {
	return AdjustmentTagManager(service).Find(context, &types.AdjustmentTag{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}
