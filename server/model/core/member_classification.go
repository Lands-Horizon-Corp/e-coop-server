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
	MemberClassification struct {
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

		OrganizationID uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_member_classification"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_member_classification"`
		Branch         *Branch       `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"branch,omitempty"`

		Name        string `gorm:"type:varchar(255);not null"`
		Icon        string `gorm:"type:varchar(255)"`
		Description string `gorm:"type:text"`
	}

	MemberClassificationResponse struct {
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
		Icon           string                `json:"icon"`
		Description    string                `json:"description"`
	}

	MemberClassificationRequest struct {
		Name        string `json:"name" validate:"required,min=1,max=255"`
		Icon        string `json:"icon,omitempty"`
		Description string `json:"description,omitempty"`
	}
)

func (m *Core) memberClassification() {
	m.Migration = append(m.Migration, &MemberClassification{})
	m.MemberClassificationManager() = registry.NewRegistry(registry.RegistryParams[MemberClassification, MemberClassificationResponse, MemberClassificationRequest]{
		Preloads: []string{"CreatedBy", "UpdatedBy", "Branch", "Organization"},
		Database: m.provider.Service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return m.provider.Service.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *MemberClassification) *MemberClassificationResponse {
			if data == nil {
				return nil
			}
			return &MemberClassificationResponse{
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
				Icon:           data.Icon,
				Description:    data.Description,
			}
		},

		Created: func(data *MemberClassification) registry.Topics {
			return []string{
				"member_classification.create",
				fmt.Sprintf("member_classification.create.%s", data.ID),
				fmt.Sprintf("member_classification.create.branch.%s", data.BranchID),
				fmt.Sprintf("member_classification.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *MemberClassification) registry.Topics {
			return []string{
				"member_classification.update",
				fmt.Sprintf("member_classification.update.%s", data.ID),
				fmt.Sprintf("member_classification.update.branch.%s", data.BranchID),
				fmt.Sprintf("member_classification.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *MemberClassification) registry.Topics {
			return []string{
				"member_classification.delete",
				fmt.Sprintf("member_classification.delete.%s", data.ID),
				fmt.Sprintf("member_classification.delete.branch.%s", data.BranchID),
				fmt.Sprintf("member_classification.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func (m *Core) memberClassificationSeed(context context.Context, tx *gorm.DB, userID uuid.UUID, organizationID uuid.UUID, branchID uuid.UUID) error {
	now := time.Now().UTC()
	memberClassifications := []*MemberClassification{
		{

			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Gold",
			Icon:           "sunrise",
			Description:    "Gold membership is reserved for top-tier members with excellent credit scores and consistent loyalty.",
		},
		{

			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Silver",
			Icon:           "moon-star",
			Description:    "Silver membership is designed for members with good credit history and regular engagement.",
		},
		{

			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Bronze",
			Icon:           "cloud",
			Description:    "Bronze membership is for new or casual members who are starting their journey with us.",
		},
		{

			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedByID:    userID,
			UpdatedByID:    userID,
			OrganizationID: organizationID,
			BranchID:       branchID,
			Name:           "Platinum",
			Icon:           "gem",
			Description:    "Platinum membership offers ZEDE benefits to elite members with outstanding history and contributions.",
		},
	}
	for _, data := range memberClassifications {
		if err := m.MemberClassificationManager().CreateWithTx(context, tx, data); err != nil {
			return eris.Wrapf(err, "failed to seed member classification %s", data.Name)
		}
	}
	return nil
}

func (m *Core) MemberClassificationCurrentBranch(context context.Context, organizationID uuid.UUID, branchID uuid.UUID) ([]*MemberClassification, error) {
	return m.MemberClassificationManager().Find(context, &MemberClassification{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}
