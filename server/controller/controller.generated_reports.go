package v1

import (
	"net/http"

	"github.com/Lands-Horizon-Corp/e-coop-server/server/model/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/services/handlers"
	"github.com/labstack/echo/v4"
)

// GeneratedReports manages endpoints for generated report resources.
func (c *Controller) generatedReports() {
	req := c.provider.Service.Request

	// GET /generated-report/:generated_report_id: Get a specific generated report by ID. (NO footstep)
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/generated-report/:generated_report_id",
		Method:       "GET",
		ResponseType: core.GeneratedReportResponse{},
		Note:         "Returns a specific generated report by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		generatedReportID, err := handlers.EngineUUIDParam(ctx, "generated_report_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid generated report ID"})
		}
		generatedReport, err := c.core.GeneratedReportManager.GetByID(context, *generatedReportID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Generated report not found"})
		}
		return ctx.JSON(http.StatusOK, c.core.GeneratedReportManager.ToModel(generatedReport))
	})

	// GET /api/v1/generated-report/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/generated-report/search",
		Method:       "GET",
		ResponseType: core.GeneratedReportResponse{},
		Note:         "Search generated reports.",
	}, func(ctx echo.Context) error {

		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if user.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		generatedReports, err := c.core.GeneratedReportManager.PaginationWithFields(context, ctx, &core.GeneratedReport{
			BranchID:       *user.BranchID,
			OrganizationID: user.OrganizationID,
		})
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "No generated reports found for the current branch"})
		}
		return ctx.JSON(http.StatusOK, generatedReports)
	})

	// GET /api/v1/generated-report/me/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/generated-report/me/search",
		Method:       "GET",
		ResponseType: core.GeneratedReportResponse{},
		Note:         "Search generated reports by current user logged in.",
	}, func(ctx echo.Context) error {

		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if user.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		generatedReports, err := c.core.GeneratedReportManager.PaginationWithFields(context, ctx, &core.GeneratedReport{
			BranchID:       *user.BranchID,
			OrganizationID: user.OrganizationID,
			CreatedByID:    user.UserID,
		})
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "No generated reports found for the current branch"})
		}
		return ctx.JSON(http.StatusOK, generatedReports)
	})

	// GET /api/v1/generated-report/pdf/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/generated-report/pdf/search",
		Method:       "GET",
		ResponseType: core.GeneratedReportResponse{},
		Note:         "Search generated PDF reports.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if user.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		generatedReports, err := c.core.GeneratedReportManager.PaginationWithFields(context, ctx, &core.GeneratedReport{
			BranchID:            *user.BranchID,
			OrganizationID:      user.OrganizationID,
			GeneratedReportType: core.GeneratedReportTypePDF,
		})
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "No generated reports found for the current branch"})
		}
		return ctx.JSON(http.StatusOK, generatedReports)
	})

	// GET /api/v1/generated-report/me/pdf/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/generated-report/me/pdf/search",
		Method:       "GET",
		ResponseType: core.GeneratedReportResponse{},
		Note:         "Search generated PDF reports by current user logged in.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if user.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		generatedReports, err := c.core.GeneratedReportManager.PaginationWithFields(context, ctx, &core.GeneratedReport{
			BranchID:            *user.BranchID,
			OrganizationID:      user.OrganizationID,
			CreatedByID:         user.UserID,
			GeneratedReportType: core.GeneratedReportTypePDF,
		})
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "No generated reports found for the current branch"})
		}
		return ctx.JSON(http.StatusOK, generatedReports)
	})

	// GET /api/v1/generated-report/excel/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/generated-report/excel/search",
		Method:       "GET",
		ResponseType: core.GeneratedReportResponse{},
		Note:         "Search generated Excel reports.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if user.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		generatedReports, err := c.core.GeneratedReportManager.PaginationWithFields(context, ctx, &core.GeneratedReport{
			BranchID:            *user.BranchID,
			OrganizationID:      user.OrganizationID,
			GeneratedReportType: core.GeneratedReportTypeExcel,
		})
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "No generated reports found for the current branch"})
		}
		return ctx.JSON(http.StatusOK, generatedReports)
	})
	// GET /api/v1/generated-report/me/excel/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/generated-report/me/excel/search",
		Method:       "GET",
		ResponseType: core.GeneratedReportResponse{},
		Note:         "Search generated Excel reports by current user logged in.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if user.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		generatedReports, err := c.core.GeneratedReportManager.PaginationWithFields(context, ctx, &core.GeneratedReport{
			BranchID:            *user.BranchID,
			OrganizationID:      user.OrganizationID,
			CreatedByID:         user.UserID,
			GeneratedReportType: core.GeneratedReportTypeExcel,
		})
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "No generated reports found for the current branch"})
		}
		return ctx.JSON(http.StatusOK, generatedReports)
	})

	// GET /api/v1/generated-report/favorites/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/generated-report/favorites/search",
		Method:       "GET",
		ResponseType: core.GeneratedReportResponse{},
		Note:         "Search favorite generated reports.",
	}, func(ctx echo.Context) error {

		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if user.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		generatedReports, err := c.core.GeneratedReportManager.PaginationWithFields(context, ctx, &core.GeneratedReport{
			BranchID:       *user.BranchID,
			OrganizationID: user.OrganizationID,
			IsFavorite:     true,
		})
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "No generated reports found for the current branch"})
		}
		return ctx.JSON(http.StatusOK, generatedReports)
	})

	// GET /api/v1/generated-report/me/favorites/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/generated-report/me/favorites/search",
		Method:       "GET",
		ResponseType: core.GeneratedReportResponse{},
		Note:         "Search favorite generated reports by current user logged in.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if user.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		generatedReports, err := c.core.GeneratedReportManager.PaginationWithFields(context, ctx, &core.GeneratedReport{
			BranchID:       *user.BranchID,
			OrganizationID: user.OrganizationID,
			CreatedByID:    user.UserID,
			IsFavorite:     true,
		})
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "No generated reports found for the current branch"})
		}
		return ctx.JSON(http.StatusOK, generatedReports)
	})

	// GET /api/v1/generated-report/available-models
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/generated-report/available-models",
		Method:       "GET",
		ResponseType: core.GeneratedReportAvailableModelsResponse{},
		Note:         "Get available generated report models with their counts for the current branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if user.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		models, err := c.core.GeneratedReportAvailableModels(context, user.OrganizationID, *user.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve available generated report models"})
		}
		return ctx.JSON(http.StatusOK, models)
	})

}
