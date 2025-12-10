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
	// MemberOtherInformationEntry represents the MemberOtherInformationEntry model.
	MemberOtherInformationEntry struct {
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

		OrganizationID uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_member_other_information_entry"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_member_other_information_entry"`
		Branch         *Branch       `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"branch,omitempty"`

		Name        string    `gorm:"type:varchar(255);not null"`
		Description string    `gorm:"type:text"`
		EntryDate   time.Time `gorm:"type:timestamp"`
	}

	// MemberOtherInformationEntryResponse represents the response structure for memberotherinformationentry data

	// MemberOtherInformationEntryResponse represents the response structure for MemberOtherInformationEntry.
	MemberOtherInformationEntryResponse struct {
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
		EntryDate      string                `json:"entry_date"`
	}

	// MemberOtherInformationEntryRequest represents the request structure for creating/updating memberotherinformationentry

	// MemberOtherInformationEntryRequest represents the request structure for MemberOtherInformationEntry.
	MemberOtherInformationEntryRequest struct {
		Name        string    `json:"name" validate:"required,min=1,max=255"`
		Description string    `json:"description,omitempty"`
		EntryDate   time.Time `json:"entry_date"`
	}
)

func (m *Core) memberOtherInformationEntry() {
	m.Migration = append(m.Migration, &MemberOtherInformationEntry{})
	m.MemberOtherInformationEntryManager = *registry.NewRegistry(registry.RegistryParams[MemberOtherInformationEntry, MemberOtherInformationEntryResponse, MemberOtherInformationEntryRequest]{
		Preloads: []string{"CreatedBy", "UpdatedBy", "Branch", "Organization"},
		Database: m.provider.Service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return m.provider.Service.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *MemberOtherInformationEntry) *MemberOtherInformationEntryResponse {
			if data == nil {
				return nil
			}
			return &MemberOtherInformationEntryResponse{
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
				EntryDate:      data.EntryDate.Format(time.RFC3339),
			}
		},

		Created: func(data *MemberOtherInformationEntry) registry.Topics {
			return []string{
				"member_other_information_entry.create",
				fmt.Sprintf("member_other_information_entry.create.%s", data.ID),
				fmt.Sprintf("member_other_information_entry.create.branch.%s", data.BranchID),
				fmt.Sprintf("member_other_information_entry.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *MemberOtherInformationEntry) registry.Topics {
			return []string{
				"member_other_information_entry.update",
				fmt.Sprintf("member_other_information_entry.update.%s", data.ID),
				fmt.Sprintf("member_other_information_entry.update.branch.%s", data.BranchID),
				fmt.Sprintf("member_other_information_entry.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *MemberOtherInformationEntry) registry.Topics {
			return []string{
				"member_other_information_entry.delete",
				fmt.Sprintf("member_other_information_entry.delete.%s", data.ID),
				fmt.Sprintf("member_other_information_entry.delete.branch.%s", data.BranchID),
				fmt.Sprintf("member_other_information_entry.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

// MemberOtherInformationEntryCurrentBranch returns MemberOtherInformationEntryCurrentBranch for the current branch or organization where applicable.
func (m *Core) MemberOtherInformationEntryCurrentBranch(context context.Context, organizationID uuid.UUID, branchID uuid.UUID) ([]*MemberOtherInformationEntry, error) {
	return m.MemberOtherInformationEntryManager.Find(context, &MemberOtherInformationEntry{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}
