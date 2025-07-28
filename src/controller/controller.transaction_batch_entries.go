package controller

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/lands-horizon/horizon-server/services/handlers"
	"github.com/lands-horizon/horizon-server/services/horizon"
	"github.com/lands-horizon/horizon-server/src/model"
)

func (c *Controller) TransactionBatchEntriesController() {

	req := c.provider.Service.Request

	req.RegisterRoute(handlers.Route{
		Route:        "/check-entry/transaction-batch/:transaction_batch_id/search",
		Method:       "GET",
		ResponseType: model.CheckEntryResponse{},
		Note:         "Returns paginated check entries for the specified transaction batch.",
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

	req.RegisterRoute(handlers.Route{
		Route:        "/check-entry/search",
		Method:       "GET",
		ResponseType: model.CheckEntryResponse{},
		Note:         "Returns paginated check entries",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}

		check, err := c.model.CheckEntryManager.Find(context, &model.CheckEntry{
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve check entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.CheckEntryManager.Pagination(context, ctx, check))
	})

	// Returns paginated withdrawal entries for a given transaction batch.
	req.RegisterRoute(handlers.Route{
		Route:        "/withdrawal-entry/transaction-batch/:transaction_batch_id/search",
		Method:       "GET",
		ResponseType: model.WithdrawalEntryResponse{},
		Note:         "Returns paginated withdrawal entries for the specified transaction batch.",
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

	// Returns paginated deposit entries for a given transaction batch.
	req.RegisterRoute(handlers.Route{
		Route:        "/deposit-entry/transaction-batch/:transaction_batch_id/search",
		Method:       "GET",
		ResponseType: model.DepositEntryResponse{},
		Note:         "Returns paginated deposit entries for the specified transaction batch.",
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

}
