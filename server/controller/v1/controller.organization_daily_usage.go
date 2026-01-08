package v1

import (
	"net/http"

	"github.com/Lands-Horizon-Corp/e-coop-server/server/model/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/services/handlers"
	"github.com/labstack/echo/v4"
)

func (c *Controller) organizationDailyUsage() {
	req := c.provider.Service.Request

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/organization-daily-usage",
		Method:       "GET",
		Note:         "Returns all daily usage records for the current user's organization.",
		ResponseType: core.OrganizationDailyUsageResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.event.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		dailyUsage, err := c.core.GetOrganizationDailyUsageByOrganization(context, userOrg.OrganizationID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve organization daily usage records: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.core.OrganizationDailyUsageManager().ToModels(dailyUsage))
	})

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/organization-daily-usage/:organization_daily_usage_id",
		Method:       "GET",
		Note:         "Returns a specific organization daily usage record by its ID.",
		ResponseType: core.OrganizationDailyUsageResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		dailyUsageID, err := handlers.EngineUUIDParam(ctx, "organization_daily_usage_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid organization_daily_usage_id: " + err.Error()})
		}
		dailyUsage, err := c.core.OrganizationDailyUsageManager().GetByIDRaw(context, *dailyUsageID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve organization daily usage by ID: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, dailyUsage)
	})
}
