package modelcore

import (
	"context"
	"fmt"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/services"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type (
	MemberRelativeAccount struct {
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

		OrganizationID uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_member_relative_account"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_member_relative_account"`
		Branch         *Branch       `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"branch,omitempty"`

		MemberProfileID uuid.UUID      `gorm:"type:uuid;not null"`
		MemberProfile   *MemberProfile `gorm:"foreignKey:MemberProfileID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"member_profile,omitempty"`

		RelativeMemberProfileID uuid.UUID      `gorm:"type:uuid;not null"`
		RelativeMemberProfile   *MemberProfile `gorm:"foreignKey:RelativeMemberProfileID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"relative_member_profile,omitempty"`

		FamilyRelationship string `gorm:"type:varchar(255);not null"` // Enum handled in frontend/validation
		Description        string `gorm:"type:text"`
	}

	MemberRelativeAccountResponse struct {
		ID                      uuid.UUID              `json:"id"`
		CreatedAt               string                 `json:"created_at"`
		CreatedByID             uuid.UUID              `json:"created_by_id"`
		CreatedBy               *UserResponse          `json:"created_by,omitempty"`
		UpdatedAt               string                 `json:"updated_at"`
		UpdatedByID             uuid.UUID              `json:"updated_by_id"`
		UpdatedBy               *UserResponse          `json:"updated_by,omitempty"`
		OrganizationID          uuid.UUID              `json:"organization_id"`
		Organization            *OrganizationResponse  `json:"organization,omitempty"`
		BranchID                uuid.UUID              `json:"branch_id"`
		Branch                  *BranchResponse        `json:"branch,omitempty"`
		MemberProfileID         uuid.UUID              `json:"member_profile_id"`
		MemberProfile           *MemberProfileResponse `json:"member_profile,omitempty"`
		RelativeMemberProfileID uuid.UUID              `json:"relative_member_profile_id"`
		RelativeMemberProfile   *MemberProfileResponse `json:"relative_member_profile,omitempty"`
		FamilyRelationship      string                 `json:"family_relationship"`
		Description             string                 `json:"description"`
	}

	MemberRelativeAccountRequest struct {
		MemberProfileID         uuid.UUID `json:"member_profile_id" validate:"required"`
		RelativeMemberProfileID uuid.UUID `json:"relative_member_profile_id" validate:"required"`
		FamilyRelationship      string    `json:"family_relationship" validate:"required,min=1,max=255"`
		Description             string    `json:"description,omitempty"`
	}
)

func (m *ModelCore) memberRelativeAccount() {
	m.Migration = append(m.Migration, &MemberRelativeAccount{})
	m.MemberRelativeAccountManager = services.NewRepository(services.RepositoryParams[MemberRelativeAccount, MemberRelativeAccountResponse, MemberRelativeAccountRequest]{
		Preloads: []string{"CreatedBy", "UpdatedBy", "MemberProfile", "RelativeMemberProfile"},
		Service:  m.provider.Service,
		Resource: func(data *MemberRelativeAccount) *MemberRelativeAccountResponse {
			if data == nil {
				return nil
			}
			return &MemberRelativeAccountResponse{
				ID:                      data.ID,
				CreatedAt:               data.CreatedAt.Format(time.RFC3339),
				CreatedByID:             data.CreatedByID,
				CreatedBy:               m.UserManager.ToModel(data.CreatedBy),
				UpdatedAt:               data.UpdatedAt.Format(time.RFC3339),
				UpdatedByID:             data.UpdatedByID,
				UpdatedBy:               m.UserManager.ToModel(data.UpdatedBy),
				OrganizationID:          data.OrganizationID,
				Organization:            m.OrganizationManager.ToModel(data.Organization),
				BranchID:                data.BranchID,
				Branch:                  m.BranchManager.ToModel(data.Branch),
				MemberProfileID:         data.MemberProfileID,
				MemberProfile:           m.MemberProfileManager.ToModel(data.MemberProfile),
				RelativeMemberProfileID: data.RelativeMemberProfileID,
				RelativeMemberProfile:   m.MemberProfileManager.ToModel(data.RelativeMemberProfile),
				FamilyRelationship:      data.FamilyRelationship,
				Description:             data.Description,
			}
		},

		Created: func(data *MemberRelativeAccount) []string {
			return []string{
				"member_relative_account.create",
				fmt.Sprintf("member_relative_account.create.%s", data.ID),
				fmt.Sprintf("member_relative_account.create.branch.%s", data.BranchID),
				fmt.Sprintf("member_relative_account.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *MemberRelativeAccount) []string {
			return []string{
				"member_relative_account.update",
				fmt.Sprintf("member_relative_account.update.%s", data.ID),
				fmt.Sprintf("member_relative_account.update.branch.%s", data.BranchID),
				fmt.Sprintf("member_relative_account.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *MemberRelativeAccount) []string {
			return []string{
				"member_relative_account.delete",
				fmt.Sprintf("member_relative_account.delete.%s", data.ID),
				fmt.Sprintf("member_relative_account.delete.branch.%s", data.BranchID),
				fmt.Sprintf("member_relative_account.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func (m *ModelCore) memberRelativeAccountCurrentBranch(context context.Context, orgId uuid.UUID, branchId uuid.UUID) ([]*MemberRelativeAccount, error) {
	return m.MemberRelativeAccountManager.Find(context, &MemberRelativeAccount{
		OrganizationID: orgId,
		BranchID:       branchId,
	})
}
