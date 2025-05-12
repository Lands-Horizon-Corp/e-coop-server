package controllers

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"horizon.com/server/horizon"
)

// GET /organization_daily_usage
func (c *Controller) OrganizationDailyUsageList(ctx echo.Context) error {
	organization_daily_usage, err := c.organizationDailyUsage.Manager.List()
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return ctx.JSON(http.StatusOK, c.model.OrganizationDailyUsageModels(organization_daily_usage))
}

// GET /organization_daily_usage/:organization_daily_usage_id
func (c *Controller) OrganizationDailyUsageGetByID(ctx echo.Context) error {
	id, err := horizon.EngineUUIDParam(ctx, "organization_daily_usage_id")
	if err != nil {
		return err
	}
	organization_daily_usage, err := c.organizationDailyUsage.Manager.GetByID(*id)
	if err != nil {
		return ctx.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
	}
	return ctx.JSON(http.StatusOK, c.model.OrganizationDailyUsageModel(organization_daily_usage))
}

// DELETE /organization_daily_usage/:organization_daily_usage_id
func (c *Controller) OrganizationDailyUsageDelete(ctx echo.Context) error {
	id, err := horizon.EngineUUIDParam(ctx, "organization_daily_usage_id")
	if err != nil {
		return err
	}
	if err := c.organizationDailyUsage.Manager.DeleteByID(*id); err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return ctx.NoContent(http.StatusNoContent)
}

// GET /organization_daily_usage/:organization_daily_usage_id/organization/:organization_id
func (c *Controller) OrganizationDailyUsageListByOrganization(ctx echo.Context) error {
	id, err := horizon.EngineUUIDParam(ctx, "organization_daily_usage_id")
	if err != nil {
		return err
	}
	organization_daily_usage, err := c.organizationDailyUsage.ListByOrganization(*id)
	if err != nil {
		return ctx.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
	}
	return ctx.JSON(http.StatusOK, c.model.OrganizationDailyUsageModels(organization_daily_usage))
}
