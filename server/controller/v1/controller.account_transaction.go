package v1

import (
	"net/http"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/server/model/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/services/handlers"
	"github.com/labstack/echo/v4"
)

func (c *Controller) accountTransactionController() {

	// GET api/v1/account-transaction/year/:year/month/:month
	// POST api/v1/account-transaction/process-gl
	// GET api/v1/account-transaction/account/:account_id/year/:year

	req := c.provider.Service.Request

	// LIST (current branch)
	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/account-transaction",
		Method:       "GET",
		Note:         "Returns all account transactions for the current user's organization and branch.",
		ResponseType: core.AccountTransactionResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := c.event.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if userOrg.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		transactions, err := c.core.AccountTransactionManager().Find(context, &core.AccountTransaction{
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "No account transactions found"})
		}
		return ctx.JSON(http.StatusOK, c.core.AccountTransactionManager().ToModels(transactions))
	})

	// SEARCH / PAGINATION
	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/account-transaction/search",
		Method:       "GET",
		Note:         "Returns a paginated list of account transactions.",
		ResponseType: core.AccountTransactionResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()

		userOrg, err := c.event.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{
				"error": "User organization not found or authentication failed",
			})
		}
		if userOrg.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		result, err := c.core.AccountTransactionManager().NormalPagination(
			context,
			ctx,
			&core.AccountTransaction{
				OrganizationID: userOrg.OrganizationID,
				BranchID:       *userOrg.BranchID,
			},
		)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{
				"error": "Failed to fetch account transactions: " + err.Error(),
			})
		}

		return ctx.JSON(http.StatusOK, result)
	})

	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/account-transaction/:transaction_id",
		Method:       "GET",
		Note:         "Returns a single account transaction by ID.",
		ResponseType: core.AccountTransactionResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()

		id, err := handlers.EngineUUIDParam(ctx, "transaction_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{
				"error": "Invalid transaction ID",
			})
		}

		transaction, err := c.core.AccountTransactionManager().GetByIDRaw(context, *id)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{
				"error": "Account transaction not found",
			})
		}

		return ctx.JSON(http.StatusOK, transaction)
	})

	// UPDATE
	req.RegisterWebRoute(handlers.Route{
		Route:        "/api/v1/account-transaction/:transaction_id",
		Method:       "PUT",
		Note:         "Updates an account transaction.",
		RequestType:  core.AccountTransactionRequest{},
		ResponseType: core.AccountTransactionResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		id, err := handlers.EngineUUIDParam(ctx, "transaction_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{
				"error": "Invalid transaction ID",
			})
		}
		reqBody, err := c.core.AccountTransactionManager().Validate(ctx)
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{
				"error": err.Error(),
			})
		}
		userOrg, err := c.event.CurrentUserOrganization(context, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized"})
		}
		transaction, err := c.core.AccountTransactionManager().GetByID(context, *id)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Account transaction not found"})
		}
		transaction.Description = reqBody.Description
		transaction.UpdatedAt = time.Now().UTC()
		transaction.UpdatedByID = userOrg.UserID
		if err := c.core.AccountTransactionManager().UpdateByID(context, transaction.ID, transaction); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{
				"error": "Failed to update account transaction",
			})
		}

		return ctx.JSON(http.StatusOK, c.core.AccountTransactionManager().ToModel(transaction))
	})

	// DELETE
	req.RegisterWebRoute(handlers.Route{
		Route:  "/api/v1/account-transaction/:transaction_id",
		Method: "DELETE",
		Note:   "Deletes an account transaction.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		id, err := handlers.EngineUUIDParam(ctx, "transaction_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{
				"error": "Invalid transaction ID",
			})
		}
		if err := c.core.AccountTransactionManager().Delete(context, *id); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete account transaction"})
		}
		return ctx.NoContent(http.StatusNoContent)
	})

}
