package v1

import (
	"fmt"
	"net/http"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/helpers"
	"github.com/Lands-Horizon-Corp/e-coop-server/horizon"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/event"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/usecase"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func currencyController(service *horizon.HorizonService) {
	req := service.API

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/currency",
		Method:       "GET",
		ResponseType: core.CurrencyResponse{},
		Note:         "Returns all currencies.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		currencies, err := core.CurrencyManager(service).List(context)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve currencies: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, core.CurrencyManager(service).ToModels(currencies))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/currency/blotter-available",
		Method:       "GET",
		ResponseType: core.CurrencyResponse{},
		Note:         "Returns all available currencies on unbalance accounts.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		currency := []*core.Currency{}
		for _, unbal := range userOrg.Branch.BranchSetting.UnbalancedAccounts {
			if unbal.Currency != nil {
				currency = append(currency, unbal.Currency)
			}
		}
		return ctx.JSON(http.StatusOK, core.CurrencyManager(service).ToModels(currency))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/currency/available",
		Method:       "GET",
		ResponseType: core.CurrencyResponse{},
		Note:         "Returns all available currencies.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		userOrg, err := event.CurrentUserOrganization(context, service, ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Bank update failed (/bank/:bank_id), user org error: " + err.Error(),
				Module:      "Bank",
			})
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User organization not found or authentication failed"})
		}
		accounts, err := core.AccountManager(service).Find(context, &core.Account{
			OrganizationID: userOrg.OrganizationID,
			BranchID:       *userOrg.BranchID,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve accounts: " + err.Error()})
		}
		currencies := []*core.Currency{}
		currencyMap := make(map[uuid.UUID]*core.Currency)
		for _, account := range accounts {
			if account.Currency != nil {
				currencyMap[account.Currency.ID] = account.Currency
			}
		}
		for _, currency := range currencyMap {
			currencies = append(currencies, currency)
		}
		return ctx.JSON(http.StatusOK, core.CurrencyManager(service).ToModels(currencies))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/currency/:currency_id",
		Method:       "GET",
		ResponseType: core.CurrencyResponse{},
		Note:         "Returns a specific currency by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		currencyID, err := helpers.EngineUUIDParam(ctx, "currency_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid currency_id: " + err.Error()})
		}

		currency, err := core.CurrencyManager(service).GetByIDRaw(context, *currencyID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Currency not found: " + err.Error()})
		}

		return ctx.JSON(http.StatusOK, currency)
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/currency/code/:currency_code",
		Method:       "GET",
		ResponseType: core.CurrencyResponse{},
		Note:         "Returns a specific currency by its code (e.g., USD, EUR).",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		currencyCode := ctx.Param("currency_code")
		if currencyCode == "" {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Currency code is required"})
		}

		currency, err := core.CurrencyFindByCode(context, currencyCode)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Currency not found: " + err.Error()})
		}

		return ctx.JSON(http.StatusOK, core.CurrencyManager(service).ToModel(currency))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/currency",
		Method:       "POST",
		ResponseType: core.CurrencyResponse{},
		RequestType:  core.CurrencyRequest{},
		Note:         "Creates a new currency.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := core.CurrencyManager(service).Validate(ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create currency failed: validation error: " + err.Error(),
				Module:      "Currency",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}

		currency := &core.Currency{
			Name:         req.Name,
			Country:      req.Country,
			CurrencyCode: req.CurrencyCode,
			Symbol:       req.Symbol,
			Emoji:        req.Emoji,
			Timezone:     req.Timezone,
			CreatedAt:    time.Now().UTC(),
			UpdatedAt:    time.Now().UTC(),
		}

		if err := core.CurrencyManager(service).Create(context, currency); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create currency failed: create error: " + err.Error(),
				Module:      "Currency",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create currency: " + err.Error()})
		}

		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "create-success",
			Description: "Created currency: " + currency.Name + " (" + currency.CurrencyCode + ")",
			Module:      "Currency",
		})

		return ctx.JSON(http.StatusOK, core.CurrencyManager(service).ToModel(currency))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/currency/:currency_id",
		Method:       "PUT",
		ResponseType: core.CurrencyResponse{},
		RequestType:  core.CurrencyRequest{},
		Note:         "Updates an existing currency by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		currencyID, err := helpers.EngineUUIDParam(ctx, "currency_id")
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update currency failed: invalid currency_id: " + err.Error(),
				Module:      "Currency",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid currency_id: " + err.Error()})
		}

		req, err := core.CurrencyManager(service).Validate(ctx)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update currency failed: validation error: " + err.Error(),
				Module:      "Currency",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}

		currency, err := core.CurrencyManager(service).GetByID(context, *currencyID)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update currency failed: not found: " + err.Error(),
				Module:      "Currency",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Currency not found: " + err.Error()})
		}

		currency.Name = req.Name
		currency.Country = req.Country
		currency.CurrencyCode = req.CurrencyCode
		currency.Symbol = req.Symbol
		currency.Emoji = req.Emoji
		currency.Timezone = req.Timezone
		currency.UpdatedAt = time.Now().UTC()

		if err := core.CurrencyManager(service).UpdateByID(context, currency.ID, currency); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update currency failed: update error: " + err.Error(),
				Module:      "Currency",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update currency: " + err.Error()})
		}

		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Updated currency: " + currency.Name + " (" + currency.CurrencyCode + ")",
			Module:      "Currency",
		})

		return ctx.JSON(http.StatusOK, core.CurrencyManager(service).ToModel(currency))
	})

	req.RegisterWebRoute(horizon.Route{
		Route:  "/api/v1/currency/:currency_id",
		Method: "DELETE",
		Note:   "Deletes a currency by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		currencyID, err := helpers.EngineUUIDParam(ctx, "currency_id")
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete currency failed: invalid currency_id: " + err.Error(),
				Module:      "Currency",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid currency_id: " + err.Error()})
		}

		currency, err := core.CurrencyManager(service).GetByID(context, *currencyID)
		if err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete currency failed: not found: " + err.Error(),
				Module:      "Currency",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Currency not found: " + err.Error()})
		}

		if err := core.CurrencyManager(service).Delete(context, *currencyID); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete currency failed: delete error: " + err.Error(),
				Module:      "Currency",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete currency: " + err.Error()})
		}

		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "delete-success",
			Description: "Deleted currency: " + currency.Name + " (" + currency.CurrencyCode + ")",
			Module:      "Currency",
		})

		return ctx.NoContent(http.StatusNoContent)
	})

	req.RegisterWebRoute(horizon.Route{
		Route:       "/api/v1/currency/bulk-delete",
		Method:      "DELETE",
		RequestType: core.IDSRequest{},
		Note:        "Deletes multiple currency records.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var reqBody core.IDSRequest

		if err := ctx.Bind(&reqBody); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete currencies failed: invalid request body. " + err.Error(),
				Module:      "Currency",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body: " + err.Error()})
		}

		if len(reqBody.IDs) == 0 {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete currencies failed: no IDs provided.",
				Module:      "Currency",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "No IDs provided for deletion."})
		}

		ids := make([]any, len(reqBody.IDs))
		for i, id := range reqBody.IDs {
			ids[i] = id
		}
		if err := core.CurrencyManager(service).BulkDelete(context, ids); err != nil {
			event.Footstep(ctx, service, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete currencies failed: " + err.Error(),
				Module:      "Currency",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to bulk delete currencies: " + err.Error()})
		}

		event.Footstep(ctx, service, event.FootstepEvent{
			Activity:    "bulk-delete-success",
			Description: "Bulk deleted currencies.",
			Module:      "Currency",
		})

		return ctx.NoContent(http.StatusNoContent)
	})
	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/currency/exchange-rate/:currency_from_id/:currency_to_id/:amount",
		Method:       "POST",
		ResponseType: usecase.ExchangeResult{},
		Note:         "Computes exchange rate between two currencies for a given amount.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		currencyFromID, err := helpers.EngineUUIDParam(ctx, "currency_from_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid currency_from_id: " + err.Error()})
		}
		currencyToID, err := helpers.EngineUUIDParam(ctx, "currency_to_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid currency_to_id: " + err.Error()})
		}
		amountParam := ctx.Param("amount")
		var amount float64
		_, err = fmt.Sscanf(amountParam, "%f", &amount)
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid amount: " + err.Error()})
		}

		fromCurrency, err := core.CurrencyManager(service).GetByID(context, *currencyFromID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Currency from not found: " + err.Error()})
		}
		toCurrency, err := core.CurrencyManager(service).GetByID(context, *currencyToID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Currency to not found: " + err.Error()})
		}

		result, err := usecase.ExchangeRateComputeAmount(*fromCurrency, *toCurrency, amount)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to compute exchange rate: " + err.Error()})
		}

		return ctx.JSON(http.StatusOK, result)
	})

	req.RegisterWebRoute(horizon.Route{
		Route:        "/api/v1/currency/timezone/:timezone",
		Method:       "GET",
		ResponseType: core.CurrencyResponse{},
		Note:         "Returns the currency for a given timezone.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		timezone := ctx.Param("timezone")
		if timezone == "" {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Timezone is required"})
		}
		currency, err := core.CurrencyManager(service).FindOneRaw(context, &core.Currency{Timezone: timezone})
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Currency not found for timezone: " + err.Error()})
		}

		return ctx.JSON(http.StatusOK, currency)
	})
}
