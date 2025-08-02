package controller_v1

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/lands-horizon/horizon-server/services/handlers"
	"github.com/lands-horizon/horizon-server/src/event"
	"github.com/lands-horizon/horizon-server/src/model"
)

func (c *Controller) PaymentController() {
	req := c.provider.Service.Request

	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/transaction/:transaction_id/payment",
		Method:       "POST",
		Note:         "Processes a payment for the specified transaction by transaction_id and records it in the general ledger.",
		ResponseType: model.GeneralLedgerResponse{},
		RequestType:  model.PaymentRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var req model.PaymentRequest
		if err := ctx.Bind(&req); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "payment-bind-error",
				Description: "Payment failed: invalid payload: " + err.Error(),
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid payment payload: " + err.Error()})
		}
		if err := c.provider.Service.Validator.Struct(req); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "payment-validation-error",
				Description: "Payment failed: validation error: " + err.Error(),
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Payment validation failed: " + err.Error()})
		}

		tx := c.provider.Service.Database.Client().Begin()
		if tx.Error != nil {
			tx.Rollback()
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "payment-db-error",
				Description: "Payment failed (/transaction/:transaction_id/payment), begin tx error: " + tx.Error.Error(),
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to start database transaction: " + tx.Error.Error()})
		}

		transactionId, err := handlers.EngineUUIDParam(ctx, "transaction_id")
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "payment-param-error",
				Description: fmt.Sprintf("Invalid transaction id for POST /transaction/:transaction_id/payment: %v", err),
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid transaction ID: " + err.Error()})
		}

		generalLedger, err := c.event.Payment(context, ctx, tx, &model.PaymentQuickRequest{
			Amount:                req.Amount,
			SignatureMediaID:      req.SignatureMediaID,
			ProofOfPaymentMediaID: req.ProofOfPaymentMediaID,
			BankID:                req.BankID,
			BankReferenceNumber:   req.BankReferenceNumber,
			EntryDate:             req.EntryDate,
			AccountID:             req.AccountID,
			PaymentTypeID:         req.PaymentTypeID,
			Description:           req.Description,
		}, transactionId, model.GeneralLedgerSourcePayment)
		if err != nil {
			tx.Rollback()
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "payment-error",
				Description: "Payment processing failed: " + err.Error(),
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Payment processing failed: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.GeneralLedgerManager.ToModel(generalLedger))
	})

	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/transaction/:transaction_id/withdraw",
		Method:       "POST",
		Note:         "Processes a withdrawal for the specified transaction by transaction_id and updates the general ledger accordingly.",
		ResponseType: model.GeneralLedgerResponse{},
		RequestType:  model.PaymentRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var req model.PaymentRequest
		if err := ctx.Bind(&req); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "withdraw-bind-error",
				Description: "Withdrawal failed: invalid payload: " + err.Error(),
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid withdrawal payload: " + err.Error()})
		}
		if err := c.provider.Service.Validator.Struct(req); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "withdraw-validation-error",
				Description: "Withdrawal failed: validation error: " + err.Error(),
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Withdrawal validation failed: " + err.Error()})
		}

		tx := c.provider.Service.Database.Client().Begin()
		if tx.Error != nil {
			tx.Rollback()
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "withdraw-db-error",
				Description: "Withdrawal failed (/transaction/:transaction_id/withdraw), begin tx error: " + tx.Error.Error(),
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to start database transaction: " + tx.Error.Error()})
		}

		transactionId, err := handlers.EngineUUIDParam(ctx, "transaction_id")
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "withdraw-param-error",
				Description: fmt.Sprintf("Invalid transaction id for POST /transaction/:transaction_id/withdraw: %v", err),
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid transaction ID: " + err.Error()})
		}

		generalLedger, err := c.event.Payment(context, ctx, tx, &model.PaymentQuickRequest{
			Amount:                req.Amount,
			SignatureMediaID:      req.SignatureMediaID,
			ProofOfPaymentMediaID: req.ProofOfPaymentMediaID,
			BankID:                req.BankID,
			BankReferenceNumber:   req.BankReferenceNumber,
			EntryDate:             req.EntryDate,
			AccountID:             req.AccountID,
			PaymentTypeID:         req.PaymentTypeID,
			Description:           req.Description,
		}, transactionId, model.GeneralLedgerSourceWithdraw)
		if err != nil {
			tx.Rollback()
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "withdraw-error",
				Description: "Withdrawal processing failed: " + err.Error(),
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Withdrawal processing failed: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.GeneralLedgerManager.ToModel(generalLedger))
	})

	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/transaction/:transaction_id/deposit",
		Method:       "POST",
		Note:         "Processes a deposit for the specified transaction by transaction_id and updates the general ledger accordingly.",
		ResponseType: model.GeneralLedgerResponse{},
		RequestType:  model.PaymentRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var req model.PaymentRequest
		if err := ctx.Bind(&req); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "deposit-bind-error",
				Description: "Deposit failed: invalid payload: " + err.Error(),
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid deposit payload: " + err.Error()})
		}
		if err := c.provider.Service.Validator.Struct(req); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "deposit-validation-error",
				Description: "Deposit failed: validation error: " + err.Error(),
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Deposit validation failed: " + err.Error()})
		}

		tx := c.provider.Service.Database.Client().Begin()
		if tx.Error != nil {
			tx.Rollback()
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "deposit-db-error",
				Description: "Deposit failed (/transaction/:transaction_id/deposit), begin tx error: " + tx.Error.Error(),
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to start database transaction: " + tx.Error.Error()})
		}

		transactionId, err := handlers.EngineUUIDParam(ctx, "transaction_id")
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "deposit-param-error",
				Description: fmt.Sprintf("Invalid transaction id for POST /transaction/:transaction_id/deposit: %v", err),
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid transaction ID: " + err.Error()})
		}

		generalLedger, err := c.event.Payment(context, ctx, tx, &model.PaymentQuickRequest{
			Amount:                req.Amount,
			SignatureMediaID:      req.SignatureMediaID,
			ProofOfPaymentMediaID: req.ProofOfPaymentMediaID,
			BankID:                req.BankID,
			BankReferenceNumber:   req.BankReferenceNumber,
			EntryDate:             req.EntryDate,
			AccountID:             req.AccountID,
			PaymentTypeID:         req.PaymentTypeID,
			Description:           req.Description,
		}, transactionId, model.GeneralLedgerSourceDeposit)
		if err != nil {
			tx.Rollback()
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "deposit-error",
				Description: "Deposit processing failed: " + err.Error(),
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Deposit processing failed: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.GeneralLedgerManager.ToModel(generalLedger))
	})

	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/transaction/payment",
		Method:       "POST",
		Note:         "Processes a payment for a transaction without specifying transaction_id in the route. Used for general payments.",
		ResponseType: model.GeneralLedger{},
		RequestType:  model.PaymentQuickRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var req model.PaymentQuickRequest
		if err := ctx.Bind(&req); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "general-payment-bind-error",
				Description: "General payment failed: invalid payload: " + err.Error(),
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid payment payload: " + err.Error()})
		}
		if err := c.provider.Service.Validator.Struct(req); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "general-payment-validation-error",
				Description: "General payment failed: validation error: " + err.Error(),
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Payment validation failed: " + err.Error()})
		}

		tx := c.provider.Service.Database.Client().Begin()
		if tx.Error != nil {
			tx.Rollback()
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "general-payment-db-error",
				Description: "General payment failed (/transaction/payment), begin tx error: " + tx.Error.Error(),
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to start database transaction: " + tx.Error.Error()})
		}

		generalLedger, err := c.event.Payment(context, ctx, tx, &req, nil, model.GeneralLedgerSourcePayment)
		if err != nil {
			tx.Rollback()
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "general-payment-error",
				Description: "General payment processing failed: " + err.Error(),
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Payment processing failed: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.GeneralLedgerManager.ToModel(generalLedger))
	})

	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/transaction/withdraw",
		Method:       "POST",
		Note:         "Processes a withdrawal for a transaction without specifying transaction_id in the route. Used for general withdrawals.",
		ResponseType: model.GeneralLedgerResponse{},
		RequestType:  model.PaymentQuickRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var req model.PaymentQuickRequest
		if err := ctx.Bind(&req); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "general-withdraw-bind-error",
				Description: "General withdrawal failed: invalid payload: " + err.Error(),
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid withdrawal payload: " + err.Error()})
		}
		if err := c.provider.Service.Validator.Struct(req); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "general-withdraw-validation-error",
				Description: "General withdrawal failed: validation error: " + err.Error(),
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Withdrawal validation failed: " + err.Error()})
		}

		tx := c.provider.Service.Database.Client().Begin()
		if tx.Error != nil {
			tx.Rollback()
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "general-withdraw-db-error",
				Description: "General withdrawal failed (/transaction/withdraw), begin tx error: " + tx.Error.Error(),
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to start database transaction: " + tx.Error.Error()})
		}

		generalLedger, err := c.event.Payment(context, ctx, tx, &req, nil, model.GeneralLedgerSourceWithdraw)
		if err != nil {
			tx.Rollback()
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "general-withdraw-error",
				Description: "General withdrawal processing failed: " + err.Error(),
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Withdrawal processing failed: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.GeneralLedgerManager.ToModel(generalLedger))
	})

	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/transaction/deposit",
		Method:       "POST",
		Note:         "Processes a deposit for a transaction without specifying transaction_id in the route. Used for general deposits.",
		ResponseType: model.GeneralLedgerResponse{},
		RequestType:  model.PaymentQuickRequest{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var req model.PaymentQuickRequest
		if err := ctx.Bind(&req); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "general-deposit-bind-error",
				Description: "General deposit failed: invalid payload: " + err.Error(),
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid deposit payload: " + err.Error()})
		}
		if err := c.provider.Service.Validator.Struct(req); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "general-deposit-validation-error",
				Description: "General deposit failed: validation error: " + err.Error(),
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Deposit validation failed: " + err.Error()})
		}

		tx := c.provider.Service.Database.Client().Begin()
		if tx.Error != nil {
			tx.Rollback()
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "general-deposit-db-error",
				Description: "General deposit failed (/transaction/deposit), begin tx error: " + tx.Error.Error(),
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to start database transaction: " + tx.Error.Error()})
		}

		generalLedger, err := c.event.Payment(context, ctx, tx, &req, nil, model.GeneralLedgerSourceDeposit)
		if err != nil {
			tx.Rollback()
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "general-deposit-error",
				Description: "General deposit processing failed: " + err.Error(),
				Module:      "Transaction",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Deposit processing failed: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.GeneralLedgerManager.ToModel(generalLedger))
	})
}
