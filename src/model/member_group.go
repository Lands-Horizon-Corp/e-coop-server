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
	MemberGroup struct {
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

		OrganizationID uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_member_group"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_member_group"`
		Branch         *Branch       `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"branch,omitempty"`

		Name        string `gorm:"type:varchar(50);not null"`
		Description string `gorm:"type:text;not null"`
	}

	MemberGroupResponse struct {
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
		Name           string                `json:"name"`
		Description    string                `json:"description"`
	}

	MemberGroupRequest struct {
		Name        string `json:"name" validate:"required,min=1,max=50"`
		Description string `json:"description" validate:"required"`
	}
)

func (m *Model) MemberGroup() {
	m.Migration = append(m.Migration, &MemberGroup{})
	m.MemberGroupManager = horizon_services.NewRepository(horizon_services.RepositoryParams[MemberGroup, MemberGroupResponse, MemberGroupRequest]{
		Preloads: []string{"CreatedBy", "UpdatedBy", "Branch", "Organization"},
		Service:  m.provider.Service,
		Resource: func(data *MemberGroup) *MemberGroupResponse {
			if data == nil {
				return nil
			}
			return &MemberGroupResponse{
				ID:             data.ID,
				CreatedAt:      data.CreatedAt.Format(time.RFC3339),
				CreatedByID:    data.CreatedByID,
				CreatedBy:      m.UserManager.ToModel(data.CreatedBy),
				UpdatedAt:      data.UpdatedAt.Format(time.RFC3339),
				UpdatedByID:    data.UpdatedByID,
				UpdatedBy:      m.UserManager.ToModel(data.UpdatedBy),
				OrganizationID: data.OrganizationID,
				Organization:   m.OrganizationManager.ToModel(data.Organization),
				BranchID:       data.BranchID,
				Branch:         m.BranchManager.ToModel(data.Branch),
				Name:           data.Name,
				Description:    data.Description,
			}
		},

		Created: func(data *MemberGroup) []string {
			return []string{
				"member_group.create",
				fmt.Sprintf("member_group.create.%s", data.ID),
				fmt.Sprintf("member_group.create.branch.%s", data.BranchID),
				fmt.Sprintf("member_group.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *MemberGroup) []string {
			return []string{
				"member_group.update",
				fmt.Sprintf("member_group.update.%s", data.ID),
				fmt.Sprintf("member_group.update.branch.%s", data.BranchID),
				fmt.Sprintf("member_group.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *MemberGroup) []string {
			return []string{
				"member_group.delete",
				fmt.Sprintf("member_group.delete.%s", data.ID),
				fmt.Sprintf("member_group.delete.branch.%s", data.BranchID),
				fmt.Sprintf("member_group.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func (m *Model) MemberGroupCurrentBranch(context context.Context, orgId uuid.UUID, branchId uuid.UUID) ([]*MemberGroup, error) {
	return m.MemberGroupManager.Find(context, &MemberGroup{
		OrganizationID: orgId,
		BranchID:       branchId,
	})
}
