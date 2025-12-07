package v1

import (
	"net/http"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/server/event"
	"github.com/Lands-Horizon-Corp/e-coop-server/server/model/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/services/handlers"
	"github.com/labstack/echo/v4"
)

func (c *Controller) subscriptionPlanController() {
	req := c.provider.Service.Request

	// Get all subscription plans
	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/subscription-plan",
		Method:       "GET",
		ResponseType: core.SubscriptionPlanResponse{},
		Note:         "Returns all subscription plans.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		categories, err := c.core.SubscriptionPlanManager.List(context)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve subscription plans: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.core.SubscriptionPlanManager.ToModels(categories))
	})

	// Get a subscription plan by its ID
	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/subscription-plan/:subscription_plan_id",
		Method:       "GET",
		ResponseType: core.SubscriptionPlanResponse{},
		Note:         "Returns a specific subscription plan by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		subscriptionPlanID, err := handlers.EngineUUIDParam(ctx, "subscription_plan_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid subscription_plan_id: " + err.Error()})
		}

		subscriptionPlan, err := c.core.SubscriptionPlanManager.GetByIDRaw(context, *subscriptionPlanID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "SubscriptionPlan not found: " + err.Error()})
		}

		return ctx.JSON(http.StatusOK, subscriptionPlan)
	})

	// Create a new subscription plan
	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/subscription-plan",
		Method:       "POST",
		ResponseType: core.SubscriptionPlanResponse{},
		RequestType:  core.SubscriptionPlanRequest{},
		Note:         "Creates a new subscription plan.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := c.core.SubscriptionPlanManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create subscription plan failed: validation error: " + err.Error(),
				Module:      "SubscriptionPlan",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}

		subscriptionPlan := &core.SubscriptionPlan{
			Name:                     req.Name,
			Description:              req.Description,
			Cost:                     req.Cost,
			Timespan:                 req.Timespan,
			MaxBranches:              req.MaxBranches,
			MaxEmployees:             req.MaxEmployees,
			MaxMembersPerBranch:      req.MaxMembersPerBranch,
			Discount:                 req.Discount,
			YearlyDiscount:           req.YearlyDiscount,
			IsRecommended:            req.IsRecommended,
			HasAPIAccess:             req.HasAPIAccess,
			HasFlexibleOrgStructures: req.HasFlexibleOrgStructures,
			HasAIEnabled:             req.HasAIEnabled,
			HasMachineLearning:       req.HasMachineLearning,
			MaxAPICallsPerMonth:      req.MaxAPICallsPerMonth,
			CurrencyID:               req.CurrencyID,
			CreatedAt:                time.Now().UTC(),
			UpdatedAt:                time.Now().UTC(),
		}

		if err := c.core.SubscriptionPlanManager.Create(context, subscriptionPlan); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create subscription plan failed: create error: " + err.Error(),
				Module:      "SubscriptionPlan",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create subscription plan: " + err.Error()})
		}

		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "create-success",
			Description: "Created subscription plan: " + subscriptionPlan.Name,
			Module:      "SubscriptionPlan",
		})

		return ctx.JSON(http.StatusOK, c.core.SubscriptionPlanManager.ToModel(subscriptionPlan))
	})

	// Update a subscription plan by its ID
	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/subscription-plan/:subscription_plan_id",
		Method:       "PUT",
		ResponseType: core.SubscriptionPlanResponse{},
		RequestType:  core.SubscriptionPlanRequest{},
		Note:         "Updates an existing subscription plan by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		subscriptionPlanID, err := handlers.EngineUUIDParam(ctx, "subscription_plan_id")
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update subscription plan failed: invalid subscription_plan_id: " + err.Error(),
				Module:      "SubscriptionPlan",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid subscription_plan_id: " + err.Error()})
		}

		req, err := c.core.SubscriptionPlanManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update subscription plan failed: validation error: " + err.Error(),
				Module:      "SubscriptionPlan",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}

		subscriptionPlan, err := c.core.SubscriptionPlanManager.GetByID(context, *subscriptionPlanID)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update subscription plan failed: not found: " + err.Error(),
				Module:      "SubscriptionPlan",
			})
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
		subscriptionPlan.IsRecommended = req.IsRecommended
		subscriptionPlan.HasAPIAccess = req.HasAPIAccess
		subscriptionPlan.HasFlexibleOrgStructures = req.HasFlexibleOrgStructures
		subscriptionPlan.HasAIEnabled = req.HasAIEnabled
		subscriptionPlan.HasMachineLearning = req.HasMachineLearning
		subscriptionPlan.MaxAPICallsPerMonth = req.MaxAPICallsPerMonth
		subscriptionPlan.CurrencyID = req.CurrencyID
		subscriptionPlan.UpdatedAt = time.Now().UTC()

		if err := c.core.SubscriptionPlanManager.UpdateByID(context, subscriptionPlan.ID, subscriptionPlan); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update subscription plan failed: update error: " + err.Error(),
				Module:      "SubscriptionPlan",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update subscription plan: " + err.Error()})
		}

		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Updated subscription plan: " + subscriptionPlan.Name,
			Module:      "SubscriptionPlan",
		})

		return ctx.JSON(http.StatusOK, c.core.SubscriptionPlanManager.ToModel(subscriptionPlan))
	})

	// Delete a subscription plan by its ID
	req.RegisterWebRoute(handlers.Route{
		Route:  "/api/v1/subscription-plan/:subscription_plan_id",
		Method: "DELETE",
		Note:   "Deletes a subscription plan by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		subscriptionPlanID, err := handlers.EngineUUIDParam(ctx, "subscription_plan_id")
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete subscription plan failed: invalid subscription_plan_id: " + err.Error(),
				Module:      "SubscriptionPlan",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid subscription_plan_id: " + err.Error()})
		}

		subscriptionPlan, err := c.core.SubscriptionPlanManager.GetByID(context, *subscriptionPlanID)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete subscription plan failed: not found: " + err.Error(),
				Module:      "SubscriptionPlan",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "SubscriptionPlan not found: " + err.Error()})
		}

		if err := c.core.SubscriptionPlanManager.Delete(context, *subscriptionPlanID); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete subscription plan failed: delete error: " + err.Error(),
				Module:      "SubscriptionPlan",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete subscription plan: " + err.Error()})
		}

		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "delete-success",
			Description: "Deleted subscription plan: " + subscriptionPlan.Name,
			Module:      "SubscriptionPlan",
		})

		return ctx.NoContent(http.StatusNoContent)
	})

	// Simplified bulk-delete handler for subscription plans (mirrors feedback/holiday pattern)
	req.RegisterWebRoute(handlers.Route{
		Route:       "/api/v1/subscription-plan/bulk-delete",
		Method:      "DELETE",
		Note:        "Deletes multiple subscription plan records.",
		RequestType: core.IDSRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var reqBody core.IDSRequest

		if err := ctx.Bind(&reqBody); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Subscription plan bulk delete failed (/subscription-plan/bulk-delete) | invalid request body: " + err.Error(),
				Module:      "SubscriptionPlan",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body: " + err.Error()})
		}

		if len(reqBody.IDs) == 0 {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Subscription plan bulk delete failed (/subscription-plan/bulk-delete) | no IDs provided",
				Module:      "SubscriptionPlan",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "No IDs provided for bulk delete"})
		}

		// Delegate deletion to the manager. Manager should handle transactions, validations and DeletedBy bookkeeping.
		if err := c.core.SubscriptionPlanManager.BulkDelete(context, reqBody.IDs); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Subscription plan bulk delete failed (/subscription-plan/bulk-delete) | error: " + err.Error(),
				Module:      "SubscriptionPlan",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to bulk delete subscription plans: " + err.Error()})
		}

		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "bulk-delete-success",
			Description: "Bulk deleted subscription plans (/subscription-plan/bulk-delete)",
			Module:      "SubscriptionPlan",
		})

		return ctx.NoContent(http.StatusNoContent)
	})

	// GET /api/v1/subscription-plan/currency/:currency_id
	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/subscription-plan/currency/:currency_id",
		Method:       "GET",
		ResponseType: core.SubscriptionPlanResponse{},
		Note:         "Returns all subscription plans for a specific currency.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		currencyID, err := handlers.EngineUUIDParam(ctx, "currency_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid currency_id: " + err.Error()})
		}
		subscriptionPlans, err := c.core.SubscriptionPlanManager.FindRaw(context, &core.SubscriptionPlan{
			CurrencyID: currencyID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve subscription plans: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, subscriptionPlans)
	})

}
