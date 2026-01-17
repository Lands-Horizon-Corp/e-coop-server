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

func VoucherPayToManager(service *horizon.HorizonService) *registry.Registry[types.VoucherPayTo, types.VoucherPayToResponse, types.VoucherPayToRequest] {
	return registry.NewRegistry(registry.RegistryParams[
		types.VoucherPayTo, types.VoucherPayToResponse, types.VoucherPayToRequest,
	]{
		Preloads: []string{
			"CreatedBy", "UpdatedBy", "Media",
		},
		Database: service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return service.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *types.VoucherPayTo) *types.VoucherPayToResponse {
			if data == nil {
				return nil
			}
			return &types.VoucherPayToResponse{
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
				MediaID:        data.MediaID,
				Media:          MediaManager(service).ToModel(data.Media),
				Description:    data.Description,
			}
		},
		Created: func(data *types.VoucherPayTo) registry.Topics {
			return []string{
				"voucher_pay_to.create",
				fmt.Sprintf("voucher_pay_to.create.%s", data.ID),
				fmt.Sprintf("voucher_pay_to.create.branch.%s", data.BranchID),
				fmt.Sprintf("voucher_pay_to.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *types.VoucherPayTo) registry.Topics {
			return []string{
				"voucher_pay_to.update",
				fmt.Sprintf("voucher_pay_to.update.%s", data.ID),
				fmt.Sprintf("voucher_pay_to.update.branch.%s", data.BranchID),
				fmt.Sprintf("voucher_pay_to.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *types.VoucherPayTo) registry.Topics {
			return []string{
				"voucher_pay_to.delete",
				fmt.Sprintf("voucher_pay_to.delete.%s", data.ID),
				fmt.Sprintf("voucher_pay_to.delete.branch.%s", data.BranchID),
				fmt.Sprintf("voucher_pay_to.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func VoucherPayToCurrentBranch(context context.Context, service *horizon.HorizonService, organizationID uuid.UUID, branchID uuid.UUID) ([]*types.VoucherPayTo, error) {
	return VoucherPayToManager(service).Find(context, &types.VoucherPayTo{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}
