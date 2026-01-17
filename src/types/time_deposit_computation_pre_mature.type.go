package types

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type (
	TimeDepositComputationPreMature struct {
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

		OrganizationID uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_time_deposit_computation_pre_mature"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_time_deposit_computation_pre_mature"`
		Branch         *Branch       `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"branch,omitempty"`

		TimeDepositTypeID uuid.UUID        `gorm:"type:uuid;not null"`
		TimeDepositType   *TimeDepositType `gorm:"foreignKey:TimeDepositTypeID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"time_deposit_type,omitempty"`

		Terms int     `gorm:"default:0"`
		From  int     `gorm:"default:0"`
		To    int     `gorm:"default:0"`
		Rate  float64 `gorm:"type:decimal;default:0"`
	}

	TimeDepositComputationPreMatureResponse struct {
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
		TimeDepositTypeID uuid.UUID                `json:"time_deposit_type_id"`
		TimeDepositType   *TimeDepositTypeResponse `json:"time_deposit_type,omitempty"`
		Terms             int                      `json:"terms"`
		From              int                      `json:"from"`
		To                int                      `json:"to"`
		Rate              float64                  `json:"rate"`
	}

	TimeDepositComputationPreMatureRequest struct {
		ID                *uuid.UUID `json:"id,omitempty"`
		TimeDepositTypeID uuid.UUID  `json:"time_deposit_type_id" validate:"required"`
		Terms             int        `json:"terms,omitempty"`
		From              int        `json:"from,omitempty"`
		To                int        `json:"to,omitempty"`
		Rate              float64    `json:"rate,omitempty"`
	}
)
