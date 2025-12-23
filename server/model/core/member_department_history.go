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
	MemberDepartmentHistory struct {
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

		OrganizationID uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_member_department_history"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_member_department_history"`
		Branch         *Branch       `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"branch,omitempty"`

		MemberProfileID uuid.UUID      `gorm:"type:uuid;not null"`
		MemberProfile   *MemberProfile `gorm:"foreignKey:MemberProfileID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"member_profile,omitempty"`

		MemberDepartmentID uuid.UUID         `gorm:"type:uuid;not null"`
		MemberDepartment   *MemberDepartment `gorm:"foreignKey:MemberDepartmentID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"member_department,omitempty"`
	}

	MemberDepartmentHistoryResponse struct {
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

		MemberDepartmentID uuid.UUID                 `json:"member_department_id"`
		MemberDepartment   *MemberDepartmentResponse `json:"member_department,omitempty"`

		MemberProfileID uuid.UUID              `json:"member_profile_id"`
		MemberProfile   *MemberProfileResponse `json:"member_profile,omitempty"`
	}

	MemberDepartmentHistoryRequest struct {
		MemberDepartmentID uuid.UUID `json:"member_department_id" validate:"required"`
		MemberProfileID    uuid.UUID `json:"member_profile_id" validate:"required"`
		BranchID           uuid.UUID `json:"branch_id" validate:"required"`
		OrganizationID     uuid.UUID `json:"organization_id" validate:"required"`
	}
)

func (m *Core) MemberDepartmentHistoryManager() *registry.Registry[MemberDepartmentHistory, MemberDepartmentHistoryResponse, MemberDepartmentHistoryRequest] {
	return registry.NewRegistry(registry.RegistryParams[
		MemberDepartmentHistory,
		MemberDepartmentHistoryResponse,
		MemberDepartmentHistoryRequest,
	]{
		Preloads: []string{
			"CreatedBy", "UpdatedBy",
			"Organization", "Branch", "MemberDepartment", "MemberProfile",
		},
		Database: m.provider.Service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return m.provider.Service.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *MemberDepartmentHistory) *MemberDepartmentHistoryResponse {
			if data == nil {
				return nil
			}
			return &MemberDepartmentHistoryResponse{
				ID:                 data.ID,
				CreatedAt:          data.CreatedAt.Format(time.RFC3339),
				CreatedByID:        data.CreatedByID,
				CreatedBy:          m.UserManager().ToModel(data.CreatedBy),
				UpdatedAt:          data.UpdatedAt.Format(time.RFC3339),
				UpdatedByID:        data.UpdatedByID,
				UpdatedBy:          m.UserManager().ToModel(data.UpdatedBy),
				OrganizationID:     data.OrganizationID,
				Organization:       m.OrganizationManager().ToModel(data.Organization),
				BranchID:           data.BranchID,
				Branch:             m.BranchManager().ToModel(data.Branch),
				MemberDepartmentID: data.MemberDepartmentID,
				MemberDepartment:   m.MemberDepartmentManager().ToModel(data.MemberDepartment),
				MemberProfileID:    data.MemberProfileID,
				MemberProfile:      m.MemberProfileManager().ToModel(data.MemberProfile),
			}
		},
		Created: func(data *MemberDepartmentHistory) registry.Topics {
			return []string{
				"member_department_history.create",
				fmt.Sprintf("member_department_history.create.%s", data.ID),
				fmt.Sprintf("member_department_history.create.branch.%s", data.BranchID),
				fmt.Sprintf("member_department_history.create.organization.%s", data.OrganizationID),
				fmt.Sprintf("member_department_history.create.member_profile.%s", data.MemberProfileID),
			}
		},
		Updated: func(data *MemberDepartmentHistory) registry.Topics {
			return []string{
				"member_department_history.update",
				fmt.Sprintf("member_department_history.update.%s", data.ID),
				fmt.Sprintf("member_department_history.update.branch.%s", data.BranchID),
				fmt.Sprintf("member_department_history.update.organization.%s", data.OrganizationID),
				fmt.Sprintf("member_department_history.update.member_profile.%s", data.MemberProfileID),
			}
		},
		Deleted: func(data *MemberDepartmentHistory) registry.Topics {
			return []string{
				"member_department_history.delete",
				fmt.Sprintf("member_department_history.delete.%s", data.ID),
				fmt.Sprintf("member_department_history.delete.branch.%s", data.BranchID),
				fmt.Sprintf("member_department_history.delete.organization.%s", data.OrganizationID),
				fmt.Sprintf("member_department_history.delete.member_profile.%s", data.MemberProfileID),
			}
		},
	})
}

func (m *Core) MemberDepartmentHistoryCurrentBranch(context context.Context, organizationID uuid.UUID, branchID uuid.UUID) ([]*MemberDepartmentHistory, error) {
	return m.MemberDepartmentHistoryManager().Find(context, &MemberDepartmentHistory{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}

func (m *Core) MemberDepartmentHistoryMemberProfileID(context context.Context, memberProfileID, organizationID, branchID uuid.UUID) ([]*MemberDepartmentHistory, error) {
	return m.MemberDepartmentHistoryManager().Find(context, &MemberDepartmentHistory{
		OrganizationID:  organizationID,
		BranchID:        branchID,
		MemberProfileID: memberProfileID,
	})
}
