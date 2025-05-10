package collection

import (
	"net/http"
	"time"

	"github.com/go-playground/validator"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

type (
	Branch struct {
		ID          uuid.UUID      `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
		CreatedAt   time.Time      `gorm:"not null;default:now()"`
		CreatedByID uuid.UUID      `gorm:"type:uuid"`
		CreatedBy   *User          `gorm:"foreignKey:CreatedByID;constraint:OnDelete:SET NULL;" json:"created_by,omitempty"`
		UpdatedAt   time.Time      `gorm:"not null;default:now()"`
		UpdatedByID uuid.UUID      `gorm:"type:uuid"`
		UpdatedBy   *User          `gorm:"foreignKey:UpdatedByID;constraint:OnDelete:SET NULL;" json:"updated_by,omitempty"`
		DeletedAt   gorm.DeletedAt `gorm:"index"`
		DeletedByID *uuid.UUID     `gorm:"type:uuid"`
		DeletedBy   *User          `gorm:"foreignKey:DeletedByID;constraint:OnDelete:SET NULL;" json:"deleted_by,omitempty"`

		MediaID *uuid.UUID `gorm:"type:uuid"`
		Media   *Media     `gorm:"foreignKey:MediaID;constraint:OnDelete:SET NULL;" json:"media,omitempty"`

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

		Footsteps           []*Footstep           `gorm:"foreignKey:BranchID" json:"footsteps,omitempty"`
		GeneratedReports    []*GeneratedReport    `gorm:"foreignKey:BranchID" json:"generated_reports,omitempty"`
		InvitationCodes     []*InvitationCode     `gorm:"foreignKey:BranchID" json:"invitation_codes,omitempty"`
		PermissionTemplates []*PermissionTemplate `gorm:"foreignKey:BranchID" json:"permission_templates,omitempty"`
		UserOrganizations   []*UserOrganization   `gorm:"foreignKey:BranchID" json:"user_organizations,omitempty"`
	}

	BranchRequest struct {
		OrganizationID uuid.UUID  `json:"organization_id" validate:"required"`
		MediaID        *uuid.UUID `json:"media_id,omitempty"`

		Type          string   `json:"type" validate:"required"`
		Name          string   `json:"name" validate:"required"`
		Email         string   `json:"email" validate:"required,email"`
		Description   *string  `json:"description,omitempty"`
		CountryCode   string   `json:"country_code" validate:"required"`
		ContactNumber *string  `json:"contact_number,omitempty"`
		Address       string   `json:"address" validate:"required"`
		Province      string   `json:"province" validate:"required"`
		City          string   `json:"city" validate:"required"`
		Region        string   `json:"region" validate:"required"`
		Barangay      string   `json:"barangay" validate:"required"`
		PostalCode    string   `json:"postal_code" validate:"required"`
		Latitude      *float64 `json:"latitude,omitempty"`
		Longitude     *float64 `json:"longitude,omitempty"`
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
		validator          *validator.Validate
		media              *MediaCollection
		user               *UserCollection
		footstep           *FootstepCollection
		generatedReport    *GeneratedReportCollection
		invitationCode     *InvitationCodeCollection
		permissionTemplate *PermissionTemplateCollection
		userOrganization   *UserOrganizationCollection
	}
)

func NewBranchCollection(
	user *UserCollection,
	media *MediaCollection,
	footstep *FootstepCollection,
	generatedReport *GeneratedReportCollection,
	invitationCode *InvitationCodeCollection,
	permissionTemplate *PermissionTemplateCollection,
	userOrganization *UserOrganizationCollection,
) (*BranchCollection, error) {
	return &BranchCollection{
		validator:          validator.New(),
		media:              media,
		user:               user,
		footstep:           footstep,
		generatedReport:    generatedReport,
		invitationCode:     invitationCode,
		permissionTemplate: permissionTemplate,
		userOrganization:   userOrganization,
	}, nil
}

func (bc *BranchCollection) ValidateCreate(c echo.Context) (*BranchRequest, error) {
	req := new(BranchRequest)
	if err := c.Bind(req); err != nil {
		return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := bc.validator.Struct(req); err != nil {
		return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return req, nil
}

func (bc *BranchCollection) ToModel(branch *Branch) *BranchResponse {
	if branch == nil {
		return nil
	}

	return &BranchResponse{
		ID:          branch.ID,
		CreatedAt:   branch.CreatedAt.Format(time.RFC3339),
		CreatedByID: branch.CreatedByID,
		CreatedBy:   bc.user.ToModel(branch.CreatedBy),
		UpdatedAt:   branch.UpdatedAt.Format(time.RFC3339),
		UpdatedByID: branch.UpdatedByID,
		UpdatedBy:   bc.user.ToModel(branch.UpdatedBy),

		MediaID:       branch.MediaID,
		Media:         bc.media.ToModel(branch.Media),
		Type:          branch.Type,
		Name:          branch.Name,
		Email:         branch.Email,
		Description:   branch.Description,
		CountryCode:   branch.CountryCode,
		ContactNumber: branch.ContactNumber,
		Address:       branch.Address,
		Province:      branch.Province,
		City:          branch.City,
		Region:        branch.Region,
		Barangay:      branch.Barangay,
		PostalCode:    branch.PostalCode,
		Latitude:      branch.Latitude,
		Longitude:     branch.Longitude,

		Footsteps:           bc.footstep.ToModels(branch.Footsteps),
		GeneratedReports:    bc.generatedReport.ToModels(branch.GeneratedReports),
		InvitationCodes:     bc.invitationCode.ToModels(branch.InvitationCodes),
		PermissionTemplates: bc.permissionTemplate.ToModels(branch.PermissionTemplates),
		UserOrganizations:   bc.userOrganization.ToModels(branch.UserOrganizations),
	}
}

func (bc *BranchCollection) ToModels(data []*Branch) []*BranchResponse {
	if data == nil {
		return []*BranchResponse{}
	}
	var branches []*BranchResponse
	for _, b := range data {
		res := bc.ToModel(b)
		if res != nil {
			branches = append(branches, res)
		}
	}
	return branches
}
