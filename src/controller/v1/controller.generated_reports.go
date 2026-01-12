package v1

import (
	"net/http"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/helpers"
	"github.com/Lands-Horizon-Corp/e-coop-server/horizon"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/event"
	"github.com/labstack/echo/v4"
)

func generatedReports(service *horizon.HorizonService) {
	req := service.API

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/generated-report/:generated_report_id/download",
		Method:       "POST",
		ResponseType: core.Media{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		generatedReportID, err := helpers.EngineUUIDParam(ctx, "generated_report_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid generated report ID: " + err.Error()})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Generated report creation failed (/generated-report), user org error: " + err.Error(),
				Module:      "GeneratedReport",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		generatedReport, err := core.GeneratedReportManager(service).GetByID(context, *generatedReportID)
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
		if err := core.GeneratedReportsDownloadUsersManager(service).Create(context, generatedReportsDownloadUsers); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Generated reports download user creation failed (/generated-report/download), db error: " + err.Error(),
				Module:      "GeneratedReportsDownloadUsers",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create generated reports download user " + err.Error()})
		}

		media, err := core.MediaManager(service).GetByID(context, *generatedReport.MediaID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Media not found"})
		}
		return ctx.JSON(http.StatusOK, core.MediaManager(service).ToModel(media))
	})

	// req.RegisterWebRoute(horizon.Route{
	// 	Route:        "/api/v1/generated-report",
	// 	Method:       "POST",
	// 	RequestType:  core.GeneratedReportRequest{},
	// 	ResponseType: core.GeneratedReportResponse{},
	// 	Note:         "Create a new generated report.",
	// }, func(ctx echo.Context) error {
	// 	context := ctx.Request().Context()
	// 	req, err := core.GeneratedReportManager(service).Validate(ctx)
	// 	if err != nil {
	// 		event.Footstep(ctx, service, event.FootstepEvent{
	// 			Activity:    "create-error",
	// 			Description: "Generated report creation failed (/generated-report), validation error: " + err.Error(),
	// 			Module:      "GeneratedReport",
	// 		})
	// 		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid generated report data: " + err.Error()})
	// 	}
	// 	userOrg, err := event.CurrentUserOrganization(context, service, ctx)
	// 	if err != nil {
	// 		event.Footstep(ctx, service, event.FootstepEvent{
	// 			Activity:    "create-error",
	// 			Description: "Generated report creation failed (/generated-report), user org error: " + err.Error(),
	// 			Module:      "GeneratedReport",
	// 		})
	// 		return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
	// 	}
	// 	if userOrg.BranchID == nil {
	// 		event.Footstep(ctx, service, event.FootstepEvent{
	// 			Activity:    "create-error",
	// 			Description: "Generated report creation failed (/generated-report), user not assigned to branch.",
	// 			Module:      "GeneratedReport",
	// 		})
	// 		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
	// 	}
	// 	generatedReport := &core.GeneratedReport{
	// 		Name:                req.Name,
	// 		Description:         req.Description,
	// 		FilterSearch:        req.FilterSearch,
	// 		Model:               req.Model,
	// 		CreatedAt:           time.Now().UTC(),
	// 		CreatedByID:         userOrg.UserID,
	// 		UpdatedAt:           time.Now().UTC(),
	// 		UpdatedByID:         userOrg.UserID,
	// 		BranchID:            *userOrg.BranchID,
	// 		OrganizationID:      userOrg.OrganizationID,
	// 		Status:              core.GeneratedReportStatusPending,
	// 		GeneratedReportType: req.GeneratedReportType,
	// 		URL:                 req.URL,
	// 		UserID:              &userOrg.UserID,

	// 		Template:  req.Template,
	// 		PaperSize: req.PaperSize,
	// 		Width:     req.Width,
	// 		Height:    req.Height,
	// 		Unit:      req.Unit,
	// 		Landscape: req.Landscape,
	// 	}
	// 	data, err := GeneratedReportDownload(context, generatedReport)
	// 	if err != nil {
	// 		event.Footstep(ctx, service, event.FootstepEvent{
	// 			Activity:    "create-error",
	// 			Description: "Generated report creation failed (/generated-report), download error: " + err.Error(),
	// 			Module:      "GeneratedReport",
	// 		})
	// 		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create generated report: " + err.Error()})
	// 	}
	// 	return ctx.JSON(http.StatusCreated, core.GeneratedReportManager(service).ToModel(data))

	// })

	req.RegisterWebRoute(horizon.Route{
		Route:  "/api/v1/generated-report/:generated_report_id",
		Method: "DELETE",
		Note:   "Delete a generated report by ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		generatedReportID, err := helpers.EngineUUIDParam(ctx, "generated_report_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid generated report ID: " + err.Error()})
		}
		generatedReport, err := core.GeneratedReportManager(service).GetByID(context, *generatedReportID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Generated report not found"})
		}
		if generatedReport.MediaID != nil {
			if err := core.MediaDelete(context, service, *generatedReport.MediaID); err != nil {
				event.Footstep(ctx, service, event.FootstepEvent{
					Activity:    "delete-error",
					Description: "Media delete failed (/media/:media_id), db error: " + err.Error(),
					Module:      "Media",
				})
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete media record: " + err.Error()})
			}
		}
		if err := core.GeneratedReportManager(service).Delete(context, *generatedReportID); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete generated report: " + err.Error()})
		}
		return ctx.NoContent(http.StatusNoContent)
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/generated-report/:generated_report_id",
		Method:       "PUT",
		RequestType:  core.GeneratedReportUpdateRequest{},
		ResponseType: core.GeneratedReportResponse{},
		Note:         "Update an existing generated report.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()

		generatedReportID, err := helpers.EngineUUIDParam(ctx, "generated_report_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid generated report ID: " + err.Error()})
		}
		var req core.GeneratedReportUpdateRequest
		if err := ctx.Bind(&req); err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid generated report update payload: " + err.Error()})
		}
		if err := service.Validator.Struct(req); err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Generated report update failed (/generated-report/:generated_report_id), user org error: " + err.Error(),
				Module:      "GeneratedReport",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		generatedReport, err := core.GeneratedReportManager(service).GetByID(context, *generatedReportID)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
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

		if err := core.GeneratedReportManager(service).UpdateByID(context, generatedReport.ID, generatedReport); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Generated report update failed (/generated-report/:generated_report_id), db error: " + err.Error(),
				Module:      "GeneratedReport",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update generated report: " + err.Error()})
		}
		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Updated generated report (/generated-report/:generated_report_id): " + generatedReport.Name,
			Module:      "GeneratedReport",
		})
		return ctx.JSON(http.StatusOK, core.GeneratedReportManager(service).ToModel(generatedReport))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/generated-report/download-user",
		Method:       "POST",
		Note:         "Creates a new generated report download user entry for the current user's organization and branch.",
		RequestType:  core.GeneratedReportsDownloadUsersRequest{},
		ResponseType: core.GeneratedReportsDownloadUsersResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := core.GeneratedReportsDownloadUsersManager(service).Validate(ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Generated reports download user creation failed (/generated-report/download-user), validation error: " + err.Error(),
				Module:      "GeneratedReportsDownloadUsers",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid generated reports download user data: " + err.Error()})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Generated reports download user creation failed (/generated-report/download-user), user org error: " + err.Error(),
				Module:      "GeneratedReportsDownloadUsers",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if userOrg.BranchID == nil {
			event.Footstep(ctx, service, event.FootstepEvent{
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

		if err := core.GeneratedReportsDownloadUsersManager(service).Create(context, downloadUser); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Generated reports download user creation failed (/generated-report/download-user), db error: " + err.Error(),
				Module:      "GeneratedReportsDownloadUsers",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create generated reports download user: " + err.Error()})
		}
		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "create-success",
			Description: "Created generated reports download user (/generated-report/download-user)",
			Module:      "GeneratedReportsDownloadUsers",
		})
		return ctx.JSON(http.StatusCreated, core.GeneratedReportsDownloadUsersManager(service).ToModel(downloadUser))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/generated-report/:generated_report_id/favorite",
		Method:       "PUT",
		ResponseType: core.GeneratedReportResponse{},
		Note:         "Mark or unmark a generated report as favorite.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		generatedReportID, err := helpers.EngineUUIDParam(ctx, "generated_report_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid generated report ID"})
		}
		generatedReport, err := core.GeneratedReportManager(service).GetByID(context, *generatedReportID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Generated report not found"})
		}
		generatedReport.IsFavorite = !generatedReport.IsFavorite
		if err := core.GeneratedReportManager(service).UpdateByID(context, generatedReport.ID, generatedReport); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update favorite status: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, core.GeneratedReportManager(service).ToModel(generatedReport))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/generated-report/:generated_report_id",
		Method:       "GET",
		ResponseType: core.GeneratedReportResponse{},
		Note:         "Returns a specific generated report by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		generatedReportID, err := helpers.EngineUUIDParam(ctx, "generated_report_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid generated report ID"})
		}
		generatedReport, err := core.GeneratedReportManager(service).GetByID(context, *generatedReportID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Generated report not found"})
		}
		return ctx.JSON(http.StatusOK, core.GeneratedReportManager(service).ToModel(generatedReport))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/generated-report/search",
		Method:       "GET",
		ResponseType: core.GeneratedReportResponse{},
		Note:         "Search generated reports.",
	}, func(ctx echo.Context) error {

		context := ctx.Request().Context()
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if userOrg.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		generatedReports, err := core.GeneratedReportManager(service).NormalPagination(context, ctx, &core.GeneratedReport{
			BranchID:       *userOrg.BranchID,
			OrganizationID: userOrg.OrganizationID,
		})
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "No generated reports found for the current branch: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, generatedReports)
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/generated-report/me/search",
		Method:       "GET",
		ResponseType: core.GeneratedReportResponse{},
		Note:         "Search generated reports by current user logged in.",
	}, func(ctx echo.Context) error {

		context := ctx.Request().Context()
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if userOrg.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		generatedReports, err := core.GeneratedReportManager(service).NormalPagination(context, ctx, &core.GeneratedReport{
			BranchID:       *userOrg.BranchID,
			OrganizationID: userOrg.OrganizationID,
			CreatedByID:    userOrg.UserID,
		})
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "No generated reports found for the current branch: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, generatedReports)
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/generated-report/pdf/search",
		Method:       "GET",
		ResponseType: core.GeneratedReportResponse{},
		Note:         "Search generated PDF reports.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if userOrg.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		generatedReports, err := core.GeneratedReportManager(service).NormalPagination(context, ctx, &core.GeneratedReport{
			BranchID:            *userOrg.BranchID,
			OrganizationID:      userOrg.OrganizationID,
			GeneratedReportType: core.GeneratedReportTypePDF,
		})
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "No generated reports found for the current branch: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, generatedReports)
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/generated-report/me/pdf/search",
		Method:       "GET",
		ResponseType: core.GeneratedReportResponse{},
		Note:         "Search generated PDF reports by current user logged in.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if userOrg.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		generatedReports, err := core.GeneratedReportManager(service).NormalPagination(context, ctx, &core.GeneratedReport{
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

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/generated-report/excel/search",
		Method:       "GET",
		ResponseType: core.GeneratedReportResponse{},
		Note:         "Search generated Excel reports.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if userOrg.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		generatedReports, err := core.GeneratedReportManager(service).NormalPagination(context, ctx, &core.GeneratedReport{
			BranchID:            *userOrg.BranchID,
			OrganizationID:      userOrg.OrganizationID,
			GeneratedReportType: core.GeneratedReportTypeExcel,
		})
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "No generated reports found for the current branch: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, generatedReports)
	})
	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/generated-report/me/excel/search",
		Method:       "GET",
		ResponseType: core.GeneratedReportResponse{},
		Note:         "Search generated Excel reports by current user logged in.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if userOrg.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		generatedReports, err := core.GeneratedReportManager(service).NormalPagination(context, ctx, &core.GeneratedReport{
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

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/generated-report/favorites/search",
		Method:       "GET",
		ResponseType: core.GeneratedReportResponse{},
		Note:         "Search favorite generated reports.",
	}, func(ctx echo.Context) error {

		context := ctx.Request().Context()
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if userOrg.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		generatedReports, err := core.GeneratedReportManager(service).NormalPagination(context, ctx, &core.GeneratedReport{
			BranchID:       *userOrg.BranchID,
			OrganizationID: userOrg.OrganizationID,
			IsFavorite:     true,
		})
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "No generated reports found for the current branch: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, generatedReports)
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/generated-report/me/favorites/search",
		Method:       "GET",
		ResponseType: core.GeneratedReportResponse{},
		Note:         "Search favorite generated reports by current user logged in.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if userOrg.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		generatedReports, err := core.GeneratedReportManager(service).NormalPagination(context, ctx, &core.GeneratedReport{
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

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/generated-report/available-models",
		Method:       "GET",
		ResponseType: core.GeneratedReportAvailableModelsResponse{},
		Note:         "Get available generated report models with their counts for the current branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if userOrg.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		models, err := core.GeneratedReportAvailableModels(context, service, userOrg.OrganizationID, *userOrg.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve available generated report models: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, models)
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/generated-report/model/:model/search",
		Method:       "GET",
		ResponseType: core.GeneratedReportResponse{},
		Note:         "Search generated reports by model.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if userOrg.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		model := ctx.Param("model")
		generatedReports, err := core.GeneratedReportManager(service).NormalPagination(context, ctx, &core.GeneratedReport{
			BranchID:       *userOrg.BranchID,
			OrganizationID: userOrg.OrganizationID,
			Model:          model,
		})
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "No generated reports found for the current branch: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, generatedReports)
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/generated-report/me/model/:model/search",
		Method:       "GET",
		ResponseType: core.GeneratedReportResponse{},
		Note:         "Search generated reports by model for current user logged in.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if userOrg.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		model := ctx.Param("model")
		generatedReports, err := core.GeneratedReportManager(service).NormalPagination(context, ctx, &core.GeneratedReport{
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

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/generated-report/pdf/model/:model/search",
		Method:       "GET",
		ResponseType: core.GeneratedReportResponse{},
		Note:         "Search generated PDF reports by model.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if userOrg.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		model := ctx.Param("model")
		generatedReports, err := core.GeneratedReportManager(service).NormalPagination(context, ctx, &core.GeneratedReport{
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

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/generated-report/me/pdf/model/:model/search",
		Method:       "GET",
		ResponseType: core.GeneratedReportResponse{},
		Note:         "Search generated PDF reports by model for current user logged in.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if userOrg.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		model := ctx.Param("model")
		generatedReports, err := core.GeneratedReportManager(service).NormalPagination(context, ctx, &core.GeneratedReport{
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

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/generated-report/excel/model/:model/search",
		Method:       "GET",
		ResponseType: core.GeneratedReportResponse{},
		Note:         "Search generated Excel reports by model.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if userOrg.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		model := ctx.Param("model")
		generatedReports, err := core.GeneratedReportManager(service).NormalPagination(context, ctx, &core.GeneratedReport{
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

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/generated-report/me/excel/model/:model/search",
		Method:       "GET",
		ResponseType: core.GeneratedReportResponse{},
		Note:         "Search generated Excel reports by model for current user logged in.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed: " + err.Error()})
		}
		if userOrg.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		model := ctx.Param("model")
		generatedReports, err := core.GeneratedReportManager(service).NormalPagination(context, ctx, &core.GeneratedReport{
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

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/generated-report/favorites/model/:model/search",
		Method:       "GET",
		ResponseType: core.GeneratedReportResponse{},
		Note:         "Search favorite generated reports by model.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed: " + err.Error()})
		}
		if userOrg.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		model := ctx.Param("model")
		generatedReports, err := core.GeneratedReportManager(service).NormalPagination(context, ctx, &core.GeneratedReport{
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

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/generated-report/me/favorites/model/:model/search",
		Method:       "GET",
		ResponseType: core.GeneratedReportResponse{},
		Note:         "Search favorite generated reports by model for current user logged in.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed: " + err.Error()})
		}
		if userOrg.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		model := ctx.Param("model")
		generatedReports, err := core.GeneratedReportManager(service).NormalPagination(context, ctx, &core.GeneratedReport{
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
