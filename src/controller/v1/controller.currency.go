package controller_v1

import (
	"fmt"
	"net/http"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/services/handlers"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/event"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/model/model_core"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/service"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func (c *Controller) CurrencyController() {
	req := c.provider.Service.Request

	// Get all currencies
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/currency",
		Method:       "GET",
		ResponseType: model_core.CurrencyResponse{},
		Note:         "Returns all currencies.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		currencies, err := c.model_core.CurrencyManager.List(context)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve currencies: " + err.Error()})
		}
		return ctx.JSON(http.StatusOK, c.model_core.CurrencyManager.Filtered(context, ctx, currencies))
	})

	// Get a currency by its ID
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/currency/:currency_id",
		Method:       "GET",
		ResponseType: model_core.CurrencyResponse{},
		Note:         "Returns a specific currency by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		currencyID, err := handlers.EngineUUIDParam(ctx, "currency_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid currency_id: " + err.Error()})
		}

		currency, err := c.model_core.CurrencyManager.GetByIDRaw(context, *currencyID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Currency not found: " + err.Error()})
		}

		return ctx.JSON(http.StatusOK, currency)
	})

	// Get a currency by its code
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/currency/code/:currency_code",
		Method:       "GET",
		ResponseType: model_core.CurrencyResponse{},
		Note:         "Returns a specific currency by its code (e.g., USD, EUR).",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		currencyCode := ctx.Param("currency_code")
		if currencyCode == "" {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Currency code is required"})
		}

		currency, err := c.model_core.CurrencyFindByCode(context, currencyCode)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Currency not found: " + err.Error()})
		}

		return ctx.JSON(http.StatusOK, c.model_core.CurrencyManager.ToModel(currency))
	})

	// Create a new currency
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/currency",
		Method:       "POST",
		ResponseType: model_core.CurrencyResponse{},
		RequestType:  model_core.CurrencyRequest{},
		Note:         "Creates a new currency.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		req, err := c.model_core.CurrencyManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create currency failed: validation error: " + err.Error(),
				Module:      "Currency",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}

		currency := &model_core.Currency{
			Name:         req.Name,
			Country:      req.Country,
			CurrencyCode: req.CurrencyCode,
			Symbol:       req.Symbol,
			Emoji:        req.Emoji,
			CreatedAt:    time.Now().UTC(),
			UpdatedAt:    time.Now().UTC(),
		}

		if err := c.model_core.CurrencyManager.Create(context, currency); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "create-error",
				Description: "Create currency failed: create error: " + err.Error(),
				Module:      "Currency",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create currency: " + err.Error()})
		}

		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "create-success",
			Description: "Created currency: " + currency.Name + " (" + currency.CurrencyCode + ")",
			Module:      "Currency",
		})

		return ctx.JSON(http.StatusOK, c.model_core.CurrencyManager.ToModel(currency))
	})

	// Update a currency by its ID
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/currency/:currency_id",
		Method:       "PUT",
		ResponseType: model_core.CurrencyResponse{},
		RequestType:  model_core.CurrencyRequest{},
		Note:         "Updates an existing currency by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		currencyID, err := handlers.EngineUUIDParam(ctx, "currency_id")
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update currency failed: invalid currency_id: " + err.Error(),
				Module:      "Currency",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid currency_id: " + err.Error()})
		}

		req, err := c.model_core.CurrencyManager.Validate(ctx)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update currency failed: validation error: " + err.Error(),
				Module:      "Currency",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed: " + err.Error()})
		}

		currency, err := c.model_core.CurrencyManager.GetByID(context, *currencyID)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
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
		currency.UpdatedAt = time.Now().UTC()

		if err := c.model_core.CurrencyManager.UpdateFields(context, currency.ID, currency); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "update-error",
				Description: "Update currency failed: update error: " + err.Error(),
				Module:      "Currency",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update currency: " + err.Error()})
		}

		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "update-success",
			Description: "Updated currency: " + currency.Name + " (" + currency.CurrencyCode + ")",
			Module:      "Currency",
		})

		return ctx.JSON(http.StatusOK, c.model_core.CurrencyManager.ToModel(currency))
	})

	// Delete a currency by its ID
	req.RegisterRoute(handlers.Route{
		Route:  "/api/v1/currency/:currency_id",
		Method: "DELETE",
		Note:   "Deletes a currency by its ID.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		currencyID, err := handlers.EngineUUIDParam(ctx, "currency_id")
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete currency failed: invalid currency_id: " + err.Error(),
				Module:      "Currency",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid currency_id: " + err.Error()})
		}

		currency, err := c.model_core.CurrencyManager.GetByID(context, *currencyID)
		if err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete currency failed: not found: " + err.Error(),
				Module:      "Currency",
			})
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Currency not found: " + err.Error()})
		}

		if err := c.model_core.CurrencyManager.DeleteByID(context, *currencyID); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "delete-error",
				Description: "Delete currency failed: delete error: " + err.Error(),
				Module:      "Currency",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete currency: " + err.Error()})
		}

		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "delete-success",
			Description: "Deleted currency: " + currency.Name + " (" + currency.CurrencyCode + ")",
			Module:      "Currency",
		})

		return ctx.NoContent(http.StatusNoContent)
	})

	// Bulk delete currencies by IDs
	req.RegisterRoute(handlers.Route{
		Route:       "/api/v1/currency/bulk-delete",
		Method:      "DELETE",
		RequestType: model_core.IDSRequest{},
		Note:        "Deletes multiple currency records.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		var reqBody model_core.IDSRequest

		if err := ctx.Bind(&reqBody); err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete currencies failed: invalid request body.",
				Module:      "Currency",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body: " + err.Error()})
		}

		if len(reqBody.IDs) == 0 {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete currencies failed: no IDs provided.",
				Module:      "Currency",
			})
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "No IDs provided for deletion."})
		}

		tx := c.provider.Service.Database.Client().Begin()
		if tx.Error != nil {
			tx.Rollback()
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete currencies failed: begin tx error: " + tx.Error.Error(),
				Module:      "Currency",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to begin transaction: " + tx.Error.Error()})
		}

		names := ""
		for _, rawID := range reqBody.IDs {
			currencyID, err := uuid.Parse(rawID)
			if err != nil {
				tx.Rollback()
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "bulk-delete-error",
					Description: "Bulk delete currencies failed: invalid UUID: " + rawID,
					Module:      "Currency",
				})
				return ctx.JSON(http.StatusBadRequest, map[string]string{"error": fmt.Sprintf("Invalid UUID: %s - %v", rawID, err)})
			}

			currency, err := c.model_core.CurrencyManager.GetByID(context, currencyID)
			if err != nil {
				tx.Rollback()
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "bulk-delete-error",
					Description: "Bulk delete currencies failed: not found: " + rawID,
					Module:      "Currency",
				})
				return ctx.JSON(http.StatusNotFound, map[string]string{"error": fmt.Sprintf("Currency with ID %s not found: %v", rawID, err)})
			}

			names += currency.Name + " (" + currency.CurrencyCode + "),"
			if err := c.model_core.CurrencyManager.DeleteByIDWithTx(context, tx, currencyID); err != nil {
				tx.Rollback()
				c.event.Footstep(context, ctx, event.FootstepEvent{
					Activity:    "bulk-delete-error",
					Description: "Bulk delete currencies failed: delete error: " + err.Error(),
					Module:      "Currency",
				})
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": fmt.Sprintf("Failed to delete currency with ID %s: %v", rawID, err)})
			}
		}

		if err := tx.Commit().Error; err != nil {
			c.event.Footstep(context, ctx, event.FootstepEvent{
				Activity:    "bulk-delete-error",
				Description: "Bulk delete currencies failed: commit tx error: " + err.Error(),
				Module:      "Currency",
			})
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to commit transaction: " + err.Error()})
		}

		c.event.Footstep(context, ctx, event.FootstepEvent{
			Activity:    "bulk-delete-success",
			Description: "Bulk deleted currencies: " + names,
			Module:      "Currency",
		})

		return ctx.NoContent(http.StatusNoContent)
	})

	// POST /api/v1/currency/exchange-rate/:currency_from_id/:currency_to_id/:amount
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/currency/exchange-rate/:currency_from_id/:currency_to_id/:amount",
		Method:       "POST",
		ResponseType: service.ExchangeResult{},
		Note:         "Computes exchange rate between two currencies for a given amount.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		currencyFromID, err := handlers.EngineUUIDParam(ctx, "currency_from_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid currency_from_id: " + err.Error()})
		}
		currencyToID, err := handlers.EngineUUIDParam(ctx, "currency_to_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid currency_to_id: " + err.Error()})
		}
		amountParam := ctx.Param("amount")
		var amount float64
		_, err = fmt.Sscanf(amountParam, "%f", &amount)
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid amount: " + err.Error()})
		}

		fromCurrency, err := c.model_core.CurrencyManager.GetByID(context, *currencyFromID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Currency from not found: " + err.Error()})
		}
		toCurrency, err := c.model_core.CurrencyManager.GetByID(context, *currencyToID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Currency to not found: " + err.Error()})
		}

		result, err := c.service.ExchangeRateComputeAmount(*fromCurrency, *toCurrency, amount)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to compute exchange rate: " + err.Error()})
		}

		return ctx.JSON(http.StatusOK, result)
	})

	// GET /api/v1/currency/country-code/:country_code
	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/currency/country-code/:country_code",
		Method:       "GET",
		ResponseType: model_core.CurrencyResponse{},
		Note:         "Returns the currency for a given country code.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		countryCode := ctx.Param("country_code")
		if countryCode == "" {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Country code is required"})
		}
		currency, err := c.model_core.CurrencyFindByCode(context, countryCode)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Currency not found for country code: " + err.Error()})
		}

		return ctx.JSON(http.StatusOK, c.model_core.CurrencyManager.ToModel(currency))
	})
}
