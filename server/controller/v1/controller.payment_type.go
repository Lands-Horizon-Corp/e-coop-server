package v1

import (
	"net/http"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/server/event"
	"github.com/Lands-Horizon-Corp/e-coop-server/server/model/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/services/handlers"
	"github.com/labstack/echo/v4"
)

func (c *Controller) paymentTypeController() {
	req := c.provider.Service.Request

	// Get all payment types for the current branch
	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/payment-type",
		Method:       "GET",
		ResponseType: core.PaymentTypeResponse{},
		Note:         "Returns all payment types for the current user's branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		paymentTypes, err := c.core.PaymentTypeCurrentBranch(context, userOrg.OrganizationID, *userOrg.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve payment types: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.core.PaymentTypeManager.ToModels(paymentTypes))
	})

	// Paginate payment types for the current branch
	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/payment-type/search",
		Method:       "GET",
		ResponseType: core.PaymentTypeResponse{},
		Note:         "Returns paginated payment types for the current user's branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		value, err := c.core.PaymentTypeManager.PaginationWithFields(context, ctx, &core.PaymentType{
			BranchID:       *userOrg.BranchID,
			OrganizationID: userOrg.OrganizationID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve payment types for pagination: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, value)
	})

	// Get a payment type by its ID
	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/payment-type/:payment_type_id",
		Method:       "GET",
		Note:         "Returns a specific payment type by its ID.",
		ResponseType: core.PaymentTypeResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		paymentTypeID, err := handlers.EngineUUIDParam(ctx, "payment_type_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid payment_type_id: " + err.Error()})
		}
		paymentType, err := c.core.PaymentTypeManager.GetByIDRaw(context, *paymentTypeID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "PaymentType not found: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, paymentType)
	})

	// Create a new payment type
	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/payment-type",
		Method:       "POST",
		ResponseType: core.PaymentTypeResponse{},
		RequestType:  core.PaymentTypeRequest{},
		Note:         "Creates a new payment type record for the current user's branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := c.core.PaymentTypeManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create payment type failed: validation error: " + err.Error(),
				Module:      "PaymentType",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create payment type failed: user org error: " + err.Error(),
				Module:      "PaymentType",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}

		paymentType := &core.PaymentType{
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

		if err := c.core.PaymentTypeManager.Create(context, paymentType); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create payment type failed: create error: " + err.Error(),
				Module:      "PaymentType",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create payment type: " + err.Error()})
		}

		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "create-success",
			Description: "Created payment type: " + paymentType.Name,
			Module:      "PaymentType",
		})

		return ctx.JSON(http.StatusOK, c.core.PaymentTypeManager.ToModel(paymentType))
	})

	// Update a payment type by its ID
	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/payment-type/:payment_type_id",
		Method:       "PUT",
		ResponseType: core.PaymentTypeResponse{},
		RequestType:  core.PaymentTypeRequest{},
		Note:         "Updates an existing payment type by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		paymentTypeID, err := handlers.EngineUUIDParam(ctx, "payment_type_id")
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update payment type failed: invalid payment_type_id: " + err.Error(),
				Module:      "PaymentType",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid payment_type_id: " + err.Error()})
		}

		req, err := c.core.PaymentTypeManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update payment type failed: validation error: " + err.Error(),
				Module:      "PaymentType",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update payment type failed: user org error: " + err.Error(),
				Module:      "PaymentType",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}

		paymentType, err := c.core.PaymentTypeManager.GetByID(context, *paymentTypeID)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
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
		if err := c.core.PaymentTypeManager.UpdateByID(context, paymentType.ID, paymentType); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update payment type failed: update error: " + err.Error(),
				Module:      "PaymentType",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update payment type: " + err.Error()})
		}
		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Updated payment type: " + paymentType.Name,
			Module:      "PaymentType",
		})
		return ctx.JSON(http.StatusOK, c.core.PaymentTypeManager.ToModel(paymentType))
	})

	// Delete a payment type by its ID
	req.RegisterWebRoute(handlers.Route{
		Route:  "/api/v1/payment-type/:payment_type_id",
		Method: "DELETE",
		Note:   "Deletes a payment type record by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		paymentTypeID, err := handlers.EngineUUIDParam(ctx, "payment_type_id")
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete payment type failed: invalid payment_type_id: " + err.Error(),
				Module:      "PaymentType",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid payment_type_id: " + err.Error()})
		}
		paymentType, err := c.core.PaymentTypeManager.GetByID(context, *paymentTypeID)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete payment type failed: not found: " + err.Error(),
				Module:      "PaymentType",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "PaymentType not found: " + err.Error()})
		}
		if err := c.core.PaymentTypeManager.Delete(context, *paymentTypeID); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete payment type failed: delete error: " + err.Error(),
				Module:      "PaymentType",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete payment type: " + err.Error()})
		}
		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "delete-success",
			Description: "Deleted payment type: " + paymentType.Name,
			Module:      "PaymentType",
		})
		return ctx.NoContent(http.StatusNoContent)
	})

	// Simplified bulk-delete handler for payment types (mirrors feedback/holiday pattern)
	req.RegisterWebRoute(handlers.Route{
		Route:       "/api/v1/payment-type/bulk-delete",
		Method:      "DELETE",
		Note:        "Deletes multiple payment type records by their IDs.",
		RequestType: core.IDSRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var reqBody core.IDSRequest

		if err := ctx.Bind(&reqBody); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Payment type bulk delete failed (/payment-type/bulk-delete) | invalid request body: " + err.Error(),
				Module:      "PaymentType",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body: " + err.Error()})
		}

		if len(reqBody.IDs) == 0 {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Payment type bulk delete failed (/payment-type/bulk-delete) | no IDs provided",
				Module:      "PaymentType",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "No IDs provided for bulk delete"})
		}

		// Delegate deletion to the manager. Manager should handle transactions, validations and DeletedBy bookkeeping.
		if err := c.core.PaymentTypeManager.BulkDelete(context, reqBody.IDs); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Payment type bulk delete failed (/payment-type/bulk-delete) | error: " + err.Error(),
				Module:      "PaymentType",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to bulk delete payment types: " + err.Error()})
		}

		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "bulk-delete-success",
			Description: "Bulk deleted payment types (/payment-type/bulk-delete)",
			Module:      "PaymentType",
		})

		return ctx.NoContent(http.StatusNoContent)
	})
}
