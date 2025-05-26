package seeder

import (
	"context"
	"time"

	"github.com/lands-horizon/horizon-server/src"
	"github.com/lands-horizon/horizon-server/src/model"
)

type Seeder struct {
	provider *src.Provider
	model    *model.Model
}

func NewSeeder(provider *src.Provider, model *model.Model) (*Seeder, error) {
	return &Seeder{
		provider: provider,
		model:    model,
	}, nil
}

func (s *Seeder) Run(ctx context.Context) error {
	if err := s.SeedSubscription(ctx); err != nil {
		return err
	}
	if err := s.SeedCategory(ctx); err != nil {
		return err
	}
	return nil
}

func (s *Seeder) SeedCategory(ctx context.Context) error {
	category, err := s.model.CategoryManager.List(ctx)
	if err != nil {
		return err
	}
	if len(category) >= 1 {
		return nil
	}

	categories := []model.Category{
		{
			Name:        "Loaning",
			Description: "Loan-related cooperative services",
			Color:       "#FF5733",
			Icon:        "loan",
			CreatedAt:   time.Now().UTC(),
			UpdatedAt:   time.Now().UTC(),
		},
		{
			Name:        "Membership",
			Description: "Member registration and benefits",
			Color:       "#33C1FF",
			Icon:        "user-group",
			CreatedAt:   time.Now().UTC(),
			UpdatedAt:   time.Now().UTC(),
		},
		{
			Name:        "Team Building",
			Description: "Events and programs to strengthen teamwork",
			Color:       "#33FF6F",
			Icon:        "team",
			CreatedAt:   time.Now().UTC(),
			UpdatedAt:   time.Now().UTC(),
		},
		{
			Name:        "Farming",
			Description: "Agricultural and farming initiatives",
			Color:       "#A3D633",
			Icon:        "tractor",
			CreatedAt:   time.Now().UTC(),
			UpdatedAt:   time.Now().UTC(),
		},
		{
			Name:        "Technology",
			Description: "Tech support and infrastructure",
			Color:       "#8E44AD",
			Icon:        "chip",
			CreatedAt:   time.Now().UTC(),
			UpdatedAt:   time.Now().UTC(),
		},
		{
			Name:        "Education",
			Description: "Training and educational programs",
			Color:       "#FFC300",
			Icon:        "book-open",
			CreatedAt:   time.Now().UTC(),
			UpdatedAt:   time.Now().UTC(),
		},
		{
			Name:        "Livelihood",
			Description: "Community livelihood support",
			Color:       "#2ECC71",
			Icon:        "briefcase",
			CreatedAt:   time.Now().UTC(),
			UpdatedAt:   time.Now().UTC(),
		},
	}

	for _, category := range categories {
		if err := s.model.CategoryManager.Create(ctx, &category); err != nil {
			return err
		}
	}
	return nil
}

func (s *Seeder) SeedSubscription(ctx context.Context) error {
	subscriptionPlan, err := s.model.SubscriptionPlanManager.List(ctx)
	if err != nil {
		return err
	}
	if len(subscriptionPlan) >= 1 {
		return nil
	}
	subscriptionPlans := []model.SubscriptionPlan{
		{
			Name:                "Basic Plan",
			Description:         "A basic plan with limited features.",
			Cost:                99.99,
			Timespan:            12, // 12 months
			MaxBranches:         5,
			MaxEmployees:        50,
			MaxMembersPerBranch: 5,
			Discount:            5.00,  // 5% discount
			YearlyDiscount:      10.00, // 10% yearly discount
		},
		{
			Name:                "Pro Plan",
			Description:         "A professional plan with additional features.",
			Cost:                199.99,
			Timespan:            12, // 12 months
			MaxBranches:         10,
			MaxEmployees:        100,
			MaxMembersPerBranch: 10,
			Discount:            10.00, // 10% discount
			YearlyDiscount:      15.00, // 15% yearly discount
		},
		{
			Name:                "Enterprise Plan",
			Description:         "An enterprise-level plan with unlimited features.",
			Cost:                499.99,
			Timespan:            12, // 12 months
			MaxBranches:         20,
			MaxEmployees:        500,
			MaxMembersPerBranch: 50,
			Discount:            15.00, // 15% discount
			YearlyDiscount:      20.00, // 20% yearly discount
		},
	}
	for _, subscriptionPlan := range subscriptionPlans {
		if err := s.model.SubscriptionPlanManager.Create(ctx, &subscriptionPlan); err != nil {
			return err // optionally log and continue
		}
	}
	return nil
}
