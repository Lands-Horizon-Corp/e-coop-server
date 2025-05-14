package controllers

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"horizon.com/server/horizon"
	"horizon.com/server/server/model"
)

// GET /member-gender
func (c *Controller) MemberGenderList(ctx echo.Context) error {
	member_gender, err := c.memberGender.Manager.List()
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return ctx.JSON(http.StatusOK, c.model.MemberGenderModels(member_gender))
}

// GET /member-gender/:member_gender_id
func (c *Controller) MemberGenderGetByID(ctx echo.Context) error {
	id, err := horizon.EngineUUIDParam(ctx, "member_gender_id")
	if err != nil {
		return err
	}
	member_gender, err := c.memberGender.Manager.GetByID(*id)
	if err != nil {
		return ctx.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
	}
	return ctx.JSON(http.StatusOK, c.model.MemberGenderModel(member_gender))
}

// POST /member-gender
func (c *Controller) MemberGenderCreate(ctx echo.Context) error {
	req, err := c.model.MemberGenderValidate(ctx)
	if err != nil {
		return err
	}
	user, err := c.provider.CurrentUserOrganization(ctx)
	if err != nil {
		return err
	}
	model := &model.MemberGender{
		CreatedAt:      time.Now().UTC(),
		CreatedByID:    user.UserID,
		UpdatedAt:      time.Now().UTC(),
		UpdatedByID:    user.UserID,
		BranchID:       *user.BranchID,
		OrganizationID: user.OrganizationID,

		Name:        req.Name,
		Description: req.Description,
	}
	if err := c.memberGender.Manager.Create(model); err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	c.provider.UserFootstep(ctx, "member-gender", "creating member center", model)
	return ctx.JSON(http.StatusCreated, c.model.MemberGenderModel(model))
}

// PUT /member-gender/member_gender_id
func (c *Controller) MemberGenderUpdate(ctx echo.Context) error {
	id, err := horizon.EngineUUIDParam(ctx, "member_gender_id")
	if err != nil {
		return err
	}

	req, err := c.model.MemberGenderValidate(ctx)
	if err != nil {
		return err
	}

	user, err := c.provider.CurrentUserOrganization(ctx)
	if err != nil {
		return err
	}
	model := &model.MemberGender{
		UpdatedAt:      time.Now().UTC(),
		UpdatedByID:    user.UserID,
		BranchID:       *user.BranchID,
		OrganizationID: user.OrganizationID,

		Name:        req.Name,
		Description: req.Description,
	}
	if err := c.memberGender.Manager.UpdateByID(*id, model); err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	c.provider.UserFootstep(ctx, "member-gender", "updating member center", model)
	return ctx.JSON(http.StatusCreated, c.model.MemberGenderModel(model))
}

// DELETE /member-gender/member_gender_id
func (c *Controller) MemberGenderDelete(ctx echo.Context) error {
	id, err := horizon.EngineUUIDParam(ctx, "member_gender_id")
	if err != nil {
		return err
	}
	if err := c.memberGender.Manager.DeleteByID(*id); err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return ctx.NoContent(http.StatusNoContent)
}

// GET member-gender/branch/:branch_id
func (c *Controller) MemberGenderListByBranch(ctx echo.Context) error {
	id, err := horizon.EngineUUIDParam(ctx, "branch_id")
	if err != nil {
		return err
	}
	member_gender, err := c.memberGender.ListByBranch(*id)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return ctx.JSON(http.StatusOK, c.model.MemberGenderModels(member_gender))
}

// GET member-gender/organization/:organization_id
func (c *Controller) MemberGenderListByOrganization(ctx echo.Context) error {
	id, err := horizon.EngineUUIDParam(ctx, "organization_id")
	if err != nil {
		return err
	}
	member_gender, err := c.memberGender.ListByOrganization(*id)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return ctx.JSON(http.StatusOK, c.model.MemberGenderModels(member_gender))
}

// GET member_gender/organization/:organization_id/branch/:branch_id
func (c *Controller) MemberGenderListByOrganizationBranch(ctx echo.Context) error {
	orgId, err := horizon.EngineUUIDParam(ctx, "organization_id")
	if err != nil {
		return err
	}
	branchId, err := horizon.EngineUUIDParam(ctx, "branch_id")
	if err != nil {
		return err
	}
	member_gender, err := c.memberGender.ListByOrganizationBranch(*branchId, *orgId)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return ctx.JSON(http.StatusOK, c.model.MemberGenderModels(member_gender))
}
