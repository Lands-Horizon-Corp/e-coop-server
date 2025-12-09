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
	// LoanClearanceAnalysis represents the LoanClearanceAnalysis model.
	LoanClearanceAnalysis struct {
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

		OrganizationID uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_loan_clearance_analysis"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_loan_clearance_analysis"`
		Branch         *Branch       `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"branch,omitempty"`

		LoanTransactionID uuid.UUID        `gorm:"type:uuid;not null"`
		LoanTransaction   *LoanTransaction `gorm:"foreignKey:LoanTransactionID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"loan_transaction,omitempty"`

		RegularDeductionDescription string  `gorm:"type:text"`
		RegularDeductionAmount      float64 `gorm:"type:decimal"`

		BalancesDescription string  `gorm:"type:text"`
		BalancesAmount      float64 `gorm:"type:decimal"`
		BalancesCount       int     `gorm:"type:int"`
	}

	// LoanClearanceAnalysisResponse represents the response structure for LoanClearanceAnalysis.
	LoanClearanceAnalysisResponse struct {
		ID                          uuid.UUID                `json:"id"`
		CreatedAt                   string                   `json:"created_at"`
		CreatedByID                 uuid.UUID                `json:"created_by_id"`
		CreatedBy                   *UserResponse            `json:"created_by,omitempty"`
		UpdatedAt                   string                   `json:"updated_at"`
		UpdatedByID                 uuid.UUID                `json:"updated_by_id"`
		UpdatedBy                   *UserResponse            `json:"updated_by,omitempty"`
		OrganizationID              uuid.UUID                `json:"organization_id"`
		Organization                *OrganizationResponse    `json:"organization,omitempty"`
		BranchID                    uuid.UUID                `json:"branch_id"`
		Branch                      *BranchResponse          `json:"branch,omitempty"`
		LoanTransactionID           uuid.UUID                `json:"loan_transaction_id"`
		LoanTransaction             *LoanTransactionResponse `json:"loan_transaction,omitempty"`
		RegularDeductionDescription string                   `json:"regular_deduction_description"`
		RegularDeductionAmount      float64                  `json:"regular_deduction_amount"`
		BalancesDescription         string                   `json:"balances_description"`
		BalancesAmount              float64                  `json:"balances_amount"`
		BalancesCount               int                      `json:"balances_count"`
	}

	// LoanClearanceAnalysisRequest represents the request structure for LoanClearanceAnalysis.
	LoanClearanceAnalysisRequest struct {
		ID                          *uuid.UUID `json:"id"`
		RegularDeductionDescription string     `json:"regular_deduction_description,omitempty"`
		RegularDeductionAmount      float64    `json:"regular_deduction_amount,omitempty"`
		BalancesDescription         string     `json:"balances_description,omitempty"`
		BalancesAmount              float64    `json:"balances_amount,omitempty"`
		BalancesCount               int        `json:"balances_count,omitempty"`
	}
)

func (m *Core) loanClearanceAnalysis() {
	m.Migration = append(m.Migration, &LoanClearanceAnalysis{})
	m.LoanClearanceAnalysisManager = *registry.NewRegistry(registry.RegistryParams[
		LoanClearanceAnalysis, LoanClearanceAnalysisResponse, LoanClearanceAnalysisRequest,
	]{
		Preloads: []string{
			"CreatedBy", "UpdatedBy", "LoanTransaction",
		},
		Service: m.provider.Service,
		Resource: func(data *LoanClearanceAnalysis) *LoanClearanceAnalysisResponse {
			if data == nil {
				return nil
			}
			return &LoanClearanceAnalysisResponse{
				ID:                          data.ID,
				CreatedAt:                   data.CreatedAt.Format(time.RFC3339),
				CreatedByID:                 data.CreatedByID,
				CreatedBy:                   m.UserManager.ToModel(data.CreatedBy),
				UpdatedAt:                   data.UpdatedAt.Format(time.RFC3339),
				UpdatedByID:                 data.UpdatedByID,
				UpdatedBy:                   m.UserManager.ToModel(data.UpdatedBy),
				OrganizationID:              data.OrganizationID,
				Organization:                m.OrganizationManager.ToModel(data.Organization),
				BranchID:                    data.BranchID,
				Branch:                      m.BranchManager.ToModel(data.Branch),
				LoanTransactionID:           data.LoanTransactionID,
				LoanTransaction:             m.LoanTransactionManager.ToModel(data.LoanTransaction),
				RegularDeductionDescription: data.RegularDeductionDescription,
				RegularDeductionAmount:      data.RegularDeductionAmount,
				BalancesDescription:         data.BalancesDescription,
				BalancesAmount:              data.BalancesAmount,
				BalancesCount:               data.BalancesCount,
			}
		},

		Created: func(data *LoanClearanceAnalysis) []string {
			return []string{
				"loan_clearance_analysis.create",
				fmt.Sprintf("loan_clearance_analysis.create.%s", data.ID),
				fmt.Sprintf("loan_clearance_analysis.create.branch.%s", data.BranchID),
				fmt.Sprintf("loan_clearance_analysis.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *LoanClearanceAnalysis) []string {
			return []string{
				"loan_clearance_analysis.update",
				fmt.Sprintf("loan_clearance_analysis.update.%s", data.ID),
				fmt.Sprintf("loan_clearance_analysis.update.branch.%s", data.BranchID),
				fmt.Sprintf("loan_clearance_analysis.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *LoanClearanceAnalysis) []string {
			return []string{
				"loan_clearance_analysis.delete",
				fmt.Sprintf("loan_clearance_analysis.delete.%s", data.ID),
				fmt.Sprintf("loan_clearance_analysis.delete.branch.%s", data.BranchID),
				fmt.Sprintf("loan_clearance_analysis.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

// LoanClearanceAnalysisCurrentBranch retrieves loan clearance analyses for the given organization and branch.
func (m *Core) LoanClearanceAnalysisCurrentBranch(context context.Context, organizationID uuid.UUID, branchID uuid.UUID) ([]*LoanClearanceAnalysis, error) {
	return m.LoanClearanceAnalysisManager.Find(context, &LoanClearanceAnalysis{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}
