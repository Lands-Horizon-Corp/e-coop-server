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
	MemberOccupationHistory struct {
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

		OrganizationID uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_member_occupation_history"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_member_occupation_history"`
		Branch         *Branch       `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"branch,omitempty"`

		MemberProfileID uuid.UUID      `gorm:"type:uuid;not null"`
		MemberProfile   *MemberProfile `gorm:"foreignKey:MemberProfileID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"member_profile,omitempty"`

		MemberOccupationID uuid.UUID         `gorm:"type:uuid;not null"`
		MemberOccupation   *MemberOccupation `gorm:"foreignKey:MemberOccupationID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"member_occupation,omitempty"`
	}

	MemberOccupationHistoryResponse struct {
		ID                 uuid.UUID                 `json:"id"`
		CreatedAt          string                    `json:"created_at"`
		CreatedByID        uuid.UUID                 `json:"created_by_id"`
		CreatedBy          *UserResponse             `json:"created_by,omitempty"`
		UpdatedAt          string                    `json:"updated_at"`
		UpdatedByID        uuid.UUID                 `json:"updated_by_id"`
		UpdatedBy          *UserResponse             `json:"updated_by,omitempty"`
		OrganizationID     uuid.UUID                 `json:"organization_id"`
		Organization       *OrganizationResponse     `json:"organization,omitempty"`
		BranchID           uuid.UUID                 `json:"branch_id"`
		Branch             *BranchResponse           `json:"branch,omitempty"`
		MemberProfileID    uuid.UUID                 `json:"member_profile_id"`
		MemberProfile      *MemberProfileResponse    `json:"member_profile,omitempty"`
		MemberOccupationID uuid.UUID                 `json:"member_occupation_id"`
		MemberOccupation   *MemberOccupationResponse `json:"member_occupation,omitempty"`
	}

	MemberOccupationHistoryRequest struct {
		MemberProfileID    uuid.UUID `json:"member_profile_id" validate:"required"`
		MemberOccupationID uuid.UUID `json:"member_occupation_id" validate:"required"`
	}
)

func (m *ModelCore) MemberOccupationHistory() {
	m.Migration = append(m.Migration, &MemberOccupationHistory{})
	m.MemberOccupationHistoryManager = horizon_services.NewRepository(horizon_services.RepositoryParams[MemberOccupationHistory, MemberOccupationHistoryResponse, MemberOccupationHistoryRequest]{
		Preloads: []string{
			"CreatedBy", "UpdatedBy", "Branch", "Organization", "MemberProfile", "MemberOccupation",
		},
		Service: m.provider.Service,
		Resource: func(data *MemberOccupationHistory) *MemberOccupationHistoryResponse {
			if data == nil {
				return nil
			}
			return &MemberOccupationHistoryResponse{
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
				MemberOccupationID: data.MemberOccupationID,
				MemberOccupation:   m.MemberOccupationManager.ToModel(data.MemberOccupation),
			}
		},
		Created: func(data *MemberOccupationHistory) []string {
			return []string{
				"member_occupation_history.create",
				fmt.Sprintf("member_occupation_history.create.%s", data.ID),
				fmt.Sprintf("member_occupation_history.create.branch.%s", data.BranchID),
				fmt.Sprintf("member_occupation_history.create.organization.%s", data.OrganizationID),
				fmt.Sprintf("member_occupation_history.create.member_profile.%s", data.MemberProfileID),
			}
		},
		Updated: func(data *MemberOccupationHistory) []string {
			return []string{
				"member_occupation_history.update",
				fmt.Sprintf("member_occupation_history.update.%s", data.ID),
				fmt.Sprintf("member_occupation_history.update.branch.%s", data.BranchID),
				fmt.Sprintf("member_occupation_history.update.organization.%s", data.OrganizationID),
				fmt.Sprintf("member_occupation_history.update.member_profile.%s", data.MemberProfileID),
			}
		},
		Deleted: func(data *MemberOccupationHistory) []string {
			return []string{
				"member_occupation_history.delete",
				fmt.Sprintf("member_occupation_history.delete.%s", data.ID),
				fmt.Sprintf("member_occupation_history.delete.branch.%s", data.BranchID),
				fmt.Sprintf("member_occupation_history.delete.organization.%s", data.OrganizationID),
				fmt.Sprintf("member_occupation_history.update.member_profile.%s", data.MemberProfileID),
			}
		},
	})
}

func (m *ModelCore) MemberOccupationHistoryCurrentBranch(context context.Context, orgId uuid.UUID, branchId uuid.UUID) ([]*MemberOccupationHistory, error) {
	return m.MemberOccupationHistoryManager.Find(context, &MemberOccupationHistory{
		OrganizationID: orgId,
		BranchID:       branchId,
	})
}
func (m *ModelCore) MemberOccupationHistoryMemberProfileID(context context.Context, memberProfileId, orgId, branchId uuid.UUID) ([]*MemberOccupationHistory, error) {
	return m.MemberOccupationHistoryManager.Find(context, &MemberOccupationHistory{
		OrganizationID:  orgId,
		BranchID:        branchId,
		MemberProfileID: memberProfileId,
	})
}
