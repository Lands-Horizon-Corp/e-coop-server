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
	LoanLedger struct {
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

		OrganizationID uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_loan_ledger"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_loan_ledger"`
		Branch         *Branch       `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"branch,omitempty"`
	}

	LoanLedgerResponse struct {
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
	}

	LoanLedgerRequest struct{}
)

func (m *Model) LoanLedger() {
	m.Migration = append(m.Migration, &LoanLedger{})
	m.LoanLedgerManager = horizon_services.NewRepository(horizon_services.RepositoryParams[
		LoanLedger, LoanLedgerResponse, LoanLedgerRequest,
	]{
		Preloads: []string{
			"CreatedBy", "UpdatedBy", "DeletedBy", "Branch", "Organization",
		},
		Service: m.provider.Service,
		Resource: func(data *LoanLedger) *LoanLedgerResponse {
			if data == nil {
				return nil
			}
			return &LoanLedgerResponse{
				ID:             data.ID,
				CreatedAt:      data.CreatedAt.Format(time.RFC3339),
				CreatedByID:    data.CreatedByID,
				CreatedBy:      m.UserManager.ToModel(data.CreatedBy),
				UpdatedAt:      data.UpdatedAt.Format(time.RFC3339),
				UpdatedByID:    data.UpdatedByID,
				UpdatedBy:      m.UserManager.ToModel(data.UpdatedBy),
				OrganizationID: data.OrganizationID,
				Organization:   m.OrganizationManager.ToModel(data.Organization),
				BranchID:       data.BranchID,
				Branch:         m.BranchManager.ToModel(data.Branch),
			}
		},

		Created: func(data *LoanLedger) []string {
			return []string{
				"loan_ledger.create",
				fmt.Sprintf("loan_ledger.create.%s", data.ID),
				fmt.Sprintf("loan_ledger.create.branch.%s", data.BranchID),
				fmt.Sprintf("loan_ledger.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *LoanLedger) []string {
			return []string{
				"loan_ledger.update",
				fmt.Sprintf("loan_ledger.update.%s", data.ID),
				fmt.Sprintf("loan_ledger.update.branch.%s", data.BranchID),
				fmt.Sprintf("loan_ledger.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *LoanLedger) []string {
			return []string{
				"loan_ledger.delete",
				fmt.Sprintf("loan_ledger.delete.%s", data.ID),
				fmt.Sprintf("loan_ledger.delete.branch.%s", data.BranchID),
				fmt.Sprintf("loan_ledger.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func (m *Model) LoanLedgerCurrentBranch(context context.Context, orgId uuid.UUID, branchId uuid.UUID) ([]*LoanLedger, error) {
	return m.LoanLedgerManager.Find(context, &LoanLedger{
		OrganizationID: orgId,
		BranchID:       branchId,
	})
}
