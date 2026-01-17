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
	"gorm.io/gorm"
)

func AccountTransactionManager(service *horizon.HorizonService) *registry.Registry[
	types.AccountTransaction,
	types.AccountTransactionResponse,
	types.AccountTransactionRequest,
] {
	return registry.GetRegistry(registry.RegistryParams[
		types.AccountTransaction,
		types.AccountTransactionResponse,
		types.AccountTransactionRequest,
	]{
		Preloads: []string{
			"CreatedBy",
			"UpdatedBy",
			"Entries",
			"Entries.Account",
		},
		Database: service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return service.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *types.AccountTransaction) *types.AccountTransactionResponse {
			if data == nil {
				return nil
			}
			return &types.AccountTransactionResponse{
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
				JVNumber:       data.JVNumber,
				Date:           data.Date.Format("2006-01-02"),
				Description:    data.Description,
				Debit:          data.Debit,
				Credit:         data.Credit,
				Source:         data.Source,
				Entries:        AccountTransactionEntryManager(service).ToModels(data.Entries),
			}
		},
		Created: func(data *types.AccountTransaction) registry.Topics {
			return []string{
				"account_transaction.create",
				fmt.Sprintf("account_transaction.create.%s", data.ID),
				fmt.Sprintf("account_transaction.create.branch.%s", data.BranchID),
				fmt.Sprintf("account_transaction.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *types.AccountTransaction) registry.Topics {
			return []string{
				"account_transaction.update",
				fmt.Sprintf("account_transaction.update.%s", data.ID),
				fmt.Sprintf("account_transaction.update.branch.%s", data.BranchID),
				fmt.Sprintf("account_transaction.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *types.AccountTransaction) registry.Topics {
			return []string{
				"account_transaction.delete",
				fmt.Sprintf("account_transaction.delete.%s", data.ID),
				fmt.Sprintf("account_transaction.delete.branch.%s", data.BranchID),
				fmt.Sprintf("account_transaction.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func AccountTransactionByMonthYear(
	ctx context.Context,
	service *horizon.HorizonService,
	year int,
	month int,
	organizationID uuid.UUID,
	branchID uuid.UUID,
) ([]*types.AccountTransaction, error) {
	normalizedMonth := ((month - 1) % 12) + 1
	if normalizedMonth <= 0 {
		normalizedMonth += 12
	}
	start := time.Date(year, time.Month(normalizedMonth), 1, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 1, 0).Add(-time.Nanosecond)
	filters := []query.ArrFilterSQL{
		{Field: "organization_id", Op: query.ModeEqual, Value: organizationID},
		{Field: "branch_id", Op: query.ModeEqual, Value: branchID},
		{Field: "created_at", Op: query.ModeGTE, Value: start},
		{Field: "created_at", Op: query.ModeLTE, Value: end},
	}
	sorts := []query.ArrFilterSortSQL{
		{Field: "created_at", Order: query.SortOrderAsc},
	}
	return AccountTransactionManager(service).ArrFind(ctx, filters, sorts, "Entries", "Entries.Account")
}

func AccountTransactionDestroyer(
	ctx context.Context,
	service *horizon.HorizonService,
	tx *gorm.DB,
	day time.Time,
	organizationID, branchID uuid.UUID,
) error {
	startOfDay := time.Date(day.Year(), day.Month(), day.Day(), 0, 0, 0, 0, day.Location())
	endOfDay := time.Date(day.Year(), day.Month(), day.Day(), 23, 59, 59, 999999999, day.Location())
	filters := []query.ArrFilterSQL{
		{Field: "organization_id", Op: query.ModeEqual, Value: organizationID},
		{Field: "branch_id", Op: query.ModeEqual, Value: branchID},
		{Field: "date", Op: query.ModeGTE, Value: startOfDay},
		{Field: "date", Op: query.ModeLTE, Value: endOfDay},
	}
	sorts := []query.ArrFilterSortSQL{
		{Field: "created_at", Order: query.SortOrderAsc},
	}
	transactions, err := AccountTransactionManager(service).ArrFind(ctx, filters, sorts, "Entries")
	if err != nil {
		return err
	}
	for _, item := range transactions {
		if item == nil {
			continue
		}
		for _, entry := range item.Entries {
			if err := AccountTransactionEntryManager(service).DeleteWithTx(ctx, tx, entry.ID); err != nil {
				return err
			}
		}
		if err := AccountTransactionManager(service).DeleteWithTx(ctx, tx, item.ID); err != nil {
			return err
		}
	}
	return nil
}
