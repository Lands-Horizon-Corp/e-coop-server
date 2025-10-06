package controller_v1

import (
	"net/http"

	"github.com/Lands-Horizon-Corp/e-coop-server/services/handlers"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/model"
	"github.com/labstack/echo/v4"
)

func (c *Controller) LoanTransactionEntryController() {
	req := c.provider.Service.Request

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
