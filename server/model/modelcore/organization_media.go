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
	// OrganizationMedia represents media files associated with cooperative organizations
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
		Media   Media     `gorm:"foreignKey:MediaID;constraint:OnDelete:CASCADE;" json:"media,omitempty"`
	}

	// OrganizationMediaResponse represents the response structure for organization media data
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

	// OrganizationMediaRequest represents the request structure for creating organization media associations
	OrganizationMediaRequest struct {
		Name        string  `json:"name" validate:"required,min=1,max=255"`
		Description *string `json:"description,omitempty"`

		OrganizationID uuid.UUID `json:"organization_id" validate:"required,uuid4"`
		MediaID        uuid.UUID `json:"media_id" validate:"required,uuid4"`
	}
)

// OrganizationMedia initializes the organization media model and its repository manager
func (m *ModelCore) organizationMedia() {
	m.migration = append(m.migration, &OrganizationMedia{})
	m.organizationMediaManager = horizon_services.NewRepository(horizon_services.RepositoryParams[OrganizationMedia, OrganizationMediaResponse, OrganizationMediaRequest]{
		Preloads: []string{"Media"},
		Service:  m.provider.Service,
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
				Organization:   m.organizationManager.ToModel(&data.Organization),

				MediaID: data.MediaID,
				Media:   m.mediaManager.ToModel(&data.Media),
			}
		},
		Created: func(data *OrganizationMedia) []string {
			return []string{
				"organization_media.create",
				fmt.Sprintf("organization_media.create.%s", data.ID),
				fmt.Sprintf("organization_media.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *OrganizationMedia) []string {
			return []string{
				"organization_media.update",
				fmt.Sprintf("organization_media.update.%s", data.ID),
				fmt.Sprintf("organization_media.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *OrganizationMedia) []string {
			return []string{
				"organization_media.delete",
				fmt.Sprintf("organization_media.delete.%s", data.ID),
				fmt.Sprintf("organization_media.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

// OrganizationMediaFindByOrganization retrieves all media files associated with a specific organization
func (m *ModelCore) organizationMediaFindByOrganization(context context.Context, organizationID uuid.UUID) ([]*OrganizationMedia, error) {
	return m.organizationMediaManager.Find(context, &OrganizationMedia{
		OrganizationID: organizationID,
	})
}

// OrganizationMediaCreateForOrganization creates a new media association for an organization
func (m *ModelCore) organizationMediaCreateForOrganization(context context.Context, organizationID uuid.UUID, mediaID uuid.UUID, name string, description *string) (*OrganizationMedia, error) {
	organizationMedia := &OrganizationMedia{
		Name:           name,
		Description:    description,
		OrganizationID: organizationID,
		MediaID:        mediaID,
	}

	if err := m.organizationMediaManager.Create(context, organizationMedia); err != nil {
		return nil, err
	}

	return organizationMedia, nil
}

// OrganizationMediaDeleteByID deletes an organization media association by its ID
func (m *ModelCore) organizationMediaDeleteByID(context context.Context, id uuid.UUID) error {
	return m.organizationMediaManager.DeleteByID(context, id)
}
