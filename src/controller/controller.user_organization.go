package controller

import (
	"context"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/lands-horizon/horizon-server/services/horizon"
)

func (c *Controller) UserOrganinzationController() {

	req := c.provider.Service.Request

	req.RegisterRoute(horizon.Route{
		Route:    "/user-organization/user/:user_id",
		Method:   "GET",
		Response: "TUserOrganization",
		Note:     "Retrieve all user organizations. Use query param `pending=true` to include pending organizations.",
	}, func(ctx echo.Context) error {
		context := context.Background()
		userId, err := horizon.EngineUUIDParam(ctx, "user_id")
		isPending := ctx.QueryParam("pending") == "true"
		if err != nil {
			return err
		}
		user, err := c.model.UserManager.GetByID(context, *userId)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
		}
		userOrganization, err := c.model.GetUserOrganizationByUser(context, user.ID, isPending)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.UserOrganizationManager.ToModels(userOrganization))
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/user-organization/organization/:organization_id",
		Method:   "GET",
		Response: "TUserOrganization",
		Note:     "Retrieve all user organizations across all branches of a specific organization. Use query param `pending=true` to include pending organizations.",
	}, func(ctx echo.Context) error {
		context := context.Background()
		organizationId, err := horizon.EngineUUIDParam(ctx, "organization_id")
		isPending := ctx.QueryParam("pending") == "true"
		if err != nil {
			return err
		}

		organization, err := c.model.OrganizationManager.GetByID(context, *organizationId)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
		}

		userOrganization, err := c.model.GetUserOrganizationByOrganization(context, organization.ID, isPending)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}

		return ctx.JSON(http.StatusOK, c.model.UserOrganizationManager.ToModels(userOrganization))
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/user-organization/branch/:branch_id",
		Method:   "GET",
		Response: "TUserOrganization",
		Note:     "Retrieve all user organizations from a specific branch. Use query param `pending=true` to include pending organizations.",
	}, func(ctx echo.Context) error {
		context := context.Background()
		branchId, err := horizon.EngineUUIDParam(ctx, "branch_id")
		isPending := ctx.QueryParam("pending") == "true"
		if err != nil {
			return err
		}

		branch, err := c.model.BranchManager.GetByID(context, *branchId)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
		}

		userOrganization, err := c.model.GetUserOrganizationByBranch(context, branch.OrganizationID, branch.ID, isPending)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}

		return ctx.JSON(http.StatusOK, c.model.UserOrganizationManager.ToModels(userOrganization))
	})

	req.RegisterRoute(horizon.Route{
		Route:  "/user-organization/:user_organization_id/switch",
		Method: "POST",
		Note:   "Switch organization and branch stored in JWT (no database impact).",
	}, func(ctx echo.Context) error {
		context := context.Background()
		organizationId, err := horizon.EngineUUIDParam(ctx, "user_organization_id")
		if err != nil {
			return err
		}
		userOrganization, err := c.model.UserOrganizationManager.GetByID(context, *organizationId)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		// witch here
		return ctx.NoContent(http.StatusNoContent)
	})

	req.RegisterRoute(horizon.Route{
		Route:  "/user-organization/:user_organization_id/unswitch",
		Method: "POST",
		Note:   "Remove organization and branch from JWT (no database impact).",
	}, func(ctx echo.Context) error {
		return nil
	})

	req.RegisterRoute(horizon.Route{
		Route:  "/user-organization/:organization_id/seed",
		Method: "POST",
		Note:   "Seed all branches inside an organization when first created.",
	}, func(ctx echo.Context) error {
		return nil
	})

	req.RegisterRoute(horizon.Route{
		Route:  "/user-organization/:user_organization_id/developer-key-refresh",
		Method: "POST",
		Note:   "Refresh developer key associated with the user organization.",
	}, func(ctx echo.Context) error {
		return nil
	})

	req.RegisterRoute(horizon.Route{
		Route:  "/user-organization/:user_organization_id/invitation-code/:code",
		Method: "POST",
		Note:   "Join organization and branch using an invitation code.",
	}, func(ctx echo.Context) error {
		return nil
	})

	req.RegisterRoute(horizon.Route{
		Route:  "/user-organization/:user_organization_id/join",
		Method: "POST",
		Note:   "Join an organization and branch that is already created.",
	}, func(ctx echo.Context) error {
		return nil
	})

	req.RegisterRoute(horizon.Route{
		Route:  "/user-organization/:user_organization_id/leave",
		Method: "POST",
		Note:   "Leave a specific organization and branch that is already joined.",
	}, func(ctx echo.Context) error {
		return nil
	})

	req.RegisterRoute(horizon.Route{
		Route:  "/user-organization/:user_organization_id/can-join-employee",
		Method: "GET",
		Note:   "Check if the user can join as an employee.",
	}, func(ctx echo.Context) error {
		return nil
	})

	req.RegisterRoute(horizon.Route{
		Route:  "/user-organization/:user_organization_id/can-join-member",
		Method: "GET",
		Note:   "Check if the user can join as a member.",
	}, func(ctx echo.Context) error {
		return nil
	})
}
