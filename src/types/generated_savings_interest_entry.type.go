package types

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type (
	GeneratedSavingsInterestEntry struct {
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

		OrganizationID uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_generate_savings_interest_entry"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_generate_savings_interest_entry"`
		Branch         *Branch       `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"branch,omitempty"`

		GeneratedSavingsInterestID uuid.UUID                 `gorm:"type:uuid;not null;index:idx_generated_savings_interest_entry"`
		GeneratedSavingsInterest   *GeneratedSavingsInterest `gorm:"foreignKey:GeneratedSavingsInterestID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"generated_savings_interest,omitempty"`

		AccountID uuid.UUID `gorm:"type:uuid;not null;index:idx_account_member_profile_entry"`
		Account   *Account  `gorm:"foreignKey:AccountID;constraint:OnDelete:RESTRICT;" json:"account,omitempty"`

		MemberProfileID uuid.UUID      `gorm:"type:uuid;not null;index:idx_account_member_profile_entry"`
		MemberProfile   *MemberProfile `gorm:"foreignKey:MemberProfileID;constraint:OnDelete:RESTRICT;" json:"member_profile,omitempty"`

		EndingBalance  float64 `gorm:"type:decimal(15,2);not null" json:"ending_balance" validate:"required"`
		InterestAmount float64 `gorm:"type:decimal(15,2);not null" json:"interest_amount" validate:"required"`
		InterestTax    float64 `gorm:"type:decimal(15,2);not null" json:"interest_tax" validate:"required"`
	}

	GeneratedSavingsInterestEntryResponse struct {
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

		GeneratedSavingsInterestID uuid.UUID                         `json:"generated_savings_interest_id"`
		GeneratedSavingsInterest   *GeneratedSavingsInterestResponse `json:"generated_savings_interest,omitempty"`
		AccountID                  uuid.UUID                         `json:"account_id"`
		Account                    *AccountResponse                  `json:"account,omitempty"`
		MemberProfileID            uuid.UUID                         `json:"member_profile_id"`
		MemberProfile              *MemberProfileResponse            `json:"member_profile,omitempty"`
		EndingBalance              float64                           `json:"ending_balance"`
		InterestAmount             float64                           `json:"interest_amount"`
		InterestTax                float64                           `json:"interest_tax"`
	}

	GeneratedSavingsInterestEntryRequest struct {
		AccountID       uuid.UUID `json:"account_id" validate:"required"`
		MemberProfileID uuid.UUID `json:"member_profile_id" validate:"required"`
		InterestAmount  float64   `json:"interest_amount" validate:"required"`
		InterestTax     float64   `json:"interest_tax" validate:"required"`
	}

	GeneratedSavingsInterestEntryDailyBalance struct {
		Balance float64 `json:"balance"`
		Date    string  `json:"date"`
		Type    string  `json:"type"` // "increase", "decrease", "no_change"
	}
	GeneratedSavingsInterestEntryDailyBalanceResponse struct {
		BeginningBalance    float64                                     `json:"beginning_balance"`
		EndingBalance       float64                                     `json:"ending_balance"`
		AverageDailyBalance float64                                     `json:"average_daily_balance"`
		LowestBalance       float64                                     `json:"lowest_balance"`
		HighestBalance      float64                                     `json:"highest_balance"`
		DailyBalance        []GeneratedSavingsInterestEntryDailyBalance `json:"daily_balance"`
		Account             *AccountResponse                            `json:"account,omitempty"`
		MemberProfile       *MemberProfileResponse                      `json:"member_profile,omitempty"`
	}
)
