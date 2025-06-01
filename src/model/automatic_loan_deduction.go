package model

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	horizon_services "github.com/lands-horizon/horizon-server/services"
	"gorm.io/gorm"
)

type (
	AutomaticLoanDeduction struct {
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

		OrganizationID uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_automatic_loan_deduction"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_automatic_loan_deduction"`
		Branch         *Branch       `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"branch,omitempty"`

		AccountID          *uuid.UUID        `gorm:"type:uuid"`
		Account            *Account          `gorm:"foreignKey:AccountID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"account,omitempty"`
		ComputationSheetID *uuid.UUID        `gorm:"type:uuid"`
		ComputationSheet   *ComputationSheet `gorm:"foreignKey:ComputationSheetID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"computation_sheet,omitempty"`
		LinkAccountID      *uuid.UUID        `gorm:"type:uuid"`
		LinkAccount        *Account          `gorm:"foreignKey:LinkAccountID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"link_account,omitempty"`

		ChargesPercentage1 float64 `gorm:"type:decimal"`
		ChargesPercentage2 float64 `gorm:"type:decimal"`
		ChargesAmount      float64 `gorm:"type:decimal"`
		ChargesDivisor     float64 `gorm:"type:decimal"`

		MinAmount float64 `gorm:"type:decimal"`
		MaxAmount float64 `gorm:"type:decimal"`

		Anum int `gorm:"type:int"` // months

		AddOn          bool `gorm:"type:boolean;default:false"`
		AoRest         bool `gorm:"type:boolean;default:false"`
		ExcludeRenewal bool `gorm:"type:boolean;default:false"`
		Ct             int  `gorm:"type:int"`

		Name        string `gorm:"type:varchar(255)"`
		Description string `gorm:"type:text"`
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

		AccountID          *uuid.UUID                `json:"account_id,omitempty"`
		Account            *AccountResponse          `json:"account,omitempty"`
		ComputationSheetID *uuid.UUID                `json:"computation_sheet_id,omitempty"`
		ComputationSheet   *ComputationSheetResponse `json:"computation_sheet,omitempty"`
		LinkAccountID      *uuid.UUID                `json:"link_account_id,omitempty"`
		LinkAccount        *AccountResponse          `json:"link_account,omitempty"`

		ChargesPercentage1 float64 `json:"charges_percentage_1"`
		ChargesPercentage2 float64 `json:"charges_percentage_2"`
		ChargesAmount      float64 `json:"charges_amount"`
		ChargesDivisor     float64 `json:"charges_divisor"`

		MinAmount float64 `json:"min_amount"`
		MaxAmount float64 `json:"max_amount"`

		Anum int `json:"anum"`

		AddOn          bool `json:"add_on"`
		AoRest         bool `json:"ao_rest"`
		ExcludeRenewal bool `json:"exclude_renewal"`
		Ct             int  `json:"ct"`

		Name        string `json:"name"`
		Description string `json:"description"`
	}

	AutomaticLoanDeductionRequest struct {
		AccountID          *uuid.UUID `json:"account_id,omitempty"`
		ComputationSheetID *uuid.UUID `json:"computation_sheet_id,omitempty"`
		LinkAccountID      *uuid.UUID `json:"link_account_id,omitempty"`
		ChargesPercentage1 float64    `json:"charges_percentage_1,omitempty"`
		ChargesPercentage2 float64    `json:"charges_percentage_2,omitempty"`
		ChargesAmount      float64    `json:"charges_amount,omitempty"`
		ChargesDivisor     float64    `json:"charges_divisor,omitempty"`
		MinAmount          float64    `json:"min_amount,omitempty"`
		MaxAmount          float64    `json:"max_amount,omitempty"`
		Anum               int        `json:"anum,omitempty"`
		AddOn              bool       `json:"add_on,omitempty"`
		AoRest             bool       `json:"ao_rest,omitempty"`
		ExcludeRenewal     bool       `json:"exclude_renewal,omitempty"`
		Ct                 int        `json:"ct,omitempty"`
		Name               string     `json:"name" validate:"required,min=1,max=255"`
		Description        string     `json:"description,omitempty"`
	}
)

func (m *Model) AutomaticLoanDeduction() {
	m.Migration = append(m.Migration, &AutomaticLoanDeduction{})
	m.AutomaticLoanDeductionManager = horizon_services.NewRepository(horizon_services.RepositoryParams[
		AutomaticLoanDeduction, AutomaticLoanDeductionResponse, AutomaticLoanDeductionRequest,
	]{
		Preloads: []string{
			"CreatedBy", "UpdatedBy", "DeletedBy", "Branch", "Organization",
			"Account", "ComputationSheet", "LinkAccount",
		},
		Service: m.provider.Service,
		Resource: func(data *AutomaticLoanDeduction) *AutomaticLoanDeductionResponse {
			if data == nil {
				return nil
			}
			return &AutomaticLoanDeductionResponse{
				ID:                 data.ID,
				CreatedAt:          data.CreatedAt.Format(time.RFC3339),
				CreatedByID:        data.CreatedByID,
				CreatedBy:          m.UserManager.ToModel(data.CreatedBy),
				UpdatedAt:          data.UpdatedAt.Format(time.RFC3339),
				UpdatedByID:        data.UpdatedByID,
				UpdatedBy:          m.UserManager.ToModel(data.UpdatedBy),
				OrganizationID:     data.OrganizationID,
				Organization:       m.OrganizationManager.ToModel(data.Organization),
				BranchID:           data.BranchID,
				Branch:             m.BranchManager.ToModel(data.Branch),
				AccountID:          data.AccountID,
				Account:            m.AccountManager.ToModel(data.Account),
				ComputationSheetID: data.ComputationSheetID,
				ComputationSheet:   m.ComputationSheetManager.ToModel(data.ComputationSheet),
				LinkAccountID:      data.LinkAccountID,
				LinkAccount:        m.AccountManager.ToModel(data.LinkAccount),
				ChargesPercentage1: data.ChargesPercentage1,
				ChargesPercentage2: data.ChargesPercentage2,
				ChargesAmount:      data.ChargesAmount,
				ChargesDivisor:     data.ChargesDivisor,
				MinAmount:          data.MinAmount,
				MaxAmount:          data.MaxAmount,
				Anum:               data.Anum,
				AddOn:              data.AddOn,
				AoRest:             data.AoRest,
				ExcludeRenewal:     data.ExcludeRenewal,
				Ct:                 data.Ct,
				Name:               data.Name,
				Description:        data.Description,
			}
		},
		Created: func(data *AutomaticLoanDeduction) []string {
			return []string{
				"automatic_loan_deduction.create",
				fmt.Sprintf("automatic_loan_deduction.create.%s", data.ID),
				fmt.Sprintf("automatic_loan_deduction.create.branch.%s", data.BranchID),
				fmt.Sprintf("automatic_loan_deduction.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *AutomaticLoanDeduction) []string {
			return []string{
				"automatic_loan_deduction.update",
				fmt.Sprintf("automatic_loan_deduction.update.%s", data.ID),
				fmt.Sprintf("automatic_loan_deduction.update.branch.%s", data.BranchID),
				fmt.Sprintf("automatic_loan_deduction.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *AutomaticLoanDeduction) []string {
			return []string{
				"automatic_loan_deduction.update",
				fmt.Sprintf("automatic_loan_deduction.delete.%s", data.ID),
				fmt.Sprintf("automatic_loan_deduction.delete.branch.%s", data.BranchID),
				fmt.Sprintf("automatic_loan_deduction.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func (m *Model) AutomaticLoanDeductionCurrentBranch(context context.Context, orgId uuid.UUID, branchId uuid.UUID) ([]*AutomaticLoanDeduction, error) {
	return m.AutomaticLoanDeductionManager.Find(context, &AutomaticLoanDeduction{
		OrganizationID: orgId,
		BranchID:       branchId,
	})
}
