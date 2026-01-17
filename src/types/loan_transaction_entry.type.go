package types

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	LoanTransactionStatic             LoanTransactionEntryType = "static"
	LoanTransactionDeduction          LoanTransactionEntryType = "deduction"
	LoanTransactionAddOn              LoanTransactionEntryType = "add-on"
	LoanTransactionAutomaticDeduction LoanTransactionEntryType = "automatic-deduction"
	LoanTransactionPrevious           LoanTransactionEntryType = "previous"
)

type (
	LoanTransactionEntryType string
	LoanTransactionEntry     struct {
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

		OrganizationID uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_loan_transaction_entry"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_loan_transaction_entry"`
		Branch         *Branch       `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"branch,omitempty"`

		LoanTransactionID uuid.UUID        `gorm:"type:uuid;not null;index:idx_loan_transaction_entry_loan_transaction"`
		LoanTransaction   *LoanTransaction `gorm:"foreignKey:LoanTransactionID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"loan_transaction,omitempty"`

		Index int `gorm:"type:int;default:0" json:"index"`

		Type    LoanTransactionEntryType `gorm:"type:varchar(20);not null;default:'static'" json:"type"`
		IsAddOn bool                     `gorm:"type:boolean;not null;default:false" json:"is_add_on"`

		AccountID *uuid.UUID `gorm:"type:uuid"`
		Account   *Account   `gorm:"foreignKey:AccountID;constraint:OnDelete:SET NULL;" json:"account,omitempty"`

		AutomaticLoanDeductionID        *uuid.UUID              `gorm:"type:uuid"`
		AutomaticLoanDeduction          *AutomaticLoanDeduction `gorm:"foreignKey:AutomaticLoanDeductionID;constraint:OnDelete:SET NULL;" json:"automatic_loan_deduction,omitempty"`
		IsAutomaticLoanDeductionDeleted bool                    `gorm:"type:boolean;not null;default:false" json:"is_automatic_loan_deduction_deleted"`

		Name        string  `gorm:"type:varchar(255)" json:"name"`
		Description string  `gorm:"type:varchar(500)" json:"description"`
		Credit      float64 `gorm:"type:decimal"`
		Debit       float64 `gorm:"type:decimal"`

		Amount float64 `gorm:"type:decimal;default:0" json:"amount,omitempty"`
	}

	LoanTransactionEntryResponse struct {
		ID                              uuid.UUID                       `json:"id"`
		CreatedAt                       string                          `json:"created_at"`
		CreatedByID                     uuid.UUID                       `json:"created_by_id"`
		CreatedBy                       *UserResponse                   `json:"created_by,omitempty"`
		UpdatedAt                       string                          `json:"updated_at"`
		UpdatedByID                     uuid.UUID                       `json:"updated_by_id"`
		UpdatedBy                       *UserResponse                   `json:"updated_by,omitempty"`
		OrganizationID                  uuid.UUID                       `json:"organization_id"`
		Organization                    *OrganizationResponse           `json:"organization,omitempty"`
		BranchID                        uuid.UUID                       `json:"branch_id"`
		Branch                          *BranchResponse                 `json:"branch,omitempty"`
		LoanTransactionID               uuid.UUID                       `json:"loan_transaction_id"`
		LoanTransaction                 *LoanTransactionResponse        `json:"loan_transaction,omitempty"`
		Index                           int                             `json:"index"`
		Type                            LoanTransactionEntryType        `json:"type"`
		IsAddOn                         bool                            `json:"is_add_on"`
		AccountID                       *uuid.UUID                      `json:"account_id,omitempty"`
		Account                         *AccountResponse                `json:"account,omitempty"`
		AutomaticLoanDeductionID        *uuid.UUID                      `json:"automatic_loan_deduction_id,omitempty"`
		AutomaticLoanDeduction          *AutomaticLoanDeductionResponse `json:"automatic_loan_deduction,omitempty"`
		IsAutomaticLoanDeductionDeleted bool                            `json:"is_automatic_loan_deduction_deleted"`
		Name                            string                          `json:"name"`
		Description                     string                          `json:"description"`
		Credit                          float64                         `json:"credit"`
		Debit                           float64                         `json:"debit"`
		Amount                          float64                         `json:"amount"`
	}

	LoanTransactionEntryRequest struct {
		ID                *uuid.UUID               `json:"id"`
		LoanTransactionID uuid.UUID                `json:"loan_transaction_id" validate:"required"`
		Index             int                      `json:"index,omitempty"`
		Type              LoanTransactionEntryType `json:"type" validate:"required,oneof=static deduction add-on"`
		IsAddOn           bool                     `json:"is_add_on,omitempty"`
		AccountID         *uuid.UUID               `json:"account_id,omitempty"`
		Name              string                   `json:"name,omitempty"`
		Description       string                   `json:"description,omitempty"`
		Credit            float64                  `json:"credit,omitempty"`
		Debit             float64                  `json:"debit,omitempty"`
	}

	LoanTransactionDeductionRequest struct {
		AccountID   uuid.UUID `json:"account_id" validate:"required"`
		Amount      float64   `json:"amount"`
		Description string    `json:"description,omitempty"`
		IsAddOn     bool      `json:"is_add_on,omitempty"`
	}
)
