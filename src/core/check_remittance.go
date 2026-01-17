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

func CheckRemittanceManager(service *horizon.HorizonService) *registry.Registry[
	types.CheckRemittance, types.CheckRemittanceResponse, types.CheckRemittanceRequest] {
	return registry.NewRegistry(registry.RegistryParams[
		types.CheckRemittance, types.CheckRemittanceResponse, types.CheckRemittanceRequest,
	]{
		Preloads: []string{
			"CreatedBy", "UpdatedBy",
			"Bank", "Media", "EmployeeUser", "TransactionBatch", "Currency",
			"Bank.Media",
		},
		Database: service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return service.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *types.CheckRemittance) *types.CheckRemittanceResponse {
			if data == nil {
				return nil
			}
			var dateEntry *string
			if data.DateEntry != nil {
				s := data.DateEntry.Format(time.RFC3339)
				dateEntry = &s
			}
			return &types.CheckRemittanceResponse{
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
				BankID:             data.BankID,
				Bank:               BankManager(service).ToModel(data.Bank),
				MediaID:            data.MediaID,
				Media:              MediaManager(service).ToModel(data.Media),
				EmployeeUserID:     data.EmployeeUserID,
				EmployeeUser:       UserManager(service).ToModel(data.EmployeeUser),
				TransactionBatchID: data.TransactionBatchID,
				TransactionBatch:   TransactionBatchManager(service).ToModel(data.TransactionBatch),
				CurrencyID:         data.CurrencyID,
				Currency:           CurrencyManager(service).ToModel(data.Currency),
				ReferenceNumber:    data.ReferenceNumber,
				AccountName:        data.AccountName,
				Amount:             data.Amount,
				DateEntry:          dateEntry,
				Description:        data.Description,
			}
		},
		Created: func(data *types.CheckRemittance) registry.Topics {
			return []string{
				"check_remittance.create",
				fmt.Sprintf("check_remittance.create.%s", data.ID),
				fmt.Sprintf("check_remittance.create.branch.%s", data.BranchID),
				fmt.Sprintf("check_remittance.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *types.CheckRemittance) registry.Topics {
			return []string{
				"check_remittance.update",
				fmt.Sprintf("check_remittance.update.%s", data.ID),
				fmt.Sprintf("check_remittance.update.branch.%s", data.BranchID),
				fmt.Sprintf("check_remittance.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *types.CheckRemittance) registry.Topics {
			return []string{
				"check_remittance.delete",
				fmt.Sprintf("check_remittance.delete.%s", data.ID),
				fmt.Sprintf("check_remittance.delete.branch.%s", data.BranchID),
				fmt.Sprintf("check_remittance.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func CheckRemittanceCurrentBranch(context context.Context, service *horizon.HorizonService,
	organizationID uuid.UUID, branchID uuid.UUID) ([]*types.CheckRemittance, error) {
	return CheckRemittanceManager(service).Find(context, &types.CheckRemittance{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}
