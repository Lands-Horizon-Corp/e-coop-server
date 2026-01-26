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

func JournalVoucherTagManager(service *horizon.HorizonService) *registry.Registry[
	types.JournalVoucherTag, types.JournalVoucherTagResponse, types.JournalVoucherTagRequest] {
	return registry.NewRegistry(registry.RegistryParams[
		types.JournalVoucherTag, types.JournalVoucherTagResponse, types.JournalVoucherTagRequest,
	]{
		Preloads: []string{},
		Database: service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return service.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *types.JournalVoucherTag) *types.JournalVoucherTagResponse {
			if data == nil {
				return nil
			}
			return &types.JournalVoucherTagResponse{
				ID:               data.ID,
				CreatedAt:        data.CreatedAt.Format(time.RFC3339),
				CreatedByID:      data.CreatedByID,
				CreatedBy:        UserManager(service).ToModel(data.CreatedBy),
				UpdatedAt:        data.UpdatedAt.Format(time.RFC3339),
				UpdatedByID:      data.UpdatedByID,
				UpdatedBy:        UserManager(service).ToModel(data.UpdatedBy),
				OrganizationID:   data.OrganizationID,
				Organization:     OrganizationManager(service).ToModel(data.Organization),
				BranchID:         data.BranchID,
				Branch:           BranchManager(service).ToModel(data.Branch),
				JournalVoucherID: data.JournalVoucherID,
				Name:             data.Name,
				Description:      data.Description,
				Category:         data.Category,
				Color:            data.Color,
				Icon:             data.Icon,
			}
		},

		Created: func(data *types.JournalVoucherTag) registry.Topics {
			return []string{
				"journal_voucher_tag.create",
				fmt.Sprintf("journal_voucher_tag.create.%s", data.ID),
				fmt.Sprintf("journal_voucher_tag.create.branch.%s", data.BranchID),
				fmt.Sprintf("journal_voucher_tag.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *types.JournalVoucherTag) registry.Topics {
			return []string{
				"journal_voucher_tag.update",
				fmt.Sprintf("journal_voucher_tag.update.%s", data.ID),
				fmt.Sprintf("journal_voucher_tag.update.branch.%s", data.BranchID),
				fmt.Sprintf("journal_voucher_tag.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *types.JournalVoucherTag) registry.Topics {
			return []string{
				"journal_voucher_tag.delete",
				fmt.Sprintf("journal_voucher_tag.delete.%s", data.ID),
				fmt.Sprintf("journal_voucher_tag.delete.branch.%s", data.BranchID),
				fmt.Sprintf("journal_voucher_tag.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func JournalVoucherTagCurrentBranch(context context.Context, service *horizon.HorizonService,
	organizationID uuid.UUID, branchID uuid.UUID) ([]*types.JournalVoucherTag, error) {
	return JournalVoucherTagManager(service).Find(context, &types.JournalVoucherTag{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}
