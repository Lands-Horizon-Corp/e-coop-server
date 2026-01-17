package types

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type (
	AdjustmentEntry struct {
		ID          uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
		CreatedAt   time.Time      `gorm:"not null;default:now()" json:"created_at"`
		CreatedByID uuid.UUID      `gorm:"type:uuid" json:"created_by_id"`
		CreatedBy   *User          `gorm:"foreignKey:CreatedByID;constraint:OnDelete:SET NULL;" json:"created_by,omitempty"`
		UpdatedAt   time.Time      `gorm:"not null;default:now()" json:"updated_at"`
		UpdatedByID uuid.UUID      `gorm:"type:uuid" json:"updated_by_id"`
		UpdatedBy   *User          `gorm:"foreignKey:UpdatedByID;constraint:OnDelete:SET NULL;" json:"updated_by,omitempty"`
		DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at"`
		DeletedByID *uuid.UUID     `gorm:"type:uuid" json:"deleted_by_id"`
		DeletedBy   *User          `gorm:"foreignKey:DeletedByID;constraint:OnDelete:SET NULL;" json:"deleted_by,omitempty"`

		OrganizationID uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_adjustment_entry" json:"organization_id"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_adjustment_entry" json:"branch_id"`
		Branch         *Branch       `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"branch,omitempty"`

		TransactionBatchID *uuid.UUID        `gorm:"type:uuid" json:"transaction_batch_id"`
		TransactionBatch   *TransactionBatch `gorm:"foreignKey:TransactionBatchID;constraint:OnDelete:SET NULL,OnUpdate:CASCADE;" json:"transaction_batch,omitempty"`

		SignatureMediaID *uuid.UUID `gorm:"type:uuid" json:"signature_media_id"`
		SignatureMedia   *Media     `gorm:"foreignKey:SignatureMediaID;constraint:OnDelete:SET NULL;" json:"signature_media,omitempty"`

		AccountID uuid.UUID `gorm:"type:uuid;not null" json:"account_id"`
		Account   *Account  `gorm:"foreignKey:AccountID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"account,omitempty"`

		MemberProfileID *uuid.UUID     `gorm:"type:uuid" json:"member_profile_id"`
		MemberProfile   *MemberProfile `gorm:"foreignKey:MemberProfileID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"member_profile,omitempty"`

		EmployeeUserID *uuid.UUID `gorm:"type:uuid" json:"employee_user_id"`
		EmployeeUser   *User      `gorm:"foreignKey:EmployeeUserID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"employee_user,omitempty"`

		PaymentTypeID *uuid.UUID   `gorm:"type:uuid" json:"payment_type_id"`
		PaymentType   *PaymentType `gorm:"foreignKey:PaymentTypeID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"payment_type,omitempty"`

		TypeOfPaymentType string `gorm:"type:varchar(50)" json:"type_of_payment_type"`

		Description     string     `gorm:"type:text" json:"description"`
		ReferenceNumber string     `gorm:"type:varchar(255)" json:"reference_number"`
		EntryDate       *time.Time `gorm:"type:date" json:"entry_date"`

		Debit  float64 `gorm:"type:decimal" json:"debit"`
		Credit float64 `gorm:"type:decimal" json:"credit"`

		AdjustmentTags []*AdjustmentTag `gorm:"foreignKey:AdjustmentEntryID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"adjustment_tags,omitempty"`

		LoanTransactionID *uuid.UUID       `gorm:"type:uuid" json:"loan_transaction_id"`
		LoanTransaction   *LoanTransaction `gorm:"foreignKey:LoanTransactionID;constraint:OnDelete:SET NULL,OnUpdate:CASCADE;" json:"loan_transaction,omitempty"`
	}

	AdjustmentEntryResponse struct {
		ID                 uuid.UUID                 `json:"id"`
		CreatedAt          string                    `json:"created_at"`
		CreatedByID        uuid.UUID                 `json:"created_by_id"`
		CreatedBy          *UserResponse             `json:"created_by,omitempty"`
		UpdatedAt          string                    `json:"updated_at"`
		UpdatedByID        uuid.UUID                 `json:"updated_by_id"`
		UpdatedBy          *UserResponse             `json:"updated_by,omitempty"`
		OrganizationID     uuid.UUID                 `json:"organization_id"`
		Organization       *OrganizationResponse     `json:"organization,omitempty"`
		BranchID           uuid.UUID                 `json:"branch_id"`
		Branch             *BranchResponse           `json:"branch,omitempty"`
		TransactionBatchID *uuid.UUID                `json:"transaction_batch_id,omitempty"`
		TransactionBatch   *TransactionBatchResponse `json:"transaction_batch,omitempty"`
		SignatureMediaID   *uuid.UUID                `json:"signature_media_id,omitempty"`
		SignatureMedia     *MediaResponse            `json:"signature_media,omitempty"`
		AccountID          uuid.UUID                 `json:"account_id"`
		Account            *AccountResponse          `json:"account,omitempty"`
		MemberProfileID    *uuid.UUID                `json:"member_profile_id,omitempty"`
		MemberProfile      *MemberProfileResponse    `json:"member_profile,omitempty"`
		EmployeeUserID     *uuid.UUID                `json:"employee_user_id,omitempty"`
		EmployeeUser       *UserResponse             `json:"employee_user,omitempty"`
		PaymentTypeID      *uuid.UUID                `json:"payment_type_id,omitempty"`
		PaymentType        *PaymentTypeResponse      `json:"payment_type,omitempty"`
		TypeOfPaymentType  string                    `json:"type_of_payment_type"`
		Description        string                    `json:"description"`
		ReferenceNumber    string                    `json:"reference_number"`
		EntryDate          *string                   `json:"entry_date,omitempty"`
		Debit              float64                   `json:"debit"`
		Credit             float64                   `json:"credit"`
		AdjustmentTags     []*AdjustmentTagResponse  `json:"adjustment_tags,omitempty"`
		LoanTransactionID  *uuid.UUID                `json:"loan_transaction_id,omitempty"`
		LoanTransaction    *LoanTransactionResponse  `json:"loan_transaction,omitempty"`
	}

	AdjustmentEntryRequest struct {
		TransactionBatchID *uuid.UUID `json:"transaction_batch_id,omitempty"`
		SignatureMediaID   *uuid.UUID `json:"signature_media_id,omitempty"`
		AccountID          uuid.UUID  `json:"account_id" validate:"required"`
		MemberProfileID    *uuid.UUID `json:"member_profile_id,omitempty"`
		PaymentTypeID      *uuid.UUID `json:"payment_type_id,omitempty"`
		TypeOfPaymentType  string     `json:"type_of_payment_type,omitempty"`
		Description        string     `json:"description,omitempty"`
		ReferenceNumber    string     `json:"reference_number,omitempty"`
		EntryDate          *time.Time `json:"entry_date,omitempty"`
		Debit              float64    `json:"debit,omitempty"`
		Credit             float64    `json:"credit,omitempty"`
		LoanTransactionID  *uuid.UUID `json:"loan_transaction_id,omitempty"`
	}

	AdjustmentEntryTotalResponse struct {
		TotalDebit  float64 `json:"total_debit"`
		TotalCredit float64 `json:"total_credit"`
		Balance     float64 `json:"balance"`
		IsBalanced  bool    `json:"is_balanced"`
	}

	AdjustmentEntrySummaryRequest struct {
		CurrencyID         uuid.UUID  `json:"currency_id" validate:"required"`
		UserOrganizationID *uuid.UUID `json:"user_organization_id,omitempty"`
	}
)
