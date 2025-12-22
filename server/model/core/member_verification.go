package core

import (
	"context"
	"fmt"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/pkg/registry"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type (
	MemberVerification struct {
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

		OrganizationID uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_member_verification"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"organization,omitempty"`
		BranchID       *uuid.UUID    `gorm:"type:uuid;not null;index:idx_organization_branch_member_verification"`
		Branch         *Branch       `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"branch,omitempty"`

		MemberProfileID *uuid.UUID     `gorm:"type:uuid;not null"`
		MemberProfile   *MemberProfile `gorm:"foreignKey:MemberProfileID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"member_profile,omitempty"`

		VerifiedByUserID *uuid.UUID `gorm:"type:uuid"`
		VerifiedByUser   *User      `gorm:"foreignKey:VerifiedByUserID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"verified_by_user,omitempty"`

		Status string `gorm:"type:varchar(50);not null;default:'pending'"`
	}

	MemberVerificationResponse struct {
		ID               uuid.UUID              `json:"id"`
		CreatedAt        string                 `json:"created_at"`
		CreatedByID      uuid.UUID              `json:"created_by_id"`
		CreatedBy        *UserResponse          `json:"created_by,omitempty"`
		UpdatedAt        string                 `json:"updated_at"`
		UpdatedByID      uuid.UUID              `json:"updated_by_id"`
		UpdatedBy        *UserResponse          `json:"updated_by,omitempty"`
		OrganizationID   uuid.UUID              `json:"organization_id"`
		Organization     *OrganizationResponse  `json:"organization,omitempty"`
		BranchID         uuid.UUID              `json:"branch_id"`
		Branch           *BranchResponse        `json:"branch,omitempty"`
		MemberProfileID  uuid.UUID              `json:"member_profile_id"`
		MemberProfile    *MemberProfileResponse `json:"member_profile,omitempty"`
		VerifiedByUserID uuid.UUID              `json:"verified_by_user_id"`
		VerifiedByUser   *UserResponse          `json:"verified_by_user,omitempty"`
		Status           string                 `json:"status"`
	}

	MemberVerificationRequest struct {
		MemberProfileID  uuid.UUID `json:"member_profile_id" validate:"required"`
		VerifiedByUserID uuid.UUID `json:"verified_by_user_id,omitempty"`
		Status           string    `json:"status,omitempty"`
	}
)

func (m *Core) memberVerification() {
	m.Migration = append(m.Migration, &MemberVerification{})
	m.MemberVerificationManager() = registry.NewRegistry(registry.RegistryParams[MemberVerification, MemberVerificationResponse, MemberVerificationRequest]{
		Preloads: []string{"CreatedBy", "UpdatedBy", "MemberProfile", "VerifiedByUser"},
		Database: m.provider.Service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return m.provider.Service.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *MemberVerification) *MemberVerificationResponse {
			if data == nil {
				return nil
			}
			return &MemberVerificationResponse{
				ID:               data.ID,
				CreatedAt:        data.CreatedAt.Format(time.RFC3339),
				CreatedByID:      data.CreatedByID,
				CreatedBy:        m.UserManager().ToModel(data.CreatedBy),
				UpdatedAt:        data.UpdatedAt.Format(time.RFC3339),
				UpdatedByID:      data.UpdatedByID,
				UpdatedBy:        m.UserManager().ToModel(data.UpdatedBy),
				OrganizationID:   data.OrganizationID,
				Organization:     m.OrganizationManager().ToModel(data.Organization),
				BranchID:         *data.BranchID,
				Branch:           m.BranchManager().ToModel(data.Branch),
				MemberProfileID:  *data.MemberProfileID,
				MemberProfile:    m.MemberProfileManager().ToModel(data.MemberProfile),
				VerifiedByUserID: *data.VerifiedByUserID,
				VerifiedByUser:   m.UserManager().ToModel(data.VerifiedByUser),
				Status:           data.Status,
			}
		},

		Created: func(data *MemberVerification) registry.Topics {
			return []string{
				"member_verification.create",
				fmt.Sprintf("member_verification.create.%s", data.ID),
				fmt.Sprintf("member_verification.create.branch.%s", data.BranchID),
				fmt.Sprintf("member_verification.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *MemberVerification) registry.Topics {
			return []string{
				"member_verification.update",
				fmt.Sprintf("member_verification.update.%s", data.ID),
				fmt.Sprintf("member_verification.update.branch.%s", data.BranchID),
				fmt.Sprintf("member_verification.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *MemberVerification) registry.Topics {
			return []string{
				"member_verification.delete",
				fmt.Sprintf("member_verification.delete.%s", data.ID),
				fmt.Sprintf("member_verification.delete.branch.%s", data.BranchID),
				fmt.Sprintf("member_verification.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func (m *Core) MemberVerificationCurrentBranch(context context.Context, organizationID uuid.UUID, branchID uuid.UUID) ([]*MemberVerification, error) {
	return m.MemberVerificationManager().Find(context, &MemberVerification{
		OrganizationID: organizationID,
		BranchID:       &branchID,
	})
}
