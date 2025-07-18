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

	// Get all subscription plans
	req.RegisterRoute(horizon.Route{
		Route:    "/subscription-plan",
		Method:   "GET",
		Response: "TSubscriptionPlan[]",
		Note:     "Returns all subscription plans.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		categories, err := c.model.SubscriptionPlanManager.ListRaw(context)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve subscription plans: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, categories)
	})

	// Get a subscription plan by its ID
	req.RegisterRoute(horizon.Route{
		Route:    "/subscription-plan/:subscription_plan_id",
		Method:   "GET",
		Response: "TSubscriptionPlan",
		Note:     "Returns a specific subscription plan by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		subscriptionPlanID, err := horizon.EngineUUIDParam(ctx, "subscription_plan_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid subscription_plan_id: " + err.Error()})
		}

		subscriptionPlan, err := c.model.SubscriptionPlanManager.GetByIDRaw(context, *subscriptionPlanID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "SubscriptionPlan not found: " + err.Error()})
		}

		return ctx.JSON(http.StatusOK, subscriptionPlan)
	})

	// Create a new subscription plan
	req.RegisterRoute(horizon.Route{
		Route:    "/subscription-plan",
		Method:   "POST",
		Request:  "TSubscriptionPlan",
		Response: "TSubscriptionPlan",
		Note:     "Creates a new subscription plan.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := c.model.SubscriptionPlanManager.Validate(ctx)
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}

		subscriptionPlan := &model.SubscriptionPlan{
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

		if err := c.model.SubscriptionPlanManager.Create(context, subscriptionPlan); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create subscription plan: " + err.Error()})
		}

		return ctx.JSON(http.StatusOK, c.model.SubscriptionPlanManager.ToModel(subscriptionPlan))
	})

	// Update a subscription plan by its ID
	req.RegisterRoute(horizon.Route{
		Route:    "/subscription-plan/:subscription_plan_id",
		Method:   "PUT",
		Request:  "TSubscriptionPlan",
		Response: "TSubscriptionPlan",
		Note:     "Updates an existing subscription plan by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		subscriptionPlanID, err := horizon.EngineUUIDParam(ctx, "subscription_plan_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid subscription_plan_id: " + err.Error()})
		}

		req, err := c.model.SubscriptionPlanManager.Validate(ctx)
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}

		subscriptionPlan, err := c.model.SubscriptionPlanManager.GetByID(context, *subscriptionPlanID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "SubscriptionPlan not found: " + err.Error()})
		}

		subscriptionPlan.Name = req.Name
		subscriptionPlan.Description = req.Description
		subscriptionPlan.Cost = req.Cost
		subscriptionPlan.Timespan = req.Timespan
		subscriptionPlan.MaxBranches = req.MaxBranches
		subscriptionPlan.MaxEmployees = req.MaxEmployees
		subscriptionPlan.MaxMembersPerBranch = req.MaxMembersPerBranch
		subscriptionPlan.Discount = req.Discount
		subscriptionPlan.YearlyDiscount = req.YearlyDiscount
		subscriptionPlan.UpdatedAt = time.Now().UTC()

		if err := c.model.SubscriptionPlanManager.UpdateFields(context, subscriptionPlan.ID, subscriptionPlan); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update subscription plan: " + err.Error()})
		}

		return ctx.JSON(http.StatusOK, c.model.SubscriptionPlanManager.ToModel(subscriptionPlan))
	})

	// Delete a subscription plan by its ID
	req.RegisterRoute(horizon.Route{
		Route:  "/subscription-plan/:subscription_plan_id",
		Method: "DELETE",
		Note:   "Deletes a subscription plan by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		subscriptionPlanID, err := horizon.EngineUUIDParam(ctx, "subscription_plan_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid subscription_plan_id: " + err.Error()})
		}

		if err := c.model.SubscriptionPlanManager.DeleteByID(context, *subscriptionPlanID); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete subscription plan: " + err.Error()})
		}

		return ctx.NoContent(http.StatusNoContent)
	})

	// Bulk delete subscription plans by IDs
	req.RegisterRoute(horizon.Route{
		Route:   "/subscription-plan/bulk-delete",
		Method:  "DELETE",
		Request: "string[]",
		Note:    "Deletes multiple subscription plan records.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var reqBody struct {
			IDs []string `json:"ids"`
		}

		if err := ctx.Bind(&reqBody); err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body: " + err.Error()})
		}

		if len(reqBody.IDs) == 0 {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "No IDs provided for deletion."})
		}

		tx := c.provider.Service.Database.Client().Begin()
		if tx.Error != nil {
			tx.Rollback()
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to begin transaction: " + tx.Error.Error()})
		}

		for _, rawID := range reqBody.IDs {
			subscriptionPlanID, err := uuid.Parse(rawID)
			if err != nil {
				tx.Rollback()
				return ctx.JSON(http.StatusBadRequest, map[string]string{"error": fmt.Sprintf("Invalid UUID: %s - %v", rawID, err)})
			}

			if _, err := c.model.SubscriptionPlanManager.GetByID(context, subscriptionPlanID); err != nil {
				tx.Rollback()
				return ctx.JSON(http.StatusNotFound, map[string]string{"error": fmt.Sprintf("SubscriptionPlan with ID %s not found: %v", rawID, err)})
			}

			if err := c.model.SubscriptionPlanManager.DeleteByIDWithTx(context, tx, subscriptionPlanID); err != nil {
				tx.Rollback()
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": fmt.Sprintf("Failed to delete subscription plan with ID %s: %v", rawID, err)})
			}
		}

		if err := tx.Commit().Error; err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to commit transaction: " + err.Error()})
		}

		return ctx.NoContent(http.StatusNoContent)
	})
}
