package journal

import (
	"net/http"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/helpers"
	"github.com/Lands-Horizon-Corp/e-coop-server/horizon"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/event"
	"github.com/labstack/echo/v4"
)

func JournalVoucherTagController(service *horizon.HorizonService) {
	req := service.API

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/journal-voucher-tag",
		Method:       "GET",
		Note:         "Returns all journal voucher tags for the current user's organization and branch. Returns empty if not authenticated.",
		ResponseType: core.JournalVoucherTagResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if userOrg.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		tags, err := core.JournalVoucherTagCurrentBranch(context, service, userOrg.OrganizationID, *userOrg.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "No journal voucher tags found for the current branch"})
		}
		return ctx.JSON(http.StatusOK, core.JournalVoucherTagManager(service).ToModels(tags))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/journal-voucher-tag/search",
		Method:       "GET",
		Note:         "Returns a paginated list of journal voucher tags for the current user's organization and branch.",
		ResponseType: core.JournalVoucherTagResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if userOrg.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		tags, err := core.JournalVoucherTagManager(service).NormalPagination(context, ctx, &core.JournalVoucherTag{
			BranchID:       *userOrg.BranchID,
			OrganizationID: userOrg.OrganizationID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch journal voucher tags for pagination: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, tags)
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/journal-voucher-tag/:tag_id",
		Method:       "GET",
		Note:         "Returns a single journal voucher tag by its ID.",
		ResponseType: core.JournalVoucherTagResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		tagID, err := helpers.EngineUUIDParam(ctx, "tag_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid journal voucher tag ID"})
		}
		tag, err := core.JournalVoucherTagManager(service).GetByIDRaw(context, *tagID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Journal voucher tag not found"})
		}
		return ctx.JSON(http.StatusOK, tag)
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/journal-voucher-tag",
		Method:       "POST",
		Note:         "Creates a new journal voucher tag for the current user's organization and branch.",
		RequestType:  core.JournalVoucherTagRequest{},
		ResponseType: core.JournalVoucherTagResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := core.JournalVoucherTagManager(service).Validate(ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Journal voucher tag creation failed (/journal-voucher-tag), validation error: " + err.Error(),
				Module:      "JournalVoucherTag",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid journal voucher tag data: " + err.Error()})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Journal voucher tag creation failed (/journal-voucher-tag), user org error: " + err.Error(),
				Module:      "JournalVoucherTag",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if userOrg.BranchID == nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Journal voucher tag creation failed (/journal-voucher-tag), user not assigned to branch.",
				Module:      "JournalVoucherTag",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}

		tag := &core.JournalVoucherTag{
			JournalVoucherID: req.JournalVoucherID,
			Name:             req.Name,
			Description:      req.Description,
			Category:         req.Category,
			Color:            req.Color,
			Icon:             req.Icon,
			CreatedAt:        time.Now().UTC(),
			CreatedByID:      userOrg.UserID,
			UpdatedAt:        time.Now().UTC(),
			UpdatedByID:      userOrg.UserID,
			BranchID:         *userOrg.BranchID,
			OrganizationID:   userOrg.OrganizationID,
		}

		if err := core.JournalVoucherTagManager(service).Create(context, tag); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Journal voucher tag creation failed (/journal-voucher-tag), db error: " + err.Error(),
				Module:      "JournalVoucherTag",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create journal voucher tag: " + err.Error()})
		}
		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "create-success",
			Description: "Created journal voucher tag (/journal-voucher-tag): " + tag.Name,
			Module:      "JournalVoucherTag",
		})
		return ctx.JSON(http.StatusCreated, core.JournalVoucherTagManager(service).ToModel(tag))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/journal-voucher-tag/:tag_id",
		Method:       "PUT",
		Note:         "Updates an existing journal voucher tag by its ID.",
		RequestType:  core.JournalVoucherTagRequest{},
		ResponseType: core.JournalVoucherTagResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		tagID, err := helpers.EngineUUIDParam(ctx, "tag_id")
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Journal voucher tag update failed (/journal-voucher-tag/:tag_id), invalid tag ID.",
				Module:      "JournalVoucherTag",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid journal voucher tag ID"})
		}

		req, err := core.JournalVoucherTagManager(service).Validate(ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Journal voucher tag update failed (/journal-voucher-tag/:tag_id), validation error: " + err.Error(),
				Module:      "JournalVoucherTag",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid journal voucher tag data: " + err.Error()})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Journal voucher tag update failed (/journal-voucher-tag/:tag_id), user org error: " + err.Error(),
				Module:      "JournalVoucherTag",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		tag, err := core.JournalVoucherTagManager(service).GetByID(context, *tagID)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
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
		tag.UpdatedByID = userOrg.UserID
		if err := core.JournalVoucherTagManager(service).UpdateByID(context, tag.ID, tag); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Journal voucher tag update failed (/journal-voucher-tag/:tag_id), db error: " + err.Error(),
				Module:      "JournalVoucherTag",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update journal voucher tag: " + err.Error()})
		}
		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Updated journal voucher tag (/journal-voucher-tag/:tag_id): " + tag.Name,
			Module:      "JournalVoucherTag",
		})
		return ctx.JSON(http.StatusOK, core.JournalVoucherTagManager(service).ToModel(tag))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/journal-voucher-tag/journal-voucher/:journal_voucher_id",
		Method:       "GET",
		Note:         "Returns all journal voucher tags associated with the specified journal voucher ID.",
		ResponseType: core.JournalVoucherTagResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		journalVoucherID, err := helpers.EngineUUIDParam(ctx, "journal_voucher_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid journal voucher ID"})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if userOrg.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}

		tags, err := core.JournalVoucherTagManager(service).Find(context, &core.JournalVoucherTag{
			JournalVoucherID: journalVoucherID,
			OrganizationID:   userOrg.OrganizationID,
			BranchID:         *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "No journal voucher tags found for the given journal voucher ID"})
		}
		return ctx.JSON(http.StatusOK, core.JournalVoucherTagManager(service).ToModels(tags))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:  "/api/v1/journal-voucher-tag/:tag_id",
		Method: "DELETE",
		Note:   "Deletes the specified journal voucher tag by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		tagID, err := helpers.EngineUUIDParam(ctx, "tag_id")
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Journal voucher tag delete failed (/journal-voucher-tag/:tag_id), invalid tag ID.",
				Module:      "JournalVoucherTag",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid journal voucher tag ID"})
		}
		tag, err := core.JournalVoucherTagManager(service).GetByID(context, *tagID)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Journal voucher tag delete failed (/journal-voucher-tag/:tag_id), not found.",
				Module:      "JournalVoucherTag",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Journal voucher tag not found"})
		}
		if err := core.JournalVoucherTagManager(service).Delete(context, *tagID); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Journal voucher tag delete failed (/journal-voucher-tag/:tag_id), db error: " + err.Error(),
				Module:      "JournalVoucherTag",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete journal voucher tag: " + err.Error()})
		}
		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "delete-success",
			Description: "Deleted journal voucher tag (/journal-voucher-tag/:tag_id): " + tag.Name,
			Module:      "JournalVoucherTag",
		})
		return ctx.NoContent(http.StatusNoContent)
	})

	req.RegisterWebRoute(horizon.Route{
		Route:       "/api/v1/journal-voucher-tag/bulk-delete",
		Method:      "DELETE",
		Note:        "Deletes multiple journal voucher tags by their IDs. Expects a JSON body: { \"ids\": [\"id1\", \"id2\", ...] }",
		RequestType: core.IDSRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var reqBody core.IDSRequest

		if err := ctx.Bind(&reqBody); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete failed (/journal-voucher-tag/bulk-delete) | invalid request body: " + err.Error(),
				Module:      "JournalVoucherTag",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body: " + err.Error()})
		}

		if len(reqBody.IDs) == 0 {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete failed (/journal-voucher-tag/bulk-delete) | no IDs provided",
				Module:      "JournalVoucherTag",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "No IDs provided for bulk delete"})
		}
		ids := make([]any, len(reqBody.IDs))
		for i, id := range reqBody.IDs {
			ids[i] = id
		}
		if err := core.JournalVoucherTagManager(service).BulkDelete(context, ids); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete failed (/journal-voucher-tag/bulk-delete) | error: " + err.Error(),
				Module:      "JournalVoucherTag",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to bulk delete journal voucher tags: " + err.Error()})
		}

		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "bulk-delete-success",
			Description: "Bulk deleted journal voucher tags (/journal-voucher-tag/bulk-delete)",
			Module:      "JournalVoucherTag",
		})

		return ctx.NoContent(http.StatusNoContent)
	})

}
