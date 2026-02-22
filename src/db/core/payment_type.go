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

func PaymentTypeManager(service *horizon.HorizonService) *registry.Registry[types.PaymentType, types.PaymentTypeResponse, types.PaymentTypeRequest] {
	return registry.NewRegistry(registry.RegistryParams[
		types.PaymentType, types.PaymentTypeResponse, types.PaymentTypeRequest,
	]{
		Preloads: []string{"Account"},
		Database: service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return service.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *types.PaymentType) *types.PaymentTypeResponse {
			if data == nil {
				return nil
			}
			return &types.PaymentTypeResponse{
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
				NumberOfDays:   data.NumberOfDays,
				Type:           data.Type,
				AccountID:      data.AccountID,
				Account:        AccountManager(service).ToModel(data.Account),
			}
		},
		Created: func(data *types.PaymentType) registry.Topics {
			return []string{
				"payment_type.create",
				fmt.Sprintf("payment_type.create.%s", data.ID),
				fmt.Sprintf("payment_type.create.branch.%s", data.BranchID),
				fmt.Sprintf("payment_type.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *types.PaymentType) registry.Topics {
			return []string{
				"payment_type.update",
				fmt.Sprintf("payment_type.update.%s", data.ID),
				fmt.Sprintf("payment_type.update.branch.%s", data.BranchID),
				fmt.Sprintf("payment_type.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *types.PaymentType) registry.Topics {
			return []string{
				"payment_type.delete",
				fmt.Sprintf("payment_type.delete.%s", data.ID),
				fmt.Sprintf("payment_type.delete.branch.%s", data.BranchID),
				fmt.Sprintf("payment_type.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func PaymentTypeCurrentBranch(context context.Context, service *horizon.HorizonService,
	organizationID uuid.UUID, branchID uuid.UUID) ([]*types.PaymentType, error) {
	return PaymentTypeManager(service).Find(context, &types.PaymentType{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}
