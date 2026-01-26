package transactions

import (
	"net/http"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/helpers"
	"github.com/Lands-Horizon-Corp/e-coop-server/horizon"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/db/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/event"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/types"
	"github.com/labstack/echo/v4"
)

func PaymentTypeController(service *horizon.HorizonService) {
	req := service.API

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/payment-type",
		Method:       "GET",
		ResponseType: types.PaymentTypeResponse{},
		Note:         "Returns all payment types for the current user's branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		paymentTypes, err := core.PaymentTypeCurrentBranch(context, service, userOrg.OrganizationID, *userOrg.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve payment types: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, core.PaymentTypeManager(service).ToModels(paymentTypes))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/payment-type/search",
		Method:       "GET",
		ResponseType: types.PaymentTypeResponse{},
		Note:         "Returns paginated payment types for the current user's branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		value, err := core.PaymentTypeManager(service).NormalPagination(context, ctx, &types.PaymentType{
			BranchID:       *userOrg.BranchID,
			OrganizationID: userOrg.OrganizationID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve payment types for pagination: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, value)
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/payment-type/:payment_type_id",
		Method:       "GET",
		Note:         "Returns a specific payment type by its ID.",
		ResponseType: types.PaymentTypeResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		paymentTypeID, err := helpers.EngineUUIDParam(ctx, "payment_type_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid payment_type_id: " + err.Error()})
		}
		paymentType, err := core.PaymentTypeManager(service).GetByIDRaw(context, *paymentTypeID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "PaymentType not found: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, paymentType)
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/payment-type",
		Method:       "POST",
		ResponseType: types.PaymentTypeResponse{},
		RequestType:  types.PaymentTypeRequest{},
		Note:         "Creates a new payment type record for the current user's branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := core.PaymentTypeManager(service).Validate(ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create payment type failed: validation error: " + err.Error(),
				Module:      "PaymentType",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create payment type failed: user org error: " + err.Error(),
				Module:      "PaymentType",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}

		paymentType := &types.PaymentType{
			Name:           req.Name,
			Description:    req.Description,
			NumberOfDays:   req.NumberOfDays,
			Type:           req.Type,
			CreatedAt:      time.Now().UTC(),
			CreatedByID:    userOrg.UserID,
			UpdatedAt:      time.Now().UTC(),
			UpdatedByID:    userOrg.UserID,
			BranchID:       *userOrg.BranchID,
			OrganizationID: userOrg.OrganizationID,
		}

		if err := core.PaymentTypeManager(service).Create(context, paymentType); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create payment type failed: create error: " + err.Error(),
				Module:      "PaymentType",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create payment type: " + err.Error()})
		}

		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "create-success",
			Description: "Created payment type: " + paymentType.Name,
			Module:      "PaymentType",
		})

		return ctx.JSON(http.StatusOK, core.PaymentTypeManager(service).ToModel(paymentType))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/payment-type/:payment_type_id",
		Method:       "PUT",
		ResponseType: types.PaymentTypeResponse{},
		RequestType:  types.PaymentTypeRequest{},
		Note:         "Updates an existing payment type by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		paymentTypeID, err := helpers.EngineUUIDParam(ctx, "payment_type_id")
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update payment type failed: invalid payment_type_id: " + err.Error(),
				Module:      "PaymentType",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid payment_type_id: " + err.Error()})
		}

		req, err := core.PaymentTypeManager(service).Validate(ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update payment type failed: validation error: " + err.Error(),
				Module:      "PaymentType",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update payment type failed: user org error: " + err.Error(),
				Module:      "PaymentType",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}

		paymentType, err := core.PaymentTypeManager(service).GetByID(context, *paymentTypeID)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update payment type failed: not found: " + err.Error(),
				Module:      "PaymentType",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "PaymentType not found: " + err.Error()})
		}
		paymentType.Name = req.Name
		paymentType.Description = req.Description
		paymentType.NumberOfDays = req.NumberOfDays
		paymentType.Type = req.Type
		paymentType.UpdatedAt = time.Now().UTC()
		paymentType.UpdatedByID = userOrg.UserID
		if err := core.PaymentTypeManager(service).UpdateByID(context, paymentType.ID, paymentType); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update payment type failed: update error: " + err.Error(),
				Module:      "PaymentType",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update payment type: " + err.Error()})
		}
		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Updated payment type: " + paymentType.Name,
			Module:      "PaymentType",
		})
		return ctx.JSON(http.StatusOK, core.PaymentTypeManager(service).ToModel(paymentType))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:  "/api/v1/payment-type/:payment_type_id",
		Method: "DELETE",
		Note:   "Deletes a payment type record by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		paymentTypeID, err := helpers.EngineUUIDParam(ctx, "payment_type_id")
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete payment type failed: invalid payment_type_id: " + err.Error(),
				Module:      "PaymentType",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid payment_type_id: " + err.Error()})
		}
		paymentType, err := core.PaymentTypeManager(service).GetByID(context, *paymentTypeID)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete payment type failed: not found: " + err.Error(),
				Module:      "PaymentType",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "PaymentType not found: " + err.Error()})
		}
		if err := core.PaymentTypeManager(service).Delete(context, *paymentTypeID); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete payment type failed: delete error: " + err.Error(),
				Module:      "PaymentType",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete payment type: " + err.Error()})
		}
		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "delete-success",
			Description: "Deleted payment type: " + paymentType.Name,
			Module:      "PaymentType",
		})
		return ctx.NoContent(http.StatusNoContent)
	})

	req.RegisterWebRoute(horizon.Route{
		Route:       "/api/v1/payment-type/bulk-delete",
		Method:      "DELETE",
		Note:        "Deletes multiple payment type records by their IDs.",
		RequestType: types.IDSRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var reqBody types.IDSRequest

		if err := ctx.Bind(&reqBody); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Payment type bulk delete failed (/payment-type/bulk-delete) | invalid request body: " + err.Error(),
				Module:      "PaymentType",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body: " + err.Error()})
		}

		if len(reqBody.IDs) == 0 {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Payment type bulk delete failed (/payment-type/bulk-delete) | no IDs provided",
				Module:      "PaymentType",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "No IDs provided for bulk delete"})
		}

		ids := make([]any, len(reqBody.IDs))
		for i, id := range reqBody.IDs {
			ids[i] = id
		}
		if err := core.PaymentTypeManager(service).BulkDelete(context, ids); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Payment type bulk delete failed (/payment-type/bulk-delete) | error: " + err.Error(),
				Module:      "PaymentType",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to bulk delete payment types: " + err.Error()})
		}

		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "bulk-delete-success",
			Description: "Bulk deleted payment types (/payment-type/bulk-delete)",
			Module:      "PaymentType",
		})

		return ctx.NoContent(http.StatusNoContent)
	})
}
