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
	// MemberTypeHistory represents the MemberTypeHistory model.
	MemberTypeHistory struct {
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

		OrganizationID uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_member_type_history"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_member_type_history"`
		Branch         *Branch       `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"branch,omitempty"`

		MemberTypeID uuid.UUID   `gorm:"type:uuid;not null"`
		MemberType   *MemberType `gorm:"foreignKey:MemberTypeID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"member_type,omitempty"`

		MemberProfileID uuid.UUID      `gorm:"type:uuid;not null"`
		MemberProfile   *MemberProfile `gorm:"foreignKey:MemberProfileID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"member_profile,omitempty"`
	}

	// MemberTypeHistoryResponse represents the response structure for membertypehistory data

	// MemberTypeHistoryResponse represents the response structure for MemberTypeHistory.
	MemberTypeHistoryResponse struct {
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
		MemberTypeID    uuid.UUID              `json:"member_type_id"`
		MemberType      *MemberTypeResponse    `json:"member_type,omitempty"`
		MemberProfileID uuid.UUID              `json:"member_profile_id"`
		MemberProfile   *MemberProfileResponse `json:"member_profile,omitempty"`
	}

	// MemberTypeHistoryRequest represents the request structure for creating/updating membertypehistory

	// MemberTypeHistoryRequest represents the request structure for MemberTypeHistory.
	MemberTypeHistoryRequest struct {
		MemberTypeID    uuid.UUID `json:"member_type_id" validate:"required"`
		MemberProfileID uuid.UUID `json:"member_profile_id" validate:"required"`
	}
)

func (m *ModelCore) memberTypeHistory() {
	m.Migration = append(m.Migration, &MemberTypeHistory{})
	m.MemberTypeHistoryManager = services.NewRepository(services.RepositoryParams[MemberTypeHistory, MemberTypeHistoryResponse, MemberTypeHistoryRequest]{
		Preloads: []string{"CreatedBy", "UpdatedBy", "MemberType", "MemberProfile"},
		Service:  m.provider.Service,
		Resource: func(data *MemberTypeHistory) *MemberTypeHistoryResponse {
			if data == nil {
				return nil
			}
			return &MemberTypeHistoryResponse{
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
				MemberTypeID:    data.MemberTypeID,
				MemberType:      m.MemberTypeManager.ToModel(data.MemberType),
				MemberProfileID: data.MemberProfileID,
				MemberProfile:   m.MemberProfileManager.ToModel(data.MemberProfile),
			}
		},

		Created: func(data *MemberTypeHistory) []string {
			return []string{
				"member_type_history.create",
				fmt.Sprintf("member_type_history.create.%s", data.ID),
				fmt.Sprintf("member_type_history.create.branch.%s", data.BranchID),
				fmt.Sprintf("member_type_history.create.organization.%s", data.OrganizationID),
				fmt.Sprintf("member_type_history.create.member_profile.%s", data.MemberProfileID),
			}
		},
		Updated: func(data *MemberTypeHistory) []string {
			return []string{
				"member_type_history.update",
				fmt.Sprintf("member_type_history.update.%s", data.ID),
				fmt.Sprintf("member_type_history.update.branch.%s", data.BranchID),
				fmt.Sprintf("member_type_history.update.organization.%s", data.OrganizationID),
				fmt.Sprintf("member_type_history.update.member_profile.%s", data.MemberProfileID),
			}
		},
		Deleted: func(data *MemberTypeHistory) []string {
			return []string{
				"member_type_history.delete",
				fmt.Sprintf("member_type_history.delete.%s", data.ID),
				fmt.Sprintf("member_type_history.delete.branch.%s", data.BranchID),
				fmt.Sprintf("member_type_history.delete.organization.%s", data.OrganizationID),
				fmt.Sprintf("member_type_history.delete.member_profile.%s", data.MemberProfileID),
			}
		},
	})
}

// MemberTypeHistoryCurrentBranch returns MemberTypeHistoryCurrentBranch for the current branch or organization where applicable.
func (m *ModelCore) MemberTypeHistoryCurrentBranch(context context.Context, orgID uuid.UUID, branchID uuid.UUID) ([]*MemberTypeHistory, error) {
	return m.MemberTypeHistoryManager.Find(context, &MemberTypeHistory{
		OrganizationID: orgID,
		BranchID:       branchID,
	})
}

// MemberTypeHistoryMemberProfileID returns MemberTypeHistoryMemberProfileID for the current branch or organization where applicable.
func (m *ModelCore) MemberTypeHistoryMemberProfileID(context context.Context, memberProfileId, orgID, branchID uuid.UUID) ([]*MemberTypeHistory, error) {
	return m.MemberTypeHistoryManager.Find(context, &MemberTypeHistory{
		OrganizationID:  orgID,
		BranchID:        branchID,
		MemberProfileID: memberProfileId,
	})
}
