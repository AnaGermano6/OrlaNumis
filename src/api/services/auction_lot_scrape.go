package services

import (
	"strings"

	"golang.org/x/net/html"
)

// An auction lot as read from a provider's public lot page (CNG's), to be
// saved as an auction lot.
type WatchlistLot struct {
	URL          string   `json:"url"`
	SourceLotID  string   `json:"sourceLotId"`
	SourceSaleID string   `json:"sourceSaleId"`
	SaleID       string   `json:"saleId"`
	LotNumber    int      `json:"lotNumber"`
	Title        string   `json:"title"`
	ImageURL     string   `json:"imageUrl"`
	Estimate     *float64 `json:"estimate"`
	CurrentBid   *float64 `json:"currentBid"`
	MaxBid       *float64 `json:"maxBid"`
	Currency     string   `json:"currency"`
	AuctionHouse string   `json:"auctionHouse"`
	SaleName     string   `json:"saleName"`
	SaleDate     string   `json:"saleDate"`
	Description  string   `json:"description"`

	// ProviderStatus, SoldPrice, and WinningCustomerRowID are populated by providers that
	// expose a definitive closed-lot outcome (currently CNG only) and are used by the
	// watchlist sync service to auto-detect won/lost instead of requiring a manual override.
	ProviderStatus       string   `json:"providerStatus,omitempty"`
	SoldPrice            *float64 `json:"soldPrice,omitempty"`
	WinningCustomerRowID string   `json:"-"`
}

// cleanHTML is the text of an HTML fragment, its whitespace collapsed.
func cleanHTML(s string) string {
	tokenizer := html.NewTokenizer(strings.NewReader(s))
	var result strings.Builder
	for {
		tt := tokenizer.Next()
		if tt == html.ErrorToken {
			break
		}
		if tt == html.TextToken {
			result.WriteString(tokenizer.Token().Data)
		}
	}
	// Normalize whitespace
	text := result.String()
	text = strings.Join(strings.Fields(text), " ")
	return strings.TrimSpace(text)
}
