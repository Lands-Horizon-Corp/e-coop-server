package user

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

func InvitationCodeController(service *horizon.HorizonService) {
	req := service.API

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/invitation-code",
		Method:       "GET",
		ResponseType: types.InvitationCodeResponse{},
		Note:         "Returns all invitation codes for the current user's organization and branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization/branch not found"})
		}
		if userOrg.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		invitationCode, err := core.GetInvitationCodeByBranch(context, service, userOrg.OrganizationID, *userOrg.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve invitation codes: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, core.InvitationCodeManager(service).ToModels(invitationCode))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:       "/api/v1/invitation-code/search",
		Method:      "GET",
		RequestType: types.InvitationCodeRequest{},
		Note:        "Returns a paginated list of invitation codes for the current user's organization and branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization/branch not found"})
		}
		if userOrg.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		invitationCode, err := core.InvitationCodeManager(service).NormalPagination(context, ctx, &types.InvitationCode{
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve invitation codes: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, invitationCode)
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/invitation-code/code/:code",
		Method:       "GET",
		Note:         "Returns the invitation code matching the specified code for the current user's organization.",
		ResponseType: types.InvitationCodeResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		code := ctx.Param("code")
		invitationCode, err := core.GetInvitationCodeByCode(context, service, code)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Invitation code not found"})
		}
		return ctx.JSON(http.StatusOK, core.InvitationCodeManager(service).ToModel(invitationCode))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/invitation-code/:invitation_code_id",
		Method:       "GET",
		Note:         "Returns the details of a specific invitation code by its ID.",
		ResponseType: types.InvitationCodeResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		invitationCodeID, err := helpers.EngineUUIDParam(ctx, "invitation_code_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid invitation code ID"})
		}
		invitationCode, err := core.InvitationCodeManager(service).GetByID(context, *invitationCodeID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Invitation code not found"})
		}
		return ctx.JSON(http.StatusOK, core.InvitationCodeManager(service).ToModel(invitationCode))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/invitation-code",
		Method:       "POST",
		ResponseType: types.InvitationCodeResponse{},
		RequestType: types.InvitationCodeRequest{},
		Note:         "Creates a new invitation code under the current user's organization and branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := core.InvitationCodeManager(service).Validate(ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Invitation code creation failed (/invitation-code), validation error: " + err.Error(),
				Module:      "InvitationCode",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid invitation code data: " + err.Error()})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Invitation code creation failed (/invitation-code), user org error: " + err.Error(),
				Module:      "InvitationCode",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization/branch not found"})
		}
		if userOrg.UserType != core.UserOrganizationTypeOwner && userOrg.UserType != core.UserOrganizationTypeEmployee {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Unauthorized create attempt for invitation code (/invitation-code)",
				Module:      "InvitationCode",
			})
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Only owners and employees can create invitation codes"})
		}
		if core.UserOrganizationType(req.UserType) == core.UserOrganizationTypeOwner {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Invitation code creation failed (/invitation-code), attempted to create user type 'owner'",
				Module:      "InvitationCode",
			})
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Cannot create invitation code with user type 'owner'"})
		}
		if userOrg.BranchID == nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Invitation code creation failed (/invitation-code), user not assigned to branch.",
				Module:      "InvitationCode",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		data := &types.InvitationCode{
			CreatedAt:      time.Now().UTC(),
			CreatedByID:    userOrg.UserID,
			UpdatedAt:      time.Now().UTC(),
			UpdatedByID:    userOrg.UserID,
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
			UserType:       core.UserOrganizationType(req.UserType),
			Code:           req.Code,
			ExpirationDate: req.ExpirationDate,
			MaxUse:         req.MaxUse,
			CurrentUse:     0,
			Description:    req.Description,
		}
		if err := core.InvitationCodeManager(service).Create(context, data); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Invitation code creation failed (/invitation-code), db error: " + err.Error(),
				Module:      "InvitationCode",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create invitation code: " + err.Error()})
		}
		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "create-success",
			Description: "Created invitation code (/invitation-code): " + data.Code,
			Module:      "InvitationCode",
		})
		return ctx.JSON(http.StatusCreated, core.InvitationCodeManager(service).ToModel(data))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/invitation-code/:invitation_code_id",
		Method:       "PUT",
		ResponseType: types.InvitationCodeResponse{},
		RequestType: types.InvitationCodeRequest{},
		Note:         "Updates an existing invitation code identified by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		invitationCodeID, err := helpers.EngineUUIDParam(ctx, "invitation_code_id")
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Invitation code update failed (/invitation-code/:invitation_code_id), invalid invitation code ID.",
				Module:      "InvitationCode",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid invitation code ID"})
		}
		req, err := core.InvitationCodeManager(service).Validate(ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Invitation code update failed (/invitation-code/:invitation_code_id), validation error: " + err.Error(),
				Module:      "InvitationCode",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid invitation code data: " + err.Error()})
		}
		invitationCode, err := core.InvitationCodeManager(service).GetByID(context, *invitationCodeID)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Invitation code update failed (/invitation-code/:invitation_code_id), not found.",
				Module:      "InvitationCode",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Invitation code not found"})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Invitation code update failed (/invitation-code/:invitation_code_id), user org error: " + err.Error(),
				Module:      "InvitationCode",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization/branch not found"})
		}
		if userOrg.BranchID == nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Invitation code update failed (/invitation-code/:invitation_code_id), user not assigned to branch.",
				Module:      "InvitationCode",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		invitationCode.UpdatedAt = time.Now().UTC()
		invitationCode.UpdatedByID = userOrg.UserID
		invitationCode.OrganizationID = userOrg.OrganizationID
		invitationCode.BranchID = *userOrg.BranchID
		invitationCode.UserType = core.UserOrganizationType(req.UserType)
		invitationCode.Code = req.Code
		invitationCode.ExpirationDate = req.ExpirationDate
		invitationCode.MaxUse = req.MaxUse
		invitationCode.Description = req.Description
		invitationCode.PermissionDescription = req.PermissionDescription
		invitationCode.Permissions = req.Permissions
		invitationCode.PermissionName = req.PermissionName

		if err := core.InvitationCodeManager(service).UpdateByID(context, invitationCode.ID, invitationCode); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Invitation code update failed (/invitation-code/:invitation_code_id), db error: " + err.Error(),
				Module:      "InvitationCode",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update invitation code: " + err.Error()})
		}
		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Updated invitation code (/invitation-code/:invitation_code_id): " + invitationCode.Code,
			Module:      "InvitationCode",
		})
		return ctx.JSON(http.StatusOK, core.InvitationCodeManager(service).ToModel(invitationCode))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:  "/api/v1/invitation-code/:invitation_code_id",
		Method: "DELETE",
		Note:   "Deletes a specific invitation code identified by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		invitationCodeID, err := helpers.EngineUUIDParam(ctx, "invitation_code_id")
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Invitation code delete failed (/invitation-code/:invitation_code_id), invalid invitation code ID.",
				Module:      "InvitationCode",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid invitation code ID"})
		}
		codeModel, err := core.InvitationCodeManager(service).GetByID(context, *invitationCodeID)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Invitation code delete failed (/invitation-code/:invitation_code_id), not found.",
				Module:      "InvitationCode",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Invitation code not found"})
		}
		if err := core.InvitationCodeManager(service).Delete(context, *invitationCodeID); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Invitation code delete failed (/invitation-code/:invitation_code_id), db error: " + err.Error(),
				Module:      "InvitationCode",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete invitation code: " + err.Error()})
		}
		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "delete-success",
			Description: "Deleted invitation code (/invitation-code/:invitation_code_id): " + codeModel.Code,
			Module:      "InvitationCode",
		})
		return ctx.NoContent(http.StatusNoContent)
	})

	req.RegisterWebRoute(horizon.Route{
		Route:       "/api/v1/invitation-code/bulk-delete",
		Method:      "DELETE",
		Note:        "Deletes multiple invitation codes by their IDs. Expects a JSON body: { \"ids\": [\"id1\", \"id2\", ...] }",
		RequestType: types.IDSRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var reqBody core.IDSRequest

		if err := ctx.Bind(&reqBody); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Invitation code bulk delete failed (/invitation-code/bulk-delete) | invalid request body: " + err.Error(),
				Module:      "InvitationCode",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body: " + err.Error()})
		}

		if len(reqBody.IDs) == 0 {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Invitation code bulk delete failed (/invitation-code/bulk-delete) | no IDs provided",
				Module:      "InvitationCode",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "No IDs provided for bulk delete"})
		}

		ids := make([]any, len(reqBody.IDs))
		for i, id := range reqBody.IDs {
			ids[i] = id
		}

		if err := core.InvitationCodeManager(service).BulkDelete(context, ids); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Invitation code bulk delete failed (/invitation-code/bulk-delete) | error: " + err.Error(),
				Module:      "InvitationCode",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to bulk delete invitation codes: " + err.Error()})
		}

		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "bulk-delete-success",
			Description: "Bulk deleted invitation codes (/invitation-code/bulk-delete)",
			Module:      "InvitationCode",
		})

		return ctx.NoContent(http.StatusNoContent)
	})
}
