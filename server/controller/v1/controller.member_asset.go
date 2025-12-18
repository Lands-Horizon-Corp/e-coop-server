package v1

import (
	"net/http"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/server/event"
	"github.com/Lands-Horizon-Corp/e-coop-server/server/model/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/services/handlers"
	"github.com/labstack/echo/v4"
)

func (c *Controller) memberAssetController() {
	req := c.provider.Service.Request

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/member-asset/member-profile/:member_profile_id",
		Method:       "POST",
		RequestType:  core.MemberAsset{},
		ResponseType: core.MemberAsset{},
		Note:         "Creates a new asset record for a member profile.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberProfileID, err := handlers.EngineUUIDParam(ctx, "member_profile_id")
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create member asset failed (/member-asset/member-profile/:member_profile_id), invalid member profile ID.",
				Module:      "MemberAsset",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member profile ID"})
		}
		req, err := c.core.MemberAssetManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create member asset failed (/member-asset/member-profile/:member_profile_id), validation error: " + err.Error(),
				Module:      "MemberAsset",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid asset data: " + err.Error()})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create member asset failed (/member-asset/member-profile/:member_profile_id), user org error: " + err.Error(),
				Module:      "MemberAsset",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization/branch not found"})
		}
		if userOrg.BranchID == nil {
			c.event.Footstep(ctx, event.FootstepEvent{
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
		if err := c.core.MemberAssetManager.Create(context, value); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create member asset failed (/member-asset/member-profile/:member_profile_id), db error: " + err.Error(),
				Module:      "MemberAsset",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create member asset record: " + err.Error()})
		}
		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "create-success",
			Description: "Created member asset (/member-asset/member-profile/:member_profile_id): " + value.Name,
			Module:      "MemberAsset",
		})
		return ctx.JSON(http.StatusCreated, c.core.MemberAssetManager.ToModel(value))
	})

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/member-asset/:member_asset_id",
		Method:       "PUT",
		RequestType:  core.MemberAsset{},
		ResponseType: core.MemberAsset{},
		Note:         "Updates an existing asset record for a member profile.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberAssetID, err := handlers.EngineUUIDParam(ctx, "member_asset_id")
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update member asset failed (/member-asset/:member_asset_id), invalid member asset ID.",
				Module:      "MemberAsset",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member asset ID"})
		}
		req, err := c.core.MemberAssetManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update member asset failed (/member-asset/:member_asset_id), validation error: " + err.Error(),
				Module:      "MemberAsset",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid asset data: " + err.Error()})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update member asset failed (/member-asset/:member_asset_id), user org error: " + err.Error(),
				Module:      "MemberAsset",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization/branch not found"})
		}
		if userOrg.BranchID == nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update member asset failed (/member-asset/:member_asset_id), user not assigned to branch.",
				Module:      "MemberAsset",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		value, err := c.core.MemberAssetManager.GetByID(context, *memberAssetID)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
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

		if err := c.core.MemberAssetManager.UpdateByID(context, value.ID, value); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update member asset failed (/member-asset/:member_asset_id), db error: " + err.Error(),
				Module:      "MemberAsset",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update member asset record: " + err.Error()})
		}
		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Updated member asset (/member-asset/:member_asset_id): " + value.Name,
			Module:      "MemberAsset",
		})
		return ctx.JSON(http.StatusOK, c.core.MemberAssetManager.ToModel(value))
	})

	req.RegisterWebRoute(handlers.Route{
		Route:  "/api/v1/member-asset/:member_asset_id",
		Method: "DELETE",
		Note:   "Deletes a member's asset record by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberAssetID, err := handlers.EngineUUIDParam(ctx, "member_asset_id")
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete member asset failed (/member-asset/:member_asset_id), invalid member asset ID.",
				Module:      "MemberAsset",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member asset ID"})
		}
		value, err := c.core.MemberAssetManager.GetByID(context, *memberAssetID)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete member asset failed (/member-asset/:member_asset_id), record not found.",
				Module:      "MemberAsset",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Member asset record not found"})
		}
		if err := c.core.MemberAssetManager.Delete(context, *memberAssetID); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete member asset failed (/member-asset/:member_asset_id), db error: " + err.Error(),
				Module:      "MemberAsset",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete member asset record: " + err.Error()})
		}
		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "delete-success",
			Description: "Deleted member asset (/member-asset/:member_asset_id): " + value.Name,
			Module:      "MemberAsset",
		})
		return ctx.NoContent(http.StatusNoContent)
	})
}
