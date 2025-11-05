package v1

import (
	"net/http"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/server/event"
	"github.com/Lands-Horizon-Corp/e-coop-server/server/model/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/services/handlers"
	"github.com/labstack/echo/v4"
)

// CashCheckVoucherTagController registers routes for managing cash check voucher tags.
func (c *Controller) cashCheckVoucherTagController() {
	req := c.provider.Service.Request

	// GET /cash-check-voucher-tag: List all cash check voucher tags for the current user's branch. (NO footstep)
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/cash-check-voucher-tag",
		Method:       "GET",
		Note:         "Returns all cash check voucher tags for the current user's organization and branch. Returns empty if not authenticated.",
		ResponseType: core.CashCheckVoucherTagResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if user.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		tags, err := c.core.CashCheckVoucherTagCurrentBranch(context, user.OrganizationID, *user.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "No cash check voucher tags found for the current branch"})
		}
		return ctx.JSON(http.StatusOK, c.core.CashCheckVoucherTagManager.ToModels(tags))
	})

	// GET /cash-check-voucher-tag/search: Paginated search of cash check voucher tags for the current branch. (NO footstep)
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/cash-check-voucher-tag/search",
		Method:       "GET",
		Note:         "Returns a paginated list of cash check voucher tags for the current user's organization and branch.",
		ResponseType: core.CashCheckVoucherTagResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if user.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		tags, err := c.core.CashCheckVoucherTagManager.PaginationWithFields(context, ctx, &core.CashCheckVoucherTag{
			OrganizationID: user.OrganizationID,
			BranchID:       *user.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch cash check voucher tags for pagination: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, tags)
	})

	// GET /cash-check-voucher-tag/:tag_id: Get specific cash check voucher tag by ID. (NO footstep)
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/cash-check-voucher-tag/:tag_id",
		Method:       "GET",
		Note:         "Returns a single cash check voucher tag by its ID.",
		ResponseType: core.CashCheckVoucherTagResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		tagID, err := handlers.EngineUUIDParam(ctx, "tag_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid cash check voucher tag ID"})
		}
		tag, err := c.core.CashCheckVoucherTagManager.GetByIDRaw(context, *tagID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Cash check voucher tag not found"})
		}
		return ctx.JSON(http.StatusOK, tag)
	})

	// POST /cash-check-voucher-tag: Create a new cash check voucher tag. (WITH footstep)
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/cash-check-voucher-tag",
		Method:       "POST",
		Note:         "Creates a new cash check voucher tag for the current user's organization and branch.",
		RequestType:  core.CashCheckVoucherTagRequest{},
		ResponseType: core.CashCheckVoucherTagResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := c.core.CashCheckVoucherTagManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Cash check voucher tag creation failed (/cash-check-voucher-tag), validation error: " + err.Error(),
				Module:      "CashCheckVoucherTag",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid cash check voucher tag data: " + err.Error()})
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Cash check voucher tag creation failed (/cash-check-voucher-tag), user org error: " + err.Error(),
				Module:      "CashCheckVoucherTag",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if user.BranchID == nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Cash check voucher tag creation failed (/cash-check-voucher-tag), user not assigned to branch.",
				Module:      "CashCheckVoucherTag",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}

		tag := &core.CashCheckVoucherTag{
			CashCheckVoucherID: req.CashCheckVoucherID,
			Name:               req.Name,
			Description:        req.Description,
			Category:           req.Category,
			Color:              req.Color,
			Icon:               req.Icon,
			CreatedAt:          time.Now().UTC(),
			CreatedByID:        user.UserID,
			UpdatedAt:          time.Now().UTC(),
			UpdatedByID:        user.UserID,
			BranchID:           *user.BranchID,
			OrganizationID:     user.OrganizationID,
		}

		if err := c.core.CashCheckVoucherTagManager.Create(context, tag); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Cash check voucher tag creation failed (/cash-check-voucher-tag), db error: " + err.Error(),
				Module:      "CashCheckVoucherTag",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create cash check voucher tag: " + err.Error()})
		}
		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "create-success",
			Description: "Created cash check voucher tag (/cash-check-voucher-tag): " + tag.Name,
			Module:      "CashCheckVoucherTag",
		})
		return ctx.JSON(http.StatusCreated, c.core.CashCheckVoucherTagManager.ToModel(tag))
	})

	// PUT /cash-check-voucher-tag/:tag_id: Update cash check voucher tag by ID. (WITH footstep)
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/cash-check-voucher-tag/:tag_id",
		Method:       "PUT",
		Note:         "Updates an existing cash check voucher tag by its ID.",
		RequestType:  core.CashCheckVoucherTagRequest{},
		ResponseType: core.CashCheckVoucherTagResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		tagID, err := handlers.EngineUUIDParam(ctx, "tag_id")
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Cash check voucher tag update failed (/cash-check-voucher-tag/:tag_id), invalid tag ID.",
				Module:      "CashCheckVoucherTag",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid cash check voucher tag ID"})
		}

		req, err := c.core.CashCheckVoucherTagManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Cash check voucher tag update failed (/cash-check-voucher-tag/:tag_id), validation error: " + err.Error(),
				Module:      "CashCheckVoucherTag",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid cash check voucher tag data: " + err.Error()})
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Cash check voucher tag update failed (/cash-check-voucher-tag/:tag_id), user org error: " + err.Error(),
				Module:      "CashCheckVoucherTag",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		tag, err := c.core.CashCheckVoucherTagManager.GetByID(context, *tagID)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Cash check voucher tag update failed (/cash-check-voucher-tag/:tag_id), tag not found.",
				Module:      "CashCheckVoucherTag",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Cash check voucher tag not found"})
		}
		tag.CashCheckVoucherID = req.CashCheckVoucherID
		tag.Name = req.Name
		tag.Description = req.Description
		tag.Category = req.Category
		tag.Color = req.Color
		tag.Icon = req.Icon
		tag.UpdatedAt = time.Now().UTC()
		tag.UpdatedByID = user.UserID
		if err := c.core.CashCheckVoucherTagManager.UpdateByID(context, tag.ID, tag); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Cash check voucher tag update failed (/cash-check-voucher-tag/:tag_id), db error: " + err.Error(),
				Module:      "CashCheckVoucherTag",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update cash check voucher tag: " + err.Error()})
		}
		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Updated cash check voucher tag (/cash-check-voucher-tag/:tag_id): " + tag.Name,
			Module:      "CashCheckVoucherTag",
		})
		return ctx.JSON(http.StatusOK, c.core.CashCheckVoucherTagManager.ToModel(tag))
	})

	// DELETE /cash-check-voucher-tag/:tag_id: Delete a cash check voucher tag by ID. (WITH footstep)
	req.RegisterRoute(handlers.Route{
		Route:  "/api/v1/cash-check-voucher-tag/:tag_id",
		Method: "DELETE",
		Note:   "Deletes the specified cash check voucher tag by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		tagID, err := handlers.EngineUUIDParam(ctx, "tag_id")
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Cash check voucher tag delete failed (/cash-check-voucher-tag/:tag_id), invalid tag ID.",
				Module:      "CashCheckVoucherTag",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid cash check voucher tag ID"})
		}
		tag, err := c.core.CashCheckVoucherTagManager.GetByID(context, *tagID)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Cash check voucher tag delete failed (/cash-check-voucher-tag/:tag_id), not found.",
				Module:      "CashCheckVoucherTag",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Cash check voucher tag not found"})
		}
		if err := c.core.CashCheckVoucherTagManager.Delete(context, *tagID); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Cash check voucher tag delete failed (/cash-check-voucher-tag/:tag_id), db error: " + err.Error(),
				Module:      "CashCheckVoucherTag",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete cash check voucher tag: " + err.Error()})
		}
		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "delete-success",
			Description: "Deleted cash check voucher tag (/cash-check-voucher-tag/:tag_id): " + tag.Name,
			Module:      "CashCheckVoucherTag",
		})
		return ctx.NoContent(http.StatusNoContent)
	})

	req.RegisterRoute(handlers.Route{
		Route:       "/api/v1/cash-check-voucher-tag/bulk-delete",
		Method:      "DELETE",
		Note:        "Deletes multiple cash check voucher tags by their IDs. Expects a JSON body: { \"ids\": [\"id1\", \"id2\", ...] }",
		RequestType: core.IDSRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var reqBody core.IDSRequest
		if err := ctx.Bind(&reqBody); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Failed bulk delete cash check voucher tags (/cash-check-voucher-tag/bulk-delete) | invalid request body: " + err.Error(),
				Module:      "CashCheckVoucherTag",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body: " + err.Error()})
		}
		if len(reqBody.IDs) == 0 {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Failed bulk delete cash check voucher tags (/cash-check-voucher-tag/bulk-delete) | no IDs provided",
				Module:      "CashCheckVoucherTag",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "No cash check voucher tag IDs provided for bulk delete"})
		}

		if err := c.core.CashCheckVoucherTagManager.BulkDelete(context, reqBody.IDs); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Failed bulk delete cash check voucher tags (/cash-check-voucher-tag/bulk-delete) | error: " + err.Error(),
				Module:      "CashCheckVoucherTag",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to bulk delete cash check voucher tags: " + err.Error()})
		}

		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "bulk-delete-success",
			Description: "Bulk deleted cash check voucher tags (/cash-check-voucher-tag/bulk-delete)",
			Module:      "CashCheckVoucherTag",
		})
		return ctx.NoContent(http.StatusNoContent)
	})
	// cash check voucher tag
	// GET /api/v1/cash-check-voucher-tag/cash-check-voucher/:cash_check_voucher_id
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/cash-check-voucher-tag/cash-check-voucher/:cash_check_voucher_id",
		Method:       "GET",
		Note:         "Returns all cash check voucher tags for the specified cash check voucher ID.",
		ResponseType: core.CashCheckVoucherTagResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		cashCheckVoucherID, err := handlers.EngineUUIDParam(ctx, "cash_check_voucher_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid cash check voucher ID"})
		}
		tags, err := c.core.CashCheckVoucherTagManager.Find(context, &core.CashCheckVoucherTag{
			CashCheckVoucherID: cashCheckVoucherID,
		})
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "No cash check voucher tags found for the specified cash check voucher ID"})
		}
		return ctx.JSON(http.StatusOK, c.core.CashCheckVoucherTagManager.ToModels(tags))
	})
}
