package v1

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/server/event"
	"github.com/Lands-Horizon-Corp/e-coop-server/server/model/modelcore"
	"github.com/Lands-Horizon-Corp/e-coop-server/services/handlers"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func (c *Controller) memberGroupController() {
	req := c.provider.Service.Request

	// Get all member group history for the current branch
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/member-group-history",
		Method:       "GET",
		ResponseType: modelcore.MemberGroupHistoryResponse{},
		Note:         "Returns all member group history entries for the current user's branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		memberGroupHistory, err := c.modelcore.MemberGroupHistoryCurrentBranch(context, user.OrganizationID, *user.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get member group history: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.modelcore.MemberGroupHistoryManager.Filtered(context, ctx, memberGroupHistory))
	})

	// Get member group history by member profile ID
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/member-group-history/member-profile/:member_profile_id/search",
		Method:       "GET",
		ResponseType: modelcore.MemberGroupHistoryResponse{},
		Note:         "Returns member group history for a specific member profile ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberProfileID, err := handlers.EngineUUIDParam(ctx, "member_profile_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member_profile_id: " + err.Error()})
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		memberGroupHistory, err := c.modelcore.MemberGroupHistoryMemberProfileID(context, *memberProfileID, user.OrganizationID, *user.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get member group history by profile: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.modelcore.MemberGroupHistoryManager.Pagination(context, ctx, memberGroupHistory))
	})

	// Get all member groups for the current branch
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/member-group",
		Method:       "GET",
		ResponseType: modelcore.MemberGroupResponse{},
		Note:         "Returns all member groups for the current user's branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		memberGroup, err := c.modelcore.MemberGroupCurrentBranch(context, user.OrganizationID, *user.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get member groups: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.modelcore.MemberGroupManager.Filtered(context, ctx, memberGroup))
	})

	// Get paginated member groups
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/member-group/search",
		Method:       "GET",
		RequestType:  modelcore.MemberGroupRequest{},
		ResponseType: modelcore.MemberGroupResponse{},
		Note:         "Returns paginated member groups for the current user's branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		value, err := c.modelcore.MemberGroupCurrentBranch(context, user.OrganizationID, *user.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get member groups for pagination: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.modelcore.MemberGroupManager.Pagination(context, ctx, value))
	})

	// Create a new member group
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/member-group",
		Method:       "POST",
		ResponseType: modelcore.MemberGroupResponse{},
		RequestType:  modelcore.MemberGroupRequest{},
		Note:         "Creates a new member group record.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := c.modelcore.MemberGroupManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create member group failed (/member-group), validation error: " + err.Error(),
				Module:      "MemberGroup",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create member group failed (/member-group), user org error: " + err.Error(),
				Module:      "MemberGroup",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}

		memberGroup := &modelcore.MemberGroup{
			Name:           req.Name,
			Description:    req.Description,
			CreatedAt:      time.Now().UTC(),
			CreatedByID:    user.UserID,
			UpdatedAt:      time.Now().UTC(),
			UpdatedByID:    user.UserID,
			BranchID:       *user.BranchID,
			OrganizationID: user.OrganizationID,
		}

		if err := c.modelcore.MemberGroupManager.Create(context, memberGroup); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create member group failed (/member-group), db error: " + err.Error(),
				Module:      "MemberGroup",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create member group: " + err.Error()})
		}

		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "create-success",
			Description: "Created member group (/member-group): " + memberGroup.Name,
			Module:      "MemberGroup",
		})

		return ctx.JSON(http.StatusOK, c.modelcore.MemberGroupManager.ToModel(memberGroup))
	})

	// Update an existing member group by ID
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/member-group/:member_group_id",
		Method:       "PUT",
		ResponseType: modelcore.MemberGroupResponse{},
		RequestType:  modelcore.MemberGroupRequest{},
		Note:         "Updates an existing member group record by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberGroupID, err := handlers.EngineUUIDParam(ctx, "member_group_id")
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update member group failed (/member-group/:member_group_id), invalid member_group_id: " + err.Error(),
				Module:      "MemberGroup",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member_group_id: " + err.Error()})
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update member group failed (/member-group/:member_group_id), user org error: " + err.Error(),
				Module:      "MemberGroup",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		req, err := c.modelcore.MemberGroupManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update member group failed (/member-group/:member_group_id), validation error: " + err.Error(),
				Module:      "MemberGroup",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		memberGroup, err := c.modelcore.MemberGroupManager.GetByID(context, *memberGroupID)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update member group failed (/member-group/:member_group_id), not found: " + err.Error(),
				Module:      "MemberGroup",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Member group not found: " + err.Error()})
		}
		memberGroup.UpdatedAt = time.Now().UTC()
		memberGroup.UpdatedByID = user.UserID
		memberGroup.OrganizationID = user.OrganizationID
		memberGroup.BranchID = *user.BranchID
		memberGroup.Name = req.Name
		memberGroup.Description = req.Description
		if err := c.modelcore.MemberGroupManager.UpdateFields(context, memberGroup.ID, memberGroup); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update member group failed (/member-group/:member_group_id), db error: " + err.Error(),
				Module:      "MemberGroup",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update member group: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Updated member group (/member-group/:member_group_id): " + memberGroup.Name,
			Module:      "MemberGroup",
		})
		return ctx.JSON(http.StatusOK, c.modelcore.MemberGroupManager.ToModel(memberGroup))
	})

	// Delete a member group by ID
	req.RegisterRoute(handlers.Route{
		Route:  "/api/v1/member-group/:member_group_id",
		Method: "DELETE",
		Note:   "Deletes a member group record by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberGroupID, err := handlers.EngineUUIDParam(ctx, "member_group_id")
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete member group failed (/member-group/:member_group_id), invalid member_group_id: " + err.Error(),
				Module:      "MemberGroup",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member_group_id: " + err.Error()})
		}
		value, err := c.modelcore.MemberGroupManager.GetByID(context, *memberGroupID)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete member group failed (/member-group/:member_group_id), record not found: " + err.Error(),
				Module:      "MemberGroup",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Member group not found: " + err.Error()})
		}
		if err := c.modelcore.MemberGroupManager.DeleteByID(context, *memberGroupID); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete member group failed (/member-group/:member_group_id), db error: " + err.Error(),
				Module:      "MemberGroup",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete member group: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "delete-success",
			Description: "Deleted member group (/member-group/:member_group_id): " + value.Name,
			Module:      "MemberGroup",
		})
		return ctx.NoContent(http.StatusNoContent)
	})

	// Bulk delete member groups by IDs
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/member-group/bulk-delete",
		Method:       "DELETE",
		ResponseType: modelcore.IDSRequest{},
		Note:         "Deletes multiple member group records by their IDs.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var reqBody modelcore.IDSRequest

		if err := ctx.Bind(&reqBody); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete member groups failed (/member-group/bulk-delete), invalid request body.",
				Module:      "MemberGroup",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body: " + err.Error()})
		}

		if len(reqBody.IDs) == 0 {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete member groups failed (/member-group/bulk-delete), no IDs provided.",
				Module:      "MemberGroup",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "No IDs provided for deletion."})
		}

		tx := c.provider.Service.Database.Client().Begin()
		if tx.Error != nil {
			tx.Rollback()
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete member groups failed (/member-group/bulk-delete), begin tx error: " + tx.Error.Error(),
				Module:      "MemberGroup",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to begin transaction: " + tx.Error.Error()})
		}

		var namesSlice []string
		for _, rawID := range reqBody.IDs {
			memberGroupID, err := uuid.Parse(rawID)
			if err != nil {
				tx.Rollback()
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "bulk-delete-error",
					Description: "Bulk delete member groups failed (/member-group/bulk-delete), invalid UUID: " + rawID,
					Module:      "MemberGroup",
				})
				return ctx.JSON(http.StatusBadRequest, map[string]string{"error": fmt.Sprintf("Invalid UUID '%s': %s", rawID, err.Error())})
			}

			value, err := c.modelcore.MemberGroupManager.GetByID(context, memberGroupID)
			if err != nil {
				tx.Rollback()
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "bulk-delete-error",
					Description: "Bulk delete member groups failed (/member-group/bulk-delete), not found: " + rawID,
					Module:      "MemberGroup",
				})
				return ctx.JSON(http.StatusNotFound, map[string]string{"error": fmt.Sprintf("Member group with ID '%s' not found: %s", rawID, err.Error())})
			}

			namesSlice = append(namesSlice, value.Name)
			if err := c.modelcore.MemberGroupManager.DeleteByIDWithTx(context, tx, memberGroupID); err != nil {
				tx.Rollback()
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "bulk-delete-error",
					Description: "Bulk delete member groups failed (/member-group/bulk-delete), db error: " + err.Error(),
					Module:      "MemberGroup",
				})
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": fmt.Sprintf("Failed to delete member group with ID '%s': %s", rawID, err.Error())})
			}
		}
		names := strings.Join(namesSlice, ",")

		if err := tx.Commit().Error; err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete member groups failed (/member-group/bulk-delete), commit error: " + err.Error(),
				Module:      "MemberGroup",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to commit transaction: " + err.Error()})
		}

		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "bulk-delete-success",
			Description: "Bulk deleted member groups (/member-group/bulk-delete): " + names,
			Module:      "MemberGroup",
		})

		return ctx.NoContent(http.StatusNoContent)
	})
}
