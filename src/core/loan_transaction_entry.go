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

func LoanTransactionEntryManager(service *horizon.HorizonService) *registry.Registry[
	types.LoanTransactionEntry, types.LoanTransactionEntryResponse, types.LoanTransactionEntryRequest] {
	return registry.NewRegistry(registry.RegistryParams[
		types.LoanTransactionEntry, types.LoanTransactionEntryResponse, types.LoanTransactionEntryRequest,
	]{
		Preloads: []string{
			"CreatedBy", "UpdatedBy", "LoanTransaction", "Account", "AutomaticLoanDeduction",
		},
		Database: service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return service.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *types.LoanTransactionEntry) *types.LoanTransactionEntryResponse {
			if data == nil {
				return nil
			}
			return &types.LoanTransactionEntryResponse{
				ID:                              data.ID,
				CreatedAt:                       data.CreatedAt.Format(time.RFC3339),
				CreatedByID:                     data.CreatedByID,
				CreatedBy:                       UserManager(service).ToModel(data.CreatedBy),
				UpdatedAt:                       data.UpdatedAt.Format(time.RFC3339),
				UpdatedByID:                     data.UpdatedByID,
				UpdatedBy:                       UserManager(service).ToModel(data.UpdatedBy),
				OrganizationID:                  data.OrganizationID,
				Organization:                    OrganizationManager(service).ToModel(data.Organization),
				BranchID:                        data.BranchID,
				Branch:                          BranchManager(service).ToModel(data.Branch),
				LoanTransactionID:               data.LoanTransactionID,
				LoanTransaction:                 LoanTransactionManager(service).ToModel(data.LoanTransaction),
				Index:                           data.Index,
				Type:                            data.Type,
				IsAddOn:                         data.IsAddOn,
				AccountID:                       data.AccountID,
				Account:                         AccountManager(service).ToModel(data.Account),
				AutomaticLoanDeductionID:        data.AutomaticLoanDeductionID,
				AutomaticLoanDeduction:          AutomaticLoanDeductionManager(service).ToModel(data.AutomaticLoanDeduction),
				IsAutomaticLoanDeductionDeleted: data.IsAutomaticLoanDeductionDeleted,

				Name:        data.Name,
				Description: data.Description,
				Credit:      data.Credit,
				Debit:       data.Debit,
				Amount:      data.Amount,
			}
		},

		Created: func(data *types.LoanTransactionEntry) registry.Topics {
			return []string{
				"loan_transaction_entry.create",
				fmt.Sprintf("loan_transaction_entry.create.%s", data.ID),
				fmt.Sprintf("loan_transaction_entry.create.branch.%s", data.BranchID),
				fmt.Sprintf("loan_transaction_entry.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *types.LoanTransactionEntry) registry.Topics {
			return []string{
				"loan_transaction_entry.update",
				fmt.Sprintf("loan_transaction_entry.update.%s", data.ID),
				fmt.Sprintf("loan_transaction_entry.update.branch.%s", data.BranchID),
				fmt.Sprintf("loan_transaction_entry.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *types.LoanTransactionEntry) registry.Topics {
			return []string{
				"loan_transaction_entry.delete",
				fmt.Sprintf("loan_transaction_entry.delete.%s", data.ID),
				fmt.Sprintf("loan_transaction_entry.delete.branch.%s", data.BranchID),
				fmt.Sprintf("loan_transaction_entry.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func LoanTransactionEntryCurrentBranch(context context.Context, service *horizon.HorizonService,
	organizationID uuid.UUID, branchID uuid.UUID) ([]*types.LoanTransactionEntry, error) {
	filters := []query.ArrFilterSQL{
		{Field: "organization_id", Op: query.ModeEqual, Value: organizationID},
		{Field: "branch_id", Op: query.ModeEqual, Value: branchID},
	}

	return LoanTransactionEntryManager(service).ArrFind(context, filters, nil)
}

func GetCashOnCashEquivalence(ctx context.Context, service *horizon.HorizonService,
	loanTransactionID, organizationID, branchID uuid.UUID) (*types.LoanTransactionEntry, error) {
	filters := []query.ArrFilterSQL{
		{Field: "organization_id", Op: query.ModeEqual, Value: organizationID},
		{Field: "branch_id", Op: query.ModeEqual, Value: branchID},
		{Field: "index", Op: query.ModeEqual, Value: 0},
		{Field: "debit", Op: query.ModeEqual, Value: 0},
		{Field: "loan_transaction_id", Op: query.ModeEqual, Value: loanTransactionID},
	}

	return LoanTransactionEntryManager(service).ArrFindOne(
		ctx, filters, nil, "Account", "Account.DefaultPaymentType",
	)
}

func GetLoanEntryAccount(ctx context.Context, service *horizon.HorizonService, loanTransactionID,
	organizationID, branchID uuid.UUID) (*types.LoanTransactionEntry, error) {
	filters := []query.ArrFilterSQL{
		{Field: "organization_id", Op: query.ModeEqual, Value: organizationID},
		{Field: "branch_id", Op: query.ModeEqual, Value: branchID},
		{Field: "index", Op: query.ModeEqual, Value: 1},
		{Field: "credit", Op: query.ModeEqual, Value: 0},
		{Field: "loan_transaction_id", Op: query.ModeEqual, Value: loanTransactionID},
	}

	return LoanTransactionEntryManager(service).ArrFindOne(ctx, filters, nil)
}
