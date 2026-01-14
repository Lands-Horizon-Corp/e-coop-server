package transactions

import (
	"fmt"
	"net/http"

	"github.com/Lands-Horizon-Corp/e-coop-server/helpers"
	"github.com/Lands-Horizon-Corp/e-coop-server/horizon"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/event"
	"github.com/labstack/echo/v4"
)

func PaymentController(service *horizon.HorizonService) {
	req := service.API
	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/transaction/:transaction_id/multipayment",
		Method:       "POST",
		Note:         "Processes multiple payments for the specified transaction by transaction_id and records them in the general ledger.",
		ResponseType: core.GeneralLedgerResponse{},
		RequestType:  core.PaymentRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()

		transactionID, err := helpers.EngineUUIDParam(ctx, "transaction_id")
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "multipayment-param-error",
				Description: fmt.Sprintf("Invalid transaction id for POST /transaction/:transaction_id/multipayment: %v", err),
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid transaction ID: " + err.Error()})
		}

		transaction, err := core.TransactionManager(service).GetByID(context, *transactionID)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "multipayment-transaction-not-found",
				Description: fmt.Sprintf("Transaction not found for ID %v: %v", transactionID, err),
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Transaction not found: " + err.Error()})
		}
		var req []core.PaymentRequest
		if err := ctx.Bind(&req); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "multipayment-bind-error",
				Description: "Multiple payment failed: invalid payload: " + err.Error(),
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid multipayment payload: " + err.Error()})
		}

		if len(req) == 0 {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "multipayment-empty-error",
				Description: "Multiple payment failed: no payment entries provided",
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "No payment entries provided"})
		}

		for i, payment := range req {
			if err := service.Validator.Struct(payment); err != nil {
				event.Footstep(ctx, service, event.FootstepEvent{
					Activity:    "multipayment-validation-error",
					Description: fmt.Sprintf("Multiple payment failed: validation error for payment %d: %v", i+1, err),
					Module:      "Transaction",
				})
				return ctx.JSON(http.StatusBadRequest, map[string]string{"error": fmt.Sprintf("Payment %d validation failed: %v", i+1, err.Error())})
			}
		}

		var generalLedgers []*core.GeneralLedger

		for i, payment := range req {
			tx, endTx := service.Database.StartTransaction(context)
			if tx.Error != nil {
				event.Footstep(ctx, service, event.FootstepEvent{
					Activity:    "multipayment-db-error",
					Description: "Multiple payment failed (/transaction/:transaction_id/multipayment), begin tx error: " + tx.Error.Error(),
					Module:      "Transaction",
				})
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to start database transaction: " + endTx(tx.Error).Error()})
			}

			generalLedger, err := event.TransactionPayment(context, service, ctx, tx, endTx, event.TransactionEvent{
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
				event.Footstep(ctx, service, event.FootstepEvent{
					Activity:    "multipayment-error",
					Description: fmt.Sprintf("Multiple payment processing failed for payment %d: %v", i+1, err),
					Module:      "Transaction",
				})
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": fmt.Sprintf("Payment %d processing failed: %v", i+1, err.Error())})
			}

			if err := endTx(nil); err != nil {
				event.Footstep(ctx, service, event.FootstepEvent{
					Activity:    "multipayment-commit-error",
					Description: fmt.Sprintf("Multiple payment commit failed for payment %d: %v", i+1, err),
					Module:      "Transaction",
				})
			}

			generalLedgers = append(generalLedgers, generalLedger)
		}

		var response []core.GeneralLedgerResponse
		for _, gl := range generalLedgers {
			response = append(response, *core.GeneralLedgerManager(service).ToModel(gl))
		}

		return ctx.JSON(http.StatusOK, response)
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/transaction/general-ledger/:general_ledger_id/print",
		Method:       "POST",
		Note:         "Processes print number for the specified general ledger by general_ledger_id.",
		ResponseType: core.GeneralLedgerResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		generalLedgerID, err := helpers.EngineUUIDParam(ctx, "general_ledger_id")
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "payment-general-ledger-param-error",
				Description: fmt.Sprintf("Invalid general ledger id for POST /transaction/general-ledger/:general_ledger_id/payment: %v", err),
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid general ledger ID: " + err.Error()})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User authentication failed or organization not found"})
		}

		// Get general ledger
		generalLedger, err := core.GeneralLedgerManager(service).GetByID(context, *generalLedgerID)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "payment-general-ledger-not-found",
				Description: fmt.Sprintf("General ledger not found for ID %v: %v", generalLedgerID, err),
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "General ledger not found: " + err.Error()})
		}
		maxNumber, err := core.GeneralLedgerPrintMaxNumber(
			context, service,
			*generalLedger.MemberProfileID,
			*generalLedger.AccountID,
			*userOrg.BranchID,
			userOrg.OrganizationID,
		)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "payment-general-ledger-max-number-error",
				Description: fmt.Sprintf("Failed to get max print number for general ledger ID %v: %v", generalLedgerID, err),
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get max print number: " + err.Error()})
		}
		generalLedger.PrintNumber = maxNumber + 1
		if err := core.GeneralLedgerManager(service).UpdateByID(context, generalLedger.ID, generalLedger); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Connect account to FS definition failed (/financial-statement-definition/:financial_statement_definition_id/account/:account_id/connect), account db error: " + err.Error(),
				Module:      "FinancialStatement",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to connect account: " + err.Error()})
		}

		response := core.GeneralLedgerManager(service).ToModel(generalLedger)
		return ctx.JSON(http.StatusOK, response)
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/transaction/:transaction_id/payment",
		Method:       "POST",
		Note:         "Processes a payment for the specified transaction by transaction_id and records it in the general ledger.",
		ResponseType: core.GeneralLedgerResponse{},
		RequestType:  core.PaymentRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()

		var req core.PaymentRequest
		if err := ctx.Bind(&req); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "payment-bind-error",
				Description: "Payment failed: invalid payload: " + err.Error(),
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid payment payload: " + err.Error()})
		}

		if err := service.Validator.Struct(req); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "payment-validation-error",
				Description: "Payment failed: validation error: " + err.Error(),
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Payment validation failed: " + err.Error()})
		}

		tx, endTx := service.Database.StartTransaction(context)
		if tx.Error != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "payment-db-error",
				Description: "Payment failed (/transaction/:transaction_id/payment), begin tx error: " + tx.Error.Error(),
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to start database transaction: " + endTx(tx.Error).Error()})
		}

		transactionID, err := helpers.EngineUUIDParam(ctx, "transaction_id")
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "payment-param-error",
				Description: fmt.Sprintf("Invalid transaction id for POST /transaction/:transaction_id/payment: %v", err),
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid transaction ID: " + err.Error()})
		}

		generalLedger, err := event.TransactionPayment(context, service, ctx, tx, endTx, event.TransactionEvent{
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
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "payment-error",
				Description: "Payment processing failed: " + err.Error(),
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Payment processing failed: " + err.Error()})
		}

		response := core.GeneralLedgerManager(service).ToModel(generalLedger)

		return ctx.JSON(http.StatusOK, response)
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/transaction/:transaction_id/withdraw",
		Method:       "POST",
		Note:         "Processes a withdrawal for the specified transaction by transaction_id and updates the general ledger accordingly.",
		ResponseType: core.GeneralLedgerResponse{},
		RequestType:  core.PaymentRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var req core.PaymentRequest
		if err := ctx.Bind(&req); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "withdraw-bind-error",
				Description: "Withdrawal failed: invalid payload: " + err.Error(),
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid withdrawal payload: " + err.Error()})
		}
		if err := service.Validator.Struct(req); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "withdraw-validation-error",
				Description: "Withdrawal failed: validation error: " + err.Error(),
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Withdrawal validation failed: " + err.Error()})
		}

		tx, endTx := service.Database.StartTransaction(context)
		if tx.Error != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "withdraw-db-error",
				Description: "Withdrawal failed (/transaction/:transaction_id/withdraw), begin tx error: " + tx.Error.Error(),
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to start database transaction: " + endTx(tx.Error).Error()})
		}

		transactionID, err := helpers.EngineUUIDParam(ctx, "transaction_id")
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "withdraw-param-error",
				Description: fmt.Sprintf("Invalid transaction id for POST /transaction/:transaction_id/withdraw: %v", err),
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid transaction ID: " + err.Error()})
		}

		generalLedger, err := event.TransactionPayment(context, service, ctx, tx, endTx, event.TransactionEvent{
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
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "withdraw-error",
				Description: "Withdrawal processing failed: " + err.Error(),
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Withdrawal processing failed: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, core.GeneralLedgerManager(service).ToModel(generalLedger))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/transaction/:transaction_id/deposit",
		Method:       "POST",
		Note:         "Processes a deposit for the specified transaction by transaction_id and updates the general ledger accordingly.",
		ResponseType: core.GeneralLedgerResponse{},
		RequestType:  core.PaymentRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var req core.PaymentRequest
		if err := ctx.Bind(&req); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "deposit-bind-error",
				Description: "Deposit failed: invalid payload: " + err.Error(),
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid deposit payload: " + err.Error()})
		}
		if err := service.Validator.Struct(req); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "deposit-validation-error",
				Description: "Deposit failed: validation error: " + err.Error(),
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Deposit validation failed: " + err.Error()})
		}

		tx, endTx := service.Database.StartTransaction(context)
		if tx.Error != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "deposit-db-error",
				Description: "Deposit failed (/transaction/:transaction_id/deposit), begin tx error: " + tx.Error.Error(),
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to start database transaction: " + endTx(tx.Error).Error()})
		}

		transactionID, err := helpers.EngineUUIDParam(ctx, "transaction_id")
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "deposit-param-error",
				Description: fmt.Sprintf("Invalid transaction id for POST /transaction/:transaction_id/deposit: %v", err),
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid transaction ID: " + err.Error()})
		}

		generalLedger, err := event.TransactionPayment(context, service, ctx, tx, endTx, event.TransactionEvent{
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
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "deposit-error",
				Description: "Deposit processing failed: " + err.Error(),
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Deposit processing failed: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, core.GeneralLedgerManager(service).ToModel(generalLedger))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/transaction/payment",
		Method:       "POST",
		Note:         "Processes a payment for a transaction without specifying transaction_id in the route. Used for general payments.",
		ResponseType: core.GeneralLedger{},
		RequestType:  core.PaymentQuickRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var req core.PaymentQuickRequest
		if err := ctx.Bind(&req); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "general-payment-bind-error",
				Description: "General payment failed: invalid payload: " + err.Error(),
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid payment payload: " + err.Error()})
		}
		if err := service.Validator.Struct(req); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "general-payment-validation-error",
				Description: "General payment failed: validation error: " + err.Error(),
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Payment validation failed: " + err.Error()})
		}

		tx, endTx := service.Database.StartTransaction(context)
		if tx.Error != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "general-payment-db-error",
				Description: "General payment failed (/transaction/payment), begin tx error: " + tx.Error.Error(),
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to start database transaction: " + endTx(tx.Error).Error()})
		}

		generalLedger, err := event.TransactionPayment(context, service, ctx, tx, endTx, event.TransactionEvent{
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
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "general-payment-error",
				Description: "General payment processing failed: " + err.Error(),
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Payment processing failed: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, core.GeneralLedgerManager(service).ToModel(generalLedger))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/transaction/withdraw",
		Method:       "POST",
		Note:         "Processes a withdrawal for a transaction without specifying transaction_id in the route. Used for general withdrawals.",
		ResponseType: core.GeneralLedgerResponse{},
		RequestType:  core.PaymentQuickRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var req core.PaymentQuickRequest
		if err := ctx.Bind(&req); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "general-withdraw-bind-error",
				Description: "General withdrawal failed: invalid payload: " + err.Error(),
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid withdrawal payload: " + err.Error()})
		}
		if err := service.Validator.Struct(req); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "general-withdraw-validation-error",
				Description: "General withdrawal failed: validation error: " + err.Error(),
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Withdrawal validation failed: " + err.Error()})
		}

		tx, endTx := service.Database.StartTransaction(context)
		if tx.Error != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "general-withdraw-db-error",
				Description: "General withdrawal failed (/transaction/withdraw), begin tx error: " + tx.Error.Error(),
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to start database transaction: " + endTx(tx.Error).Error()})
		}

		generalLedger, err := event.TransactionPayment(context, service, ctx, tx, endTx, event.TransactionEvent{
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
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "general-withdraw-error",
				Description: "General withdrawal processing failed: " + err.Error(),
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Withdrawal processing failed: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, core.GeneralLedgerManager(service).ToModel(generalLedger))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/transaction/deposit",
		Method:       "POST",
		Note:         "Processes a deposit for a transaction without specifying transaction_id in the route. Used for general deposits.",
		ResponseType: core.GeneralLedgerResponse{},
		RequestType:  core.PaymentQuickRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var req core.PaymentQuickRequest
		if err := ctx.Bind(&req); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "general-deposit-bind-error",
				Description: "General deposit failed: invalid payload: " + err.Error(),
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid deposit payload: " + err.Error()})
		}
		if err := service.Validator.Struct(req); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "general-deposit-validation-error",
				Description: "General deposit failed: validation error: " + err.Error(),
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Deposit validation failed: " + err.Error()})
		}

		tx, endTx := service.Database.StartTransaction(context)
		if tx.Error != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "general-deposit-db-error",
				Description: "General deposit failed (/transaction/deposit), begin tx error: " + tx.Error.Error(),
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to start database transaction: " + endTx(tx.Error).Error()})
		}

		generalLedger, err := event.TransactionPayment(context, service, ctx, tx, endTx, event.TransactionEvent{
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
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "general-deposit-error",
				Description: "General deposit processing failed: " + err.Error(),
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Deposit processing failed: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, core.GeneralLedgerManager(service).ToModel(generalLedger))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:  "/api/v1/transaction/general-ledger/:general_ledger_id/reverse",
		Method: "POST",
		Note:   "Reverses a specific general ledger transaction.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		generalLedgerID, err := helpers.EngineUUIDParam(ctx, "general_ledger_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid general ledger ID: " + err.Error()})
		}
		generalLedger, err := core.GeneralLedgerManager(service).GetByID(context, *generalLedgerID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "General ledger not found: " + err.Error()})
		}
		tx, endTx := service.Database.StartTransaction(context)
		if tx.Error != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
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
		newGeneralLedger, err := event.TransactionPayment(context, service, ctx, tx, endTx, event.TransactionEvent{
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
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "payment-error",
				Description: "Payment processing failed: " + err.Error(),
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Payment processing failed: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, core.GeneralLedgerManager(service).ToModel(newGeneralLedger))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/transaction/:transaction_id/reverse",
		Method:       "POST",
		Note:         "Reverses all general ledger entries for a specific transaction by transaction_id.",
		ResponseType: core.TransactionResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		transactionID, err := helpers.EngineUUIDParam(ctx, "transaction_id")
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "transaction-reverse-param-error",
				Description: fmt.Sprintf("Invalid transaction id for POST /general-ledger/transaction/:transaction_id/reverse: %v", err),
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid transaction ID: " + err.Error()})
		}

		generalLedgers, err := core.GeneralLedgerManager(service).Find(context, &core.GeneralLedger{
			TransactionID: transactionID,
		})
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "transaction-reverse-ledger-error",
				Description: fmt.Sprintf("Failed to get general ledger entries for transaction ID %v: %v", transactionID, err),
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get general ledger entries: " + err.Error()})
		}

		if len(generalLedgers) == 0 {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "No general ledger entries found for this transaction"})
		}
		tx, endTx := service.Database.StartTransaction(context)
		if tx.Error != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
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
			newGeneralLedger, err := event.TransactionPayment(context, service, ctx, tx, endTx, event.TransactionEvent{
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
				event.Footstep(ctx, service, event.FootstepEvent{
					Activity:    "transaction-reverse-error",
					Description: fmt.Sprintf("Transaction reversal failed for ledger %v: %v", generalLedger.ID, err),
					Module:      "Transaction",
				})
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Transaction reversal failed: " + err.Error()})
			}
			reversedLedgers = append(reversedLedgers, newGeneralLedger)
		}
		transaction, err := core.TransactionManager(service).GetByIDRaw(context, *transactionID)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "transaction-reverse-fetch-error",
				Description: fmt.Sprintf("Failed to fetch transaction %v after reversal: %v", transactionID, err),
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch transaction after reversal: " + err.Error()})
		}
		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "transaction-reverse-success",
			Description: fmt.Sprintf("Successfully reversed transaction %v with %d general ledger entries", transactionID, len(reversedLedgers)),
			Module:      "Transaction",
		})

		return ctx.JSON(http.StatusOK, transaction)
	})

}
