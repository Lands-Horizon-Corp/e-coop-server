package types

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type (
	MemberBankCard struct {
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

		OrganizationID uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_member_card"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_member_card"`
		Branch         *Branch       `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"branch,omitempty"`

		BankID *uuid.UUID `gorm:"type:uuid"`
		Bank   *Bank      `gorm:"foreignKey:BankID;constraint:OnDelete:SET NULL,OnUpdate:CASCADE;" json:"bank,omitempty"`

		MemberProfileID *uuid.UUID     `gorm:"type:uuid"`
		MemberProfile   *MemberProfile `gorm:"foreignKey:MemberProfileID;constraint:OnDelete:SET NULL,OnUpdate:CASCADE;" json:"member_profile,omitempty"`

		AccountNumber  string    `gorm:"type:varchar(50);not null"`
		CardName       string    `gorm:"type:varchar(50);not null"`
		ExpirationDate time.Time `gorm:"not null"`
		IsDefault      bool      `gorm:"not null;default:false"`
	}

	MemberBankCardResponse struct {
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

		BankID *uuid.UUID    `json:"bank_id,omitempty"`
		Bank   *BankResponse `json:"bank,omitempty"`

		MemberProfileID *uuid.UUID             `json:"member_profile_id,omitempty"`
		MemberProfile   *MemberProfileResponse `json:"member_profile,omitempty"`

		AccountNumber  string `json:"account_number"`
		CardName       string `json:"card_name"`
		ExpirationDate string `json:"expiration_date"`
		IsDefault      bool   `json:"is_default"`
	}

	MemberBankCardRequest struct {
		AccountNumber   string     `json:"account_number" validate:"required,min=1,max=50"`
		CardName        string     `json:"card_name" validate:"required,min=1,max=50"`
		ExpirationDate  time.Time  `json:"expiration_date" validate:"required"`
		IsDefault       bool       `json:"is_default"`
		BankID          *uuid.UUID `json:"bank_id,omitempty"`
		MemberProfileID *uuid.UUID `json:"member_profile_id,omitempty"`
	}
)
