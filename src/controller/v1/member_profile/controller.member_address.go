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

func MemberAddressController(service *horizon.HorizonService) {
	req := service.API

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/member-address/member-profile/:member_profile_id",
		Method:       "POST",
		RequestType: types.MemberAddress{},
		ResponseType: types.MemberAddress{},
		Note:         "Creates a new address record for a member profile.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberProfileID, err := helpers.EngineUUIDParam(ctx, "member_profile_id")
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create member address failed (/member-address/member-profile/:member_profile_id), invalid member profile ID.",
				Module:      "MemberAddress",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member profile ID"})
		}
		req, err := core.MemberAddressManager(service).Validate(ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create member address failed (/member-address/member-profile/:member_profile_id), validation error: " + err.Error(),
				Module:      "MemberAddress",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid address data: " + err.Error()})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create member address failed (/member-address/member-profile/:member_profile_id), user org error: " + err.Error(),
				Module:      "MemberAddress",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization/branch not found"})
		}
		if userOrg.BranchID == nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create member address failed (/member-address/member-profile/:member_profile_id), user not assigned to branch.",
				Module:      "MemberAddress",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		value := &types.MemberAddress{
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
			CreatedByID:     &userOrg.UserID,
			UpdatedAt:       time.Now().UTC(),
			UpdatedByID:     &userOrg.UserID,
			BranchID:        *userOrg.BranchID,
			OrganizationID:  userOrg.OrganizationID,
			Longitude:       req.Longitude,
			Latitude:        req.Latitude,
		}
		if err := core.MemberAddressManager(service).Create(context, value); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create member address failed (/member-address/member-profile/:member_profile_id), db error: " + err.Error(),
				Module:      "MemberAddress",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create member address record: " + err.Error()})
		}
		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "create-success",
			Description: "Created member address (/member-address/member-profile/:member_profile_id): " + value.Label,
			Module:      "MemberAddress",
		})
		return ctx.JSON(http.StatusCreated, core.MemberAddressManager(service).ToModel(value))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/member-address/:member_address_id",
		Method:       "PUT",
		RequestType: types.MemberAddress{},
		ResponseType: types.MemberAddress{},
		Note:         "Updates an existing address record for a member in the current branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberAddressID, err := helpers.EngineUUIDParam(ctx, "member_address_id")
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update member address failed (/member-address/:member_address_id), invalid member address ID.",
				Module:      "MemberAddress",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member address ID"})
		}
		req, err := core.MemberAddressManager(service).Validate(ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update member address failed (/member-address/:member_address_id), validation error: " + err.Error(),
				Module:      "MemberAddress",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid address data: " + err.Error()})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update member address failed (/member-address/:member_address_id), user org error: " + err.Error(),
				Module:      "MemberAddress",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization/branch not found"})
		}
		if userOrg.BranchID == nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update member address failed (/member-address/:member_address_id), user not assigned to branch.",
				Module:      "MemberAddress",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		value, err := core.MemberAddressManager(service).GetByID(context, *memberAddressID)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update member address failed (/member-address/:member_address_id), record not found.",
				Module:      "MemberAddress",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Member address record not found"})
		}
		value.UpdatedAt = time.Now().UTC()
		value.UpdatedByID = &userOrg.UserID
		value.OrganizationID = userOrg.OrganizationID
		value.BranchID = *userOrg.BranchID

		value.MemberProfileID = req.MemberProfileID
		value.Label = req.Label
		value.City = req.City
		value.CountryCode = req.CountryCode
		value.PostalCode = req.PostalCode
		value.ProvinceState = req.ProvinceState
		value.Barangay = req.Barangay
		value.Landmark = req.Landmark
		value.Address = req.Address
		value.Longitude = req.Longitude
		value.Latitude = req.Latitude
		if err := core.MemberAddressManager(service).UpdateByID(context, value.ID, value); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update member address failed (/member-address/:member_address_id), db error: " + err.Error(),
				Module:      "MemberAddress",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update member address record: " + err.Error()})
		}
		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Updated member address (/member-address/:member_address_id): " + value.Label,
			Module:      "MemberAddress",
		})
		return ctx.JSON(http.StatusOK, core.MemberAddressManager(service).ToModel(value))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:  "/api/v1/member-address/:member_address_id",
		Method: "DELETE",
		Note:   "Deletes a member's address record by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		memberAddressID, err := helpers.EngineUUIDParam(ctx, "member_address_id")
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete member address failed (/member-address/:member_address_id), invalid member address ID.",
				Module:      "MemberAddress",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member address ID"})
		}
		value, err := core.MemberAddressManager(service).GetByID(context, *memberAddressID)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete member address failed (/member-address/:member_address_id), record not found.",
				Module:      "MemberAddress",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Member address record not found"})
		}
		if err := core.MemberAddressManager(service).Delete(context, *memberAddressID); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete member address failed (/member-address/:member_address_id), db error: " + err.Error(),
				Module:      "MemberAddress",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete member address record: " + err.Error()})
		}
		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "delete-success",
			Description: "Deleted member address (/member-address/:member_address_id): " + value.Label,
			Module:      "MemberAddress",
		})
		return ctx.NoContent(http.StatusNoContent)
	})
}
