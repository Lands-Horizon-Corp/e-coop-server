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
		Route:    "/timehseet/current",
		Method:   "GET",
		Response: "Ttimesheet",
		Note:     "use to identify if time in and time out action",
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
		Route:    "/timehseet/time-in-and-out",
		Method:   "POST",
		Request:  "{media_id: string}",
		Response: "Ttimesheet",
		Note:     "Current user time in and out",
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
				return echo.NewHTTPError(http.StatusInternalServerError, "failed to update timesheet: "+err.Error())
			}
			return ctx.JSON(http.StatusOK, c.model.TimesheetManager.ToModel(timesheet))

		}
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/timehseet/:timeSheet_id",
		Method:   "GET",
		Response: "TTimesheet",
		Note:     "get specifc timesgeet for view",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		timesheetId, err := horizon.EngineUUIDParam(ctx, "timeSheet_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid timehseet ID")
		}
		if err := c.model.TimesheetManager.DeleteByID(context, *timesheetId); err != nil {
			return c.InternalServerError(ctx, err)
		}
		return ctx.NoContent(http.StatusNoContent)
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/timehseet/branch/:branch_id",
		Method:   "GET",
		Response: "Ttimesheet",
		Note:     "get all timesheet of users on specific branch",
	}, func(ctx echo.Context) error {
		return nil
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/timehseet/me",
		Method:   "GET",
		Response: "Ttimesheet[]",
		Note:     "list of the users timesheet",
	}, func(ctx echo.Context) error {
		return nil
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/timehseet/user/:user-id",
		Method:   "GET",
		Response: "Ttimesheet[]",
		Note:     "list of timesheet of specific user",
	}, func(ctx echo.Context) error {
		return nil
	})

}
