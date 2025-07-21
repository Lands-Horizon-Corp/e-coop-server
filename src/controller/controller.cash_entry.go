package controller

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/lands-horizon/horizon-server/services/horizon"
	"github.com/lands-horizon/horizon-server/src/model"
)

func (c *Controller) CashEntryController() {
	req := c.provider.Service.Request

	// Returns paginated cash entries for a given transaction batch.
	req.RegisterRoute(horizon.Route{
		Route:    "/cash-entry/transaction-batch/:transaction_batch_id/search",
		Method:   "GET",
		Response: "CashEntry[]",
		Note:     "Returns paginated cash entries for the specified transaction batch.",
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
		cash, err := c.model.CashEntryManager.Find(context, &model.CashEntry{
			TransactionBatchID: transactionBatchID,
			OrganizationID:     userOrg.OrganizationID,
			BranchID:           *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve cash entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.CashEntryManager.Pagination(context, ctx, cash))
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/cash-entry/transaction-batch/:transaction_batch_id",
		Method:   "GET",
		Response: "CashEntry[]",
		Note:     "Returns paginated cash entries for the specified transaction batch.",
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
		cash, err := c.model.CashEntryManager.Find(context, &model.CashEntry{
			TransactionBatchID: transactionBatchID,
			OrganizationID:     userOrg.OrganizationID,
			BranchID:           *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve cash entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.CashEntryManager.Filtered(context, ctx, cash))
	})
}
