package types

import (
	"fmt"
	"time"

	"github.com/Lands-Horizon-Corp/numi18n/numi18n"
	"github.com/google/uuid"
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
		Timezone       string `gorm:"type:varchar(50)" json:"timezone"`        // Country timezone
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
		Timezone       string    `json:"timezone"`
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
		Timezone       string `json:"timezone,omitempty" validate:"omitempty,max=50"`
	}
)

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
