package controller_v1

import (
	"net/http"

	"github.com/Lands-Horizon-Corp/e-coop-server/services/handlers"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/event"
	modelCore "github.com/Lands-Horizon-Corp/e-coop-server/src/model/model_core"
	"github.com/labstack/echo/v4"
)

// GeneratedReports manages endpoints for generated report resources.
func (c *Controller) GeneratedReports() {
	req := c.provider.Service.Request

	// GET /generated-report: Get all generated reports for the current user. (NO footstep)
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/generated-report",
		Method:       "GET",
		ResponseType: modelCore.GeneratedReportResponse{},
		Note:         "Returns all generated reports for the currently authenticated user.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userToken.CurrentUser(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or user not found"})
		}
		generatedReports, err := c.modelCore.GetGenerationReportByUser(context, user.ID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve generated reports: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.modelCore.GeneratedReportManager.Filtered(context, ctx, generatedReports))
	})

	// GET /generated-report/:generated_report_id: Get a specific generated report by ID. (NO footstep)
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/generated-report/:generated_report_id",
		Method:       "GET",
		ResponseType: modelCore.GeneratedReportResponse{},
		Note:         "Returns a specific generated report by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		generatedReportID, err := handlers.EngineUUIDParam(ctx, "generated_report_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid generated report ID"})
		}
		generatedReport, err := c.modelCore.GeneratedReportManager.GetByID(context, *generatedReportID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Generated report not found"})
		}
		return ctx.JSON(http.StatusOK, c.modelCore.GeneratedReportManager.ToModel(generatedReport))
	})

	// DELETE /generated-report/:generated_report_id: Delete a specific generated report by ID and its associated file. (WITH footstep)
	req.RegisterRoute(handlers.Route{
		Route:  "/api/v1/generated-report/:generated_report_id",
		Method: "DELETE",
		Note:   "Deletes the specified generated report by its ID and the associated file.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		generatedReportID, err := handlers.EngineUUIDParam(ctx, "generated_report_id")
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Generated report delete failed (/generated-report/:generated_report_id), invalid generated report ID.",
				Module:      "GeneratedReport",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid generated report ID"})
		}
		generatedReport, err := c.modelCore.GeneratedReportManager.GetByID(context, *generatedReportID)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Generated report delete failed (/generated-report/:generated_report_id), not found.",
				Module:      "GeneratedReport",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Generated report not found"})
		}
		if generatedReport.MediaID != nil {
			if err := c.modelCore.MediaDelete(context, *generatedReport.MediaID); err != nil {
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "delete-error",
					Description: "Generated report delete failed (/generated-report/:generated_report_id), media delete error: " + err.Error(),
					Module:      "GeneratedReport",
				})
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete associated media file: " + err.Error()})
			}
		}
		if err := c.modelCore.GeneratedReportManager.DeleteByID(context, generatedReport.ID); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Generated report delete failed (/generated-report/:generated_report_id), db error: " + err.Error(),
				Module:      "GeneratedReport",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete generated report: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "delete-success",
			Description: "Deleted generated report (/generated-report/:generated_report_id): " + generatedReport.Name,
			Module:      "GeneratedReport",
		})
		return ctx.NoContent(http.StatusNoContent)
	})
}
