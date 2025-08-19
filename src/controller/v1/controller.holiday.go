package controller_v1

import (
	"fmt"
	"net/http"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/services/handlers"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/event"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/model"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// HolidayController manages endpoints for holiday records.
func (c *Controller) HolidayController() {
	req := c.provider.Service.Request

	// GET /holiday: List all holidays for the current user's branch. (NO footstep)
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/holiday",
		Method:       "GET",
		ResponseType: model.HolidayResponse{},
		Note:         "Returns all holiday records for the current user's organization and branch.",
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
		return ctx.JSON(http.StatusOK, c.model.HolidayManager.Filtered(context, ctx, holiday))
	})

	// GET /holiday/search: Paginated search of holidays for current branch. (NO footstep)
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/holiday/search",
		Method:       "GET",
		ResponseType: model.HolidayResponse{},
		Note:         "Returns a paginated list of holiday records for the current user's organization and branch.",
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

	// GET /holiday/:holiday_id: Get a specific holiday record by ID. (NO footstep)
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/holiday/:holiday_id",
		Method:       "GET",
		ResponseType: model.HolidayResponse{},
		RequestType:  model.HolidayRequest{},
		Note:         "Returns a holiday record by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		holidayID, err := handlers.EngineUUIDParam(ctx, "holiday_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid holiday ID"})
		}
		holiday, err := c.model.HolidayManager.GetByIDRaw(context, *holidayID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Holiday record not found"})
		}
		return ctx.JSON(http.StatusOK, holiday)
	})

	// POST /holiday: Create a new holiday record. (WITH footstep)
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/holiday",
		Method:       "POST",
		ResponseType: model.HolidayResponse{},
		RequestType:  model.HolidayRequest{},
		Note:         "Creates a new holiday record for the current user's organization and branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := c.model.HolidayManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Holiday creation failed (/holiday), validation error: " + err.Error(),
				Module:      "Holiday",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid holiday data: " + err.Error()})
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Holiday creation failed (/holiday), user org error: " + err.Error(),
				Module:      "Holiday",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization/branch not found"})
		}
		if user.BranchID == nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Holiday creation failed (/holiday), user not assigned to branch.",
				Module:      "Holiday",
			})
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
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Holiday creation failed (/holiday), db error: " + err.Error(),
				Module:      "Holiday",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create holiday record: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "create-success",
			Description: "Created holiday (/holiday): " + holiday.Name,
			Module:      "Holiday",
		})
		return ctx.JSON(http.StatusCreated, c.model.HolidayManager.ToModel(holiday))
	})

	// PUT /holiday/:holiday_id: Update a holiday record by ID. (WITH footstep)
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/holiday/:holiday_id",
		Method:       "PUT",
		ResponseType: model.HolidayResponse{},
		RequestType:  model.HolidayRequest{},
		Note:         "Updates an existing holiday record by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		holidayID, err := handlers.EngineUUIDParam(ctx, "holiday_id")
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Holiday update failed (/holiday/:holiday_id), invalid holiday ID.",
				Module:      "Holiday",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid holiday ID"})
		}
		req, err := c.model.HolidayManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Holiday update failed (/holiday/:holiday_id), validation error: " + err.Error(),
				Module:      "Holiday",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid holiday data: " + err.Error()})
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Holiday update failed (/holiday/:holiday_id), user org error: " + err.Error(),
				Module:      "Holiday",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization/branch not found"})
		}
		if user.BranchID == nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Holiday update failed (/holiday/:holiday_id), user not assigned to branch.",
				Module:      "Holiday",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		holiday, err := c.model.HolidayManager.GetByID(context, *holidayID)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Holiday update failed (/holiday/:holiday_id), not found.",
				Module:      "Holiday",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Holiday record not found"})
		}
		holiday.EntryDate = req.EntryDate
		holiday.Name = req.Name
		holiday.Description = req.Description
		holiday.UpdatedAt = time.Now().UTC()
		holiday.UpdatedByID = user.UserID
		if err := c.model.HolidayManager.UpdateFields(context, holiday.ID, holiday); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Holiday update failed (/holiday/:holiday_id), db error: " + err.Error(),
				Module:      "Holiday",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update holiday record: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Updated holiday (/holiday/:holiday_id): " + holiday.Name,
			Module:      "Holiday",
		})
		return ctx.JSON(http.StatusOK, c.model.HolidayManager.ToModel(holiday))
	})

	// DELETE /holiday/:holiday_id: Delete a holiday record by ID. (WITH footstep)
	req.RegisterRoute(handlers.Route{
		Route:  "/api/v1/holiday/:holiday_id",
		Method: "DELETE",
		Note:   "Deletes the specified holiday record by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		holidayID, err := handlers.EngineUUIDParam(ctx, "holiday_id")
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Holiday delete failed (/holiday/:holiday_id), invalid holiday ID.",
				Module:      "Holiday",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid holiday ID"})
		}
		holiday, err := c.model.HolidayManager.GetByID(context, *holidayID)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Holiday delete failed (/holiday/:holiday_id), not found.",
				Module:      "Holiday",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Holiday record not found"})
		}
		if err := c.model.HolidayManager.DeleteByID(context, *holidayID); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Holiday delete failed (/holiday/:holiday_id), db error: " + err.Error(),
				Module:      "Holiday",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete holiday record: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "delete-success",
			Description: "Deleted holiday (/holiday/:holiday_id): " + holiday.Name,
			Module:      "Holiday",
		})
		return ctx.NoContent(http.StatusNoContent)
	})

	// DELETE /holiday/bulk-delete: Bulk delete holiday records by IDs. (WITH footstep)
	req.RegisterRoute(handlers.Route{
		Route:       "/api/v1/holiday/bulk-delete",
		Method:      "DELETE",
		Note:        "Deletes multiple holiday records by their IDs. Expects a JSON body: { \"ids\": [\"id1\", \"id2\", ...] }",
		RequestType: model.IDSRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var reqBody model.IDSRequest
		if err := ctx.Bind(&reqBody); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Holiday bulk delete failed (/holiday/bulk-delete), invalid request body.",
				Module:      "Holiday",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
		}
		if len(reqBody.IDs) == 0 {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Holiday bulk delete failed (/holiday/bulk-delete), no IDs provided.",
				Module:      "Holiday",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "No IDs provided for bulk delete"})
		}
		tx := c.provider.Service.Database.Client().Begin()
		if tx.Error != nil {
			tx.Rollback()
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Holiday bulk delete failed (/holiday/bulk-delete), begin tx error: " + tx.Error.Error(),
				Module:      "Holiday",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to start database transaction: " + tx.Error.Error()})
		}
		names := ""
		for _, rawID := range reqBody.IDs {
			holidayID, err := uuid.Parse(rawID)
			if err != nil {
				tx.Rollback()
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "bulk-delete-error",
					Description: "Holiday bulk delete failed (/holiday/bulk-delete), invalid UUID: " + rawID,
					Module:      "Holiday",
				})
				return ctx.JSON(http.StatusBadRequest, map[string]string{"error": fmt.Sprintf("Invalid UUID: %s", rawID)})
			}
			holiday, err := c.model.HolidayManager.GetByID(context, holidayID)
			if err != nil {
				tx.Rollback()
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "bulk-delete-error",
					Description: "Holiday bulk delete failed (/holiday/bulk-delete), not found: " + rawID,
					Module:      "Holiday",
				})
				return ctx.JSON(http.StatusNotFound, map[string]string{"error": fmt.Sprintf("Holiday record not found with ID: %s", rawID)})
			}
			names += holiday.Name + ","
			if err := c.model.HolidayManager.DeleteByIDWithTx(context, tx, holidayID); err != nil {
				tx.Rollback()
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "bulk-delete-error",
					Description: "Holiday bulk delete failed (/holiday/bulk-delete), db error: " + err.Error(),
					Module:      "Holiday",
				})
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete holiday record: " + err.Error()})
			}
		}
		if err := tx.Commit().Error; err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Holiday bulk delete failed (/holiday/bulk-delete), commit error: " + err.Error(),
				Module:      "Holiday",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to commit transaction: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "bulk-delete-success",
			Description: "Bulk deleted holidays (/holiday/bulk-delete): " + names,
			Module:      "Holiday",
		})
		return ctx.NoContent(http.StatusNoContent)
	})
}
