package types

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type (
	DisbursementTransaction struct {
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

		OrganizationID uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_disbursement_transaction"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_disbursement_transaction"`
		Branch         *Branch       `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"branch,omitempty"`

		Disbursement       *Disbursement     `gorm:"foreignKey:DisbursementID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"disbursement,omitempty"`
		TransactionBatchID uuid.UUID         `gorm:"type:uuid;not null"`
		TransactionBatch   *TransactionBatch `gorm:"foreignKey:TransactionBatchID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"transaction_batch,omitempty"`
		EmployeeUserID     uuid.UUID         `gorm:"type:uuid;not null"`
		EmployeeUser       *User             `gorm:"foreignKey:EmployeeUserID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"employee_user,omitempty"`

		DisbursementID  uuid.UUID `gorm:"type:uuid;not null"`
		ReferenceNumber string    `gorm:"type:varchar(50)"`
		Amount          float64   `gorm:"type:decimal"`
		Description     string    `gorm:"type:text"`
		EmployeeName    string    `gorm:"type:varchar(100)"`
	}

	DisbursementTransactionResponse struct {
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

		DisbursementID     uuid.UUID                 `json:"disbursement_id"`
		Disbursement       *DisbursementResponse     `json:"disbursement,omitempty"`
		TransactionBatchID uuid.UUID                 `json:"transaction_batch_id"`
		TransactionBatch   *TransactionBatchResponse `json:"transaction_batch,omitempty"`
		EmployeeUserID     uuid.UUID                 `json:"employee_user_id"`
		EmployeeUser       *UserResponse             `json:"employee_user,omitempty"`

		ReferenceNumber string  `json:"reference_number"`
		Amount          float64 `json:"amount"`
	}

	DisbursementTransactionRequest struct {
		DisbursementID *uuid.UUID `json:"disbursement_id" validate:"required"`

		Description              string `json:"description,omitempty"`
		IsReferenceNumberChecked bool   `json:"is_reference_number_checked,omitempty"`

		ReferenceNumber string  `json:"reference_number,omitempty"`
		Amount          float64 `json:"amount,omitempty"`
	}
)
