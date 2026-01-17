package types

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type (
	AutomaticLoanDeduction struct {
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

		OrganizationID uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_automatic_loan_deduction" json:"organization_id"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_automatic_loan_deduction" json:"branch_id"`
		Branch         *Branch       `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"branch,omitempty"`

		AccountID           *uuid.UUID         `gorm:"type:uuid" json:"account_id"`
		Account             *Account           `gorm:"foreignKey:AccountID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"account,omitempty"`
		ComputationSheetID  *uuid.UUID         `gorm:"type:uuid" json:"computation_sheet_id"`
		ComputationSheet    *ComputationSheet  `gorm:"foreignKey:ComputationSheetID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"computation_sheet,omitempty"`
		ChargesRateSchemeID *uuid.UUID         `gorm:"type:uuid" json:"charges_rate_scheme_id"`
		ChargesRateScheme   *ChargesRateScheme `gorm:"foreignKey:ChargesRateSchemeID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"charges_rate_scheme,omitempty"`

		ChargesPercentage1 float64 `gorm:"type:decimal" json:"charges_percentage_1"`
		ChargesPercentage2 float64 `gorm:"type:decimal" json:"charges_percentage_2"`
		ChargesAmount      float64 `gorm:"type:decimal" json:"charges_amount"`
		ChargesDivisor     float64 `gorm:"type:decimal" json:"charges_divisor"`

		MinAmount float64 `gorm:"type:decimal" json:"min_amount"`
		MaxAmount float64 `gorm:"type:decimal" json:"max_amount"`

		Anum int16 `gorm:"type:int;default:0" json:"anum"`

		NumberOfMonths int `gorm:"type:int" json:"number_of_months"`

		AddOn          bool `gorm:"type:boolean;default:false" json:"add_on"`
		AoRest         bool `gorm:"type:boolean;default:false" json:"ao_rest"`
		ExcludeRenewal bool `gorm:"type:boolean;default:false" json:"exclude_renewal"`
		Ct             int  `gorm:"type:int" json:"ct"`

		Name        string `gorm:"type:varchar(255)" json:"name"`
		Description string `gorm:"type:text" json:"description"`
	}

	AutomaticLoanDeductionResponse struct {
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

		AccountID           *uuid.UUID                 `json:"account_id,omitempty"`
		Account             *AccountResponse           `json:"account,omitempty"`
		ComputationSheetID  *uuid.UUID                 `json:"computation_sheet_id,omitempty"`
		ComputationSheet    *ComputationSheetResponse  `json:"computation_sheet,omitempty"`
		ChargesRateSchemeID *uuid.UUID                 `json:"charges_rate_scheme_id,omitempty"`
		ChargesRateScheme   *ChargesRateSchemeResponse `json:"charges_rate_scheme,omitempty"`

		ChargesPercentage1 float64 `json:"charges_percentage_1"`
		ChargesPercentage2 float64 `json:"charges_percentage_2"`
		ChargesAmount      float64 `json:"charges_amount"`
		ChargesDivisor     float64 `json:"charges_divisor"`

		MinAmount float64 `json:"min_amount"`
		MaxAmount float64 `json:"max_amount"`

		Anum int16 `json:"anum"`

		NumberOfMonths int `json:"number_of_months"`

		AddOn          bool `json:"add_on"`
		AoRest         bool `json:"ao_rest"`
		ExcludeRenewal bool `json:"exclude_renewal"`
		Ct             int  `json:"ct"`

		Name        string `json:"name"`
		Description string `json:"description"`
	}

	AutomaticLoanDeductionRequest struct {
		AccountID           *uuid.UUID `json:"account_id" validate:"required"`
		ComputationSheetID  *uuid.UUID `json:"computation_sheet_id,omitempty"`
		ChargesRateSchemeID *uuid.UUID `json:"charges_rate_scheme_id,omitempty"`
		ChargesPercentage1  float64    `json:"charges_percentage_1,omitempty"`
		ChargesPercentage2  float64    `json:"charges_percentage_2,omitempty"`
		ChargesAmount       float64    `json:"charges_amount,omitempty"`
		ChargesDivisor      float64    `json:"charges_divisor,omitempty"`
		MinAmount           float64    `json:"min_amount,omitempty"`
		MaxAmount           float64    `json:"max_amount,omitempty"`
		Anum                int16      `json:"anum,omitempty"`
		NumberOfMonths      int        `json:"number_of_months,omitempty"`
		AddOn               bool       `json:"add_on,omitempty"`
		AoRest              bool       `json:"ao_rest,omitempty"`
		ExcludeRenewal      bool       `json:"exclude_renewal,omitempty"`
		Ct                  int        `json:"ct,omitempty"`
		Name                string     `json:"name,omitempty"`
		Description         string     `json:"description,omitempty"`
	}
)
