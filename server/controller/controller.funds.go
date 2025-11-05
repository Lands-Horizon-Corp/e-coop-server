package v1

import (
	"net/http"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/server/event"
	"github.com/Lands-Horizon-Corp/e-coop-server/server/model/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/services/handlers"
	"github.com/labstack/echo/v4"
)

func (c *Controller) fundsController() {
	req := c.provider.Service.Request

	// Get all funds for the current branch
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/funds",
		Method:       "GET",
		ResponseType: core.FundsResponse{},
		Note:         "Returns all funds for the current user's branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		funds, err := c.core.FundsCurrentBranch(context, user.OrganizationID, *user.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get funds: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.core.FundsManager.ToModels(funds))
	})

	// Get paginated funds
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/funds/search",
		Method:       "GET",
		ResponseType: core.FundsResponse{},
		Note:         "Returns paginated funds for the current user's branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		funds, err := c.core.FundsManager.PaginationWithFields(context, ctx, &core.Funds{
			BranchID:       *user.BranchID,
			OrganizationID: user.OrganizationID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get funds for pagination: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, funds)
	})

	// Create a new funds record
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/funds",
		Method:       "POST",
		ResponseType: core.FundsResponse{},
		RequestType:  core.FundsRequest{},
		Note:         "Creates a new funds record.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := c.core.FundsManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create funds failed (/funds), validation error: " + err.Error(),
				Module:      "Funds",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create funds failed (/funds), user org error: " + err.Error(),
				Module:      "Funds",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}

		funds := &core.Funds{
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

		if err := c.core.FundsManager.Create(context, funds); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create funds failed (/funds), db error: " + err.Error(),
				Module:      "Funds",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create funds: " + err.Error()})
		}

		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "create-success",
			Description: "Created funds (/funds): " + funds.Type,
			Module:      "Funds",
		})

		return ctx.JSON(http.StatusOK, c.core.FundsManager.ToModel(funds))
	})

	// Update an existing funds record by ID
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/funds/:funds_id",
		Method:       "PUT",
		ResponseType: core.FundsResponse{},
		RequestType:  core.FundsRequest{},
		Note:         "Updates an existing funds record by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		fundsID, err := handlers.EngineUUIDParam(ctx, "funds_id")
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update funds failed (/funds/:funds_id), invalid funds_id: " + err.Error(),
				Module:      "Funds",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid funds_id: " + err.Error()})
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update funds failed (/funds/:funds_id), user org error: " + err.Error(),
				Module:      "Funds",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		req, err := c.core.FundsManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update funds failed (/funds/:funds_id), validation error: " + err.Error(),
				Module:      "Funds",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		funds, err := c.core.FundsManager.GetByID(context, *fundsID)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
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
		if err := c.core.FundsManager.UpdateByID(context, funds.ID, funds); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update funds failed (/funds/:funds_id), db error: " + err.Error(),
				Module:      "Funds",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update funds: " + err.Error()})
		}
		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Updated funds (/funds/:funds_id): " + funds.Type,
			Module:      "Funds",
		})
		return ctx.JSON(http.StatusOK, c.core.FundsManager.ToModel(funds))
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
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete funds failed (/funds/:funds_id), invalid funds_id: " + err.Error(),
				Module:      "Funds",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid funds_id: " + err.Error()})
		}
		value, err := c.core.FundsManager.GetByID(context, *fundsID)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete funds failed (/funds/:funds_id), record not found: " + err.Error(),
				Module:      "Funds",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Funds not found: " + err.Error()})
		}
		if err := c.core.FundsManager.Delete(context, *fundsID); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete funds failed (/funds/:funds_id), db error: " + err.Error(),
				Module:      "Funds",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete funds: " + err.Error()})
		}
		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "delete-success",
			Description: "Deleted funds (/funds/:funds_id): " + value.Type,
			Module:      "Funds",
		})
		return ctx.NoContent(http.StatusNoContent)
	})

	req.RegisterRoute(handlers.Route{
		Route:       "/api/v1/funds/bulk-delete",
		Method:      "DELETE",
		Note:        "Deletes multiple funds records by their IDs.",
		RequestType: core.IDSRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var reqBody core.IDSRequest

		if err := ctx.Bind(&reqBody); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete funds failed (/funds/bulk-delete) | invalid request body: " + err.Error(),
				Module:      "Funds",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body: " + err.Error()})
		}

		if len(reqBody.IDs) == 0 {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete funds failed (/funds/bulk-delete) | no IDs provided",
				Module:      "Funds",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "No IDs provided for deletion."})
		}

		if err := c.core.FundsManager.BulkDelete(context, reqBody.IDs); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete funds failed (/funds/bulk-delete) | error: " + err.Error(),
				Module:      "Funds",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to bulk delete funds: " + err.Error()})
		}

		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "bulk-delete-success",
			Description: "Bulk deleted funds (/funds/bulk-delete)",
			Module:      "Funds",
		})

		return ctx.NoContent(http.StatusNoContent)
	})
}
