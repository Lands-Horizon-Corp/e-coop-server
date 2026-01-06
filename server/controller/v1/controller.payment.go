package v1

import (
	"fmt"
	"net/http"

	"github.com/Lands-Horizon-Corp/e-coop-server/server/event"
	"github.com/Lands-Horizon-Corp/e-coop-server/server/model/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/services/handlers"
	"github.com/labstack/echo/v4"
)

func (c *Controller) paymentController() {
	req := c.provider.Service.Request
	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/transaction/:transaction_id/multipayment",
		Method:       "POST",
		Note:         "Processes multiple payments for the specified transaction by transaction_id and records them in the general ledger.",
		ResponseType: core.GeneralLedgerResponse{},
		RequestType:  core.PaymentRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()

		transactionID, err := handlers.EngineUUIDParam(ctx, "transaction_id")
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "multipayment-param-error",
				Description: fmt.Sprintf("Invalid transaction id for POST /transaction/:transaction_id/multipayment: %v", err),
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid transaction ID: " + err.Error()})
		}

		transaction, err := c.core.TransactionManager().GetByID(context, *transactionID)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "multipayment-transaction-not-found",
				Description: fmt.Sprintf("Transaction not found for ID %v: %v", transactionID, err),
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Transaction not found: " + err.Error()})
		}
		var req []core.PaymentRequest
		if err := ctx.Bind(&req); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "multipayment-bind-error",
				Description: "Multiple payment failed: invalid payload: " + err.Error(),
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid multipayment payload: " + err.Error()})
		}

		if len(req) == 0 {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "multipayment-empty-error",
				Description: "Multiple payment failed: no payment entries provided",
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "No payment entries provided"})
		}

		for i, payment := range req {
			if err := c.provider.Service.Validator.Struct(payment); err != nil {
				c.event.Footstep(ctx, event.FootstepEvent{
					Activity:    "multipayment-validation-error",
					Description: fmt.Sprintf("Multiple payment failed: validation error for payment %d: %v", i+1, err),
					Module:      "Transaction",
				})
				return ctx.JSON(http.StatusBadRequest, map[string]string{"error": fmt.Sprintf("Payment %d validation failed: %v", i+1, err.Error())})
			}
		}

		var generalLedgers []*core.GeneralLedger

		for i, payment := range req {
			tx, endTx := c.provider.Service.Database.StartTransaction(context)
			if tx.Error != nil {
				c.event.Footstep(ctx, event.FootstepEvent{
					Activity:    "multipayment-db-error",
					Description: "Multiple payment failed (/transaction/:transaction_id/multipayment), begin tx error: " + tx.Error.Error(),
					Module:      "Transaction",
				})
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to start database transaction: " + endTx(tx.Error).Error()})
			}

			generalLedger, err := c.event.TransactionPayment(context, ctx, tx, endTx, event.TransactionEvent{
				TransactionID:        &transaction.ID,
				MemberProfileID:      transaction.MemberProfileID,
				MemberJointAccountID: transaction.MemberJointAccountID,
				ReferenceNumber:      transaction.ReferenceNumber,

				Source:                core.GeneralLedgerSourcePayment,
				Amount:                payment.Amount,
				AccountID:             payment.AccountID,
				PaymentTypeID:         payment.PaymentTypeID,
				SignatureMediaID:      payment.SignatureMediaID,
				EntryDate:             payment.EntryDate,
				BankID:                payment.BankID,
				ProofOfPaymentMediaID: payment.ProofOfPaymentMediaID,
				Description:           payment.Description,
				BankReferenceNumber:   payment.BankReferenceNumber,

				LoanTransactionID: payment.LoanTransactionID,
			})
			if err != nil {
				c.event.Footstep(ctx, event.FootstepEvent{
					Activity:    "multipayment-error",
					Description: fmt.Sprintf("Multiple payment processing failed for payment %d: %v", i+1, err),
					Module:      "Transaction",
				})
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": fmt.Sprintf("Payment %d processing failed: %v", i+1, err.Error())})
			}

			if err := endTx(nil); err != nil {
				c.event.Footstep(ctx, event.FootstepEvent{
					Activity:    "multipayment-commit-error",
					Description: fmt.Sprintf("Multiple payment commit failed for payment %d: %v", i+1, err),
					Module:      "Transaction",
				})
			}

			generalLedgers = append(generalLedgers, generalLedger)
		}

		var response []core.GeneralLedgerResponse
		for _, gl := range generalLedgers {
			response = append(response, *c.core.GeneralLedgerManager().ToModel(gl))
		}

		return ctx.JSON(http.StatusOK, response)
	})

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/transaction/general-ledger/:general_ledger_id/print",
		Method:       "POST",
		Note:         "Processes print number for the specified general ledger by general_ledger_id.",
		ResponseType: core.GeneralLedgerResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		generalLedgerID, err := handlers.EngineUUIDParam(ctx, "general_ledger_id")
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "payment-general-ledger-param-error",
				Description: fmt.Sprintf("Invalid general ledger id for POST /transaction/general-ledger/:general_ledger_id/payment: %v", err),
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid general ledger ID: " + err.Error()})
		}
		userOrg, err := c.token.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}
		generalLedger, err := c.core.GeneralLedgerManager().GetByID(context, *generalLedgerID)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "payment-general-ledger-not-found",
				Description: fmt.Sprintf("General ledger not found for ID %v: %v", generalLedgerID, err),
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "General ledger not found: " + err.Error()})
		}
		maxNumber, err := c.core.GeneralLedgerPrintMaxNumber(context, *generalLedger.MemberProfileID, *generalLedger.AccountID, *userOrg.BranchID, userOrg.OrganizationID)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "payment-general-ledger-max-number-error",
				Description: fmt.Sprintf("Failed to get max print number for general ledger ID %v: %v", generalLedgerID, err),
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get max print number: " + err.Error()})
		}

		generalLedger.PrintNumber = maxNumber + 1
		if err := c.core.GeneralLedgerManager().UpdateByID(context, generalLedger.ID, generalLedger); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Connect account to FS definition failed (/financial-statement-definition/:financial_statement_definition_id/account/:account_id/connect), account db error: " + err.Error(),
				Module:      "FinancialStatement",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to connect account: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.core.GeneralLedgerManager().ToModel(generalLedger))
	})

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/transaction/:transaction_id/payment",
		Method:       "POST",
		Note:         "Processes a payment for the specified transaction by transaction_id and records it in the general ledger.",
		ResponseType: core.GeneralLedgerResponse{},
		RequestType:  core.PaymentRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()

		var req core.PaymentRequest
		if err := ctx.Bind(&req); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "payment-bind-error",
				Description: "Payment failed: invalid payload: " + err.Error(),
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid payment payload: " + err.Error()})
		}

		if err := c.provider.Service.Validator.Struct(req); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "payment-validation-error",
				Description: "Payment failed: validation error: " + err.Error(),
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Payment validation failed: " + err.Error()})
		}

		tx, endTx := c.provider.Service.Database.StartTransaction(context)
		if tx.Error != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "payment-db-error",
				Description: "Payment failed (/transaction/:transaction_id/payment), begin tx error: " + tx.Error.Error(),
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to start database transaction: " + endTx(tx.Error).Error()})
		}

		transactionID, err := handlers.EngineUUIDParam(ctx, "transaction_id")
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "payment-param-error",
				Description: fmt.Sprintf("Invalid transaction id for POST /transaction/:transaction_id/payment: %v", err),
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid transaction ID: " + err.Error()})
		}

		generalLedger, err := c.event.TransactionPayment(context, ctx, tx, endTx, event.TransactionEvent{
			TransactionID:        transactionID,
			MemberProfileID:      nil,
			MemberJointAccountID: nil,
			ReferenceNumber:      "",

			Source:                core.GeneralLedgerSourcePayment,
			Amount:                req.Amount,
			AccountID:             req.AccountID,
			PaymentTypeID:         req.PaymentTypeID,
			SignatureMediaID:      req.SignatureMediaID,
			EntryDate:             req.EntryDate,
			BankID:                req.BankID,
			ProofOfPaymentMediaID: req.ProofOfPaymentMediaID,
			Description:           req.Description,
			BankReferenceNumber:   req.BankReferenceNumber,

			LoanTransactionID: req.LoanTransactionID,
		})
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "payment-error",
				Description: "Payment processing failed: " + err.Error(),
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Payment processing failed: " + err.Error()})
		}

		response := c.core.GeneralLedgerManager().ToModel(generalLedger)

		return ctx.JSON(http.StatusOK, response)
	})

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/transaction/:transaction_id/withdraw",
		Method:       "POST",
		Note:         "Processes a withdrawal for the specified transaction by transaction_id and updates the general ledger accordingly.",
		ResponseType: core.GeneralLedgerResponse{},
		RequestType:  core.PaymentRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var req core.PaymentRequest
		if err := ctx.Bind(&req); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "withdraw-bind-error",
				Description: "Withdrawal failed: invalid payload: " + err.Error(),
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid withdrawal payload: " + err.Error()})
		}
		if err := c.provider.Service.Validator.Struct(req); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "withdraw-validation-error",
				Description: "Withdrawal failed: validation error: " + err.Error(),
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Withdrawal validation failed: " + err.Error()})
		}

		tx, endTx := c.provider.Service.Database.StartTransaction(context)
		if tx.Error != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "withdraw-db-error",
				Description: "Withdrawal failed (/transaction/:transaction_id/withdraw), begin tx error: " + tx.Error.Error(),
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to start database transaction: " + endTx(tx.Error).Error()})
		}

		transactionID, err := handlers.EngineUUIDParam(ctx, "transaction_id")
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "withdraw-param-error",
				Description: fmt.Sprintf("Invalid transaction id for POST /transaction/:transaction_id/withdraw: %v", err),
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid transaction ID: " + err.Error()})
		}

		generalLedger, err := c.event.TransactionPayment(context, ctx, tx, endTx, event.TransactionEvent{
			TransactionID:        transactionID,
			MemberProfileID:      nil,
			MemberJointAccountID: nil,
			ReferenceNumber:      "",

			Source:                core.GeneralLedgerSourceWithdraw,
			Amount:                req.Amount,
			AccountID:             req.AccountID,
			PaymentTypeID:         req.PaymentTypeID,
			SignatureMediaID:      req.SignatureMediaID,
			EntryDate:             req.EntryDate,
			BankID:                req.BankID,
			ProofOfPaymentMediaID: req.ProofOfPaymentMediaID,
			Description:           req.Description,
			BankReferenceNumber:   req.BankReferenceNumber,
			LoanTransactionID:     req.LoanTransactionID,
		})
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "withdraw-error",
				Description: "Withdrawal processing failed: " + err.Error(),
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Withdrawal processing failed: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.core.GeneralLedgerManager().ToModel(generalLedger))
	})

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/transaction/:transaction_id/deposit",
		Method:       "POST",
		Note:         "Processes a deposit for the specified transaction by transaction_id and updates the general ledger accordingly.",
		ResponseType: core.GeneralLedgerResponse{},
		RequestType:  core.PaymentRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var req core.PaymentRequest
		if err := ctx.Bind(&req); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "deposit-bind-error",
				Description: "Deposit failed: invalid payload: " + err.Error(),
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid deposit payload: " + err.Error()})
		}
		if err := c.provider.Service.Validator.Struct(req); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "deposit-validation-error",
				Description: "Deposit failed: validation error: " + err.Error(),
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Deposit validation failed: " + err.Error()})
		}

		tx, endTx := c.provider.Service.Database.StartTransaction(context)
		if tx.Error != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "deposit-db-error",
				Description: "Deposit failed (/transaction/:transaction_id/deposit), begin tx error: " + tx.Error.Error(),
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to start database transaction: " + endTx(tx.Error).Error()})
		}

		transactionID, err := handlers.EngineUUIDParam(ctx, "transaction_id")
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "deposit-param-error",
				Description: fmt.Sprintf("Invalid transaction id for POST /transaction/:transaction_id/deposit: %v", err),
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid transaction ID: " + err.Error()})
		}

		generalLedger, err := c.event.TransactionPayment(context, ctx, tx, endTx, event.TransactionEvent{
			TransactionID:        transactionID,
			MemberProfileID:      nil,
			MemberJointAccountID: nil,
			ReferenceNumber:      "",

			Source:                core.GeneralLedgerSourceDeposit,
			Amount:                req.Amount,
			AccountID:             req.AccountID,
			PaymentTypeID:         req.PaymentTypeID,
			SignatureMediaID:      req.SignatureMediaID,
			EntryDate:             req.EntryDate,
			BankID:                req.BankID,
			ProofOfPaymentMediaID: req.ProofOfPaymentMediaID,
			Description:           req.Description,
			BankReferenceNumber:   req.BankReferenceNumber,
			LoanTransactionID:     req.LoanTransactionID,
		})
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "deposit-error",
				Description: "Deposit processing failed: " + err.Error(),
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Deposit processing failed: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.core.GeneralLedgerManager().ToModel(generalLedger))
	})

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/transaction/payment",
		Method:       "POST",
		Note:         "Processes a payment for a transaction without specifying transaction_id in the route. Used for general payments.",
		ResponseType: core.GeneralLedger{},
		RequestType:  core.PaymentQuickRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var req core.PaymentQuickRequest
		if err := ctx.Bind(&req); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "general-payment-bind-error",
				Description: "General payment failed: invalid payload: " + err.Error(),
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid payment payload: " + err.Error()})
		}
		if err := c.provider.Service.Validator.Struct(req); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "general-payment-validation-error",
				Description: "General payment failed: validation error: " + err.Error(),
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Payment validation failed: " + err.Error()})
		}

		tx, endTx := c.provider.Service.Database.StartTransaction(context)
		if tx.Error != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "general-payment-db-error",
				Description: "General payment failed (/transaction/payment), begin tx error: " + tx.Error.Error(),
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to start database transaction: " + endTx(tx.Error).Error()})
		}

		generalLedger, err := c.event.TransactionPayment(context, ctx, tx, endTx, event.TransactionEvent{
			TransactionID:        nil,
			MemberProfileID:      req.MemberProfileID,
			MemberJointAccountID: req.MemberJointAccountID,
			ReferenceNumber:      req.BankReferenceNumber,

			Source:                core.GeneralLedgerSourcePayment,
			Amount:                req.Amount,
			AccountID:             req.AccountID,
			PaymentTypeID:         req.PaymentTypeID,
			SignatureMediaID:      req.SignatureMediaID,
			EntryDate:             req.EntryDate,
			BankID:                req.BankID,
			ProofOfPaymentMediaID: req.ProofOfPaymentMediaID,
			Description:           req.Description,
			BankReferenceNumber:   req.BankReferenceNumber,
			ORAutoGenerated:       req.ORAutoGenerated,
			LoanTransactionID:     req.LoanTransactionID,
		})
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "general-payment-error",
				Description: "General payment processing failed: " + err.Error(),
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Payment processing failed: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.core.GeneralLedgerManager().ToModel(generalLedger))
	})

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/transaction/withdraw",
		Method:       "POST",
		Note:         "Processes a withdrawal for a transaction without specifying transaction_id in the route. Used for general withdrawals.",
		ResponseType: core.GeneralLedgerResponse{},
		RequestType:  core.PaymentQuickRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var req core.PaymentQuickRequest
		if err := ctx.Bind(&req); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "general-withdraw-bind-error",
				Description: "General withdrawal failed: invalid payload: " + err.Error(),
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid withdrawal payload: " + err.Error()})
		}
		if err := c.provider.Service.Validator.Struct(req); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "general-withdraw-validation-error",
				Description: "General withdrawal failed: validation error: " + err.Error(),
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Withdrawal validation failed: " + err.Error()})
		}

		tx, endTx := c.provider.Service.Database.StartTransaction(context)
		if tx.Error != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "general-withdraw-db-error",
				Description: "General withdrawal failed (/transaction/withdraw), begin tx error: " + tx.Error.Error(),
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to start database transaction: " + endTx(tx.Error).Error()})
		}

		generalLedger, err := c.event.TransactionPayment(context, ctx, tx, endTx, event.TransactionEvent{
			TransactionID:        nil,
			MemberProfileID:      req.MemberProfileID,
			MemberJointAccountID: req.MemberJointAccountID,
			ReferenceNumber:      req.BankReferenceNumber,

			Source:                core.GeneralLedgerSourceWithdraw,
			Amount:                req.Amount,
			AccountID:             req.AccountID,
			PaymentTypeID:         req.PaymentTypeID,
			SignatureMediaID:      req.SignatureMediaID,
			EntryDate:             req.EntryDate,
			BankID:                req.BankID,
			ProofOfPaymentMediaID: req.ProofOfPaymentMediaID,
			Description:           req.Description,
			BankReferenceNumber:   req.BankReferenceNumber,
			ORAutoGenerated:       req.ORAutoGenerated,
			LoanTransactionID:     req.LoanTransactionID,
		})
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "general-withdraw-error",
				Description: "General withdrawal processing failed: " + err.Error(),
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Withdrawal processing failed: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.core.GeneralLedgerManager().ToModel(generalLedger))
	})

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/transaction/deposit",
		Method:       "POST",
		Note:         "Processes a deposit for a transaction without specifying transaction_id in the route. Used for general deposits.",
		ResponseType: core.GeneralLedgerResponse{},
		RequestType:  core.PaymentQuickRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var req core.PaymentQuickRequest
		if err := ctx.Bind(&req); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "general-deposit-bind-error",
				Description: "General deposit failed: invalid payload: " + err.Error(),
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid deposit payload: " + err.Error()})
		}
		if err := c.provider.Service.Validator.Struct(req); err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "general-deposit-validation-error",
				Description: "General deposit failed: validation error: " + err.Error(),
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Deposit validation failed: " + err.Error()})
		}

		tx, endTx := c.provider.Service.Database.StartTransaction(context)
		if tx.Error != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "general-deposit-db-error",
				Description: "General deposit failed (/transaction/deposit), begin tx error: " + tx.Error.Error(),
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to start database transaction: " + endTx(tx.Error).Error()})
		}

		generalLedger, err := c.event.TransactionPayment(context, ctx, tx, endTx, event.TransactionEvent{
			TransactionID:        nil,
			MemberProfileID:      req.MemberProfileID,
			MemberJointAccountID: req.MemberJointAccountID,
			ReferenceNumber:      req.BankReferenceNumber,

			Source:                core.GeneralLedgerSourceDeposit,
			Amount:                req.Amount,
			AccountID:             req.AccountID,
			PaymentTypeID:         req.PaymentTypeID,
			SignatureMediaID:      req.SignatureMediaID,
			EntryDate:             req.EntryDate,
			BankID:                req.BankID,
			ProofOfPaymentMediaID: req.ProofOfPaymentMediaID,
			Description:           req.Description,
			BankReferenceNumber:   req.BankReferenceNumber,
			ORAutoGenerated:       req.ORAutoGenerated,
			LoanTransactionID:     req.LoanTransactionID,
		})
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "general-deposit-error",
				Description: "General deposit processing failed: " + err.Error(),
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Deposit processing failed: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.core.GeneralLedgerManager().ToModel(generalLedger))
	})

	req.RegisterWebRoute(handlers.Route{
		Route:  "/api/v1/transaction/general-ledger/:general_ledger_id/reverse",
		Method: "POST",
		Note:   "Reverses a specific general ledger transaction.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		generalLedgerID, err := handlers.EngineUUIDParam(ctx, "general_ledger_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid general ledger ID: " + err.Error()})
		}
		generalLedger, err := c.core.GeneralLedgerManager().GetByID(context, *generalLedgerID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "General ledger not found: " + err.Error()})
		}
		tx, endTx := c.provider.Service.Database.StartTransaction(context)
		if tx.Error != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "payment-db-error",
				Description: "Payment failed (/transaction/:transaction_id/payment), begin tx error: " + tx.Error.Error(),
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to start database transaction: " + endTx(tx.Error).Error()})
		}
		amount := 0.0
		switch {
		case generalLedger.Credit > 0:
			amount = generalLedger.Credit
		case generalLedger.Debit > 0:
			amount = -generalLedger.Debit
		default:
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "General ledger entry is neither debit nor credit"})
		}
		newGeneralLedger, err := c.event.TransactionPayment(context, ctx, tx, endTx, event.TransactionEvent{
			Amount:                amount,
			AccountID:             generalLedger.AccountID,
			PaymentTypeID:         generalLedger.PaymentTypeID,
			TransactionID:         generalLedger.TransactionID,
			MemberProfileID:       generalLedger.MemberProfileID,
			SignatureMediaID:      generalLedger.SignatureMediaID,
			MemberJointAccountID:  generalLedger.MemberJointAccountID,
			Source:                generalLedger.Source,
			EntryDate:             &generalLedger.EntryDate,
			BankID:                generalLedger.BankID,
			ProofOfPaymentMediaID: generalLedger.ProofOfPaymentMediaID,
			ReferenceNumber:       generalLedger.ReferenceNumber,
			Description:           "REVERSAL: " + generalLedger.Description,
			BankReferenceNumber:   generalLedger.BankReferenceNumber,
			ORAutoGenerated:       false,
			Reverse:               true,
			LoanTransactionID:     generalLedger.LoanTransactionID,
		})
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "payment-error",
				Description: "Payment processing failed: " + err.Error(),
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Payment processing failed: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.core.GeneralLedgerManager().ToModel(newGeneralLedger))
	})

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/transaction/:transaction_id/reverse",
		Method:       "POST",
		Note:         "Reverses all general ledger entries for a specific transaction by transaction_id.",
		ResponseType: core.TransactionResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		transactionID, err := handlers.EngineUUIDParam(ctx, "transaction_id")
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "transaction-reverse-param-error",
				Description: fmt.Sprintf("Invalid transaction id for POST /general-ledger/transaction/:transaction_id/reverse: %v", err),
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid transaction ID: " + err.Error()})
		}

		generalLedgers, err := c.core.GeneralLedgerManager().Find(context, &core.GeneralLedger{
			TransactionID: transactionID,
		})
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "transaction-reverse-ledger-error",
				Description: fmt.Sprintf("Failed to get general ledger entries for transaction ID %v: %v", transactionID, err),
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get general ledger entries: " + err.Error()})
		}

		if len(generalLedgers) == 0 {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "No general ledger entries found for this transaction"})
		}
		tx, endTx := c.provider.Service.Database.StartTransaction(context)
		if tx.Error != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "transaction-reverse-db-error",
				Description: "Transaction reverse failed, begin tx error: " + tx.Error.Error(),
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to start database transaction: " + endTx(tx.Error).Error()})
		}

		var reversedLedgers []*core.GeneralLedger

		for _, generalLedger := range generalLedgers {
			amount := 0.0
			switch {
			case generalLedger.Credit > 0:
				amount = generalLedger.Credit
			case generalLedger.Debit > 0:
				amount = generalLedger.Debit
			default:
				continue
			}
			newGeneralLedger, err := c.event.TransactionPayment(context, ctx, tx, endTx, event.TransactionEvent{
				Amount:                amount,
				AccountID:             generalLedger.AccountID,
				PaymentTypeID:         generalLedger.PaymentTypeID,
				TransactionID:         generalLedger.TransactionID,
				MemberProfileID:       generalLedger.MemberProfileID,
				SignatureMediaID:      generalLedger.SignatureMediaID,
				MemberJointAccountID:  generalLedger.MemberJointAccountID,
				Source:                generalLedger.Source,
				EntryDate:             &generalLedger.EntryDate,
				BankID:                generalLedger.BankID,
				ProofOfPaymentMediaID: generalLedger.ProofOfPaymentMediaID,
				ReferenceNumber:       generalLedger.ReferenceNumber,
				Description:           "REVERSAL: " + generalLedger.Description,
				BankReferenceNumber:   generalLedger.BankReferenceNumber,
				ORAutoGenerated:       false,
				Reverse:               true,
				LoanTransactionID:     generalLedger.LoanTransactionID,
			})
			if err != nil {
				c.event.Footstep(ctx, event.FootstepEvent{
					Activity:    "transaction-reverse-error",
					Description: fmt.Sprintf("Transaction reversal failed for ledger %v: %v", generalLedger.ID, err),
					Module:      "Transaction",
				})
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Transaction reversal failed: " + err.Error()})
			}
			reversedLedgers = append(reversedLedgers, newGeneralLedger)
		}
		transaction, err := c.core.TransactionManager().GetByIDRaw(context, *transactionID)
		if err != nil {
			c.event.Footstep(ctx, event.FootstepEvent{
				Activity:    "transaction-reverse-fetch-error",
				Description: fmt.Sprintf("Failed to fetch transaction %v after reversal: %v", transactionID, err),
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch transaction after reversal: " + err.Error()})
		}
		c.event.Footstep(ctx, event.FootstepEvent{
			Activity:    "transaction-reverse-success",
			Description: fmt.Sprintf("Successfully reversed transaction %v with %d general ledger entries", transactionID, len(reversedLedgers)),
			Module:      "Transaction",
		})

		return ctx.JSON(http.StatusOK, transaction)
	})

}
