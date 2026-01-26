package core

import (
	"context"
	"fmt"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/horizon"
	"github.com/Lands-Horizon-Corp/e-coop-server/pkg/registry"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/types"
)

func categorySeed(ctx context.Context, service *horizon.HorizonService) error {
	category, err := CategoryManager(service).List(ctx)

	if err != nil {
		return err
	}
	if len(category) >= 1 {
		return nil
	}

	categories := []types.Category{
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
		if err := CategoryManager(service).Create(ctx, &category); err != nil {
			return err
		}
	}
	return nil
}

func CategoryManager(service *horizon.HorizonService) *registry.Registry[types.Category, types.CategoryResponse, types.CategoryRequest] {
	return registry.NewRegistry(registry.RegistryParams[types.Category, types.CategoryResponse, types.CategoryRequest]{
		Preloads: []string{"OrganizationCategories"},
		Database: service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return service.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *types.Category) *types.CategoryResponse {
			if data == nil {
				return nil
			}
			return &types.CategoryResponse{
				ID:        data.ID,
				CreatedAt: data.CreatedAt.Format(time.RFC3339),
				UpdatedAt: data.UpdatedAt.Format(time.RFC3339),

				Name:        data.Name,
				Description: data.Description,
				Color:       data.Color,
				Icon:        data.Icon,

				OrganizationCategories: OrganizationCategoryManager(service).ToModels(data.OrganizationCategories),
			}
		},
		Created: func(data *types.Category) registry.Topics {
			return []string{
				"category.create",
				fmt.Sprintf("category.create.%s", data.ID),
			}
		},
		Updated: func(data *types.Category) registry.Topics {
			return []string{
				"category.update",
				fmt.Sprintf("category.update.%s", data.ID),
			}
		},
		Deleted: func(data *types.Category) registry.Topics {
			return []string{
				"category.delete",
				fmt.Sprintf("category.delete.%s", data.ID),
			}
		},
	})
}
