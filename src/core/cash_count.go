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

func CashCountManager(service *horizon.HorizonService) *registry.Registry[types.CashCount, types.CashCountResponse, types.CashCountRequest] {
	return registry.NewRegistry(registry.RegistryParams[
		types.CashCount, types.CashCountResponse, types.CashCountRequest,
	]{
		Preloads: []string{
			"CreatedBy", "UpdatedBy",
			"EmployeeUser", "TransactionBatch", "Currency",
		},
		Database: service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return service.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *types.CashCount) *types.CashCountResponse {
			if data == nil {
				return nil
			}
			return &types.CashCountResponse{
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
				EmployeeUserID:     data.EmployeeUserID,
				EmployeeUser:       UserManager(service).ToModel(data.EmployeeUser),
				TransactionBatchID: data.TransactionBatchID,
				TransactionBatch:   TransactionBatchManager(service).ToModel(data.TransactionBatch),
				CurrencyID:         data.CurrencyID,
				Currency:           CurrencyManager(service).ToModel(data.Currency),
				BillAmount:         data.BillAmount,
				Quantity:           data.Quantity,
				Amount:             data.Amount,
				Name:               data.Name,
			}
		},
		Created: func(data *types.CashCount) registry.Topics {
			return []string{
				"cash_count.create",
				fmt.Sprintf("cash_count.create.%s", data.ID),
				fmt.Sprintf("cash_count.create.branch.%s", data.BranchID),
				fmt.Sprintf("cash_count.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *types.CashCount) registry.Topics {
			return []string{
				"cash_count.update",
				fmt.Sprintf("cash_count.update.%s", data.ID),
				fmt.Sprintf("cash_count.update.branch.%s", data.BranchID),
				fmt.Sprintf("cash_count.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *types.CashCount) registry.Topics {
			return []string{
				"cash_count.delete",
				fmt.Sprintf("cash_count.delete.%s", data.ID),
				fmt.Sprintf("cash_count.delete.branch.%s", data.BranchID),
				fmt.Sprintf("cash_count.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func CashCountCurrentBranch(context context.Context, service *horizon.HorizonService, organizationID uuid.UUID, branchID uuid.UUID) ([]*types.CashCount, error) {
	return CashCountManager(service).Find(context, &types.CashCount{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}
