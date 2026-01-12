package core

import (
	"context"
	"fmt"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/horizon"
	"github.com/Lands-Horizon-Corp/e-coop-server/pkg/registry"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type (
	MemberProfileMedia struct {
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
	}

	MemberProfileMediaResponse struct {
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
	}

	MemberProfileMediaRequest struct {
		Name            string     `json:"name" validate:"required,min=1,max=255"`
		Description     string     `json:"description,omitempty"`
		MemberProfileID *uuid.UUID `json:"member_profile_id,omitempty"`
		MediaID         *uuid.UUID `json:"media_id,omitempty"`
		OrganizationID  *uuid.UUID `json:"organization_id,omitempty"`
		BranchID        *uuid.UUID `json:"branch_id,omitempty"`
	}
)

func MemberProfileMediaManager(service *horizon.HorizonService) *registry.Registry[MemberProfileMedia, MemberProfileMediaResponse, MemberProfileMediaRequest] {
	return registry.NewRegistry(registry.RegistryParams[MemberProfileMedia, MemberProfileMediaResponse, MemberProfileMediaRequest]{
		Preloads: []string{"CreatedBy", "UpdatedBy", "Media", "MemberProfile"},
		Database: service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return service.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *MemberProfileMedia) *MemberProfileMediaResponse {
			if data == nil {
				return nil
			}
			return &MemberProfileMediaResponse{
				ID:              data.ID,
				CreatedAt:       data.CreatedAt.Format(time.RFC3339),
				CreatedByID:     data.CreatedByID,
				CreatedBy:       UserManager(service).ToModel(data.CreatedBy),
				UpdatedAt:       data.UpdatedAt.Format(time.RFC3339),
				UpdatedByID:     data.UpdatedByID,
				UpdatedBy:       UserManager(service).ToModel(data.UpdatedBy),
				OrganizationID:  data.OrganizationID,
				Organization:    OrganizationManager(service).ToModel(data.Organization),
				BranchID:        data.BranchID,
				Branch:          BranchManager(service).ToModel(data.Branch),
				MemberProfileID: data.MemberProfileID,
				MemberProfile:   MemberProfileManager(service).ToModel(data.MemberProfile),
				MediaID:         data.MediaID,
				Media:           MediaManager(service).ToModel(data.Media),
				Name:            data.Name,
				Description:     data.Description,
			}
		},
		Created: func(data *MemberProfileMedia) registry.Topics {
			events := []string{
				"member_profile_media.create",
				fmt.Sprintf("member_profile_media.create.%s", data.ID),
			}
			if data.BranchID != nil {
				events = append(events, fmt.Sprintf("member_profile_media.create.branch.%s", *data.BranchID))
			}
			if data.OrganizationID != nil {
				events = append(events, fmt.Sprintf("member_profile_media.create.organization.%s", *data.OrganizationID))
			}
			return events
		},
		Updated: func(data *MemberProfileMedia) registry.Topics {
			events := []string{
				"member_profile_media.update",
				fmt.Sprintf("member_profile_media.update.%s", data.ID),
			}
			if data.BranchID != nil {
				events = append(events, fmt.Sprintf("member_profile_media.update.branch.%s", *data.BranchID))
			}
			if data.OrganizationID != nil {
				events = append(events, fmt.Sprintf("member_profile_media.update.organization.%s", *data.OrganizationID))
			}
			return events
		},
		Deleted: func(data *MemberProfileMedia) registry.Topics {
			events := []string{
				"member_profile_media.delete",
				fmt.Sprintf("member_profile_media.delete.%s", data.ID),
			}
			if data.BranchID != nil {
				events = append(events, fmt.Sprintf("member_profile_media.delete.branch.%s", *data.BranchID))
			}
			if data.OrganizationID != nil {
				events = append(events, fmt.Sprintf("member_profile_media.delete.organization.%s", *data.OrganizationID))
			}
			return events
		},
	})
}

func MemberProfileMediaCurrentBranch(context context.Context, service *horizon.HorizonService, organizationID *uuid.UUID, branchID *uuid.UUID) ([]*MemberProfileMedia, error) {
	return MemberProfileMediaManager(service).Find(context, &MemberProfileMedia{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}
