package types

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type (
	JournalVoucher struct {
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

		OrganizationID uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_journal_voucher"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_journal_voucher"`
		Branch         *Branch       `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"branch,omitempty"`
		CurrencyID     uuid.UUID     `gorm:"type:uuid;not null"`
		Currency       *Currency     `gorm:"foreignKey:CurrencyID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"currency,omitempty"`

		Name              string     `gorm:"type:varchar(255)"`
		CashVoucherNumber string     `gorm:"type:varchar(255)"`
		Date              time.Time  `gorm:"not null;default:now()"`
		Description       string     `gorm:"type:text"`
		Reference         string     `gorm:"type:varchar(255)"`
		Status            string     `gorm:"type:varchar(50);default:'draft'"`
		PostedAt          *time.Time `gorm:"type:timestamp"`
		PostedByID        *uuid.UUID `gorm:"type:uuid"`
		PostedBy          *User      `gorm:"foreignKey:PostedByID;constraint:OnDelete:SET NULL;" json:"posted_by,omitempty"`

		EmployeeUserID     *uuid.UUID        `gorm:"type:uuid" json:"employee_user_id,omitempty"`
		EmployeeUser       *User             `gorm:"foreignKey:EmployeeUserID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"employee_user,omitempty"`
		TransactionBatchID *uuid.UUID        `gorm:"type:uuid" json:"transaction_batch_id,omitempty"`
		TransactionBatch   *TransactionBatch `gorm:"foreignKey:TransactionBatchID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"transaction_batch,omitempty"`

		PrintedDate  *time.Time `gorm:"type:timestamp"`
		PrintedByID  *uuid.UUID `gorm:"type:uuid"`
		PrintedBy    *User      `gorm:"foreignKey:PrintedByID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"printed_by,omitempty"`
		PrintNumber  int        `gorm:"type:int;default:0"`
		ApprovedDate *time.Time `gorm:"type:timestamp"`
		ApprovedByID *uuid.UUID `gorm:"type:uuid"`
		ApprovedBy   *User      `gorm:"foreignKey:ApprovedByID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"approved_by,omitempty"`
		ReleasedDate *time.Time `gorm:"type:timestamp"`
		ReleasedByID *uuid.UUID `gorm:"type:uuid"`
		ReleasedBy   *User      `gorm:"foreignKey:ReleasedByID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"released_by,omitempty"`

		JournalVoucherTags []*JournalVoucherTag `gorm:"foreignKey:JournalVoucherID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"journal_voucher_tags,omitempty"`

		JournalVoucherEntries []*JournalVoucherEntry `gorm:"foreignKey:JournalVoucherID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"journal_voucher_entries,omitempty"`

		TotalDebit  float64 `gorm:"type:decimal" json:"total_debit"`
		TotalCredit float64 `gorm:"type:decimal" json:"total_credit"`
	}

	JournalVoucherResponse struct {
		ID                uuid.UUID             `json:"id"`
		CreatedAt         string                `json:"created_at"`
		CreatedByID       uuid.UUID             `json:"created_by_id"`
		CreatedBy         *UserResponse         `json:"created_by,omitempty"`
		UpdatedAt         string                `json:"updated_at"`
		UpdatedByID       uuid.UUID             `json:"updated_by_id"`
		UpdatedBy         *UserResponse         `json:"updated_by,omitempty"`
		OrganizationID    uuid.UUID             `json:"organization_id"`
		Organization      *OrganizationResponse `json:"organization,omitempty"`
		BranchID          uuid.UUID             `json:"branch_id"`
		Branch            *BranchResponse       `json:"branch,omitempty"`
		CurrencyID        uuid.UUID             `json:"currency_id"`
		Currency          *CurrencyResponse     `json:"currency,omitempty"`
		Name              string                `json:"name"`
		VoucherNumber     string                `json:"voucher_number"`
		CashVoucherNumber string                `json:"cash_voucher_number"`
		Date              string                `json:"date"`
		Description       string                `json:"description"`
		Reference         string                `json:"reference"`
		Status            string                `json:"status"`
		PostedAt          *string               `json:"posted_at,omitempty"`
		PostedByID        *uuid.UUID            `json:"posted_by_id,omitempty"`
		PostedBy          *UserResponse         `json:"posted_by,omitempty"`

		EmployeeUserID     *uuid.UUID                `json:"employee_user_id,omitempty"`
		EmployeeUser       *UserResponse             `json:"employee_user,omitempty"`
		TransactionBatchID *uuid.UUID                `json:"transaction_batch_id,omitempty"`
		TransactionBatch   *TransactionBatchResponse `json:"transaction_batch,omitempty"`

		PrintedDate  *string       `json:"printed_date,omitempty"`
		PrintedByID  *uuid.UUID    `json:"printed_by_id,omitempty"`
		PrintedBy    *UserResponse `json:"printed_by,omitempty"`
		PrintNumber  int           `json:"print_number"`
		ApprovedDate *string       `json:"approved_date,omitempty"`
		ApprovedByID *uuid.UUID    `json:"approved_by_id,omitempty"`
		ApprovedBy   *UserResponse `json:"approved_by,omitempty"`
		ReleasedDate *string       `json:"released_date,omitempty"`
		ReleasedByID *uuid.UUID    `json:"released_by_id,omitempty"`
		ReleasedBy   *UserResponse `json:"released_by,omitempty"`

		JournalVoucherTags []*JournalVoucherTagResponse `json:"journal_voucher_tags,omitempty"`

		JournalVoucherEntries []*JournalVoucherEntryResponse `json:"journal_voucher_entries,omitempty"`

		TotalDebit  float64 `json:"total_debit"`
		TotalCredit float64 `json:"total_credit"`
	}

	JournalVoucherRequest struct {
		Name              string    `json:"name" validate:"required"`
		CashVoucherNumber string    `json:"cash_voucher_number,omitempty"`
		Date              time.Time `json:"date"`
		Description       string    `json:"description,omitempty"`
		Reference         string    `json:"reference,omitempty"`
		Status            string    `json:"status,omitempty"`
		CurrencyID        uuid.UUID `json:"currency_id" validate:"required"`

		JournalVoucherEntries        []*JournalVoucherEntryRequest `json:"journal_voucher_entries,omitempty"`
		JournalVoucherEntriesDeleted uuid.UUIDs                    `json:"journal_voucher_entries_deleted,omitempty"`
	}

	JournalVoucherPrintRequest struct {
		CashVoucherNumber string `json:"cash_voucher_number,omitempty"`
		ORAutoGenerated   bool   `json:"or_auto_generated"`
	}
)
