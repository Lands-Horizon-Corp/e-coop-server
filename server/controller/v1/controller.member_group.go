package v1

import (
	"net/http"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/server/event"
	"github.com/Lands-Horizon-Corp/e-coop-server/server/model/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/services/handlers"
	"github.com/labstack/echo/v4"
)

func (c *Controller) memberGroupController() {
	req := c.provider.Service.Request

	// Get all member group history for the current branch
	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/member-group-history",
		Method:       "GET",
		ResponseType: core.MemberGroupHistoryResponse{},
		Note:         "Returns all member group history entries for the current user's branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		memberGroupHistory, err := c.core.MemberGroupHistoryCurrentBranch(context, userOrg.OrganizationID, *userOrg.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get member group history: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.core.MemberGroupHistoryManager.ToModels(memberGroupHistory))
	})

	// Get member group history by member profile ID
	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/member-group-history/member-profile/:member_profile_id/search",
		Method:       "GET",
		ResponseType: core.MemberGroupHistoryResponse{},
		Note:         "Returns member group history for a specific member profile ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberProfileID, err := handlers.EngineUUIDParam(ctx, "member_profile_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member_profile_id: " + err.Error()})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		memberGroupHistory, err := c.core.MemberGroupHistoryManager.PaginationWithFields(context, ctx, &core.MemberGroupHistory{
			MemberProfileID: *memberProfileID,
			BranchID:        *userOrg.BranchID,
			OrganizationID:  userOrg.OrganizationID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get member group history by profile: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, memberGroupHistory)
	})

	// Get all member groups for the current branch
	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/member-group",
		Method:       "GET",
		ResponseType: core.MemberGroupResponse{},
		Note:         "Returns all member groups for the current user's branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		memberGroup, err := c.core.MemberGroupManager.FindRaw(context, &core.MemberGroup{
			BranchID:       *userOrg.BranchID,
			OrganizationID: userOrg.OrganizationID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get member groups: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, memberGroup)
	})

	// Get paginated member groups
	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/member-group/search",
		Method:       "GET",
		RequestType:  core.MemberGroupRequest{},
		ResponseType: core.MemberGroupResponse{},
		Note:         "Returns paginated member groups for the current user's branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		memberGroup, err := c.core.MemberGroupManager.PaginationWithFields(context, ctx, &core.MemberGroup{
			BranchID:       *userOrg.BranchID,
			OrganizationID: userOrg.OrganizationID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get member groups: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, memberGroup)
	})

	// Create a new member group
	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/member-group",
		Method:       "POST",
		ResponseType: core.MemberGroupResponse{},
		RequestType:  core.MemberGroupRequest{},
		Note:         "Creates a new member group record.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := c.core.MemberGroupManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create member group failed (/member-group), validation error: " + err.Error(),
				Module:      "MemberGroup",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create member group failed (/member-group), user org error: " + err.Error(),
				Module:      "MemberGroup",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}

		memberGroup := &core.MemberGroup{
			Name:           req.Name,
			Description:    req.Description,
			CreatedAt:      time.Now().UTC(),
			CreatedByID:    userOrg.UserID,
			UpdatedAt:      time.Now().UTC(),
			UpdatedByID:    userOrg.UserID,
			BranchID:       *userOrg.BranchID,
			OrganizationID: userOrg.OrganizationID,
		}

		if err := c.core.MemberGroupManager.Create(context, memberGroup); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create member group failed (/member-group), db error: " + err.Error(),
				Module:      "MemberGroup",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create member group: " + err.Error()})
		}

		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "create-success",
			Description: "Created member group (/member-group): " + memberGroup.Name,
			Module:      "MemberGroup",
		})

		return ctx.JSON(http.StatusOK, c.core.MemberGroupManager.ToModel(memberGroup))
	})

	// Update an existing member group by ID
	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/member-group/:member_group_id",
		Method:       "PUT",
		ResponseType: core.MemberGroupResponse{},
		RequestType:  core.MemberGroupRequest{},
		Note:         "Updates an existing member group record by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberGroupID, err := handlers.EngineUUIDParam(ctx, "member_group_id")
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update member group failed (/member-group/:member_group_id), invalid member_group_id: " + err.Error(),
				Module:      "MemberGroup",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member_group_id: " + err.Error()})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update member group failed (/member-group/:member_group_id), user org error: " + err.Error(),
				Module:      "MemberGroup",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		req, err := c.core.MemberGroupManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update member group failed (/member-group/:member_group_id), validation error: " + err.Error(),
				Module:      "MemberGroup",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		memberGroup, err := c.core.MemberGroupManager.GetByID(context, *memberGroupID)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
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
		if err := c.core.MemberGroupManager.UpdateByID(context, memberGroup.ID, memberGroup); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update member group failed (/member-group/:member_group_id), db error: " + err.Error(),
				Module:      "MemberGroup",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update member group: " + err.Error()})
		}
		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Updated member group (/member-group/:member_group_id): " + memberGroup.Name,
			Module:      "MemberGroup",
		})
		return ctx.JSON(http.StatusOK, c.core.MemberGroupManager.ToModel(memberGroup))
	})

	// Delete a member group by ID
	req.RegisterWebRoute(handlers.Route{
		Route:  "/api/v1/member-group/:member_group_id",
		Method: "DELETE",
		Note:   "Deletes a member group record by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberGroupID, err := handlers.EngineUUIDParam(ctx, "member_group_id")
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete member group failed (/member-group/:member_group_id), invalid member_group_id: " + err.Error(),
				Module:      "MemberGroup",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member_group_id: " + err.Error()})
		}
		value, err := c.core.MemberGroupManager.GetByID(context, *memberGroupID)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete member group failed (/member-group/:member_group_id), record not found: " + err.Error(),
				Module:      "MemberGroup",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Member group not found: " + err.Error()})
		}
		if err := c.core.MemberGroupManager.Delete(context, *memberGroupID); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete member group failed (/member-group/:member_group_id), db error: " + err.Error(),
				Module:      "MemberGroup",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete member group: " + err.Error()})
		}
		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "delete-success",
			Description: "Deleted member group (/member-group/:member_group_id): " + value.Name,
			Module:      "MemberGroup",
		})
		return ctx.NoContent(http.StatusNoContent)
	})

	// Simplified bulk-delete handler for member groups (mirrors feedback/holiday pattern)
	req.RegisterWebRoute(handlers.Route{
		Route:       "/api/v1/member-group/bulk-delete",
		Method:      "DELETE",
		Note:        "Deletes multiple member group records by their IDs.",
		RequestType: core.IDSRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var reqBody core.IDSRequest

		if err := ctx.Bind(&reqBody); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete member groups failed (/member-group/bulk-delete) | invalid request body: " + err.Error(),
				Module:      "MemberGroup",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body: " + err.Error()})
		}

		if len(reqBody.IDs) == 0 {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete member groups failed (/member-group/bulk-delete) | no IDs provided",
				Module:      "MemberGroup",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "No IDs provided for bulk delete"})
		}

		// Delegate deletion to the manager. Manager should handle transactions, validations and DeletedBy bookkeeping.
		if err := c.core.MemberGroupManager.BulkDelete(context, reqBody.IDs); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete member groups failed (/member-group/bulk-delete) | error: " + err.Error(),
				Module:      "MemberGroup",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to bulk delete member groups: " + err.Error()})
		}

		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "bulk-delete-success",
			Description: "Bulk deleted member groups (/member-group/bulk-delete)",
			Module:      "MemberGroup",
		})

		return ctx.NoContent(http.StatusNoContent)
	})
}
