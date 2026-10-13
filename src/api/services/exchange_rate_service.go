package services

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	ecbDailyRatesURL      = "https://www.ecb.europa.eu/stats/eurofxref/eurofxref-daily.xml"
	ecbResponseLimit      = 64 << 10
	exchangeRatesCacheKey = "fx:ecb:all:v1"
	// The ECB publishes once per working day (around 16:00 CET).
	exchangeRateCacheTTL = 6 * time.Hour
)

// ExchangeRates are the ECB euro reference rates of a day: how much of each
// currency one euro buys.
type ExchangeRates struct {
	// Date is the ECB reference date, YYYY-MM-DD.
	Date string `json:"date"`
	// PerEUR is how much of each currency one euro buys ("USD": 1.1206).
	PerEUR map[string]float64 `json:"perEur"`
}

// ExchangeRate is the ECB euro reference rate for the US dollar.
type ExchangeRate struct {
	// USDPerEUR is how many US dollars one euro buys (e.g. 1.1206).
	USDPerEUR float64 `json:"usdPerEur"`
	// Date is the ECB reference date, YYYY-MM-DD.
	Date string `json:"date"`
}

// ECBRatesClient fetches the ECB daily reference rates XML.
type ECBRatesClient interface {
	DailyRates(ctx context.Context) ([]byte, error)
}

// HTTPECBClient downloads the ECB daily reference rates.
type HTTPECBClient struct {
	client *http.Client
	url    string
}

// NewHTTPECBClient creates a client; a nil client uses a default with a timeout.
func NewHTTPECBClient(client *http.Client) *HTTPECBClient {
	if client == nil {
		client = &http.Client{Timeout: 10 * time.Second}
	}
	return &HTTPECBClient{client: client, url: ecbDailyRatesURL}
}

// DailyRates returns the raw XML of today's ECB reference rates.
func (c *HTTPECBClient) DailyRates(ctx context.Context) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", CommonsUserAgent)
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ECB rates returned %d", resp.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, ecbResponseLimit+1))
	if err != nil {
		return nil, err
	}
	if len(body) > ecbResponseLimit {
		return nil, errors.New("ECB response exceeded limit")
	}
	return body, nil
}

// ExchangeRateService gives the ECB daily euro reference rates, to convert
// amounts between currencies and to switch the collection currency.
type ExchangeRateService struct {
	client ECBRatesClient
	cache  *NumistaResponseCache
}

// NewExchangeRateService creates the service. The response cache is shared
// with the Numista lookups; it stores any small JSON payload by key.
func NewExchangeRateService(client ECBRatesClient, cache *NumistaResponseCache) *ExchangeRateService {
	return &ExchangeRateService{client: client, cache: cache}
}

// Rate is how much of currency one euro buys: 1 for the euro; ok is false
// for a currency the ECB does not quote.
func (r ExchangeRates) Rate(currency string) (float64, bool) {
	currency = strings.ToUpper(strings.TrimSpace(currency))
	if currency == CurrencyEUR {
		return 1, true
	}
	rate, ok := r.PerEUR[currency]
	return rate, ok && rate > 0
}

// Rates returns the latest ECB reference rates, cached for a few hours; when
// the ECB cannot be reached the last stored rates are used.
func (s *ExchangeRateService) Rates(ctx context.Context) (ExchangeRates, error) {
	body, _, _, err := s.cache.Get(ctx, exchangeRatesCacheKey, exchangeRateCacheTTL, func(ctx context.Context) ([]byte, error) {
		raw, err := s.client.DailyRates(ctx)
		if err != nil {
			return nil, err
		}
		rates, err := parseECBRates(raw)
		if err != nil {
			return nil, err
		}
		return json.Marshal(rates)
	})
	if err != nil {
		return ExchangeRates{}, err
	}
	var rates ExchangeRates
	if err := json.Unmarshal(body, &rates); err != nil || len(rates.PerEUR) == 0 {
		return ExchangeRates{}, errors.New("invalid stored exchange rates")
	}
	return rates, nil
}

// USDPerEUR returns the latest ECB rate for the US dollar (see Rates).
func (s *ExchangeRateService) USDPerEUR(ctx context.Context) (ExchangeRate, error) {
	rates, err := s.Rates(ctx)
	if err != nil {
		return ExchangeRate{}, err
	}
	usd, ok := rates.Rate(CurrencyUSD)
	if !ok {
		return ExchangeRate{}, errors.New("USD rate not found")
	}
	return ExchangeRate{USDPerEUR: usd, Date: rates.Date}, nil
}

// parseECBRates reads every rate and their date from the ECB XML
// (<Cube><Cube time="…"><Cube currency="USD" rate="…"/>…).
func parseECBRates(raw []byte) (ExchangeRates, error) {
	var envelope struct {
		Cube struct {
			Daily struct {
				Time  string `xml:"time,attr"`
				Rates []struct {
					Currency string `xml:"currency,attr"`
					Rate     string `xml:"rate,attr"`
				} `xml:"Cube"`
			} `xml:"Cube"`
		} `xml:"Cube"`
	}
	if err := xml.Unmarshal(raw, &envelope); err != nil {
		return ExchangeRates{}, err
	}
	rates := ExchangeRates{Date: envelope.Cube.Daily.Time, PerEUR: map[string]float64{}}
	for _, rate := range envelope.Cube.Daily.Rates {
		value, err := strconv.ParseFloat(strings.TrimSpace(rate.Rate), 64)
		code := strings.ToUpper(strings.TrimSpace(rate.Currency))
		if err == nil && value > 0 && len(code) == 3 {
			rates.PerEUR[code] = value
		}
	}
	if len(rates.PerEUR) == 0 {
		return ExchangeRates{}, errors.New("no rates found")
	}
	return rates, nil
}
