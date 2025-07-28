package controller

import (
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/lands-horizon/horizon-server/services/handlers"
	"github.com/lands-horizon/horizon-server/src/event"
	"github.com/lands-horizon/horizon-server/src/model"
)

// InvitationCode manages endpoints for invitation code resources.
func (c *Controller) InvitationCode() {
	req := c.provider.Service.Request

	// GET /invitation-code: Retrieve all invitation codes for the current user's organization and branch. (NO footstep)
	req.RegisterRoute(handlers.Route{
		Route:        "/invitation-code",
		Method:       "GET",
		ResponseType: model.InvitationCodeResponse{},
		Note:         "Returns all invitation codes for the current user's organization and branch.",
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
		return ctx.JSON(http.StatusOK, c.model.InvitationCodeManager.Filtered(context, ctx, invitationCode))
	})

	// GET /invitation-code/search: Paginated search of invitation codes for current branch. (NO footstep)
	req.RegisterRoute(handlers.Route{
		Route:       "/invitation-code/search",
		Method:      "GET",
		RequestType: model.InvitationCodeRequest{},
		Note:        "Returns a paginated list of invitation codes for the current user's organization and branch.",
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

	// GET /invitation-code/code/:code: Retrieve an invitation code by its code string (for current organization). (NO footstep)
	req.RegisterRoute(handlers.Route{
		Route:        "/invitation-code/code/:code",
		Method:       "GET",
		Note:         "Returns the invitation code matching the specified code for the current user's organization.",
		ResponseType: model.InvitationCodeResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		code := ctx.Param("code")
		invitationCode, err := c.model.GetInvitationCodeByCode(context, code)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Invitation code not found"})
		}
		return ctx.JSON(http.StatusOK, c.model.InvitationCodeManager.ToModel(invitationCode))
	})

	// GET /invitation-code/:invitation_code_id: Retrieve a specific invitation code by its ID. (NO footstep)
	req.RegisterRoute(handlers.Route{
		Route:        "/invitation-code/:invitation_code_id",
		Method:       "GET",
		Note:         "Returns the details of a specific invitation code by its ID.",
		ResponseType: model.InvitationCodeResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		invitationCodeId, err := handlers.EngineUUIDParam(ctx, "invitation_code_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid invitation code ID"})
		}
		invitationCode, err := c.model.InvitationCodeManager.GetByID(context, *invitationCodeId)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Invitation code not found"})
		}
		return ctx.JSON(http.StatusOK, c.model.InvitationCodeManager.ToModel(invitationCode))
	})

	// POST /invitation-code: Create a new invitation code for the current user's organization and branch. (WITH footstep)
	req.RegisterRoute(handlers.Route{
		Route:        "/invitation-code",
		Method:       "POST",
		ResponseType: model.InvitationCodeResponse{},
		RequestType:  model.InvitationCodeRequest{},
		Note:         "Creates a new invitation code under the current user's organization and branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := c.model.InvitationCodeManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Invitation code creation failed (/invitation-code), validation error: " + err.Error(),
				Module:      "InvitationCode",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid invitation code data: " + err.Error()})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Invitation code creation failed (/invitation-code), user org error: " + err.Error(),
				Module:      "InvitationCode",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization/branch not found"})
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Unauthorized create attempt for invitation code (/invitation-code)",
				Module:      "InvitationCode",
			})
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Only owners and employees can create invitation codes"})
		}
		if req.UserType == "owner" {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Invitation code creation failed (/invitation-code), attempted to create user type 'owner'",
				Module:      "InvitationCode",
			})
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Cannot create invitation code with user type 'owner'"})
		}
		if userOrg.BranchID == nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Invitation code creation failed (/invitation-code), user not assigned to branch.",
				Module:      "InvitationCode",
			})
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
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Invitation code creation failed (/invitation-code), db error: " + err.Error(),
				Module:      "InvitationCode",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create invitation code: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "create-success",
			Description: "Created invitation code (/invitation-code): " + data.Code,
			Module:      "InvitationCode",
		})
		return ctx.JSON(http.StatusCreated, c.model.InvitationCodeManager.ToModel(data))
	})

	// PUT /invitation-code/:invitation_code_id: Update an existing invitation code by its ID. (WITH footstep)
	req.RegisterRoute(handlers.Route{
		Route:        "/invitation-code/:invitation_code_id",
		Method:       "PUT",
		ResponseType: model.InvitationCodeResponse{},
		RequestType:  model.InvitationCodeRequest{},
		Note:         "Updates an existing invitation code identified by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		invitationCodeId, err := handlers.EngineUUIDParam(ctx, "invitation_code_id")
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Invitation code update failed (/invitation-code/:invitation_code_id), invalid invitation code ID.",
				Module:      "InvitationCode",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid invitation code ID"})
		}
		req, err := c.model.InvitationCodeManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Invitation code update failed (/invitation-code/:invitation_code_id), validation error: " + err.Error(),
				Module:      "InvitationCode",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid invitation code data: " + err.Error()})
		}
		invitationCode, err := c.model.InvitationCodeManager.GetByID(context, *invitationCodeId)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Invitation code update failed (/invitation-code/:invitation_code_id), not found.",
				Module:      "InvitationCode",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Invitation code not found"})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Invitation code update failed (/invitation-code/:invitation_code_id), user org error: " + err.Error(),
				Module:      "InvitationCode",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization/branch not found"})
		}
		if userOrg.BranchID == nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
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
		invitationCode.UserType = req.UserType
		invitationCode.Code = req.Code
		invitationCode.ExpirationDate = req.ExpirationDate
		invitationCode.MaxUse = req.MaxUse
		invitationCode.Description = req.Description
		invitationCode.PermissionDescription = req.PermissionDescription
		invitationCode.Permissions = req.Permissions
		invitationCode.PermissionName = req.PermissionName

		if err := c.model.InvitationCodeManager.UpdateFields(context, invitationCode.ID, invitationCode); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Invitation code update failed (/invitation-code/:invitation_code_id), db error: " + err.Error(),
				Module:      "InvitationCode",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update invitation code: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Updated invitation code (/invitation-code/:invitation_code_id): " + invitationCode.Code,
			Module:      "InvitationCode",
		})
		return ctx.JSON(http.StatusOK, c.model.InvitationCodeManager.ToModel(invitationCode))
	})

	// DELETE /invitation-code/:invitation_code_id: Delete a specific invitation code by its ID. (WITH footstep)
	req.RegisterRoute(handlers.Route{
		Route:  "/invitation-code/:invitation_code_id",
		Method: "DELETE",
		Note:   "Deletes a specific invitation code identified by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		invitationCodeId, err := handlers.EngineUUIDParam(ctx, "invitation_code_id")
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Invitation code delete failed (/invitation-code/:invitation_code_id), invalid invitation code ID.",
				Module:      "InvitationCode",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid invitation code ID"})
		}
		codeModel, err := c.model.InvitationCodeManager.GetByID(context, *invitationCodeId)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Invitation code delete failed (/invitation-code/:invitation_code_id), not found.",
				Module:      "InvitationCode",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Invitation code not found"})
		}
		if err := c.model.InvitationCodeManager.DeleteByID(context, *invitationCodeId); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Invitation code delete failed (/invitation-code/:invitation_code_id), db error: " + err.Error(),
				Module:      "InvitationCode",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete invitation code: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "delete-success",
			Description: "Deleted invitation code (/invitation-code/:invitation_code_id): " + codeModel.Code,
			Module:      "InvitationCode",
		})
		return ctx.NoContent(http.StatusNoContent)
	})

	// DELETE /invitation-code/bulk-delete: Bulk delete invitation codes by IDs. (WITH footstep)
	req.RegisterRoute(handlers.Route{
		Route:       "/invitation-code/bulk-delete",
		Method:      "DELETE",
		RequestType: model.IDSRequest{},
		Note:        "Deletes multiple invitation codes by their IDs. Expects a JSON body: { \"ids\": [\"id1\", \"id2\", ...] }",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var reqBody model.IDSRequest
		if err := ctx.Bind(&reqBody); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Invitation code bulk delete failed (/invitation-code/bulk-delete), invalid request body.",
				Module:      "InvitationCode",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
		}
		if len(reqBody.IDs) == 0 {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Invitation code bulk delete failed (/invitation-code/bulk-delete), no IDs provided.",
				Module:      "InvitationCode",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "No IDs provided for bulk delete"})
		}
		tx := c.provider.Service.Database.Client().Begin()
		if tx.Error != nil {
			tx.Rollback()
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Invitation code bulk delete failed (/invitation-code/bulk-delete), begin tx error: " + tx.Error.Error(),
				Module:      "InvitationCode",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to start database transaction: " + tx.Error.Error()})
		}
		codes := ""
		for _, rawID := range reqBody.IDs {
			invitationCodeId, err := uuid.Parse(rawID)
			if err != nil {
				tx.Rollback()
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "bulk-delete-error",
					Description: "Invitation code bulk delete failed (/invitation-code/bulk-delete), invalid UUID: " + rawID,
					Module:      "InvitationCode",
				})
				return ctx.JSON(http.StatusBadRequest, map[string]string{"error": fmt.Sprintf("Invalid UUID: %s", rawID)})
			}
			codeModel, err := c.model.InvitationCodeManager.GetByID(context, invitationCodeId)
			if err != nil {
				tx.Rollback()
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "bulk-delete-error",
					Description: "Invitation code bulk delete failed (/invitation-code/bulk-delete), not found: " + rawID,
					Module:      "InvitationCode",
				})
				return ctx.JSON(http.StatusNotFound, map[string]string{"error": fmt.Sprintf("Invitation code not found with ID: %s", rawID)})
			}
			codes += codeModel.Code + ","
			if err := c.model.InvitationCodeManager.DeleteByIDWithTx(context, tx, invitationCodeId); err != nil {
				tx.Rollback()
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "bulk-delete-error",
					Description: "Invitation code bulk delete failed (/invitation-code/bulk-delete), db error: " + err.Error(),
					Module:      "InvitationCode",
				})
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete invitation code: " + err.Error()})
			}
		}
		if err := tx.Commit().Error; err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Invitation code bulk delete failed (/invitation-code/bulk-delete), commit error: " + err.Error(),
				Module:      "InvitationCode",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to commit transaction: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "bulk-delete-success",
			Description: "Bulk deleted invitation codes (/invitation-code/bulk-delete): " + codes,
			Module:      "InvitationCode",
		})
		return ctx.NoContent(http.StatusNoContent)
	})
}
