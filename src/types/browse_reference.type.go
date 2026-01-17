package types

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	InterestTypeYear   InterestType = "year"
	InterestTypeDate   InterestType = "date"
	InterestTypeAmount InterestType = "amount"
	InterestTypeNone   InterestType = "none"
)

type (
	InterestType    string
	BrowseReference struct {
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

		OrganizationID uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_browse_reference"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_browse_reference"`
		Branch         *Branch       `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"branch,omitempty"`

		Name        string `gorm:"type:varchar(255);not null" json:"name" validate:"required,max=255"`
		Description string `gorm:"type:text" json:"description"`

		InterestRate   float64 `gorm:"type:decimal(15,6);default:0" json:"interest_rate"`
		MinimumBalance float64 `gorm:"type:decimal(15,2);default:0" json:"minimum_balance"`
		Charges        float64 `gorm:"type:decimal(15,2);default:0" json:"charges"`

		AccountID *uuid.UUID `gorm:"type:uuid"`
		Account   *Account   `gorm:"foreignKey:AccountID;constraint:OnDelete:SET NULL;" json:"account,omitempty"`

		MemberTypeID *uuid.UUID  `gorm:"type:uuid"`
		MemberType   *MemberType `gorm:"foreignKey:MemberTypeID;constraint:OnDelete:SET NULL;" json:"member_type,omitempty"`

		InterestType InterestType `gorm:"type:varchar(20);not null;default:'year'" json:"interest_type" validate:"required,oneof=year date amount"`

		DefaultMinimumBalance float64 `gorm:"type:decimal(15,2);default:0" json:"default_minimum_balance"`
		DefaultInterestRate   float64 `gorm:"type:decimal(15,6);default:0" json:"default_interest_rate"`

		InterestRatesByYear   []*InterestRateByYear   `gorm:"foreignKey:BrowseReferenceID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"interest_rates_by_year,omitempty"`
		InterestRatesByDate   []*InterestRateByDate   `gorm:"foreignKey:BrowseReferenceID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"interest_rates_by_date,omitempty"`
		InterestRatesByAmount []*InterestRateByAmount `gorm:"foreignKey:BrowseReferenceID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"interest_rates_by_amount,omitempty"`
	}

	BrowseReferenceResponse struct {
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

		Name                  string              `json:"name"`
		Description           string              `json:"description"`
		InterestRate          float64             `json:"interest_rate"`
		MinimumBalance        float64             `json:"minimum_balance"`
		Charges               float64             `json:"charges"`
		AccountID             *uuid.UUID          `json:"account_id,omitempty"`
		Account               *AccountResponse    `json:"account,omitempty"`
		MemberTypeID          *uuid.UUID          `json:"member_type_id,omitempty"`
		MemberType            *MemberTypeResponse `json:"member_type,omitempty"`
		InterestType          InterestType        `json:"interest_type"`
		DefaultMinimumBalance float64             `json:"default_minimum_balance"`
		DefaultInterestRate   float64             `json:"default_interest_rate"`

		InterestRatesByYear   []*InterestRateByYearResponse   `json:"interest_rates_by_year,omitempty"`
		InterestRatesByDate   []*InterestRateByDateResponse   `json:"interest_rates_by_date,omitempty"`
		InterestRatesByAmount []*InterestRateByAmountResponse `json:"interest_rates_by_amount,omitempty"`
	}

	BrowseReferenceRequest struct {
		ID          *uuid.UUID `json:"id"`
		Name        string     `json:"name" validate:"required,max=255"`
		Description string     `json:"description"`

		InterestRate   float64 `json:"interest_rate"`
		MinimumBalance float64 `json:"minimum_balance"`
		Charges        float64 `json:"charges"`

		AccountID    *uuid.UUID   `json:"account_id"`
		MemberTypeID *uuid.UUID   `json:"member_type_id"`
		InterestType InterestType `json:"interest_type" validate:"required,oneof=year date amount none"`

		DefaultMinimumBalance float64 `json:"default_minimum_balance"`
		DefaultInterestRate   float64 `json:"default_interest_rate"`

		InterestRatesByYear   []*InterestRateByYearRequest   `json:"interest_rates_by_year,omitempty"`
		InterestRatesByDate   []*InterestRateByDateRequest   `json:"interest_rates_by_date,omitempty"`
		InterestRatesByAmount []*InterestRateByAmountRequest `json:"interest_rates_by_amount,omitempty"`

		InterestRatesByYearDeleted   uuid.UUIDs `json:"interest_rates_by_year_deleted,omitempty"`
		InterestRatesByDateDeleted   uuid.UUIDs `json:"interest_rates_by_date_deleted,omitempty"`
		InterestRatesByAmountDeleted uuid.UUIDs `json:"interest_rates_by_amount_deleted,omitempty"`
	}
)
