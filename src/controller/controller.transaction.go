package controller

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/lands-horizon/horizon-server/services/horizon"
	"github.com/lands-horizon/horizon-server/src/event"
	"github.com/lands-horizon/horizon-server/src/model"
)

func (c *Controller) TransactionController() {
	req := c.provider.Service.Request
	req.RegisterRoute(horizon.Route{
		Route:        "/transaction/deposit/:transaction_id",
		Method:       "POST",
		RequestType:  model.PaymentRequest{},
		ResponseType: model.GeneralLedgerResponse{},
		Note:         "Registers an online deposit against a specific transaction, updating both the general ledger and transaction record.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "auth-error",
				Description: "Failed to get user organization (/transaction/deposit/:transaction_id): " + err.Error(),
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		transactionID, err := horizon.EngineUUIDParam(ctx, "transaction_id")
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "param-error",
				Description: "Invalid transaction ID (/transaction/deposit/:transaction_id): " + err.Error(),
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid transaction ID: " + err.Error()})
		}
		var req model.PaymentRequest
		if err := ctx.Bind(&req); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bind-error",
				Description: "Invalid request body (/transaction/deposit/:transaction_id): " + err.Error(),
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body: " + err.Error()})
		}
		transactionBatch, err := c.model.TransactionBatchCurrent(context, userOrg.UserID, userOrg.OrganizationID, *userOrg.BranchID)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "batch-error",
				Description: "Failed to retrieve transaction batch (/transaction/deposit/:transaction_id): " + err.Error(),
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Failed to retrieve transaction batch: " + err.Error()})
		}
		transaction, err := c.model.TransactionManager.GetByID(context, *transactionID)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "transaction-error",
				Description: "Failed to retrieve transaction (/transaction/deposit/:transaction_id): " + err.Error(),
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Failed to retrieve transaction: " + err.Error()})
		}
		account, err := c.model.AccountManager.GetByID(context, *req.AccountID)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "account-error",
				Description: "Failed to retrieve account (/transaction/deposit/:transaction_id): " + err.Error(),
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Failed to retrieve account: " + err.Error()})
		}
		paymentType, err := c.model.PaymentTypeManager.GetByID(context, *req.PaymentTypeID)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "payment-type-error",
				Description: "Failed to retrieve payment type (/transaction/deposit/:transaction_id): " + err.Error(),
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Failed to retrieve payment type: " + err.Error()})
		}

		// Transaction block starts here
		tx := c.provider.Service.Database.Client().Begin()
		if tx.Error != nil {
			tx.Rollback()
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "db-error",
				Description: "Failed to start transaction (/transaction/deposit/:transaction_id): " + tx.Error.Error(),
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to start transaction: " + tx.Error.Error()})
		}
		defer func() {
			if r := recover(); r != nil {
				tx.Rollback()
			}
		}()

		// >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>
		// Use FOR UPDATE to lock the latest general ledger entry for this account/member/org/branch
		generalLedger, err := c.model.GeneralLedgerCurrentMemberAccountForUpdate(
			context, tx, *req.MemberProfileID, *req.AccountID, userOrg.OrganizationID, *userOrg.BranchID)
		if err != nil {
			tx.Rollback()
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "ledger-error",
				Description: "Failed to retrieve general ledger (FOR UPDATE) (/transaction/deposit/:transaction_id): " + err.Error(),
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Failed to retrieve general ledger: " + err.Error()})
		}
		// <<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<

		var credit, debit, newBalance float64
		switch account.Type {
		case model.AccountTypeDeposit:
			credit = req.Amount
			debit = 0
			if generalLedger != nil {
				newBalance = generalLedger.Balance + req.Amount
			} else {
				newBalance = req.Amount
			}
		case model.AccountTypeLoan:
			credit = 0
			debit = req.Amount
			if generalLedger != nil {
				newBalance = generalLedger.Balance - req.Amount
			} else {
				newBalance = -req.Amount
			}
		case model.AccountTypeARLedger, model.AccountTypeARAging, model.AccountTypeFines, model.AccountTypeInterest, model.AccountTypeSVFLedger, model.AccountTypeWOff, model.AccountTypeAPLedger:
			credit = 0
			debit = req.Amount
			if generalLedger != nil {
				newBalance = generalLedger.Balance - req.Amount
			} else {
				newBalance = -req.Amount
			}
		case model.AccountTypeOther:
			credit = req.Amount
			debit = 0
			if generalLedger != nil {
				newBalance = generalLedger.Balance + req.Amount
			} else {
				newBalance = req.Amount
			}
		default:
			credit = req.Amount
			debit = 0
			if generalLedger != nil {
				newBalance = generalLedger.Balance + req.Amount
			} else {
				newBalance = req.Amount
			}
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
			Description:                req.Description,
		}
		if err := c.model.GeneralLedgerManager.CreateWithTx(context, tx, ledgerReq); err != nil {
			tx.Rollback()
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "ledger-create-error",
				Description: "Failed to create online payment entry (/transaction/deposit/:transaction_id): " + err.Error(),
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create online payment entry: " + err.Error()})
		}
		transaction.Amount += req.Amount
		if err := c.model.TransactionManager.UpdateFields(context, transaction.ID, transaction); err != nil {
			tx.Rollback()
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "transaction-update-error",
				Description: "Transaction update failed (/transaction/deposit/:transaction_id): " + err.Error(),
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update transaction: " + err.Error()})
		}
		if err := tx.Commit().Error; err != nil {
			tx.Rollback()
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "commit-error",
				Description: "Failed to commit transaction (/transaction/deposit/:transaction_id): " + err.Error(),
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to commit transaction: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "deposit-success",
			Description: "Online deposit successfully registered (/transaction/deposit/:transaction_id), transaction_id: " + transaction.ID.String(),
			Module:      "Transaction",
		})
		return ctx.JSON(http.StatusOK, c.model.GeneralLedgerManager.ToModel(ledgerReq))
	})

	req.RegisterRoute(horizon.Route{
		Route:        "/transaction/withdraw/:transaction_id",
		Method:       "POST",
		RequestType:  model.PaymentRequest{},
		ResponseType: model.GeneralLedgerResponse{},
		Note:         "Processes an online withdrawal for the given transaction, updating the general ledger and transaction amounts.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "auth-error",
				Description: "Failed to get user organization (/transaction/withdraw/:transaction_id): " + err.Error(),
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		transactionID, err := horizon.EngineUUIDParam(ctx, "transaction_id")
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "param-error",
				Description: "Invalid transaction ID (/transaction/withdraw/:transaction_id): " + err.Error(),
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid transaction ID: " + err.Error()})
		}
		var req model.PaymentRequest
		if err := ctx.Bind(&req); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bind-error",
				Description: "Invalid request body (/transaction/withdraw/:transaction_id): " + err.Error(),
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body: " + err.Error()})
		}
		transactionBatch, err := c.model.TransactionBatchCurrent(context, userOrg.UserID, userOrg.OrganizationID, *userOrg.BranchID)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "batch-error",
				Description: "Failed to retrieve transaction batch (/transaction/withdraw/:transaction_id): " + err.Error(),
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Failed to retrieve transaction batch: " + err.Error()})
		}
		transaction, err := c.model.TransactionManager.GetByID(context, *transactionID)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "transaction-error",
				Description: "Failed to retrieve transaction (/transaction/withdraw/:transaction_id): " + err.Error(),
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Failed to retrieve transaction: " + err.Error()})
		}
		account, err := c.model.AccountManager.GetByID(context, *req.AccountID)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "account-error",
				Description: "Failed to retrieve account (/transaction/withdraw/:transaction_id): " + err.Error(),
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Failed to retrieve account: " + err.Error()})
		}
		paymentType, err := c.model.PaymentTypeManager.GetByID(context, *req.PaymentTypeID)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "payment-type-error",
				Description: "Failed to retrieve payment type (/transaction/withdraw/:transaction_id): " + err.Error(),
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Failed to retrieve payment type: " + err.Error()})
		}

		// Begin DB transaction for race condition protection
		tx := c.provider.Service.Database.Client().Begin()
		if tx.Error != nil {
			tx.Rollback()
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "db-error",
				Description: "Withdraw failed: begin tx error (/transaction/withdraw/:transaction_id): " + tx.Error.Error(),
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to start transaction: " + tx.Error.Error()})
		}
		defer func() {
			if r := recover(); r != nil {
				tx.Rollback()
			}
		}()

		// >>>>>>>>>>>>>>>> RACE CONDITION PROTECTION <<<<<<<<<<<<<<<<
		// Lock the latest general ledger entry for update, so concurrent withdrawals can't double-spend.
		generalLedger, err := c.model.GeneralLedgerCurrentMemberAccountForUpdate(
			context, tx, *req.MemberProfileID, *req.AccountID, userOrg.OrganizationID, *userOrg.BranchID)
		if err != nil {
			tx.Rollback()
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "ledger-error",
				Description: "Failed to retrieve general ledger (FOR UPDATE) (/transaction/withdraw/:transaction_id): " + err.Error(),
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Failed to retrieve general ledger: " + err.Error()})
		}
		// <<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<

		var credit, debit, newBalance float64
		switch account.Type {
		case model.AccountTypeDeposit:
			credit = 0
			debit = req.Amount
			if generalLedger != nil {
				newBalance = generalLedger.Balance - req.Amount
			} else {
				newBalance = -req.Amount
			}
		case model.AccountTypeLoan:
			credit = req.Amount
			debit = 0
			if generalLedger != nil {
				newBalance = generalLedger.Balance + req.Amount
			} else {
				newBalance = req.Amount
			}
		case model.AccountTypeARLedger, model.AccountTypeARAging, model.AccountTypeFines, model.AccountTypeInterest, model.AccountTypeSVFLedger, model.AccountTypeWOff, model.AccountTypeAPLedger:
			credit = req.Amount
			debit = 0
			if generalLedger != nil {
				newBalance = generalLedger.Balance + req.Amount
			} else {
				newBalance = req.Amount
			}
		case model.AccountTypeOther:
			credit = 0
			debit = req.Amount
			if generalLedger != nil {
				newBalance = generalLedger.Balance - req.Amount
			} else {
				newBalance = -req.Amount
			}
		default:
			credit = 0
			debit = req.Amount
			if generalLedger != nil {
				newBalance = generalLedger.Balance - req.Amount
			} else {
				newBalance = -req.Amount
			}
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
			Source:                     model.GeneralLedgerSourceWithdraw,
			BankReferenceNumber:        req.BankReferenceNumber,
			EmployeeUserID:             &userOrg.UserID,
			Description:                req.Description,
		}
		if err := c.model.GeneralLedgerManager.CreateWithTx(context, tx, ledgerReq); err != nil {
			tx.Rollback()
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "ledger-create-error",
				Description: "Failed to create withdraw entry (/transaction/withdraw/:transaction_id): " + err.Error(),
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create withdraw entry: " + err.Error()})
		}
		transaction.Amount -= req.Amount
		if err := c.model.TransactionManager.UpdateFields(context, transaction.ID, transaction); err != nil {
			tx.Rollback()
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "transaction-update-error",
				Description: "Transaction update failed (/transaction/withdraw/:transaction_id): " + err.Error(),
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update transaction: " + err.Error()})
		}
		if err := tx.Commit().Error; err != nil {
			tx.Rollback()
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "commit-error",
				Description: "Withdraw failed: commit tx error (/transaction/withdraw/:transaction_id): " + err.Error(),
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to commit transaction: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "withdraw-success",
			Description: "Online withdrawal successfully processed (/transaction/withdraw/:transaction_id), transaction_id: " + transaction.ID.String(),
			Module:      "Transaction",
		})
		return ctx.JSON(http.StatusOK, c.model.GeneralLedgerManager.ToModel(ledgerReq))
	})

	// Create transaction
	req.RegisterRoute(horizon.Route{
		Route:        "/transaction",
		Method:       "POST",
		RequestType:  model.TransactionRequest{},
		ResponseType: model.TransactionResponse{},
		Note:         "Creates a new transaction record with provided details, allowing subsequent deposit or withdrawal actions.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "auth-error",
				Description: "Failed to get user organization (/transaction): " + err.Error(),
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		var req model.TransactionRequest
		if err := ctx.Bind(&req); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bind-error",
				Description: "Invalid request body (/transaction): " + err.Error(),
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body: " + err.Error()})
		}
		if err := ctx.Validate(&req); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "validate-error",
				Description: "Validation failed (/transaction): " + err.Error(),
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		transactionBatch, err := c.model.TransactionBatchCurrent(context, userOrg.UserID, userOrg.OrganizationID, *userOrg.BranchID)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "batch-error",
				Description: "Failed to retrieve transaction batch (/transaction): " + err.Error(),
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Failed to retrieve transaction batch: " + err.Error()})
		}
		transaction := &model.Transaction{
			CreatedAt:            time.Now().UTC(),
			CreatedByID:          userOrg.UserID,
			UpdatedAt:            time.Now().UTC(),
			UpdatedByID:          userOrg.UserID,
			BranchID:             *userOrg.BranchID,
			OrganizationID:       userOrg.OrganizationID,
			SignatureMediaID:     req.SignatureMediaID,
			TransactionBatchID:   &transactionBatch.ID,
			EmployeeUserID:       &userOrg.UserID,
			MemberProfileID:      req.MemberProfileID,
			MemberJointAccountID: req.MemberJointAccountID,
			LoanBalance:          0,
			LoanDue:              0,
			TotalDue:             0,
			FinesDue:             0,
			TotalLoan:            0,
			InterestDue:          0,
			Amount:               0,
			ReferenceNumber:      req.ReferenceNumber,
			Source:               req.Source,
			Description:          req.Description,
		}
		if err := c.model.TransactionManager.Create(context, transaction); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Transaction creation failed (/transaction), db error: " + err.Error(),
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create transaction: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "create-success",
			Description: "Transaction created successfully (/transaction), transaction_id: " + transaction.ID.String(),
			Module:      "Transaction",
		})
		return ctx.JSON(http.StatusCreated, c.model.TransactionManager.ToModel(transaction))
	})

	req.RegisterRoute(horizon.Route{
		Route:        "/transaction/deposit",
		Method:       "POST",
		RequestType:  model.PaymentQuickRequest{},
		ResponseType: model.GeneralLedgerResponse{},
		Note:         "Performs a quick deposit operation using minimal information, creating both a transaction and related ledger entry.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "auth-error",
				Description: "Failed to get user organization (/transaction/deposit): " + err.Error(),
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		var req model.PaymentQuickRequest
		if err := ctx.Bind(&req); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bind-error",
				Description: "Invalid request body (/transaction/deposit): " + err.Error(),
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body: " + err.Error()})
		}
		transactionBatch, err := c.model.TransactionBatchCurrent(context, userOrg.UserID, userOrg.OrganizationID, *userOrg.BranchID)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "batch-error",
				Description: "Failed to retrieve transaction batch (/transaction/deposit): " + err.Error(),
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Failed to retrieve transaction batch: " + err.Error()})
		}
		account, err := c.model.AccountManager.GetByID(context, *req.AccountID)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "account-error",
				Description: "Failed to retrieve account (/transaction/deposit): " + err.Error(),
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Failed to retrieve account: " + err.Error()})
		}
		paymentType, err := c.model.PaymentTypeManager.GetByID(context, *req.PaymentTypeID)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "payment-type-error",
				Description: "Failed to retrieve payment type (/transaction/deposit): " + err.Error(),
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Failed to retrieve payment type: " + err.Error()})
		}

		// Begin transaction with race condition protection
		tx := c.provider.Service.Database.Client().Begin()
		if tx.Error != nil {
			tx.Rollback()
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "db-error",
				Description: "Failed to start transaction (/transaction/deposit): " + tx.Error.Error(),
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to start transaction: " + tx.Error.Error()})
		}
		defer func() {
			if r := recover(); r != nil {
				tx.Rollback()
			}
		}()

		// >>>>>>>>>>>>>>>> RACE CONDITION PROTECTION <<<<<<<<<<<<<<<<
		// Lock the latest general ledger entry for update, so concurrent deposits can't double-spend.
		generalLedger, err := c.model.GeneralLedgerCurrentMemberAccountForUpdate(
			context, tx, *req.MemberProfileID, *req.AccountID, userOrg.OrganizationID, *userOrg.BranchID)
		if err != nil {
			tx.Rollback()
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "ledger-error",
				Description: "Failed to retrieve general ledger (FOR UPDATE) (/transaction/deposit): " + err.Error(),
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Failed to retrieve general ledger: " + err.Error()})
		}
		// <<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<

		var credit, debit, newBalance float64
		switch account.Type {
		case model.AccountTypeDeposit:
			credit = req.Amount
			debit = 0
			if generalLedger != nil {
				newBalance = generalLedger.Balance + req.Amount
			} else {
				newBalance = req.Amount
			}
		case model.AccountTypeLoan:
			credit = 0
			debit = req.Amount
			if generalLedger != nil {
				newBalance = generalLedger.Balance - req.Amount
			} else {
				newBalance = -req.Amount
			}
		case model.AccountTypeARLedger, model.AccountTypeARAging, model.AccountTypeFines, model.AccountTypeInterest, model.AccountTypeSVFLedger, model.AccountTypeWOff, model.AccountTypeAPLedger:
			credit = 0
			debit = req.Amount
			if generalLedger != nil {
				newBalance = generalLedger.Balance - req.Amount
			} else {
				newBalance = -req.Amount
			}
		case model.AccountTypeOther:
			credit = req.Amount
			debit = 0
			if generalLedger != nil {
				newBalance = generalLedger.Balance + req.Amount
			} else {
				newBalance = req.Amount
			}
		default:
			credit = req.Amount
			debit = 0
			if generalLedger != nil {
				newBalance = generalLedger.Balance + req.Amount
			} else {
				newBalance = req.Amount
			}
		}
		transaction := &model.Transaction{
			CreatedAt:            time.Now().UTC(),
			CreatedByID:          userOrg.UserID,
			UpdatedAt:            time.Now().UTC(),
			UpdatedByID:          userOrg.UserID,
			BranchID:             *userOrg.BranchID,
			OrganizationID:       userOrg.OrganizationID,
			SignatureMediaID:     req.SignatureMediaID,
			TransactionBatchID:   &transactionBatch.ID,
			EmployeeUserID:       &userOrg.UserID,
			MemberProfileID:      req.MemberProfileID,
			MemberJointAccountID: req.MemberJointAccountID,
			Amount:               req.Amount,
			ReferenceNumber:      req.ReferenceNumber,
			Source:               model.GeneralLedgerSourceDeposit,
			Description:          req.Description,
			LoanBalance:          0,
			LoanDue:              0,
			TotalDue:             0,
			FinesDue:             0,
			TotalLoan:            0,
			InterestDue:          0,
		}
		if err := c.model.TransactionManager.CreateWithTx(context, tx, transaction); err != nil {
			tx.Rollback()
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "transaction-create-error",
				Description: "Failed to create transaction (/transaction/deposit): " + err.Error(),
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create transaction: " + err.Error()})
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
			TransactionReferenceNumber: req.ReferenceNumber,
			Source:                     model.GeneralLedgerSourceDeposit,
			BankReferenceNumber:        req.BankReferenceNumber,
			EmployeeUserID:             &userOrg.UserID,
			Description:                req.Description,
		}
		if err := c.model.GeneralLedgerManager.CreateWithTx(context, tx, ledgerReq); err != nil {
			tx.Rollback()
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "ledger-create-error",
				Description: "Failed to create quick deposit entry (/transaction/deposit): " + err.Error(),
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create quick deposit entry: " + err.Error()})
		}
		if err := tx.Commit().Error; err != nil {
			tx.Rollback()
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "commit-error",
				Description: "Failed to commit transaction (/transaction/deposit): " + err.Error(),
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to commit transaction: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "deposit-success",
			Description: "Quick deposit entry successfully created (/transaction/deposit), transaction_id: " + transaction.ID.String(),
			Module:      "Transaction",
		})
		return ctx.JSON(http.StatusOK, c.model.GeneralLedgerManager.ToModel(ledgerReq))
	})

	req.RegisterRoute(horizon.Route{
		Route:        "/transaction/withdraw",
		Method:       "POST",
		RequestType:  model.PaymentQuickRequest{},
		ResponseType: model.GeneralLedgerResponse{},
		Note:         "Executes a quick withdrawal with minimal required info, generating both a transaction and related ledger entry.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "auth-error",
				Description: "Failed to get user organization (/transaction/withdraw): " + err.Error(),
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		var req model.PaymentQuickRequest
		if err := ctx.Bind(&req); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bind-error",
				Description: "Invalid request body (/transaction/withdraw): " + err.Error(),
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body: " + err.Error()})
		}
		transactionBatch, err := c.model.TransactionBatchCurrent(context, userOrg.UserID, userOrg.OrganizationID, *userOrg.BranchID)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "batch-error",
				Description: "Failed to retrieve transaction batch (/transaction/withdraw): " + err.Error(),
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Failed to retrieve transaction batch: " + err.Error()})
		}

		account, err := c.model.AccountManager.GetByID(context, *req.AccountID)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "account-error",
				Description: "Failed to retrieve account (/transaction/withdraw): " + err.Error(),
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Failed to retrieve account: " + err.Error()})
		}
		paymentType, err := c.model.PaymentTypeManager.GetByID(context, *req.PaymentTypeID)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "payment-type-error",
				Description: "Failed to retrieve payment type (/transaction/withdraw): " + err.Error(),
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Failed to retrieve payment type: " + err.Error()})
		}

		// Begin transaction with race condition protection
		tx := c.provider.Service.Database.Client().Begin()
		if tx.Error != nil {
			tx.Rollback()
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "db-error",
				Description: "Failed to start transaction (/transaction/withdraw): " + tx.Error.Error(),
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to start transaction: " + tx.Error.Error()})
		}
		defer func() {
			if r := recover(); r != nil {
				tx.Rollback()
			}
		}()

		// >>>>>>>>>>>>>>>> RACE CONDITION PROTECTION <<<<<<<<<<<<<<<<
		// Lock the latest general ledger entry for update, so concurrent withdrawals can't double-spend.
		generalLedger, err := c.model.GeneralLedgerCurrentMemberAccountForUpdate(
			context, tx, *req.MemberProfileID, *req.AccountID, userOrg.OrganizationID, *userOrg.BranchID)
		if err != nil {
			tx.Rollback()
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "ledger-error",
				Description: "Failed to retrieve general ledger (FOR UPDATE) (/transaction/withdraw): " + err.Error(),
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Failed to retrieve general ledger: " + err.Error()})
		}
		// <<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<

		// Withdraw logic
		var credit, debit, newBalance float64
		switch account.Type {
		case model.AccountTypeDeposit:
			credit = 0
			debit = req.Amount
			if generalLedger != nil {
				newBalance = generalLedger.Balance - req.Amount
			} else {
				newBalance = -req.Amount
			}
		case model.AccountTypeLoan:
			credit = req.Amount
			debit = 0
			if generalLedger != nil {
				newBalance = generalLedger.Balance + req.Amount
			} else {
				newBalance = req.Amount
			}
		case model.AccountTypeARLedger, model.AccountTypeARAging, model.AccountTypeFines, model.AccountTypeInterest, model.AccountTypeSVFLedger, model.AccountTypeWOff, model.AccountTypeAPLedger:
			credit = req.Amount
			debit = 0
			if generalLedger != nil {
				newBalance = generalLedger.Balance + req.Amount
			} else {
				newBalance = req.Amount
			}
		case model.AccountTypeOther:
			credit = 0
			debit = req.Amount
			if generalLedger != nil {
				newBalance = generalLedger.Balance - req.Amount
			} else {
				newBalance = -req.Amount
			}
		default:
			credit = 0
			debit = req.Amount
			if generalLedger != nil {
				newBalance = generalLedger.Balance - req.Amount
			} else {
				newBalance = -req.Amount
			}
		}
		transaction := &model.Transaction{
			CreatedAt:            time.Now().UTC(),
			CreatedByID:          userOrg.UserID,
			UpdatedAt:            time.Now().UTC(),
			UpdatedByID:          userOrg.UserID,
			BranchID:             *userOrg.BranchID,
			OrganizationID:       userOrg.OrganizationID,
			SignatureMediaID:     req.SignatureMediaID,
			TransactionBatchID:   &transactionBatch.ID,
			EmployeeUserID:       &userOrg.UserID,
			MemberProfileID:      req.MemberProfileID,
			MemberJointAccountID: req.MemberJointAccountID,
			Amount:               req.Amount,
			ReferenceNumber:      req.ReferenceNumber,
			Source:               model.GeneralLedgerSourceWithdraw,
			Description:          req.Description,
			LoanBalance:          0,
			LoanDue:              0,
			TotalDue:             0,
			FinesDue:             0,
			TotalLoan:            0,
			InterestDue:          0,
		}
		if err := c.model.TransactionManager.CreateWithTx(context, tx, transaction); err != nil {
			tx.Rollback()
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "transaction-create-error",
				Description: "Failed to create transaction (/transaction/withdraw): " + err.Error(),
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create transaction: " + err.Error()})
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
			TransactionReferenceNumber: req.ReferenceNumber,
			Source:                     model.GeneralLedgerSourceWithdraw,
			BankReferenceNumber:        req.BankReferenceNumber,
			EmployeeUserID:             &userOrg.UserID,
			Description:                req.Description,
		}
		if err := c.model.GeneralLedgerManager.CreateWithTx(context, tx, ledgerReq); err != nil {
			tx.Rollback()
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "ledger-create-error",
				Description: "Failed to create quick withdraw entry (/transaction/withdraw): " + err.Error(),
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create quick withdraw entry: " + err.Error()})
		}
		if err := tx.Commit().Error; err != nil {
			tx.Rollback()
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "commit-error",
				Description: "Failed to commit transaction (/transaction/withdraw): " + err.Error(),
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to commit transaction: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "withdraw-success",
			Description: "Quick withdraw entry successfully created (/transaction/withdraw), transaction_id: " + transaction.ID.String(),
			Module:      "Transaction",
		})
		return ctx.JSON(http.StatusOK, c.model.GeneralLedgerManager.ToModel(ledgerReq))
	})

	req.RegisterRoute(horizon.Route{
		Route:        "/transaction/:transaction_id",
		Method:       "PUT",
		RequestType:  model.TransactionRequestEdit{},
		ResponseType: model.TransactionResponse{},
		Note:         "Modifies the description of an existing transaction, allowing updates to its memo or comment field.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "auth-error",
				Description: "Failed to get user organization (/transaction/:transaction_id): " + err.Error(),
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		transactionID, err := horizon.EngineUUIDParam(ctx, "transaction_id")
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "param-error",
				Description: "Invalid transaction ID (/transaction/:transaction_id): " + err.Error(),
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid transaction ID: " + err.Error()})
		}
		var req model.TransactionRequestEdit
		if err := ctx.Bind(&req); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bind-error",
				Description: "Invalid request body (/transaction/:transaction_id): " + err.Error(),
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body: " + err.Error()})
		}

		// Begin transaction for row-level locking
		tx := c.provider.Service.Database.Client().Begin()
		if tx.Error != nil {
			tx.Rollback()
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "db-error",
				Description: "Failed to start transaction (/transaction/:transaction_id): " + tx.Error.Error(),
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to start transaction: " + tx.Error.Error()})
		}
		defer func() {
			if r := recover(); r != nil {
				tx.Rollback()
			}
		}()

		transaction, err := c.model.TransactionManager.GetByID(context, *transactionID)
		if err != nil {
			tx.Rollback()
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "not-found-error",
				Description: "Transaction not found or lock failed (/transaction/:transaction_id): " + err.Error(),
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Transaction not found: " + err.Error()})
		}
		transaction.Description = req.Description
		transaction.UpdatedAt = time.Now().UTC()
		transaction.UpdatedByID = userOrg.UserID
		if err := c.model.TransactionManager.UpdateFieldsWithTx(context, tx, transaction.ID, transaction); err != nil {
			tx.Rollback()
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Failed to update transaction (/transaction/:transaction_id): " + err.Error(),
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update transaction: " + err.Error()})
		}
		if err := tx.Commit().Error; err != nil {
			tx.Rollback()
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "commit-error",
				Description: "Failed to commit transaction (/transaction/:transaction_id): " + err.Error(),
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to commit transaction: " + err.Error()})
		}
		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Transaction description updated successfully (/transaction/:transaction_id), transaction_id: " + transaction.ID.String(),
			Module:      "Transaction",
		})
		return ctx.JSON(http.StatusOK, c.model.TransactionManager.ToModel(transaction))
	})

	req.RegisterRoute(horizon.Route{
		Route:        "/transaction/:transaction_id",
		Method:       "GET",
		ResponseType: model.TransactionResponse{},
		Note:         "Retrieves detailed information for the specified transaction by its unique identifier.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Access denied"})
		}
		transactionID, err := horizon.EngineUUIDParam(ctx, "transaction_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid transaction ID: " + err.Error()})
		}
		transaction, err := c.model.TransactionManager.GetByID(context, *transactionID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Transaction not found: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.TransactionManager.ToModel(transaction))
	})

	req.RegisterRoute(horizon.Route{
		Route:        "/transaction/current/search",
		Method:       "GET",
		ResponseType: model.TransactionResponse{},
		Note:         "Lists all transactions associated with the currently authenticated user within their organization and branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Access denied"})
		}
		transactions, err := c.model.TransactionManager.Find(context, &model.Transaction{
			EmployeeUserID: &userOrg.UserID,
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve transactions: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.TransactionManager.Filtered(context, ctx, transactions))
	})

	req.RegisterRoute(horizon.Route{
		Route:        "/transaction/employee/:employee_id/search",
		Method:       "GET",
		ResponseType: model.TransactionResponse{},
		Note:         "Fetches all transactions handled by the specified employee, filtered by organization and branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Access denied"})
		}
		employeeID, err := horizon.EngineUUIDParam(ctx, "employee_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid employee ID: " + err.Error()})
		}
		employee, err := c.model.UserManager.GetByID(context, *employeeID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Employee not found: " + err.Error()})
		}
		transactions, err := c.model.TransactionManager.Find(context, &model.Transaction{
			EmployeeUserID: &employee.ID,
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve transactions: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.TransactionManager.Filtered(context, ctx, transactions))
	})

	req.RegisterRoute(horizon.Route{
		Route:        "/transaction/member-profile/:member_profile_id/search",
		Method:       "GET",
		ResponseType: model.TransactionResponse{},
		Note:         "Retrieves all transactions related to the given member profile within the user's organization and branch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Access denied"})
		}
		memberProfileID, err := horizon.EngineUUIDParam(ctx, "member_profile_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid member profile ID: " + err.Error()})
		}
		memberProfile, err := c.model.MemberProfileManager.GetByID(context, *memberProfileID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Member profile not found: " + err.Error()})
		}
		transactions, err := c.model.TransactionManager.Find(context, &model.Transaction{
			MemberProfileID: &memberProfile.ID,
			OrganizationID:  userOrg.OrganizationID,
			BranchID:        *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve transactions: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.TransactionManager.Filtered(context, ctx, transactions))
	})

	req.RegisterRoute(horizon.Route{
		Route:        "/transaction/branch/search",
		Method:       "GET",
		ResponseType: model.TransactionResponse{},
		Note:         "Provides a paginated list of all transactions recorded for the current branch of the user's organization.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		if userOrg.UserType != "owner" && userOrg.UserType != "employee" {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Access denied"})
		}
		transactions, err := c.model.TransactionManager.Find(context, &model.Transaction{
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve branch transactions: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.TransactionManager.Pagination(context, ctx, transactions))
	})

}
