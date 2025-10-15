package model_core

import (
	"context"
	"fmt"
	"time"

	horizon_services "github.com/Lands-Horizon-Corp/e-coop-server/services"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type (
	MemberJointAccount struct {
		ID          uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
		CreatedAt   time.Time      `gorm:"not null;default:now()"`
		CreatedByID uuid.UUID      `gorm:"type:uuid"`
		CreatedBy   *User          `gorm:"foreignKey:CreatedByID;constraint:OnDelete:SET NULL;" json:"created_by,omitempty"`
		UpdatedAt   time.Time      `gorm:"not null;default:now()"`
		UpdatedByID uuid.UUID      `gorm:"type:uuid"`
		UpdatedBy   *User          `gorm:"foreignKey:UpdatedByID;constraint:OnDelete:SET NULL;" json:"updated_by,omitempty"`
		DeletedAt   gorm.DeletedAt `gorm:"index"`
		DeletedByID *uuid.UUID     `gorm:"type:uuid"`
		DeletedBy   *User          `gorm:"foreignKey:DeletedByID;constraint:OnDelete:SET NULL;" json:"deleted_by,omitempty"`

		OrganizationID uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_member_join_account"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_member_join_account"`
		Branch         *Branch       `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"branch,omitempty"`

		MemberProfileID uuid.UUID      `gorm:"type:uuid;not null"`
		MemberProfile   *MemberProfile `gorm:"foreignKey:MemberProfileID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"member_profile,omitempty"`

		PictureMediaID uuid.UUID `gorm:"type:uuid;not null"`
		PictureMedia   *Media    `gorm:"foreignKey:PictureMediaID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"picture_media,omitempty"`

		SignatureMediaID uuid.UUID `gorm:"type:uuid;not null"`
		SignatureMedia   *Media    `gorm:"foreignKey:SignatureMediaID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"signature_media,omitempty"`

		Description        string    `gorm:"type:text"`
		FirstName          string    `gorm:"type:varchar(255);not null"`
		MiddleName         string    `gorm:"type:varchar(255)"`
		LastName           string    `gorm:"type:varchar(255);not null"`
		FullName           string    `gorm:"type:varchar(255);not null"`
		Suffix             string    `gorm:"type:varchar(255)"`
		Birthday           time.Time `gorm:"not null"`
		FamilyRelationship string    `gorm:"type:varchar(255);not null"` // Enum handled on frontend/validation
	}

	MemberJointAccountResponse struct {
		ID                 uuid.UUID              `json:"id"`
		CreatedAt          string                 `json:"created_at"`
		CreatedByID        uuid.UUID              `json:"created_by_id"`
		CreatedBy          *UserResponse          `json:"created_by,omitempty"`
		UpdatedAt          string                 `json:"updated_at"`
		UpdatedByID        uuid.UUID              `json:"updated_by_id"`
		UpdatedBy          *UserResponse          `json:"updated_by,omitempty"`
		OrganizationID     uuid.UUID              `json:"organization_id"`
		Organization       *OrganizationResponse  `json:"organization,omitempty"`
		BranchID           uuid.UUID              `json:"branch_id"`
		Branch             *BranchResponse        `json:"branch,omitempty"`
		MemberProfileID    uuid.UUID              `json:"member_profile_id"`
		MemberProfile      *MemberProfileResponse `json:"member_profile,omitempty"`
		PictureMediaID     uuid.UUID              `json:"picture_media_id"`
		PictureMedia       *MediaResponse         `json:"picture_media,omitempty"`
		SignatureMediaID   uuid.UUID              `json:"signature_media_id"`
		SignatureMedia     *MediaResponse         `json:"signature_media,omitempty"`
		Description        string                 `json:"description"`
		FirstName          string                 `json:"first_name"`
		MiddleName         string                 `json:"middle_name"`
		LastName           string                 `json:"last_name"`
		FullName           string                 `json:"full_name"`
		Suffix             string                 `json:"suffix"`
		Birthday           string                 `json:"birthday"`
		FamilyRelationship string                 `json:"family_relationship"`
	}

	MemberJointAccountRequest struct {
		PictureMediaID     uuid.UUID `json:"picture_media_id" validate:"required"`
		SignatureMediaID   uuid.UUID `json:"signature_media_id" validate:"required"`
		Description        string    `json:"description,omitempty"`
		FirstName          string    `json:"first_name" validate:"required,min=1,max=255"`
		MiddleName         string    `json:"middle_name,omitempty"`
		LastName           string    `json:"last_name" validate:"required,min=1,max=255"`
		FullName           string    `json:"full_name" validate:"required,min=1,max=255"`
		Suffix             string    `json:"suffix,omitempty"`
		Birthday           time.Time `json:"birthday" validate:"required"`
		FamilyRelationship string    `json:"family_relationship" validate:"required,min=1,max=255"`
	}
)

func (m *ModelCore) MemberJointAccount() {
	m.Migration = append(m.Migration, &MemberJointAccount{})
	m.MemberJointAccountManager = horizon_services.NewRepository(horizon_services.RepositoryParams[MemberJointAccount, MemberJointAccountResponse, MemberJointAccountRequest]{
		Preloads: []string{
			"CreatedBy", "UpdatedBy", "Branch", "Organization",
			"MemberProfile", "PictureMedia", "SignatureMedia",
		},
		Service: m.provider.Service,
		Resource: func(data *MemberJointAccount) *MemberJointAccountResponse {
			if data == nil {
				return nil
			}
			return &MemberJointAccountResponse{
				ID:                 data.ID,
				CreatedAt:          data.CreatedAt.Format(time.RFC3339),
				CreatedByID:        data.CreatedByID,
				CreatedBy:          m.UserManager.ToModel(data.CreatedBy),
				UpdatedAt:          data.UpdatedAt.Format(time.RFC3339),
				UpdatedByID:        data.UpdatedByID,
				UpdatedBy:          m.UserManager.ToModel(data.UpdatedBy),
				OrganizationID:     data.OrganizationID,
				Organization:       m.OrganizationManager.ToModel(data.Organization),
				BranchID:           data.BranchID,
				Branch:             m.BranchManager.ToModel(data.Branch),
				MemberProfileID:    data.MemberProfileID,
				MemberProfile:      m.MemberProfileManager.ToModel(data.MemberProfile),
				PictureMediaID:     data.PictureMediaID,
				PictureMedia:       m.MediaManager.ToModel(data.PictureMedia),
				SignatureMediaID:   data.SignatureMediaID,
				SignatureMedia:     m.MediaManager.ToModel(data.SignatureMedia),
				Description:        data.Description,
				FirstName:          data.FirstName,
				MiddleName:         data.MiddleName,
				LastName:           data.LastName,
				FullName:           data.FullName,
				Suffix:             data.Suffix,
				Birthday:           data.Birthday.Format(time.RFC3339),
				FamilyRelationship: data.FamilyRelationship,
			}
		},

		Created: func(data *MemberJointAccount) []string {
			return []string{
				"member_joint_account.create",
				fmt.Sprintf("member_joint_account.create.%s", data.ID),
				fmt.Sprintf("member_joint_account.create.branch.%s", data.BranchID),
				fmt.Sprintf("member_joint_account.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *MemberJointAccount) []string {
			return []string{
				"member_joint_account.update",
				fmt.Sprintf("member_joint_account.update.%s", data.ID),
				fmt.Sprintf("member_joint_account.update.branch.%s", data.BranchID),
				fmt.Sprintf("member_joint_account.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *MemberJointAccount) []string {
			return []string{
				"member_joint_account.delete",
				fmt.Sprintf("member_joint_account.delete.%s", data.ID),
				fmt.Sprintf("member_joint_account.delete.branch.%s", data.BranchID),
				fmt.Sprintf("member_joint_account.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func (m *ModelCore) MemberJointAccountCurrentBranch(context context.Context, orgId uuid.UUID, branchId uuid.UUID) ([]*MemberJointAccount, error) {
	return m.MemberJointAccountManager.Find(context, &MemberJointAccount{
		OrganizationID: orgId,
		BranchID:       branchId,
	})
}
