package model

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
	"horizon.com/server/horizon"
	horizon_manager "horizon.com/server/horizon/manager"
)

type (
	MemberGroup struct {
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
		OrganizationID uuid.UUID      `gorm:"type:uuid;not null;index:idx_branch_org_member_group"`
		Organization   *Organization  `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID      `gorm:"type:uuid;not null;index:idx_branch_org_member_group"`
		Branch         *Branch        `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE;" json:"branch,omitempty"`

		Name        string `gorm:"type:varchar(255);not null"`
		Description string `gorm:"type:text;not null"`
	}

	MemberGroupResponse struct {
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

		Name        string `json:"name"`
		Description string `json:"description"`
	}

	MemberGroupRequest struct {
		Name        string `json:"name,omitempty" validate:"required,max=255"`
		Description string `json:"description,omitempty" validate:"max=1024"`
	}

	MemberGroupCollection struct {
		Manager horizon_manager.CollectionManager[MemberGroup]
	}
)

func (m *Model) MemberGroupValidate(ctx echo.Context) (*MemberGroupRequest, error) {
	return horizon_manager.Validate[MemberGroupRequest](ctx, m.validator)
}

func (m *Model) MemberGroupModel(data *MemberGroup) *MemberGroupResponse {
	if data == nil {
		return nil
	}
	return horizon_manager.ToModel(data, func(data *MemberGroup) *MemberGroupResponse {
		return &MemberGroupResponse{
			ID:             data.ID,
			CreatedAt:      data.CreatedAt.Format(time.RFC3339),
			CreatedByID:    data.CreatedByID,
			CreatedBy:      m.UserModel(data.CreatedBy),
			UpdatedAt:      data.UpdatedAt.Format(time.RFC3339),
			UpdatedByID:    data.UpdatedByID,
			UpdatedBy:      m.UserModel(data.UpdatedBy),
			OrganizationID: data.OrganizationID,
			Organization:   m.OrganizationModel(data.Organization),
			BranchID:       data.BranchID,
			Branch:         m.BranchModel(data.Branch),
			Name:           data.Name,
			Description:    data.Description,
		}
	})
}

func NewMemberGroupCollection(
	broadcast *horizon.HorizonBroadcast,
	database *horizon.HorizonDatabase,
	model *Model,
) (*MemberGroupCollection, error) {
	manager := horizon_manager.NewcollectionManager(
		database,
		broadcast,
		func(data *MemberGroup) ([]string, any) {
			return []string{
				fmt.Sprintf("member_group.create.%s", data.ID),
				fmt.Sprintf("member_group.create.banch.%s", data.BranchID),
				fmt.Sprintf("member_group.create.organization.%s", data.OrganizationID),
			}, model.MemberGroupModel(data)
		},
		func(data *MemberGroup) ([]string, any) {
			return []string{
				"member_group.update",
				fmt.Sprintf("member_group.update.%s", data.ID),
				fmt.Sprintf("member_group.update.banch.%s", data.BranchID),
				fmt.Sprintf("member_group.update.organization.%s", data.OrganizationID),
			}, model.MemberGroupModel(data)
		},
		func(data *MemberGroup) ([]string, any) {
			return []string{
				"member_group.delete",
				fmt.Sprintf("member_group.delete.%s", data.ID),
				fmt.Sprintf("member_group.delete.banch.%s", data.BranchID),
				fmt.Sprintf("member_group.delete.organization.%s", data.OrganizationID),
			}, model.MemberGroupModel(data)
		},
		[]string{
			"CreatedBy",
			"UpdatedBy",
			"Organization",
			"Branch",
		},
	)
	return &MemberGroupCollection{
		Manager: manager,
	}, nil
}

func (m *Model) MemberGroupModels(data []*MemberGroup) []*MemberGroupResponse {
	return horizon_manager.ToModels(data, m.MemberGroupModel)
}

// member-group/branch/:branch_id
func (fc *MemberGroupCollection) ListByBranch(branchID uuid.UUID) ([]*MemberGroup, error) {
	return fc.Manager.Find(&MemberGroup{
		BranchID: branchID,
	})
}

// member-group/organization/:organization_id
func (fc *MemberGroupCollection) ListByOrganization(organizationID uuid.UUID) ([]*MemberGroup, error) {
	return fc.Manager.Find(&MemberGroup{
		OrganizationID: organizationID,
	})
}

// member-group/organization/:organization_id/branch/:branch_id
func (fc *MemberGroupCollection) ListByOrganizationBranch(organizationID uuid.UUID, branchID uuid.UUID) ([]*MemberGroup, error) {
	return fc.Manager.Find(&MemberGroup{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}
