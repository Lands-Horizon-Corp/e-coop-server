package types

import (
	"time"

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
