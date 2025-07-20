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
		Route:    "/footstep/me/search",
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
		return ctx.JSON(http.StatusOK, c.model.FootstepManager.Pagination(context, ctx, footstep))
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/footstep/member-profile/:member_profile_id/search",
		Method:   "GET",
		Response: "TFootstep[]",
		Note:     "Returns all footsteps for the specified employee (user) on the current branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization/branch not found"})
		}
		memberProfileId, err := horizon.EngineUUIDParam(ctx, "member_profile_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member profile ID"})
		}
		memberProfile, err := c.model.MemberProfileManager.GetByID(context, *memberProfileId)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Member profile not found: " + err.Error()})
		}
		footstep, err := c.model.GetFootstepByUserOrganization(context, *memberProfile.UserID, userOrg.OrganizationID, *userOrg.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve footsteps for employee: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.FootstepManager.Pagination(context, ctx, footstep))
	})

	// GET /footstep/branch: Get all footsteps for the current user's branch.
	req.RegisterRoute(horizon.Route{
		Route:    "/footstep/branch/search",
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
		return ctx.JSON(http.StatusOK, c.model.FootstepManager.Pagination(context, ctx, footstep))
	})

	// GET /footstep/user-organization/:user_organization_id/search: Get footsteps for a user organization on the current branch.
	req.RegisterRoute(horizon.Route{
		Route:    "/footstep/user-organization/:user_organization_id/search",
		Method:   "GET",
		Response: "TFootstep[]",
		Note:     "Returns footsteps for the specified user-organization on the current branch if the user is a member, employee, or owner.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrgId, err := horizon.EngineUUIDParam(ctx, "user_organization_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid user_organization_id"})
		}
		// You may want to add permission checks here to ensure the current user can access this user organization

		// Retrieve the user organization record to get its UserID, OrganizationID, and BranchID
		targetUserOrg, err := c.model.UserOrganizationManager.GetByID(context, *userOrgId)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "User organization not found"})
		}

		footstep, err := c.model.GetFootstepByUserOrganization(context, targetUserOrg.UserID, targetUserOrg.OrganizationID, *targetUserOrg.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve footsteps for user organization: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.FootstepManager.Pagination(context, ctx, footstep))
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
