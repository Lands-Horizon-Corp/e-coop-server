package modelcore

import (
	"context"
	"fmt"
	"time"

	horizon_services "github.com/Lands-Horizon-Corp/e-coop-server/services"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type (
	MemberGroupHistory struct {
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

		OrganizationID uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_member_group_history"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_member_group_history"`
		Branch         *Branch       `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"branch,omitempty"`

		MemberProfileID uuid.UUID      `gorm:"type:uuid;not null"`
		MemberProfile   *MemberProfile `gorm:"foreignKey:MemberProfileID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"member_profile,omitempty"`

		MemberGroupID uuid.UUID    `gorm:"type:uuid;not null"`
		MemberGroup   *MemberGroup `gorm:"foreignKey:MemberGroupID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"member_group,omitempty"`
	}

	MemberGroupHistoryResponse struct {
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
		MemberGroupID   uuid.UUID              `json:"member_group_id"`
		MemberGroup     *MemberGroupResponse   `json:"member_group,omitempty"`
	}

	MemberGroupHistoryRequest struct {
		MemberProfileID uuid.UUID `json:"member_profile_id" validate:"required"`
		MemberGroupID   uuid.UUID `json:"member_group_id" validate:"required"`
	}
)

func (m *ModelCore) MemberGroupHistory() {
	m.Migration = append(m.Migration, &MemberGroupHistory{})
	m.MemberGroupHistoryManager = horizon_services.NewRepository(horizon_services.RepositoryParams[MemberGroupHistory, MemberGroupHistoryResponse, MemberGroupHistoryRequest]{
		Preloads: []string{"CreatedBy", "UpdatedBy", "MemberProfile", "MemberGroup"},
		Service:  m.provider.Service,
		Resource: func(data *MemberGroupHistory) *MemberGroupHistoryResponse {
			if data == nil {
				return nil
			}
			return &MemberGroupHistoryResponse{
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
				MemberGroupID:   data.MemberGroupID,
				MemberGroup:     m.MemberGroupManager.ToModel(data.MemberGroup),
			}
		},

		Created: func(data *MemberGroupHistory) []string {
			return []string{
				"member_group_history.create",
				fmt.Sprintf("member_group_history.create.%s", data.ID),
				fmt.Sprintf("member_group_history.create.branch.%s", data.BranchID),
				fmt.Sprintf("member_group_history.create.organization.%s", data.OrganizationID),
				fmt.Sprintf("member_group_history.create.member_profile.%s", data.MemberProfileID),
			}
		},
		Updated: func(data *MemberGroupHistory) []string {
			return []string{
				"member_group_history.update",
				fmt.Sprintf("member_group_history.update.%s", data.ID),
				fmt.Sprintf("member_group_history.update.branch.%s", data.BranchID),
				fmt.Sprintf("member_group_history.update.organization.%s", data.OrganizationID),
				fmt.Sprintf("member_group_history.update.member_profile.%s", data.MemberProfileID),
			}
		},
		Deleted: func(data *MemberGroupHistory) []string {
			return []string{
				"member_group_history.delete",
				fmt.Sprintf("member_group_history.delete.%s", data.ID),
				fmt.Sprintf("member_group_history.delete.branch.%s", data.BranchID),
				fmt.Sprintf("member_group_history.delete.organization.%s", data.OrganizationID),
				fmt.Sprintf("member_group_history.delete.member_profile.%s", data.MemberProfileID),
			}
		},
	})
}

func (m *ModelCore) MemberGroupHistoryCurrentBranch(context context.Context, orgId uuid.UUID, branchId uuid.UUID) ([]*MemberGroupHistory, error) {
	return m.MemberGroupHistoryManager.Find(context, &MemberGroupHistory{
		OrganizationID: orgId,
		BranchID:       branchId,
	})
}

func (m *ModelCore) MemberGroupHistoryMemberProfileID(context context.Context, memberProfileId, orgId, branchId uuid.UUID) ([]*MemberGroupHistory, error) {
	return m.MemberGroupHistoryManager.Find(context, &MemberGroupHistory{
		OrganizationID:  orgId,
		BranchID:        branchId,
		MemberProfileID: memberProfileId,
	})
}
