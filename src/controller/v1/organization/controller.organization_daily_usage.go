package organization

import (
	"net/http"

	"github.com/Lands-Horizon-Corp/e-coop-server/helpers"
	"github.com/Lands-Horizon-Corp/e-coop-server/horizon"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/db/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/event"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/types"
	"github.com/labstack/echo/v4"
)

func OrganizationDailyUsageController(service *horizon.HorizonService) {
	req := service.API

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/organization-daily-usage",
		Method:       "GET",
		Note:         "Returns all daily usage records for the current user's organization.",
		ResponseType: types.OrganizationDailyUsageResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		dailyUsage, err := core.GetOrganizationDailyUsageByOrganization(context, service, userOrg.OrganizationID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve organization daily usage records: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, core.OrganizationDailyUsageManager(service).ToModels(dailyUsage))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/organization-daily-usage/:organization_daily_usage_id",
		Method:       "GET",
		Note:         "Returns a specific organization daily usage record by its ID.",
		ResponseType: types.OrganizationDailyUsageResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		dailyUsageID, err := helpers.EngineUUIDParam(ctx, "organization_daily_usage_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid organization_daily_usage_id: " + err.Error()})
		}
		dailyUsage, err := core.OrganizationDailyUsageManager(service).GetByIDRaw(context, *dailyUsageID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve organization daily usage by ID: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, dailyUsage)
	})
}
