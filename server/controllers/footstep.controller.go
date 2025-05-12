package controllers

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"horizon.com/server/horizon"
)

// GET /footstep
func (c *Controller) FootstepList(ctx echo.Context) error {
	footstep, err := c.footstep.Manager.List()
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return ctx.JSON(http.StatusOK, c.model.FootstepModels(footstep))
}

// GET /footstep/:footstep_id
func (c *Controller) FootstepGetByID(ctx echo.Context) error {
	id, err := horizon.EngineUUIDParam(ctx, "footstep_id")
	if err != nil {
		return err
	}
	footstep, err := c.footstep.Manager.GetByID(*id)
	if err != nil {
		return ctx.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
	}
	return ctx.JSON(http.StatusOK, c.model.FootstepModel(footstep))
}

// DELETE /footstep/footstep_id
func (c *Controller) FootstepDelete(ctx echo.Context) error {
	id, err := horizon.EngineUUIDParam(ctx, "footstep_id")
	if err != nil {
		return err
	}
	if err := c.footstep.Manager.DeleteByID(*id); err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return ctx.NoContent(http.StatusNoContent)
}

// GET footstep/user/:user_id
func (c *Controller) FootstepListByUser(ctx echo.Context) error {
	id, err := horizon.EngineUUIDParam(ctx, "user_id")
	if err != nil {
		return err
	}
	footstep, err := c.footstep.ListByUser(id)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return ctx.JSON(http.StatusOK, c.model.FootstepModels(footstep))
}

// GET footstep/branch/:branch_id
func (c *Controller) FootstepListByBranch(ctx echo.Context) error {
	id, err := horizon.EngineUUIDParam(ctx, "branch_id")
	if err != nil {
		return err
	}
	footstep, err := c.footstep.ListByBranch(*id)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return ctx.JSON(http.StatusOK, c.model.FootstepModels(footstep))
}

// GET footstep/organization/:organization_id
func (c *Controller) FootstepListByOrganization(ctx echo.Context) error {
	id, err := horizon.EngineUUIDParam(ctx, "organization_id")
	if err != nil {
		return err
	}
	footstep, err := c.footstep.ListByOrganization(*id)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return ctx.JSON(http.StatusOK, c.model.FootstepModels(footstep))
}

// GET footstep/organization/:organization_id/branch/:branch_id
func (c *Controller) FootstepListByOrganizationBranch(ctx echo.Context) error {
	orgId, err := horizon.EngineUUIDParam(ctx, "organization_id")
	if err != nil {
		return err
	}
	branchId, err := horizon.EngineUUIDParam(ctx, "branch_id")
	if err != nil {
		return err
	}
	footstep, err := c.footstep.ListByOrganizationBranch(*branchId, *orgId)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return ctx.JSON(http.StatusOK, c.model.FootstepModels(footstep))
}

// GET footstep/user/:user_id/organization/:organization_id/branch/:branch_id
func (c *Controller) FootstepListByUserOrganizationBranch(ctx echo.Context) error {
	userId, err := horizon.EngineUUIDParam(ctx, "user_id")
	if err != nil {
		return err
	}
	orgId, err := horizon.EngineUUIDParam(ctx, "organization_id")
	if err != nil {
		return err
	}
	branchId, err := horizon.EngineUUIDParam(ctx, "branch_id")
	if err != nil {
		return err
	}
	footstep, err := c.footstep.ListByUserOrganizationBranch(userId, *branchId, *orgId)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return ctx.JSON(http.StatusOK, c.model.FootstepModels(footstep))

}

// GET footstep/user/:user_id/branch/:branch_id
func (c *Controller) FootstepUserBranch(ctx echo.Context) error {
	userId, err := horizon.EngineUUIDParam(ctx, "user_id")
	if err != nil {
		return err
	}
	branchId, err := horizon.EngineUUIDParam(ctx, "branch_id")
	if err != nil {
		return err
	}
	footstep, err := c.footstep.ListByUserBranch(userId, *branchId)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return ctx.JSON(http.StatusOK, c.model.FootstepModels(footstep))
}

// GET footstep/user/:user_id/organization/:organization_id
func (c *Controller) FootstepListByUserOrganization(ctx echo.Context) error {
	userId, err := horizon.EngineUUIDParam(ctx, "user_id")
	if err != nil {
		return err
	}
	orgId, err := horizon.EngineUUIDParam(ctx, "organization_id")
	if err != nil {
		return err
	}
	footstep, err := c.footstep.ListByUserOrganization(userId, *orgId)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return ctx.JSON(http.StatusOK, c.model.FootstepModels(footstep))
}
