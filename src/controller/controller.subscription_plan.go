package controller

import (
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/lands-horizon/horizon-server/services/horizon"
	"github.com/lands-horizon/horizon-server/src/model"
)

func (c *Controller) SubscriptionPlanController() {
	req := c.provider.Service.Request

	req.RegisterRoute(horizon.Route{
		Route:    "/subscription-plan",
		Method:   "GET",
		Response: "TSubscriptionPlan[]",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		categories, err := c.model.SubscriptionPlanManager.ListRaw(context)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})

		}
		return ctx.JSON(http.StatusOK, categories)
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/subscription-plan/:subscription_plan_id",
		Method:   "GET",
		Response: "TSubscriptionPlan",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		subscription_planID, err := horizon.EngineUUIDParam(ctx, "subscription_plan_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid subscription_plan ID")
		}

		subscription_plan, err := c.model.SubscriptionPlanManager.GetByIDRaw(context, *subscription_planID)
		if err != nil {
			return c.NotFound(ctx, "SubscriptionPlan")
		}

		return ctx.JSON(http.StatusOK, subscription_plan)
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/subscription-plan",
		Method:   "POST",
		Request:  "TSubscriptionPlan",
		Response: "TSubscriptionPlan",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := c.model.SubscriptionPlanManager.Validate(ctx)
		if err != nil {
			return c.BadRequest(ctx, err.Error())
		}

		subscription_plan := &model.SubscriptionPlan{
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

		if err := c.model.SubscriptionPlanManager.Create(context, subscription_plan); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})

		}

		return ctx.JSON(http.StatusOK, c.model.SubscriptionPlanManager.ToModel(subscription_plan))
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/subscription-plan/:subscription_plan_id",
		Method:   "PUT",
		Request:  "TSubscriptionPlan",
		Response: "TSubscriptionPlan",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		subscription_planID, err := horizon.EngineUUIDParam(ctx, "subscription_plan_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid subscription_plan ID")
		}

		req, err := c.model.SubscriptionPlanManager.Validate(ctx)
		if err != nil {
			return c.BadRequest(ctx, err.Error())
		}

		subscription_plan, err := c.model.SubscriptionPlanManager.GetByID(context, *subscription_planID)
		if err != nil {
			return c.NotFound(ctx, "SubscriptionPlan")
		}

		subscription_plan.Name = req.Name
		subscription_plan.Description = req.Description
		subscription_plan.Cost = req.Cost
		subscription_plan.Timespan = req.Timespan
		subscription_plan.MaxBranches = req.MaxBranches
		subscription_plan.MaxEmployees = req.MaxEmployees
		subscription_plan.MaxMembersPerBranch = req.MaxMembersPerBranch
		subscription_plan.Discount = req.Discount
		subscription_plan.YearlyDiscount = req.YearlyDiscount
		subscription_plan.UpdatedAt = time.Now().UTC()

		if err := c.model.SubscriptionPlanManager.UpdateFields(context, subscription_plan.ID, subscription_plan); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})

		}

		return ctx.JSON(http.StatusOK, c.model.SubscriptionPlanManager.ToModel(subscription_plan))
	})

	req.RegisterRoute(horizon.Route{
		Route:  "/subscription-plan/:subscription_plan_id",
		Method: "DELETE",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		subscription_planID, err := horizon.EngineUUIDParam(ctx, "subscription_plan_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid subscription_plan ID")
		}

		if err := c.model.SubscriptionPlanManager.DeleteByID(context, *subscription_planID); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})

		}

		return ctx.NoContent(http.StatusNoContent)
	})

	req.RegisterRoute(horizon.Route{
		Route:   "/subscription-plan/bulk-delete",
		Method:  "DELETE",
		Request: "string[]",
		Note:    "Delete multiple subscription_plan records",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var reqBody struct {
			IDs []string `json:"ids"`
		}

		if err := ctx.Bind(&reqBody); err != nil {
			return c.BadRequest(ctx, "Invalid request body")
		}

		if len(reqBody.IDs) == 0 {
			return c.BadRequest(ctx, "No IDs provided")
		}

		tx := c.provider.Service.Database.Client().Begin()
		if tx.Error != nil {
			tx.Rollback()
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": tx.Error.Error()})
		}

		for _, rawID := range reqBody.IDs {
			subscription_planID, err := uuid.Parse(rawID)
			if err != nil {
				tx.Rollback()
				return c.BadRequest(ctx, fmt.Sprintf("Invalid UUID: %s", rawID))
			}

			if _, err := c.model.SubscriptionPlanManager.GetByID(context, subscription_planID); err != nil {
				tx.Rollback()
				return c.NotFound(ctx, fmt.Sprintf("SubscriptionPlan with ID %s", rawID))
			}

			if err := c.model.SubscriptionPlanManager.DeleteByIDWithTx(context, tx, subscription_planID); err != nil {
				tx.Rollback()
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})

			}
		}

		if err := tx.Commit().Error; err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})

		}

		return ctx.NoContent(http.StatusNoContent)
	})
}
