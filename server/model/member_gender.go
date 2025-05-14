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
	MemberGender struct {
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
		OrganizationID uuid.UUID      `gorm:"type:uuid;not null;index:idx_branch_org_member_gender"`
		Organization   *Organization  `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID      `gorm:"type:uuid;not null;index:idx_branch_org_member_gender"`
		Branch         *Branch        `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE;" json:"branch,omitempty"`

		Name        string `gorm:"type:varchar(255);not null"`
		Description string `gorm:"type:text;not null"`

		MemberProfiles []*MemberProfile `gorm:"foreignKey:MemberGenderID;references:ID" json:"member_profiles,omitempty"`
	}

	MemberGenderResponse struct {
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

	MemberGenderRequest struct {
		Name        string `json:"name,omitempty" validate:"required,max=255"`
		Description string `json:"description,omitempty" validate:"max=1024"`
	}

	MemberGenderCollection struct {
		Manager horizon_manager.CollectionManager[MemberGender]
	}
)

func (m *Model) MemberGenderValidate(ctx echo.Context) (*MemberGenderRequest, error) {
	return horizon_manager.Validate[MemberGenderRequest](ctx, m.validator)
}

func (m *Model) MemberGenderModel(data *MemberGender) *MemberGenderResponse {
	if data == nil {
		return nil
	}
	return horizon_manager.ToModel(data, func(data *MemberGender) *MemberGenderResponse {
		return &MemberGenderResponse{
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

func NewMemberGenderCollection(
	broadcast *horizon.HorizonBroadcast,
	database *horizon.HorizonDatabase,
	model *Model,
) (*MemberGenderCollection, error) {
	manager := horizon_manager.NewcollectionManager(
		database,
		broadcast,
		func(data *MemberGender) ([]string, any) {
			return []string{
				fmt.Sprintf("member_gender.create.%s", data.ID),
				fmt.Sprintf("member_gender.create.banch.%s", data.BranchID),
				fmt.Sprintf("member_gender.create.organization.%s", data.OrganizationID),
			}, model.MemberGenderModel(data)
		},
		func(data *MemberGender) ([]string, any) {
			return []string{
				"member_gender.update",
				fmt.Sprintf("member_gender.update.%s", data.ID),
				fmt.Sprintf("member_gender.update.banch.%s", data.BranchID),
				fmt.Sprintf("member_gender.update.organization.%s", data.OrganizationID),
			}, model.MemberGenderModel(data)
		},
		func(data *MemberGender) ([]string, any) {
			return []string{
				"member_gender.delete",
				fmt.Sprintf("member_gender.delete.%s", data.ID),
				fmt.Sprintf("member_gender.delete.banch.%s", data.BranchID),
				fmt.Sprintf("member_gender.delete.organization.%s", data.OrganizationID),
			}, model.MemberGenderModel(data)
		},
		[]string{
			"CreatedBy",
			"UpdatedBy",
			"Organization",
			"Branch",
		},
	)
	return &MemberGenderCollection{
		Manager: manager,
	}, nil
}

func (m *Model) MemberGenderModels(data []*MemberGender) []*MemberGenderResponse {
	return horizon_manager.ToModels(data, m.MemberGenderModel)
}

// member-gender/branch/:branch_id
func (fc *MemberGenderCollection) ListByBranch(branchID uuid.UUID) ([]*MemberGender, error) {
	return fc.Manager.Find(&MemberGender{
		BranchID: branchID,
	})
}

// member-gender/organization/:organization_id
func (fc *MemberGenderCollection) ListByOrganization(organizationID uuid.UUID) ([]*MemberGender, error) {
	return fc.Manager.Find(&MemberGender{
		OrganizationID: organizationID,
	})
}

// member-gender/organization/:organization_id/branch/:branch_id
func (fc *MemberGenderCollection) ListByOrganizationBranch(organizationID uuid.UUID, branchID uuid.UUID) ([]*MemberGender, error) {
	return fc.Manager.Find(&MemberGender{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}

func (fc *MemberGenderCollection) Seeder(userID uuid.UUID, organizationID uuid.UUID, branchID uuid.UUID) ([]*MemberGender, error) {
	now := time.Now()

	genders := []*MemberGender{
		{
			ID:             uuid.New(),
			CreatedAt:      now,
			CreatedByID:    userID,
			UpdatedAt:      now,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Male",
			Description:    "Identifies as male.",
		},
		{
			ID:             uuid.New(),
			CreatedAt:      now,
			CreatedByID:    userID,
			UpdatedAt:      now,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Female",
			Description:    "Identifies as female.",
		},
		{
			ID:             uuid.New(),
			CreatedAt:      now,
			CreatedByID:    userID,
			UpdatedAt:      now,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Other",
			Description:    "Identifies outside the binary gender categories.",
		},
	}

	if err := fc.Manager.CreateMany(genders); err != nil {
		return nil, err
	}

	return genders, nil
}
