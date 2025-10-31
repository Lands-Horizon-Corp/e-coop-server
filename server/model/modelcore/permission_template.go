package modelcore

import (
	"context"
	"fmt"
	"time"

	horizon_services "github.com/Lands-Horizon-Corp/e-coop-server/services"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

type (
	// PermissionTemplate represents predefined sets of permissions for users within organizations and branches
	PermissionTemplate struct {
		ID             uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
		CreatedAt      time.Time      `gorm:"not null;default:now()"`
		CreatedByID    uuid.UUID      `gorm:"type:uuid"`
		CreatedBy      *User          `gorm:"foreignKey:CreatedByID;constraint:OnDelete:SET NULL;" json:"created_by,omitempty"`
		UpdatedAt      time.Time      `gorm:"not null;default:now()"`
		UpdatedByID    uuid.UUID      `gorm:"type:uuid"`
		UpdatedBy      *User          `gorm:"foreignKey:UpdatedByID;constraint:OnDelete:SET NULL;" json:"updated_by,omitempty"`
		DeletedAt      gorm.DeletedAt `gorm:"index"`
		DeletedByID    *uuid.UUID     `gorm:"type:uuid"`
		DeletedBy      *User          `gorm:"foreignKey:DeletedByID;constraint:OnDelete:SET NULL;" json:"deleted_by,omitempty"`
		OrganizationID uuid.UUID      `gorm:"type:uuid;not null;index:idx_branch_org_permission_template"`
		Organization   *Organization  `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID      `gorm:"type:uuid;not null;index:idx_branch_org_permission_template"`
		Branch         *Branch        `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE;" json:"branch,omitempty"`

		Name        string         `gorm:"type:varchar(255);not null"`
		Description string         `gorm:"type:text"`
		Permissions pq.StringArray `gorm:"type:varchar[];default:'{}'"`
	}

	// PermissionTemplateRequest represents the request structure for creating or updating permission templates
	PermissionTemplateRequest struct {
		ID *uuid.UUID `json:"id,omitempty"`

		Name        string   `json:"name" validate:"required,min=1,max=255"`
		Description string   `json:"description,omitempty"`
		Permissions []string `json:"permissions,omitempty"`
	}

	// PermissionTemplateResponse represents the response structure for permission template data
	PermissionTemplateResponse struct {
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

		Name        string   `json:"name"`
		Description string   `json:"description,omitempty"`
		Permissions []string `json:"permissions"`
	}
)

// PermissionTemplate initializes the permission template model and its repository manager
func (m *ModelCore) permissionTemplate() {
	m.migration = append(m.migration, &PermissionTemplate{})
	m.permissionTemplateManager = horizon_services.NewRepository(horizon_services.RepositoryParams[PermissionTemplate, PermissionTemplateResponse, PermissionTemplateRequest]{
		Preloads: []string{
			"CreatedBy",
			"UpdatedBy",
			"Branch",
			"Branch.Media",
			"Organization",
			"Organization.Media",
			"Organization.CoverMedia",
			"Organization.OrganizationCategories",
			"Organization.OrganizationCategories.Category",
		},
		Service: m.provider.Service,
		Resource: func(data *PermissionTemplate) *PermissionTemplateResponse {
			if data == nil {
				return nil
			}
			if data.Permissions == nil {
				data.Permissions = []string{}
			}
			return &PermissionTemplateResponse{
				ID:             data.ID,
				CreatedAt:      data.CreatedAt.Format(time.RFC3339),
				CreatedByID:    data.CreatedByID,
				CreatedBy:      m.userManager.ToModel(data.CreatedBy),
				UpdatedAt:      data.UpdatedAt.Format(time.RFC3339),
				UpdatedByID:    data.UpdatedByID,
				UpdatedBy:      m.userManager.ToModel(data.UpdatedBy),
				OrganizationID: data.OrganizationID,
				Organization:   m.organizationManager.ToModel(data.Organization),
				BranchID:       data.BranchID,
				Branch:         m.branchManager.ToModel(data.Branch),

				Name:        data.Name,
				Description: data.Description,
				Permissions: data.Permissions,
			}
		},

		Created: func(data *PermissionTemplate) []string {
			return []string{
				"permission_template.create",
				fmt.Sprintf("permission_template.create.%s", data.ID),
				fmt.Sprintf("permission_template.create.branch.%s", data.BranchID),
				fmt.Sprintf("permission_template.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *PermissionTemplate) []string {
			return []string{
				"permission_template.update",
				fmt.Sprintf("permission_template.update.%s", data.ID),
				fmt.Sprintf("permission_template.update.branch.%s", data.BranchID),
				fmt.Sprintf("permission_template.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *PermissionTemplate) []string {
			return []string{
				"permission_template.delete",
				fmt.Sprintf("permission_template.delete.%s", data.ID),
				fmt.Sprintf("permission_template.delete.branch.%s", data.BranchID),
				fmt.Sprintf("permission_template.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

// GetPermissionTemplateByBranch retrieves permission templates for a specific branch within an organization
func (m *ModelCore) getPermissionTemplateBybranch(context context.Context, organizationID uuid.UUID, branchID uuid.UUID) ([]*PermissionTemplate, error) {
	return m.permissionTemplateManager.Find(context, &PermissionTemplate{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}
