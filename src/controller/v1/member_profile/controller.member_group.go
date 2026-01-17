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

func MemberGroupController(service *horizon.HorizonService) {
	req := service.API

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/member-group-history",
		Method:       "GET",
		ResponseType: types.MemberGroupHistoryResponse{},
		Note:         "Returns all member group history entries for the current user's branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		memberGroupHistory, err := core.MemberGroupHistoryCurrentBranch(context, service, userOrg.OrganizationID, *userOrg.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get member group history: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, core.MemberGroupHistoryManager(service).ToModels(memberGroupHistory))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/member-group-history/member-profile/:member_profile_id/search",
		Method:       "GET",
		ResponseType: types.MemberGroupHistoryResponse{},
		Note:         "Returns member group history for a specific member profile ID.",
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
		memberGroupHistory, err := core.MemberGroupHistoryManager(service).NormalPagination(context, ctx, &types.MemberGroupHistory{
			MemberProfileID: *memberProfileID,
			BranchID:        *userOrg.BranchID,
			OrganizationID:  userOrg.OrganizationID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get member group history by profile: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, memberGroupHistory)
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/member-group",
		Method:       "GET",
		ResponseType: types.MemberGroupResponse{},
		Note:         "Returns all member groups for the current user's branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		memberGroup, err := core.MemberGroupManager(service).FindRaw(context, &types.MemberGroup{
			BranchID:       *userOrg.BranchID,
			OrganizationID: userOrg.OrganizationID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get member groups: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, memberGroup)
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/member-group/search",
		Method:       "GET",
		RequestType: types.MemberGroupRequest{},
		ResponseType: types.MemberGroupResponse{},
		Note:         "Returns paginated member groups for the current user's branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		memberGroup, err := core.MemberGroupManager(service).NormalPagination(context, ctx, &types.MemberGroup{
			BranchID:       *userOrg.BranchID,
			OrganizationID: userOrg.OrganizationID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get member groups: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, memberGroup)
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/member-group",
		Method:       "POST",
		ResponseType: types.MemberGroupResponse{},
		RequestType: types.MemberGroupRequest{},
		Note:         "Creates a new member group record.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := core.MemberGroupManager(service).Validate(ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create member group failed (/member-group), validation error: " + err.Error(),
				Module:      "MemberGroup",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create member group failed (/member-group), user org error: " + err.Error(),
				Module:      "MemberGroup",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}

		memberGroup := &types.MemberGroup{
			Name:           req.Name,
			Description:    req.Description,
			CreatedAt:      time.Now().UTC(),
			CreatedByID:    userOrg.UserID,
			UpdatedAt:      time.Now().UTC(),
			UpdatedByID:    userOrg.UserID,
			BranchID:       *userOrg.BranchID,
			OrganizationID: userOrg.OrganizationID,
		}

		if err := core.MemberGroupManager(service).Create(context, memberGroup); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create member group failed (/member-group), db error: " + err.Error(),
				Module:      "MemberGroup",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create member group: " + err.Error()})
		}

		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "create-success",
			Description: "Created member group (/member-group): " + memberGroup.Name,
			Module:      "MemberGroup",
		})

		return ctx.JSON(http.StatusOK, core.MemberGroupManager(service).ToModel(memberGroup))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/member-group/:member_group_id",
		Method:       "PUT",
		ResponseType: types.MemberGroupResponse{},
		RequestType: types.MemberGroupRequest{},
		Note:         "Updates an existing member group record by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberGroupID, err := helpers.EngineUUIDParam(ctx, "member_group_id")
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update member group failed (/member-group/:member_group_id), invalid member_group_id: " + err.Error(),
				Module:      "MemberGroup",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member_group_id: " + err.Error()})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update member group failed (/member-group/:member_group_id), user org error: " + err.Error(),
				Module:      "MemberGroup",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		req, err := core.MemberGroupManager(service).Validate(ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update member group failed (/member-group/:member_group_id), validation error: " + err.Error(),
				Module:      "MemberGroup",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		memberGroup, err := core.MemberGroupManager(service).GetByID(context, *memberGroupID)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update member group failed (/member-group/:member_group_id), not found: " + err.Error(),
				Module:      "MemberGroup",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Member group not found: " + err.Error()})
		}
		memberGroup.UpdatedAt = time.Now().UTC()
		memberGroup.UpdatedByID = userOrg.UserID
		memberGroup.OrganizationID = userOrg.OrganizationID
		memberGroup.BranchID = *userOrg.BranchID
		memberGroup.Name = req.Name
		memberGroup.Description = req.Description
		if err := core.MemberGroupManager(service).UpdateByID(context, memberGroup.ID, memberGroup); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update member group failed (/member-group/:member_group_id), db error: " + err.Error(),
				Module:      "MemberGroup",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update member group: " + err.Error()})
		}
		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Updated member group (/member-group/:member_group_id): " + memberGroup.Name,
			Module:      "MemberGroup",
		})
		return ctx.JSON(http.StatusOK, core.MemberGroupManager(service).ToModel(memberGroup))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:  "/api/v1/member-group/:member_group_id",
		Method: "DELETE",
		Note:   "Deletes a member group record by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberGroupID, err := helpers.EngineUUIDParam(ctx, "member_group_id")
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete member group failed (/member-group/:member_group_id), invalid member_group_id: " + err.Error(),
				Module:      "MemberGroup",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member_group_id: " + err.Error()})
		}
		value, err := core.MemberGroupManager(service).GetByID(context, *memberGroupID)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete member group failed (/member-group/:member_group_id), record not found: " + err.Error(),
				Module:      "MemberGroup",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Member group not found: " + err.Error()})
		}
		if err := core.MemberGroupManager(service).Delete(context, *memberGroupID); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete member group failed (/member-group/:member_group_id), db error: " + err.Error(),
				Module:      "MemberGroup",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete member group: " + err.Error()})
		}
		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "delete-success",
			Description: "Deleted member group (/member-group/:member_group_id): " + value.Name,
			Module:      "MemberGroup",
		})
		return ctx.NoContent(http.StatusNoContent)
	})

	req.RegisterWebRoute(horizon.Route{
		Route:       "/api/v1/member-group/bulk-delete",
		Method:      "DELETE",
		Note:        "Deletes multiple member group records by their IDs.",
		RequestType: types.IDSRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var reqBody types.IDSRequest

		if err := ctx.Bind(&reqBody); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete member groups failed (/member-group/bulk-delete) | invalid request body: " + err.Error(),
				Module:      "MemberGroup",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body: " + err.Error()})
		}

		if len(reqBody.IDs) == 0 {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete member groups failed (/member-group/bulk-delete) | no IDs provided",
				Module:      "MemberGroup",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "No IDs provided for bulk delete"})
		}

		ids := make([]any, len(reqBody.IDs))
		for i, id := range reqBody.IDs {
			ids[i] = id
		}
		if err := core.MemberGroupManager(service).BulkDelete(context, ids); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete member groups failed (/member-group/bulk-delete) | error: " + err.Error(),
				Module:      "MemberGroup",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to bulk delete member groups: " + err.Error()})
		}

		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "bulk-delete-success",
			Description: "Bulk deleted member groups (/member-group/bulk-delete)",
			Module:      "MemberGroup",
		})

		return ctx.NoContent(http.StatusNoContent)
	})
}
