package services

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

// The euro and the US dollar, the collection's usual currencies.
const (
	CurrencyUSD = "USD"
	CurrencyEUR = "EUR"
)

// CollectionCurrencies are the currencies a collection can be kept in: the
// euro and the currencies the ECB publishes a daily euro reference rate for.
var CollectionCurrencies = []string{
	"EUR", "USD", "GBP", "CHF", "JPY", "CAD", "AUD", "NZD", "BRL", "CNY", "HKD", "SGD", "KRW", "INR", "IDR",
	"MYR", "PHP", "THB", "MXN", "ZAR", "ILS", "TRY", "SEK", "NOK", "DKK", "ISK", "PLN", "CZK", "HUF", "RON",
}

var (
	// ErrUnsupportedCollectionCurrency is returned for a currency the ECB gives no rate for.
	ErrUnsupportedCollectionCurrency = errors.New("collection currency must be the euro or a currency of the ECB reference rates")
	// ErrExchangeRateUnavailable means no ECB rate could be read, so nothing was converted.
	ErrExchangeRateUnavailable = errors.New("exchange rate unavailable")
)

// NormalizeCollectionCurrency returns the code of a collection currency
// ("eur" → "EUR"), and false for any other code.
func NormalizeCollectionCurrency(code string) (string, bool) {
	code = strings.ToUpper(strings.TrimSpace(code))
	for _, currency := range CollectionCurrencies {
		if currency == code {
			return currency, true
		}
	}
	return "", false
}

// CollectionCurrency is the currency coin prices and values are kept in and
// every page shows; USD unless an admin switched the collection to another.
func (s *SettingsService) CollectionCurrency() string {
	if s == nil {
		return CurrencyUSD
	}
	if currency, ok := NormalizeCollectionCurrency(s.GetSetting(SettingCollectionCurrency)); ok {
		return currency
	}
	return CurrencyUSD
}

// CurrencyInstruction tells an AI model which currency to give prices in. It
// is empty for US dollars, the currency the built-in prompts already ask for.
func CurrencyInstruction(currency string) string {
	if currency == "" || currency == CurrencyUSD {
		return ""
	}
	return fmt.Sprintf("Give estimatedValue and every comparable price in %s, not USD; convert when a source is priced in another currency.", currency)
}

// CollectionCurrencyChange reports a currency switch.
type CollectionCurrencyChange struct {
	Currency string `json:"currency"`
	Previous string `json:"previous"`
	// Factor is what every amount was multiplied by (how much of the new
	// currency one unit of the previous one buys), at the ECB rates of
	// RateDate; both are empty when nothing was converted.
	Factor   float64 `json:"factor,omitempty"`
	RateDate string  `json:"rateDate,omitempty"`
	// CoinsConverted counts coins with a price or value that was converted.
	CoinsConverted int64 `json:"coinsConverted"`
}

type collectionCurrencyRates interface {
	Rates(ctx context.Context) (ExchangeRates, error)
}

// collectionAmountConverter converts the stored amounts and records the new
// currency in one transaction (repository.CollectionCurrencyRepository).
type collectionAmountConverter interface {
	ConvertAmounts(factor float64, settingKey, value string) (int64, error)
}

// CollectionCurrencyService switches the currency the collection is kept in,
// converting every stored price and value with the ECB reference rate in one
// transaction so no amount is ever read in the wrong currency.
type CollectionCurrencyService struct {
	amounts  collectionAmountConverter
	settings *SettingsService
	rates    collectionCurrencyRates
}

// NewCollectionCurrencyService creates the service.
func NewCollectionCurrencyService(amounts collectionAmountConverter, settings *SettingsService, rates collectionCurrencyRates) *CollectionCurrencyService {
	return &CollectionCurrencyService{amounts: amounts, settings: settings, rates: rates}
}

// Change switches the collection to currency, converting stored amounts.
// Switching to the current currency changes nothing.
func (s *CollectionCurrencyService) Change(ctx context.Context, currency string) (CollectionCurrencyChange, error) {
	target, ok := NormalizeCollectionCurrency(currency)
	if !ok {
		return CollectionCurrencyChange{}, ErrUnsupportedCollectionCurrency
	}
	previous := s.settings.CollectionCurrency()
	change := CollectionCurrencyChange{Currency: target, Previous: previous}
	if target == previous {
		return change, nil
	}

	rates, err := s.rates.Rates(ctx)
	if err != nil {
		return change, fmt.Errorf("%w: %v", ErrExchangeRateUnavailable, err)
	}
	// Both rates are euros' worth: the ECB quotes every currency against the euro.
	from, fromOK := rates.Rate(previous)
	to, toOK := rates.Rate(target)
	if !fromOK || !toOK {
		return change, ErrExchangeRateUnavailable
	}
	factor := to / from
	coins, err := s.amounts.ConvertAmounts(factor, SettingCollectionCurrency, target)
	if err != nil {
		return CollectionCurrencyChange{Currency: previous, Previous: previous}, err
	}
	change.CoinsConverted = coins
	change.Factor = factor
	change.RateDate = rates.Date
	return change, nil
}
