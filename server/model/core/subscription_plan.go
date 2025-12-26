package core

import (
	"context"
	"fmt"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/pkg/registry"
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

		HasAPIAccess             bool `gorm:"not null;default:false"` // False for free
		HasFlexibleOrgStructures bool `gorm:"not null;default:false"` // False for free
		HasAIEnabled             bool `gorm:"not null;default:false"`
		HasMachineLearning       bool `gorm:"not null;default:false"`

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

		HasAPIAccess             bool `json:"has_api_access"`
		HasFlexibleOrgStructures bool `json:"has_flexible_org_structures"`
		HasAIEnabled             bool `json:"has_ai_enabled"`
		HasMachineLearning       bool `json:"has_machine_learning"`

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

		HasAPIAccess             bool `json:"has_api_access"`
		HasFlexibleOrgStructures bool `json:"has_flexible_org_structures"`
		HasAIEnabled             bool `json:"has_ai_enabled"`
		HasMachineLearning       bool `json:"has_machine_learning"`

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

func (m *Core) SubscriptionPlanManager() *registry.Registry[SubscriptionPlan, SubscriptionPlanResponse, SubscriptionPlanRequest] {
	return registry.NewRegistry(registry.RegistryParams[SubscriptionPlan, SubscriptionPlanResponse, SubscriptionPlanRequest]{
		Preloads: []string{"Currency"},
		Database: m.provider.Service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return m.provider.Service.Broker.Dispatch(topics, payload)
		},
		Resource: func(sp *SubscriptionPlan) *SubscriptionPlanResponse {
			if sp == nil {
				return nil
			}

			decimal := m.provider.Service.Decimal

			monthlyPrice := m.provider.Service.Decimal.RoundToDecimalPlaces(sp.Cost, 2)

			yearlyPrice := m.provider.Service.Decimal.RoundToDecimalPlaces(
				decimal.Multiply(sp.Cost, 12), 2)

			discountedMonthlyPrice := m.provider.Service.Decimal.RoundToDecimalPlaces(
				decimal.SubtractPercentage(sp.Cost, sp.Discount), 2)

			discountedYearlyPrice := m.provider.Service.Decimal.RoundToDecimalPlaces(
				decimal.SubtractPercentage(yearlyPrice, sp.YearlyDiscount), 2)

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

				HasAPIAccess:             sp.HasAPIAccess,
				HasFlexibleOrgStructures: sp.HasFlexibleOrgStructures,
				HasAIEnabled:             sp.HasAIEnabled,
				HasMachineLearning:       sp.HasMachineLearning,

				MaxAPICallsPerMonth: sp.MaxAPICallsPerMonth,

				MonthlyPrice:           monthlyPrice,
				YearlyPrice:            yearlyPrice,
				DiscountedMonthlyPrice: discountedMonthlyPrice,
				DiscountedYearlyPrice:  discountedYearlyPrice,
				CreatedAt:              sp.CreatedAt.Format(time.RFC3339),
				UpdatedAt:              sp.UpdatedAt.Format(time.RFC3339),
				CurrencyID:             sp.CurrencyID,
				Currency:               m.CurrencyManager().ToModel(sp.Currency),
			}
		},

		Created: func(data *SubscriptionPlan) registry.Topics {
			return []string{
				"subscription_plan.create",
				fmt.Sprintf("subscription_plan.create.%s", data.ID),
			}
		},
		Updated: func(data *SubscriptionPlan) registry.Topics {
			return []string{
				"subscription_plan.update",
				fmt.Sprintf("subscription_plan.update.%s", data.ID),
			}
		},
		Deleted: func(data *SubscriptionPlan) registry.Topics {
			return []string{
				"subscription_plan.delete",
				fmt.Sprintf("subscription_plan.delete.%s", data.ID),
			}
		},
	})
}

func newSubscriptionPlan(name, description string, cost, discount, yearlyDiscount float64, tier string, currencyID *uuid.UUID) *SubscriptionPlan {
	p := &SubscriptionPlan{
		Name:           name,
		Description:    description,
		Cost:           cost,
		Discount:       discount,
		YearlyDiscount: yearlyDiscount,
		Timespan:       int64(30 * 24 * time.Hour),
		CurrencyID:     currencyID,
	}

	switch tier {
	case "enterprise":
		p.MaxBranches = 50
		p.MaxEmployees = 1000
		p.MaxMembersPerBranch = 500
		p.HasAPIAccess = true
		p.HasFlexibleOrgStructures = true
		p.HasAIEnabled = true
		p.HasMachineLearning = true
		p.MaxAPICallsPerMonth = 0 // Unlimited
		p.IsRecommended = false
	case "pro":
		p.MaxBranches = 20
		p.MaxEmployees = 200
		p.MaxMembersPerBranch = 100
		p.HasAPIAccess = true
		p.HasFlexibleOrgStructures = true
		p.HasAIEnabled = true
		p.HasMachineLearning = false
		p.MaxAPICallsPerMonth = 0 // Unlimited
		p.IsRecommended = true
	case "growth":
		p.MaxBranches = 8
		p.MaxEmployees = 75
		p.MaxMembersPerBranch = 50
		p.HasAPIAccess = true
		p.HasFlexibleOrgStructures = true
		p.HasAIEnabled = false
		p.HasMachineLearning = false
		p.MaxAPICallsPerMonth = 10000
		p.IsRecommended = false
	case "starter":
		p.MaxBranches = 3
		p.MaxEmployees = 25
		p.MaxMembersPerBranch = 25
		p.HasAPIAccess = true
		p.HasFlexibleOrgStructures = false
		p.HasAIEnabled = false
		p.HasMachineLearning = false
		p.MaxAPICallsPerMonth = 1000
		p.IsRecommended = false
	case "free":
		p.MaxBranches = 1
		p.MaxEmployees = 3
		p.MaxMembersPerBranch = 10
		p.HasAPIAccess = false
		p.HasFlexibleOrgStructures = false
		p.HasAIEnabled = false
		p.HasMachineLearning = false
		p.MaxAPICallsPerMonth = 100
		p.IsRecommended = false
	}

	return p
}

func (m *Core) subscriptionPlanSeed(ctx context.Context) error {
	subscriptionPlans, err := m.SubscriptionPlanManager().List(ctx)
	if err != nil {
		return err
	}
	if len(subscriptionPlans) >= 1 {
		return nil
	}

	currencies, err := m.CurrencyManager().List(ctx)
	if err != nil {
		return err
	}

	for _, currency := range currencies {
		var subscriptions []*SubscriptionPlan

		switch currency.ISO3166Alpha3 {
		case "USA": // United States
			subscriptions = []*SubscriptionPlan{
				newSubscriptionPlan("Enterprise Plan", "An enterprise-level plan with unlimited features, AI/ML capabilities, and priority support.", 399.99, 15.00, 25.00, "enterprise", &currency.ID),
				newSubscriptionPlan("Pro Plan", "A professional plan perfect for growing cooperatives with AI features.", 199.99, 10.00, 20.00, "pro", &currency.ID),
				newSubscriptionPlan("Growth Plan", "A balanced plan for mid-sized co-ops ready to scale with flexible structures.", 99.99, 7.50, 17.50, "growth", &currency.ID),
				newSubscriptionPlan("Starter Plan", "An affordable plan for small organizations just getting started.", 49.99, 5.00, 15.00, "starter", &currency.ID),
				newSubscriptionPlan("Free Plan", "A basic trial plan with essential features to get you started.", 0.00, 0.00, 0.00, "free", &currency.ID),
			}
		case "DEU": // Germany
			subscriptions = []*SubscriptionPlan{
				newSubscriptionPlan("Enterprise Plan", "Ein Unternehmensplan mit unbegrenzten Funktionen, KI-/ML-Fähigkeiten und priorisiertem Support.", 399, 15, 25, "enterprise", &currency.ID),
				newSubscriptionPlan("Pro Plan", "Ein professioneller Plan, ideal für wachsende Genossenschaften mit KI-Funktionen.", 199, 10, 20, "pro", &currency.ID),
				newSubscriptionPlan("Growth Plan", "Ein ausgewogener Plan für mittelgroße Genossenschaften, die bereit sind zu wachsen.", 99, 7, 17, "growth", &currency.ID),
				newSubscriptionPlan("Starter Plan", "Ein günstiger Plan für kleine Organisationen, die gerade erst anfangen.", 49, 5, 15, "starter", &currency.ID),
				newSubscriptionPlan("Free Plan", "Ein kostenloser Plan mit grundlegenden Funktionen für den Einstieg.", 0, 0, 0, "free", &currency.ID),
			}
		case "JPN": // Japan
			subscriptions = []*SubscriptionPlan{
				newSubscriptionPlan("エンタープライズプラン", "無制限の機能、AI/ML機能、そして優先サポートを備えた企業向けプランです。", 50000, 15, 25, "enterprise", &currency.ID),
				newSubscriptionPlan("プロプラン", "成長中の協同組合に最適な、AI機能を備えたプロフェッショナルプランです。", 25000, 10, 20, "pro", &currency.ID),
				newSubscriptionPlan("グロースプラン", "中規模の協同組合がスケールアップするためのバランスの取れたプランです。", 12000, 7, 17, "growth", &currency.ID),
				newSubscriptionPlan("スタータープラン", "小規模組織のための手頃な入門プランです。", 6000, 5, 15, "starter", &currency.ID),
				newSubscriptionPlan("無料プラン", "基本的な機能を備えた無料トライアルプランです。", 0, 0, 0, "free", &currency.ID),
			}
		case "GBR": // United Kingdom
			subscriptions = []*SubscriptionPlan{
				newSubscriptionPlan("Enterprise Plan", "An enterprise-level plan with unlimited features, AI/ML capabilities, and priority support.", 400, 15, 25, "enterprise", &currency.ID),
				newSubscriptionPlan("Pro Plan", "A professional plan ideal for growing cooperatives with AI support.", 200, 10, 20, "pro", &currency.ID),
				newSubscriptionPlan("Growth Plan", "A balanced plan for mid-sized co-ops aiming to expand with flexibility.", 100, 8, 18, "growth", &currency.ID),
				newSubscriptionPlan("Starter Plan", "An affordable plan for small organizations beginning their digital journey.", 50, 5, 15, "starter", &currency.ID),
				newSubscriptionPlan("Free Plan", "A basic plan with essential tools to help you start your cooperative journey.", 0, 0, 0, "free", &currency.ID),
			}
		case "AUS": // Australia
			subscriptions = []*SubscriptionPlan{
				newSubscriptionPlan("Enterprise Plan", "An enterprise-grade plan with unlimited access, AI and machine learning tools, and top-tier support.", 550, 15, 25, "enterprise", &currency.ID),
				newSubscriptionPlan("Pro Plan", "A professional plan ideal for expanding cooperatives with AI capabilities.", 280, 10, 20, "pro", &currency.ID),
				newSubscriptionPlan("Growth Plan", "A flexible plan for medium-sized co-ops ready to expand and modernise.", 140, 8, 18, "growth", &currency.ID),
				newSubscriptionPlan("Starter Plan", "A cost-effective plan for small teams starting their cooperative journey.", 70, 5, 15, "starter", &currency.ID),
				newSubscriptionPlan("Free Plan", "A simple free plan with essential features for testing and evaluation.", 0, 0, 0, "free", &currency.ID),
			}
		case "CAN": // Canada
			subscriptions = []*SubscriptionPlan{
				newSubscriptionPlan("Enterprise Plan", "A top-tier plan with unlimited access, advanced AI/ML tools, and premium support for large cooperatives.", 540, 15, 25, "enterprise", &currency.ID),
				newSubscriptionPlan("Pro Plan", "A professional plan for growing cooperatives with AI features and flexibility.", 270, 10, 20, "pro", &currency.ID),
				newSubscriptionPlan("Growth Plan", "A flexible plan for mid-sized co-ops looking to scale efficiently.", 135, 8, 18, "growth", &currency.ID),
				newSubscriptionPlan("Starter Plan", "An affordable plan designed for small organizations getting started.", 65, 5, 15, "starter", &currency.ID),
				newSubscriptionPlan("Free Plan", "A basic free plan offering essential tools to help you get started.", 0, 0, 0, "free", &currency.ID),
			}
		case "CHE": // Switzerland
			subscriptions = []*SubscriptionPlan{
				newSubscriptionPlan("Enterprise Plan", "An enterprise-level plan with full AI and machine learning capabilities.", 399, 15, 25, "enterprise", &currency.ID),
				newSubscriptionPlan("Pro Plan", "A professional plan for scaling organizations with AI features.", 199, 10, 20, "pro", &currency.ID),
				newSubscriptionPlan("Growth Plan", "A balanced plan for growing organizations.", 99, 7, 17, "growth", &currency.ID),
				newSubscriptionPlan("Starter Plan", "An affordable plan for small cooperatives.", 49, 5, 15, "starter", &currency.ID),
				newSubscriptionPlan("Free Plan", "Try our platform free for 30 days.", 0, 0, 0, "free", &currency.ID),
			}
		case "CHN": // China
			subscriptions = []*SubscriptionPlan{
				newSubscriptionPlan("企业版", "适用于大型企业，包含全部 AI 和机器学习功能。", 2800, 15, 25, "enterprise", &currency.ID),
				newSubscriptionPlan("专业版", "适合成长型合作社，包含AI功能。", 1400, 10, 20, "pro", &currency.ID),
				newSubscriptionPlan("成长版", "适用于中型合作社的灵活方案。", 700, 7, 17, "growth", &currency.ID),
				newSubscriptionPlan("入门版", "适合小型组织的经济型方案。", 300, 5, 15, "starter", &currency.ID),
				newSubscriptionPlan("免费版", "提供30天免费试用。", 0, 0, 0, "free", &currency.ID),
			}
		case "SWE": // Sweden
			subscriptions = []*SubscriptionPlan{
				newSubscriptionPlan("Enterprise Plan", "Complete AI and ML package for large cooperatives.", 3999, 15, 25, "enterprise", &currency.ID),
				newSubscriptionPlan("Pro Plan", "Perfect for expanding organizations with AI support.", 1999, 10, 20, "pro", &currency.ID),
				newSubscriptionPlan("Growth Plan", "For mid-sized cooperatives aiming to scale.", 999, 7, 17, "growth", &currency.ID),
				newSubscriptionPlan("Starter Plan", "Simple plan for small organizations.", 499, 5, 15, "starter", &currency.ID),
				newSubscriptionPlan("Free Plan", "Free 30-day access to essential tools.", 0, 0, 0, "free", &currency.ID),
			}
		case "NZL": // New Zealand
			subscriptions = []*SubscriptionPlan{
				newSubscriptionPlan("Enterprise Plan", "Unlimited features with AI/ML and enterprise tools.", 599, 15, 25, "enterprise", &currency.ID),
				newSubscriptionPlan("Pro Plan", "Best for professional growth with AI integration.", 299, 10, 20, "pro", &currency.ID),
				newSubscriptionPlan("Growth Plan", "Ideal for scaling co-ops and small businesses.", 149, 7, 17, "growth", &currency.ID),
				newSubscriptionPlan("Starter Plan", "Affordable plan for beginners.", 79, 5, 15, "starter", &currency.ID),
				newSubscriptionPlan("Free Plan", "Free access for 30 days to try our platform.", 0, 0, 0, "free", &currency.ID),
			}
		case "PHL": // Philippines
			subscriptions = []*SubscriptionPlan{
				newSubscriptionPlan("Enterprise Plan", "For large cooperatives with full AI and machine learning tools.", 24999, 15, 25, "enterprise", &currency.ID),
				newSubscriptionPlan("Pro Plan", "Perfect for growing cooperatives with AI tools.", 12499, 10, 20, "pro", &currency.ID),
				newSubscriptionPlan("Growth Plan", "Ideal for mid-sized organizations ready to scale.", 6499, 7, 17, "growth", &currency.ID),
				newSubscriptionPlan("Starter Plan", "A great choice for small and new cooperatives.", 2999, 5, 15, "starter", &currency.ID),
				newSubscriptionPlan("Free Plan", "Free access for 30 days to explore our platform.", 0, 0, 0, "free", &currency.ID),
			}
		case "IND": // India
			subscriptions = []*SubscriptionPlan{
				newSubscriptionPlan("Enterprise Plan", "Full-featured AI and ML suite for large organizations.", 24999, 15, 25, "enterprise", &currency.ID),
				newSubscriptionPlan("Pro Plan", "AI-ready plan for growing cooperatives.", 12499, 10, 20, "pro", &currency.ID),
				newSubscriptionPlan("Growth Plan", "Mid-tier plan for scaling cooperatives.", 6999, 7, 17, "growth", &currency.ID),
				newSubscriptionPlan("Starter Plan", "Affordable plan for small organizations.", 2999, 5, 15, "starter", &currency.ID),
				newSubscriptionPlan("Free Plan", "Free 30-day trial with basic tools.", 0, 0, 0, "free", &currency.ID),
			}
		case "KOR": // South Korea
			subscriptions = []*SubscriptionPlan{
				newSubscriptionPlan("Enterprise Plan", "대기업을 위한 AI/ML 기능과 무제한 지원을 제공합니다.", 499000, 15, 25, "enterprise", &currency.ID),
				newSubscriptionPlan("Pro Plan", "AI 기능이 포함된 성장형 협동조합을 위한 전문 플랜입니다.", 249000, 10, 20, "pro", &currency.ID),
				newSubscriptionPlan("Growth Plan", "중형 조직을 위한 유연한 확장형 플랜입니다.", 129000, 7, 17, "growth", &currency.ID),
				newSubscriptionPlan("Starter Plan", "소규모 협동조합을 위한 경제적인 플랜입니다.", 59000, 5, 15, "starter", &currency.ID),
				newSubscriptionPlan("Free Plan", "기본 기능으로 30일 무료 체험이 가능합니다.", 0, 0, 0, "free", &currency.ID),
			}
		case "THA": // Thailand
			subscriptions = []*SubscriptionPlan{
				newSubscriptionPlan("Enterprise Plan", "Complete enterprise solution with AI/ML capabilities.", 13999, 15, 25, "enterprise", &currency.ID),
				newSubscriptionPlan("Pro Plan", "Ideal for growing cooperatives with AI features.", 6999, 10, 20, "pro", &currency.ID),
				newSubscriptionPlan("Growth Plan", "Balanced plan for mid-sized co-ops.", 3499, 7, 17, "growth", &currency.ID),
				newSubscriptionPlan("Starter Plan", "Affordable plan for small organizations.", 1499, 5, 15, "starter", &currency.ID),
				newSubscriptionPlan("Free Plan", "30-day free trial with basic tools.", 0, 0, 0, "free", &currency.ID),
			}
		case "SGP": // Singapore
			subscriptions = []*SubscriptionPlan{
				newSubscriptionPlan("Enterprise Plan", "Comprehensive AI/ML plan for enterprise-level co-ops.", 499, 15, 25, "enterprise", &currency.ID),
				newSubscriptionPlan("Pro Plan", "AI-powered plan for growing cooperatives.", 249, 10, 20, "pro", &currency.ID),
				newSubscriptionPlan("Growth Plan", "Perfect for medium-sized co-ops ready to scale.", 129, 7, 17, "growth", &currency.ID),
				newSubscriptionPlan("Starter Plan", "Affordable option for small co-ops.", 59, 5, 15, "starter", &currency.ID),
				newSubscriptionPlan("Free Plan", "Free 30-day access with basic tools.", 0, 0, 0, "free", &currency.ID),
			}
		case "HKG": // Hong Kong
			subscriptions = []*SubscriptionPlan{
				newSubscriptionPlan("Enterprise Plan", "Full-featured enterprise plan with AI/ML tools.", 2999, 15, 25, "enterprise", &currency.ID),
				newSubscriptionPlan("Pro Plan", "Advanced plan for growing co-ops with AI tools.", 1499, 10, 20, "pro", &currency.ID),
				newSubscriptionPlan("Growth Plan", "Balanced plan for scaling organizations.", 799, 7, 17, "growth", &currency.ID),
				newSubscriptionPlan("Starter Plan", "Great choice for small co-ops.", 399, 5, 15, "starter", &currency.ID),
				newSubscriptionPlan("Free Plan", "30-day free trial to explore our platform.", 0, 0, 0, "free", &currency.ID),
			}
		case "MYS": // Malaysia
			subscriptions = []*SubscriptionPlan{
				newSubscriptionPlan("Enterprise Plan", "Enterprise-level plan with AI and ML features.", 1499, 15, 25, "enterprise", &currency.ID),
				newSubscriptionPlan("Pro Plan", "For growing organizations with AI features.", 799, 10, 20, "pro", &currency.ID),
				newSubscriptionPlan("Growth Plan", "Balanced plan for mid-sized co-ops.", 399, 7, 17, "growth", &currency.ID),
				newSubscriptionPlan("Starter Plan", "Affordable plan for small organizations.", 199, 5, 15, "starter", &currency.ID),
				newSubscriptionPlan("Free Plan", "Free access for 30 days with essential tools.", 0, 0, 0, "free", &currency.ID),
			}
		case "IDN": // Indonesia
			subscriptions = []*SubscriptionPlan{
				newSubscriptionPlan("Paket Enterprise", "Solusi lengkap untuk koperasi besar dengan fitur AI/ML.", 3990000, 15, 25, "enterprise", &currency.ID),
				newSubscriptionPlan("Paket Pro", "Untuk koperasi yang sedang berkembang dengan fitur AI.", 1990000, 10, 20, "pro", &currency.ID),
				newSubscriptionPlan("Paket Growth", "Cocok untuk koperasi menengah yang ingin berkembang.", 999000, 7, 17, "growth", &currency.ID),
				newSubscriptionPlan("Paket Starter", "Pilihan hemat untuk koperasi kecil.", 499000, 5, 15, "starter", &currency.ID),
				newSubscriptionPlan("Paket Gratis", "Uji coba gratis selama 30 hari dengan fitur dasar.", 0, 0, 0, "free", &currency.ID),
			}
		case "VNM": // Vietnam
			subscriptions = []*SubscriptionPlan{
				newSubscriptionPlan("Gói Doanh Nghiệp", "Gói cao cấp với đầy đủ tính năng AI và học máy.", 4990000, 15, 25, "enterprise", &currency.ID),
				newSubscriptionPlan("Gói Chuyên Nghiệp", "Phù hợp với hợp tác xã đang phát triển cùng AI.", 2490000, 10, 20, "pro", &currency.ID),
				newSubscriptionPlan("Gói Tăng Trưởng", "Dành cho hợp tác xã quy mô vừa muốn mở rộng.", 1290000, 7, 17, "growth", &currency.ID),
				newSubscriptionPlan("Gói Khởi Đầu", "Gói tiết kiệm cho tổ chức nhỏ.", 599000, 5, 15, "starter", &currency.ID),
				newSubscriptionPlan("Gói Miễn Phí", "Dùng thử 30 ngày miễn phí với tính năng cơ bản.", 0, 0, 0, "free", &currency.ID),
			}
		case "TWN": // Taiwan
			subscriptions = []*SubscriptionPlan{
				newSubscriptionPlan("Enterprise Plan", "Comprehensive enterprise solution with AI and ML capabilities.", 11999, 15, 25, "enterprise", &currency.ID),
				newSubscriptionPlan("Pro Plan", "For growing cooperatives with AI tools and premium support.", 5999, 10, 20, "pro", &currency.ID),
				newSubscriptionPlan("Growth Plan", "Mid-tier plan for scaling co-ops with flexibility.", 2999, 7, 17, "growth", &currency.ID),
				newSubscriptionPlan("Starter Plan", "Affordable entry plan for small organizations.", 1299, 5, 15, "starter", &currency.ID),
				newSubscriptionPlan("Free Plan", "Try the essential tools free for 30 days.", 0, 0, 0, "free", &currency.ID),
			}
		case "BRN": // Brunei
			subscriptions = []*SubscriptionPlan{
				newSubscriptionPlan("Enterprise Plan", "Enterprise-grade plan with full AI and ML capabilities.", 499, 15, 25, "enterprise", &currency.ID),
				newSubscriptionPlan("Pro Plan", "For growing cooperatives with AI integration.", 249, 10, 20, "pro", &currency.ID),
				newSubscriptionPlan("Growth Plan", "Flexible plan for expanding co-ops.", 129, 7, 17, "growth", &currency.ID),
				newSubscriptionPlan("Starter Plan", "Affordable plan for small organizations.", 59, 5, 15, "starter", &currency.ID),
				newSubscriptionPlan("Free Plan", "Free 30-day trial with core features.", 0, 0, 0, "free", &currency.ID),
			}
		case "SAU": // Saudi Arabia
			subscriptions = []*SubscriptionPlan{
				newSubscriptionPlan("Enterprise Plan", "Full-featured enterprise solution with AI/ML.", 1499, 15, 25, "enterprise", &currency.ID),
				newSubscriptionPlan("Pro Plan", "AI-ready plan for growing organizations.", 799, 10, 20, "pro", &currency.ID),
				newSubscriptionPlan("Growth Plan", "Ideal for mid-sized co-ops looking to expand.", 399, 7, 17, "growth", &currency.ID),
				newSubscriptionPlan("Starter Plan", "Affordable plan for small organizations.", 199, 5, 15, "starter", &currency.ID),
				newSubscriptionPlan("Free Plan", "Free trial for 30 days with basic tools.", 0, 0, 0, "free", &currency.ID),
			}
		case "ARE": // United Arab Emirates
			subscriptions = []*SubscriptionPlan{
				newSubscriptionPlan("Enterprise Plan", "Advanced enterprise-level solution with AI/ML.", 1499, 15, 25, "enterprise", &currency.ID),
				newSubscriptionPlan("Pro Plan", "AI-powered plan for growing cooperatives.", 749, 10, 20, "pro", &currency.ID),
				newSubscriptionPlan("Growth Plan", "Balanced plan for scaling co-ops.", 349, 7, 17, "growth", &currency.ID),
				newSubscriptionPlan("Starter Plan", "Starter option for small organizations.", 179, 5, 15, "starter", &currency.ID),
				newSubscriptionPlan("Free Plan", "Free trial for 30 days with core tools.", 0, 0, 0, "free", &currency.ID),
			}
		case "ISR": // Israel
			subscriptions = []*SubscriptionPlan{
				newSubscriptionPlan("Enterprise Plan", "Comprehensive enterprise plan with AI/ML and unlimited features.", 1499, 15, 25, "enterprise", &currency.ID),
				newSubscriptionPlan("Pro Plan", "Perfect for growing cooperatives with AI support.", 799, 10, 20, "pro", &currency.ID),
				newSubscriptionPlan("Growth Plan", "Mid-tier flexible plan for scaling cooperatives.", 399, 7, 17, "growth", &currency.ID),
				newSubscriptionPlan("Starter Plan", "Entry-level plan for small organizations.", 179, 5, 15, "starter", &currency.ID),
				newSubscriptionPlan("Free Plan", "Free 30-day access with basic features.", 0, 0, 0, "free", &currency.ID),
			}
		case "ZAF": // South Africa
			subscriptions = []*SubscriptionPlan{
				newSubscriptionPlan("Enterprise Plan", "Top-tier plan with AI, machine learning, and unlimited features.", 7000, 15, 25, "enterprise", &currency.ID),
				newSubscriptionPlan("Pro Plan", "Perfect for growing cooperatives with AI tools.", 3500, 10, 20, "pro", &currency.ID),
				newSubscriptionPlan("Growth Plan", "For mid-sized co-ops ready to expand operations.", 1800, 7, 17, "growth", &currency.ID),
				newSubscriptionPlan("Starter Plan", "Affordable option for new organizations.", 900, 5, 15, "starter", &currency.ID),
				newSubscriptionPlan("Free Plan", "Free 30-day trial with basic features.", 0, 0, 0, "free", &currency.ID),
			}
		case "EGY": // Egypt
			subscriptions = []*SubscriptionPlan{
				newSubscriptionPlan("خطة المؤسسات", "الخطة الشاملة للمؤسسات الكبيرة مع ميزات الذكاء الاصطناعي والدعم المميز.", 12000, 15, 25, "enterprise", &currency.ID),
				newSubscriptionPlan("الخطة الاحترافية", "الخطة المثالية للمؤسسات النامية مع أدوات الذكاء الاصطناعي.", 6000, 10, 20, "pro", &currency.ID),
				newSubscriptionPlan("خطة النمو", "للمؤسسات المتوسطة التي تستعد للتوسع.", 3000, 7, 17, "growth", &currency.ID),
				newSubscriptionPlan("الخطة المبتدئة", "خطة مناسبة للمؤسسات الصغيرة.", 1500, 5, 15, "starter", &currency.ID),
				newSubscriptionPlan("الخطة المجانية", "تجربة مجانية لمدة 30 يومًا مع ميزات أساسية.", 0, 0, 0, "free", &currency.ID),
			}
		case "TUR": // Turkey
			subscriptions = []*SubscriptionPlan{
				newSubscriptionPlan("Kurumsal Plan", "Sınırsız özellikler, AI/ML desteği ve öncelikli destek içeren üst düzey plan.", 9000, 15, 25, "enterprise", &currency.ID),
				newSubscriptionPlan("Profesyonel Plan", "Büyüyen kooperatifler için profesyonel plan.", 4500, 10, 20, "pro", &currency.ID),
				newSubscriptionPlan("Büyüme Planı", "Orta ölçekli kurumlar için dengeli bir plan.", 2200, 7, 17, "growth", &currency.ID),
				newSubscriptionPlan("Başlangıç Planı", "Yeni başlayanlar için uygun fiyatlı plan.", 1000, 5, 15, "starter", &currency.ID),
				newSubscriptionPlan("Ücretsiz Plan", "Temel özelliklerle 30 günlük ücretsiz deneme.", 0, 0, 0, "free", &currency.ID),
			}
		case "SEN": // Senegal (West African CFA)
			subscriptions = []*SubscriptionPlan{
				newSubscriptionPlan("Plan Entreprise", "Plan complet avec IA et assistance prioritaire.", 2400000, 15, 25, "enterprise", &currency.ID),
				newSubscriptionPlan("Plan Pro", "Idéal pour les coopératives en croissance avec des outils d'IA.", 1200000, 10, 20, "pro", &currency.ID),
				newSubscriptionPlan("Plan Croissance", "Pour les structures prêtes à se développer.", 600000, 7, 17, "growth", &currency.ID),
				newSubscriptionPlan("Plan Débutant", "Plan économique pour les petites structures.", 300000, 5, 15, "starter", &currency.ID),
				newSubscriptionPlan("Plan Gratuit", "Essai gratuit de 30 jours avec fonctions de base.", 0, 0, 0, "free", &currency.ID),
			}
		case "CMR": // Cameroon (Central African CFA)
			subscriptions = []*SubscriptionPlan{
				newSubscriptionPlan("Plan Entreprise", "Plan complet avec IA, ML et support premium.", 2400000, 15, 25, "enterprise", &currency.ID),
				newSubscriptionPlan("Plan Pro", "Plan professionnel pour coopératives en expansion.", 1200000, 10, 20, "pro", &currency.ID),
				newSubscriptionPlan("Plan Croissance", "Plan équilibré pour structures moyennes.", 600000, 7, 17, "growth", &currency.ID),
				newSubscriptionPlan("Plan Débutant", "Plan abordable pour petites organisations.", 300000, 5, 15, "starter", &currency.ID),
				newSubscriptionPlan("Plan Gratuit", "Essai gratuit de 30 jours avec fonctionnalités limitées.", 0, 0, 0, "free", &currency.ID),
			}
		case "MUS": // Mauritius
			subscriptions = []*SubscriptionPlan{
				newSubscriptionPlan("Enterprise Plan", "For large cooperatives requiring advanced AI and automation tools.", 18000, 15, 25, "enterprise", &currency.ID),
				newSubscriptionPlan("Pro Plan", "Ideal for growing cooperatives with AI support.", 9000, 10, 20, "pro", &currency.ID),
				newSubscriptionPlan("Growth Plan", "Balanced plan for mid-sized cooperatives ready to scale.", 4500, 7, 17, "growth", &currency.ID),
				newSubscriptionPlan("Starter Plan", "Affordable plan for small cooperatives just starting out.", 2200, 5, 15, "starter", &currency.ID),
				newSubscriptionPlan("Free Plan", "Basic trial plan to explore core features.", 0, 0, 0, "free", &currency.ID),
			}
		case "MDV": // Maldives
			subscriptions = []*SubscriptionPlan{
				newSubscriptionPlan("Enterprise Plan", "Complete plan for large organizations needing AI and automation.", 6200, 15, 25, "enterprise", &currency.ID),
				newSubscriptionPlan("Pro Plan", "Professional plan with AI features for scaling cooperatives.", 3100, 10, 20, "pro", &currency.ID),
				newSubscriptionPlan("Growth Plan", "Mid-tier plan for flexible and expanding cooperatives.", 1600, 7, 17, "growth", &currency.ID),
				newSubscriptionPlan("Starter Plan", "Basic plan for small teams and new projects.", 800, 5, 15, "starter", &currency.ID),
				newSubscriptionPlan("Free Plan", "Try essential tools for free.", 0, 0, 0, "free", &currency.ID),
			}
		case "NOR": // Norway
			subscriptions = []*SubscriptionPlan{
				newSubscriptionPlan("Enterprise Plan", "Advanced AI and ML plan for large organizations.", 4200, 15, 25, "enterprise", &currency.ID),
				newSubscriptionPlan("Pro Plan", "Perfect for expanding organizations with AI capabilities.", 2100, 10, 20, "pro", &currency.ID),
				newSubscriptionPlan("Growth Plan", "Designed for mid-sized cooperatives ready to scale.", 1100, 7, 17, "growth", &currency.ID),
				newSubscriptionPlan("Starter Plan", "Budget-friendly plan for small cooperatives.", 600, 5, 15, "starter", &currency.ID),
				newSubscriptionPlan("Free Plan", "Free plan with limited features for testing.", 0, 0, 0, "free", &currency.ID),
			}
		case "DNK": // Denmark
			subscriptions = []*SubscriptionPlan{
				newSubscriptionPlan("Enterprise Plan", "Enterprise plan with AI, ML, and automation tools.", 2800, 15, 25, "enterprise", &currency.ID),
				newSubscriptionPlan("Pro Plan", "AI-supported plan for fast-growing cooperatives.", 1400, 10, 20, "pro", &currency.ID),
				newSubscriptionPlan("Growth Plan", "Flexible plan for mid-sized cooperatives.", 700, 7, 17, "growth", &currency.ID),
				newSubscriptionPlan("Starter Plan", "Simple plan for small and starting co-ops.", 350, 5, 15, "starter", &currency.ID),
				newSubscriptionPlan("Free Plan", "Free trial for essential features.", 0, 0, 0, "free", &currency.ID),
			}
		case "POL": // Poland
			subscriptions = []*SubscriptionPlan{
				newSubscriptionPlan("Enterprise Plan", "Full AI-enabled plan for large-scale cooperatives.", 1650, 15, 25, "enterprise", &currency.ID),
				newSubscriptionPlan("Pro Plan", "AI-enabled plan for growing cooperatives.", 850, 10, 20, "pro", &currency.ID),
				newSubscriptionPlan("Growth Plan", "Flexible mid-tier plan for expanding teams.", 430, 7, 17, "growth", &currency.ID),
				newSubscriptionPlan("Starter Plan", "Basic plan for small cooperatives starting out.", 210, 5, 15, "starter", &currency.ID),
				newSubscriptionPlan("Free Plan", "Free plan with essential features.", 0, 0, 0, "free", &currency.ID),
			}
		case "CZE": // Czech Republic
			subscriptions = []*SubscriptionPlan{
				newSubscriptionPlan("Enterprise Plan", "Enterprise-level plan with unlimited features, AI/ML capabilities, and priority support.", 10000, 15.00, 25.00, "enterprise", &currency.ID),
				newSubscriptionPlan("Pro Plan", "Professional plan perfect for growing cooperatives with AI features.", 5000, 10.00, 20.00, "pro", &currency.ID),
				newSubscriptionPlan("Growth Plan", "Balanced plan for mid‐sized co-ops ready to scale with flexible structures.", 2500, 7.50, 17.50, "growth", &currency.ID),
				newSubscriptionPlan("Starter Plan", "Affordable plan for small organizations just getting started.", 1200, 5.00, 15.00, "starter", &currency.ID),
				newSubscriptionPlan("Free Plan", "Basic trial plan with essential features to get you started.", 0, 0, 0, "free", &currency.ID),
			}
		case "HUN": // Hungary
			subscriptions = []*SubscriptionPlan{
				newSubscriptionPlan("Enterprise Plan", "Enterprise-level plan with unlimited features, AI/ML capabilities, and priority support.", 400000, 15.00, 25.00, "enterprise", &currency.ID),
				newSubscriptionPlan("Pro Plan", "Professional plan perfect for growing cooperatives with AI features.", 200000, 10.00, 20.00, "pro", &currency.ID),
				newSubscriptionPlan("Growth Plan", "Balanced plan for mid‐sized co-ops ready to scale with flexible structures.", 100000, 7.50, 17.50, "growth", &currency.ID),
				newSubscriptionPlan("Starter Plan", "Affordable plan for small organizations just getting started.", 50000, 5.00, 15.00, "starter", &currency.ID),
				newSubscriptionPlan("Free Plan", "Basic trial plan with essential features to get you started.", 0, 0, 0, "free", &currency.ID),
			}
		case "RUS": // Russia
			subscriptions = []*SubscriptionPlan{
				newSubscriptionPlan("Enterprise Plan", "Enterprise-level plan with unlimited features, AI/ML capabilities, and priority support.", 35000, 15.00, 25.00, "enterprise", &currency.ID),
				newSubscriptionPlan("Pro Plan", "Professional plan perfect for growing cooperatives with AI features.", 18000, 10.00, 20.00, "pro", &currency.ID),
				newSubscriptionPlan("Growth Plan", "Balanced plan for mid-sized co-ops ready to scale with flexible structures.", 9000, 7.50, 17.50, "growth", &currency.ID),
				newSubscriptionPlan("Starter Plan", "Affordable plan for small organizations just getting started.", 5000, 5.00, 15.00, "starter", &currency.ID),
				newSubscriptionPlan("Free Plan", "Basic trial plan with essential features to get you started.", 0, 0, 0, "free", &currency.ID),
			}
		case "HRV": // Croatia
			subscriptions = []*SubscriptionPlan{
				newSubscriptionPlan("Enterprise Plan", "Enterprise-level plan with unlimited features, AI/ML capabilities, and priority support.", 399, 15.00, 25.00, "enterprise", &currency.ID),
				newSubscriptionPlan("Pro Plan", "Professional plan perfect for growing cooperatives with AI features.", 199, 10.00, 20.00, "pro", &currency.ID),
				newSubscriptionPlan("Growth Plan", "Balanced plan for mid-sized co-ops ready to scale with flexible structures.", 99, 7.50, 17.50, "growth", &currency.ID),
				newSubscriptionPlan("Starter Plan", "Affordable plan for small organizations just getting started.", 49, 5.00, 15.00, "starter", &currency.ID),
				newSubscriptionPlan("Free Plan", "Basic trial plan with essential features to get you started.", 0, 0, 0, "free", &currency.ID),
			}
		case "BRA": // Brazil
			subscriptions = []*SubscriptionPlan{
				newSubscriptionPlan("Plano Empresarial", "Plano empresarial com recursos ilimitados, IA e suporte prioritário.", 1999, 15, 25, "enterprise", &currency.ID),
				newSubscriptionPlan("Plano Profissional", "Plano ideal para cooperativas em crescimento com recursos de IA.", 999, 10, 20, "pro", &currency.ID),
				newSubscriptionPlan("Plano Crescimento", "Plano equilibrado para cooperativas médias que desejam expandir.", 499, 8, 18, "growth", &currency.ID),
				newSubscriptionPlan("Plano Inicial", "Plano acessível para pequenas organizações iniciantes.", 249, 5, 15, "starter", &currency.ID),
				newSubscriptionPlan("Plano Gratuito", "Plano básico com recursos essenciais para começar.", 0, 0, 0, "free", &currency.ID),
			}
		case "MEX": // Mexico
			subscriptions = []*SubscriptionPlan{
				newSubscriptionPlan("Plan Empresarial", "Plan empresarial con funciones ilimitadas, IA y soporte prioritario.", 6999, 15, 25, "enterprise", &currency.ID),
				newSubscriptionPlan("Plan Profesional", "Plan ideal para cooperativas en crecimiento con funciones de IA.", 3499, 10, 20, "pro", &currency.ID),
				newSubscriptionPlan("Plan Crecimiento", "Plan equilibrado para cooperativas medianas listas para expandirse.", 1799, 8, 18, "growth", &currency.ID),
				newSubscriptionPlan("Plan Inicial", "Plan asequible para pequeñas organizaciones que comienzan.", 899, 5, 15, "starter", &currency.ID),
				newSubscriptionPlan("Plan Gratis", "Plan básico con funciones esenciales para comenzar.", 0, 0, 0, "free", &currency.ID),
			}
		case "ARG": // Argentina
			subscriptions = []*SubscriptionPlan{
				newSubscriptionPlan("Plan Empresarial", "Plan empresarial con funciones ilimitadas, IA y soporte prioritario.", 250000, 15, 25, "enterprise", &currency.ID),
				newSubscriptionPlan("Plan Profesional", "Plan ideal para cooperativas en crecimiento con IA.", 125000, 10, 20, "pro", &currency.ID),
				newSubscriptionPlan("Plan Crecimiento", "Plan equilibrado para cooperativas medianas listas para expandirse.", 65000, 8, 18, "growth", &currency.ID),
				newSubscriptionPlan("Plan Inicial", "Plan asequible para pequeñas organizaciones que comienzan.", 30000, 5, 15, "starter", &currency.ID),
				newSubscriptionPlan("Plan Gratis", "Plan básico con funciones esenciales para comenzar.", 0, 0, 0, "free", &currency.ID),
			}
		case "CHL": // Chile
			subscriptions = []*SubscriptionPlan{
				newSubscriptionPlan("Plan Empresarial", "Plan empresarial con funciones ilimitadas, IA y soporte prioritario.", 350000, 15, 25, "enterprise", &currency.ID),
				newSubscriptionPlan("Plan Profesional", "Plan ideal para cooperativas en crecimiento con IA.", 175000, 10, 20, "pro", &currency.ID),
				newSubscriptionPlan("Plan Crecimiento", "Plan equilibrado para cooperativas medianas listas para expandirse.", 90000, 8, 18, "growth", &currency.ID),
				newSubscriptionPlan("Plan Inicial", "Plan asequible para pequeñas organizaciones que comienzan.", 45000, 5, 15, "starter", &currency.ID),
				newSubscriptionPlan("Plan Gratis", "Plan básico con funciones esenciales para comenzar.", 0, 0, 0, "free", &currency.ID),
			}
		case "COL": // Colombia
			subscriptions = []*SubscriptionPlan{
				newSubscriptionPlan("Plan Empresarial", "Plan empresarial con funciones ilimitadas, IA y soporte prioritario.", 1500000, 15, 25, "enterprise", &currency.ID),
				newSubscriptionPlan("Plan Profesional", "Plan ideal para cooperativas en crecimiento con IA.", 750000, 10, 20, "pro", &currency.ID),
				newSubscriptionPlan("Plan Crecimiento", "Plan equilibrado para cooperativas medianas listas para expandirse.", 400000, 8, 18, "growth", &currency.ID),
				newSubscriptionPlan("Plan Inicial", "Plan asequible para pequeñas organizaciones que comienzan.", 200000, 5, 15, "starter", &currency.ID),
				newSubscriptionPlan("Plan Gratis", "Plan básico con funciones esenciales para comenzar.", 0, 0, 0, "free", &currency.ID),
			}
		case "PER": // Peru
			subscriptions = []*SubscriptionPlan{
				newSubscriptionPlan("Plan Empresarial", "Un plan empresarial con funciones ilimitadas, IA y soporte prioritario.", 1500, 15, 25, "enterprise", &currency.ID),
				newSubscriptionPlan("Plan Profesional", "Ideal para cooperativas en crecimiento con funciones de IA.", 750, 10, 20, "pro", &currency.ID),
				newSubscriptionPlan("Plan Crecimiento", "Un plan equilibrado para cooperativas medianas listas para expandirse.", 380, 8, 17, "growth", &currency.ID),
				newSubscriptionPlan("Plan Inicial", "Un plan accesible para organizaciones pequeñas que empiezan.", 190, 5, 15, "starter", &currency.ID),
				newSubscriptionPlan("Plan Gratuito", "Plan básico de prueba con funciones esenciales.", 0, 0, 0, "free", &currency.ID),
			}
		case "URY": // Uruguay
			subscriptions = []*SubscriptionPlan{
				newSubscriptionPlan("Plan Empresarial", "Plan empresarial con todas las funciones y soporte prioritario.", 16000, 15, 25, "enterprise", &currency.ID),
				newSubscriptionPlan("Plan Profesional", "Ideal para cooperativas en expansión con IA integrada.", 8000, 10, 20, "pro", &currency.ID),
				newSubscriptionPlan("Plan Crecimiento", "Diseñado para cooperativas medianas listas para escalar.", 4000, 8, 17, "growth", &currency.ID),
				newSubscriptionPlan("Plan Inicial", "Plan básico para organizaciones pequeñas.", 2000, 5, 15, "starter", &currency.ID),
				newSubscriptionPlan("Plan Gratuito", "Plan de prueba con funciones limitadas.", 0, 0, 0, "free", &currency.ID),
			}
		case "DOM": // Dominican Republic
			subscriptions = []*SubscriptionPlan{
				newSubscriptionPlan("Plan Empresarial", "Plan empresarial con todas las funciones y soporte prioritario.", 22000, 15, 25, "enterprise", &currency.ID),
				newSubscriptionPlan("Plan Profesional", "Ideal para cooperativas en crecimiento con herramientas de IA.", 11000, 10, 20, "pro", &currency.ID),
				newSubscriptionPlan("Plan Crecimiento", "Un plan equilibrado para cooperativas medianas.", 5500, 8, 17, "growth", &currency.ID),
				newSubscriptionPlan("Plan Inicial", "Plan accesible para organizaciones pequeñas.", 2700, 5, 15, "starter", &currency.ID),
				newSubscriptionPlan("Plan Gratuito", "Prueba gratuita con funciones básicas.", 0, 0, 0, "free", &currency.ID),
			}
		case "PRY": // Paraguay
			subscriptions = []*SubscriptionPlan{
				newSubscriptionPlan("Plan Empresarial", "Plan empresarial con todas las funciones y soporte premium.", 1600000, 15, 25, "enterprise", &currency.ID),
				newSubscriptionPlan("Plan Profesional", "Ideal para cooperativas en expansión con IA.", 800000, 10, 20, "pro", &currency.ID),
				newSubscriptionPlan("Plan Crecimiento", "Plan equilibrado para cooperativas medianas.", 400000, 8, 17, "growth", &currency.ID),
				newSubscriptionPlan("Plan Inicial", "Plan básico para pequeñas organizaciones.", 200000, 5, 15, "starter", &currency.ID),
				newSubscriptionPlan("Plan Gratuito", "Plan gratuito de prueba con funciones básicas.", 0, 0, 0, "free", &currency.ID),
			}
		case "BOL": // Bolivia
			subscriptions = []*SubscriptionPlan{
				newSubscriptionPlan("Plan Empresarial", "Plan empresarial con IA y soporte prioritario.", 2700, 15, 25, "enterprise", &currency.ID),
				newSubscriptionPlan("Plan Profesional", "Plan ideal para cooperativas en crecimiento con IA.", 1350, 10, 20, "pro", &currency.ID),
				newSubscriptionPlan("Plan Crecimiento", "Plan equilibrado para cooperativas medianas.", 700, 8, 17, "growth", &currency.ID),
				newSubscriptionPlan("Plan Inicial", "Plan accesible para organizaciones pequeñas.", 350, 5, 15, "starter", &currency.ID),
				newSubscriptionPlan("Plan Gratuito", "Plan básico de prueba con funciones limitadas.", 0, 0, 0, "free", &currency.ID),
			}
		case "VEN": // Venezuela
			subscriptions = []*SubscriptionPlan{
				newSubscriptionPlan("Enterprise Plan", "A complete plan for large cooperatives with full AI and automation tools.", 16000, 15, 25, "enterprise", &currency.ID),
				newSubscriptionPlan("Pro Plan", "Best for growing co-ops with smart AI features and flexibility.", 8000, 10, 20, "pro", &currency.ID),
				newSubscriptionPlan("Growth Plan", "A practical plan for mid-sized cooperatives ready to scale.", 4000, 8, 17, "growth", &currency.ID),
				newSubscriptionPlan("Starter Plan", "For small co-ops just starting out.", 2000, 5, 15, "starter", &currency.ID),
				newSubscriptionPlan("Free Plan", "Basic trial plan to explore essential features.", 0, 0, 0, "free", &currency.ID),
			}
		case "PAK": // Pakistan
			subscriptions = []*SubscriptionPlan{
				newSubscriptionPlan("Enterprise Plan", "Enterprise solution with AI, automation, and large-scale management tools.", 115000, 15, 25, "enterprise", &currency.ID),
				newSubscriptionPlan("Pro Plan", "Smart and affordable plan for expanding cooperatives.", 58000, 10, 20, "pro", &currency.ID),
				newSubscriptionPlan("Growth Plan", "For expanding co-ops looking for flexibility.", 29000, 8, 17, "growth", &currency.ID),
				newSubscriptionPlan("Starter Plan", "Affordable plan for small teams.", 15000, 5, 15, "starter", &currency.ID),
				newSubscriptionPlan("Free Plan", "Basic trial version with limited access.", 0, 0, 0, "free", &currency.ID),
			}
		case "BGD": // Bangladesh
			subscriptions = []*SubscriptionPlan{
				newSubscriptionPlan("Enterprise Plan", "Enterprise-grade plan with automation and AI tools for co-ops.", 44000, 15, 25, "enterprise", &currency.ID),
				newSubscriptionPlan("Pro Plan", "Smart and affordable plan for expanding cooperatives.", 22000, 10, 20, "pro", &currency.ID),
				newSubscriptionPlan("Growth Plan", "Perfect for mid-sized co-ops aiming to scale.", 11000, 8, 17, "growth", &currency.ID),
				newSubscriptionPlan("Starter Plan", "A simple plan to get started.", 6000, 5, 15, "starter", &currency.ID),
				newSubscriptionPlan("Free Plan", "Free 30-day plan with basic tools.", 0, 0, 0, "free", &currency.ID),
			}
		case "LKA": // Sri Lanka
			subscriptions = []*SubscriptionPlan{
				newSubscriptionPlan("Enterprise Plan", "Full-featured enterprise plan for large organizations.", 125000, 15, 25, "enterprise", &currency.ID),
				newSubscriptionPlan("Pro Plan", "Professional plan with smart tools for co-ops.", 63000, 10, 20, "pro", &currency.ID),
				newSubscriptionPlan("Growth Plan", "A mid-tier plan for scaling co-ops.", 32000, 8, 17, "growth", &currency.ID),
				newSubscriptionPlan("Starter Plan", "Simple and affordable entry-level plan.", 16000, 5, 15, "starter", &currency.ID),
				newSubscriptionPlan("Free Plan", "Free plan with limited access.", 0, 0, 0, "free", &currency.ID),
			}
		case "NPL": // Nepal
			subscriptions = []*SubscriptionPlan{
				newSubscriptionPlan("Enterprise Plan", "Advanced enterprise plan with full AI tools.", 53000, 15, 25, "enterprise", &currency.ID),
				newSubscriptionPlan("Pro Plan", "Perfect for growing co-ops with modern tools.", 27000, 10, 20, "pro", &currency.ID),
				newSubscriptionPlan("Growth Plan", "A mid-level plan for scaling operations.", 14000, 8, 17, "growth", &currency.ID),
				newSubscriptionPlan("Starter Plan", "Entry-level plan for small cooperatives.", 7000, 5, 15, "starter", &currency.ID),
				newSubscriptionPlan("Free Plan", "Free 30-day plan with limited access for testing.", 0, 0, 0, "free", &currency.ID),
			}
		case "MMR": // Myanmar
			subscriptions = []*SubscriptionPlan{
				newSubscriptionPlan("Enterprise Plan", "An enterprise-level plan with unlimited features, AI/ML capabilities, and priority support.", 840000, 15, 25, "enterprise", &currency.ID),
				newSubscriptionPlan("Pro Plan", "A professional plan perfect for growing cooperatives with AI features.", 420000, 10, 20, "pro", &currency.ID),
				newSubscriptionPlan("Growth Plan", "A balanced plan for mid-sized co-ops ready to scale with flexible structures.", 210000, 7.5, 17.5, "growth", &currency.ID),
				newSubscriptionPlan("Starter Plan", "An affordable plan for small organizations just getting started.", 105000, 5, 15, "starter", &currency.ID),
				newSubscriptionPlan("Free Plan", "A basic trial plan with essential features to get you started.", 0, 0, 0, "free", &currency.ID),
			}
		case "KHM": // Cambodia
			subscriptions = []*SubscriptionPlan{
				newSubscriptionPlan("Enterprise Plan", "Enterprise plan with unlimited features, AI/ML and priority support.", 1600000, 15, 25, "enterprise", &currency.ID),
				newSubscriptionPlan("Pro Plan", "Professional plan ideal for growing co-ops with AI features.", 800000, 10, 20, "pro", &currency.ID),
				newSubscriptionPlan("Growth Plan", "Balanced plan for mid-sized co-ops ready to grow.", 400000, 8, 18, "growth", &currency.ID),
				newSubscriptionPlan("Starter Plan", "Affordable plan for small organizations just starting out.", 200000, 5, 15, "starter", &currency.ID),
				newSubscriptionPlan("Free Plan", "Basic trial plan with essential features to get started.", 0, 0, 0, "free", &currency.ID),
			}
		case "LAO": // Laos
			subscriptions = []*SubscriptionPlan{
				newSubscriptionPlan("Enterprise Plan", "Enterprise-level plan with unlimited features, AI/ML support.", 8600000, 15, 25, "enterprise", &currency.ID),
				newSubscriptionPlan("Pro Plan", "Professional plan for growing organisations with AI capabilities.", 4300000, 10, 20, "pro", &currency.ID),
				newSubscriptionPlan("Growth Plan", "Balanced plan for mid-sized organisations ready to scale.", 2150000, 8, 18, "growth", &currency.ID),
				newSubscriptionPlan("Starter Plan", "Affordable plan for small teams just starting.", 1080000, 5, 15, "starter", &currency.ID),
				newSubscriptionPlan("Free Plan", "Basic trial plan with essential features to get you started.", 0, 0, 0, "free", &currency.ID),
			}
		case "NGA": // Nigeria
			subscriptions = []*SubscriptionPlan{
				newSubscriptionPlan("Enterprise Plan", "Enterprise-level plan with unlimited features, AI/ML and priority support.", 590000, 15, 25, "enterprise", &currency.ID),
				newSubscriptionPlan("Pro Plan", "Professional plan ideal for growing cooperatives with AI features.", 295000, 10, 20, "pro", &currency.ID),
				newSubscriptionPlan("Growth Plan", "Balanced plan for mid-sized co-ops ready to scale with flexible structures.", 147000, 8, 18, "growth", &currency.ID),
				newSubscriptionPlan("Starter Plan", "Affordable plan for small organisations just getting started.", 74000, 5, 15, "starter", &currency.ID),
				newSubscriptionPlan("Free Plan", "Basic trial plan with essential features to get you started.", 0, 0, 0, "free", &currency.ID),
			}
		case "KEN": // Kenya
			subscriptions = []*SubscriptionPlan{
				newSubscriptionPlan("Enterprise Plan", "An advanced plan for large cooperatives with full AI and analytics capabilities.", 48000, 15, 25, "enterprise", &currency.ID),
				newSubscriptionPlan("Pro Plan", "A professional plan ideal for expanding SACCOs and cooperatives.", 24000, 10, 20, "pro", &currency.ID),
				newSubscriptionPlan("Growth Plan", "Designed for mid-sized SACCOs seeking scalable operations.", 12000, 8, 17, "growth", &currency.ID),
				newSubscriptionPlan("Starter Plan", "An affordable plan for small SACCOs starting their digital journey.", 6000, 5, 15, "starter", &currency.ID),
				newSubscriptionPlan("Free Plan", "A basic plan with limited features to get started.", 0, 0, 0, "free", &currency.ID),
			}
		case "GHA": // Ghana
			subscriptions = []*SubscriptionPlan{
				newSubscriptionPlan("Enterprise Plan", "A complete plan for large credit unions and cooperatives with AI features.", 4800, 15, 25, "enterprise", &currency.ID),
				newSubscriptionPlan("Pro Plan", "Perfect for expanding cooperatives and credit unions.", 2400, 10, 20, "pro", &currency.ID),
				newSubscriptionPlan("Growth Plan", "Ideal for growing organizations seeking flexibility.", 1200, 8, 17, "growth", &currency.ID),
				newSubscriptionPlan("Starter Plan", "An entry-level plan for small cooperatives.", 600, 5, 15, "starter", &currency.ID),
				newSubscriptionPlan("Free Plan", "A basic plan with limited access for beginners.", 0, 0, 0, "free", &currency.ID),
			}
		case "MAR": // Morocco
			subscriptions = []*SubscriptionPlan{
				newSubscriptionPlan("Plan Entreprise", "Une solution complète pour les grandes coopératives avec IA et analyses avancées.", 4000, 15, 25, "enterprise", &currency.ID),
				newSubscriptionPlan("Plan Pro", "Idéal pour les coopératives en expansion avec fonctions IA.", 2000, 10, 20, "pro", &currency.ID),
				newSubscriptionPlan("Plan Croissance", "Un plan équilibré pour les coopératives de taille moyenne.", 1000, 8, 17, "growth", &currency.ID),
				newSubscriptionPlan("Plan Débutant", "Une offre abordable pour les petites organisations.", 500, 5, 15, "starter", &currency.ID),
				newSubscriptionPlan("Plan Gratuit", "Plan de base avec des fonctionnalités limitées.", 0, 0, 0, "free", &currency.ID),
			}
		case "TUN": // Tunisia
			subscriptions = []*SubscriptionPlan{
				newSubscriptionPlan("Plan Entreprise", "Solution complète pour les grandes entreprises avec IA et apprentissage automatique.", 1200, 15, 25, "enterprise", &currency.ID),
				newSubscriptionPlan("Plan Pro", "Plan professionnel pour les organisations en croissance.", 600, 10, 20, "pro", &currency.ID),
				newSubscriptionPlan("Plan Croissance", "Plan pour les coopératives de taille moyenne.", 300, 8, 17, "growth", &currency.ID),
				newSubscriptionPlan("Plan Débutant", "Plan abordable pour les petites coopératives.", 150, 5, 15, "starter", &currency.ID),
				newSubscriptionPlan("Plan Gratuit", "Plan d’essai avec fonctionnalités limitées.", 0, 0, 0, "free", &currency.ID),
			}
		case "ETH": // Ethiopia
			subscriptions = []*SubscriptionPlan{
				newSubscriptionPlan("Enterprise Plan", "Comprehensive enterprise plan with unlimited features, AI/ML tools, and full support.", 23000, 15, 25, "enterprise", &currency.ID),
				newSubscriptionPlan("Pro Plan", "Ideal for growing cooperatives with AI-powered insights.", 11500, 10, 20, "pro", &currency.ID),
				newSubscriptionPlan("Growth Plan", "Flexible plan for medium organizations ready to expand.", 6000, 8, 18, "growth", &currency.ID),
				newSubscriptionPlan("Starter Plan", "Basic plan for small organizations starting their journey.", 3000, 5, 15, "starter", &currency.ID),
				newSubscriptionPlan("Free Plan", "Trial plan with essential features for 30 days.", 0, 0, 0, "free", &currency.ID),
			}
		case "DZA": // Algeria
			subscriptions = []*SubscriptionPlan{
				newSubscriptionPlan("Enterprise Plan", "Offre complète avec toutes les fonctionnalités, IA/ML et assistance prioritaire.", 55000, 15, 25, "enterprise", &currency.ID),
				newSubscriptionPlan("Pro Plan", "Parfait pour les coopératives en croissance avec des fonctions IA.", 28000, 10, 20, "pro", &currency.ID),
				newSubscriptionPlan("Growth Plan", "Solution flexible pour les organisations de taille moyenne.", 14000, 8, 18, "growth", &currency.ID),
				newSubscriptionPlan("Starter Plan", "Offre abordable pour les petites structures débutantes.", 7000, 5, 15, "starter", &currency.ID),
				newSubscriptionPlan("Free Plan", "Essai gratuit de 30 jours avec les fonctions de base.", 0, 0, 0, "free", &currency.ID),
			}
		case "UKR": // Ukraine
			subscriptions = []*SubscriptionPlan{
				newSubscriptionPlan("Enterprise Plan", "Повний корпоративний план з усіма функціями, AI/ML та пріоритетною підтримкою.", 16000, 15, 25, "enterprise", &currency.ID),
				newSubscriptionPlan("Pro Plan", "Професійний план для зростаючих кооперативів із підтримкою AI.", 8000, 10, 20, "pro", &currency.ID),
				newSubscriptionPlan("Growth Plan", "Збалансований план для середніх організацій, готових до розширення.", 4000, 8, 18, "growth", &currency.ID),
				newSubscriptionPlan("Starter Plan", "Початковий план для малих організацій.", 2000, 5, 15, "starter", &currency.ID),
				newSubscriptionPlan("Free Plan", "Безкоштовний план із базовими функціями на 30 днів.", 0, 0, 0, "free", &currency.ID),
			}
		case "ROU": // Romania
			subscriptions = []*SubscriptionPlan{
				newSubscriptionPlan("Enterprise Plan", "Plan complet pentru organizații mari, cu funcții AI/ML și suport prioritar.", 1800, 15, 25, "enterprise", &currency.ID),
				newSubscriptionPlan("Pro Plan", "Plan profesional pentru cooperative în creștere, cu funcții AI.", 900, 10, 20, "pro", &currency.ID),
				newSubscriptionPlan("Growth Plan", "Plan echilibrat pentru organizații mijlocii care doresc extindere.", 450, 8, 18, "growth", &currency.ID),
				newSubscriptionPlan("Starter Plan", "Plan accesibil pentru organizații mici.", 230, 5, 15, "starter", &currency.ID),
				newSubscriptionPlan("Free Plan", "Plan gratuit cu funcții de bază, valabil 30 de zile.", 0, 0, 0, "free", &currency.ID),
			}
		case "BGR": // Bulgaria
			subscriptions = []*SubscriptionPlan{
				newSubscriptionPlan("Enterprise Plan", "План за големи организации с неограничени функции и поддръжка с приоритет.", 700, 15, 25, "enterprise", &currency.ID),
				newSubscriptionPlan("Pro Plan", "Идеален план за растящи кооперации с AI функции.", 350, 10, 20, "pro", &currency.ID),
				newSubscriptionPlan("Growth Plan", "Подходящ за средни кооперации, готови да се разрастват.", 180, 8, 18, "growth", &currency.ID),
				newSubscriptionPlan("Starter Plan", "Достъпен план за малки организации.", 90, 5, 15, "starter", &currency.ID),
				newSubscriptionPlan("Free Plan", "Безплатен пробен план с основни функции.", 0, 0, 0, "free", &currency.ID),
			}
		case "SRB": // Serbia
			subscriptions = []*SubscriptionPlan{
				newSubscriptionPlan("Enterprise Plan", "Enterprise plan with unlimited tools and premium support.", 46000, 15, 25, "enterprise", &currency.ID),
				newSubscriptionPlan("Pro Plan", "Perfect for growing cooperatives with AI features.", 23000, 10, 20, "pro", &currency.ID),
				newSubscriptionPlan("Growth Plan", "Great value for mid-sized organizations.", 11500, 8, 18, "growth", &currency.ID),
				newSubscriptionPlan("Starter Plan", "Affordable plan for new co-ops.", 5500, 5, 15, "starter", &currency.ID),
				newSubscriptionPlan("Free Plan", "Basic trial plan for 30 days.", 0, 0, 0, "free", &currency.ID),
			}
		case "ISL": // Iceland
			subscriptions = []*SubscriptionPlan{
				newSubscriptionPlan("Enterprise Plan", "Full-scale plan with unlimited access and AI tools.", 56000, 15, 25, "enterprise", &currency.ID),
				newSubscriptionPlan("Pro Plan", "Ideal for professional cooperatives.", 28000, 10, 20, "pro", &currency.ID),
				newSubscriptionPlan("Growth Plan", "Balanced option for scaling up.", 14000, 8, 18, "growth", &currency.ID),
				newSubscriptionPlan("Starter Plan", "Simple and affordable plan.", 7000, 5, 15, "starter", &currency.ID),
				newSubscriptionPlan("Free Plan", "Free trial plan with core features.", 0, 0, 0, "free", &currency.ID),
			}
		case "BLR": // Belarus
			subscriptions = []*SubscriptionPlan{
				newSubscriptionPlan("Enterprise Plan", "План для крупных кооперативов с неограниченными возможностями и поддержкой.", 1300, 15, 25, "enterprise", &currency.ID),
				newSubscriptionPlan("Pro Plan", "Подходит для развивающихся организаций с AI.", 650, 10, 20, "pro", &currency.ID),
				newSubscriptionPlan("Growth Plan", "Хороший выбор для среднего бизнеса.", 320, 8, 18, "growth", &currency.ID),
				newSubscriptionPlan("Starter Plan", "Доступный стартовый план.", 160, 5, 15, "starter", &currency.ID),
				newSubscriptionPlan("Free Plan", "Бесплатный план с основными функциями.", 0, 0, 0, "free", &currency.ID),
			}
		case "FJI": // Fiji
			subscriptions = []*SubscriptionPlan{
				newSubscriptionPlan("Enterprise Plan", "An enterprise-level plan with unlimited features, AI/ML capabilities, and priority support.", 900, 15, 25, "enterprise", &currency.ID),
				newSubscriptionPlan("Pro Plan", "A professional plan perfect for growing cooperatives with AI features.", 450, 10, 20, "pro", &currency.ID),
				newSubscriptionPlan("Growth Plan", "A balanced plan for mid-sized co-ops ready to scale with flexible structures.", 220, 8, 18, "growth", &currency.ID),
				newSubscriptionPlan("Starter Plan", "An affordable plan for small organizations just getting started.", 110, 5, 15, "starter", &currency.ID),
				newSubscriptionPlan("Free Plan", "A basic trial plan with essential features to get you started.", 0, 0, 0, "free", &currency.ID),
			}
		case "PNG": // Papua New Guinea
			subscriptions = []*SubscriptionPlan{
				newSubscriptionPlan("Enterprise Plan", "An enterprise-level plan with unlimited features and AI/ML support for large cooperatives.", 1300, 15, 25, "enterprise", &currency.ID),
				newSubscriptionPlan("Pro Plan", "A professional plan designed for growing organizations with AI tools.", 650, 10, 20, "pro", &currency.ID),
				newSubscriptionPlan("Growth Plan", "A balanced plan for cooperatives aiming to expand operations efficiently.", 300, 8, 18, "growth", &currency.ID),
				newSubscriptionPlan("Starter Plan", "An entry-level plan for small cooperatives.", 150, 5, 15, "starter", &currency.ID),
				newSubscriptionPlan("Free Plan", "A free plan to explore the platform's essential features.", 0, 0, 0, "free", &currency.ID),
			}
		case "JAM": // Jamaica
			subscriptions = []*SubscriptionPlan{
				newSubscriptionPlan("Enterprise Plan", "Top-tier plan with advanced AI tools, machine learning, and premium support.", 65000, 15, 25, "enterprise", &currency.ID),
				newSubscriptionPlan("Pro Plan", "Perfect for professional cooperatives looking to grow with AI-driven tools.", 32000, 10, 20, "pro", &currency.ID),
				newSubscriptionPlan("Growth Plan", "Ideal for mid-size organizations expanding their operations.", 15000, 8, 18, "growth", &currency.ID),
				newSubscriptionPlan("Starter Plan", "Affordable plan for startups and small cooperatives.", 7500, 5, 15, "starter", &currency.ID),
				newSubscriptionPlan("Free Plan", "Free 30-day trial with essential tools and limited access.", 0, 0, 0, "free", &currency.ID),
			}
		case "CRI": // Costa Rica
			subscriptions = []*SubscriptionPlan{
				newSubscriptionPlan("Plan Empresarial", "Plan empresarial con funciones ilimitadas, inteligencia artificial y soporte prioritario.", 250000, 15, 25, "enterprise", &currency.ID),
				newSubscriptionPlan("Plan Profesional", "Un plan ideal para cooperativas en crecimiento con herramientas de IA.", 120000, 10, 20, "pro", &currency.ID),
				newSubscriptionPlan("Plan de Crecimiento", "Plan equilibrado para cooperativas medianas que desean escalar.", 60000, 8, 18, "growth", &currency.ID),
				newSubscriptionPlan("Plan Inicial", "Plan económico para organizaciones pequeñas que comienzan.", 30000, 5, 15, "starter", &currency.ID),
				newSubscriptionPlan("Plan Gratuito", "Plan básico gratuito con acceso limitado para probar la plataforma.", 0, 0, 0, "free", &currency.ID),
			}
		case "GTM": // Guatemala
			subscriptions = []*SubscriptionPlan{
				newSubscriptionPlan("Enterprise Plan", "An enterprise-level plan with unlimited features, AI/ML capabilities, and priority support.", 400, 15, 25, "enterprise", &currency.ID),
				newSubscriptionPlan("Pro Plan", "A professional plan perfect for growing cooperatives with AI features.", 200, 10, 20, "pro", &currency.ID),
				newSubscriptionPlan("Growth Plan", "A balanced plan for mid-sized co-ops ready to scale with flexible structures.", 100, 8, 18, "growth", &currency.ID),
				newSubscriptionPlan("Starter Plan", "An affordable plan for small organizations just getting started.", 50, 5, 15, "starter", &currency.ID),
				newSubscriptionPlan("Free Plan", "A basic trial plan with essential features to get you started.", 0, 0, 0, "free", &currency.ID),
			}

		case "KWT": // Kuwait
			subscriptions = []*SubscriptionPlan{
				newSubscriptionPlan("Enterprise Plan", "Enterprise-level plan with unlimited features, AI/ML capabilities, and priority support.", 120, 15, 25, "enterprise", &currency.ID),
				newSubscriptionPlan("Pro Plan", "Professional plan for growing cooperatives with AI features.", 60, 10, 20, "pro", &currency.ID),
				newSubscriptionPlan("Growth Plan", "Balanced plan for mid-sized organizations ready to scale with flexible structures.", 30, 8, 18, "growth", &currency.ID),
				newSubscriptionPlan("Starter Plan", "Affordable starter plan for small organizations just getting started.", 15, 5, 15, "starter", &currency.ID),
				newSubscriptionPlan("Free Plan", "Basic free trial plan with essential features to get you started.", 0, 0, 0, "free", &currency.ID),
			}
		case "QAT": // Qatar
			subscriptions = []*SubscriptionPlan{
				newSubscriptionPlan("Enterprise Plan", "Enterprise-level plan with unlimited features, AI/ML capabilities, and priority support.", 1500, 15, 25, "enterprise", &currency.ID),
				newSubscriptionPlan("Pro Plan", "Professional plan for growing cooperatives with AI features.", 800, 10, 20, "pro", &currency.ID),
				newSubscriptionPlan("Growth Plan", "Balanced plan for mid-sized organizations ready to scale with flexible structures.", 400, 7, 17, "growth", &currency.ID),
				newSubscriptionPlan("Starter Plan", "Affordable starter plan for small organizations just getting started.", 200, 5, 15, "starter", &currency.ID),
				newSubscriptionPlan("Free Plan", "Basic free trial plan with essential features to get you started.", 0, 0, 0, "free", &currency.ID),
			}
		case "OMN": // Oman
			subscriptions = []*SubscriptionPlan{
				newSubscriptionPlan("Enterprise Plan", "An enterprise-level plan with unlimited features, AI/ML capabilities, and priority support.", 150, 15, 25, "enterprise", &currency.ID),
				newSubscriptionPlan("Pro Plan", "A professional plan perfect for growing cooperatives with AI features.", 75, 10, 20, "pro", &currency.ID),
				newSubscriptionPlan("Growth Plan", "A balanced plan for mid-sized co-ops ready to scale with flexible structures.", 35, 8, 18, "growth", &currency.ID),
				newSubscriptionPlan("Starter Plan", "An affordable plan for small organizations just getting started.", 20, 5, 15, "starter", &currency.ID),
				newSubscriptionPlan("Free Plan", "A basic trial plan with essential features to get you started.", 0, 0, 0, "free", &currency.ID),
			}
		case "BHR": // Bahrain
			subscriptions = []*SubscriptionPlan{
				newSubscriptionPlan("Enterprise Plan", "Enterprise-grade plan with all premium features, AI/ML tools, and top-tier support.", 150, 15, 25, "enterprise", &currency.ID),
				newSubscriptionPlan("Pro Plan", "Perfect for growing cooperatives with advanced AI features.", 75, 10, 20, "pro", &currency.ID),
				newSubscriptionPlan("Growth Plan", "Ideal for scaling organizations with flexibility and analytics.", 35, 8, 18, "growth", &currency.ID),
				newSubscriptionPlan("Starter Plan", "A simple plan for small cooperatives and startups.", 20, 5, 15, "starter", &currency.ID),
				newSubscriptionPlan("Free Plan", "Trial plan to explore core features before upgrading.", 0, 0, 0, "free", &currency.ID),
			}
		case "JOR": // Jordan
			subscriptions = []*SubscriptionPlan{
				newSubscriptionPlan("Enterprise Plan", "الخطة المتقدمة للمؤسسات — تتضمن جميع الميزات المتقدمة ودعمًا أولوياً.", 110, 15, 25, "enterprise", &currency.ID),
				newSubscriptionPlan("Pro Plan", "الخطة الاحترافية — مثالية للمؤسسات المتنامية بميزات الذكاء الاصطناعي.", 55, 10, 20, "pro", &currency.ID),
				newSubscriptionPlan("Growth Plan", "خطة النمو — للشركات المتوسطة التي تسعى للتوسع.", 30, 8, 18, "growth", &currency.ID),
				newSubscriptionPlan("Starter Plan", "خطة البداية — للشركات الصغيرة والمبتدئة.", 15, 5, 15, "starter", &currency.ID),
				newSubscriptionPlan("Free Plan", "خطة مجانية لتجربة الأساسيات قبل الاشتراك الكامل.", 0, 0, 0, "free", &currency.ID),
			}
		case "KAZ": // Kazakhstan
			subscriptions = []*SubscriptionPlan{
				newSubscriptionPlan("Enterprise Plan", "Жоғары деңгейдегі жоспар — барлық мүмкіндіктер мен басым қолдау қамтылған.", 180000, 15, 25, "enterprise", &currency.ID),
				newSubscriptionPlan("Pro Plan", "Кәсіби жоспар — өсіп келе жатқан ұйымдар үшін мінсіз таңдау.", 90000, 10, 20, "pro", &currency.ID),
				newSubscriptionPlan("Growth Plan", "Орта деңгейлі ұйымдарға арналған теңгерімді жоспар.", 45000, 8, 18, "growth", &currency.ID),
				newSubscriptionPlan("Starter Plan", "Кіші ұйымдар үшін қолжетімді жоспар.", 20000, 5, 15, "starter", &currency.ID),
				newSubscriptionPlan("Free Plan", "Негізгі мүмкіндіктерді сынауға арналған тегін жоспар.", 0, 0, 0, "free", &currency.ID),
			}
		default:
			continue
		}

		for _, sub := range subscriptions {
			if err := m.SubscriptionPlanManager().Create(ctx, sub); err != nil {
				return eris.Wrapf(err, "failed to seed subscription %s for  %s", sub.Name, currency.ISO3166Alpha3)
			}
		}
	}

	return nil
}
