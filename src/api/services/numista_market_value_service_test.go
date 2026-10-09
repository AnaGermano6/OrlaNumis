package services

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/briandenicola/ancient-coins-api/models"
)

type fakeNumistaPriceClient struct {
	issues      string
	prices      map[int]string
	issuesErr   error
	pricesErr   error
	issueCalls  int
	priceCalls  int
	lastIssueID int
	lastCurr    string
}

func (f *fakeNumistaPriceClient) Issues(_ context.Context, _ int) ([]byte, error) {
	f.issueCalls++
	if f.issuesErr != nil {
		return nil, f.issuesErr
	}
	return []byte(f.issues), nil
}

func (f *fakeNumistaPriceClient) IssuePrices(_ context.Context, _ int, issueID int, currency string) ([]byte, error) {
	f.priceCalls++
	f.lastIssueID = issueID
	f.lastCurr = currency
	if f.pricesErr != nil {
		return nil, f.pricesErr
	}
	body, ok := f.prices[issueID]
	if !ok {
		body = `{"currency":"EUR","prices":[]}`
	}
	return []byte(body), nil
}

type memoryNumistaCacheStore struct {
	entries map[string]models.NumistaCacheEntry
}

func (m *memoryNumistaCacheStore) Find(key string) (*models.NumistaCacheEntry, error) {
	entry, ok := m.entries[key]
	if !ok {
		return nil, nil
	}
	return &entry, nil
}

func (m *memoryNumistaCacheStore) Save(entry *models.NumistaCacheEntry) error {
	m.entries[entry.Key] = *entry
	return nil
}

type manualNumistaClock struct{ now time.Time }

func (c *manualNumistaClock) Now() time.Time { return c.now }

const testNumistaIssues = `[
	{"id": 101, "is_dated": true, "year": 1967, "gregorian_year": 1967, "mintage": 1000000},
	{"id": 102, "is_dated": true, "year": 1968, "gregorian_year": 1968, "mint_letter": "L", "mintage": 5000000, "comment": "Lisbon"}
]`

func newMarketValueFixture() (*NumistaMarketValueService, *fakeNumistaPriceClient, *memoryNumistaCacheStore, *manualNumistaClock) {
	client := &fakeNumistaPriceClient{
		issues: testNumistaIssues,
		prices: map[int]string{
			101: `{"currency":"EUR","prices":[{"grade":"vf","price":3}]}`,
			102: `{"currency":"EUR","prices":[{"grade":"unc","price":2.75},{"grade":"f","price":0.5},{"grade":"xf","price":0},{"grade":"vf","price":1}]}`,
		},
	}
	store := &memoryNumistaCacheStore{entries: map[string]models.NumistaCacheEntry{}}
	clock := &manualNumistaClock{now: time.Date(2026, 10, 9, 10, 0, 0, 0, time.UTC)}
	return NewNumistaMarketValueService(client, NewNumistaResponseCache(store, clock, nil)), client, store, clock
}

func linkedCoin(dateRange string) *models.Coin {
	return &models.Coin{
		DateRange:  dateRange,
		References: []models.CoinReference{{Catalog: "RIC", Number: "207"}, {Catalog: "Numista", Number: "2933"}},
	}
}

func TestNumistaMarketValue_NotLinkedSkipsProvider(t *testing.T) {
	svc, client, _, _ := newMarketValueFixture()

	got := svc.ForCoin(context.Background(), &models.Coin{References: []models.CoinReference{{Catalog: "RIC", Number: "1"}}}, 0, "EUR", "")

	if got.Status != models.NumistaDataStatusNotLinked || client.issueCalls+client.priceCalls != 0 {
		t.Fatalf("got %+v with %d calls", got, client.issueCalls+client.priceCalls)
	}
}

func TestNumistaMarketValue_FetchesMatchingIssueAndSortsGrades(t *testing.T) {
	svc, client, _, clock := newMarketValueFixture()

	got := svc.ForCoin(context.Background(), linkedCoin("1968"), 0, "EUR", "")

	if got.Status != models.NumistaDataStatusOK || got.Stale {
		t.Fatalf("status = %q stale = %v", got.Status, got.Stale)
	}
	if got.NumistaTypeID != 2933 || got.NumistaURL == "" || len(got.Issues) != 2 {
		t.Fatalf("unexpected type/issues: %+v", got)
	}
	if got.SelectedIssueID != 102 || client.lastIssueID != 102 || client.lastCurr != "EUR" {
		t.Fatalf("selected %d, requested %d %s; want issue 102 in EUR", got.SelectedIssueID, client.lastIssueID, client.lastCurr)
	}
	want := []models.NumistaGradePrice{{Grade: "f", Price: 0.5}, {Grade: "vf", Price: 1}, {Grade: "unc", Price: 2.75}}
	if fmt.Sprint(got.Prices) != fmt.Sprint(want) {
		t.Fatalf("prices = %v, want %v (sorted, zero prices dropped)", got.Prices, want)
	}
	if got.FetchedAt == nil || !got.FetchedAt.Equal(clock.now) {
		t.Fatalf("fetchedAt = %v", got.FetchedAt)
	}
	if issue := got.Issues[1]; issue.MintLetter != "L" || issue.Mintage == nil || *issue.Mintage != 5000000 || issue.Year == nil || *issue.Year != 1968 {
		t.Fatalf("issue summary = %+v", issue)
	}
}

func TestNumistaMarketValue_ReusesCacheUntilEachTTLExpires(t *testing.T) {
	svc, client, _, clock := newMarketValueFixture()
	ctx := context.Background()

	svc.ForCoin(ctx, linkedCoin("1968"), 0, "EUR", "")
	clock.now = clock.now.Add(6 * 24 * time.Hour)
	svc.ForCoin(ctx, linkedCoin("1968"), 0, "EUR", "")
	if client.issueCalls != 1 || client.priceCalls != 1 {
		t.Fatalf("within TTL: %d issue / %d price calls, want 1/1", client.issueCalls, client.priceCalls)
	}

	clock.now = clock.now.Add(2 * 24 * time.Hour) // prices are now 8 days old, issues 8 days
	got := svc.ForCoin(ctx, linkedCoin("1968"), 0, "EUR", "")
	if client.issueCalls != 1 || client.priceCalls != 2 {
		t.Fatalf("after price TTL: %d issue / %d price calls, want 1/2", client.issueCalls, client.priceCalls)
	}
	if !got.FetchedAt.Equal(clock.now) {
		t.Fatalf("fetchedAt not refreshed: %v", got.FetchedAt)
	}

	svc.ForCoin(ctx, linkedCoin("1968"), 0, "USD", "")
	if client.priceCalls != 3 || client.lastCurr != "USD" {
		t.Fatalf("a new currency must be fetched separately: %d calls, %s", client.priceCalls, client.lastCurr)
	}
}

func TestNumistaMarketValue_ServesStoredValuesWhenRefreshFails(t *testing.T) {
	svc, client, _, clock := newMarketValueFixture()
	ctx := context.Background()
	first := svc.ForCoin(ctx, linkedCoin("1968"), 0, "EUR", "")

	clock.now = clock.now.Add(40 * 24 * time.Hour)
	client.issuesErr = &NumistaError{Kind: NumistaErrorQuotaLimited}
	client.pricesErr = &NumistaError{Kind: NumistaErrorQuotaLimited}
	got := svc.ForCoin(ctx, linkedCoin("1968"), 0, "EUR", "")

	if got.Status != models.NumistaDataStatusOK || !got.Stale {
		t.Fatalf("status = %q stale = %v, want ok + stale", got.Status, got.Stale)
	}
	if fmt.Sprint(got.Prices) != fmt.Sprint(first.Prices) || !got.FetchedAt.Equal(*first.FetchedAt) {
		t.Fatalf("stale answer differs from stored one: %+v", got)
	}
}

func TestNumistaMarketValue_ReportsProviderProblemsWithoutCache(t *testing.T) {
	cases := map[NumistaErrorKind]string{
		NumistaErrorQuotaLimited: models.NumistaDataStatusQuotaLimited,
		NumistaErrorUnconfigured: models.NumistaDataStatusUnconfigured,
		NumistaErrorUnauthorized: models.NumistaDataStatusUnauthorized,
		NumistaErrorUnavailable:  models.NumistaDataStatusUnavailable,
	}
	for kind, want := range cases {
		svc, client, _, _ := newMarketValueFixture()
		client.issuesErr = &NumistaError{Kind: kind}

		got := svc.ForCoin(context.Background(), linkedCoin("1968"), 0, "EUR", "")

		if got.Status != want || len(got.Prices) != 0 || got.NumistaTypeID != 2933 {
			t.Errorf("%s: got %+v, want status %s", kind, got, want)
		}
	}
}

func TestNumistaMarketValue_HonoursRequestedIssueAndReportsMissingPrices(t *testing.T) {
	svc, client, _, _ := newMarketValueFixture()
	ctx := context.Background()

	got := svc.ForCoin(ctx, linkedCoin("1968"), 101, "EUR", "")
	if got.SelectedIssueID != 101 || len(got.Prices) != 1 || got.Prices[0].Grade != "vf" {
		t.Fatalf("requested issue ignored: %+v", got)
	}

	got = svc.ForCoin(ctx, linkedCoin(""), 999, "EUR", "")
	if got.SelectedIssueID != 101 {
		t.Fatalf("unknown issue should fall back to the first, got %d", got.SelectedIssueID)
	}

	delete(client.prices, 102)
	got = svc.ForCoin(ctx, linkedCoin("1968"), 0, "GBP", "")
	if got.Status != models.NumistaDataStatusNoPrices || len(got.Prices) != 0 {
		t.Fatalf("status = %q prices = %v, want no_prices", got.Status, got.Prices)
	}
}

func TestSingleCoinYear(t *testing.T) {
	cases := []struct {
		in   string
		year int
		ok   bool
	}{
		{"1968", 1968, true},
		{"AD 312", 312, true},
		{"44 BC", -44, true},
		{"c. 300 a.C.", -300, true},
		{"27 BC - AD 14", 0, false},
		{"1967-1968", 0, false},
		{"Imperial", 0, false},
		{"", 0, false},
	}
	for _, tc := range cases {
		year, ok := singleCoinYear(tc.in)
		if year != tc.year || ok != tc.ok {
			t.Errorf("singleCoinYear(%q) = %d, %v; want %d, %v", tc.in, year, ok, tc.year, tc.ok)
		}
	}
}
