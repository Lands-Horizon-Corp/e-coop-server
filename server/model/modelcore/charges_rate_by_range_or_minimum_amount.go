package modelcore

import (
	"context"
	"fmt"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/services"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type (
	ChargesRateByRangeOrMinimumAmount struct {
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

		OrganizationID uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_charges_rate_by_range_or_minimum_amount"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_charges_rate_by_range_or_minimum_amount"`
		Branch         *Branch       `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"branch,omitempty"`

		ChargesRateSchemeID uuid.UUID          `gorm:"type:uuid;not null"`
		ChargesRateScheme   *ChargesRateScheme `gorm:"foreignKey:ChargesRateSchemeID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"charges_rate_scheme,omitempty"`

		From          float64 `gorm:"type:decimal;default:0"`
		To            float64 `gorm:"type:decimal;default:0"`
		Charge        float64 `gorm:"type:decimal;default:0"`
		Amount        float64 `gorm:"type:decimal;default:0"`
		MinimumAmount float64 `gorm:"type:decimal;default:0"`
	}

	ChargesRateByRangeOrMinimumAmountResponse struct {
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
		Charge              float64                    `json:"charge"`
		Amount              float64                    `json:"amount"`
		MinimumAmount       float64                    `json:"minimum_amount"`
	}

	ChargesRateByRangeOrMinimumAmountRequest struct {
		ID            *uuid.UUID `json:"id,omitempty"`
		From          float64    `json:"from,omitempty"`
		To            float64    `json:"to,omitempty"`
		Charge        float64    `json:"charge,omitempty"`
		Amount        float64    `json:"amount,omitempty"`
		MinimumAmount float64    `json:"minimum_amount,omitempty"`
	}
)

func (m *ModelCore) chargesRateByRangeOrMinimumAmount() {
	m.Migration = append(m.Migration, &ChargesRateByRangeOrMinimumAmount{})
	m.ChargesRateByRangeOrMinimumAmountManager = services.NewRepository(services.RepositoryParams[
		ChargesRateByRangeOrMinimumAmount, ChargesRateByRangeOrMinimumAmountResponse, ChargesRateByRangeOrMinimumAmountRequest,
	]{
		Preloads: []string{
			"CreatedBy", "UpdatedBy", "ChargesRateScheme",
		},
		Service: m.provider.Service,
		Resource: func(data *ChargesRateByRangeOrMinimumAmount) *ChargesRateByRangeOrMinimumAmountResponse {
			if data == nil {
				return nil
			}
			return &ChargesRateByRangeOrMinimumAmountResponse{
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
				From:                data.From,
				To:                  data.To,
				Charge:              data.Charge,
				Amount:              data.Amount,
				MinimumAmount:       data.MinimumAmount,
			}
		},
		Created: func(data *ChargesRateByRangeOrMinimumAmount) []string {
			return []string{
				"charges_rate_by_range_or_minimum_amount.create",
				fmt.Sprintf("charges_rate_by_range_or_minimum_amount.create.%s", data.ID),
				fmt.Sprintf("charges_rate_by_range_or_minimum_amount.create.branch.%s", data.BranchID),
				fmt.Sprintf("charges_rate_by_range_or_minimum_amount.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *ChargesRateByRangeOrMinimumAmount) []string {
			return []string{
				"charges_rate_by_range_or_minimum_amount.update",
				fmt.Sprintf("charges_rate_by_range_or_minimum_amount.update.%s", data.ID),
				fmt.Sprintf("charges_rate_by_range_or_minimum_amount.update.branch.%s", data.BranchID),
				fmt.Sprintf("charges_rate_by_range_or_minimum_amount.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *ChargesRateByRangeOrMinimumAmount) []string {
			return []string{
				"charges_rate_by_range_or_minimum_amount.delete",
				fmt.Sprintf("charges_rate_by_range_or_minimum_amount.delete.%s", data.ID),
				fmt.Sprintf("charges_rate_by_range_or_minimum_amount.delete.branch.%s", data.BranchID),
				fmt.Sprintf("charges_rate_by_range_or_minimum_amount.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

// ChargesRateByRangeOrMinimumAmountCurrentBranch retrieves all charges rate by range or minimum amount for the specified organization and branch
func (m *ModelCore) ChargesRateByRangeOrMinimumAmountCurrentBranch(context context.Context, orgID uuid.UUID, branchID uuid.UUID) ([]*ChargesRateByRangeOrMinimumAmount, error) {
	return m.ChargesRateByRangeOrMinimumAmountManager.Find(context, &ChargesRateByRangeOrMinimumAmount{
		OrganizationID: orgID,
		BranchID:       branchID,
	})
}
