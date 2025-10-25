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
		case "CHF": // Switzerland
			subscription = []*SubscriptionPlan{}
		case "CNY": // China
			subscription = []*SubscriptionPlan{}
		case "SEK": // Sweden
			subscription = []*SubscriptionPlan{}
		case "NZD": // New Zealand
			subscription = []*SubscriptionPlan{}
		case "PHP": // Philippines
			subscription = []*SubscriptionPlan{}
		case "INR": // India
			subscription = []*SubscriptionPlan{}
		case "KRW": // South Korea
			subscription = []*SubscriptionPlan{}
		case "THB": // Thailand
			subscription = []*SubscriptionPlan{}
		case "SGD": // Singapore
			subscription = []*SubscriptionPlan{}
		case "HKD": // Hong Kong
			subscription = []*SubscriptionPlan{}
		case "MYR": // Malaysia
			subscription = []*SubscriptionPlan{}
		case "IDR": // Indonesia
			subscription = []*SubscriptionPlan{}
		case "VND": // Vietnam
			subscription = []*SubscriptionPlan{}
		case "TWD": // Taiwan
			subscription = []*SubscriptionPlan{}
		case "BND": // Brunei
			subscription = []*SubscriptionPlan{}
		case "SAR": // Saudi Arabia
			subscription = []*SubscriptionPlan{}
		case "AED": // United Arab Emirates
			subscription = []*SubscriptionPlan{}
		case "ILS": // Israel
			subscription = []*SubscriptionPlan{}
		case "ZAR": // South Africa
			subscription = []*SubscriptionPlan{}
		case "EGP": // Egypt
			subscription = []*SubscriptionPlan{}
		case "TRY": // Turkey
			subscription = []*SubscriptionPlan{}
		case "XOF": // West African CFA Franc
			subscription = []*SubscriptionPlan{}
		case "XAF": // Central African CFA Franc
			subscription = []*SubscriptionPlan{}
		case "MUR": // Mauritius
			subscription = []*SubscriptionPlan{}
		case "MVR": // Maldives
			subscription = []*SubscriptionPlan{}
		case "NOK": // Norway
			subscription = []*SubscriptionPlan{}
		case "DKK": // Denmark
			subscription = []*SubscriptionPlan{}
		case "PLN": // Poland
			subscription = []*SubscriptionPlan{}
		case "CZK": // Czech Republic
			subscription = []*SubscriptionPlan{}
		case "HUF": // Hungary
			subscription = []*SubscriptionPlan{}
		case "RUB": // Russia
			subscription = []*SubscriptionPlan{}
		case "EUR-HR": // Croatia (Euro)
			subscription = []*SubscriptionPlan{}
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
