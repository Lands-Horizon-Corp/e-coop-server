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

	// Create a new relative account record for a member profile
	req.RegisterRoute(horizon.Route{
		Route:    "/member-relative-account/member-profile/:member_profile_id",
		Method:   "POST",
		Request:  "TMemberRelativeAccount",
		Response: "TMemberRelativeAccount",
		Note:     "Creates a new relative account record for the specified member profile.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberProfileID, err := horizon.EngineUUIDParam(ctx, "member_profile_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member_profile_id: " + err.Error()})
		}
		req, err := c.model.MemberRelativeAccountManager.Validate(ctx)
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
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
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create relative account record: " + err.Error()})
		}

		return ctx.JSON(http.StatusOK, c.model.MemberRelativeAccountManager.ToModel(value))
	})

	// Update an existing relative account record by its ID
	req.RegisterRoute(horizon.Route{
		Route:    "/member-relative-account/:member_relative_account_id",
		Method:   "PUT",
		Request:  "TMemberRelativeAccount",
		Response: "TMemberRelativeAccount",
		Note:     "Updates an existing relative account record by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberRelativeAccountID, err := horizon.EngineUUIDParam(ctx, "member_relative_account_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member_relative_account_id: " + err.Error()})
		}
		req, err := c.model.MemberRelativeAccountManager.Validate(ctx)
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}

		value, err := c.model.MemberRelativeAccountManager.GetByID(context, *memberRelativeAccountID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Relative account record not found: " + err.Error()})
		}

		value.UpdatedAt = time.Now().UTC()
		value.UpdatedByID = user.UserID
		value.OrganizationID = user.OrganizationID
		value.BranchID = *user.BranchID
		value.MemberProfileID = req.MemberProfileID
		value.RelativeMemberProfileID = req.RelativeMemberProfileID
		value.FamilyRelationship = req.FamilyRelationship
		value.Description = req.Description

		if err := c.model.MemberRelativeAccountManager.UpdateFields(context, value.ID, value); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update relative account record: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.MemberRelativeAccountManager.ToModel(value))
	})

	// Delete a member's relative account record by its ID
	req.RegisterRoute(horizon.Route{
		Route:  "/member-relative-account/:member_relative_account_id",
		Method: "DELETE",
		Note:   "Deletes a relative account record by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberRelativeAccountID, err := horizon.EngineUUIDParam(ctx, "member_relative_account_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member_relative_account_id: " + err.Error()})
		}
		if err := c.model.MemberRelativeAccountManager.DeleteByID(context, *memberRelativeAccountID); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete relative account record: " + err.Error()})
		}
		return ctx.NoContent(http.StatusNoContent)
	})
}
