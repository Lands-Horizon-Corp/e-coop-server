package controller_v1

import (
	"net/http"

	"github.com/Lands-Horizon-Corp/e-coop-server/services/handlers"
	modelcore "github.com/Lands-Horizon-Corp/e-coop-server/src/model/modelcore"
	"github.com/labstack/echo/v4"
)

func (c *Controller) OrganizationDailyUsage() {
	req := c.provider.Service.Request

	// Get daily usage records for the current user's organization
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/organization-daily-usage",
		Method:       "GET",
		Note:         "Returns all daily usage records for the current user's organization.",
		ResponseType: modelcore.OrganizationDailyUsageResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		dailyUsage, err := c.modelcore.GetOrganizationDailyUsageByOrganization(context, userOrg.OrganizationID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve organization daily usage records: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.modelcore.OrganizationDailyUsageManager.Filtered(context, ctx, dailyUsage))
	})

	// Get a specific organization daily usage record by its ID
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/organization-daily-usage/:organization_daily_usage_id",
		Method:       "GET",
		Note:         "Returns a specific organization daily usage record by its ID.",
		ResponseType: modelcore.OrganizationDailyUsageResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		dailyUsageId, err := handlers.EngineUUIDParam(ctx, "organization_daily_usage_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid organization_daily_usage_id: " + err.Error()})
		}
		dailyUsage, err := c.modelcore.OrganizationDailyUsageManager.GetByIDRaw(context, *dailyUsageId)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve organization daily usage by ID: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, dailyUsage)
	})
}
