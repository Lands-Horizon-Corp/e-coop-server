package controllers

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"horizon.com/server/horizon"
	"horizon.com/server/server/model"
)

// GET /invitation-code
func (c *Controller) InvitationCode(ctx echo.Context) error {
	invitationCode, err := c.invitationCode.Manager.List()
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return ctx.JSON(http.StatusOK, c.model.InvitationCodeModels(invitationCode))
}

// GET /invitation-code/:invitation_code_id
func (c *Controller) InvitationCodeGetByID(ctx echo.Context) error {
	id, err := horizon.EngineUUIDParam(ctx, "invitation_code_id")
	if err != nil {
		return err
	}
	contact_us, err := c.invitationCode.Manager.GetByID(*id)
	if err != nil {
		return ctx.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
	}
	return ctx.JSON(http.StatusOK, c.model.InvitationCodeModel(contact_us))
}

// POST /invitation-code/organization/:organization_id/branch/branch_id
func (c *Controller) InvitationCodeCreate(ctx echo.Context) error {
	organizationId, err := horizon.EngineUUIDParam(ctx, "organization_id")
	if err != nil {
		return err
	}
	branchId, err := horizon.EngineUUIDParam(ctx, "branch_id")
	if err != nil {
		return err
	}
	req, err := c.model.InvitationCodeValidate(ctx)
	if err != nil {
		return err
	}
	userOrg, err := c.provider.UserOwnerEmployee(ctx, organizationId.String(), branchId.String())
	if err != nil {
		return err
	}
	model := &model.InvitationCode{
		CreatedByID:    userOrg.UserID,
		UpdatedByID:    userOrg.User.ID,
		OrganizationID: userOrg.OrganizationID,
		BranchID:       *userOrg.BranchID,
		CreatedAt:      time.Now().UTC(),
		UpdatedAt:      time.Now().UTC(),

		UserType:       req.UserType,
		Code:           req.Code,
		ExpirationDate: req.ExpirationDate,
		MaxUse:         req.MaxUse,
		Description:    req.Description,
	}

	if err := c.invitationCode.Manager.Create(model); err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return ctx.JSON(http.StatusCreated, c.model.InvitationCodeModel(model))
}

// PUT /invitation-code/:invitation_code_id
func (c *Controller) InvitationCodeUpdate(ctx echo.Context) error {
	// Extract the invitation_code_id from the route
	invitationCodeId, err := horizon.EngineUUIDParam(ctx, "invitation_code_id")
	if err != nil {
		return err
	}

	// Validate the request body
	req, err := c.model.InvitationCodeValidate(ctx)
	if err != nil {
		return err
	}

	// ✅ Fetch the existing invitation code
	existing, err := c.invitationCode.Manager.GetByID(*invitationCodeId)
	if err != nil {
		return ctx.JSON(http.StatusNotFound, map[string]string{"error": "invitation code not found"})
	}

	// ✅ Get user ownership based on current invitation code's org and branch
	userOrg, err := c.provider.UserOwnerEmployee(ctx, existing.OrganizationID.String(), existing.BranchID.String())
	if err != nil {
		return err
	}

	// ✅ Update fields (preserving CreatedAt/CreatedByID)
	existing.UpdatedByID = userOrg.User.ID
	existing.UpdatedAt = time.Now().UTC()
	existing.UserType = req.UserType
	existing.Code = req.Code
	existing.ExpirationDate = req.ExpirationDate
	existing.MaxUse = req.MaxUse
	existing.Description = req.Description

	// Save updated invitation code
	if err := c.invitationCode.Manager.UpdateByID(*invitationCodeId, existing); err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return ctx.JSON(http.StatusOK, c.model.InvitationCodeModel(existing))
}

// DELETE /invitation-code/:invitation_code_id/organization/:organization_id/branch/branch_id
func (c *Controller) InvitationCodeDelete(ctx echo.Context) error {
	invitationCodeId, err := horizon.EngineUUIDParam(ctx, "invitation_code_id")
	if err != nil {
		return err
	}
	organizationId, err := horizon.EngineUUIDParam(ctx, "organization_id")
	if err != nil {
		return err
	}
	branchId, err := horizon.EngineUUIDParam(ctx, "branch_id")
	if err != nil {
		return err
	}
	_, err = c.provider.UserOwnerEmployee(ctx, organizationId.String(), branchId.String())
	if err != nil {
		return err
	}
	if err := c.invitationCode.Manager.DeleteByID(*invitationCodeId); err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return ctx.NoContent(http.StatusNoContent)
}

// GET invitation-code/branch/:branch_id
func (c *Controller) InvitationCodeListByBranch(ctx echo.Context) error {
	branchId, err := horizon.EngineUUIDParam(ctx, "branch_id")
	if err != nil {
		return err
	}
	invitationCode, err := c.invitationCode.ListByBranch(*branchId)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return ctx.JSON(http.StatusOK, c.model.InvitationCodeModels(invitationCode))
}

// GET invitation-code/organization/:organization_id
func (c *Controller) InvitationCodeListByOrganization(ctx echo.Context) error {
	orgId, err := horizon.EngineUUIDParam(ctx, "organization_id")
	if err != nil {
		return err
	}
	invitationCode, err := c.invitationCode.ListByOrganization(*orgId)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return ctx.JSON(http.StatusOK, c.model.InvitationCodeModels(invitationCode))
}

// GET invitation-code/organization/:organization_id/branch/:branch_id
func (c *Controller) InvitationCodeListByOrganizationBranch(ctx echo.Context) error {
	orgId, err := horizon.EngineUUIDParam(ctx, "organization_id")
	if err != nil {
		return err
	}
	branchId, err := horizon.EngineUUIDParam(ctx, "branch_id")
	if err != nil {
		return err
	}
	invitationCode, err := c.invitationCode.ListByOrganizationBranch(*orgId, *branchId)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return ctx.JSON(http.StatusOK, c.model.InvitationCodeModels(invitationCode))
}

// GET nvitation-code/exists/:code
func (c *Controller) InvitationCodeExists(ctx echo.Context) error {
	code := ctx.Param("code")

	exists, err := c.invitationCode.Exists(code)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{
			"error": "failed to check invitation code",
		})
	}

	if !exists {
		return ctx.JSON(http.StatusNotFound, map[string]string{
			"error": "invitation code not found",
		})
	}
	return ctx.NoContent(http.StatusOK)
}

// GET invitation-code/code/:code
func (c *Controller) InvitationCodeByCode(ctx echo.Context) error {
	code := ctx.Param("code")
	exists, err := c.invitationCode.Exists(code)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{
			"error": "failed to check invitation code",
		})
	}
	if !exists {
		return ctx.JSON(http.StatusNotFound, map[string]string{
			"error": "invitation code not found",
		})
	}
	invitationCode, err := c.invitationCode.ByCode(code)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to retrieve invitation code"})
	}
	return ctx.JSON(http.StatusOK, c.model.InvitationCodeModel(invitationCode))
}

// GET invitation-code/verfiy/:code
func (c *Controller) InvitationCodeVerify(ctx echo.Context) error {
	code := ctx.Param("code")
	exists, err := c.invitationCode.Exists(code)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{
			"error": "failed to check invitation code",
		})
	}
	if !exists {
		return ctx.JSON(http.StatusNotFound, map[string]string{
			"error": "invitation code not found",
		})
	}
	invitationCode, err := c.invitationCode.Verify(code)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to retrieve invitation code"})
	}
	return ctx.JSON(http.StatusOK, c.model.InvitationCodeModel(invitationCode))
}
