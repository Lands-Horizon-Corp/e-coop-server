package model

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	horizon_services "github.com/lands-horizon/horizon-server/services"
	"gorm.io/gorm"
)

// Enum for charges_mode_of_payment_type
type ChargesModeOfPaymentType string

const (
	Daily       ChargesModeOfPaymentType = "daily"
	Weekly      ChargesModeOfPaymentType = "weekly"
	Monthly     ChargesModeOfPaymentType = "monthly"
	SemiMonthly ChargesModeOfPaymentType = "semi-monthly"
	Quarterly   ChargesModeOfPaymentType = "quarterly"
	SemiAnnual  ChargesModeOfPaymentType = "semi-annual"
	LumpSum     ChargesModeOfPaymentType = "lumpsum"
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

		Name          string                   `gorm:"type:varchar(255)"`
		Description   string                   `gorm:"type:text"`
		ModeOfPayment ChargesModeOfPaymentType `gorm:"type:varchar(20);default:monthly"`

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
		ModeOfPayment       ChargesModeOfPaymentType   `json:"mode_of_payment"`
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
		ChargesRateSchemeID uuid.UUID                `json:"charges_rate_scheme_id" validate:"required"`
		Name                string                   `json:"name,omitempty"`
		Description         string                   `json:"description,omitempty"`
		ModeOfPayment       ChargesModeOfPaymentType `json:"mode_of_payment,omitempty"`
		Rate1               float64                  `json:"rate_1,omitempty"`
		Rate2               float64                  `json:"rate_2,omitempty"`
		Rate3               float64                  `json:"rate_3,omitempty"`
		Rate4               float64                  `json:"rate_4,omitempty"`
		Rate5               float64                  `json:"rate_5,omitempty"`
		Rate6               float64                  `json:"rate_6,omitempty"`
		Rate7               float64                  `json:"rate_7,omitempty"`
		Rate8               float64                  `json:"rate_8,omitempty"`
		Rate9               float64                  `json:"rate_9,omitempty"`
		Rate10              float64                  `json:"rate_10,omitempty"`
		Rate11              float64                  `json:"rate_11,omitempty"`
		Rate12              float64                  `json:"rate_12,omitempty"`
		Rate13              float64                  `json:"rate_13,omitempty"`
		Rate14              float64                  `json:"rate_14,omitempty"`
		Rate15              float64                  `json:"rate_15,omitempty"`
		Rate16              float64                  `json:"rate_16,omitempty"`
		Rate17              float64                  `json:"rate_17,omitempty"`
		Rate18              float64                  `json:"rate_18,omitempty"`
		Rate19              float64                  `json:"rate_19,omitempty"`
		Rate20              float64                  `json:"rate_20,omitempty"`
		Rate21              float64                  `json:"rate_21,omitempty"`
		Rate22              float64                  `json:"rate_22,omitempty"`
	}
)

func (m *Model) ChargesRateByTerm() {
	m.Migration = append(m.Migration, &ChargesRateByTerm{})
	m.ChargesRateByTermManager = horizon_services.NewRepository(horizon_services.RepositoryParams[
		ChargesRateByTerm, ChargesRateByTermResponse, ChargesRateByTermRequest,
	]{
		Preloads: []string{
			"CreatedBy", "UpdatedBy", "DeletedBy", "Branch", "Organization", "ChargesRateScheme",
		},
		Service: m.provider.Service,
		Resource: func(data *ChargesRateByTerm) *ChargesRateByTermResponse {
			if data == nil {
				return nil
			}
			return &ChargesRateByTermResponse{
				ID:                  data.ID,
				CreatedAt:           data.CreatedAt.Format(time.RFC3339),
				CreatedByID:         data.CreatedByID,
				CreatedBy:           m.UserManager.ToModel(data.CreatedBy),
				UpdatedAt:           data.UpdatedAt.Format(time.RFC3339),
				UpdatedByID:         data.UpdatedByID,
				UpdatedBy:           m.UserManager.ToModel(data.UpdatedBy),
				OrganizationID:      data.OrganizationID,
				Organization:        m.OrganizationManager.ToModel(data.Organization),
				BranchID:            data.BranchID,
				Branch:              m.BranchManager.ToModel(data.Branch),
				ChargesRateSchemeID: data.ChargesRateSchemeID,
				ChargesRateScheme:   m.ChargesRateSchemeManager.ToModel(data.ChargesRateScheme),
				Name:                data.Name,
				Description:         data.Description,
				ModeOfPayment:       data.ModeOfPayment,
				Rate1:               data.Rate1,
				Rate2:               data.Rate2,
				Rate3:               data.Rate3,
				Rate4:               data.Rate4,
				Rate5:               data.Rate5,
				Rate6:               data.Rate6,
				Rate7:               data.Rate7,
				Rate8:               data.Rate8,
				Rate9:               data.Rate9,
				Rate10:              data.Rate10,
				Rate11:              data.Rate11,
				Rate12:              data.Rate12,
				Rate13:              data.Rate13,
				Rate14:              data.Rate14,
				Rate15:              data.Rate15,
				Rate16:              data.Rate16,
				Rate17:              data.Rate17,
				Rate18:              data.Rate18,
				Rate19:              data.Rate19,
				Rate20:              data.Rate20,
				Rate21:              data.Rate21,
				Rate22:              data.Rate22,
			}
		},
		Created: func(data *ChargesRateByTerm) []string {
			return []string{
				"charges_rate_by_term.create",
				fmt.Sprintf("charges_rate_by_term.create.%s", data.ID),
				fmt.Sprintf("charges_rate_by_term.create.branch.%s", data.BranchID),
				fmt.Sprintf("charges_rate_by_term.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *ChargesRateByTerm) []string {
			return []string{
				"charges_rate_by_term.update",
				fmt.Sprintf("charges_rate_by_term.update.%s", data.ID),
				fmt.Sprintf("charges_rate_by_term.update.branch.%s", data.BranchID),
				fmt.Sprintf("charges_rate_by_term.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *ChargesRateByTerm) []string {
			return []string{
				"charges_rate_by_term.delete",
				fmt.Sprintf("charges_rate_by_term.delete.%s", data.ID),
				fmt.Sprintf("charges_rate_by_term.delete.branch.%s", data.BranchID),
				fmt.Sprintf("charges_rate_by_term.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}
