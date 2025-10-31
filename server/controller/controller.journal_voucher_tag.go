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

// JournalVoucherTagController registers routes for managing journal voucher tags.
func (c *Controller) journalVoucherTagController(
	req := c.provider.Service.Request

	// GET /journal-voucher-tag: List all journal voucher tags for the current user's branch. (NO footstep)
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/journal-voucher-tag",
		Method:       "GET",
		Note:         "Returns all journal voucher tags for the current user's organization and branch. Returns empty if not authenticated.",
		ResponseType: modelcore.JournalVoucherTagResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if user.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		tags, err := c.modelcore.JournalVoucherTagCurrentbranch(context, user.OrganizationID, *user.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "No journal voucher tags found for the current branch"})
		}
		return ctx.JSON(http.StatusOK, c.modelcore.JournalVoucherTagManager.Filtered(context, ctx, tags))
	})

	// GET /journal-voucher-tag/search: Paginated search of journal voucher tags for the current branch. (NO footstep)
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/journal-voucher-tag/search",
		Method:       "GET",
		Note:         "Returns a paginated list of journal voucher tags for the current user's organization and branch.",
		ResponseType: modelcore.JournalVoucherTagResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if user.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		tags, err := c.modelcore.JournalVoucherTagCurrentbranch(context, user.OrganizationID, *user.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch journal voucher tags for pagination: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.modelcore.JournalVoucherTagManager.Pagination(context, ctx, tags))
	})

	// GET /journal-voucher-tag/:tag_id: Get specific journal voucher tag by ID. (NO footstep)
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/journal-voucher-tag/:tag_id",
		Method:       "GET",
		Note:         "Returns a single journal voucher tag by its ID.",
		ResponseType: modelcore.JournalVoucherTagResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		tagID, err := handlers.EngineUUIDParam(ctx, "tag_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid journal voucher tag ID"})
		}
		tag, err := c.modelcore.JournalVoucherTagManager.GetByIDRaw(context, *tagID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Journal voucher tag not found"})
		}
		return ctx.JSON(http.StatusOK, tag)
	})

	// POST /journal-voucher-tag: Create a new journal voucher tag. (WITH footstep)
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/journal-voucher-tag",
		Method:       "POST",
		Note:         "Creates a new journal voucher tag for the current user's organization and branch.",
		RequestType:  modelcore.JournalVoucherTagRequest{},
		ResponseType: modelcore.JournalVoucherTagResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := c.modelcore.JournalVoucherTagManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Journal voucher tag creation failed (/journal-voucher-tag), validation error: " + err.Error(),
				Module:      "JournalVoucherTag",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid journal voucher tag data: " + err.Error()})
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Journal voucher tag creation failed (/journal-voucher-tag), user org error: " + err.Error(),
				Module:      "JournalVoucherTag",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if user.BranchID == nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Journal voucher tag creation failed (/journal-voucher-tag), user not assigned to branch.",
				Module:      "JournalVoucherTag",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}

		tag := &modelcore.JournalVoucherTag{
			JournalVoucherID: req.JournalVoucherID,
			Name:             req.Name,
			Description:      req.Description,
			Category:         req.Category,
			Color:            req.Color,
			Icon:             req.Icon,
			CreatedAt:        time.Now().UTC(),
			CreatedByID:      user.UserID,
			UpdatedAt:        time.Now().UTC(),
			UpdatedByID:      user.UserID,
			BranchID:         *user.BranchID,
			OrganizationID:   user.OrganizationID,
		}

		if err := c.modelcore.JournalVoucherTagManager.Create(context, tag); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Journal voucher tag creation failed (/journal-voucher-tag), db error: " + err.Error(),
				Module:      "JournalVoucherTag",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create journal voucher tag: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "create-success",
			Description: "Created journal voucher tag (/journal-voucher-tag): " + tag.Name,
			Module:      "JournalVoucherTag",
		})
		return ctx.JSON(http.StatusCreated, c.modelcore.JournalVoucherTagManager.ToModel(tag))
	})

	// PUT /journal-voucher-tag/:tag_id: Update journal voucher tag by ID. (WITH footstep)
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/journal-voucher-tag/:tag_id",
		Method:       "PUT",
		Note:         "Updates an existing journal voucher tag by its ID.",
		RequestType:  modelcore.JournalVoucherTagRequest{},
		ResponseType: modelcore.JournalVoucherTagResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		tagID, err := handlers.EngineUUIDParam(ctx, "tag_id")
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Journal voucher tag update failed (/journal-voucher-tag/:tag_id), invalid tag ID.",
				Module:      "JournalVoucherTag",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid journal voucher tag ID"})
		}

		req, err := c.modelcore.JournalVoucherTagManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Journal voucher tag update failed (/journal-voucher-tag/:tag_id), validation error: " + err.Error(),
				Module:      "JournalVoucherTag",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid journal voucher tag data: " + err.Error()})
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Journal voucher tag update failed (/journal-voucher-tag/:tag_id), user org error: " + err.Error(),
				Module:      "JournalVoucherTag",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		tag, err := c.modelcore.JournalVoucherTagManager.GetByID(context, *tagID)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Journal voucher tag update failed (/journal-voucher-tag/:tag_id), tag not found.",
				Module:      "JournalVoucherTag",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Journal voucher tag not found"})
		}
		tag.JournalVoucherID = req.JournalVoucherID
		tag.Name = req.Name
		tag.Description = req.Description
		tag.Category = req.Category
		tag.Color = req.Color
		tag.Icon = req.Icon
		tag.UpdatedAt = time.Now().UTC()
		tag.UpdatedByID = user.UserID
		if err := c.modelcore.JournalVoucherTagManager.UpdateFields(context, tag.ID, tag); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Journal voucher tag update failed (/journal-voucher-tag/:tag_id), db error: " + err.Error(),
				Module:      "JournalVoucherTag",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update journal voucher tag: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Updated journal voucher tag (/journal-voucher-tag/:tag_id): " + tag.Name,
			Module:      "JournalVoucherTag",
		})
		return ctx.JSON(http.StatusOK, c.modelcore.JournalVoucherTagManager.ToModel(tag))
	})

	// GET  "/api/v1/journal-voucher-tag/journal-voucher/:journal_voucher_id" - List journal voucher tags by journal voucher ID. (NO footstep)
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/journal-voucher-tag/journal-voucher/:journal_voucher_id",
		Method:       "GET",
		Note:         "Returns all journal voucher tags associated with the specified journal voucher ID.",
		ResponseType: modelcore.JournalVoucherTagResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		journalVoucherID, err := handlers.EngineUUIDParam(ctx, "journal_voucher_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid journal voucher ID"})
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if user.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}

		tags, err := c.modelcore.JournalVoucherTagManager.Find(context, &modelcore.JournalVoucherTag{
			JournalVoucherID: journalVoucherID,
			OrganizationID:   user.OrganizationID,
			BranchID:         *user.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "No journal voucher tags found for the given journal voucher ID"})
		}
		return ctx.JSON(http.StatusOK, c.modelcore.JournalVoucherTagManager.Filtered(context, ctx, tags))
	})

	// DELETE /journal-voucher-tag/:tag_id: Delete a journal voucher tag by ID. (WITH footstep)
	req.RegisterRoute(handlers.Route{
		Route:  "/api/v1/journal-voucher-tag/:tag_id",
		Method: "DELETE",
		Note:   "Deletes the specified journal voucher tag by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		tagID, err := handlers.EngineUUIDParam(ctx, "tag_id")
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Journal voucher tag delete failed (/journal-voucher-tag/:tag_id), invalid tag ID.",
				Module:      "JournalVoucherTag",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid journal voucher tag ID"})
		}
		tag, err := c.modelcore.JournalVoucherTagManager.GetByID(context, *tagID)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Journal voucher tag delete failed (/journal-voucher-tag/:tag_id), not found.",
				Module:      "JournalVoucherTag",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Journal voucher tag not found"})
		}
		if err := c.modelcore.JournalVoucherTagManager.DeleteByID(context, *tagID); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Journal voucher tag delete failed (/journal-voucher-tag/:tag_id), db error: " + err.Error(),
				Module:      "JournalVoucherTag",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete journal voucher tag: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "delete-success",
			Description: "Deleted journal voucher tag (/journal-voucher-tag/:tag_id): " + tag.Name,
			Module:      "JournalVoucherTag",
		})
		return ctx.NoContent(http.StatusNoContent)
	})

	// DELETE /journal-voucher-tag/bulk-delete: Bulk delete journal voucher tags by IDs. (WITH footstep)
	req.RegisterRoute(handlers.Route{
		Route:       "/api/v1/journal-voucher-tag/bulk-delete",
		Method:      "DELETE",
		Note:        "Deletes multiple journal voucher tags by their IDs. Expects a JSON body: { \"ids\": [\"id1\", \"id2\", ...] }",
		RequestType: modelcore.IDSRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var reqBody modelcore.IDSRequest
		if err := ctx.Bind(&reqBody); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete failed (/journal-voucher-tag/bulk-delete), invalid request body.",
				Module:      "JournalVoucherTag",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
		}
		if len(reqBody.IDs) == 0 {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete failed (/journal-voucher-tag/bulk-delete), no IDs provided.",
				Module:      "JournalVoucherTag",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "No journal voucher tag IDs provided for bulk delete"})
		}
		tx := c.provider.Service.Database.Client().Begin()
		if tx.Error != nil {
			tx.Rollback()
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete failed (/journal-voucher-tag/bulk-delete), begin tx error: " + tx.Error.Error(),
				Module:      "JournalVoucherTag",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to start database transaction: " + tx.Error.Error()})
		}
		names := ""
		for _, rawID := range reqBody.IDs {
			tagID, err := uuid.Parse(rawID)
			if err != nil {
				tx.Rollback()
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "bulk-delete-error",
					Description: "Bulk delete failed (/journal-voucher-tag/bulk-delete), invalid UUID: " + rawID,
					Module:      "JournalVoucherTag",
				})
				return ctx.JSON(http.StatusBadRequest, map[string]string{"error": fmt.Sprintf("Invalid UUID: %s", rawID)})
			}
			tag, err := c.modelcore.JournalVoucherTagManager.GetByID(context, tagID)
			if err != nil {
				tx.Rollback()
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "bulk-delete-error",
					Description: "Bulk delete failed (/journal-voucher-tag/bulk-delete), not found: " + rawID,
					Module:      "JournalVoucherTag",
				})
				return ctx.JSON(http.StatusNotFound, map[string]string{"error": fmt.Sprintf("Journal voucher tag not found with ID: %s", rawID)})
			}
			names += tag.Name + ","
			if err := c.modelcore.JournalVoucherTagManager.DeleteByIDWithTx(context, tx, tagID); err != nil {
				tx.Rollback()
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "bulk-delete-error",
					Description: "Bulk delete failed (/journal-voucher-tag/bulk-delete), db error: " + err.Error(),
					Module:      "JournalVoucherTag",
				})
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete journal voucher tag: " + err.Error()})
			}
		}
		if err := tx.Commit().Error; err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete failed (/journal-voucher-tag/bulk-delete), commit error: " + err.Error(),
				Module:      "JournalVoucherTag",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to commit bulk delete: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "bulk-delete-success",
			Description: "Bulk deleted journal voucher tags (/journal-voucher-tag/bulk-delete): " + names,
			Module:      "JournalVoucherTag",
		})
		return ctx.NoContent(http.StatusNoContent)
	})

}
