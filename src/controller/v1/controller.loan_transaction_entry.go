package controller_v1

import (
	"net/http"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/services/handlers"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/model"
	"github.com/labstack/echo/v4"
)

func (c *Controller) LoanTransactionEntryController() {
	req := c.provider.Service.Request

	// GET /api/v1/loan-transaction-entry/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/loan-transaction-entry/search",
		Method:       "GET",
		ResponseType: model.LoanTransactionEntryResponse{},
		Note:         "Returns all loan transaction entries for the current user's branch with pagination and filtering.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != model.UserOrganizationTypeOwner && userOrg.UserType != model.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view loan transaction entries"})
		}

		loanTransactionEntries, err := c.model.LoanTransactionEntryCurrentBranch(context, userOrg.OrganizationID, *userOrg.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve loan transaction entries: " + err.Error()})
		}

		return ctx.JSON(http.StatusOK, c.model.LoanTransactionEntryManager.Pagination(context, ctx, loanTransactionEntries))
	})

	// GET /api/v1/loan-transaction-entry/:loan_transaction_entry_id
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/loan-transaction-entry/:loan_transaction_entry_id",
		Method:       "GET",
		ResponseType: model.LoanTransactionEntryResponse{},
		Note:         "Returns a specific loan transaction entry by ID.",
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
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to view loan transaction entries"})
		}

		loanTransactionEntry, err := c.model.LoanTransactionEntryManager.GetByID(context, *loanTransactionEntryID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Loan transaction entry not found"})
		}

		// Check if the loan transaction entry belongs to the user's organization and branch
		if loanTransactionEntry.OrganizationID != userOrg.OrganizationID || loanTransactionEntry.BranchID != *userOrg.BranchID {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Access denied to this loan transaction entry"})
		}

		return ctx.JSON(http.StatusOK, c.model.LoanTransactionEntryManager.ToModel(loanTransactionEntry))
	})

	// POST /api/v1/loan-transaction-entry
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/loan-transaction-entry",
		Method:       "POST",
		ResponseType: model.LoanTransactionEntryResponse{},
		Note:         "Creates a new loan transaction entry.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != model.UserOrganizationTypeOwner && userOrg.UserType != model.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to create loan transaction entries"})
		}

		request, err := c.model.LoanTransactionEntryManager.Validate(ctx)
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}

		// Verify that the loan transaction exists and belongs to the user's organization and branch
		loanTransaction, err := c.model.LoanTransactionManager.GetByID(context, request.LoanTransactionID)
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid loan transaction ID"})
		}
		if loanTransaction.OrganizationID != userOrg.OrganizationID || loanTransaction.BranchID != *userOrg.BranchID {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Access denied to this loan transaction"})
		}

		loanTransactionEntry := &model.LoanTransactionEntry{
			CreatedByID: userOrg.UserID,
			UpdatedByID: userOrg.UserID,
			CreatedAt:   time.Now().UTC(),
			UpdatedAt:   time.Now().UTC(),

			OrganizationID:    userOrg.OrganizationID,
			BranchID:          *userOrg.BranchID,
			LoanTransactionID: request.LoanTransactionID,
			AccountID:         request.AccountID,
			Description:       request.Description,
			Credit:            request.Credit,
			Debit:             request.Debit,
			Type:              request.Type,
			IsAddOn:           request.IsAddOn,
			Name:              request.Name,
		}

		if err := c.model.LoanTransactionEntryManager.Create(context, loanTransactionEntry); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create loan transaction entry: " + err.Error()})
		}

		return ctx.JSON(http.StatusCreated, c.model.LoanTransactionEntryManager.ToModel(loanTransactionEntry))
	})

	// POST /api/v1/loan-transaction-entry/loan_tra/:loan_transaction_id
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/loan-transaction-entry/loan-transaction/:loan_transaction_id",
		Method:       "POST",
		ResponseType: model.LoanTransactionEntryResponse{},
		Note:         "Creates a new loan transaction entry for a specific loan transaction.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		loanTransactionID, err := handlers.EngineUUIDParam(ctx, "loan_transaction_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid loan transaction ID"})
		}

		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		if userOrg.UserType != model.UserOrganizationTypeOwner && userOrg.UserType != model.UserOrganizationTypeEmployee {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to create loan transaction entries"})
		}

		// Verify that the loan transaction exists and belongs to the user's organization and branch
		loanTransaction, err := c.model.LoanTransactionManager.GetByID(context, *loanTransactionID)
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Loan transaction not found"})
		}
		if loanTransaction.OrganizationID != userOrg.OrganizationID || loanTransaction.BranchID != *userOrg.BranchID {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Access denied to this loan transaction"})
		}

		request, err := c.model.LoanTransactionEntryManager.Validate(ctx)
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}

		loanTransactionEntry := &model.LoanTransactionEntry{
			CreatedByID:       userOrg.UserID,
			UpdatedByID:       userOrg.UserID,
			OrganizationID:    userOrg.OrganizationID,
			BranchID:          *userOrg.BranchID,
			LoanTransactionID: *loanTransactionID,
			AccountID:         request.AccountID,
			Description:       request.Description,
			Credit:            request.Credit,
			Debit:             request.Debit,
			Type:              request.Type,
			IsAddOn:           request.IsAddOn,
			CreatedAt:         time.Now().UTC(),
			UpdatedAt:         time.Now().UTC(),

			Name: request.Name,
		}

		if err := c.model.LoanTransactionEntryManager.Create(context, loanTransactionEntry); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create loan transaction entry: " + err.Error()})
		}

		return ctx.JSON(http.StatusCreated, c.model.LoanTransactionEntryManager.ToModel(loanTransactionEntry))
	})

	// PUT /api/v1/loan-transaction-entry/:loan_transaction_entry_id
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/loan-transaction-entry/:loan_transaction_entry_id",
		Method:       "PUT",
		ResponseType: model.LoanTransactionEntryResponse{},
		Note:         "Updates an existing loan transaction entry.",
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
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "User is not authorized to update loan transaction entries"})
		}

		request, err := c.model.LoanTransactionEntryManager.Validate(ctx)
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}

		loanTransactionEntry, err := c.model.LoanTransactionEntryManager.GetByID(context, *loanTransactionEntryID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Loan transaction entry not found"})
		}

		// Check if the loan transaction entry belongs to the user's organization and branch
		if loanTransactionEntry.OrganizationID != userOrg.OrganizationID || loanTransactionEntry.BranchID != *userOrg.BranchID {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Access denied to this loan transaction entry"})
		}

		// Verify that the new loan transaction exists and belongs to the user's organization and branch
		if request.LoanTransactionID != loanTransactionEntry.LoanTransactionID {
			loanTransaction, err := c.model.LoanTransactionManager.GetByID(context, request.LoanTransactionID)
			if err != nil {
				return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid loan transaction ID"})
			}
			if loanTransaction.OrganizationID != userOrg.OrganizationID || loanTransaction.BranchID != *userOrg.BranchID {
				return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Access denied to this loan transaction"})
			}
		}

		// Update fields
		loanTransactionEntry.UpdatedAt = time.Now().UTC()
		loanTransactionEntry.UpdatedByID = userOrg.UserID
		loanTransactionEntry.LoanTransactionID = request.LoanTransactionID
		loanTransactionEntry.AccountID = request.AccountID
		loanTransactionEntry.Description = request.Description
		loanTransactionEntry.Credit = request.Credit
		loanTransactionEntry.Debit = request.Debit
		loanTransactionEntry.Type = request.Type
		loanTransactionEntry.IsAddOn = request.IsAddOn
		loanTransactionEntry.Name = request.Name

		if err := c.model.LoanTransactionEntryManager.Update(context, loanTransactionEntry); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update loan transaction entry: " + err.Error()})
		}

		return ctx.JSON(http.StatusOK, c.model.LoanTransactionEntryManager.ToModel(loanTransactionEntry))
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
