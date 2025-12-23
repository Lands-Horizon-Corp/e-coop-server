package core

import (
	"context"
	"fmt"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/pkg/registry"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type (
	GroceryComputationSheetMonthly struct {
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

		OrganizationID uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_grocery_computation_sheet_monthly"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_grocery_computation_sheet_monthly"`
		Branch         *Branch       `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"branch,omitempty"`

		GroceryComputationSheetID uuid.UUID                `gorm:"type:uuid;not null"`
		GroceryComputationSheet   *GroceryComputationSheet `gorm:"foreignKey:GroceryComputationSheetID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"grocery_computation_sheet,omitempty"`

		Months                 int     `gorm:"default:0"`
		InterestRate           float64 `gorm:"type:decimal;default:0"`
		LoanGuaranteedFundRate float64 `gorm:"type:decimal;default:0"`
	}

	GroceryComputationSheetMonthlyResponse struct {
		ID                        uuid.UUID                        `json:"id"`
		CreatedAt                 string                           `json:"created_at"`
		CreatedByID               uuid.UUID                        `json:"created_by_id"`
		CreatedBy                 *UserResponse                    `json:"created_by,omitempty"`
		UpdatedAt                 string                           `json:"updated_at"`
		UpdatedByID               uuid.UUID                        `json:"updated_by_id"`
		UpdatedBy                 *UserResponse                    `json:"updated_by,omitempty"`
		OrganizationID            uuid.UUID                        `json:"organization_id"`
		Organization              *OrganizationResponse            `json:"organization,omitempty"`
		BranchID                  uuid.UUID                        `json:"branch_id"`
		Branch                    *BranchResponse                  `json:"branch,omitempty"`
		GroceryComputationSheetID uuid.UUID                        `json:"grocery_computation_sheet_id"`
		GroceryComputationSheet   *GroceryComputationSheetResponse `json:"grocery_computation_sheet,omitempty"`
		Months                    int                              `json:"months"`
		InterestRate              float64                          `json:"interest_rate"`
		LoanGuaranteedFundRate    float64                          `json:"loan_guaranteed_fund_rate"`
	}

	GroceryComputationSheetMonthlyRequest struct {
		GroceryComputationSheetID uuid.UUID `json:"grocery_computation_sheet_id" validate:"required"`
		Months                    int       `json:"months,omitempty"`
		InterestRate              float64   `json:"interest_rate,omitempty"`
		LoanGuaranteedFundRate    float64   `json:"loan_guaranteed_fund_rate,omitempty"`
	}
)

func (m *Core) GroceryComputationSheetMonthlyManager() *registry.Registry[GroceryComputationSheetMonthly, GroceryComputationSheetMonthlyResponse, GroceryComputationSheetMonthlyRequest] {
	return registry.NewRegistry(registry.RegistryParams[
		GroceryComputationSheetMonthly, GroceryComputationSheetMonthlyResponse, GroceryComputationSheetMonthlyRequest,
	]{
		Preloads: []string{
			"CreatedBy", "UpdatedBy", "GroceryComputationSheet",
		},
		Database: m.provider.Service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return m.provider.Service.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *GroceryComputationSheetMonthly) *GroceryComputationSheetMonthlyResponse {
			if data == nil {
				return nil
			}
			return &GroceryComputationSheetMonthlyResponse{
				ID:                        data.ID,
				CreatedAt:                 data.CreatedAt.Format(time.RFC3339),
				CreatedByID:               data.CreatedByID,
				CreatedBy:                 m.UserManager().ToModel(data.CreatedBy),
				UpdatedAt:                 data.UpdatedAt.Format(time.RFC3339),
				UpdatedByID:               data.UpdatedByID,
				UpdatedBy:                 m.UserManager().ToModel(data.UpdatedBy),
				OrganizationID:            data.OrganizationID,
				Organization:              m.OrganizationManager().ToModel(data.Organization),
				BranchID:                  data.BranchID,
				Branch:                    m.BranchManager().ToModel(data.Branch),
				GroceryComputationSheetID: data.GroceryComputationSheetID,
				GroceryComputationSheet:   m.GroceryComputationSheetManager().ToModel(data.GroceryComputationSheet),
				Months:                    data.Months,
				InterestRate:              data.InterestRate,
				LoanGuaranteedFundRate:    data.LoanGuaranteedFundRate,
			}
		},
		Created: func(data *GroceryComputationSheetMonthly) registry.Topics {
			return []string{
				"grocery_computation_sheet_monthly.create",
				fmt.Sprintf("grocery_computation_sheet_monthly.create.%s", data.ID),
				fmt.Sprintf("grocery_computation_sheet_monthly.create.branch.%s", data.BranchID),
				fmt.Sprintf("grocery_computation_sheet_monthly.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *GroceryComputationSheetMonthly) registry.Topics {
			return []string{
				"grocery_computation_sheet_monthly.update",
				fmt.Sprintf("grocery_computation_sheet_monthly.update.%s", data.ID),
				fmt.Sprintf("grocery_computation_sheet_monthly.update.branch.%s", data.BranchID),
				fmt.Sprintf("grocery_computation_sheet_monthly.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *GroceryComputationSheetMonthly) registry.Topics {
			return []string{
				"grocery_computation_sheet_monthly.delete",
				fmt.Sprintf("grocery_computation_sheet_monthly.delete.%s", data.ID),
				fmt.Sprintf("grocery_computation_sheet_monthly.delete.branch.%s", data.BranchID),
				fmt.Sprintf("grocery_computation_sheet_monthly.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func (m *Core) GroceryComputationSheetMonthlyCurrentBranch(context context.Context, organizationID uuid.UUID, branchID uuid.UUID) ([]*GroceryComputationSheetMonthly, error) {
	return m.GroceryComputationSheetMonthlyManager().Find(context, &GroceryComputationSheetMonthly{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}
