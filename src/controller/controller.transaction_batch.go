package controller

import (
	"github.com/labstack/echo/v4"
	"github.com/lands-horizon/horizon-server/services/horizon"
)

func (c *Controller) TransactionBatchController() {
	req := c.provider.Service.Request

	req.RegisterRoute(horizon.Route{
		Route:    "/transaction-batch",
		Method:   "GET",
		Response: "ITransactionBatch[]",
		Note:     "List all transaction batches for the current branch",
	}, func(ctx echo.Context) error {
		return nil
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/transaction-batch/current",
		Method:   "GET",
		Response: "ITransactionBatch",
		Note:     "Get the current active transaction batch for the user",
	}, func(ctx echo.Context) error {
		return nil
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/transaction-batch",
		Method:   "POST",
		Response: "ITransactionBatch",
		Request:  "ITransactionBatch",
		Note:     "Create and start a new transaction batch; returns the created batch. (Will populate Cashcount)",
	}, func(ctx echo.Context) error {
		return nil
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/transaction-batch/end",
		Method:   "POST",
		Response: "ITransactionBatch",
		Request:  "ITransactionBatch",
		Note:     "End the current transaction batch for the authenticated user",
	}, func(ctx echo.Context) error {
		return nil
	})

	req.RegisterRoute(horizon.Route{
		Route:  "/transaction-batch/:transaction_batch_id",
		Method: "GET",
		Note:   "Retrieve a transaction batch by its ID",
	}, func(ctx echo.Context) error {
		return nil
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/transaction-batch/:transaction_batch_id/view-request",
		Method:   "PUT",
		Response: "ITransactionBatch",
		Note:     "Submit a request to view (blotter) a specific transaction batch",
	}, func(ctx echo.Context) error {
		return nil
	})

	req.RegisterRoute(horizon.Route{
		Route:  "/transaction-batch/:transaction_batch_id/deposit-in-bank",
		Method: "PUT",
		Note:   "Update the 'deposit in bank' value for a transaction batch",
	}, func(ctx echo.Context) error {
		return nil
	})

	req.RegisterRoute(horizon.Route{
		Route:  "/transaction-batch/view-request",
		Method: "GET",
		Note:   "List all pending view (blotter) requests for transaction batches",
	}, func(ctx echo.Context) error {
		return nil
	})

	req.RegisterRoute(horizon.Route{
		Route:  "/transaction-batch/:transaction_batch_id/view-accept",
		Method: "PUT",
		Note:   "Accept a view (blotter) request for a transaction batch by ID",
	}, func(ctx echo.Context) error {
		return nil
	})
}

func (c *Controller) CashCountController() {
	req := c.provider.Service.Request

	req.RegisterRoute(horizon.Route{
		Route:    "/cash-count",
		Method:   "GET",
		Response: "ICashCount[]",
		Note:     "Retrieve batch cash count bills (JWT) for the current transaction batch before ending.",
	}, func(ctx echo.Context) error {
		return nil
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/cash-count",
		Method:   "POST",
		Response: "ICashCount",
		Request:  "ICashCount",
		Note:     "Add a cash count bill to the current transaction batch before ending.",
	}, func(ctx echo.Context) error {
		return nil
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/cash-count/:id",
		Method:   "DELETE",
		Response: "ICashCount",
		Note:     "Delete cash count (JWT) with the specified ID from the current transaction batch before ending.",
	}, func(ctx echo.Context) error {
		return nil
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/cash-count/:id",
		Method:   "GET",
		Response: "ICashCount",
		Note:     "Retrieve specific cash count information based on ID from the current transaction batch before ending.",
	}, func(ctx echo.Context) error {
		return nil
	})
}

func (c *Controller) BatchFunding() {
	req := c.provider.Service.Request

	req.RegisterRoute(horizon.Route{
		Route:    "/batch-funding",
		Method:   "POST",
		Request:  "IBatchFunding",
		Response: "IBatchFunding",
		Note:     "Sart: create batch funding based on current transaction batch",
	}, func(ctx echo.Context) error {
		return nil
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/batch-funding/transaction-batch/:transaction_batch_id",
		Method:   "GET",
		Response: "IBatchFunding[]",
		Note:     "Get all batch funding of transaction batch.",
	}, func(ctx echo.Context) error {
		return nil
	})
}

func (C *Controller) CheckRemittance() {

}
