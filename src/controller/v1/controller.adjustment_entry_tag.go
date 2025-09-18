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

// AdjustmentEntryTagController registers routes for managing adjustment entry tags.
func (c *Controller) AdjustmentEntryTagController() {
	req := c.provider.Service.Request

	// GET /adjustment-entry-tag: List all adjustment entry tags for the current user's branch. (NO footstep)
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/adjustment-entry-tag",
		Method:       "GET",
		Note:         "Returns all adjustment entry tags for the current user's organization and branch. Returns empty if not authenticated.",
		ResponseType: model.AdjustmentEntryTagResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if user.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		adjustmentEntryTags, err := c.model.AdjustmentEntryTagCurrentBranch(context, user.OrganizationID, *user.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "No adjustment entry tags found for the current branch"})
		}
		return ctx.JSON(http.StatusOK, c.model.AdjustmentEntryTagManager.Filtered(context, ctx, adjustmentEntryTags))
	})

	// GET /adjustment-entry-tag/search: Paginated search of adjustment entry tags for the current branch. (NO footstep)
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/adjustment-entry-tag/search",
		Method:       "GET",
		Note:         "Returns a paginated list of adjustment entry tags for the current user's organization and branch.",
		ResponseType: model.AdjustmentEntryTagResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if user.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		adjustmentEntryTags, err := c.model.AdjustmentEntryTagCurrentBranch(context, user.OrganizationID, *user.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch adjustment entry tags for pagination: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.AdjustmentEntryTagManager.Pagination(context, ctx, adjustmentEntryTags))
	})

	// GET /adjustment-entry-tag/:adjustment_entry_tag_id: Get specific adjustment entry tag by ID. (NO footstep)
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/adjustment-entry-tag/:adjustment_entry_tag_id",
		Method:       "GET",
		Note:         "Returns a single adjustment entry tag by its ID.",
		ResponseType: model.AdjustmentEntryTagResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		adjustmentEntryTagID, err := handlers.EngineUUIDParam(ctx, "adjustment_entry_tag_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid adjustment entry tag ID"})
		}
		adjustmentEntryTag, err := c.model.AdjustmentEntryTagManager.GetByIDRaw(context, *adjustmentEntryTagID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Adjustment entry tag not found"})
		}
		return ctx.JSON(http.StatusOK, adjustmentEntryTag)
	})

	// POST /adjustment-entry-tag: Create a new adjustment entry tag. (WITH footstep)
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/adjustment-entry-tag",
		Method:       "POST",
		Note:         "Creates a new adjustment entry tag for the current user's organization and branch.",
		RequestType:  model.AdjustmentEntryTagRequest{},
		ResponseType: model.AdjustmentEntryTagResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := c.model.AdjustmentEntryTagManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Adjustment entry tag creation failed (/adjustment-entry-tag), validation error: " + err.Error(),
				Module:      "AdjustmentEntryTag",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid adjustment entry tag data: " + err.Error()})
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Adjustment entry tag creation failed (/adjustment-entry-tag), user org error: " + err.Error(),
				Module:      "AdjustmentEntryTag",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if user.BranchID == nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Adjustment entry tag creation failed (/adjustment-entry-tag), user not assigned to branch.",
				Module:      "AdjustmentEntryTag",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}

		adjustmentEntryTag := &model.AdjustmentEntryTag{
			AdjustmentEntryID: req.AdjustmentEntryID,
			Name:              req.Name,
			Description:       req.Description,
			Category:          req.Category,
			Color:             req.Color,
			Icon:              req.Icon,
			CreatedAt:         time.Now().UTC(),
			CreatedByID:       user.UserID,
			UpdatedAt:         time.Now().UTC(),
			UpdatedByID:       user.UserID,
			BranchID:          *user.BranchID,
			OrganizationID:    user.OrganizationID,
		}

		if err := c.model.AdjustmentEntryTagManager.Create(context, adjustmentEntryTag); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Adjustment entry tag creation failed (/adjustment-entry-tag), db error: " + err.Error(),
				Module:      "AdjustmentEntryTag",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create adjustment entry tag: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "create-success",
			Description: "Created adjustment entry tag (/adjustment-entry-tag): " + adjustmentEntryTag.Name,
			Module:      "AdjustmentEntryTag",
		})
		return ctx.JSON(http.StatusCreated, c.model.AdjustmentEntryTagManager.ToModel(adjustmentEntryTag))
	})

	// PUT /adjustment-entry-tag/:adjustment_entry_tag_id: Update adjustment entry tag by ID. (WITH footstep)
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/adjustment-entry-tag/:adjustment_entry_tag_id",
		Method:       "PUT",
		Note:         "Updates an existing adjustment entry tag by its ID.",
		RequestType:  model.AdjustmentEntryTagRequest{},
		ResponseType: model.AdjustmentEntryTagResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		adjustmentEntryTagID, err := handlers.EngineUUIDParam(ctx, "adjustment_entry_tag_id")
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Adjustment entry tag update failed (/adjustment-entry-tag/:adjustment_entry_tag_id), invalid adjustment entry tag ID.",
				Module:      "AdjustmentEntryTag",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid adjustment entry tag ID"})
		}

		req, err := c.model.AdjustmentEntryTagManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Adjustment entry tag update failed (/adjustment-entry-tag/:adjustment_entry_tag_id), validation error: " + err.Error(),
				Module:      "AdjustmentEntryTag",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid adjustment entry tag data: " + err.Error()})
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Adjustment entry tag update failed (/adjustment-entry-tag/:adjustment_entry_tag_id), user org error: " + err.Error(),
				Module:      "AdjustmentEntryTag",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		adjustmentEntryTag, err := c.model.AdjustmentEntryTagManager.GetByID(context, *adjustmentEntryTagID)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Adjustment entry tag update failed (/adjustment-entry-tag/:adjustment_entry_tag_id), adjustment entry tag not found.",
				Module:      "AdjustmentEntryTag",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Adjustment entry tag not found"})
		}
		adjustmentEntryTag.AdjustmentEntryID = req.AdjustmentEntryID
		adjustmentEntryTag.Name = req.Name
		adjustmentEntryTag.Description = req.Description
		adjustmentEntryTag.Category = req.Category
		adjustmentEntryTag.Color = req.Color
		adjustmentEntryTag.Icon = req.Icon
		adjustmentEntryTag.UpdatedAt = time.Now().UTC()
		adjustmentEntryTag.UpdatedByID = user.UserID
		if err := c.model.AdjustmentEntryTagManager.UpdateFields(context, adjustmentEntryTag.ID, adjustmentEntryTag); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Adjustment entry tag update failed (/adjustment-entry-tag/:adjustment_entry_tag_id), db error: " + err.Error(),
				Module:      "AdjustmentEntryTag",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update adjustment entry tag: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Updated adjustment entry tag (/adjustment-entry-tag/:adjustment_entry_tag_id): " + adjustmentEntryTag.Name,
			Module:      "AdjustmentEntryTag",
		})
		return ctx.JSON(http.StatusOK, c.model.AdjustmentEntryTagManager.ToModel(adjustmentEntryTag))
	})

	// DELETE /adjustment-entry-tag/:adjustment_entry_tag_id: Delete an adjustment entry tag by ID. (WITH footstep)
	req.RegisterRoute(handlers.Route{
		Route:  "/api/v1/adjustment-entry-tag/:adjustment_entry_tag_id",
		Method: "DELETE",
		Note:   "Deletes the specified adjustment entry tag by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		adjustmentEntryTagID, err := handlers.EngineUUIDParam(ctx, "adjustment_entry_tag_id")
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Adjustment entry tag delete failed (/adjustment-entry-tag/:adjustment_entry_tag_id), invalid adjustment entry tag ID.",
				Module:      "AdjustmentEntryTag",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid adjustment entry tag ID"})
		}
		adjustmentEntryTag, err := c.model.AdjustmentEntryTagManager.GetByID(context, *adjustmentEntryTagID)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Adjustment entry tag delete failed (/adjustment-entry-tag/:adjustment_entry_tag_id), not found.",
				Module:      "AdjustmentEntryTag",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Adjustment entry tag not found"})
		}
		if err := c.model.AdjustmentEntryTagManager.DeleteByID(context, *adjustmentEntryTagID); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Adjustment entry tag delete failed (/adjustment-entry-tag/:adjustment_entry_tag_id), db error: " + err.Error(),
				Module:      "AdjustmentEntryTag",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete adjustment entry tag: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "delete-success",
			Description: "Deleted adjustment entry tag (/adjustment-entry-tag/:adjustment_entry_tag_id): " + adjustmentEntryTag.Name,
			Module:      "AdjustmentEntryTag",
		})
		return ctx.NoContent(http.StatusNoContent)
	})

	// DELETE /adjustment-entry-tag/bulk-delete: Bulk delete adjustment entry tags by IDs. (WITH footstep)
	req.RegisterRoute(handlers.Route{
		Route:       "/api/v1/adjustment-entry-tag/bulk-delete",
		Method:      "DELETE",
		Note:        "Deletes multiple adjustment entry tags by their IDs. Expects a JSON body: { \"ids\": [\"id1\", \"id2\", ...] }",
		RequestType: model.IDSRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var reqBody model.IDSRequest
		if err := ctx.Bind(&reqBody); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete failed (/adjustment-entry-tag/bulk-delete), invalid request body.",
				Module:      "AdjustmentEntryTag",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
		}
		if len(reqBody.IDs) == 0 {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete failed (/adjustment-entry-tag/bulk-delete), no IDs provided.",
				Module:      "AdjustmentEntryTag",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "No adjustment entry tag IDs provided for bulk delete"})
		}
		tx := c.provider.Service.Database.Client().Begin()
		if tx.Error != nil {
			tx.Rollback()
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete failed (/adjustment-entry-tag/bulk-delete), begin tx error: " + tx.Error.Error(),
				Module:      "AdjustmentEntryTag",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to start database transaction: " + tx.Error.Error()})
		}
		names := ""
		for _, rawID := range reqBody.IDs {
			adjustmentEntryTagID, err := uuid.Parse(rawID)
			if err != nil {
				tx.Rollback()
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "bulk-delete-error",
					Description: "Bulk delete failed (/adjustment-entry-tag/bulk-delete), invalid UUID: " + rawID,
					Module:      "AdjustmentEntryTag",
				})
				return ctx.JSON(http.StatusBadRequest, map[string]string{"error": fmt.Sprintf("Invalid UUID: %s", rawID)})
			}
			adjustmentEntryTag, err := c.model.AdjustmentEntryTagManager.GetByID(context, adjustmentEntryTagID)
			if err != nil {
				tx.Rollback()
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "bulk-delete-error",
					Description: "Bulk delete failed (/adjustment-entry-tag/bulk-delete), not found: " + rawID,
					Module:      "AdjustmentEntryTag",
				})
				return ctx.JSON(http.StatusNotFound, map[string]string{"error": fmt.Sprintf("Adjustment entry tag not found with ID: %s", rawID)})
			}
			names += adjustmentEntryTag.Name + ","
			if err := c.model.AdjustmentEntryTagManager.DeleteByIDWithTx(context, tx, adjustmentEntryTagID); err != nil {
				tx.Rollback()
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "bulk-delete-error",
					Description: "Bulk delete failed (/adjustment-entry-tag/bulk-delete), db error: " + err.Error(),
					Module:      "AdjustmentEntryTag",
				})
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete adjustment entry tag: " + err.Error()})
			}
		}
		if err := tx.Commit().Error; err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete failed (/adjustment-entry-tag/bulk-delete), commit error: " + err.Error(),
				Module:      "AdjustmentEntryTag",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to commit bulk delete: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "bulk-delete-success",
			Description: "Bulk deleted adjustment entry tags (/adjustment-entry-tag/bulk-delete): " + names,
			Module:      "AdjustmentEntryTag",
		})
		return ctx.NoContent(http.StatusNoContent)
	})
}
