package core

import (
	"context"
	"fmt"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/horizon"
	"github.com/Lands-Horizon-Corp/e-coop-server/pkg/query"
	"github.com/Lands-Horizon-Corp/e-coop-server/pkg/registry"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/types"
	"github.com/google/uuid"
)

func AccountTransactionEntryManager(service *horizon.HorizonService) *registry.Registry[
	types.AccountTransactionEntry,
	types.AccountTransactionEntryResponse,
	any,
] {
	return registry.GetRegistry(registry.RegistryParams[
		types.AccountTransactionEntry,
		types.AccountTransactionEntryResponse,
		any,
	]{
		Preloads: []string{
			"CreatedBy",
			"UpdatedBy",
			"Organization",
			"Branch",
			"Account",
			"AccountTransaction",
		},
		Database: service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return service.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *types.AccountTransactionEntry) *types.AccountTransactionEntryResponse {
			if data == nil {
				return nil
			}

			return &types.AccountTransactionEntryResponse{
				ID:             data.ID,
				CreatedAt:      data.CreatedAt.Format(time.RFC3339),
				OrganizationID: data.OrganizationID,
				BranchID:       data.BranchID,
				AccountID:      data.AccountID,
				Account:        AccountManager(service).ToModel(data.Account),
				Debit:          data.Debit,
				Credit:         data.Credit,
				Date:           data.Date.Format("2006-01-02"),
				JVNumber:       data.JVNumber,
			}
		},
		Created: func(data *types.AccountTransactionEntry) registry.Topics {
			return []string{
				"account_transaction_entry.create",
				fmt.Sprintf("account_transaction_entry.create.%s", data.ID),
				fmt.Sprintf("account_transaction_entry.create.transaction.%s", data.AccountTransactionID),
				fmt.Sprintf("account_transaction_entry.create.branch.%s", data.BranchID),
				fmt.Sprintf("account_transaction_entry.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *types.AccountTransactionEntry) registry.Topics {
			return []string{
				"account_transaction_entry.update",
				fmt.Sprintf("account_transaction_entry.update.%s", data.ID),
				fmt.Sprintf("account_transaction_entry.update.transaction.%s", data.AccountTransactionID),
				fmt.Sprintf("account_transaction_entry.update.branch.%s", data.BranchID),
				fmt.Sprintf("account_transaction_entry.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *types.AccountTransactionEntry) registry.Topics {
			return []string{
				"account_transaction_entry.delete",
				fmt.Sprintf("account_transaction_entry.delete.%s", data.ID),
				fmt.Sprintf("account_transaction_entry.delete.transaction.%s", data.AccountTransactionID),
				fmt.Sprintf("account_transaction_entry.delete.branch.%s", data.BranchID),
				fmt.Sprintf("account_transaction_entry.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func AccountingEntryByAccountMonthYear(
	ctx context.Context,
	service *horizon.HorizonService,
	accountID uuid.UUID,
	month int,
	year int,
	organizationID uuid.UUID,
	branchID uuid.UUID,
) ([]*types.AccountTransactionEntry, error) {
	normalizedMonth := ((month-1)%12+12)%12 + 1
	startDate := time.Date(year, time.Month(normalizedMonth), 1, 0, 0, 0, 0, time.UTC)
	endDate := startDate.AddDate(0, 1, 0).Add(-time.Nanosecond)
	filters := []query.ArrFilterSQL{
		{Field: "organization_id", Op: query.ModeEqual, Value: organizationID},
		{Field: "branch_id", Op: query.ModeEqual, Value: branchID},
		{Field: "account_id", Op: query.ModeEqual, Value: accountID},
		{Field: "date", Op: query.ModeGTE, Value: startDate},
		{Field: "date", Op: query.ModeLTE, Value: endDate},
	}
	sorts := []query.ArrFilterSortSQL{
		{Field: "date", Order: query.SortOrderAsc},
	}
	return AccountTransactionEntryManager(service).ArrFind(ctx, filters, sorts, "Account", "Account.Currency")
}
