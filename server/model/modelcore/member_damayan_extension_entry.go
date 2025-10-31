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
	MemberDamayanExtensionEntry struct {
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

		OrganizationID uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_member_damayan_extension_entry"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_member_damayan_extension_entry"`
		Branch         *Branch       `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"branch,omitempty"`

		MemberProfileID uuid.UUID      `gorm:"type:uuid;not null"`
		MemberProfile   *MemberProfile `gorm:"foreignKey:MemberProfileID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE;" json:"member_profile,omitempty"`

		Name        string     `gorm:"type:varchar(255)"`
		Description string     `gorm:"type:text"`
		Birthdate   *time.Time `gorm:"type:timestamp"`
	}

	MemberDamayanExtensionEntryResponse struct {
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
		Birthdate       *string                `json:"birthdate,omitempty"`
	}

	MemberDamayanExtensionEntryRequest struct {
		MemberProfileID uuid.UUID  `json:"member_profile_id" validate:"required"`
		Name            string     `json:"name" validate:"required,min=1,max=255"`
		Description     string     `json:"description,omitempty"`
		Birthdate       *time.Time `json:"birthdate,omitempty"`
	}
)

func (m *ModelCore) memberDamayanExtensionEntry() {
	m.Migration = append(m.Migration, &MemberDamayanExtensionEntry{})
	m.MemberDamayanExtensionEntryManager = services.NewRepository(services.RepositoryParams[MemberDamayanExtensionEntry, MemberDamayanExtensionEntryResponse, MemberDamayanExtensionEntryRequest]{
		Preloads: []string{"CreatedBy", "UpdatedBy", "MemberProfile"},
		Service:  m.provider.Service,
		Resource: func(data *MemberDamayanExtensionEntry) *MemberDamayanExtensionEntryResponse {
			if data == nil {
				return nil
			}
			var birthdateStr *string
			if data.Birthdate != nil {
				s := data.Birthdate.Format(time.RFC3339)
				birthdateStr = &s
			}
			return &MemberDamayanExtensionEntryResponse{
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
				Birthdate:       birthdateStr,
			}
		},

		Created: func(data *MemberDamayanExtensionEntry) []string {
			return []string{
				"member_damayan_extension_entry.create",
				fmt.Sprintf("member_damayan_extension_entry.create.%s", data.ID),
				fmt.Sprintf("member_damayan_extension_entry.create.branch.%s", data.BranchID),
				fmt.Sprintf("member_damayan_extension_entry.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *MemberDamayanExtensionEntry) []string {
			return []string{
				"member_damayan_extension_entry.update",
				fmt.Sprintf("member_damayan_extension_entry.update.%s", data.ID),
				fmt.Sprintf("member_damayan_extension_entry.update.branch.%s", data.BranchID),
				fmt.Sprintf("member_damayan_extension_entry.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *MemberDamayanExtensionEntry) []string {
			return []string{
				"member_damayan_extension_entry.delete",
				fmt.Sprintf("member_damayan_extension_entry.delete.%s", data.ID),
				fmt.Sprintf("member_damayan_extension_entry.delete.branch.%s", data.BranchID),
				fmt.Sprintf("member_damayan_extension_entry.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func (m *ModelCore) memberDamayanExtensionEntryCurrentBranch(context context.Context, orgId uuid.UUID, branchId uuid.UUID) ([]*MemberDamayanExtensionEntry, error) {
	return m.MemberDamayanExtensionEntryManager.Find(context, &MemberDamayanExtensionEntry{
		OrganizationID: orgId,
		BranchID:       branchId,
	})
}
