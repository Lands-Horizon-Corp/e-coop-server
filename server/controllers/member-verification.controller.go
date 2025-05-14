package controllers

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"horizon.com/server/horizon"
)

// GET /member-verification
func (c *Controller) MemberVerificationList(ctx echo.Context) error {
	member_verification, err := c.memberVerification.Manager.List()
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return ctx.JSON(http.StatusOK, c.model.MemberVerificationModels(member_verification))
}

// GET /member-verification/:member_verification_id
func (c *Controller) MemberVerificationGetByID(ctx echo.Context) error {
	id, err := horizon.EngineUUIDParam(ctx, "member_verification_id")
	if err != nil {
		return err
	}
	member_verification, err := c.memberVerification.Manager.GetByID(*id)
	if err != nil {
		return ctx.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
	}
	return ctx.JSON(http.StatusOK, c.model.MemberVerificationModel(member_verification))
}

// DELETE /member-verification/member_verification_id
func (c *Controller) MemberVerificationDelete(ctx echo.Context) error {
	id, err := horizon.EngineUUIDParam(ctx, "member_verification_id")
	if err != nil {
		return err
	}
	if err := c.memberVerification.Manager.DeleteByID(*id); err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return ctx.NoContent(http.StatusNoContent)
}

// GET member-verification/branch/:branch_id
func (c *Controller) MemberVerificationListByBranch(ctx echo.Context) error {
	id, err := horizon.EngineUUIDParam(ctx, "branch_id")
	if err != nil {
		return err
	}
	member_verification, err := c.memberVerification.ListByBranch(*id)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return ctx.JSON(http.StatusOK, c.model.MemberVerificationModels(member_verification))
}

// GET member-verification/organization/:organization_id
func (c *Controller) MemberVerificationListByOrganization(ctx echo.Context) error {
	id, err := horizon.EngineUUIDParam(ctx, "organization_id")
	if err != nil {
		return err
	}
	member_verification, err := c.memberVerification.ListByOrganization(*id)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return ctx.JSON(http.StatusOK, c.model.MemberVerificationModels(member_verification))
}

// GET member_verification/organization/:organization_id/branch/:branch_id
func (c *Controller) MemberVerificationListByOrganizationBranch(ctx echo.Context) error {
	orgId, err := horizon.EngineUUIDParam(ctx, "organization_id")
	if err != nil {
		return err
	}
	branchId, err := horizon.EngineUUIDParam(ctx, "branch_id")
	if err != nil {
		return err
	}
	member_verification, err := c.memberVerification.ListByOrganizationBranch(*branchId, *orgId)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return ctx.JSON(http.StatusOK, c.model.MemberVerificationModels(member_verification))
}
