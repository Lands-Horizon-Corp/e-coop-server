package model

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/lib/pq"
	"horizon.com/server/horizon"
	horizon_manager "horizon.com/server/horizon/manager"

	"gorm.io/gorm"
)

type (
	PermissionTemplate struct {
		ID             uuid.UUID      `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
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

	PermissionTemplateRequest struct {
		ID *string `json:"id,omitempty"`

		Name        string   `json:"name" validate:"required,min=1,max=255"`
		Description string   `json:"description,omitempty"`
		Permissions []string `json:"permissions,omitempty"`
	}

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

	PermissionTemplateCollection struct {
		Manager horizon_manager.CollectionManager[PermissionTemplate]
	}
)

func (m *Model) PermissionTemplateValidate(ctx echo.Context) (*PermissionTemplateRequest, error) {
	return horizon_manager.Validate[PermissionTemplateRequest](ctx, m.validator)
}

func (m *Model) PermissionTemplateModel(data *PermissionTemplate) *PermissionTemplateResponse {
	if data == nil {
		return nil
	}
	return horizon_manager.ToModel(data, func(data *PermissionTemplate) *PermissionTemplateResponse {
		return &PermissionTemplateResponse{
			ID:             data.ID,
			CreatedAt:      data.CreatedAt.Format(time.RFC3339),
			CreatedByID:    data.CreatedByID,
			CreatedBy:      m.UserModel(data.CreatedBy),
			UpdatedAt:      data.UpdatedAt.Format(time.RFC3339),
			UpdatedByID:    data.UpdatedByID,
			UpdatedBy:      m.UserModel(data.CreatedBy),
			OrganizationID: data.OrganizationID,
			Organization:   m.OrganizationModel(data.Organization),
			BranchID:       data.BranchID,
			Branch:         m.BranchModel(data.Branch),

			Name:        data.Name,
			Description: data.Description,
			Permissions: data.Permissions,
		}
	})
}

func (m *Model) PermissionTemplateModels(data []*PermissionTemplate) []*PermissionTemplateResponse {
	return horizon_manager.ToModels(data, m.PermissionTemplateModel)
}

func NewPermissionTemplateCollection(
	broadcast *horizon.HorizonBroadcast,
	database *horizon.HorizonDatabase,
	model *Model,
) (*PermissionTemplateCollection, error) {
	manager := horizon_manager.NewcollectionManager(
		database,
		broadcast,
		func(data *PermissionTemplate) ([]string, any) {
			return []string{
				"permission_template.create",
				fmt.Sprintf("permission_template.create.%s", data.ID),
				fmt.Sprintf("permission_template.create.organization.%s", data.OrganizationID),
				fmt.Sprintf("permission_template.create.branch.%s", data.BranchID),
			}, model.PermissionTemplateModel(data)
		},
		func(data *PermissionTemplate) ([]string, any) {
			return []string{
				"permission_template.delete",
				fmt.Sprintf("permission_template.delete.%s", data.ID),
				fmt.Sprintf("permission_template.delete.organization.%s", data.OrganizationID),
				fmt.Sprintf("permission_template.delete.branch.%s", data.BranchID),
			}, model.PermissionTemplateModel(data)
		},
		func(data *PermissionTemplate) ([]string, any) {
			return []string{
				"permission_template.delete",
				fmt.Sprintf("permission_template.delete.%s", data.ID),
				fmt.Sprintf("permission_template.delete.organization.%s", data.OrganizationID),
				fmt.Sprintf("permission_template.delete.branch.%s", data.BranchID),
			}, model.PermissionTemplateModel(data)
		},
		[]string{
			"CreatedBy",
			"UpdatedBy",
			"Organization",
			"Branch",
		},
	)
	return &PermissionTemplateCollection{
		Manager: manager,
	}, nil
}

// permission-template/branch/:branch_id
func (fc *PermissionTemplateCollection) ListByBranch(branchID uuid.UUID) ([]*PermissionTemplate, error) {
	return fc.Manager.Find(&PermissionTemplate{
		BranchID: branchID,
	})
}

// permission-template/organization/:organization_id
func (fc *PermissionTemplateCollection) ListByOrganization(organizationID uuid.UUID) ([]*PermissionTemplate, error) {
	return fc.Manager.Find(&PermissionTemplate{
		OrganizationID: organizationID,
	})
}

// permission-template/organization/:organization_id/branch/:branch_id
func (fc *PermissionTemplateCollection) ListByOrganizationBranch(organizationID uuid.UUID, branchID uuid.UUID) ([]*PermissionTemplate, error) {
	return fc.Manager.Find(&PermissionTemplate{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}
