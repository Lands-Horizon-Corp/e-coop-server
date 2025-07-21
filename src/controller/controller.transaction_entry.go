package controller

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/lands-horizon/horizon-server/services/horizon"
	"github.com/lands-horizon/horizon-server/src/model"
)

func (c *Controller) TransactionEntryController() {
	req := c.provider.Service.Request

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
