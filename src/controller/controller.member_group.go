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

func (c *Controller) MemberGroupController() {
	req := c.provider.Service.Request

	req.RegisterRoute(horizon.Route{
		Route:    "/member-group-history",
		Method:   "GET",
		Response: "TMemberGroupHistory[]",
		Note:     "Get member group history for the current branch",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.NoContent(http.StatusNoContent)
		}
		memberGroupHistory, err := c.model.MemberGroupHistoryCurrentBranch(context, user.OrganizationID, *user.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.MemberGroupHistoryManager.ToModels(memberGroupHistory))
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/member-group-history/member-profile/:member_profile_id/search",
		Method:   "GET",
		Response: "TMemberGroupHistory[]",
		Note:     "Get member group history by member profile ID",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberProfileID, err := horizon.EngineUUIDParam(ctx, "member_profile_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid member group ID")
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.NoContent(http.StatusNoContent)
		}
		memberGroupHistory, err := c.model.MemberGroupHistoryMemberProfileID(context, *memberProfileID, user.OrganizationID, *user.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.MemberGroupHistoryManager.Pagination(context, ctx, memberGroupHistory))

	})

	req.RegisterRoute(horizon.Route{
		Route:    "/member-group",
		Method:   "GET",
		Response: "TMemberGroup[]",
		Note:     "Get all member group records for the current branch",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.NoContent(http.StatusNoContent)
		}
		memberGroup, err := c.model.MemberGroupCurrentBranch(context, user.OrganizationID, *user.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.MemberGroupManager.ToModels(memberGroup))
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/member-group/search",
		Method:   "GET",
		Request:  "Filter<IMemberGroup>",
		Response: "Paginated<IMemberGroup>",
		Note:     "Get pagination member group",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.NoContent(http.StatusNoContent)
		}
		value, err := c.model.MemberGroupCurrentBranch(context, user.OrganizationID, *user.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.MemberGroupManager.Pagination(context, ctx, value))
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/member-group",
		Method:   "POST",
		Request:  "TMemberGroup",
		Response: "TMemberGroup",
		Note:     "Create a new member group record",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := c.model.MemberGroupManager.Validate(ctx)
		if err != nil {
			return c.BadRequest(ctx, err.Error())
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.NoContent(http.StatusNoContent)
		}

		memberGroup := &model.MemberGroup{
			Name:        req.Name,
			Description: req.Description,

			CreatedAt:      time.Now().UTC(),
			CreatedByID:    user.UserID,
			UpdatedAt:      time.Now().UTC(),
			UpdatedByID:    user.UserID,
			BranchID:       *user.BranchID,
			OrganizationID: user.OrganizationID,
		}

		if err := c.model.MemberGroupManager.Create(context, memberGroup); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})

		}

		return ctx.JSON(http.StatusOK, c.model.MemberGroupManager.ToModel(memberGroup))
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/member-group/:member_group_id",
		Method:   "PUT",
		Request:  "TMemberGroup",
		Response: "TMemberGroup",
		Note:     "Update an existing member group record by ID",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberGroupID, err := horizon.EngineUUIDParam(ctx, "member_group_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid member group ID")
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.NoContent(http.StatusNoContent)
		}

		req, err := c.model.MemberGroupManager.Validate(ctx)
		if err != nil {
			return c.BadRequest(ctx, err.Error())
		}

		memberGroup, err := c.model.MemberGroupManager.GetByID(context, *memberGroupID)
		if err != nil {
			return c.NotFound(ctx, "MemberGroup")
		}

		memberGroup.UpdatedAt = time.Now().UTC()
		memberGroup.UpdatedByID = user.UserID
		memberGroup.OrganizationID = user.OrganizationID
		memberGroup.BranchID = *user.BranchID
		memberGroup.Name = req.Name
		memberGroup.Description = req.Description
		if err := c.model.MemberGroupManager.UpdateFields(context, memberGroup.ID, memberGroup); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})

		}
		return ctx.JSON(http.StatusOK, c.model.MemberGroupManager.ToModel(memberGroup))
	})

	req.RegisterRoute(horizon.Route{
		Route:  "/member-group/:member_group_id",
		Method: "DELETE",
		Note:   "Delete a member group record by ID",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberGroupID, err := horizon.EngineUUIDParam(ctx, "member_group_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid member group ID")
		}
		if err := c.model.MemberGroupManager.DeleteByID(context, *memberGroupID); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})

		}
		return ctx.NoContent(http.StatusNoContent)
	})

	req.RegisterRoute(horizon.Route{
		Route:   "/member-group/bulk-delete",
		Method:  "DELETE",
		Request: "string[]",
		Note:    "Delete multiple member group records by their IDs",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var reqBody struct {
			IDs []string `json:"ids"`
		}

		if err := ctx.Bind(&reqBody); err != nil {
			return c.BadRequest(ctx, "Invalid request body")
		}

		if len(reqBody.IDs) == 0 {
			return c.BadRequest(ctx, "No IDs provided")
		}

		tx := c.provider.Service.Database.Client().Begin()
		if tx.Error != nil {
			tx.Rollback()
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": tx.Error.Error()})
		}

		for _, rawID := range reqBody.IDs {
			memberGroupID, err := uuid.Parse(rawID)
			if err != nil {
				tx.Rollback()
				return c.BadRequest(ctx, fmt.Sprintf("Invalid UUID: %s", rawID))
			}

			if _, err := c.model.MemberGroupManager.GetByID(context, memberGroupID); err != nil {
				tx.Rollback()
				return c.NotFound(ctx, fmt.Sprintf("MemberGroup with ID %s", rawID))
			}

			if err := c.model.MemberGroupManager.DeleteByIDWithTx(context, tx, memberGroupID); err != nil {
				tx.Rollback()
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})

			}
		}

		if err := tx.Commit().Error; err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})

		}

		return ctx.NoContent(http.StatusNoContent)
	})
}
