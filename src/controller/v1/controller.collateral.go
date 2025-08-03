package controller_v1

import (
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/lands-horizon/horizon-server/services/handlers"
	"github.com/lands-horizon/horizon-server/src/event"
	"github.com/lands-horizon/horizon-server/src/model"
)

// CollateralController manages endpoints for collateral operations.
func (c *Controller) CollateralController() {
	req := c.provider.Service.Request

	// GET /collateral: List all collaterals for the current user's branch. (NO footstep)
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/collateral",
		Method:       "GET",
		Note:         "Returns all collateral records for the current user's organization and branch. Returns error if not authenticated.",
		ResponseType: model.CollateralResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if user.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		collaterals, err := c.model.CollateralCurrentBranch(context, user.OrganizationID, *user.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "No collateral records found for the current branch"})
		}
		return ctx.JSON(http.StatusOK, c.model.CollateralManager.Filtered(context, ctx, collaterals))
	})

	// GET /collateral/search: Paginated search of collaterals for current branch. (NO footstep)
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/collateral/search",
		Method:       "GET",
		ResponseType: model.CollateralResponse{},
		Note:         "Returns a paginated list of collateral records for the current user's organization and branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if user.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		collaterals, err := c.model.CollateralCurrentBranch(context, user.OrganizationID, *user.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch collateral records: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.CollateralManager.Pagination(context, ctx, collaterals))
	})

	// GET /collateral/:collateral_id: Get a specific collateral record by ID. (NO footstep)
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/collateral/:collateral_id",
		Method:       "GET",
		Note:         "Returns a collateral record by its ID.",
		ResponseType: model.CollateralResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		collateralID, err := handlers.EngineUUIDParam(ctx, "collateral_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid collateral ID"})
		}
		collateral, err := c.model.CollateralManager.GetByIDRaw(context, *collateralID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Collateral record not found"})
		}
		return ctx.JSON(http.StatusOK, collateral)
	})

	// POST /collateral: Create a new collateral record. (WITH footstep)
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/collateral",
		Method:       "POST",
		RequestType:  model.CollateralRequest{},
		ResponseType: model.CollateralResponse{},
		Note:         "Creates a new collateral record for the current user's organization and branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := c.model.CollateralManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Collateral creation failed (/collateral), validation error: " + err.Error(),
				Module:      "Collateral",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid collateral data: " + err.Error()})
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Collateral creation failed (/collateral), user org error: " + err.Error(),
				Module:      "Collateral",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if user.BranchID == nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Collateral creation failed (/collateral), user not assigned to branch.",
				Module:      "Collateral",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}

		collateral := &model.Collateral{
			Icon:           req.Icon,
			Name:           req.Name,
			Description:    req.Description,
			CreatedAt:      time.Now().UTC(),
			CreatedByID:    user.UserID,
			UpdatedAt:      time.Now().UTC(),
			UpdatedByID:    user.UserID,
			BranchID:       *user.BranchID,
			OrganizationID: user.OrganizationID,
		}

		if err := c.model.CollateralManager.Create(context, collateral); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Collateral creation failed (/collateral), db error: " + err.Error(),
				Module:      "Collateral",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create collateral record: " + err.Error()})
		}

		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "create-success",
			Description: "Created collateral (/collateral): " + collateral.Name,
			Module:      "Collateral",
		})
		return ctx.JSON(http.StatusCreated, c.model.CollateralManager.ToModel(collateral))
	})

	// PUT /collateral/:collateral_id: Update a collateral record by ID. (WITH footstep)
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/collateral/:collateral_id",
		Method:       "PUT",
		RequestType:  model.CollateralRequest{},
		ResponseType: model.CollateralResponse{},
		Note:         "Updates an existing collateral record by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		collateralID, err := handlers.EngineUUIDParam(ctx, "collateral_id")
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Collateral update failed (/collateral/:collateral_id), invalid collateral ID.",
				Module:      "Collateral",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid collateral ID"})
		}

		req, err := c.model.CollateralManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Collateral update failed (/collateral/:collateral_id), validation error: " + err.Error(),
				Module:      "Collateral",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid collateral data: " + err.Error()})
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Collateral update failed (/collateral/:collateral_id), user org error: " + err.Error(),
				Module:      "Collateral",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		collateral, err := c.model.CollateralManager.GetByID(context, *collateralID)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Collateral update failed (/collateral/:collateral_id), record not found.",
				Module:      "Collateral",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Collateral record not found"})
		}
		collateral.Icon = req.Icon
		collateral.Name = req.Name
		collateral.Description = req.Description
		collateral.UpdatedAt = time.Now().UTC()
		collateral.UpdatedByID = user.UserID
		if err := c.model.CollateralManager.UpdateFields(context, collateral.ID, collateral); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Collateral update failed (/collateral/:collateral_id), db error: " + err.Error(),
				Module:      "Collateral",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update collateral record: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Updated collateral (/collateral/:collateral_id): " + collateral.Name,
			Module:      "Collateral",
		})
		return ctx.JSON(http.StatusOK, c.model.CollateralManager.ToModel(collateral))
	})

	// DELETE /collateral/:collateral_id: Delete a collateral record by ID. (WITH footstep)
	req.RegisterRoute(handlers.Route{
		Route:  "/api/v1/collateral/:collateral_id",
		Method: "DELETE",
		Note:   "Deletes the specified collateral record by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		collateralID, err := handlers.EngineUUIDParam(ctx, "collateral_id")
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Collateral delete failed (/collateral/:collateral_id), invalid collateral ID.",
				Module:      "Collateral",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid collateral ID"})
		}
		collateral, err := c.model.CollateralManager.GetByID(context, *collateralID)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Collateral delete failed (/collateral/:collateral_id), record not found.",
				Module:      "Collateral",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Collateral record not found"})
		}
		if err := c.model.CollateralManager.DeleteByID(context, *collateralID); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Collateral delete failed (/collateral/:collateral_id), db error: " + err.Error(),
				Module:      "Collateral",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete collateral record: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "delete-success",
			Description: "Deleted collateral (/collateral/:collateral_id): " + collateral.Name,
			Module:      "Collateral",
		})
		return ctx.NoContent(http.StatusNoContent)
	})

	// DELETE /collateral/bulk-delete: Bulk delete collateral records by IDs. (WITH footstep)
	req.RegisterRoute(handlers.Route{
		Route:       "/api/v1/collateral/bulk-delete",
		Method:      "DELETE",
		Note:        "Deletes multiple collateral records by their IDs. Expects a JSON body: { \"ids\": [\"id1\", \"id2\", ...] }",
		RequestType: model.IDSRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var reqBody model.IDSRequest
		if err := ctx.Bind(&reqBody); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Collateral bulk delete failed (/collateral/bulk-delete), invalid request body.",
				Module:      "Collateral",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
		}
		if len(reqBody.IDs) == 0 {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Collateral bulk delete failed (/collateral/bulk-delete), no IDs provided.",
				Module:      "Collateral",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "No IDs provided for bulk delete"})
		}
		tx := c.provider.Service.Database.Client().Begin()
		if tx.Error != nil {
			tx.Rollback()
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Collateral bulk delete failed (/collateral/bulk-delete), begin tx error: " + tx.Error.Error(),
				Module:      "Collateral",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to start database transaction: " + tx.Error.Error()})
		}
		names := ""
		for _, rawID := range reqBody.IDs {
			collateralID, err := uuid.Parse(rawID)
			if err != nil {
				tx.Rollback()
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "bulk-delete-error",
					Description: "Collateral bulk delete failed (/collateral/bulk-delete), invalid UUID: " + rawID,
					Module:      "Collateral",
				})
				return ctx.JSON(http.StatusBadRequest, map[string]string{"error": fmt.Sprintf("Invalid UUID: %s", rawID)})
			}
			collateral, err := c.model.CollateralManager.GetByID(context, collateralID)
			if err != nil {
				tx.Rollback()
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "bulk-delete-error",
					Description: "Collateral bulk delete failed (/collateral/bulk-delete), record not found: " + rawID,
					Module:      "Collateral",
				})
				return ctx.JSON(http.StatusNotFound, map[string]string{"error": fmt.Sprintf("Collateral record not found with ID: %s", rawID)})
			}
			names += collateral.Name + ","
			if err := c.model.CollateralManager.DeleteByIDWithTx(context, tx, collateralID); err != nil {
				tx.Rollback()
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "bulk-delete-error",
					Description: "Collateral bulk delete failed (/collateral/bulk-delete), db error: " + err.Error(),
					Module:      "Collateral",
				})
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete collateral record: " + err.Error()})
			}
		}
		if err := tx.Commit().Error; err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Collateral bulk delete failed (/collateral/bulk-delete), commit error: " + err.Error(),
				Module:      "Collateral",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to commit bulk delete transaction: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "bulk-delete-success",
			Description: "Bulk deleted collaterals (/collateral/bulk-delete): " + names,
			Module:      "Collateral",
		})
		return ctx.NoContent(http.StatusNoContent)
	})
}
