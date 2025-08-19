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
	LoanStatus struct {
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

		OrganizationID uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_loan_status"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_loan_status"`
		Branch         *Branch       `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"branch,omitempty"`

		Name        string `gorm:"type:varchar(255);not null"`
		Icon        string `gorm:"type:varchar(255)"`
		Color       string `gorm:"type:varchar(255)"`
		Description string `gorm:"type:text"`
	}

	LoanStatusResponse struct {
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
		Name           string                `json:"name"`
		Icon           string                `json:"icon"`
		Color          string                `json:"color"`
		Description    string                `json:"description"`
	}

	LoanStatusRequest struct {
		Name        string `json:"name" validate:"required,min=1,max=255"`
		Icon        string `json:"icon,omitempty"`
		Color       string `json:"color,omitempty"`
		Description string `json:"description,omitempty"`
	}
)

func (m *Model) LoanStatus() {
	m.Migration = append(m.Migration, &LoanStatus{})
	m.LoanStatusManager = horizon_services.NewRepository(horizon_services.RepositoryParams[
		LoanStatus, LoanStatusResponse, LoanStatusRequest,
	]{
		Preloads: []string{
			"CreatedBy", "UpdatedBy", "Branch", "Organization",
		},
		Service: m.provider.Service,
		Resource: func(data *LoanStatus) *LoanStatusResponse {
			if data == nil {
				return nil
			}
			return &LoanStatusResponse{
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
				Name:           data.Name,
				Icon:           data.Icon,
				Color:          data.Color,
				Description:    data.Description,
			}
		},

		Created: func(data *LoanStatus) []string {
			return []string{
				"loan_status.create",
				fmt.Sprintf("loan_status.create.%s", data.ID),
				fmt.Sprintf("loan_status.create.branch.%s", data.BranchID),
				fmt.Sprintf("loan_status.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *LoanStatus) []string {
			return []string{
				"loan_status.update",
				fmt.Sprintf("loan_status.update.%s", data.ID),
				fmt.Sprintf("loan_status.update.branch.%s", data.BranchID),
				fmt.Sprintf("loan_status.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *LoanStatus) []string {
			return []string{
				"loan_status.delete",
				fmt.Sprintf("loan_status.delete.%s", data.ID),
				fmt.Sprintf("loan_status.delete.branch.%s", data.BranchID),
				fmt.Sprintf("loan_status.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func (m *Model) LoanStatusCurrentBranch(context context.Context, orgId uuid.UUID, branchId uuid.UUID) ([]*LoanStatus, error) {
	return m.LoanStatusManager.Find(context, &LoanStatus{
		OrganizationID: orgId,
		BranchID:       branchId,
	})
}
