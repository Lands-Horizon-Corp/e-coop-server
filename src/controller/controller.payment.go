package controller

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
		Route:        "/transaction/:transaction_id/payment",
		Method:       "POST",
		Note:         "Processes a payment for the specified transaction by transaction_id and records it in the general ledger.",
		ResponseType: model.GeneralLedger{},
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
				Activity:    "bulk-delete-error",
				Description: "Feedback bulk delete failed (/feedback/bulk-delete), begin tx error: " + tx.Error.Error(),
				Module:      "Feedback",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to start database transaction: " + tx.Error.Error()})
		}

		transactionId, err := handlers.EngineUUIDParam(ctx, "transaction_id")
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update error",
				Description: fmt.Sprintf("Invalid transaction id for POST /transaction/:transaction_id/payment: %v", err),
				Module:      "transaction",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid branch ID: " + err.Error()})
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
		Route:        "/transaction/:transaction_id/withdraw",
		Method:       "POST",
		Note:         "Processes a withdrawal for the specified transaction by transaction_id and updates the general ledger accordingly.",
		ResponseType: model.GeneralLedger{},
		RequestType:  model.PaymentRequest{},
	}, func(ctx echo.Context) error {

		return nil
	})

	req.RegisterRoute(handlers.Route{
		Route:        "/transaction/:transaction_id/deposit",
		Method:       "POST",
		Note:         "Processes a deposit for the specified transaction by transaction_id and updates the general ledger accordingly.",
		ResponseType: model.GeneralLedger{},
		RequestType:  model.PaymentRequest{},
	}, func(ctx echo.Context) error {
		return nil
	})

	req.RegisterRoute(handlers.Route{
		Route:        "/transaction/payment",
		Method:       "POST",
		Note:         "Processes a payment for a transaction without specifying transaction_id in the route. Used for general payments.",
		ResponseType: model.GeneralLedger{},
		RequestType:  model.PaymentQuickRequest{},
	}, func(ctx echo.Context) error {
		return nil
	})

	req.RegisterRoute(handlers.Route{
		Route:        "/transaction/withdraw",
		Method:       "POST",
		Note:         "Processes a withdrawal for a transaction without specifying transaction_id in the route. Used for general withdrawals.",
		ResponseType: model.GeneralLedger{},
		RequestType:  model.PaymentQuickRequest{},
	}, func(ctx echo.Context) error {
		return nil
	})

	req.RegisterRoute(handlers.Route{
		Route:        "/transaction/deposit",
		Method:       "POST",
		Note:         "Processes a deposit for a transaction without specifying transaction_id in the route. Used for general deposits.",
		ResponseType: model.GeneralLedger{},
		RequestType:  model.PaymentQuickRequest{},
	}, func(ctx echo.Context) error {
		return nil
	})
}
