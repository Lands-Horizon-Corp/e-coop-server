package types

import (
	"time"

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
