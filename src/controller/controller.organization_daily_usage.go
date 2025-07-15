package controller

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/lands-horizon/horizon-server/services/horizon"
)

func (c *Controller) OrganizationDailyUsage() {
	req := c.provider.Service.Request

	req.RegisterRoute(horizon.Route{
		Route:    "/organization-daily-usage",
		Method:   "GET",
		Response: "TOrganizationDailyUsage[]",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return err
		}
		dailyUsage, err := c.model.GetOrganizationDailyUsageByOrganization(context, userOrg.OrganizationID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.OrganizationDailyUsageManager.ToModels(dailyUsage))
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/organization-daily-usage/:organization_daily_usage_id",
		Method:   "GET",
		Response: "TOrganizationDailyUsage",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		dailyUsageId, err := horizon.EngineUUIDParam(ctx, "organization_daily_usage_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid Organization daily usage ID")
		}
		dailyUsage, err := c.model.OrganizationDailyUsageManager.GetByIDRaw(context, *dailyUsageId)
		if err != nil {
			return c.InternalServerError(ctx, err)
		}
		return ctx.JSON(http.StatusOK, dailyUsage)
	})
}
