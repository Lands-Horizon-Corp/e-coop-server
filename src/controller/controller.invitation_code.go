package controller

import (
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/lands-horizon/horizon-server/services/horizon"
	"github.com/lands-horizon/horizon-server/src/model"
)

func (c *Controller) InvitationCode() {
	req := c.provider.Service.Request

	// Retrieve all invitation codes for the current user's organization
	req.RegisterRoute(horizon.Route{
		Route:    "/invitation-code",
		Method:   "GET",
		Response: "IInvitationCode[]",
		Note:     "Retrieves a list of all invitation codes for the current organization (based on JWT user organization).",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return err
		}
		invitationCode, err := c.model.GetInvitationCodeByBranch(context, userOrg.OrganizationID, *userOrg.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.InvitationCodeManager.ToModels(invitationCode))
	})
	req.RegisterRoute(horizon.Route{
		Route:    "/invitation-code/search",
		Method:   "GET",
		Request:  "Filter<TInvitationCode>",
		Response: "Paginated<TInvitationCode>",
		Note:     "Get pagination gender",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.NoContent(http.StatusNoContent)
		}
		invitationCode, err := c.model.GetInvitationCodeByBranch(context, user.OrganizationID, *user.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.InvitationCodeManager.Pagination(context, ctx, invitationCode))
	})
	// Retrieve all invitation codes that match a specific code in the current organization
	req.RegisterRoute(horizon.Route{
		Route:    "/invitation-code/code/:code",
		Method:   "GET",
		Response: "IInvitationCode",
		Note:     "Retrieves invitation code matching the specified code for the current organization (based on JWT user organization).",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		code := ctx.Param("code")
		invitationCode, err := c.model.GetInvitationCodeByCode(context, code)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return ctx.JSON(http.StatusAccepted, c.model.InvitationCodeManager.ToModel(invitationCode))
	})

	// Retrieve a specific invitation code by its ID
	req.RegisterRoute(horizon.Route{
		Route:    "/invitation-code/:invitation_code_id",
		Method:   "GET",
		Response: "IInvitationCode",
		Note:     "Retrieves details of a specific invitation code by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		invitationCodeId, err := horizon.EngineUUIDParam(ctx, "invitation_code_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid invitation code ID")
		}
		invitationCode, err := c.model.InvitationCodeManager.GetByID(context, *invitationCodeId)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return ctx.JSON(http.StatusAccepted, c.model.InvitationCodeManager.ToModel(invitationCode))
	})

	// Create a new invitation code for the current user's organization
	req.RegisterRoute(horizon.Route{
		Route:    "/invitation-code",
		Method:   "POST",
		Response: "IInvitationCode",
		Request:  "IInvitationCode",
		Note:     "Creates a new invitation code under the current organization (based on JWT user organization).",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := c.model.InvitationCodeManager.Validate(ctx)
		if err != nil {
			return c.BadRequest(ctx, err.Error())
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return err
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "only owners and employees can perform this action"})
		}
		if req.UserType == "owner" {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "cannot create invitation code type owner"})
		}
		data := &model.InvitationCode{
			CreatedAt:      time.Now().UTC(),
			CreatedByID:    userOrg.UserID,
			UpdatedAt:      time.Now().UTC(),
			UpdatedByID:    userOrg.UserID,
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
			UserType:       req.UserType,
			Code:           req.Code,
			ExpirationDate: req.ExpirationDate,
			MaxUse:         req.MaxUse,
			CurrentUse:     0,
			Description:    req.Description,
		}

		if err := c.model.InvitationCodeManager.Create(context, data); err != nil {
			return c.InternalServerError(ctx, err)
		}
		return ctx.JSON(http.StatusOK, c.model.InvitationCodeManager.ToModel(data))
	})

	// Update an existing invitation code by its ID
	req.RegisterRoute(horizon.Route{
		Route:    "/invitation-code/:invitation_code_id",
		Method:   "PUT",
		Response: "IInvitationCode",
		Request:  "IInvitationCode",
		Note:     "Updates an existing invitation code identified by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		invitationCodeId, err := horizon.EngineUUIDParam(ctx, "invitation_code_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid invitation code ID")
		}
		req, err := c.model.InvitationCodeManager.Validate(ctx)
		if err != nil {
			return c.BadRequest(ctx, err.Error())
		}
		invitationCode, err := c.model.InvitationCodeManager.GetByID(context, *invitationCodeId)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return err
		}
		invitationCode.UpdatedAt = time.Now().UTC()
		invitationCode.UpdatedByID = userOrg.UserID
		invitationCode.OrganizationID = userOrg.OrganizationID
		invitationCode.BranchID = *userOrg.BranchID
		invitationCode.UserType = req.UserType
		invitationCode.Code = req.Code
		invitationCode.ExpirationDate = req.ExpirationDate
		invitationCode.MaxUse = req.MaxUse
		invitationCode.Description = req.Description
		invitationCode.PermissionDescription = req.PermissionDescription
		invitationCode.Permissions = req.Permissions
		invitationCode.PermissionName = req.PermissionName

		if err := c.model.InvitationCodeManager.UpdateFields(context, invitationCode.ID, invitationCode); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to update user: "+err.Error())
		}
		return ctx.JSON(http.StatusOK, c.model.InvitationCodeManager.ToModel(invitationCode))
	})

	// Delete a specific invitation code by its ID
	req.RegisterRoute(horizon.Route{
		Route:  "/invitation-code/:invitation_code_id",
		Method: "DELETE",
		Note:   "Deletes a specific invitation code identified by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		invitationCodeId, err := horizon.EngineUUIDParam(ctx, "invitation_code_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid invitation code ID")
		}
		if err := c.model.InvitationCodeManager.DeleteByID(context, *invitationCodeId); err != nil {
			return c.InternalServerError(ctx, err)
		}
		return ctx.NoContent(http.StatusNoContent)
	})

	req.RegisterRoute(horizon.Route{
		Route:   "/invitation-code/bulk-delete",
		Method:  "DELETE",
		Request: "string[]",
		Note:    "Delete multiple member occupation records by their IDs",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var reqBody struct {
			IDs []string `json:"ids"`
		}

		if err := ctx.Bind(&reqBody); err != nil {
			return c.BadRequest(ctx, "Invalid request body")
		}

		if len(reqBody.IDs) == 0 {
			return c.BadRequest(ctx, "No IDs provided")
		}

		tx := c.provider.Service.Database.Client().Begin()
		if tx.Error != nil {
			tx.Rollback()
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": tx.Error.Error()})
		}

		for _, rawID := range reqBody.IDs {
			invitationCodeId, err := uuid.Parse(rawID)
			if err != nil {
				tx.Rollback()
				return c.BadRequest(ctx, fmt.Sprintf("Invalid UUID: %s", rawID))
			}

			if _, err := c.model.InvitationCodeManager.GetByID(context, invitationCodeId); err != nil {
				tx.Rollback()
				return c.NotFound(ctx, fmt.Sprintf("InvitationCode with ID %s", rawID))
			}

			if err := c.model.InvitationCodeManager.DeleteByIDWithTx(context, tx, invitationCodeId); err != nil {
				tx.Rollback()
				return c.InternalServerError(ctx, err)
			}
		}

		if err := tx.Commit().Error; err != nil {
			return c.InternalServerError(ctx, err)
		}

		return ctx.NoContent(http.StatusNoContent)
	})
}
