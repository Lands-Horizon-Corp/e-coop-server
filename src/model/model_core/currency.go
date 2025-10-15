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
		CurrencyCode string `gorm:"type:varchar(10);not null" json:"currency_code"`
		Symbol       string `gorm:"type:varchar(10)" json:"symbol"`
		Emoji        string `gorm:"type:varchar(10)" json:"emoji"`
	}

	CurrencyResponse struct {
		ID           uuid.UUID `json:"id"`
		CreatedAt    string    `json:"created_at"`
		UpdatedAt    string    `json:"updated_at"`
		Name         string    `json:"name"`
		Country      string    `json:"country"`
		CurrencyCode string    `json:"currency_code"`
		Symbol       string    `json:"symbol"`
		Emoji        string    `json:"emoji"`
	}

	CurrencyRequest struct {
		Name         string `json:"name" validate:"required,min=1,max=255"`
		Country      string `json:"country" validate:"required,min=1,max=255"`
		CurrencyCode string `json:"currency_code" validate:"required,min=2,max=10"`
		Symbol       string `json:"symbol,omitempty"`
		Emoji        string `json:"emoji,omitempty"`
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
				Emoji:        data.Emoji,
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
		{CreatedAt: now, UpdatedAt: now, Name: "US Dollar", Country: "United States", CurrencyCode: "USD", Symbol: "US$", Emoji: "🇺🇸"},
		{CreatedAt: now, UpdatedAt: now, Name: "Euro", Country: "European Union", CurrencyCode: "EUR", Symbol: "€", Emoji: "🇪🇺"},
		{CreatedAt: now, UpdatedAt: now, Name: "Japanese Yen", Country: "Japan", CurrencyCode: "JPY", Symbol: "¥", Emoji: "🇯🇵"},
		{CreatedAt: now, UpdatedAt: now, Name: "British Pound Sterling", Country: "United Kingdom", CurrencyCode: "GBP", Symbol: "£", Emoji: "🇬🇧"},
		{CreatedAt: now, UpdatedAt: now, Name: "Australian Dollar", Country: "Australia", CurrencyCode: "AUD", Symbol: "AU$", Emoji: "🇦🇺"},
		{CreatedAt: now, UpdatedAt: now, Name: "Canadian Dollar", Country: "Canada", CurrencyCode: "CAD", Symbol: "CA$", Emoji: "🇨🇦"},
		{CreatedAt: now, UpdatedAt: now, Name: "Swiss Franc", Country: "Switzerland", CurrencyCode: "CHF", Symbol: "Fr", Emoji: "🇨🇭"},
		{CreatedAt: now, UpdatedAt: now, Name: "Chinese Yuan", Country: "China", CurrencyCode: "CNY", Symbol: "CN¥", Emoji: "🇨🇳"},
		{CreatedAt: now, UpdatedAt: now, Name: "Swedish Krona", Country: "Sweden", CurrencyCode: "SEK", Symbol: "kr", Emoji: "🇸🇪"},
		{CreatedAt: now, UpdatedAt: now, Name: "New Zealand Dollar", Country: "New Zealand", CurrencyCode: "NZD", Symbol: "NZ$", Emoji: "🇳🇿"},

		// Asian Currencies
		{CreatedAt: now, UpdatedAt: now, Name: "Philippine Peso", Country: "Philippines", CurrencyCode: "PHP", Symbol: "₱", Emoji: "🇵🇭"},
		{CreatedAt: now, UpdatedAt: now, Name: "Indian Rupee", Country: "India", CurrencyCode: "INR", Symbol: "₹", Emoji: "🇮🇳"},
		{CreatedAt: now, UpdatedAt: now, Name: "South Korean Won", Country: "South Korea", CurrencyCode: "KRW", Symbol: "₩", Emoji: "🇰🇷"},
		{CreatedAt: now, UpdatedAt: now, Name: "Thai Baht", Country: "Thailand", CurrencyCode: "THB", Symbol: "฿", Emoji: "🇹🇭"},
		{CreatedAt: now, UpdatedAt: now, Name: "Singapore Dollar", Country: "Singapore", CurrencyCode: "SGD", Symbol: "S$", Emoji: "🇸🇬"},
		{CreatedAt: now, UpdatedAt: now, Name: "Hong Kong Dollar", Country: "Hong Kong", CurrencyCode: "HKD", Symbol: "HK$", Emoji: "🇭🇰"},
		{CreatedAt: now, UpdatedAt: now, Name: "Malaysian Ringgit", Country: "Malaysia", CurrencyCode: "MYR", Symbol: "RM", Emoji: "🇲🇾"},
		{CreatedAt: now, UpdatedAt: now, Name: "Indonesian Rupiah", Country: "Indonesia", CurrencyCode: "IDR", Symbol: "Rp", Emoji: "🇮🇩"},
		{CreatedAt: now, UpdatedAt: now, Name: "Vietnamese Dong", Country: "Vietnam", CurrencyCode: "VND", Symbol: "₫", Emoji: "🇻🇳"},
		{CreatedAt: now, UpdatedAt: now, Name: "Taiwan Dollar", Country: "Taiwan", CurrencyCode: "TWD", Symbol: "NT$", Emoji: "🇹🇼"},
		{CreatedAt: now, UpdatedAt: now, Name: "Brunei Dollar", Country: "Brunei", CurrencyCode: "BND", Symbol: "B$", Emoji: "🇧🇳"},

		// Middle Eastern & African Currencies
		{CreatedAt: now, UpdatedAt: now, Name: "Saudi Riyal", Country: "Saudi Arabia", CurrencyCode: "SAR", Symbol: "ر.س", Emoji: "🇸🇦"},
		{CreatedAt: now, UpdatedAt: now, Name: "UAE Dirham", Country: "United Arab Emirates", CurrencyCode: "AED", Symbol: "د.إ", Emoji: "🇦🇪"},
		{CreatedAt: now, UpdatedAt: now, Name: "Israeli New Shekel", Country: "Israel", CurrencyCode: "ILS", Symbol: "₪", Emoji: "🇮🇱"},
		{CreatedAt: now, UpdatedAt: now, Name: "South African Rand", Country: "South Africa", CurrencyCode: "ZAR", Symbol: "R", Emoji: "🇿🇦"},
		{CreatedAt: now, UpdatedAt: now, Name: "Egyptian Pound", Country: "Egypt", CurrencyCode: "EGP", Symbol: "ج.م", Emoji: "🇪🇬"},
		{CreatedAt: now, UpdatedAt: now, Name: "Turkish Lira", Country: "Turkey", CurrencyCode: "TRY", Symbol: "₺", Emoji: "🇹🇷"},
		{CreatedAt: now, UpdatedAt: now, Name: "West African CFA Franc", Country: "West African States", CurrencyCode: "XOF", Symbol: "CFA", Emoji: "🌍"},
		{CreatedAt: now, UpdatedAt: now, Name: "Central African CFA Franc", Country: "Central African States", CurrencyCode: "XAF", Symbol: "CFA", Emoji: "🌍"},
		{CreatedAt: now, UpdatedAt: now, Name: "Mauritian Rupee", Country: "Mauritius", CurrencyCode: "MUR", Symbol: "₨", Emoji: "🇲🇺"},
		{CreatedAt: now, UpdatedAt: now, Name: "Maldivian Rufiyaa", Country: "Maldives", CurrencyCode: "MVR", Symbol: "Rf", Emoji: "🇲🇻"},

		// European Currencies (Non-Euro)
		{CreatedAt: now, UpdatedAt: now, Name: "Norwegian Krone", Country: "Norway", CurrencyCode: "NOK", Symbol: "kr", Emoji: "🇳🇴"},
		{CreatedAt: now, UpdatedAt: now, Name: "Danish Krone", Country: "Denmark", CurrencyCode: "DKK", Symbol: "kr", Emoji: "🇩🇰"},
		{CreatedAt: now, UpdatedAt: now, Name: "Polish Zloty", Country: "Poland", CurrencyCode: "PLN", Symbol: "zł", Emoji: "🇵🇱"},
		{CreatedAt: now, UpdatedAt: now, Name: "Czech Koruna", Country: "Czech Republic", CurrencyCode: "CZK", Symbol: "Kč", Emoji: "🇨🇿"},
		{CreatedAt: now, UpdatedAt: now, Name: "Hungarian Forint", Country: "Hungary", CurrencyCode: "HUF", Symbol: "Ft", Emoji: "🇭🇺"},
		{CreatedAt: now, UpdatedAt: now, Name: "Russian Ruble", Country: "Russia", CurrencyCode: "RUB", Symbol: "₽", Emoji: "🇷🇺"},
		{CreatedAt: now, UpdatedAt: now, Name: "Euro (Croatia)", Country: "Croatia", CurrencyCode: "EUR", Symbol: "€", Emoji: "🇭🇷"},

		// Latin American Currencies
		{CreatedAt: now, UpdatedAt: now, Name: "Brazilian Real", Country: "Brazil", CurrencyCode: "BRL", Symbol: "R$", Emoji: "🇧🇷"},
		{CreatedAt: now, UpdatedAt: now, Name: "Mexican Peso", Country: "Mexico", CurrencyCode: "MXN", Symbol: "MX$", Emoji: "🇲🇽"},
		{CreatedAt: now, UpdatedAt: now, Name: "Argentine Peso", Country: "Argentina", CurrencyCode: "ARS", Symbol: "AR$", Emoji: "🇦🇷"},
		{CreatedAt: now, UpdatedAt: now, Name: "Chilean Peso", Country: "Chile", CurrencyCode: "CLP", Symbol: "CL$", Emoji: "🇨🇱"},
		{CreatedAt: now, UpdatedAt: now, Name: "Colombian Peso", Country: "Colombia", CurrencyCode: "COP", Symbol: "CO$", Emoji: "🇨🇴"},
		{CreatedAt: now, UpdatedAt: now, Name: "Peruvian Sol", Country: "Peru", CurrencyCode: "PEN", Symbol: "S/", Emoji: "🇵🇪"},
		{CreatedAt: now, UpdatedAt: now, Name: "Uruguayan Peso", Country: "Uruguay", CurrencyCode: "UYU", Symbol: "$U", Emoji: "🇺🇾"},
		{CreatedAt: now, UpdatedAt: now, Name: "Dominican Peso", Country: "Dominican Republic", CurrencyCode: "DOP", Symbol: "RD$", Emoji: "🇩🇴"},
		{CreatedAt: now, UpdatedAt: now, Name: "Paraguayan Guarani", Country: "Paraguay", CurrencyCode: "PYG", Symbol: "₲", Emoji: "🇵🇾"},
		{CreatedAt: now, UpdatedAt: now, Name: "Bolivian Boliviano", Country: "Bolivia", CurrencyCode: "BOB", Symbol: "Bs", Emoji: "🇧🇴"},
		{CreatedAt: now, UpdatedAt: now, Name: "Venezuelan Bolívar", Country: "Venezuela", CurrencyCode: "VES", Symbol: "Bs.S", Emoji: "🇻🇪"},

		// Other Major Currencies
		{CreatedAt: now, UpdatedAt: now, Name: "Bitcoin", Country: "Global", CurrencyCode: "BTC", Symbol: "₿", Emoji: "₿"},
		{CreatedAt: now, UpdatedAt: now, Name: "Ethereum", Country: "Global", CurrencyCode: "ETH", Symbol: "Ξ", Emoji: "⟠"},
		{CreatedAt: now, UpdatedAt: now, Name: "Gold Ounce", Country: "Global", CurrencyCode: "XAU", Symbol: "Au", Emoji: "🥇"},
		{CreatedAt: now, UpdatedAt: now, Name: "Silver Ounce", Country: "Global", CurrencyCode: "XAG", Symbol: "Ag", Emoji: "🥈"},

		// Additional Asian Currencies
		{CreatedAt: now, UpdatedAt: now, Name: "Pakistani Rupee", Country: "Pakistan", CurrencyCode: "PKR", Symbol: "₨", Emoji: "🇵🇰"},
		{CreatedAt: now, UpdatedAt: now, Name: "Bangladeshi Taka", Country: "Bangladesh", CurrencyCode: "BDT", Symbol: "৳", Emoji: "🇧🇩"},
		{CreatedAt: now, UpdatedAt: now, Name: "Sri Lankan Rupee", Country: "Sri Lanka", CurrencyCode: "LKR", Symbol: "Rs", Emoji: "🇱🇰"},
		{CreatedAt: now, UpdatedAt: now, Name: "Nepalese Rupee", Country: "Nepal", CurrencyCode: "NPR", Symbol: "Rs", Emoji: "🇳🇵"},
		{CreatedAt: now, UpdatedAt: now, Name: "Myanmar Kyat", Country: "Myanmar", CurrencyCode: "MMK", Symbol: "K", Emoji: "🇲🇲"},
		{CreatedAt: now, UpdatedAt: now, Name: "Cambodian Riel", Country: "Cambodia", CurrencyCode: "KHR", Symbol: "៛", Emoji: "🇰🇭"},
		{CreatedAt: now, UpdatedAt: now, Name: "Laotian Kip", Country: "Laos", CurrencyCode: "LAK", Symbol: "₭", Emoji: "🇱🇦"},

		// Additional African Currencies
		{CreatedAt: now, UpdatedAt: now, Name: "Nigerian Naira", Country: "Nigeria", CurrencyCode: "NGN", Symbol: "₦", Emoji: "🇳🇬"},
		{CreatedAt: now, UpdatedAt: now, Name: "Kenyan Shilling", Country: "Kenya", CurrencyCode: "KES", Symbol: "KSh", Emoji: "🇰🇪"},
		{CreatedAt: now, UpdatedAt: now, Name: "Ghanaian Cedi", Country: "Ghana", CurrencyCode: "GHS", Symbol: "₵", Emoji: "🇬🇭"},
		{CreatedAt: now, UpdatedAt: now, Name: "Moroccan Dirham", Country: "Morocco", CurrencyCode: "MAD", Symbol: "د.م.", Emoji: "🇲🇦"},
		{CreatedAt: now, UpdatedAt: now, Name: "Tunisian Dinar", Country: "Tunisia", CurrencyCode: "TND", Symbol: "د.ت", Emoji: "🇹🇳"},
		{CreatedAt: now, UpdatedAt: now, Name: "Ethiopian Birr", Country: "Ethiopia", CurrencyCode: "ETB", Symbol: "Br", Emoji: "🇪🇹"},
		{CreatedAt: now, UpdatedAt: now, Name: "Algerian Dinar", Country: "Algeria", CurrencyCode: "DZD", Symbol: "د.ج", Emoji: "🇩🇿"},

		// Additional European Currencies
		{CreatedAt: now, UpdatedAt: now, Name: "Ukrainian Hryvnia", Country: "Ukraine", CurrencyCode: "UAH", Symbol: "₴", Emoji: "🇺🇦"},
		{CreatedAt: now, UpdatedAt: now, Name: "Romanian Leu", Country: "Romania", CurrencyCode: "RON", Symbol: "lei", Emoji: "🇷🇴"},
		{CreatedAt: now, UpdatedAt: now, Name: "Bulgarian Lev", Country: "Bulgaria", CurrencyCode: "BGN", Symbol: "лв", Emoji: "🇧🇬"},
		{CreatedAt: now, UpdatedAt: now, Name: "Serbian Dinar", Country: "Serbia", CurrencyCode: "RSD", Symbol: "дин", Emoji: "🇷🇸"},
		{CreatedAt: now, UpdatedAt: now, Name: "Icelandic Krona", Country: "Iceland", CurrencyCode: "ISK", Symbol: "kr", Emoji: "🇮🇸"},
		{CreatedAt: now, UpdatedAt: now, Name: "Belarusian Ruble", Country: "Belarus", CurrencyCode: "BYN", Symbol: "Br", Emoji: "🇧🇾"},

		// Oceania & Others
		{CreatedAt: now, UpdatedAt: now, Name: "Fijian Dollar", Country: "Fiji", CurrencyCode: "FJD", Symbol: "FJ$", Emoji: "🇫🇯"},
		{CreatedAt: now, UpdatedAt: now, Name: "Papua New Guinea Kina", Country: "Papua New Guinea", CurrencyCode: "PGK", Symbol: "K", Emoji: "🇵🇬"},

		// Caribbean & Central America
		{CreatedAt: now, UpdatedAt: now, Name: "Jamaican Dollar", Country: "Jamaica", CurrencyCode: "JMD", Symbol: "J$", Emoji: "🇯🇲"},
		{CreatedAt: now, UpdatedAt: now, Name: "Costa Rican Colon", Country: "Costa Rica", CurrencyCode: "CRC", Symbol: "₡", Emoji: "🇨🇷"},
		{CreatedAt: now, UpdatedAt: now, Name: "Guatemalan Quetzal", Country: "Guatemala", CurrencyCode: "GTQ", Symbol: "Q", Emoji: "🇬🇹"},

		// Special Drawing Rights
		{CreatedAt: now, UpdatedAt: now, Name: "Special Drawing Rights", Country: "IMF", CurrencyCode: "XDR", Symbol: "SDR", Emoji: "🏦"},

		// Middle Eastern Currencies
		{CreatedAt: now, UpdatedAt: now, Name: "Kuwaiti Dinar", Country: "Kuwait", CurrencyCode: "KWD", Symbol: "د.ك", Emoji: "🇰🇼"},
		{CreatedAt: now, UpdatedAt: now, Name: "Qatari Riyal", Country: "Qatar", CurrencyCode: "QAR", Symbol: "ر.ق", Emoji: "🇶🇦"},
		{CreatedAt: now, UpdatedAt: now, Name: "Omani Rial", Country: "Oman", CurrencyCode: "OMR", Symbol: "ر.ع", Emoji: "🇴🇲"},
		{CreatedAt: now, UpdatedAt: now, Name: "Bahraini Dinar", Country: "Bahrain", CurrencyCode: "BHD", Symbol: "ب.د", Emoji: "🇧🇭"},
		{CreatedAt: now, UpdatedAt: now, Name: "Jordanian Dinar", Country: "Jordan", CurrencyCode: "JOD", Symbol: "د.ا", Emoji: "🇯🇴"},

		// Central Asian Currencies
		{CreatedAt: now, UpdatedAt: now, Name: "Kazakhstani Tenge", Country: "Kazakhstan", CurrencyCode: "KZT", Symbol: "₸", Emoji: "🇰🇿"},
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

func (m *ModelCore) CurrencyGetDefault(context context.Context) (*Currency, error) {
	// Default to PHP (Philippine Peso) as the system default currency
	return m.CurrencyFindByCode(context, "PHP")
}

func (m *ModelCore) CurrencySetDefaultForUserOrganization(context context.Context, userOrgID uuid.UUID, currencyID uuid.UUID) error {
	return m.UserOrganizationManager.UpdateFields(context, userOrgID, &UserOrganization{
		SettingsCurrencyDefaultValueID: &currencyID,
	})
}
