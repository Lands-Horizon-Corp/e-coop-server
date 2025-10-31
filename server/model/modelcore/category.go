package modelcore

import (
	"context"
	"fmt"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/services"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type (
	Category struct {
		ID        uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
		CreatedAt time.Time      `gorm:"not null;default:now()"`
		UpdatedAt time.Time      `gorm:"not null;default:now()"`
		DeletedAt gorm.DeletedAt `gorm:"index"`

		Name                   string                  `gorm:"type:varchar(255);not null"`
		Description            string                  `gorm:"type:text"`
		Color                  string                  `gorm:"type:varchar(50)"`
		Icon                   string                  `gorm:"type:varchar(50)"`
		OrganizationCategories []*OrganizationCategory `gorm:"foreignKey:CategoryID"` // organization category
	}

	// CategoryResponse represents the response structure for category data

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

	// CategoryRequest represents the request structure for creating/updating category

	CategoryRequest struct {
		ID *uuid.UUID `json:"id,omitempty"`

		Name        string `json:"name" validate:"required,min=1,max=255"`
		Description string `json:"description" validate:"required,min=1,max=2048"`
		Color       string `json:"color" validate:"required,min=1,max=50"`
		Icon        string `json:"icon" validate:"required,min=1,max=50"`
	}
)

func (m *ModelCore) categorySeed(ctx context.Context) error {
	category, err := m.CategoryManager.List(ctx)

	if err != nil {
		return err
	}
	if len(category) >= 1 {
		return nil
	}

	categories := []Category{
		{
			Name:        "Loaning",
			Description: "Loan-related cooperative services",
			Color:       "#FF5733",
			Icon:        "Money Bag",
			CreatedAt:   time.Now().UTC(),
			UpdatedAt:   time.Now().UTC(),
		},
		{
			Name:        "Membership",
			Description: "Member registration and benefits",
			Color:       "#33C1FF",
			Icon:        "User Group",
			CreatedAt:   time.Now().UTC(),
			UpdatedAt:   time.Now().UTC(),
		},
		{
			Name:        "Team Building",
			Description: "Events and programs to strengthen teamwork",
			Color:       "#33FF6F",
			Icon:        "People Group",
			CreatedAt:   time.Now().UTC(),
			UpdatedAt:   time.Now().UTC(),
		},
		{
			Name:        "Farming",
			Description: "Agricultural and farming initiatives",
			Color:       "#A3D633",
			Icon:        "Plant Growth",
			CreatedAt:   time.Now().UTC(),
			UpdatedAt:   time.Now().UTC(),
		},
		{
			Name:        "Technology",
			Description: "Tech support and infrastructure",
			Color:       "#8E44AD",
			Icon:        "Gear",
			CreatedAt:   time.Now().UTC(),
			UpdatedAt:   time.Now().UTC(),
		},
		{
			Name:        "Education",
			Description: "Training and educational programs",
			Color:       "#FFC300",
			Icon:        "Book Open",
			CreatedAt:   time.Now().UTC(),
			UpdatedAt:   time.Now().UTC(),
		},
		{
			Name:        "Livelihood",
			Description: "Community livelihood support",
			Color:       "#2ECC71",
			Icon:        "Brief Case",
			CreatedAt:   time.Now().UTC(),
			UpdatedAt:   time.Now().UTC(),
		},
		{
			Name:        "Banking & Finance",
			Description: "Banking services, deposits, and financial management",
			Color:       "#1E3A8A",
			Icon:        "Bank",
			CreatedAt:   time.Now().UTC(),
			UpdatedAt:   time.Now().UTC(),
		},
		{
			Name:        "Savings & Investment",
			Description: "Savings programs and investment opportunities",
			Color:       "#059669",
			Icon:        "Savings",
			CreatedAt:   time.Now().UTC(),
			UpdatedAt:   time.Now().UTC(),
		},
		{
			Name:        "Healthcare",
			Description: "Health services and medical assistance programs",
			Color:       "#DC2626",
			Icon:        "Shield",
			CreatedAt:   time.Now().UTC(),
			UpdatedAt:   time.Now().UTC(),
		},
		{
			Name:        "Insurance",
			Description: "Life, health, and property insurance services",
			Color:       "#7C2D12",
			Icon:        "Umbrella",
			CreatedAt:   time.Now().UTC(),
			UpdatedAt:   time.Now().UTC(),
		},
		{
			Name:        "Business Development",
			Description: "Small business loans and entrepreneurship support",
			Color:       "#B45309",
			Icon:        "Building",
			CreatedAt:   time.Now().UTC(),
			UpdatedAt:   time.Now().UTC(),
		},
		{
			Name:        "Housing & Real Estate",
			Description: "Housing loans and real estate services",
			Color:       "#6B7280",
			Icon:        "House",
			CreatedAt:   time.Now().UTC(),
			UpdatedAt:   time.Now().UTC(),
		},
		{
			Name:        "Transportation",
			Description: "Vehicle loans and transportation services",
			Color:       "#4338CA",
			Icon:        "Navigation",
			CreatedAt:   time.Now().UTC(),
			UpdatedAt:   time.Now().UTC(),
		},
		{
			Name:        "Consumer Goods",
			Description: "Appliance and consumer goods financing",
			Color:       "#9333EA",
			Icon:        "Store",
			CreatedAt:   time.Now().UTC(),
			UpdatedAt:   time.Now().UTC(),
		},
		{
			Name:        "Emergency Fund",
			Description: "Emergency financial assistance and calamity loans",
			Color:       "#EF4444",
			Icon:        "Warning",
			CreatedAt:   time.Now().UTC(),
			UpdatedAt:   time.Now().UTC(),
		},
		{
			Name:        "Senior Citizens",
			Description: "Special programs for elderly members",
			Color:       "#78716C",
			Icon:        "Crown",
			CreatedAt:   time.Now().UTC(),
			UpdatedAt:   time.Now().UTC(),
		},
		{
			Name:        "Youth Development",
			Description: "Programs for young adults and students",
			Color:       "#10B981",
			Icon:        "Graduation Cap",
			CreatedAt:   time.Now().UTC(),
			UpdatedAt:   time.Now().UTC(),
		},
		{
			Name:        "Women Empowerment",
			Description: "Programs specifically for women members",
			Color:       "#EC4899",
			Icon:        "User Plus",
			CreatedAt:   time.Now().UTC(),
			UpdatedAt:   time.Now().UTC(),
		},
		{
			Name:        "Digital Services",
			Description: "Online banking and digital financial services",
			Color:       "#3B82F6",
			Icon:        "Smartphone",
			CreatedAt:   time.Now().UTC(),
			UpdatedAt:   time.Now().UTC(),
		},
		{
			Name:        "Community Development",
			Description: "Community projects and social responsibility programs",
			Color:       "#F59E0B",
			Icon:        "Tree City",
			CreatedAt:   time.Now().UTC(),
			UpdatedAt:   time.Now().UTC(),
		},
		{
			Name:        "Funeral Services",
			Description: "Death benefits and funeral assistance",
			Color:       "#374151",
			Icon:        "Shield Fill",
			CreatedAt:   time.Now().UTC(),
			UpdatedAt:   time.Now().UTC(),
		},
		{
			Name:        "Microfinance",
			Description: "Small-scale financial services for low-income members",
			Color:       "#84CC16",
			Icon:        "Hand Coins",
			CreatedAt:   time.Now().UTC(),
			UpdatedAt:   time.Now().UTC(),
		},
		{
			Name:        "Remittance",
			Description: "Money transfer and remittance services",
			Color:       "#06B6D4",
			Icon:        "Paper Plane",
			CreatedAt:   time.Now().UTC(),
			UpdatedAt:   time.Now().UTC(),
		},
		{
			Name:        "Financial Literacy",
			Description: "Financial education and literacy programs",
			Color:       "#8B5CF6",
			Icon:        "Book Stack",
			CreatedAt:   time.Now().UTC(),
			UpdatedAt:   time.Now().UTC(),
		},
	}

	for _, category := range categories {
		if err := m.CategoryManager.Create(ctx, &category); err != nil {
			return err
		}
	}
	return nil
}

func (m *ModelCore) category() {
	m.Migration = append(m.Migration, &Category{})
	m.CategoryManager = services.NewRepository(services.RepositoryParams[Category, CategoryResponse, CategoryRequest]{
		Preloads: []string{"OrganizationCategories"},
		Service:  m.provider.Service,
		Resource: func(data *Category) *CategoryResponse {
			if data == nil {
				return nil
			}
			return &CategoryResponse{
				ID:        data.ID,
				CreatedAt: data.CreatedAt.Format(time.RFC3339),
				UpdatedAt: data.UpdatedAt.Format(time.RFC3339),

				Name:        data.Name,
				Description: data.Description,
				Color:       data.Color,
				Icon:        data.Icon,

				OrganizationCategories: m.OrganizationCategoryManager.ToModels(data.OrganizationCategories),
			}
		},
		Created: func(data *Category) []string {
			return []string{
				"category.create",
				fmt.Sprintf("category.create.%s", data.ID),
			}
		},
		Updated: func(data *Category) []string {
			return []string{
				"category.update",
				fmt.Sprintf("category.update.%s", data.ID),
			}
		},
		Deleted: func(data *Category) []string {
			return []string{
				"category.delete",
				fmt.Sprintf("category.delete.%s", data.ID),
			}
		},
	})
}
