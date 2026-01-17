package member_profile

import (
	"net/http"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/helpers"
	"github.com/Lands-Horizon-Corp/e-coop-server/horizon"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/event"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/types"
	"github.com/labstack/echo/v4"
)

func MemberRelativeAccountController(service *horizon.HorizonService) {
	req := service.API

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/member-relative-account/member-profile/:member_profile_id",
		Method:       "POST",
		RequestType: types.MemberRelativeAccountRequest{},
		ResponseType: types.MemberRelativeAccountResponse{},
		Note:         "Creates a new relative account record for the specified member profile.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberProfileID, err := helpers.EngineUUIDParam(ctx, "member_profile_id")
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create member relative account failed: invalid member_profile_id: " + err.Error(),
				Module:      "MemberRelativeAccount",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member_profile_id: " + err.Error()})
		}
		req, err := core.MemberRelativeAccountManager(service).Validate(ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create member relative account failed: validation error: " + err.Error(),
				Module:      "MemberRelativeAccount",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create member relative account failed: user org error: " + err.Error(),
				Module:      "MemberRelativeAccount",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}

		value := &types.MemberRelativeAccount{
			MemberProfileID:         *memberProfileID,
			RelativeMemberProfileID: req.RelativeMemberProfileID,
			FamilyRelationship:      req.FamilyRelationship,
			Description:             req.Description,
			CreatedAt:               time.Now().UTC(),
			CreatedByID:             userOrg.UserID,
			UpdatedAt:               time.Now().UTC(),
			UpdatedByID:             userOrg.UserID,
			BranchID:                *userOrg.BranchID,
			OrganizationID:          userOrg.OrganizationID,
		}

		if err := core.MemberRelativeAccountManager(service).Create(context, value); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create member relative account failed: " + err.Error(),
				Module:      "MemberRelativeAccount",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create relative account record: " + err.Error()})
		}

		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "create-success",
			Description: "Created member relative account for member_profile_id: " + memberProfileID.String(),
			Module:      "MemberRelativeAccount",
		})

		return ctx.JSON(http.StatusOK, core.MemberRelativeAccountManager(service).ToModel(value))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/member-relative-account/:member_relative_account_id",
		Method:       "PUT",
		RequestType: types.MemberRelativeAccountRequest{},
		ResponseType: types.MemberRelativeAccountResponse{},
		Note:         "Updates an existing relative account record by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberRelativeAccountID, err := helpers.EngineUUIDParam(ctx, "member_relative_account_id")
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update member relative account failed: invalid member_relative_account_id: " + err.Error(),
				Module:      "MemberRelativeAccount",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member_relative_account_id: " + err.Error()})
		}
		req, err := core.MemberRelativeAccountManager(service).Validate(ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update member relative account failed: validation error: " + err.Error(),
				Module:      "MemberRelativeAccount",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update member relative account failed: user org error: " + err.Error(),
				Module:      "MemberRelativeAccount",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}

		value, err := core.MemberRelativeAccountManager(service).GetByID(context, *memberRelativeAccountID)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update member relative account failed: record not found: " + err.Error(),
				Module:      "MemberRelativeAccount",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Relative account record not found: " + err.Error()})
		}

		value.UpdatedAt = time.Now().UTC()
		value.UpdatedByID = userOrg.UserID
		value.OrganizationID = userOrg.OrganizationID
		value.BranchID = *userOrg.BranchID
		value.MemberProfileID = req.MemberProfileID
		value.RelativeMemberProfileID = req.RelativeMemberProfileID
		value.FamilyRelationship = req.FamilyRelationship
		value.Description = req.Description

		if err := core.MemberRelativeAccountManager(service).UpdateByID(context, value.ID, value); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update member relative account failed: update error: " + err.Error(),
				Module:      "MemberRelativeAccount",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update relative account record: " + err.Error()})
		}
		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Updated member relative account ID: " + memberRelativeAccountID.String(),
			Module:      "MemberRelativeAccount",
		})
		return ctx.JSON(http.StatusOK, core.MemberRelativeAccountManager(service).ToModel(value))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:  "/api/v1/member-relative-account/:member_relative_account_id",
		Method: "DELETE",
		Note:   "Deletes a relative account record by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberRelativeAccountID, err := helpers.EngineUUIDParam(ctx, "member_relative_account_id")
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete member relative account failed: invalid member_relative_account_id: " + err.Error(),
				Module:      "MemberRelativeAccount",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member_relative_account_id: " + err.Error()})
		}
		if err := core.MemberRelativeAccountManager(service).Delete(context, *memberRelativeAccountID); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete member relative account failed: " + err.Error(),
				Module:      "MemberRelativeAccount",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete relative account record: " + err.Error()})
		}
		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "delete-success",
			Description: "Deleted member relative account ID: " + memberRelativeAccountID.String(),
			Module:      "MemberRelativeAccount",
		})
		return ctx.NoContent(http.StatusNoContent)
	})
}
