package v1

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

func (c *Controller) paymentTypeController(
	req := c.provider.Service.Request

	// Get all payment types for the current branch
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/payment-type",
		Method:       "GET",
		ResponseType: modelcore.PaymentTypeResponse{},
		Note:         "Returns all payment types for the current user's branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		paymentTypes, err := c.modelcore.PaymentTypeCurrentbranch(context, user.OrganizationID, *user.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve payment types: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.modelcore.PaymentTypeManager.Filtered(context, ctx, paymentTypes))
	})

	// Paginate payment types for the current branch
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/payment-type/search",
		Method:       "GET",
		ResponseType: modelcore.PaymentTypeResponse{},
		Note:         "Returns paginated payment types for the current user's branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		value, err := c.modelcore.PaymentTypeCurrentbranch(context, user.OrganizationID, *user.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve payment types for pagination: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.modelcore.PaymentTypeManager.Pagination(context, ctx, value))
	})

	// Get a payment type by its ID
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/payment-type/:payment_type_id",
		Method:       "GET",
		Note:         "Returns a specific payment type by its ID.",
		ResponseType: modelcore.PaymentTypeResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		paymentTypeID, err := handlers.EngineUUIDParam(ctx, "payment_type_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid payment_type_id: " + err.Error()})
		}
		paymentType, err := c.modelcore.PaymentTypeManager.GetByIDRaw(context, *paymentTypeID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "PaymentType not found: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, paymentType)
	})

	// Create a new payment type
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/payment-type",
		Method:       "POST",
		ResponseType: modelcore.PaymentTypeResponse{},
		RequestType:  modelcore.PaymentTypeRequest{},
		Note:         "Creates a new payment type record for the current user's branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := c.modelcore.PaymentTypeManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create payment type failed: validation error: " + err.Error(),
				Module:      "PaymentType",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create payment type failed: user org error: " + err.Error(),
				Module:      "PaymentType",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}

		paymentType := &modelcore.PaymentType{
			Name:           req.Name,
			Description:    req.Description,
			NumberOfDays:   req.NumberOfDays,
			Type:           req.Type,
			CreatedAt:      time.Now().UTC(),
			CreatedByID:    user.UserID,
			UpdatedAt:      time.Now().UTC(),
			UpdatedByID:    user.UserID,
			BranchID:       *user.BranchID,
			OrganizationID: user.OrganizationID,
		}

		if err := c.modelcore.PaymentTypeManager.Create(context, paymentType); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create payment type failed: create error: " + err.Error(),
				Module:      "PaymentType",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create payment type: " + err.Error()})
		}

		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "create-success",
			Description: "Created payment type: " + paymentType.Name,
			Module:      "PaymentType",
		})

		return ctx.JSON(http.StatusOK, c.modelcore.PaymentTypeManager.ToModel(paymentType))
	})

	// Update a payment type by its ID
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/payment-type/:payment_type_id",
		Method:       "PUT",
		ResponseType: modelcore.PaymentTypeResponse{},
		RequestType:  modelcore.PaymentTypeRequest{},
		Note:         "Updates an existing payment type by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		paymentTypeID, err := handlers.EngineUUIDParam(ctx, "payment_type_id")
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update payment type failed: invalid payment_type_id: " + err.Error(),
				Module:      "PaymentType",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid payment_type_id: " + err.Error()})
		}

		req, err := c.modelcore.PaymentTypeManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update payment type failed: validation error: " + err.Error(),
				Module:      "PaymentType",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update payment type failed: user org error: " + err.Error(),
				Module:      "PaymentType",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}

		paymentType, err := c.modelcore.PaymentTypeManager.GetByID(context, *paymentTypeID)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
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
		paymentType.UpdatedByID = user.UserID
		if err := c.modelcore.PaymentTypeManager.UpdateFields(context, paymentType.ID, paymentType); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update payment type failed: update error: " + err.Error(),
				Module:      "PaymentType",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update payment type: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Updated payment type: " + paymentType.Name,
			Module:      "PaymentType",
		})
		return ctx.JSON(http.StatusOK, c.modelcore.PaymentTypeManager.ToModel(paymentType))
	})

	// Delete a payment type by its ID
	req.RegisterRoute(handlers.Route{
		Route:  "/api/v1/payment-type/:payment_type_id",
		Method: "DELETE",
		Note:   "Deletes a payment type record by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		paymentTypeID, err := handlers.EngineUUIDParam(ctx, "payment_type_id")
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete payment type failed: invalid payment_type_id: " + err.Error(),
				Module:      "PaymentType",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid payment_type_id: " + err.Error()})
		}
		paymentType, err := c.modelcore.PaymentTypeManager.GetByID(context, *paymentTypeID)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete payment type failed: not found: " + err.Error(),
				Module:      "PaymentType",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "PaymentType not found: " + err.Error()})
		}
		if err := c.modelcore.PaymentTypeManager.DeleteByID(context, *paymentTypeID); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete payment type failed: delete error: " + err.Error(),
				Module:      "PaymentType",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete payment type: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "delete-success",
			Description: "Deleted payment type: " + paymentType.Name,
			Module:      "PaymentType",
		})
		return ctx.NoContent(http.StatusNoContent)
	})

	// Bulk delete payment types by IDs
	req.RegisterRoute(handlers.Route{
		Route:       "/api/v1/payment-type/bulk-delete",
		Method:      "DELETE",
		RequestType: modelcore.IDSRequest{},
		Note:        "Deletes multiple payment type records by their IDs.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var reqBody modelcore.IDSRequest
		if err := ctx.Bind(&reqBody); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete payment types failed: invalid request body.",
				Module:      "PaymentType",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body: " + err.Error()})
		}
		if len(reqBody.IDs) == 0 {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete payment types failed: no IDs provided.",
				Module:      "PaymentType",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "No IDs provided for deletion."})
		}
		tx := c.provider.Service.Database.Client().Begin()
		if tx.Error != nil {
			tx.Rollback()
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete payment types failed: begin tx error: " + tx.Error.Error(),
				Module:      "PaymentType",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to begin transaction: " + tx.Error.Error()})
		}
		names := ""
		for _, rawID := range reqBody.IDs {
			paymentTypeID, err := uuid.Parse(rawID)
			if err != nil {
				tx.Rollback()
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "bulk-delete-error",
					Description: "Bulk delete payment types failed: invalid UUID: " + rawID,
					Module:      "PaymentType",
				})
				return ctx.JSON(http.StatusBadRequest, map[string]string{"error": fmt.Sprintf("Invalid UUID: %s - %v", rawID, err)})
			}
			paymentType, err := c.modelcore.PaymentTypeManager.GetByID(context, paymentTypeID)
			if err != nil {
				tx.Rollback()
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "bulk-delete-error",
					Description: "Bulk delete payment types failed: not found: " + rawID,
					Module:      "PaymentType",
				})
				return ctx.JSON(http.StatusNotFound, map[string]string{"error": fmt.Sprintf("PaymentType with ID %s not found: %v", rawID, err)})
			}
			names += paymentType.Name + ","
			if err := c.modelcore.PaymentTypeManager.DeleteByIDWithTx(context, tx, paymentTypeID); err != nil {
				tx.Rollback()
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "bulk-delete-error",
					Description: "Bulk delete payment types failed: delete error: " + err.Error(),
					Module:      "PaymentType",
				})
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": fmt.Sprintf("Failed to delete payment type with ID %s: %v", rawID, err)})
			}
		}
		if err := tx.Commit().Error; err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete payment types failed: commit tx error: " + err.Error(),
				Module:      "PaymentType",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to commit transaction: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "bulk-delete-success",
			Description: "Bulk deleted payment types: " + names,
			Module:      "PaymentType",
		})
		return ctx.NoContent(http.StatusNoContent)
	})
}
