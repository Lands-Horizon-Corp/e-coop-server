package model

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	horizon_services "github.com/lands-horizon/horizon-server/services"
	"gorm.io/gorm"
)

// TagCategory is an enum for tag_category

type (
	TagTemplate struct {
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

		OrganizationID uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_tag_template"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_tag_template"`
		Branch         *Branch       `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"branch,omitempty"`

		Name        string `gorm:"type:varchar(50)"`
		Description string `gorm:"type:text"`
		Category    string `gorm:"type:varchar(50)"`
		Color       string `gorm:"type:varchar(20)"`
		Icon        string `gorm:"type:varchar(20)"`
	}

	TagTemplateResponse struct {
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
		Category       TagCategory           `json:"category"`
		Color          string                `json:"color"`
		Icon           string                `json:"icon"`
	}

	TagTemplateRequest struct {
		Name        string      `json:"name" validate:"required,min=1,max=50"`
		Description string      `json:"description,omitempty"`
		Category    TagCategory `json:"category,omitempty"`
		Color       string      `json:"color,omitempty"`
		Icon        string      `json:"icon,omitempty"`
	}
)

func (m *Model) TagTemplate() {
	m.Migration = append(m.Migration, &TagTemplate{})
	m.TagTemplateManager = horizon_services.NewRepository(horizon_services.RepositoryParams[
		TagTemplate, TagTemplateResponse, TagTemplateRequest,
	]{
		Preloads: []string{
			"CreatedBy", "UpdatedBy", "DeletedBy", "Branch", "Organization",
		},
		Service: m.provider.Service,
		Resource: func(data *TagTemplate) *TagTemplateResponse {
			if data == nil {
				return nil
			}
			return &TagTemplateResponse{
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
				Category:       TagCategory(data.Category),
				Color:          data.Color,
				Icon:           data.Icon,
			}
		},

		Created: func(data *TagTemplate) []string {
			return []string{
				"tag_template.create",
				fmt.Sprintf("tag_template.create.%s", data.ID),
				fmt.Sprintf("tag_template.create.branch.%s", data.BranchID),
				fmt.Sprintf("tag_template.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *TagTemplate) []string {
			return []string{
				"tag_template.update",
				fmt.Sprintf("tag_template.update.%s", data.ID),
				fmt.Sprintf("tag_template.update.branch.%s", data.BranchID),
				fmt.Sprintf("tag_template.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *TagTemplate) []string {
			return []string{
				"tag_template.delete",
				fmt.Sprintf("tag_template.delete.%s", data.ID),
				fmt.Sprintf("tag_template.delete.branch.%s", data.BranchID),
				fmt.Sprintf("tag_template.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func (m *Model) TagTemplateCurrentBranch(context context.Context, orgId uuid.UUID, branchId uuid.UUID) ([]*TagTemplate, error) {
	return m.TagTemplateManager.Find(context, &TagTemplate{
		OrganizationID: orgId,
		BranchID:       branchId,
	})
}
