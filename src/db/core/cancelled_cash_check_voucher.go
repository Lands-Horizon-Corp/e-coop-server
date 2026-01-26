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

func CancelledCashCheckVoucherManager(service *horizon.HorizonService) *registry.Registry[types.CancelledCashCheckVoucher, types.CancelledCashCheckVoucherResponse, types.CancelledCashCheckVoucherRequest] {
	return registry.NewRegistry(registry.RegistryParams[
		types.CancelledCashCheckVoucher, types.CancelledCashCheckVoucherResponse, types.CancelledCashCheckVoucherRequest,
	]{
		Preloads: []string{"CreatedBy", "UpdatedBy"},
		Database: service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return service.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *types.CancelledCashCheckVoucher) *types.CancelledCashCheckVoucherResponse {
			if data == nil {
				return nil
			}
			return &types.CancelledCashCheckVoucherResponse{
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
				CheckNumber:    data.CheckNumber,
				EntryDate:      data.EntryDate.Format(time.RFC3339),
				Description:    data.Description,
			}
		},
		Created: func(data *types.CancelledCashCheckVoucher) registry.Topics {
			return []string{
				"cancelled_cash_check_voucher.create",
				fmt.Sprintf("cancelled_cash_check_voucher.create.%s", data.ID),
				fmt.Sprintf("cancelled_cash_check_voucher.create.branch.%s", data.BranchID),
				fmt.Sprintf("cancelled_cash_check_voucher.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *types.CancelledCashCheckVoucher) registry.Topics {
			return []string{
				"cancelled_cash_check_voucher.update",
				fmt.Sprintf("cancelled_cash_check_voucher.update.%s", data.ID),
				fmt.Sprintf("cancelled_cash_check_voucher.update.branch.%s", data.BranchID),
				fmt.Sprintf("cancelled_cash_check_voucher.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *types.CancelledCashCheckVoucher) registry.Topics {
			return []string{
				"cancelled_cash_check_voucher.delete",
				fmt.Sprintf("cancelled_cash_check_voucher.delete.%s", data.ID),
				fmt.Sprintf("cancelled_cash_check_voucher.delete.branch.%s", data.BranchID),
				fmt.Sprintf("cancelled_cash_check_voucher.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func CancelledCashCheckVoucherCurrentBranch(context context.Context, service *horizon.HorizonService, organizationID uuid.UUID,
	branchID uuid.UUID) ([]*types.CancelledCashCheckVoucher, error) {
	return CancelledCashCheckVoucherManager(service).Find(context, &types.CancelledCashCheckVoucher{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}
