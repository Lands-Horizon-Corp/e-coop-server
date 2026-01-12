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
	Branch struct {
		ID             uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
		CreatedAt      time.Time      `gorm:"not null;default:now()" json:"created_at"`
		CreatedByID    uuid.UUID      `gorm:"type:uuid" json:"created_by_id"`
		CreatedBy      *User          `gorm:"foreignKey:CreatedByID;constraint:OnDelete:SET NULL;" json:"created_by,omitempty"`
		UpdatedAt      time.Time      `gorm:"not null;default:now()" json:"updated_at"`
		UpdatedByID    uuid.UUID      `gorm:"type:uuid" json:"updated_by_id"`
		UpdatedBy      *User          `gorm:"foreignKey:UpdatedByID;constraint:OnDelete:SET NULL;" json:"updated_by,omitempty"`
		DeletedAt      gorm.DeletedAt `gorm:"index" json:"deleted_at"`
		DeletedByID    *uuid.UUID     `gorm:"type:uuid" json:"deleted_by_id"`
		DeletedBy      *User          `gorm:"foreignKey:DeletedByID;constraint:OnDelete:SET NULL;" json:"deleted_by,omitempty"`
		OrganizationID uuid.UUID      `gorm:"type:uuid;not null" json:"organization_id"`
		Organization   *Organization  `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE;" json:"organization,omitempty"`
		MediaID        *uuid.UUID     `gorm:"type:uuid" json:"media_id"`
		Media          *Media         `gorm:"foreignKey:MediaID;constraint:OnDelete:SET NULL;" json:"media,omitempty"`

		Type        string  `gorm:"type:varchar(100);not null" json:"type"`
		Name        string  `gorm:"type:varchar(255);not null" json:"name"`
		Email       string  `gorm:"type:varchar(255);not null" json:"email"`
		Description *string `gorm:"type:text" json:"description,omitempty"`

		CurrencyID *uuid.UUID `gorm:"type:uuid" json:"currency_id"`
		Currency   *Currency  `gorm:"foreignKey:CurrencyID;constraint:OnDelete:SET NULL;" json:"currency"`

		ContactNumber *string  `gorm:"type:varchar(20)" json:"contact_number,omitempty"`
		Address       string   `gorm:"type:varchar(500);not null" json:"address"`
		Province      string   `gorm:"type:varchar(100);not null" json:"province"`
		City          string   `gorm:"type:varchar(100);not null" json:"city"`
		Region        string   `gorm:"type:varchar(100);not null" json:"region"`
		Barangay      string   `gorm:"type:varchar(100);not null" json:"barangay"`
		PostalCode    string   `gorm:"type:varchar(20);not null" json:"postal_code"`
		Latitude      *float64 `gorm:"type:double precision" json:"latitude,omitempty"`
		Longitude     *float64 `gorm:"type:double precision" json:"longitude,omitempty"`
		IsMainBranch  bool     `gorm:"not null;default:false" json:"is_main_branch"`

		BranchSetting *BranchSetting `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE;" json:"branch_setting,omitempty"`

		Footsteps           []*Footstep           `gorm:"foreignKey:BranchID" json:"footsteps,omitempty"`
		GeneratedReports    []*GeneratedReport    `gorm:"foreignKey:BranchID" json:"generated_reports,omitempty"`
		InvitationCodes     []*InvitationCode     `gorm:"foreignKey:BranchID" json:"invitation_codes,omitempty"`
		PermissionTemplates []*PermissionTemplate `gorm:"foreignKey:BranchID" json:"permission_templates,omitempty"`
		UserOrganizations   []*UserOrganization   `gorm:"foreignKey:BranchID" json:"user_organizations,omitempty"`

		TaxIdentificationNumber *string `gorm:"type:varchar(100)" json:"tax_identification_number,omitempty"`
	}

	BranchRequest struct {
		ID *uuid.UUID `json:"id,omitempty"`

		MediaID       *uuid.UUID `json:"media_id,omitempty"`
		Type          string     `json:"type" validate:"required"`
		Name          string     `json:"name" validate:"required"`
		Email         string     `json:"email" validate:"required,email"`
		Description   *string    `json:"description,omitempty"`
		CurrencyID    *uuid.UUID `json:"currency_id" validate:"required"` // Changed from Currency string to CurrencyID
		ContactNumber *string    `json:"contact_number,omitempty"`
		Address       string     `json:"address" validate:"required"`
		Province      string     `json:"province" validate:"required"`
		City          string     `json:"city" validate:"required"`
		Region        string     `json:"region" validate:"required"`
		Barangay      string     `json:"barangay" validate:"required"`
		PostalCode    string     `json:"postal_code" validate:"required"`
		Latitude      *float64   `json:"latitude,omitempty"`
		Longitude     *float64   `json:"longitude,omitempty"`

		IsMainBranch            bool    `json:"is_main_branch,omitempty"`
		TaxIdentificationNumber *string `json:"tax_identification_number,omitempty"`
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

		MediaID       *uuid.UUID        `json:"media_id,omitempty"`
		Media         *MediaResponse    `json:"media,omitempty"`
		Type          string            `json:"type"`
		Name          string            `json:"name"`
		Email         string            `json:"email"`
		Description   *string           `json:"description,omitempty"`
		CurrencyID    *uuid.UUID        `json:"currency_id,omitempty"`
		Currency      *CurrencyResponse `json:"currency,omitempty"` // Changed from string to Currency object
		ContactNumber *string           `json:"contact_number,omitempty"`
		Address       string            `json:"address"`
		Province      string            `json:"province"`
		City          string            `json:"city"`
		Region        string            `json:"region"`
		Barangay      string            `json:"barangay"`
		PostalCode    string            `json:"postal_code"`
		Latitude      *float64          `json:"latitude,omitempty"`
		Longitude     *float64          `json:"longitude,omitempty"`

		IsMainBranch            bool    `json:"is_main_branch,omitempty"`
		TaxIdentificationNumber *string `json:"tax_identification_number,omitempty"`

		BranchSetting *BranchSettingResponse `json:"branch_setting,omitempty"`

		Footsteps           []*FootstepResponse           `json:"footsteps,omitempty"`
		GeneratedReports    []*GeneratedReportResponse    `json:"generated_reports,omitempty"`
		InvitationCodes     []*InvitationCodeResponse     `json:"invitation_codes,omitempty"`
		PermissionTemplates []*PermissionTemplateResponse `json:"permission_templates,omitempty"`
		UserOrganizations   []*UserOrganizationResponse   `json:"user_organizations,omitempty"`
	}
)

func BranchManager(service *horizon.HorizonService) *registry.Registry[Branch, BranchResponse, BranchRequest] {
	return registry.GetRegistry(registry.RegistryParams[Branch, BranchResponse, BranchRequest]{
		Preloads: []string{
			"Media",
			"CreatedBy",
			"UpdatedBy",
			"Currency",
			"BranchSetting",
			"Organization.Media",
			"Organization.CreatedBy",
			"Organization.CoverMedia",
		},
		Database: service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return service.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *Branch) *BranchResponse {
			if data == nil {
				return nil
			}
			return &BranchResponse{
				ID:           data.ID,
				CreatedAt:    data.CreatedAt.Format(time.RFC3339),
				CreatedByID:  data.CreatedByID,
				CreatedBy:    UserManager(service).ToModel(data.CreatedBy),
				UpdatedAt:    data.UpdatedAt.Format(time.RFC3339),
				UpdatedByID:  data.UpdatedByID,
				UpdatedBy:    UserManager(service).ToModel(data.UpdatedBy),
				Organization: OrganizationManager(service).ToModel(data.Organization),

				MediaID:       data.MediaID,
				Media:         MediaManager(service).ToModel(data.Media),
				Type:          data.Type,
				Name:          data.Name,
				Email:         data.Email,
				Description:   data.Description,
				CurrencyID:    data.CurrencyID,
				Currency:      CurrencyManager(service).ToModel(data.Currency), // Use the Currency relationship
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

				BranchSetting: BranchSettingManager(service).ToModel(data.BranchSetting),

				Footsteps:           FootstepManager(service).ToModels(data.Footsteps),
				GeneratedReports:    GeneratedReportManager(service).ToModels(data.GeneratedReports),
				InvitationCodes:     InvitationCodeManager(service).ToModels(data.InvitationCodes),
				PermissionTemplates: PermissionTemplateManager(service).ToModels(data.PermissionTemplates),
				UserOrganizations:   UserOrganizationManager(service).ToModels(data.UserOrganizations),

				TaxIdentificationNumber: data.TaxIdentificationNumber,
			}
		},
		Created: func(data *Branch) registry.Topics {
			return []string{
				"branch.create",
				fmt.Sprintf("branch.create.%s", data.ID),
				fmt.Sprintf("branch.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *Branch) registry.Topics {
			return []string{
				"branch.update",
				fmt.Sprintf("branch.update.%s", data.ID),
				fmt.Sprintf("branch.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *Branch) registry.Topics {
			return []string{
				"branch.delete",
				fmt.Sprintf("branch.delete.%s", data.ID),
				fmt.Sprintf("branch.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func GetBranchesByOrganization(context context.Context, service *horizon.HorizonService, organizationID uuid.UUID) ([]*Branch, error) {
	return BranchManager(service).Find(context, &Branch{OrganizationID: organizationID})
}

func GetBranchesByOrganizationCount(context context.Context, service *horizon.HorizonService, organizationID uuid.UUID) (int64, error) {
	return BranchManager(service).Count(context, &Branch{OrganizationID: organizationID})
}
