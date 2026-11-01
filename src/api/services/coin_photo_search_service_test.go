package services

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/briandenicola/ancient-coins-api/models"
)

type fakeCommonsClient struct {
	responses map[string]string
	err       error
	queries   []string
}

func (f *fakeCommonsClient) SearchFiles(_ context.Context, query string, _ int) ([]byte, error) {
	f.queries = append(f.queries, query)
	if f.err != nil {
		return nil, f.err
	}
	if body, ok := f.responses[query]; ok {
		return []byte(body), nil
	}
	return []byte(`{"batchcomplete": true}`), nil
}

// Shaped like a real formatversion=2 response; values are made up.
const testCommonsResponse = `{"query": {"pages": [
	{"title": "File:Catalogue of coins.pdf", "index": 1, "imageinfo": [{"mime": "application/pdf", "url": "https://upload.wikimedia.org/wikipedia/commons/a/aa/Catalogue.pdf"}]},
	{"title": "File:50 Escudos 1968 reverse.jpg", "index": 3, "imageinfo": [{"mime": "image/jpeg",
		"url": "https://upload.wikimedia.org/wikipedia/commons/b/bb/Rev.jpg", "thumburl": "https://upload.wikimedia.org/wikipedia/commons/thumb/b/bb/Rev.jpg/400px-Rev.jpg",
		"descriptionurl": "https://commons.wikimedia.org/wiki/File:Rev.jpg",
		"extmetadata": {"Artist": {"value": "<a href=\"//commons.wikimedia.org/wiki/User:Jane\" title=\"User:Jane\">Jane &amp; Co</a>"}, "LicenseShortName": {"value": "CC BY-SA 4.0"}}}]},
	{"title": "File:50 Escudos 1968 obverse.png", "index": 2, "imageinfo": [{"mime": "image/png",
		"url": "https://upload.wikimedia.org/wikipedia/commons/c/cc/Obv.png",
		"descriptionurl": "//commons.wikimedia.org/wiki/File:Obv.png",
		"extmetadata": {"LicenseShortName": {"value": "CC0"}}}]}
]}}`

func newPhotoSearchFixture() (*CoinPhotoSearchService, *fakeCommonsClient, *fakeNumistaCatalogClient) {
	catalog, _, numista, store := newCatalogFixture()
	_, _, _, clock := newMarketValueFixture()
	commons := &fakeCommonsClient{responses: map[string]string{"50 Escudos 1968 filetype:bitmap": testCommonsResponse}}
	return NewCoinPhotoSearchService(catalog, commons, NewNumistaResponseCache(store, clock, nil)), commons, numista
}

func TestCoinPhotoSearch_CombinesNumistaAndCommonsWithCredit(t *testing.T) {
	svc, commons, _ := newPhotoSearchFixture()

	got := svc.Search(context.Background(), "  50 Escudos   1968 ", 2933, "pt")

	if got.Status != models.NumistaDataStatusOK || got.Query != "50 Escudos 1968" || len(got.Photos) != 3 {
		t.Fatalf("got %+v", got)
	}
	numista := got.Photos[0]
	if numista.Source != models.CoinPhotoSourceNumista || numista.Side != "obverse" || numista.ImageURL != "https://en.numista.com/photos/a-original.jpg" || numista.Author != "Ana Example" {
		t.Fatalf("numista photo = %+v", numista)
	}
	obverse, reverse := got.Photos[1], got.Photos[2]
	if obverse.Title != "50 Escudos 1968 obverse" || obverse.License != "CC0" || obverse.Thumbnail != obverse.ImageURL || obverse.PageURL != "https://commons.wikimedia.org/wiki/File:Obv.png" {
		t.Fatalf("commons photos are not in search order or lack fields: %+v", obverse)
	}
	if reverse.Author != "Jane & Co" || reverse.License != "CC BY-SA 4.0" || !strings.HasSuffix(reverse.Thumbnail, "400px-Rev.jpg") {
		t.Fatalf("credit not cleaned: %+v", reverse)
	}

	svc.Search(context.Background(), "50 escudos 1968", 0, "pt")
	if len(commons.queries) != 1 {
		t.Fatalf("repeated search must be cached, got %v", commons.queries)
	}
}

func TestCoinPhotoSearch_RetriesWithoutPunctuationAndReportsFailures(t *testing.T) {
	svc, commons, _ := newPhotoSearchFixture()
	commons.responses["Semis Augustus Gaius Caesar filetype:bitmap"] = testCommonsResponse

	got := svc.Search(context.Background(), "Semis - Augustus (Gaius Caesar)", 0, "en")
	if len(got.Photos) != 2 || strings.Join(commons.queries, "|") != "Semis - Augustus (Gaius Caesar) filetype:bitmap|Semis Augustus Gaius Caesar filetype:bitmap" {
		t.Fatalf("photos = %d, queries = %v", len(got.Photos), commons.queries)
	}

	failing, failingCommons, _ := newPhotoSearchFixture()
	failingCommons.err = errors.New("offline")
	if result := failing.Search(context.Background(), "denarius", 0, "en"); result.Status != models.NumistaDataStatusUnavailable || len(result.Photos) != 0 {
		t.Fatalf("offline without Numista photos: %+v", result)
	}
	if result := failing.Search(context.Background(), "denarius", 2933, "en"); result.Status != models.NumistaDataStatusOK || len(result.Photos) != 1 {
		t.Fatalf("offline with Numista photos keeps them: %+v", result)
	}
}

func TestPlainText(t *testing.T) {
	if got := plainText(`<span>Foto: <b>José</b>&nbsp;Silva</span>`, 200); got != "Foto: José Silva" {
		t.Fatalf("plainText = %q", got)
	}
	if got := plainText(strings.Repeat("a", 10), 5); got != "aaaa…" {
		t.Fatalf("plainText truncation = %q", got)
	}
}

// A €2 commemorative's ECB image comes first; a regular €2 coin, a coin of
// another denomination or one without a year gets none.
func TestCoinPhotoSearch_AddsTheECBImageOfACommemorative(t *testing.T) {
	svc, _, _ := newPhotoSearchFixture()
	ecb := newTestEcbService(t, &fakeECBCoinsClient{page: ecb2022Page(t)})
	svc.WithECB(ecb)
	ctx := context.Background()

	got := svc.SearchCoin(ctx, CoinPhotoQuery{Year: 2022, Country: "Alemanha", Denomination: "2 euro", Subject: "Bundesländer – Thuringia", Lang: "pt"})
	if len(got.Photos) != 1 || got.Photos[0].Source != models.CoinPhotoSourceECB || got.Photos[0].Match != models.CoinPhotoMatchExact {
		t.Fatalf("commemorative = %+v", got.Photos)
	}
	for _, query := range []CoinPhotoQuery{
		{Year: 2022, Country: "Alemanha", Denomination: "2 euro"},
		{Year: 2022, Country: "Alemanha", Denomination: "5 euro", Subject: "Turíngia"},
		{Year: 2022, Country: "Alemanha", Denomination: "2 cêntimos de euro", Subject: "Turíngia"},
		{Country: "Alemanha", Denomination: "2 euro", Subject: "Turíngia"},
	} {
		if got := svc.SearchCoin(ctx, query); len(got.Photos) != 0 {
			t.Fatalf("%+v = %+v", query, got.Photos)
		}
	}
}

// Commons is searched in English first, as it names coin photos, with the
// year and then without it, then for the text typed; images only.
func TestCommonsPhotoQueries(t *testing.T) {
	coin := CoinPhotoQuery{Query: "200 liras 1995 – San Marino", Year: 1995, Country: "San Marino", Denomination: "200 liras"}
	want := []string{
		"200 lire San Marino 1995 coin filetype:bitmap",
		"200 lire San Marino coin filetype:bitmap",
		"200 liras 1995 – San Marino filetype:bitmap",
		"200 liras 1995 San Marino filetype:bitmap",
	}
	if got := commonsPhotoQueries(coin); strings.Join(got, "|") != strings.Join(want, "|") {
		t.Fatalf("queries = %q", got)
	}
	banknote := CoinPhotoQuery{Query: "20 escudos", Country: "Angola", Denomination: "20 escudos", Banknote: true}
	if got := commonsPhotoQueries(banknote); strings.Join(got, "|") != "20 escudos Angola banknote filetype:bitmap|20 escudos filetype:bitmap" {
		t.Fatalf("banknote without a year = %q", got)
	}
	if got := commonsPhotoQueries(CoinPhotoQuery{Query: "Angola 1969 coin"}); strings.Join(got, "|") != "Angola 1969 coin filetype:bitmap" {
		t.Fatalf("text only = %q", got)
	}
}

func TestEnglishDenomination(t *testing.T) {
	for denomination, want := range map[string]string{
		"200 liras":           "200 lire",
		"2,5 escudos":         "2.5 escudos",
		"7½ euro":             "7.5 euro",
		"½ rúpia":             "0.5 rupee",
		"10.000 reis":         "10000 reis",
		"10 cêntimos de euro": "10 euro cent",
		"1 penny novo":        "1 new penny",
		"5 Xelins":            "5 shillings",
		"1 Dólar":             "1 dollar",
		"5 kopiyok":           "5 kopiyok",
		"Two pounds":          "Two pounds",
		"":                    "",
	} {
		if got := englishDenomination(denomination); got != want {
			t.Errorf("englishDenomination(%q) = %q, want %q", denomination, got, want)
		}
	}
}

// The Commons photos of every search are kept in their order, once each; the
// search is unavailable only when Commons failed and nothing was found.
func TestCoinPhotoSearch_SearchesCommonsInEnglishFirst(t *testing.T) {
	svc, commons, _ := newPhotoSearchFixture()
	commons.responses["5 kopiyok Ukraine 1992 coin filetype:bitmap"] = testCommonsResponse
	commons.responses["5 kopiyok Ukraine coin filetype:bitmap"] = testCommonsResponse

	got := svc.SearchCoin(context.Background(), CoinPhotoQuery{Query: "5 kopiyok 1992 – Ucrânia", Year: 1992, Country: "Ukraine", Denomination: "5 kopiyok"})
	if got.Status != models.NumistaDataStatusOK || len(got.Photos) != 2 || got.Photos[0].Title != "50 Escudos 1968 obverse" {
		t.Fatalf("got %+v", got)
	}
	if len(commons.queries) != 4 || commons.queries[0] != "5 kopiyok Ukraine 1992 coin filetype:bitmap" {
		t.Fatalf("queries = %q", commons.queries)
	}
}
