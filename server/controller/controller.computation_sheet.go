package v1

import (
	"fmt"
	"net/http"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/server/event"
	"github.com/Lands-Horizon-Corp/e-coop-server/server/model/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/services/handlers"
	"github.com/labstack/echo/v4"
)

// ComputationSheetController registers routes for managing computation sheets.
func (c *Controller) computationSheetController() {
	req := c.provider.Service.Request

	// POST /computation-sheet/:computation_sheet_id/calculator: Returns sample calculation data.
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/computation-sheet/:computation_sheet_id/calculator",
		Method:       "POST",
		Note:         "Returns sample payment calculation data for a computation sheet.",
		RequestType:  core.LoanComputationSheetCalculatorRequest{},
		ResponseType: core.ComputationSheetAmortizationResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var request core.LoanComputationSheetCalculatorRequest
		computationSheetID, err := handlers.EngineUUIDParam(ctx, "computation_sheet_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid computation sheet ID"})
		}
		if computationSheetID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Computation sheet ID is required"})
		}
		if err := ctx.Bind(&request); err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid login payload: " + err.Error()})
		}
		if err := c.provider.Service.Validator.Struct(request); err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		computationSheet, err := c.core.ComputationSheetManager.GetByID(context, *computationSheetID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Computation sheet not found"})
		}
		automaticLoanDeductionEntries, err := c.core.AutomaticLoanDeductionManager.Find(context, &core.AutomaticLoanDeduction{
			ComputationSheetID: &computationSheet.ID,
			BranchID:           computationSheet.BranchID,
			OrganizationID:     computationSheet.OrganizationID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve automatic loan deduction entries: " + err.Error()})
		}
		account, err := c.core.AccountManager.GetByID(context, *request.AccountID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Account not found"})
		}
		cashOnHandAccountID := userOrg.Branch.BranchSetting.CashOnHandAccountID
		cashOnHand, err := c.core.AccountManager.GetByID(context, *cashOnHandAccountID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve cash on hand account: " + err.Error()})
		}
		loanTransactionEntries := []*core.LoanTransactionEntry{
			{
				Account: cashOnHand,
				IsAddOn: false,
				Type:    core.LoanTransactionStatic,
				Debit:   0,
				Credit:  request.Applied1,
				Name:    account.Name,
			},
			{
				Account: account,
				IsAddOn: false,
				Type:    core.LoanTransactionStatic,
				Debit:   request.Applied1,
				Credit:  0,
				Name:    cashOnHand.Name,
			},
		}
		addOnEntry := &core.LoanTransactionEntry{
			Account: nil,
			Credit:  0,
			Debit:   0,
			Name:    "ADD ON INTEREST",
			Type:    core.LoanTransactionAddOn,
			IsAddOn: true,
		}
		totalNonAddOns, totalAddOns := 0.0, 0.0
		for _, ald := range automaticLoanDeductionEntries {
			if ald.AccountID == nil {
				continue
			}
			ald.Account, err = c.core.AccountManager.GetByID(context, *ald.AccountID)
			if err != nil {
				continue
			}
			entry := &core.LoanTransactionEntry{
				Credit:  0,
				Debit:   0,
				Name:    ald.Name,
				Type:    core.LoanTransactionDeduction,
				IsAddOn: ald.AddOn,
				Account: ald.Account,
			}
			if entry.AutomaticLoanDeduction.ChargesRateSchemeID != nil {
				chargesRateScheme, err := c.core.ChargesRateSchemeManager.GetByID(context, *entry.AutomaticLoanDeduction.ChargesRateSchemeID)
				if err != nil {
					return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": fmt.Sprintf("Failed to retrieve charges rate scheme for %s: %s", entry.Name, err.Error())})
				}
				entry.Credit = c.usecase.LoanChargesRateComputation(context, *chargesRateScheme, core.LoanTransaction{
					Applied1: request.Applied1,
					Terms:    request.Terms,
					MemberProfile: &core.MemberProfile{
						MemberTypeID: request.MemberTypeID,
					},
				})

			}
			if entry.Credit <= 0 {
				entry.Credit = c.usecase.LoanComputation(*ald, core.LoanTransaction{
					Terms:    request.Terms,
					Applied1: request.Applied1,
				})
			}

			if !entry.IsAddOn {
				totalNonAddOns = c.provider.Service.Decimal.Add(totalNonAddOns, entry.Credit)
			} else {
				totalAddOns = c.provider.Service.Decimal.Add(totalAddOns, entry.Credit)
			}
			if entry.Credit > 0 {
				loanTransactionEntries = append(loanTransactionEntries, entry)
			}
		}
		if request.IsAddOn {
			loanTransactionEntries[0].Credit = c.provider.Service.Decimal.Subtract(request.Applied1, totalNonAddOns)
		} else {
			loanTransactionEntries[0].Credit = c.provider.Service.Decimal.Subtract(request.Applied1, c.provider.Service.Decimal.Add(totalNonAddOns, totalAddOns))
		}
		if request.IsAddOn {
			addOnEntry.Debit = totalAddOns
			loanTransactionEntries = append(loanTransactionEntries, addOnEntry)
		}
		totalDebit, totalCredit := 0.0, 0.0
		for _, entry := range loanTransactionEntries {
			totalDebit = c.provider.Service.Decimal.Add(totalDebit, entry.Debit)
			totalCredit = c.provider.Service.Decimal.Add(totalCredit, entry.Credit)
		}
		return ctx.JSON(http.StatusOK, core.ComputationSheetAmortizationResponse{
			Entries:     c.core.LoanTransactionEntryManager.ToModels(loanTransactionEntries),
			TotalDebit:  totalDebit,
			TotalCredit: totalCredit,
		})
	})

	// GET /computation-sheet: List all computation sheets for the current user's branch.
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/computation-sheet",
		Method:       "GET",
		Note:         "Returns all computation sheets for the current user's organization and branch.",
		ResponseType: core.ComputationSheetResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if user.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		sheets, err := c.core.ComputationSheetCurrentBranch(context, user.OrganizationID, *user.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "No computation sheets found for the current branch"})
		}
		return ctx.JSON(http.StatusOK, c.core.ComputationSheetManager.ToModels(sheets))
	})

	// GET /computation-sheet/:id: Get specific computation sheet by ID.
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/computation-sheet/:id",
		Method:       "GET",
		ResponseType: core.ComputationSheetResponse{},
		Note:         "Returns a single computation sheet by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		id, err := handlers.EngineUUIDParam(ctx, "id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid computation sheet ID"})
		}
		sheet, err := c.core.ComputationSheetManager.GetByIDRaw(context, *id)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Computation sheet not found"})
		}
		return ctx.JSON(http.StatusOK, sheet)
	})

	// POST /computation-sheet: Create a new computation sheet.
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/computation-sheet",
		Method:       "POST",
		RequestType:  core.ComputationSheetRequest{},
		ResponseType: core.ComputationSheetResponse{},
		Note:         "Creates a new computation sheet for the current user's organization and branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := c.core.ComputationSheetManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Computation sheet creation failed (/computation-sheet), validation error: " + err.Error(),
				Module:      "ComputationSheet",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid computation sheet data: " + err.Error()})
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Computation sheet creation failed (/computation-sheet), user org error: " + err.Error(),
				Module:      "ComputationSheet",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if user.BranchID == nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Computation sheet creation failed (/computation-sheet), user not assigned to branch.",
				Module:      "ComputationSheet",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}

		sheet := &core.ComputationSheet{
			Name:              req.Name,
			Description:       req.Description,
			DeliquentAccount:  req.DeliquentAccount,
			FinesAccount:      req.FinesAccount,
			InterestAccountID: req.InterestAccountID,
			ComakerAccount:    req.ComakerAccount,
			ExistAccount:      req.ExistAccount,
			CreatedAt:         time.Now().UTC(),
			CreatedByID:       user.UserID,
			UpdatedAt:         time.Now().UTC(),
			UpdatedByID:       user.UserID,
			BranchID:          *user.BranchID,
			OrganizationID:    user.OrganizationID,
			CurrencyID:        req.CurrencyID,
		}

		if err := c.core.ComputationSheetManager.Create(context, sheet); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Computation sheet creation failed (/computation-sheet), db error: " + err.Error(),
				Module:      "ComputationSheet",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create computation sheet: " + err.Error()})
		}
		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "create-success",
			Description: "Created computation sheet (/computation-sheet): " + sheet.Name,
			Module:      "ComputationSheet",
		})
		return ctx.JSON(http.StatusCreated, c.core.ComputationSheetManager.ToModel(sheet))
	})

	// PUT /computation-sheet/:id: Update computation sheet by ID.
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/computation-sheet/:id",
		Method:       "PUT",
		RequestType:  core.ComputationSheetRequest{},
		ResponseType: core.ComputationSheetResponse{},
		Note:         "Updates an existing computation sheet by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		id, err := handlers.EngineUUIDParam(ctx, "id")
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Computation sheet update failed (/computation-sheet/:id), invalid ID.",
				Module:      "ComputationSheet",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid computation sheet ID"})
		}

		req, err := c.core.ComputationSheetManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Computation sheet update failed (/computation-sheet/:id), validation error: " + err.Error(),
				Module:      "ComputationSheet",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid computation sheet data: " + err.Error()})
		}
		user, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Computation sheet update failed (/computation-sheet/:id), user org error: " + err.Error(),
				Module:      "ComputationSheet",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		sheet, err := c.core.ComputationSheetManager.GetByID(context, *id)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
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
		sheet.UpdatedByID = user.UserID

		if err := c.core.ComputationSheetManager.UpdateByID(context, sheet.ID, sheet); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Computation sheet update failed (/computation-sheet/:id), db error: " + err.Error(),
				Module:      "ComputationSheet",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update computation sheet: " + err.Error()})
		}
		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Updated computation sheet (/computation-sheet/:id): " + sheet.Name,
			Module:      "ComputationSheet",
		})
		return ctx.JSON(http.StatusOK, c.core.ComputationSheetManager.ToModel(sheet))
	})

	// DELETE /computation-sheet/:id: Delete a computation sheet by ID.
	req.RegisterRoute(handlers.Route{
		Route:  "/api/v1/computation-sheet/:id",
		Method: "DELETE",
		Note:   "Deletes the specified computation sheet by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		id, err := handlers.EngineUUIDParam(ctx, "id")
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Computation sheet delete failed (/computation-sheet/:id), invalid ID.",
				Module:      "ComputationSheet",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid computation sheet ID"})
		}
		sheet, err := c.core.ComputationSheetManager.GetByID(context, *id)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Computation sheet delete failed (/computation-sheet/:id), not found.",
				Module:      "ComputationSheet",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Computation sheet not found"})
		}
		if err := c.core.ComputationSheetManager.Delete(context, *id); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Computation sheet delete failed (/computation-sheet/:id), db error: " + err.Error(),
				Module:      "ComputationSheet",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete computation sheet: " + err.Error()})
		}
		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "delete-success",
			Description: "Deleted computation sheet (/computation-sheet/:id): " + sheet.Name,
			Module:      "ComputationSheet",
		})
		return ctx.NoContent(http.StatusNoContent)
	})

	req.RegisterRoute(handlers.Route{
		Route:       "/api/v1/computation-sheet/bulk-delete",
		Method:      "DELETE",
		Note:        "Deletes multiple computation sheets by their IDs. Expects a JSON body: { \"ids\": [\"id1\", \"id2\", ...] }",
		RequestType: core.IDSRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var reqBody core.IDSRequest

		if err := ctx.Bind(&reqBody); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete failed (/computation-sheet/bulk-delete) | invalid request body: " + err.Error(),
				Module:      "ComputationSheet",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body: " + err.Error()})
		}
		if len(reqBody.IDs) == 0 {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete failed (/computation-sheet/bulk-delete) | no IDs provided",
				Module:      "ComputationSheet",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "No computation sheet IDs provided for bulk delete"})
		}

		if err := c.core.ComputationSheetManager.BulkDelete(context, reqBody.IDs); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete failed (/computation-sheet/bulk-delete) | error: " + err.Error(),
				Module:      "ComputationSheet",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to bulk delete computation sheets: " + err.Error()})
		}

		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "bulk-delete-success",
			Description: "Bulk deleted computation sheets (/computation-sheet/bulk-delete)",
			Module:      "ComputationSheet",
		})
		return ctx.NoContent(http.StatusNoContent)
	})

	// POST - api/v1/computation-sheeet/:computation-sheet-id/account/:account-id/connect
	// PUT - api/v1/computation-sheeet/:computation-sheet-id/account/:account-id/disconnect
}
