package service

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	modelCore "github.com/Lands-Horizon-Corp/e-coop-server/src/model/model_core"
)

// ExchangeResult represents the result of a currency exchange operation
type ExchangeResult struct {
	From      string    `json:"from"`
	To        string    `json:"to"`
	Amount    float64   `json:"amount"`
	Rate      float64   `json:"rate"`
	Converted float64   `json:"converted"`
	Date      string    `json:"date"`
	FetchedAt time.Time `json:"fetched_at"`
}

func fetchJSON(url string) (map[string]any, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected response: %s\nBody: %s", resp.Status, string(body))
	}

	var data map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}
	return data, nil
}

// GetExchangeRate fetches the current exchange rate and converts the given amount between currencies
func GetExchangeRate(currencyFrom, currencyTo string, amount float64) (*ExchangeResult, error) {
	base := strings.ToLower(currencyFrom)
	target := strings.ToLower(currencyTo)

	// 1️⃣ Primary source (jsDelivr)
	mainURL := fmt.Sprintf("https://cdn.jsdelivr.net/npm/@fawazahmed0/currency-api@latest/v1/currencies/%s.json", base)
	// 2️⃣ Fallback (Cloudflare mirror)
	fallbackURL := fmt.Sprintf("https://latest.currency-api.pages.dev/v1/currencies/%s.json", base)

	data, err := fetchJSON(mainURL)
	if err != nil {
		log.Printf("Primary source failed, trying fallback... (%v)", err)
		data, err = fetchJSON(fallbackURL)
		if err != nil {
			return nil, fmt.Errorf("both sources failed: %w", err)
		}
	}

	// Get date and nested currency map
	dateStr, _ := data["date"].(string)
	currencies, ok := data[base].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("invalid base currency data for %s", base)
	}

	// Get rate
	rateVal, ok := currencies[target].(float64)
	if !ok {
		return nil, fmt.Errorf("invalid or missing rate for %s", target)
	}

	result := &ExchangeResult{
		From:      strings.ToUpper(currencyFrom),
		To:        strings.ToUpper(currencyTo),
		Amount:    amount,
		Rate:      rateVal,
		Converted: amount * rateVal,
		Date:      dateStr,
		FetchedAt: time.Now(),
	}
	return result, nil
}

// ExchangeRateComputeAmount computes the exchange rate and converts amount between two currencies
func (s *TransactionService) ExchangeRateComputeAmount(
	fromCurrency modelCore.Currency,
	toCurrency modelCore.Currency,
	amount float64) (*ExchangeResult, error) {

	fromCurrencyStr := fromCurrency.CurrencyCode
	toCurrencyStr := toCurrency.CurrencyCode

	result, err := GetExchangeRate(fromCurrencyStr, toCurrencyStr, amount)
	if err != nil {
		return nil, fmt.Errorf("failed to get exchange rate: %w", err)
	}

	return result, nil
}
