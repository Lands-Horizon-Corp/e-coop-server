package v1

import (
	"net/http"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/server/event"
	"github.com/Lands-Horizon-Corp/e-coop-server/server/model/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/services/handlers"
	"github.com/labstack/echo/v4"
)

func (c *Controller) memberCenterController() {
	req := c.provider.Service.WebRequest

	// Get all member center history for the current branch
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/member-center-history",
		Method:       "GET",
		ResponseType: core.MemberCenterResponse{},
		Note:         "Returns all member center history entries for the current user's branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		memberCenterHistory, err := c.core.MemberCenterHistoryCurrentBranch(context, userOrg.OrganizationID, *userOrg.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get member center history: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.core.MemberCenterHistoryManager.ToModels(memberCenterHistory))
	})

	// Get member center history by member profile ID
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/member-center-history/member-profile/:member_profile_id/search",
		Method:       "GET",
		ResponseType: core.MemberCenterHistoryResponse{},
		Note:         "Returns member center history for a specific member profile ID.",
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
		memberCenterHistory, err := c.core.MemberCenterHistoryManager.PaginationWithFields(context, ctx, &core.MemberCenterHistory{
			OrganizationID:  userOrg.OrganizationID,
			BranchID:        *userOrg.BranchID,
			MemberProfileID: *memberProfileID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get member center history by profile: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, memberCenterHistory)
	})

	// Get all member centers for the current branch
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/member-center",
		Method:       "GET",
		ResponseType: core.MemberCenterResponse{},
		Note:         "Returns all member centers for the current user's branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		memberCenter, err := c.core.MemberCenterCurrentBranch(context, userOrg.OrganizationID, *userOrg.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get member centers: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.core.MemberCenterManager.ToModels(memberCenter))
	})

	// Get paginated member centers
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/member-center/search",
		Method:       "GET",
		ResponseType: core.MemberCenterResponse{},
		Note:         "Returns paginated member centers for the current user's branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		value, err := c.core.MemberCenterManager.PaginationWithFields(context, ctx, &core.MemberCenter{
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get member centers for pagination: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, value)
	})

	// Create a new member center
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/member-center",
		Method:       "POST",
		ResponseType: core.MemberCenterResponse{},
		RequestType:  core.MemberCenterRequest{},
		Note:         "Creates a new member center record.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := c.core.MemberCenterManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create member center failed (/member-center), validation error: " + err.Error(),
				Module:      "MemberCenter",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create member center failed (/member-center), user org error: " + err.Error(),
				Module:      "MemberCenter",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}

		memberCenter := &core.MemberCenter{
			Name:           req.Name,
			Description:    req.Description,
			CreatedAt:      time.Now().UTC(),
			CreatedByID:    userOrg.UserID,
			UpdatedAt:      time.Now().UTC(),
			UpdatedByID:    userOrg.UserID,
			BranchID:       *userOrg.BranchID,
			OrganizationID: userOrg.OrganizationID,
		}

		if err := c.core.MemberCenterManager.Create(context, memberCenter); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create member center failed (/member-center), db error: " + err.Error(),
				Module:      "MemberCenter",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create member center: " + err.Error()})
		}

		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "create-success",
			Description: "Created member center (/member-center): " + memberCenter.Name,
			Module:      "MemberCenter",
		})

		return ctx.JSON(http.StatusOK, c.core.MemberCenterManager.ToModel(memberCenter))
	})

	// Update an existing member center by ID
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/member-center/:member_center_id",
		Method:       "PUT",
		ResponseType: core.MemberCenterResponse{},
		RequestType:  core.MemberCenterRequest{},
		Note:         "Updates an existing member center record by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberCenterID, err := handlers.EngineUUIDParam(ctx, "member_center_id")
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update member center failed (/member-center/:member_center_id), invalid member_center_id: " + err.Error(),
				Module:      "MemberCenter",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member_center_id: " + err.Error()})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update member center failed (/member-center/:member_center_id), user org error: " + err.Error(),
				Module:      "MemberCenter",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		req, err := c.core.MemberCenterManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update member center failed (/member-center/:member_center_id), validation error: " + err.Error(),
				Module:      "MemberCenter",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		memberCenter, err := c.core.MemberCenterManager.GetByID(context, *memberCenterID)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
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
		if err := c.core.MemberCenterManager.UpdateByID(context, memberCenter.ID, memberCenter); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update member center failed (/member-center/:member_center_id), db error: " + err.Error(),
				Module:      "MemberCenter",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update member center: " + err.Error()})
		}
		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Updated member center (/member-center/:member_center_id): " + memberCenter.Name,
			Module:      "MemberCenter",
		})
		return ctx.JSON(http.StatusOK, c.core.MemberCenterManager.ToModel(memberCenter))
	})

	// Delete a member center by ID (cleaned up messages to match other handlers)
	req.RegisterRoute(handlers.Route{
		Route:  "/api/v1/member-center/:member_center_id",
		Method: "DELETE",
		Note:   "Deletes a member center record by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()

		memberCenterID, err := handlers.EngineUUIDParam(ctx, "member_center_id")
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete member center failed (/member-center/:member_center_id) | invalid member_center_id: " + err.Error(),
				Module:      "MemberCenter",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member_center_id: " + err.Error()})
		}

		value, err := c.core.MemberCenterManager.GetByID(context, *memberCenterID)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete member center failed (/member-center/:member_center_id) | not found: " + err.Error(),
				Module:      "MemberCenter",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Member center not found: " + err.Error()})
		}

		if err := c.core.MemberCenterManager.Delete(context, *memberCenterID); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete member center failed (/member-center/:member_center_id) | db error: " + err.Error(),
				Module:      "MemberCenter",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete member center: " + err.Error()})
		}

		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "delete-success",
			Description: "Deleted member center (/member-center/:member_center_id): " + value.Name,
			Module:      "MemberCenter",
		})

		return ctx.NoContent(http.StatusNoContent)
	})

	// Simplified bulk-delete handler for member centers (mirrors the feedback/holiday pattern)
	req.RegisterRoute(handlers.Route{
		Route:       "/api/v1/member-center/bulk-delete",
		Method:      "DELETE",
		Note:        "Deletes multiple member center records by their IDs.",
		RequestType: core.IDSRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var reqBody core.IDSRequest

		if err := ctx.Bind(&reqBody); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete member centers failed (/member-center/bulk-delete) | invalid request body: " + err.Error(),
				Module:      "MemberCenter",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body: " + err.Error()})
		}

		if len(reqBody.IDs) == 0 {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete member centers failed (/member-center/bulk-delete) | no IDs provided",
				Module:      "MemberCenter",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "No IDs provided for bulk delete"})
		}

		// Delegate deletion to the manager. Manager should handle transactions, validations and DeletedBy bookkeeping.
		if err := c.core.MemberCenterManager.BulkDelete(context, reqBody.IDs); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete member centers failed (/member-center/bulk-delete) | error: " + err.Error(),
				Module:      "MemberCenter",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to bulk delete member centers: " + err.Error()})
		}

		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "bulk-delete-success",
			Description: "Bulk deleted member centers (/member-center/bulk-delete)",
			Module:      "MemberCenter",
		})

		return ctx.NoContent(http.StatusNoContent)
	})
}
