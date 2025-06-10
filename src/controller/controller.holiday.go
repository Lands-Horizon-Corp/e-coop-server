package controller

import (
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/lands-horizon/horizon-server/services/horizon"
	"github.com/lands-horizon/horizon-server/src/model"
)

func (c *Controller) HolidayController() {
	req := c.provider.Service.Request

	req.RegisterRoute(horizon.Route{
		Route:    "/holiday",
		Method:   "GET",
		Response: "THoliday[]",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.NoContent(http.StatusNoContent)
		}
		holiday, err := c.model.HolidayCurrentBranch(context, user.OrganizationID, *user.BranchID)
		if err != nil {
			return c.NotFound(ctx, "Holiday")
		}

		return ctx.JSON(http.StatusOK, c.model.HolidayManager.ToModels(holiday))
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/holiday/search",
		Method:   "GET",
		Request:  "Filter<IHoliday>",
		Response: "Paginated<IHoliday>",
		Note:     "Get pagination holiday",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.NoContent(http.StatusNoContent)
		}
		value, err := c.model.HolidayCurrentBranch(context, user.OrganizationID, *user.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.HolidayManager.Pagination(context, ctx, value))
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/holiday/:holiday_id",
		Method:   "GET",
		Response: "THoliday",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		holidayID, err := horizon.EngineUUIDParam(ctx, "holiday_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid holiday ID")
		}
		holiday, err := c.model.HolidayManager.GetByIDRaw(context, *holidayID)
		if err != nil {
			return c.NotFound(ctx, "Holiday")
		}
		return ctx.JSON(http.StatusOK, holiday)
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/holiday",
		Method:   "POST",
		Request:  "THoliday",
		Response: "THoliday",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := c.model.HolidayManager.Validate(ctx)
		if err != nil {
			return c.BadRequest(ctx, err.Error())
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.NoContent(http.StatusNoContent)
		}

		holiday := &model.Holiday{
			EntryDate:   req.EntryDate,
			Name:        req.Name,
			Description: req.Description,

			CreatedAt:      time.Now().UTC(),
			CreatedByID:    user.UserID,
			UpdatedAt:      time.Now().UTC(),
			UpdatedByID:    user.UserID,
			BranchID:       *user.BranchID,
			OrganizationID: user.OrganizationID,
		}

		if err := c.model.HolidayManager.Create(context, holiday); err != nil {
			return c.InternalServerError(ctx, err)
		}

		return ctx.JSON(http.StatusOK, c.model.HolidayManager.ToModel(holiday))
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/holiday/:holiday_id",
		Method:   "PUT",
		Request:  "THoliday",
		Response: "THoliday",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		holidayID, err := horizon.EngineUUIDParam(ctx, "holiday_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid holiday ID")
		}

		req, err := c.model.HolidayManager.Validate(ctx)
		if err != nil {
			return c.BadRequest(ctx, err.Error())
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.NoContent(http.StatusNoContent)
		}

		holiday, err := c.model.HolidayManager.GetByID(context, *holidayID)
		if err != nil {
			return c.NotFound(ctx, "Holiday")
		}
		holiday.EntryDate = req.EntryDate
		holiday.Name = req.Name
		holiday.Description = req.Description
		holiday.UpdatedAt = time.Now().UTC()
		holiday.UpdatedByID = user.UserID
		if err := c.model.HolidayManager.UpdateFields(context, holiday.ID, holiday); err != nil {
			return c.InternalServerError(ctx, err)
		}
		return ctx.JSON(http.StatusOK, c.model.HolidayManager.ToModel(holiday))
	})

	req.RegisterRoute(horizon.Route{
		Route:  "/holiday/:holiday_id",
		Method: "DELETE",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		holidayID, err := horizon.EngineUUIDParam(ctx, "holiday_id")
		if err != nil {
			return c.BadRequest(ctx, "Invalid holiday ID")
		}
		if err := c.model.HolidayManager.DeleteByID(context, *holidayID); err != nil {
			return c.InternalServerError(ctx, err)
		}
		return ctx.NoContent(http.StatusNoContent)
	})

	req.RegisterRoute(horizon.Route{
		Route:   "/holiday/bulk-delete",
		Method:  "DELETE",
		Request: "string[]",
		Note:    "Delete multiple holiday records",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var reqBody struct {
			IDs []string `json:"ids"`
		}
		if err := ctx.Bind(&reqBody); err != nil {
			return c.BadRequest(ctx, "Invalid request body")
		}
		if len(reqBody.IDs) == 0 {
			return c.BadRequest(ctx, "No IDs provided")
		}
		tx := c.provider.Service.Database.Client().Begin()
		if tx.Error != nil {
			tx.Rollback()
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": tx.Error.Error()})
		}
		for _, rawID := range reqBody.IDs {
			holidayID, err := uuid.Parse(rawID)
			if err != nil {
				tx.Rollback()
				return c.BadRequest(ctx, fmt.Sprintf("Invalid UUID: %s", rawID))
			}
			if _, err := c.model.HolidayManager.GetByID(context, holidayID); err != nil {
				tx.Rollback()
				return c.NotFound(ctx, fmt.Sprintf("Holiday with ID %s", rawID))
			}
			if err := c.model.HolidayManager.DeleteByIDWithTx(context, tx, holidayID); err != nil {
				tx.Rollback()
				return c.InternalServerError(ctx, err)
			}
		}
		if err := tx.Commit().Error; err != nil {
			return c.InternalServerError(ctx, err)
		}
		return ctx.NoContent(http.StatusNoContent)
	})
}
