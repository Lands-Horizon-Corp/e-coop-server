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
	MemberOccupationHistory struct {
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
		OrganizationID uuid.UUID      `gorm:"type:uuid;not null;index:idx_branch_org_member_occupation_history"`
		Organization   *Organization  `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID      `gorm:"type:uuid;not null;index:idx_branch_org_member_occupation_history"`
		Branch         *Branch        `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE;" json:"branch,omitempty"`

		MemberProfileID *uuid.UUID     `gorm:"type:uuid"`
		MemberProfile   *MemberProfile `gorm:"foreignKey:MemberProfileID;constraint:OnDelete:SET NULL;" json:"member_profile,omitempty"`

		MemberOccupationID *uuid.UUID        `gorm:"type:uuid"`
		MemberOccupation   *MemberOccupation `gorm:"foreignKey:MemberOccupationID;constraint:OnDelete:SET NULL;" json:"member_occupation,omitempty"`
	}

	MemberOccupationHistoryResponse struct {
		ID                 uuid.UUID                 `json:"id"`
		CreatedAt          string                    `json:"created_at"`
		CreatedByID        uuid.UUID                 `json:"created_by_id"`
		CreatedBy          *UserResponse             `json:"created_by,omitempty"`
		UpdatedAt          string                    `json:"updated_at"`
		UpdatedByID        uuid.UUID                 `json:"updated_by_id"`
		UpdatedBy          *UserResponse             `json:"updated_by,omitempty"`
		OrganizationID     uuid.UUID                 `json:"organization_id"`
		Organization       *OrganizationResponse     `json:"organization,omitempty"`
		BranchID           uuid.UUID                 `json:"branch_id"`
		Branch             *BranchResponse           `json:"branch,omitempty"`
		MemberProfileID    uuid.UUID                 `json:"member_profile_id,omitempty"`
		MemberProfile      *MemberProfileResponse    `json:"member_profile,omitempty"`
		MemberOccupationID uuid.UUID                 `json:"member_occupation_id,omitempty"`
		MemberOccupation   *MemberOccupationResponse `json:"member_occupation,omitempty"`
	}

	MemberOccupationHistoryRequest struct {
		Name        string `json:"name,omitempty" validate:"required,max=255"`
		Description string `json:"description,omitempty" validate:"max=1024"`
	}

	MemberOccupationHistoryCollection struct {
		Manager horizon_manager.CollectionManager[MemberOccupationHistory]
	}
)

func (m *Model) MemberOccupationHistoryValidate(ctx echo.Context) (*MemberOccupationHistoryRequest, error) {
	return horizon_manager.Validate[MemberOccupationHistoryRequest](ctx, m.validator)
}

func (m *Model) MemberOccupationHistoryModel(data *MemberOccupationHistory) *MemberOccupationHistoryResponse {
	if data == nil {
		return nil
	}
	return horizon_manager.ToModel(data, func(data *MemberOccupationHistory) *MemberOccupationHistoryResponse {
		return &MemberOccupationHistoryResponse{
			ID:                 data.ID,
			CreatedAt:          data.CreatedAt.Format(time.RFC3339),
			CreatedByID:        data.CreatedByID,
			CreatedBy:          m.UserModel(data.CreatedBy),
			UpdatedAt:          data.UpdatedAt.Format(time.RFC3339),
			UpdatedByID:        data.UpdatedByID,
			UpdatedBy:          m.UserModel(data.UpdatedBy),
			OrganizationID:     data.OrganizationID,
			Organization:       m.OrganizationModel(data.Organization),
			BranchID:           data.BranchID,
			Branch:             m.BranchModel(data.Branch),
			MemberProfileID:    *data.MemberProfileID,
			MemberProfile:      m.MemberProfileModel(data.MemberProfile),
			MemberOccupationID: *data.MemberOccupationID,
			MemberOccupation:   m.MemberOccupationModel(data.MemberOccupation),
		}
	})
}

func NewMemberOccupationHistoryCollection(
	broadcast *horizon.HorizonBroadcast,
	database *horizon.HorizonDatabase,
	model *Model,
) (*MemberOccupationHistoryCollection, error) {
	manager := horizon_manager.NewcollectionManager(
		database,
		broadcast,
		func(data *MemberOccupationHistory) ([]string, any) {
			return []string{
				fmt.Sprintf("member_occupation_history.create.%s", data.ID),
				fmt.Sprintf("member_occupation_history.create.banch.%s", data.BranchID),
				fmt.Sprintf("member_occupation_history.create.organization.%s", data.OrganizationID),
			}, model.MemberOccupationHistoryModel(data)
		},
		func(data *MemberOccupationHistory) ([]string, any) {
			return []string{
				"member_occupation_history.update",
				fmt.Sprintf("member_occupation_history.update.%s", data.ID),
				fmt.Sprintf("member_occupation_history.update.banch.%s", data.BranchID),
				fmt.Sprintf("member_occupation_history.update.organization.%s", data.OrganizationID),
			}, model.MemberOccupationHistoryModel(data)
		},
		func(data *MemberOccupationHistory) ([]string, any) {
			return []string{
				"member_occupation_history.delete",
				fmt.Sprintf("member_occupation_history.delete.%s", data.ID),
				fmt.Sprintf("member_occupation_history.delete.banch.%s", data.BranchID),
				fmt.Sprintf("member_occupation_history.delete.organization.%s", data.OrganizationID),
			}, model.MemberOccupationHistoryModel(data)
		},
		[]string{
			"CreatedBy",
			"UpdatedBy",
			"Organization",
			"Branch",
		},
	)
	return &MemberOccupationHistoryCollection{
		Manager: manager,
	}, nil
}

func (m *Model) MemberOccupationHistoryModels(data []*MemberOccupationHistory) []*MemberOccupationHistoryResponse {
	return horizon_manager.ToModels(data, m.MemberOccupationHistoryModel)
}

// member-occupation-history/branch/:branch_id
func (fc *MemberOccupationHistoryCollection) ListByBranch(branchID uuid.UUID) ([]*MemberOccupationHistory, error) {
	return fc.Manager.Find(&MemberOccupationHistory{
		BranchID: branchID,
	})
}

// member-occupation-history/organization/:organization_id
func (fc *MemberOccupationHistoryCollection) ListByOrganization(organizationID uuid.UUID) ([]*MemberOccupationHistory, error) {
	return fc.Manager.Find(&MemberOccupationHistory{
		OrganizationID: organizationID,
	})
}

// member-occupation-history/organization/:organization_id/branch/:branch_id
func (fc *MemberOccupationHistoryCollection) ListByOrganizationBranch(organizationID uuid.UUID, branchID uuid.UUID) ([]*MemberOccupationHistory, error) {
	return fc.Manager.Find(&MemberOccupationHistory{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}
