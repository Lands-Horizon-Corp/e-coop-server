package model_core

import (
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
		IsRecommended       bool    `gorm:"not null;default:false"` // <-- Added field

		Organizations []*Organization `gorm:"foreignKey:SubscriptionPlanID" json:"organizations,omitempty"` // organization
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

		YearlyDiscount float64         `json:"yearly_discount" validate:"gte=0"`
		Organizations  []*Organization `json:"organizations,omitempty"`
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

		MonthlyPrice           float64 `json:"monthly_price"`
		YearlyPrice            float64 `json:"yearly_price"`
		DiscountedMonthlyPrice float64 `json:"discounted_monthly_price"`
		DiscountedYearlyPrice  float64 `json:"discounted_yearly_price"`

		CreatedAt string `json:"created_at"`
		UpdatedAt string `json:"updated_at"`
	}
)

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
				ID:                     sp.ID,
				Name:                   sp.Name,
				Description:            sp.Description,
				Cost:                   sp.Cost,
				Timespan:               sp.Timespan,
				MaxBranches:            sp.MaxBranches,
				MaxEmployees:           sp.MaxEmployees,
				MaxMembersPerBranch:    sp.MaxMembersPerBranch,
				Discount:               sp.Discount,
				YearlyDiscount:         sp.YearlyDiscount,
				IsRecommended:          sp.IsRecommended,
				MonthlyPrice:           monthlyPrice,
				YearlyPrice:            yearlyPrice,
				DiscountedMonthlyPrice: discountedMonthlyPrice,
				DiscountedYearlyPrice:  discountedYearlyPrice,
				CreatedAt:              sp.CreatedAt.Format(time.RFC3339),
				UpdatedAt:              sp.UpdatedAt.Format(time.RFC3339),
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
