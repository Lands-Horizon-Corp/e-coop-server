package model

import (
	"context"
	"fmt"
	"time"

	horizon_services "github.com/Lands-Horizon-Corp/e-coop-server/services"
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
		Name              string                `json:"name"`
		Description       string                `json:"description"`
		DeliquentAccount  bool                  `json:"deliquent_account"`
		FinesAccount      bool                  `json:"fines_account"`
		InterestAccountID bool                  `json:"interest_account_id"`
		ComakerAccount    float64               `json:"comaker_account"`
		ExistAccount      bool                  `json:"exist_account"`
	}

	ComputationSheetRequest struct {
		Name              string  `json:"name" validate:"required,min=1,max=254"`
		Description       string  `json:"description,omitempty"`
		DeliquentAccount  bool    `json:"deliquent_account,omitempty"`
		FinesAccount      bool    `json:"fines_account,omitempty"`
		InterestAccountID bool    `json:"interest_account_id,omitempty"`
		ComakerAccount    float64 `json:"comaker_account,omitempty"`
		ExistAccount      bool    `json:"exist_account,omitempty"`
	}

	ComputationSheetAmortizationRequest struct {
		Terms              int               `json:"terms" validate:"required,min=1"`
		AccountID          uuid.UUID         `json:"account_id" validate:"required"`
		ModeOfPayment      LoanModeOfPayment `json:"mode_of_payment" validate:"required"`
		FixedDays          *int              `json:"fixed_days,omitempty"`
		Weekdays           *int              `json:"weekdays,omitempty"`
		Pay1               *int              `json:"pay_1,omitempty"`
		Pay2               *int              `json:"pay_2,omitempty"`
		Applied1           float64           `json:"applied_1" validate:"required"`
		IsAddOn            bool              `json:"is_add_on"`
		ExcludeHolidays    *bool             `json:"exclude_holidays,omitempty"`
		ExcludeSaturdays   *bool             `json:"exclude_saturdays,omitempty"`
		ExcludeSundays     *bool             `json:"exclude_sundays,omitempty"`
		ComputationSheetID uuid.UUID         `json:"computation_sheet_id" validate:"required"`
	}

	ComputationSheetAmortizationEntry struct {
		Account *AccountResponse         `json:"account,omitempty"`
		IsAddOn bool                     `json:"is_add_on"`
		Type    LoanTransactionEntryType `json:"type"`
		Credit  float64                  `json:"credit"`
		Debit   float64                  `json:"debit"`
		Name    string                   `json:"name,omitempty"`
	}

	ComputationSheetAmortizationResponse struct {
		Entries      []ComputationSheetAmortizationEntry `json:"entries"`
		Amortization struct {
			Amortizations       []AmortizationPayment `json:"amortizations"`
			AmortizationSummary AmortizationSummary   `json:"amortization_summary"`
		} `json:"amortization"`
	}
)

func (m *Model) ComputationSheet() {
	m.Migration = append(m.Migration, &ComputationSheet{})
	m.ComputationSheetManager = horizon_services.NewRepository(horizon_services.RepositoryParams[
		ComputationSheet, ComputationSheetResponse, ComputationSheetRequest,
	]{
		Preloads: []string{
			"CreatedBy", "UpdatedBy", "Branch", "Organization",
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

func (m *Model) ComputationSheetCurrentBranch(context context.Context, orgId uuid.UUID, branchId uuid.UUID) ([]*ComputationSheet, error) {
	return m.ComputationSheetManager.Find(context, &ComputationSheet{
		OrganizationID: orgId,
		BranchID:       branchId,
	})
}
