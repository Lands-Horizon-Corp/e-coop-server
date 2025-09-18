package model

import (
	"context"
	"fmt"
	"time"

	horizon_services "github.com/Lands-Horizon-Corp/e-coop-server/services"
	"github.com/google/uuid"
	"github.com/rotisserie/eris"
	"gorm.io/gorm"
)

type (
	MemberGender struct {
		ID          uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
		CreatedAt   time.Time      `gorm:"not null;default:now()" json:"created_at"`
		CreatedByID uuid.UUID      `gorm:"type:uuid"`
		CreatedBy   *User          `gorm:"foreignKey:CreatedByID;constraint:OnDelete:SET NULL;" json:"created_by,omitempty"`
		UpdatedAt   time.Time      `gorm:"not null;default:now()"`
		UpdatedByID uuid.UUID      `gorm:"type:uuid"`
		UpdatedBy   *User          `gorm:"foreignKey:UpdatedByID;constraint:OnDelete:SET NULL;" json:"updated_by,omitempty"`
		DeletedAt   gorm.DeletedAt `gorm:"index"`
		DeletedByID *uuid.UUID     `gorm:"type:uuid"`
		DeletedBy   *User          `gorm:"foreignKey:DeletedByID;constraint:OnDelete:SET NULL;" json:"deleted_by,omitempty"`

		OrganizationID uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_member_gender"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_member_gender"`
		Branch         *Branch       `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"branch,omitempty"`

		Name        string `gorm:"type:varchar(255)"`
		Description string `gorm:"type:text"`
	}

	MemberGenderResponse struct {
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

	MemberGenderRequest struct {
		Name        string `json:"name" validate:"required,min=1,max=255"`
		Description string `json:"description,omitempty"`
	}
)

func (m *Model) MemberGender() {
	m.Migration = append(m.Migration, &MemberGender{})
	m.MemberGenderManager = horizon_services.NewRepository(horizon_services.RepositoryParams[MemberGender, MemberGenderResponse, MemberGenderRequest]{
		Preloads: []string{"CreatedBy", "UpdatedBy", "Branch", "Organization"},
		Service:  m.provider.Service,
		Resource: func(data *MemberGender) *MemberGenderResponse {
			if data == nil {
				return nil
			}
			return &MemberGenderResponse{
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

		Created: func(data *MemberGender) []string {
			return []string{
				"member_gender.create",
				fmt.Sprintf("member_gender.create.%s", data.ID),
				fmt.Sprintf("member_gender.create.branch.%s", data.BranchID),
				fmt.Sprintf("member_gender.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *MemberGender) []string {
			return []string{
				"member_gender.update",
				fmt.Sprintf("member_gender.update.%s", data.ID),
				fmt.Sprintf("member_gender.update.branch.%s", data.BranchID),
				fmt.Sprintf("member_gender.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *MemberGender) []string {
			return []string{
				"member_gender.delete",
				fmt.Sprintf("member_gender.delete.%s", data.ID),
				fmt.Sprintf("member_gender.delete.branch.%s", data.BranchID),
				fmt.Sprintf("member_gender.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func (m *Model) MemberGenderSeed(context context.Context, tx *gorm.DB, userID uuid.UUID, organizationID uuid.UUID, branchID uuid.UUID) error {
	now := time.Now().UTC()
	memberGenders := []*MemberGender{
		{

			CreatedAt:      now,
			CreatedByID:    userID,
			UpdatedAt:      now,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Male",
			Description:    "Identifies as male.",
		},
		{

			CreatedAt:      now,
			CreatedByID:    userID,
			UpdatedAt:      now,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Female",
			Description:    "Identifies as female.",
		},
		{

			CreatedAt:      now,
			CreatedByID:    userID,
			UpdatedAt:      now,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Other",
			Description:    "Identifies outside the binary gender categories.",
		},
	}
	for _, data := range memberGenders {
		if err := m.MemberGenderManager.CreateWithTx(context, tx, data); err != nil {
			return eris.Wrapf(err, "failed to seed member gender %s", data.Name)
		}
	}

	return nil
}

func (m *Model) MemberGenderCurrentBranch(context context.Context, orgId uuid.UUID, branchId uuid.UUID) ([]*MemberGender, error) {
	return m.MemberGenderManager.Find(context, &MemberGender{
		OrganizationID: orgId,
		BranchID:       branchId,
	})
}
