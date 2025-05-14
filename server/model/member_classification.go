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
	MemberClassification struct {
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
		OrganizationID uuid.UUID      `gorm:"type:uuid;not null;index:idx_branch_org_member_classification"`
		Organization   *Organization  `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID      `gorm:"type:uuid;not null;index:idx_branch_org_member_classification"`
		Branch         *Branch        `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE;" json:"branch,omitempty"`

		Name        string `gorm:"type:varchar(255);not null"`
		Icon        string `gorm:"type:varchar(255);default:''"`
		Description string `gorm:"type:text;not null"`
	}

	MemberClassificationResponse struct {
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
		Icon        string `json:"icon"`
	}

	MemberClassificationRequest struct {
		Name        string `json:"name,omitempty" validate:"required,max=255"`
		Icon        string `json:"icon,omitempty" validate:"max=255"`
		Description string `json:"description,omitempty" validate:"max=1024"`
	}

	MemberClassificationCollection struct {
		Manager horizon_manager.CollectionManager[MemberClassification]
	}
)

func (m *Model) MemberClassificationValidate(ctx echo.Context) (*MemberClassificationRequest, error) {
	return horizon_manager.Validate[MemberClassificationRequest](ctx, m.validator)
}

func (m *Model) MemberClassificationModel(data *MemberClassification) *MemberClassificationResponse {
	if data == nil {
		return nil
	}
	return horizon_manager.ToModel(data, func(data *MemberClassification) *MemberClassificationResponse {
		return &MemberClassificationResponse{
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
			Icon:           data.Icon,
		}
	})
}

func NewMemberClassificationCollection(
	broadcast *horizon.HorizonBroadcast,
	database *horizon.HorizonDatabase,
	model *Model,
) (*MemberClassificationCollection, error) {
	manager := horizon_manager.NewcollectionManager(
		database,
		broadcast,
		func(data *MemberClassification) ([]string, any) {
			return []string{
				fmt.Sprintf("member_classification.create.%s", data.ID),
				fmt.Sprintf("member_classification.create.banch.%s", data.BranchID),
				fmt.Sprintf("member_classification.create.organization.%s", data.OrganizationID),
			}, model.MemberClassificationModel(data)
		},
		func(data *MemberClassification) ([]string, any) {
			return []string{
				"member_classification.update",
				fmt.Sprintf("member_classification.update.%s", data.ID),
				fmt.Sprintf("member_classification.update.banch.%s", data.BranchID),
				fmt.Sprintf("member_classification.update.organization.%s", data.OrganizationID),
			}, model.MemberClassificationModel(data)
		},
		func(data *MemberClassification) ([]string, any) {
			return []string{
				"member_classification.delete",
				fmt.Sprintf("member_classification.delete.%s", data.ID),
				fmt.Sprintf("member_classification.delete.banch.%s", data.BranchID),
				fmt.Sprintf("member_classification.delete.organization.%s", data.OrganizationID),
			}, model.MemberClassificationModel(data)
		},
		[]string{
			"CreatedBy",
			"UpdatedBy",
			"Organization",
			"Branch",
		},
	)
	return &MemberClassificationCollection{
		Manager: manager,
	}, nil
}

func (m *Model) MemberClassificationModels(data []*MemberClassification) []*MemberClassificationResponse {
	return horizon_manager.ToModels(data, m.MemberClassificationModel)
}

// member-classification/branch/:branch_id
func (fc *MemberClassificationCollection) ListByBranch(branchID uuid.UUID) ([]*MemberClassification, error) {
	return fc.Manager.Find(&MemberClassification{
		BranchID: branchID,
	})
}

// member-classification/organization/:organization_id
func (fc *MemberClassificationCollection) ListByOrganization(organizationID uuid.UUID) ([]*MemberClassification, error) {
	return fc.Manager.Find(&MemberClassification{
		OrganizationID: organizationID,
	})
}

// member-classification/organization/:organization_id/branch/:branch_id
func (fc *MemberClassificationCollection) ListByOrganizationBranch(organizationID uuid.UUID, branchID uuid.UUID) ([]*MemberClassification, error) {
	return fc.Manager.Find(&MemberClassification{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}

func (fc *MemberClassificationCollection) Seeder(userID uuid.UUID, organizationID uuid.UUID, branchID uuid.UUID) ([]*MemberClassification, error) {
	now := time.Now()
	classifications := []*MemberClassification{
		{
			ID:             uuid.New(),
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Gold",
			Icon:           "sunrise",
			Description:    "Gold membership is reserved for top-tier members with excellent credit scores and consistent loyalty.",
		},
		{
			ID:             uuid.New(),
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Silver",
			Icon:           "moon-star",
			Description:    "Silver membership is designed for members with good credit history and regular engagement.",
		},
		{
			ID:             uuid.New(),
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Bronze",
			Icon:           "cloud",
			Description:    "Bronze membership is for new or casual members who are starting their journey with us.",
		},
		{
			ID:             uuid.New(),
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Platinum",
			Icon:           "gem",
			Description:    "Platinum membership offers exclusive benefits to elite members with outstanding history and contributions.",
		},
	}

	for _, mc := range classifications {
		if err := fc.Manager.Create(mc); err != nil {
			return nil, fmt.Errorf("failed to seed member classification %s: %w", mc.Name, err)
		}
	}

	return classifications, nil
}
