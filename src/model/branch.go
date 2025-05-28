package model

import (
	"context"
	"time"

	"github.com/google/uuid"
	horizon_services "github.com/lands-horizon/horizon-server/services"
	"gorm.io/gorm"
)

type (
	Branch struct {
		ID             uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
		CreatedAt      time.Time      `gorm:"not null;default:now()"`
		CreatedByID    uuid.UUID      `gorm:"type:uuid"`
		CreatedBy      *User          `gorm:"foreignKey:CreatedByID;constraint:OnDelete:SET NULL;" json:"created_by,omitempty"`
		UpdatedAt      time.Time      `gorm:"not null;default:now()"`
		UpdatedByID    uuid.UUID      `gorm:"type:uuid"`
		UpdatedBy      *User          `gorm:"foreignKey:UpdatedByID;constraint:OnDelete:SET NULL;" json:"updated_by,omitempty"`
		DeletedAt      gorm.DeletedAt `gorm:"index"`
		DeletedByID    *uuid.UUID     `gorm:"type:uuid"`
		DeletedBy      *User          `gorm:"foreignKey:DeletedByID;constraint:OnDelete:SET NULL;" json:"deleted_by,omitempty"`
		OrganizationID uuid.UUID      `gorm:"type:uuid;not null" json:"organization_id"`
		Organization   *Organization  `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE;" json:"organization,omitempty"`
		MediaID        *uuid.UUID     `gorm:"type:uuid"`
		Media          *Media         `gorm:"foreignKey:MediaID;constraint:OnDelete:SET NULL;" json:"media,omitempty"`

		Type          string   `gorm:"type:varchar(100);not null"`
		Name          string   `gorm:"type:varchar(255);not null"`
		Email         string   `gorm:"type:varchar(255);not null"`
		Description   *string  `gorm:"type:text"`
		CountryCode   string   `gorm:"type:varchar(10);not null"`
		ContactNumber *string  `gorm:"type:varchar(20)"`
		Address       string   `gorm:"type:varchar(500);not null"`
		Province      string   `gorm:"type:varchar(100);not null"`
		City          string   `gorm:"type:varchar(100);not null"`
		Region        string   `gorm:"type:varchar(100);not null"`
		Barangay      string   `gorm:"type:varchar(100);not null"`
		PostalCode    string   `gorm:"type:varchar(20);not null"`
		Latitude      *float64 `gorm:"type:double precision" json:"latitude,omitempty"`
		Longitude     *float64 `gorm:"type:double precision" json:"longitude,omitempty"`
		IsMainBranch  bool     `gorm:"not null;default:false"`

		Footsteps           []*Footstep           `gorm:"foreignKey:BranchID" json:"footsteps,omitempty"`            // Footsteps
		GeneratedReports    []*GeneratedReport    `gorm:"foreignKey:BranchID" json:"generated_reports,omitempty"`    // Generated reports
		InvitationCodes     []*InvitationCode     `gorm:"foreignKey:BranchID" json:"invitation_codes,omitempty"`     // Invitation codes
		PermissionTemplates []*PermissionTemplate `gorm:"foreignKey:BranchID" json:"permission_templates,omitempty"` // permission templates
		UserOrganizations   []*UserOrganization   `gorm:"foreignKey:BranchID" json:"user_organizations,omitempty"`   // user organizations
	}

	BranchRequest struct {
		ID *uuid.UUID `json:"id,omitempty"`

		MediaID       *uuid.UUID `json:"media_id,omitempty"`
		Type          string     `json:"type" validate:"required"`
		Name          string     `json:"name" validate:"required"`
		Email         string     `json:"email" validate:"required,email"`
		Description   *string    `json:"description,omitempty"`
		CountryCode   string     `json:"country_code" validate:"required"`
		ContactNumber *string    `json:"contact_number,omitempty"`
		Address       string     `json:"address" validate:"required"`
		Province      string     `json:"province" validate:"required"`
		City          string     `json:"city" validate:"required"`
		Region        string     `json:"region" validate:"required"`
		Barangay      string     `json:"barangay" validate:"required"`
		PostalCode    string     `json:"postal_code" validate:"required"`
		Latitude      *float64   `json:"latitude,omitempty"`
		Longitude     *float64   `json:"longitude,omitempty"`

		IsMainBranch bool `json:"is_main_branch,omitempty"`
	}

	BranchResponse struct {
		ID           uuid.UUID             `json:"id"`
		CreatedAt    string                `json:"created_at"`
		CreatedByID  uuid.UUID             `json:"created_by_id"`
		CreatedBy    *UserResponse         `json:"created_by,omitempty"`
		UpdatedAt    string                `json:"updated_at"`
		UpdatedByID  uuid.UUID             `json:"updated_by_id"`
		UpdatedBy    *UserResponse         `json:"updated_by,omitempty"`
		Organization *OrganizationResponse `json:"organization,omitempty"`

		MediaID       *uuid.UUID     `json:"media_id,omitempty"`
		Media         *MediaResponse `json:"media,omitempty"`
		Type          string         `json:"type"`
		Name          string         `json:"name"`
		Email         string         `json:"email"`
		Description   *string        `json:"description,omitempty"`
		CountryCode   string         `json:"country_code"`
		ContactNumber *string        `json:"contact_number,omitempty"`
		Address       string         `json:"address"`
		Province      string         `json:"province"`
		City          string         `json:"city"`
		Region        string         `json:"region"`
		Barangay      string         `json:"barangay"`
		PostalCode    string         `json:"postal_code"`
		Latitude      *float64       `json:"latitude,omitempty"`
		Longitude     *float64       `json:"longitude,omitempty"`

		IsMainBranch bool `json:"is_main_branch,omitempty"`

		Footsteps           []*FootstepResponse           `json:"footsteps,omitempty"`
		GeneratedReports    []*GeneratedReportResponse    `json:"generated_reports,omitempty"`
		InvitationCodes     []*InvitationCodeResponse     `json:"invitation_codes,omitempty"`
		PermissionTemplates []*PermissionTemplateResponse `json:"permission_templates,omitempty"`
		UserOrganizations   []*UserOrganizationResponse   `json:"user_organizations,omitempty"`
	}
)

func (m *Model) Branch() {
	m.Migration = append(m.Migration, &Branch{})
	m.BranchManager = horizon_services.NewRepository(horizon_services.RepositoryParams[Branch, BranchResponse, BranchRequest]{
		Preloads: []string{
			"Media",
			"CreatedBy",
			"UpdatedBy",
			"Footsteps",
			"GeneratedReports",
			"InvitationCodes",
			"PermissionTemplates",
			"UserOrganizations",
			"Organization",
			"Organization.Media",
			"Organization.CreatedBy",
			"Organization.Media",
			"Organization.CoverMedia",
		},
		Service: m.provider.Service,
		Resource: func(data *Branch) *BranchResponse {
			if data == nil {
				return nil
			}
			return &BranchResponse{
				ID:           data.ID,
				CreatedAt:    data.CreatedAt.Format(time.RFC3339),
				CreatedByID:  data.CreatedByID,
				CreatedBy:    m.UserManager.ToModel(data.CreatedBy),
				UpdatedAt:    data.UpdatedAt.Format(time.RFC3339),
				UpdatedByID:  data.UpdatedByID,
				UpdatedBy:    m.UserManager.ToModel(data.UpdatedBy),
				Organization: m.OrganizationManager.ToModel(data.Organization),

				MediaID:       data.MediaID,
				Media:         m.MediaManager.ToModel(data.Media),
				Type:          data.Type,
				Name:          data.Name,
				Email:         data.Email,
				Description:   data.Description,
				CountryCode:   data.CountryCode,
				ContactNumber: data.ContactNumber,
				Address:       data.Address,
				Province:      data.Province,
				City:          data.City,
				Region:        data.Region,
				Barangay:      data.Barangay,
				PostalCode:    data.PostalCode,
				Latitude:      data.Latitude,
				Longitude:     data.Longitude,

				IsMainBranch: data.IsMainBranch,

				Footsteps:           m.FootstepManager.ToModels(data.Footsteps),
				GeneratedReports:    m.GeneratedReportManager.ToModels(data.GeneratedReports),
				InvitationCodes:     m.InvitationCodeManager.ToModels(data.InvitationCodes),
				PermissionTemplates: m.PermissionTemplateManager.ToModels(data.PermissionTemplates),
				UserOrganizations:   m.UserOrganizationManager.ToModels(data.UserOrganizations),
			}
		},
	})
}

func (m *Model) GetBranchesByOrganization(context context.Context, organizationId uuid.UUID) ([]*Branch, error) {
	return m.BranchManager.Find(context, &Branch{OrganizationID: organizationId})
}

func (m *Model) GetBranchesByOrganizationCount(context context.Context, organizationId uuid.UUID) (int64, error) {
	return m.BranchManager.Count(context, &Branch{OrganizationID: organizationId})
}
