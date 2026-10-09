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
	ecbDailyRatesURL     = "https://www.ecb.europa.eu/stats/eurofxref/eurofxref-daily.xml"
	ecbResponseLimit     = 64 << 10
	exchangeRateCacheKey = "fx:ecb:usd"
	// The ECB publishes once per working day (around 16:00 CET).
	exchangeRateCacheTTL = 6 * time.Hour
)

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

// ExchangeRateService converts collection values, which the app keeps in US
// dollars, to euros with the ECB daily reference rate.
type ExchangeRateService struct {
	client ECBRatesClient
	cache  *NumistaResponseCache
}

// NewExchangeRateService creates the service. The response cache is shared
// with the Numista lookups; it stores any small JSON payload by key.
func NewExchangeRateService(client ECBRatesClient, cache *NumistaResponseCache) *ExchangeRateService {
	return &ExchangeRateService{client: client, cache: cache}
}

// USDPerEUR returns the latest ECB rate, cached for a few hours; when the ECB
// cannot be reached the last stored rate is used.
func (s *ExchangeRateService) USDPerEUR(ctx context.Context) (ExchangeRate, error) {
	body, _, _, err := s.cache.Get(ctx, exchangeRateCacheKey, exchangeRateCacheTTL, func(ctx context.Context) ([]byte, error) {
		raw, err := s.client.DailyRates(ctx)
		if err != nil {
			return nil, err
		}
		rate, err := parseECBDollarRate(raw)
		if err != nil {
			return nil, err
		}
		return json.Marshal(rate)
	})
	if err != nil {
		return ExchangeRate{}, err
	}
	var rate ExchangeRate
	if err := json.Unmarshal(body, &rate); err != nil || rate.USDPerEUR <= 0 {
		return ExchangeRate{}, errors.New("invalid stored exchange rate")
	}
	return rate, nil
}

// parseECBDollarRate reads the USD rate and its date from the ECB XML
// (<Cube><Cube time="…"><Cube currency="USD" rate="…"/>…).
func parseECBDollarRate(raw []byte) (ExchangeRate, error) {
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
		return ExchangeRate{}, err
	}
	for _, rate := range envelope.Cube.Daily.Rates {
		if strings.EqualFold(rate.Currency, "USD") {
			value, err := strconv.ParseFloat(strings.TrimSpace(rate.Rate), 64)
			if err != nil || value <= 0 {
				return ExchangeRate{}, errors.New("invalid USD rate")
			}
			return ExchangeRate{USDPerEUR: value, Date: envelope.Cube.Daily.Time}, nil
		}
	}
	return ExchangeRate{}, errors.New("USD rate not found")
}
