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

// HolidayController manages endpoints for holiday records.
func (c *Controller) HolidayController() {
	req := c.provider.Service.Request

	// GET /holiday: List all holidays for the current user's branch.
	req.RegisterRoute(horizon.Route{
		Route:    "/holiday",
		Method:   "GET",
		Response: "THoliday[]",
		Note:     "Returns all holiday records for the current user's organization and branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization/branch not found"})
		}
		if user.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		holiday, err := c.model.HolidayCurrentBranch(context, user.OrganizationID, *user.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "No holiday records found for the current branch"})
		}
		return ctx.JSON(http.StatusOK, c.model.HolidayManager.ToModels(holiday))
	})

	// GET /holiday/search: Paginated search of holidays for current branch.
	req.RegisterRoute(horizon.Route{
		Route:    "/holiday/search",
		Method:   "GET",
		Request:  "Filter<IHoliday>",
		Response: "Paginated<IHoliday>",
		Note:     "Returns a paginated list of holiday records for the current user's organization and branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization/branch not found"})
		}
		if user.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		holidays, err := c.model.HolidayCurrentBranch(context, user.OrganizationID, *user.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch holiday records: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.HolidayManager.Pagination(context, ctx, holidays))
	})

	// GET /holiday/:holiday_id: Get a specific holiday record by ID.
	req.RegisterRoute(horizon.Route{
		Route:    "/holiday/:holiday_id",
		Method:   "GET",
		Response: "THoliday",
		Note:     "Returns a holiday record by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		holidayID, err := horizon.EngineUUIDParam(ctx, "holiday_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid holiday ID"})
		}
		holiday, err := c.model.HolidayManager.GetByIDRaw(context, *holidayID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Holiday record not found"})
		}
		return ctx.JSON(http.StatusOK, holiday)
	})

	// POST /holiday: Create a new holiday record.
	req.RegisterRoute(horizon.Route{
		Route:    "/holiday",
		Method:   "POST",
		Request:  "THoliday",
		Response: "THoliday",
		Note:     "Creates a new holiday record for the current user's organization and branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := c.model.HolidayManager.Validate(ctx)
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid holiday data: " + err.Error()})
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization/branch not found"})
		}
		if user.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		holiday := &model.Holiday{
			EntryDate:      req.EntryDate,
			Name:           req.Name,
			Description:    req.Description,
			CreatedAt:      time.Now().UTC(),
			CreatedByID:    user.UserID,
			UpdatedAt:      time.Now().UTC(),
			UpdatedByID:    user.UserID,
			BranchID:       *user.BranchID,
			OrganizationID: user.OrganizationID,
		}
		if err := c.model.HolidayManager.Create(context, holiday); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create holiday record: " + err.Error()})
		}
		return ctx.JSON(http.StatusCreated, c.model.HolidayManager.ToModel(holiday))
	})

	// PUT /holiday/:holiday_id: Update a holiday record by ID.
	req.RegisterRoute(horizon.Route{
		Route:    "/holiday/:holiday_id",
		Method:   "PUT",
		Request:  "THoliday",
		Response: "THoliday",
		Note:     "Updates an existing holiday record by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		holidayID, err := horizon.EngineUUIDParam(ctx, "holiday_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid holiday ID"})
		}
		req, err := c.model.HolidayManager.Validate(ctx)
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid holiday data: " + err.Error()})
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization/branch not found"})
		}
		if user.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		holiday, err := c.model.HolidayManager.GetByID(context, *holidayID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Holiday record not found"})
		}
		holiday.EntryDate = req.EntryDate
		holiday.Name = req.Name
		holiday.Description = req.Description
		holiday.UpdatedAt = time.Now().UTC()
		holiday.UpdatedByID = user.UserID
		if err := c.model.HolidayManager.UpdateFields(context, holiday.ID, holiday); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update holiday record: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.HolidayManager.ToModel(holiday))
	})

	// DELETE /holiday/:holiday_id: Delete a holiday record by ID.
	req.RegisterRoute(horizon.Route{
		Route:  "/holiday/:holiday_id",
		Method: "DELETE",
		Note:   "Deletes the specified holiday record by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		holidayID, err := horizon.EngineUUIDParam(ctx, "holiday_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid holiday ID"})
		}
		if err := c.model.HolidayManager.DeleteByID(context, *holidayID); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete holiday record: " + err.Error()})
		}
		return ctx.NoContent(http.StatusNoContent)
	})

	// DELETE /holiday/bulk-delete: Bulk delete holiday records by IDs.
	req.RegisterRoute(horizon.Route{
		Route:   "/holiday/bulk-delete",
		Method:  "DELETE",
		Request: "string[]",
		Note:    "Deletes multiple holiday records by their IDs. Expects a JSON body: { \"ids\": [\"id1\", \"id2\", ...] }",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var reqBody struct {
			IDs []string `json:"ids"`
		}
		if err := ctx.Bind(&reqBody); err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
		}
		if len(reqBody.IDs) == 0 {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "No IDs provided for bulk delete"})
		}
		tx := c.provider.Service.Database.Client().Begin()
		if tx.Error != nil {
			tx.Rollback()
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to start database transaction: " + tx.Error.Error()})
		}
		for _, rawID := range reqBody.IDs {
			holidayID, err := uuid.Parse(rawID)
			if err != nil {
				tx.Rollback()
				return ctx.JSON(http.StatusBadRequest, map[string]string{"error": fmt.Sprintf("Invalid UUID: %s", rawID)})
			}
			if _, err := c.model.HolidayManager.GetByID(context, holidayID); err != nil {
				tx.Rollback()
				return ctx.JSON(http.StatusNotFound, map[string]string{"error": fmt.Sprintf("Holiday record not found with ID: %s", rawID)})
			}
			if err := c.model.HolidayManager.DeleteByIDWithTx(context, tx, holidayID); err != nil {
				tx.Rollback()
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete holiday record: " + err.Error()})
			}
		}
		if err := tx.Commit().Error; err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to commit transaction: " + err.Error()})
		}
		return ctx.NoContent(http.StatusNoContent)
	})
}
