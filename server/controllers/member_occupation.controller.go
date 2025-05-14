package controllers

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"horizon.com/server/horizon"
	"horizon.com/server/server/model"
)

// GET /member-occupation
func (c *Controller) MemberOccupationList(ctx echo.Context) error {
	member_occupation, err := c.memberOccupation.Manager.List()
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return ctx.JSON(http.StatusOK, c.model.MemberOccupationModels(member_occupation))
}

// GET /member-occupation/:member_occupation_id
func (c *Controller) MemberOccupationGetByID(ctx echo.Context) error {
	id, err := horizon.EngineUUIDParam(ctx, "member_occupation_id")
	if err != nil {
		return err
	}
	member_occupation, err := c.memberOccupation.Manager.GetByID(*id)
	if err != nil {
		return ctx.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
	}
	return ctx.JSON(http.StatusOK, c.model.MemberOccupationModel(member_occupation))
}

// POST /member-occupation
func (c *Controller) MemberOccupationCreate(ctx echo.Context) error {
	req, err := c.model.MemberOccupationValidate(ctx)
	if err != nil {
		return err
	}
	user, err := c.provider.CurrentUserOrganization(ctx)
	if err != nil {
		return err
	}
	model := &model.MemberOccupation{
		CreatedAt:      time.Now().UTC(),
		CreatedByID:    user.UserID,
		UpdatedAt:      time.Now().UTC(),
		UpdatedByID:    user.UserID,
		BranchID:       *user.BranchID,
		OrganizationID: user.OrganizationID,

		Name:        req.Name,
		Description: req.Description,
	}
	if err := c.memberOccupation.Manager.Create(model); err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	c.provider.UserFootstep(ctx, "member-occupation", "creating member center", model)
	return ctx.JSON(http.StatusCreated, c.model.MemberOccupationModel(model))
}

// PUT /member-occupation/member_occupation_id
func (c *Controller) MemberOccupationUpdate(ctx echo.Context) error {
	id, err := horizon.EngineUUIDParam(ctx, "member_occupation_id")
	if err != nil {
		return err
	}

	req, err := c.model.MemberOccupationValidate(ctx)
	if err != nil {
		return err
	}

	user, err := c.provider.CurrentUserOrganization(ctx)
	if err != nil {
		return err
	}
	model := &model.MemberOccupation{
		UpdatedAt:      time.Now().UTC(),
		UpdatedByID:    user.UserID,
		BranchID:       *user.BranchID,
		OrganizationID: user.OrganizationID,

		Name:        req.Name,
		Description: req.Description,
	}
	if err := c.memberOccupation.Manager.UpdateByID(*id, model); err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	c.provider.UserFootstep(ctx, "member-occupation", "updating member center", model)
	return ctx.JSON(http.StatusCreated, c.model.MemberOccupationModel(model))
}

// DELETE /member-occupation/member_occupation_id
func (c *Controller) MemberOccupationDelete(ctx echo.Context) error {
	id, err := horizon.EngineUUIDParam(ctx, "member_occupation_id")
	if err != nil {
		return err
	}
	if err := c.memberOccupation.Manager.DeleteByID(*id); err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return ctx.NoContent(http.StatusNoContent)
}

// GET member-occupation/branch/:branch_id
func (c *Controller) MemberOccupationListByBranch(ctx echo.Context) error {
	id, err := horizon.EngineUUIDParam(ctx, "branch_id")
	if err != nil {
		return err
	}
	member_occupation, err := c.memberOccupation.ListByBranch(*id)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return ctx.JSON(http.StatusOK, c.model.MemberOccupationModels(member_occupation))
}

// GET member-occupation/organization/:organization_id
func (c *Controller) MemberOccupationListByOrganization(ctx echo.Context) error {
	id, err := horizon.EngineUUIDParam(ctx, "organization_id")
	if err != nil {
		return err
	}
	member_occupation, err := c.memberOccupation.ListByOrganization(*id)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return ctx.JSON(http.StatusOK, c.model.MemberOccupationModels(member_occupation))
}

// GET member_occupation/organization/:organization_id/branch/:branch_id
func (c *Controller) MemberOccupationListByOrganizationBranch(ctx echo.Context) error {
	orgId, err := horizon.EngineUUIDParam(ctx, "organization_id")
	if err != nil {
		return err
	}
	branchId, err := horizon.EngineUUIDParam(ctx, "branch_id")
	if err != nil {
		return err
	}
	member_occupation, err := c.memberOccupation.ListByOrganizationBranch(*branchId, *orgId)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return ctx.JSON(http.StatusOK, c.model.MemberOccupationModels(member_occupation))
}
