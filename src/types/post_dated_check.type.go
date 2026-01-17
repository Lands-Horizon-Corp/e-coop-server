package types

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type (
	PostDatedCheck struct {
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

		OrganizationID uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_post_dated_check"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_post_dated_check"`
		Branch         *Branch       `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"branch,omitempty"`

		MemberProfileID uuid.UUID      `gorm:"type:uuid"`
		MemberProfile   *MemberProfile `gorm:"foreignKey:MemberProfileID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"member_profile,omitempty"`

		FullName       string `gorm:"type:varchar(255)"`
		PassbookNumber string `gorm:"type:varchar(255)"`

		CheckNumber string    `gorm:"type:varchar(255)"`
		CheckDate   time.Time `gorm:"type:timestamp"`
		ClearDays   int       `gorm:"type:int"`
		DateCleared time.Time `gorm:"type:timestamp"`
		BankID      uuid.UUID `gorm:"type:uuid"`
		Bank        *Bank     `gorm:"foreignKey:BankID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"bank,omitempty"`
		Amount      float64   `gorm:"type:decimal;default:0"`

		ReferenceNumber     string    `gorm:"type:varchar(255)"`
		OfficialReceiptDate time.Time `gorm:"type:timestamp"`
		CollateralUserID    uuid.UUID `gorm:"type:uuid"`

		Description string `gorm:"type:text"`
	}

	PostDatedCheckResponse struct {
		ID                  uuid.UUID              `json:"id"`
		CreatedAt           string                 `json:"created_at"`
		CreatedByID         uuid.UUID              `json:"created_by_id"`
		CreatedBy           *UserResponse          `json:"created_by,omitempty"`
		UpdatedAt           string                 `json:"updated_at"`
		UpdatedByID         uuid.UUID              `json:"updated_by_id"`
		UpdatedBy           *UserResponse          `json:"updated_by,omitempty"`
		OrganizationID      uuid.UUID              `json:"organization_id"`
		Organization        *OrganizationResponse  `json:"organization,omitempty"`
		BranchID            uuid.UUID              `json:"branch_id"`
		Branch              *BranchResponse        `json:"branch,omitempty"`
		MemberProfileID     uuid.UUID              `json:"member_profile_id"`
		MemberProfile       *MemberProfileResponse `json:"member_profile,omitempty"`
		FullName            string                 `json:"full_name"`
		PassbookNumber      string                 `json:"passbook_number"`
		CheckNumber         string                 `json:"check_number"`
		CheckDate           string                 `json:"check_date"`
		ClearDays           int                    `json:"clear_days"`
		DateCleared         string                 `json:"date_cleared"`
		BankID              uuid.UUID              `json:"bank_id"`
		Bank                *BankResponse          `json:"bank,omitempty"`
		Amount              float64                `json:"amount"`
		ReferenceNumber     string                 `json:"reference_number"`
		OfficialReceiptDate string                 `json:"official_receipt_date"`
		CollateralUserID    uuid.UUID              `json:"collateral_user_id"`
		Description         string                 `json:"description"`
	}

	PostDatedCheckRequest struct {
		MemberProfileID     uuid.UUID `json:"member_profile_id,omitempty"`
		FullName            string    `json:"full_name,omitempty"`
		PassbookNumber      string    `json:"passbook_number,omitempty"`
		CheckNumber         string    `json:"check_number,omitempty"`
		CheckDate           time.Time `json:"check_date"`
		ClearDays           int       `json:"clear_days,omitempty"`
		DateCleared         time.Time `json:"date_cleared"`
		BankID              uuid.UUID `json:"bank_id,omitempty"`
		Amount              float64   `json:"amount,omitempty"`
		ReferenceNumber     string    `json:"reference_number,omitempty"`
		OfficialReceiptDate time.Time `json:"official_receipt_date"`
		CollateralUserID    uuid.UUID `json:"collateral_user_id,omitempty"`
		Description         string    `json:"description,omitempty"`
	}
)
