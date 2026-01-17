package types

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type (
	JournalVoucherTag struct {
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

		OrganizationID uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_journal_voucher_tag"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_journal_voucher_tag"`
		Branch         *Branch       `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"branch,omitempty"`

		JournalVoucherID *uuid.UUID      `gorm:"type:uuid"`
		JournalVoucher   *JournalVoucher `gorm:"foreignKey:JournalVoucherID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"journal_voucher,omitempty"`

		Name        string `gorm:"type:varchar(50)"`
		Description string `gorm:"type:text"`
		Category    string `gorm:"type:varchar(50)"`
		Color       string `gorm:"type:varchar(20)"`
		Icon        string `gorm:"type:varchar(20)"`
	}

	JournalVoucherTagResponse struct {
		ID               uuid.UUID             `json:"id"`
		CreatedAt        string                `json:"created_at"`
		CreatedByID      uuid.UUID             `json:"created_by_id"`
		CreatedBy        *UserResponse         `json:"created_by,omitempty"`
		UpdatedAt        string                `json:"updated_at"`
		UpdatedByID      uuid.UUID             `json:"updated_by_id"`
		UpdatedBy        *UserResponse         `json:"updated_by,omitempty"`
		OrganizationID   uuid.UUID             `json:"organization_id"`
		Organization     *OrganizationResponse `json:"organization,omitempty"`
		BranchID         uuid.UUID             `json:"branch_id"`
		Branch           *BranchResponse       `json:"branch,omitempty"`
		JournalVoucherID *uuid.UUID            `json:"journal_voucher_id,omitempty"`
		Name             string                `json:"name"`
		Description      string                `json:"description"`
		Category         string                `json:"category"`
		Color            string                `json:"color"`
		Icon             string                `json:"icon"`
	}

	JournalVoucherTagRequest struct {
		JournalVoucherID *uuid.UUID `json:"journal_voucher_id,omitempty"`
		Name             string     `json:"name,omitempty"`
		Description      string     `json:"description,omitempty"`
		Category         string     `json:"category,omitempty"`
		Color            string     `json:"color,omitempty"`
		Icon             string     `json:"icon,omitempty"`
	}
)
