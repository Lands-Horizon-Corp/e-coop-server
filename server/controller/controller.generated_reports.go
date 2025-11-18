package v1

import (
	"net/http"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/server/event"
	"github.com/Lands-Horizon-Corp/e-coop-server/server/model/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/services/handlers"
	"github.com/labstack/echo/v4"
)

// GeneratedReports manages endpoints for generated report resources.
func (c *Controller) generatedReports() {
	req := c.provider.Service.Request

	// GET /api/v1/generated-report/:generated_report_id/download
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/generated-report/:generated_report_id/download",
		Method:       "POST",
		ResponseType: core.Media{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		generatedReportID, err := handlers.EngineUUIDParam(ctx, "generated_report_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid generated report ID: " + err.Error()})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Generated report creation failed (/generated-report), user org error: " + err.Error(),
				Module:      "GeneratedReport",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		generatedReport, err := c.core.GeneratedReportManager.GetByID(context, *generatedReportID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Generated report not found"})
		}
		if generatedReport.MediaID == nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "No media associated with this generated report"})
		}
		generatedReportsDownloadUsers := &core.GeneratedReportsDownloadUsers{
			CreatedAt:          time.Now().UTC(),
			CreatedByID:        userOrg.UserID,
			UpdatedAt:          time.Now().UTC(),
			UpdatedByID:        userOrg.UserID,
			OrganizationID:     userOrg.OrganizationID,
			BranchID:           *userOrg.BranchID,
			UserID:             userOrg.UserID,
			UserOrganizationID: userOrg.ID,
			GeneratedReportID:  generatedReport.ID,
		}
		if err := c.core.GeneratedReportsDownloadUsersManager.Create(context, generatedReportsDownloadUsers); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Generated reports download user creation failed (/generated-report/download), db error: " + err.Error(),
				Module:      "GeneratedReportsDownloadUsers",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create generated reports download user " + err.Error()})
		}

		media, err := c.core.MediaManager.GetByID(context, *generatedReport.MediaID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Media not found"})
		}
		return ctx.JSON(http.StatusOK, c.core.MediaManager.ToModel(media))
	})

	// POST
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/generated-report",
		Method:       "POST",
		RequestType:  core.GeneratedReportRequest{},
		ResponseType: core.GeneratedReportResponse{},
		Note:         "Create a new generated report.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := c.core.GeneratedReportManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Generated report creation failed (/generated-report), validation error: " + err.Error(),
				Module:      "GeneratedReport",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid generated report data: " + err.Error()})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Generated report creation failed (/generated-report), user org error: " + err.Error(),
				Module:      "GeneratedReport",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if userOrg.BranchID == nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Generated report creation failed (/generated-report), user not assigned to branch.",
				Module:      "GeneratedReport",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		generatedReport := &core.GeneratedReport{
			Name:                req.Name,
			Description:         req.Description,
			FilterSearch:        req.FilterSearch,
			Model:               req.Model,
			CreatedAt:           time.Now().UTC(),
			CreatedByID:         userOrg.UserID,
			UpdatedAt:           time.Now().UTC(),
			UpdatedByID:         userOrg.UserID,
			BranchID:            *userOrg.BranchID,
			OrganizationID:      userOrg.OrganizationID,
			Status:              core.GeneratedReportStatusPending,
			GeneratedReportType: req.GeneratedReportType,
			URL:                 req.URL,
			Template:            req.Template,
			PaperSize:           req.PaperSize,
			UserID:              &userOrg.UserID,
		}
		data, err := c.event.GeneratedReportDownload(context, generatedReport)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Generated report creation failed (/generated-report), download error: " + err.Error(),
				Module:      "GeneratedReport",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create generated report: " + err.Error()})
		}
		return ctx.JSON(http.StatusCreated, c.core.GeneratedReportManager.ToModel(data))

	})

	req.RegisterRoute(handlers.Route{
		Route:  "/api/v1/generated-report/:generated_report_id",
		Method: "DELETE",
		Note:   "Delete a generated report by ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		generatedReportID, err := handlers.EngineUUIDParam(ctx, "generated_report_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid generated report ID: " + err.Error()})
		}
		generatedReport, err := c.core.GeneratedReportManager.GetByID(context, *generatedReportID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Generated report not found"})
		}
		// Only delete media if it exists
		if generatedReport.MediaID != nil {
			if err := c.core.MediaDelete(context, *generatedReport.MediaID); err != nil {
				c.event.Footstep(ctx, event.FootstepEvent{
					Activity:    "delete-error",
					Description: "Media delete failed (/media/:media_id), db error: " + err.Error(),
					Module:      "Media",
				})
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete media record: " + err.Error()})
			}
		}
		if err := c.core.GeneratedReportManager.Delete(context, *generatedReportID); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete generated report: " + err.Error()})
		}
		return ctx.NoContent(http.StatusNoContent)
	})

	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/generated-report/:generated_report_id",
		Method:       "PUT",
		RequestType:  core.GeneratedReportUpdateRequest{},
		ResponseType: core.GeneratedReportResponse{},
		Note:         "Update an existing generated report.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()

		generatedReportID, err := handlers.EngineUUIDParam(ctx, "generated_report_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid generated report ID: " + err.Error()})
		}
		var req core.GeneratedReportUpdateRequest
		if err := ctx.Bind(&req); err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid generated report update payload: " + err.Error()})
		}
		if err := c.provider.Service.Validator.Struct(req); err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Generated report update failed (/generated-report/:generated_report_id), user org error: " + err.Error(),
				Module:      "GeneratedReport",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		generatedReport, err := c.core.GeneratedReportManager.GetByID(context, *generatedReportID)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Generated report update failed (/generated-report/:generated_report_id), generated report not found.",
				Module:      "GeneratedReport",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "GeneratedReport not found"})
		}
		generatedReport.Name = req.Name
		generatedReport.Description = req.Description
		generatedReport.UpdatedAt = time.Now().UTC()
		generatedReport.UpdatedByID = userOrg.UserID

		if err := c.core.GeneratedReportManager.UpdateByID(context, generatedReport.ID, generatedReport); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Generated report update failed (/generated-report/:generated_report_id), db error: " + err.Error(),
				Module:      "GeneratedReport",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update generated report: " + err.Error()})
		}
		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Updated generated report (/generated-report/:generated_report_id): " + generatedReport.Name,
			Module:      "GeneratedReport",
		})
		return ctx.JSON(http.StatusOK, c.core.GeneratedReportManager.ToModel(generatedReport))
	})

	// POST /api/v1/generated-report/download-user
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/generated-report/download-user",
		Method:       "POST",
		Note:         "Creates a new generated report download user entry for the current user's organization and branch.",
		RequestType:  core.GeneratedReportsDownloadUsersRequest{},
		ResponseType: core.GeneratedReportsDownloadUsersResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := c.core.GeneratedReportsDownloadUsersManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Generated reports download user creation failed (/generated-report/download-user), validation error: " + err.Error(),
				Module:      "GeneratedReportsDownloadUsers",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid generated reports download user data: " + err.Error()})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Generated reports download user creation failed (/generated-report/download-user), user org error: " + err.Error(),
				Module:      "GeneratedReportsDownloadUsers",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if userOrg.BranchID == nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Generated reports download user creation failed (/generated-report/download-user), user not assigned to branch.",
				Module:      "GeneratedReportsDownloadUsers",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}

		downloadUser := &core.GeneratedReportsDownloadUsers{
			UserOrganizationID: req.UserOrganizationID,
			GeneratedReportID:  req.GeneratedReportID,
			CreatedAt:          time.Now().UTC(),
			CreatedByID:        userOrg.UserID,
			UpdatedAt:          time.Now().UTC(),
			UpdatedByID:        userOrg.UserID,
			BranchID:           *userOrg.BranchID,
			OrganizationID:     userOrg.OrganizationID,
		}

		if err := c.core.GeneratedReportsDownloadUsersManager.Create(context, downloadUser); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Generated reports download user creation failed (/generated-report/download-user), db error: " + err.Error(),
				Module:      "GeneratedReportsDownloadUsers",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create generated reports download user: " + err.Error()})
		}
		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "create-success",
			Description: "Created generated reports download user (/generated-report/download-user)",
			Module:      "GeneratedReportsDownloadUsers",
		})
		return ctx.JSON(http.StatusCreated, c.core.GeneratedReportsDownloadUsersManager.ToModel(downloadUser))
	})

	// PUT /api/v1/generated-report/:generated_report_id/favorite: Mark or unmark a generated report as favorite.
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/generated-report/:generated_report_id/favorite",
		Method:       "PUT",
		ResponseType: core.GeneratedReportResponse{},
		Note:         "Mark or unmark a generated report as favorite.",
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
		// Toggle the IsFavorite field
		generatedReport.IsFavorite = !generatedReport.IsFavorite
		if err := c.core.GeneratedReportManager.UpdateByID(context, generatedReport.ID, generatedReport); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update favorite status: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.core.GeneratedReportManager.ToModel(generatedReport))
	})

	// =======================================[FILTERED]==========================================================
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
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if userOrg.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		generatedReports, err := c.core.GeneratedReportManager.PaginationWithFields(context, ctx, &core.GeneratedReport{
			BranchID:       *userOrg.BranchID,
			OrganizationID: userOrg.OrganizationID,
		})
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "No generated reports found for the current branch: " + err.Error()})
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
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if userOrg.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		generatedReports, err := c.core.GeneratedReportManager.PaginationWithFields(context, ctx, &core.GeneratedReport{
			BranchID:       *userOrg.BranchID,
			OrganizationID: userOrg.OrganizationID,
			CreatedByID:    userOrg.UserID,
		})
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "No generated reports found for the current branch: " + err.Error()})
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
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if userOrg.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		generatedReports, err := c.core.GeneratedReportManager.PaginationWithFields(context, ctx, &core.GeneratedReport{
			BranchID:            *userOrg.BranchID,
			OrganizationID:      userOrg.OrganizationID,
			GeneratedReportType: core.GeneratedReportTypePDF,
		})
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "No generated reports found for the current branch: " + err.Error()})
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
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if userOrg.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		generatedReports, err := c.core.GeneratedReportManager.PaginationWithFields(context, ctx, &core.GeneratedReport{
			BranchID:            *userOrg.BranchID,
			OrganizationID:      userOrg.OrganizationID,
			CreatedByID:         userOrg.UserID,
			GeneratedReportType: core.GeneratedReportTypePDF,
		})
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "No generated reports found for the current branch: " + err.Error()})
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
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if userOrg.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		generatedReports, err := c.core.GeneratedReportManager.PaginationWithFields(context, ctx, &core.GeneratedReport{
			BranchID:            *userOrg.BranchID,
			OrganizationID:      userOrg.OrganizationID,
			GeneratedReportType: core.GeneratedReportTypeExcel,
		})
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "No generated reports found for the current branch: " + err.Error()})
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
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if userOrg.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		generatedReports, err := c.core.GeneratedReportManager.PaginationWithFields(context, ctx, &core.GeneratedReport{
			BranchID:            *userOrg.BranchID,
			OrganizationID:      userOrg.OrganizationID,
			CreatedByID:         userOrg.UserID,
			GeneratedReportType: core.GeneratedReportTypeExcel,
		})
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "No generated reports found for the current branch: " + err.Error()})
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
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if userOrg.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		generatedReports, err := c.core.GeneratedReportManager.PaginationWithFields(context, ctx, &core.GeneratedReport{
			BranchID:       *userOrg.BranchID,
			OrganizationID: userOrg.OrganizationID,
			IsFavorite:     true,
		})
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "No generated reports found for the current branch: " + err.Error()})
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
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if userOrg.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		generatedReports, err := c.core.GeneratedReportManager.PaginationWithFields(context, ctx, &core.GeneratedReport{
			BranchID:       *userOrg.BranchID,
			OrganizationID: userOrg.OrganizationID,
			CreatedByID:    userOrg.UserID,
			IsFavorite:     true,
		})
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "No generated reports found for the current branch:" + err.Error()})
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
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if userOrg.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		models, err := c.core.GeneratedReportAvailableModels(context, userOrg.OrganizationID, *userOrg.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve available generated report models: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, models)
	})

	// GET /api/v1/generated-report/model/:model/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/generated-report/model/:model/search",
		Method:       "GET",
		ResponseType: core.GeneratedReportResponse{},
		Note:         "Search generated reports by model.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if userOrg.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		model := ctx.Param("model")
		generatedReports, err := c.core.GeneratedReportManager.PaginationWithFields(context, ctx, &core.GeneratedReport{
			BranchID:       *userOrg.BranchID,
			OrganizationID: userOrg.OrganizationID,
			Model:          model,
		})
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "No generated reports found for the current branch: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, generatedReports)
	})

	// GET /api/v1/generated-report/me/model/:model/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/generated-report/me/model/:model/search",
		Method:       "GET",
		ResponseType: core.GeneratedReportResponse{},
		Note:         "Search generated reports by model for current user logged in.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if userOrg.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		model := ctx.Param("model")
		generatedReports, err := c.core.GeneratedReportManager.PaginationWithFields(context, ctx, &core.GeneratedReport{
			BranchID:       *userOrg.BranchID,
			OrganizationID: userOrg.OrganizationID,
			CreatedByID:    userOrg.UserID,
			Model:          model,
		})
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "No generated reports found for the current branch: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, generatedReports)
	})

	// GET /api/v1/generated-report/pdf/model/:model/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/generated-report/pdf/model/:model/search",
		Method:       "GET",
		ResponseType: core.GeneratedReportResponse{},
		Note:         "Search generated PDF reports by model.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if userOrg.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		model := ctx.Param("model")
		generatedReports, err := c.core.GeneratedReportManager.PaginationWithFields(context, ctx, &core.GeneratedReport{
			BranchID:            *userOrg.BranchID,
			OrganizationID:      userOrg.OrganizationID,
			GeneratedReportType: core.GeneratedReportTypePDF,
			Model:               model,
		})
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "No generated reports found for the current branch: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, generatedReports)
	})

	// GET /api/v1/generated-report/me/pdf/model/:model/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/generated-report/me/pdf/model/:model/search",
		Method:       "GET",
		ResponseType: core.GeneratedReportResponse{},
		Note:         "Search generated PDF reports by model for current user logged in.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if userOrg.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		model := ctx.Param("model")
		generatedReports, err := c.core.GeneratedReportManager.PaginationWithFields(context, ctx, &core.GeneratedReport{
			BranchID:            *userOrg.BranchID,
			OrganizationID:      userOrg.OrganizationID,
			CreatedByID:         userOrg.UserID,
			GeneratedReportType: core.GeneratedReportTypePDF,
			Model:               model,
		})
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "No generated reports found for the current branch: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, generatedReports)
	})

	// GET /api/v1/generated-report/excel/model/:model/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/generated-report/excel/model/:model/search",
		Method:       "GET",
		ResponseType: core.GeneratedReportResponse{},
		Note:         "Search generated Excel reports by model.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if userOrg.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		model := ctx.Param("model")
		generatedReports, err := c.core.GeneratedReportManager.PaginationWithFields(context, ctx, &core.GeneratedReport{
			BranchID:            *userOrg.BranchID,
			OrganizationID:      userOrg.OrganizationID,
			GeneratedReportType: core.GeneratedReportTypeExcel,
			Model:               model,
		})
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "No generated reports found for the current branch: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, generatedReports)
	})

	// GET /api/v1/generated-report/me/excel/model/:model/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/generated-report/me/excel/model/:model/search",
		Method:       "GET",
		ResponseType: core.GeneratedReportResponse{},
		Note:         "Search generated Excel reports by model for current user logged in.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed: " + err.Error()})
		}
		if userOrg.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		model := ctx.Param("model")
		generatedReports, err := c.core.GeneratedReportManager.PaginationWithFields(context, ctx, &core.GeneratedReport{
			BranchID:            *userOrg.BranchID,
			OrganizationID:      userOrg.OrganizationID,
			CreatedByID:         userOrg.UserID,
			GeneratedReportType: core.GeneratedReportTypeExcel,
			Model:               model,
		})
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "No generated reports found for the current branch: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, generatedReports)
	})

	// GET /api/v1/generated-report/favorites/model/:model/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/generated-report/favorites/model/:model/search",
		Method:       "GET",
		ResponseType: core.GeneratedReportResponse{},
		Note:         "Search favorite generated reports by model.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed: " + err.Error()})
		}
		if userOrg.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		model := ctx.Param("model")
		generatedReports, err := c.core.GeneratedReportManager.PaginationWithFields(context, ctx, &core.GeneratedReport{
			BranchID:       *userOrg.BranchID,
			OrganizationID: userOrg.OrganizationID,
			IsFavorite:     true,
			Model:          model,
		})
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "No generated reports found for the current branch: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, generatedReports)
	})

	// GET /api/v1/generated-report/me/favorites/model/:model/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/generated-report/me/favorites/model/:model/search",
		Method:       "GET",
		ResponseType: core.GeneratedReportResponse{},
		Note:         "Search favorite generated reports by model for current user logged in.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed: " + err.Error()})
		}
		if userOrg.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		model := ctx.Param("model")
		generatedReports, err := c.core.GeneratedReportManager.PaginationWithFields(context, ctx, &core.GeneratedReport{
			BranchID:       *userOrg.BranchID,
			OrganizationID: userOrg.OrganizationID,
			CreatedByID:    userOrg.UserID,
			IsFavorite:     true,
			Model:          model,
		})
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "No generated reports found for the current branch: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, generatedReports)
	})

}
