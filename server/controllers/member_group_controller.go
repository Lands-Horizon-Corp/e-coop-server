package controllers

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"horizon.com/server/horizon"
	"horizon.com/server/server/model"
)

// GET /member-group
func (c *Controller) MemberGroupList(ctx echo.Context) error {
	member_group, err := c.memberGroup.Manager.List()
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return ctx.JSON(http.StatusOK, c.model.MemberGroupModels(member_group))
}

// GET /member-group/:member_group_id
func (c *Controller) MemberGroupGetByID(ctx echo.Context) error {
	id, err := horizon.EngineUUIDParam(ctx, "member_group_id")
	if err != nil {
		return err
	}
	member_group, err := c.memberGroup.Manager.GetByID(*id)
	if err != nil {
		return ctx.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
	}
	return ctx.JSON(http.StatusOK, c.model.MemberGroupModel(member_group))
}

// POST /member-group
func (c *Controller) MemberGroupCreate(ctx echo.Context) error {
	req, err := c.model.MemberGroupValidate(ctx)
	if err != nil {
		return err
	}
	user, err := c.provider.CurrentUserOrganization(ctx)
	if err != nil {
		return err
	}
	model := &model.MemberGroup{
		CreatedAt:      time.Now().UTC(),
		CreatedByID:    user.UserID,
		UpdatedAt:      time.Now().UTC(),
		UpdatedByID:    user.UserID,
		BranchID:       *user.BranchID,
		OrganizationID: user.OrganizationID,

		Name:        req.Name,
		Description: req.Description,
	}
	if err := c.memberGroup.Manager.Create(model); err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	c.provider.UserFootstep(ctx, "member-group", "creating member center", model)
	return ctx.JSON(http.StatusCreated, c.model.MemberGroupModel(model))
}

// PUT /member-group/member_group_id
func (c *Controller) MemberGroupUpdate(ctx echo.Context) error {
	id, err := horizon.EngineUUIDParam(ctx, "member_group_id")
	if err != nil {
		return err
	}

	req, err := c.model.MemberGroupValidate(ctx)
	if err != nil {
		return err
	}

	user, err := c.provider.CurrentUserOrganization(ctx)
	if err != nil {
		return err
	}
	model := &model.MemberGroup{
		UpdatedAt:      time.Now().UTC(),
		UpdatedByID:    user.UserID,
		BranchID:       *user.BranchID,
		OrganizationID: user.OrganizationID,

		Name:        req.Name,
		Description: req.Description,
	}
	if err := c.memberGroup.Manager.UpdateByID(*id, model); err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	c.provider.UserFootstep(ctx, "member-group", "updating member center", model)
	return ctx.JSON(http.StatusCreated, c.model.MemberGroupModel(model))
}

// DELETE /member-group/member_group_id
func (c *Controller) MemberGroupDelete(ctx echo.Context) error {
	id, err := horizon.EngineUUIDParam(ctx, "member_group_id")
	if err != nil {
		return err
	}
	if err := c.memberGroup.Manager.DeleteByID(*id); err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return ctx.NoContent(http.StatusNoContent)
}

// GET member-group/branch/:branch_id
func (c *Controller) MemberGroupListByBranch(ctx echo.Context) error {
	id, err := horizon.EngineUUIDParam(ctx, "branch_id")
	if err != nil {
		return err
	}
	member_group, err := c.memberGroup.ListByBranch(*id)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return ctx.JSON(http.StatusOK, c.model.MemberGroupModels(member_group))
}

// GET member-group/organization/:organization_id
func (c *Controller) MemberGroupListByOrganization(ctx echo.Context) error {
	id, err := horizon.EngineUUIDParam(ctx, "organization_id")
	if err != nil {
		return err
	}
	member_group, err := c.memberGroup.ListByOrganization(*id)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return ctx.JSON(http.StatusOK, c.model.MemberGroupModels(member_group))
}

// GET member_group/organization/:organization_id/branch/:branch_id
func (c *Controller) MemberGroupListByOrganizationBranch(ctx echo.Context) error {
	orgId, err := horizon.EngineUUIDParam(ctx, "organization_id")
	if err != nil {
		return err
	}
	branchId, err := horizon.EngineUUIDParam(ctx, "branch_id")
	if err != nil {
		return err
	}
	member_group, err := c.memberGroup.ListByOrganizationBranch(*branchId, *orgId)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return ctx.JSON(http.StatusOK, c.model.MemberGroupModels(member_group))
}
