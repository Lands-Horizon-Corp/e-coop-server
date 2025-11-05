package core

import (
	"context"
	"fmt"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/services/registry"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type (
	// LoanClearanceAnalysisInstitution represents the LoanClearanceAnalysisInstitution model.
	LoanClearanceAnalysisInstitution struct {
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

		OrganizationID uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_loan_clearance_analysis_institution"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_loan_clearance_analysis_institution"`
		Branch         *Branch       `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"branch,omitempty"`

		LoanTransactionID uuid.UUID        `gorm:"type:uuid;not null"`
		LoanTransaction   *LoanTransaction `gorm:"foreignKey:LoanTransactionID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"loan_transaction,omitempty"`

		Name        string `gorm:"type:varchar(50)"`
		Description string `gorm:"type:text"`
	}

	// LoanClearanceAnalysisInstitutionResponse represents the response structure for loanclearanceanalysisinstitution data

	// LoanClearanceAnalysisInstitutionResponse represents the response structure for LoanClearanceAnalysisInstitution.
	LoanClearanceAnalysisInstitutionResponse struct {
		ID                uuid.UUID                `json:"id"`
		CreatedAt         string                   `json:"created_at"`
		CreatedByID       uuid.UUID                `json:"created_by_id"`
		CreatedBy         *UserResponse            `json:"created_by,omitempty"`
		UpdatedAt         string                   `json:"updated_at"`
		UpdatedByID       uuid.UUID                `json:"updated_by_id"`
		UpdatedBy         *UserResponse            `json:"updated_by,omitempty"`
		OrganizationID    uuid.UUID                `json:"organization_id"`
		Organization      *OrganizationResponse    `json:"organization,omitempty"`
		BranchID          uuid.UUID                `json:"branch_id"`
		Branch            *BranchResponse          `json:"branch,omitempty"`
		LoanTransactionID uuid.UUID                `json:"loan_transaction_id"`
		LoanTransaction   *LoanTransactionResponse `json:"loan_transaction,omitempty"`
		Name              string                   `json:"name"`
		Description       string                   `json:"description"`
	}

	// LoanClearanceAnalysisInstitutionRequest represents the request structure for creating/updating loanclearanceanalysisinstitution

	// LoanClearanceAnalysisInstitutionRequest represents the request structure for LoanClearanceAnalysisInstitution.
	LoanClearanceAnalysisInstitutionRequest struct {
		ID                *uuid.UUID `json:"id"`
		LoanTransactionID uuid.UUID  `json:"loan_transaction_id"`
		Name              string     `json:"name" validate:"required,min=1,max=50"`
		Description       string     `json:"description,omitempty"`
	}
)

func (m *Core) loanClearanceAnalysisInstitution() {
	m.Migration = append(m.Migration, &LoanClearanceAnalysisInstitution{})
	m.LoanClearanceAnalysisInstitutionManager = *registry.NewRegistry(registry.RegistryParams[
		LoanClearanceAnalysisInstitution, LoanClearanceAnalysisInstitutionResponse, LoanClearanceAnalysisInstitutionRequest,
	]{
		Preloads: []string{
			"CreatedBy", "UpdatedBy", "LoanTransaction",
		},
		Service: m.provider.Service,
		Resource: func(data *LoanClearanceAnalysisInstitution) *LoanClearanceAnalysisInstitutionResponse {
			if data == nil {
				return nil
			}
			return &LoanClearanceAnalysisInstitutionResponse{
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
				LoanTransactionID: data.LoanTransactionID,
				LoanTransaction:   m.LoanTransactionManager.ToModel(data.LoanTransaction),
				Name:              data.Name,
				Description:       data.Description,
			}
		},

		Created: func(data *LoanClearanceAnalysisInstitution) []string {
			return []string{
				"loan_clearance_analysis_institution.create",
				fmt.Sprintf("loan_clearance_analysis_institution.create.%s", data.ID),
				fmt.Sprintf("loan_clearance_analysis_institution.create.branch.%s", data.BranchID),
				fmt.Sprintf("loan_clearance_analysis_institution.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *LoanClearanceAnalysisInstitution) []string {
			return []string{
				"loan_clearance_analysis_institution.update",
				fmt.Sprintf("loan_clearance_analysis_institution.update.%s", data.ID),
				fmt.Sprintf("loan_clearance_analysis_institution.update.branch.%s", data.BranchID),
				fmt.Sprintf("loan_clearance_analysis_institution.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *LoanClearanceAnalysisInstitution) []string {
			return []string{
				"loan_clearance_analysis_institution.delete",
				fmt.Sprintf("loan_clearance_analysis_institution.delete.%s", data.ID),
				fmt.Sprintf("loan_clearance_analysis_institution.delete.branch.%s", data.BranchID),
				fmt.Sprintf("loan_clearance_analysis_institution.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

// LoanClearanceAnalysisInstitutionCurrentBranch returns LoanClearanceAnalysisInstitutionCurrentBranch for the current branch or organization where applicable.
func (m *Core) LoanClearanceAnalysisInstitutionCurrentBranch(context context.Context, organizationID uuid.UUID, branchID uuid.UUID) ([]*LoanClearanceAnalysisInstitution, error) {
	return m.LoanClearanceAnalysisInstitutionManager.Find(context, &LoanClearanceAnalysisInstitution{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}
