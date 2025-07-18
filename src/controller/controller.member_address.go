package controller

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/lands-horizon/horizon-server/services/horizon"
	"github.com/lands-horizon/horizon-server/src/model"
)

// MemberAddressController manages endpoints for member address records.
func (c *Controller) MemberAddressController() {
	req := c.provider.Service.Request

	// POST /member-address/member-profile/:member_profile_id: Create a new address record for a member.
	req.RegisterRoute(horizon.Route{
		Route:    "/member-address/member-profile/:member_profile_id",
		Method:   "POST",
		Request:  "TMemberAddress",
		Response: "TMemberAddress",
		Note:     "Creates a new address record for a member profile.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberProfileID, err := horizon.EngineUUIDParam(ctx, "member_profile_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member profile ID"})
		}
		req, err := c.model.MemberAddressManager.Validate(ctx)
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid address data: " + err.Error()})
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization/branch not found"})
		}
		if user.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		value := &model.MemberAddress{
			MemberProfileID: memberProfileID,
			Label:           req.Label,
			City:            req.City,
			CountryCode:     req.CountryCode,
			PostalCode:      req.PostalCode,
			ProvinceState:   req.ProvinceState,
			Barangay:        req.Barangay,
			Landmark:        req.Landmark,
			Address:         req.Address,
			CreatedAt:       time.Now().UTC(),
			CreatedByID:     user.UserID,
			UpdatedAt:       time.Now().UTC(),
			UpdatedByID:     user.UserID,
			BranchID:        *user.BranchID,
			OrganizationID:  user.OrganizationID,
		}
		if err := c.model.MemberAddressManager.Create(context, value); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create member address record: " + err.Error()})
		}
		return ctx.JSON(http.StatusCreated, c.model.MemberAddressManager.ToModel(value))
	})

	// PUT /member-address/:member_address_id: Update an existing address record for a member.
	req.RegisterRoute(horizon.Route{
		Route:    "/member-address/:member_address_id",
		Method:   "PUT",
		Request:  "TMemberAddress",
		Response: "TMemberAddress",
		Note:     "Updates an existing address record for a member in the current branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberAddressID, err := horizon.EngineUUIDParam(ctx, "member_address_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member address ID"})
		}
		req, err := c.model.MemberAddressManager.Validate(ctx)
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid address data: " + err.Error()})
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization/branch not found"})
		}
		if user.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		value, err := c.model.MemberAddressManager.GetByID(context, *memberAddressID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Member address record not found"})
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
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update member address record: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.MemberAddressManager.ToModel(value))
	})

	// DELETE /member-address/:member_address_id: Delete a member's address record by ID.
	req.RegisterRoute(horizon.Route{
		Route:  "/member-address/:member_address_id",
		Method: "DELETE",
		Note:   "Deletes a member's address record by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberAddressID, err := horizon.EngineUUIDParam(ctx, "member_address_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member address ID"})
		}
		if err := c.model.MemberAddressManager.DeleteByID(context, *memberAddressID); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete member address record: " + err.Error()})
		}
		return ctx.NoContent(http.StatusNoContent)
	})
}
