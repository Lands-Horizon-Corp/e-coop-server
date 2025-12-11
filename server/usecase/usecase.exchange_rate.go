package usecase

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/server/model/core"
	"github.com/rotisserie/eris"
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

func fetchJSON(rawURL string) (map[string]any, error) {
	// Validate URL before making request to reduce risk flagged by gosec (G107)
	parsed, err := url.ParseRequestURI(rawURL)
	if err != nil {
		return nil, eris.Wrap(err, "invalid url")
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return nil, eris.Errorf("unsupported url scheme: %s", parsed.Scheme)
	}
	if parsed.Host == "" {
		return nil, eris.New("invalid url host")
	}

	resp, err := http.Get(parsed.String())
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, eris.Errorf("unexpected response: %s\nBody: %s", resp.Status, string(body))
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
			return nil, eris.Wrap(err, "both sources failed")
		}
	}

	// Get date and nested currency map
	dateStr, _ := data["date"].(string)
	currencies, ok := data[base].(map[string]any)
	if !ok {
		return nil, eris.Errorf("invalid base currency data for %s", base)
	}

	// Get rate
	rateVal, ok := currencies[target].(float64)
	if !ok {
		return nil, eris.Errorf("invalid or missing rate for %s", target)
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
func (s *UsecaseService) ExchangeRateComputeAmount(
	fromCurrency core.Currency,
	toCurrency core.Currency,
	amount float64) (*ExchangeResult, error) {

	fromCurrencyStr := fromCurrency.CurrencyCode
	toCurrencyStr := toCurrency.CurrencyCode

	result, err := GetExchangeRate(fromCurrencyStr, toCurrencyStr, amount)
	if err != nil {
		return nil, eris.Wrap(err, "failed to get exchange rate")
	}

	return result, nil
}
