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
	AccountCategory struct {
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

		OrganizationID uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_account_category" json:"organization_id"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_account_category" json:"branch_id"`
		Branch         *Branch       `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"branch,omitempty"`

		Name        string `gorm:"type:varchar(255)" json:"name"`
		Description string `gorm:"type:text" json:"description"`
	}

	AccountCategoryResponse struct {
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

	AccountCategoryRequest struct {
		Name        string `json:"name" validate:"required,min=1,max=255"`
		Description string `json:"description,omitempty"`
	}
)

func (m *Model) AccountCategory() {
	m.Migration = append(m.Migration, &AccountCategory{})
	m.AccountCategoryManager = horizon_services.NewRepository(horizon_services.RepositoryParams[
		AccountCategory, AccountCategoryResponse, AccountCategoryRequest,
	]{
		Preloads: []string{"CreatedBy", "UpdatedBy", "Branch", "Organization"},
		Service:  m.provider.Service,
		Resource: func(data *AccountCategory) *AccountCategoryResponse {
			if data == nil {
				return nil
			}
			return &AccountCategoryResponse{
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
		Created: func(data *AccountCategory) []string {
			return []string{
				"account_category.create",
				fmt.Sprintf("account_category.create.%s", data.ID),
				fmt.Sprintf("account_category.create.branch.%s", data.BranchID),
				fmt.Sprintf("account_category.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *AccountCategory) []string {
			return []string{
				"account_category.update",
				fmt.Sprintf("account_category.update.%s", data.ID),
				fmt.Sprintf("account_category.update.branch.%s", data.BranchID),
				fmt.Sprintf("account_category.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *AccountCategory) []string {
			return []string{
				"account_category.delete",
				fmt.Sprintf("account_category.delete.%s", data.ID),
				fmt.Sprintf("account_category.delete.branch.%s", data.BranchID),
				fmt.Sprintf("account_category.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func (m *Model) AccountCategoryCurrentBranch(context context.Context, orgId uuid.UUID, branchId uuid.UUID) ([]*AccountCategory, error) {
	return m.AccountCategoryManager.Find(context, &AccountCategory{
		OrganizationID: orgId,
		BranchID:       branchId,
	})
}
