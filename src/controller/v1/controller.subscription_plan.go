package controller_v1

import (
	"fmt"
	"net/http"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/services/handlers"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/event"
	modelcore "github.com/Lands-Horizon-Corp/e-coop-server/src/model/modelcore"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func (c *Controller) SubscriptionPlanController() {
	req := c.provider.Service.Request

	// Get all subscription plans
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/subscription-plan",
		Method:       "GET",
		ResponseType: modelcore.SubscriptionPlanResponse{},
		Note:         "Returns all subscription plans.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		categories, err := c.modelcore.SubscriptionPlanManager.List(context)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve subscription plans: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.modelcore.SubscriptionPlanManager.Filtered(context, ctx, categories))
	})

	// Get a subscription plan by its ID
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/subscription-plan/:subscription_plan_id",
		Method:       "GET",
		ResponseType: modelcore.SubscriptionPlanResponse{},
		Note:         "Returns a specific subscription plan by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		subscriptionPlanID, err := handlers.EngineUUIDParam(ctx, "subscription_plan_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid subscription_plan_id: " + err.Error()})
		}

		subscriptionPlan, err := c.modelcore.SubscriptionPlanManager.GetByIDRaw(context, *subscriptionPlanID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "SubscriptionPlan not found: " + err.Error()})
		}

		return ctx.JSON(http.StatusOK, subscriptionPlan)
	})

	// Create a new subscription plan
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/subscription-plan",
		Method:       "POST",
		ResponseType: modelcore.SubscriptionPlanResponse{},
		RequestType:  modelcore.SubscriptionPlanRequest{},
		Note:         "Creates a new subscription plan.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := c.modelcore.SubscriptionPlanManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create subscription plan failed: validation error: " + err.Error(),
				Module:      "SubscriptionPlan",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}

		subscriptionPlan := &modelcore.SubscriptionPlan{
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

		if err := c.modelcore.SubscriptionPlanManager.Create(context, subscriptionPlan); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create subscription plan failed: create error: " + err.Error(),
				Module:      "SubscriptionPlan",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create subscription plan: " + err.Error()})
		}

		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "create-success",
			Description: "Created subscription plan: " + subscriptionPlan.Name,
			Module:      "SubscriptionPlan",
		})

		return ctx.JSON(http.StatusOK, c.modelcore.SubscriptionPlanManager.ToModel(subscriptionPlan))
	})

	// Update a subscription plan by its ID
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/subscription-plan/:subscription_plan_id",
		Method:       "PUT",
		ResponseType: modelcore.SubscriptionPlanResponse{},
		RequestType:  modelcore.SubscriptionPlanRequest{},
		Note:         "Updates an existing subscription plan by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		subscriptionPlanID, err := handlers.EngineUUIDParam(ctx, "subscription_plan_id")
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update subscription plan failed: invalid subscription_plan_id: " + err.Error(),
				Module:      "SubscriptionPlan",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid subscription_plan_id: " + err.Error()})
		}

		req, err := c.modelcore.SubscriptionPlanManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update subscription plan failed: validation error: " + err.Error(),
				Module:      "SubscriptionPlan",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}

		subscriptionPlan, err := c.modelcore.SubscriptionPlanManager.GetByID(context, *subscriptionPlanID)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
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

		if err := c.modelcore.SubscriptionPlanManager.UpdateFields(context, subscriptionPlan.ID, subscriptionPlan); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update subscription plan failed: update error: " + err.Error(),
				Module:      "SubscriptionPlan",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update subscription plan: " + err.Error()})
		}

		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Updated subscription plan: " + subscriptionPlan.Name,
			Module:      "SubscriptionPlan",
		})

		return ctx.JSON(http.StatusOK, c.modelcore.SubscriptionPlanManager.ToModel(subscriptionPlan))
	})

	// Delete a subscription plan by its ID
	req.RegisterRoute(handlers.Route{
		Route:  "/api/v1/subscription-plan/:subscription_plan_id",
		Method: "DELETE",
		Note:   "Deletes a subscription plan by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		subscriptionPlanID, err := handlers.EngineUUIDParam(ctx, "subscription_plan_id")
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete subscription plan failed: invalid subscription_plan_id: " + err.Error(),
				Module:      "SubscriptionPlan",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid subscription_plan_id: " + err.Error()})
		}

		subscriptionPlan, err := c.modelcore.SubscriptionPlanManager.GetByID(context, *subscriptionPlanID)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete subscription plan failed: not found: " + err.Error(),
				Module:      "SubscriptionPlan",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "SubscriptionPlan not found: " + err.Error()})
		}

		if err := c.modelcore.SubscriptionPlanManager.DeleteByID(context, *subscriptionPlanID); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete subscription plan failed: delete error: " + err.Error(),
				Module:      "SubscriptionPlan",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete subscription plan: " + err.Error()})
		}

		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "delete-success",
			Description: "Deleted subscription plan: " + subscriptionPlan.Name,
			Module:      "SubscriptionPlan",
		})

		return ctx.NoContent(http.StatusNoContent)
	})

	// Bulk delete subscription plans by IDs
	req.RegisterRoute(handlers.Route{
		Route:       "/api/v1/subscription-plan/bulk-delete",
		Method:      "DELETE",
		RequestType: modelcore.IDSRequest{},
		Note:        "Deletes multiple subscription plan records.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var reqBody modelcore.IDSRequest

		if err := ctx.Bind(&reqBody); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete subscription plans failed: invalid request body.",
				Module:      "SubscriptionPlan",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body: " + err.Error()})
		}

		if len(reqBody.IDs) == 0 {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete subscription plans failed: no IDs provided.",
				Module:      "SubscriptionPlan",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "No IDs provided for deletion."})
		}

		tx := c.provider.Service.Database.Client().Begin()
		if tx.Error != nil {
			tx.Rollback()
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete subscription plans failed: begin tx error: " + tx.Error.Error(),
				Module:      "SubscriptionPlan",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to begin transaction: " + tx.Error.Error()})
		}

		names := ""
		for _, rawID := range reqBody.IDs {
			subscriptionPlanID, err := uuid.Parse(rawID)
			if err != nil {
				tx.Rollback()
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "bulk-delete-error",
					Description: "Bulk delete subscription plans failed: invalid UUID: " + rawID,
					Module:      "SubscriptionPlan",
				})
				return ctx.JSON(http.StatusBadRequest, map[string]string{"error": fmt.Sprintf("Invalid UUID: %s - %v", rawID, err)})
			}

			subscriptionPlan, err := c.modelcore.SubscriptionPlanManager.GetByID(context, subscriptionPlanID)
			if err != nil {
				tx.Rollback()
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "bulk-delete-error",
					Description: "Bulk delete subscription plans failed: not found: " + rawID,
					Module:      "SubscriptionPlan",
				})
				return ctx.JSON(http.StatusNotFound, map[string]string{"error": fmt.Sprintf("SubscriptionPlan with ID %s not found: %v", rawID, err)})
			}

			names += subscriptionPlan.Name + ","
			if err := c.modelcore.SubscriptionPlanManager.DeleteByIDWithTx(context, tx, subscriptionPlanID); err != nil {
				tx.Rollback()
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "bulk-delete-error",
					Description: "Bulk delete subscription plans failed: delete error: " + err.Error(),
					Module:      "SubscriptionPlan",
				})
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": fmt.Sprintf("Failed to delete subscription plan with ID %s: %v", rawID, err)})
			}
		}

		if err := tx.Commit().Error; err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete subscription plans failed: commit tx error: " + err.Error(),
				Module:      "SubscriptionPlan",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to commit transaction: " + err.Error()})
		}

		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "bulk-delete-success",
			Description: "Bulk deleted subscription plans: " + names,
			Module:      "SubscriptionPlan",
		})

		return ctx.NoContent(http.StatusNoContent)
	})

	// GET /api/v1/subscription-plan/currency/:currency_id
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/subscription-plan/currency/:currency_id",
		Method:       "GET",
		ResponseType: modelcore.SubscriptionPlanResponse{},
		Note:         "Returns all subscription plans for a specific currency.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		currencyID, err := handlers.EngineUUIDParam(ctx, "currency_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid currency_id: " + err.Error()})
		}
		subscriptionPlans, err := c.modelcore.SubscriptionPlanManager.FindRaw(context, &modelcore.SubscriptionPlan{
			CurrencyID: currencyID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve subscription plans: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, subscriptionPlans)
	})

}
