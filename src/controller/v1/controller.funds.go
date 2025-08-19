package controller_v1

import (
	"fmt"
	"net/http"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/services/handlers"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/event"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/model"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func (c *Controller) FundsController() {
	req := c.provider.Service.Request

	// Get all funds for the current branch
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/funds",
		Method:       "GET",
		ResponseType: model.FundsResponse{},
		Note:         "Returns all funds for the current user's branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		funds, err := c.model.FundsCurrentBranch(context, user.OrganizationID, *user.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get funds: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.FundsManager.Filtered(context, ctx, funds))
	})

	// Get paginated funds
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/funds/search",
		Method:       "GET",
		ResponseType: model.FundsResponse{},
		Note:         "Returns paginated funds for the current user's branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		funds, err := c.model.FundsCurrentBranch(context, user.OrganizationID, *user.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get funds for pagination: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.FundsManager.Pagination(context, ctx, funds))
	})

	// Create a new funds record
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/funds",
		Method:       "POST",
		ResponseType: model.FundsResponse{},
		RequestType:  model.FundsRequest{},
		Note:         "Creates a new funds record.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := c.model.FundsManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create funds failed (/funds), validation error: " + err.Error(),
				Module:      "Funds",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create funds failed (/funds), user org error: " + err.Error(),
				Module:      "Funds",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}

		funds := &model.Funds{
			AccountID:      req.AccountID,
			Type:           req.Type,
			Description:    req.Description,
			Icon:           req.Icon,
			GLBooks:        req.GLBooks,
			CreatedAt:      time.Now().UTC(),
			CreatedByID:    user.UserID,
			UpdatedAt:      time.Now().UTC(),
			UpdatedByID:    user.UserID,
			BranchID:       *user.BranchID,
			OrganizationID: user.OrganizationID,
		}

		if err := c.model.FundsManager.Create(context, funds); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create funds failed (/funds), db error: " + err.Error(),
				Module:      "Funds",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create funds: " + err.Error()})
		}

		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "create-success",
			Description: "Created funds (/funds): " + funds.Type,
			Module:      "Funds",
		})

		return ctx.JSON(http.StatusOK, c.model.FundsManager.ToModel(funds))
	})

	// Update an existing funds record by ID
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/funds/:funds_id",
		Method:       "PUT",
		ResponseType: model.FundsResponse{},
		RequestType:  model.FundsRequest{},
		Note:         "Updates an existing funds record by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		fundsID, err := handlers.EngineUUIDParam(ctx, "funds_id")
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update funds failed (/funds/:funds_id), invalid funds_id: " + err.Error(),
				Module:      "Funds",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid funds_id: " + err.Error()})
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update funds failed (/funds/:funds_id), user org error: " + err.Error(),
				Module:      "Funds",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		req, err := c.model.FundsManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update funds failed (/funds/:funds_id), validation error: " + err.Error(),
				Module:      "Funds",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		funds, err := c.model.FundsManager.GetByID(context, *fundsID)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update funds failed (/funds/:funds_id), not found: " + err.Error(),
				Module:      "Funds",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Funds not found: " + err.Error()})
		}
		funds.UpdatedAt = time.Now().UTC()
		funds.UpdatedByID = user.UserID
		funds.OrganizationID = user.OrganizationID
		funds.BranchID = *user.BranchID
		funds.AccountID = req.AccountID
		funds.Type = req.Type
		funds.Description = req.Description
		funds.Icon = req.Icon
		funds.GLBooks = req.GLBooks
		if err := c.model.FundsManager.UpdateFields(context, funds.ID, funds); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update funds failed (/funds/:funds_id), db error: " + err.Error(),
				Module:      "Funds",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update funds: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Updated funds (/funds/:funds_id): " + funds.Type,
			Module:      "Funds",
		})
		return ctx.JSON(http.StatusOK, c.model.FundsManager.ToModel(funds))
	})

	// Delete a funds record by ID
	req.RegisterRoute(handlers.Route{
		Route:  "/api/v1/funds/:funds_id",
		Method: "DELETE",
		Note:   "Deletes a funds record by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		fundsID, err := handlers.EngineUUIDParam(ctx, "funds_id")
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete funds failed (/funds/:funds_id), invalid funds_id: " + err.Error(),
				Module:      "Funds",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid funds_id: " + err.Error()})
		}
		value, err := c.model.FundsManager.GetByID(context, *fundsID)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete funds failed (/funds/:funds_id), record not found: " + err.Error(),
				Module:      "Funds",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Funds not found: " + err.Error()})
		}
		if err := c.model.FundsManager.DeleteByID(context, *fundsID); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete funds failed (/funds/:funds_id), db error: " + err.Error(),
				Module:      "Funds",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete funds: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "delete-success",
			Description: "Deleted funds (/funds/:funds_id): " + value.Type,
			Module:      "Funds",
		})
		return ctx.NoContent(http.StatusNoContent)
	})

	// Bulk delete funds by IDs
	req.RegisterRoute(handlers.Route{
		Route:       "/api/v1/funds/bulk-delete",
		Method:      "DELETE",
		Note:        "Deletes multiple funds records by their IDs.",
		RequestType: model.IDSRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var reqBody model.IDSRequest

		if err := ctx.Bind(&reqBody); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete funds failed (/funds/bulk-delete), invalid request body.",
				Module:      "Funds",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body: " + err.Error()})
		}

		if len(reqBody.IDs) == 0 {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete funds failed (/funds/bulk-delete), no IDs provided.",
				Module:      "Funds",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "No IDs provided for deletion."})
		}

		tx := c.provider.Service.Database.Client().Begin()
		if tx.Error != nil {
			tx.Rollback()
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete funds failed (/funds/bulk-delete), begin tx error: " + tx.Error.Error(),
				Module:      "Funds",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to begin transaction: " + tx.Error.Error()})
		}

		types := ""
		for _, rawID := range reqBody.IDs {
			fundsID, err := uuid.Parse(rawID)
			if err != nil {
				tx.Rollback()
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "bulk-delete-error",
					Description: "Bulk delete funds failed (/funds/bulk-delete), invalid UUID: " + rawID,
					Module:      "Funds",
				})
				return ctx.JSON(http.StatusBadRequest, map[string]string{"error": fmt.Sprintf("Invalid UUID '%s': %s", rawID, err.Error())})
			}

			value, err := c.model.FundsManager.GetByID(context, fundsID)
			if err != nil {
				tx.Rollback()
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "bulk-delete-error",
					Description: "Bulk delete funds failed (/funds/bulk-delete), not found: " + rawID,
					Module:      "Funds",
				})
				return ctx.JSON(http.StatusNotFound, map[string]string{"error": fmt.Sprintf("Funds with ID '%s' not found: %s", rawID, err.Error())})
			}

			types += value.Type + ","
			if err := c.model.FundsManager.DeleteByIDWithTx(context, tx, fundsID); err != nil {
				tx.Rollback()
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "bulk-delete-error",
					Description: "Bulk delete funds failed (/funds/bulk-delete), db error: " + err.Error(),
					Module:      "Funds",
				})
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": fmt.Sprintf("Failed to delete funds with ID '%s': %s", rawID, err.Error())})
			}
		}

		if err := tx.Commit().Error; err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete funds failed (/funds/bulk-delete), commit error: " + err.Error(),
				Module:      "Funds",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to commit transaction: " + err.Error()})
		}

		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "bulk-delete-success",
			Description: "Bulk deleted funds (/funds/bulk-delete): " + types,
			Module:      "Funds",
		})

		return ctx.NoContent(http.StatusNoContent)
	})
}
