package v1

import (
	"net/http"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/helpers"
	"github.com/Lands-Horizon-Corp/e-coop-server/horizon"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/event"
	"github.com/labstack/echo/v4"
)

func automaticLoanDeductionController(service *horizon.HorizonService) {
	req := service.API

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/automatic-loan-deduction/computation-sheet/:computation_sheet_id",
		Method:       "GET",
		Note:         "Returns all automatic loan deductions for a computation sheet in the current user's org/branch.",
		ResponseType: core.AutomaticLoanDeductionResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if userOrg.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		sheetID, err := helpers.EngineUUIDParam(ctx, "computation_sheet_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid computation sheet ID"})
		}
		alds, err := core.AutomaticLoanDeductionManager(service).Find(context, &core.AutomaticLoanDeduction{
			OrganizationID:     userOrg.OrganizationID,
			BranchID:           *userOrg.BranchID,
			ComputationSheetID: sheetID,
		})
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "No automatic loan deductions found for this computation sheet"})
		}
		return ctx.JSON(http.StatusOK, core.AutomaticLoanDeductionManager(service).ToModels(alds))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/automatic-loan-deduction/computation-sheet/:computation_sheet_id/search",
		Method:       "GET",
		Note:         "Returns all automatic loan deductions for a computation sheet in the current user's org/branch.",
		ResponseType: core.AutomaticLoanDeductionResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if userOrg.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		sheetID, err := helpers.EngineUUIDParam(ctx, "computation_sheet_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid computation sheet ID"})
		}
		alds, err := core.AutomaticLoanDeductionManager(service).NormalPagination(context, ctx, &core.AutomaticLoanDeduction{
			OrganizationID:     userOrg.OrganizationID,
			BranchID:           *userOrg.BranchID,
			ComputationSheetID: sheetID,
		})
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "No automatic loan deductions found for this computation sheet"})
		}
		return ctx.JSON(http.StatusOK, alds)
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/automatic-loan-deduction",
		Method:       "POST",
		Note:         "Creates a new automatic loan deduction for the current user's org/branch.",
		RequestType:  core.AutomaticLoanDeductionRequest{},
		ResponseType: core.AutomaticLoanDeductionResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		request, err := core.AutomaticLoanDeductionManager(service).Validate(ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Automatic loan deduction creation failed (/automatic-loan-deduction), validation error: " + err.Error(),
				Module:      "AutomaticLoanDeduction",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid data: " + err.Error()})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Automatic loan deduction creation failed (/automatic-loan-deduction), user org error: " + err.Error(),
				Module:      "AutomaticLoanDeduction",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if userOrg.BranchID == nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Automatic loan deduction creation failed (/automatic-loan-deduction), user not assigned to branch.",
				Module:      "AutomaticLoanDeduction",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		name := request.Name
		if name == "" {
			account, err := core.AccountManager(service).GetByID(context, *request.AccountID)
			if err != nil {
				event.Footstep(ctx, service, event.FootstepEvent{
					Activity:    "create-error",
					Description: "Automatic loan deduction creation failed (/automatic-loan-deduction), account fetch error: " + err.Error(),
					Module:      "AutomaticLoanDeduction",
				})
				return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid account ID: " + err.Error()})
			}
			name = account.Name
		}
		ald := &core.AutomaticLoanDeduction{
			AccountID:           request.AccountID,
			ComputationSheetID:  request.ComputationSheetID,
			ChargesRateSchemeID: request.ChargesRateSchemeID,
			ChargesPercentage1:  request.ChargesPercentage1,
			ChargesPercentage2:  request.ChargesPercentage2,
			ChargesAmount:       request.ChargesAmount,
			ChargesDivisor:      request.ChargesDivisor,
			MinAmount:           request.MinAmount,
			MaxAmount:           request.MaxAmount,
			Anum:                request.Anum,
			AddOn:               request.AddOn,
			AoRest:              request.AoRest,
			ExcludeRenewal:      request.ExcludeRenewal,
			Ct:                  request.Ct,
			Name:                name,
			Description:         request.Description,
			CreatedAt:           time.Now().UTC(),
			CreatedByID:         userOrg.UserID,
			UpdatedAt:           time.Now().UTC(),
			UpdatedByID:         userOrg.UserID,
			BranchID:            *userOrg.BranchID,
			OrganizationID:      userOrg.OrganizationID,
			NumberOfMonths:      request.NumberOfMonths,
		}

		if err := core.AutomaticLoanDeductionManager(service).Create(context, ald); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Automatic loan deduction creation failed (/automatic-loan-deduction), db error: " + err.Error(),
				Module:      "AutomaticLoanDeduction",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create automatic loan deduction: " + err.Error()})
		}
		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "create-success",
			Description: "Created automatic loan deduction (/automatic-loan-deduction): " + ald.Name,
			Module:      "AutomaticLoanDeduction",
		})
		return ctx.JSON(http.StatusCreated, core.AutomaticLoanDeductionManager(service).ToModel(ald))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/automatic-loan-deduction/:automatic_loan_deduction_id",
		Method:       "PUT",
		Note:         "Updates an existing automatic loan deduction by its ID.",
		RequestType:  core.AutomaticLoanDeductionRequest{},
		ResponseType: core.AutomaticLoanDeductionResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		id, err := helpers.EngineUUIDParam(ctx, "automatic_loan_deduction_id")
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Automatic loan deduction update failed (/automatic-loan-deduction/:automatic_loan_deduction_id), invalid ID.",
				Module:      "AutomaticLoanDeduction",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid automatic loan deduction ID"})
		}

		request, err := core.AutomaticLoanDeductionManager(service).Validate(ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Automatic loan deduction update failed (/automatic-loan-deduction/:automatic_loan_deduction_id), validation error: " + err.Error(),
				Module:      "AutomaticLoanDeduction",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid data: " + err.Error()})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Automatic loan deduction update failed (/automatic-loan-deduction/:automatic_loan_deduction_id), user org error: " + err.Error(),
				Module:      "AutomaticLoanDeduction",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		ald, err := core.AutomaticLoanDeductionManager(service).GetByID(context, *id)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Automatic loan deduction update failed (/automatic-loan-deduction/:automatic_loan_deduction_id), not found.",
				Module:      "AutomaticLoanDeduction",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Automatic loan deduction not found"})
		}
		name := request.Name
		if name == "" {
			account, err := core.AccountManager(service).GetByID(context, *request.AccountID)
			if err != nil {
				event.Footstep(ctx, service, event.FootstepEvent{
					Activity:    "create-error",
					Description: "Automatic loan deduction creation failed (/automatic-loan-deduction), account fetch error: " + err.Error(),
					Module:      "AutomaticLoanDeduction",
				})
				return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid account ID: " + err.Error()})
			}
			name = account.Name
		}
		ald.AccountID = request.AccountID
		ald.ComputationSheetID = request.ComputationSheetID
		ald.ChargesRateSchemeID = request.ChargesRateSchemeID
		ald.ChargesPercentage1 = request.ChargesPercentage1
		ald.ChargesPercentage2 = request.ChargesPercentage2
		ald.ChargesAmount = request.ChargesAmount
		ald.ChargesDivisor = request.ChargesDivisor
		ald.MinAmount = request.MinAmount
		ald.MaxAmount = request.MaxAmount
		ald.Anum = request.Anum
		ald.AddOn = request.AddOn
		ald.AoRest = request.AoRest
		ald.ExcludeRenewal = request.ExcludeRenewal
		ald.Ct = request.Ct
		ald.Name = name
		ald.Description = request.Description
		ald.UpdatedAt = time.Now().UTC()
		ald.UpdatedByID = userOrg.UserID
		ald.NumberOfMonths = request.NumberOfMonths

		if err := core.AutomaticLoanDeductionManager(service).UpdateByID(context, ald.ID, ald); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Automatic loan deduction update failed (/automatic-loan-deduction/:automatic_loan_deduction_id), db error: " + err.Error(),
				Module:      "AutomaticLoanDeduction",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update automatic loan deduction: " + err.Error()})
		}
		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Updated automatic loan deduction (/automatic-loan-deduction/:automatic_loan_deduction_id): " + ald.Name,
			Module:      "AutomaticLoanDeduction",
		})
		return ctx.JSON(http.StatusOK, core.AutomaticLoanDeductionManager(service).ToModel(ald))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:  "/api/v1/automatic-loan-deduction/:automatic_loan_deduction_id",
		Method: "DELETE",
		Note:   "Deletes the specified automatic loan deduction by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		id, err := helpers.EngineUUIDParam(ctx, "automatic_loan_deduction_id")
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Automatic loan deduction delete failed (/automatic-loan-deduction/:automatic_loan_deduction_id), invalid ID.",
				Module:      "AutomaticLoanDeduction",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid automatic loan deduction ID"})
		}
		ald, err := core.AutomaticLoanDeductionManager(service).GetByID(context, *id)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Automatic loan deduction delete failed (/automatic-loan-deduction/:automatic_loan_deduction_id), not found.",
				Module:      "AutomaticLoanDeduction",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Automatic loan deduction not found"})
		}
		if err := core.AutomaticLoanDeductionManager(service).Delete(context, *id); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Automatic loan deduction delete failed (/automatic-loan-deduction/:automatic_loan_deduction_id), db error: " + err.Error(),
				Module:      "AutomaticLoanDeduction",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete automatic loan deduction: " + err.Error()})
		}
		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "delete-success",
			Description: "Deleted automatic loan deduction (/automatic-loan-deduction/:automatic_loan_deduction_id): " + ald.Name,
			Module:      "AutomaticLoanDeduction",
		})
		return ctx.NoContent(http.StatusNoContent)
	})

	req.RegisterWebRoute(horizon.Route{
		Route:       "/api/v1/automatic-loan-deduction/bulk-delete",
		Method:      "DELETE",
		Note:        "Deletes multiple automatic loan deductions by their IDs. Expects a JSON body: { \"ids\": [\"id1\", \"id2\", ...] }",
		RequestType: core.IDSRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var reqBody core.IDSRequest
		if err := ctx.Bind(&reqBody); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Failed bulk delete automatic loan deductions (/automatic-loan-deduction/bulk-delete) | invalid request body: " + err.Error(),
				Module:      "AutomaticLoanDeduction",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body: " + err.Error()})
		}
		if len(reqBody.IDs) == 0 {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Failed bulk delete automatic loan deductions (/automatic-loan-deduction/bulk-delete) | no IDs provided",
				Module:      "AutomaticLoanDeduction",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "No IDs provided for bulk delete"})
		}
		ids := make([]any, len(reqBody.IDs))
		for i, id := range reqBody.IDs {
			ids[i] = id
		}
		if err := core.AutomaticLoanDeductionManager(service).BulkDelete(context, ids); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Failed bulk delete automatic loan deductions (/automatic-loan-deduction/bulk-delete) | error: " + err.Error(),
				Module:      "AutomaticLoanDeduction",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to bulk delete automatic loan deductions: " + err.Error()})
		}

		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "bulk-delete-success",
			Description: "Bulk deleted automatic loan deductions (/automatic-loan-deduction/bulk-delete)",
			Module:      "AutomaticLoanDeduction",
		})
		return ctx.NoContent(http.StatusNoContent)
	})
}
