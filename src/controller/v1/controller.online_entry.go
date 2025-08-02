package controller_v1

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/lands-horizon/horizon-server/services/handlers"
	"github.com/lands-horizon/horizon-server/src/model"
)

func (c *Controller) OnlineEntryController() {
	req := c.provider.Service.Request

	// Returns paginated online entries for a given transaction batch.
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/online-entry/transaction-batch/:transaction_batch_id/search",
		Method:       "GET",
		ResponseType: model.OnlineEntryResponse{},
		Note:         "Returns paginated online entries for the specified transaction batch.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}
		transactionBatchID, err := handlers.EngineUUIDParam(ctx, "transaction_batch_id")
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

	// Returns paginated online entries for a given transaction batch.
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/online-entry/search",
		Method:       "GET",
		ResponseType: model.OnlineEntryResponse{},
		Note:         "Returns paginated online entries",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user organization: " + err.Error()})
		}

		online, err := c.model.OnlineEntryManager.Find(context, &model.OnlineEntry{
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve online entries: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model.OnlineEntryManager.Pagination(context, ctx, online))
	})

}
