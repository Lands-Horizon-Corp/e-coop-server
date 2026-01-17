package transactions

import (
	"fmt"
	"net/http"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/helpers"
	"github.com/Lands-Horizon-Corp/e-coop-server/horizon"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/event"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/types"
	"github.com/labstack/echo/v4"
)

func AdjustmentTagController(service *horizon.HorizonService) {
	req := service.API

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/adjustment-tag",
		Method:       "GET",
		Note:         "Returns all adjustment tags for the current user's organization and branch. Returns empty if not authenticated.",
		ResponseType: types.AdjustmentTagResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if userOrg.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		tags, err := core.AdjustmentTagCurrentBranch(context, service, userOrg.OrganizationID, *userOrg.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "No adjustment tags found for the current branch"})
		}
		return ctx.JSON(http.StatusOK, core.AdjustmentTagManager(service).ToModels(tags))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/adjustment-tag/search",
		Method:       "GET",
		Note:         "Returns a paginated list of adjustment tags for the current user's organization and branch.",
		ResponseType: types.AdjustmentTagResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if userOrg.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		tags, err := core.AdjustmentTagManager(service).NormalPagination(context, ctx, &types.AdjustmentTag{
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch adjustment tags for pagination: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, tags)
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/adjustment-tag/:tag_id",
		Method:       "GET",
		Note:         "Returns a single adjustment tag by its ID.",
		ResponseType: types.AdjustmentTagResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		tagID, err := helpers.EngineUUIDParam(ctx, "tag_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid adjustment tag ID"})
		}
		tag, err := core.AdjustmentTagManager(service).GetByIDRaw(context, *tagID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "adjustment tag not found"})
		}
		return ctx.JSON(http.StatusOK, tag)
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/adjustment-tag",
		Method:       "POST",
		Note:         "Creates a new adjustment tag for the current user's organization and branch.",
		RequestType: types.AdjustmentTagRequest{},
		ResponseType: types.AdjustmentTagResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := core.AdjustmentTagManager(service).Validate(ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "adjustment tag creation failed (/adjustment-tag), validation error: " + err.Error(),
				Module:      "AdjustmentTag",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid adjustment tag data: " + err.Error()})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "adjustment tag creation failed (/adjustment-tag), user org error: " + err.Error(),
				Module:      "AdjustmentTag",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if userOrg.BranchID == nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "adjustment tag creation failed (/adjustment-tag), user not assigned to branch.",
				Module:      "AdjustmentTag",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}

		tag := &types.AdjustmentTag{
			AdjustmentEntryID: req.AdjustmentEntryID,
			Name:              req.Name,
			Description:       req.Description,
			Category:          req.Category,
			Color:             req.Color,
			Icon:              req.Icon,
			CreatedAt:         time.Now().UTC(),
			CreatedByID:       userOrg.UserID,
			UpdatedAt:         time.Now().UTC(),
			UpdatedByID:       userOrg.UserID,
			BranchID:          *userOrg.BranchID,
			OrganizationID:    userOrg.OrganizationID,
		}

		if err := core.AdjustmentTagManager(service).Create(context, tag); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "adjustment tag creation failed (/adjustment-tag), db error: " + err.Error(),
				Module:      "AdjustmentTag",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create adjustment tag: " + err.Error()})
		}
		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "create-success",
			Description: "Created adjustment tag (/adjustment-tag): " + tag.Name,
			Module:      "AdjustmentTag",
		})

		event.OrganizationAdminsNotification(ctx, service, event.NotificationEvent{
			Description:      fmt.Sprintf("Journal vouchers approved list has been accessed by %s", userOrg.User.FullName),
			Title:            "Journal Vouchers - Approved List Accessed",
			NotificationType: core.NotificationSystem,
		})
		return ctx.JSON(http.StatusCreated, core.AdjustmentTagManager(service).ToModel(tag))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/adjustment-tag/adjustment-entry/:adjustment_entry_id",
		Method:       "GET",
		Note:         "Returns all adjustment tags for the given adjustment entry ID.",
		ResponseType: types.AdjustmentTagResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		adjustmentEntryID, err := helpers.EngineUUIDParam(ctx, "adjustment_entry_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid adjustment entry ID"})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if userOrg.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}

		tags, err := core.AdjustmentTagManager(service).Find(context, &types.AdjustmentTag{
			AdjustmentEntryID: adjustmentEntryID,
			OrganizationID:    userOrg.OrganizationID,
			BranchID:          *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "No adjustment tags found for the given adjustment entry ID"})
		}
		return ctx.JSON(http.StatusOK, core.AdjustmentTagManager(service).ToModels(tags))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/adjustment-tag/:tag_id",
		Method:       "PUT",
		Note:         "Updates an existing adjustment tag by its ID.",
		RequestType: types.AdjustmentTagRequest{},
		ResponseType: types.AdjustmentTagResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		tagID, err := helpers.EngineUUIDParam(ctx, "tag_id")
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "adjustment tag update failed (/adjustment-tag/:tag_id), invalid tag ID.",
				Module:      "AdjustmentTag",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid adjustment tag ID"})
		}

		req, err := core.AdjustmentTagManager(service).Validate(ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "adjustment tag update failed (/adjustment-tag/:tag_id), validation error: " + err.Error(),
				Module:      "AdjustmentTag",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid adjustment tag data: " + err.Error()})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "adjustment tag update failed (/adjustment-tag/:tag_id), user org error: " + err.Error(),
				Module:      "AdjustmentTag",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		tag, err := core.AdjustmentTagManager(service).GetByID(context, *tagID)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
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
		tag.UpdatedByID = userOrg.UserID
		if err := core.AdjustmentTagManager(service).UpdateByID(context, tag.ID, tag); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "adjustment tag update failed (/adjustment-tag/:tag_id), db error: " + err.Error(),
				Module:      "AdjustmentTag",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update adjustment tag: " + err.Error()})
		}
		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Updated adjustment tag (/adjustment-tag/:tag_id): " + tag.Name,
			Module:      "AdjustmentTag",
		})
		return ctx.JSON(http.StatusOK, core.AdjustmentTagManager(service).ToModel(tag))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:  "/api/v1/adjustment-tag/:tag_id",
		Method: "DELETE",
		Note:   "Deletes the specified adjustment tag by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		tagID, err := helpers.EngineUUIDParam(ctx, "tag_id")
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "adjustment tag delete failed (/adjustment-tag/:tag_id), invalid tag ID.",
				Module:      "AdjustmentTag",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid adjustment tag ID"})
		}
		tag, err := core.AdjustmentTagManager(service).GetByID(context, *tagID)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "adjustment tag delete failed (/adjustment-tag/:tag_id), not found.",
				Module:      "AdjustmentTag",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "adjustment tag not found"})
		}
		if err := core.AdjustmentTagManager(service).Delete(context, *tagID); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "adjustment tag delete failed (/adjustment-tag/:tag_id), db error: " + err.Error(),
				Module:      "AdjustmentTag",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete adjustment tag: " + err.Error()})
		}
		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "delete-success",
			Description: "Deleted adjustment tag (/adjustment-tag/:tag_id): " + tag.Name,
			Module:      "AdjustmentTag",
		})
		return ctx.NoContent(http.StatusNoContent)
	})

	req.RegisterWebRoute(horizon.Route{
		Route:       "/api/v1/adjustment-tag/bulk-delete",
		Method:      "DELETE",
		Note:        "Deletes multiple adjustment tags by their IDs. Expects a JSON body: { \"ids\": [\"id1\", \"id2\", ...] }",
		RequestType: types.IDSRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var reqBody core.IDSRequest
		if err := ctx.Bind(&reqBody); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Failed bulk delete adjustment tags (/adjustment-tag/bulk-delete) | invalid request body: " + err.Error(),
				Module:      "AdjustmentTag",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body: " + err.Error()})
		}
		if len(reqBody.IDs) == 0 {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Failed bulk delete adjustment tags (/adjustment-tag/bulk-delete) | no IDs provided",
				Module:      "AdjustmentTag",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "No adjustment tag IDs provided for bulk delete"})
		}
		ids := make([]any, len(reqBody.IDs))
		for i, id := range reqBody.IDs {
			ids[i] = id
		}
		if err := core.AdjustmentTagManager(service).BulkDelete(context, ids); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Failed bulk delete adjustment tags (/adjustment-tag/bulk-delete) | error: " + err.Error(),
				Module:      "AdjustmentTag",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to bulk delete adjustment tags: " + err.Error()})
		}

		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "bulk-delete-success",
			Description: "Bulk deleted adjustment tags (/adjustment-tag/bulk-delete)",
			Module:      "AdjustmentTag",
		})
		return ctx.NoContent(http.StatusNoContent)
	})
}
