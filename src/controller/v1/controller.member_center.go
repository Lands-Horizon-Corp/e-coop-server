package controller_v1

import (
	"fmt"
	"net/http"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/services/handlers"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/event"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/model"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func (c *Controller) MemberCenterController() {
	req := c.provider.Service.Request

	// Get all member center history for the current branch
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/member-center-history",
		Method:       "GET",
		ResponseType: model.MemberCenterResponse{},
		Note:         "Returns all member center history entries for the current user's branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		memberCenterHistory, err := c.model.MemberCenterHistoryCurrentBranch(context, user.OrganizationID, *user.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get member center history: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.MemberCenterHistoryManager.Filtered(context, ctx, memberCenterHistory))
	})

	// Get member center history by member profile ID
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/member-center-history/member-profile/:member_profile_id/search",
		Method:       "GET",
		ResponseType: model.MemberCenterHistoryResponse{},
		Note:         "Returns member center history for a specific member profile ID.",
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
		memberCenterHistory, err := c.model.MemberCenterHistoryMemberProfileID(context, *memberProfileID, user.OrganizationID, *user.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get member center history by profile: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.MemberCenterHistoryManager.Pagination(context, ctx, memberCenterHistory))
	})

	// Get all member centers for the current branch
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/member-center",
		Method:       "GET",
		ResponseType: model.MemberCenterResponse{},
		Note:         "Returns all member centers for the current user's branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		memberCenter, err := c.model.MemberCenterCurrentBranch(context, user.OrganizationID, *user.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get member centers: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.MemberCenterManager.Filtered(context, ctx, memberCenter))
	})

	// Get paginated member centers
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/member-center/search",
		Method:       "GET",
		ResponseType: model.MemberCenterResponse{},
		Note:         "Returns paginated member centers for the current user's branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		value, err := c.model.MemberCenterCurrentBranch(context, user.OrganizationID, *user.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get member centers for pagination: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.MemberCenterManager.Pagination(context, ctx, value))
	})

	// Create a new member center
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/member-center",
		Method:       "POST",
		ResponseType: model.MemberCenterResponse{},
		RequestType:  model.MemberCenterRequest{},
		Note:         "Creates a new member center record.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := c.model.MemberCenterManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create member center failed (/member-center), validation error: " + err.Error(),
				Module:      "MemberCenter",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create member center failed (/member-center), user org error: " + err.Error(),
				Module:      "MemberCenter",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}

		memberCenter := &model.MemberCenter{
			Name:           req.Name,
			Description:    req.Description,
			CreatedAt:      time.Now().UTC(),
			CreatedByID:    user.UserID,
			UpdatedAt:      time.Now().UTC(),
			UpdatedByID:    user.UserID,
			BranchID:       *user.BranchID,
			OrganizationID: user.OrganizationID,
		}

		if err := c.model.MemberCenterManager.Create(context, memberCenter); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create member center failed (/member-center), db error: " + err.Error(),
				Module:      "MemberCenter",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create member center: " + err.Error()})
		}

		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "create-success",
			Description: "Created member center (/member-center): " + memberCenter.Name,
			Module:      "MemberCenter",
		})

		return ctx.JSON(http.StatusOK, c.model.MemberCenterManager.ToModel(memberCenter))
	})

	// Update an existing member center by ID
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/member-center/:member_center_id",
		Method:       "PUT",
		ResponseType: model.MemberCenterResponse{},
		RequestType:  model.MemberCenterRequest{},
		Note:         "Updates an existing member center record by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberCenterID, err := handlers.EngineUUIDParam(ctx, "member_center_id")
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update member center failed (/member-center/:member_center_id), invalid member_center_id: " + err.Error(),
				Module:      "MemberCenter",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member_center_id: " + err.Error()})
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update member center failed (/member-center/:member_center_id), user org error: " + err.Error(),
				Module:      "MemberCenter",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		req, err := c.model.MemberCenterManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update member center failed (/member-center/:member_center_id), validation error: " + err.Error(),
				Module:      "MemberCenter",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		memberCenter, err := c.model.MemberCenterManager.GetByID(context, *memberCenterID)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update member center failed (/member-center/:member_center_id), not found: " + err.Error(),
				Module:      "MemberCenter",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Member center not found: " + err.Error()})
		}
		memberCenter.UpdatedAt = time.Now().UTC()
		memberCenter.UpdatedByID = user.UserID
		memberCenter.OrganizationID = user.OrganizationID
		memberCenter.BranchID = *user.BranchID
		memberCenter.Name = req.Name
		memberCenter.Description = req.Description
		if err := c.model.MemberCenterManager.UpdateFields(context, memberCenter.ID, memberCenter); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update member center failed (/member-center/:member_center_id), db error: " + err.Error(),
				Module:      "MemberCenter",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update member center: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Updated member center (/member-center/:member_center_id): " + memberCenter.Name,
			Module:      "MemberCenter",
		})
		return ctx.JSON(http.StatusOK, c.model.MemberCenterManager.ToModel(memberCenter))
	})

	// Delete a member center by ID
	req.RegisterRoute(handlers.Route{
		Route:  "/api/v1/member-center/:member_center_id",
		Method: "DELETE",
		Note:   "Deletes a member center record by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberCenterID, err := handlers.EngineUUIDParam(ctx, "member_center_id")
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete member center failed (/member-center/:member_center_id), invalid member_center_id: " + err.Error(),
				Module:      "MemberCenter",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member_center_id: " + err.Error()})
		}
		value, err := c.model.MemberCenterManager.GetByID(context, *memberCenterID)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete member center failed (/member-center/:member_center_id), not found: " + err.Error(),
				Module:      "MemberCenter",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Member center not found: " + err.Error()})
		}
		if err := c.model.MemberCenterManager.DeleteByID(context, *memberCenterID); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete member center failed (/member-center/:member_center_id), db error: " + err.Error(),
				Module:      "MemberCenter",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete member center: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "delete-success",
			Description: "Deleted member center (/member-center/:member_center_id): " + value.Name,
			Module:      "MemberCenter",
		})
		return ctx.NoContent(http.StatusNoContent)
	})

	// Bulk delete member centers by IDs
	req.RegisterRoute(handlers.Route{
		Route:       "/api/v1/member-center/bulk-delete",
		Method:      "DELETE",
		Note:        "Deletes multiple member center records by their IDs.",
		RequestType: model.IDSRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var reqBody model.IDSRequest

		if err := ctx.Bind(&reqBody); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete member centers failed (/member-center/bulk-delete), invalid request body.",
				Module:      "MemberCenter",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body: " + err.Error()})
		}
		if len(reqBody.IDs) == 0 {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete member centers failed (/member-center/bulk-delete), no IDs provided.",
				Module:      "MemberCenter",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "No IDs provided for deletion."})
		}

		tx := c.provider.Service.Database.Client().Begin()
		if tx.Error != nil {
			tx.Rollback()
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete member centers failed (/member-center/bulk-delete), begin tx error: " + tx.Error.Error(),
				Module:      "MemberCenter",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to begin transaction: " + tx.Error.Error()})
		}

		names := ""
		for _, rawID := range reqBody.IDs {
			memberCenterID, err := uuid.Parse(rawID)
			if err != nil {
				tx.Rollback()
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "bulk-delete-error",
					Description: "Bulk delete member centers failed (/member-center/bulk-delete), invalid UUID: " + rawID,
					Module:      "MemberCenter",
				})
				return ctx.JSON(http.StatusBadRequest, map[string]string{"error": fmt.Sprintf("Invalid UUID '%s': %s", rawID, err.Error())})
			}
			value, err := c.model.MemberCenterManager.GetByID(context, memberCenterID)
			if err != nil {
				tx.Rollback()
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "bulk-delete-error",
					Description: "Bulk delete member centers failed (/member-center/bulk-delete), not found: " + rawID,
					Module:      "MemberCenter",
				})
				return ctx.JSON(http.StatusNotFound, map[string]string{"error": fmt.Sprintf("Member center with ID '%s' not found: %s", rawID, err.Error())})
			}
			names += value.Name + ","
			if err := c.model.MemberCenterManager.DeleteByIDWithTx(context, tx, memberCenterID); err != nil {
				tx.Rollback()
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "bulk-delete-error",
					Description: "Bulk delete member centers failed (/member-center/bulk-delete), db error: " + err.Error(),
					Module:      "MemberCenter",
				})
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": fmt.Sprintf("Failed to delete member center with ID '%s': %s", rawID, err.Error())})
			}
		}

		if err := tx.Commit().Error; err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete member centers failed (/member-center/bulk-delete), commit error: " + err.Error(),
				Module:      "MemberCenter",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to commit transaction: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "bulk-delete-success",
			Description: "Bulk deleted member centers (/member-center/bulk-delete): " + names,
			Module:      "MemberCenter",
		})
		return ctx.NoContent(http.StatusNoContent)
	})
}
