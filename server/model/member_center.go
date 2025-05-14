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
	MemberCenter struct {
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
		OrganizationID uuid.UUID      `gorm:"type:uuid;not null;index:idx_branch_org_member_center"`
		Organization   *Organization  `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID      `gorm:"type:uuid;not null;index:idx_branch_org_member_center"`
		Branch         *Branch        `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE;" json:"branch,omitempty"`

		Name        string `gorm:"type:varchar(255);not null"`
		Description string `gorm:"type:text;not null"`

		MemberProfiles []*MemberProfile `gorm:"foreignKey:MemberCenterID;references:ID" json:"member_profiles,omitempty"`
	}

	MemberCenterResponse struct {
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

		Name           string                   `json:"name"`
		Description    string                   `json:"description"`
		MemberProfiles []*MemberProfileResponse `json:"member_profiles"`
	}

	MemberCenterRequest struct {
		Name        string `json:"name,omitempty" validate:"required,max=255"`
		Description string `json:"description,omitempty" validate:"max=1024"`
	}

	MemberCenterCollection struct {
		Manager horizon_manager.CollectionManager[MemberCenter]
	}
)

func (m *Model) MemberCenterValidate(ctx echo.Context) (*MemberCenterRequest, error) {
	return horizon_manager.Validate[MemberCenterRequest](ctx, m.validator)
}

func (m *Model) MemberCenterModel(data *MemberCenter) *MemberCenterResponse {
	if data == nil {
		return nil
	}
	return horizon_manager.ToModel(data, func(data *MemberCenter) *MemberCenterResponse {
		return &MemberCenterResponse{
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

			MemberProfiles: m.MemberProfileModels(data.MemberProfiles),
		}
	})
}

func NewMemberCenterCollection(
	broadcast *horizon.HorizonBroadcast,
	database *horizon.HorizonDatabase,
	model *Model,
) (*MemberCenterCollection, error) {
	manager := horizon_manager.NewcollectionManager(
		database,
		broadcast,
		func(data *MemberCenter) ([]string, any) {
			return []string{
				fmt.Sprintf("member_center.create.%s", data.ID),
				fmt.Sprintf("member_center.create.banch.%s", data.BranchID),
				fmt.Sprintf("member_center.create.organization.%s", data.OrganizationID),
			}, model.MemberCenterModel(data)
		},
		func(data *MemberCenter) ([]string, any) {
			return []string{
				"member_center.update",
				fmt.Sprintf("member_center.update.%s", data.ID),
				fmt.Sprintf("member_center.update.banch.%s", data.BranchID),
				fmt.Sprintf("member_center.update.organization.%s", data.OrganizationID),
			}, model.MemberCenterModel(data)
		},
		func(data *MemberCenter) ([]string, any) {
			return []string{
				"member_center.delete",
				fmt.Sprintf("member_center.delete.%s", data.ID),
				fmt.Sprintf("member_center.delete.banch.%s", data.BranchID),
				fmt.Sprintf("member_center.delete.organization.%s", data.OrganizationID),
			}, model.MemberCenterModel(data)
		},
		[]string{
			"CreatedBy",
			"UpdatedBy",
			"Organization",
			"Branch",
		},
	)
	return &MemberCenterCollection{
		Manager: manager,
	}, nil
}

func (m *Model) MemberCenterModels(data []*MemberCenter) []*MemberCenterResponse {
	return horizon_manager.ToModels(data, m.MemberCenterModel)
}

// member-center/branch/:branch_id
func (fc *MemberCenterCollection) ListByBranch(branchID uuid.UUID) ([]*MemberCenter, error) {
	return fc.Manager.Find(&MemberCenter{
		BranchID: branchID,
	})
}

// member-center/organization/:organization_id
func (fc *MemberCenterCollection) ListByOrganization(organizationID uuid.UUID) ([]*MemberCenter, error) {
	return fc.Manager.Find(&MemberCenter{
		OrganizationID: organizationID,
	})
}

// member-center/organization/:organization_id/branch/:branch_id
func (fc *MemberCenterCollection) ListByOrganizationBranch(organizationID uuid.UUID, branchID uuid.UUID) ([]*MemberCenter, error) {
	return fc.Manager.Find(&MemberCenter{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}

func (fc *MemberCenterCollection) Seeder(userID uuid.UUID, organizationID uuid.UUID, branchID uuid.UUID) ([]*MemberCenter, error) {
	seedData := []*MemberCenter{
		{
			Name:           "Main Wellness Center",
			Description:    "Provides health and wellness programs.",
			OrganizationID: organizationID,
			BranchID:       branchID,
			CreatedAt:      time.Now(),
			CreatedByID:    userID,
			UpdatedAt:      time.Now(),
			UpdatedByID:    userID,
		},
		{
			ID:             uuid.New(),
			Name:           "Training Hub",
			Description:    "Offers skill-building and training for members.",
			OrganizationID: organizationID,
			BranchID:       branchID,
			CreatedAt:      time.Now(),
			CreatedByID:    userID,
			UpdatedAt:      time.Now(),
			UpdatedByID:    userID,
		},
		{
			ID:             uuid.New(),
			Name:           "Community Support Center",
			Description:    "Focuses on community support services and events.",
			OrganizationID: organizationID,
			BranchID:       branchID,
			CreatedAt:      time.Now(),
			CreatedByID:    userID,
			UpdatedAt:      time.Now(),
			UpdatedByID:    userID,
		},
	}
	for _, center := range seedData {
		if err := fc.Manager.Create(center); err != nil {
			return nil, err
		}
	}
	return seedData, nil
}
