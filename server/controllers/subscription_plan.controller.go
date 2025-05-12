package controllers

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"horizon.com/server/horizon"
	"horizon.com/server/server/model"
)

// GET /subscription-plan
func (c *Controller) SubscriptionPlanList(ctx echo.Context) error {
	subscription_plan, err := c.subscriptionPlan.Manager.List()
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return ctx.JSON(http.StatusOK, c.model.SubscriptionPlanModels(subscription_plan))
}

// GET /subscription-plan/:subscription_plan_id
func (c *Controller) SubscriptionPlanGetByID(ctx echo.Context) error {
	id, err := horizon.EngineUUIDParam(ctx, "subscription_plan_id")
	if err != nil {
		return err
	}
	subscription_plan, err := c.subscriptionPlan.Manager.GetByID(*id)
	if err != nil {
		return ctx.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
	}
	return ctx.JSON(http.StatusOK, c.model.SubscriptionPlanModel(subscription_plan))
}

// POST /subscription_plan
func (c *Controller) SubscriptionPlanCreate(ctx echo.Context) error {
	req, err := c.model.SubscriptionPlanValidate(ctx)
	if err != nil {
		return err
	}
	model := &model.SubscriptionPlan{
		Name:                req.Name,
		Description:         req.Description,
		Cost:                req.Cost,
		Timespan:            req.Timespan,
		MaxBranches:         req.MaxBranches,
		MaxEmployees:        req.MaxEmployees,
		MaxMembersPerBranch: req.MaxMembersPerBranch,
		Discount:            req.Discount,
		YearlyDiscount:      req.YearlyDiscount,
		CreatedAt:           time.Now().UTC(),
		UpdatedAt:           time.Now().UTC(),
	}
	if err := c.subscriptionPlan.Manager.Create(model); err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return ctx.JSON(http.StatusCreated, c.model.SubscriptionPlanModel(model))
}

// PUT /subscription-plan/subscription_plan_id
func (c *Controller) SubscriptionPlanUpdate(ctx echo.Context) error {
	id, err := horizon.EngineUUIDParam(ctx, "subscription_plan_id")
	if err != nil {
		return err
	}
	req, err := c.model.SubscriptionPlanValidate(ctx)
	if err != nil {
		return err
	}
	model := &model.SubscriptionPlan{
		Name:                req.Name,
		Description:         req.Description,
		Cost:                req.Cost,
		Timespan:            req.Timespan,
		MaxBranches:         req.MaxBranches,
		MaxEmployees:        req.MaxEmployees,
		MaxMembersPerBranch: req.MaxMembersPerBranch,
		Discount:            req.Discount,
		YearlyDiscount:      req.YearlyDiscount,
		UpdatedAt:           time.Now().UTC(),
	}
	if err := c.subscriptionPlan.Manager.UpdateByID(*id, model); err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return ctx.JSON(http.StatusCreated, c.model.SubscriptionPlanModel(model))
}

// DELETE /subscription-plan/subscription_plan_id
func (c *Controller) SubscriptionPlanDelete(ctx echo.Context) error {
	id, err := horizon.EngineUUIDParam(ctx, "subscription_plan_id")
	if err != nil {
		return err
	}
	if err := c.subscriptionPlan.Manager.DeleteByID(*id); err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return ctx.NoContent(http.StatusNoContent)
}
