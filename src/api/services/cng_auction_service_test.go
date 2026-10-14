package services

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCNGAuctionService_ParseLotPage(t *testing.T) {
	svc := NewCNGAuctionService(nil)
	lot, err := svc.parseLotPage(cngLotFixture())
	if err != nil {
		t.Fatalf("parseLotPage returned error: %v", err)
	}

	if lot.SourceLotID != "4-LOTID" {
		t.Fatalf("SourceLotID = %q, want 4-LOTID", lot.SourceLotID)
	}
	if lot.SourceSaleID != "4-SALEID" {
		t.Fatalf("SourceSaleID = %q, want 4-SALEID", lot.SourceSaleID)
	}
	if lot.LotNumber != 43 {
		t.Fatalf("LotNumber = %d, want 43", lot.LotNumber)
	}
	if lot.URL != "https://auctions.cngcoins.com/lots/view/4-LOTID/test-lot" {
		t.Fatalf("URL = %q", lot.URL)
	}
	if lot.ImageURL != "https://images.example/43_1.jpg" {
		t.Fatalf("ImageURL = %q", lot.ImageURL)
	}
	if lot.Estimate == nil || *lot.Estimate != 100 {
		t.Fatalf("Estimate = %v, want 100", lot.Estimate)
	}
	// No bid_amount in fixture — should fall back to starting_price (60)
	if lot.CurrentBid == nil || *lot.CurrentBid != 60 {
		t.Fatalf("CurrentBid = %v, want 60 (starting_price fallback)", lot.CurrentBid)
	}
	if lot.MaxBid != nil {
		t.Fatalf("MaxBid = %v, want nil (no autobid in fixture)", lot.MaxBid)
	}
	if lot.Currency != "USD" {
		t.Fatalf("Currency = %q, want USD", lot.Currency)
	}
	if lot.SaleName != "Electronic Auction 612" {
		t.Fatalf("SaleName = %q", lot.SaleName)
	}
	if lot.Description == "" || strings.Contains(lot.Description, "<b>") {
		t.Fatalf("Description was not cleaned: %q", lot.Description)
	}
}

// TestCNGAuctionService_ParseLotPage_UsesRealBidFields verifies that timed_auction_bid.amount
// takes precedence over starting_price for CurrentBid, that absentee_bid.max_bid is mapped to
// MaxBid, and that the lot's own extended_end_time is used for the sale date rather than the
// whole auction's effective_end_time. CNG has no "bid_amount"/"autobid" fields; earlier
// versions of this test asserted against those dead keys and never caught that they'd stopped
// existing in CNG's real API response (see F021 / the CNG provider audit for issue #482).
func TestCNGAuctionService_ParseLotPage_UsesRealBidFields(t *testing.T) {
	svc := NewCNGAuctionService(nil)
	lot, err := svc.parseLotPage(cngLotWithBidsFixture())
	if err != nil {
		t.Fatalf("parseLotPage returned error: %v", err)
	}
	// timed_auction_bid.amount (1500) must take precedence over starting_price (300)
	if lot.CurrentBid == nil || *lot.CurrentBid != 1500 {
		t.Fatalf("CurrentBid = %v, want 1500 (timed_auction_bid.amount)", lot.CurrentBid)
	}
	// absentee_bid.max_bid should be mapped to MaxBid
	if lot.MaxBid == nil || *lot.MaxBid != 1800 {
		t.Fatalf("MaxBid = %v, want 1800 (absentee_bid.max_bid)", lot.MaxBid)
	}
	// The lot's own extended_end_time (14:04) must win over the auction-wide
	// effective_end_time (16:04) for a multi-lot staggered timed auction.
	if lot.SaleDate != "2026-07-22T14:04:20Z" {
		t.Fatalf("SaleDate = %q, want lot.extended_end_time, not auction.effective_end_time", lot.SaleDate)
	}
	if lot.WinningCustomerRowID != "4-KQ7PSC" {
		t.Fatalf("WinningCustomerRowID = %q, want the timed_auction_bid.registration.customer.row_id", lot.WinningCustomerRowID)
	}
}

// TestCNGAuctionService_ScrapeLotWithClientUsesRealBidFields verifies that ScrapeLotWithClient
// uses the provided HTTP client and returns real current-bid/max-bid values. This path backs
// the manual "paste a lot URL" import (ImportFromURL), which fetches a single lot page directly
// rather than going through the watched-lots list. Note: timed_auction_bid.amount is public
// data present even without authentication — only absentee_bid.max_bid requires a logged-in
// client, since it's the specific user's own bid ceiling.
func TestCNGAuctionService_ScrapeLotWithClientUsesRealBidFields(t *testing.T) {
	var lotPageRequests []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lotPageRequests = append(lotPageRequests, r.URL.Path)
		w.Write([]byte(cngLotWithBidsFixture()))
	}))
	defer server.Close()

	restore := overrideCNGURLs(server.URL)
	defer restore()

	svc := NewCNGAuctionService(nil)
	client := server.Client()

	lot, err := svc.ScrapeLotWithClient(client, "https://auctions.cngcoins.com/lots/view/4-LOTID/thasos-drachm")
	if err != nil {
		t.Fatalf("ScrapeLotWithClient returned error: %v", err)
	}
	// timed_auction_bid.amount (1500) from the lot page must be returned as CurrentBid.
	if lot.CurrentBid == nil || *lot.CurrentBid != 1500 {
		t.Fatalf("CurrentBid = %v, want 1500 (timed_auction_bid.amount from lot page)", lot.CurrentBid)
	}
	// absentee_bid.max_bid (1800) must be returned as MaxBid.
	if lot.MaxBid == nil || *lot.MaxBid != 1800 {
		t.Fatalf("MaxBid = %v, want 1800 (absentee_bid.max_bid from lot page)", lot.MaxBid)
	}
	if len(lotPageRequests) == 0 {
		t.Fatal("no request was made to the test server — ScrapeLotWithClient did not fetch the page")
	}
}

func TestCNGAuctionService_ScrapeLotRejectsNonCNGURL(t *testing.T) {
	svc := NewCNGAuctionService(nil)
	tests := []string{
		"http://auctions.cngcoins.com/lots/view/4-LOT/test",
		"https://example.com/lots/view/4-LOT/test",
		"https://localhost/lots/view/4-LOT/test",
		"https://127.0.0.1/lots/view/4-LOT/test",
		"https://auctions.cngcoins.com.evil.example/lots/view/4-LOT/test",
		"https://attacker@example.com@auctions.cngcoins.com/lots/view/4-LOT/test",
		"https://auctions.cngcoins.com:8443/lots/view/4-LOT/test",
		"https://auctions.cngcoins.com/lots/view/4-LOT/test?next=https://example.com",
		"https://auctions.cngcoins.com/lots/view/4-LOT/test#fragment",
		"https://auctions.cngcoins.com/auctions/4-SALE/test",
		"://bad",
	}
	for _, rawURL := range tests {
		t.Run(rawURL, func(t *testing.T) {
			if _, err := svc.ScrapeLot(rawURL); err == nil {
				t.Fatal("ScrapeLot succeeded, want URL validation error")
			}
		})
	}
}

func TestCanonicalCNGLotPath(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want string
	}{
		{
			name: "basic lot path",
			raw:  "https://auctions.cngcoins.com/lots/view/4-LOT/test-lot",
			want: "/lots/view/4-LOT/test-lot",
		},
		{
			name: "standard https port",
			raw:  "https://auctions.cngcoins.com:443/lots/view/4-LOT/test-lot",
			want: "/lots/view/4-LOT/test-lot",
		},
		{
			name: "trailing slash",
			raw:  "https://auctions.cngcoins.com/lots/view/4-LOT/test-lot/",
			want: "/lots/view/4-LOT/test-lot/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := canonicalCNGLotPath(tt.raw)
			if err != nil {
				t.Fatalf("canonicalCNGLotPath returned error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("canonicalCNGLotPath = %q, want %q", got, tt.want)
			}
		})
	}
}

func cngLotFixture() string {
	return `<!doctype html><html><script>
viewVars = {
  "currentRouteName":"lot-detail-slug",
  "lot":{
    "row_id":"4-LOTID",
    "lot_number":43,
    "lot_number_extension":"",
    "title":"CARTHAGE. Half-Shekel. Good VF.",
    "description":"<b>CARTHAGE.</b> Second Punic War. Good VF.",
    "estimate_low":"100.00",
    "estimate_high":"150.00",
    "currency_code":"USD",
    "starting_price":"60.00",
    "sold_price":null,
    "status":"active",
    "_detail_url":"/lots/view/4-LOTID/test-lot",
    "cover_thumbnail":"",
    "images":[{"detail_url":"https://images.example/43_1.jpg","thumbnail_url":"https://images.example/thumb.jpg"}],
    "auction":{
      "row_id":"4-SALEID",
      "title":"Electronic Auction 612",
      "currency_code":"USD",
      "time_start":"2026-06-17T20:00:00Z",
      "effective_end_time":"2026-07-01T21:15:00Z"
    }
  }
};
</script></html>`
}

// cngLotWithBidsFixture mirrors the real viewVars.lot shape captured from a live,
// authenticated CNG lot page with active bidding (see F021 / the CNG provider audit for
// issue #482). CNG has no "bid_amount"/"autobid" fields — the current bid lives at
// timed_auction_bid.amount (public) and the user's own ceiling at absentee_bid.max_bid
// (auth-only). extended_end_time is this lot's own close time and is deliberately
// different here from auction.effective_end_time (the whole sale's close time) to catch
// any regression back to reading the wrong one.
func cngLotWithBidsFixture() string {
	return `<!doctype html><html><script>
viewVars = {
  "currentRouteName":"lot-detail-slug",
  "lot":{
    "row_id":"4-LOTID",
    "lot_number":14,
    "lot_number_extension":"",
    "title":"ISLANDS off THRACE, Thasos. AR Drachm.",
    "description":"<b>ISLANDS off THRACE.</b> Thasos. Circa 412-404 BC. Good VF.",
    "estimate_low":"500.00",
    "estimate_high":"750.00",
    "currency_code":"USD",
    "starting_price":"300.00",
    "sold_price":null,
    "status":"active",
    "extended_end_time":"2026-07-22T14:04:20Z",
    "timed_auction_bid":{
      "amount":"1500.00",
      "registration":{"customer":{"row_id":"4-KQ7PSC"}}
    },
    "absentee_bid":{"max_bid":"1800.00"},
    "_detail_url":"/lots/view/4-LOTID/thasos-drachm",
    "cover_thumbnail":"https://images.example/14_1.jpg",
    "images":[],
    "auction":{
      "row_id":"4-SALEID",
      "title":"Keystone 17 - The W. Toliver Besson Collection",
      "currency_code":"USD",
      "time_start":"2026-07-01T12:00:00Z",
      "effective_end_time":"2026-07-22T16:04:20Z"
    }
  }
};
</script></html>`
}

func overrideCNGURLs(base string) func() {
	oldBase := cngBase
	cngBase = base
	return func() { cngBase = oldBase }
}
