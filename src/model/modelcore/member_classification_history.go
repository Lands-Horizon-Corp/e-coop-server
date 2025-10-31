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
	MemberClassificationHistory struct {
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

		OrganizationID uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_member_classification_history"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_member_classification_history"`
		Branch         *Branch       `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"branch,omitempty"`

		MemberProfileID uuid.UUID      `gorm:"type:uuid;not null"`
		MemberProfile   *MemberProfile `gorm:"foreignKey:MemberProfileID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"member_profile,omitempty"`

		MemberClassificationID uuid.UUID             `gorm:"type:uuid;not null"`
		MemberClassification   *MemberClassification `gorm:"foreignKey:MemberClassificationID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"member_classification,omitempty"`
	}

	MemberClassificationHistoryResponse struct {
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

		MemberClassificationID uuid.UUID                     `json:"member_classification_id"`
		MemberClassification   *MemberClassificationResponse `json:"member_classification,omitempty"`

		MemberProfileID uuid.UUID              `json:"member_profile_id"`
		MemberProfile   *MemberProfileResponse `json:"member_profile,omitempty"`
	}

	MemberClassificationHistoryRequest struct {
		MemberClassificationID uuid.UUID `json:"member_classification_id" validate:"required"`
		MemberProfileID        uuid.UUID `json:"member_profile_id" validate:"required"`
		BranchID               uuid.UUID `json:"branch_id" validate:"required"`
		OrganizationID         uuid.UUID `json:"organization_id" validate:"required"`
	}
)

func (m *modelcore) MemberClassificationHistory() {
	m.Migration = append(m.Migration, &MemberClassificationHistory{})
	m.MemberClassificationHistoryManager = horizon_services.NewRepository(horizon_services.RepositoryParams[
		MemberClassificationHistory,
		MemberClassificationHistoryResponse,
		MemberClassificationHistoryRequest,
	]{
		Preloads: []string{
			"CreatedBy", "UpdatedBy",
			"Organization", "Branch", "MemberClassification", "MemberProfile",
		},
		Service: m.provider.Service,
		Resource: func(data *MemberClassificationHistory) *MemberClassificationHistoryResponse {
			if data == nil {
				return nil
			}
			return &MemberClassificationHistoryResponse{
				ID:                     data.ID,
				CreatedAt:              data.CreatedAt.Format(time.RFC3339),
				CreatedByID:            data.CreatedByID,
				CreatedBy:              m.UserManager.ToModel(data.CreatedBy),
				UpdatedAt:              data.UpdatedAt.Format(time.RFC3339),
				UpdatedByID:            data.UpdatedByID,
				UpdatedBy:              m.UserManager.ToModel(data.UpdatedBy),
				OrganizationID:         data.OrganizationID,
				Organization:           m.OrganizationManager.ToModel(data.Organization),
				BranchID:               data.BranchID,
				Branch:                 m.BranchManager.ToModel(data.Branch),
				MemberClassificationID: data.MemberClassificationID,
				MemberClassification:   m.MemberClassificationManager.ToModel(data.MemberClassification),
				MemberProfileID:        data.MemberProfileID,
				MemberProfile:          m.MemberProfileManager.ToModel(data.MemberProfile),
			}
		},
		Created: func(data *MemberClassificationHistory) []string {
			return []string{
				"member_classification_history.create",
				fmt.Sprintf("member_classification_history.create.%s", data.ID),
				fmt.Sprintf("member_classification_history.create.branch.%s", data.BranchID),
				fmt.Sprintf("member_classification_history.create.organization.%s", data.OrganizationID),
				fmt.Sprintf("member_classification_history.create.member_profile.%s", data.MemberProfileID),
			}
		},
		Updated: func(data *MemberClassificationHistory) []string {
			return []string{
				"member_classification_history.update",
				fmt.Sprintf("member_classification_history.update.%s", data.ID),
				fmt.Sprintf("member_classification_history.update.branch.%s", data.BranchID),
				fmt.Sprintf("member_classification_history.update.organization.%s", data.OrganizationID),
				fmt.Sprintf("member_classification_history.update.member_profile.%s", data.MemberProfileID),
			}
		},
		Deleted: func(data *MemberClassificationHistory) []string {
			return []string{
				"member_classification_history.delete",
				fmt.Sprintf("member_classification_history.delete.%s", data.ID),
				fmt.Sprintf("member_classification_history.delete.branch.%s", data.BranchID),
				fmt.Sprintf("member_classification_history.delete.organization.%s", data.OrganizationID),
				fmt.Sprintf("member_classification_history.delete.member_profile.%s", data.MemberProfileID),
			}
		},
	})
}

func (m *modelcore) MemberClassificationHistoryCurrentBranch(context context.Context, orgId uuid.UUID, branchId uuid.UUID) ([]*MemberClassificationHistory, error) {
	return m.MemberClassificationHistoryManager.Find(context, &MemberClassificationHistory{
		OrganizationID: orgId,
		BranchID:       branchId,
	})
}

func (m *modelcore) MemberClassificationHistoryMemberProfileID(context context.Context, memberProfileId, orgId, branchId uuid.UUID) ([]*MemberClassificationHistory, error) {
	return m.MemberClassificationHistoryManager.Find(context, &MemberClassificationHistory{
		OrganizationID:  orgId,
		BranchID:        branchId,
		MemberProfileID: memberProfileId,
	})
}
