package modelcore

import (
	"fmt"
	"time"

	horizon_services "github.com/Lands-Horizon-Corp/e-coop-server/services"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type (
	ContactUs struct {
		ID            uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
		FirstName     string         `gorm:"type:varchar(255);not null"`
		LastName      string         `gorm:"type:varchar(255)"`
		Email         string         `gorm:"type:varchar(255)"`
		ContactNumber string         `gorm:"type:varchar(20)"`
		Description   string         `gorm:"type:text;not null"`
		CreatedAt     time.Time      `gorm:"not null;default:now()"`
		UpdatedAt     time.Time      `gorm:"not null;default:now()"`
		DeletedAt     gorm.DeletedAt `gorm:"index"`
	}

	ContactUsResponse struct {
		ID            uuid.UUID `json:"id"`
		FirstName     string    `json:"first_name"`
		LastName      string    `json:"last_name,omitempty"`
		Email         string    `json:"email,omitempty"`
		ContactNumber string    `json:"contact_number,omitempty"`
		Description   string    `json:"description"`
		CreatedAt     string    `json:"created_at"`
		UpdatedAt     string    `json:"updated_at"`
	}

	ContactUsRequest struct {
		ID            *uuid.UUID `json:"id,omitempty"`
		FirstName     string     `json:"first_name" validate:"required,min=1,max=255"`
		LastName      string     `json:"last_name,omitempty" validate:"omitempty,min=1,max=255"`
		Email         string     `json:"email,omitempty" validate:"omitempty,email,max=255"`
		ContactNumber string     `json:"contact_number,omitempty" validate:"omitempty,min=1,max=20"`
		Description   string     `json:"description" validate:"required,min=1"`
	}
)

func (m *ModelCore) contactUs() {
	m.migration = append(m.migration, &ContactUs{})
	m.contactUsManager = horizon_services.NewRepository(horizon_services.RepositoryParams[ContactUs, ContactUsResponse, ContactUsRequest]{
		Preloads: nil,
		Service:  m.provider.Service,
		Resource: func(cu *ContactUs) *ContactUsResponse {
			if cu == nil {
				return nil
			}
			return &ContactUsResponse{
				ID:            cu.ID,
				FirstName:     cu.FirstName,
				LastName:      cu.LastName,
				Email:         cu.Email,
				ContactNumber: cu.ContactNumber,
				Description:   cu.Description,
				CreatedAt:     cu.CreatedAt.Format(time.RFC3339),
				UpdatedAt:     cu.UpdatedAt.Format(time.RFC3339),
			}
		},
		Created: func(cu *ContactUs) []string {
			return []string{
				"contact_us.create",
				fmt.Sprintf("feedback.create.%s", cu.ID),
			}
		},
		Deleted: func(cu *ContactUs) []string {
			return []string{
				"contact_us.delete",
				fmt.Sprintf("feedback.delete.%s", cu.ID),
			}
		},
		Updated: func(cu *ContactUs) []string {
			return []string{
				"contact_us.update",
				fmt.Sprintf("feedback.update.%s", cu.ID),
			}
		},
	})
}
