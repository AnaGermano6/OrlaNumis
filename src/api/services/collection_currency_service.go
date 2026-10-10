package services

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

// Collection currencies: the ones the ECB euro reference rate converts between.
const (
	CurrencyUSD = "USD"
	CurrencyEUR = "EUR"
)

var (
	// ErrUnsupportedCollectionCurrency is returned for a currency other than USD or EUR.
	ErrUnsupportedCollectionCurrency = errors.New("collection currency must be USD or EUR")
	// ErrExchangeRateUnavailable means no ECB rate could be read, so nothing was converted.
	ErrExchangeRateUnavailable = errors.New("exchange rate unavailable")
)

// NormalizeCollectionCurrency returns USD or EUR for a currency code, and false
// for anything else.
func NormalizeCollectionCurrency(code string) (string, bool) {
	switch strings.ToUpper(strings.TrimSpace(code)) {
	case CurrencyUSD:
		return CurrencyUSD, true
	case CurrencyEUR:
		return CurrencyEUR, true
	}
	return "", false
}

// CollectionCurrency is the currency coin prices and values are kept in;
// USD unless an admin switched the collection to euros.
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
	return fmt.Sprintf("Give estimatedValue and every comparable price in %s (euros), not USD; convert when a source is priced in another currency.", currency)
}

// CollectionCurrencyChange reports a currency switch.
type CollectionCurrencyChange struct {
	Currency string `json:"currency"`
	Previous string `json:"previous"`
	// ExchangeRate is the ECB rate used; nil when nothing was converted.
	ExchangeRate *ExchangeRate `json:"exchangeRate,omitempty"`
	// CoinsConverted counts coins with a price or value that was converted.
	CoinsConverted int64 `json:"coinsConverted"`
}

type collectionCurrencyRates interface {
	USDPerEUR(ctx context.Context) (ExchangeRate, error)
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

	rate, err := s.rates.USDPerEUR(ctx)
	if err != nil {
		return change, fmt.Errorf("%w: %v", ErrExchangeRateUnavailable, err)
	}
	if rate.USDPerEUR <= 0 {
		return change, ErrExchangeRateUnavailable
	}
	factor := rate.USDPerEUR // EUR → USD
	if target == CurrencyEUR {
		factor = 1 / rate.USDPerEUR // USD → EUR
	}
	coins, err := s.amounts.ConvertAmounts(factor, SettingCollectionCurrency, target)
	if err != nil {
		return CollectionCurrencyChange{Currency: previous, Previous: previous}, err
	}
	change.CoinsConverted = coins
	change.ExchangeRate = &rate
	return change, nil
}
