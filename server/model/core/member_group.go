package core

import (
	"context"
	"fmt"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/pkg/registry"
	"github.com/google/uuid"
	"github.com/rotisserie/eris"
	"gorm.io/gorm"
)

type (
	// MemberGroup represents the MemberGroup model.
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

	// MemberGroupResponse represents the response structure for membergroup data

	// MemberGroupResponse represents the response structure for MemberGroup.
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

	// MemberGroupRequest represents the request structure for creating/updating membergroup

	// MemberGroupRequest represents the request structure for MemberGroup.
	MemberGroupRequest struct {
		Name        string `json:"name" validate:"required,min=1,max=50"`
		Description string `json:"description" validate:"required"`
	}
)

func (m *Core) memberGroup() {
	m.Migration = append(m.Migration, &MemberGroup{})
	m.MemberGroupManager = *registry.NewRegistry(registry.RegistryParams[MemberGroup, MemberGroupResponse, MemberGroupRequest]{
		Preloads: []string{"CreatedBy", "UpdatedBy", "Branch", "Organization"},
		Database: m.provider.Service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return m.provider.Service.Broker.Dispatch(topics, payload)
		},
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

		Created: func(data *MemberGroup) registry.Topics {
			return []string{
				"member_group.create",
				fmt.Sprintf("member_group.create.%s", data.ID),
				fmt.Sprintf("member_group.create.branch.%s", data.BranchID),
				fmt.Sprintf("member_group.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *MemberGroup) registry.Topics {
			return []string{
				"member_group.update",
				fmt.Sprintf("member_group.update.%s", data.ID),
				fmt.Sprintf("member_group.update.branch.%s", data.BranchID),
				fmt.Sprintf("member_group.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *MemberGroup) registry.Topics {
			return []string{
				"member_group.delete",
				fmt.Sprintf("member_group.delete.%s", data.ID),
				fmt.Sprintf("member_group.delete.branch.%s", data.BranchID),
				fmt.Sprintf("member_group.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func (m *Core) memberGroupSeed(context context.Context, tx *gorm.DB, userID uuid.UUID, organizationID uuid.UUID, branchID uuid.UUID) error {
	now := time.Now().UTC()
	memberGroup := []*MemberGroup{
		{

			CreatedAt:      now,
			CreatedByID:    userID,
			UpdatedAt:      now,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Single Moms",
			Description:    "Support group for single mothers in the community.",
		},
		{

			CreatedAt:      now,
			CreatedByID:    userID,
			UpdatedAt:      now,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Athletes",
			Description:    "Members who actively participate in sports and fitness.",
		},
		{

			CreatedAt:      now,
			CreatedByID:    userID,
			UpdatedAt:      now,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Tech",
			Description:    "Members involved in information technology or development.",
		},
		{

			CreatedAt:      now,
			CreatedByID:    userID,
			UpdatedAt:      now,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Graphics Artists",
			Description:    "Creative members who specialize in digital and graphic design.",
		},
		{

			CreatedAt:      now,
			CreatedByID:    userID,
			UpdatedAt:      now,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Accountants",
			Description:    "Finance-focused members responsible for budgeting and auditing.",
		},
	}
	for _, data := range memberGroup {
		if err := m.MemberGroupManager.CreateWithTx(context, tx, data); err != nil {
			return eris.Wrapf(err, "failed to seed member group %s", data.Name)
		}
	}
	return nil
}

// MemberGroupCurrentBranch returns MemberGroupCurrentBranch for the current branch or organization where applicable.
func (m *Core) MemberGroupCurrentBranch(context context.Context, organizationID uuid.UUID, branchID uuid.UUID) ([]*MemberGroup, error) {
	return m.MemberGroupManager.Find(context, &MemberGroup{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}
