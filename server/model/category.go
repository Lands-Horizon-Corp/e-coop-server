package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type (
	Category struct {
		ID        uuid.UUID      `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
		CreatedAt time.Time      `gorm:"not null;default:now()"`
		UpdatedAt time.Time      `gorm:"not null;default:now()"`
		DeletedAt gorm.DeletedAt `gorm:"index"`

		Name                   string                  `gorm:"type:varchar(255);not null"`
		Description            string                  `gorm:"type:text"`
		Color                  string                  `gorm:"type:varchar(50)"`
		Icon                   string                  `gorm:"type:varchar(50)"`
		OrganizationCategories []*OrganizationCategory `gorm:"foreignKey:CategoryID"`
	}

	CategoryResponse struct {
		ID        uuid.UUID `json:"id"`
		CreatedAt string    `json:"createdAt"`
		UpdatedAt string    `json:"updatedAt"`

		Name                   string                          `json:"name"`
		Description            string                          `json:"description"`
		Color                  string                          `json:"color"`
		Icon                   string                          `json:"icon"`
		OrganizationCategories []*OrganizationCategoryResponse `json:"organizaton_categories"`
	}

	CategoryRequest struct {
		Name        string `json:"name" validate:"required,min=1,max=255"`
		Description string `json:"description" validate:"required,min=1,max=2048"`
		Color       string `json:"color" validate:"required,min=1,max=50"`
		Icon        string `json:"icon" validate:"required,min=1,max=50"`
	}
)

// func NewCategoryCollection(
// 	organizationCategory *OrganizationCategoryCollection,
// ) (*CategoryCollection, error) {
// 	return &CategoryCollection{
// 		validator:            validator.New(),
// 		organizationCategory: organizationCategory,
// 	}, nil
// }

// func (c *CategoryCollection) ValidateCreate(ctx echo.Context) (*CategoryRequest, error) {
// 	req := new(CategoryRequest)
// 	if err := ctx.Bind(req); err != nil {
// 		return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
// 	}
// 	if err := c.validator.Struct(req); err != nil {
// 		return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
// 	}
// 	return req, nil
// }
// func (c *CategoryCollection) ToModel(m *Category) *CategoryResponse {
// 	if m == nil {
// 		return nil
// 	}
// 	return &CategoryResponse{
// 		ID:                     m.ID,
// 		CreatedAt:              m.CreatedAt.Format(time.RFC3339),
// 		UpdatedAt:              m.UpdatedAt.Format(time.RFC3339),
// 		Name:                   m.Name,
// 		Description:            m.Description,
// 		Color:                  m.Color,
// 		Icon:                   m.Icon,
// 		OrganizationCategories: c.organizationCategory.ToModels(m.OrganizationCategories),
// 	}
// }

// func (oc *CategoryCollection) ToModels(data []*Category) []*CategoryResponse {
// 	if data == nil {
// 		return []*CategoryResponse{}
// 	}
// 	var out []*CategoryResponse
// 	for _, o := range data {
// 		if m := oc.ToModel(o); m != nil {
// 			out = append(out, m)
// 		}
// 	}
// 	return out
// }
