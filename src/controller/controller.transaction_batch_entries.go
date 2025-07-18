package controller

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/lands-horizon/horizon-server/services/horizon"
	"github.com/lands-horizon/horizon-server/src/model"
)

func (c *Controller) TransactionBatchEntriesController() {

	req := c.provider.Service.Request

	// Returns paginated check entries for a given transaction batch.
	req.RegisterRoute(horizon.Route{
		Route:    "/check-entry/transaction-batch/:transaction_batch_id/search",
		Method:   "GET",
		Response: "CheckEntry[]",
		Note:     "Returns paginated check entries for the specified transaction batch.",
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
		check, err := c.model.CheckEntryManager.Find(context, &model.CheckEntry{
			TransactionBatchID: transactionBatchID,
			OrganizationID:     userOrg.OrganizationID,
			BranchID:           *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve check entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.CheckEntryManager.Pagination(context, ctx, check))
	})

	// Returns paginated withdrawal entries for a given transaction batch.
	req.RegisterRoute(horizon.Route{
		Route:    "/withdrawal-entry/transaction-batch/:transaction_batch_id/search",
		Method:   "GET",
		Response: "WithdrawalEntry[]",
		Note:     "Returns paginated withdrawal entries for the specified transaction batch.",
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
		withdrawal, err := c.model.WithdrawalEntryManager.Find(context, &model.WithdrawalEntry{
			TransactionBatchID: transactionBatchID,
			OrganizationID:     userOrg.OrganizationID,
			BranchID:           *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve withdrawal entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.WithdrawalEntryManager.Pagination(context, ctx, withdrawal))
	})

	// Returns paginated online entries for a given transaction batch.
	req.RegisterRoute(horizon.Route{
		Route:    "/online-entry/transaction-batch/:transaction_batch_id/search",
		Method:   "GET",
		Response: "OnlineEntry[]",
		Note:     "Returns paginated online entries for the specified transaction batch.",
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
		online, err := c.model.OnlineEntryManager.Find(context, &model.OnlineEntry{
			TransactionBatchID: transactionBatchID,
			OrganizationID:     userOrg.OrganizationID,
			BranchID:           *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve online entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.OnlineEntryManager.Pagination(context, ctx, online))
	})

	// Returns paginated deposit entries for a given transaction batch.
	req.RegisterRoute(horizon.Route{
		Route:    "/deposit-entry/transaction-batch/:transaction_batch_id/search",
		Method:   "GET",
		Response: "DepositEntry[]",
		Note:     "Returns paginated deposit entries for the specified transaction batch.",
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
		deposit, err := c.model.DepositEntryManager.Find(context, &model.DepositEntry{
			TransactionBatchID: transactionBatchID,
			OrganizationID:     userOrg.OrganizationID,
			BranchID:           *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve deposit entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.DepositEntryManager.Pagination(context, ctx, deposit))
	})

	// Returns paginated transaction entries for a given transaction batch.
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

}
