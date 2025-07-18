package controller

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/lands-horizon/horizon-server/services/horizon"
	"github.com/lands-horizon/horizon-server/src/model"
)

func (c *Controller) MemberRelativeAccountController() {
	req := c.provider.Service.Request

	req.RegisterRoute(horizon.Route{
		Route:    "/member-relative-account/member-profile/:member_profile_id",
		Method:   "POST",
		Request:  "TMemberRelativeAccount",
		Response: "TMemberRelativeAccount",
		Note:     "Create a new relative account record for a member.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberProfileID, err := horizon.EngineUUIDParam(ctx, "member_profile_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid member profile ID")
		}
		req, err := c.model.MemberRelativeAccountManager.Validate(ctx)
		if err != nil {
			return c.BadRequest(ctx, err.Error())
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.NoContent(http.StatusNoContent)
		}

		value := &model.MemberRelativeAccount{
			MemberProfileID:         *memberProfileID,
			RelativeMemberProfileID: req.RelativeMemberProfileID,
			FamilyRelationship:      req.FamilyRelationship,
			Description:             req.Description,
			CreatedAt:               time.Now().UTC(),
			CreatedByID:             user.UserID,
			UpdatedAt:               time.Now().UTC(),
			UpdatedByID:             user.UserID,
			BranchID:                *user.BranchID,
			OrganizationID:          user.OrganizationID,
		}

		if err := c.model.MemberRelativeAccountManager.Create(context, value); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})

		}

		return ctx.JSON(http.StatusOK, c.model.MemberRelativeAccountManager.ToModel(value))
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/member-relative-account/:member_relative_account_id",
		Method:   "PUT",
		Request:  "TMemberRelativeAccount",
		Response: "TMemberRelativeAccount",
		Note:     "Update an existing relative account record for a member.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberRelativeAccountID, err := horizon.EngineUUIDParam(ctx, "member_relative_account_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid member relative account ID")
		}
		req, err := c.model.MemberRelativeAccountManager.Validate(ctx)
		if err != nil {
			return c.BadRequest(ctx, err.Error())
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.NoContent(http.StatusNoContent)
		}

		value, err := c.model.MemberRelativeAccountManager.GetByID(context, *memberRelativeAccountID)
		if err != nil {
			return c.NotFound(ctx, "MemberRelativeAccount")
		}

		value.UpdatedAt = time.Now().UTC()
		value.UpdatedByID = user.UserID
		value.OrganizationID = user.OrganizationID
		value.BranchID = *user.BranchID
		value.MemberProfileID = req.MemberProfileID
		value.RelativeMemberProfileID = req.RelativeMemberProfileID
		value.FamilyRelationship = req.FamilyRelationship
		value.Description = req.Description
		value.FamilyRelationship = req.FamilyRelationship

		if err := c.model.MemberRelativeAccountManager.UpdateFields(context, value.ID, value); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})

		}
		return ctx.JSON(http.StatusOK, c.model.MemberRelativeAccountManager.ToModel(value))
	})

	req.RegisterRoute(horizon.Route{
		Route:  "/member-relative-account/:member_relative_account_id",
		Method: "DELETE",
		Note:   "Delete a member's relative account record by ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberRelativeAccountID, err := horizon.EngineUUIDParam(ctx, "member_relative_account_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid member relative account ID")
		}
		if err := c.model.MemberRelativeAccountManager.DeleteByID(context, *memberRelativeAccountID); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})

		}
		return ctx.NoContent(http.StatusNoContent)
	})
}
