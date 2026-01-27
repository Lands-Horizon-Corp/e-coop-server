package settings

import (
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/helpers"
	"github.com/Lands-Horizon-Corp/e-coop-server/horizon"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/db/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/event"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/types"
	"github.com/labstack/echo/v4"
)

func HolidayController(service *horizon.HorizonService) {

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/holiday",
		Method:       "GET",
		ResponseType: types.HolidayResponse{},
		Note:         "Returns all holiday records for the current user's organization and branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization/branch not found"})
		}
		if userOrg.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		holiday, err := core.HolidayCurrentBranch(context, service, userOrg.OrganizationID, *userOrg.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "No holiday records found for the current branch"})
		}
		return ctx.JSON(http.StatusOK, core.HolidayManager(service).ToModels(holiday))
	})

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/holiday/search",
		Method:       "GET",
		ResponseType: types.HolidayResponse{},
		Note:         "Returns a paginated list of holiday records for the current user's organization and branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization/branch not found"})
		}
		if userOrg.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		holidays, err := core.HolidayManager(service).NormalPagination(context, ctx, &types.Holiday{
			BranchID:       *userOrg.BranchID,
			OrganizationID: userOrg.OrganizationID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch holiday records: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, holidays)
	})

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/holiday/:holiday_id",
		Method:       "GET",
		ResponseType: types.HolidayResponse{},
		RequestType:  types.HolidayRequest{},
		Note:         "Returns a holiday record by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		holidayID, err := helpers.EngineUUIDParam(ctx, "holiday_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid holiday ID"})
		}
		holiday, err := core.HolidayManager(service).GetByIDRaw(context, *holidayID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Holiday record not found"})
		}
		return ctx.JSON(http.StatusOK, holiday)
	})

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/holiday",
		Method:       "POST",
		ResponseType: types.HolidayResponse{},
		RequestType:  types.HolidayRequest{},
		Note:         "Creates a new holiday record for the current user's organization and branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := core.HolidayManager(service).Validate(ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Holiday creation failed (/holiday), validation error: " + err.Error(),
				Module:      "Holiday",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid holiday data: " + err.Error()})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Holiday creation failed (/holiday), user org error: " + err.Error(),
				Module:      "Holiday",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization/branch not found"})
		}
		if userOrg.BranchID == nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Holiday creation failed (/holiday), user not assigned to branch.",
				Module:      "Holiday",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		holiday := &types.Holiday{
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
		if err := core.HolidayManager(service).Create(context, holiday); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Holiday creation failed (/holiday), db error: " + err.Error(),
				Module:      "Holiday",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create holiday record: " + err.Error()})
		}
		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "create-success",
			Description: "Created holiday (/holiday): " + holiday.Name,
			Module:      "Holiday",
		})
		return ctx.JSON(http.StatusCreated, core.HolidayManager(service).ToModel(holiday))
	})

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/holiday/:holiday_id",
		Method:       "PUT",
		ResponseType: types.HolidayResponse{},
		RequestType:  types.HolidayRequest{},
		Note:         "Updates an existing holiday record by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		holidayID, err := helpers.EngineUUIDParam(ctx, "holiday_id")
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Holiday update failed (/holiday/:holiday_id), invalid holiday ID.",
				Module:      "Holiday",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid holiday ID"})
		}
		req, err := core.HolidayManager(service).Validate(ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Holiday update failed (/holiday/:holiday_id), validation error: " + err.Error(),
				Module:      "Holiday",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid holiday data: " + err.Error()})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Holiday update failed (/holiday/:holiday_id), user org error: " + err.Error(),
				Module:      "Holiday",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization/branch not found"})
		}
		if userOrg.BranchID == nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Holiday update failed (/holiday/:holiday_id), user not assigned to branch.",
				Module:      "Holiday",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		holiday, err := core.HolidayManager(service).GetByID(context, *holidayID)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
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
		if err := core.HolidayManager(service).UpdateByID(context, holiday.ID, holiday); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Holiday update failed (/holiday/:holiday_id), db error: " + err.Error(),
				Module:      "Holiday",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update holiday record: " + err.Error()})
		}
		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Updated holiday (/holiday/:holiday_id): " + holiday.Name,
			Module:      "Holiday",
		})
		return ctx.JSON(http.StatusOK, core.HolidayManager(service).ToModel(holiday))
	})

	service.API.RegisterWebRoute(horizon.Route{
		Route:  "/api/v1/holiday/:holiday_id",
		Method: "DELETE",
		Note:   "Deletes the specified holiday record by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		holidayID, err := helpers.EngineUUIDParam(ctx, "holiday_id")
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Holiday delete failed (/holiday/:holiday_id), invalid holiday ID.",
				Module:      "Holiday",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid holiday ID"})
		}
		holiday, err := core.HolidayManager(service).GetByID(context, *holidayID)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Holiday delete failed (/holiday/:holiday_id), not found.",
				Module:      "Holiday",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Holiday record not found"})
		}
		if err := core.HolidayManager(service).Delete(context, *holidayID); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Holiday delete failed (/holiday/:holiday_id), db error: " + err.Error(),
				Module:      "Holiday",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete holiday record: " + err.Error()})
		}
		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "delete-success",
			Description: "Deleted holiday (/holiday/:holiday_id): " + holiday.Name,
			Module:      "Holiday",
		})
		return ctx.NoContent(http.StatusNoContent)
	})

	service.API.RegisterWebRoute(horizon.Route{
		Route:       "/api/v1/holiday/bulk-delete",
		Method:      "DELETE",
		Note:        "Deletes multiple holiday records by their IDs. Expects a JSON body: { \"ids\": [\"id1\", \"id2\", ...] }",
		RequestType: types.IDSRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var reqBody types.IDSRequest

		if err := ctx.Bind(&reqBody); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Holiday bulk delete failed (/holiday/bulk-delete) | invalid request body: " + err.Error(),
				Module:      "Holiday",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body: " + err.Error()})
		}

		if len(reqBody.IDs) == 0 {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Holiday bulk delete failed (/holiday/bulk-delete) | no IDs provided",
				Module:      "Holiday",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "No IDs provided for bulk delete"})
		}
		ids := make([]any, len(reqBody.IDs))
		for i, id := range reqBody.IDs {
			ids[i] = id
		}
		if err := core.HolidayManager(service).BulkDelete(context, ids); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Holiday bulk delete failed (/holiday/bulk-delete) | error: " + err.Error(),
				Module:      "Holiday",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to bulk delete holiday records: " + err.Error()})
		}

		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "bulk-delete-success",
			Description: "Bulk deleted holidays (/holiday/bulk-delete)",
			Module:      "Holiday",
		})

		return ctx.NoContent(http.StatusNoContent)
	})

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/holiday/year-available",
		Method:       "GET",
		ResponseType: types.HoldayYearAvaiable{},
		Note:         "Returns years with available holiday records for the current user's organization and branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization/branch not found"})
		}
		if userOrg.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		holidays, err := core.HolidayManager(service).Find(context, &types.Holiday{
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch years with holiday records: " + err.Error()})
		}

		yearCount := make(map[int]int)
		maxYear := 0
		for _, holiday := range holidays {
			year := holiday.EntryDate.Year()
			yearCount[year]++
			if year > maxYear {
				maxYear = year
			}
		}

		if len(yearCount) == 0 {
			currentYear := time.Now().UTC().Year()
			yearCount[currentYear] = 0
			yearCount[currentYear+1] = 0 // Add next year as well
		} else {
			yearCount[maxYear+1] = 0
		}

		var response []types.HoldayYearAvaiable
		for year, count := range yearCount {
			response = append(response, types.HoldayYearAvaiable{
				Year:  year,
				Count: count,
			})
		}
		sort.SliceStable(response, func(i, j int) bool {
			return response[i].Year < response[j].Year
		})
		return ctx.JSON(http.StatusOK, response)
	})

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/holiday/currency/:currency_id/year-available",
		Method:       "GET",
		ResponseType: types.HoldayYearAvaiable{},
		Note:         "Returns years with available holiday records for a specific currency for the current user's organization and branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		currencyID, err := helpers.EngineUUIDParam(ctx, "currency_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid currency ID parameter"})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization/branch not found"})
		}
		if userOrg.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		holidays, err := core.HolidayManager(service).Find(context, &types.Holiday{
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
			CurrencyID:     *currencyID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch years with holiday records: " + err.Error()})
		}

		yearCount := make(map[int]int)
		maxYear := 0
		for _, holiday := range holidays {
			year := holiday.EntryDate.Year()
			yearCount[year]++
			if year > maxYear {
				maxYear = year
			}
		}

		if len(yearCount) == 0 {
			currentYear := time.Now().UTC().Year()
			yearCount[currentYear] = 0
			yearCount[currentYear+1] = 0 // Add next year as well
		} else {
			yearCount[maxYear+1] = 0
		}

		var response []types.HoldayYearAvaiable
		for year, count := range yearCount {
			response = append(response, types.HoldayYearAvaiable{
				Year:  year,
				Count: count,
			})
		}
		sort.SliceStable(response, func(i, j int) bool {
			return response[i].Year < response[j].Year
		})
		return ctx.JSON(http.StatusOK, response)
	})

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/holiday/year/:year",
		Method:       "GET",
		ResponseType: types.HolidayResponse{},
		Note:         "Returns holiday records for a specific year for the current user's organization and branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		yearParam := ctx.Param("year")
		year, err := strconv.Atoi(yearParam)
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid year parameter"})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization/branch not found"})
		}
		if userOrg.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		holiday, err := core.HolidayManager(service).Find(context, &types.Holiday{
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
		})
		result := []*types.Holiday{}
		for _, h := range holiday {
			if h.EntryDate.Year() == year {
				result = append(result, h)
			}
		}
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch holiday records for the year: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, core.HolidayManager(service).ToModels(result))
	})

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/holiday/currency/:currency_id",
		Method:       "GET",
		ResponseType: types.HolidayResponse{},
		Note:         "Returns holiday records for a specific currency for the current user's organization and branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		currencyID, err := helpers.EngineUUIDParam(ctx, "currency_id")
		if currencyID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid currency ID parameter"})
		}
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid currency ID parameter"})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization/branch not found"})
		}
		if userOrg.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		holiday, err := core.HolidayManager(service).Find(context, &types.Holiday{
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
			CurrencyID:     *currencyID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch holiday records for the currency: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, core.HolidayManager(service).ToModels(holiday))
	})

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/holiday/year/:year/currency/:currency_id",
		Method:       "GET",
		ResponseType: types.HolidayResponse{},
		Note:         "Returns holiday records for a specific year and currency for the current user's organization and branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		yearParam := ctx.Param("year")
		year, err := strconv.Atoi(yearParam)
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid year parameter"})
		}
		currencyID, err := helpers.EngineUUIDParam(ctx, "currency_id")
		if currencyID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid currency ID parameter"})
		}
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid currency ID parameter"})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization/branch not found"})
		}
		if userOrg.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		holiday, err := core.HolidayManager(service).Find(context, &types.Holiday{
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
			CurrencyID:     *currencyID,
		})
		result := []*types.Holiday{}
		for _, h := range holiday {
			if h.EntryDate.Year() == year {
				result = append(result, h)
			}
		}
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch holiday records for the year and currency: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, core.HolidayManager(service).ToModels(result))
	})

	service.API.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/holiday/year/:year/currency/:currency_id/copy/:source_year",
		Method:       "POST",
		ResponseType: types.HolidayResponse{},
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
		currencyID, err := helpers.EngineUUIDParam(ctx, "currency_id")
		if currencyID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid currency ID parameter"})
		}
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid currency ID parameter"})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization/branch not found"})
		}
		if userOrg.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		holidays, err := core.HolidayManager(service).Find(context, &types.Holiday{
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
			CurrencyID:     *currencyID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch holiday records for the currency: " + err.Error()})
		}
		var copiedHolidays []*types.Holiday
		for _, h := range holidays {
			if h.EntryDate.Year() == sourceYear {
				newHoliday := &types.Holiday{
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
				if err := core.HolidayManager(service).Create(context, newHoliday); err != nil {
					return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to copy holiday record: " + err.Error()})
				}
				copiedHolidays = append(copiedHolidays, newHoliday)
			}
		}
		return ctx.JSON(http.StatusCreated, core.HolidayManager(service).ToModels(copiedHolidays))
	})

}
