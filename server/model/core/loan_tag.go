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
	// LoanTag represents an arbitrary tag that can be attached to a loan transaction.
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

		LoanTransactionID *uuid.UUID       `gorm:"type:uuid"`
		LoanTransaction   *LoanTransaction `gorm:"foreignKey:LoanTransactionID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"loan_transaction,omitempty"`

		Name        string      `gorm:"type:varchar(50);not null"`
		Description string      `gorm:"type:text"`
		Category    TagCategory `gorm:"type:varchar(50)"`
		Color       string      `gorm:"type:varchar(20)"`
		Icon        string      `gorm:"type:varchar(20)"`
	}

	// LoanTagResponse represents the response structure for loantag data

	// LoanTagResponse represents the response structure for LoanTag.
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
		LoanTransactionID *uuid.UUID               `json:"loan_transaction_id,omitempty"`
		LoanTransaction   *LoanTransactionResponse `json:"loan_transaction,omitempty"`
		Name              string                   `json:"name"`
		Description       string                   `json:"description"`
		Category          TagCategory              `json:"category"`
		Color             string                   `json:"color"`
		Icon              string                   `json:"icon"`
	}

	// LoanTagRequest represents the request structure for creating/updating loantag

	// LoanTagRequest represents the request structure for LoanTag.
	LoanTagRequest struct {
		LoanTransactionID *uuid.UUID  `json:"loan_transaction_id" validate:"required"`
		Name              string      `json:"name" validate:"required,min=1,max=50"`
		Description       string      `json:"description,omitempty"`
		Category          TagCategory `json:"category,omitempty"`
		Color             string      `json:"color,omitempty"`
		Icon              string      `json:"icon,omitempty"`
	}
)

func (m *Core) loanTag() {
	m.Migration = append(m.Migration, &LoanTag{})
	m.LoanTagManager = *registry.NewRegistry(registry.RegistryParams[
		LoanTag, LoanTagResponse, LoanTagRequest,
	]{
		Preloads: []string{
			"CreatedBy", "UpdatedBy", "LoanTransaction",
		},
		Database: m.provider.Service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return m.provider.Service.Broker.Dispatch(topics, payload)
		},
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
				Category:          data.Category,
				Color:             data.Color,
				Icon:              data.Icon,
			}
		},

		Created: func(data *LoanTag) registry.Topics {
			return []string{
				"loan_tag.create",
				fmt.Sprintf("loan_tag.create.%s", data.ID),
				fmt.Sprintf("loan_tag.create.branch.%s", data.BranchID),
				fmt.Sprintf("loan_tag.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *LoanTag) registry.Topics {
			return []string{
				"loan_tag.update",
				fmt.Sprintf("loan_tag.update.%s", data.ID),
				fmt.Sprintf("loan_tag.update.branch.%s", data.BranchID),
				fmt.Sprintf("loan_tag.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *LoanTag) registry.Topics {
			return []string{
				"loan_tag.delete",
				fmt.Sprintf("loan_tag.delete.%s", data.ID),
				fmt.Sprintf("loan_tag.delete.branch.%s", data.BranchID),
				fmt.Sprintf("loan_tag.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

// LoanTagCurrentBranch retrieves loan tags for the specified organization and branch.
func (m *Core) LoanTagCurrentBranch(context context.Context, organizationID uuid.UUID, branchID uuid.UUID) ([]*LoanTag, error) {
	return m.LoanTagManager.Find(context, &LoanTag{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}
