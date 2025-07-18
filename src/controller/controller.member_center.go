package controller

import (
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/lands-horizon/horizon-server/services/horizon"
	"github.com/lands-horizon/horizon-server/src/model"
)

func (c *Controller) MemberCenterController() {
	req := c.provider.Service.Request

	// Get all member center history for the current branch
	req.RegisterRoute(horizon.Route{
		Route:    "/member-center-history",
		Method:   "GET",
		Response: "TMemberCenterHistory[]",
		Note:     "Returns all member center history entries for the current user's branch.",
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
		return ctx.JSON(http.StatusOK, c.model.MemberCenterHistoryManager.ToModels(memberCenterHistory))
	})

	// Get member center history by member profile ID
	req.RegisterRoute(horizon.Route{
		Route:    "/member-center-history/member-profile/:member_profile_id/search",
		Method:   "GET",
		Response: "TMemberCenterHistory[]",
		Note:     "Returns member center history for a specific member profile ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberProfileID, err := horizon.EngineUUIDParam(ctx, "member_profile_id")
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
	req.RegisterRoute(horizon.Route{
		Route:    "/member-center",
		Method:   "GET",
		Response: "TMemberCenter[]",
		Note:     "Returns all member centers for the current user's branch.",
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
		return ctx.JSON(http.StatusOK, c.model.MemberCenterManager.ToModels(memberCenter))
	})

	// Get paginated member centers
	req.RegisterRoute(horizon.Route{
		Route:    "/member-center/search",
		Method:   "GET",
		Request:  "Filter<IMemberCenter>",
		Response: "Paginated<IMemberCenter>",
		Note:     "Returns paginated member centers for the current user's branch.",
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
	req.RegisterRoute(horizon.Route{
		Route:    "/member-center",
		Method:   "POST",
		Request:  "TMemberCenter",
		Response: "TMemberCenter",
		Note:     "Creates a new member center record.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := c.model.MemberCenterManager.Validate(ctx)
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
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
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create member center: " + err.Error()})
		}

		return ctx.JSON(http.StatusOK, c.model.MemberCenterManager.ToModel(memberCenter))
	})

	// Update an existing member center by ID
	req.RegisterRoute(horizon.Route{
		Route:    "/member-center/:member_center_id",
		Method:   "PUT",
		Request:  "TMemberCenter",
		Response: "TMemberCenter",
		Note:     "Updates an existing member center record by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberCenterID, err := horizon.EngineUUIDParam(ctx, "member_center_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member_center_id: " + err.Error()})
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		req, err := c.model.MemberCenterManager.Validate(ctx)
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		memberCenter, err := c.model.MemberCenterManager.GetByID(context, *memberCenterID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Member center not found: " + err.Error()})
		}
		memberCenter.UpdatedAt = time.Now().UTC()
		memberCenter.UpdatedByID = user.UserID
		memberCenter.OrganizationID = user.OrganizationID
		memberCenter.BranchID = *user.BranchID
		memberCenter.Name = req.Name
		memberCenter.Description = req.Description
		if err := c.model.MemberCenterManager.UpdateFields(context, memberCenter.ID, memberCenter); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update member center: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.MemberCenterManager.ToModel(memberCenter))
	})

	// Delete a member center by ID
	req.RegisterRoute(horizon.Route{
		Route:  "/member-center/:member_center_id",
		Method: "DELETE",
		Note:   "Deletes a member center record by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberCenterID, err := horizon.EngineUUIDParam(ctx, "member_center_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member_center_id: " + err.Error()})
		}
		if err := c.model.MemberCenterManager.DeleteByID(context, *memberCenterID); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete member center: " + err.Error()})
		}
		return ctx.NoContent(http.StatusNoContent)
	})

	// Bulk delete member centers by IDs
	req.RegisterRoute(horizon.Route{
		Route:   "/member-center/bulk-delete",
		Method:  "DELETE",
		Request: "string[]",
		Note:    "Deletes multiple member center records by their IDs.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var reqBody struct {
			IDs []string `json:"ids"`
		}

		if err := ctx.Bind(&reqBody); err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body: " + err.Error()})
		}
		if len(reqBody.IDs) == 0 {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "No IDs provided for deletion."})
		}

		tx := c.provider.Service.Database.Client().Begin()
		if tx.Error != nil {
			tx.Rollback()
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to begin transaction: " + tx.Error.Error()})
		}

		for _, rawID := range reqBody.IDs {
			memberCenterID, err := uuid.Parse(rawID)
			if err != nil {
				tx.Rollback()
				return ctx.JSON(http.StatusBadRequest, map[string]string{"error": fmt.Sprintf("Invalid UUID '%s': %s", rawID, err.Error())})
			}
			if _, err := c.model.MemberCenterManager.GetByID(context, memberCenterID); err != nil {
				tx.Rollback()
				return ctx.JSON(http.StatusNotFound, map[string]string{"error": fmt.Sprintf("Member center with ID '%s' not found: %s", rawID, err.Error())})
			}
			if err := c.model.MemberCenterManager.DeleteByIDWithTx(context, tx, memberCenterID); err != nil {
				tx.Rollback()
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": fmt.Sprintf("Failed to delete member center with ID '%s': %s", rawID, err.Error())})
			}
		}

		if err := tx.Commit().Error; err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to commit transaction: " + err.Error()})
		}
		return ctx.NoContent(http.StatusNoContent)
	})
}
