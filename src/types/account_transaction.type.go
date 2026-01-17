package types

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AccountTransactionSource string

const (
	AccountTransactionSourceDailyCollectionBook       AccountTransactionSource = "daily_collection_book"
	AccountTransactionSourceCashCheckDisbursementBook AccountTransactionSource = "cash_check_disbursement_book"
	AccountTransactionSourceGeneralJournal            AccountTransactionSource = "general_journal"
)

type (
	AccountTransaction struct {
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

		OrganizationID uuid.UUID     `gorm:"type:uuid;not null;index:idx_org_branch_account_tx" json:"organization_id"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID     `gorm:"type:uuid;not null;index:idx_org_branch_account_tx" json:"branch_id"`
		Branch         *Branch       `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"branch,omitempty"`

		Source AccountTransactionSource `gorm:"type:varchar(50);not null" json:"source"`

		JVNumber    string    `gorm:"type:varchar(255);not null" json:"jv_number"`
		Date        time.Time `gorm:"type:date;not null" json:"date"`
		Description string    `gorm:"type:text" json:"description"`
		Debit       float64   `gorm:"type:numeric(18,2);default:0" json:"debit"`
		Credit      float64   `gorm:"type:numeric(18,2);default:0" json:"credit"`

		Entries []*AccountTransactionEntry `gorm:"foreignKey:AccountTransactionID" json:"entries,omitempty"`
	}

	AccountTransactionResponse struct {
		ID             uuid.UUID             `json:"id"`
		CreatedAt      string                `json:"created_at"`
		CreatedByID    uuid.UUID             `json:"created_by_id"`
		CreatedBy      *UserResponse         `json:"created_by,omitempty"`
		UpdatedAt      string                `json:"updated_at"`
		UpdatedByID    uuid.UUID             `json:"updated_by_id"`
		UpdatedBy      *UserResponse         `json:"updated_by,omitempty"`
		OrganizationID uuid.UUID             `json:"organization_id"`
		Organization   *OrganizationResponse `json:"organization,omitempty"`
		BranchID       uuid.UUID             `json:"branch_id"`
		Branch         *BranchResponse       `json:"branch,omitempty"`

		JVNumber    string                   `json:"jv_number"`
		Date        string                   `json:"date"`
		Description string                   `json:"description"`
		Debit       float64                  `json:"debit"`
		Credit      float64                  `json:"credit"`
		Source      AccountTransactionSource `json:"source"`

		Entries []*AccountTransactionEntryResponse `json:"entries"`
	}

	AccountTransactionRequest struct {
		JVNumber    string `json:"jv_number" validate:"required,min=1,max=255"`
		Description string `json:"description,omitempty"`
	}

	AccountTransactionProcessGLRequest struct {
		StartDate time.Time `json:"start_date" validate:"required"`
		EndDate   time.Time `json:"end_date" validate:"required,gtefield=StartDate"`
	}

	AccountTransactionLedgerResponse struct {
		AccountTransactionEntry []*AccountTransactionEntryResponse `json:"account_transaction_entry"`
		Month                   int                                `json:"month"`
		Debit                   float64                            `json:"debit"`
		Credit                  float64                            `json:"credit"`
	}
)
