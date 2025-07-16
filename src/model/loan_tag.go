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
	LoanTag struct {
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

		OrganizationID uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_loan_tag"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_loan_tag"`
		Branch         *Branch       `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"branch,omitempty"`

		LoanTransactionID uuid.UUID        `gorm:"type:uuid;not null"`
		LoanTransaction   *LoanTransaction `gorm:"foreignKey:LoanTransactionID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"loan_transaction,omitempty"`

		Name        string `gorm:"type:varchar(50);not null"`
		Description string `gorm:"type:text"`
		Category    string `gorm:"type:varchar(50)"`
		Color       string `gorm:"type:varchar(20)"`
		Icon        string `gorm:"type:varchar(20)"`
	}

	LoanTagResponse struct {
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
		Category          TagCategory              `json:"category"`
		Color             string                   `json:"color"`
		Icon              string                   `json:"icon"`
	}

	LoanTagRequest struct {
		LoanTransactionID uuid.UUID   `json:"loan_transaction_id" validate:"required"`
		Name              string      `json:"name" validate:"required,min=1,max=50"`
		Description       string      `json:"description,omitempty"`
		Category          TagCategory `json:"category,omitempty"`
		Color             string      `json:"color,omitempty"`
		Icon              string      `json:"icon,omitempty"`
	}
)

func (m *Model) LoanTag() {
	m.Migration = append(m.Migration, &LoanTag{})
	m.LoanTagManager = horizon_services.NewRepository(horizon_services.RepositoryParams[
		LoanTag, LoanTagResponse, LoanTagRequest,
	]{
		Preloads: []string{
			"CreatedBy", "UpdatedBy", "Branch", "Organization", "LoanTransaction",
		},
		Service: m.provider.Service,
		Resource: func(data *LoanTag) *LoanTagResponse {
			if data == nil {
				return nil
			}
			return &LoanTagResponse{
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
				Category:          TagCategory(data.Category),
				Color:             data.Color,
				Icon:              data.Icon,
			}
		},

		Created: func(data *LoanTag) []string {
			return []string{
				"loan_tag.create",
				fmt.Sprintf("loan_tag.create.%s", data.ID),
				fmt.Sprintf("loan_tag.create.branch.%s", data.BranchID),
				fmt.Sprintf("loan_tag.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *LoanTag) []string {
			return []string{
				"loan_tag.update",
				fmt.Sprintf("loan_tag.update.%s", data.ID),
				fmt.Sprintf("loan_tag.update.branch.%s", data.BranchID),
				fmt.Sprintf("loan_tag.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *LoanTag) []string {
			return []string{
				"loan_tag.delete",
				fmt.Sprintf("loan_tag.delete.%s", data.ID),
				fmt.Sprintf("loan_tag.delete.branch.%s", data.BranchID),
				fmt.Sprintf("loan_tag.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func (m *Model) LoanTagCurrentBranch(context context.Context, orgId uuid.UUID, branchId uuid.UUID) ([]*LoanTag, error) {
	return m.LoanTagManager.Find(context, &LoanTag{
		OrganizationID: orgId,
		BranchID:       branchId,
	})
}
