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
		ID        uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
		CreatedAt time.Time
		UpdatedAt time.Time
		DeletedAt gorm.DeletedAt `gorm:"index"`

		OrganizationID uuid.UUID    `gorm:"type:uuid;not null"`
		Organization   Organization `gorm:"foreignKey:OrganizationID"`

		MediaID *uuid.UUID `gorm:"type:uuid"`
		Media   *Media     `gorm:"foreignKey:MediaID;constraint:OnDelete:SET NULL;" json:"media,omitempty"`

		Type          string  `gorm:"type:varchar(100);not null"`
		Name          string  `gorm:"type:varchar(255);not null"`
		Email         string  `gorm:"type:varchar(255);not null"`
		Description   *string `gorm:"type:text"`
		CountryCode   string  `gorm:"type:varchar(10);not null"`
		ContactNumber *string `gorm:"type:varchar(20)"`

		Address    string `gorm:"type:varchar(500);not null"`
		Province   string `gorm:"type:varchar(100);not null"`
		City       string `gorm:"type:varchar(100);not null"`
		Region     string `gorm:"type:varchar(100);not null"`
		Barangay   string `gorm:"type:varchar(100);not null"`
		PostalCode string `gorm:"type:varchar(20);not null"`

		Latitude  *float64
		Longitude *float64

		IsMainBranch    bool
		IsAdminVerified *bool
	}

	BranchRequest struct {
		OrganizationID uuid.UUID  `json:"organization_id" validate:"required"`
		MediaID        *uuid.UUID `json:"media_id,omitempty"`

		Type          string  `json:"type" validate:"required"`
		Name          string  `json:"name" validate:"required"`
		Email         string  `json:"email" validate:"required,email"`
		Description   *string `json:"description,omitempty"`
		CountryCode   string  `json:"country_code" validate:"required"`
		ContactNumber *string `json:"contact_number,omitempty"`

		Address    string `json:"address" validate:"required"`
		Province   string `json:"province" validate:"required"`
		City       string `json:"city" validate:"required"`
		Region     string `json:"region" validate:"required"`
		Barangay   string `json:"barangay" validate:"required"`
		PostalCode string `json:"postal_code" validate:"required"`

		Latitude        *float64 `json:"latitude,omitempty"`
		Longitude       *float64 `json:"longitude,omitempty"`
		IsMainBranch    bool     `json:"is_main_branch"`
		IsAdminVerified *bool    `json:"is_admin_verified,omitempty"`
	}

	BranchResponse struct {
		ID              uuid.UUID      `json:"id"`
		OrganizationID  uuid.UUID      `json:"organization_id"`
		MediaID         *uuid.UUID     `json:"media_id,omitempty"`
		Media           *MediaResponse `json:"media,omitempty"`
		Type            string         `json:"type"`
		Name            string         `json:"name"`
		Email           string         `json:"email"`
		Description     *string        `json:"description,omitempty"`
		CountryCode     string         `json:"country_code"`
		ContactNumber   *string        `json:"contact_number,omitempty"`
		Address         string         `json:"address"`
		Province        string         `json:"province"`
		City            string         `json:"city"`
		Region          string         `json:"region"`
		Barangay        string         `json:"barangay"`
		PostalCode      string         `json:"postal_code"`
		Latitude        *float64       `json:"latitude,omitempty"`
		Longitude       *float64       `json:"longitude,omitempty"`
		IsMainBranch    bool           `json:"is_main_branch"`
		IsAdminVerified *bool          `json:"is_admin_verified,omitempty"`
		CreatedAt       string         `json:"created_at"`
		UpdatedAt       string         `json:"updated_at"`
		DeletedAt       *string        `json:"deleted_at,omitempty"`
	}

	BranchCollection struct {
		validator *validator.Validate
		media     *MediaCollection
	}
)

func NewBranchCollection(
	media *MediaCollection,
) (*BranchCollection, error) {
	return &BranchCollection{
		validator: validator.New(),
		media:     media,
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
	var deletedAt *string
	if branch.DeletedAt.Valid {
		t := branch.DeletedAt.Time.Format(time.RFC3339)
		deletedAt = &t
	}
	return &BranchResponse{
		ID:             branch.ID,
		OrganizationID: branch.OrganizationID,
		MediaID:        branch.MediaID,
		Media:          bc.media.ToModel(branch.Media),

		Type:            branch.Type,
		Name:            branch.Name,
		Email:           branch.Email,
		Description:     branch.Description,
		CountryCode:     branch.CountryCode,
		ContactNumber:   branch.ContactNumber,
		Address:         branch.Address,
		Province:        branch.Province,
		City:            branch.City,
		Region:          branch.Region,
		Barangay:        branch.Barangay,
		PostalCode:      branch.PostalCode,
		Latitude:        branch.Latitude,
		Longitude:       branch.Longitude,
		IsMainBranch:    branch.IsMainBranch,
		IsAdminVerified: branch.IsAdminVerified,
		CreatedAt:       branch.CreatedAt.Format(time.RFC3339),
		UpdatedAt:       branch.UpdatedAt.Format(time.RFC3339),
		DeletedAt:       deletedAt,
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
