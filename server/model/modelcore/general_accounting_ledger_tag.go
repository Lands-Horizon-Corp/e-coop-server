package modelcore

import (
	"context"
	"fmt"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/services"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Enum for tag_category (customize as needed)

type (
	GeneralLedgerTag struct {
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

		OrganizationID uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_general_ledger_tag"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_general_ledger_tag"`
		Branch         *Branch       `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"branch,omitempty"`

		GeneralLedgerID uuid.UUID      `gorm:"type:uuid;not null"`
		GeneralLedger   *GeneralLedger `gorm:"foreignKey:GeneralLedgerID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"general_ledger,omitempty"`

		Name        string      `gorm:"type:varchar(50)"`
		Description string      `gorm:"type:text"`
		Category    TagCategory `gorm:"type:varchar(50)"`
		Color       string      `gorm:"type:varchar(20)"`
		Icon        string      `gorm:"type:varchar(20)"`
	}

	GeneralLedgerTagResponse struct {
		ID              uuid.UUID              `json:"id"`
		CreatedAt       string                 `json:"created_at"`
		CreatedByID     uuid.UUID              `json:"created_by_id"`
		CreatedBy       *UserResponse          `json:"created_by,omitempty"`
		UpdatedAt       string                 `json:"updated_at"`
		UpdatedByID     uuid.UUID              `json:"updated_by_id"`
		UpdatedBy       *UserResponse          `json:"updated_by,omitempty"`
		OrganizationID  uuid.UUID              `json:"organization_id"`
		Organization    *OrganizationResponse  `json:"organization,omitempty"`
		BranchID        uuid.UUID              `json:"branch_id"`
		Branch          *BranchResponse        `json:"branch,omitempty"`
		GeneralLedgerID uuid.UUID              `json:"general_ledger_id"`
		GeneralLedger   *GeneralLedgerResponse `json:"general_ledger,omitempty"`
		Name            string                 `json:"name"`
		Description     string                 `json:"description"`
		Category        TagCategory            `json:"category"`
		Color           string                 `json:"color"`
		Icon            string                 `json:"icon"`
	}

	GeneralLedgerTagRequest struct {
		GeneralLedgerID uuid.UUID   `json:"general_ledger_id" validate:"required"`
		Name            string      `json:"name" validate:"required,min=1,max=50"`
		Description     string      `json:"description,omitempty"`
		Category        TagCategory `json:"category,omitempty"`
		Color           string      `json:"color,omitempty"`
		Icon            string      `json:"icon,omitempty"`
	}
)

func (m *ModelCore) generalLedgerTag() {
	m.Migration = append(m.Migration, &GeneralLedgerTag{})
	m.GeneralLedgerTagManager = services.NewRepository(services.RepositoryParams[
		GeneralLedgerTag, GeneralLedgerTagResponse, GeneralLedgerTagRequest,
	]{
		Preloads: []string{
			"CreatedBy", "UpdatedBy", "GeneralLedger",
		},
		Service: m.provider.Service,
		Resource: func(data *GeneralLedgerTag) *GeneralLedgerTagResponse {
			if data == nil {
				return nil
			}
			return &GeneralLedgerTagResponse{
				ID:              data.ID,
				CreatedAt:       data.CreatedAt.Format(time.RFC3339),
				CreatedByID:     data.CreatedByID,
				CreatedBy:       m.UserManager.ToModel(data.CreatedBy),
				UpdatedAt:       data.UpdatedAt.Format(time.RFC3339),
				UpdatedByID:     data.UpdatedByID,
				UpdatedBy:       m.UserManager.ToModel(data.UpdatedBy),
				OrganizationID:  data.OrganizationID,
				Organization:    m.OrganizationManager.ToModel(data.Organization),
				BranchID:        data.BranchID,
				Branch:          m.BranchManager.ToModel(data.Branch),
				GeneralLedgerID: data.GeneralLedgerID,
				GeneralLedger:   m.GeneralLedgerManager.ToModel(data.GeneralLedger),
				Name:            data.Name,
				Description:     data.Description,
				Category:        data.Category,
				Color:           data.Color,
				Icon:            data.Icon,
			}
		},
		Created: func(data *GeneralLedgerTag) []string {
			return []string{
				"general_ledger_tag.create",
				fmt.Sprintf("general_ledger_tag.create.%s", data.ID),
				fmt.Sprintf("general_ledger_tag.create.branch.%s", data.BranchID),
				fmt.Sprintf("general_ledger_tag.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *GeneralLedgerTag) []string {
			return []string{
				"general_ledger_tag.update",
				fmt.Sprintf("general_ledger_tag.update.%s", data.ID),
				fmt.Sprintf("general_ledger_tag.update.branch.%s", data.BranchID),
				fmt.Sprintf("general_ledger_tag.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *GeneralLedgerTag) []string {
			return []string{
				"general_ledger_tag.delete",
				fmt.Sprintf("general_ledger_tag.delete.%s", data.ID),
				fmt.Sprintf("general_ledger_tag.delete.branch.%s", data.BranchID),
				fmt.Sprintf("general_ledger_tag.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func (m *ModelCore) generalLedgerTagCurrentbranch(context context.Context, orgId uuid.UUID, branchId uuid.UUID) ([]*GeneralLedgerTag, error) {
	return m.GeneralLedgerTagManager.Find(context, &GeneralLedgerTag{
		OrganizationID: orgId,
		BranchID:       branchId,
	})
}
