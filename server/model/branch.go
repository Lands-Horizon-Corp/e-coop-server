package model

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
	"horizon.com/server/horizon"
)

type (
	Branch struct {
		ID             uuid.UUID      `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
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

		Footsteps           []*Footstep           `gorm:"foreignKey:BranchID" json:"footsteps,omitempty"`            // Footsteps
		GeneratedReports    []*GeneratedReport    `gorm:"foreignKey:BranchID" json:"generated_reports,omitempty"`    // Generated reports
		InvitationCodes     []*InvitationCode     `gorm:"foreignKey:BranchID" json:"invitation_codes,omitempty"`     // Invitation codes
		PermissionTemplates []*PermissionTemplate `gorm:"foreignKey:BranchID" json:"permission_templates,omitempty"` // permission templates
		UserOrganizations   []*UserOrganization   `gorm:"foreignKey:BranchID" json:"user_organizations,omitempty"`   // user organizations
	}

	BranchRequest struct {
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
	}

	BranchResponse struct {
		ID          uuid.UUID     `json:"id"`
		CreatedAt   string        `json:"created_at"`
		CreatedByID uuid.UUID     `json:"created_by_id"`
		CreatedBy   *UserResponse `json:"created_by,omitempty"`
		UpdatedAt   string        `json:"updated_at"`
		UpdatedByID uuid.UUID     `json:"updated_by_id"`
		UpdatedBy   *UserResponse `json:"updated_by,omitempty"`

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

		Footsteps           []*FootstepResponse           `json:"footsteps,omitempty"`
		GeneratedReports    []*GeneratedReportResponse    `json:"generated_reports,omitempty"`
		InvitationCodes     []*InvitationCodeResponse     `json:"invitation_codes,omitempty"`
		PermissionTemplates []*PermissionTemplateResponse `json:"permission_templates,omitempty"`
		UserOrganizations   []*UserOrganizationResponse   `json:"user_organizations,omitempty"`
	}
	BranchCollection struct {
		Manager CollectionManager[Branch]
	}
)

func (m *Model) BranchModel(data *Branch) *BranchResponse {
	return ToModel(data, func(data *Branch) *BranchResponse {
		return &BranchResponse{
			ID:          data.ID,
			CreatedAt:   data.CreatedAt.Format(time.RFC3339),
			CreatedByID: data.CreatedByID,
			CreatedBy:   m.UserModel(data.CreatedBy),
			UpdatedAt:   data.UpdatedAt.Format(time.RFC3339),
			UpdatedByID: data.UpdatedByID,
			UpdatedBy:   m.UserModel(data.UpdatedBy),

			MediaID:       data.MediaID,
			Media:         m.MediaModel(data.Media),
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

			Footsteps:           m.FootstepModels(data.Footsteps),
			GeneratedReports:    m.GeneratedReportModels(data.GeneratedReports),
			InvitationCodes:     m.InvitationCodeModels(data.InvitationCodes),
			PermissionTemplates: m.PermissionTemplateModels(data.PermissionTemplates),
			UserOrganizations:   m.UserOrganizationModels(data.UserOrganizations),
		}
	})
}

func (m *Model) BranchValidate(ctx echo.Context) (*BranchRequest, error) {
	return Validate[BranchRequest](ctx, m.validator)
}

func (m *Model) BranchModels(data []*Branch) []*BranchResponse {
	return ToModels(data, m.BranchModel)
}

func NewBranchCollection(
	broadcast *horizon.HorizonBroadcast,
	database *horizon.HorizonDatabase,
	model *Model,
) (*BranchCollection, error) {
	manager := NewcollectionManager(
		database,
		broadcast,
		func(data *Branch) ([]string, any) {
			return []string{
				"branch.create",
				fmt.Sprintf("branch.create.%s", data.ID),
				fmt.Sprintf("branch.create.organization.%s", data.OrganizationID),
			}, model.BranchModel(data)
		},
		func(data *Branch) ([]string, any) {
			return []string{
				"branch.update",
				fmt.Sprintf("branch.update.%s", data.ID),
				fmt.Sprintf("branch.update.organization.%s", data.OrganizationID),
			}, model.BranchModel(data)
		},
		func(data *Branch) ([]string, any) {
			return []string{
				"branch.delete",
				fmt.Sprintf("branch.delete.%s", data.ID),
				fmt.Sprintf("branch.delete.organization.%s", data.OrganizationID),
			}, model.BranchModel(data)
		},
		[]string{"CreatedBy", "UpdatedBy", "Organization", "Media"},
	)
	return &BranchCollection{
		Manager: manager,
	}, nil
}

// branch/organization/:organization_id
func (bc *BranchCollection) ByOrganizations(orgID uuid.UUID) ([]*Branch, error) {
	return bc.Manager.Find(&Branch{
		OrganizationID: orgID,
	})
}
