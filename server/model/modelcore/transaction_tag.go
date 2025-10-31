package modelcore

import (
	"context"
	"fmt"
	"time"

	horizon_services "github.com/Lands-Horizon-Corp/e-coop-server/services"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type (
	TransactionTag struct {
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

		OrganizationID uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_transaction_tag"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_transaction_tag"`
		Branch         *Branch       `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"branch,omitempty"`

		TransactionID uuid.UUID    `gorm:"type:uuid;not null"`
		Transaction   *Transaction `gorm:"foreignKey:TransactionID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"transaction,omitempty"`

		Name        string      `gorm:"type:varchar(50)"`
		Description string      `gorm:"type:text"`
		Category    TagCategory `gorm:"type:varchar(50)"`
		Color       string      `gorm:"type:varchar(20)"`
		Icon        string      `gorm:"type:varchar(20)"`
	}

	TransactionTagResponse struct {
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
		TransactionID  uuid.UUID             `json:"transaction_id"`
		Transaction    *TransactionResponse  `json:"transaction,omitempty"`
		Name           string                `json:"name"`
		Description    string                `json:"description"`
		Category       TagCategory           `json:"category"`
		Color          string                `json:"color"`
		Icon           string                `json:"icon"`
	}

	TransactionTagRequest struct {
		OrganizationID uuid.UUID   `json:"organization_id" validate:"required"`
		BranchID       uuid.UUID   `json:"branch_id" validate:"required"`
		TransactionID  uuid.UUID   `json:"transaction_id" validate:"required"`
		Name           string      `json:"name" validate:"required,min=1,max=50"`
		Description    string      `json:"description,omitempty"`
		Category       TagCategory `json:"category,omitempty"`
		Color          string      `json:"color,omitempty"`
		Icon           string      `json:"icon,omitempty"`
	}
)

func (m *ModelCore) TransactionTag() {
	m.Migration = append(m.Migration, &TransactionTag{})
	m.TransactionTagManager = horizon_services.NewRepository(horizon_services.RepositoryParams[
		TransactionTag, TransactionTagResponse, TransactionTagRequest,
	]{
		Preloads: []string{
			"CreatedBy", "UpdatedBy", "Transaction",
		},
		Service: m.provider.Service,
		Resource: func(data *TransactionTag) *TransactionTagResponse {
			if data == nil {
				return nil
			}
			return &TransactionTagResponse{
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
				TransactionID:  data.TransactionID,
				Transaction:    m.TransactionManager.ToModel(data.Transaction),
				Name:           data.Name,
				Description:    data.Description,
				Category:       data.Category,
				Color:          data.Color,
				Icon:           data.Icon,
			}
		},

		Created: func(data *TransactionTag) []string {
			return []string{
				"transaction_tag.create",
				fmt.Sprintf("transaction_tag.create.%s", data.ID),
				fmt.Sprintf("transaction_tag.create.branch.%s", data.BranchID),
				fmt.Sprintf("transaction_tag.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *TransactionTag) []string {
			return []string{
				"transaction_tag.update",
				fmt.Sprintf("transaction_tag.update.%s", data.ID),
				fmt.Sprintf("transaction_tag.update.branch.%s", data.BranchID),
				fmt.Sprintf("transaction_tag.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *TransactionTag) []string {
			return []string{
				"transaction_tag.delete",
				fmt.Sprintf("transaction_tag.delete.%s", data.ID),
				fmt.Sprintf("transaction_tag.delete.branch.%s", data.BranchID),
				fmt.Sprintf("transaction_tag.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func (m *ModelCore) TransactionTagCurrentBranch(context context.Context, orgId uuid.UUID, branchId uuid.UUID) ([]*TransactionTag, error) {
	return m.TransactionTagManager.Find(context, &TransactionTag{
		OrganizationID: orgId,
		BranchID:       branchId,
	})
}
