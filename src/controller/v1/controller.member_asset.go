package v1

import (
	"net/http"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/helpers"
	"github.com/Lands-Horizon-Corp/e-coop-server/horizon"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/event"
	"github.com/labstack/echo/v4"
)

func memberAssetController(service *horizon.HorizonService) {
	req := service.API

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/member-asset/member-profile/:member_profile_id",
		Method:       "POST",
		RequestType:  core.MemberAsset{},
		ResponseType: core.MemberAsset{},
		Note:         "Creates a new asset record for a member profile.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberProfileID, err := helpers.EngineUUIDParam(ctx, "member_profile_id")
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create member asset failed (/member-asset/member-profile/:member_profile_id), invalid member profile ID.",
				Module:      "MemberAsset",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member profile ID"})
		}
		req, err := core.MemberAssetManager(service).Validate(ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create member asset failed (/member-asset/member-profile/:member_profile_id), validation error: " + err.Error(),
				Module:      "MemberAsset",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid asset data: " + err.Error()})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create member asset failed (/member-asset/member-profile/:member_profile_id), user org error: " + err.Error(),
				Module:      "MemberAsset",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization/branch not found"})
		}
		if userOrg.BranchID == nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create member asset failed (/member-asset/member-profile/:member_profile_id), user not assigned to branch.",
				Module:      "MemberAsset",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		value := &core.MemberAsset{
			MemberProfileID: memberProfileID,
			MediaID:         req.MediaID,
			Name:            req.Name,
			EntryDate:       req.EntryDate,
			Description:     req.Description,
			Cost:            req.Cost,
			CreatedAt:       time.Now().UTC(),
			CreatedByID:     userOrg.UserID,
			UpdatedAt:       time.Now().UTC(),
			UpdatedByID:     userOrg.UserID,
			BranchID:        *userOrg.BranchID,
			OrganizationID:  userOrg.OrganizationID,
		}
		if err := core.MemberAssetManager(service).Create(context, value); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create member asset failed (/member-asset/member-profile/:member_profile_id), db error: " + err.Error(),
				Module:      "MemberAsset",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create member asset record: " + err.Error()})
		}
		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "create-success",
			Description: "Created member asset (/member-asset/member-profile/:member_profile_id): " + value.Name,
			Module:      "MemberAsset",
		})
		return ctx.JSON(http.StatusCreated, core.MemberAssetManager(service).ToModel(value))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/member-asset/:member_asset_id",
		Method:       "PUT",
		RequestType:  core.MemberAsset{},
		ResponseType: core.MemberAsset{},
		Note:         "Updates an existing asset record for a member profile.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberAssetID, err := helpers.EngineUUIDParam(ctx, "member_asset_id")
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update member asset failed (/member-asset/:member_asset_id), invalid member asset ID.",
				Module:      "MemberAsset",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member asset ID"})
		}
		req, err := core.MemberAssetManager(service).Validate(ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update member asset failed (/member-asset/:member_asset_id), validation error: " + err.Error(),
				Module:      "MemberAsset",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid asset data: " + err.Error()})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update member asset failed (/member-asset/:member_asset_id), user org error: " + err.Error(),
				Module:      "MemberAsset",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization/branch not found"})
		}
		if userOrg.BranchID == nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update member asset failed (/member-asset/:member_asset_id), user not assigned to branch.",
				Module:      "MemberAsset",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		value, err := core.MemberAssetManager(service).GetByID(context, *memberAssetID)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update member asset failed (/member-asset/:member_asset_id), record not found.",
				Module:      "MemberAsset",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Member asset record not found"})
		}
		value.UpdatedAt = time.Now().UTC()
		value.UpdatedByID = userOrg.UserID
		value.OrganizationID = userOrg.OrganizationID
		value.BranchID = *userOrg.BranchID
		value.MediaID = req.MediaID
		value.Name = req.Name
		value.EntryDate = req.EntryDate
		value.Description = req.Description
		value.Cost = req.Cost

		if err := core.MemberAssetManager(service).UpdateByID(context, value.ID, value); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update member asset failed (/member-asset/:member_asset_id), db error: " + err.Error(),
				Module:      "MemberAsset",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update member asset record: " + err.Error()})
		}
		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Updated member asset (/member-asset/:member_asset_id): " + value.Name,
			Module:      "MemberAsset",
		})
		return ctx.JSON(http.StatusOK, core.MemberAssetManager(service).ToModel(value))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:  "/api/v1/member-asset/:member_asset_id",
		Method: "DELETE",
		Note:   "Deletes a member's asset record by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberAssetID, err := helpers.EngineUUIDParam(ctx, "member_asset_id")
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete member asset failed (/member-asset/:member_asset_id), invalid member asset ID.",
				Module:      "MemberAsset",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member asset ID"})
		}
		value, err := core.MemberAssetManager(service).GetByID(context, *memberAssetID)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete member asset failed (/member-asset/:member_asset_id), record not found.",
				Module:      "MemberAsset",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Member asset record not found"})
		}
		if err := core.MemberAssetManager(service).Delete(context, *memberAssetID); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete member asset failed (/member-asset/:member_asset_id), db error: " + err.Error(),
				Module:      "MemberAsset",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete member asset record: " + err.Error()})
		}
		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "delete-success",
			Description: "Deleted member asset (/member-asset/:member_asset_id): " + value.Name,
			Module:      "MemberAsset",
		})
		return ctx.NoContent(http.StatusNoContent)
	})
}
