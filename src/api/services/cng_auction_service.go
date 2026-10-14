package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var ErrCNGAuthenticationRequired = errors.New("cng authentication required")

const (
	cngHost      = "auctions.cngcoins.com"
	cngUserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) " +
		"AppleWebKit/537.36 (KHTML, like Gecko) " +
		"Chrome/138.0.0.0 Safari/537.36"
)

var (
	cngBase          = "https://auctions.cngcoins.com"
	cngLotPathRe     = regexp.MustCompile(`^/lots/view/([^/]+)(?:/|$)`)
	cngLotSafePathRe = regexp.MustCompile(`^/lots/view/[A-Za-z0-9._~-]+(?:/[A-Za-z0-9._~%-]+)?/?$`)
)

// CNGAuctionService handles HTTP interactions with auctions.cngcoins.com.
type CNGAuctionService struct {
	logger *Logger
}

// NewCNGAuctionService creates a new CNGAuctionService.
func NewCNGAuctionService(logger *Logger) *CNGAuctionService {
	return &CNGAuctionService{logger: logger}
}

// ScrapeLot fetches a CNG lot page using the default (unauthenticated) HTTP client.
func (s *CNGAuctionService) ScrapeLot(lotURL string) (WatchlistLot, error) {
	return s.ScrapeLotWithClient(http.DefaultClient, lotURL)
}

// ScrapeLotWithClient fetches a CNG lot page using the provided HTTP client and extracts a
// source-aware lot summary. Pass the authenticated client from Login during watchlist sync
// to obtain user-specific data (current bid_amount and autobid) that are only returned on
// the individual lot page, not the watched-lots list.
func (s *CNGAuctionService) ScrapeLotWithClient(client *http.Client, lotURL string) (WatchlistLot, error) {
	lotPath, err := canonicalCNGLotPath(lotURL)
	if err != nil {
		return WatchlistLot{}, err
	}
	req, err := newScraperRequest(http.MethodGet, cngBase, nil, cngDefaultHeaders())
	if err != nil {
		return WatchlistLot{}, err
	}
	req.URL.Path = lotPath

	body, err := doScraperRequest(client, req, "lot page")
	if err != nil {
		return WatchlistLot{}, err
	}
	lot, err := s.parseLotPage(string(body))
	if err != nil {
		return WatchlistLot{}, err
	}
	return lot, nil
}

func (s *CNGAuctionService) parseLotPage(rawHTML string) (WatchlistLot, error) {
	var root cngViewVars
	if err := parseCNGViewVars(rawHTML, &root); err != nil {
		return WatchlistLot{}, err
	}
	if root.Lot == nil {
		return WatchlistLot{}, fmt.Errorf("cng lot page missing viewVars.lot")
	}
	return cngLotToWatchlistLot(*root.Lot), nil
}

func parseCNGViewVars(rawHTML string, target interface{}) error {
	idx := strings.Index(rawHTML, "viewVars = ")
	if idx < 0 {
		return fmt.Errorf("cng page missing viewVars")
	}
	start := strings.Index(rawHTML[idx:], "{")
	if start < 0 {
		return fmt.Errorf("cng page missing viewVars object")
	}
	start += idx
	end, err := findJSONObjectEnd(rawHTML, start)
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(rawHTML[start:end+1]), target)
}

func findJSONObjectEnd(raw string, start int) (int, error) {
	depth := 0
	inString := false
	escaped := false
	for i := start; i < len(raw); i++ {
		ch := raw[i]
		if inString {
			if escaped {
				escaped = false
				continue
			}
			if ch == '\\' {
				escaped = true
				continue
			}
			if ch == '"' {
				inString = false
			}
			continue
		}
		switch ch {
		case '"':
			inString = true
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return i, nil
			}
		}
	}
	return 0, fmt.Errorf("cng viewVars object is incomplete")
}

type cngViewVars struct {
	Lot  *cngLot `json:"lot"`
	Lots struct {
		ResultPage []cngLot     `json:"result_page"`
		QueryInfo  cngQueryInfo `json:"query_info"`
	} `json:"lots"`
}

type cngQueryInfo struct {
	TotalNumResults int `json:"total_num_results"`
	PageSize        int `json:"page_size"`
}

type cngLot struct {
	RowID                string              `json:"row_id"`
	LotNumber            int                 `json:"lot_number"`
	LotNumberExtension   string              `json:"lot_number_extension"`
	Title                string              `json:"title"`
	Description          string              `json:"description"`
	TruncatedDescription string              `json:"truncated_description"`
	EstimateLow          string              `json:"estimate_low"`
	EstimateHigh         string              `json:"estimate_high"`
	CurrencyCode         string              `json:"currency_code"`
	StartingPrice        string              `json:"starting_price"`
	SoldPrice            string              `json:"sold_price"`
	Status               string              `json:"status"`
	DetailURL            string              `json:"_detail_url"`
	CoverThumbnail       string              `json:"cover_thumbnail"`
	Images               []cngImage          `json:"images"`
	Auction              cngAuction          `json:"auction"`
	ExtendedEndTime      string              `json:"extended_end_time"`
	TimedAuctionBid      *cngTimedAuctionBid `json:"timed_auction_bid"`
	AbsenteeBid          *cngAbsenteeBid     `json:"absentee_bid"`
}

// cngTimedAuctionBid is the current (or, once the lot closes, final) bid on a timed-auction
// lot. It is public data: present even for an unauthenticated request. The nested
// registration/customer identifies whose bid it is, which is how a closed lot's winner is
// determined — compare against the logged-in user's own customer row_id.
type cngTimedAuctionBid struct {
	Amount       string `json:"amount"`
	Registration struct {
		Customer struct {
			RowID string `json:"row_id"`
		} `json:"customer"`
	} `json:"registration"`
}

// cngAbsenteeBid is the authenticated user's own bid ceiling on a lot. Only populated when
// fetched with a logged-in client for a lot the user has actually placed a bid on.
type cngAbsenteeBid struct {
	MaxBid string `json:"max_bid"`
}

type cngImage struct {
	DetailURL    string `json:"detail_url"`
	ThumbnailURL string `json:"thumbnail_url"`
}

type cngAuction struct {
	RowID            string `json:"row_id"`
	Title            string `json:"title"`
	CurrencyCode     string `json:"currency_code"`
	TimeStart        string `json:"time_start"`
	EffectiveEndTime string `json:"effective_end_time"`
}

func cngLotToWatchlistLot(lot cngLot) WatchlistLot {
	currency := firstNonEmpty(lot.CurrencyCode, lot.Auction.CurrencyCode, "USD")
	imageURL := lot.CoverThumbnail
	if imageURL == "" && len(lot.Images) > 0 {
		imageURL = firstNonEmpty(lot.Images[0].DetailURL, lot.Images[0].ThumbnailURL)
	}

	// timed_auction_bid.amount is the current (or, once the lot closes, final) bid and is
	// public data; fall back to starting_price only when no bid has been placed yet.
	currentBidStr := lot.StartingPrice
	winningCustomerRowID := ""
	if lot.TimedAuctionBid != nil {
		if lot.TimedAuctionBid.Amount != "" {
			currentBidStr = lot.TimedAuctionBid.Amount
		}
		winningCustomerRowID = lot.TimedAuctionBid.Registration.Customer.RowID
	}
	currentBid, _ := parseCNGDecimal(currentBidStr)

	// absentee_bid.max_bid is the authenticated user's own bid ceiling; nil unless fetched
	// with a logged-in client for a lot the user has actually bid on.
	var maxBid *float64
	if lot.AbsenteeBid != nil {
		maxBid, _ = parseCNGDecimal(lot.AbsenteeBid.MaxBid)
	}

	soldPrice, _ := parseCNGDecimal(lot.SoldPrice)
	estimate, _ := parseCNGDecimal(firstNonEmpty(lot.EstimateLow, lot.EstimateHigh))
	// extended_end_time is this specific lot's own close time. auction.effective_end_time is
	// when the entire (often multi-hundred-lot) sale ends, which can be hours later for any
	// individual lot in a staggered timed auction.
	saleDate := firstNonEmpty(lot.ExtendedEndTime, lot.Auction.EffectiveEndTime, lot.Auction.TimeStart)
	description := cleanHTML(firstNonEmpty(lot.Description, lot.TruncatedDescription))
	if len(description) > 2000 {
		description = description[:2000]
	}

	return WatchlistLot{
		URL:                  normalizeCNGURL(lot.DetailURL),
		SourceLotID:          lot.RowID,
		SourceSaleID:         lot.Auction.RowID,
		SaleID:               lot.Auction.RowID,
		LotNumber:            lot.LotNumber,
		Title:                strings.TrimSpace(lot.Title),
		ImageURL:             imageURL,
		Estimate:             estimate,
		CurrentBid:           currentBid,
		MaxBid:               maxBid,
		Currency:             strings.ToUpper(currency),
		AuctionHouse:         "Classical Numismatic Group",
		SaleName:             strings.TrimSpace(lot.Auction.Title),
		SaleDate:             saleDate,
		Description:          description,
		ProviderStatus:       strings.ToLower(strings.TrimSpace(lot.Status)),
		SoldPrice:            soldPrice,
		WinningCustomerRowID: winningCustomerRowID,
	}
}

func normalizeCNGURL(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}

	parsed, err := url.Parse(raw)
	if err == nil && parsed.IsAbs() {
		return parsed.String()
	}
	if strings.HasPrefix(raw, "/") {
		return cngBase + raw
	}
	return cngBase + "/" + raw
}

func parseCNGLotID(rawURL string) string {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}
	if match := cngLotPathRe.FindStringSubmatch(parsed.Path); match != nil {
		return match[1]
	}
	return ""
}

func canonicalCNGLotPath(rawURL string) (string, error) {
	parsed, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil {
		return "", fmt.Errorf("invalid CNG lot URL: %w", err)
	}
	if parsed.Scheme != "https" || strings.ToLower(parsed.Hostname()) != cngHost || parsed.User != nil {
		return "", fmt.Errorf("CNG lot URL must be on https://auctions.cngcoins.com")
	}
	if port := parsed.Port(); port != "" && port != "443" {
		return "", fmt.Errorf("CNG lot URL must use the standard HTTPS port")
	}
	if parsed.RawQuery != "" || parsed.Fragment != "" {
		return "", fmt.Errorf("CNG lot URL must not include query parameters or fragments")
	}
	path := parsed.EscapedPath()
	if !cngLotSafePathRe.MatchString(path) || parseCNGLotID(path) == "" {
		return "", fmt.Errorf("CNG lot URL must be a /lots/view/ URL")
	}
	return path, nil
}

func parseCNGDecimal(raw string) (*float64, string) {
	raw = strings.TrimSpace(strings.ReplaceAll(raw, ",", ""))
	if raw == "" {
		return nil, ""
	}
	val, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return nil, ""
	}
	return &val, ""
}

func ParseCNGDate(raw string) *time.Time {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	for _, layout := range []string{time.RFC3339, "2006-01-02T15:04:05Z"} {
		if t, err := time.Parse(layout, raw); err == nil {
			return &t
		}
	}
	return nil
}

func cngDefaultHeaders() map[string]string {
	return map[string]string{
		"User-Agent": cngUserAgent,
	}
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func (s *CNGAuctionService) trace(format string, args ...interface{}) {
	if s.logger != nil {
		s.logger.Trace("cng", format, args...)
	}
}

func (s *CNGAuctionService) debug(format string, args ...interface{}) {
	if s.logger != nil {
		s.logger.Debug("cng", format, args...)
	}
}

func (s *CNGAuctionService) info(format string, args ...interface{}) {
	if s.logger != nil {
		s.logger.Info("cng", format, args...)
	}
}

func (s *CNGAuctionService) warn(format string, args ...interface{}) {
	if s.logger != nil {
		s.logger.Warn("cng", format, args...)
	}
}

func (s *CNGAuctionService) error(format string, args ...interface{}) {
	if s.logger != nil {
		s.logger.Error("cng", format, args...)
	}
}
