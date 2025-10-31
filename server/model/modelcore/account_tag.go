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
	AccountTag struct {
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

		OrganizationID uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_account_tag" json:"organization_id"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_account_tag" json:"branch_id"`
		Branch         *Branch       `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"branch,omitempty"`

		AccountID uuid.UUID `gorm:"type:uuid;not null" json:"account_id"`
		Account   *Account  `gorm:"foreignKey:AccountID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"account,omitempty"`

		Name        string      `gorm:"type:varchar(50)" json:"name"`
		Description string      `gorm:"type:text" json:"description"`
		Category    TagCategory `gorm:"type:varchar(50)" json:"category"`
		Color       string      `gorm:"type:varchar(20)" json:"color"`
		Icon        string      `gorm:"type:varchar(20)" json:"icon"`
	}

	// AccountTagResponse represents the response structure for accounttag data

	AccountTagResponse struct {
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
		AccountID      uuid.UUID             `json:"account_id"`
		Account        *AccountResponse      `json:"account,omitempty"`
		Name           string                `json:"name"`
		Description    string                `json:"description"`
		Category       TagCategory           `json:"category"`
		Color          string                `json:"color"`
		Icon           string                `json:"icon"`
	}

	// AccountTagRequest represents the request structure for creating/updating accounttag

	AccountTagRequest struct {
		AccountID   uuid.UUID   `json:"account_id" validate:"required"`
		Name        string      `json:"name" validate:"required,min=1,max=50"`
		Description string      `json:"description,omitempty"`
		Category    TagCategory `json:"category,omitempty"`
		Color       string      `json:"color,omitempty"`
		Icon        string      `json:"icon,omitempty"`
	}
)

func (m *ModelCore) accountTag() {
	m.Migration = append(m.Migration, &AccountTag{})
	m.AccountTagManager = services.NewRepository(services.RepositoryParams[
		AccountTag, AccountTagResponse, AccountTagRequest,
	]{
		Preloads: []string{"CreatedBy", "UpdatedBy", "Account"},
		Service:  m.provider.Service,
		Resource: func(data *AccountTag) *AccountTagResponse {
			if data == nil {
				return nil
			}
			return &AccountTagResponse{
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
				AccountID:      data.AccountID,
				Account:        m.AccountManager.ToModel(data.Account),
				Name:           data.Name,
				Description:    data.Description,
				Category:       data.Category,
				Color:          data.Color,
				Icon:           data.Icon,
			}
		},
		Created: func(data *AccountTag) []string {
			return []string{
				"account_tag.create",
				fmt.Sprintf("account_tag.create.%s", data.ID),
				fmt.Sprintf("account_tag.create.branch.%s", data.BranchID),
				fmt.Sprintf("account_tag.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *AccountTag) []string {
			return []string{
				"account_tag.update",
				fmt.Sprintf("account_tag.update.%s", data.ID),
				fmt.Sprintf("account_tag.update.branch.%s", data.BranchID),
				fmt.Sprintf("account_tag.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *AccountTag) []string {
			return []string{
				"account_tag.delete",
				fmt.Sprintf("account_tag.delete.%s", data.ID),
				fmt.Sprintf("account_tag.delete.branch.%s", data.BranchID),
				fmt.Sprintf("account_tag.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

// AccountTagCurrentBranch retrieves account tags for a specific organization and branch.
func (m *ModelCore) AccountTagCurrentBranch(context context.Context, orgID uuid.UUID, branchID uuid.UUID) ([]*AccountTag, error) {
	return m.AccountTagManager.Find(context, &AccountTag{
		OrganizationID: orgID,
		BranchID:       branchID,
	})
}
