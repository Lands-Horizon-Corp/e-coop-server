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
	MemberCenterHistory struct {
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
		OrganizationID uuid.UUID      `gorm:"type:uuid;not null;index:idx_branch_org_member_center_history"`
		Organization   *Organization  `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID      `gorm:"type:uuid;not null;index:idx_branch_org_member_center_history"`
		Branch         *Branch        `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE;" json:"branch,omitempty"`

		MemberProfileID *uuid.UUID     `gorm:"type:uuid"`
		MemberProfile   *MemberProfile `gorm:"foreignKey:MemberProfileID;constraint:OnDelete:SET NULL;" json:"member_profile,omitempty"`

		MemberCenterID *uuid.UUID    `gorm:"type:uuid"`
		MemberCenter   *MemberCenter `gorm:"foreignKey:MemberCenterID;constraint:OnDelete:SET NULL;" json:"member_center,omitempty"`
	}

	MemberCenterHistoryResponse struct {
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
		MemberCenterID  uuid.UUID              `json:"member_center_id,omitempty"`
		MemberCenter    *MemberCenterResponse  `json:"member_center,omitempty"`
	}

	MemberCenterHistoryRequest struct {
		Name        string `json:"name,omitempty" validate:"required,max=255"`
		Description string `json:"description,omitempty" validate:"max=1024"`
	}

	MemberCenterHistoryCollection struct {
		Manager horizon_manager.CollectionManager[MemberCenterHistory]
	}
)

func (m *Model) MemberCenterHistoryValidate(ctx echo.Context) (*MemberCenterHistoryRequest, error) {
	return horizon_manager.Validate[MemberCenterHistoryRequest](ctx, m.validator)
}

func (m *Model) MemberCenterHistoryModel(data *MemberCenterHistory) *MemberCenterHistoryResponse {
	if data == nil {
		return nil
	}
	return horizon_manager.ToModel(data, func(data *MemberCenterHistory) *MemberCenterHistoryResponse {
		return &MemberCenterHistoryResponse{
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
			MemberCenterID:  *data.MemberCenterID,
			MemberCenter:    m.MemberCenterModel(data.MemberCenter),
		}
	})
}

func NewMemberCenterHistoryCollection(
	broadcast *horizon.HorizonBroadcast,
	database *horizon.HorizonDatabase,
	model *Model,
) (*MemberCenterHistoryCollection, error) {
	manager := horizon_manager.NewcollectionManager(
		database,
		broadcast,
		func(data *MemberCenterHistory) ([]string, any) {
			return []string{
				fmt.Sprintf("member_center_history.create.%s", data.ID),
				fmt.Sprintf("member_center_history.create.banch.%s", data.BranchID),
				fmt.Sprintf("member_center_history.create.organization.%s", data.OrganizationID),
			}, model.MemberCenterHistoryModel(data)
		},
		func(data *MemberCenterHistory) ([]string, any) {
			return []string{
				"member_center_history.update",
				fmt.Sprintf("member_center_history.update.%s", data.ID),
				fmt.Sprintf("member_center_history.update.banch.%s", data.BranchID),
				fmt.Sprintf("member_center_history.update.organization.%s", data.OrganizationID),
			}, model.MemberCenterHistoryModel(data)
		},
		func(data *MemberCenterHistory) ([]string, any) {
			return []string{
				"member_center_history.delete",
				fmt.Sprintf("member_center_history.delete.%s", data.ID),
				fmt.Sprintf("member_center_history.delete.banch.%s", data.BranchID),
				fmt.Sprintf("member_center_history.delete.organization.%s", data.OrganizationID),
			}, model.MemberCenterHistoryModel(data)
		},
		[]string{
			"CreatedBy",
			"UpdatedBy",
			"Organization",
			"Branch",
		},
	)
	return &MemberCenterHistoryCollection{
		Manager: manager,
	}, nil
}

func (m *Model) MemberCenterHistoryModels(data []*MemberCenterHistory) []*MemberCenterHistoryResponse {
	return horizon_manager.ToModels(data, m.MemberCenterHistoryModel)
}

// member-center-history/branch/:branch_id
func (fc *MemberCenterHistoryCollection) ListByBranch(branchID uuid.UUID) ([]*MemberCenterHistory, error) {
	return fc.Manager.Find(&MemberCenterHistory{
		BranchID: branchID,
	})
}

// member-center-history/organization/:organization_id
func (fc *MemberCenterHistoryCollection) ListByOrganization(organizationID uuid.UUID) ([]*MemberCenterHistory, error) {
	return fc.Manager.Find(&MemberCenterHistory{
		OrganizationID: organizationID,
	})
}

// member-center-history/organization/:organization_id/branch/:branch_id
func (fc *MemberCenterHistoryCollection) ListByOrganizationBranch(organizationID uuid.UUID, branchID uuid.UUID) ([]*MemberCenterHistory, error) {
	return fc.Manager.Find(&MemberCenterHistory{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}
