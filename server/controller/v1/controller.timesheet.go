package v1

import (
	"net/http"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/server/event"
	"github.com/Lands-Horizon-Corp/e-coop-server/server/model/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/services/handlers"
	"github.com/labstack/echo/v4"
)

func (c *Controller) timesheetController() {
	req := c.provider.Service.Request

	// Returns the current timesheet entry for the user, if any (for time in/out determination)
	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/timesheet/current",
		Method:       "GET",
		ResponseType: core.TimesheetResponse{},
		Note:         "Returns the current timesheet entry (not timed out yet) for the user, if any.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		timesheet, _ := c.core.TimesheetManager.FindOne(context, &core.Timesheet{
			UserID:         userOrg.UserID,
			BranchID:       *userOrg.BranchID,
			OrganizationID: userOrg.OrganizationID,
		})
		if timesheet == nil || timesheet.TimeOut != nil {
			return ctx.NoContent(http.StatusNoContent)
		}
		return ctx.JSON(http.StatusOK, c.core.TimesheetManager.ToModel(timesheet))
	})

	// Records a time in or time out for the user.
	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/timesheet/time-in-and-out",
		Method:       "POST",
		RequestType:  core.TimesheetRequest{},
		ResponseType: core.TimesheetResponse{},
		Note:         "Records a time-in or time-out for the current user depending on the last timesheet entry.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()

		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Time-in/out failed: user org error: " + err.Error(),
				Module:      "Timesheet",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}

		req, err := c.core.TimesheetManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Time-in/out failed: validation error: " + err.Error(),
				Module:      "Timesheet",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}

		timesheet, _ := c.core.TimesheetManager.FindOne(context, &core.Timesheet{
			UserID:         userOrg.UserID,
			BranchID:       *userOrg.BranchID,
			OrganizationID: userOrg.OrganizationID,
		})

		now := time.Now().UTC()

		if timesheet == nil || timesheet.TimeOut != nil {
			newTimesheet := &core.Timesheet{
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

			if err := c.core.TimesheetManager.Create(context, newTimesheet); err != nil {
				c.event.Footstep(ctx, event.FootstepEvent{
					Activity:    "create-error",
					Description: "Time-in failed: create error: " + err.Error(),
					Module:      "Timesheet",
				})
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create timesheet: " + err.Error()})
			}
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-success",
				Description: "Time-in: new timesheet created for user " + userOrg.UserID.String(),
				Module:      "Timesheet",
			})
			return ctx.JSON(http.StatusOK, c.core.TimesheetManager.ToModel(newTimesheet))
		}

		timesheet.MediaOutID = req.MediaID
		timesheet.TimeOut = &now
		timesheet.UpdatedAt = now

		if err := c.core.TimesheetManager.UpdateByID(context, timesheet.ID, timesheet); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Time-out failed: update error: " + err.Error(),
				Module:      "Timesheet",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update timesheet: " + err.Error()})
		}
		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Time-out: timesheet updated for user " + userOrg.UserID.String(),
			Module:      "Timesheet",
		})
		return ctx.JSON(http.StatusOK, c.core.TimesheetManager.ToModel(timesheet))
	})

	// Get a specific timesheet by its ID
	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/timesheet/:timesheet_id",
		Method:       "GET",
		ResponseType: core.TimesheetResponse{},
		Note:         "Returns the specific timesheet entry by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		timesheetID, err := handlers.EngineUUIDParam(ctx, "timesheet_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid timesheet_id: " + err.Error()})
		}
		timesheet, err := c.core.TimesheetManager.GetByIDRaw(context, *timesheetID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve timesheet: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, timesheet)
	})

	// Get all timesheets for users/employees on current branch
	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/timesheet",
		Method:       "GET",
		ResponseType: core.TimesheetResponse{},
		Note:         "Returns all timesheets of users/employees for the current branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		timesheets, err := c.core.TimesheetCurrentBranch(context, userOrg.OrganizationID, *userOrg.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve timesheets: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.core.TimesheetManager.ToModels(timesheets))
	})

	// Get paginated timesheets for current branch
	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/timesheet/search",
		Method:       "GET",
		ResponseType: core.TimesheetResponse{},
		Note:         "Returns paginated timesheets for the current branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		value, err := c.core.TimesheetManager.NormalPagination(context, ctx, &core.Timesheet{
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve timesheets for pagination: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, value)
	})

	// Get the user's own timesheets in the current branch
	req.RegisterWebRoute(handlers.Route{
		Route:       "/api/v1/timesheet/me",
		Method:      "GET",
		Note:        "Returns timesheets of the current user for the current branch.",
		RequestType: core.TimesheetRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		timesheets, err := c.core.GetUserTimesheet(context, userOrg.UserID, userOrg.OrganizationID, *userOrg.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve user timesheets: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.core.TimesheetManager.ToModels(timesheets))
	})

	// Get paginated list of the user's own timesheets in the current branch
	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/timesheet/me/search",
		Method:       "GET",
		ResponseType: core.TimesheetResponse{},
		Note:         "Returns paginated timesheets of the current user for the current branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		value, err := c.core.TimesheetManager.NormalPagination(context, ctx, &core.Timesheet{
			UserID:         userOrg.UserID,
			BranchID:       *userOrg.BranchID,
			OrganizationID: userOrg.OrganizationID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve user timesheets for pagination: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, value)
	})

	// List all timesheets of a specific user in the current branch
	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/timesheet/user/:user_id",
		Method:       "GET",
		ResponseType: core.TimesheetResponse{},
		Note:         "Returns all timesheets of the specified user for the current branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userID, err := handlers.EngineUUIDParam(ctx, "user_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid user_id: " + err.Error()})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		timesheets, err := c.core.GetUserTimesheet(context, *userID, userOrg.OrganizationID, *userOrg.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve user timesheets: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.core.TimesheetManager.ToModels(timesheets))
	})

	// Paginated timesheets of a specific user in the current branch
	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/timesheet/user/:user_id/search",
		Method:       "GET",
		ResponseType: core.TimesheetResponse{},
		Note:         "Returns paginated timesheets of the specified user for the current branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userID, err := handlers.EngineUUIDParam(ctx, "user_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid user_id: " + err.Error()})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		value, err := c.core.TimesheetManager.NormalPagination(context, ctx, &core.Timesheet{
			UserID:         *userID,
			BranchID:       *userOrg.BranchID,
			OrganizationID: userOrg.OrganizationID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve user timesheets for pagination: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, value)
	})

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/timesheet/employee/:user_organization_id/search",
		Method:       "GET",
		ResponseType: core.TimesheetResponse{},
		Note:         "Returns paginated timesheets of the specified employeee for the current branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrgID, err := handlers.EngineUUIDParam(ctx, "user_organization_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid user_id: " + err.Error()})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		userOrganization, err := c.core.UserOrganizationManager.GetByID(context, *userOrgID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "User organization not found: " + err.Error()})
		}

		value, err := c.core.TimesheetManager.NormalPagination(context, ctx, &core.Timesheet{
			UserID:         userOrganization.UserID,
			BranchID:       *userOrg.BranchID,
			OrganizationID: userOrg.OrganizationID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve user timesheets for pagination: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, value)
	})

	// Get currently timed-in users for the current branch
	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/timesheet/current/users",
		Method:       "GET",
		ResponseType: core.TimesheetResponse{},
		Note:         "Returns all currently timed-in users for the current branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		timesheets, err := c.core.TimeSheetActiveUsers(context, userOrg.OrganizationID, *userOrg.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve current timesheets: " + err.Error()})
		}

		return ctx.JSON(http.StatusOK, c.core.TimesheetManager.ToModels(timesheets))
	})
}
