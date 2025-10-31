package modelcore

import (
	"context"
	"fmt"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/services"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type (
	AccountClassification struct {
		ID          uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
		CreatedAt   time.Time      `gorm:"not null;default:now()" json:"created_at"`
		CreatedByID uuid.UUID      `gorm:"type:uuid" json:"created_by_id"`
		CreatedBy   *User          `gorm:"foreignKey:CreatedByID;constraint:OnDelete:SET NULL;" json:"created_by,omitempty"`
		UpdatedAt   time.Time      `gorm:"not null;default:now()" json:"updated_at"`
		UpdatedByID uuid.UUID      `gorm:"type:uuid" json:"updated_by_id"`
		UpdatedBy   *User          `gorm:"foreignKey:UpdatedByID;constraint:OnDelete:SET NULL;" json:"updated_by,omitempty"`
		DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at"`
		DeletedByID *uuid.UUID     `gorm:"type:uuid" json:"deleted_by_id"`
		DeletedBy   *User          `gorm:"foreignKey:DeletedByID;constraint:OnDelete:SET NULL;" json:"deleted_by,omitempty"`

		OrganizationID uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_account_classification" json:"organization_id"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_account_classification" json:"branch_id"`
		Branch         *Branch       `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"branch,omitempty"`

		Name        string `gorm:"type:varchar(255)" json:"name"`
		Description string `gorm:"type:text" json:"description"`
	}

	AccountClassificationResponse struct {
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
		Description    string                `json:"description"`
	}

	AccountClassificationRequest struct {
		Name        string `json:"name" validate:"required,min=1,max=255"`
		Description string `json:"description,omitempty"`
	}
)

func (m *ModelCore) accountClassification() {
	m.Migration = append(m.Migration, &AccountClassification{})
	m.AccountClassificationManager = services.NewRepository(services.RepositoryParams[
		AccountClassification, AccountClassificationResponse, AccountClassificationRequest,
	]{
		Preloads: []string{"CreatedBy", "UpdatedBy", "Branch", "Organization"},
		Service:  m.provider.Service,
		Resource: func(data *AccountClassification) *AccountClassificationResponse {
			if data == nil {
				return nil
			}
			return &AccountClassificationResponse{
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
				Description:    data.Description,
			}
		},
		Created: func(data *AccountClassification) []string {
			return []string{
				"account_classification.create",
				fmt.Sprintf("account_classification.create.%s", data.ID),
				fmt.Sprintf("account_classification.create.branch.%s", data.BranchID),
				fmt.Sprintf("account_classification.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *AccountClassification) []string {
			return []string{
				"account_classification.update",
				fmt.Sprintf("account_classification.update.%s", data.ID),
				fmt.Sprintf("account_classification.update.branch.%s", data.BranchID),
				fmt.Sprintf("account_classification.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *AccountClassification) []string {
			return []string{
				"account_classification.delete",
				fmt.Sprintf("account_classification.delete.%s", data.ID),
				fmt.Sprintf("account_classification.delete.branch.%s", data.BranchID),
				fmt.Sprintf("account_classification.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

// AccountCategoryCurrentBranch retrieves all account categories for the specified organization and branch
func (m *ModelCore) AccountClassificationCurrentBranch(context context.Context, orgId uuid.UUID, branchId uuid.UUID) ([]*AccountClassification, error) {
	return m.AccountClassificationManager.Find(context, &AccountClassification{
		OrganizationID: orgId,
		BranchID:       branchId,
	})
}
