package controller_v1

import (
	"fmt"
	"net/http"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/services/handlers"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/event"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/model"
	"github.com/labstack/echo/v4"
)

func (c *Controller) LoanTransactionEntryController() {
	req := c.provider.Service.Request

	// POST /api/v1/loan-transaction/:loan_transaction_id/deduction
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/loan-transaction-entry/loan-transaction/:loan_transaction_id/deduction",
		Method:       "POST",
		Note:         "Adds a deduction to a loan transaction by ID.",
		RequestType:  model.LoanTransactionDeductionRequest{},
		ResponseType: model.LoanTransaction{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		loanTransactionID, err := handlers.EngineUUIDParam(ctx, "loan_transaction_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid loan transaction ID"})
		}
		var req model.LoanTransactionDeductionRequest
		if err := ctx.Bind(&req); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Loan transaction deduction failed: invalid payload: " + err.Error(),
				Module:      "LoanTransaction",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid loan transaction deduction payload: " + err.Error()})
		}
		if err := c.provider.Service.Validator.Struct(req); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Loan transaction deduction failed: validation error: " + err.Error(),
				Module:      "LoanTransaction",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		account, err := c.model.AccountManager.GetByID(context, req.AccountID)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "not-found",
				Description: "Account not found for loan transaction deduction: " + err.Error(),
				Module:      "LoanTransaction",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Account not found: " + err.Error()})
		}

		loanTransaction := &model.LoanTransactionEntry{
			CreatedByID:       userOrg.UserID,
			UpdatedByID:       userOrg.UserID,
			CreatedAt:         time.Now().UTC(),
			UpdatedAt:         time.Now().UTC(),
			OrganizationID:    userOrg.OrganizationID,
			BranchID:          *userOrg.BranchID,
			LoanTransactionID: *loanTransactionID,
			Type:              model.LoanTransactionAutomaticDeduction,
			Debit:             0,
			Credit:            req.Amount,
			IsAddOn:           req.IsAddOn,
			AccountID:         &req.AccountID,
			Name:              account.Name,
			Description:       account.Description,
		}
		if err := c.model.LoanTransactionEntryManager.Create(context, loanTransaction); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Loan transaction deduction creation failed: " + err.Error(),
				Module:      "LoanTransaction",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create loan transaction deduction: " + err.Error()})
		}

		tx := c.provider.Service.Database.Client().Begin()
		if tx.Error != nil {
			tx.Rollback()
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Failed to start database transaction: " + tx.Error.Error(),
				Module:      "LoanTransaction",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to start database transaction: " + tx.Error.Error()})
		}
		newLoanTransaction, err := c.event.LoanBalancing(context, ctx, tx, event.LoanBalanceEvent{
			LoanTransactionID: *loanTransactionID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": fmt.Sprintf("Failed to balance loan transaction: %v", err)})
		}
		return ctx.JSON(http.StatusOK, newLoanTransaction)
	})

	// PUT /api/v1/loan-transaction/deduction/:loan_transaction_entry_id
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/loan-transaction-entry/:loan_transaction_entry_id/deduction",
		Method:       "PUT",
		Note:         "Adds a deduction to a loan transaction by ID.",
		RequestType:  model.LoanTransactionDeductionRequest{},
		ResponseType: model.LoanTransaction{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		loanTransactionEntryId, err := handlers.EngineUUIDParam(ctx, "loan_transaction_entry_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid loan transaction ID"})
		}
		var req model.LoanTransactionDeductionRequest
		if err := ctx.Bind(&req); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Loan transaction deduction failed: invalid payload: " + err.Error(),
				Module:      "LoanTransaction",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid loan transaction deduction payload: " + err.Error()})
		}
		if err := c.provider.Service.Validator.Struct(req); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Loan transaction deduction failed: validation error: " + err.Error(),
				Module:      "LoanTransaction",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		account, err := c.model.AccountManager.GetByID(context, req.AccountID)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "not-found",
				Description: "Account not found for loan transaction deduction: " + err.Error(),
				Module:      "LoanTransaction",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Account not found: " + err.Error()})
		}
		loanTransactionEntry, err := c.model.LoanTransactionEntryManager.GetByID(context, *loanTransactionEntryId)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "not-found",
				Description: "Loan transaction entry not found for deduction update: " + err.Error(),
				Module:      "LoanTransaction",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Loan transaction entry not found for deduction update: " + err.Error()})
		}
		loanTransactionEntry.Credit = req.Amount
		loanTransactionEntry.IsAddOn = req.IsAddOn
		loanTransactionEntry.AccountID = &req.AccountID
		loanTransactionEntry.Name = account.Name
		loanTransactionEntry.UpdatedAt = time.Now().UTC()
		loanTransactionEntry.UpdatedByID = userOrg.UserID
		if err := c.model.LoanTransactionEntryManager.UpdateFields(context, *loanTransactionEntryId, loanTransactionEntry); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Loan transaction deduction creation failed: " + err.Error(),
				Module:      "LoanTransaction",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create loan transaction deduction: " + err.Error()})
		}

		newLoanTransaction, err := c.model.LoanTransactionManager.GetByIDRaw(context, loanTransactionEntry.LoanTransactionID)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "not-found",
				Description: "Loan transaction not found after deduction creation: " + err.Error(),
				Module:      "LoanTransaction",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Loan transaction not found after deduction creation: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, newLoanTransaction)
	})

	// DELETE /api/v1/loan-transaction-entry/:loan_transaction_entry_id
	req.RegisterRoute(handlers.Route{
		Route:  "/api/v1/loan-transaction-entry/:loan_transaction_entry_id",
		Method: "DELETE",
		Note:   "Deletes a loan transaction entry by ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		loanTransactionEntryID, err := handlers.EngineUUIDParam(ctx, "loan_transaction_entry_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid loan transaction entry ID"})
		}

		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != model.UserOrganizationTypeOwner && userOrg.UserType != model.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to delete loan transaction entries"})
		}

		loanTransactionEntry, err := c.model.LoanTransactionEntryManager.GetByID(context, *loanTransactionEntryID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Loan transaction entry not found"})
		}

		// Check if the loan transaction entry belongs to the user's organization and branch
		if loanTransactionEntry.OrganizationID != userOrg.OrganizationID || loanTransactionEntry.BranchID != *userOrg.BranchID {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Access denied to this loan transaction entry"})
		}

		// Set deleted by user
		loanTransactionEntry.DeletedByID = &userOrg.UserID

		if err := c.model.LoanTransactionEntryManager.Delete(context, loanTransactionEntry); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete loan transaction entry: " + err.Error()})
		}

		return ctx.JSON(http.StatusOK, map[string]string{"message": "Loan transaction entry deleted successfully"})
	})
}
