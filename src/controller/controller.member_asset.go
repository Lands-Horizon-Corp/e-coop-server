package controller

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/lands-horizon/horizon-server/services/horizon"
	"github.com/lands-horizon/horizon-server/src/model"
)

func (c *Controller) MemberAssetController() {
	req := c.provider.Service.Request

	req.RegisterRoute(horizon.Route{
		Route:    "/member-asset/member-profile/:member_profile_id",
		Method:   "POST",
		Request:  "TMemberAsset",
		Response: "TMemberAsset",
		Note:     "Create a new asset record for a member.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberProfileID, err := horizon.EngineUUIDParam(ctx, "member_profile_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid member profile ID")
		}
		req, err := c.model.MemberAssetManager.Validate(ctx)
		if err != nil {
			return c.BadRequest(ctx, err.Error())
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.NoContent(http.StatusNoContent)
		}

		value := &model.MemberAsset{
			MemberProfileID: memberProfileID,

			MediaID:     req.MediaID,
			Name:        req.Name,
			EntryDate:   req.EntryDate,
			Description: req.Description,
			Cost:        req.Cost,

			CreatedAt:      time.Now().UTC(),
			CreatedByID:    user.UserID,
			UpdatedAt:      time.Now().UTC(),
			UpdatedByID:    user.UserID,
			BranchID:       *user.BranchID,
			OrganizationID: user.OrganizationID,
		}

		if err := c.model.MemberAssetManager.Create(context, value); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})

		}

		return ctx.JSON(http.StatusOK, c.model.MemberAssetManager.ToModel(value))
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/member-asset/:member_asset_id",
		Method:   "PUT",
		Request:  "TMemberAsset",
		Response: "TMemberAsset",
		Note:     "Update an existing asset record for a member.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberAssetID, err := horizon.EngineUUIDParam(ctx, "member_asset_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid member asset ID")
		}
		req, err := c.model.MemberAssetManager.Validate(ctx)
		if err != nil {
			return c.BadRequest(ctx, err.Error())
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.NoContent(http.StatusNoContent)
		}

		value, err := c.model.MemberAssetManager.GetByID(context, *memberAssetID)
		if err != nil {
			return c.NotFound(ctx, "MemberAsset")
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
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})

		}
		return ctx.JSON(http.StatusOK, c.model.MemberAssetManager.ToModel(value))
	})

	req.RegisterRoute(horizon.Route{
		Route:  "/member-asset/:member_asset_id",
		Method: "DELETE",
		Note:   "Delete a member's asset record by ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberAssetID, err := horizon.EngineUUIDParam(ctx, "member_asset_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid member asset ID")
		}
		if err := c.model.MemberAssetManager.DeleteByID(context, *memberAssetID); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})

		}
		return ctx.NoContent(http.StatusNoContent)
	})
}
