package controller

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/lands-horizon/horizon-server/services/horizon"
	"github.com/lands-horizon/horizon-server/src/model"
)

func (c *Controller) TransactionBatchEntriesController() {

	req := c.provider.Service.Request

	req.RegisterRoute(horizon.Route{
		Route:    "/check-entry/transaction-batch/:transaction_batch_id/search",
		Method:   "GET",
		Response: "CheckEntry[]",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.NoContent(http.StatusNoContent)
		}
		transactionBatchID, err := horizon.EngineUUIDParam(ctx, "transaction_batch_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid transaction batch ID"})
		}
		check, err := c.model.CheckEntryManager.Find(context, &model.CheckEntry{
			TransactionBatchID: transactionBatchID,
			OrganizationID:     userOrg.OrganizationID,
			BranchID:           *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.CheckEntryManager.Pagination(context, ctx, check))
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/withdrawal-entry/transaction-batch/:transaction_batch_id/search",
		Method:   "GET",
		Response: "WithdrawalEntry[]",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.NoContent(http.StatusNoContent)
		}
		transactionBatchID, err := horizon.EngineUUIDParam(ctx, "transaction_batch_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid transaction batch ID"})
		}
		withdrawal, err := c.model.WithdrawalEntryManager.Find(context, &model.WithdrawalEntry{
			TransactionBatchID: transactionBatchID,
			OrganizationID:     userOrg.OrganizationID,
			BranchID:           *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.WithdrawalEntryManager.Pagination(context, ctx, withdrawal))
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/online-entry/transaction-batch/:transaction_batch_id/search",
		Method:   "GET",
		Response: "OnlineEntry[]",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.NoContent(http.StatusNoContent)
		}
		transactionBatchID, err := horizon.EngineUUIDParam(ctx, "transaction_batch_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid transaction batch ID"})
		}
		online, err := c.model.OnlineEntryManager.Find(context, &model.OnlineEntry{
			TransactionBatchID: transactionBatchID,
			OrganizationID:     userOrg.OrganizationID,
			BranchID:           *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.OnlineEntryManager.Pagination(context, ctx, online))
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/deposit-entry/transaction-batch/:transaction_batch_id/search",
		Method:   "GET",
		Response: "DepositEntry[]",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.NoContent(http.StatusNoContent)
		}
		transactionBatchID, err := horizon.EngineUUIDParam(ctx, "transaction_batch_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid transaction batch ID"})
		}
		deposit, err := c.model.DepositEntryManager.Find(context, &model.DepositEntry{
			TransactionBatchID: transactionBatchID,
			OrganizationID:     userOrg.OrganizationID,
			BranchID:           *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.DepositEntryManager.Pagination(context, ctx, deposit))
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/transaction-entry/transaction-batch/:transaction_batch_id/search",
		Method:   "GET",
		Response: "TransactionEntry[]",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.NoContent(http.StatusNoContent)
		}
		transactionBatchID, err := horizon.EngineUUIDParam(ctx, "transaction_batch_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid transaction batch ID"})
		}
		transaction, err := c.model.TransactionEntryManager.Find(context, &model.TransactionEntry{
			TransactionBatchID: transactionBatchID,
			OrganizationID:     userOrg.OrganizationID,
			BranchID:           *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.TransactionEntryManager.Pagination(context, ctx, transaction))
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/cash-entry/transaction-batch/:transaction_batch_id/search",
		Method:   "GET",
		Response: "CashEntry[]",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.NoContent(http.StatusNoContent)
		}
		transactionBatchID, err := horizon.EngineUUIDParam(ctx, "transaction_batch_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid transaction batch ID"})
		}
		cash, err := c.model.CashEntryManager.Find(context, &model.CashEntry{
			TransactionBatchID: transactionBatchID,
			OrganizationID:     userOrg.OrganizationID,
			BranchID:           *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.CashEntryManager.Pagination(context, ctx, cash))
	})

}
