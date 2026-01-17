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

func CashCheckVoucherTagManager(service *horizon.HorizonService) *registry.Registry[
	types.CashCheckVoucherTag, types.CashCheckVoucherTagResponse, types.CashCheckVoucherTagRequest] {
	return registry.NewRegistry(registry.RegistryParams[
		types.CashCheckVoucherTag, types.CashCheckVoucherTagResponse, types.CashCheckVoucherTagRequest,
	]{
		Preloads: []string{
			"CreatedBy", "UpdatedBy",
		},
		Database: service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return service.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *types.CashCheckVoucherTag) *types.CashCheckVoucherTagResponse {
			if data == nil {
				return nil
			}
			return &types.CashCheckVoucherTagResponse{
				ID:                 data.ID,
				CreatedAt:          data.CreatedAt.Format(time.RFC3339),
				CreatedByID:        data.CreatedByID,
				CreatedBy:          UserManager(service).ToModel(data.CreatedBy),
				UpdatedAt:          data.UpdatedAt.Format(time.RFC3339),
				UpdatedByID:        data.UpdatedByID,
				UpdatedBy:          UserManager(service).ToModel(data.UpdatedBy),
				OrganizationID:     data.OrganizationID,
				Organization:       OrganizationManager(service).ToModel(data.Organization),
				BranchID:           data.BranchID,
				Branch:             BranchManager(service).ToModel(data.Branch),
				CashCheckVoucherID: data.CashCheckVoucherID,
				Name:               data.Name,
				Description:        data.Description,
				Category:           data.Category,
				Color:              data.Color,
				Icon:               data.Icon,
			}
		},
		Created: func(data *types.CashCheckVoucherTag) registry.Topics {
			return []string{
				"cash_check_voucher_tag.create",
				fmt.Sprintf("cash_check_voucher_tag.create.%s", data.ID),
				fmt.Sprintf("cash_check_voucher_tag.create.branch.%s", data.BranchID),
				fmt.Sprintf("cash_check_voucher_tag.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *types.CashCheckVoucherTag) registry.Topics {
			return []string{
				"cash_check_voucher_tag.create",
				fmt.Sprintf("cash_check_voucher_tag.update.%s", data.ID),
				fmt.Sprintf("cash_check_voucher_tag.update.branch.%s", data.BranchID),
				fmt.Sprintf("cash_check_voucher_tag.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *types.CashCheckVoucherTag) registry.Topics {
			return []string{
				"cash_check_voucher_tag.create",
				fmt.Sprintf("cash_check_voucher_tag.delete.%s", data.ID),
				fmt.Sprintf("cash_check_voucher_tag.delete.branch.%s", data.BranchID),
				fmt.Sprintf("cash_check_voucher_tag.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func CashCheckVoucherTagCurrentBranch(context context.Context, service *horizon.HorizonService,
	organizationID uuid.UUID, branchID uuid.UUID) ([]*types.CashCheckVoucherTag, error) {
	return CashCheckVoucherTagManager(service).Find(context, &types.CashCheckVoucherTag{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}
