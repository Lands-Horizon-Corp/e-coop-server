package controller

import (
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/lands-horizon/horizon-server/services/horizon"
	"github.com/lands-horizon/horizon-server/src/event"
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
		categories, err := c.model.SubscriptionPlanManager.List(context)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve subscription plans: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.SubscriptionPlanManager.Filtered(context, ctx, categories))
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
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create subscription plan failed: validation error: " + err.Error(),
				Module:      "SubscriptionPlan",
			})
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
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update subscription plan failed: invalid subscription_plan_id: " + err.Error(),
				Module:      "SubscriptionPlan",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid subscription_plan_id: " + err.Error()})
		}

		req, err := c.model.SubscriptionPlanManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update subscription plan failed: validation error: " + err.Error(),
				Module:      "SubscriptionPlan",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}

		subscriptionPlan, err := c.model.SubscriptionPlanManager.GetByID(context, *subscriptionPlanID)
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
		subscriptionPlan.UpdatedAt = time.Now().UTC()

		if err := c.model.SubscriptionPlanManager.UpdateFields(context, subscriptionPlan.ID, subscriptionPlan); err != nil {
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
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete subscription plan failed: invalid subscription_plan_id: " + err.Error(),
				Module:      "SubscriptionPlan",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid subscription_plan_id: " + err.Error()})
		}

		subscriptionPlan, err := c.model.SubscriptionPlanManager.GetByID(context, *subscriptionPlanID)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete subscription plan failed: not found: " + err.Error(),
				Module:      "SubscriptionPlan",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "SubscriptionPlan not found: " + err.Error()})
		}

		if err := c.model.SubscriptionPlanManager.DeleteByID(context, *subscriptionPlanID); err != nil {
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

			subscriptionPlan, err := c.model.SubscriptionPlanManager.GetByID(context, subscriptionPlanID)
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
			if err := c.model.SubscriptionPlanManager.DeleteByIDWithTx(context, tx, subscriptionPlanID); err != nil {
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
}
