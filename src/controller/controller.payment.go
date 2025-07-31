package controller

import (
	"github.com/labstack/echo/v4"
	"github.com/lands-horizon/horizon-server/services/handlers"
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
	}, func(c echo.Context) error {
		return nil
	})

	req.RegisterRoute(handlers.Route{
		Route:        "/transaction/:transaction_id/withdraw",
		Method:       "POST",
		Note:         "Processes a withdrawal for the specified transaction by transaction_id and updates the general ledger accordingly.",
		ResponseType: model.GeneralLedger{},
		RequestType:  model.PaymentRequest{},
	}, func(c echo.Context) error {
		return nil
	})

	req.RegisterRoute(handlers.Route{
		Route:        "/transaction/:transaction_id/deposit",
		Method:       "POST",
		Note:         "Processes a deposit for the specified transaction by transaction_id and updates the general ledger accordingly.",
		ResponseType: model.GeneralLedger{},
		RequestType:  model.PaymentRequest{},
	}, func(c echo.Context) error {
		return nil
	})

	req.RegisterRoute(handlers.Route{
		Route:        "/transaction/payment",
		Method:       "POST",
		Note:         "Processes a payment for a transaction without specifying transaction_id in the route. Used for general payments.",
		ResponseType: model.GeneralLedger{},
		RequestType:  model.PaymentQuickRequest{},
	}, func(c echo.Context) error {
		return nil
	})

	req.RegisterRoute(handlers.Route{
		Route:        "/transaction/withdraw",
		Method:       "POST",
		Note:         "Processes a withdrawal for a transaction without specifying transaction_id in the route. Used for general withdrawals.",
		ResponseType: model.GeneralLedger{},
		RequestType:  model.PaymentQuickRequest{},
	}, func(c echo.Context) error {
		return nil
	})

	req.RegisterRoute(handlers.Route{
		Route:        "/transaction/deposit",
		Method:       "POST",
		Note:         "Processes a deposit for a transaction without specifying transaction_id in the route. Used for general deposits.",
		ResponseType: model.GeneralLedger{},
		RequestType:  model.PaymentQuickRequest{},
	}, func(c echo.Context) error {
		return nil
	})
}
