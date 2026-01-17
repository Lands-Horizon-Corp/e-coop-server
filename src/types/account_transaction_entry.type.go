package types

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type (
	AccountTransactionEntry struct {
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

		OrganizationID uuid.UUID     `gorm:"type:uuid;not null;index:idx_org_branch_account_tx_entry" json:"organization_id"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"organization,omitempty"`

		BranchID uuid.UUID `gorm:"type:uuid;not null;index:idx_org_branch_account_tx_entry" json:"branch_id"`
		Branch   *Branch   `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"branch,omitempty"`

		AccountTransactionID uuid.UUID           `gorm:"type:uuid;not null;index" json:"account_transaction_id"`
		AccountTransaction   *AccountTransaction `gorm:"foreignKey:AccountTransactionID;constraint:OnDelete:CASCADE;" json:"account_transaction,omitempty"`

		AccountID uuid.UUID `gorm:"type:uuid;not null;index" json:"account_id"`
		Account   *Account  `gorm:"foreignKey:AccountID;constraint:OnDelete:RESTRICT;" json:"account,omitempty"`

		Debit  float64 `gorm:"type:numeric(18,2);default:0" json:"debit"`
		Credit float64 `gorm:"type:numeric(18,2);default:0" json:"credit"`

		Date     time.Time `gorm:"type:date;not null;index:idx_date_account_tx_entry" json:"date"`
		JVNumber string    `gorm:"type:varchar(255);not null" json:"jv_number"`
	}

	AccountTransactionEntryResponse struct {
		ID             uuid.UUID        `json:"id"`
		CreatedAt      string           `json:"created_at"`
		OrganizationID uuid.UUID        `json:"organization_id"`
		BranchID       uuid.UUID        `json:"branch_id"`
		AccountID      uuid.UUID        `json:"account_id"`
		Account        *AccountResponse `json:"account,omitempty"`
		Debit          float64          `json:"debit"`
		Credit         float64          `json:"credit"`
		Date           string           `json:"date"`
		JVNumber       string           `json:"jv_number"`

		Balance float64 `json:"balance"`
	}
)
