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

	req.RegisterRoute(horizon.Route{
		Route:    "/member-center-history",
		Method:   "GET",
		Response: "TMemberCenterHistory[]",
		Note:     "Get member center history for the current branch",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.NoContent(http.StatusNoContent)
		}
		memberCenterHistory, err := c.model.MemberCenterHistoryCurrentBranch(context, user.OrganizationID, *user.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.MemberCenterHistoryManager.ToModels(memberCenterHistory))
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/member-center-history/member-profile/:member_profile_id/search",
		Method:   "GET",
		Response: "TMemberCenterHistory[]",
		Note:     "Get member center history by member profile ID",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberProfileID, err := horizon.EngineUUIDParam(ctx, "member_profile_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid member center ID")
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.NoContent(http.StatusNoContent)
		}
		memberCenterHistory, err := c.model.MemberCenterHistoryMemberProfileID(context, *memberProfileID, user.OrganizationID, *user.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.MemberCenterHistoryManager.Pagination(context, ctx, memberCenterHistory))
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/member-center",
		Method:   "GET",
		Response: "TMemberCenter[]",
		Note:     "Get all member center records for the current branch",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.NoContent(http.StatusNoContent)
		}
		memberCenter, err := c.model.MemberCenterCurrentBranch(context, user.OrganizationID, *user.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.MemberCenterManager.ToModels(memberCenter))
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/member-center/search",
		Method:   "GET",
		Request:  "Filter<IMemberCenter>",
		Response: "Paginated<IMemberCenter>",
		Note:     "Get pagination member center",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.NoContent(http.StatusNoContent)
		}
		value, err := c.model.MemberCenterCurrentBranch(context, user.OrganizationID, *user.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.MemberCenterManager.Pagination(context, ctx, value))
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/member-center",
		Method:   "POST",
		Request:  "TMemberCenter",
		Response: "TMemberCenter",
		Note:     "Create a new member center record",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := c.model.MemberCenterManager.Validate(ctx)
		if err != nil {
			return c.BadRequest(ctx, err.Error())
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.NoContent(http.StatusNoContent)
		}

		memberCenter := &model.MemberCenter{
			Name:        req.Name,
			Description: req.Description,

			CreatedAt:      time.Now().UTC(),
			CreatedByID:    user.UserID,
			UpdatedAt:      time.Now().UTC(),
			UpdatedByID:    user.UserID,
			BranchID:       *user.BranchID,
			OrganizationID: user.OrganizationID,
		}

		if err := c.model.MemberCenterManager.Create(context, memberCenter); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})

		}

		return ctx.JSON(http.StatusOK, c.model.MemberCenterManager.ToModel(memberCenter))
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/member-center/:member_center_id",
		Method:   "PUT",
		Request:  "TMemberCenter",
		Response: "TMemberCenter",
		Note:     "Update an existing member center record by ID",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberCenterID, err := horizon.EngineUUIDParam(ctx, "member_center_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid member center ID")
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.NoContent(http.StatusNoContent)
		}

		req, err := c.model.MemberCenterManager.Validate(ctx)
		if err != nil {
			return c.BadRequest(ctx, err.Error())
		}

		memberCenter, err := c.model.MemberCenterManager.GetByID(context, *memberCenterID)
		if err != nil {
			return c.NotFound(ctx, "MemberCenter")
		}

		memberCenter.UpdatedAt = time.Now().UTC()
		memberCenter.UpdatedByID = user.UserID
		memberCenter.OrganizationID = user.OrganizationID
		memberCenter.BranchID = *user.BranchID
		memberCenter.Name = req.Name
		memberCenter.Description = req.Description
		if err := c.model.MemberCenterManager.UpdateFields(context, memberCenter.ID, memberCenter); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})

		}
		return ctx.JSON(http.StatusOK, c.model.MemberCenterManager.ToModel(memberCenter))
	})

	req.RegisterRoute(horizon.Route{
		Route:  "/member-center/:member_center_id",
		Method: "DELETE",
		Note:   "Delete a member center record by ID",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberCenterID, err := horizon.EngineUUIDParam(ctx, "member_center_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid member center ID")
		}
		if err := c.model.MemberCenterManager.DeleteByID(context, *memberCenterID); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})

		}
		return ctx.NoContent(http.StatusNoContent)
	})

	req.RegisterRoute(horizon.Route{
		Route:   "/member-center/bulk-delete",
		Method:  "DELETE",
		Request: "string[]",
		Note:    "Delete multiple member center records by their IDs",
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
			memberCenterID, err := uuid.Parse(rawID)
			if err != nil {
				tx.Rollback()
				return c.BadRequest(ctx, fmt.Sprintf("Invalid UUID: %s", rawID))
			}

			if _, err := c.model.MemberCenterManager.GetByID(context, memberCenterID); err != nil {
				tx.Rollback()
				return c.NotFound(ctx, fmt.Sprintf("MemberCenter with ID %s", rawID))
			}

			if err := c.model.MemberCenterManager.DeleteByIDWithTx(context, tx, memberCenterID); err != nil {
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
