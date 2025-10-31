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
	AdjustmentTag struct {
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

		OrganizationID uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_adjustment_entry_tag" json:"organization_id"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_adjustment_entry_tag" json:"branch_id"`
		Branch         *Branch       `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"branch,omitempty"`

		AdjustmentEntryID *uuid.UUID       `gorm:"type:uuid" json:"adjustment_entry_id,omitempty"`
		AdjustmentEntry   *AdjustmentEntry `gorm:"foreignKey:AdjustmentEntryID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"adjustment_entry,omitempty"`

		Name        string `gorm:"type:varchar(50)" json:"name"`
		Description string `gorm:"type:text" json:"description"`
		Category    string `gorm:"type:varchar(50)" json:"category"`
		Color       string `gorm:"type:varchar(20)" json:"color"`
		Icon        string `gorm:"type:varchar(20)" json:"icon"`
	}

	AdjustmentTagResponse struct {
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
		AdjustmentEntryID *uuid.UUID               `json:"adjustment_entry_id,omitempty"`
		AdjustmentEntry   *AdjustmentEntryResponse `json:"adjustment_entry,omitempty"`
		Name              string                   `json:"name"`
		Description       string                   `json:"description"`
		Category          string                   `json:"category"`
		Color             string                   `json:"color"`
		Icon              string                   `json:"icon"`
	}

	AdjustmentTagRequest struct {
		AdjustmentEntryID *uuid.UUID `json:"adjustment_entry_id,omitempty"`
		Name              string     `json:"name,omitempty"`
		Description       string     `json:"description,omitempty"`
		Category          string     `json:"category,omitempty"`
		Color             string     `json:"color,omitempty"`
		Icon              string     `json:"icon,omitempty"`
	}
)

func (m *ModelCore) adjustmentTag() {
	m.Migration = append(m.Migration, &AdjustmentTag{})
	m.AdjustmentTagManager = services.NewRepository(services.RepositoryParams[
		AdjustmentTag, AdjustmentTagResponse, AdjustmentTagRequest,
	]{
		Preloads: []string{
			"CreatedBy", "UpdatedBy", "AdjustmentEntry",
		},
		Service: m.provider.Service,
		Resource: func(data *AdjustmentTag) *AdjustmentTagResponse {
			if data == nil {
				return nil
			}
			return &AdjustmentTagResponse{
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
				AdjustmentEntryID: data.AdjustmentEntryID,
				AdjustmentEntry:   m.AdjustmentEntryManager.ToModel(data.AdjustmentEntry),
				Name:              data.Name,
				Description:       data.Description,
				Category:          data.Category,
				Color:             data.Color,
				Icon:              data.Icon,
			}
		},
		Created: func(data *AdjustmentTag) []string {
			return []string{
				"adjustment_entry_tag.create",
				fmt.Sprintf("adjustment_entry_tag.create.%s", data.ID),
				fmt.Sprintf("adjustment_entry_tag.create.branch.%s", data.BranchID),
				fmt.Sprintf("adjustment_entry_tag.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *AdjustmentTag) []string {
			return []string{
				"adjustment_entry_tag.update",
				fmt.Sprintf("adjustment_entry_tag.update.%s", data.ID),
				fmt.Sprintf("adjustment_entry_tag.update.branch.%s", data.BranchID),
				fmt.Sprintf("adjustment_entry_tag.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *AdjustmentTag) []string {
			return []string{
				"adjustment_entry_tag.delete",
				fmt.Sprintf("adjustment_entry_tag.delete.%s", data.ID),
				fmt.Sprintf("adjustment_entry_tag.delete.branch.%s", data.BranchID),
				fmt.Sprintf("adjustment_entry_tag.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

// adjustmentTagCurrentbranch retrieves adjustment tags for a specific organization and branch.
func (m *ModelCore) AdjustmentTagCurrentbranch(context context.Context, orgId uuid.UUID, branchId uuid.UUID) ([]*AdjustmentTag, error) {
	return m.AdjustmentTagManager.Find(context, &AdjustmentTag{
		OrganizationID: orgId,
		BranchID:       branchId,
	})
}
