package model

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
	"horizon.com/server/horizon"
)

type (
	SubscriptionPlan struct {
		ID        uuid.UUID      `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
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

		Organizations []*Organization `gorm:"foreignKey:SubscriptionPlanID" json:"organizations,omitempty"`
	}

	SubscriptionPlanRequest struct {
		Name                string  `json:"name" validate:"required,min=1,max=255"`
		Description         string  `json:"description" validate:"required"`
		Cost                float64 `json:"cost" validate:"required,gt=0"`
		Timespan            int     `json:"timespan" validate:"required,gt=0"`
		MaxBranches         int     `json:"max_branches" validate:"required,gte=0"`
		MaxEmployees        int     `json:"max_employees" validate:"required,gte=0"`
		MaxMembersPerBranch int     `json:"max_members_per_branch" validate:"required,gte=0"`
		Discount            float64 `json:"discount" validate:"gte=0"`
		YearlyDiscount      float64 `json:"yearly_discount" validate:"gte=0"`

		Organizations []*Organization `json:"organizations,omitempty"`
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

	SubscriptionPlanCollection struct {
		Manager CollectionManager[SubscriptionPlan]
	}
)

func (m *Model) SubscriptionPlanValidate(ctx echo.Context) (*SubscriptionPlanRequest, error) {
	return Validate[SubscriptionPlanRequest](ctx, m.validator)
}

func (m *Model) SubscriptionPlanModel(data *SubscriptionPlan) *SubscriptionPlanResponse {
	return ToModel(data, func(data *SubscriptionPlan) *SubscriptionPlanResponse {
		return &SubscriptionPlanResponse{
			ID:                  data.ID,
			Name:                data.Name,
			Description:         data.Description,
			Cost:                data.Cost,
			Timespan:            data.Timespan,
			MaxBranches:         data.MaxBranches,
			MaxEmployees:        data.MaxEmployees,
			MaxMembersPerBranch: data.MaxMembersPerBranch,
			Discount:            data.Discount,
			YearlyDiscount:      data.YearlyDiscount,
			CreatedAt:           data.CreatedAt.Format(time.RFC3339),
			UpdatedAt:           data.UpdatedAt.Format(time.RFC3339),
		}
	})
}

func (m *Model) SubscriptionPlanModels(data []*SubscriptionPlan) []*SubscriptionPlanResponse {
	return ToModels(data, m.SubscriptionPlanModel)
}

func NewSubscriptionPlanCollection(
	broadcast *horizon.HorizonBroadcast,
	database *horizon.HorizonDatabase,
	model *Model,
) (*SubscriptionPlanCollection, error) {
	manager := NewcollectionManager(
		database,
		broadcast,
		func(data *SubscriptionPlan) ([]string, any) {
			return []string{
				"subscription_plan.create",
				fmt.Sprintf("subscription_plan.create.%s", data.ID),
			}, model.SubscriptionPlanModel(data)
		},
		func(data *SubscriptionPlan) ([]string, any) {
			return []string{
				"subscription_plan.update",
				fmt.Sprintf("subscription_plan.update.%s", data.ID),
			}, model.SubscriptionPlanModel(data)
		},
		func(data *SubscriptionPlan) ([]string, any) {
			return []string{
				"subscription_plan.delete",
				fmt.Sprintf("subscription_plan.delete.%s", data.ID),
			}, model.SubscriptionPlanModel(data)
		},
		[]string{},
	)
	return &SubscriptionPlanCollection{
		Manager: manager,
	}, nil
}
