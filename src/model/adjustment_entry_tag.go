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
	AdjustmentEntryTag struct {
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

		OrganizationID uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_adjustment_entry_tag"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_adjustment_entry_tag"`
		Branch         *Branch       `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"branch,omitempty"`

		AdjustmentEntryID uuid.UUID        `gorm:"type:uuid;not null"`
		AdjustmentEntry   *AdjustmentEntry `gorm:"foreignKey:AdjustmentEntryID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"adjustment_entry,omitempty"`

		Name        string `gorm:"type:varchar(50)"`
		Description string `gorm:"type:text"`
		Category    string `gorm:"type:varchar(50)"` // If you have a Go enum for tag_category, replace this type
		Color       string `gorm:"type:varchar(20)"`
		Icon        string `gorm:"type:varchar(20)"`
	}

	AdjustmentEntryTagResponse struct {
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
		AdjustmentEntryID uuid.UUID                `json:"adjustment_entry_id"`
		AdjustmentEntry   *AdjustmentEntryResponse `json:"adjustment_entry,omitempty"`
		Name              string                   `json:"name"`
		Description       string                   `json:"description"`
		Category          string                   `json:"category"`
		Color             string                   `json:"color"`
		Icon              string                   `json:"icon"`
	}

	AdjustmentEntryTagRequest struct {
		AdjustmentEntryID uuid.UUID `json:"adjustment_entry_id" validate:"required"`
		Name              string    `json:"name,omitempty"`
		Description       string    `json:"description,omitempty"`
		Category          string    `json:"category,omitempty"`
		Color             string    `json:"color,omitempty"`
		Icon              string    `json:"icon,omitempty"`
	}
)

func (m *Model) AdjustmentEntryTag() {
	m.Migration = append(m.Migration, &AdjustmentEntryTag{})
	m.AdjustmentEntryTagManager = horizon_services.NewRepository(horizon_services.RepositoryParams[
		AdjustmentEntryTag, AdjustmentEntryTagResponse, AdjustmentEntryTagRequest,
	]{
		Preloads: []string{
			"CreatedBy", "UpdatedBy", "DeletedBy", "Branch", "Organization", "AdjustmentEntry",
		},
		Service: m.provider.Service,
		Resource: func(data *AdjustmentEntryTag) *AdjustmentEntryTagResponse {
			if data == nil {
				return nil
			}
			return &AdjustmentEntryTagResponse{
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
		Created: func(data *AdjustmentEntryTag) []string {
			return []string{
				"adjustment_entry_tag.create",
				fmt.Sprintf("adjustment_entry_tag.create.%s", data.ID),
				fmt.Sprintf("adjustment_entry_tag.create.branch.%s", data.BranchID),
				fmt.Sprintf("adjustment_entry_tag.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *AdjustmentEntryTag) []string {
			return []string{
				"adjustment_entry_tag.update",
				fmt.Sprintf("adjustment_entry_tag.update.%s", data.ID),
				fmt.Sprintf("adjustment_entry_tag.update.branch.%s", data.BranchID),
				fmt.Sprintf("adjustment_entry_tag.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *AdjustmentEntryTag) []string {
			return []string{
				"adjustment_entry_tag.delete",
				fmt.Sprintf("adjustment_entry_tag.delete.%s", data.ID),
				fmt.Sprintf("adjustment_entry_tag.delete.branch.%s", data.BranchID),
				fmt.Sprintf("adjustment_entry_tag.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func (m *Model) AdjustmentEntryTagCurrentBranch(context context.Context, orgId uuid.UUID, branchId uuid.UUID) ([]*AdjustmentEntryTag, error) {
	return m.AdjustmentEntryTagManager.Find(context, &AdjustmentEntryTag{
		OrganizationID: orgId,
		BranchID:       branchId,
	})
}
