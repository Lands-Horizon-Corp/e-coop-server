package controller

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/lands-horizon/horizon-server/services/horizon"
	"github.com/lands-horizon/horizon-server/src/model"
)

func (c *Controller) MemberAddressController() {
	req := c.provider.Service.Request

	req.RegisterRoute(horizon.Route{
		Route:    "/member-address/member-profile/:member_profile_id",
		Method:   "POST",
		Request:  "TMemberAddress",
		Response: "TMemberAddress",
		Note:     "Create a new address record for a member.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberProfileID, err := horizon.EngineUUIDParam(ctx, "member_profile_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid member profile ID")
		}
		req, err := c.model.MemberAddressManager.Validate(ctx)
		if err != nil {
			return c.BadRequest(ctx, err.Error())
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.NoContent(http.StatusNoContent)
		}

		value := &model.MemberAddress{
			MemberProfileID: memberProfileID,

			Label:         req.Label,
			City:          req.City,
			CountryCode:   req.CountryCode,
			PostalCode:    req.PostalCode,
			ProvinceState: req.ProvinceState,
			Barangay:      req.Barangay,
			Landmark:      req.Landmark,
			Address:       req.Address,

			CreatedAt:      time.Now().UTC(),
			CreatedByID:    user.UserID,
			UpdatedAt:      time.Now().UTC(),
			UpdatedByID:    user.UserID,
			BranchID:       *user.BranchID,
			OrganizationID: user.OrganizationID,
		}

		if err := c.model.MemberAddressManager.Create(context, value); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})

		}

		return ctx.JSON(http.StatusOK, c.model.MemberAddressManager.ToModel(value))
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/member-address/:member_address_id",
		Method:   "PUT",
		Request:  "TMemberAddress",
		Response: "TMemberAddress",
		Note:     "Update an existing address record for a member in the current branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberAddressID, err := horizon.EngineUUIDParam(ctx, "member_address_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid member address ID")
		}
		req, err := c.model.MemberAddressManager.Validate(ctx)
		if err != nil {
			return c.BadRequest(ctx, err.Error())
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.NoContent(http.StatusNoContent)
		}

		value, err := c.model.MemberAddressManager.GetByID(context, *memberAddressID)
		if err != nil {
			return c.NotFound(ctx, "MemberAddress")
		}

		value.UpdatedAt = time.Now().UTC()
		value.UpdatedByID = user.UserID
		value.OrganizationID = user.OrganizationID
		value.BranchID = *user.BranchID

		value.MemberProfileID = req.MemberProfileID
		value.Label = req.Label
		value.City = req.City
		value.CountryCode = req.CountryCode
		value.PostalCode = req.PostalCode
		value.ProvinceState = req.ProvinceState
		value.Barangay = req.Barangay
		value.Landmark = req.Landmark
		value.Address = req.Address
		if err := c.model.MemberAddressManager.UpdateFields(context, value.ID, value); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})

		}
		return ctx.JSON(http.StatusOK, c.model.MemberAddressManager.ToModel(value))
	})

	req.RegisterRoute(horizon.Route{
		Route:  "/member-address/:member_address_id",
		Method: "DELETE",
		Note:   "Delete a member's address record by ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberAddressID, err := horizon.EngineUUIDParam(ctx, "member_address_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid member address ID")
		}
		if err := c.model.MemberAddressManager.DeleteByID(context, *memberAddressID); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})

		}
		return ctx.NoContent(http.StatusNoContent)
	})
}
