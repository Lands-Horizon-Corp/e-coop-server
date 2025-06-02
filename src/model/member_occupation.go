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
	MemberOccupation struct {
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

		OrganizationID uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_member_occupation"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_member_occupation"`
		Branch         *Branch       `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"branch,omitempty"`

		Name        string `gorm:"type:varchar(255)"`
		Description string `gorm:"type:text"`
	}

	MemberOccupationResponse struct {
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

	MemberOccupationRequest struct {
		Name        string `json:"name" validate:"required,min=1,max=255"`
		Description string `json:"description,omitempty"`
	}
)

func (m *Model) MemberOccupation() {
	m.Migration = append(m.Migration, &MemberOccupation{})
	m.MemberOccupationManager = horizon_services.NewRepository(horizon_services.RepositoryParams[MemberOccupation, MemberOccupationResponse, MemberOccupationRequest]{
		Preloads: []string{"CreatedBy", "UpdatedBy", "Branch", "Organization"},
		Service:  m.provider.Service,
		Resource: func(data *MemberOccupation) *MemberOccupationResponse {
			if data == nil {
				return nil
			}
			return &MemberOccupationResponse{
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

		Created: func(data *MemberOccupation) []string {
			return []string{
				"member_occupation.create",
				fmt.Sprintf("member_occupation.create.%s", data.ID),
				fmt.Sprintf("member_occupation.create.branch.%s", data.BranchID),
				fmt.Sprintf("member_occupation.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *MemberOccupation) []string {
			return []string{
				"member_occupation.update",
				fmt.Sprintf("member_occupation.update.%s", data.ID),
				fmt.Sprintf("member_occupation.update.branch.%s", data.BranchID),
				fmt.Sprintf("member_occupation.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *MemberOccupation) []string {
			return []string{
				"member_occupation.delete",
				fmt.Sprintf("member_occupation.delete.%s", data.ID),
				fmt.Sprintf("member_occupation.delete.branch.%s", data.BranchID),
				fmt.Sprintf("member_occupation.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func (m *Model) MemberOccupationCurrentBranch(context context.Context, orgId uuid.UUID, branchId uuid.UUID) ([]*MemberOccupation, error) {
	return m.MemberOccupationManager.Find(context, &MemberOccupation{
		OrganizationID: orgId,
		BranchID:       branchId,
	})
}
