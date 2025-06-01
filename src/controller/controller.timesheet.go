package controller

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/lands-horizon/horizon-server/services/horizon"
	"github.com/lands-horizon/horizon-server/src/model"
)

func (c *Controller) TimesheetController() {
	req := c.provider.Service.Request

	req.RegisterRoute(horizon.Route{
		Route:    "/timesheet/current",
		Method:   "GET",
		Response: "Ttimesheet",
		Note:     "Use to identify current time in and time out action",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.NoContent(http.StatusNoContent)
		}
		timesheet, _ := c.model.TimesheetManager.FindOne(context, &model.Timesheet{
			UserID:         user.UserID,
			BranchID:       *user.BranchID,
			OrganizationID: user.OrganizationID,
		})
		if timesheet == nil {
			return ctx.NoContent(http.StatusNoContent)
		}
		return ctx.JSON(http.StatusOK, c.model.TimesheetManager.ToModel(timesheet))
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/timesheet/time-in-and-out",
		Method:   "POST",
		Request:  "{media_id: string}",
		Response: "Ttimesheet",
		Note:     "Record current user time in and out",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.NoContent(http.StatusNoContent)
		}
		req, err := c.model.TimesheetManager.Validate(ctx)
		if err != nil {
			return c.BadRequest(ctx, err.Error())
		}
		timesheet, _ := c.model.TimesheetManager.FindOne(context, &model.Timesheet{
			UserID: user.UserID,
		})
		now := time.Now().UTC()
		if timesheet == nil {
			model := &model.Timesheet{
				CreatedAt:      now,
				CreatedByID:    user.UserID,
				UpdatedAt:      now,
				UpdatedByID:    user.UserID,
				BranchID:       *user.BranchID,
				OrganizationID: user.OrganizationID,
				TimeIn:         now,
				MediaInID:      req.MediaID,
			}
			if err := c.model.TimesheetManager.Create(context, model); err != nil {
				return c.InternalServerError(ctx, err)
			}
			return ctx.JSON(http.StatusOK, c.model.TimesheetManager.ToModel(model))
		} else {
			timesheet.MediaOutID = req.MediaID
			timesheet.TimeOut = &now
			timesheet.UpdatedAt = now
			if err := c.model.TimesheetManager.UpdateByID(context, timesheet.ID, timesheet); err != nil {
				return echo.NewHTTPError(http.StatusInternalServerError, "Failed to update timesheet: "+err.Error())
			}
			return ctx.JSON(http.StatusOK, c.model.TimesheetManager.ToModel(timesheet))
		}
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/timesheet/:timesheet_id",
		Method:   "GET",
		Response: "TTimesheet",
		Note:     "Get specific timesheet for viewing",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		timesheetID, err := horizon.EngineUUIDParam(ctx, "timesheet_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid timesheet ID")
		}
		timesheet, err := c.model.TimesheetManager.GetByIDRaw(context, *timesheetID)
		if err != nil {
			return c.InternalServerError(ctx, err)
		}
		return ctx.JSON(http.StatusOK, timesheet)
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/timesheet",
		Method:   "GET",
		Response: "Ttimesheet",
		Note:     "Get all timesheets of users/employees on current branch",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.NoContent(http.StatusNoContent)
		}
		timesheets, err := c.model.GetAllUserTimesheet(context, user.OrganizationID, *user.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.TimesheetManager.ToModels(timesheets))
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/timesheet/me",
		Method:   "GET",
		Response: "Ttimesheet[]",
		Note:     "List of the user's timesheets on current branch",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.NoContent(http.StatusNoContent)
		}
		timesheets, err := c.model.GetUserTimesheet(context, user.UserID, user.OrganizationID, *user.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.TimesheetManager.ToModels(timesheets))
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/timesheet/user/:user_id",
		Method:   "GET",
		Response: "Ttimesheet[]",
		Note:     "List of timesheets of specific user",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userID, err := horizon.EngineUUIDParam(ctx, "user_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid user ID")
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.NoContent(http.StatusNoContent)
		}
		timesheets, err := c.model.GetUserTimesheet(context, *userID, user.OrganizationID, *user.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.TimesheetManager.ToModels(timesheets))
	})
}
