package model_core

import (
	"context"
	"fmt"
	"math"
	"time"

	horizon_services "github.com/Lands-Horizon-Corp/e-coop-server/services"
	"github.com/google/uuid"
	"github.com/rotisserie/eris"
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

const PRICING_COUNTRY_CODE = "PH"

func (m *ModelCore) SubscriptionPlanSeed(ctx context.Context) error {

	subscriptionPlans, err := m.SubscriptionPlanManager.List(ctx)
	if err != nil {
		return err
	}
	if len(subscriptionPlans) >= 1 {
		return nil
	}
	currency, err := m.CurrencyManager.List(ctx)
	if err != nil {
		return err
	}
	for _, currency := range currency {
		var subscription []*SubscriptionPlan

		switch currency.CurrencyCode {
		case "USD": // United States
			subscription = []*SubscriptionPlan{
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
					MaxAPICallsPerMonth:      0,            // Unlimited
					CurrencyID:               &currency.ID, // Set to default USD UUID if available
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
					CurrencyID:               &currency.ID,
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
					CurrencyID:               &currency.ID,
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
					CurrencyID:               &currency.ID,
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
					CurrencyID:               &currency.ID,
				},
			}
		case "EUR": // European Union (Germany as representative)
			subscription = []*SubscriptionPlan{
				{
					Name:                     "Enterprise Plan",
					Description:              "Ein Unternehmensplan mit unbegrenzten Funktionen, KI-/ML-Fähigkeiten und priorisiertem Support.",
					Cost:                     399, // €399 per month
					Timespan:                 int64(30 * 24 * time.Hour),
					MaxBranches:              50,
					MaxEmployees:             1000,
					MaxMembersPerBranch:      500,
					Discount:                 15, // 15% discount
					YearlyDiscount:           25, // 25% yearly discount
					IsRecommended:            false,
					HasAPIAccess:             true,
					HasFlexibleOrgStructures: true,
					HasAIEnabled:             true,
					HasMachineLearning:       true,
					MaxAPICallsPerMonth:      0,            // Unlimited
					CurrencyID:               &currency.ID, // EUR UUID
				},
				{
					Name:                     "Pro Plan",
					Description:              "Ein professioneller Plan, ideal für wachsende Genossenschaften mit KI-Funktionen.",
					Cost:                     199, // €199 per month
					Timespan:                 int64(30 * 24 * time.Hour),
					MaxBranches:              20,
					MaxEmployees:             200,
					MaxMembersPerBranch:      100,
					Discount:                 10,
					YearlyDiscount:           20,
					IsRecommended:            true,
					HasAPIAccess:             true,
					HasFlexibleOrgStructures: true,
					HasAIEnabled:             true,
					HasMachineLearning:       false,
					MaxAPICallsPerMonth:      0, // Unlimited
					CurrencyID:               &currency.ID,
				},
				{
					Name:                     "Growth Plan",
					Description:              "Ein ausgewogener Plan für mittelgroße Genossenschaften, die bereit sind zu wachsen.",
					Cost:                     99, // €99 per month
					Timespan:                 int64(30 * 24 * time.Hour),
					MaxBranches:              8,
					MaxEmployees:             75,
					MaxMembersPerBranch:      50,
					Discount:                 7,
					YearlyDiscount:           17,
					IsRecommended:            false,
					HasAPIAccess:             true,
					HasFlexibleOrgStructures: true,
					HasAIEnabled:             false,
					HasMachineLearning:       false,
					MaxAPICallsPerMonth:      10000,
					CurrencyID:               &currency.ID,
				},
				{
					Name:                     "Starter Plan",
					Description:              "Ein günstiger Plan für kleine Organisationen, die gerade erst anfangen.",
					Cost:                     49, // €49 per month
					Timespan:                 int64(30 * 24 * time.Hour),
					MaxBranches:              3,
					MaxEmployees:             25,
					MaxMembersPerBranch:      25,
					Discount:                 5,
					YearlyDiscount:           15,
					IsRecommended:            false,
					HasAPIAccess:             true,
					HasFlexibleOrgStructures: false,
					HasAIEnabled:             false,
					HasMachineLearning:       false,
					MaxAPICallsPerMonth:      1000,
					CurrencyID:               &currency.ID,
				},
				{
					Name:                     "Free Plan",
					Description:              "Ein kostenloser Plan mit grundlegenden Funktionen für den Einstieg.",
					Cost:                     0, // Free trial
					Timespan:                 int64(30 * 24 * time.Hour),
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
					CurrencyID:               &currency.ID,
				},
			}
		case "JPY": // Japan
			subscription = []*SubscriptionPlan{
				{
					Name:                     "エンタープライズプラン",
					Description:              "無制限の機能、AI/ML機能、そして優先サポートを備えた企業向けプランです。",
					Cost:                     50000, // ¥50,000 / month
					Timespan:                 int64(30 * 24 * time.Hour),
					MaxBranches:              50,
					MaxEmployees:             1000,
					MaxMembersPerBranch:      500,
					Discount:                 15, // 15% discount
					YearlyDiscount:           25, // 25% yearly discount
					IsRecommended:            false,
					HasAPIAccess:             true,
					HasFlexibleOrgStructures: true,
					HasAIEnabled:             true,
					HasMachineLearning:       true,
					MaxAPICallsPerMonth:      0,            // Unlimited
					CurrencyID:               &currency.ID, // JPY UUID
				},
				{
					Name:                     "プロプラン",
					Description:              "成長中の協同組合に最適な、AI機能を備えたプロフェッショナルプランです。",
					Cost:                     25000, // ¥25,000 / month
					Timespan:                 int64(30 * 24 * time.Hour),
					MaxBranches:              20,
					MaxEmployees:             200,
					MaxMembersPerBranch:      100,
					Discount:                 10,
					YearlyDiscount:           20,
					IsRecommended:            true,
					HasAPIAccess:             true,
					HasFlexibleOrgStructures: true,
					HasAIEnabled:             true,
					HasMachineLearning:       false,
					MaxAPICallsPerMonth:      0, // Unlimited
					CurrencyID:               &currency.ID,
				},
				{
					Name:                     "グロースプラン",
					Description:              "中規模の協同組合がスケールアップするためのバランスの取れたプランです。",
					Cost:                     12000, // ¥12,000 / month
					Timespan:                 int64(30 * 24 * time.Hour),
					MaxBranches:              8,
					MaxEmployees:             75,
					MaxMembersPerBranch:      50,
					Discount:                 7,
					YearlyDiscount:           17,
					IsRecommended:            false,
					HasAPIAccess:             true,
					HasFlexibleOrgStructures: true,
					HasAIEnabled:             false,
					HasMachineLearning:       false,
					MaxAPICallsPerMonth:      10000,
					CurrencyID:               &currency.ID,
				},
				{
					Name:                     "スタータープラン",
					Description:              "小規模組織のための手頃な入門プランです。",
					Cost:                     6000, // ¥6,000 / month
					Timespan:                 int64(30 * 24 * time.Hour),
					MaxBranches:              3,
					MaxEmployees:             25,
					MaxMembersPerBranch:      25,
					Discount:                 5,
					YearlyDiscount:           15,
					IsRecommended:            false,
					HasAPIAccess:             true,
					HasFlexibleOrgStructures: false,
					HasAIEnabled:             false,
					HasMachineLearning:       false,
					MaxAPICallsPerMonth:      1000,
					CurrencyID:               &currency.ID,
				},
				{
					Name:                     "無料プラン",
					Description:              "基本的な機能を備えた無料トライアルプランです。",
					Cost:                     0, // Free
					Timespan:                 int64(30 * 24 * time.Hour),
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
					CurrencyID:               &currency.ID,
				},
			}
		case "GBP": // United Kingdom
			subscription = []*SubscriptionPlan{
				{
					Name:                     "Enterprise Plan",
					Description:              "An enterprise-level plan with unlimited features, AI/ML capabilities, and priority support.",
					Cost:                     400, // GBP 400/month
					Timespan:                 int64(30 * 24 * time.Hour),
					MaxBranches:              50,
					MaxEmployees:             1000,
					MaxMembersPerBranch:      500,
					Discount:                 15, // 15% discount
					YearlyDiscount:           25, // 25% yearly discount
					IsRecommended:            false,
					HasAPIAccess:             true,
					HasFlexibleOrgStructures: true,
					HasAIEnabled:             true,
					HasMachineLearning:       true,
					MaxAPICallsPerMonth:      0, // Unlimited
					CurrencyID:               &currency.ID,
				},
				{
					Name:                     "Pro Plan",
					Description:              "A professional plan ideal for growing cooperatives with AI support.",
					Cost:                     200, // GBP 200/month
					Timespan:                 int64(30 * 24 * time.Hour),
					MaxBranches:              20,
					MaxEmployees:             200,
					MaxMembersPerBranch:      100,
					Discount:                 10, // 10% discount
					YearlyDiscount:           20, // 20% yearly discount
					IsRecommended:            true,
					HasAPIAccess:             true,
					HasFlexibleOrgStructures: true,
					HasAIEnabled:             true,
					HasMachineLearning:       false,
					MaxAPICallsPerMonth:      0, // Unlimited
					CurrencyID:               &currency.ID,
				},
				{
					Name:                     "Growth Plan",
					Description:              "A balanced plan for mid-sized co-ops aiming to expand with flexibility.",
					Cost:                     100, // GBP 100/month
					Timespan:                 int64(30 * 24 * time.Hour),
					MaxBranches:              8,
					MaxEmployees:             75,
					MaxMembersPerBranch:      50,
					Discount:                 8,  // 8% discount
					YearlyDiscount:           18, // 18% yearly discount
					IsRecommended:            false,
					HasAPIAccess:             true,
					HasFlexibleOrgStructures: true,
					HasAIEnabled:             false,
					HasMachineLearning:       false,
					MaxAPICallsPerMonth:      10000,
					CurrencyID:               &currency.ID,
				},
				{
					Name:                     "Starter Plan",
					Description:              "An affordable plan for small organizations beginning their digital journey.",
					Cost:                     50, // GBP 50/month
					Timespan:                 int64(30 * 24 * time.Hour),
					MaxBranches:              3,
					MaxEmployees:             25,
					MaxMembersPerBranch:      25,
					Discount:                 5,
					YearlyDiscount:           15,
					IsRecommended:            false,
					HasAPIAccess:             true,
					HasFlexibleOrgStructures: false,
					HasAIEnabled:             false,
					HasMachineLearning:       false,
					MaxAPICallsPerMonth:      1000,
					CurrencyID:               &currency.ID,
				},
				{
					Name:                     "Free Plan",
					Description:              "A basic plan with essential tools to help you start your cooperative journey.",
					Cost:                     0, // Free
					Timespan:                 int64(30 * 24 * time.Hour),
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
					CurrencyID:               &currency.ID,
				},
			}
		case "AUD": // Australia
			subscription = []*SubscriptionPlan{
				{
					Name:                     "Enterprise Plan",
					Description:              "An enterprise-grade plan with unlimited access, AI and machine learning tools, and top-tier support.",
					Cost:                     550, // AUD 550/month
					Timespan:                 int64(30 * 24 * time.Hour),
					MaxBranches:              50,
					MaxEmployees:             1000,
					MaxMembersPerBranch:      500,
					Discount:                 15, // 15% discount
					YearlyDiscount:           25, // 25% yearly discount
					IsRecommended:            false,
					HasAPIAccess:             true,
					HasFlexibleOrgStructures: true,
					HasAIEnabled:             true,
					HasMachineLearning:       true,
					MaxAPICallsPerMonth:      0, // Unlimited
					CurrencyID:               &currency.ID,
				},
				{
					Name:                     "Pro Plan",
					Description:              "A professional plan ideal for expanding cooperatives with AI capabilities.",
					Cost:                     280, // AUD 280/month
					Timespan:                 int64(30 * 24 * time.Hour),
					MaxBranches:              20,
					MaxEmployees:             200,
					MaxMembersPerBranch:      100,
					Discount:                 10, // 10% discount
					YearlyDiscount:           20, // 20% yearly discount
					IsRecommended:            true,
					HasAPIAccess:             true,
					HasFlexibleOrgStructures: true,
					HasAIEnabled:             true,
					HasMachineLearning:       false,
					MaxAPICallsPerMonth:      0, // Unlimited
					CurrencyID:               &currency.ID,
				},
				{
					Name:                     "Growth Plan",
					Description:              "A flexible plan for medium-sized co-ops ready to expand and modernise.",
					Cost:                     140, // AUD 140/month
					Timespan:                 int64(30 * 24 * time.Hour),
					MaxBranches:              8,
					MaxEmployees:             75,
					MaxMembersPerBranch:      50,
					Discount:                 8,  // 8% discount
					YearlyDiscount:           18, // 18% yearly discount
					IsRecommended:            false,
					HasAPIAccess:             true,
					HasFlexibleOrgStructures: true,
					HasAIEnabled:             false,
					HasMachineLearning:       false,
					MaxAPICallsPerMonth:      10000,
					CurrencyID:               &currency.ID,
				},
				{
					Name:                     "Starter Plan",
					Description:              "A cost-effective plan for small teams starting their cooperative journey.",
					Cost:                     70, // AUD 70/month
					Timespan:                 int64(30 * 24 * time.Hour),
					MaxBranches:              3,
					MaxEmployees:             25,
					MaxMembersPerBranch:      25,
					Discount:                 5,
					YearlyDiscount:           15,
					IsRecommended:            false,
					HasAPIAccess:             true,
					HasFlexibleOrgStructures: false,
					HasAIEnabled:             false,
					HasMachineLearning:       false,
					MaxAPICallsPerMonth:      1000,
					CurrencyID:               &currency.ID,
				},
				{
					Name:                     "Free Plan",
					Description:              "A simple free plan with essential features for testing and evaluation.",
					Cost:                     0, // Free
					Timespan:                 int64(30 * 24 * time.Hour),
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
					CurrencyID:               &currency.ID,
				},
			}
		case "CAD": // Canada
			subscription = []*SubscriptionPlan{
				{
					Name:                     "Enterprise Plan",
					Description:              "A top-tier plan with unlimited access, advanced AI/ML tools, and premium support for large cooperatives.",
					Cost:                     540, // CAD 540/month (~USD 400 equivalent)
					Timespan:                 int64(30 * 24 * time.Hour),
					MaxBranches:              50,
					MaxEmployees:             1000,
					MaxMembersPerBranch:      500,
					Discount:                 15, // 15% discount
					YearlyDiscount:           25, // 25% yearly discount
					IsRecommended:            false,
					HasAPIAccess:             true,
					HasFlexibleOrgStructures: true,
					HasAIEnabled:             true,
					HasMachineLearning:       true,
					MaxAPICallsPerMonth:      0, // Unlimited
					CurrencyID:               &currency.ID,
				},
				{
					Name:                     "Pro Plan",
					Description:              "A professional plan for growing cooperatives with AI features and flexibility.",
					Cost:                     270, // CAD 270/month (~USD 200 equivalent)
					Timespan:                 int64(30 * 24 * time.Hour),
					MaxBranches:              20,
					MaxEmployees:             200,
					MaxMembersPerBranch:      100,
					Discount:                 10, // 10% discount
					YearlyDiscount:           20, // 20% yearly discount
					IsRecommended:            true,
					HasAPIAccess:             true,
					HasFlexibleOrgStructures: true,
					HasAIEnabled:             true,
					HasMachineLearning:       false,
					MaxAPICallsPerMonth:      0, // Unlimited
					CurrencyID:               &currency.ID,
				},
				{
					Name:                     "Growth Plan",
					Description:              "A flexible plan for mid-sized co-ops looking to scale efficiently.",
					Cost:                     135, // CAD 135/month (~USD 100 equivalent)
					Timespan:                 int64(30 * 24 * time.Hour),
					MaxBranches:              8,
					MaxEmployees:             75,
					MaxMembersPerBranch:      50,
					Discount:                 8,  // 8% discount
					YearlyDiscount:           18, // 18% yearly discount
					IsRecommended:            false,
					HasAPIAccess:             true,
					HasFlexibleOrgStructures: true,
					HasAIEnabled:             false,
					HasMachineLearning:       false,
					MaxAPICallsPerMonth:      10000,
					CurrencyID:               &currency.ID,
				},
				{
					Name:                     "Starter Plan",
					Description:              "An affordable plan designed for small organizations getting started.",
					Cost:                     65, // CAD 65/month (~USD 50 equivalent)
					Timespan:                 int64(30 * 24 * time.Hour),
					MaxBranches:              3,
					MaxEmployees:             25,
					MaxMembersPerBranch:      25,
					Discount:                 5,
					YearlyDiscount:           15,
					IsRecommended:            false,
					HasAPIAccess:             true,
					HasFlexibleOrgStructures: false,
					HasAIEnabled:             false,
					HasMachineLearning:       false,
					MaxAPICallsPerMonth:      1000,
					CurrencyID:               &currency.ID,
				},
				{
					Name:                     "Free Plan",
					Description:              "A basic free plan offering essential tools to help you get started.",
					Cost:                     0, // Free
					Timespan:                 int64(30 * 24 * time.Hour),
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
					CurrencyID:               &currency.ID,
				},
			}
		case "CHF": // Switzerland (high-income economy)
			subscription = []*SubscriptionPlan{
				{
					Name:                     "Enterprise Plan",
					Description:              "An enterprise-level plan with full AI and machine learning capabilities.",
					Cost:                     399, // CHF
					Timespan:                 int64(30 * 24 * time.Hour),
					MaxBranches:              50,
					MaxEmployees:             1000,
					MaxMembersPerBranch:      500,
					Discount:                 15,
					YearlyDiscount:           25,
					IsRecommended:            false,
					HasAPIAccess:             true,
					HasFlexibleOrgStructures: true,
					HasAIEnabled:             true,
					HasMachineLearning:       true,
					MaxAPICallsPerMonth:      0,
					CurrencyID:               &currency.ID,
				},
				{
					Name:                     "Pro Plan",
					Description:              "A professional plan for scaling organizations with AI features.",
					Cost:                     199,
					Timespan:                 int64(30 * 24 * time.Hour),
					MaxBranches:              20,
					MaxEmployees:             200,
					MaxMembersPerBranch:      100,
					Discount:                 10,
					YearlyDiscount:           20,
					IsRecommended:            true,
					HasAPIAccess:             true,
					HasFlexibleOrgStructures: true,
					HasAIEnabled:             true,
					CurrencyID:               &currency.ID,
				},
				{
					Name:                     "Growth Plan",
					Description:              "A balanced plan for growing organizations.",
					Cost:                     99,
					Timespan:                 int64(30 * 24 * time.Hour),
					MaxBranches:              8,
					MaxEmployees:             75,
					MaxMembersPerBranch:      50,
					Discount:                 7,
					YearlyDiscount:           17,
					HasAPIAccess:             true,
					HasFlexibleOrgStructures: true,
					CurrencyID:               &currency.ID,
				},
				{
					Name:                "Starter Plan",
					Description:         "An affordable plan for small cooperatives.",
					Cost:                49,
					Timespan:            int64(30 * 24 * time.Hour),
					MaxBranches:         3,
					MaxEmployees:        25,
					MaxMembersPerBranch: 25,
					Discount:            5,
					YearlyDiscount:      15,
					HasAPIAccess:        true,
					CurrencyID:          &currency.ID,
				},
				{
					Name:                "Free Plan",
					Description:         "Try our platform free for 30 days.",
					Cost:                0,
					Timespan:            int64(30 * 24 * time.Hour),
					MaxBranches:         1,
					MaxEmployees:        3,
					MaxMembersPerBranch: 10,
					CurrencyID:          &currency.ID,
				},
			}
		case "CNY": // China (use RMB, simplified Chinese)
			subscription = []*SubscriptionPlan{
				{
					Name:        "企业版",
					Description: "适用于大型企业，包含全部 AI 和机器学习功能。",
					Cost:        2800, // RMB
					Timespan:    int64(30 * 24 * time.Hour),
					MaxBranches: 50, MaxEmployees: 1000, MaxMembersPerBranch: 500,
					Discount: 15, YearlyDiscount: 25, HasAPIAccess: true, HasFlexibleOrgStructures: true, HasAIEnabled: true, HasMachineLearning: true,
					CurrencyID: &currency.ID,
				},
				{
					Name:        "专业版",
					Description: "适合成长型合作社，包含AI功能。",
					Cost:        1400,
					Timespan:    int64(30 * 24 * time.Hour),
					MaxBranches: 20, MaxEmployees: 200, MaxMembersPerBranch: 100,
					Discount: 10, YearlyDiscount: 20, HasAPIAccess: true, HasFlexibleOrgStructures: true, HasAIEnabled: true,
					IsRecommended: true,
					CurrencyID:    &currency.ID,
				},
				{
					Name:        "成长版",
					Description: "适用于中型合作社的灵活方案。",
					Cost:        700,
					Timespan:    int64(30 * 24 * time.Hour),
					MaxBranches: 8, MaxEmployees: 75, MaxMembersPerBranch: 50,
					Discount: 7, YearlyDiscount: 17, HasAPIAccess: true, HasFlexibleOrgStructures: true,
					CurrencyID: &currency.ID,
				},
				{
					Name:        "入门版",
					Description: "适合小型组织的经济型方案。",
					Cost:        300,
					Timespan:    int64(30 * 24 * time.Hour),
					MaxBranches: 3, MaxEmployees: 25, MaxMembersPerBranch: 25,
					Discount: 5, YearlyDiscount: 15, HasAPIAccess: true,
					CurrencyID: &currency.ID,
				},
				{
					Name:        "免费版",
					Description: "提供30天免费试用。",
					Cost:        0,
					Timespan:    int64(30 * 24 * time.Hour),
					MaxBranches: 1, MaxEmployees: 3, MaxMembersPerBranch: 10,
					CurrencyID: &currency.ID,
				},
			}
		case "SEK": // Sweden (English)
			subscription = []*SubscriptionPlan{
				{
					Name:        "Enterprise Plan",
					Description: "Complete AI and ML package for large cooperatives.",
					Cost:        3999, // SEK
					Timespan:    int64(30 * 24 * time.Hour),
					MaxBranches: 50, MaxEmployees: 1000, MaxMembersPerBranch: 500,
					Discount: 15, YearlyDiscount: 25, HasAPIAccess: true, HasFlexibleOrgStructures: true, HasAIEnabled: true, HasMachineLearning: true,
					CurrencyID: &currency.ID,
				},
				{
					Name:          "Pro Plan",
					Description:   "Perfect for expanding organizations with AI support.",
					Cost:          1999,
					Timespan:      int64(30 * 24 * time.Hour),
					IsRecommended: true,
					MaxBranches:   20, MaxEmployees: 200, MaxMembersPerBranch: 100,
					Discount: 10, YearlyDiscount: 20, HasAPIAccess: true, HasFlexibleOrgStructures: true, HasAIEnabled: true,
					CurrencyID: &currency.ID,
				},
				{
					Name:        "Growth Plan",
					Description: "For mid-sized cooperatives aiming to scale.",
					Cost:        999,
					Timespan:    int64(30 * 24 * time.Hour),
					MaxBranches: 8, MaxEmployees: 75, MaxMembersPerBranch: 50,
					Discount: 7, YearlyDiscount: 17, HasAPIAccess: true, HasFlexibleOrgStructures: true,
					CurrencyID: &currency.ID,
				},
				{
					Name:        "Starter Plan",
					Description: "Simple plan for small organizations.",
					Cost:        499,
					Timespan:    int64(30 * 24 * time.Hour),
					MaxBranches: 3, MaxEmployees: 25, MaxMembersPerBranch: 25,
					Discount: 5, YearlyDiscount: 15, HasAPIAccess: true,
					CurrencyID: &currency.ID,
				},
				{
					Name:        "Free Plan",
					Description: "Free 30-day access to essential tools.",
					Cost:        0,
					Timespan:    int64(30 * 24 * time.Hour),
					MaxBranches: 1, MaxEmployees: 3, MaxMembersPerBranch: 10,
					CurrencyID: &currency.ID,
				},
			}
		case "NZD": // New Zealand (English)
			subscription = []*SubscriptionPlan{
				{
					Name:        "Enterprise Plan",
					Description: "Unlimited features with AI/ML and enterprise tools.",
					Cost:        599, // NZD
					Timespan:    int64(30 * 24 * time.Hour),
					MaxBranches: 50, MaxEmployees: 1000, MaxMembersPerBranch: 500,
					Discount: 15, YearlyDiscount: 25, HasAPIAccess: true, HasFlexibleOrgStructures: true, HasAIEnabled: true, HasMachineLearning: true,
					CurrencyID: &currency.ID,
				},
				{
					Name:          "Pro Plan",
					Description:   "Best for professional growth with AI integration.",
					Cost:          299,
					IsRecommended: true,
					Timespan:      int64(30 * 24 * time.Hour),
					MaxBranches:   20, MaxEmployees: 200, MaxMembersPerBranch: 100,
					Discount: 10, YearlyDiscount: 20, HasAPIAccess: true, HasFlexibleOrgStructures: true, HasAIEnabled: true,
					CurrencyID: &currency.ID,
				},
				{
					Name:        "Growth Plan",
					Description: "Ideal for scaling co-ops and small businesses.",
					Cost:        149,
					Timespan:    int64(30 * 24 * time.Hour),
					MaxBranches: 8, MaxEmployees: 75, MaxMembersPerBranch: 50,
					Discount: 7, YearlyDiscount: 17, HasAPIAccess: true, HasFlexibleOrgStructures: true,
					CurrencyID: &currency.ID,
				},
				{
					Name:        "Starter Plan",
					Description: "Affordable plan for beginners.",
					Cost:        79,
					Timespan:    int64(30 * 24 * time.Hour),
					MaxBranches: 3, MaxEmployees: 25, MaxMembersPerBranch: 25,
					Discount: 5, YearlyDiscount: 15, HasAPIAccess: true,
					CurrencyID: &currency.ID,
				},
				{
					Name:        "Free Plan",
					Description: "Free access for 30 days to try our platform.",
					Cost:        0,
					Timespan:    int64(30 * 24 * time.Hour),
					MaxBranches: 1, MaxEmployees: 3, MaxMembersPerBranch: 10,
					CurrencyID: &currency.ID,
				},
			}
		case "PHP": // Philippines (English)
			subscription = []*SubscriptionPlan{
				{
					Name:        "Enterprise Plan",
					Description: "For large cooperatives with full AI and machine learning tools.",
					Cost:        24999, // PHP
					Timespan:    int64(30 * 24 * time.Hour),
					MaxBranches: 50, MaxEmployees: 1000, MaxMembersPerBranch: 500,
					Discount: 15, YearlyDiscount: 25, HasAPIAccess: true, HasFlexibleOrgStructures: true, HasAIEnabled: true, HasMachineLearning: true,
					CurrencyID: &currency.ID,
				},
				{
					Name:          "Pro Plan",
					Description:   "Perfect for growing cooperatives with AI tools.",
					Cost:          12499,
					IsRecommended: true,
					Timespan:      int64(30 * 24 * time.Hour),
					MaxBranches:   20, MaxEmployees: 200, MaxMembersPerBranch: 100,
					Discount: 10, YearlyDiscount: 20, HasAPIAccess: true, HasFlexibleOrgStructures: true, HasAIEnabled: true,
					CurrencyID: &currency.ID,
				},
				{
					Name:        "Growth Plan",
					Description: "Ideal for mid-sized organizations ready to scale.",
					Cost:        6499,
					Timespan:    int64(30 * 24 * time.Hour),
					MaxBranches: 8, MaxEmployees: 75, MaxMembersPerBranch: 50,
					Discount: 7, YearlyDiscount: 17, HasAPIAccess: true, HasFlexibleOrgStructures: true,
					CurrencyID: &currency.ID,
				},
				{
					Name:        "Starter Plan",
					Description: "A great choice for small and new cooperatives.",
					Cost:        2999,
					Timespan:    int64(30 * 24 * time.Hour),
					MaxBranches: 3, MaxEmployees: 25, MaxMembersPerBranch: 25,
					Discount: 5, YearlyDiscount: 15, HasAPIAccess: true,
					CurrencyID: &currency.ID,
				},
				{
					Name:        "Free Plan",
					Description: "Free access for 30 days to explore our platform.",
					Cost:        0,
					Timespan:    int64(30 * 24 * time.Hour),
					MaxBranches: 1, MaxEmployees: 3, MaxMembersPerBranch: 10,
					CurrencyID: &currency.ID,
				},
			}
		case "INR": // India (English)
			subscription = []*SubscriptionPlan{
				{
					Name:        "Enterprise Plan",
					Description: "Full-featured AI and ML suite for large organizations.",
					Cost:        24999, // INR
					Timespan:    int64(30 * 24 * time.Hour),
					MaxBranches: 50, MaxEmployees: 1000, MaxMembersPerBranch: 500,
					Discount: 15, YearlyDiscount: 25, HasAPIAccess: true, HasFlexibleOrgStructures: true, HasAIEnabled: true, HasMachineLearning: true,
					CurrencyID: &currency.ID,
				},
				{
					Name:          "Pro Plan",
					Description:   "AI-ready plan for growing cooperatives.",
					Cost:          12499,
					IsRecommended: true,
					Timespan:      int64(30 * 24 * time.Hour),
					MaxBranches:   20, MaxEmployees: 200, MaxMembersPerBranch: 100,
					Discount: 10, YearlyDiscount: 20, HasAPIAccess: true, HasFlexibleOrgStructures: true, HasAIEnabled: true,
					CurrencyID: &currency.ID,
				},
				{
					Name:        "Growth Plan",
					Description: "Mid-tier plan for scaling cooperatives.",
					Cost:        6999,
					Timespan:    int64(30 * 24 * time.Hour),
					MaxBranches: 8, MaxEmployees: 75, MaxMembersPerBranch: 50,
					Discount: 7, YearlyDiscount: 17, HasAPIAccess: true, HasFlexibleOrgStructures: true,
					CurrencyID: &currency.ID,
				},
				{
					Name:        "Starter Plan",
					Description: "Affordable plan for small organizations.",
					Cost:        2999,
					Timespan:    int64(30 * 24 * time.Hour),
					MaxBranches: 3, MaxEmployees: 25, MaxMembersPerBranch: 25,
					Discount: 5, YearlyDiscount: 15, HasAPIAccess: true,
					CurrencyID: &currency.ID,
				},
				{
					Name:        "Free Plan",
					Description: "Free 30-day trial with basic tools.",
					Cost:        0,
					Timespan:    int64(30 * 24 * time.Hour),
					MaxBranches: 1, MaxEmployees: 3, MaxMembersPerBranch: 10,
					CurrencyID: &currency.ID,
				},
			}
		case "KRW": // South Korea (₩) — local-friendly pricing
			subscription = []*SubscriptionPlan{
				{
					Name:        "Enterprise Plan",
					Description: "대기업을 위한 AI/ML 기능과 무제한 지원을 제공합니다.",
					Cost:        499000, // ₩499,000
					Timespan:    int64(30 * 24 * time.Hour),
					MaxBranches: 50, MaxEmployees: 1000, MaxMembersPerBranch: 500,
					Discount: 15, YearlyDiscount: 25,
					HasAPIAccess: true, HasFlexibleOrgStructures: true, HasAIEnabled: true, HasMachineLearning: true,
					CurrencyID: &currency.ID,
				},
				{
					Name:          "Pro Plan",
					Description:   "AI 기능이 포함된 성장형 협동조합을 위한 전문 플랜입니다.",
					Cost:          249000, // ₩249,000
					IsRecommended: true,
					Timespan:      int64(30 * 24 * time.Hour),
					MaxBranches:   20, MaxEmployees: 200, MaxMembersPerBranch: 100,
					Discount: 10, YearlyDiscount: 20,
					HasAPIAccess: true, HasFlexibleOrgStructures: true, HasAIEnabled: true,
					CurrencyID: &currency.ID,
				},
				{
					Name:        "Growth Plan",
					Description: "중형 조직을 위한 유연한 확장형 플랜입니다.",
					Cost:        129000, // ₩129,000
					Timespan:    int64(30 * 24 * time.Hour),
					MaxBranches: 8, MaxEmployees: 75, MaxMembersPerBranch: 50,
					Discount: 7, YearlyDiscount: 17,
					HasAPIAccess: true, HasFlexibleOrgStructures: true,
					CurrencyID: &currency.ID,
				},
				{
					Name:        "Starter Plan",
					Description: "소규모 협동조합을 위한 경제적인 플랜입니다.",
					Cost:        59000, // ₩59,000
					Timespan:    int64(30 * 24 * time.Hour),
					MaxBranches: 3, MaxEmployees: 25, MaxMembersPerBranch: 25,
					Discount: 5, YearlyDiscount: 15,
					HasAPIAccess: true,
					CurrencyID:   &currency.ID,
				},
				{
					Name:        "Free Plan",
					Description: "기본 기능으로 30일 무료 체험이 가능합니다.",
					Cost:        0,
					Timespan:    int64(30 * 24 * time.Hour),
					MaxBranches: 1, MaxEmployees: 3, MaxMembersPerBranch: 10,
					CurrencyID: &currency.ID,
				},
			}
		case "THB": // Thailand (฿) — local-friendly pricing, English
			subscription = []*SubscriptionPlan{
				{
					Name:        "Enterprise Plan",
					Description: "Complete enterprise solution with AI/ML capabilities.",
					Cost:        13999, // ฿13,999
					Timespan:    int64(30 * 24 * time.Hour),
					MaxBranches: 50, MaxEmployees: 1000, MaxMembersPerBranch: 500,
					Discount: 15, YearlyDiscount: 25,
					HasAPIAccess: true, HasFlexibleOrgStructures: true, HasAIEnabled: true, HasMachineLearning: true,
					CurrencyID: &currency.ID,
				},
				{
					Name:          "Pro Plan",
					Description:   "Ideal for growing cooperatives with AI features.",
					Cost:          6999, // ฿6,999
					IsRecommended: true,
					Timespan:      int64(30 * 24 * time.Hour),
					MaxBranches:   20, MaxEmployees: 200, MaxMembersPerBranch: 100,
					Discount: 10, YearlyDiscount: 20,
					HasAPIAccess: true, HasFlexibleOrgStructures: true, HasAIEnabled: true,
					CurrencyID: &currency.ID,
				},
				{
					Name:        "Growth Plan",
					Description: "Balanced plan for mid-sized co-ops.",
					Cost:        3499, // ฿3,499
					Timespan:    int64(30 * 24 * time.Hour),
					MaxBranches: 8, MaxEmployees: 75, MaxMembersPerBranch: 50,
					Discount: 7, YearlyDiscount: 17,
					HasAPIAccess: true, HasFlexibleOrgStructures: true,
					CurrencyID: &currency.ID,
				},
				{
					Name:        "Starter Plan",
					Description: "Affordable plan for small organizations.",
					Cost:        1499, // ฿1,499
					Timespan:    int64(30 * 24 * time.Hour),
					MaxBranches: 3, MaxEmployees: 25, MaxMembersPerBranch: 25,
					Discount: 5, YearlyDiscount: 15,
					HasAPIAccess: true,
					CurrencyID:   &currency.ID,
				},
				{
					Name:        "Free Plan",
					Description: "30-day free trial with basic tools.",
					Cost:        0,
					Timespan:    int64(30 * 24 * time.Hour),
					MaxBranches: 1, MaxEmployees: 3, MaxMembersPerBranch: 10,
					CurrencyID: &currency.ID,
				},
			}
		case "SGD": // Singapore (English)
			subscription = []*SubscriptionPlan{
				{
					Name:        "Enterprise Plan",
					Description: "Comprehensive AI/ML plan for enterprise-level co-ops.",
					Cost:        499, // SGD
					Timespan:    int64(30 * 24 * time.Hour),
					MaxBranches: 50, MaxEmployees: 1000, MaxMembersPerBranch: 500,
					Discount: 15, YearlyDiscount: 25,
					HasAPIAccess: true, HasFlexibleOrgStructures: true, HasAIEnabled: true, HasMachineLearning: true,
					CurrencyID: &currency.ID,
				},
				{
					Name:          "Pro Plan",
					Description:   "AI-powered plan for growing cooperatives.",
					Cost:          249,
					IsRecommended: true,
					Timespan:      int64(30 * 24 * time.Hour),
					MaxBranches:   20, MaxEmployees: 200, MaxMembersPerBranch: 100,
					Discount: 10, YearlyDiscount: 20,
					HasAPIAccess: true, HasFlexibleOrgStructures: true, HasAIEnabled: true,
					CurrencyID: &currency.ID,
				},
				{
					Name:        "Growth Plan",
					Description: "Perfect for medium-sized co-ops ready to scale.",
					Cost:        129,
					Timespan:    int64(30 * 24 * time.Hour),
					MaxBranches: 8, MaxEmployees: 75, MaxMembersPerBranch: 50,
					Discount: 7, YearlyDiscount: 17,
					HasAPIAccess: true, HasFlexibleOrgStructures: true,
					CurrencyID: &currency.ID,
				},
				{
					Name:        "Starter Plan",
					Description: "Affordable option for small co-ops.",
					Cost:        59,
					Timespan:    int64(30 * 24 * time.Hour),
					MaxBranches: 3, MaxEmployees: 25, MaxMembersPerBranch: 25,
					Discount: 5, YearlyDiscount: 15,
					HasAPIAccess: true,
					CurrencyID:   &currency.ID,
				},
				{
					Name:        "Free Plan",
					Description: "Free 30-day access with basic tools.",
					Cost:        0,
					Timespan:    int64(30 * 24 * time.Hour),
					MaxBranches: 1, MaxEmployees: 3, MaxMembersPerBranch: 10,
					CurrencyID: &currency.ID,
				},
			}
		case "HKD": // Hong Kong (English)
			subscription = []*SubscriptionPlan{
				{
					Name:        "Enterprise Plan",
					Description: "Full-featured enterprise plan with AI/ML tools.",
					Cost:        2999, // HKD
					Timespan:    int64(30 * 24 * time.Hour),
					MaxBranches: 50, MaxEmployees: 1000, MaxMembersPerBranch: 500,
					Discount: 15, YearlyDiscount: 25,
					HasAPIAccess: true, HasFlexibleOrgStructures: true, HasAIEnabled: true, HasMachineLearning: true,
					CurrencyID: &currency.ID,
				},
				{
					Name:          "Pro Plan",
					Description:   "Advanced plan for growing co-ops with AI tools.",
					Cost:          1499,
					IsRecommended: true,
					Timespan:      int64(30 * 24 * time.Hour),
					MaxBranches:   20, MaxEmployees: 200, MaxMembersPerBranch: 100,
					Discount: 10, YearlyDiscount: 20,
					HasAPIAccess: true, HasFlexibleOrgStructures: true, HasAIEnabled: true,
					CurrencyID: &currency.ID,
				},
				{
					Name:        "Growth Plan",
					Description: "Balanced plan for scaling organizations.",
					Cost:        799,
					Timespan:    int64(30 * 24 * time.Hour),
					MaxBranches: 8, MaxEmployees: 75, MaxMembersPerBranch: 50,
					Discount: 7, YearlyDiscount: 17,
					HasAPIAccess: true, HasFlexibleOrgStructures: true,
					CurrencyID: &currency.ID,
				},
				{
					Name:        "Starter Plan",
					Description: "Great choice for small co-ops.",
					Cost:        399,
					Timespan:    int64(30 * 24 * time.Hour),
					MaxBranches: 3, MaxEmployees: 25, MaxMembersPerBranch: 25,
					Discount: 5, YearlyDiscount: 15,
					HasAPIAccess: true,
					CurrencyID:   &currency.ID,
				},
				{
					Name:        "Free Plan",
					Description: "30-day free trial to explore our platform.",
					Cost:        0,
					Timespan:    int64(30 * 24 * time.Hour),
					MaxBranches: 1, MaxEmployees: 3, MaxMembersPerBranch: 10,
					CurrencyID: &currency.ID,
				},
			}
		case "MYR": // Malaysia (English)
			subscription = []*SubscriptionPlan{
				{
					Name:        "Enterprise Plan",
					Description: "Enterprise-level plan with AI and ML features.",
					Cost:        1499, // RM1,499
					Timespan:    int64(30 * 24 * time.Hour),
					MaxBranches: 50, MaxEmployees: 1000, MaxMembersPerBranch: 500,
					Discount: 15, YearlyDiscount: 25,
					HasAPIAccess: true, HasFlexibleOrgStructures: true, HasAIEnabled: true, HasMachineLearning: true,
					CurrencyID: &currency.ID,
				},
				{
					Name:          "Pro Plan",
					Description:   "For growing organizations with AI features.",
					Cost:          799, // RM799
					IsRecommended: true,
					Timespan:      int64(30 * 24 * time.Hour),
					MaxBranches:   20, MaxEmployees: 200, MaxMembersPerBranch: 100,
					Discount: 10, YearlyDiscount: 20,
					HasAPIAccess: true, HasFlexibleOrgStructures: true, HasAIEnabled: true,
					CurrencyID: &currency.ID,
				},
				{
					Name:        "Growth Plan",
					Description: "Balanced plan for mid-sized co-ops.",
					Cost:        399, // RM399
					Timespan:    int64(30 * 24 * time.Hour),
					MaxBranches: 8, MaxEmployees: 75, MaxMembersPerBranch: 50,
					Discount: 7, YearlyDiscount: 17,
					HasAPIAccess: true, HasFlexibleOrgStructures: true,
					CurrencyID: &currency.ID,
				},
				{
					Name:        "Starter Plan",
					Description: "Affordable plan for small organizations.",
					Cost:        199, // RM199
					Timespan:    int64(30 * 24 * time.Hour),
					MaxBranches: 3, MaxEmployees: 25, MaxMembersPerBranch: 25,
					Discount: 5, YearlyDiscount: 15,
					HasAPIAccess: true,
					CurrencyID:   &currency.ID,
				},
				{
					Name:        "Free Plan",
					Description: "Free access for 30 days with essential tools.",
					Cost:        0,
					Timespan:    int64(30 * 24 * time.Hour),
					MaxBranches: 1, MaxEmployees: 3, MaxMembersPerBranch: 10,
					CurrencyID: &currency.ID,
				},
			}
		case "IDR": // Indonesia (Rp) — Bahasa Indonesia
			subscription = []*SubscriptionPlan{
				{
					Name:        "Paket Enterprise",
					Description: "Solusi lengkap untuk koperasi besar dengan fitur AI/ML.",
					Cost:        3990000, // Rp3,990,000
					Timespan:    int64(30 * 24 * time.Hour),
					MaxBranches: 50, MaxEmployees: 1000, MaxMembersPerBranch: 500,
					Discount: 15, YearlyDiscount: 25,
					HasAPIAccess: true, HasFlexibleOrgStructures: true, HasAIEnabled: true, HasMachineLearning: true,
					CurrencyID: &currency.ID,
				},
				{
					Name:          "Paket Pro",
					Description:   "Untuk koperasi yang sedang berkembang dengan fitur AI.",
					Cost:          1990000,
					IsRecommended: true,
					Timespan:      int64(30 * 24 * time.Hour),
					MaxBranches:   20, MaxEmployees: 200, MaxMembersPerBranch: 100,
					Discount: 10, YearlyDiscount: 20,
					HasAPIAccess: true, HasFlexibleOrgStructures: true, HasAIEnabled: true,
					CurrencyID: &currency.ID,
				},
				{
					Name:        "Paket Growth",
					Description: "Cocok untuk koperasi menengah yang ingin berkembang.",
					Cost:        999000,
					Timespan:    int64(30 * 24 * time.Hour),
					MaxBranches: 8, MaxEmployees: 75, MaxMembersPerBranch: 50,
					Discount: 7, YearlyDiscount: 17,
					HasAPIAccess: true, HasFlexibleOrgStructures: true,
					CurrencyID: &currency.ID,
				},
				{
					Name:        "Paket Starter",
					Description: "Pilihan hemat untuk koperasi kecil.",
					Cost:        499000,
					Timespan:    int64(30 * 24 * time.Hour),
					MaxBranches: 3, MaxEmployees: 25, MaxMembersPerBranch: 25,
					Discount: 5, YearlyDiscount: 15,
					HasAPIAccess: true,
					CurrencyID:   &currency.ID,
				},
				{
					Name:        "Paket Gratis",
					Description: "Uji coba gratis selama 30 hari dengan fitur dasar.",
					Cost:        0,
					Timespan:    int64(30 * 24 * time.Hour),
					MaxBranches: 1, MaxEmployees: 3, MaxMembersPerBranch: 10,
					CurrencyID: &currency.ID,
				},
			}
		case "VND": // Vietnam (₫) — Vietnamese
			subscription = []*SubscriptionPlan{
				{
					Name:        "Gói Doanh Nghiệp",
					Description: "Gói cao cấp với đầy đủ tính năng AI và học máy.",
					Cost:        4990000, // ₫4,990,000
					Timespan:    int64(30 * 24 * time.Hour),
					MaxBranches: 50, MaxEmployees: 1000, MaxMembersPerBranch: 500,
					Discount: 15, YearlyDiscount: 25,
					HasAPIAccess: true, HasFlexibleOrgStructures: true, HasAIEnabled: true, HasMachineLearning: true,
					CurrencyID: &currency.ID,
				},
				{
					Name:          "Gói Chuyên Nghiệp",
					Description:   "Phù hợp với hợp tác xã đang phát triển cùng AI.",
					Cost:          2490000, // ₫2,490,000
					IsRecommended: true,
					Timespan:      int64(30 * 24 * time.Hour),
					MaxBranches:   20, MaxEmployees: 200, MaxMembersPerBranch: 100,
					Discount: 10, YearlyDiscount: 20,
					HasAPIAccess: true, HasFlexibleOrgStructures: true, HasAIEnabled: true,
					CurrencyID: &currency.ID,
				},
				{
					Name:        "Gói Tăng Trưởng",
					Description: "Dành cho hợp tác xã quy mô vừa muốn mở rộng.",
					Cost:        1290000,
					Timespan:    int64(30 * 24 * time.Hour),
					MaxBranches: 8, MaxEmployees: 75, MaxMembersPerBranch: 50,
					Discount: 7, YearlyDiscount: 17,
					HasAPIAccess: true, HasFlexibleOrgStructures: true,
					CurrencyID: &currency.ID,
				},
				{
					Name:        "Gói Khởi Đầu",
					Description: "Gói tiết kiệm cho tổ chức nhỏ.",
					Cost:        599000,
					Timespan:    int64(30 * 24 * time.Hour),
					MaxBranches: 3, MaxEmployees: 25, MaxMembersPerBranch: 25,
					Discount: 5, YearlyDiscount: 15,
					HasAPIAccess: true,
					CurrencyID:   &currency.ID,
				},
				{
					Name:        "Gói Miễn Phí",
					Description: "Dùng thử 30 ngày miễn phí với tính năng cơ bản.",
					Cost:        0,
					Timespan:    int64(30 * 24 * time.Hour),
					MaxBranches: 1, MaxEmployees: 3, MaxMembersPerBranch: 10,
					CurrencyID: &currency.ID,
				},
			}
		case "TWD": // Taiwan (NT$) — English
			subscription = []*SubscriptionPlan{
				{
					Name:        "Enterprise Plan",
					Description: "Comprehensive enterprise solution with AI and ML capabilities.",
					Cost:        11999, // NT$11,999
					Timespan:    int64(30 * 24 * time.Hour),
					MaxBranches: 50, MaxEmployees: 1000, MaxMembersPerBranch: 500,
					Discount: 15, YearlyDiscount: 25,
					IsRecommended: false,
					HasAPIAccess:  true, HasFlexibleOrgStructures: true,
					HasAIEnabled: true, HasMachineLearning: true,
					MaxAPICallsPerMonth: 0, CurrencyID: &currency.ID,
				},
				{
					Name:          "Pro Plan",
					Description:   "For growing cooperatives with AI tools and premium support.",
					Cost:          5999, // NT$5,999
					IsRecommended: true,
					Timespan:      int64(30 * 24 * time.Hour),
					MaxBranches:   20, MaxEmployees: 200, MaxMembersPerBranch: 100,
					Discount: 10, YearlyDiscount: 20,
					HasAPIAccess: true, HasFlexibleOrgStructures: true, HasAIEnabled: true,
					MaxAPICallsPerMonth: 0, CurrencyID: &currency.ID,
				},
				{
					Name:        "Growth Plan",
					Description: "Mid-tier plan for scaling co-ops with flexibility.",
					Cost:        2999, // NT$2,999
					Timespan:    int64(30 * 24 * time.Hour),
					MaxBranches: 8, MaxEmployees: 75, MaxMembersPerBranch: 50,
					Discount: 7, YearlyDiscount: 17,
					HasAPIAccess: true, HasFlexibleOrgStructures: true,
					MaxAPICallsPerMonth: 10000, CurrencyID: &currency.ID,
				},
				{
					Name:        "Starter Plan",
					Description: "Affordable entry plan for small organizations.",
					Cost:        1299, // NT$1,299
					Timespan:    int64(30 * 24 * time.Hour),
					MaxBranches: 3, MaxEmployees: 25, MaxMembersPerBranch: 25,
					Discount: 5, YearlyDiscount: 15,
					HasAPIAccess: true, HasFlexibleOrgStructures: false,
					MaxAPICallsPerMonth: 1000, CurrencyID: &currency.ID,
				},
				{
					Name:        "Free Plan",
					Description: "Try the essential tools free for 30 days.",
					Cost:        0,
					Timespan:    int64(30 * 24 * time.Hour),
					MaxBranches: 1, MaxEmployees: 3, MaxMembersPerBranch: 10,
					MaxAPICallsPerMonth: 100, CurrencyID: &currency.ID,
				},
			}
		case "BND": // Brunei (B$) — English
			subscription = []*SubscriptionPlan{
				{
					Name:        "Enterprise Plan",
					Description: "Enterprise-grade plan with full AI and ML capabilities.",
					Cost:        499, // B$499
					Timespan:    int64(30 * 24 * time.Hour),
					MaxBranches: 50, MaxEmployees: 1000, MaxMembersPerBranch: 500,
					Discount: 15, YearlyDiscount: 25,
					HasAPIAccess: true, HasFlexibleOrgStructures: true,
					HasAIEnabled: true, HasMachineLearning: true,
					CurrencyID: &currency.ID,
				},
				{
					Name:          "Pro Plan",
					Description:   "For growing cooperatives with AI integration.",
					Cost:          249, // B$249
					IsRecommended: true,
					Timespan:      int64(30 * 24 * time.Hour),
					MaxBranches:   20, MaxEmployees: 200, MaxMembersPerBranch: 100,
					Discount: 10, YearlyDiscount: 20,
					HasAPIAccess: true, HasFlexibleOrgStructures: true, HasAIEnabled: true,
					CurrencyID: &currency.ID,
				},
				{
					Name:        "Growth Plan",
					Description: "Flexible plan for expanding co-ops.",
					Cost:        129, // B$129
					Timespan:    int64(30 * 24 * time.Hour),
					MaxBranches: 8, MaxEmployees: 75, MaxMembersPerBranch: 50,
					Discount: 7, YearlyDiscount: 17,
					HasAPIAccess: true, HasFlexibleOrgStructures: true,
					CurrencyID: &currency.ID,
				},
				{
					Name:        "Starter Plan",
					Description: "Affordable plan for small organizations.",
					Cost:        59, // B$59
					Timespan:    int64(30 * 24 * time.Hour),
					MaxBranches: 3, MaxEmployees: 25, MaxMembersPerBranch: 25,
					Discount: 5, YearlyDiscount: 15,
					HasAPIAccess: true,
					CurrencyID:   &currency.ID,
				},
				{
					Name:        "Free Plan",
					Description: "Free 30-day trial with core features.",
					Cost:        0,
					Timespan:    int64(30 * 24 * time.Hour),
					MaxBranches: 1, MaxEmployees: 3, MaxMembersPerBranch: 10,
					CurrencyID: &currency.ID,
				},
			}
		case "SAR": // Saudi Arabia (﷼) — English
			subscription = []*SubscriptionPlan{
				{
					Name:        "Enterprise Plan",
					Description: "Full-featured enterprise solution with AI/ML.",
					Cost:        1499, // SAR 1,499
					Timespan:    int64(30 * 24 * time.Hour),
					MaxBranches: 50, MaxEmployees: 1000, MaxMembersPerBranch: 500,
					Discount: 15, YearlyDiscount: 25,
					HasAPIAccess: true, HasFlexibleOrgStructures: true,
					HasAIEnabled: true, HasMachineLearning: true,
					CurrencyID: &currency.ID,
				},
				{
					Name:          "Pro Plan",
					Description:   "AI-ready plan for growing organizations.",
					Cost:          799,
					IsRecommended: true,
					Timespan:      int64(30 * 24 * time.Hour),
					MaxBranches:   20, MaxEmployees: 200, MaxMembersPerBranch: 100,
					Discount: 10, YearlyDiscount: 20,
					HasAPIAccess: true, HasFlexibleOrgStructures: true, HasAIEnabled: true,
					CurrencyID: &currency.ID,
				},
				{
					Name:        "Growth Plan",
					Description: "Ideal for mid-sized co-ops looking to expand.",
					Cost:        399,
					Timespan:    int64(30 * 24 * time.Hour),
					MaxBranches: 8, MaxEmployees: 75, MaxMembersPerBranch: 50,
					Discount: 7, YearlyDiscount: 17,
					HasAPIAccess: true, HasFlexibleOrgStructures: true,
					CurrencyID: &currency.ID,
				},
				{
					Name:        "Starter Plan",
					Description: "Affordable plan for small organizations.",
					Cost:        199,
					Timespan:    int64(30 * 24 * time.Hour),
					MaxBranches: 3, MaxEmployees: 25, MaxMembersPerBranch: 25,
					Discount: 5, YearlyDiscount: 15,
					HasAPIAccess: true,
					CurrencyID:   &currency.ID,
				},
				{
					Name:        "Free Plan",
					Description: "Free trial for 30 days with basic tools.",
					Cost:        0,
					Timespan:    int64(30 * 24 * time.Hour),
					MaxBranches: 1, MaxEmployees: 3, MaxMembersPerBranch: 10,
					CurrencyID: &currency.ID,
				},
			}
		case "AED": // United Arab Emirates (د.إ) — English
			subscription = []*SubscriptionPlan{
				{
					Name:        "Enterprise Plan",
					Description: "Advanced enterprise-level solution with AI/ML.",
					Cost:        1499, // AED 1,499
					Timespan:    int64(30 * 24 * time.Hour),
					MaxBranches: 50, MaxEmployees: 1000, MaxMembersPerBranch: 500,
					Discount: 15, YearlyDiscount: 25,
					HasAPIAccess: true, HasFlexibleOrgStructures: true,
					HasAIEnabled: true, HasMachineLearning: true,
					CurrencyID: &currency.ID,
				},
				{
					Name:          "Pro Plan",
					Description:   "AI-powered plan for growing cooperatives.",
					Cost:          749,
					IsRecommended: true,
					Timespan:      int64(30 * 24 * time.Hour),
					MaxBranches:   20, MaxEmployees: 200, MaxMembersPerBranch: 100,
					Discount: 10, YearlyDiscount: 20,
					HasAPIAccess: true, HasFlexibleOrgStructures: true, HasAIEnabled: true,
					CurrencyID: &currency.ID,
				},
				{
					Name:        "Growth Plan",
					Description: "Balanced plan for scaling co-ops.",
					Cost:        349,
					Timespan:    int64(30 * 24 * time.Hour),
					MaxBranches: 8, MaxEmployees: 75, MaxMembersPerBranch: 50,
					Discount: 7, YearlyDiscount: 17,
					HasAPIAccess: true, HasFlexibleOrgStructures: true,
					CurrencyID: &currency.ID,
				},
				{
					Name:        "Starter Plan",
					Description: "Starter option for small organizations.",
					Cost:        179,
					Timespan:    int64(30 * 24 * time.Hour),
					MaxBranches: 3, MaxEmployees: 25, MaxMembersPerBranch: 25,
					Discount: 5, YearlyDiscount: 15,
					HasAPIAccess: true,
					CurrencyID:   &currency.ID,
				},
				{
					Name:        "Free Plan",
					Description: "Free trial for 30 days with core tools.",
					Cost:        0,
					Timespan:    int64(30 * 24 * time.Hour),
					MaxBranches: 1, MaxEmployees: 3, MaxMembersPerBranch: 10,
					CurrencyID: &currency.ID,
				},
			}
		case "ILS": // Israel (₪) — English
			subscription = []*SubscriptionPlan{
				{
					Name:        "Enterprise Plan",
					Description: "Comprehensive enterprise plan with AI/ML and unlimited features.",
					Cost:        1499, // ₪1,499
					Timespan:    int64(30 * 24 * time.Hour),
					MaxBranches: 50, MaxEmployees: 1000, MaxMembersPerBranch: 500,
					Discount: 15, YearlyDiscount: 25,
					HasAPIAccess: true, HasFlexibleOrgStructures: true,
					HasAIEnabled: true, HasMachineLearning: true,
					CurrencyID: &currency.ID,
				},
				{
					Name:          "Pro Plan",
					Description:   "Perfect for growing cooperatives with AI support.",
					Cost:          799, // ₪799
					IsRecommended: true,
					Timespan:      int64(30 * 24 * time.Hour),
					MaxBranches:   20, MaxEmployees: 200, MaxMembersPerBranch: 100,
					Discount: 10, YearlyDiscount: 20,
					HasAPIAccess: true, HasFlexibleOrgStructures: true, HasAIEnabled: true,
					CurrencyID: &currency.ID,
				},
				{
					Name:        "Growth Plan",
					Description: "Mid-tier flexible plan for scaling cooperatives.",
					Cost:        399, // ₪399
					Timespan:    int64(30 * 24 * time.Hour),
					MaxBranches: 8, MaxEmployees: 75, MaxMembersPerBranch: 50,
					Discount: 7, YearlyDiscount: 17,
					HasAPIAccess: true, HasFlexibleOrgStructures: true,
					CurrencyID: &currency.ID,
				},
				{
					Name:        "Starter Plan",
					Description: "Entry-level plan for small organizations.",
					Cost:        179, // ₪179
					Timespan:    int64(30 * 24 * time.Hour),
					MaxBranches: 3, MaxEmployees: 25, MaxMembersPerBranch: 25,
					Discount: 5, YearlyDiscount: 15,
					HasAPIAccess: true,
					CurrencyID:   &currency.ID,
				},
				{
					Name:        "Free Plan",
					Description: "Free 30-day access with basic features.",
					Cost:        0,
					Timespan:    int64(30 * 24 * time.Hour),
					MaxBranches: 1, MaxEmployees: 3, MaxMembersPerBranch: 10,
					CurrencyID: &currency.ID,
				},
			}
		case "ZAR": // South Africa
			subscription = []*SubscriptionPlan{
				{
					Name:        "Enterprise Plan",
					Description: "Top-tier plan with AI, machine learning, and unlimited features.",
					Cost:        7000,
					Timespan:    int64(30 * 24 * time.Hour),
					MaxBranches: 50, MaxEmployees: 1000, MaxMembersPerBranch: 500,
					Discount: 15, YearlyDiscount: 25,
					HasAPIAccess: true, HasFlexibleOrgStructures: true,
					HasAIEnabled: true, HasMachineLearning: true,
					MaxAPICallsPerMonth: 0, // Unlimited
					CurrencyID:          &currency.ID,
				},
				{
					Name:        "Pro Plan",
					Description: "Perfect for growing cooperatives with AI tools.",
					Cost:        3500,
					Timespan:    int64(30 * 24 * time.Hour),
					MaxBranches: 20, MaxEmployees: 200, MaxMembersPerBranch: 100,
					Discount: 10, YearlyDiscount: 20,
					IsRecommended: true,
					HasAPIAccess:  true, HasFlexibleOrgStructures: true,
					HasAIEnabled: true, CurrencyID: &currency.ID,
				},
				{
					Name:        "Growth Plan",
					Description: "For mid-sized co-ops ready to expand operations.",
					Cost:        1800,
					Timespan:    int64(30 * 24 * time.Hour),
					MaxBranches: 8, MaxEmployees: 75, MaxMembersPerBranch: 50,
					Discount: 7, YearlyDiscount: 17,
					HasAPIAccess: true, HasFlexibleOrgStructures: true,
					CurrencyID: &currency.ID,
				},
				{
					Name:        "Starter Plan",
					Description: "Affordable option for new organizations.",
					Cost:        900,
					Timespan:    int64(30 * 24 * time.Hour),
					MaxBranches: 3, MaxEmployees: 25, MaxMembersPerBranch: 25,
					Discount: 5, YearlyDiscount: 15,
					HasAPIAccess: true, CurrencyID: &currency.ID,
				},
				{
					Name:        "Free Plan",
					Description: "Free 30-day trial with basic features.",
					Cost:        0,
					Timespan:    int64(30 * 24 * time.Hour),
					MaxBranches: 1, MaxEmployees: 3, MaxMembersPerBranch: 10,
					CurrencyID: &currency.ID,
				},
			}
		case "EGP": // Egypt
			subscription = []*SubscriptionPlan{
				{
					Name:        "خطة المؤسسات",
					Description: "الخطة الشاملة للمؤسسات الكبيرة مع ميزات الذكاء الاصطناعي والدعم المميز.",
					Cost:        12000,
					Timespan:    int64(30 * 24 * time.Hour),
					MaxBranches: 50, MaxEmployees: 1000, MaxMembersPerBranch: 500,
					Discount: 15, YearlyDiscount: 25,
					HasAPIAccess: true, HasFlexibleOrgStructures: true,
					HasAIEnabled: true, HasMachineLearning: true,
					MaxAPICallsPerMonth: 0, CurrencyID: &currency.ID,
				},
				{
					Name:        "الخطة الاحترافية",
					Description: "الخطة المثالية للمؤسسات النامية مع أدوات الذكاء الاصطناعي.",
					Cost:        6000,
					Timespan:    int64(30 * 24 * time.Hour),
					MaxBranches: 20, MaxEmployees: 200, MaxMembersPerBranch: 100,
					IsRecommended: true,
					Discount:      10, YearlyDiscount: 20,
					HasAPIAccess: true, HasAIEnabled: true,
					CurrencyID: &currency.ID,
				},
				{
					Name:        "خطة النمو",
					Description: "للمؤسسات المتوسطة التي تستعد للتوسع.",
					Cost:        3000,
					Timespan:    int64(30 * 24 * time.Hour),
					MaxBranches: 8, MaxEmployees: 75, MaxMembersPerBranch: 50,
					Discount: 7, YearlyDiscount: 17,
					HasAPIAccess: true, CurrencyID: &currency.ID,
				},
				{
					Name:        "الخطة المبتدئة",
					Description: "خطة مناسبة للمؤسسات الصغيرة.",
					Cost:        1500,
					Timespan:    int64(30 * 24 * time.Hour),
					MaxBranches: 3, MaxEmployees: 25, MaxMembersPerBranch: 25,
					CurrencyID: &currency.ID,
				},
				{
					Name:        "الخطة المجانية",
					Description: "تجربة مجانية لمدة 30 يومًا مع ميزات أساسية.",
					Cost:        0,
					Timespan:    int64(30 * 24 * time.Hour),
					MaxBranches: 1, MaxEmployees: 3, MaxMembersPerBranch: 10,
					CurrencyID: &currency.ID,
				},
			}
		case "TRY": // Turkey
			subscription = []*SubscriptionPlan{
				{
					Name:        "Kurumsal Plan",
					Description: "Sınırsız özellikler, AI/ML desteği ve öncelikli destek içeren üst düzey plan.",
					Cost:        9000,
					Timespan:    int64(30 * 24 * time.Hour),
					MaxBranches: 50, MaxEmployees: 1000, MaxMembersPerBranch: 500,
					Discount: 15, YearlyDiscount: 25,
					HasAPIAccess: true, HasAIEnabled: true, HasMachineLearning: true,
					CurrencyID: &currency.ID,
				},
				{
					Name:          "Profesyonel Plan",
					Description:   "Büyüyen kooperatifler için profesyonel plan.",
					Cost:          4500,
					Timespan:      int64(30 * 24 * time.Hour),
					IsRecommended: true,
					MaxBranches:   20, MaxEmployees: 200, MaxMembersPerBranch: 100,
					Discount: 10, YearlyDiscount: 20,
					HasAPIAccess: true, HasAIEnabled: true,
					CurrencyID: &currency.ID,
				},
				{
					Name:        "Büyüme Planı",
					Description: "Orta ölçekli kurumlar için dengeli bir plan.",
					Cost:        2200,
					Timespan:    int64(30 * 24 * time.Hour),
					MaxBranches: 8, MaxEmployees: 75, MaxMembersPerBranch: 50,
					CurrencyID: &currency.ID,
				},
				{
					Name:        "Başlangıç Planı",
					Description: "Yeni başlayanlar için uygun fiyatlı plan.",
					Cost:        1000,
					Timespan:    int64(30 * 24 * time.Hour),
					CurrencyID:  &currency.ID,
				},
				{
					Name:        "Ücretsiz Plan",
					Description: "Temel özelliklerle 30 günlük ücretsiz deneme.",
					Cost:        0,
					Timespan:    int64(30 * 24 * time.Hour),
					CurrencyID:  &currency.ID,
				},
			}
		case "XOF": // West African CFA Franc
			subscription = []*SubscriptionPlan{
				{
					Name:        "Plan Entreprise",
					Description: "Plan complet avec IA et assistance prioritaire.",
					Cost:        2400000,
					Timespan:    int64(30 * 24 * time.Hour),
					MaxBranches: 50, MaxEmployees: 1000,
					Discount: 15, YearlyDiscount: 25,
					HasAPIAccess: true, HasAIEnabled: true, HasMachineLearning: true,
					CurrencyID: &currency.ID,
				},
				{
					Name:          "Plan Pro",
					Description:   "Idéal pour les coopératives en croissance avec des outils d'IA.",
					Cost:          1200000,
					Timespan:      int64(30 * 24 * time.Hour),
					IsRecommended: true,
					HasAPIAccess:  true, HasAIEnabled: true,
					CurrencyID: &currency.ID,
				},
				{
					Name:        "Plan Croissance",
					Description: "Pour les structures prêtes à se développer.",
					Cost:        600000,
					Timespan:    int64(30 * 24 * time.Hour),
					CurrencyID:  &currency.ID,
				},
				{
					Name:        "Plan Débutant",
					Description: "Plan économique pour les petites structures.",
					Cost:        300000,
					Timespan:    int64(30 * 24 * time.Hour),
					CurrencyID:  &currency.ID,
				},
				{
					Name:        "Plan Gratuit",
					Description: "Essai gratuit de 30 jours avec fonctions de base.",
					Cost:        0,
					Timespan:    int64(30 * 24 * time.Hour),
					CurrencyID:  &currency.ID,
				},
			}
		case "XAF": // Central African CFA Franc
			subscription = []*SubscriptionPlan{
				{
					Name:        "Plan Entreprise",
					Description: "Plan complet avec IA, ML et support premium.",
					Cost:        2400000,
					Timespan:    int64(30 * 24 * time.Hour),
					MaxBranches: 50, MaxEmployees: 1000,
					Discount: 15, YearlyDiscount: 25,
					HasAPIAccess: true, HasAIEnabled: true, HasMachineLearning: true,
					CurrencyID: &currency.ID,
				},
				{
					Name:          "Plan Pro",
					Description:   "Plan professionnel pour coopératives en expansion.",
					Cost:          1200000,
					Timespan:      int64(30 * 24 * time.Hour),
					IsRecommended: true,
					HasAPIAccess:  true, HasAIEnabled: true,
					CurrencyID: &currency.ID,
				},
				{
					Name:        "Plan Croissance",
					Description: "Plan équilibré pour structures moyennes.",
					Cost:        600000,
					Timespan:    int64(30 * 24 * time.Hour),
					CurrencyID:  &currency.ID,
				},
				{
					Name:        "Plan Débutant",
					Description: "Plan abordable pour petites organisations.",
					Cost:        300000,
					Timespan:    int64(30 * 24 * time.Hour),
					CurrencyID:  &currency.ID,
				},
				{
					Name:        "Plan Gratuit",
					Description: "Essai gratuit de 30 jours avec fonctionnalités limitées.",
					Cost:        0,
					Timespan:    int64(30 * 24 * time.Hour),
					CurrencyID:  &currency.ID,
				},
			}
		case "MUR": // Mauritius
			subscription = []*SubscriptionPlan{
				{
					Name:                     "Enterprise Plan",
					Description:              "For large cooperatives requiring advanced AI and automation tools.",
					Cost:                     18000, // ~USD 400 equivalent
					Timespan:                 int64(30 * 24 * time.Hour),
					MaxBranches:              50,
					MaxEmployees:             1000,
					MaxMembersPerBranch:      500,
					Discount:                 15,
					YearlyDiscount:           25,
					IsRecommended:            false,
					HasAPIAccess:             true,
					HasFlexibleOrgStructures: true,
					HasAIEnabled:             true,
					HasMachineLearning:       true,
					MaxAPICallsPerMonth:      0,
					CurrencyID:               &currency.ID,
				},
				{
					Name:                     "Pro Plan",
					Description:              "Ideal for growing cooperatives with AI support.",
					Cost:                     9000,
					Timespan:                 int64(30 * 24 * time.Hour),
					MaxBranches:              20,
					MaxEmployees:             200,
					MaxMembersPerBranch:      100,
					Discount:                 10,
					YearlyDiscount:           20,
					IsRecommended:            true,
					HasAPIAccess:             true,
					HasFlexibleOrgStructures: true,
					HasAIEnabled:             true,
					HasMachineLearning:       false,
					MaxAPICallsPerMonth:      0,
					CurrencyID:               &currency.ID,
				},
				{
					Name:                     "Growth Plan",
					Description:              "Balanced plan for mid-sized cooperatives ready to scale.",
					Cost:                     4500,
					Timespan:                 int64(30 * 24 * time.Hour),
					MaxBranches:              8,
					MaxEmployees:             75,
					MaxMembersPerBranch:      50,
					Discount:                 7,
					YearlyDiscount:           17,
					IsRecommended:            false,
					HasAPIAccess:             true,
					HasFlexibleOrgStructures: true,
					HasAIEnabled:             false,
					HasMachineLearning:       false,
					MaxAPICallsPerMonth:      10000,
					CurrencyID:               &currency.ID,
				},
				{
					Name:                     "Starter Plan",
					Description:              "Affordable plan for small cooperatives just starting out.",
					Cost:                     2200,
					Timespan:                 int64(30 * 24 * time.Hour),
					MaxBranches:              3,
					MaxEmployees:             25,
					MaxMembersPerBranch:      25,
					Discount:                 5,
					YearlyDiscount:           15,
					IsRecommended:            false,
					HasAPIAccess:             true,
					HasFlexibleOrgStructures: false,
					HasAIEnabled:             false,
					HasMachineLearning:       false,
					MaxAPICallsPerMonth:      1000,
					CurrencyID:               &currency.ID,
				},
				{
					Name:                     "Free Plan",
					Description:              "Basic trial plan to explore core features.",
					Cost:                     0,
					Timespan:                 int64(30 * 24 * time.Hour),
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
					CurrencyID:               &currency.ID,
				},
			}
		case "MVR": // Maldives
			subscription = []*SubscriptionPlan{
				{
					Name:                     "Enterprise Plan",
					Description:              "Complete plan for large organizations needing AI and automation.",
					Cost:                     6200,
					Timespan:                 int64(30 * 24 * time.Hour),
					MaxBranches:              50,
					MaxEmployees:             1000,
					MaxMembersPerBranch:      500,
					Discount:                 15,
					YearlyDiscount:           25,
					IsRecommended:            false,
					HasAPIAccess:             true,
					HasFlexibleOrgStructures: true,
					HasAIEnabled:             true,
					HasMachineLearning:       true,
					MaxAPICallsPerMonth:      0,
					CurrencyID:               &currency.ID,
				},
				{
					Name:                     "Pro Plan",
					Description:              "Professional plan with AI features for scaling cooperatives.",
					Cost:                     3100,
					Timespan:                 int64(30 * 24 * time.Hour),
					MaxBranches:              20,
					MaxEmployees:             200,
					MaxMembersPerBranch:      100,
					Discount:                 10,
					YearlyDiscount:           20,
					IsRecommended:            true,
					HasAPIAccess:             true,
					HasFlexibleOrgStructures: true,
					HasAIEnabled:             true,
					HasMachineLearning:       false,
					MaxAPICallsPerMonth:      0,
					CurrencyID:               &currency.ID,
				},
				{
					Name:                     "Growth Plan",
					Description:              "Mid-tier plan for flexible and expanding cooperatives.",
					Cost:                     1600,
					Timespan:                 int64(30 * 24 * time.Hour),
					MaxBranches:              8,
					MaxEmployees:             75,
					MaxMembersPerBranch:      50,
					Discount:                 7,
					YearlyDiscount:           17,
					IsRecommended:            false,
					HasAPIAccess:             true,
					HasFlexibleOrgStructures: true,
					HasAIEnabled:             false,
					HasMachineLearning:       false,
					MaxAPICallsPerMonth:      10000,
					CurrencyID:               &currency.ID,
				},
				{
					Name:                     "Starter Plan",
					Description:              "Basic plan for small teams and new projects.",
					Cost:                     800,
					Timespan:                 int64(30 * 24 * time.Hour),
					MaxBranches:              3,
					MaxEmployees:             25,
					MaxMembersPerBranch:      25,
					Discount:                 5,
					YearlyDiscount:           15,
					IsRecommended:            false,
					HasAPIAccess:             true,
					HasFlexibleOrgStructures: false,
					HasAIEnabled:             false,
					HasMachineLearning:       false,
					MaxAPICallsPerMonth:      1000,
					CurrencyID:               &currency.ID,
				},
				{
					Name:                     "Free Plan",
					Description:              "Try essential tools for free.",
					Cost:                     0,
					Timespan:                 int64(30 * 24 * time.Hour),
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
					CurrencyID:               &currency.ID,
				},
			}
		case "NOK": // Norway
			subscription = []*SubscriptionPlan{
				{
					Name:                     "Enterprise Plan",
					Description:              "Advanced AI and ML plan for large organizations.",
					Cost:                     4200,
					Timespan:                 int64(30 * 24 * time.Hour),
					MaxBranches:              50,
					MaxEmployees:             1000,
					MaxMembersPerBranch:      500,
					Discount:                 15,
					YearlyDiscount:           25,
					IsRecommended:            false,
					HasAPIAccess:             true,
					HasFlexibleOrgStructures: true,
					HasAIEnabled:             true,
					HasMachineLearning:       true,
					MaxAPICallsPerMonth:      0,
					CurrencyID:               &currency.ID,
				},
				{
					Name:                     "Pro Plan",
					Description:              "Perfect for expanding organizations with AI capabilities.",
					Cost:                     2100,
					Timespan:                 int64(30 * 24 * time.Hour),
					MaxBranches:              20,
					MaxEmployees:             200,
					MaxMembersPerBranch:      100,
					Discount:                 10,
					YearlyDiscount:           20,
					IsRecommended:            true,
					HasAPIAccess:             true,
					HasFlexibleOrgStructures: true,
					HasAIEnabled:             true,
					HasMachineLearning:       false,
					MaxAPICallsPerMonth:      0,
					CurrencyID:               &currency.ID,
				},
				{
					Name:                     "Growth Plan",
					Description:              "Designed for mid-sized cooperatives ready to scale.",
					Cost:                     1100,
					Timespan:                 int64(30 * 24 * time.Hour),
					MaxBranches:              8,
					MaxEmployees:             75,
					MaxMembersPerBranch:      50,
					Discount:                 7,
					YearlyDiscount:           17,
					IsRecommended:            false,
					HasAPIAccess:             true,
					HasFlexibleOrgStructures: true,
					HasAIEnabled:             false,
					HasMachineLearning:       false,
					MaxAPICallsPerMonth:      10000,
					CurrencyID:               &currency.ID,
				},
				{
					Name:                     "Starter Plan",
					Description:              "Budget-friendly plan for small cooperatives.",
					Cost:                     600,
					Timespan:                 int64(30 * 24 * time.Hour),
					MaxBranches:              3,
					MaxEmployees:             25,
					MaxMembersPerBranch:      25,
					Discount:                 5,
					YearlyDiscount:           15,
					IsRecommended:            false,
					HasAPIAccess:             true,
					HasFlexibleOrgStructures: false,
					HasAIEnabled:             false,
					HasMachineLearning:       false,
					MaxAPICallsPerMonth:      1000,
					CurrencyID:               &currency.ID,
				},
				{
					Name:                     "Free Plan",
					Description:              "Free plan with limited features for testing.",
					Cost:                     0,
					Timespan:                 int64(30 * 24 * time.Hour),
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
					CurrencyID:               &currency.ID,
				},
			}
		case "DKK": // Denmark
			subscription = []*SubscriptionPlan{
				{
					Name:                     "Enterprise Plan",
					Description:              "Enterprise plan with AI, ML, and automation tools.",
					Cost:                     2800,
					Timespan:                 int64(30 * 24 * time.Hour),
					MaxBranches:              50,
					MaxEmployees:             1000,
					MaxMembersPerBranch:      500,
					Discount:                 15,
					YearlyDiscount:           25,
					IsRecommended:            false,
					HasAPIAccess:             true,
					HasFlexibleOrgStructures: true,
					HasAIEnabled:             true,
					HasMachineLearning:       true,
					MaxAPICallsPerMonth:      0,
					CurrencyID:               &currency.ID,
				},
				{
					Name:                     "Pro Plan",
					Description:              "AI-supported plan for fast-growing cooperatives.",
					Cost:                     1400,
					Timespan:                 int64(30 * 24 * time.Hour),
					MaxBranches:              20,
					MaxEmployees:             200,
					MaxMembersPerBranch:      100,
					Discount:                 10,
					YearlyDiscount:           20,
					IsRecommended:            true,
					HasAPIAccess:             true,
					HasFlexibleOrgStructures: true,
					HasAIEnabled:             true,
					HasMachineLearning:       false,
					MaxAPICallsPerMonth:      0,
					CurrencyID:               &currency.ID,
				},
				{
					Name:                     "Growth Plan",
					Description:              "Flexible plan for mid-sized cooperatives.",
					Cost:                     700,
					Timespan:                 int64(30 * 24 * time.Hour),
					MaxBranches:              8,
					MaxEmployees:             75,
					MaxMembersPerBranch:      50,
					Discount:                 7,
					YearlyDiscount:           17,
					IsRecommended:            false,
					HasAPIAccess:             true,
					HasFlexibleOrgStructures: true,
					HasAIEnabled:             false,
					HasMachineLearning:       false,
					MaxAPICallsPerMonth:      10000,
					CurrencyID:               &currency.ID,
				},
				{
					Name:                     "Starter Plan",
					Description:              "Simple plan for small and starting co-ops.",
					Cost:                     350,
					Timespan:                 int64(30 * 24 * time.Hour),
					MaxBranches:              3,
					MaxEmployees:             25,
					MaxMembersPerBranch:      25,
					Discount:                 5,
					YearlyDiscount:           15,
					IsRecommended:            false,
					HasAPIAccess:             true,
					HasFlexibleOrgStructures: false,
					HasAIEnabled:             false,
					HasMachineLearning:       false,
					MaxAPICallsPerMonth:      1000,
					CurrencyID:               &currency.ID,
				},
				{
					Name:                     "Free Plan",
					Description:              "Free trial for essential features.",
					Cost:                     0,
					Timespan:                 int64(30 * 24 * time.Hour),
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
					CurrencyID:               &currency.ID,
				},
			}
		case "PLN": // Poland
			subscription = []*SubscriptionPlan{
				{
					Name:                     "Enterprise Plan",
					Description:              "Full AI-enabled plan for large-scale cooperatives.",
					Cost:                     1650,
					Timespan:                 int64(30 * 24 * time.Hour),
					MaxBranches:              50,
					MaxEmployees:             1000,
					MaxMembersPerBranch:      500,
					Discount:                 15,
					YearlyDiscount:           25,
					IsRecommended:            false,
					HasAPIAccess:             true,
					HasFlexibleOrgStructures: true,
					HasAIEnabled:             true,
					HasMachineLearning:       true,
					MaxAPICallsPerMonth:      0,
					CurrencyID:               &currency.ID,
				},
				{
					Name:                     "Pro Plan",
					Description:              "AI-enabled plan for growing cooperatives.",
					Cost:                     850,
					Timespan:                 int64(30 * 24 * time.Hour),
					MaxBranches:              20,
					MaxEmployees:             200,
					MaxMembersPerBranch:      100,
					Discount:                 10,
					YearlyDiscount:           20,
					IsRecommended:            true,
					HasAPIAccess:             true,
					HasFlexibleOrgStructures: true,
					HasAIEnabled:             true,
					HasMachineLearning:       false,
					MaxAPICallsPerMonth:      0,
					CurrencyID:               &currency.ID,
				},
				{
					Name:                     "Growth Plan",
					Description:              "Flexible mid-tier plan for expanding teams.",
					Cost:                     430,
					Timespan:                 int64(30 * 24 * time.Hour),
					MaxBranches:              8,
					MaxEmployees:             75,
					MaxMembersPerBranch:      50,
					Discount:                 7,
					YearlyDiscount:           17,
					IsRecommended:            false,
					HasAPIAccess:             true,
					HasFlexibleOrgStructures: true,
					HasAIEnabled:             false,
					HasMachineLearning:       false,
					MaxAPICallsPerMonth:      10000,
					CurrencyID:               &currency.ID,
				},
				{
					Name:                     "Starter Plan",
					Description:              "Basic plan for small cooperatives starting out.",
					Cost:                     210,
					Timespan:                 int64(30 * 24 * time.Hour),
					MaxBranches:              3,
					MaxEmployees:             25,
					MaxMembersPerBranch:      25,
					Discount:                 5,
					YearlyDiscount:           15,
					IsRecommended:            false,
					HasAPIAccess:             true,
					HasFlexibleOrgStructures: false,
					HasAIEnabled:             false,
					HasMachineLearning:       false,
					MaxAPICallsPerMonth:      1000,
					CurrencyID:               &currency.ID,
				},
				{
					Name:                     "Free Plan",
					Description:              "Free plan with essential features.",
					Cost:                     0,
					Timespan:                 int64(30 * 24 * time.Hour),
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
					CurrencyID:               &currency.ID,
				},
			}
		case "CZK": // Czech Republic
			subscription = []*SubscriptionPlan{
				{
					Name:                     "Enterprise Plan",
					Description:              "Enterprise-level plan with unlimited features, AI/ML capabilities, and priority support.",
					Cost:                     10000, // in CZK
					Timespan:                 int64(30 * 24 * time.Hour),
					MaxBranches:              50,
					MaxEmployees:             1000,
					MaxMembersPerBranch:      500,
					Discount:                 15.00,
					YearlyDiscount:           25.00,
					IsRecommended:            false,
					HasAPIAccess:             true,
					HasFlexibleOrgStructures: true,
					HasAIEnabled:             true,
					HasMachineLearning:       true,
					MaxAPICallsPerMonth:      0,
					CurrencyID:               &currency.ID,
				},
				{
					Name:                     "Pro Plan",
					Description:              "Professional plan perfect for growing cooperatives with AI features.",
					Cost:                     5000, // in CZK
					Timespan:                 int64(30 * 24 * time.Hour),
					MaxBranches:              20,
					MaxEmployees:             200,
					MaxMembersPerBranch:      100,
					Discount:                 10.00,
					YearlyDiscount:           20.00,
					IsRecommended:            true,
					HasAPIAccess:             true,
					HasFlexibleOrgStructures: true,
					HasAIEnabled:             true,
					HasMachineLearning:       false,
					MaxAPICallsPerMonth:      0,
					CurrencyID:               &currency.ID,
				},
				{
					Name:                     "Growth Plan",
					Description:              "Balanced plan for mid‐sized co-ops ready to scale with flexible structures.",
					Cost:                     2500, // in CZK
					Timespan:                 int64(30 * 24 * time.Hour),
					MaxBranches:              8,
					MaxEmployees:             75,
					MaxMembersPerBranch:      50,
					Discount:                 7.50,
					YearlyDiscount:           17.50,
					IsRecommended:            false,
					HasAPIAccess:             true,
					HasFlexibleOrgStructures: true,
					HasAIEnabled:             false,
					HasMachineLearning:       false,
					MaxAPICallsPerMonth:      10000,
					CurrencyID:               &currency.ID,
				},
				{
					Name:                     "Starter Plan",
					Description:              "Affordable plan for small organizations just getting started.",
					Cost:                     1200, // in CZK
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
					CurrencyID:               &currency.ID,
				},
				{
					Name:                     "Free Plan",
					Description:              "Basic trial plan with essential features to get you started.",
					Cost:                     0,
					Timespan:                 int64(30 * 24 * time.Hour),
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
					CurrencyID:               &currency.ID,
				},
			}
		case "HUF": // Hungary
			subscription = []*SubscriptionPlan{
				{
					Name:                     "Enterprise Plan",
					Description:              "Enterprise-level plan with unlimited features, AI/ML capabilities, and priority support.",
					Cost:                     400000, // in HUF
					Timespan:                 int64(30 * 24 * time.Hour),
					MaxBranches:              50,
					MaxEmployees:             1000,
					MaxMembersPerBranch:      500,
					Discount:                 15.00,
					YearlyDiscount:           25.00,
					IsRecommended:            false,
					HasAPIAccess:             true,
					HasFlexibleOrgStructures: true,
					HasAIEnabled:             true,
					HasMachineLearning:       true,
					MaxAPICallsPerMonth:      0,
					CurrencyID:               &currency.ID,
				},
				{
					Name:                     "Pro Plan",
					Description:              "Professional plan perfect for growing cooperatives with AI features.",
					Cost:                     200000, // in HUF
					Timespan:                 int64(30 * 24 * time.Hour),
					MaxBranches:              20,
					MaxEmployees:             200,
					MaxMembersPerBranch:      100,
					Discount:                 10.00,
					YearlyDiscount:           20.00,
					IsRecommended:            true,
					HasAPIAccess:             true,
					HasFlexibleOrgStructures: true,
					HasAIEnabled:             true,
					HasMachineLearning:       false,
					MaxAPICallsPerMonth:      0,
					CurrencyID:               &currency.ID,
				},
				{
					Name:                     "Growth Plan",
					Description:              "Balanced plan for mid‐sized co-ops ready to scale with flexible structures.",
					Cost:                     100000, // in HUF
					Timespan:                 int64(30 * 24 * time.Hour),
					MaxBranches:              8,
					MaxEmployees:             75,
					MaxMembersPerBranch:      50,
					Discount:                 7.50,
					YearlyDiscount:           17.50,
					IsRecommended:            false,
					HasAPIAccess:             true,
					HasFlexibleOrgStructures: true,
					HasAIEnabled:             false,
					HasMachineLearning:       false,
					MaxAPICallsPerMonth:      10000,
					CurrencyID:               &currency.ID,
				},
				{
					Name:                     "Starter Plan",
					Description:              "Affordable plan for small organizations just getting started.",
					Cost:                     50000, // in HUF
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
					CurrencyID:               &currency.ID,
				},
				{
					Name:                     "Free Plan",
					Description:              "Basic trial plan with essential features to get you started.",
					Cost:                     0,
					Timespan:                 int64(30 * 24 * time.Hour),
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
					CurrencyID:               &currency.ID,
				},
			}
		case "RUB": // Russia
			subscription = []*SubscriptionPlan{
				{
					Name:                     "Enterprise Plan",
					Description:              "Enterprise-level plan with unlimited features, AI/ML capabilities, and priority support.",
					Cost:                     35000, // in RUB
					Timespan:                 int64(30 * 24 * time.Hour),
					MaxBranches:              50,
					MaxEmployees:             1000,
					MaxMembersPerBranch:      500,
					Discount:                 15.00,
					YearlyDiscount:           25.00,
					IsRecommended:            false,
					HasAPIAccess:             true,
					HasFlexibleOrgStructures: true,
					HasAIEnabled:             true,
					HasMachineLearning:       true,
					MaxAPICallsPerMonth:      0,
					CurrencyID:               &currency.ID,
				},
				{
					Name:                     "Pro Plan",
					Description:              "Professional plan perfect for growing cooperatives with AI features.",
					Cost:                     18000, // in RUB
					Timespan:                 int64(30 * 24 * time.Hour),
					MaxBranches:              20,
					MaxEmployees:             200,
					MaxMembersPerBranch:      100,
					Discount:                 10.00,
					YearlyDiscount:           20.00,
					IsRecommended:            true,
					HasAPIAccess:             true,
					HasFlexibleOrgStructures: true,
					HasAIEnabled:             true,
					HasMachineLearning:       false,
					MaxAPICallsPerMonth:      0,
					CurrencyID:               &currency.ID,
				},
				{
					Name:                     "Growth Plan",
					Description:              "Balanced plan for mid-sized co-ops ready to scale with flexible structures.",
					Cost:                     9000, // in RUB
					Timespan:                 int64(30 * 24 * time.Hour),
					MaxBranches:              8,
					MaxEmployees:             75,
					MaxMembersPerBranch:      50,
					Discount:                 7.50,
					YearlyDiscount:           17.50,
					IsRecommended:            false,
					HasAPIAccess:             true,
					HasFlexibleOrgStructures: true,
					HasAIEnabled:             false,
					HasMachineLearning:       false,
					MaxAPICallsPerMonth:      10000,
					CurrencyID:               &currency.ID,
				},
				{
					Name:                     "Starter Plan",
					Description:              "Affordable plan for small organizations just getting started.",
					Cost:                     5000, // in RUB
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
					CurrencyID:               &currency.ID,
				},
				{
					Name:                     "Free Plan",
					Description:              "Basic trial plan with essential features to get you started.",
					Cost:                     0,
					Timespan:                 int64(30 * 24 * time.Hour),
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
					CurrencyID:               &currency.ID,
				},
			}
		case "EUR-HR": // Croatia (uses Euro)
			subscription = []*SubscriptionPlan{
				{
					Name:                     "Enterprise Plan",
					Description:              "Enterprise-level plan with unlimited features, AI/ML capabilities, and priority support.",
					Cost:                     399, // in EUR
					Timespan:                 int64(30 * 24 * time.Hour),
					MaxBranches:              50,
					MaxEmployees:             1000,
					MaxMembersPerBranch:      500,
					Discount:                 15.00,
					YearlyDiscount:           25.00,
					IsRecommended:            false,
					HasAPIAccess:             true,
					HasFlexibleOrgStructures: true,
					HasAIEnabled:             true,
					HasMachineLearning:       true,
					MaxAPICallsPerMonth:      0,
					CurrencyID:               &currency.ID,
				},
				{
					Name:                     "Pro Plan",
					Description:              "Professional plan perfect for growing cooperatives with AI features.",
					Cost:                     199, // in EUR
					Timespan:                 int64(30 * 24 * time.Hour),
					MaxBranches:              20,
					MaxEmployees:             200,
					MaxMembersPerBranch:      100,
					Discount:                 10.00,
					YearlyDiscount:           20.00,
					IsRecommended:            true,
					HasAPIAccess:             true,
					HasFlexibleOrgStructures: true,
					HasAIEnabled:             true,
					HasMachineLearning:       false,
					MaxAPICallsPerMonth:      0,
					CurrencyID:               &currency.ID,
				},
				{
					Name:                     "Growth Plan",
					Description:              "Balanced plan for mid-sized co-ops ready to scale with flexible structures.",
					Cost:                     99, // in EUR
					Timespan:                 int64(30 * 24 * time.Hour),
					MaxBranches:              8,
					MaxEmployees:             75,
					MaxMembersPerBranch:      50,
					Discount:                 7.50,
					YearlyDiscount:           17.50,
					IsRecommended:            false,
					HasAPIAccess:             true,
					HasFlexibleOrgStructures: true,
					HasAIEnabled:             false,
					HasMachineLearning:       false,
					MaxAPICallsPerMonth:      10000,
					CurrencyID:               &currency.ID,
				},
				{
					Name:                     "Starter Plan",
					Description:              "Affordable plan for small organizations just getting started.",
					Cost:                     49, // in EUR
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
					CurrencyID:               &currency.ID,
				},
				{
					Name:                     "Free Plan",
					Description:              "Basic trial plan with essential features to get you started.",
					Cost:                     0,
					Timespan:                 int64(30 * 24 * time.Hour),
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
					CurrencyID:               &currency.ID,
				},
			}
		case "BRL": // Brazil
			subscription = []*SubscriptionPlan{}
		case "MXN": // Mexico
			subscription = []*SubscriptionPlan{}
		case "ARS": // Argentina
			subscription = []*SubscriptionPlan{}
		case "CLP": // Chile
			subscription = []*SubscriptionPlan{}
		case "COP": // Colombia
			subscription = []*SubscriptionPlan{}
		case "PEN": // Peru
			subscription = []*SubscriptionPlan{}
		case "UYU": // Uruguay
			subscription = []*SubscriptionPlan{}
		case "DOP": // Dominican Republic
			subscription = []*SubscriptionPlan{}
		case "PYG": // Paraguay
			subscription = []*SubscriptionPlan{}
		case "BOB": // Bolivia
			subscription = []*SubscriptionPlan{}
		case "VES": // Venezuela
			subscription = []*SubscriptionPlan{}
		case "PKR": // Pakistan
			subscription = []*SubscriptionPlan{}
		case "BDT": // Bangladesh
			subscription = []*SubscriptionPlan{}
		case "LKR": // Sri Lanka
			subscription = []*SubscriptionPlan{}
		case "NPR": // Nepal
			subscription = []*SubscriptionPlan{}
		case "MMK": // Myanmar
			subscription = []*SubscriptionPlan{}
		case "KHR": // Cambodia
			subscription = []*SubscriptionPlan{}
		case "LAK": // Laos
			subscription = []*SubscriptionPlan{}
		case "NGN": // Nigeria
			subscription = []*SubscriptionPlan{}
		case "KES": // Kenya
			subscription = []*SubscriptionPlan{}
		case "GHS": // Ghana
			subscription = []*SubscriptionPlan{}
		case "MAD": // Morocco
			subscription = []*SubscriptionPlan{}
		case "TND": // Tunisia
			subscription = []*SubscriptionPlan{}
		case "ETB": // Ethiopia
			subscription = []*SubscriptionPlan{}
		case "DZD": // Algeria
			subscription = []*SubscriptionPlan{}
		case "UAH": // Ukraine
			subscription = []*SubscriptionPlan{}
		case "RON": // Romania
			subscription = []*SubscriptionPlan{}
		case "BGN": // Bulgaria
			subscription = []*SubscriptionPlan{}
		case "RSD": // Serbia
			subscription = []*SubscriptionPlan{}
		case "ISK": // Iceland
			subscription = []*SubscriptionPlan{}
		case "BYN": // Belarus
			subscription = []*SubscriptionPlan{}
		case "FJD": // Fiji
			subscription = []*SubscriptionPlan{}
		case "PGK": // Papua New Guinea
			subscription = []*SubscriptionPlan{}
		case "JMD": // Jamaica
			subscription = []*SubscriptionPlan{}
		case "CRC": // Costa Rica
			subscription = []*SubscriptionPlan{}
		case "GTQ": // Guatemala
			subscription = []*SubscriptionPlan{}
		case "XDR": // Special Drawing Rights (IMF)
			subscription = []*SubscriptionPlan{}
		case "KWD": // Kuwait
			subscription = []*SubscriptionPlan{}
		case "QAR": // Qatar
			subscription = []*SubscriptionPlan{}
		case "OMR": // Oman
			subscription = []*SubscriptionPlan{}
		case "BHD": // Bahrain
			subscription = []*SubscriptionPlan{}
		case "JOD": // Jordan
			subscription = []*SubscriptionPlan{}
		case "KZT": // Kazakhstan
			subscription = []*SubscriptionPlan{}
		default:
		}
		for _, sub := range subscription {
			sub.CurrencyID = &currency.ID
			if err := m.SubscriptionPlanManager.Create(ctx, sub); err != nil {
				return eris.Wrapf(err, "failed to seed subscription %s for currency %s", sub.Name, currency.CurrencyCode)
			}
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
