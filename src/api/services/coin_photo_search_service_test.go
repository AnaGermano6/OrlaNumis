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
	commons := &fakeCommonsClient{responses: map[string]string{"50 Escudos 1968": testCommonsResponse}}
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
	commons.responses["Semis Augustus Gaius Caesar"] = testCommonsResponse

	got := svc.Search(context.Background(), "Semis - Augustus (Gaius Caesar)", 0, "en")
	if len(got.Photos) != 2 || strings.Join(commons.queries, "|") != "Semis - Augustus (Gaius Caesar)|Semis Augustus Gaius Caesar" {
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
