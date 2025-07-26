package controller

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/lands-horizon/horizon-server/services/horizon"
	"github.com/lands-horizon/horizon-server/src/model"
)

func (c *Controller) TransactionEntryController() {
	req := c.provider.Service.Request
	req.RegisterRoute(horizon.Route{
		Route:    "/transaction/payment/:transaction_id",
		Method:   "POST",
		Request:  "PaymentOnlineRequest",
		Response: "GeneralLedgerResponse",
		Note:     "Creates an online payment entry for the specified transaction.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		transactionID, err := horizon.EngineUUIDParam(ctx, "transaction_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid transaction ID: " + err.Error()})
		}
		var req model.PaymentRequest
		if err := ctx.Bind(&req); err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body: " + err.Error()})
		}
		transactionBatch, err := c.model.TransactionBatchCurrent(context, userOrg.UserID, userOrg.OrganizationID, *userOrg.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Failed to retrieve transaction batch: " + err.Error()})
		}
		transaction, err := c.model.TransactionEntryManager.GetByID(context, *transactionID)
		if err != nil {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Failed to retrieve transaction: " + err.Error()})
		}
		generalLedger, err := c.model.GeneralLedgerCurrentMemberAccount(context, *req.MemberProfileID, *req.AccountID, userOrg.OrganizationID, *userOrg.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Failed to retrieve general ledger: " + err.Error()})
		}
		account, err := c.model.AccountManager.GetByID(context, *req.AccountID)
		if err != nil {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Failed to retrieve account: " + err.Error()})
		}
		paymentType, err := c.model.PaymentTypeManager.GetByID(context, *req.PaymentTypeID)
		if err != nil {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Failed to retrieve payment type: " + err.Error()})
		}
		var credit, debit, newBalance float64
		switch account.Type {
		case model.AccountTypeDeposit:
			credit = req.Amount
			debit = 0
			newBalance = generalLedger.Balance + req.Amount
		case model.AccountTypeLoan:
			credit = 0
			debit = req.Amount
			newBalance = generalLedger.Balance - req.Amount
		case model.AccountTypeARLedger, model.AccountTypeARAging, model.AccountTypeFines, model.AccountTypeInterest, model.AccountTypeSVFLedger, model.AccountTypeWOff, model.AccountTypeAPLedger:
			credit = 0
			debit = req.Amount
			newBalance = generalLedger.Balance - req.Amount
		case model.AccountTypeOther:
			credit = req.Amount
			debit = 0
			newBalance = generalLedger.Balance + req.Amount
		default:
			credit = req.Amount
			debit = 0
			newBalance = generalLedger.Balance + req.Amount
		}
		ledgerReq := &model.GeneralLedger{
			CreatedAt:                  time.Now().UTC(),
			CreatedByID:                userOrg.UserID,
			UpdatedAt:                  time.Now().UTC(),
			UpdatedByID:                userOrg.UserID,
			BranchID:                   *userOrg.BranchID,
			OrganizationID:             userOrg.OrganizationID,
			TransactionBatchID:         &transactionBatch.ID,
			ReferenceNumber:            req.ReferenceNumber,
			TransactionID:              &transaction.ID,
			EntryDate:                  req.EntryDate,
			SignatureMediaID:           req.SignatureMediaID,
			ProofOfPaymentMediaID:      req.ProofOfPaymentMediaID,
			BankID:                     req.BankID,
			AccountID:                  &account.ID,
			Credit:                     credit,
			Debit:                      debit,
			Balance:                    newBalance,
			MemberProfileID:            req.MemberProfileID,
			MemberJointAccountID:       req.MemberJointAccountID,
			PaymentTypeID:              &paymentType.ID,
			TransactionReferenceNumber: transaction.ReferenceNumber,
			Source:                     model.GeneralLedgerSourcePayment,
			BankReferenceNumber:        req.BankReferenceNumber,
			EmployeeUserID:             &userOrg.UserID,
		}
		if err := c.model.GeneralLedgerManager.Create(context, ledgerReq); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create online payment entry: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.GeneralLedgerManager.ToModel(ledgerReq))
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/transaction-entry/transaction-batch/:transaction_batch_id/search",
		Method:   "GET",
		Response: "TransactionEntry[]",
		Note:     "Returns paginated transaction entries for the specified transaction batch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		transactionBatchID, err := horizon.EngineUUIDParam(ctx, "transaction_batch_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid transaction batch ID: " + err.Error()})
		}
		transaction, err := c.model.TransactionEntryManager.Find(context, &model.TransactionEntry{
			TransactionBatchID: transactionBatchID,
			OrganizationID:     userOrg.OrganizationID,
			BranchID:           *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve transaction entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.TransactionEntryManager.Pagination(context, ctx, transaction))
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/transaction-entry/transaction/:transaction_id",
		Method:   "GET",
		Response: "TransactionEntry[]",
		Note:     "Returns all transaction entries for the specified transaction ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		transactionID, err := horizon.EngineUUIDParam(ctx, "transaction_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid transaction ID: " + err.Error()})
		}
		entries, err := c.model.TransactionEntryManager.Find(context, &model.TransactionEntry{
			TransactionID:  transactionID,
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve transaction entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.TransactionEntryManager.Filtered(context, ctx, entries))
	})
}
