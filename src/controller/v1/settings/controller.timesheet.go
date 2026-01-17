package settings

import (
	"net/http"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/helpers"
	"github.com/Lands-Horizon-Corp/e-coop-server/horizon"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/event"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/types"
	"github.com/labstack/echo/v4"
)

func TimesheetController(service *horizon.HorizonService) {
	req := service.API

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/timesheet/current",
		Method:       "GET",
		ResponseType: types.TimesheetResponse{},
		Note:         "Returns the current timesheet entry (not timed out yet) for the user, if any.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		timesheet, _ := core.TimesheetManager(service).FindOne(context, &types.Timesheet{
			UserID:         userOrg.UserID,
			BranchID:       *userOrg.BranchID,
			OrganizationID: userOrg.OrganizationID,
		})
		if timesheet == nil || timesheet.TimeOut != nil {
			return ctx.NoContent(http.StatusNoContent)
		}
		return ctx.JSON(http.StatusOK, core.TimesheetManager(service).ToModel(timesheet))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/timesheet/time-in-and-out",
		Method:       "POST",
		RequestType:  types.TimesheetRequest{},
		ResponseType: types.TimesheetResponse{},
		Note:         "Records a time-in or time-out for the current user depending on the last timesheet entry.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()

		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Time-in/out failed: user org error: " + err.Error(),
				Module:      "Timesheet",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}

		req, err := core.TimesheetManager(service).Validate(ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Time-in/out failed: validation error: " + err.Error(),
				Module:      "Timesheet",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}

		timesheet, _ := core.TimesheetManager(service).FindOne(context, &types.Timesheet{
			UserID:         userOrg.UserID,
			BranchID:       *userOrg.BranchID,
			OrganizationID: userOrg.OrganizationID,
		})

		now := time.Now().UTC()

		if timesheet == nil || timesheet.TimeOut != nil {
			newTimesheet := &types.Timesheet{
				CreatedAt:      now,
				CreatedByID:    userOrg.UserID,
				UpdatedAt:      now,
				UpdatedByID:    userOrg.UserID,
				BranchID:       *userOrg.BranchID,
				OrganizationID: userOrg.OrganizationID,
				TimeIn:         now,
				MediaInID:      req.MediaID,
				UserID:         userOrg.UserID,
			}

			if err := core.TimesheetManager(service).Create(context, newTimesheet); err != nil {
				event.Footstep(ctx, service, event.FootstepEvent{
					Activity:    "create-error",
					Description: "Time-in failed: create error: " + err.Error(),
					Module:      "Timesheet",
				})
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create timesheet: " + err.Error()})
			}
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-success",
				Description: "Time-in: new timesheet created for user " + userOrg.UserID.String(),
				Module:      "Timesheet",
			})
			return ctx.JSON(http.StatusOK, core.TimesheetManager(service).ToModel(newTimesheet))
		}

		timesheet.MediaOutID = req.MediaID
		timesheet.TimeOut = &now
		timesheet.UpdatedAt = now

		if err := core.TimesheetManager(service).UpdateByID(context, timesheet.ID, timesheet); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Time-out failed: update error: " + err.Error(),
				Module:      "Timesheet",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update timesheet: " + err.Error()})
		}
		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Time-out: timesheet updated for user " + userOrg.UserID.String(),
			Module:      "Timesheet",
		})
		return ctx.JSON(http.StatusOK, core.TimesheetManager(service).ToModel(timesheet))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/timesheet/:timesheet_id",
		Method:       "GET",
		ResponseType: types.TimesheetResponse{},
		Note:         "Returns the specific timesheet entry by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		timesheetID, err := helpers.EngineUUIDParam(ctx, "timesheet_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid timesheet_id: " + err.Error()})
		}
		timesheet, err := core.TimesheetManager(service).GetByIDRaw(context, *timesheetID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve timesheet: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, timesheet)
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/timesheet",
		Method:       "GET",
		ResponseType: types.TimesheetResponse{},
		Note:         "Returns all timesheets of users/employees for the current branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		timesheets, err := core.TimesheetCurrentBranch(context, service, userOrg.OrganizationID, *userOrg.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve timesheets: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, core.TimesheetManager(service).ToModels(timesheets))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/timesheet/search",
		Method:       "GET",
		ResponseType: types.TimesheetResponse{},
		Note:         "Returns paginated timesheets for the current branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		value, err := core.TimesheetManager(service).NormalPagination(context, ctx, &types.Timesheet{
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve timesheets for pagination: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, value)
	})

	req.RegisterWebRoute(horizon.Route{
		Route:       "/api/v1/timesheet/me",
		Method:      "GET",
		Note:        "Returns timesheets of the current user for the current branch.",
		RequestType: types.TimesheetRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		timesheets, err := core.GetUserTimesheet(context, service, userOrg.UserID, userOrg.OrganizationID, *userOrg.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve user timesheets: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, core.TimesheetManager(service).ToModels(timesheets))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/timesheet/me/search",
		Method:       "GET",
		ResponseType: types.TimesheetResponse{},
		Note:         "Returns paginated timesheets of the current user for the current branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		value, err := core.TimesheetManager(service).NormalPagination(context, ctx, &types.Timesheet{
			UserID:         userOrg.UserID,
			BranchID:       *userOrg.BranchID,
			OrganizationID: userOrg.OrganizationID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve user timesheets for pagination: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, value)
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/timesheet/user/:user_id",
		Method:       "GET",
		ResponseType: types.TimesheetResponse{},
		Note:         "Returns all timesheets of the specified user for the current branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userID, err := helpers.EngineUUIDParam(ctx, "user_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid user_id: " + err.Error()})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		timesheets, err := core.GetUserTimesheet(context, service, *userID, userOrg.OrganizationID, *userOrg.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve user timesheets: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, core.TimesheetManager(service).ToModels(timesheets))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/timesheet/user/:user_id/search",
		Method:       "GET",
		ResponseType: types.TimesheetResponse{},
		Note:         "Returns paginated timesheets of the specified user for the current branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userID, err := helpers.EngineUUIDParam(ctx, "user_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid user_id: " + err.Error()})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		value, err := core.TimesheetManager(service).NormalPagination(context, ctx, &types.Timesheet{
			UserID:         *userID,
			BranchID:       *userOrg.BranchID,
			OrganizationID: userOrg.OrganizationID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve user timesheets for pagination: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, value)
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/timesheet/employee/:user_organization_id/search",
		Method:       "GET",
		ResponseType: types.TimesheetResponse{},
		Note:         "Returns paginated timesheets of the specified employeee for the current branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrgID, err := helpers.EngineUUIDParam(ctx, "user_organization_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid user_id: " + err.Error()})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		userOrganization, err := core.UserOrganizationManager(service).GetByID(context, *userOrgID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "User organization not found: " + err.Error()})
		}

		value, err := core.TimesheetManager(service).NormalPagination(context, ctx, &types.Timesheet{
			UserID:         userOrganization.UserID,
			BranchID:       *userOrg.BranchID,
			OrganizationID: userOrg.OrganizationID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve user timesheets for pagination: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, value)
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/timesheet/current/users",
		Method:       "GET",
		ResponseType: types.TimesheetResponse{},
		Note:         "Returns all currently timed-in users for the current branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		timesheets, err := core.TimeSheetActiveUsers(context, service, userOrg.OrganizationID, *userOrg.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve current timesheets: " + err.Error()})
		}

		return ctx.JSON(http.StatusOK, core.TimesheetManager(service).ToModels(timesheets))
	})
}
