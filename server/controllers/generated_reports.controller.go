package controllers

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"horizon.com/server/horizon"
)

// GET /generated-report
func (c *Controller) GeneratedReportList(ctx echo.Context) error {
	generated_report, err := c.generatedReport.Manager.List()
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return ctx.JSON(http.StatusOK, c.model.GeneratedReportModels(generated_report))
}

// GET /generated-report/:generated_report_id
func (c *Controller) GeneratedReportGetByID(ctx echo.Context) error {
	id, err := horizon.EngineUUIDParam(ctx, "generated_report_id")
	if err != nil {
		return err
	}
	generated_report, err := c.generatedReport.Manager.GetByID(*id)
	if err != nil {
		return ctx.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
	}
	return ctx.JSON(http.StatusOK, c.model.GeneratedReportModel(generated_report))
}

// DELETE /generated-report/generated_report_id
func (c *Controller) GeneratedReportDelete(ctx echo.Context) error {
	id, err := horizon.EngineUUIDParam(ctx, "generated_report_id")
	if err != nil {
		return err
	}
	if err := c.generatedReport.Manager.DeleteByID(*id); err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return ctx.NoContent(http.StatusNoContent)
}

// GET generated-report/user/:user_id
func (c *Controller) GeneratedReportListByUser(ctx echo.Context) error {
	id, err := horizon.EngineUUIDParam(ctx, "user_id")
	if err != nil {
		return err
	}
	generated_report, err := c.generatedReport.ListByUser(id)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return ctx.JSON(http.StatusOK, c.model.GeneratedReportModels(generated_report))
}

// GET generated-report/branch/:branch_id
func (c *Controller) GeneratedReportListByBranch(ctx echo.Context) error {
	id, err := horizon.EngineUUIDParam(ctx, "branch_id")
	if err != nil {
		return err
	}
	generated_report, err := c.generatedReport.ListByBranch(*id)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return ctx.JSON(http.StatusOK, c.model.GeneratedReportModels(generated_report))
}

// GET generated-report/organization/:organization_id
func (c *Controller) GeneratedReportListByOrganization(ctx echo.Context) error {
	id, err := horizon.EngineUUIDParam(ctx, "organization_id")
	if err != nil {
		return err
	}
	generated_report, err := c.generatedReport.ListByOrganization(*id)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return ctx.JSON(http.StatusOK, c.model.GeneratedReportModels(generated_report))
}

// GET generated-report/organization/:organization_id/branch/:branch_id
func (c *Controller) GeneratedReportListByOrganizationBranch(ctx echo.Context) error {
	orgId, err := horizon.EngineUUIDParam(ctx, "organization_id")
	if err != nil {
		return err
	}
	branchId, err := horizon.EngineUUIDParam(ctx, "branch_id")
	if err != nil {
		return err
	}
	generated_report, err := c.generatedReport.ListByOrganizationBranch(*branchId, *orgId)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return ctx.JSON(http.StatusOK, c.model.GeneratedReportModels(generated_report))
}

// GET generated-report/user/:user_id/organization/:organization_id/branch/:branch_id
func (c *Controller) GeneratedReportListByUserOrganizationBranch(ctx echo.Context) error {
	userId, err := horizon.EngineUUIDParam(ctx, "user_id")
	if err != nil {
		return err
	}
	orgId, err := horizon.EngineUUIDParam(ctx, "organization_id")
	if err != nil {
		return err
	}
	branchId, err := horizon.EngineUUIDParam(ctx, "branch_id")
	if err != nil {
		return err
	}
	generated_report, err := c.generatedReport.ListByUserOrganizationBranch(userId, *branchId, *orgId)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return ctx.JSON(http.StatusOK, c.model.GeneratedReportModels(generated_report))

}

// GET generated-report/user/:user_id/branch/:branch_id
func (c *Controller) GeneratedReportUserBranch(ctx echo.Context) error {
	userId, err := horizon.EngineUUIDParam(ctx, "user_id")
	if err != nil {
		return err
	}
	branchId, err := horizon.EngineUUIDParam(ctx, "branch_id")
	if err != nil {
		return err
	}
	generated_report, err := c.generatedReport.ListByUserBranch(userId, *branchId)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return ctx.JSON(http.StatusOK, c.model.GeneratedReportModels(generated_report))
}

// GET generated-report/user/:user_id/organization/:organization_id
func (c *Controller) GeneratedReportListByUserOrganization(ctx echo.Context) error {
	userId, err := horizon.EngineUUIDParam(ctx, "user_id")
	if err != nil {
		return err
	}
	orgId, err := horizon.EngineUUIDParam(ctx, "organization_id")
	if err != nil {
		return err
	}
	generated_report, err := c.generatedReport.ListByUserOrganization(userId, *orgId)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return ctx.JSON(http.StatusOK, c.model.GeneratedReportModels(generated_report))
}
