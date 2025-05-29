package model

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	horizon_services "github.com/lands-horizon/horizon-server/services"
	"gorm.io/gorm"
)

type (
	Holiday struct {
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

		OrganizationID uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_holidays"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"organization,omitempty"`
		BranchID       uuid.UUID     `gorm:"type:uuid;not null;index:idx_organization_branch_holidays"`
		Branch         *Branch       `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"branch,omitempty"`

		EntryDate   time.Time `gorm:"not null"`
		Name        string    `gorm:"type:varchar(255)"`
		Description string    `gorm:"type:text"`
	}

	HolidayResponse struct {
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
		EntryDate      string                `json:"entry_date"`
		Name           string                `json:"name"`
		Description    string                `json:"description"`
	}

	HolidayRequest struct {
		EntryDate   time.Time `json:"entry_date" validate:"required"`
		Name        string    `json:"name" validate:"required,min=1,max=255"`
		Description string    `json:"description,omitempty"`
	}
)

func (m *Model) Holiday() {
	m.Migration = append(m.Migration, &Holiday{})
	m.HolidayManager = horizon_services.NewRepository(horizon_services.RepositoryParams[
		Holiday, HolidayResponse, HolidayRequest,
	]{
		Preloads: []string{
			"CreatedBy", "UpdatedBy", "DeletedBy", "Branch", "Organization",
		},
		Service: m.provider.Service,
		Resource: func(data *Holiday) *HolidayResponse {
			if data == nil {
				return nil
			}
			return &HolidayResponse{
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
				EntryDate:      data.EntryDate.Format(time.RFC3339),
				Name:           data.Name,
				Description:    data.Description,
			}
		},
		Created: func(data *Holiday) []string {
			return []string{
				"holidays.create",
				fmt.Sprintf("holidays.create.%s", data.ID),
			}
		},
		Updated: func(data *Holiday) []string {
			return []string{
				"holidays.update",
				fmt.Sprintf("holidays.update.%s", data.ID),
			}
		},
		Deleted: func(data *Holiday) []string {
			return []string{
				"holidays.delete",
				fmt.Sprintf("holidays.delete.%s", data.ID),
			}
		},
	})
}
