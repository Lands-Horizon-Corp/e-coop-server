package controller_v1

import (
	"net/http"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/services/handlers"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/event"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/model/modelCore"
	"github.com/labstack/echo/v4"
)

// FootstepController manages endpoints related to footstep records.
func (c *Controller) FootstepController() {
	req := c.provider.Service.Request

	// POST /footstep: Create a new footstep. (WITH footstep)
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/footstep",
		Method:       "POST",
		Note:         "Creates a new footstep record for the current user's organization and branch.",
		ResponseType: modelCore.FootstepResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := c.modelCore.FootstepManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Footstep creation failed (/footstep), validation error: " + err.Error(),
				Module:      "Footstep",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid footstep data: " + err.Error()})
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Footstep creation failed (/footstep), user org error: " + err.Error(),
				Module:      "Footstep",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if user.BranchID == nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Footstep creation failed (/footstep), user not assigned to branch.",
				Module:      "Footstep",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		longitude := handlers.ParseCoordinate(ctx.Request().Header.Get("X-Longitude"))
		latitude := handlers.ParseCoordinate(ctx.Request().Header.Get("X-Latitude"))
		footstep := &modelCore.Footstep{
			Activity: req.Activity,
			UserType: user.UserType,
			Module:   req.Module,

			Description:    req.Description,
			Latitude:       &latitude,
			Longitude:      &longitude,
			IPAddress:      ctx.RealIP(),
			UserAgent:      ctx.Request().UserAgent(),
			Referer:        ctx.Request().Referer(),
			Location:       ctx.Request().Header.Get("Location"),
			AcceptLanguage: ctx.Request().Header.Get("Accept-Language"),
			Timestamp:      time.Now().UTC(),
			CreatedAt:      time.Now().UTC(),
			CreatedByID:    user.UserID,
			UpdatedAt:      time.Now().UTC(),
			UpdatedByID:    user.UserID,
			UserID:         &user.UserID,
			BranchID:       user.BranchID,
			OrganizationID: &user.OrganizationID,
		}

		if err := c.modelCore.FootstepManager.Create(context, footstep); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Footstep creation failed (/footstep), db error: " + err.Error(),
				Module:      "Footstep",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create footstep: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "create-success",
			Description: "Created footstep (/footstep): " + footstep.Activity,
			Module:      "Footstep",
		})
		return ctx.JSON(http.StatusCreated, c.modelCore.FootstepManager.ToModel(footstep))
	})

	// GET /footstep/me: Get all footsteps for the currently logged-in user.
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/footstep/me/search",
		Method:       "GET",
		ResponseType: modelCore.FootstepResponse{},
		Note:         "Returns all footsteps for the currently authenticated user.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userToken.CurrentUser(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or user not found"})
		}
		footstep, err := c.modelCore.GetFootstepByUser(context, user.ID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve user's footsteps: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.modelCore.FootstepManager.Pagination(context, ctx, footstep))
	})

	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/footstep/member-profile/:member_profile_id/search",
		Method:       "GET",
		ResponseType: modelCore.FootstepResponse{},
		Note:         "Returns all footsteps for the specified employee (user) on the current branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization/branch not found"})
		}
		memberProfileId, err := handlers.EngineUUIDParam(ctx, "member_profile_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member profile ID"})
		}
		memberProfile, err := c.modelCore.MemberProfileManager.GetByID(context, *memberProfileId)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Member profile not found: " + err.Error()})
		}
		if memberProfile.UserID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Member profile UserID is missing"})
		}
		if userOrg.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User branch ID is missing"})
		}
		footstep, err := c.modelCore.GetFootstepByUserOrganization(context, *memberProfile.UserID, userOrg.OrganizationID, *userOrg.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve footsteps for employee: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.modelCore.FootstepManager.Pagination(context, ctx, footstep))
	})

	// GET /footstep/branch: Get all footsteps for the current user's branch.
	req.RegisterRoute(handlers.Route{
		Route:  "/api/v1/footstep/branch/search",
		Method: "GET",
		Note:   "Returns all footsteps for the current user's organization and branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization/branch not found"})
		}
		footstep, err := c.modelCore.GetFootstepByBranch(context, userOrg.OrganizationID, *userOrg.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve footsteps for branch: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.modelCore.FootstepManager.Pagination(context, ctx, footstep))
	})

	// GET /footstep/user-organization/:user_organization_id/search: Get footsteps for a user organization on the current branch.
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/footstep/user-organization/:user_organization_id/search",
		Method:       "GET",
		ResponseType: modelCore.FootstepResponse{},
		Note:         "Returns footsteps for the specified user-organization on the current branch if the user is a member, employee, or owner.",
	}, func(ctx echo.Context) error {

		context := ctx.Request().Context()
		userOrgId, err := handlers.EngineUUIDParam(ctx, "user_organization_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid user_organization_id"})
		}
		targetUserOrg, err := c.modelCore.UserOrganizationManager.GetByID(context, *userOrgId)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "User organization not found"})
		}

		footstep, err := c.modelCore.GetFootstepByUserOrganization(context, targetUserOrg.UserID, targetUserOrg.OrganizationID, *targetUserOrg.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve footsteps for user organization: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.modelCore.FootstepManager.Pagination(context, ctx, footstep))
	})

	// GET /footstep/:footstep_id: Get a specific footstep by ID.
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/footstep/:footstep_id",
		Method:       "GET",
		Note:         "Returns a specific footstep record by its ID.",
		ResponseType: modelCore.FootstepResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		footstepId, err := handlers.EngineUUIDParam(ctx, "footstep_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid footstep ID"})
		}
		footstep, err := c.modelCore.FootstepManager.GetByIDRaw(context, *footstepId)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Footstep record not found"})
		}
		return ctx.JSON(http.StatusOK, footstep)
	})

	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/footstep/current/me/branch/search",
		Method:       "GET",
		Note:         "Returns footsteps for the currently authenticated user on their current branch.",
		ResponseType: modelCore.FootstepResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()

		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization/branch not found"})
		}
		if userOrg.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User branch ID is missing"})
		}
		footstep, err := c.modelCore.GetFootstepByUserOrganization(context, userOrg.UserID, userOrg.OrganizationID, *userOrg.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve footsteps for user on branch: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.modelCore.FootstepManager.Pagination(context, ctx, footstep))
	})
}
