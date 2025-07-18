package controller

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/lands-horizon/horizon-server/services/horizon"
	"github.com/lands-horizon/horizon-server/src/event"
	"github.com/lands-horizon/horizon-server/src/model"
)

// MemberAssetController manages endpoints for member asset records.
func (c *Controller) MemberAssetController() {
	req := c.provider.Service.Request

	// POST /member-asset/member-profile/:member_profile_id: Create a new asset record for a member.
	req.RegisterRoute(horizon.Route{
		Route:    "/member-asset/member-profile/:member_profile_id",
		Method:   "POST",
		Request:  "TMemberAsset",
		Response: "TMemberAsset",
		Note:     "Creates a new asset record for a member profile.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberProfileID, err := horizon.EngineUUIDParam(ctx, "member_profile_id")
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create member asset failed (/member-asset/member-profile/:member_profile_id), invalid member profile ID.",
				Module:      "MemberAsset",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member profile ID"})
		}
		req, err := c.model.MemberAssetManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create member asset failed (/member-asset/member-profile/:member_profile_id), validation error: " + err.Error(),
				Module:      "MemberAsset",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid asset data: " + err.Error()})
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create member asset failed (/member-asset/member-profile/:member_profile_id), user org error: " + err.Error(),
				Module:      "MemberAsset",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization/branch not found"})
		}
		if user.BranchID == nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create member asset failed (/member-asset/member-profile/:member_profile_id), user not assigned to branch.",
				Module:      "MemberAsset",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		value := &model.MemberAsset{
			MemberProfileID: memberProfileID,
			MediaID:         req.MediaID,
			Name:            req.Name,
			EntryDate:       req.EntryDate,
			Description:     req.Description,
			Cost:            req.Cost,
			CreatedAt:       time.Now().UTC(),
			CreatedByID:     user.UserID,
			UpdatedAt:       time.Now().UTC(),
			UpdatedByID:     user.UserID,
			BranchID:        *user.BranchID,
			OrganizationID:  user.OrganizationID,
		}
		if err := c.model.MemberAssetManager.Create(context, value); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create member asset failed (/member-asset/member-profile/:member_profile_id), db error: " + err.Error(),
				Module:      "MemberAsset",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create member asset record: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "create-success",
			Description: "Created member asset (/member-asset/member-profile/:member_profile_id): " + value.Name,
			Module:      "MemberAsset",
		})
		return ctx.JSON(http.StatusCreated, c.model.MemberAssetManager.ToModel(value))
	})

	// PUT /member-asset/:member_asset_id: Update an existing asset record for a member.
	req.RegisterRoute(horizon.Route{
		Route:    "/member-asset/:member_asset_id",
		Method:   "PUT",
		Request:  "TMemberAsset",
		Response: "TMemberAsset",
		Note:     "Updates an existing asset record for a member profile.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberAssetID, err := horizon.EngineUUIDParam(ctx, "member_asset_id")
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update member asset failed (/member-asset/:member_asset_id), invalid member asset ID.",
				Module:      "MemberAsset",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member asset ID"})
		}
		req, err := c.model.MemberAssetManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update member asset failed (/member-asset/:member_asset_id), validation error: " + err.Error(),
				Module:      "MemberAsset",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid asset data: " + err.Error()})
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update member asset failed (/member-asset/:member_asset_id), user org error: " + err.Error(),
				Module:      "MemberAsset",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization/branch not found"})
		}
		if user.BranchID == nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update member asset failed (/member-asset/:member_asset_id), user not assigned to branch.",
				Module:      "MemberAsset",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		value, err := c.model.MemberAssetManager.GetByID(context, *memberAssetID)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update member asset failed (/member-asset/:member_asset_id), record not found.",
				Module:      "MemberAsset",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Member asset record not found"})
		}
		value.UpdatedAt = time.Now().UTC()
		value.UpdatedByID = user.UserID
		value.OrganizationID = user.OrganizationID
		value.BranchID = *user.BranchID
		value.MediaID = req.MediaID
		value.Name = req.Name
		value.EntryDate = req.EntryDate
		value.Description = req.Description
		value.Cost = req.Cost

		if err := c.model.MemberAssetManager.UpdateFields(context, value.ID, value); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update member asset failed (/member-asset/:member_asset_id), db error: " + err.Error(),
				Module:      "MemberAsset",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update member asset record: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Updated member asset (/member-asset/:member_asset_id): " + value.Name,
			Module:      "MemberAsset",
		})
		return ctx.JSON(http.StatusOK, c.model.MemberAssetManager.ToModel(value))
	})

	// DELETE /member-asset/:member_asset_id: Delete a member's asset record by ID.
	req.RegisterRoute(horizon.Route{
		Route:  "/member-asset/:member_asset_id",
		Method: "DELETE",
		Note:   "Deletes a member's asset record by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberAssetID, err := horizon.EngineUUIDParam(ctx, "member_asset_id")
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete member asset failed (/member-asset/:member_asset_id), invalid member asset ID.",
				Module:      "MemberAsset",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member asset ID"})
		}
		value, err := c.model.MemberAssetManager.GetByID(context, *memberAssetID)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete member asset failed (/member-asset/:member_asset_id), record not found.",
				Module:      "MemberAsset",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Member asset record not found"})
		}
		if err := c.model.MemberAssetManager.DeleteByID(context, *memberAssetID); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete member asset failed (/member-asset/:member_asset_id), db error: " + err.Error(),
				Module:      "MemberAsset",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete member asset record: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "delete-success",
			Description: "Deleted member asset (/member-asset/:member_asset_id): " + value.Name,
			Module:      "MemberAsset",
		})
		return ctx.NoContent(http.StatusNoContent)
	})
}
