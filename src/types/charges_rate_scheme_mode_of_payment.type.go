package types

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type (
	ChargesRateSchemeModeOfPayment struct {
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

		OrganizationID uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_charges_rate_scheme_mode_of_payment"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_charges_rate_scheme_mode_of_payment"`
		Branch         *Branch       `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"branch,omitempty"`

		ChargesRateSchemeID uuid.UUID          `gorm:"type:uuid;not null"`
		ChargesRateScheme   *ChargesRateScheme `gorm:"foreignKey:ChargesRateSchemeID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"charges_rate_scheme,omitempty"`

		From float64 `gorm:"type:decimal;default:0"`
		To   float64 `gorm:"type:decimal;default:0"`

		Column1  float64 `gorm:"type:decimal;default:0"`
		Column2  float64 `gorm:"type:decimal;default:0"`
		Column3  float64 `gorm:"type:decimal;default:0"`
		Column4  float64 `gorm:"type:decimal;default:0"`
		Column5  float64 `gorm:"type:decimal;default:0"`
		Column6  float64 `gorm:"type:decimal;default:0"`
		Column7  float64 `gorm:"type:decimal;default:0"`
		Column8  float64 `gorm:"type:decimal;default:0"`
		Column9  float64 `gorm:"type:decimal;default:0"`
		Column10 float64 `gorm:"type:decimal;default:0"`
		Column11 float64 `gorm:"type:decimal;default:0"`
		Column12 float64 `gorm:"type:decimal;default:0"`
		Column13 float64 `gorm:"type:decimal;default:0"`
		Column14 float64 `gorm:"type:decimal;default:0"`
		Column15 float64 `gorm:"type:decimal;default:0"`
		Column16 float64 `gorm:"type:decimal;default:0"`
		Column17 float64 `gorm:"type:decimal;default:0"`
		Column18 float64 `gorm:"type:decimal;default:0"`
		Column19 float64 `gorm:"type:decimal;default:0"`
		Column20 float64 `gorm:"type:decimal;default:0"`
		Column21 float64 `gorm:"type:decimal;default:0"`
		Column22 float64 `gorm:"type:decimal;default:0"`
	}

	ChargesRateSchemeModeOfPaymentResponse struct {
		ID                  uuid.UUID                  `json:"id"`
		CreatedAt           string                     `json:"created_at"`
		CreatedByID         uuid.UUID                  `json:"created_by_id"`
		CreatedBy           *UserResponse              `json:"created_by,omitempty"`
		UpdatedAt           string                     `json:"updated_at"`
		UpdatedByID         uuid.UUID                  `json:"updated_by_id"`
		UpdatedBy           *UserResponse              `json:"updated_by,omitempty"`
		OrganizationID      uuid.UUID                  `json:"organization_id"`
		Organization        *OrganizationResponse      `json:"organization,omitempty"`
		BranchID            uuid.UUID                  `json:"branch_id"`
		Branch              *BranchResponse            `json:"branch,omitempty"`
		ChargesRateSchemeID uuid.UUID                  `json:"charges_rate_scheme_id"`
		ChargesRateScheme   *ChargesRateSchemeResponse `json:"charges_rate_scheme,omitempty"`
		From                float64                    `json:"from"`
		To                  float64                    `json:"to"`

		Column1  float64 `json:"column1"`
		Column2  float64 `json:"column2"`
		Column3  float64 `json:"column3"`
		Column4  float64 `json:"column4"`
		Column5  float64 `json:"column5"`
		Column6  float64 `json:"column6"`
		Column7  float64 `json:"column7"`
		Column8  float64 `json:"column8"`
		Column9  float64 `json:"column9"`
		Column10 float64 `json:"column10"`
		Column11 float64 `json:"column11"`
		Column12 float64 `json:"column12"`
		Column13 float64 `json:"column13"`
		Column14 float64 `json:"column14"`
		Column15 float64 `json:"column15"`
		Column16 float64 `json:"column16"`
		Column17 float64 `json:"column17"`
		Column18 float64 `json:"column18"`
		Column19 float64 `json:"column19"`
		Column20 float64 `json:"column20"`
		Column21 float64 `json:"column21"`
		Column22 float64 `json:"column22"`
	}

	ChargesRateSchemeModeOfPaymentRequest struct {
		ID   *uuid.UUID `json:"id,omitempty"`
		From float64    `json:"from,omitempty"`
		To   float64    `json:"to,omitempty"`

		Column1  float64 `json:"column1,omitempty"`
		Column2  float64 `json:"column2,omitempty"`
		Column3  float64 `json:"column3,omitempty"`
		Column4  float64 `json:"column4,omitempty"`
		Column5  float64 `json:"column5,omitempty"`
		Column6  float64 `json:"column6,omitempty"`
		Column7  float64 `json:"column7,omitempty"`
		Column8  float64 `json:"column8,omitempty"`
		Column9  float64 `json:"column9,omitempty"`
		Column10 float64 `json:"column10,omitempty"`
		Column11 float64 `json:"column11,omitempty"`
		Column12 float64 `json:"column12,omitempty"`
		Column13 float64 `json:"column13,omitempty"`
		Column14 float64 `json:"column14,omitempty"`
		Column15 float64 `json:"column15,omitempty"`
		Column16 float64 `json:"column16,omitempty"`
		Column17 float64 `json:"column17,omitempty"`
		Column18 float64 `json:"column18,omitempty"`
		Column19 float64 `json:"column19,omitempty"`
		Column20 float64 `json:"column20,omitempty"`
		Column21 float64 `json:"column21,omitempty"`
		Column22 float64 `json:"column22,omitempty"`
	}
)
