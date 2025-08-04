package controller_v1

import (
	"github.com/labstack/echo/v4"
	"github.com/lands-horizon/horizon-server/services/handlers"
	"github.com/lands-horizon/horizon-server/src/model"
)

func (c *Controller) DisbursementTransactionController() {
	req := c.provider.Service.Request
	// /disbursement-transaction/transaction-batch/:transaction_batch_id/search
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/disbursement-transaction/transaction-batch/:transaction_batch_id/search",
		Method:       "GET",
		Note:         "Returns all disbursement transactions for a specific transaction batch.",
		ResponseType: model.DisbursementResponse{},
	}, func(ctx echo.Context) error {
		return nil
	})

	req.RegisterRoute(handlers.Route{
		Route:        "/disbursement-transaction/employee/:user_organization_id/search",
		Method:       "GET",
		Note:         "Returns all disbursement transactions handled by a specific employee.",
		ResponseType: model.DisbursementResponse{},
	}, func(ctx echo.Context) error {
		return nil
	})

	req.RegisterRoute(handlers.Route{
		Route:        "/disbursement-transaction/current/search",
		Method:       "GET",
		Note:         "Returns all disbursement transactions for the currently authenticated user.",
		ResponseType: model.DisbursementResponse{},
	}, func(ctx echo.Context) error {
		return nil
	})

	req.RegisterRoute(handlers.Route{
		Route:        "/disbursement-transaction/branch/search",
		Method:       "GET",
		Note:         "Returns all disbursement transactions for the current user's branch.",
		ResponseType: model.DisbursementResponse{},
	}, func(ctx echo.Context) error {
		return nil
	})

	req.RegisterRoute(handlers.Route{
		Route:        "/disbursement-transaction/disbursement/:disbursement_id/search",
		Method:       "GET",
		Note:         "Returns all disbursement transactions for a specific disbursement ID.",
		ResponseType: model.DisbursementResponse{},
	}, func(ctx echo.Context) error {
		return nil
	})
}
