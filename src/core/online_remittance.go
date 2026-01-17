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

func OnlineRemittanceManager(service *horizon.HorizonService) *registry.Registry[
	types.OnlineRemittance, types.OnlineRemittanceResponse, types.OnlineRemittanceRequest] {
	return registry.NewRegistry(registry.RegistryParams[
		types.OnlineRemittance, types.OnlineRemittanceResponse, types.OnlineRemittanceRequest,
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
		Resource: func(data *types.OnlineRemittance) *types.OnlineRemittanceResponse {
			if data == nil {
				return nil
			}
			var dateEntry *string
			if data.DateEntry != nil {
				s := data.DateEntry.Format(time.RFC3339)
				dateEntry = &s
			}
			return &types.OnlineRemittanceResponse{
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
				Amount:             data.Amount,
				AccountName:        data.AccountName,
				DateEntry:          dateEntry,
				Description:        data.Description,
			}
		},
		Created: func(data *types.OnlineRemittance) registry.Topics {
			return []string{
				"online_remittance.create",
				fmt.Sprintf("online_remittance.create.%s", data.ID),
				fmt.Sprintf("online_remittance.create.branch.%s", data.BranchID),
				fmt.Sprintf("online_remittance.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *types.OnlineRemittance) registry.Topics {
			return []string{
				"online_remittance.update",
				fmt.Sprintf("online_remittance.update.%s", data.ID),
				fmt.Sprintf("online_remittance.update.branch.%s", data.BranchID),
				fmt.Sprintf("online_remittance.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *types.OnlineRemittance) registry.Topics {
			return []string{
				"online_remittance.delete",
				fmt.Sprintf("online_remittance.delete.%s", data.ID),
				fmt.Sprintf("online_remittance.delete.branch.%s", data.BranchID),
				fmt.Sprintf("online_remittance.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func OnlineRemittanceCurrentBranch(context context.Context, service *horizon.HorizonService, organizationID uuid.UUID, branchID uuid.UUID) ([]*types.OnlineRemittance, error) {
	return OnlineRemittanceManager(service).Find(context, &types.OnlineRemittance{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}
