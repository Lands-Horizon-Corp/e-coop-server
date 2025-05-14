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
	MemberOccupation struct {
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
		OrganizationID uuid.UUID      `gorm:"type:uuid;not null;index:idx_branch_org_member_occupation"`
		Organization   *Organization  `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID      `gorm:"type:uuid;not null;index:idx_branch_org_member_occupation"`
		Branch         *Branch        `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE;" json:"branch,omitempty"`

		Name        string `gorm:"type:varchar(255);not null"`
		Description string `gorm:"type:text;not null"`
	}

	MemberOccupationResponse struct {
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

	MemberOccupationRequest struct {
		Name        string `json:"name,omitempty" validate:"required,max=255"`
		Description string `json:"description,omitempty" validate:"max=1024"`
	}

	MemberOccupationCollection struct {
		Manager horizon_manager.CollectionManager[MemberOccupation]
	}
)

func (m *Model) MemberOccupationValidate(ctx echo.Context) (*MemberOccupationRequest, error) {
	return horizon_manager.Validate[MemberOccupationRequest](ctx, m.validator)
}

func (m *Model) MemberOccupationModel(data *MemberOccupation) *MemberOccupationResponse {
	if data == nil {
		return nil
	}
	return horizon_manager.ToModel(data, func(data *MemberOccupation) *MemberOccupationResponse {
		return &MemberOccupationResponse{
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

func NewMemberOccupationCollection(
	broadcast *horizon.HorizonBroadcast,
	database *horizon.HorizonDatabase,
	model *Model,
) (*MemberOccupationCollection, error) {
	manager := horizon_manager.NewcollectionManager(
		database,
		broadcast,
		func(data *MemberOccupation) ([]string, any) {
			return []string{
				fmt.Sprintf("member_occupation.create.%s", data.ID),
				fmt.Sprintf("member_occupation.create.banch.%s", data.BranchID),
				fmt.Sprintf("member_occupation.create.organization.%s", data.OrganizationID),
			}, model.MemberOccupationModel(data)
		},
		func(data *MemberOccupation) ([]string, any) {
			return []string{
				"member_occupation.update",
				fmt.Sprintf("member_occupation.update.%s", data.ID),
				fmt.Sprintf("member_occupation.update.banch.%s", data.BranchID),
				fmt.Sprintf("member_occupation.update.organization.%s", data.OrganizationID),
			}, model.MemberOccupationModel(data)
		},
		func(data *MemberOccupation) ([]string, any) {
			return []string{
				"member_occupation.delete",
				fmt.Sprintf("member_occupation.delete.%s", data.ID),
				fmt.Sprintf("member_occupation.delete.banch.%s", data.BranchID),
				fmt.Sprintf("member_occupation.delete.organization.%s", data.OrganizationID),
			}, model.MemberOccupationModel(data)
		},
		[]string{
			"CreatedBy",
			"UpdatedBy",
			"Organization",
			"Branch",
		},
	)
	return &MemberOccupationCollection{
		Manager: manager,
	}, nil
}

func (m *Model) MemberOccupationModels(data []*MemberOccupation) []*MemberOccupationResponse {
	return horizon_manager.ToModels(data, m.MemberOccupationModel)
}

// member-occupation/branch/:branch_id
func (fc *MemberOccupationCollection) ListByBranch(branchID uuid.UUID) ([]*MemberOccupation, error) {
	return fc.Manager.Find(&MemberOccupation{
		BranchID: branchID,
	})
}

// member-occupation/organization/:organization_id
func (fc *MemberOccupationCollection) ListByOrganization(organizationID uuid.UUID) ([]*MemberOccupation, error) {
	return fc.Manager.Find(&MemberOccupation{
		OrganizationID: organizationID,
	})
}

// member-occupation/organization/:organization_id/branch/:branch_id
func (fc *MemberOccupationCollection) ListByOrganizationBranch(organizationID uuid.UUID, branchID uuid.UUID) ([]*MemberOccupation, error) {
	return fc.Manager.Find(&MemberOccupation{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}
