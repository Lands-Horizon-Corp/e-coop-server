package types

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type (
	ChargesRateByTerm struct {
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

		OrganizationID uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_charges_rate_by_term"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_charges_rate_by_term"`
		Branch         *Branch       `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"branch,omitempty"`

		ChargesRateSchemeID uuid.UUID          `gorm:"type:uuid;not null"`
		ChargesRateScheme   *ChargesRateScheme `gorm:"foreignKey:ChargesRateSchemeID;constraint:OnDelete:SET NULL,OnUpdate:CASCADE;" json:"charges_rate_scheme,omitempty"`

		Name          string            `gorm:"type:varchar(255)"`
		Description   string            `gorm:"type:text"`
		ModeOfPayment LoanModeOfPayment `gorm:"type:varchar(20);default:monthly"`

		Rate1  float64 `gorm:"type:decimal;default:0"`
		Rate2  float64 `gorm:"type:decimal;default:0"`
		Rate3  float64 `gorm:"type:decimal;default:0"`
		Rate4  float64 `gorm:"type:decimal;default:0"`
		Rate5  float64 `gorm:"type:decimal;default:0"`
		Rate6  float64 `gorm:"type:decimal;default:0"`
		Rate7  float64 `gorm:"type:decimal;default:0"`
		Rate8  float64 `gorm:"type:decimal;default:0"`
		Rate9  float64 `gorm:"type:decimal;default:0"`
		Rate10 float64 `gorm:"type:decimal;default:0"`
		Rate11 float64 `gorm:"type:decimal;default:0"`
		Rate12 float64 `gorm:"type:decimal;default:0"`
		Rate13 float64 `gorm:"type:decimal;default:0"`
		Rate14 float64 `gorm:"type:decimal;default:0"`
		Rate15 float64 `gorm:"type:decimal;default:0"`
		Rate16 float64 `gorm:"type:decimal;default:0"`
		Rate17 float64 `gorm:"type:decimal;default:0"`
		Rate18 float64 `gorm:"type:decimal;default:0"`
		Rate19 float64 `gorm:"type:decimal;default:0"`
		Rate20 float64 `gorm:"type:decimal;default:0"`
		Rate21 float64 `gorm:"type:decimal;default:0"`
		Rate22 float64 `gorm:"type:decimal;default:0"`
	}

	ChargesRateByTermResponse struct {
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
		Name                string                     `json:"name"`
		Description         string                     `json:"description"`
		ModeOfPayment       LoanModeOfPayment          `json:"mode_of_payment"`
		Rate1               float64                    `json:"rate_1"`
		Rate2               float64                    `json:"rate_2"`
		Rate3               float64                    `json:"rate_3"`
		Rate4               float64                    `json:"rate_4"`
		Rate5               float64                    `json:"rate_5"`
		Rate6               float64                    `json:"rate_6"`
		Rate7               float64                    `json:"rate_7"`
		Rate8               float64                    `json:"rate_8"`
		Rate9               float64                    `json:"rate_9"`
		Rate10              float64                    `json:"rate_10"`
		Rate11              float64                    `json:"rate_11"`
		Rate12              float64                    `json:"rate_12"`
		Rate13              float64                    `json:"rate_13"`
		Rate14              float64                    `json:"rate_14"`
		Rate15              float64                    `json:"rate_15"`
		Rate16              float64                    `json:"rate_16"`
		Rate17              float64                    `json:"rate_17"`
		Rate18              float64                    `json:"rate_18"`
		Rate19              float64                    `json:"rate_19"`
		Rate20              float64                    `json:"rate_20"`
		Rate21              float64                    `json:"rate_21"`
		Rate22              float64                    `json:"rate_22"`
	}

	ChargesRateByTermRequest struct {
		ID *uuid.UUID `json:"id,omitempty"`

		Name          string            `json:"name,omitempty"`
		Description   string            `json:"description,omitempty"`
		ModeOfPayment LoanModeOfPayment `json:"mode_of_payment,omitempty"`
		Rate1         float64           `json:"rate_1,omitempty"`
		Rate2         float64           `json:"rate_2,omitempty"`
		Rate3         float64           `json:"rate_3,omitempty"`
		Rate4         float64           `json:"rate_4,omitempty"`
		Rate5         float64           `json:"rate_5,omitempty"`
		Rate6         float64           `json:"rate_6,omitempty"`
		Rate7         float64           `json:"rate_7,omitempty"`
		Rate8         float64           `json:"rate_8,omitempty"`
		Rate9         float64           `json:"rate_9,omitempty"`
		Rate10        float64           `json:"rate_10,omitempty"`
		Rate11        float64           `json:"rate_11,omitempty"`
		Rate12        float64           `json:"rate_12,omitempty"`
		Rate13        float64           `json:"rate_13,omitempty"`
		Rate14        float64           `json:"rate_14,omitempty"`
		Rate15        float64           `json:"rate_15,omitempty"`
		Rate16        float64           `json:"rate_16,omitempty"`
		Rate17        float64           `json:"rate_17,omitempty"`
		Rate18        float64           `json:"rate_18,omitempty"`
		Rate19        float64           `json:"rate_19,omitempty"`
		Rate20        float64           `json:"rate_20,omitempty"`
		Rate21        float64           `json:"rate_21,omitempty"`
		Rate22        float64           `json:"rate_22,omitempty"`
	}
)
