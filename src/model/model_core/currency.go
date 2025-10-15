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

		Name           string `gorm:"type:varchar(255);not null;unique" json:"name"`
		Country        string `gorm:"type:varchar(255);not null" json:"country"`
		CurrencyCode   string `gorm:"type:varchar(10);not null" json:"currency_code"`
		Symbol         string `gorm:"type:varchar(10)" json:"symbol"`
		Emoji          string `gorm:"type:varchar(10)" json:"emoji"`
		ISO3166Alpha2  string `gorm:"type:varchar(2)" json:"iso_3166_alpha2"`  // ISO 3166-1 alpha-2
		ISO3166Alpha3  string `gorm:"type:varchar(3)" json:"iso_3166_alpha3"`  // ISO 3166-1 alpha-3
		ISO3166Numeric string `gorm:"type:varchar(3)" json:"iso_3166_numeric"` // ISO 3166-1 numeric
		PhoneCode      string `gorm:"type:varchar(10)" json:"phone_code"`      // Country phone code
		Domain         string `gorm:"type:varchar(10)" json:"domain"`          // Country top-level domain
		Locale         string `gorm:"type:varchar(10)" json:"locale"`          // Country locale code
	}

	CurrencyResponse struct {
		ID             uuid.UUID `json:"id"`
		CreatedAt      string    `json:"created_at"`
		UpdatedAt      string    `json:"updated_at"`
		Name           string    `json:"name"`
		Country        string    `json:"country"`
		CurrencyCode   string    `json:"currency_code"`
		Symbol         string    `json:"symbol"`
		Emoji          string    `json:"emoji"`
		ISO3166Alpha2  string    `json:"iso_3166_alpha2"`
		ISO3166Alpha3  string    `json:"iso_3166_alpha3"`
		ISO3166Numeric string    `json:"iso_3166_numeric"`
		PhoneCode      string    `json:"phone_code"`
		Domain         string    `json:"domain"`
		Locale         string    `json:"locale"`
	}

	CurrencyRequest struct {
		Name           string `json:"name" validate:"required,min=1,max=255"`
		Country        string `json:"country" validate:"required,min=1,max=255"`
		CurrencyCode   string `json:"currency_code" validate:"required,min=2,max=10"`
		Symbol         string `json:"symbol,omitempty"`
		Emoji          string `json:"emoji,omitempty"`
		ISO3166Alpha2  string `json:"iso_3166_alpha2,omitempty" validate:"omitempty,len=2"`
		ISO3166Alpha3  string `json:"iso_3166_alpha3,omitempty" validate:"omitempty,len=3"`
		ISO3166Numeric string `json:"iso_3166_numeric,omitempty" validate:"omitempty,len=3"`
		PhoneCode      string `json:"phone_code,omitempty" validate:"omitempty,max=10"`
		Domain         string `json:"domain,omitempty" validate:"omitempty,max=10"`
		Locale         string `json:"locale,omitempty" validate:"omitempty,max=10"`
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
				ID:             data.ID,
				CreatedAt:      data.CreatedAt.Format(time.RFC3339),
				UpdatedAt:      data.UpdatedAt.Format(time.RFC3339),
				Name:           data.Name,
				Country:        data.Country,
				CurrencyCode:   data.CurrencyCode,
				Symbol:         data.Symbol,
				Emoji:          data.Emoji,
				ISO3166Alpha2:  data.ISO3166Alpha2,
				ISO3166Alpha3:  data.ISO3166Alpha3,
				ISO3166Numeric: data.ISO3166Numeric,
				PhoneCode:      data.PhoneCode,
				Domain:         data.Domain,
				Locale:         data.Locale,
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
		{CreatedAt: now, UpdatedAt: now, Name: "US Dollar", Country: "United States", CurrencyCode: "USD", Symbol: "US$", Emoji: "🇺🇸", ISO3166Alpha2: "US", ISO3166Alpha3: "USA", ISO3166Numeric: "840", PhoneCode: "+1", Domain: ".us", Locale: "en_US"},
		{CreatedAt: now, UpdatedAt: now, Name: "Euro", Country: "European Union", CurrencyCode: "EUR", Symbol: "€", Emoji: "🇪🇺", ISO3166Alpha2: "EU", ISO3166Alpha3: "EUR", ISO3166Numeric: "978", PhoneCode: "", Domain: ".eu", Locale: "en_EU"},
		{CreatedAt: now, UpdatedAt: now, Name: "Japanese Yen", Country: "Japan", CurrencyCode: "JPY", Symbol: "¥", Emoji: "🇯🇵", ISO3166Alpha2: "JP", ISO3166Alpha3: "JPN", ISO3166Numeric: "392", PhoneCode: "+81", Domain: ".jp", Locale: "ja_JP"},
		{CreatedAt: now, UpdatedAt: now, Name: "British Pound Sterling", Country: "United Kingdom", CurrencyCode: "GBP", Symbol: "£", Emoji: "🇬🇧", ISO3166Alpha2: "GB", ISO3166Alpha3: "GBR", ISO3166Numeric: "826", PhoneCode: "+44", Domain: ".uk", Locale: "en_GB"},
		{CreatedAt: now, UpdatedAt: now, Name: "Australian Dollar", Country: "Australia", CurrencyCode: "AUD", Symbol: "AU$", Emoji: "🇦🇺", ISO3166Alpha2: "AU", ISO3166Alpha3: "AUS", ISO3166Numeric: "036", PhoneCode: "+61", Domain: ".au", Locale: "en_AU"},
		{CreatedAt: now, UpdatedAt: now, Name: "Canadian Dollar", Country: "Canada", CurrencyCode: "CAD", Symbol: "CA$", Emoji: "🇨🇦", ISO3166Alpha2: "CA", ISO3166Alpha3: "CAN", ISO3166Numeric: "124", PhoneCode: "+1", Domain: ".ca", Locale: "en_CA"},
		{CreatedAt: now, UpdatedAt: now, Name: "Swiss Franc", Country: "Switzerland", CurrencyCode: "CHF", Symbol: "Fr", Emoji: "🇨🇭", ISO3166Alpha2: "CH", ISO3166Alpha3: "CHE", ISO3166Numeric: "756", PhoneCode: "+41", Domain: ".ch", Locale: "de_CH"},
		{CreatedAt: now, UpdatedAt: now, Name: "Chinese Yuan", Country: "China", CurrencyCode: "CNY", Symbol: "CN¥", Emoji: "🇨🇳", ISO3166Alpha2: "CN", ISO3166Alpha3: "CHN", ISO3166Numeric: "156", PhoneCode: "+86", Domain: ".cn", Locale: "zh_CN"},
		{CreatedAt: now, UpdatedAt: now, Name: "Swedish Krona", Country: "Sweden", CurrencyCode: "SEK", Symbol: "kr", Emoji: "🇸🇪", ISO3166Alpha2: "SE", ISO3166Alpha3: "SWE", ISO3166Numeric: "752", PhoneCode: "+46", Domain: ".se", Locale: "sv_SE"},
		{CreatedAt: now, UpdatedAt: now, Name: "New Zealand Dollar", Country: "New Zealand", CurrencyCode: "NZD", Symbol: "NZ$", Emoji: "🇳🇿", ISO3166Alpha2: "NZ", ISO3166Alpha3: "NZL", ISO3166Numeric: "554", PhoneCode: "+64", Domain: ".nz", Locale: "en_NZ"},

		// Asian Currencies
		{CreatedAt: now, UpdatedAt: now, Name: "Philippine Peso", Country: "Philippines", CurrencyCode: "PHP", Symbol: "₱", Emoji: "🇵🇭", ISO3166Alpha2: "PH", ISO3166Alpha3: "PHL", ISO3166Numeric: "608", PhoneCode: "+63", Domain: ".ph", Locale: "en_PH"},
		{CreatedAt: now, UpdatedAt: now, Name: "Indian Rupee", Country: "India", CurrencyCode: "INR", Symbol: "₹", Emoji: "🇮🇳", ISO3166Alpha2: "IN", ISO3166Alpha3: "IND", ISO3166Numeric: "356", PhoneCode: "+91", Domain: ".in", Locale: "hi_IN"},
		{CreatedAt: now, UpdatedAt: now, Name: "South Korean Won", Country: "South Korea", CurrencyCode: "KRW", Symbol: "₩", Emoji: "🇰🇷", ISO3166Alpha2: "KR", ISO3166Alpha3: "KOR", ISO3166Numeric: "410", PhoneCode: "+82", Domain: ".kr", Locale: "ko_KR"},
		{CreatedAt: now, UpdatedAt: now, Name: "Thai Baht", Country: "Thailand", CurrencyCode: "THB", Symbol: "฿", Emoji: "🇹🇭", ISO3166Alpha2: "TH", ISO3166Alpha3: "THA", ISO3166Numeric: "764", PhoneCode: "+66", Domain: ".th", Locale: "th_TH"},
		{CreatedAt: now, UpdatedAt: now, Name: "Singapore Dollar", Country: "Singapore", CurrencyCode: "SGD", Symbol: "S$", Emoji: "🇸🇬", ISO3166Alpha2: "SG", ISO3166Alpha3: "SGP", ISO3166Numeric: "702", PhoneCode: "+65", Domain: ".sg", Locale: "en_SG"},
		{CreatedAt: now, UpdatedAt: now, Name: "Hong Kong Dollar", Country: "Hong Kong", CurrencyCode: "HKD", Symbol: "HK$", Emoji: "🇭🇰", ISO3166Alpha2: "HK", ISO3166Alpha3: "HKG", ISO3166Numeric: "344", PhoneCode: "+852", Domain: ".hk", Locale: "zh_HK"},
		{CreatedAt: now, UpdatedAt: now, Name: "Malaysian Ringgit", Country: "Malaysia", CurrencyCode: "MYR", Symbol: "RM", Emoji: "🇲🇾", ISO3166Alpha2: "MY", ISO3166Alpha3: "MYS", ISO3166Numeric: "458", PhoneCode: "+60", Domain: ".my", Locale: "ms_MY"},
		{CreatedAt: now, UpdatedAt: now, Name: "Indonesian Rupiah", Country: "Indonesia", CurrencyCode: "IDR", Symbol: "Rp", Emoji: "🇮🇩", ISO3166Alpha2: "ID", ISO3166Alpha3: "IDN", ISO3166Numeric: "360", PhoneCode: "+62", Domain: ".id", Locale: "id_ID"},
		{CreatedAt: now, UpdatedAt: now, Name: "Vietnamese Dong", Country: "Vietnam", CurrencyCode: "VND", Symbol: "₫", Emoji: "🇻🇳", ISO3166Alpha2: "VN", ISO3166Alpha3: "VNM", ISO3166Numeric: "704", PhoneCode: "+84", Domain: ".vn", Locale: "vi_VN"},
		{CreatedAt: now, UpdatedAt: now, Name: "Taiwan Dollar", Country: "Taiwan", CurrencyCode: "TWD", Symbol: "NT$", Emoji: "🇹🇼", ISO3166Alpha2: "TW", ISO3166Alpha3: "TWN", ISO3166Numeric: "158", PhoneCode: "+886", Domain: ".tw", Locale: "zh_TW"},
		{CreatedAt: now, UpdatedAt: now, Name: "Brunei Dollar", Country: "Brunei", CurrencyCode: "BND", Symbol: "B$", Emoji: "🇧🇳", ISO3166Alpha2: "BN", ISO3166Alpha3: "BRN", ISO3166Numeric: "096", PhoneCode: "+673", Domain: ".bn", Locale: "ms_BN"},

		// Middle Eastern & African Currencies
		{CreatedAt: now, UpdatedAt: now, Name: "Saudi Riyal", Country: "Saudi Arabia", CurrencyCode: "SAR", Symbol: "ر.س", Emoji: "🇸🇦", ISO3166Alpha2: "SA", ISO3166Alpha3: "SAU", ISO3166Numeric: "682", PhoneCode: "+966", Domain: ".sa", Locale: "ar_SA"},
		{CreatedAt: now, UpdatedAt: now, Name: "UAE Dirham", Country: "United Arab Emirates", CurrencyCode: "AED", Symbol: "د.إ", Emoji: "🇦🇪", ISO3166Alpha2: "AE", ISO3166Alpha3: "ARE", ISO3166Numeric: "784", PhoneCode: "+971", Domain: ".ae", Locale: "ar_AE"},
		{CreatedAt: now, UpdatedAt: now, Name: "Israeli New Shekel", Country: "Israel", CurrencyCode: "ILS", Symbol: "₪", Emoji: "🇮🇱", ISO3166Alpha2: "IL", ISO3166Alpha3: "ISR", ISO3166Numeric: "376", PhoneCode: "+972", Domain: ".il", Locale: "he_IL"},
		{CreatedAt: now, UpdatedAt: now, Name: "South African Rand", Country: "South Africa", CurrencyCode: "ZAR", Symbol: "R", Emoji: "🇿🇦", ISO3166Alpha2: "ZA", ISO3166Alpha3: "ZAF", ISO3166Numeric: "710", PhoneCode: "+27", Domain: ".za", Locale: "af_ZA"},
		{CreatedAt: now, UpdatedAt: now, Name: "Egyptian Pound", Country: "Egypt", CurrencyCode: "EGP", Symbol: "ج.م", Emoji: "🇪🇬", ISO3166Alpha2: "EG", ISO3166Alpha3: "EGY", ISO3166Numeric: "818", PhoneCode: "+20", Domain: ".eg", Locale: "ar_EG"},
		{CreatedAt: now, UpdatedAt: now, Name: "Turkish Lira", Country: "Turkey", CurrencyCode: "TRY", Symbol: "₺", Emoji: "🇹🇷", ISO3166Alpha2: "TR", ISO3166Alpha3: "TUR", ISO3166Numeric: "792", PhoneCode: "+90", Domain: ".tr", Locale: "tr_TR"},
		{CreatedAt: now, UpdatedAt: now, Name: "West African CFA Franc", Country: "West African States", CurrencyCode: "XOF", Symbol: "CFA", Emoji: "🌍", ISO3166Alpha2: "", ISO3166Alpha3: "", ISO3166Numeric: "952", PhoneCode: "", Domain: "", Locale: "fr_FR"},
		{CreatedAt: now, UpdatedAt: now, Name: "Central African CFA Franc", Country: "Central African States", CurrencyCode: "XAF", Symbol: "CFA", Emoji: "🌍", ISO3166Alpha2: "", ISO3166Alpha3: "", ISO3166Numeric: "950", PhoneCode: "", Domain: "", Locale: "fr_FR"},
		{CreatedAt: now, UpdatedAt: now, Name: "Mauritian Rupee", Country: "Mauritius", CurrencyCode: "MUR", Symbol: "₨", Emoji: "🇲🇺", ISO3166Alpha2: "MU", ISO3166Alpha3: "MUS", ISO3166Numeric: "480", PhoneCode: "+230", Domain: ".mu", Locale: "en_MU"},
		{CreatedAt: now, UpdatedAt: now, Name: "Maldivian Rufiyaa", Country: "Maldives", CurrencyCode: "MVR", Symbol: "Rf", Emoji: "🇲🇻", ISO3166Alpha2: "MV", ISO3166Alpha3: "MDV", ISO3166Numeric: "462", PhoneCode: "+960", Domain: ".mv", Locale: "dv_MV"},

		// European Currencies (Non-Euro)
		{CreatedAt: now, UpdatedAt: now, Name: "Norwegian Krone", Country: "Norway", CurrencyCode: "NOK", Symbol: "kr", Emoji: "🇳🇴", ISO3166Alpha2: "NO", ISO3166Alpha3: "NOR", ISO3166Numeric: "578", PhoneCode: "+47", Domain: ".no", Locale: "nb_NO"},
		{CreatedAt: now, UpdatedAt: now, Name: "Danish Krone", Country: "Denmark", CurrencyCode: "DKK", Symbol: "kr", Emoji: "🇩🇰", ISO3166Alpha2: "DK", ISO3166Alpha3: "DNK", ISO3166Numeric: "208", PhoneCode: "+45", Domain: ".dk", Locale: "da_DK"},
		{CreatedAt: now, UpdatedAt: now, Name: "Polish Zloty", Country: "Poland", CurrencyCode: "PLN", Symbol: "zł", Emoji: "🇵🇱", ISO3166Alpha2: "PL", ISO3166Alpha3: "POL", ISO3166Numeric: "616", PhoneCode: "+48", Domain: ".pl", Locale: "pl_PL"},
		{CreatedAt: now, UpdatedAt: now, Name: "Czech Koruna", Country: "Czech Republic", CurrencyCode: "CZK", Symbol: "Kč", Emoji: "🇨🇿", ISO3166Alpha2: "CZ", ISO3166Alpha3: "CZE", ISO3166Numeric: "203", PhoneCode: "+420", Domain: ".cz", Locale: "cs_CZ"},
		{CreatedAt: now, UpdatedAt: now, Name: "Hungarian Forint", Country: "Hungary", CurrencyCode: "HUF", Symbol: "Ft", Emoji: "🇭🇺", ISO3166Alpha2: "HU", ISO3166Alpha3: "HUN", ISO3166Numeric: "348", PhoneCode: "+36", Domain: ".hu", Locale: "hu_HU"},
		{CreatedAt: now, UpdatedAt: now, Name: "Russian Ruble", Country: "Russia", CurrencyCode: "RUB", Symbol: "₽", Emoji: "🇷🇺", ISO3166Alpha2: "RU", ISO3166Alpha3: "RUS", ISO3166Numeric: "643", PhoneCode: "+7", Domain: ".ru", Locale: "ru_RU"},
		{CreatedAt: now, UpdatedAt: now, Name: "Euro (Croatia)", Country: "Croatia", CurrencyCode: "EUR", Symbol: "€", Emoji: "🇭🇷", ISO3166Alpha2: "HR", ISO3166Alpha3: "HRV", ISO3166Numeric: "191", PhoneCode: "+385", Domain: ".hr", Locale: "hr_HR"},

		// Latin American Currencies
		{CreatedAt: now, UpdatedAt: now, Name: "Brazilian Real", Country: "Brazil", CurrencyCode: "BRL", Symbol: "R$", Emoji: "🇧🇷", ISO3166Alpha2: "BR", ISO3166Alpha3: "BRA", ISO3166Numeric: "076", PhoneCode: "+55", Domain: ".br", Locale: "pt_BR"},
		{CreatedAt: now, UpdatedAt: now, Name: "Mexican Peso", Country: "Mexico", CurrencyCode: "MXN", Symbol: "MX$", Emoji: "🇲🇽", ISO3166Alpha2: "MX", ISO3166Alpha3: "MEX", ISO3166Numeric: "484", PhoneCode: "+52", Domain: ".mx", Locale: "es_MX"},
		{CreatedAt: now, UpdatedAt: now, Name: "Argentine Peso", Country: "Argentina", CurrencyCode: "ARS", Symbol: "AR$", Emoji: "🇦🇷", ISO3166Alpha2: "AR", ISO3166Alpha3: "ARG", ISO3166Numeric: "032", PhoneCode: "+54", Domain: ".ar", Locale: "es_AR"},
		{CreatedAt: now, UpdatedAt: now, Name: "Chilean Peso", Country: "Chile", CurrencyCode: "CLP", Symbol: "CL$", Emoji: "🇨🇱", ISO3166Alpha2: "CL", ISO3166Alpha3: "CHL", ISO3166Numeric: "152", PhoneCode: "+56", Domain: ".cl", Locale: "es_CL"},
		{CreatedAt: now, UpdatedAt: now, Name: "Colombian Peso", Country: "Colombia", CurrencyCode: "COP", Symbol: "CO$", Emoji: "🇨🇴", ISO3166Alpha2: "CO", ISO3166Alpha3: "COL", ISO3166Numeric: "170", PhoneCode: "+57", Domain: ".co", Locale: "es_CO"},
		{CreatedAt: now, UpdatedAt: now, Name: "Peruvian Sol", Country: "Peru", CurrencyCode: "PEN", Symbol: "S/", Emoji: "🇵🇪", ISO3166Alpha2: "PE", ISO3166Alpha3: "PER", ISO3166Numeric: "604", PhoneCode: "+51", Domain: ".pe", Locale: "es_PE"},
		{CreatedAt: now, UpdatedAt: now, Name: "Uruguayan Peso", Country: "Uruguay", CurrencyCode: "UYU", Symbol: "$U", Emoji: "🇺🇾", ISO3166Alpha2: "UY", ISO3166Alpha3: "URY", ISO3166Numeric: "858", PhoneCode: "+598", Domain: ".uy", Locale: "es_UY"},
		{CreatedAt: now, UpdatedAt: now, Name: "Dominican Peso", Country: "Dominican Republic", CurrencyCode: "DOP", Symbol: "RD$", Emoji: "🇩🇴", ISO3166Alpha2: "DO", ISO3166Alpha3: "DOM", ISO3166Numeric: "214", PhoneCode: "+1", Domain: ".do", Locale: "es_DO"},
		{CreatedAt: now, UpdatedAt: now, Name: "Paraguayan Guarani", Country: "Paraguay", CurrencyCode: "PYG", Symbol: "₲", Emoji: "🇵🇾", ISO3166Alpha2: "PY", ISO3166Alpha3: "PRY", ISO3166Numeric: "600", PhoneCode: "+595", Domain: ".py", Locale: "es_PY"},
		{CreatedAt: now, UpdatedAt: now, Name: "Bolivian Boliviano", Country: "Bolivia", CurrencyCode: "BOB", Symbol: "Bs", Emoji: "🇧🇴", ISO3166Alpha2: "BO", ISO3166Alpha3: "BOL", ISO3166Numeric: "068", PhoneCode: "+591", Domain: ".bo", Locale: "es_BO"},
		{CreatedAt: now, UpdatedAt: now, Name: "Venezuelan Bolívar", Country: "Venezuela", CurrencyCode: "VES", Symbol: "Bs.S", Emoji: "🇻🇪", ISO3166Alpha2: "VE", ISO3166Alpha3: "VEN", ISO3166Numeric: "928", PhoneCode: "+58", Domain: ".ve", Locale: "es_VE"},

		// Additional Asian Currencies
		{CreatedAt: now, UpdatedAt: now, Name: "Pakistani Rupee", Country: "Pakistan", CurrencyCode: "PKR", Symbol: "₨", Emoji: "🇵🇰", ISO3166Alpha2: "PK", ISO3166Alpha3: "PAK", ISO3166Numeric: "586", PhoneCode: "+92", Domain: ".pk", Locale: "ur_PK"},
		{CreatedAt: now, UpdatedAt: now, Name: "Bangladeshi Taka", Country: "Bangladesh", CurrencyCode: "BDT", Symbol: "৳", Emoji: "🇧🇩", ISO3166Alpha2: "BD", ISO3166Alpha3: "BGD", ISO3166Numeric: "050", PhoneCode: "+880", Domain: ".bd", Locale: "bn_BD"},
		{CreatedAt: now, UpdatedAt: now, Name: "Sri Lankan Rupee", Country: "Sri Lanka", CurrencyCode: "LKR", Symbol: "Rs", Emoji: "🇱🇰", ISO3166Alpha2: "LK", ISO3166Alpha3: "LKA", ISO3166Numeric: "144", PhoneCode: "+94", Domain: ".lk", Locale: "si_LK"},
		{CreatedAt: now, UpdatedAt: now, Name: "Nepalese Rupee", Country: "Nepal", CurrencyCode: "NPR", Symbol: "Rs", Emoji: "🇳🇵", ISO3166Alpha2: "NP", ISO3166Alpha3: "NPL", ISO3166Numeric: "524", PhoneCode: "+977", Domain: ".np", Locale: "ne_NP"},
		{CreatedAt: now, UpdatedAt: now, Name: "Myanmar Kyat", Country: "Myanmar", CurrencyCode: "MMK", Symbol: "K", Emoji: "🇲🇲", ISO3166Alpha2: "MM", ISO3166Alpha3: "MMR", ISO3166Numeric: "104", PhoneCode: "+95", Domain: ".mm", Locale: "my_MM"},
		{CreatedAt: now, UpdatedAt: now, Name: "Cambodian Riel", Country: "Cambodia", CurrencyCode: "KHR", Symbol: "៛", Emoji: "🇰🇭", ISO3166Alpha2: "KH", ISO3166Alpha3: "KHM", ISO3166Numeric: "116", PhoneCode: "+855", Domain: ".kh", Locale: "km_KH"},
		{CreatedAt: now, UpdatedAt: now, Name: "Laotian Kip", Country: "Laos", CurrencyCode: "LAK", Symbol: "₭", Emoji: "🇱🇦", ISO3166Alpha2: "LA", ISO3166Alpha3: "LAO", ISO3166Numeric: "418", PhoneCode: "+856", Domain: ".la", Locale: "lo_LA"},

		// Additional African Currencies
		{CreatedAt: now, UpdatedAt: now, Name: "Nigerian Naira", Country: "Nigeria", CurrencyCode: "NGN", Symbol: "₦", Emoji: "🇳🇬", ISO3166Alpha2: "NG", ISO3166Alpha3: "NGA", ISO3166Numeric: "566", PhoneCode: "+234", Domain: ".ng", Locale: "en_NG"},
		{CreatedAt: now, UpdatedAt: now, Name: "Kenyan Shilling", Country: "Kenya", CurrencyCode: "KES", Symbol: "KSh", Emoji: "🇰🇪", ISO3166Alpha2: "KE", ISO3166Alpha3: "KEN", ISO3166Numeric: "404", PhoneCode: "+254", Domain: ".ke", Locale: "sw_KE"},
		{CreatedAt: now, UpdatedAt: now, Name: "Ghanaian Cedi", Country: "Ghana", CurrencyCode: "GHS", Symbol: "₵", Emoji: "🇬🇭", ISO3166Alpha2: "GH", ISO3166Alpha3: "GHA", ISO3166Numeric: "288", PhoneCode: "+233", Domain: ".gh", Locale: "en_GH"},
		{CreatedAt: now, UpdatedAt: now, Name: "Moroccan Dirham", Country: "Morocco", CurrencyCode: "MAD", Symbol: "د.م.", Emoji: "🇲🇦", ISO3166Alpha2: "MA", ISO3166Alpha3: "MAR", ISO3166Numeric: "504", PhoneCode: "+212", Domain: ".ma", Locale: "ar_MA"},
		{CreatedAt: now, UpdatedAt: now, Name: "Tunisian Dinar", Country: "Tunisia", CurrencyCode: "TND", Symbol: "د.ت", Emoji: "🇹🇳", ISO3166Alpha2: "TN", ISO3166Alpha3: "TUN", ISO3166Numeric: "788", PhoneCode: "+216", Domain: ".tn", Locale: "ar_TN"},
		{CreatedAt: now, UpdatedAt: now, Name: "Ethiopian Birr", Country: "Ethiopia", CurrencyCode: "ETB", Symbol: "Br", Emoji: "🇪🇹", ISO3166Alpha2: "ET", ISO3166Alpha3: "ETH", ISO3166Numeric: "230", PhoneCode: "+251", Domain: ".et", Locale: "am_ET"},
		{CreatedAt: now, UpdatedAt: now, Name: "Algerian Dinar", Country: "Algeria", CurrencyCode: "DZD", Symbol: "د.ج", Emoji: "🇩🇿", ISO3166Alpha2: "DZ", ISO3166Alpha3: "DZA", ISO3166Numeric: "012", PhoneCode: "+213", Domain: ".dz", Locale: "ar_DZ"},

		// Additional European Currencies
		{CreatedAt: now, UpdatedAt: now, Name: "Ukrainian Hryvnia", Country: "Ukraine", CurrencyCode: "UAH", Symbol: "₴", Emoji: "🇺🇦", ISO3166Alpha2: "UA", ISO3166Alpha3: "UKR", ISO3166Numeric: "804", PhoneCode: "+380", Domain: ".ua", Locale: "uk_UA"},
		{CreatedAt: now, UpdatedAt: now, Name: "Romanian Leu", Country: "Romania", CurrencyCode: "RON", Symbol: "lei", Emoji: "🇷🇴", ISO3166Alpha2: "RO", ISO3166Alpha3: "ROU", ISO3166Numeric: "642", PhoneCode: "+40", Domain: ".ro", Locale: "ro_RO"},
		{CreatedAt: now, UpdatedAt: now, Name: "Bulgarian Lev", Country: "Bulgaria", CurrencyCode: "BGN", Symbol: "лв", Emoji: "🇧🇬", ISO3166Alpha2: "BG", ISO3166Alpha3: "BGR", ISO3166Numeric: "100", PhoneCode: "+359", Domain: ".bg", Locale: "bg_BG"},
		{CreatedAt: now, UpdatedAt: now, Name: "Serbian Dinar", Country: "Serbia", CurrencyCode: "RSD", Symbol: "дин", Emoji: "🇷🇸", ISO3166Alpha2: "RS", ISO3166Alpha3: "SRB", ISO3166Numeric: "941", PhoneCode: "+381", Domain: ".rs", Locale: "sr_RS"},
		{CreatedAt: now, UpdatedAt: now, Name: "Icelandic Krona", Country: "Iceland", CurrencyCode: "ISK", Symbol: "kr", Emoji: "🇮🇸", ISO3166Alpha2: "IS", ISO3166Alpha3: "ISL", ISO3166Numeric: "352", PhoneCode: "+354", Domain: ".is", Locale: "is_IS"},
		{CreatedAt: now, UpdatedAt: now, Name: "Belarusian Ruble", Country: "Belarus", CurrencyCode: "BYN", Symbol: "Br", Emoji: "🇧🇾", ISO3166Alpha2: "BY", ISO3166Alpha3: "BLR", ISO3166Numeric: "933", PhoneCode: "+375", Domain: ".by", Locale: "be_BY"},

		// Oceania & Others
		{CreatedAt: now, UpdatedAt: now, Name: "Fijian Dollar", Country: "Fiji", CurrencyCode: "FJD", Symbol: "FJ$", Emoji: "🇫🇯", ISO3166Alpha2: "FJ", ISO3166Alpha3: "FJI", ISO3166Numeric: "242", PhoneCode: "+679", Domain: ".fj", Locale: "en_FJ"},
		{CreatedAt: now, UpdatedAt: now, Name: "Papua New Guinea Kina", Country: "Papua New Guinea", CurrencyCode: "PGK", Symbol: "K", Emoji: "🇵🇬", ISO3166Alpha2: "PG", ISO3166Alpha3: "PNG", ISO3166Numeric: "598", PhoneCode: "+675", Domain: ".pg", Locale: "en_PG"},

		// Caribbean & Central America
		{CreatedAt: now, UpdatedAt: now, Name: "Jamaican Dollar", Country: "Jamaica", CurrencyCode: "JMD", Symbol: "J$", Emoji: "🇯🇲", ISO3166Alpha2: "JM", ISO3166Alpha3: "JAM", ISO3166Numeric: "388", PhoneCode: "+1", Domain: ".jm", Locale: "en_JM"},
		{CreatedAt: now, UpdatedAt: now, Name: "Costa Rican Colon", Country: "Costa Rica", CurrencyCode: "CRC", Symbol: "₡", Emoji: "🇨🇷", ISO3166Alpha2: "CR", ISO3166Alpha3: "CRI", ISO3166Numeric: "188", PhoneCode: "+506", Domain: ".cr", Locale: "es_CR"},
		{CreatedAt: now, UpdatedAt: now, Name: "Guatemalan Quetzal", Country: "Guatemala", CurrencyCode: "GTQ", Symbol: "Q", Emoji: "🇬🇹", ISO3166Alpha2: "GT", ISO3166Alpha3: "GTM", ISO3166Numeric: "320", PhoneCode: "+502", Domain: ".gt", Locale: "es_GT"},

		// Special Drawing Rights
		{CreatedAt: now, UpdatedAt: now, Name: "Special Drawing Rights", Country: "IMF", CurrencyCode: "XDR", Symbol: "SDR", Emoji: "🏦", ISO3166Alpha2: "", ISO3166Alpha3: "", ISO3166Numeric: "960", PhoneCode: "", Domain: "", Locale: "en_US"},

		// Middle Eastern Currencies
		{CreatedAt: now, UpdatedAt: now, Name: "Kuwaiti Dinar", Country: "Kuwait", CurrencyCode: "KWD", Symbol: "د.ك", Emoji: "🇰🇼", ISO3166Alpha2: "KW", ISO3166Alpha3: "KWT", ISO3166Numeric: "414", PhoneCode: "+965", Domain: ".kw", Locale: "ar_KW"},
		{CreatedAt: now, UpdatedAt: now, Name: "Qatari Riyal", Country: "Qatar", CurrencyCode: "QAR", Symbol: "ر.ق", Emoji: "🇶🇦", ISO3166Alpha2: "QA", ISO3166Alpha3: "QAT", ISO3166Numeric: "634", PhoneCode: "+974", Domain: ".qa", Locale: "ar_QA"},
		{CreatedAt: now, UpdatedAt: now, Name: "Omani Rial", Country: "Oman", CurrencyCode: "OMR", Symbol: "ر.ع", Emoji: "🇴🇲", ISO3166Alpha2: "OM", ISO3166Alpha3: "OMN", ISO3166Numeric: "512", PhoneCode: "+968", Domain: ".om", Locale: "ar_OM"},
		{CreatedAt: now, UpdatedAt: now, Name: "Bahraini Dinar", Country: "Bahrain", CurrencyCode: "BHD", Symbol: "ب.د", Emoji: "🇧🇭", ISO3166Alpha2: "BH", ISO3166Alpha3: "BHR", ISO3166Numeric: "048", PhoneCode: "+973", Domain: ".bh", Locale: "ar_BH"},
		{CreatedAt: now, UpdatedAt: now, Name: "Jordanian Dinar", Country: "Jordan", CurrencyCode: "JOD", Symbol: "د.ا", Emoji: "🇯🇴", ISO3166Alpha2: "JO", ISO3166Alpha3: "JOR", ISO3166Numeric: "400", PhoneCode: "+962", Domain: ".jo", Locale: "ar_JO"},

		// Central Asian Currencies
		{CreatedAt: now, UpdatedAt: now, Name: "Kazakhstani Tenge", Country: "Kazakhstan", CurrencyCode: "KZT", Symbol: "₸", Emoji: "🇰🇿", ISO3166Alpha2: "KZ", ISO3166Alpha3: "KAZ", ISO3166Numeric: "398", PhoneCode: "+7", Domain: ".kz", Locale: "kk_KZ"},
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

func (m *ModelCore) CurrencyFindByAlpha2(context context.Context, iso3166Alpha2 string) (*Currency, error) {
	currencies, err := m.CurrencyManager.FindOne(context, &Currency{ISO3166Alpha2: iso3166Alpha2})
	if err != nil {
		return nil, err
	}
	return currencies, nil
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
