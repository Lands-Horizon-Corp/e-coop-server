package collection

import (
	"time"

	"github.com/go-playground/validator"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type (
	GeneratedReport struct {
		ID uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`

		UserID         *uuid.UUID    `gorm:"type:uuid;not null"`
		User           *User         `gorm:"foreignKey:UserID"`
		OrganizationID *uuid.UUID    `gorm:"type:uuid;not null"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID"`
		BranchID       *uuid.UUID    `gorm:"type:uuid;not null"`
		Branch         *Branch       `gorm:"foreignKey:BranchID"`
		MediaID        *uuid.UUID    `gorm:"type:uuid;not null"`
		Media          *Media        `gorm:"foreignKey:MediaID"`

		Status    string         `gorm:"type:varchar(50);not null"`
		Progress  float64        `gorm:"not null"`
		CreatedAt time.Time      `gorm:"not null;default:now()"`
		UpdatedAt time.Time      `gorm:"not null;default:now()"`
		DeletedAt gorm.DeletedAt `gorm:"index"`
	}

	GeneratedReportRequest struct {
		UserID         *uuid.UUID `json:"userId" validate:"required"`
		OrganizationID *uuid.UUID `json:"organizationId" validate:"required"`
		BranchID       *uuid.UUID `json:"branchId" validate:"required"`
		Status         *string    `json:"status" validate:"required,min=1,max=50"`
		Progress       *float64   `json:"progress" validate:"required"`
	}

	GeneratedReportResponse struct {
		ID             uuid.UUID             `json:"id"`
		UserID         *uuid.UUID            `json:"user_id,omitempty"`
		User           *UserResponse         `json:"user"`
		OrganizationID *uuid.UUID            `json:"organization_id,omitempty"`
		Organization   *OrganizationResponse `json:"organization"`
		BranchID       *uuid.UUID            `json:"branch_id,omitempty"`
		Branch         *BranchResponse       `json:"branch"`
		MediaID        *uuid.UUID            `json:"media_id,omitempty"`
		Media          *MediaResponse        `json:"media"`

		CreatedAt string  `json:"createdAt"`
		UpdatedAt string  `json:"updatedAt"`
		Status    string  `json:"status"`
		Progress  float64 `json:"progress"`
	}

	GeneratedReportCollection struct {
		validator    *validator.Validate
		user         *UserCollection
		media        *MediaCollection
		organization *OrganizationCollection
		branch       *BranchCollection
	}
)

func NewGeneratedReportCollection(
	user *UserCollection,
	media *MediaCollection,
	organization *OrganizationCollection,
	branch *BranchCollection,
) *GeneratedReportCollection {
	return &GeneratedReportCollection{
		validator:    validator.New(),
		user:         user,
		media:        media,
		organization: organization,
		branch:       branch,
	}
}

func (c *GeneratedReportCollection) ToModel(m *GeneratedReport) *GeneratedReportResponse {
	if m == nil {
		return nil
	}

	return &GeneratedReportResponse{
		ID:             m.ID,
		UserID:         m.UserID,
		User:           c.user.ToModel(m.User),
		OrganizationID: m.OrganizationID,
		Organization:   c.organization.ToModel(m.Organization),
		BranchID:       m.BranchID,
		Branch:         c.branch.ToModel(m.Branch),
		MediaID:        m.MediaID,
		Media:          c.media.ToModel(m.Media),
		CreatedAt:      m.CreatedAt.Format(time.RFC3339),
		UpdatedAt:      m.UpdatedAt.Format(time.RFC3339),
		Status:         m.Status,
		Progress:       m.Progress,
	}
}

func (c *GeneratedReportCollection) ToModels(data []*GeneratedReport) []*GeneratedReportResponse {
	if data == nil {
		return []*GeneratedReportResponse{}
	}

	responses := make([]*GeneratedReportResponse, 0, len(data))
	for _, report := range data {
		if response := c.ToModel(report); response != nil {
			responses = append(responses, response)
		}
	}
	return responses
}
