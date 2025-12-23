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
	OrganizationMedia struct {
		ID        uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
		CreatedAt time.Time      `gorm:"not null;default:now()" json:"created_at"`
		UpdatedAt time.Time      `gorm:"not null;default:now()" json:"updated_at"`
		DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at"`

		Name        string  `gorm:"type:varchar(255);not null" json:"name"`
		Description *string `gorm:"type:text" json:"description,omitempty"`

		OrganizationID uuid.UUID    `gorm:"type:uuid;not null;index" json:"organization_id"`
		Organization   Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE;" json:"organization"`

		MediaID uuid.UUID `gorm:"type:uuid;not null" json:"media_id"`
		Media   Media     `gorm:"foreignKey:MediaID;constraint:OnDelete:CASCADE;" json:"media"`
	}

	OrganizationMediaResponse struct {
		ID        uuid.UUID `json:"id"`
		CreatedAt string    `json:"created_at"`
		UpdatedAt string    `json:"updated_at"`

		Name        string  `json:"name"`
		Description *string `json:"description,omitempty"`

		OrganizationID uuid.UUID             `json:"organization_id"`
		Organization   *OrganizationResponse `json:"organization,omitempty"`

		MediaID uuid.UUID      `json:"media_id"`
		Media   *MediaResponse `json:"media,omitempty"`
	}

	OrganizationMediaRequest struct {
		Name        string  `json:"name" validate:"required,min=1,max=255"`
		Description *string `json:"description,omitempty"`

		OrganizationID uuid.UUID `json:"organization_id" validate:"required,uuid4"`
		MediaID        uuid.UUID `json:"media_id" validate:"required,uuid4"`
	}
)

func (m *Core) OrganizationMediaManager() *registry.Registry[OrganizationMedia, OrganizationMediaResponse, OrganizationMediaRequest] {
	return registry.NewRegistry(registry.RegistryParams[OrganizationMedia, OrganizationMediaResponse, OrganizationMediaRequest]{
		Preloads: []string{"Media"},
		Database: m.provider.Service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return m.provider.Service.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *OrganizationMedia) *OrganizationMediaResponse {
			if data == nil {
				return nil
			}
			return &OrganizationMediaResponse{
				ID:        data.ID,
				CreatedAt: data.CreatedAt.Format(time.RFC3339),
				UpdatedAt: data.UpdatedAt.Format(time.RFC3339),

				Name:        data.Name,
				Description: data.Description,

				OrganizationID: data.OrganizationID,
				Organization:   m.OrganizationManager().ToModel(&data.Organization),

				MediaID: data.MediaID,
				Media:   m.MediaManager().ToModel(&data.Media),
			}
		},
		Created: func(data *OrganizationMedia) registry.Topics {
			return []string{
				"organization_media.create",
				fmt.Sprintf("organization_media.create.%s", data.ID),
				fmt.Sprintf("organization_media.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *OrganizationMedia) registry.Topics {
			return []string{
				"organization_media.update",
				fmt.Sprintf("organization_media.update.%s", data.ID),
				fmt.Sprintf("organization_media.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *OrganizationMedia) registry.Topics {
			return []string{
				"organization_media.delete",
				fmt.Sprintf("organization_media.delete.%s", data.ID),
				fmt.Sprintf("organization_media.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func (m *Core) OrganizationMediaFindByOrganization(context context.Context, organizationID uuid.UUID) ([]*OrganizationMedia, error) {
	return m.OrganizationMediaManager().Find(context, &OrganizationMedia{
		OrganizationID: organizationID,
	})
}

func (m *Core) OrganizationMediaCreateForOrganization(context context.Context, organizationID uuid.UUID, mediaID uuid.UUID, name string, description *string) (*OrganizationMedia, error) {
	organizationMedia := &OrganizationMedia{
		Name:           name,
		Description:    description,
		OrganizationID: organizationID,
		MediaID:        mediaID,
	}

	if err := m.OrganizationMediaManager().Create(context, organizationMedia); err != nil {
		return nil, err
	}

	return organizationMedia, nil
}
