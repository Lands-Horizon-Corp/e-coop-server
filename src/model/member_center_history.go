package model

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	horizon_services "github.com/lands-horizon/horizon-server/services"
	"gorm.io/gorm"
)

type (
	MemberCenterHistory struct {
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

		OrganizationID uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_member_center_history"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_member_center_history"`
		Branch         *Branch       `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"branch,omitempty"`

		MemberCenterID uuid.UUID     `gorm:"type:uuid;not null"`
		MemberCenter   *MemberCenter `gorm:"foreignKey:MemberCenterID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"member_center,omitempty"`

		MemberProfileID uuid.UUID      `gorm:"type:uuid;not null"`
		MemberProfile   *MemberProfile `gorm:"foreignKey:MemberProfileID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"member_profile,omitempty"`
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
		MemberCenterID  uuid.UUID              `json:"member_center_id"`
		MemberCenter    *MemberCenterResponse  `json:"member_center,omitempty"`
		MemberProfileID uuid.UUID              `json:"member_profile_id"`
		MemberProfile   *MemberProfileResponse `json:"member_profile,omitempty"`
	}

	MemberCenterHistoryRequest struct {
		MemberCenterID  uuid.UUID `json:"member_center_id" validate:"required"`
		MemberProfileID uuid.UUID `json:"member_profile_id" validate:"required"`
	}
)

func (m *Model) MemberCenterHistory() {
	m.Migration = append(m.Migration, &MemberCenterHistory{})
	m.MemberCenterHistoryManager = horizon_services.NewRepository(horizon_services.RepositoryParams[MemberCenterHistory, MemberCenterHistoryResponse, MemberCenterHistoryRequest]{
		Preloads: []string{"CreatedBy", "UpdatedBy", "Branch", "Organization", "MemberCenter", "MemberProfile"},
		Service:  m.provider.Service,
		Resource: func(data *MemberCenterHistory) *MemberCenterHistoryResponse {
			if data == nil {
				return nil
			}
			return &MemberCenterHistoryResponse{
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
				MemberCenterID:  data.MemberCenterID,
				MemberCenter:    m.MemberCenterManager.ToModel(data.MemberCenter),
				MemberProfileID: data.MemberProfileID,
				MemberProfile:   m.MemberProfileManager.ToModel(data.MemberProfile),
			}
		},

		Created: func(data *MemberCenterHistory) []string {
			return []string{
				"member_center_history.create",
				fmt.Sprintf("member_center_history.create.%s", data.ID),
				fmt.Sprintf("member_center_history.create.branch.%s", data.BranchID),
				fmt.Sprintf("member_center_history.create.organization.%s", data.OrganizationID),
				fmt.Sprintf("member_center_history.create.member_profile.%s", data.MemberProfileID),
			}
		},
		Updated: func(data *MemberCenterHistory) []string {
			return []string{
				"member_center_history.update",
				fmt.Sprintf("member_center_history.update.%s", data.ID),
				fmt.Sprintf("member_center_history.update.branch.%s", data.BranchID),
				fmt.Sprintf("member_center_history.update.organization.%s", data.OrganizationID),
				fmt.Sprintf("member_center_history.update.member_profile.%s", data.MemberProfileID),
			}
		},
		Deleted: func(data *MemberCenterHistory) []string {
			return []string{
				"member_center_history.delete",
				fmt.Sprintf("member_center_history.delete.%s", data.ID),
				fmt.Sprintf("member_center_history.delete.branch.%s", data.BranchID),
				fmt.Sprintf("member_center_history.delete.organization.%s", data.OrganizationID),
				fmt.Sprintf("member_center_history.delete.member_profile.%s", data.MemberProfileID),
			}
		},
	})
}

func (m *Model) MemberCenterHistoryCurrentBranch(context context.Context, orgId uuid.UUID, branchId uuid.UUID) ([]*MemberCenterHistory, error) {
	return m.MemberCenterHistoryManager.Find(context, &MemberCenterHistory{
		OrganizationID: orgId,
		BranchID:       branchId,
	})
}

func (m *Model) MemberCenterHistoryMemberProfileID(context context.Context, memberProfileId, orgId, branchId uuid.UUID) ([]*MemberCenterHistory, error) {
	return m.MemberCenterHistoryManager.Find(context, &MemberCenterHistory{
		OrganizationID:  orgId,
		BranchID:        branchId,
		MemberProfileID: memberProfileId,
	})
}
