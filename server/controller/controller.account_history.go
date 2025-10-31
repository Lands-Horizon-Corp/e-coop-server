package v1

import (
	"net/http"

	"github.com/Lands-Horizon-Corp/e-coop-server/server/model/modelcore"
	"github.com/Lands-Horizon-Corp/e-coop-server/services/handlers"
	"github.com/labstack/echo/v4"
)

func (c *Controller) accountHistory() {
	req := c.provider.Service.Request

	// GET api/v1/account-history/account/:account_id
	req.RegisterRoute(handlers.Route{
		Method:       "GET",
		Route:        "/api/v1/account-history/account/:account_id",
		ResponseType: modelcore.AccountHistoryResponse{},
		Note:         "Get account history by account ID",
	},
		func(ctx echo.Context) error {
			context := ctx.Request().Context()
			accountID, err := handlers.EngineUUIDParam(ctx, "account_id")
			if err != nil {
				return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid account_id: " + err.Error()})
			}
			userOrg, err := c.userOrganizationToken.CurrentUserOrganization(context, ctx)
			if err != nil {
				return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Authorization failed: Unable to determine user organization. " + err.Error()})
			}
			accountHistory, err := c.modelcore.AccountHistoryManager.FindRaw(context, &modelcore.AccountHistory{
				AccountID:      *accountID,
				OrganizationID: userOrg.OrganizationID,
				BranchID:       *userOrg.BranchID,
			})
			if err != nil {
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve account history: " + err.Error()})
			}
			return ctx.JSON(http.StatusOK, accountHistory)
		})
}
