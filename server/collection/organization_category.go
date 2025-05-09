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
	OrganizationCategory struct {
		ID        uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
		CreatedAt time.Time
		UpdatedAt time.Time
		DeletedAt gorm.DeletedAt `gorm:"index"`

		OrganizationID *uuid.UUID    `gorm:"type:uuid;not null"`
		Organization   *Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE"`

		CategoryID *uuid.UUID `gorm:"type:uuid;not null"`
		Category   *Category  `gorm:"foreignKey:CategoryID;constraint:OnDelete:CASCADE"`
	}

	OrganizationCategoryRequest struct {
		OrganizationID uuid.UUID `json:"organization_id" validate:"required,uuid4"`
		CategoryID     uuid.UUID `json:"category_id" validate:"required,uuid4"`
	}

	OrganizationCategoryResponse struct {
		ID        uuid.UUID `json:"id"`
		CreatedAt string    `json:"created_at"`
		UpdatedAt string    `json:"updated_at"`
		DeletedAt *string   `json:"deleted_at,omitempty"`

		OrganizationID *uuid.UUID            `json:"organization_id"`
		Organization   *OrganizationResponse `json:"organization"`
		CategoryID     *uuid.UUID            `json:"category_id"`
		Category       *CategoryResponse     `json:"category"`
	}

	OrganizationCategoryCollection struct {
		validator    *validator.Validate
		organization *OrganizationCollection
		category     *CategoryCollection
	}
)

func NewOrganizationCategoryCollection(
	organization *OrganizationCollection,
) *OrganizationCategoryCollection {
	return &OrganizationCategoryCollection{
		validator:    validator.New(),
		organization: organization,
	}
}

func (c *OrganizationCategoryCollection) ValidateCreate(ctx echo.Context) (*OrganizationCategoryRequest, error) {
	req := new(OrganizationCategoryRequest)
	if err := ctx.Bind(req); err != nil {
		return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := c.validator.Struct(req); err != nil {
		return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return req, nil
}

func (c *OrganizationCategoryCollection) ToModel(req *OrganizationCategory) *OrganizationCategoryResponse {
	return &OrganizationCategoryResponse{
		OrganizationID: req.OrganizationID,
		Organization:   c.organization.ToModel(req.Organization),
		CategoryID:     req.CategoryID,
		Category:       c.category.ToModel(req.Category),
	}
}

func (fc *OrganizationCategoryCollection) ToModels(data []*OrganizationCategory) []*OrganizationCategoryResponse {
	if data == nil {
		return make([]*OrganizationCategoryResponse, 0)
	}
	var organizationCategory []*OrganizationCategoryResponse
	for _, value := range data {
		model := fc.ToModel(value)
		if model != nil {
			organizationCategory = append(organizationCategory, model)
		}
	}
	if len(organizationCategory) <= 0 {
		return make([]*OrganizationCategoryResponse, 0)
	}
	return organizationCategory
}
