package model

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"horizon.com/server/horizon"
	horizon_manager "horizon.com/server/horizon/manager"
)

type (
	MemberProfile struct {
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
		OrganizationID uuid.UUID      `gorm:"type:uuid;not null;index:idx_branch_org_member_profile"`
		Organization   *Organization  `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID      `gorm:"type:uuid;not null;index:idx_branch_org_member_profile"`
		Branch         *Branch        `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE;" json:"branch,omitempty"`

		UserID           *uuid.UUID `gorm:"type:uuid"`
		User             *User      `gorm:"foreignKey:UserID;constraint:OnDelete:SET NULL;" json:"user,omitempty"`
		MediaID          *uuid.UUID `gorm:"type:uuid"`
		Media            *Media     `gorm:"foreignKey:MediaID;constraint:OnDelete:SET NULL;" json:"media,omitempty"`
		SignatureMediaID *uuid.UUID `gorm:"type:uuid"`
		SignatureMedia   *Media     `gorm:"foreignKey:SignatureMediaID;constraint:OnDelete:SET NULL;" json:"signature,omitempty"`

		MemberCenterID         *uuid.UUID            `gorm:"type:uuid"`
		MemberCenter           *MemberCenter         `gorm:"foreignKey:MemberCenterID;constraint:OnDelete:SET NULL;" json:"member_center,omitempty"`
		MemberClassificationID *uuid.UUID            `gorm:"type:uuid"`
		MemberClassification   *MemberClassification `gorm:"foreignKey:MemberClassificationID;constraint:OnDelete:SET NULL;" json:"member_classification,omitempty"`
		MemberGenderID         *uuid.UUID            `gorm:"type:uuid"`
		MemberGender           *MemberGender         `gorm:"foreignKey:MemberGenderID;constraint:OnDelete:SET NULL;" json:"member_gender,omitempty"`
		MemberGroupID          *uuid.UUID            `gorm:"type:uuid"`
		MemberGroup            *MemberCenter         `gorm:"foreignKey:MemberGroupID;constraint:OnDelete:SET NULL;" json:"member_group,omitempty"`
		MemberOccupationID     *uuid.UUID            `gorm:"type:uuid"`
		MemberOccupation       *MemberCenter         `gorm:"foreignKey:MemberOccupationID;constraint:OnDelete:SET NULL;" json:"member_occupation,omitempty"`

		IsClosed             bool   `gorm:"not null;default:false"`
		IsMutualFundMember   bool   `gorm:"not null;default:false"`
		IsMicroFinanceMember bool   `gorm:"not null;default:false"`
		FirstName            string `gorm:"type:varchar(255);not null"`
		MiddleName           string `gorm:"type:varchar(255)"`
		LastName             string `gorm:"type:varchar(255);not null"`
		FullName             string `gorm:"type:varchar(255);not null"`
		Suffix               string `gorm:"type:varchar(50)"`
		Birthdate            *time.Time
		Status               string `gorm:"type:general_status;not null;default:'pending'"`

		Description           string `gorm:"type:text"`
		Notes                 string `gorm:"type:text"`
		ContactNumber         string `gorm:"type:varchar(255)"`
		OldReferenceID        string `gorm:"type:varchar(50)"`
		Passbook              string `gorm:"type:varchar(255)"`
		Occupation            string `gorm:"type:varchar(255)"`
		BusinessAddress       string `gorm:"type:varchar(255)"`
		BusinessContactNumber string `gorm:"type:varchar(255)"`
		CivilStatus           string `gorm:"type:civil_status;not null;default:'single'"`
	}

	MemberProfileResponse struct {
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
	}

	MemberProfileCollection struct {
		Manager horizon_manager.CollectionManager[MemberProfile]
	}
)

func (m *Model) MemberProfileModel(data *MemberProfile) *MemberProfileResponse {
	if data == nil {
		return nil
	}
	return horizon_manager.ToModel(data, func(data *MemberProfile) *MemberProfileResponse {
		return &MemberProfileResponse{
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
		}
	})
}

func (m *Model) MemberProfileModels(data []*MemberProfile) []*MemberProfileResponse {
	return horizon_manager.ToModels(data, m.MemberProfileModel)
}

func NewMemberProfileCollection(
	broadcast *horizon.HorizonBroadcast,
	database *horizon.HorizonDatabase,
	model *Model,
) (*MemberProfileCollection, error) {
	manager := horizon_manager.NewcollectionManager(
		database,
		broadcast,
		func(data *MemberProfile) ([]string, any) {
			return []string{
				fmt.Sprintf("member_profile.create.%s", data.ID),
				fmt.Sprintf("member_profile.create.banch.%s", data.BranchID),
				fmt.Sprintf("member_center.create.organization.%s", data.OrganizationID),
			}, model.MemberProfileModel(data)
		},
		func(data *MemberProfile) ([]string, any) {
			return []string{
				"member_profile.update",
				fmt.Sprintf("member_profile.update.%s", data.ID),
				fmt.Sprintf("member_profile.update.banch.%s", data.BranchID),
				fmt.Sprintf("member_profile.update.organization.%s", data.OrganizationID),
			}, model.MemberProfileModel(data)
		},
		func(data *MemberProfile) ([]string, any) {
			return []string{
				"member_profile.delete",
				fmt.Sprintf("member_profile.delete.%s", data.ID),
				fmt.Sprintf("member_profile.delete.banch.%s", data.BranchID),
				fmt.Sprintf("member_profile.delete.organization.%s", data.OrganizationID),
			}, model.MemberProfileModel(data)
		},
		[]string{
			"CreatedBy",
			"UpdatedBy",
			"Organization",
			"Branch",
			"User",
			"Media",
			"SignatureMedia",
			"MemberCenter",
			"MemberClassification",
			"MemberGender",
			"MemberGroup",
			"MemberOccupation",
		},
	)
	return &MemberProfileCollection{
		Manager: manager,
	}, nil
}

// member-group/branch/:branch_id
func (fc *MemberProfileCollection) ListByBranch(branchID uuid.UUID) ([]*MemberProfile, error) {
	return fc.Manager.Find(&MemberProfile{
		BranchID: branchID,
	})
}

// member-group/organization/:organization_id
func (fc *MemberProfileCollection) ListByOrganization(organizationID uuid.UUID) ([]*MemberProfile, error) {
	return fc.Manager.Find(&MemberProfile{
		OrganizationID: organizationID,
	})
}

// member-group/organization/:organization_id/branch/:branch_id
func (fc *MemberProfileCollection) ListByOrganizationBranch(organizationID uuid.UUID, branchID uuid.UUID) ([]*MemberProfile, error) {
	return fc.Manager.Find(&MemberProfile{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}
