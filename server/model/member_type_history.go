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
	MemberTypeHistory struct {
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
		OrganizationID uuid.UUID      `gorm:"type:uuid;not null;index:idx_branch_org_member_type_history"`
		Organization   *Organization  `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID      `gorm:"type:uuid;not null;index:idx_branch_org_member_type_history"`
		Branch         *Branch        `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE;" json:"branch,omitempty"`

		MemberProfileID *uuid.UUID     `gorm:"type:uuid"`
		MemberProfile   *MemberProfile `gorm:"foreignKey:MemberProfileID;constraint:OnDelete:SET NULL;" json:"member_profile,omitempty"`

		MemberTypeID *uuid.UUID  `gorm:"type:uuid"`
		MemberType   *MemberType `gorm:"foreignKey:MemberTypeID;constraint:OnDelete:SET NULL;" json:"member_type,omitempty"`
	}

	MemberTypeHistoryResponse struct {
		ID              uuid.UUID              `json:"id"`
		CreatedAt       string                 `json:"created_at"`
		CreatedByID     uuid.UUID              `json:"created_by_id"`
		CreatedBy       *UserResponse          `json:"created_by,omitempty"`
		UpdatedAt       string                 `json:"updated_at"`
		UpdatedByID     uuid.UUID              `json:"updated_by_id"`
		UpdatedBy       *UserResponse          `json:"updated_by,omitempty"`
		OrganizationID  uuid.UUID              `json:"organization_id"`
		Organization    *OrganizationResponse  `json:"organization,omitempty"`
		BranchID        uuid.UUID              `json:"branch_id"`
		Branch          *BranchResponse        `json:"branch,omitempty"`
		MemberProfileID uuid.UUID              `json:"member_profile_id,omitempty"`
		MemberProfile   *MemberProfileResponse `json:"member_profile,omitempty"`
		MemberTypeID    uuid.UUID              `json:"member_type_id,omitempty"`
		MemberType      *MemberTypeResponse    `json:"member_type,omitempty"`
	}

	MemberTypeHistoryRequest struct {
		Name        string `json:"name,omitempty" validate:"required,max=255"`
		Description string `json:"description,omitempty" validate:"max=1024"`
	}

	MemberTypeHistoryCollection struct {
		Manager horizon_manager.CollectionManager[MemberTypeHistory]
	}
)

func (m *Model) MemberTypeHistoryValidate(ctx echo.Context) (*MemberTypeHistoryRequest, error) {
	return horizon_manager.Validate[MemberTypeHistoryRequest](ctx, m.validator)
}

func (m *Model) MemberTypeHistoryModel(data *MemberTypeHistory) *MemberTypeHistoryResponse {
	if data == nil {
		return nil
	}
	return horizon_manager.ToModel(data, func(data *MemberTypeHistory) *MemberTypeHistoryResponse {
		return &MemberTypeHistoryResponse{
			ID:              data.ID,
			CreatedAt:       data.CreatedAt.Format(time.RFC3339),
			CreatedByID:     data.CreatedByID,
			CreatedBy:       m.UserModel(data.CreatedBy),
			UpdatedAt:       data.UpdatedAt.Format(time.RFC3339),
			UpdatedByID:     data.UpdatedByID,
			UpdatedBy:       m.UserModel(data.UpdatedBy),
			OrganizationID:  data.OrganizationID,
			Organization:    m.OrganizationModel(data.Organization),
			BranchID:        data.BranchID,
			Branch:          m.BranchModel(data.Branch),
			MemberProfileID: *data.MemberProfileID,
			MemberProfile:   m.MemberProfileModel(data.MemberProfile),
			MemberTypeID:    *data.MemberTypeID,
			MemberType:      m.MemberTypeModel(data.MemberType),
		}
	})
}

func NewMemberTypeHistoryCollection(
	broadcast *horizon.HorizonBroadcast,
	database *horizon.HorizonDatabase,
	model *Model,
) (*MemberTypeHistoryCollection, error) {
	manager := horizon_manager.NewcollectionManager(
		database,
		broadcast,
		func(data *MemberTypeHistory) ([]string, any) {
			return []string{
				fmt.Sprintf("member_type_history.create.%s", data.ID),
				fmt.Sprintf("member_type_history.create.banch.%s", data.BranchID),
				fmt.Sprintf("member_type_history.create.organization.%s", data.OrganizationID),
			}, model.MemberTypeHistoryModel(data)
		},
		func(data *MemberTypeHistory) ([]string, any) {
			return []string{
				"member_type_history.update",
				fmt.Sprintf("member_type_history.update.%s", data.ID),
				fmt.Sprintf("member_type_history.update.banch.%s", data.BranchID),
				fmt.Sprintf("member_type_history.update.organization.%s", data.OrganizationID),
			}, model.MemberTypeHistoryModel(data)
		},
		func(data *MemberTypeHistory) ([]string, any) {
			return []string{
				"member_type_history.delete",
				fmt.Sprintf("member_type_history.delete.%s", data.ID),
				fmt.Sprintf("member_type_history.delete.banch.%s", data.BranchID),
				fmt.Sprintf("member_type_history.delete.organization.%s", data.OrganizationID),
			}, model.MemberTypeHistoryModel(data)
		},
		[]string{
			"CreatedBy",
			"UpdatedBy",
			"Organization",
			"Branch",
			"MemberType",
			"MemberProfile",
		},
	)
	return &MemberTypeHistoryCollection{
		Manager: manager,
	}, nil
}

func (m *Model) MemberTypeHistoryModels(data []*MemberTypeHistory) []*MemberTypeHistoryResponse {
	return horizon_manager.ToModels(data, m.MemberTypeHistoryModel)
}

// member-type-history/branch/:branch_id
func (fc *MemberTypeHistoryCollection) ListByBranch(branchID uuid.UUID) ([]*MemberTypeHistory, error) {
	return fc.Manager.Find(&MemberTypeHistory{
		BranchID: branchID,
	})
}

// member-type-history/organization/:organization_id
func (fc *MemberTypeHistoryCollection) ListByOrganization(organizationID uuid.UUID) ([]*MemberTypeHistory, error) {
	return fc.Manager.Find(&MemberTypeHistory{
		OrganizationID: organizationID,
	})
}

// member-type-history/organization/:organization_id/branch/:branch_id
func (fc *MemberTypeHistoryCollection) ListByOrganizationBranch(organizationID uuid.UUID, branchID uuid.UUID) ([]*MemberTypeHistory, error) {
	return fc.Manager.Find(&MemberTypeHistory{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}
