package collection

import (
	"net/http"
	"time"

	"github.com/go-playground/validator"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

type (
	SubscriptionPlan struct {
		ID        uuid.UUID      `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
		CreatedAt time.Time      `gorm:"not null;default:now()"`
		UpdatedAt time.Time      `gorm:"not null;default:now()"`
		DeletedAt gorm.DeletedAt `gorm:"index"`

		Name        string `gorm:"type:varchar(255);not null"`
		Description string `gorm:"type:text;not null"`

		Cost                float64 `gorm:"type:numeric(10,2);not null"`
		Timespan            int     `gorm:"not null"`
		MaxBranches         int     `gorm:"not null"`
		MaxEmployees        int     `gorm:"not null"`
		MaxMembersPerBranch int     `gorm:"not null"`

		Discount       float64 `gorm:"type:numeric(5,2);default:0"`
		YearlyDiscount float64 `gorm:"type:numeric(5,2);default:0"`
	}

	SubscriptionPlanRequest struct {
		Name        string `json:"name" validate:"required,min=1,max=255"`
		Description string `json:"description" validate:"required"`

		Cost                float64 `json:"cost" validate:"required,gt=0"`
		Timespan            int     `json:"timespan" validate:"required,gt=0"`
		MaxBranches         int     `json:"max_branches" validate:"required,gte=0"`
		MaxEmployees        int     `json:"max_employees" validate:"required,gte=0"`
		MaxMembersPerBranch int     `json:"max_members_per_branch" validate:"required,gte=0"`

		Discount       float64 `json:"discount" validate:"gte=0"`
		YearlyDiscount float64 `json:"yearly_discount" validate:"gte=0"`
	}

	SubscriptionPlanResponse struct {
		ID          uuid.UUID `json:"id"`
		Name        string    `json:"name"`
		Description string    `json:"description"`

		Cost                float64 `json:"cost"`
		Timespan            int     `json:"timespan"`
		MaxBranches         int     `json:"max_branches"`
		MaxEmployees        int     `json:"max_employees"`
		MaxMembersPerBranch int     `json:"max_members_per_branch"`

		Discount       float64 `json:"discount"`
		YearlyDiscount float64 `json:"yearly_discount"`

		CreatedAt string `json:"created_at"`
		UpdatedAt string `json:"updated_at"`
	}

	SubscriptionPlanCollection struct {
		validator *validator.Validate
	}
)

func NewSubscriptionPlanCollection() (*SubscriptionPlanCollection, error) {
	return &SubscriptionPlanCollection{
		validator: validator.New(),
	}, nil
}

func (spc *SubscriptionPlanCollection) ValidateCreate(c echo.Context) (*SubscriptionPlanRequest, error) {
	req := new(SubscriptionPlanRequest)
	if err := c.Bind(req); err != nil {
		return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := spc.validator.Struct(req); err != nil {
		return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return req, nil
}

func (spc *SubscriptionPlanCollection) ToModel(plan *SubscriptionPlan) *SubscriptionPlanResponse {
	if plan == nil {
		return nil
	}
	return &SubscriptionPlanResponse{
		ID:                  plan.ID,
		Name:                plan.Name,
		Description:         plan.Description,
		Cost:                plan.Cost,
		Timespan:            plan.Timespan,
		MaxBranches:         plan.MaxBranches,
		MaxEmployees:        plan.MaxEmployees,
		MaxMembersPerBranch: plan.MaxMembersPerBranch,
		Discount:            plan.Discount,
		YearlyDiscount:      plan.YearlyDiscount,
		CreatedAt:           plan.CreatedAt.Format(time.RFC3339),
		UpdatedAt:           plan.UpdatedAt.Format(time.RFC3339),
	}
}

// ToModels maps multiple DB SubscriptionPlans to responses
func (spc *SubscriptionPlanCollection) ToModels(data []*SubscriptionPlan) []*SubscriptionPlanResponse {
	if data == nil {
		return []*SubscriptionPlanResponse{}
	}
	var out []*SubscriptionPlanResponse
	for _, p := range data {
		if m := spc.ToModel(p); m != nil {
			out = append(out, m)
		}
	}
	return out
}
