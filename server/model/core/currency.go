package core

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/services/registry"
	"github.com/Lands-Horizon-Corp/numi18n/numi18n"

	"github.com/google/uuid"
	"github.com/rotisserie/eris"
	"gorm.io/gorm"
)

type (
	// Currency represents the Currency model.
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
		Timezone       string `gorm:"type:varchar(50)" json:"timezone"`        // Country timezone
	}

	// CurrencyResponse represents the response structure for currency data

	// CurrencyResponse represents the response structure for Currency.
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
		Timezone       string    `json:"timezone"`
	}

	// CurrencyRequest represents the request structure for creating/updating currency

	// CurrencyRequest represents the request structure for Currency.
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
		Timezone       string `json:"timezone,omitempty" validate:"omitempty,max=50"`
	}
)

func (m *Core) currency() {
	m.Migration = append(m.Migration, &Currency{})
	m.CurrencyManager = *registry.NewRegistry(registry.RegistryParams[Currency, CurrencyResponse, CurrencyRequest]{
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
				Timezone:       data.Timezone,
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

func (c *Currency) ToFormat(value float64) string {
	if c == nil {
		return fmt.Sprintf("%.2f", value)
	}
	options := &numi18n.NumI18NOptions{
		Currency:       c.CurrencyCode,
		ISO3166Alpha2:  c.ISO3166Alpha2,
		ISO3166Alpha3:  c.ISO3166Alpha3,
		ISO3166Numeric: c.ISO3166Numeric,
		Locale:         c.Locale,
		WordDetails: &numi18n.WordDetails{
			Currency: true,
			Decimal:  true,
		},
	}
	return options.ToFormat(value)
}

func (c *Currency) ToWords(amount float64) string {
	if c == nil {
		return fmt.Sprintf("%.2f", amount)
	}
	options := &numi18n.NumI18NOptions{
		Currency:       c.CurrencyCode,
		ISO3166Alpha2:  c.ISO3166Alpha2,
		ISO3166Alpha3:  c.ISO3166Alpha3,
		ISO3166Numeric: c.ISO3166Numeric,
		Locale:         c.Locale,
		WordDetails: &numi18n.WordDetails{
			Currency:   true,
			Decimal:    true,
			Capitalize: true,
		},
	}
	return options.ToWords(amount)
}

func (m *Core) currencySeed(context context.Context) error {
	now := time.Now().UTC()
	availableLocales := numi18n.Locales()
	for _, locale := range availableLocales {
		currency := &Currency{
			CreatedAt:      now,
			UpdatedAt:      now,
			Name:           locale.Currency.Name,
			Country:        locale.NumI18Identifier.CountryName,
			CurrencyCode:   locale.NumI18Identifier.Currency,
			Symbol:         locale.Currency.Symbol,
			Emoji:          locale.NumI18Identifier.Emoji,
			ISO3166Alpha2:  locale.NumI18Identifier.ISO3166Alpha2,
			ISO3166Alpha3:  locale.NumI18Identifier.ISO3166Alpha3,
			ISO3166Numeric: locale.NumI18Identifier.ISO3166Numeric,
			PhoneCode:      "", // Not available in numi18n
			Domain:         "", // Not available in numi18n
			Locale:         locale.NumI18Identifier.Locale,
			Timezone:       strings.Join(locale.NumI18Identifier.Timezone, ","),
		}
		if err := m.CurrencyManager.Create(context, currency); err != nil {
			return eris.Wrapf(err, "failed to seed currency %s (%s)", currency.Name, currency.CurrencyCode)
		}
	}
	return nil
}

// CurrencyFindByAlpha2 retrieves a currency by its ISO 3166-1 alpha-2 code
func (m *Core) CurrencyFindByAlpha2(context context.Context, iso3166Alpha2 string) (*Currency, error) {
	currencies, err := m.CurrencyManager.FindOne(context, &Currency{ISO3166Alpha2: iso3166Alpha2})
	if err != nil {
		return nil, err
	}
	return currencies, nil
}

// CurrencyFindByCode retrieves a currency by its currency code
func (m *Core) CurrencyFindByCode(context context.Context, currencyCode string) (*Currency, error) {
	currency, err := m.CurrencyManager.FindOne(context, &Currency{CurrencyCode: currencyCode})
	if err != nil {
		return nil, err
	}
	return currency, nil
}
