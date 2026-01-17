package loan

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

func ComputationSheetController(service *horizon.HorizonService) {
	req := service.API

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/computation-sheet/:computation_sheet_id/calculator",
		Method:       "POST",
		Note:         "Returns sample payment calculation data for a computation sheet.",
		RequestType:  event.LoanComputationSheetCalculatorRequest{},
		ResponseType: event.ComputationSheetAmortizationResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var request event.LoanComputationSheetCalculatorRequest
		computationSheetID, err := helpers.EngineUUIDParam(ctx, "computation_sheet_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid computation sheet ID"})
		}
		if computationSheetID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Computation sheet ID is required"})
		}
		if err := ctx.Bind(&request); err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid computation sheet payload: " + err.Error()})
		}
		if err := service.Validator.Struct(request); err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}

		cashOnHandAccountID := userOrg.Branch.BranchSetting.CashOnHandAccountID
		computed, err := event.ComputationSheetCalculator(
			context, service,
			event.LoanComputationSheetCalculatorRequest{
				AccountID:                    request.AccountID,
				Applied1:                     request.Applied1,
				Terms:                        request.Terms,
				MemberTypeID:                 request.MemberTypeID,
				IsAddOn:                      request.IsAddOn,
				ExcludeSaturday:              request.ExcludeSaturday,
				ExcludeSunday:                request.ExcludeSunday,
				ExcludeHoliday:               request.ExcludeHoliday,
				ModeOfPaymentMonthlyExactDay: request.ModeOfPaymentMonthlyExactDay,
				ModeOfPaymentWeekly:          request.ModeOfPaymentWeekly,
				ModeOfPaymentSemiMonthlyPay1: request.ModeOfPaymentSemiMonthlyPay1,
				ModeOfPaymentSemiMonthlyPay2: request.ModeOfPaymentSemiMonthlyPay2,
				ModeOfPayment:                request.ModeOfPayment,
				Accounts:                     request.Accounts,
				CashOnHandAccountID:          cashOnHandAccountID,
				ComputationSheetID:           computationSheetID,
			},
		)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Computation failed: " + err.Error()})
		}

		return ctx.JSON(http.StatusOK, computed)
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/computation-sheet",
		Method:       "GET",
		Note:         "Returns all computation sheets for the current user's organization and branch.",
		ResponseType: types.ComputationSheetResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if userOrg.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		sheets, err := core.ComputationSheetCurrentBranch(context, service, userOrg.OrganizationID, *userOrg.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "No computation sheets found for the current branch"})
		}
		return ctx.JSON(http.StatusOK, core.ComputationSheetManager(service).ToModels(sheets))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/computation-sheet/:id",
		Method:       "GET",
		ResponseType: types.ComputationSheetResponse{},
		Note:         "Returns a single computation sheet by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		id, err := helpers.EngineUUIDParam(ctx, "id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid computation sheet ID"})
		}
		sheet, err := core.ComputationSheetManager(service).GetByIDRaw(context, *id)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Computation sheet not found"})
		}
		return ctx.JSON(http.StatusOK, sheet)
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/computation-sheet",
		Method:       "POST",
		RequestType: types.ComputationSheetRequest{},
		ResponseType: types.ComputationSheetResponse{},
		Note:         "Creates a new computation sheet for the current user's organization and branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := core.ComputationSheetManager(service).Validate(ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Computation sheet creation failed (/computation-sheet), validation error: " + err.Error(),
				Module:      "ComputationSheet",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid computation sheet data: " + err.Error()})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Computation sheet creation failed (/computation-sheet), user org error: " + err.Error(),
				Module:      "ComputationSheet",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if userOrg.BranchID == nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Computation sheet creation failed (/computation-sheet), user not assigned to branch.",
				Module:      "ComputationSheet",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}

		sheet := &types.ComputationSheet{
			Name:              req.Name,
			Description:       req.Description,
			DeliquentAccount:  req.DeliquentAccount,
			FinesAccount:      req.FinesAccount,
			InterestAccountID: req.InterestAccountID,
			ComakerAccount:    req.ComakerAccount,
			ExistAccount:      req.ExistAccount,
			CreatedAt:         time.Now().UTC(),
			CreatedByID:       userOrg.UserID,
			UpdatedAt:         time.Now().UTC(),
			UpdatedByID:       userOrg.UserID,
			BranchID:          *userOrg.BranchID,
			OrganizationID:    userOrg.OrganizationID,
			CurrencyID:        req.CurrencyID,
		}

		if err := core.ComputationSheetManager(service).Create(context, sheet); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Computation sheet creation failed (/computation-sheet), db error: " + err.Error(),
				Module:      "ComputationSheet",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create computation sheet: " + err.Error()})
		}
		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "create-success",
			Description: "Created computation sheet (/computation-sheet): " + sheet.Name,
			Module:      "ComputationSheet",
		})
		return ctx.JSON(http.StatusCreated, core.ComputationSheetManager(service).ToModel(sheet))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/computation-sheet/:id",
		Method:       "PUT",
		RequestType: types.ComputationSheetRequest{},
		ResponseType: types.ComputationSheetResponse{},
		Note:         "Updates an existing computation sheet by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		id, err := helpers.EngineUUIDParam(ctx, "id")
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Computation sheet update failed (/computation-sheet/:id), invalid ID.",
				Module:      "ComputationSheet",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid computation sheet ID"})
		}

		req, err := core.ComputationSheetManager(service).Validate(ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Computation sheet update failed (/computation-sheet/:id), validation error: " + err.Error(),
				Module:      "ComputationSheet",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid computation sheet data: " + err.Error()})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Computation sheet update failed (/computation-sheet/:id), user org error: " + err.Error(),
				Module:      "ComputationSheet",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		sheet, err := core.ComputationSheetManager(service).GetByID(context, *id)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Computation sheet update failed (/computation-sheet/:id), not found.",
				Module:      "ComputationSheet",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Computation sheet not found"})
		}
		sheet.Name = req.Name
		sheet.Description = req.Description
		sheet.DeliquentAccount = req.DeliquentAccount
		sheet.FinesAccount = req.FinesAccount
		sheet.InterestAccountID = req.InterestAccountID
		sheet.ComakerAccount = req.ComakerAccount
		sheet.ExistAccount = req.ExistAccount
		sheet.UpdatedAt = time.Now().UTC()
		sheet.UpdatedByID = userOrg.UserID

		if err := core.ComputationSheetManager(service).UpdateByID(context, sheet.ID, sheet); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Computation sheet update failed (/computation-sheet/:id), db error: " + err.Error(),
				Module:      "ComputationSheet",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update computation sheet: " + err.Error()})
		}
		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Updated computation sheet (/computation-sheet/:id): " + sheet.Name,
			Module:      "ComputationSheet",
		})
		return ctx.JSON(http.StatusOK, core.ComputationSheetManager(service).ToModel(sheet))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:  "/api/v1/computation-sheet/:id",
		Method: "DELETE",
		Note:   "Deletes the specified computation sheet by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		id, err := helpers.EngineUUIDParam(ctx, "id")
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Computation sheet delete failed (/computation-sheet/:id), invalid ID.",
				Module:      "ComputationSheet",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid computation sheet ID"})
		}
		sheet, err := core.ComputationSheetManager(service).GetByID(context, *id)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Computation sheet delete failed (/computation-sheet/:id), not found.",
				Module:      "ComputationSheet",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Computation sheet not found"})
		}
		if err := core.ComputationSheetManager(service).Delete(context, *id); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Computation sheet delete failed (/computation-sheet/:id), db error: " + err.Error(),
				Module:      "ComputationSheet",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete computation sheet: " + err.Error()})
		}
		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "delete-success",
			Description: "Deleted computation sheet (/computation-sheet/:id): " + sheet.Name,
			Module:      "ComputationSheet",
		})
		return ctx.NoContent(http.StatusNoContent)
	})

	req.RegisterWebRoute(horizon.Route{
		Route:       "/api/v1/computation-sheet/bulk-delete",
		Method:      "DELETE",
		Note:        "Deletes multiple computation sheets by their IDs. Expects a JSON body: { \"ids\": [\"id1\", \"id2\", ...] }",
		RequestType: types.IDSRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var reqBody types.IDSRequest

		if err := ctx.Bind(&reqBody); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete failed (/computation-sheet/bulk-delete) | invalid request body: " + err.Error(),
				Module:      "ComputationSheet",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body: " + err.Error()})
		}
		if len(reqBody.IDs) == 0 {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete failed (/computation-sheet/bulk-delete) | no IDs provided",
				Module:      "ComputationSheet",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "No computation sheet IDs provided for bulk delete"})
		}

		ids := make([]any, len(reqBody.IDs))
		for i, id := range reqBody.IDs {
			ids[i] = id
		}
		if err := core.ComputationSheetManager(service).BulkDelete(context, ids); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete failed (/computation-sheet/bulk-delete) | error: " + err.Error(),
				Module:      "ComputationSheet",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to bulk delete computation sheets: " + err.Error()})
		}

		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "bulk-delete-success",
			Description: "Bulk deleted computation sheets (/computation-sheet/bulk-delete)",
			Module:      "ComputationSheet",
		})
		return ctx.NoContent(http.StatusNoContent)
	})

}
