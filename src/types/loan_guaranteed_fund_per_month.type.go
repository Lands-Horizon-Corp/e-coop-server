package types

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type (
	LoanGuaranteedFundPerMonth struct {
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

		OrganizationID uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_loan_guaranteed_fund_per_month"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_loan_guaranteed_fund_per_month"`
		Branch         *Branch       `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"branch,omitempty"`

		Month              int `gorm:"type:int;default:0"`
		LoanGuaranteedFund int `gorm:"type:int;default:0"`
	}

	LoanGuaranteedFundPerMonthResponse struct {
		ID                 uuid.UUID             `json:"id"`
		CreatedAt          string                `json:"created_at"`
		CreatedByID        uuid.UUID             `json:"created_by_id"`
		CreatedBy          *UserResponse         `json:"created_by,omitempty"`
		UpdatedAt          string                `json:"updated_at"`
		UpdatedByID        uuid.UUID             `json:"updated_by_id"`
		UpdatedBy          *UserResponse         `json:"updated_by,omitempty"`
		OrganizationID     uuid.UUID             `json:"organization_id"`
		Organization       *OrganizationResponse `json:"organization,omitempty"`
		BranchID           uuid.UUID             `json:"branch_id"`
		Branch             *BranchResponse       `json:"branch,omitempty"`
		Month              int                   `json:"month"`
		LoanGuaranteedFund int                   `json:"loan_guaranteed_fund"`
	}

	LoanGuaranteedFundPerMonthRequest struct {
		Month              int `json:"month,omitempty"`
		LoanGuaranteedFund int `json:"loan_guaranteed_fund,omitempty"`
	}
)
