package types

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type (
	CashCount struct {
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

		OrganizationID uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_cash_count"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_cash_count"`
		Branch         *Branch       `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"branch,omitempty"`

		EmployeeUserID     uuid.UUID         `gorm:"type:uuid;not null"`
		EmployeeUser       *User             `gorm:"foreignKey:EmployeeUserID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"employee_user,omitempty"`
		TransactionBatchID uuid.UUID         `gorm:"type:uuid;not null"`
		TransactionBatch   *TransactionBatch `gorm:"foreignKey:TransactionBatchID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"transaction_batch,omitempty"`

		CurrencyID uuid.UUID `gorm:"type:uuid;not null"`
		Currency   *Currency `gorm:"foreignKey:CurrencyID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"currency,omitempty"`

		Name       string  `gorm:"type:varchar(100);not null"`
		BillAmount float64 `gorm:"type:decimal"`
		Quantity   int     `gorm:"type:int"`
		Amount     float64 `gorm:"type:decimal"`
	}

	CashCountResponse struct {
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
		EmployeeUserID     uuid.UUID                 `json:"employee_user_id"`
		EmployeeUser       *UserResponse             `json:"employee_user,omitempty"`
		TransactionBatchID uuid.UUID                 `json:"transaction_batch_id"`
		TransactionBatch   *TransactionBatchResponse `json:"transaction_batch,omitempty"`
		CurrencyID         uuid.UUID                 `json:"currency_id"`
		Currency           *CurrencyResponse         `json:"currency,omitempty"`
		BillAmount         float64                   `json:"bill_amount"`
		Quantity           int                       `json:"quantity"`
		Amount             float64                   `json:"amount"`
		Name               string                    `json:"name"`
	}

	CashCountRequest struct {
		ID                 *uuid.UUID `json:"id,omitempty"`
		EmployeeUserID     uuid.UUID  `json:"employee_user_id" validate:"required"`
		TransactionBatchID uuid.UUID  `json:"transaction_batch_id" validate:"required"`
		CurrencyID         uuid.UUID  `json:"currency_id,omitempty"`
		BillAmount         float64    `json:"bill_amount,omitempty"`
		Quantity           int        `json:"quantity,omitempty"`
		Amount             float64    `json:"amount,omitempty"`
		Name               string     `json:"name,omitempty"`
	}
)
