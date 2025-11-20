package v1

import (
	"net/http"

	"github.com/Lands-Horizon-Corp/e-coop-server/services/handlers"
	"github.com/labstack/echo/v4"
)

type GovernmentIDResponse struct {
	Name string `json:"name"`

	HasExpiryDate bool `json:"has_expiry_date"`

	FieldName string `json:"field_name"`
	HasNumber bool   `json:"has_number"`
}

func (c *Controller) CommonController() {
	req := c.provider.Service.Request

	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/government-ids/:currency_id",
		Method:       "GET",
		ResponseType: GovernmentIDResponse{},
		Note:         "Retrieves a list of all government IDs available in the system.",
	}, func(ctx echo.Context) error {
		context := ctx.Request().Context()
		currencyID, err := handlers.EngineUUIDParam(ctx, "currency_id")
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid currency_id: " + err.Error()})
		}
		currency, err := c.core.CurrencyManager.GetByID(context, *currencyID)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Currency not found: " + err.Error()})
		}
		result := []GovernmentIDResponse{}
		switch currency.CurrencyCode {
		case "USD": // United States
		case "EUR": // European Union (Germany as representative)
		case "JPY": // Japan
		case "GBP": // United Kingdom
		case "AUD": // Australia
		case "CAD": // Canada
		case "CHF": // Switzerland
		case "CNY": // China
		case "SEK": // Sweden
		case "NZD": // New Zealand
		case "PHP": // Philippines
		case "INR": // India
		case "KRW": // South Korea
		case "THB": // Thailand
		case "SGD": // Singapore
		case "HKD": // Hong Kong
		case "MYR": // Malaysia
		case "IDR": // Indonesia
		case "VND": // Vietnam
		case "TWD": // Taiwan
		case "BND": // Brunei
		case "SAR": // Saudi Arabia
		case "AED": // United Arab Emirates
		case "ILS": // Israel
		case "ZAR": // South Africa
		case "EGP": // Egypt
		case "TRY": // Turkey
		case "XOF": // West African CFA Franc (e.g., Senegal, Côte d'Ivoire)
		case "XAF": // Central African CFA Franc (e.g., Cameroon, Gabon)
		case "MUR": // Mauritius
		case "MVR": // Maldives
		case "NOK": // Norway
		case "DKK": // Denmark
		case "PLN": // Poland
		case "CZK": // Czech Republic
		case "HUF": // Hungary
		case "RUB": // Russia
		case "EUR-HR": // Croatia (Euro)
		case "BRL": // Brazil
		case "MXN": // Mexico
		case "ARS": // Argentina
		case "CLP": // Chile
		case "COP": // Colombia
		case "PEN": // Peru
		case "UYU": // Uruguay
		case "DOP": // Dominican Republic
		case "PYG": // Paraguay
		case "BOB": // Bolivia
		case "VES": // Venezuela
		case "PKR": // Pakistan
		case "BDT": // Bangladesh
		case "LKR": // Sri Lanka
		case "NPR": // Nepal
		case "MMK": // Myanmar
		case "KHR": // Cambodia
		case "LAK": // Laos
		case "NGN": // Nigeria
		case "KES": // Kenya
		case "GHS": // Ghana
		case "MAD": // Morocco
		case "TND": // Tunisia
		case "ETB": // Ethiopia
		case "DZD": // Algeria
		case "UAH": // Ukraine
		case "RON": // Romania
		case "BGN": // Bulgaria
		case "RSD": // Serbia
		case "ISK": // Iceland
		case "BYN": // Belarus
		case "FJD": // Fiji
		case "PGK": // Papua New Guinea
		case "JMD": // Jamaica
		case "CRC": // Costa Rica
		case "GTQ": // Guatemala
		case "XDR": // Special Drawing Rights (IMF)
		case "KWD": // Kuwait
		case "QAR": // Qatar
		case "OMR": // Oman
		case "BHD": // Bahrain
		case "JOD": // Jordan
		case "KZT": // Kazakhstan
		}
		return ctx.JSON(http.StatusNoContent, result)
	})
}
