package v1

import (
	"net/http"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/server/event"
	"github.com/Lands-Horizon-Corp/e-coop-server/server/model/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/services/handlers"
	"github.com/labstack/echo/v4"
)

// AutomaticLoanDeductionController registers routes for managing automatic loan deductions.
func (c *Controller) automaticLoanDeductionController() {
	req := c.provider.Service.Request

	// GET /automatic-loan-deduction/computation-sheet/:computation_sheet_id/search
	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/automatic-loan-deduction/computation-sheet/:computation_sheet_id",
		Method:       "GET",
		Note:         "Returns all automatic loan deductions for a computation sheet in the current user's org/branch.",
		ResponseType: core.AutomaticLoanDeductionResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if userOrg.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		sheetID, err := handlers.EngineUUIDParam(ctx, "computation_sheet_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid computation sheet ID"})
		}
		// Find all for this computation sheet, org, and branch
		alds, err := c.core.AutomaticLoanDeductionManager.Find(context, &core.AutomaticLoanDeduction{
			OrganizationID:     userOrg.OrganizationID,
			BranchID:           *userOrg.BranchID,
			ComputationSheetID: sheetID,
		})
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "No automatic loan deductions found for this computation sheet"})
		}
		return ctx.JSON(http.StatusOK, c.core.AutomaticLoanDeductionManager.ToModels(alds))
	})

	// GET /automatic-loan-deduction/computation-sheet/:computation_sheet_id/search
	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/automatic-loan-deduction/computation-sheet/:computation_sheet_id/search",
		Method:       "GET",
		Note:         "Returns all automatic loan deductions for a computation sheet in the current user's org/branch.",
		ResponseType: core.AutomaticLoanDeductionResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if userOrg.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		sheetID, err := handlers.EngineUUIDParam(ctx, "computation_sheet_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid computation sheet ID"})
		}
		// Find all for this computation sheet, org, and branch
		alds, err := c.core.AutomaticLoanDeductionManager.PaginationWithFields(context, ctx, &core.AutomaticLoanDeduction{
			OrganizationID:     userOrg.OrganizationID,
			BranchID:           *userOrg.BranchID,
			ComputationSheetID: sheetID,
		})
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "No automatic loan deductions found for this computation sheet"})
		}
		return ctx.JSON(http.StatusOK, alds)
	})

	// POST /automatic-loan-deduction
	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/automatic-loan-deduction",
		Method:       "POST",
		Note:         "Creates a new automatic loan deduction for the current user's org/branch.",
		RequestType:  core.AutomaticLoanDeductionRequest{},
		ResponseType: core.AutomaticLoanDeductionResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		request, err := c.core.AutomaticLoanDeductionManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Automatic loan deduction creation failed (/automatic-loan-deduction), validation error: " + err.Error(),
				Module:      "AutomaticLoanDeduction",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid data: " + err.Error()})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Automatic loan deduction creation failed (/automatic-loan-deduction), user org error: " + err.Error(),
				Module:      "AutomaticLoanDeduction",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if userOrg.BranchID == nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Automatic loan deduction creation failed (/automatic-loan-deduction), user not assigned to branch.",
				Module:      "AutomaticLoanDeduction",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		name := request.Name
		if name == "" {
			account, err := c.core.AccountManager.GetByID(context, *request.AccountID)
			if err != nil {
				c.event.Footstep(ctx, event.FootstepEvent{
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

		if err := c.core.AutomaticLoanDeductionManager.Create(context, ald); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Automatic loan deduction creation failed (/automatic-loan-deduction), db error: " + err.Error(),
				Module:      "AutomaticLoanDeduction",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create automatic loan deduction: " + err.Error()})
		}
		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "create-success",
			Description: "Created automatic loan deduction (/automatic-loan-deduction): " + ald.Name,
			Module:      "AutomaticLoanDeduction",
		})
		return ctx.JSON(http.StatusCreated, c.core.AutomaticLoanDeductionManager.ToModel(ald))
	})

	// PUT /automatic-loan-deduction/:automatic_loan_deduction_id
	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/automatic-loan-deduction/:automatic_loan_deduction_id",
		Method:       "PUT",
		Note:         "Updates an existing automatic loan deduction by its ID.",
		RequestType:  core.AutomaticLoanDeductionRequest{},
		ResponseType: core.AutomaticLoanDeductionResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		id, err := handlers.EngineUUIDParam(ctx, "automatic_loan_deduction_id")
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Automatic loan deduction update failed (/automatic-loan-deduction/:automatic_loan_deduction_id), invalid ID.",
				Module:      "AutomaticLoanDeduction",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid automatic loan deduction ID"})
		}

		request, err := c.core.AutomaticLoanDeductionManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Automatic loan deduction update failed (/automatic-loan-deduction/:automatic_loan_deduction_id), validation error: " + err.Error(),
				Module:      "AutomaticLoanDeduction",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid data: " + err.Error()})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Automatic loan deduction update failed (/automatic-loan-deduction/:automatic_loan_deduction_id), user org error: " + err.Error(),
				Module:      "AutomaticLoanDeduction",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		ald, err := c.core.AutomaticLoanDeductionManager.GetByID(context, *id)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Automatic loan deduction update failed (/automatic-loan-deduction/:automatic_loan_deduction_id), not found.",
				Module:      "AutomaticLoanDeduction",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Automatic loan deduction not found"})
		}
		name := request.Name
		if name == "" {
			account, err := c.core.AccountManager.GetByID(context, *request.AccountID)
			if err != nil {
				c.event.Footstep(ctx, event.FootstepEvent{
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

		if err := c.core.AutomaticLoanDeductionManager.UpdateByID(context, ald.ID, ald); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Automatic loan deduction update failed (/automatic-loan-deduction/:automatic_loan_deduction_id), db error: " + err.Error(),
				Module:      "AutomaticLoanDeduction",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update automatic loan deduction: " + err.Error()})
		}
		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Updated automatic loan deduction (/automatic-loan-deduction/:automatic_loan_deduction_id): " + ald.Name,
			Module:      "AutomaticLoanDeduction",
		})
		return ctx.JSON(http.StatusOK, c.core.AutomaticLoanDeductionManager.ToModel(ald))
	})

	// DELETE /automatic-loan-deduction/:automatic_loan_deduction_id
	req.RegisterWebRoute(handlers.Route{
		Route:  "/api/v1/automatic-loan-deduction/:automatic_loan_deduction_id",
		Method: "DELETE",
		Note:   "Deletes the specified automatic loan deduction by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		id, err := handlers.EngineUUIDParam(ctx, "automatic_loan_deduction_id")
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Automatic loan deduction delete failed (/automatic-loan-deduction/:automatic_loan_deduction_id), invalid ID.",
				Module:      "AutomaticLoanDeduction",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid automatic loan deduction ID"})
		}
		ald, err := c.core.AutomaticLoanDeductionManager.GetByID(context, *id)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Automatic loan deduction delete failed (/automatic-loan-deduction/:automatic_loan_deduction_id), not found.",
				Module:      "AutomaticLoanDeduction",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Automatic loan deduction not found"})
		}
		if err := c.core.AutomaticLoanDeductionManager.Delete(context, *id); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Automatic loan deduction delete failed (/automatic-loan-deduction/:automatic_loan_deduction_id), db error: " + err.Error(),
				Module:      "AutomaticLoanDeduction",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete automatic loan deduction: " + err.Error()})
		}
		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "delete-success",
			Description: "Deleted automatic loan deduction (/automatic-loan-deduction/:automatic_loan_deduction_id): " + ald.Name,
			Module:      "AutomaticLoanDeduction",
		})
		return ctx.NoContent(http.StatusNoContent)
	})

	req.RegisterWebRoute(handlers.Route{
		Route:       "/api/v1/automatic-loan-deduction/bulk-delete",
		Method:      "DELETE",
		Note:        "Deletes multiple automatic loan deductions by their IDs. Expects a JSON body: { \"ids\": [\"id1\", \"id2\", ...] }",
		RequestType: core.IDSRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var reqBody core.IDSRequest
		if err := ctx.Bind(&reqBody); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Failed bulk delete automatic loan deductions (/automatic-loan-deduction/bulk-delete) | invalid request body: " + err.Error(),
				Module:      "AutomaticLoanDeduction",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body: " + err.Error()})
		}
		if len(reqBody.IDs) == 0 {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Failed bulk delete automatic loan deductions (/automatic-loan-deduction/bulk-delete) | no IDs provided",
				Module:      "AutomaticLoanDeduction",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "No IDs provided for bulk delete"})
		}

		if err := c.core.AutomaticLoanDeductionManager.BulkDelete(context, reqBody.IDs); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Failed bulk delete automatic loan deductions (/automatic-loan-deduction/bulk-delete) | error: " + err.Error(),
				Module:      "AutomaticLoanDeduction",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to bulk delete automatic loan deductions: " + err.Error()})
		}

		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "bulk-delete-success",
			Description: "Bulk deleted automatic loan deductions (/automatic-loan-deduction/bulk-delete)",
			Module:      "AutomaticLoanDeduction",
		})
		return ctx.NoContent(http.StatusNoContent)
	})
}
