package core

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/horizon"
	"github.com/Lands-Horizon-Corp/e-coop-server/pkg/registry"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/types"
	"github.com/Lands-Horizon-Corp/numi18n/numi18n"

	"github.com/rotisserie/eris"
)

func CurrencyManager(service *horizon.HorizonService) *registry.Registry[types.Currency, types.CurrencyResponse, types.CurrencyRequest] {
	return registry.NewRegistry(registry.RegistryParams[types.Currency, types.CurrencyResponse, types.CurrencyRequest]{
		Database: service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return service.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *types.Currency) *types.CurrencyResponse {
			if data == nil {
				return nil
			}
			return &types.CurrencyResponse{
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
		Created: func(data *types.Currency) registry.Topics {
			return []string{
				"currency.create",
				fmt.Sprintf("currency.create.%s", data.ID),
				fmt.Sprintf("currency.create.code.%s", data.CurrencyCode),
			}
		},
		Updated: func(data *types.Currency) registry.Topics {
			return []string{
				"currency.update",
				fmt.Sprintf("currency.update.%s", data.ID),
				fmt.Sprintf("currency.update.code.%s", data.CurrencyCode),
			}
		},
		Deleted: func(data *types.Currency) registry.Topics {
			return []string{
				"currency.delete",
				fmt.Sprintf("currency.delete.%s", data.ID),
				fmt.Sprintf("currency.delete.code.%s", data.CurrencyCode),
			}
		},
	})
}

func currencySeed(context context.Context, service *horizon.HorizonService) error {
	now := time.Now().UTC()
	availableLocales := numi18n.PerCountryLocales()
	for _, locales := range availableLocales {
		if len(locales.NumI18Identifier.ISO3166Alpha2) != 2 {
			continue
		}
		currency := &types.Currency{
			CreatedAt:      now,
			UpdatedAt:      now,
			Name:           locales.NumI18Identifier.Locale + " " + locales.Currency.Name,
			Country:        locales.NumI18Identifier.CountryName,
			CurrencyCode:   locales.NumI18Identifier.Currency,
			Symbol:         locales.Currency.Symbol,
			Emoji:          locales.NumI18Identifier.Emoji,
			ISO3166Alpha2:  locales.NumI18Identifier.ISO3166Alpha2,
			ISO3166Alpha3:  locales.NumI18Identifier.ISO3166Alpha3,
			ISO3166Numeric: locales.NumI18Identifier.ISO3166Numeric,
			PhoneCode:      locales.NumI18Identifier.PhoneCode,
			Domain:         locales.NumI18Identifier.Domain,
			Locale:         locales.NumI18Identifier.Locale,
			Timezone:       strings.Join(locales.NumI18Identifier.Timezone, ","),
		}
		if err := CurrencyManager(service).Create(context, currency); err != nil {
			return eris.Wrapf(err, "failed to seed currency %s (%s)", currency.Name, currency.ISO3166Alpha3)
		}
	}
	return nil
}

func CurrencyFindByAlpha2(context context.Context, service *horizon.HorizonService, iso3166Alpha2 string) (*types.Currency, error) {
	currencies, err := CurrencyManager(service).FindOne(context, &types.Currency{ISO3166Alpha2: iso3166Alpha2})
	if err != nil {
		return nil, err
	}
	return currencies, nil
}

func CurrencyFindByCode(context context.Context, service *horizon.HorizonService, currencyCode string) (*types.Currency, error) {
	currency, err := CurrencyManager(service).FindOne(context, &types.Currency{CurrencyCode: currencyCode})
	if err != nil {
		return nil, err
	}
	return currency, nil
}
