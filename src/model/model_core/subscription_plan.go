package model_core

import (
	"context"
	"fmt"
	"math"
	"time"

	horizon_services "github.com/Lands-Horizon-Corp/e-coop-server/services"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type (
	SubscriptionPlan struct {
		ID        uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
		CreatedAt time.Time      `gorm:"not null;default:now()"`
		UpdatedAt time.Time      `gorm:"not null;default:now()"`
		DeletedAt gorm.DeletedAt `gorm:"index"`

		Name                string  `gorm:"type:varchar(255);not null"`
		Description         string  `gorm:"type:text;not null"`
		Cost                float64 `gorm:"type:numeric(10,2);not null"`
		Timespan            int64   `gorm:"not null"`
		MaxBranches         int     `gorm:"not null"`
		MaxEmployees        int     `gorm:"not null"`
		MaxMembersPerBranch int     `gorm:"not null"`
		Discount            float64 `gorm:"type:numeric(5,2);default:0"`
		YearlyDiscount      float64 `gorm:"type:numeric(5,2);default:0"`
		IsRecommended       bool    `gorm:"not null;default:false"`

		// Core Features
		HasAPIAccess             bool `gorm:"not null;default:false"` // False for free
		HasFlexibleOrgStructures bool `gorm:"not null;default:false"` // False for free
		HasAIEnabled             bool `gorm:"not null;default:false"`
		HasMachineLearning       bool `gorm:"not null;default:false"`

		// Limits
		MaxAPICallsPerMonth int64 `gorm:"default:0"` // 0 for unlimited

		Organizations []*Organization `gorm:"foreignKey:SubscriptionPlanID" json:"organizations,omitempty"`
		CurrencyID    *uuid.UUID      `gorm:"type:uuid"`
		Currency      *Currency       `gorm:"foreignKey:CurrencyID"`
	}

	SubscriptionPlanRequest struct {
		ID *uuid.UUID `json:"id,omitempty"`

		Name                string  `json:"name" validate:"required,min=1,max=255"`
		Description         string  `json:"description" validate:"required"`
		Cost                float64 `json:"cost" validate:"required,gt=0"`
		Timespan            int64   `json:"timespan" validate:"required,gt=0"`
		MaxBranches         int     `json:"max_branches" validate:"required,gte=0"`
		MaxEmployees        int     `json:"max_employees" validate:"required,gte=0"`
		MaxMembersPerBranch int     `json:"max_members_per_branch" validate:"required,gte=0"`
		Discount            float64 `json:"discount" validate:"gte=0"`
		IsRecommended       bool    `json:"is_recommended"`
		YearlyDiscount      float64 `json:"yearly_discount" validate:"gte=0"`

		// Core Features
		HasAPIAccess             bool `json:"has_api_access"`
		HasFlexibleOrgStructures bool `json:"has_flexible_org_structures"`
		HasAIEnabled             bool `json:"has_ai_enabled"`
		HasMachineLearning       bool `json:"has_machine_learning"`

		// Limits
		MaxAPICallsPerMonth int64 `json:"max_api_calls_per_month" validate:"gte=0"`

		Organizations []*Organization `json:"organizations,omitempty"`
		CurrencyID    *uuid.UUID      `json:"currency_id,omitempty"`
	}

	SubscriptionPlanResponse struct {
		ID                  uuid.UUID `json:"id"`
		Name                string    `json:"name"`
		Description         string    `json:"description"`
		Cost                float64   `json:"cost"`
		Timespan            int64     `json:"timespan"`
		MaxBranches         int       `json:"max_branches"`
		MaxEmployees        int       `json:"max_employees"`
		MaxMembersPerBranch int       `json:"max_members_per_branch"`
		Discount            float64   `json:"discount"`
		YearlyDiscount      float64   `json:"yearly_discount"`
		IsRecommended       bool      `json:"is_recommended"`

		// Core Features
		HasAPIAccess             bool `json:"has_api_access"`
		HasFlexibleOrgStructures bool `json:"has_flexible_org_structures"`
		HasAIEnabled             bool `json:"has_ai_enabled"`
		HasMachineLearning       bool `json:"has_machine_learning"`

		// Limits
		MaxAPICallsPerMonth int64 `json:"max_api_calls_per_month"`

		MonthlyPrice           float64 `json:"monthly_price"`
		YearlyPrice            float64 `json:"yearly_price"`
		DiscountedMonthlyPrice float64 `json:"discounted_monthly_price"`
		DiscountedYearlyPrice  float64 `json:"discounted_yearly_price"`

		CreatedAt string `json:"created_at"`
		UpdatedAt string `json:"updated_at"`

		CurrencyID *uuid.UUID        `json:"currency_id,omitempty"`
		Currency   *CurrencyResponse `json:"currency,omitempty"`
	}
)

func (m *ModelCore) SubscriptionPlanSeed(ctx context.Context) error {
	subscriptionPlans, err := m.SubscriptionPlanManager.List(ctx)
	if err != nil {
		return err
	}
	if len(subscriptionPlans) >= 1 {
		return nil
	}
	subscriptionPlansData := []SubscriptionPlan{
		{
			Name:                     "Enterprise Plan",
			Description:              "An enterprise-level plan with unlimited features, AI/ML capabilities, and priority support.",
			Cost:                     399.99,
			Timespan:                 int64(30 * 24 * time.Hour),
			MaxBranches:              50,
			MaxEmployees:             1000,
			MaxMembersPerBranch:      500,
			Discount:                 15.00, // 15% discount
			YearlyDiscount:           25.00, // 25% yearly discount
			IsRecommended:            false,
			HasAPIAccess:             true,
			HasFlexibleOrgStructures: true,
			HasAIEnabled:             true,
			HasMachineLearning:       true,
			MaxAPICallsPerMonth:      0,   // Unlimited
			CurrencyID:               nil, // Set to default USD UUID if available
		},
		{
			Name:                     "Pro Plan",
			Description:              "A professional plan perfect for growing cooperatives with AI features.",
			Cost:                     199.99,
			Timespan:                 int64(30 * 24 * time.Hour),
			MaxBranches:              20, // Increased for competitiveness
			MaxEmployees:             200,
			MaxMembersPerBranch:      100,
			Discount:                 10.00, // 10% discount
			YearlyDiscount:           20.00, // 20% yearly discount
			IsRecommended:            true,
			HasAPIAccess:             true,
			HasFlexibleOrgStructures: true,
			HasAIEnabled:             true,
			HasMachineLearning:       false,
			MaxAPICallsPerMonth:      0, // Unlimited
			CurrencyID:               nil,
		},
		{
			Name:                     "Growth Plan",
			Description:              "A balanced plan for mid-sized co-ops ready to scale with flexible structures.",
			Cost:                     99.99,
			Timespan:                 int64(30 * 24 * time.Hour),
			MaxBranches:              8,
			MaxEmployees:             75,
			MaxMembersPerBranch:      50,
			Discount:                 7.50,  // 7.5% discount
			YearlyDiscount:           17.50, // 17.5% yearly discount
			IsRecommended:            false,
			HasAPIAccess:             true,
			HasFlexibleOrgStructures: true,
			HasAIEnabled:             false,
			HasMachineLearning:       false,
			MaxAPICallsPerMonth:      10000,
			CurrencyID:               nil,
		},
		{
			Name:                     "Starter Plan",
			Description:              "An affordable plan for small organizations just getting started.",
			Cost:                     49.99,
			Timespan:                 int64(30 * 24 * time.Hour),
			MaxBranches:              3,
			MaxEmployees:             25,
			MaxMembersPerBranch:      25,
			Discount:                 5.00,
			YearlyDiscount:           15.00,
			IsRecommended:            false,
			HasAPIAccess:             true,
			HasFlexibleOrgStructures: false,
			HasAIEnabled:             false,
			HasMachineLearning:       false,
			MaxAPICallsPerMonth:      1000,
			CurrencyID:               nil,
		},
		{
			Name:                     "Free Plan",
			Description:              "A basic trial plan with essential features to get you started.",
			Cost:                     0.00,
			Timespan:                 int64(30 * 24 * time.Hour), // Extended to 30 days for better testing
			MaxBranches:              1,
			MaxEmployees:             3,
			MaxMembersPerBranch:      10,
			Discount:                 0,
			YearlyDiscount:           0,
			IsRecommended:            false,
			HasAPIAccess:             false,
			HasFlexibleOrgStructures: false,
			HasAIEnabled:             false,
			HasMachineLearning:       false,
			MaxAPICallsPerMonth:      100,
			CurrencyID:               nil,
		},
	}
	for _, subscriptionPlan := range subscriptionPlansData {
		if err := m.SubscriptionPlanManager.Create(ctx, &subscriptionPlan); err != nil {
			return err
		}
	}
	return nil
}

func (m *ModelCore) SubscriptionPlan() {
	m.Migration = append(m.Migration, &SubscriptionPlan{})
	m.SubscriptionPlanManager = horizon_services.NewRepository(horizon_services.RepositoryParams[SubscriptionPlan, SubscriptionPlanResponse, SubscriptionPlanRequest]{
		Preloads: nil,
		Service:  m.provider.Service,
		Resource: func(sp *SubscriptionPlan) *SubscriptionPlanResponse {
			if sp == nil {
				return nil
			}

			monthlyPrice := math.Round(sp.Cost*100) / 100
			yearlyPrice := math.Round(sp.Cost*12*100) / 100
			discountedMonthlyPrice := math.Round((sp.Cost*(1-sp.Discount/100))*100) / 100
			discountedYearlyPrice := math.Round((sp.Cost*12*(1-sp.YearlyDiscount/100))*100) / 100

			return &SubscriptionPlanResponse{
				ID:                  sp.ID,
				Name:                sp.Name,
				Description:         sp.Description,
				Cost:                sp.Cost,
				Timespan:            sp.Timespan,
				MaxBranches:         sp.MaxBranches,
				MaxEmployees:        sp.MaxEmployees,
				MaxMembersPerBranch: sp.MaxMembersPerBranch,
				Discount:            sp.Discount,
				YearlyDiscount:      sp.YearlyDiscount,
				IsRecommended:       sp.IsRecommended,

				// Core Features
				HasAPIAccess:             sp.HasAPIAccess,
				HasFlexibleOrgStructures: sp.HasFlexibleOrgStructures,
				HasAIEnabled:             sp.HasAIEnabled,
				HasMachineLearning:       sp.HasMachineLearning,

				// Limits
				MaxAPICallsPerMonth: sp.MaxAPICallsPerMonth,

				MonthlyPrice:           monthlyPrice,
				YearlyPrice:            yearlyPrice,
				DiscountedMonthlyPrice: discountedMonthlyPrice,
				DiscountedYearlyPrice:  discountedYearlyPrice,
				CreatedAt:              sp.CreatedAt.Format(time.RFC3339),
				UpdatedAt:              sp.UpdatedAt.Format(time.RFC3339),
				CurrencyID:             sp.CurrencyID,
				Currency:               m.CurrencyManager.ToModel(sp.Currency),
			}
		},

		Created: func(data *SubscriptionPlan) []string {
			return []string{
				"subscription_plan.create",
				fmt.Sprintf("subscription_plan.create.%s", data.ID),
			}
		},
		Updated: func(data *SubscriptionPlan) []string {
			return []string{
				"subscription_plan.update",
				fmt.Sprintf("subscription_plan.update.%s", data.ID),
			}
		},
		Deleted: func(data *SubscriptionPlan) []string {
			return []string{
				"subscription_plan.delete",
				fmt.Sprintf("subscription_plan.delete.%s", data.ID),
			}
		},
	})
}
