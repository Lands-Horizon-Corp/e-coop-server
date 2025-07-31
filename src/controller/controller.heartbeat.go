package controller

import (
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/lands-horizon/horizon-server/services/handlers"
	"github.com/lands-horizon/horizon-server/src/model"
)

func (c *Controller) Heartbeat() {
	req := c.provider.Service.Request

	req.RegisterRoute(handlers.Route{
		Route:  "/heartbeat/online",
		Method: "GET",
		Note:   "This endpoint checks user when online status and service health.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		userOrg.Status = model.UserOrganizationStatusOnline
		userOrg.LastOnlineAt = time.Now()
		if err := c.model.UserOrganizationManager.Update(context, userOrg); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update user organization status"})
		}
		c.provider.Service.Broker.Dispatch(context, []string{
			fmt.Sprintf("user_orgainzation.status.branch.%s", userOrg.BranchID),
			fmt.Sprintf("user_orgainzation.status.organization.%s", userOrg.OrganizationID),
		}, nil)
		return ctx.NoContent(http.StatusNoContent)
	})

	req.RegisterRoute(handlers.Route{
		Route:  "/heartbeat/offline",
		Method: "GET",
		Note:   "This endpoint sets user offline status.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		userOrg.Status = model.UserOrganizationStatusOffline
		userOrg.LastOnlineAt = time.Now()
		if err := c.model.UserOrganizationManager.Update(context, userOrg); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update user organization status"})
		}
		c.provider.Service.Broker.Dispatch(context, []string{
			fmt.Sprintf("user_orgainzation.status.branch.%s", userOrg.BranchID),
			fmt.Sprintf("user_orgainzation.status.organization.%s", userOrg.OrganizationID),
		}, nil)
		return ctx.NoContent(http.StatusNoContent)
	})

	req.RegisterRoute(handlers.Route{
		Route:        "/heartbeat/status",
		Method:       "GET",
		ResponseType: model.UserOrganizationStatusResponse{},
		Note:         "This endpoint retrieves the current user organization status.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		userOrganizations, err := c.model.UserOrganizationManager.Find(context, &model.UserOrganization{
			BranchID:       userOrg.BranchID,
			OrganizationID: userOrg.OrganizationID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve user organization status"})
		}
		statuses := c.model.UserOrganizationManager.Filtered(context, ctx, userOrganizations)
		countMemberOnline := 0
		countEmployeeOnline := 0

		for _, org := range statuses {
			if org.Status == model.UserOrganizationStatusOnline {
				switch org.UserType {
				case "member":
					countMemberOnline++
				case "employee":
					countEmployeeOnline++
				}
			}
		}
		totalOnline := countMemberOnline + countEmployeeOnline
		return ctx.JSON(http.StatusOK, model.UserOrganizationStatusResponse{
			UserOrganization:    statuses,
			TotalOnline:         totalOnline,
			CountMemberOnline:   countMemberOnline,
			CountEmployeeOnline: countEmployeeOnline,
		})
	})

}
