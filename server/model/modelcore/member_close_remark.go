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
	MemberCloseRemark struct {
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

		OrganizationID uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_member_close_remark"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_member_close_remark"`
		Branch         *Branch       `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"branch,omitempty"`

		MemberProfileID *uuid.UUID     `gorm:"type:uuid" json:"member_profile_id,omitempty"`
		MemberProfile   *MemberProfile `gorm:"foreignKey:MemberProfileID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"member_profile,omitempty"`

		Reason      string `gorm:"type:varchar(255)"`
		Description string `gorm:"type:text"`
	}

	MemberCloseRemarkResponse struct {
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
		Reason          string                 `json:"reason"`
		Description     string                 `json:"description"`
	}

	MemberCloseRemarkRequest struct {
		Reason      string `json:"reason,omitempty"`
		Description string `json:"description,omitempty"`
	}
)

func (m *ModelCore) memberCloseRemark() {
	m.Migration = append(m.Migration, &MemberCloseRemark{})
	m.MemberCloseRemarkManager = services.NewRepository(services.RepositoryParams[MemberCloseRemark, MemberCloseRemarkResponse, MemberCloseRemarkRequest]{
		Preloads: []string{"CreatedBy", "UpdatedBy", "MemberProfile"},
		Service:  m.provider.Service,
		Resource: func(data *MemberCloseRemark) *MemberCloseRemarkResponse {
			if data == nil {
				return nil
			}
			return &MemberCloseRemarkResponse{
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
				MemberProfileID: *data.MemberProfileID,
				MemberProfile:   m.MemberProfileManager.ToModel(data.MemberProfile),
				Reason:          data.Reason,
				Description:     data.Description,
			}
		},

		Created: func(data *MemberCloseRemark) []string {
			return []string{
				"member_close_remark.create",
				fmt.Sprintf("member_close_remark.create.%s", data.ID),
				fmt.Sprintf("member_close_remark.create.branch.%s", data.BranchID),
				fmt.Sprintf("member_close_remark.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *MemberCloseRemark) []string {
			return []string{
				"member_close_remark.update",
				fmt.Sprintf("member_close_remark.update.%s", data.ID),
				fmt.Sprintf("member_close_remark.update.branch.%s", data.BranchID),
				fmt.Sprintf("member_close_remark.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *MemberCloseRemark) []string {
			return []string{
				"member_close_remark.delete",
				fmt.Sprintf("member_close_remark.delete.%s", data.ID),
				fmt.Sprintf("member_close_remark.delete.branch.%s", data.BranchID),
				fmt.Sprintf("member_close_remark.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func (m *ModelCore) MemberCloseRemarkCurrentBranch(context context.Context, orgID uuid.UUID, branchID uuid.UUID) ([]*MemberCloseRemark, error) {
	return m.MemberCloseRemarkManager.Find(context, &MemberCloseRemark{
		OrganizationID: orgID,
		BranchID:       branchID,
	})
}
