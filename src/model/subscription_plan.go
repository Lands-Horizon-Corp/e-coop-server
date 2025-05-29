package model

import (
	"time"

	"github.com/google/uuid"
	horizon_services "github.com/lands-horizon/horizon-server/services"
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
		Timespan            int     `gorm:"not null"`
		MaxBranches         int     `gorm:"not null"`
		MaxEmployees        int     `gorm:"not null"`
		MaxMembersPerBranch int     `gorm:"not null"`
		Discount            float64 `gorm:"type:numeric(5,2);default:0"`
		YearlyDiscount      float64 `gorm:"type:numeric(5,2);default:0"`

		Organizations []*Organization `gorm:"foreignKey:SubscriptionPlanID" json:"organizations,omitempty"` // organization
	}

	SubscriptionPlanRequest struct {
		ID *uuid.UUID `json:"id,omitempty"`

		Name                string          `json:"name" validate:"required,min=1,max=255"`
		Description         string          `json:"description" validate:"required"`
		Cost                float64         `json:"cost" validate:"required,gt=0"`
		Timespan            int             `json:"timespan" validate:"required,gt=0"`
		MaxBranches         int             `json:"max_branches" validate:"required,gte=0"`
		MaxEmployees        int             `json:"max_employees" validate:"required,gte=0"`
		MaxMembersPerBranch int             `json:"max_members_per_branch" validate:"required,gte=0"`
		Discount            float64         `json:"discount" validate:"gte=0"`
		YearlyDiscount      float64         `json:"yearly_discount" validate:"gte=0"`
		Organizations       []*Organization `json:"organizations,omitempty"`
	}

	SubscriptionPlanResponse struct {
		ID                  uuid.UUID `json:"id"`
		Name                string    `json:"name"`
		Description         string    `json:"description"`
		Cost                float64   `json:"cost"`
		Timespan            int       `json:"timespan"`
		MaxBranches         int       `json:"max_branches"`
		MaxEmployees        int       `json:"max_employees"`
		MaxMembersPerBranch int       `json:"max_members_per_branch"`
		Discount            float64   `json:"discount"`
		YearlyDiscount      float64   `json:"yearly_discount"`
		CreatedAt           string    `json:"created_at"`
		UpdatedAt           string    `json:"updated_at"`
	}
)

func (m *Model) SubscriptionPlan() {
	m.Migration = append(m.Migration, &SubscriptionPlan{})
	m.SubscriptionPlanManager = horizon_services.NewRepository(horizon_services.RepositoryParams[SubscriptionPlan, SubscriptionPlanResponse, SubscriptionPlanRequest]{
		Preloads: nil,
		Service:  m.provider.Service,
		Resource: func(sp *SubscriptionPlan) *SubscriptionPlanResponse {
			if sp == nil {
				return nil
			}
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
				CreatedAt:           sp.CreatedAt.Format(time.RFC3339),
				UpdatedAt:           sp.UpdatedAt.Format(time.RFC3339),
			}
		},
		Created: func(sp *SubscriptionPlan) []string {
			return []string{
				"subscription_plan.create",
				"subscription_plan.create." + sp.ID.String(),
			}
		},
		Updated: func(sp *SubscriptionPlan) []string {
			return []string{
				"subscription_plan.update",
				"subscription_plan.update." + sp.ID.String(),
			}
		},
		Deleted: func(sp *SubscriptionPlan) []string {
			return []string{
				"subscription_plan.delete",
				"subscription_plan.delete." + sp.ID.String(),
			}
		},
	})
}
