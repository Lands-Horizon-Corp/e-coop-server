package controllers

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"horizon.com/server/horizon"
	"horizon.com/server/server/model"
)

// GET /member-center
func (c *Controller) MemberCenterList(ctx echo.Context) error {
	member_center, err := c.memberCenter.Manager.List()
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return ctx.JSON(http.StatusOK, c.model.MemberCenterModels(member_center))
}

// GET /member-center/:member_center_id
func (c *Controller) MemberCenterGetByID(ctx echo.Context) error {
	id, err := horizon.EngineUUIDParam(ctx, "member_center_id")
	if err != nil {
		return err
	}
	member_center, err := c.memberCenter.Manager.GetByID(*id)
	if err != nil {
		return ctx.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
	}
	return ctx.JSON(http.StatusOK, c.model.MemberCenterModel(member_center))
}

// POST /member-center
func (c *Controller) MemberCenterCreate(ctx echo.Context) error {
	req, err := c.model.MemberCenterValidate(ctx)
	if err != nil {
		return err
	}
	user, err := c.provider.CurrentUserOrganization(ctx)
	if err != nil {
		return err
	}
	model := &model.MemberCenter{
		CreatedAt:      time.Now().UTC(),
		CreatedByID:    user.UserID,
		UpdatedAt:      time.Now().UTC(),
		UpdatedByID:    user.UserID,
		BranchID:       *user.BranchID,
		OrganizationID: user.OrganizationID,

		Name:        req.Name,
		Description: req.Description,
	}
	if err := c.memberCenter.Manager.Create(model); err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	c.provider.UserFootstep(ctx, "member-center", "creating member center", model)
	return ctx.JSON(http.StatusCreated, c.model.MemberCenterModel(model))
}

// PUT /member-center/member_center_id
func (c *Controller) MemberCenterUpdate(ctx echo.Context) error {
	id, err := horizon.EngineUUIDParam(ctx, "member_center_id")
	if err != nil {
		return err
	}

	req, err := c.model.MemberCenterValidate(ctx)
	if err != nil {
		return err
	}

	user, err := c.provider.CurrentUserOrganization(ctx)
	if err != nil {
		return err
	}
	model := &model.MemberCenter{
		UpdatedAt:      time.Now().UTC(),
		UpdatedByID:    user.UserID,
		BranchID:       *user.BranchID,
		OrganizationID: user.OrganizationID,

		Name:        req.Name,
		Description: req.Description,
	}
	if err := c.memberCenter.Manager.UpdateByID(*id, model); err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	c.provider.UserFootstep(ctx, "member-center", "updating member center", model)
	return ctx.JSON(http.StatusCreated, c.model.MemberCenterModel(model))
}

// DELETE /member-center/member_center_id
func (c *Controller) MemberCenterDelete(ctx echo.Context) error {
	id, err := horizon.EngineUUIDParam(ctx, "member_center_id")
	if err != nil {
		return err
	}
	if err := c.memberCenter.Manager.DeleteByID(*id); err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return ctx.NoContent(http.StatusNoContent)
}

// GET member-center/branch/:branch_id
func (c *Controller) MemberCenterListByBranch(ctx echo.Context) error {
	id, err := horizon.EngineUUIDParam(ctx, "branch_id")
	if err != nil {
		return err
	}
	member_center, err := c.memberCenter.ListByBranch(*id)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return ctx.JSON(http.StatusOK, c.model.MemberCenterModels(member_center))
}

// GET member-center/organization/:organization_id
func (c *Controller) MemberCenterListByOrganization(ctx echo.Context) error {
	id, err := horizon.EngineUUIDParam(ctx, "organization_id")
	if err != nil {
		return err
	}
	member_center, err := c.memberCenter.ListByOrganization(*id)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return ctx.JSON(http.StatusOK, c.model.MemberCenterModels(member_center))
}

// GET member_center/organization/:organization_id/branch/:branch_id
func (c *Controller) MemberCenterListByOrganizationBranch(ctx echo.Context) error {
	orgId, err := horizon.EngineUUIDParam(ctx, "organization_id")
	if err != nil {
		return err
	}
	branchId, err := horizon.EngineUUIDParam(ctx, "branch_id")
	if err != nil {
		return err
	}
	member_center, err := c.memberCenter.ListByOrganizationBranch(*branchId, *orgId)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return ctx.JSON(http.StatusOK, c.model.MemberCenterModels(member_center))
}
