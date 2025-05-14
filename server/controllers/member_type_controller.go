package controllers

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"horizon.com/server/horizon"
	"horizon.com/server/server/model"
)

// GET /member-type
func (c *Controller) MemberTypeList(ctx echo.Context) error {
	member_type, err := c.memberType.Manager.List()
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return ctx.JSON(http.StatusOK, c.model.MemberTypeModels(member_type))
}

// GET /member-type/:member_type_id
func (c *Controller) MemberTypeGetByID(ctx echo.Context) error {
	id, err := horizon.EngineUUIDParam(ctx, "member_type_id")
	if err != nil {
		return err
	}
	member_type, err := c.memberType.Manager.GetByID(*id)
	if err != nil {
		return ctx.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
	}
	return ctx.JSON(http.StatusOK, c.model.MemberTypeModel(member_type))
}

// POST /member-type
func (c *Controller) MemberTypeCreate(ctx echo.Context) error {
	req, err := c.model.MemberTypeValidate(ctx)
	if err != nil {
		return err
	}
	user, err := c.provider.CurrentUserOrganization(ctx)
	if err != nil {
		return err
	}
	model := &model.MemberType{
		CreatedAt:      time.Now().UTC(),
		CreatedByID:    user.UserID,
		UpdatedAt:      time.Now().UTC(),
		UpdatedByID:    user.UserID,
		BranchID:       *user.BranchID,
		OrganizationID: user.OrganizationID,

		Name:        req.Name,
		Description: req.Description,
		Prefix:      req.PRefix,
	}
	if err := c.memberType.Manager.Create(model); err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	c.provider.UserFootstep(ctx, "member-type", "creating member center", model)
	return ctx.JSON(http.StatusCreated, c.model.MemberTypeModel(model))
}

// PUT /member-type/member_type_id
func (c *Controller) MemberTypeUpdate(ctx echo.Context) error {
	id, err := horizon.EngineUUIDParam(ctx, "member_type_id")
	if err != nil {
		return err
	}

	req, err := c.model.MemberTypeValidate(ctx)
	if err != nil {
		return err
	}

	user, err := c.provider.CurrentUserOrganization(ctx)
	if err != nil {
		return err
	}
	model := &model.MemberType{
		UpdatedAt:      time.Now().UTC(),
		UpdatedByID:    user.UserID,
		BranchID:       *user.BranchID,
		OrganizationID: user.OrganizationID,

		Name:        req.Name,
		Description: req.Description,
		Prefix:      req.PRefix,
	}
	if err := c.memberType.Manager.UpdateByID(*id, model); err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	c.provider.UserFootstep(ctx, "member-type", "updating member center", model)
	return ctx.JSON(http.StatusCreated, c.model.MemberTypeModel(model))
}

// DELETE /member-type/member_type_id
func (c *Controller) MemberTypeDelete(ctx echo.Context) error {
	id, err := horizon.EngineUUIDParam(ctx, "member_type_id")
	if err != nil {
		return err
	}
	if err := c.memberType.Manager.DeleteByID(*id); err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return ctx.NoContent(http.StatusNoContent)
}

// GET member-type/branch/:branch_id
func (c *Controller) MemberTypeListByBranch(ctx echo.Context) error {
	id, err := horizon.EngineUUIDParam(ctx, "branch_id")
	if err != nil {
		return err
	}
	member_type, err := c.memberType.ListByBranch(*id)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return ctx.JSON(http.StatusOK, c.model.MemberTypeModels(member_type))
}

// GET member-type/organization/:organization_id
func (c *Controller) MemberTypeListByOrganization(ctx echo.Context) error {
	id, err := horizon.EngineUUIDParam(ctx, "organization_id")
	if err != nil {
		return err
	}
	member_type, err := c.memberType.ListByOrganization(*id)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return ctx.JSON(http.StatusOK, c.model.MemberTypeModels(member_type))
}

// GET member_type/organization/:organization_id/branch/:branch_id
func (c *Controller) MemberTypeListByOrganizationBranch(ctx echo.Context) error {
	orgId, err := horizon.EngineUUIDParam(ctx, "organization_id")
	if err != nil {
		return err
	}
	branchId, err := horizon.EngineUUIDParam(ctx, "branch_id")
	if err != nil {
		return err
	}
	member_type, err := c.memberType.ListByOrganizationBranch(*branchId, *orgId)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return ctx.JSON(http.StatusOK, c.model.MemberTypeModels(member_type))
}
