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
	MemberContactReference struct {
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

		OrganizationID uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_member_contact_reference"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_member_contact_reference"`
		Branch         *Branch       `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"branch,omitempty"`

		MemberProfileID uuid.UUID      `gorm:"type:uuid;not null"`
		MemberProfile   *MemberProfile `gorm:"foreignKey:MemberProfileID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"member_profile,omitempty"`

		Name          string `gorm:"type:varchar(255)"`
		Description   string `gorm:"type:text"`
		ContactNumber string `gorm:"type:varchar(30)"`
	}

	MemberContactReferenceResponse struct {
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
		Name            string                 `json:"name"`
		Description     string                 `json:"description"`
		ContactNumber   string                 `json:"contact_number"`
	}

	MemberContactReferenceRequest struct {
		MemberProfileID uuid.UUID `json:"member_profile_id" validate:"required"`
		Name            string    `json:"name" validate:"required,min=1,max=255"`
		Description     string    `json:"description,omitempty"`
		ContactNumber   string    `json:"contact_number,omitempty" validate:"omitempty,max=30"`
	}
)

func (m *ModelCore) MemberContactReference() {
	m.Migration = append(m.Migration, &MemberContactReference{})
	m.MemberContactReferenceManager = horizon_services.NewRepository(horizon_services.RepositoryParams[MemberContactReference, MemberContactReferenceResponse, MemberContactReferenceRequest]{
		Preloads: []string{"CreatedBy", "UpdatedBy", "MemberProfile"},
		Service:  m.provider.Service,
		Resource: func(data *MemberContactReference) *MemberContactReferenceResponse {
			if data == nil {
				return nil
			}
			return &MemberContactReferenceResponse{
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
				MemberProfileID: data.MemberProfileID,
				MemberProfile:   m.MemberProfileManager.ToModel(data.MemberProfile),
				Name:            data.Name,
				Description:     data.Description,
				ContactNumber:   data.ContactNumber,
			}
		},

		Created: func(data *MemberContactReference) []string {
			return []string{
				"member_contact_reference.create",
				fmt.Sprintf("member_contact_reference.create.%s", data.ID),
				fmt.Sprintf("member_contact_reference.create.branch.%s", data.BranchID),
				fmt.Sprintf("member_contact_reference.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *MemberContactReference) []string {
			return []string{
				"member_contact_reference.update",
				fmt.Sprintf("member_contact_reference.update.%s", data.ID),
				fmt.Sprintf("member_contact_reference.update.branch.%s", data.BranchID),
				fmt.Sprintf("member_contact_reference.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *MemberContactReference) []string {
			return []string{
				"member_contact_reference.delete",
				fmt.Sprintf("member_contact_reference.delete.%s", data.ID),
				fmt.Sprintf("member_contact_reference.delete.branch.%s", data.BranchID),
				fmt.Sprintf("member_contact_reference.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func (m *ModelCore) MemberContactReferenceCurrentBranch(context context.Context, orgId uuid.UUID, branchId uuid.UUID) ([]*MemberContactReference, error) {
	return m.MemberContactReferenceManager.Find(context, &MemberContactReference{
		OrganizationID: orgId,
		BranchID:       branchId,
	})
}
