package types

import (
	"time"

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
