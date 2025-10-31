package controller_v1

import (
	"net/http"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/services/handlers"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/event"
	modelCore "github.com/Lands-Horizon-Corp/e-coop-server/src/model/modelCore"
	"github.com/labstack/echo/v4"
)

func (c *Controller) MemberRelativeAccountController() {
	req := c.provider.Service.Request

	// Create a new relative account record for a member profile
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/member-relative-account/member-profile/:member_profile_id",
		Method:       "POST",
		RequestType:  modelCore.MemberRelativeAccountRequest{},
		ResponseType: modelCore.MemberRelativeAccountResponse{},
		Note:         "Creates a new relative account record for the specified member profile.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberProfileID, err := handlers.EngineUUIDParam(ctx, "member_profile_id")
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create member relative account failed: invalid member_profile_id: " + err.Error(),
				Module:      "MemberRelativeAccount",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member_profile_id: " + err.Error()})
		}
		req, err := c.modelCore.MemberRelativeAccountManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create member relative account failed: validation error: " + err.Error(),
				Module:      "MemberRelativeAccount",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create member relative account failed: user org error: " + err.Error(),
				Module:      "MemberRelativeAccount",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}

		value := &modelCore.MemberRelativeAccount{
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

		if err := c.modelCore.MemberRelativeAccountManager.Create(context, value); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create member relative account failed: " + err.Error(),
				Module:      "MemberRelativeAccount",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create relative account record: " + err.Error()})
		}

		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "create-success",
			Description: "Created member relative account for member_profile_id: " + memberProfileID.String(),
			Module:      "MemberRelativeAccount",
		})

		return ctx.JSON(http.StatusOK, c.modelCore.MemberRelativeAccountManager.ToModel(value))
	})

	// Update an existing relative account record by its ID
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/member-relative-account/:member_relative_account_id",
		Method:       "PUT",
		RequestType:  modelCore.MemberRelativeAccountRequest{},
		ResponseType: modelCore.MemberRelativeAccountResponse{},
		Note:         "Updates an existing relative account record by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberRelativeAccountID, err := handlers.EngineUUIDParam(ctx, "member_relative_account_id")
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update member relative account failed: invalid member_relative_account_id: " + err.Error(),
				Module:      "MemberRelativeAccount",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member_relative_account_id: " + err.Error()})
		}
		req, err := c.modelCore.MemberRelativeAccountManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update member relative account failed: validation error: " + err.Error(),
				Module:      "MemberRelativeAccount",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update member relative account failed: user org error: " + err.Error(),
				Module:      "MemberRelativeAccount",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}

		value, err := c.modelCore.MemberRelativeAccountManager.GetByID(context, *memberRelativeAccountID)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update member relative account failed: record not found: " + err.Error(),
				Module:      "MemberRelativeAccount",
			})
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

		if err := c.modelCore.MemberRelativeAccountManager.UpdateFields(context, value.ID, value); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update member relative account failed: update error: " + err.Error(),
				Module:      "MemberRelativeAccount",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update relative account record: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Updated member relative account ID: " + memberRelativeAccountID.String(),
			Module:      "MemberRelativeAccount",
		})
		return ctx.JSON(http.StatusOK, c.modelCore.MemberRelativeAccountManager.ToModel(value))
	})

	// Delete a member's relative account record by its ID
	req.RegisterRoute(handlers.Route{
		Route:  "/api/v1/member-relative-account/:member_relative_account_id",
		Method: "DELETE",
		Note:   "Deletes a relative account record by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberRelativeAccountID, err := handlers.EngineUUIDParam(ctx, "member_relative_account_id")
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete member relative account failed: invalid member_relative_account_id: " + err.Error(),
				Module:      "MemberRelativeAccount",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member_relative_account_id: " + err.Error()})
		}
		if err := c.modelCore.MemberRelativeAccountManager.DeleteByID(context, *memberRelativeAccountID); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete member relative account failed: " + err.Error(),
				Module:      "MemberRelativeAccount",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete relative account record: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "delete-success",
			Description: "Deleted member relative account ID: " + memberRelativeAccountID.String(),
			Module:      "MemberRelativeAccount",
		})
		return ctx.NoContent(http.StatusNoContent)
	})
}
