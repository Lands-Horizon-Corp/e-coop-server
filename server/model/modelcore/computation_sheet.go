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
	ComputationSheet struct {
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

		OrganizationID uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_computation_sheet"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_computation_sheet"`
		Branch         *Branch       `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"branch,omitempty"`
		CurrencyID     uuid.UUID     `gorm:"type:uuid;not null"`
		Currency       *Currency     `gorm:"foreignKey:CurrencyID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"currency,omitempty"`

		Name              string  `gorm:"type:varchar(254)"`
		Description       string  `gorm:"type:text"`
		DeliquentAccount  bool    `gorm:"type:boolean;default:false"`
		FinesAccount      bool    `gorm:"type:boolean;default:false"`
		InterestAccountID bool    `gorm:"type:boolean;default:false"`
		ComakerAccount    float64 `gorm:"type:decimal;default:-1"`
		NumberOfMonths    int     `gorm:"type:int;default:0"`
		ExistAccount      bool    `gorm:"type:boolean;default:false"`
	}

	ComputationSheetResponse struct {
		ID                uuid.UUID             `json:"id"`
		CreatedAt         string                `json:"created_at"`
		CreatedByID       uuid.UUID             `json:"created_by_id"`
		CreatedBy         *UserResponse         `json:"created_by,omitempty"`
		UpdatedAt         string                `json:"updated_at"`
		UpdatedByID       uuid.UUID             `json:"updated_by_id"`
		UpdatedBy         *UserResponse         `json:"updated_by,omitempty"`
		OrganizationID    uuid.UUID             `json:"organization_id"`
		Organization      *OrganizationResponse `json:"organization,omitempty"`
		BranchID          uuid.UUID             `json:"branch_id"`
		Branch            *BranchResponse       `json:"branch,omitempty"`
		CurrencyID        uuid.UUID             `json:"currency_id"`
		Currency          *CurrencyResponse     `json:"currency,omitempty"`
		Name              string                `json:"name"`
		Description       string                `json:"description"`
		DeliquentAccount  bool                  `json:"deliquent_account"`
		FinesAccount      bool                  `json:"fines_account"`
		InterestAccountID bool                  `json:"interest_account_id"`
		ComakerAccount    float64               `json:"comaker_account"`
		ExistAccount      bool                  `json:"exist_account"`
	}

	ComputationSheetRequest struct {
		Name              string    `json:"name" validate:"required,min=1,max=254"`
		Description       string    `json:"description,omitempty"`
		CurrencyID        uuid.UUID `json:"currency_id" validate:"required"`
		DeliquentAccount  bool      `json:"deliquent_account,omitempty"`
		FinesAccount      bool      `json:"fines_account,omitempty"`
		InterestAccountID bool      `json:"interest_account_id,omitempty"`
		ComakerAccount    float64   `json:"comaker_account,omitempty"`
		ExistAccount      bool      `json:"exist_account,omitempty"`
	}

	ComputationSheetAmortizationResponse struct {
		Entries      []*LoanTransactionEntryResponse `json:"entries"`
		TotalDebit   float64                         `json:"total_debit"`
		TotalCredit  float64                         `json:"total_credit"`
		Amortization struct {
			Amortizations       []AmortizationPayment `json:"amortizations"`
			AmortizationSummary AmortizationSummary   `json:"amortization_summary"`
		} `json:"amortization"`
	}
)

func (m *ModelCore) computationSheet() {
	m.Migration = append(m.Migration, &ComputationSheet{})
	m.ComputationSheetManager = services.NewRepository(services.RepositoryParams[
		ComputationSheet, ComputationSheetResponse, ComputationSheetRequest,
	]{
		Preloads: []string{
			"CreatedBy", "UpdatedBy", "Currency",
		},
		Service: m.provider.Service,
		Resource: func(data *ComputationSheet) *ComputationSheetResponse {
			if data == nil {
				return nil
			}
			return &ComputationSheetResponse{
				ID:                data.ID,
				CreatedAt:         data.CreatedAt.Format(time.RFC3339),
				CreatedByID:       data.CreatedByID,
				CreatedBy:         m.UserManager.ToModel(data.CreatedBy),
				UpdatedAt:         data.UpdatedAt.Format(time.RFC3339),
				UpdatedByID:       data.UpdatedByID,
				UpdatedBy:         m.UserManager.ToModel(data.UpdatedBy),
				OrganizationID:    data.OrganizationID,
				Organization:      m.OrganizationManager.ToModel(data.Organization),
				BranchID:          data.BranchID,
				Branch:            m.BranchManager.ToModel(data.Branch),
				Name:              data.Name,
				Description:       data.Description,
				DeliquentAccount:  data.DeliquentAccount,
				FinesAccount:      data.FinesAccount,
				InterestAccountID: data.InterestAccountID,
				ComakerAccount:    data.ComakerAccount,
				ExistAccount:      data.ExistAccount,
				CurrencyID:        data.CurrencyID,
				Currency:          m.CurrencyManager.ToModel(data.Currency),
			}
		},
		Created: func(data *ComputationSheet) []string {
			return []string{
				"computation_sheet.create",
				fmt.Sprintf("computation_sheet.create.%s", data.ID),
				fmt.Sprintf("computation_sheet.create.branch.%s", data.BranchID),
				fmt.Sprintf("computation_sheet.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *ComputationSheet) []string {
			return []string{
				"computation_sheet.update",
				fmt.Sprintf("computation_sheet.update.%s", data.ID),
				fmt.Sprintf("computation_sheet.update.branch.%s", data.BranchID),
				fmt.Sprintf("computation_sheet.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *ComputationSheet) []string {
			return []string{
				"computation_sheet.delete",
				fmt.Sprintf("computation_sheet.delete.%s", data.ID),
				fmt.Sprintf("computation_sheet.delete.branch.%s", data.BranchID),
				fmt.Sprintf("computation_sheet.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

// ComputationSheetCurrentBranch returns all computation sheets for the given organization and branch.
func (m *ModelCore) ComputationSheetCurrentBranch(context context.Context, orgID uuid.UUID, branchID uuid.UUID) ([]*ComputationSheet, error) {
	return m.ComputationSheetManager.Find(context, &ComputationSheet{
		OrganizationID: orgID,
		BranchID:       branchID,
	})
}
