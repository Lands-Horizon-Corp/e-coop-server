package controller_v1

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/lands-horizon/horizon-server/services/handlers"
	"github.com/lands-horizon/horizon-server/src/event"
	"github.com/lands-horizon/horizon-server/src/model"
)

func (c *Controller) TimesheetController() {
	req := c.provider.Service.Request

	// Returns the current timesheet entry for the user, if any (for time in/out determination)
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/timesheet/current",
		Method:       "GET",
		ResponseType: model.TimesheetResponse{},
		Note:         "Returns the current timesheet entry (not timed out yet) for the user, if any.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		timesheet, _ := c.model.TimesheetManager.FindOne(context, &model.Timesheet{
			UserID:         user.UserID,
			BranchID:       *user.BranchID,
			OrganizationID: user.OrganizationID,
		})
		if timesheet == nil || timesheet.TimeOut != nil {
			return ctx.NoContent(http.StatusNoContent)
		}
		return ctx.JSON(http.StatusOK, c.model.TimesheetManager.ToModel(timesheet))
	})

	// Records a time in or time out for the user.
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/timesheet/time-in-and-out",
		Method:       "POST",
		RequestType:  model.TimesheetRequest{},
		ResponseType: model.TimesheetResponse{},
		Note:         "Records a time-in or time-out for the current user depending on the last timesheet entry.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()

		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Time-in/out failed: user org error: " + err.Error(),
				Module:      "Timesheet",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}

		req, err := c.model.TimesheetManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Time-in/out failed: validation error: " + err.Error(),
				Module:      "Timesheet",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}

		timesheet, _ := c.model.TimesheetManager.FindOne(context, &model.Timesheet{
			UserID:         user.UserID,
			BranchID:       *user.BranchID,
			OrganizationID: user.OrganizationID,
		})

		now := time.Now().UTC()

		if timesheet == nil || timesheet.TimeOut != nil {
			newTimesheet := &model.Timesheet{
				CreatedAt:      now,
				CreatedByID:    user.UserID,
				UpdatedAt:      now,
				UpdatedByID:    user.UserID,
				BranchID:       *user.BranchID,
				OrganizationID: user.OrganizationID,
				TimeIn:         now,
				MediaInID:      req.MediaID,
				UserID:         user.UserID,
			}

			if err := c.model.TimesheetManager.Create(context, newTimesheet); err != nil {
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "create-error",
					Description: "Time-in failed: create error: " + err.Error(),
					Module:      "Timesheet",
				})
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create timesheet: " + err.Error()})
			}
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-success",
				Description: "Time-in: new timesheet created for user " + user.UserID.String(),
				Module:      "Timesheet",
			})
			return ctx.JSON(http.StatusOK, c.model.TimesheetManager.ToModel(newTimesheet))
		}

		timesheet.MediaOutID = req.MediaID
		timesheet.TimeOut = &now
		timesheet.UpdatedAt = now

		if err := c.model.TimesheetManager.UpdateFields(context, timesheet.ID, timesheet); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Time-out failed: update error: " + err.Error(),
				Module:      "Timesheet",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update timesheet: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Time-out: timesheet updated for user " + user.UserID.String(),
			Module:      "Timesheet",
		})
		return ctx.JSON(http.StatusOK, c.model.TimesheetManager.ToModel(timesheet))
	})

	// Get a specific timesheet by its ID
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/timesheet/:timesheet_id",
		Method:       "GET",
		ResponseType: model.TimesheetResponse{},
		Note:         "Returns the specific timesheet entry by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		timesheetID, err := handlers.EngineUUIDParam(ctx, "timesheet_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid timesheet_id: " + err.Error()})
		}
		timesheet, err := c.model.TimesheetManager.GetByIDRaw(context, *timesheetID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve timesheet: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, timesheet)
	})

	// Get all timesheets for users/employees on current branch
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/timesheet",
		Method:       "GET",
		ResponseType: model.TimesheetResponse{},
		Note:         "Returns all timesheets of users/employees for the current branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		timesheets, err := c.model.TimesheetCurrentBranch(context, user.OrganizationID, *user.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve timesheets: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.TimesheetManager.Filtered(context, ctx, timesheets))
	})

	// Get paginated timesheets for current branch
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/timesheet/search",
		Method:       "GET",
		ResponseType: model.TimesheetResponse{},
		Note:         "Returns paginated timesheets for the current branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		value, err := c.model.TimesheetCurrentBranch(context, user.OrganizationID, *user.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve timesheets for pagination: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.TimesheetManager.Pagination(context, ctx, value))
	})

	// Get the user's own timesheets in the current branch
	req.RegisterRoute(handlers.Route{
		Route:       "/timesheet/me",
		Method:      "GET",
		Note:        "Returns timesheets of the current user for the current branch.",
		RequestType: model.TimesheetRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		timesheets, err := c.model.GetUserTimesheet(context, user.UserID, user.OrganizationID, *user.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve user timesheets: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.TimesheetManager.Filtered(context, ctx, timesheets))
	})

	// Get paginated list of the user's own timesheets in the current branch
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/timesheet/me/search",
		Method:       "GET",
		ResponseType: model.TimesheetResponse{},
		Note:         "Returns paginated timesheets of the current user for the current branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		value, err := c.model.GetUserTimesheet(context, user.UserID, user.OrganizationID, *user.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve user timesheets for pagination: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.TimesheetManager.Pagination(context, ctx, value))
	})

	// List all timesheets of a specific user in the current branch
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/timesheet/user/:user_id",
		Method:       "GET",
		ResponseType: model.TimesheetResponse{},
		Note:         "Returns all timesheets of the specified user for the current branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userID, err := handlers.EngineUUIDParam(ctx, "user_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid user_id: " + err.Error()})
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		timesheets, err := c.model.GetUserTimesheet(context, *userID, user.OrganizationID, *user.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve user timesheets: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.TimesheetManager.Filtered(context, ctx, timesheets))
	})

	// Paginated timesheets of a specific user in the current branch
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/timesheet/user/:user_id/search",
		Method:       "GET",
		ResponseType: model.TimesheetResponse{},
		Note:         "Returns paginated timesheets of the specified user for the current branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userID, err := handlers.EngineUUIDParam(ctx, "user_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid user_id: " + err.Error()})
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		value, err := c.model.GetUserTimesheet(context, *userID, user.OrganizationID, *user.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve user timesheets for pagination: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.TimesheetManager.Pagination(context, ctx, value))
	})

	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/timesheet/employee/:user_organization_id/search",
		Method:       "GET",
		ResponseType: model.TimesheetResponse{},
		Note:         "Returns paginated timesheets of the specified employeee for the current branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrgId, err := handlers.EngineUUIDParam(ctx, "user_organization_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid user_id: " + err.Error()})
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		userOrganization, err := c.model.UserOrganizationManager.GetByID(context, *userOrgId)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "User organization not found: " + err.Error()})
		}

		value, err := c.model.GetUserTimesheet(context, userOrganization.UserID, user.OrganizationID, *user.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve user timesheets for pagination: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.TimesheetManager.Pagination(context, ctx, value))
	})

	// Get currently timed-in users for the current branch
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/timesheet/current/users",
		Method:       "GET",
		ResponseType: model.TimesheetResponse{},
		Note:         "Returns all currently timed-in users for the current branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		timesheets, err := c.model.TimeSheetActiveUsers(context, user.OrganizationID, *user.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve current timesheets: " + err.Error()})
		}

		return ctx.JSON(http.StatusOK, c.model.TimesheetManager.Filtered(context, ctx, timesheets))
	})
}
