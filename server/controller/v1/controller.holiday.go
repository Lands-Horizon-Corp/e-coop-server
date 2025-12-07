package v1

import (
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/server/event"
	"github.com/Lands-Horizon-Corp/e-coop-server/server/model/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/services/handlers"
	"github.com/labstack/echo/v4"
)

// HolidayController manages endpoints for holiday records.
func (c *Controller) holidayController() {
	req := c.provider.Service.WebRequest

	// GET /holiday: List all holidays for the current user's branch. (NO footstep)
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/holiday",
		Method:       "GET",
		ResponseType: core.HolidayResponse{},
		Note:         "Returns all holiday records for the current user's organization and branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization/branch not found"})
		}
		if userOrg.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		holiday, err := c.core.HolidayCurrentBranch(context, userOrg.OrganizationID, *userOrg.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "No holiday records found for the current branch"})
		}
		return ctx.JSON(http.StatusOK, c.core.HolidayManager.ToModels(holiday))
	})

	// GET /holiday/search: Paginated search of holidays for current branch. (NO footstep)
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/holiday/search",
		Method:       "GET",
		ResponseType: core.HolidayResponse{},
		Note:         "Returns a paginated list of holiday records for the current user's organization and branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization/branch not found"})
		}
		if userOrg.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		holidays, err := c.core.HolidayManager.PaginationWithFields(context, ctx, &core.Holiday{
			BranchID:       *userOrg.BranchID,
			OrganizationID: userOrg.OrganizationID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch holiday records: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, holidays)
	})

	// GET /holiday/:holiday_id: Get a specific holiday record by ID. (NO footstep)
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/holiday/:holiday_id",
		Method:       "GET",
		ResponseType: core.HolidayResponse{},
		RequestType:  core.HolidayRequest{},
		Note:         "Returns a holiday record by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		holidayID, err := handlers.EngineUUIDParam(ctx, "holiday_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid holiday ID"})
		}
		holiday, err := c.core.HolidayManager.GetByIDRaw(context, *holidayID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Holiday record not found"})
		}
		return ctx.JSON(http.StatusOK, holiday)
	})

	// POST /holiday: Create a new holiday record. (WITH footstep)
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/holiday",
		Method:       "POST",
		ResponseType: core.HolidayResponse{},
		RequestType:  core.HolidayRequest{},
		Note:         "Creates a new holiday record for the current user's organization and branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := c.core.HolidayManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Holiday creation failed (/holiday), validation error: " + err.Error(),
				Module:      "Holiday",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid holiday data: " + err.Error()})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Holiday creation failed (/holiday), user org error: " + err.Error(),
				Module:      "Holiday",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization/branch not found"})
		}
		if userOrg.BranchID == nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Holiday creation failed (/holiday), user not assigned to branch.",
				Module:      "Holiday",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		holiday := &core.Holiday{
			EntryDate:      req.EntryDate,
			Name:           req.Name,
			Description:    req.Description,
			CreatedAt:      time.Now().UTC(),
			CreatedByID:    userOrg.UserID,
			UpdatedAt:      time.Now().UTC(),
			UpdatedByID:    userOrg.UserID,
			BranchID:       *userOrg.BranchID,
			OrganizationID: userOrg.OrganizationID,
			CurrencyID:     req.CurrencyID,
		}
		if err := c.core.HolidayManager.Create(context, holiday); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Holiday creation failed (/holiday), db error: " + err.Error(),
				Module:      "Holiday",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create holiday record: " + err.Error()})
		}
		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "create-success",
			Description: "Created holiday (/holiday): " + holiday.Name,
			Module:      "Holiday",
		})
		return ctx.JSON(http.StatusCreated, c.core.HolidayManager.ToModel(holiday))
	})

	// PUT /holiday/:holiday_id: Update a holiday record by ID. (WITH footstep)
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/holiday/:holiday_id",
		Method:       "PUT",
		ResponseType: core.HolidayResponse{},
		RequestType:  core.HolidayRequest{},
		Note:         "Updates an existing holiday record by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		holidayID, err := handlers.EngineUUIDParam(ctx, "holiday_id")
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Holiday update failed (/holiday/:holiday_id), invalid holiday ID.",
				Module:      "Holiday",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid holiday ID"})
		}
		req, err := c.core.HolidayManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Holiday update failed (/holiday/:holiday_id), validation error: " + err.Error(),
				Module:      "Holiday",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid holiday data: " + err.Error()})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Holiday update failed (/holiday/:holiday_id), user org error: " + err.Error(),
				Module:      "Holiday",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization/branch not found"})
		}
		if userOrg.BranchID == nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Holiday update failed (/holiday/:holiday_id), user not assigned to branch.",
				Module:      "Holiday",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		holiday, err := c.core.HolidayManager.GetByID(context, *holidayID)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Holiday update failed (/holiday/:holiday_id), not found.",
				Module:      "Holiday",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Holiday record not found"})
		}
		holiday.EntryDate = req.EntryDate
		holiday.CurrencyID = req.CurrencyID
		holiday.Name = req.Name
		holiday.Description = req.Description
		holiday.UpdatedAt = time.Now().UTC()
		holiday.UpdatedByID = userOrg.UserID
		if err := c.core.HolidayManager.UpdateByID(context, holiday.ID, holiday); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Holiday update failed (/holiday/:holiday_id), db error: " + err.Error(),
				Module:      "Holiday",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update holiday record: " + err.Error()})
		}
		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Updated holiday (/holiday/:holiday_id): " + holiday.Name,
			Module:      "Holiday",
		})
		return ctx.JSON(http.StatusOK, c.core.HolidayManager.ToModel(holiday))
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
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Holiday delete failed (/holiday/:holiday_id), invalid holiday ID.",
				Module:      "Holiday",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid holiday ID"})
		}
		holiday, err := c.core.HolidayManager.GetByID(context, *holidayID)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Holiday delete failed (/holiday/:holiday_id), not found.",
				Module:      "Holiday",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Holiday record not found"})
		}
		if err := c.core.HolidayManager.Delete(context, *holidayID); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Holiday delete failed (/holiday/:holiday_id), db error: " + err.Error(),
				Module:      "Holiday",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete holiday record: " + err.Error()})
		}
		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "delete-success",
			Description: "Deleted holiday (/holiday/:holiday_id): " + holiday.Name,
			Module:      "Holiday",
		})
		return ctx.NoContent(http.StatusNoContent)
	})

	// Simplified bulk-delete handler for holidays (mirrors the feedback bulk-delete pattern)
	req.RegisterRoute(handlers.Route{
		Route:       "/api/v1/holiday/bulk-delete",
		Method:      "DELETE",
		Note:        "Deletes multiple holiday records by their IDs. Expects a JSON body: { \"ids\": [\"id1\", \"id2\", ...] }",
		RequestType: core.IDSRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var reqBody core.IDSRequest

		if err := ctx.Bind(&reqBody); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Holiday bulk delete failed (/holiday/bulk-delete) | invalid request body: " + err.Error(),
				Module:      "Holiday",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body: " + err.Error()})
		}

		if len(reqBody.IDs) == 0 {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Holiday bulk delete failed (/holiday/bulk-delete) | no IDs provided",
				Module:      "Holiday",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "No IDs provided for bulk delete"})
		}

		if err := c.core.HolidayManager.BulkDelete(context, reqBody.IDs); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Holiday bulk delete failed (/holiday/bulk-delete) | error: " + err.Error(),
				Module:      "Holiday",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to bulk delete holiday records: " + err.Error()})
		}

		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "bulk-delete-success",
			Description: "Bulk deleted holidays (/holiday/bulk-delete)",
			Module:      "Holiday",
		})

		return ctx.NoContent(http.StatusNoContent)
	})

	// api/v1/holiday/year-available
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/holiday/year-available",
		Method:       "GET",
		ResponseType: core.HoldayYearAvaiable{},
		Note:         "Returns years with available holiday records for the current user's organization and branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization/branch not found"})
		}
		if userOrg.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		holidays, err := c.core.HolidayManager.Find(context, &core.Holiday{
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch years with holiday records: " + err.Error()})
		}

		// Count holidays by year
		yearCount := make(map[int]int)
		maxYear := 0
		for _, holiday := range holidays {
			year := holiday.EntryDate.Year()
			yearCount[year]++
			if year > maxYear {
				maxYear = year
			}
		}

		// If no holidays found, add current year with count 0
		if len(yearCount) == 0 {
			currentYear := time.Now().UTC().Year()
			yearCount[currentYear] = 0
			yearCount[currentYear+1] = 0 // Add next year as well
		} else {
			// Add one more year beyond the latest existing year with count 0
			yearCount[maxYear+1] = 0
		}

		var response []core.HoldayYearAvaiable
		for year, count := range yearCount {
			response = append(response, core.HoldayYearAvaiable{
				Year:  year,
				Count: count,
			})
		}
		sort.SliceStable(response, func(i, j int) bool {
			return response[i].Year < response[j].Year
		})
		return ctx.JSON(http.StatusOK, response)
	})

	// api/v1/holiday/year-available
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/holiday/currency/:currency_id/year-available",
		Method:       "GET",
		ResponseType: core.HoldayYearAvaiable{},
		Note:         "Returns years with available holiday records for a specific currency for the current user's organization and branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		currencyID, err := handlers.EngineUUIDParam(ctx, "currency_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid currency ID parameter"})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization/branch not found"})
		}
		if userOrg.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		holidays, err := c.core.HolidayManager.Find(context, &core.Holiday{
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
			CurrencyID:     *currencyID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch years with holiday records: " + err.Error()})
		}

		// Count holidays by year
		yearCount := make(map[int]int)
		maxYear := 0
		for _, holiday := range holidays {
			year := holiday.EntryDate.Year()
			yearCount[year]++
			if year > maxYear {
				maxYear = year
			}
		}

		// If no holidays found, add current year with count 0
		if len(yearCount) == 0 {
			currentYear := time.Now().UTC().Year()
			yearCount[currentYear] = 0
			yearCount[currentYear+1] = 0 // Add next year as well
		} else {
			// Add one more year beyond the latest existing year with count 0
			yearCount[maxYear+1] = 0
		}

		var response []core.HoldayYearAvaiable
		for year, count := range yearCount {
			response = append(response, core.HoldayYearAvaiable{
				Year:  year,
				Count: count,
			})
		}
		sort.SliceStable(response, func(i, j int) bool {
			return response[i].Year < response[j].Year
		})
		return ctx.JSON(http.StatusOK, response)
	})

	// GET api/v1/holiday/year/:year
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/holiday/year/:year",
		Method:       "GET",
		ResponseType: core.HolidayResponse{},
		Note:         "Returns holiday records for a specific year for the current user's organization and branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		yearParam := ctx.Param("year")
		year, err := strconv.Atoi(yearParam)
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid year parameter"})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization/branch not found"})
		}
		if userOrg.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		holiday, err := c.core.HolidayManager.Find(context, &core.Holiday{
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
		})
		result := []*core.Holiday{}
		for _, h := range holiday {
			if h.EntryDate.Year() == year {
				result = append(result, h)
			}
		}
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch holiday records for the year: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.core.HolidayManager.ToModels(result))
	})

	// GET api/v1/holiday/currency/:currency_id
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/holiday/currency/:currency_id",
		Method:       "GET",
		ResponseType: core.HolidayResponse{},
		Note:         "Returns holiday records for a specific currency for the current user's organization and branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		currencyID, err := handlers.EngineUUIDParam(ctx, "currency_id")
		if currencyID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid currency ID parameter"})
		}
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid currency ID parameter"})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization/branch not found"})
		}
		if userOrg.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		holiday, err := c.core.HolidayManager.Find(context, &core.Holiday{
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
			CurrencyID:     *currencyID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch holiday records for the currency: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.core.HolidayManager.ToModels(holiday))
	})

	// GET api/v1/holiday/year/:year/currency/:currency_id
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/holiday/year/:year/currency/:currency_id",
		Method:       "GET",
		ResponseType: core.HolidayResponse{},
		Note:         "Returns holiday records for a specific year and currency for the current user's organization and branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		yearParam := ctx.Param("year")
		year, err := strconv.Atoi(yearParam)
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid year parameter"})
		}
		currencyID, err := handlers.EngineUUIDParam(ctx, "currency_id")
		if currencyID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid currency ID parameter"})
		}
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid currency ID parameter"})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization/branch not found"})
		}
		if userOrg.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		holiday, err := c.core.HolidayManager.Find(context, &core.Holiday{
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
			CurrencyID:     *currencyID,
		})
		result := []*core.Holiday{}
		for _, h := range holiday {
			if h.EntryDate.Year() == year {
				result = append(result, h)
			}
		}
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch holiday records for the year and currency: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.core.HolidayManager.ToModels(result))
	})

	// POST /api/v1/holiday/year/:year/currency/:currency/copy/:year
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/holiday/year/:year/currency/:currency_id/copy/:source_year",
		Method:       "POST",
		ResponseType: core.HolidayResponse{},
		Note:         "Copies holiday records from source year to target year for a specific currency in the current user's organization and branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		yearParam := ctx.Param("year")
		targetYear, err := strconv.Atoi(yearParam)
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid target year parameter"})
		}
		sourceYearParam := ctx.Param("source_year")
		sourceYear, err := strconv.Atoi(sourceYearParam)
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid source year parameter"})
		}
		currencyID, err := handlers.EngineUUIDParam(ctx, "currency_id")
		if currencyID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid currency ID parameter"})
		}
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid currency ID parameter"})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization/branch not found"})
		}
		if userOrg.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		holidays, err := c.core.HolidayManager.Find(context, &core.Holiday{
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
			CurrencyID:     *currencyID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch holiday records for the currency: " + err.Error()})
		}
		var copiedHolidays []*core.Holiday
		for _, h := range holidays {
			if h.EntryDate.Year() == sourceYear {
				newHoliday := &core.Holiday{
					EntryDate:      time.Date(targetYear, h.EntryDate.Month(), h.EntryDate.Day(), 0, 0, 0, 0, time.UTC),
					Name:           h.Name,
					Description:    h.Description,
					CreatedAt:      time.Now().UTC(),
					CreatedByID:    userOrg.UserID,
					UpdatedAt:      time.Now().UTC(),
					UpdatedByID:    userOrg.UserID,
					BranchID:       *userOrg.BranchID,
					OrganizationID: userOrg.OrganizationID,
					CurrencyID:     h.CurrencyID,
				}
				if err := c.core.HolidayManager.Create(context, newHoliday); err != nil {
					return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to copy holiday record: " + err.Error()})
				}
				copiedHolidays = append(copiedHolidays, newHoliday)
			}
		}
		return ctx.JSON(http.StatusCreated, c.core.HolidayManager.ToModels(copiedHolidays))
	})

}
