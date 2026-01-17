package settings

import (
	"net/http"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/helpers"
	"github.com/Lands-Horizon-Corp/e-coop-server/horizon"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/event"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/types"
	"github.com/labstack/echo/v4"
)

func SubscriptionPlanController(service *horizon.HorizonService) {
	req := service.API

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/subscription-plan",
		Method:       "GET",
		ResponseType: types.SubscriptionPlanResponse{},
		Note:         "Returns all subscription plans.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		categories, err := core.SubscriptionPlanManager(service).List(context)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve subscription plans: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, core.SubscriptionPlanManager(service).ToModels(categories))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/subscription-plan/:subscription_plan_id",
		Method:       "GET",
		ResponseType: types.SubscriptionPlanResponse{},
		Note:         "Returns a specific subscription plan by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		subscriptionPlanID, err := helpers.EngineUUIDParam(ctx, "subscription_plan_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid subscription_plan_id: " + err.Error()})
		}

		subscriptionPlan, err := core.SubscriptionPlanManager(service).GetByIDRaw(context, *subscriptionPlanID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "SubscriptionPlan not found: " + err.Error()})
		}

		return ctx.JSON(http.StatusOK, subscriptionPlan)
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/subscription-plan",
		Method:       "POST",
		ResponseType: types.SubscriptionPlanResponse{},
		RequestType: types.SubscriptionPlanRequest{},
		Note:         "Creates a new subscription plan.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := core.SubscriptionPlanManager(service).Validate(ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create subscription plan failed: validation error: " + err.Error(),
				Module:      "SubscriptionPlan",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}

		subscriptionPlan := &types.SubscriptionPlan{
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

		if err := core.SubscriptionPlanManager(service).Create(context, subscriptionPlan); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create subscription plan failed: create error: " + err.Error(),
				Module:      "SubscriptionPlan",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create subscription plan: " + err.Error()})
		}

		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "create-success",
			Description: "Created subscription plan: " + subscriptionPlan.Name,
			Module:      "SubscriptionPlan",
		})

		return ctx.JSON(http.StatusOK, core.SubscriptionPlanManager(service).ToModel(subscriptionPlan))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/subscription-plan/:subscription_plan_id",
		Method:       "PUT",
		ResponseType: types.SubscriptionPlanResponse{},
		RequestType: types.SubscriptionPlanRequest{},
		Note:         "Updates an existing subscription plan by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		subscriptionPlanID, err := helpers.EngineUUIDParam(ctx, "subscription_plan_id")
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update subscription plan failed: invalid subscription_plan_id: " + err.Error(),
				Module:      "SubscriptionPlan",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid subscription_plan_id: " + err.Error()})
		}

		req, err := core.SubscriptionPlanManager(service).Validate(ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update subscription plan failed: validation error: " + err.Error(),
				Module:      "SubscriptionPlan",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}

		subscriptionPlan, err := core.SubscriptionPlanManager(service).GetByID(context, *subscriptionPlanID)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
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

		if err := core.SubscriptionPlanManager(service).UpdateByID(context, subscriptionPlan.ID, subscriptionPlan); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update subscription plan failed: update error: " + err.Error(),
				Module:      "SubscriptionPlan",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update subscription plan: " + err.Error()})
		}

		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Updated subscription plan: " + subscriptionPlan.Name,
			Module:      "SubscriptionPlan",
		})

		return ctx.JSON(http.StatusOK, core.SubscriptionPlanManager(service).ToModel(subscriptionPlan))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:  "/api/v1/subscription-plan/:subscription_plan_id",
		Method: "DELETE",
		Note:   "Deletes a subscription plan by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		subscriptionPlanID, err := helpers.EngineUUIDParam(ctx, "subscription_plan_id")
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete subscription plan failed: invalid subscription_plan_id: " + err.Error(),
				Module:      "SubscriptionPlan",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid subscription_plan_id: " + err.Error()})
		}

		subscriptionPlan, err := core.SubscriptionPlanManager(service).GetByID(context, *subscriptionPlanID)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete subscription plan failed: not found: " + err.Error(),
				Module:      "SubscriptionPlan",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "SubscriptionPlan not found: " + err.Error()})
		}

		if err := core.SubscriptionPlanManager(service).Delete(context, *subscriptionPlanID); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete subscription plan failed: delete error: " + err.Error(),
				Module:      "SubscriptionPlan",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete subscription plan: " + err.Error()})
		}

		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "delete-success",
			Description: "Deleted subscription plan: " + subscriptionPlan.Name,
			Module:      "SubscriptionPlan",
		})

		return ctx.NoContent(http.StatusNoContent)
	})

	req.RegisterWebRoute(horizon.Route{
		Route:       "/api/v1/subscription-plan/bulk-delete",
		Method:      "DELETE",
		Note:        "Deletes multiple subscription plan records.",
		RequestType: types.IDSRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var reqBody core.IDSRequest

		if err := ctx.Bind(&reqBody); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Subscription plan bulk delete failed (/subscription-plan/bulk-delete) | invalid request body: " + err.Error(),
				Module:      "SubscriptionPlan",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body: " + err.Error()})
		}

		if len(reqBody.IDs) == 0 {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Subscription plan bulk delete failed (/subscription-plan/bulk-delete) | no IDs provided",
				Module:      "SubscriptionPlan",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "No IDs provided for bulk delete"})
		}

		ids := make([]any, len(reqBody.IDs))
		for i, id := range reqBody.IDs {
			ids[i] = id
		}
		if err := core.SubscriptionPlanManager(service).BulkDelete(context, ids); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Subscription plan bulk delete failed (/subscription-plan/bulk-delete) | error: " + err.Error(),
				Module:      "SubscriptionPlan",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to bulk delete subscription plans: " + err.Error()})
		}

		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "bulk-delete-success",
			Description: "Bulk deleted subscription plans (/subscription-plan/bulk-delete)",
			Module:      "SubscriptionPlan",
		})

		return ctx.NoContent(http.StatusNoContent)
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/subscription-plan/currency/:currency_id",
		Method:       "GET",
		ResponseType: types.SubscriptionPlanResponse{},
		Note:         "Returns all subscription plans for a specific currency.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		currencyID, err := helpers.EngineUUIDParam(ctx, "currency_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid currency_id: " + err.Error()})
		}
		subscriptionPlans, err := core.SubscriptionPlanManager(service).FindRaw(context, &types.SubscriptionPlan{
			CurrencyID: currencyID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve subscription plans: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, subscriptionPlans)
	})

}
