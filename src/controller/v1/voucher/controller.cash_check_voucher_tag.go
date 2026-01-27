package voucher

import (
	"net/http"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/helpers"
	"github.com/Lands-Horizon-Corp/e-coop-server/horizon"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/db/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/event"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/types"
	"github.com/labstack/echo/v4"
)

func CashCheckVoucherTagController(service *horizon.HorizonService) {

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/cash-check-voucher-tag",
		Method:       "GET",
		Note:         "Returns all cash check voucher tags for the current user's organization and branch. Returns empty if not authenticated.",
		ResponseType: types.CashCheckVoucherTagResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if userOrg.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		tags, err := core.CashCheckVoucherTagCurrentBranch(context, service, userOrg.OrganizationID, *userOrg.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "No cash check voucher tags found for the current branch"})
		}
		return ctx.JSON(http.StatusOK, core.CashCheckVoucherTagManager(service).ToModels(tags))
	})

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/cash-check-voucher-tag/search",
		Method:       "GET",
		Note:         "Returns a paginated list of cash check voucher tags for the current user's organization and branch.",
		ResponseType: types.CashCheckVoucherTagResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if userOrg.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		tags, err := core.CashCheckVoucherTagManager(service).NormalPagination(context, ctx, &types.CashCheckVoucherTag{
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch cash check voucher tags for pagination: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, tags)
	})

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/cash-check-voucher-tag/:tag_id",
		Method:       "GET",
		Note:         "Returns a single cash check voucher tag by its ID.",
		ResponseType: types.CashCheckVoucherTagResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		tagID, err := helpers.EngineUUIDParam(ctx, "tag_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid cash check voucher tag ID"})
		}
		tag, err := core.CashCheckVoucherTagManager(service).GetByIDRaw(context, *tagID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Cash check voucher tag not found"})
		}
		return ctx.JSON(http.StatusOK, tag)
	})

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/cash-check-voucher-tag",
		Method:       "POST",
		Note:         "Creates a new cash check voucher tag for the current user's organization and branch.",
		RequestType:  types.CashCheckVoucherTagRequest{},
		ResponseType: types.CashCheckVoucherTagResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := core.CashCheckVoucherTagManager(service).Validate(ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Cash check voucher tag creation failed (/cash-check-voucher-tag), validation error: " + err.Error(),
				Module:      "CashCheckVoucherTag",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid cash check voucher tag data: " + err.Error()})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Cash check voucher tag creation failed (/cash-check-voucher-tag), user org error: " + err.Error(),
				Module:      "CashCheckVoucherTag",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if userOrg.BranchID == nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Cash check voucher tag creation failed (/cash-check-voucher-tag), user not assigned to branch.",
				Module:      "CashCheckVoucherTag",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}

		tag := &types.CashCheckVoucherTag{
			CashCheckVoucherID: req.CashCheckVoucherID,
			Name:               req.Name,
			Description:        req.Description,
			Category:           req.Category,
			Color:              req.Color,
			Icon:               req.Icon,
			CreatedAt:          time.Now().UTC(),
			CreatedByID:        userOrg.UserID,
			UpdatedAt:          time.Now().UTC(),
			UpdatedByID:        userOrg.UserID,
			BranchID:           *userOrg.BranchID,
			OrganizationID:     userOrg.OrganizationID,
		}

		if err := core.CashCheckVoucherTagManager(service).Create(context, tag); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Cash check voucher tag creation failed (/cash-check-voucher-tag), db error: " + err.Error(),
				Module:      "CashCheckVoucherTag",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create cash check voucher tag: " + err.Error()})
		}
		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "create-success",
			Description: "Created cash check voucher tag (/cash-check-voucher-tag): " + tag.Name,
			Module:      "CashCheckVoucherTag",
		})
		return ctx.JSON(http.StatusCreated, core.CashCheckVoucherTagManager(service).ToModel(tag))
	})

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/cash-check-voucher-tag/:tag_id",
		Method:       "PUT",
		Note:         "Updates an existing cash check voucher tag by its ID.",
		RequestType:  types.CashCheckVoucherTagRequest{},
		ResponseType: types.CashCheckVoucherTagResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		tagID, err := helpers.EngineUUIDParam(ctx, "tag_id")
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Cash check voucher tag update failed (/cash-check-voucher-tag/:tag_id), invalid tag ID.",
				Module:      "CashCheckVoucherTag",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid cash check voucher tag ID"})
		}

		req, err := core.CashCheckVoucherTagManager(service).Validate(ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Cash check voucher tag update failed (/cash-check-voucher-tag/:tag_id), validation error: " + err.Error(),
				Module:      "CashCheckVoucherTag",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid cash check voucher tag data: " + err.Error()})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Cash check voucher tag update failed (/cash-check-voucher-tag/:tag_id), user org error: " + err.Error(),
				Module:      "CashCheckVoucherTag",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		tag, err := core.CashCheckVoucherTagManager(service).GetByID(context, *tagID)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
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
		tag.UpdatedByID = userOrg.UserID
		if err := core.CashCheckVoucherTagManager(service).UpdateByID(context, tag.ID, tag); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Cash check voucher tag update failed (/cash-check-voucher-tag/:tag_id), db error: " + err.Error(),
				Module:      "CashCheckVoucherTag",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update cash check voucher tag: " + err.Error()})
		}
		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Updated cash check voucher tag (/cash-check-voucher-tag/:tag_id): " + tag.Name,
			Module:      "CashCheckVoucherTag",
		})
		return ctx.JSON(http.StatusOK, core.CashCheckVoucherTagManager(service).ToModel(tag))
	})

	service.API.RegisterWebRoute(horizon.Route{
		Route:  "/api/v1/cash-check-voucher-tag/:tag_id",
		Method: "DELETE",
		Note:   "Deletes the specified cash check voucher tag by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		tagID, err := helpers.EngineUUIDParam(ctx, "tag_id")
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Cash check voucher tag delete failed (/cash-check-voucher-tag/:tag_id), invalid tag ID.",
				Module:      "CashCheckVoucherTag",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid cash check voucher tag ID"})
		}
		tag, err := core.CashCheckVoucherTagManager(service).GetByID(context, *tagID)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Cash check voucher tag delete failed (/cash-check-voucher-tag/:tag_id), not found.",
				Module:      "CashCheckVoucherTag",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Cash check voucher tag not found"})
		}
		if err := core.CashCheckVoucherTagManager(service).Delete(context, *tagID); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Cash check voucher tag delete failed (/cash-check-voucher-tag/:tag_id), db error: " + err.Error(),
				Module:      "CashCheckVoucherTag",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete cash check voucher tag: " + err.Error()})
		}
		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "delete-success",
			Description: "Deleted cash check voucher tag (/cash-check-voucher-tag/:tag_id): " + tag.Name,
			Module:      "CashCheckVoucherTag",
		})
		return ctx.NoContent(http.StatusNoContent)
	})

	service.API.RegisterWebRoute(horizon.Route{
		Route:       "/api/v1/cash-check-voucher-tag/bulk-delete",
		Method:      "DELETE",
		Note:        "Deletes multiple cash check voucher tags by their IDs. Expects a JSON body: { \"ids\": [\"id1\", \"id2\", ...] }",
		RequestType: types.IDSRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var reqBody types.IDSRequest
		if err := ctx.Bind(&reqBody); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Failed bulk delete cash check voucher tags (/cash-check-voucher-tag/bulk-delete) | invalid request body: " + err.Error(),
				Module:      "CashCheckVoucherTag",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body: " + err.Error()})
		}
		if len(reqBody.IDs) == 0 {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Failed bulk delete cash check voucher tags (/cash-check-voucher-tag/bulk-delete) | no IDs provided",
				Module:      "CashCheckVoucherTag",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "No cash check voucher tag IDs provided for bulk delete"})
		}

		ids := make([]any, len(reqBody.IDs))
		for i, id := range reqBody.IDs {
			ids[i] = id
		}
		if err := core.CashCheckVoucherTagManager(service).BulkDelete(context, ids); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Failed bulk delete cash check voucher tags (/cash-check-voucher-tag/bulk-delete) | error: " + err.Error(),
				Module:      "CashCheckVoucherTag",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to bulk delete cash check voucher tags: " + err.Error()})
		}

		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "bulk-delete-success",
			Description: "Bulk deleted cash check voucher tags (/cash-check-voucher-tag/bulk-delete)",
			Module:      "CashCheckVoucherTag",
		})
		return ctx.NoContent(http.StatusNoContent)
	})
	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/cash-check-voucher-tag/cash-check-voucher/:cash_check_voucher_id",
		Method:       "GET",
		Note:         "Returns all cash check voucher tags for the specified cash check voucher ID.",
		ResponseType: types.CashCheckVoucherTagResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		cashCheckVoucherID, err := helpers.EngineUUIDParam(ctx, "cash_check_voucher_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid cash check voucher ID"})
		}
		tags, err := core.CashCheckVoucherTagManager(service).Find(context, &types.CashCheckVoucherTag{
			CashCheckVoucherID: cashCheckVoucherID,
		})
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "No cash check voucher tags found for the specified cash check voucher ID"})
		}
		return ctx.JSON(http.StatusOK, core.CashCheckVoucherTagManager(service).ToModels(tags))
	})
}
