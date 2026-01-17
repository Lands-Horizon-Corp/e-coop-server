package member_profile

import (
	"net/http"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/helpers"
	"github.com/Lands-Horizon-Corp/e-coop-server/horizon"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/event"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/types"
	"github.com/labstack/echo/v4"
)

func MemberCenterController(service *horizon.HorizonService) {
	req := service.API

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/member-center-history",
		Method:       "GET",
		ResponseType: types.MemberCenterResponse{},
		Note:         "Returns all member center history entries for the current user's branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		memberCenterHistory, err := core.MemberCenterHistoryCurrentBranch(context, service, userOrg.OrganizationID, *userOrg.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get member center history: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, core.MemberCenterHistoryManager(service).ToModels(memberCenterHistory))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/member-center-history/member-profile/:member_profile_id/search",
		Method:       "GET",
		ResponseType: types.MemberCenterHistoryResponse{},
		Note:         "Returns member center history for a specific member profile ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberProfileID, err := helpers.EngineUUIDParam(ctx, "member_profile_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member_profile_id: " + err.Error()})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		memberCenterHistory, err := core.MemberCenterHistoryManager(service).NormalPagination(context, ctx, &types.MemberCenterHistory{
			OrganizationID:  userOrg.OrganizationID,
			BranchID:        *userOrg.BranchID,
			MemberProfileID: *memberProfileID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get member center history by profile: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, memberCenterHistory)
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/member-center",
		Method:       "GET",
		ResponseType: types.MemberCenterResponse{},
		Note:         "Returns all member centers for the current user's branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		memberCenter, err := core.MemberCenterCurrentBranch(context, service, userOrg.OrganizationID, *userOrg.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get member centers: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, core.MemberCenterManager(service).ToModels(memberCenter))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/member-center/search",
		Method:       "GET",
		ResponseType: types.MemberCenterResponse{},
		Note:         "Returns paginated member centers for the current user's branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		value, err := core.MemberCenterManager(service).NormalPagination(context, ctx, &types.MemberCenter{
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get member centers for pagination: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, value)
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/member-center",
		Method:       "POST",
		ResponseType: types.MemberCenterResponse{},
		RequestType: types.MemberCenterRequest{},
		Note:         "Creates a new member center record.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := core.MemberCenterManager(service).Validate(ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create member center failed (/member-center), validation error: " + err.Error(),
				Module:      "MemberCenter",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create member center failed (/member-center), user org error: " + err.Error(),
				Module:      "MemberCenter",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}

		memberCenter := &types.MemberCenter{
			Name:           req.Name,
			Description:    req.Description,
			CreatedAt:      time.Now().UTC(),
			CreatedByID:    userOrg.UserID,
			UpdatedAt:      time.Now().UTC(),
			UpdatedByID:    userOrg.UserID,
			BranchID:       *userOrg.BranchID,
			OrganizationID: userOrg.OrganizationID,
		}

		if err := core.MemberCenterManager(service).Create(context, memberCenter); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create member center failed (/member-center), db error: " + err.Error(),
				Module:      "MemberCenter",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create member center: " + err.Error()})
		}

		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "create-success",
			Description: "Created member center (/member-center): " + memberCenter.Name,
			Module:      "MemberCenter",
		})

		return ctx.JSON(http.StatusOK, core.MemberCenterManager(service).ToModel(memberCenter))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/member-center/:member_center_id",
		Method:       "PUT",
		ResponseType: types.MemberCenterResponse{},
		RequestType: types.MemberCenterRequest{},
		Note:         "Updates an existing member center record by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberCenterID, err := helpers.EngineUUIDParam(ctx, "member_center_id")
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update member center failed (/member-center/:member_center_id), invalid member_center_id: " + err.Error(),
				Module:      "MemberCenter",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member_center_id: " + err.Error()})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update member center failed (/member-center/:member_center_id), user org error: " + err.Error(),
				Module:      "MemberCenter",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		req, err := core.MemberCenterManager(service).Validate(ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update member center failed (/member-center/:member_center_id), validation error: " + err.Error(),
				Module:      "MemberCenter",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		memberCenter, err := core.MemberCenterManager(service).GetByID(context, *memberCenterID)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update member center failed (/member-center/:member_center_id), not found: " + err.Error(),
				Module:      "MemberCenter",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Member center not found: " + err.Error()})
		}
		memberCenter.UpdatedAt = time.Now().UTC()
		memberCenter.UpdatedByID = userOrg.UserID
		memberCenter.OrganizationID = userOrg.OrganizationID
		memberCenter.BranchID = *userOrg.BranchID
		memberCenter.Name = req.Name
		memberCenter.Description = req.Description
		if err := core.MemberCenterManager(service).UpdateByID(context, memberCenter.ID, memberCenter); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update member center failed (/member-center/:member_center_id), db error: " + err.Error(),
				Module:      "MemberCenter",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update member center: " + err.Error()})
		}
		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Updated member center (/member-center/:member_center_id): " + memberCenter.Name,
			Module:      "MemberCenter",
		})
		return ctx.JSON(http.StatusOK, core.MemberCenterManager(service).ToModel(memberCenter))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:  "/api/v1/member-center/:member_center_id",
		Method: "DELETE",
		Note:   "Deletes a member center record by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()

		memberCenterID, err := helpers.EngineUUIDParam(ctx, "member_center_id")
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete member center failed (/member-center/:member_center_id) | invalid member_center_id: " + err.Error(),
				Module:      "MemberCenter",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member_center_id: " + err.Error()})
		}

		value, err := core.MemberCenterManager(service).GetByID(context, *memberCenterID)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete member center failed (/member-center/:member_center_id) | not found: " + err.Error(),
				Module:      "MemberCenter",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Member center not found: " + err.Error()})
		}

		if err := core.MemberCenterManager(service).Delete(context, *memberCenterID); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete member center failed (/member-center/:member_center_id) | db error: " + err.Error(),
				Module:      "MemberCenter",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete member center: " + err.Error()})
		}

		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "delete-success",
			Description: "Deleted member center (/member-center/:member_center_id): " + value.Name,
			Module:      "MemberCenter",
		})

		return ctx.NoContent(http.StatusNoContent)
	})

	req.RegisterWebRoute(horizon.Route{
		Route:       "/api/v1/member-center/bulk-delete",
		Method:      "DELETE",
		Note:        "Deletes multiple member center records by their IDs.",
		RequestType: types.IDSRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var reqBody types.IDSRequest

		if err := ctx.Bind(&reqBody); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete member centers failed (/member-center/bulk-delete) | invalid request body: " + err.Error(),
				Module:      "MemberCenter",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body: " + err.Error()})
		}

		if len(reqBody.IDs) == 0 {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete member centers failed (/member-center/bulk-delete) | no IDs provided",
				Module:      "MemberCenter",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "No IDs provided for bulk delete"})
		}

		ids := make([]any, len(reqBody.IDs))
		for i, id := range reqBody.IDs {
			ids[i] = id
		}
		if err := core.MemberCenterManager(service).BulkDelete(context, ids); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete member centers failed (/member-center/bulk-delete) | error: " + err.Error(),
				Module:      "MemberCenter",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to bulk delete member centers: " + err.Error()})
		}

		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "bulk-delete-success",
			Description: "Bulk deleted member centers (/member-center/bulk-delete)",
			Module:      "MemberCenter",
		})

		return ctx.NoContent(http.StatusNoContent)
	})
}
