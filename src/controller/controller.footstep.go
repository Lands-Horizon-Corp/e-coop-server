package controller

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/lands-horizon/horizon-server/services/horizon"
)

// FootstepController manages endpoints related to footstep records.
func (c *Controller) FootstepController() {
	req := c.provider.Service.Request

	// GET /footstep/me: Get all footsteps for the currently logged-in user.
	req.RegisterRoute(horizon.Route{
		Route:    "/footstep/me",
		Method:   "GET",
		Response: "TFootstep[]",
		Note:     "Returns all footsteps for the currently authenticated user.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userToken.CurrentUser(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or user not found"})
		}
		footstep, err := c.model.GetFootstepByUser(context, user.ID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve user's footsteps: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.FootstepManager.ToModels(footstep))
	})

	// GET /footstep/branch: Get all footsteps for the current user's branch.
	req.RegisterRoute(horizon.Route{
		Route:    "/footstep/branch",
		Method:   "GET",
		Response: "TFootstep[]",
		Note:     "Returns all footsteps for the current user's organization and branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization/branch not found"})
		}
		footstep, err := c.model.GetFootstepByBranch(context, userOrg.OrganizationID, *userOrg.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve footsteps for branch: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.FootstepManager.ToModels(footstep))
	})

	// GET /footstep/user-organization/:user_organization_id: Get footsteps for a user organization on the current branch.
	req.RegisterRoute(horizon.Route{
		Route:    "/footstep/user-organization/:user_organization_id",
		Method:   "GET",
		Response: "TFootstep[]",
		Note:     "Returns footsteps for the specified user-organization on the current branch if the user is a member, employee, or owner.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization/branch not found"})
		}
		footstep, err := c.model.GetFootstepByUserOrganization(context, userOrg.UserID, userOrg.OrganizationID, *userOrg.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve footsteps for user organization: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.FootstepManager.ToModels(footstep))
	})

	// GET /footstep/:footstep_id: Get a specific footstep by ID.
	req.RegisterRoute(horizon.Route{
		Route:    "/footstep/:footstep_id",
		Method:   "GET",
		Response: "TFootstep",
		Note:     "Returns a specific footstep record by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		footstepId, err := horizon.EngineUUIDParam(ctx, "footstep_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid footstep ID"})
		}
		footstep, err := c.model.FootstepManager.GetByIDRaw(context, *footstepId)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Footstep record not found"})
		}
		return ctx.JSON(http.StatusOK, footstep)
	})
}
