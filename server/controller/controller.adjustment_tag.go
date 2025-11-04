package v1

import (
	"net/http"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/server/event"
	"github.com/Lands-Horizon-Corp/e-coop-server/server/model/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/services/handlers"
	"github.com/labstack/echo/v4"
)

// AdjustmentTagController registers routes for managing adjustment tags.
func (c *Controller) adjustmentTagController() {
	req := c.provider.Service.Request

	// GET /adjustment-tag: List all adjustment tags for the current user's branch. (NO footstep)
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/adjustment-tag",
		Method:       "GET",
		Note:         "Returns all adjustment tags for the current user's organization and branch. Returns empty if not authenticated.",
		ResponseType: core.AdjustmentTagResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if user.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		tags, err := c.core.AdjustmentTagCurrentBranch(context, user.OrganizationID, *user.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "No adjustment tags found for the current branch"})
		}
		return ctx.JSON(http.StatusOK, c.core.AdjustmentTagManager.ToModels(tags))
	})

	// GET /adjustment-tag/search: Paginated search of adjustment tags for the current branch. (NO footstep)
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/adjustment-tag/search",
		Method:       "GET",
		Note:         "Returns a paginated list of adjustment tags for the current user's organization and branch.",
		ResponseType: core.AdjustmentTagResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if user.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		tags, err := c.core.AdjustmentTagManager.PaginationWithFields(context, ctx, &core.AdjustmentTag{
			OrganizationID: user.OrganizationID,
			BranchID:       *user.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch adjustment tags for pagination: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, tags)
	})

	// GET /adjustment-tag/:tag_id: Get specific adjustment tag by ID. (NO footstep)
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/adjustment-tag/:tag_id",
		Method:       "GET",
		Note:         "Returns a single adjustment tag by its ID.",
		ResponseType: core.AdjustmentTagResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		tagID, err := handlers.EngineUUIDParam(ctx, "tag_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid adjustment tag ID"})
		}
		tag, err := c.core.AdjustmentTagManager.GetByIDRaw(context, *tagID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "adjustment tag not found"})
		}
		return ctx.JSON(http.StatusOK, tag)
	})

	// POST /adjustment-tag: Create a new adjustment tag. (WITH footstep)
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/adjustment-tag",
		Method:       "POST",
		Note:         "Creates a new adjustment tag for the current user's organization and branch.",
		RequestType:  core.AdjustmentTagRequest{},
		ResponseType: core.AdjustmentTagResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := c.core.AdjustmentTagManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "adjustment tag creation failed (/adjustment-tag), validation error: " + err.Error(),
				Module:      "AdjustmentTag",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid adjustment tag data: " + err.Error()})
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "adjustment tag creation failed (/adjustment-tag), user org error: " + err.Error(),
				Module:      "AdjustmentTag",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if user.BranchID == nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "adjustment tag creation failed (/adjustment-tag), user not assigned to branch.",
				Module:      "AdjustmentTag",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}

		tag := &core.AdjustmentTag{
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

		if err := c.core.AdjustmentTagManager.Create(context, tag); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "adjustment tag creation failed (/adjustment-tag), db error: " + err.Error(),
				Module:      "AdjustmentTag",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create adjustment tag: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "create-success",
			Description: "Created adjustment tag (/adjustment-tag): " + tag.Name,
			Module:      "AdjustmentTag",
		})
		return ctx.JSON(http.StatusCreated, c.core.AdjustmentTagManager.ToModel(tag))
	})

	// "/api/v1/adjustment-tag/adjustment-entry/:adjustment_entry_id",
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/adjustment-tag/adjustment-entry/:adjustment_entry_id",
		Method:       "GET",
		Note:         "Returns all adjustment tags for the given adjustment entry ID.",
		ResponseType: core.AdjustmentTagResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		adjustmentEntryID, err := handlers.EngineUUIDParam(ctx, "adjustment_entry_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid adjustment entry ID"})
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if user.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}

		tags, err := c.core.AdjustmentTagManager.Find(context, &core.AdjustmentTag{
			AdjustmentEntryID: adjustmentEntryID,
			OrganizationID:    user.OrganizationID,
			BranchID:          *user.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "No adjustment tags found for the given adjustment entry ID"})
		}
		return ctx.JSON(http.StatusOK, c.core.AdjustmentTagManager.ToModels(tags))
	})

	// PUT /adjustment-tag/:tag_id: Update adjustment tag by ID. (WITH footstep)
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/adjustment-tag/:tag_id",
		Method:       "PUT",
		Note:         "Updates an existing adjustment tag by its ID.",
		RequestType:  core.AdjustmentTagRequest{},
		ResponseType: core.AdjustmentTagResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		tagID, err := handlers.EngineUUIDParam(ctx, "tag_id")
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "adjustment tag update failed (/adjustment-tag/:tag_id), invalid tag ID.",
				Module:      "AdjustmentTag",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid adjustment tag ID"})
		}

		req, err := c.core.AdjustmentTagManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "adjustment tag update failed (/adjustment-tag/:tag_id), validation error: " + err.Error(),
				Module:      "AdjustmentTag",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid adjustment tag data: " + err.Error()})
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "adjustment tag update failed (/adjustment-tag/:tag_id), user org error: " + err.Error(),
				Module:      "AdjustmentTag",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		tag, err := c.core.AdjustmentTagManager.GetByID(context, *tagID)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "adjustment tag update failed (/adjustment-tag/:tag_id), tag not found.",
				Module:      "AdjustmentTag",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "adjustment tag not found"})
		}
		tag.AdjustmentEntryID = req.AdjustmentEntryID
		tag.Name = req.Name
		tag.Description = req.Description
		tag.Category = req.Category
		tag.Color = req.Color
		tag.Icon = req.Icon
		tag.UpdatedAt = time.Now().UTC()
		tag.UpdatedByID = user.UserID
		if err := c.core.AdjustmentTagManager.UpdateByID(context, tag.ID, tag); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "adjustment tag update failed (/adjustment-tag/:tag_id), db error: " + err.Error(),
				Module:      "AdjustmentTag",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update adjustment tag: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Updated adjustment tag (/adjustment-tag/:tag_id): " + tag.Name,
			Module:      "AdjustmentTag",
		})
		return ctx.JSON(http.StatusOK, c.core.AdjustmentTagManager.ToModel(tag))
	})

	// DELETE /adjustment-tag/:tag_id: Delete an adjustment tag by ID. (WITH footstep)
	req.RegisterRoute(handlers.Route{
		Route:  "/api/v1/adjustment-tag/:tag_id",
		Method: "DELETE",
		Note:   "Deletes the specified adjustment tag by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		tagID, err := handlers.EngineUUIDParam(ctx, "tag_id")
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "adjustment tag delete failed (/adjustment-tag/:tag_id), invalid tag ID.",
				Module:      "AdjustmentTag",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid adjustment tag ID"})
		}
		tag, err := c.core.AdjustmentTagManager.GetByID(context, *tagID)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "adjustment tag delete failed (/adjustment-tag/:tag_id), not found.",
				Module:      "AdjustmentTag",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "adjustment tag not found"})
		}
		if err := c.core.AdjustmentTagManager.Delete(context, *tagID); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "adjustment tag delete failed (/adjustment-tag/:tag_id), db error: " + err.Error(),
				Module:      "AdjustmentTag",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete adjustment tag: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "delete-success",
			Description: "Deleted adjustment tag (/adjustment-tag/:tag_id): " + tag.Name,
			Module:      "AdjustmentTag",
		})
		return ctx.NoContent(http.StatusNoContent)
	})

	req.RegisterRoute(handlers.Route{
		Route:       "/api/v1/adjustment-tag/bulk-delete",
		Method:      "DELETE",
		Note:        "Deletes multiple adjustment tags by their IDs. Expects a JSON body: { \"ids\": [\"id1\", \"id2\", ...] }",
		RequestType: core.IDSRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var reqBody core.IDSRequest
		if err := ctx.Bind(&reqBody); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Failed bulk delete adjustment tags (/adjustment-tag/bulk-delete) | invalid request body: " + err.Error(),
				Module:      "AdjustmentTag",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body: " + err.Error()})
		}
		if len(reqBody.IDs) == 0 {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Failed bulk delete adjustment tags (/adjustment-tag/bulk-delete) | no IDs provided",
				Module:      "AdjustmentTag",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "No adjustment tag IDs provided for bulk delete"})
		}

		if err := c.core.AdjustmentTagManager.BulkDelete(context, reqBody.IDs); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Failed bulk delete adjustment tags (/adjustment-tag/bulk-delete) | error: " + err.Error(),
				Module:      "AdjustmentTag",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to bulk delete adjustment tags: " + err.Error()})
		}

		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "bulk-delete-success",
			Description: "Bulk deleted adjustment tags (/adjustment-tag/bulk-delete)",
			Module:      "AdjustmentTag",
		})
		return ctx.NoContent(http.StatusNoContent)
	})
}
