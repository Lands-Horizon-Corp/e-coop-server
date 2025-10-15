package model_core

import (
	"context"
	"fmt"
	"time"

	horizon_services "github.com/Lands-Horizon-Corp/e-coop-server/services"
	"github.com/google/uuid"
	"github.com/rotisserie/eris"
	"gorm.io/gorm"
)

type (
	Currency struct {
		ID        uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
		CreatedAt time.Time      `gorm:"not null;default:now()" json:"created_at"`
		UpdatedAt time.Time      `gorm:"not null;default:now()" json:"updated_at"`
		DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at"`

		Name         string `gorm:"type:varchar(255);not null;unique" json:"name"`
		Country      string `gorm:"type:varchar(255);not null" json:"country"`
		CurrencyCode string `gorm:"type:varchar(10);not null;unique" json:"currency_code"`
		Symbol       string `gorm:"type:varchar(10)" json:"symbol"`
	}

	CurrencyResponse struct {
		ID           uuid.UUID `json:"id"`
		CreatedAt    string    `json:"created_at"`
		UpdatedAt    string    `json:"updated_at"`
		Name         string    `json:"name"`
		Country      string    `json:"country"`
		CurrencyCode string    `json:"currency_code"`
		Symbol       string    `json:"symbol"`
	}

	CurrencyRequest struct {
		Name         string `json:"name" validate:"required,min=1,max=255"`
		Country      string `json:"country" validate:"required,min=1,max=255"`
		CurrencyCode string `json:"currency_code" validate:"required,min=2,max=10"`
		Symbol       string `json:"symbol,omitempty"`
	}
)

func (m *ModelCore) Currency() {
	m.Migration = append(m.Migration, &Currency{})
	m.CurrencyManager = horizon_services.NewRepository(horizon_services.RepositoryParams[Currency, CurrencyResponse, CurrencyRequest]{
		Service: m.provider.Service,
		Resource: func(data *Currency) *CurrencyResponse {
			if data == nil {
				return nil
			}
			return &CurrencyResponse{
				ID:           data.ID,
				CreatedAt:    data.CreatedAt.Format(time.RFC3339),
				UpdatedAt:    data.UpdatedAt.Format(time.RFC3339),
				Name:         data.Name,
				Country:      data.Country,
				CurrencyCode: data.CurrencyCode,
				Symbol:       data.Symbol,
			}
		},
		Created: func(data *Currency) []string {
			return []string{
				"currency.create",
				fmt.Sprintf("currency.create.%s", data.ID),
				fmt.Sprintf("currency.create.code.%s", data.CurrencyCode),
			}
		},
		Updated: func(data *Currency) []string {
			return []string{
				"currency.update",
				fmt.Sprintf("currency.update.%s", data.ID),
				fmt.Sprintf("currency.update.code.%s", data.CurrencyCode),
			}
		},
		Deleted: func(data *Currency) []string {
			return []string{
				"currency.delete",
				fmt.Sprintf("currency.delete.%s", data.ID),
				fmt.Sprintf("currency.delete.code.%s", data.CurrencyCode),
			}
		},
	})
}

func (m *ModelCore) CurrencySeed(context context.Context, tx *gorm.DB) error {
	now := time.Now().UTC()
	currencies := []*Currency{
		// Major World Currencies
		{CreatedAt: now, UpdatedAt: now, Name: "US Dollar", Country: "United States", CurrencyCode: "USD", Symbol: "US$"},
		{CreatedAt: now, UpdatedAt: now, Name: "Euro", Country: "European Union", CurrencyCode: "EUR", Symbol: "€"},
		{CreatedAt: now, UpdatedAt: now, Name: "Japanese Yen", Country: "Japan", CurrencyCode: "JPY", Symbol: "¥"},
		{CreatedAt: now, UpdatedAt: now, Name: "British Pound Sterling", Country: "United Kingdom", CurrencyCode: "GBP", Symbol: "£"},
		{CreatedAt: now, UpdatedAt: now, Name: "Australian Dollar", Country: "Australia", CurrencyCode: "AUD", Symbol: "AU$"},
		{CreatedAt: now, UpdatedAt: now, Name: "Canadian Dollar", Country: "Canada", CurrencyCode: "CAD", Symbol: "CA$"},
		{CreatedAt: now, UpdatedAt: now, Name: "Swiss Franc", Country: "Switzerland", CurrencyCode: "CHF", Symbol: "Fr"},
		{CreatedAt: now, UpdatedAt: now, Name: "Chinese Yuan", Country: "China", CurrencyCode: "CNY", Symbol: "CN¥"},
		{CreatedAt: now, UpdatedAt: now, Name: "Swedish Krona", Country: "Sweden", CurrencyCode: "SEK", Symbol: "kr"},
		{CreatedAt: now, UpdatedAt: now, Name: "New Zealand Dollar", Country: "New Zealand", CurrencyCode: "NZD", Symbol: "NZ$"},

		// Asian Currencies
		{CreatedAt: now, UpdatedAt: now, Name: "Philippine Peso", Country: "Philippines", CurrencyCode: "PHP", Symbol: "₱"},
		{CreatedAt: now, UpdatedAt: now, Name: "Indian Rupee", Country: "India", CurrencyCode: "INR", Symbol: "₹"},
		{CreatedAt: now, UpdatedAt: now, Name: "South Korean Won", Country: "South Korea", CurrencyCode: "KRW", Symbol: "₩"},
		{CreatedAt: now, UpdatedAt: now, Name: "Thai Baht", Country: "Thailand", CurrencyCode: "THB", Symbol: "฿"},
		{CreatedAt: now, UpdatedAt: now, Name: "Singapore Dollar", Country: "Singapore", CurrencyCode: "SGD", Symbol: "S$"},
		{CreatedAt: now, UpdatedAt: now, Name: "Hong Kong Dollar", Country: "Hong Kong", CurrencyCode: "HKD", Symbol: "HK$"},
		{CreatedAt: now, UpdatedAt: now, Name: "Malaysian Ringgit", Country: "Malaysia", CurrencyCode: "MYR", Symbol: "RM"},
		{CreatedAt: now, UpdatedAt: now, Name: "Indonesian Rupiah", Country: "Indonesia", CurrencyCode: "IDR", Symbol: "Rp"},
		{CreatedAt: now, UpdatedAt: now, Name: "Vietnamese Dong", Country: "Vietnam", CurrencyCode: "VND", Symbol: "₫"},
		{CreatedAt: now, UpdatedAt: now, Name: "Taiwan Dollar", Country: "Taiwan", CurrencyCode: "TWD", Symbol: "NT$"},
		{CreatedAt: now, UpdatedAt: now, Name: "Brunei Dollar", Country: "Brunei", CurrencyCode: "BND", Symbol: "B$"},

		// Middle Eastern & African Currencies
		{CreatedAt: now, UpdatedAt: now, Name: "Saudi Riyal", Country: "Saudi Arabia", CurrencyCode: "SAR", Symbol: "ر.س"},
		{CreatedAt: now, UpdatedAt: now, Name: "UAE Dirham", Country: "United Arab Emirates", CurrencyCode: "AED", Symbol: "د.إ"},
		{CreatedAt: now, UpdatedAt: now, Name: "Israeli New Shekel", Country: "Israel", CurrencyCode: "ILS", Symbol: "₪"},
		{CreatedAt: now, UpdatedAt: now, Name: "South African Rand", Country: "South Africa", CurrencyCode: "ZAR", Symbol: "R"},
		{CreatedAt: now, UpdatedAt: now, Name: "Egyptian Pound", Country: "Egypt", CurrencyCode: "EGP", Symbol: "ج.م"},
		{CreatedAt: now, UpdatedAt: now, Name: "Turkish Lira", Country: "Turkey", CurrencyCode: "TRY", Symbol: "₺"},
		{CreatedAt: now, UpdatedAt: now, Name: "West African CFA Franc", Country: "West African States", CurrencyCode: "XOF", Symbol: "CFA"},
		{CreatedAt: now, UpdatedAt: now, Name: "Central African CFA Franc", Country: "Central African States", CurrencyCode: "XAF", Symbol: "CFA"},
		{CreatedAt: now, UpdatedAt: now, Name: "Mauritian Rupee", Country: "Mauritius", CurrencyCode: "MUR", Symbol: "₨"},
		{CreatedAt: now, UpdatedAt: now, Name: "Maldivian Rufiyaa", Country: "Maldives", CurrencyCode: "MVR", Symbol: "Rf"},

		// European Currencies (Non-Euro)
		{CreatedAt: now, UpdatedAt: now, Name: "Norwegian Krone", Country: "Norway", CurrencyCode: "NOK", Symbol: "kr"},
		{CreatedAt: now, UpdatedAt: now, Name: "Danish Krone", Country: "Denmark", CurrencyCode: "DKK", Symbol: "kr"},
		{CreatedAt: now, UpdatedAt: now, Name: "Polish Zloty", Country: "Poland", CurrencyCode: "PLN", Symbol: "zł"},
		{CreatedAt: now, UpdatedAt: now, Name: "Czech Koruna", Country: "Czech Republic", CurrencyCode: "CZK", Symbol: "Kč"},
		{CreatedAt: now, UpdatedAt: now, Name: "Hungarian Forint", Country: "Hungary", CurrencyCode: "HUF", Symbol: "Ft"},
		{CreatedAt: now, UpdatedAt: now, Name: "Russian Ruble", Country: "Russia", CurrencyCode: "RUB", Symbol: "₽"},
		{CreatedAt: now, UpdatedAt: now, Name: "Euro", Country: "Croatia", CurrencyCode: "EUR", Symbol: "€"},

		// Latin American Currencies
		{CreatedAt: now, UpdatedAt: now, Name: "Brazilian Real", Country: "Brazil", CurrencyCode: "BRL", Symbol: "R$"},
		{CreatedAt: now, UpdatedAt: now, Name: "Mexican Peso", Country: "Mexico", CurrencyCode: "MXN", Symbol: "MX$"},
		{CreatedAt: now, UpdatedAt: now, Name: "Argentine Peso", Country: "Argentina", CurrencyCode: "ARS", Symbol: "AR$"},
		{CreatedAt: now, UpdatedAt: now, Name: "Chilean Peso", Country: "Chile", CurrencyCode: "CLP", Symbol: "CL$"},
		{CreatedAt: now, UpdatedAt: now, Name: "Colombian Peso", Country: "Colombia", CurrencyCode: "COP", Symbol: "CO$"},
		{CreatedAt: now, UpdatedAt: now, Name: "Peruvian Sol", Country: "Peru", CurrencyCode: "PEN", Symbol: "S/"},
		{CreatedAt: now, UpdatedAt: now, Name: "Uruguayan Peso", Country: "Uruguay", CurrencyCode: "UYU", Symbol: "$U"},
		{CreatedAt: now, UpdatedAt: now, Name: "Dominican Peso", Country: "Dominican Republic", CurrencyCode: "DOP", Symbol: "RD$"},
		{CreatedAt: now, UpdatedAt: now, Name: "Paraguayan Guarani", Country: "Paraguay", CurrencyCode: "PYG", Symbol: "₲"},
		{CreatedAt: now, UpdatedAt: now, Name: "Bolivian Boliviano", Country: "Bolivia", CurrencyCode: "BOB", Symbol: "Bs"},
		{CreatedAt: now, UpdatedAt: now, Name: "Venezuelan Bolívar", Country: "Venezuela", CurrencyCode: "VES", Symbol: "Bs.S"},

		// Other Major Currencies
		{CreatedAt: now, UpdatedAt: now, Name: "Bitcoin", Country: "Global", CurrencyCode: "BTC", Symbol: "₿"},
		{CreatedAt: now, UpdatedAt: now, Name: "Ethereum", Country: "Global", CurrencyCode: "ETH", Symbol: "Ξ"},
		{CreatedAt: now, UpdatedAt: now, Name: "Gold Ounce", Country: "Global", CurrencyCode: "XAU", Symbol: "Au"},
		{CreatedAt: now, UpdatedAt: now, Name: "Silver Ounce", Country: "Global", CurrencyCode: "XAG", Symbol: "Ag"},

		// Additional Asian Currencies
		{CreatedAt: now, UpdatedAt: now, Name: "Pakistani Rupee", Country: "Pakistan", CurrencyCode: "PKR", Symbol: "₨"},
		{CreatedAt: now, UpdatedAt: now, Name: "Bangladeshi Taka", Country: "Bangladesh", CurrencyCode: "BDT", Symbol: "৳"},
		{CreatedAt: now, UpdatedAt: now, Name: "Sri Lankan Rupee", Country: "Sri Lanka", CurrencyCode: "LKR", Symbol: "Rs"},
		{CreatedAt: now, UpdatedAt: now, Name: "Nepalese Rupee", Country: "Nepal", CurrencyCode: "NPR", Symbol: "Rs"},
		{CreatedAt: now, UpdatedAt: now, Name: "Myanmar Kyat", Country: "Myanmar", CurrencyCode: "MMK", Symbol: "K"},
		{CreatedAt: now, UpdatedAt: now, Name: "Cambodian Riel", Country: "Cambodia", CurrencyCode: "KHR", Symbol: "៛"},
		{CreatedAt: now, UpdatedAt: now, Name: "Laotian Kip", Country: "Laos", CurrencyCode: "LAK", Symbol: "₭"},

		// Additional African Currencies
		{CreatedAt: now, UpdatedAt: now, Name: "Nigerian Naira", Country: "Nigeria", CurrencyCode: "NGN", Symbol: "₦"},
		{CreatedAt: now, UpdatedAt: now, Name: "Kenyan Shilling", Country: "Kenya", CurrencyCode: "KES", Symbol: "KSh"},
		{CreatedAt: now, UpdatedAt: now, Name: "Ghanaian Cedi", Country: "Ghana", CurrencyCode: "GHS", Symbol: "₵"},
		{CreatedAt: now, UpdatedAt: now, Name: "Moroccan Dirham", Country: "Morocco", CurrencyCode: "MAD", Symbol: "د.م."},
		{CreatedAt: now, UpdatedAt: now, Name: "Tunisian Dinar", Country: "Tunisia", CurrencyCode: "TND", Symbol: "د.ت"},
		{CreatedAt: now, UpdatedAt: now, Name: "Ethiopian Birr", Country: "Ethiopia", CurrencyCode: "ETB", Symbol: "Br"},
		{CreatedAt: now, UpdatedAt: now, Name: "Algerian Dinar", Country: "Algeria", CurrencyCode: "DZD", Symbol: "د.ج"},

		// Additional European Currencies
		{CreatedAt: now, UpdatedAt: now, Name: "Ukrainian Hryvnia", Country: "Ukraine", CurrencyCode: "UAH", Symbol: "₴"},
		{CreatedAt: now, UpdatedAt: now, Name: "Romanian Leu", Country: "Romania", CurrencyCode: "RON", Symbol: "lei"},
		{CreatedAt: now, UpdatedAt: now, Name: "Bulgarian Lev", Country: "Bulgaria", CurrencyCode: "BGN", Symbol: "лв"},
		{CreatedAt: now, UpdatedAt: now, Name: "Serbian Dinar", Country: "Serbia", CurrencyCode: "RSD", Symbol: "дин"},
		{CreatedAt: now, UpdatedAt: now, Name: "Icelandic Krona", Country: "Iceland", CurrencyCode: "ISK", Symbol: "kr"},
		{CreatedAt: now, UpdatedAt: now, Name: "Belarusian Ruble", Country: "Belarus", CurrencyCode: "BYN", Symbol: "Br"},

		// Oceania & Others
		{CreatedAt: now, UpdatedAt: now, Name: "Fijian Dollar", Country: "Fiji", CurrencyCode: "FJD", Symbol: "FJ$"},
		{CreatedAt: now, UpdatedAt: now, Name: "Papua New Guinea Kina", Country: "Papua New Guinea", CurrencyCode: "PGK", Symbol: "K"},

		// Caribbean & Central America
		{CreatedAt: now, UpdatedAt: now, Name: "Jamaican Dollar", Country: "Jamaica", CurrencyCode: "JMD", Symbol: "J$"},
		{CreatedAt: now, UpdatedAt: now, Name: "Costa Rican Colon", Country: "Costa Rica", CurrencyCode: "CRC", Symbol: "₡"},
		{CreatedAt: now, UpdatedAt: now, Name: "Guatemalan Quetzal", Country: "Guatemala", CurrencyCode: "GTQ", Symbol: "Q"},

		// Special Drawing Rights
		{CreatedAt: now, UpdatedAt: now, Name: "Special Drawing Rights", Country: "IMF", CurrencyCode: "XDR", Symbol: "SDR"},

		// Middle Eastern Currencies
		{CreatedAt: now, UpdatedAt: now, Name: "Kuwaiti Dinar", Country: "Kuwait", CurrencyCode: "KWD", Symbol: "د.ك"},
		{CreatedAt: now, UpdatedAt: now, Name: "Qatari Riyal", Country: "Qatar", CurrencyCode: "QAR", Symbol: "ر.ق"},
		{CreatedAt: now, UpdatedAt: now, Name: "Omani Rial", Country: "Oman", CurrencyCode: "OMR", Symbol: "ر.ع"},
		{CreatedAt: now, UpdatedAt: now, Name: "Bahraini Dinar", Country: "Bahrain", CurrencyCode: "BHD", Symbol: "ب.د"},
		{CreatedAt: now, UpdatedAt: now, Name: "Jordanian Dinar", Country: "Jordan", CurrencyCode: "JOD", Symbol: "د.ا"},

		// Central Asian Currencies
		{CreatedAt: now, UpdatedAt: now, Name: "Kazakhstani Tenge", Country: "Kazakhstan", CurrencyCode: "KZT", Symbol: "₸"},
	}

	for _, currency := range currencies {
		if err := m.CurrencyManager.CreateWithTx(context, tx, currency); err != nil {
			return eris.Wrapf(err, "failed to seed currency %s", currency.Name)
		}
	}

	return nil
}

func (m *ModelCore) CurrencyFindAll(context context.Context) ([]*Currency, error) {
	return m.CurrencyManager.Find(context, &Currency{})
}

func (m *ModelCore) CurrencyFindByCode(context context.Context, currencyCode string) (*Currency, error) {
	currencies, err := m.CurrencyManager.Find(context, &Currency{CurrencyCode: currencyCode})
	if err != nil {
		return nil, err
	}
	if len(currencies) == 0 {
		return nil, eris.New("currency not found")
	}
	return currencies[0], nil
}
