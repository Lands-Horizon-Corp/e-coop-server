package controller

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/lands-horizon/horizon-server/services/horizon"
)

func (c *Controller) FootstepController() {
	req := c.provider.Service.Request

	req.RegisterRoute(horizon.Route{
		Route:    "/footstep/me",
		Method:   "GET",
		Response: "TFootstep[]",
		Note:     "Getting your own footstep (the logged in user)",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userToken.CurrentUser(context, ctx)
		if err != nil {
			return err
		}
		footstep, err := c.model.GetFootstepByUser(context, user.ID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.FootstepManager.ToModels(footstep))
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/footstep/branch",
		Method:   "GET",
		Response: "TFootstep[]",
		Note:     "Get footstep on current branch",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return err
		}
		footstep, err := c.model.GetFootstepByBranch(context, userOrg.OrganizationID, *userOrg.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.FootstepManager.ToModels(footstep))
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/footstep/user-organization/:user_organization_id",
		Method:   "GET",
		Response: "TFootstep[]",
		Note:     "Getting Footstep of users that is (member or employee or owner) on current branch",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return err
		}
		footstep, err := c.model.GetFootstepByUserOrganization(context, userOrg.UserID, userOrg.OrganizationID, *userOrg.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.FootstepManager.ToModels(footstep))
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/footstep/:footstep_id",
		Method:   "GET",
		Response: "TFootstep",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		footstepId, err := horizon.EngineUUIDParam(ctx, "footstep_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid footstep ID")
		}
		footstep, err := c.model.FootstepManager.GetByIDRaw(context, *footstepId)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})

		}
		return ctx.JSON(http.StatusOK, footstep)
	})

}
