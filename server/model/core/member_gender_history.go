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
	MemberGenderHistory struct {
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

		OrganizationID uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_member_gender_history"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_member_gender_history"`
		Branch         *Branch       `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"branch,omitempty"`

		MemberProfileID uuid.UUID      `gorm:"type:uuid;not null"`
		MemberProfile   *MemberProfile `gorm:"foreignKey:MemberProfileID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"member_profile,omitempty"`

		MemberGenderID uuid.UUID     `gorm:"type:uuid;not null"`
		MemberGender   *MemberGender `gorm:"foreignKey:MemberGenderID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"member_gender,omitempty"`
	}

	MemberGenderHistoryResponse struct {
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
		MemberProfileID uuid.UUID              `json:"member_profile_id"`
		MemberProfile   *MemberProfileResponse `json:"member_profile,omitempty"`
		MemberGenderID  uuid.UUID              `json:"member_gender_id"`
		MemberGender    *MemberGenderResponse  `json:"member_gender,omitempty"`
	}

	MemberGenderHistoryRequest struct {
		MemberProfileID uuid.UUID `json:"member_profile_id" validate:"required"`
		MemberGenderID  uuid.UUID `json:"member_gender_id" validate:"required"`
	}
)

func (m *Core) memberGenderHistory() {
	m.Migration = append(m.Migration, &MemberGenderHistory{})
	m.MemberGenderHistoryManager = registry.NewRegistry(registry.RegistryParams[MemberGenderHistory, MemberGenderHistoryResponse, MemberGenderHistoryRequest]{
		Preloads: []string{"CreatedBy", "UpdatedBy", "MemberProfile", "MemberGender"},
		Database: m.provider.Service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return m.provider.Service.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *MemberGenderHistory) *MemberGenderHistoryResponse {
			if data == nil {
				return nil
			}
			return &MemberGenderHistoryResponse{
				ID:              data.ID,
				CreatedAt:       data.CreatedAt.Format(time.RFC3339),
				CreatedByID:     data.CreatedByID,
				CreatedBy:       m.UserManager.ToModel(data.CreatedBy),
				UpdatedAt:       data.UpdatedAt.Format(time.RFC3339),
				UpdatedByID:     data.UpdatedByID,
				UpdatedBy:       m.UserManager.ToModel(data.UpdatedBy),
				OrganizationID:  data.OrganizationID,
				Organization:    m.OrganizationManager.ToModel(data.Organization),
				BranchID:        data.BranchID,
				Branch:          m.BranchManager.ToModel(data.Branch),
				MemberProfileID: data.MemberProfileID,
				MemberProfile:   m.MemberProfileManager.ToModel(data.MemberProfile),
				MemberGenderID:  data.MemberGenderID,
				MemberGender:    m.MemberGenderManager.ToModel(data.MemberGender),
			}
		},

		Created: func(data *MemberGenderHistory) registry.Topics {
			return []string{
				"member_gender_history.create",
				fmt.Sprintf("member_gender_history.create.%s", data.ID),
				fmt.Sprintf("member_gender_history.create.branch.%s", data.BranchID),
				fmt.Sprintf("member_gender_history.create.organization.%s", data.OrganizationID),
				fmt.Sprintf("member_gender_history.create.member_profile.%s", data.MemberProfileID),
			}
		},
		Updated: func(data *MemberGenderHistory) registry.Topics {
			return []string{
				"member_gender_history.update",
				fmt.Sprintf("member_gender_history.update.%s", data.ID),
				fmt.Sprintf("member_gender_history.update.branch.%s", data.BranchID),
				fmt.Sprintf("member_gender_history.update.organization.%s", data.OrganizationID),
				fmt.Sprintf("member_gender_history.update.member_profile.%s", data.MemberProfileID),
			}
		},
		Deleted: func(data *MemberGenderHistory) registry.Topics {
			return []string{
				"member_gender_history.delete",
				fmt.Sprintf("member_gender_history.delete.%s", data.ID),
				fmt.Sprintf("member_gender_history.delete.branch.%s", data.BranchID),
				fmt.Sprintf("member_gender_history.delete.organization.%s", data.OrganizationID),
				fmt.Sprintf("member_gender_history.delete.member_profile.%s", data.MemberProfileID),
			}
		},
	})
}

func (m *Core) MemberGenderHistoryCurrentBranch(context context.Context, organizationID uuid.UUID, branchID uuid.UUID) ([]*MemberGenderHistory, error) {
	return m.MemberGenderHistoryManager.Find(context, &MemberGenderHistory{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}

func (m *Core) MemberGenderHistoryMemberProfileID(context context.Context, memberProfileID, organizationID, branchID uuid.UUID) ([]*MemberGenderHistory, error) {
	return m.MemberGenderHistoryManager.Find(context, &MemberGenderHistory{
		OrganizationID:  organizationID,
		BranchID:        branchID,
		MemberProfileID: memberProfileID,
	})
}
