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

// InvitationCode manages endpoints for invitation code resources.
func (c *Controller) InvitationCode() {
	req := c.provider.Service.Request

	// GET /invitation-code: Retrieve all invitation codes for the current user's organization and branch.
	req.RegisterRoute(horizon.Route{
		Route:    "/invitation-code",
		Method:   "GET",
		Response: "IInvitationCode[]",
		Note:     "Returns all invitation codes for the current user's organization and branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization/branch not found"})
		}
		if userOrg.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		invitationCode, err := c.model.GetInvitationCodeByBranch(context, userOrg.OrganizationID, *userOrg.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve invitation codes: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.InvitationCodeManager.ToModels(invitationCode))
	})

	// GET /invitation-code/search: Paginated search of invitation codes for current branch.
	req.RegisterRoute(horizon.Route{
		Route:    "/invitation-code/search",
		Method:   "GET",
		Request:  "Filter<TInvitationCode>",
		Response: "Paginated<TInvitationCode>",
		Note:     "Returns a paginated list of invitation codes for the current user's organization and branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization/branch not found"})
		}
		if userOrg.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		invitationCode, err := c.model.GetInvitationCodeByBranch(context, userOrg.OrganizationID, *userOrg.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve invitation codes: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.InvitationCodeManager.Pagination(context, ctx, invitationCode))
	})

	// GET /invitation-code/code/:code: Retrieve an invitation code by its code string (for current organization).
	req.RegisterRoute(horizon.Route{
		Route:    "/invitation-code/code/:code",
		Method:   "GET",
		Response: "IInvitationCode",
		Note:     "Returns the invitation code matching the specified code for the current user's organization.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		code := ctx.Param("code")
		invitationCode, err := c.model.GetInvitationCodeByCode(context, code)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Invitation code not found"})
		}
		return ctx.JSON(http.StatusOK, c.model.InvitationCodeManager.ToModel(invitationCode))
	})

	// GET /invitation-code/:invitation_code_id: Retrieve a specific invitation code by its ID.
	req.RegisterRoute(horizon.Route{
		Route:    "/invitation-code/:invitation_code_id",
		Method:   "GET",
		Response: "IInvitationCode",
		Note:     "Returns the details of a specific invitation code by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		invitationCodeId, err := horizon.EngineUUIDParam(ctx, "invitation_code_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid invitation code ID"})
		}
		invitationCode, err := c.model.InvitationCodeManager.GetByID(context, *invitationCodeId)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Invitation code not found"})
		}
		return ctx.JSON(http.StatusOK, c.model.InvitationCodeManager.ToModel(invitationCode))
	})

	// POST /invitation-code: Create a new invitation code for the current user's organization and branch.
	req.RegisterRoute(horizon.Route{
		Route:    "/invitation-code",
		Method:   "POST",
		Response: "IInvitationCode",
		Request:  "IInvitationCode",
		Note:     "Creates a new invitation code under the current user's organization and branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := c.model.InvitationCodeManager.Validate(ctx)
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid invitation code data: " + err.Error()})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization/branch not found"})
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Only owners and employees can create invitation codes"})
		}
		if req.UserType == "owner" {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Cannot create invitation code with user type 'owner'"})
		}
		if userOrg.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
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
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create invitation code: " + err.Error()})
		}
		return ctx.JSON(http.StatusCreated, c.model.InvitationCodeManager.ToModel(data))
	})

	// PUT /invitation-code/:invitation_code_id: Update an existing invitation code by its ID.
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
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid invitation code ID"})
		}
		req, err := c.model.InvitationCodeManager.Validate(ctx)
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid invitation code data: " + err.Error()})
		}
		invitationCode, err := c.model.InvitationCodeManager.GetByID(context, *invitationCodeId)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Invitation code not found"})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization/branch not found"})
		}
		if userOrg.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
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
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update invitation code: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.InvitationCodeManager.ToModel(invitationCode))
	})

	// DELETE /invitation-code/:invitation_code_id: Delete a specific invitation code by its ID.
	req.RegisterRoute(horizon.Route{
		Route:  "/invitation-code/:invitation_code_id",
		Method: "DELETE",
		Note:   "Deletes a specific invitation code identified by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		invitationCodeId, err := horizon.EngineUUIDParam(ctx, "invitation_code_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid invitation code ID"})
		}
		if err := c.model.InvitationCodeManager.DeleteByID(context, *invitationCodeId); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete invitation code: " + err.Error()})
		}
		return ctx.NoContent(http.StatusNoContent)
	})

	// DELETE /invitation-code/bulk-delete: Bulk delete invitation codes by IDs.
	req.RegisterRoute(horizon.Route{
		Route:   "/invitation-code/bulk-delete",
		Method:  "DELETE",
		Request: "string[]",
		Note:    "Deletes multiple invitation codes by their IDs. Expects a JSON body: { \"ids\": [\"id1\", \"id2\", ...] }",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var reqBody struct {
			IDs []string `json:"ids"`
		}
		if err := ctx.Bind(&reqBody); err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
		}
		if len(reqBody.IDs) == 0 {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "No IDs provided for bulk delete"})
		}
		tx := c.provider.Service.Database.Client().Begin()
		if tx.Error != nil {
			tx.Rollback()
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to start database transaction: " + tx.Error.Error()})
		}
		for _, rawID := range reqBody.IDs {
			invitationCodeId, err := uuid.Parse(rawID)
			if err != nil {
				tx.Rollback()
				return ctx.JSON(http.StatusBadRequest, map[string]string{"error": fmt.Sprintf("Invalid UUID: %s", rawID)})
			}
			if _, err := c.model.InvitationCodeManager.GetByID(context, invitationCodeId); err != nil {
				tx.Rollback()
				return ctx.JSON(http.StatusNotFound, map[string]string{"error": fmt.Sprintf("Invitation code not found with ID: %s", rawID)})
			}
			if err := c.model.InvitationCodeManager.DeleteByIDWithTx(context, tx, invitationCodeId); err != nil {
				tx.Rollback()
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete invitation code: " + err.Error()})
			}
		}
		if err := tx.Commit().Error; err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to commit transaction: " + err.Error()})
		}
		return ctx.NoContent(http.StatusNoContent)
	})
}
