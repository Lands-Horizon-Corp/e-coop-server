package controller

import (
	"context"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/lands-horizon/horizon-server/services/horizon"
)

func (c *Controller) BranchController() {
	req := c.provider.Service.Request
	req.RegisterRoute(horizon.Route{
		Route:    "/branch",
		Method:   "GET",
		Response: "TOrganization[]",
	}, func(ctx echo.Context) error {
		organization, err := c.model.GetPublicOrganization(context.Background())
		if err != nil {
			return c.InternalServerError(ctx, err)
		}
		return ctx.JSON(http.StatusOK, c.model.OrganizationManager.ToModels(organization))
	})
}
