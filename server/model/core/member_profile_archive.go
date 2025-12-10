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
	// MemberProfileMedia represents the MemberProfileMedia model.
	MemberProfileArchive struct {
		ID          uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
		CreatedAt   time.Time      `gorm:"not null;default:now()" json:"created_at"`
		CreatedByID uuid.UUID      `gorm:"type:uuid" json:"created_by_id"`
		CreatedBy   *User          `gorm:"foreignKey:CreatedByID;constraint:OnDelete:SET NULL;" json:"created_by,omitempty"`
		UpdatedAt   time.Time      `gorm:"not null;default:now()" json:"updated_at"`
		UpdatedByID uuid.UUID      `gorm:"type:uuid" json:"updated_by_id"`
		UpdatedBy   *User          `gorm:"foreignKey:UpdatedByID;constraint:OnDelete:SET NULL;" json:"updated_by,omitempty"`
		DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at"`
		DeletedByID *uuid.UUID     `gorm:"type:uuid" json:"deleted_by_id"`
		DeletedBy   *User          `gorm:"foreignKey:DeletedByID;constraint:OnDelete:SET NULL;" json:"deleted_by,omitempty"`

		OrganizationID *uuid.UUID    `gorm:"type:uuid;index:idx_organization_branch_member_profile_media" json:"organization_id,omitempty"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"organization,omitempty"`
		BranchID       *uuid.UUID    `gorm:"type:uuid;index:idx_organization_branch_member_profile_media" json:"branch_id,omitempty"`
		Branch         *Branch       `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"branch,omitempty"`

		MemberProfileID *uuid.UUID     `gorm:"type:uuid" json:"member_profile_id,omitempty"`
		MemberProfile   *MemberProfile `gorm:"foreignKey:MemberProfileID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"member_profile,omitempty"`

		MediaID *uuid.UUID `gorm:"type:uuid" json:"media_id"`
		Media   *Media     `gorm:"foreignKey:MediaID;constraint:OnDelete:SET NULL;" json:"media,omitempty"`

		Name        string `gorm:"type:varchar(255);not null" json:"name"`
		Description string `gorm:"type:text" json:"description"`
		Category    string `gorm:"type:varchar(100);index" json:"category"`
	}

	// MemberProfileMediaResponse represents the response structure for memberprofilemedia data

	// MemberProfileMediaResponse represents the response structure for MemberProfileMedia.
	MemberProfileArchiveResponse struct {
		ID              uuid.UUID              `json:"id"`
		CreatedAt       string                 `json:"created_at"`
		CreatedByID     uuid.UUID              `json:"created_by_id"`
		CreatedBy       *UserResponse          `json:"created_by,omitempty"`
		UpdatedAt       string                 `json:"updated_at"`
		UpdatedByID     uuid.UUID              `json:"updated_by_id"`
		UpdatedBy       *UserResponse          `json:"updated_by,omitempty"`
		OrganizationID  *uuid.UUID             `json:"organization_id,omitempty"`
		Organization    *OrganizationResponse  `json:"organization,omitempty"`
		BranchID        *uuid.UUID             `json:"branch_id,omitempty"`
		Branch          *BranchResponse        `json:"branch,omitempty"`
		MemberProfileID *uuid.UUID             `json:"member_profile_id,omitempty"`
		MemberProfile   *MemberProfileResponse `json:"member_profile,omitempty"`
		MediaID         *uuid.UUID             `json:"media_id,omitempty"`
		Media           *MediaResponse         `json:"media,omitempty"`
		Name            string                 `json:"name"`
		Description     string                 `json:"description"`
		Category        string                 `json:"category"`
	}

	// MemberProfileArchiveRequest represents the request structure for creating/updating memberprofilearchive
	MemberProfileArchiveRequest struct {
		Name            string     `json:"name" validate:"required,min=1,max=255"`
		Description     string     `json:"description,omitempty"`
		MemberProfileID *uuid.UUID `json:"member_profile_id,omitempty"`
		MediaID         *uuid.UUID `json:"media_id,omitempty"`
		OrganizationID  *uuid.UUID `json:"organization_id,omitempty"`
		BranchID        *uuid.UUID `json:"branch_id,omitempty"`
		Category        string     `json:"category,omitempty"`
	}

	MemberProfileArchiveCategoryResponse struct {
		Name  string `json:"name"`
		Count int64  `json:"count"`
	}
	MemberProfileArchiveBulkRequest struct {
		IDs      uuid.UUIDs `json:"ids"`
		Category string     `json:"category" validate:"required,min=1,max=100"`
	}
)

func (m *Core) memberProfileArchive() {
	m.Migration = append(m.Migration, &MemberProfileArchive{})
	m.MemberProfileArchiveManager = *registry.NewRegistry(registry.RegistryParams[MemberProfileArchive, MemberProfileArchiveResponse, MemberProfileArchiveRequest]{
		Preloads: []string{"CreatedBy", "UpdatedBy", "Media", "MemberProfile"},
		Database: m.provider.Service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return m.provider.Service.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *MemberProfileArchive) *MemberProfileArchiveResponse {
			if data == nil {
				return nil
			}
			return &MemberProfileArchiveResponse{
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
				MediaID:         data.MediaID,
				Media:           m.MediaManager.ToModel(data.Media),
				Name:            data.Name,
				Description:     data.Description,
				Category:        data.Category,
			}
		},
		Created: func(data *MemberProfileArchive) registry.Topics {
			events := []string{
				"member_profile_archive.create",
				fmt.Sprintf("member_profile_archive.create.%s", data.ID),
			}
			if data.BranchID != nil {
				events = append(events, fmt.Sprintf("member_profile_archive.create.branch.%s", *data.BranchID))
			}
			if data.OrganizationID != nil {
				events = append(events, fmt.Sprintf("member_profile_archive.create.organization.%s", *data.OrganizationID))
			}
			return events
		},
		Updated: func(data *MemberProfileArchive) registry.Topics {
			events := []string{
				"member_profile_archive.update",
				fmt.Sprintf("member_profile_archive.update.%s", data.ID),
			}
			if data.BranchID != nil {
				events = append(events, fmt.Sprintf("member_profile_archive.update.branch.%s", *data.BranchID))
			}
			if data.OrganizationID != nil {
				events = append(events, fmt.Sprintf("member_profile_archive.update.organization.%s", *data.OrganizationID))
			}
			return events
		},
		Deleted: func(data *MemberProfileArchive) registry.Topics {
			events := []string{
				"member_profile_archive.delete",
				fmt.Sprintf("member_profile_archive.delete.%s", data.ID),
			}
			if data.BranchID != nil {
				events = append(events, fmt.Sprintf("member_profile_archive.delete.branch.%s", *data.BranchID))
			}
			if data.OrganizationID != nil {
				events = append(events, fmt.Sprintf("member_profile_archive.delete.organization.%s", *data.OrganizationID))
			}
			return events
		},
	})
}

// MemberProfileArchiveCurrentBranch returns MemberProfileArchiveCurrentBranch for the current branch or organization where applicable.
func (m *Core) MemberProfileArchiveCurrentBranch(context context.Context, organizationID *uuid.UUID, branchID *uuid.UUID) ([]*MemberProfileArchive, error) {
	return m.MemberProfileArchiveManager.Find(context, &MemberProfileArchive{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}
