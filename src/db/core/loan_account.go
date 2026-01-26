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

func LoanAccountManager(service *horizon.HorizonService) *registry.Registry[
	types.LoanAccount, types.LoanAccountResponse, types.LoanAccountRequest] {
	return registry.NewRegistry(registry.RegistryParams[
		types.LoanAccount, types.LoanAccountResponse, types.LoanAccountRequest,
	]{
		Preloads: []string{"LoanTransaction", "Account", "Account.Currency", "AccountHistory"},
		Database: service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return service.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *types.LoanAccount) *types.LoanAccountResponse {
			if data == nil {
				return nil
			}
			return &types.LoanAccountResponse{
				ID:                data.ID,
				CreatedAt:         data.CreatedAt.Format(time.RFC3339),
				CreatedByID:       data.CreatedByID,
				CreatedBy:         UserManager(service).ToModel(data.CreatedBy),
				UpdatedAt:         data.UpdatedAt.Format(time.RFC3339),
				UpdatedByID:       data.UpdatedByID,
				UpdatedBy:         UserManager(service).ToModel(data.UpdatedBy),
				OrganizationID:    data.OrganizationID,
				Organization:      OrganizationManager(service).ToModel(data.Organization),
				BranchID:          data.BranchID,
				Branch:            BranchManager(service).ToModel(data.Branch),
				LoanTransactionID: data.LoanTransactionID,
				LoanTransaction:   LoanTransactionManager(service).ToModel(data.LoanTransaction),
				AccountID:         data.AccountID,
				Account:           AccountManager(service).ToModel(data.Account),
				AccountHistoryID:  data.AccountHistoryID,
				AccountHistory:    AccountHistoryManager(service).ToModel(data.AccountHistory),
				Amount:            data.Amount,

				TotalAdd:            data.TotalAdd,
				TotalAddCount:       data.TotalAddCount,
				TotalDeduction:      data.TotalDeduction,
				TotalDeductionCount: data.TotalDeductionCount,
				TotalPayment:        data.TotalPayment,
				TotalPaymentCount:   data.TotalPaymentCount,
			}
		},

		Created: func(data *types.LoanAccount) registry.Topics {
			return []string{
				"loan_account.create",
				fmt.Sprintf("loan_account.create.%s", data.ID),
				fmt.Sprintf("loan_account.create.branch.%s", data.BranchID),
				fmt.Sprintf("loan_account.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *types.LoanAccount) registry.Topics {
			return []string{
				"loan_account.update",
				fmt.Sprintf("loan_account.update.%s", data.ID),
				fmt.Sprintf("loan_account.update.branch.%s", data.BranchID),
				fmt.Sprintf("loan_account.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *types.LoanAccount) registry.Topics {
			return []string{
				"loan_account.delete",
				fmt.Sprintf("loan_account.delete.%s", data.ID),
				fmt.Sprintf("loan_account.delete.branch.%s", data.BranchID),
				fmt.Sprintf("loan_account.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func LoanAccountCurrentBranch(context context.Context, service *horizon.HorizonService,
	organizationID uuid.UUID, branchID uuid.UUID) ([]*types.LoanAccount, error) {
	filters := []query.ArrFilterSQL{
		{Field: "organization_id", Op: query.ModeEqual, Value: organizationID},
		{Field: "branch_id", Op: query.ModeEqual, Value: branchID},
	}

	return LoanAccountManager(service).ArrFind(context, filters, nil)
}

func GetLoanAccountByLoanTransaction(
	ctx context.Context, service *horizon.HorizonService, tx *gorm.DB,
	loanTransactionID, accountID, organizationID, branchID uuid.UUID) (*types.LoanAccount, error) {
	filters := []query.ArrFilterSQL{
		{Field: "organization_id", Op: query.ModeEqual, Value: organizationID},
		{Field: "branch_id", Op: query.ModeEqual, Value: branchID},
		{Field: "loan_transaction_id", Op: query.ModeEqual, Value: loanTransactionID},
		{Field: "account_id", Op: query.ModeEqual, Value: accountID},
	}

	return LoanAccountManager(service).ArrFindOneWithLock(
		ctx, tx, filters, nil, "Account", "Account.DefaultPaymentType", "AccountHistory",
	)
}
