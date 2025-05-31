package controller

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/lands-horizon/horizon-server/services/horizon"
)

func (c *Controller) GeneratedReports() {
	req := c.provider.Service.Request

	// Route: Get all generated reports for the current user
	req.RegisterRoute(horizon.Route{
		Route:    "/generated-report",
		Method:   "GET",
		Response: "TGeneratedReport[]",
		Note:     "Retrieves all generated reports for the current user.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userToken.CurrentUser(context, ctx)
		if err != nil {
			return err
		}
		generatedReports, err := c.model.GetGenerationReportByUser(context, user.ID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.GeneratedReportManager.ToModels(generatedReports))
	})

	// Route: Get a specific generated report by ID
	req.RegisterRoute(horizon.Route{
		Route:    "/generated-report/:generated_report_id",
		Method:   "GET",
		Response: "TGeneratedReport",
		Note:     "Retrieves a specific generated report by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		generatedReportID, err := horizon.EngineUUIDParam(ctx, "generated_report_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid generated report ID")
		}
		generatedReport, err := c.model.GeneratedReportManager.GetByID(context, *generatedReportID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return ctx.JSON(http.StatusAccepted, c.model.GeneratedReportManager.ToModel(generatedReport))
	})

	// Route: Delete a specific generated report by ID
	req.RegisterRoute(horizon.Route{
		Route:  "/generated-report/:generated_report_id",
		Method: "DELETE",
		Note:   "Deletes a specific generated report by its ID and the associated file.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		generatedReportID, err := horizon.EngineUUIDParam(ctx, "generated_report_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid generated report ID")
		}
		generatedReport, err := c.model.GeneratedReportManager.GetByID(context, *generatedReportID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		if err := c.model.MediaDelete(context, *generatedReport.MediaID); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		if err := c.model.GeneratedReportManager.DeleteByID(context, generatedReport.ID); err != nil {
			return c.InternalServerError(ctx, err)
		}
		return ctx.NoContent(http.StatusNoContent)
	})
}
