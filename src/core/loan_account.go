package core

import (
	"context"
	"fmt"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/pkg/query"
	"github.com/Lands-Horizon-Corp/e-coop-server/pkg/registry"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type (
	LoanAccount struct {
		ID          uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
		CreatedAt   time.Time      `gorm:"not null;default:now()"`
		CreatedByID uuid.UUID      `gorm:"type:uuid"`
		CreatedBy   *User          `gorm:"foreignKey:CreatedByID;constraint:OnDelete:SET NULL;" json:"created_by,omitempty"`
		UpdatedAt   time.Time      `gorm:"not null;default:now()"`
		UpdatedByID uuid.UUID      `gorm:"type:uuid"`
		UpdatedBy   *User          `gorm:"foreignKey:UpdatedByID;constraint:OnDelete:SET NULL;" json:"updated_by,omitempty"`
		DeletedAt   gorm.DeletedAt `gorm:"index"`
		DeletedByID *uuid.UUID     `gorm:"type:uuid"`
		DeletedBy   *User          `gorm:"foreignKey:DeletedByID;constraint:OnDelete:SET NULL;" json:"deleted_by,omitempty"`

		OrganizationID uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_loan_account"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_loan_account"`
		Branch         *Branch       `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"branch,omitempty"`

		LoanTransactionID uuid.UUID        `gorm:"type:uuid;not null;index:idx_loan_account_loan_transaction"`
		LoanTransaction   *LoanTransaction `gorm:"foreignKey:LoanTransactionID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"loan_transaction,omitempty"`

		AccountID        *uuid.UUID      `gorm:"type:uuid"`
		Account          *Account        `gorm:"foreignKey:AccountID;constraint:OnDelete:SET NULL;" json:"account,omitempty"`
		AccountHistoryID *uuid.UUID      `gorm:"type:uuid"`
		AccountHistory   *AccountHistory `gorm:"foreignKey:AccountHistoryID;constraint:OnDelete:SET NULL;" json:"account_history,omitempty"`

		Amount float64 `gorm:"type:decimal;default:0" json:"amount,omitempty"`

		TotalAdd            float64 `gorm:"type:decimal;default:0" json:"total_add,omitempty"`
		TotalAddCount       int     `gorm:"type:int;default:0" json:"total_add_count,omitempty"`
		TotalDeduction      float64 `gorm:"type:decimal;default:0" json:"total_deduction,omitempty"`
		TotalDeductionCount int     `gorm:"type:int;default:0" json:"total_deduction_count,omitempty"`
		TotalPayment        float64 `gorm:"type:decimal;default:0" json:"total_payment,omitempty"`
		TotalPaymentCount   int     `gorm:"type:int;default:0" json:"total_payment_count,omitempty"`
	}

	LoanAccountResponse struct {
		ID                uuid.UUID                `json:"id"`
		CreatedAt         string                   `json:"created_at"`
		CreatedByID       uuid.UUID                `json:"created_by_id"`
		CreatedBy         *UserResponse            `json:"created_by,omitempty"`
		UpdatedAt         string                   `json:"updated_at"`
		UpdatedByID       uuid.UUID                `json:"updated_by_id"`
		UpdatedBy         *UserResponse            `json:"updated_by,omitempty"`
		OrganizationID    uuid.UUID                `json:"organization_id"`
		Organization      *OrganizationResponse    `json:"organization,omitempty"`
		BranchID          uuid.UUID                `json:"branch_id"`
		Branch            *BranchResponse          `json:"branch,omitempty"`
		LoanTransactionID uuid.UUID                `json:"loan_transaction_id"`
		LoanTransaction   *LoanTransactionResponse `json:"loan_transaction,omitempty"`
		AccountID         *uuid.UUID               `json:"account_id,omitempty"`
		Account           *AccountResponse         `json:"account,omitempty"`
		AccountHistoryID  *uuid.UUID               `json:"account_history_id,omitempty"`
		AccountHistory    *AccountHistoryResponse  `json:"account_history,omitempty"`
		Amount            float64                  `json:"amount"`

		TotalAdd            float64 `json:"total_add"`
		TotalAddCount       int     `json:"total_add_count"`
		TotalDeduction      float64 `json:"total_deduction"`
		TotalDeductionCount int     `json:"total_deduction_count"`
		TotalPayment        float64 `json:"total_payment"`
		TotalPaymentCount   int     `json:"total_payment_count"`
	}

	LoanAccountRequest struct {
		ID                *uuid.UUID `json:"id"`
		LoanTransactionID uuid.UUID  `json:"loan_transaction_id" validate:"required"`
		AccountID         *uuid.UUID `json:"account_id,omitempty"`
		AccountHistoryID  *uuid.UUID `json:"account_history_id,omitempty"`
		Amount            float64    `json:"amount,omitempty"`

		TotalAdd            float64 `json:"total_add,omitempty"`
		TotalAddCount       int     `json:"total_add_count,omitempty"`
		TotalDeduction      float64 `json:"total_deduction,omitempty"`
		TotalDeductionCount int     `json:"total_deduction_count,omitempty"`
		TotalPayment        float64 `json:"total_payment,omitempty"`
		TotalPaymentCount   int     `json:"total_payment_count,omitempty"`
	}
)

func (m *Core) LoanAccountManager() *registry.Registry[LoanAccount, LoanAccountResponse, LoanAccountRequest] {
	return registry.NewRegistry(registry.RegistryParams[
		LoanAccount, LoanAccountResponse, LoanAccountRequest,
	]{
		Preloads: []string{
			"CreatedBy", "UpdatedBy", "LoanTransaction", "Account", "AccountHistory",
		},
		Database: m.provider.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return m.provider.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *LoanAccount) *LoanAccountResponse {
			if data == nil {
				return nil
			}
			return &LoanAccountResponse{
				ID:                data.ID,
				CreatedAt:         data.CreatedAt.Format(time.RFC3339),
				CreatedByID:       data.CreatedByID,
				CreatedBy:         m.UserManager().ToModel(data.CreatedBy),
				UpdatedAt:         data.UpdatedAt.Format(time.RFC3339),
				UpdatedByID:       data.UpdatedByID,
				UpdatedBy:         m.UserManager().ToModel(data.UpdatedBy),
				OrganizationID:    data.OrganizationID,
				Organization:      m.OrganizationManager().ToModel(data.Organization),
				BranchID:          data.BranchID,
				Branch:            m.BranchManager().ToModel(data.Branch),
				LoanTransactionID: data.LoanTransactionID,
				LoanTransaction:   m.LoanTransactionManager().ToModel(data.LoanTransaction),
				AccountID:         data.AccountID,
				Account:           m.AccountManager().ToModel(data.Account),
				AccountHistoryID:  data.AccountHistoryID,
				AccountHistory:    m.AccountHistoryManager().ToModel(data.AccountHistory),
				Amount:            data.Amount,

				TotalAdd:            data.TotalAdd,
				TotalAddCount:       data.TotalAddCount,
				TotalDeduction:      data.TotalDeduction,
				TotalDeductionCount: data.TotalDeductionCount,
				TotalPayment:        data.TotalPayment,
				TotalPaymentCount:   data.TotalPaymentCount,
			}
		},

		Created: func(data *LoanAccount) registry.Topics {
			return []string{
				"loan_account.create",
				fmt.Sprintf("loan_account.create.%s", data.ID),
				fmt.Sprintf("loan_account.create.branch.%s", data.BranchID),
				fmt.Sprintf("loan_account.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *LoanAccount) registry.Topics {
			return []string{
				"loan_account.update",
				fmt.Sprintf("loan_account.update.%s", data.ID),
				fmt.Sprintf("loan_account.update.branch.%s", data.BranchID),
				fmt.Sprintf("loan_account.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *LoanAccount) registry.Topics {
			return []string{
				"loan_account.delete",
				fmt.Sprintf("loan_account.delete.%s", data.ID),
				fmt.Sprintf("loan_account.delete.branch.%s", data.BranchID),
				fmt.Sprintf("loan_account.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func (m *Core) LoanAccountCurrentBranch(context context.Context, organizationID uuid.UUID, branchID uuid.UUID) ([]*LoanAccount, error) {
	filters := []registry.FilterSQL{
		{Field: "organization_id", Op: query.ModeEqual, Value: organizationID},
		{Field: "branch_id", Op: query.ModeEqual, Value: branchID},
	}

	return m.LoanAccountManager().ArrFind(context, filters, nil)
}

func (m *Core) GetLoanAccountByLoanTransaction(
	ctx context.Context, tx *gorm.DB, loanTransactionID, accountID, organizationID, branchID uuid.UUID) (*LoanAccount, error) {
	filters := []registry.FilterSQL{
		{Field: "organization_id", Op: query.ModeEqual, Value: organizationID},
		{Field: "branch_id", Op: query.ModeEqual, Value: branchID},
		{Field: "loan_transaction_id", Op: query.ModeEqual, Value: loanTransactionID},
		{Field: "account_id", Op: query.ModeEqual, Value: accountID},
	}

	return m.LoanAccountManager().ArrFindOneWithLock(
		ctx, tx, filters, nil, "Account", "Account.DefaultPaymentType", "AccountHistory",
	)
}
