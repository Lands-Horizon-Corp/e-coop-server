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
	MemberCenter struct {
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

		OrganizationID uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_member_center"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_member_center"`
		Branch         *Branch       `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"branch,omitempty"`

		Name        string `gorm:"type:varchar(255)"`
		Description string `gorm:"type:text"`
	}

	MemberCenterResponse struct {
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

	MemberCenterRequest struct {
		Name        string `json:"name" validate:"required,min=1,max=255"`
		Description string `json:"description,omitempty"`
	}
)

func (m *Core) MemberCenterManager() *registry.Registry[MemberCenter, MemberCenterResponse, MemberCenterRequest] {
	return registry.NewRegistry(registry.RegistryParams[MemberCenter, MemberCenterResponse, MemberCenterRequest]{
		Preloads: []string{"CreatedBy", "UpdatedBy", "Branch", "Organization"},
		Database: m.provider.Service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return m.provider.Service.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *MemberCenter) *MemberCenterResponse {
			if data == nil {
				return nil
			}
			return &MemberCenterResponse{
				ID:             data.ID,
				CreatedAt:      data.CreatedAt.Format(time.RFC3339),
				CreatedByID:    data.CreatedByID,
				CreatedBy:      m.UserManager().ToModel(data.CreatedBy),
				UpdatedAt:      data.UpdatedAt.Format(time.RFC3339),
				UpdatedByID:    data.UpdatedByID,
				UpdatedBy:      m.UserManager().ToModel(data.UpdatedBy),
				OrganizationID: data.OrganizationID,
				Organization:   m.OrganizationManager().ToModel(data.Organization),
				BranchID:       data.BranchID,
				Branch:         m.BranchManager().ToModel(data.Branch),
				Name:           data.Name,
				Description:    data.Description,
			}
		},

		Created: func(data *MemberCenter) registry.Topics {
			return []string{
				"member_center.create",
				fmt.Sprintf("member_center.create.%s", data.ID),
				fmt.Sprintf("member_center.create.branch.%s", data.BranchID),
				fmt.Sprintf("member_center.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *MemberCenter) registry.Topics {
			return []string{
				"member_center.update",
				fmt.Sprintf("member_center.update.%s", data.ID),
				fmt.Sprintf("member_center.update.branch.%s", data.BranchID),
				fmt.Sprintf("member_center.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *MemberCenter) registry.Topics {
			return []string{
				"member_center.delete",
				fmt.Sprintf("member_center.delete.%s", data.ID),
				fmt.Sprintf("member_center.delete.branch.%s", data.BranchID),
				fmt.Sprintf("member_center.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func (m *Core) memberCenterSeed(context context.Context, tx *gorm.DB, userID uuid.UUID, organizationID uuid.UUID, branchID uuid.UUID) error {
	now := time.Now().UTC()
	memberCenter := []*MemberCenter{
		{
			Name:           "Main Wellness Center",
			Description:    "Provides health and wellness programs.",
			OrganizationID: organizationID,
			BranchID:       branchID,
			CreatedAt:      now,
			CreatedByID:    userID,
			UpdatedAt:      now,
			UpdatedByID:    userID,
		},
		{

			Name:           "Training Hub",
			Description:    "Offers skill-building and training for members.",
			OrganizationID: organizationID,
			BranchID:       branchID,
			CreatedAt:      now,
			CreatedByID:    userID,
			UpdatedAt:      now,
			UpdatedByID:    userID,
		},
		{

			Name:           "Community Support Center",
			Description:    "Focuses on community support services and events.",
			OrganizationID: organizationID,
			BranchID:       branchID,
			CreatedAt:      now,
			CreatedByID:    userID,
			UpdatedAt:      now,
			UpdatedByID:    userID,
		},
	}
	for _, data := range memberCenter {
		if err := m.MemberCenterManager().CreateWithTx(context, tx, data); err != nil {
			return eris.Wrapf(err, "failed to seed member center %s", data.Name)
		}
	}
	return nil
}

func (m *Core) MemberCenterCurrentBranch(context context.Context, organizationID uuid.UUID, branchID uuid.UUID) ([]*MemberCenter, error) {
	return m.MemberCenterManager().Find(context, &MemberCenter{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}
