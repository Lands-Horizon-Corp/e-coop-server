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
	LoanPurpose struct {
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

		OrganizationID uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_loan_purpose"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_loan_purpose"`
		Branch         *Branch       `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"branch,omitempty"`

		Description string `gorm:"type:text"`
		Icon        string `gorm:"type:varchar(255)"`
	}

	LoanPurposeResponse struct {
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
		Description    string                `json:"description"`
		Icon           string                `json:"icon"`
	}

	LoanPurposeRequest struct {
		Description string `json:"description,omitempty"`
		Icon        string `json:"icon,omitempty"`
	}
)

func (m *Model) LoanPurpose() {
	m.Migration = append(m.Migration, &LoanPurpose{})
	m.LoanPurposeManager = horizon_services.NewRepository(horizon_services.RepositoryParams[
		LoanPurpose, LoanPurposeResponse, LoanPurposeRequest,
	]{
		Preloads: []string{
			"CreatedBy", "UpdatedBy", "DeletedBy", "Branch", "Organization",
		},
		Service: m.provider.Service,
		Resource: func(data *LoanPurpose) *LoanPurposeResponse {
			if data == nil {
				return nil
			}
			return &LoanPurposeResponse{
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
				Description:    data.Description,
				Icon:           data.Icon,
			}
		},

		Created: func(data *LoanPurpose) []string {
			return []string{
				"loan_purpose.create",
				fmt.Sprintf("loan_purpose.create.%s", data.ID),
				fmt.Sprintf("loan_purpose.create.branch.%s", data.BranchID),
				fmt.Sprintf("loan_purpose.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *LoanPurpose) []string {
			return []string{
				"loan_purpose.update",
				fmt.Sprintf("loan_purpose.update.%s", data.ID),
				fmt.Sprintf("loan_purpose.update.branch.%s", data.BranchID),
				fmt.Sprintf("loan_purpose.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *LoanPurpose) []string {
			return []string{
				"loan_purpose.delete",
				fmt.Sprintf("loan_purpose.delete.%s", data.ID),
				fmt.Sprintf("loan_purpose.delete.branch.%s", data.BranchID),
				fmt.Sprintf("loan_purpose.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func (m *Model) LoanPurposeCurrentBranch(context context.Context, orgId uuid.UUID, branchId uuid.UUID) ([]*LoanPurpose, error) {
	return m.LoanPurposeManager.Find(context, &LoanPurpose{
		OrganizationID: orgId,
		BranchID:       branchId,
	})
}
