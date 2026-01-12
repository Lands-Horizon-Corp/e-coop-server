package v1

import (
	"net/http"
	"strconv"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/helpers"
	"github.com/Lands-Horizon-Corp/e-coop-server/horizon"
	"github.com/Lands-Horizon-Corp/e-coop-server/server/event"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/core"
	"github.com/labstack/echo/v4"
)

func accountTransactionController(service *horizon.HorizonService) {

	req := service.API
	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/account-transaction/account/:account_id/year/:year",
		Method:       "GET",
		Note:         "Returns account transaction ledgers for a specific account and year.",
		ResponseType: core.AccountTransactionLedgerResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		accountID, err := helpers.EngineUUIDParam(ctx, "account_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid account ID"})
		}
		yearParam := ctx.Param("year")
		year, err := strconv.Atoi(yearParam)
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid year"})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		ledgers, err := c.event.AccountTransactionLedgers(context, *userOrg, year, accountID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch account transaction ledgers: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, ledgers)
	})

	req.RegisterWebRoute(horizon.Route{
		Route:       "/api/v1/account-transaction/process-gl",
		Method:      "POST",
		RequestType: core.AccountTransactionProcessGLRequest{},
		Note:        "Processes account transactions for the specified date range.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var req core.AccountTransactionProcessGLRequest
		if err := ctx.Bind(&req); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "process-gl-error",
				Description: "Process GL failed: invalid payload: " + err.Error(),
				Module:      "AccountTransaction",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request payload: " + err.Error()})
		}
		if err := c.provider.Service.Validator.Struct(req); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "process-gl-error",
				Description: "Process GL failed: validation error: " + err.Error(),
				Module:      "AccountTransaction",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if userOrg.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		if err := c.event.AccountTransactionProcess(context, *userOrg, req); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "process-gl-error",
				Description: "Process GL failed: " + err.Error(),
				Module:      "AccountTransaction",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to process GL: " + err.Error()})
		}

		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "process-gl-success",
			Description: "GL process initiated",
			Module:      "AccountTransaction",
		})
		return ctx.NoContent(http.StatusNoContent)
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/account-transaction/year/:year/month/:month",
		Method:       "GET",
		Note:         "Returns account transactions for the specified year and month.",
		ResponseType: core.AccountTransactionResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		yearParam := ctx.Param("year")
		monthParam := ctx.Param("month")
		year, err := strconv.Atoi(yearParam)
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid year"})
		}
		month, err := strconv.Atoi(monthParam)
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid month"})
		}
		month = month % 12
		if month == 0 {
			month = 12
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if userOrg.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		accountTransactions, err := core.AccountTransactionByMonthYear(context, year, month, userOrg.OrganizationID, *userOrg.BranchID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch account transactions: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, core.AccountTransactionManager(service).ToModels(accountTransactions))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/account-transaction",
		Method:       "GET",
		Note:         "Returns all account transactions for the current user's organization and branch.",
		ResponseType: core.AccountTransactionResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		if userOrg.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		transactions, err := core.AccountTransactionManager(service).Find(context, &core.AccountTransaction{
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "No account transactions found"})
		}
		return ctx.JSON(http.StatusOK, core.AccountTransactionManager(service).ToModels(transactions))
	})

	// SEARCH / PAGINATION
	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/account-transaction/search",
		Method:       "GET",
		Note:         "Returns a paginated list of account transactions.",
		ResponseType: core.AccountTransactionResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()

		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{
				"error": "User organization not found or authentication failed",
			})
		}
		if userOrg.BranchID == nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "User is not assigned to a branch"})
		}
		result, err := core.AccountTransactionManager(service).NormalPagination(
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

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/account-transaction/:transaction_id",
		Method:       "GET",
		Note:         "Returns a single account transaction by ID.",
		ResponseType: core.AccountTransactionResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()

		id, err := helpers.EngineUUIDParam(ctx, "transaction_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{
				"error": "Invalid transaction ID",
			})
		}

		transaction, err := core.AccountTransactionManager(service).GetByIDRaw(context, *id)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{
				"error": "Account transaction not found",
			})
		}

		return ctx.JSON(http.StatusOK, transaction)
	})

	// UPDATE
	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/account-transaction/:transaction_id",
		Method:       "PUT",
		Note:         "Updates an account transaction.",
		RequestType:  core.AccountTransactionRequest{},
		ResponseType: core.AccountTransactionResponse{},
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		id, err := helpers.EngineUUIDParam(ctx, "transaction_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{
				"error": "Invalid transaction ID",
			})
		}
		reqBody, err := core.AccountTransactionManager(service).Validate(ctx)
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{
				"error": err.Error(),
			})
		}
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized"})
		}
		transaction, err := core.AccountTransactionManager(service).GetByID(context, *id)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Account transaction not found"})
		}
		transaction.Description = reqBody.Description
		transaction.UpdatedAt = time.Now().UTC()
		transaction.UpdatedByID = userOrg.UserID
		if err := core.AccountTransactionManager(service).UpdateByID(context, transaction.ID, transaction); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{
				"error": "Failed to update account transaction",
			})
		}

		return ctx.JSON(http.StatusOK, core.AccountTransactionManager(service).ToModel(transaction))
	})

	// DELETE
	req.RegisterWebRoute(horizon.Route{
		Route:  "/api/v1/account-transaction/:transaction_id",
		Method: "DELETE",
		Note:   "Deletes an account transaction.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		id, err := helpers.EngineUUIDParam(ctx, "transaction_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{
				"error": "Invalid transaction ID",
			})
		}
		if err := core.AccountTransactionManager(service).Delete(context, *id); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete account transaction"})
		}
		return ctx.NoContent(http.StatusNoContent)
	})

}
