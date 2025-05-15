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
	MemberAsset struct {
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
		OrganizationID uuid.UUID      `gorm:"type:uuid;not null;index:idx_branch_org_member_asset"`
		Organization   *Organization  `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID      `gorm:"type:uuid;not null;index:idx_branch_org_member_asset"`
		Branch         *Branch        `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE;" json:"branch,omitempty"`

		MediaID *uuid.UUID `gorm:"type:uuid"`
		Media   *Media     `gorm:"foreignKey:MediaID;constraint:OnDelete:SET NULL;" json:"media,omitempty"`

		Name        string    `gorm:"type:varchar(255);not null"`
		EntryDate   time.Time `gorm:"type:timestamp;not null"`
		Description string    `gorm:"type:text;not null"`
		Cost        float64   `gorm:"type:decimal;default:0"`

		MemberProfileID *uuid.UUID     `gorm:"type:uuid"`
		MemberProfile   *MemberProfile `gorm:"foreignKey:MemberProfileID;constraint:OnDelete:SET NULL;" json:"member_profile,omitempty"`
	}

	MemberAssetResponse struct {
		ID                     uuid.UUID                     `json:"id"`
		CreatedAt              string                        `json:"created_at"`
		CreatedByID            uuid.UUID                     `json:"created_by_id"`
		CreatedBy              *UserResponse                 `json:"created_by,omitempty"`
		UpdatedAt              string                        `json:"updated_at"`
		UpdatedByID            uuid.UUID                     `json:"updated_by_id"`
		UpdatedBy              *UserResponse                 `json:"updated_by,omitempty"`
		OrganizationID         uuid.UUID                     `json:"organization_id"`
		Organization           *OrganizationResponse         `json:"organization,omitempty"`
		BranchID               uuid.UUID                     `json:"branch_id"`
		Branch                 *BranchResponse               `json:"branch,omitempty"`
		MemberProfileID        uuid.UUID                     `json:"member_profile_id,omitempty"`
		MemberProfile          *MemberProfileResponse        `json:"member_profile,omitempty"`
		MemberClassificationID uuid.UUID                     `json:"member_classification_id,omitempty"`
		MemberClassification   *MemberClassificationResponse `json:"member_classification,omitempty"`

		// New fields
		MediaID     *uuid.UUID     `json:"media_id,omitempty"`
		Media       *MediaResponse `json:"media,omitempty"`
		Name        string         `json:"name"`
		EntryDate   time.Time      `json:"entry_date"`
		Cost        float64        `json:"cost"`
		Description string         `json:"description,omitempty"`
	}

	MemberAssetRequest struct {
		Description string     `json:"description,omitempty" validate:"max=1024"`
		MediaID     *uuid.UUID `json:"media_id,omitempty"`
		Name        string     `json:"name" validate:"required,max=255"`
		EntryDate   time.Time  `json:"entry_date" validate:"required"`
		Cost        float64    `json:"cost" validate:"gte=0"`
	}

	MemberAssetCollection struct {
		Manager horizon_manager.CollectionManager[MemberAsset]
	}
)

func (m *Model) MemberAssetValidate(ctx echo.Context) (*MemberAssetRequest, error) {
	return horizon_manager.Validate[MemberAssetRequest](ctx, m.validator)
}

func (m *Model) MemberAssetModel(data *MemberAsset) *MemberAssetResponse {
	if data == nil {
		return nil
	}
	return horizon_manager.ToModel(data, func(data *MemberAsset) *MemberAssetResponse {
		return &MemberAssetResponse{
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

			Description: data.Description,
		}
	})
}

func (m *Model) MemberAssetModels(data []*MemberAsset) []*MemberAssetResponse {
	return horizon_manager.ToModels(data, m.MemberAssetModel)
}

func NewMemberAssetCollection(
	broadcast *horizon.HorizonBroadcast,
	database *horizon.HorizonDatabase,
	model *Model,
) (*MemberAssetCollection, error) {
	manager := horizon_manager.NewcollectionManager(
		database,
		broadcast,
		func(data *MemberAsset) ([]string, any) {
			return []string{
				fmt.Sprintf("member_asset.create.%s", data.ID),
				fmt.Sprintf("member_asset.create.banch.%s", data.BranchID),
				fmt.Sprintf("member_asset.create.member_profile.%s", data.MemberProfileID),
				fmt.Sprintf("member_asset.create.organization.%s", data.OrganizationID),
			}, model.MemberAssetModel(data)
		},
		func(data *MemberAsset) ([]string, any) {
			return []string{
				"member_asset.update",
				fmt.Sprintf("member_asset.update.%s", data.ID),
				fmt.Sprintf("member_asset.update.banch.%s", data.BranchID),
				fmt.Sprintf("member_asset.update.member_profile.%s", data.MemberProfileID),
				fmt.Sprintf("member_asset.update.organization.%s", data.OrganizationID),
			}, model.MemberAssetModel(data)
		},
		func(data *MemberAsset) ([]string, any) {
			return []string{
				"member_asset.delete",
				fmt.Sprintf("member_asset.delete.%s", data.ID),
				fmt.Sprintf("member_asset.delete.banch.%s", data.BranchID),
				fmt.Sprintf("member_asset.delete.member_profile.%s", data.MemberProfileID),
				fmt.Sprintf("member_asset.delete.organization.%s", data.OrganizationID),
			}, model.MemberAssetModel(data)
		},
		[]string{
			"CreatedBy",
			"UpdatedBy",
			"Organization",
			"Branch",
			"Media",
		},
	)
	return &MemberAssetCollection{
		Manager: manager,
	}, nil
}

// member-asset/member_profile_id
func (fc *MemberAssetCollection) ListByMemberProfile(memberProfileId uuid.UUID) ([]*MemberAsset, error) {
	return fc.Manager.Find(&MemberAsset{
		MemberProfileID: &memberProfileId,
	})
}

// member-asset/branch/:branch_id
func (fc *MemberAssetCollection) ListByBranch(branchID uuid.UUID) ([]*MemberAsset, error) {
	return fc.Manager.Find(&MemberAsset{
		BranchID: branchID,
	})
}

// member-asset/organization/:organization_id
func (fc *MemberAssetCollection) ListByOrganization(organizationID uuid.UUID) ([]*MemberAsset, error) {
	return fc.Manager.Find(&MemberAsset{
		OrganizationID: organizationID,
	})
}

// member-asset/organization/:organization_id/branch/:branch_id
func (fc *MemberAssetCollection) ListByOrganizationBranch(organizationID uuid.UUID, branchID uuid.UUID) ([]*MemberAsset, error) {
	return fc.Manager.Find(&MemberAsset{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}
