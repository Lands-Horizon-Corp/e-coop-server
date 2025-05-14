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
	MemberVerification struct {
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
		OrganizationID uuid.UUID      `gorm:"type:uuid;not null;index:idx_branch_org_member_verification"`
		Organization   *Organization  `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID      `gorm:"type:uuid;not null;index:idx_branch_org_member_verification"`
		Branch         *Branch        `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE;" json:"branch,omitempty"`

		MemberProfileID uuid.UUID      `gorm:"type:uuid;not null"`
		MemberProfile   *MemberProfile `gorm:"foreignKey:MemberProfileID;constraint:OnDelete:CASCADE;" json:"member_profile,omitempty"`

		VerifiedByUserID uuid.UUID `gorm:"type:uuid"`
		VerifiedByUser   *User     `gorm:"foreignKey:VerifiedByUserID;constraint:OnDelete:SET NULL;" json:"verified_by_user,omitempty"`

		Status string
	}

	MemberVerificationResponse struct {
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

		// New fields
		MemberProfileID uuid.UUID              `json:"member_profile_id"`
		MemberProfile   *MemberProfileResponse `json:"member_profile,omitempty"`

		VerifiedByUserID uuid.UUID     `json:"verified_by_user_id"`
		VerifiedByUser   *UserResponse `json:"verified_by_user,omitempty"`

		Status string `json:"status"`
	}

	MemberVerificationCollection struct {
		Manager horizon_manager.CollectionManager[MemberVerification]
	}
)

func (m *Model) MemberVerificationModel(data *MemberVerification) *MemberVerificationResponse {
	if data == nil {
		return nil
	}
	return horizon_manager.ToModel(data, func(data *MemberVerification) *MemberVerificationResponse {
		return &MemberVerificationResponse{
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

			MemberProfileID:  data.MemberProfileID,
			MemberProfile:    m.MemberProfileModel(data.MemberProfile),
			VerifiedByUserID: data.VerifiedByUserID,
			VerifiedByUser:   m.UserModel(data.VerifiedByUser),
		}
	})
}

func NewMemberVerificationCollection(
	broadcast *horizon.HorizonBroadcast,
	database *horizon.HorizonDatabase,
	model *Model,
) (*MemberVerificationCollection, error) {
	manager := horizon_manager.NewcollectionManager(
		database,
		broadcast,
		func(data *MemberVerification) ([]string, any) {
			return []string{
				fmt.Sprintf("member_verification.create.%s", data.ID),
				fmt.Sprintf("member_verification.create.banch.%s", data.BranchID),
				fmt.Sprintf("member_verification.create.organization.%s", data.OrganizationID),
			}, model.MemberVerificationModel(data)
		},
		func(data *MemberVerification) ([]string, any) {
			return []string{
				"member_verification.update",
				fmt.Sprintf("member_verification.update.%s", data.ID),
				fmt.Sprintf("member_verification.update.banch.%s", data.BranchID),
				fmt.Sprintf("member_verification.update.organization.%s", data.OrganizationID),
			}, model.MemberVerificationModel(data)
		},
		func(data *MemberVerification) ([]string, any) {
			return []string{
				"member_verification.delete",
				fmt.Sprintf("member_verification.delete.%s", data.ID),
				fmt.Sprintf("member_verification.delete.banch.%s", data.BranchID),
				fmt.Sprintf("member_verification.delete.organization.%s", data.OrganizationID),
			}, model.MemberVerificationModel(data)
		},
		[]string{
			"CreatedBy",
			"UpdatedBy",
			"Organization",
			"Branch",
			"MemberProfile",
			"VerifiedByUser",
		},
	)
	return &MemberVerificationCollection{
		Manager: manager,
	}, nil
}

func (m *Model) MemberVerificationModels(data []*MemberVerification) []*MemberVerificationResponse {
	return horizon_manager.ToModels(data, m.MemberVerificationModel)
}

func (fc *MemberVerificationCollection) GetByMemberProfileID(memberProfileId uuid.UUID) (*MemberVerification, error) {
	return fc.Manager.FindOne(&MemberVerification{
		MemberProfileID: memberProfileId,
	})
}

// member-verification/branch/:branch_id
func (fc *MemberVerificationCollection) ListByBranch(branchID uuid.UUID) ([]*MemberVerification, error) {
	return fc.Manager.Find(&MemberVerification{
		BranchID: branchID,
	})
}

// member-verification/organization/:organization_id
func (fc *MemberVerificationCollection) ListByOrganization(organizationID uuid.UUID) ([]*MemberVerification, error) {
	return fc.Manager.Find(&MemberVerification{
		OrganizationID: organizationID,
	})
}

// member-verification/organization/:organization_id/branch/:branch_id
func (fc *MemberVerificationCollection) ListByOrganizationBranch(organizationID uuid.UUID, branchID uuid.UUID) ([]*MemberVerification, error) {
	return fc.Manager.Find(&MemberVerification{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}
