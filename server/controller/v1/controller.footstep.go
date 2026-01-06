package v1

import (
	"net/http"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/server/event"
	"github.com/Lands-Horizon-Corp/e-coop-server/server/model/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/services/handlers"
	"github.com/labstack/echo/v4"
)

func (c *Controller) footstepController() {
	req := c.provider.Service.Request

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/footstep",
		Method:       "POST",
		Note:         "Creates a new footstep record for the current user's organization and branch.",
		ResponseType: core.FootstepResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := c.core.FootstepManager().Validate(ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Footstep creation failed (/footstep), validation error: " + err.Error(),
				Module:      "Footstep",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid footstep data: " + err.Error()})
		}
		userOrg, err := c.token.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Footstep creation failed (/footstep), user org error: " + err.Error(),
				Module:      "Footstep",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if userOrg.BranchID == nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Footstep creation failed (/footstep), user not assigned to branch.",
				Module:      "Footstep",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		longitude := handlers.ParseCoordinate(ctx.Request().Header.Get("X-Longitude"))
		latitude := handlers.ParseCoordinate(ctx.Request().Header.Get("X-Latitude"))
		footstep := &core.Footstep{
			Activity: req.Activity,
			UserType: userOrg.UserType,
			Module:   req.Module,

			Description:    req.Description,
			Latitude:       &latitude,
			Longitude:      &longitude,
			IPAddress:      handlers.GetClientIP(ctx),
			UserAgent:      handlers.GetUserAgent(ctx),
			Referer:        ctx.Request().Referer(),
			Location:       ctx.Request().Header.Get("Location"),
			AcceptLanguage: ctx.Request().Header.Get("Accept-Language"),
			Timestamp:      time.Now().UTC(),
			CreatedAt:      time.Now().UTC(),
			CreatedByID:    userOrg.UserID,
			UpdatedAt:      time.Now().UTC(),
			UpdatedByID:    userOrg.UserID,
			UserID:         &userOrg.UserID,
			BranchID:       userOrg.BranchID,
			OrganizationID: &userOrg.OrganizationID,
		}

		if err := c.core.FootstepManager().Create(context, footstep); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Footstep creation failed (/footstep), db error: " + err.Error(),
				Module:      "Footstep",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create footstep: " + err.Error()})
		}
		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "create-success",
			Description: "Created footstep (/footstep): " + footstep.Activity,
			Module:      "Footstep",
		})
		return ctx.JSON(http.StatusCreated, c.core.FootstepManager().ToModel(footstep))
	})

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/footstep/me/search",
		Method:       "GET",
		ResponseType: core.FootstepResponse{},
		Note:         "Returns all footsteps for the currently authenticated user.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.token.CurrentUser(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or user not found"})
		}
		footstep, err := c.core.FootstepManager().NormalPagination(context, ctx, &core.Footstep{
			UserID: &user.ID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve user's footsteps: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, footstep)
	})

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/footstep/member-profile/:member_profile_id/search",
		Method:       "GET",
		ResponseType: core.FootstepResponse{},
		Note:         "Returns all footsteps for the specified employee (user) on the current branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.token.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization/branch not found"})
		}
		memberProfileID, err := handlers.EngineUUIDParam(ctx, "member_profile_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member profile ID"})
		}
		memberProfile, err := c.core.MemberProfileManager().GetByID(context, *memberProfileID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Member profile not found: " + err.Error()})
		}
		if memberProfile.UserID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Member profile UserID is missing"})
		}
		if userOrg.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User branch ID is missing"})
		}
		footstep, err := c.core.FootstepManager().NormalPagination(context, ctx, &core.Footstep{
			UserID:         &userOrg.UserID,
			BranchID:       userOrg.BranchID,
			OrganizationID: &userOrg.OrganizationID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve footsteps for employee: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, footstep)
	})

	req.RegisterWebRoute(handlers.Route{
		Route:  "/api/v1/footstep/branch/search",
		Method: "GET",
		Note:   "Returns all footsteps for the current user's organization and branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.token.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization/branch not found"})
		}
		footstep, err := c.core.FootstepManager().NormalPagination(context, ctx, &core.Footstep{
			BranchID:       userOrg.BranchID,
			OrganizationID: &userOrg.OrganizationID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve footsteps for branch: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, footstep)
	})

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/footstep/user-organization/:user_organization_id/search",
		Method:       "GET",
		ResponseType: core.FootstepResponse{},
		Note:         "Returns footsteps for the specified user-organization on the current branch if the user is a member, employee, or owner.",
	}, func(ctx echo.Context) error {

		context := ctx.Request().Context()
		userOrgID, err := handlers.EngineUUIDParam(ctx, "user_organization_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid user_organization_id"})
		}
		targetUserOrg, err := c.core.UserOrganizationManager().GetByID(context, *userOrgID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "User organization not found"})
		}

		footstep, err := c.core.FootstepManager().NormalPagination(context, ctx, &core.Footstep{
			BranchID:       targetUserOrg.BranchID,
			OrganizationID: &targetUserOrg.OrganizationID,
			UserID:         &targetUserOrg.UserID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve footsteps for user organization: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, footstep)
	})

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/footstep/:footstep_id",
		Method:       "GET",
		Note:         "Returns a specific footstep record by its ID.",
		ResponseType: core.FootstepResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		footstepID, err := handlers.EngineUUIDParam(ctx, "footstep_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid footstep ID"})
		}
		footstep, err := c.core.FootstepManager().GetByIDRaw(context, *footstepID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Footstep record not found"})
		}
		return ctx.JSON(http.StatusOK, footstep)
	})

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/footstep/current/me/branch/search",
		Method:       "GET",
		Note:         "Returns footsteps for the currently authenticated user on their current branch.",
		ResponseType: core.FootstepResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()

		userOrg, err := c.token.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization/branch not found"})
		}
		if userOrg.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User branch ID is missing"})
		}
		footstep, err := c.core.FootstepManager().NormalPagination(context, ctx, &core.Footstep{
			BranchID:       userOrg.BranchID,
			OrganizationID: &userOrg.OrganizationID,
			UserID:         &userOrg.UserID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve footsteps for user on branch: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, footstep)
	})
}
