package services

import (
	"context"
	"strings"
	"testing"

	"github.com/briandenicola/ancient-coins-api/models"
)

type fakeNumistaCatalogClient struct {
	fakeNumistaPriceClient
	searchBody   string
	detailBody   string
	issuersBody  string
	searchErr    error
	detailErr    error
	searchCalls  int
	detailCalls  int
	issuerCalls  int
	lastSearch   NumistaTypeSearch
	lastLanguage string
}

func (f *fakeNumistaCatalogClient) SearchTypes(_ context.Context, search NumistaTypeSearch) ([]byte, error) {
	f.searchCalls++
	f.lastSearch = search
	if f.searchErr != nil {
		return nil, f.searchErr
	}
	return []byte(f.searchBody), nil
}

func (f *fakeNumistaCatalogClient) TypeDetail(_ context.Context, _ int, lang string) ([]byte, error) {
	f.detailCalls++
	f.lastLanguage = lang
	if f.detailErr != nil {
		return nil, f.detailErr
	}
	return []byte(f.detailBody), nil
}

func (f *fakeNumistaCatalogClient) Issuers(_ context.Context, lang string) ([]byte, error) {
	f.issuerCalls++
	f.lastLanguage = lang
	return []byte(f.issuersBody), nil
}

// Shaped like real /types/{id} responses; values are made up.
const testCatalogTypeDetail = `{
	"id": 2933, "url": "https://pt.numista.com/2933", "title": "2.50 Escudos", "category": "coin",
	"issuer": {"code": "angola", "name": "Angola"},
	"min_year": 1953, "max_year": 1974,
	"value": {"text": "2.5 Escudos", "numeric_value": 2.5, "currency": {"id": 390, "name": "Escudo"}},
	"ruler": [{"id": 1, "name": "Província ultramarina de Portugal"}],
	"composition": {"text": "Cuproníquel"}, "technique": {"text": "Fresado"},
	"weight": 3.5, "size": 20, "thickness": 1.46, "shape": "Redondo",
	"mints": [{"id": 1, "name": "Imprensa Nacional - Casa da Moeda"}],
	"obverse": {"description": "Brasão de armas", "lettering": "ANGOLA\r\n2$50",
		"picture": "https://en.numista.com/photos/a-original.jpg", "thumbnail": "https://en.numista.com/photos/a-180.jpg",
		"picture_copyright": "Ana Example", "picture_license_name": "CC BY-NC"},
	"reverse": {"description": "Brasão de Portugal", "lettering": "REPÚBLICA·PORTUGUESA",
		"thumbnail": "http://insecure.example/b-180.jpg"},
	"edge": {"description": "Serrilhado"},
	"references": [{"catalogue": {"id": 3, "code": "KM"}, "number": "77"}, {"catalogue": {"id": 24, "code": "Schön"}, "number": "17"}, {"catalogue": {"code": ""}, "number": "1"}]
}`

func newCatalogFixture() (*NumistaCatalogService, *NumistaMarketValueService, *fakeNumistaCatalogClient, *memoryNumistaCacheStore) {
	_, priceClient, store, clock := newMarketValueFixture()
	client := &fakeNumistaCatalogClient{fakeNumistaPriceClient: *priceClient, detailBody: testCatalogTypeDetail}
	cache := NewNumistaResponseCache(store, clock, nil)
	catalog := NewNumistaCatalogService(client, cache)
	marketValue := NewNumistaMarketValueService(client, cache).WithCatalog(catalog)
	return catalog, marketValue, client, store
}

func TestNumistaCatalog_TypeMapsDetailsForPrefill(t *testing.T) {
	catalog, _, client, _ := newCatalogFixture()

	got := catalog.Type(context.Background(), 2933, "pt")

	if got.Status != models.NumistaDataStatusOK || got.Type == nil {
		t.Fatalf("status = %q type = %v", got.Status, got.Type)
	}
	detail := got.Type
	if detail.URL != "https://en.numista.com/catalogue/pieces2933.html" || detail.Title != "2.50 Escudos" || detail.Issuer.Name != "Angola" {
		t.Fatalf("identity: %+v", detail)
	}
	if detail.Denomination != "2.5 Escudos" || detail.Currency != "Escudo" || detail.Composition != "Cuproníquel" || detail.Shape != "Redondo" {
		t.Fatalf("value/composition: %+v", detail)
	}
	if *detail.WeightGrams != 3.5 || *detail.DiameterMm != 20 || *detail.ThicknessMm != 1.46 {
		t.Fatalf("measures: %v %v %v", *detail.WeightGrams, *detail.DiameterMm, *detail.ThicknessMm)
	}
	if len(detail.Rulers) != 1 || len(detail.Mints) != 1 || detail.Edge.Description != "Serrilhado" {
		t.Fatalf("rulers/mints/edge: %+v", detail)
	}
	if detail.Obverse.Lettering != "ANGOLA\n2$50" {
		t.Fatalf("lettering not normalised: %q", detail.Obverse.Lettering)
	}
	if len(detail.References) != 2 || detail.References[0] != (models.NumistaCatalogReference{Catalogue: "KM", Number: "77"}) {
		t.Fatalf("references: %+v", detail.References)
	}
	if len(detail.Photos) != 1 || detail.Photos[0].Side != "obverse" || detail.Photos[0].Copyright != "Ana Example" || detail.Photos[0].License != "CC BY-NC" {
		t.Fatalf("photos (https only, with credit): %+v", detail.Photos)
	}
	if len(got.Issues) != 2 || client.lastLanguage != "pt" {
		t.Fatalf("issues = %d, lang = %q", len(got.Issues), client.lastLanguage)
	}

	catalog.Type(context.Background(), 2933, "pt")
	if client.detailCalls != 1 || client.issueCalls != 1 {
		t.Fatalf("second view must come from the cache: %d detail / %d issue calls", client.detailCalls, client.issueCalls)
	}
}

func TestNumistaCatalog_MarketValueIncludesCatalogPhotos(t *testing.T) {
	_, marketValue, _, _ := newCatalogFixture()

	got := marketValue.ForType(context.Background(), 2933, 0, "1967", "EUR", "en")

	if got.Status != models.NumistaDataStatusOK || got.SelectedIssueID != 101 || len(got.CatalogPhotos) != 1 {
		t.Fatalf("got %+v", got)
	}
	if none := marketValue.ForType(context.Background(), 2933, 0, "", "EUR", ""); len(none.CatalogPhotos) != 0 {
		t.Fatalf("photos must be skipped without a language: %+v", none.CatalogPhotos)
	}
}

func TestNumistaCatalog_SearchMapsResultsAndCachesEachQuery(t *testing.T) {
	catalog, _, client, _ := newCatalogFixture()
	client.searchBody = `{"count": 347, "types": [
		{"id": 25664, "title": "Dobra 8 Escudos", "issuer": {"code": "portugal", "name": "Portugal"}, "min_year": 1724, "max_year": 1733,
		 "obverse_thumbnail": "https://en.numista.com/o.jpg", "reverse_thumbnail": "http://insecure.example/r.jpg"},
		{"id": 0, "title": "broken"}
	]}`
	ctx := context.Background()
	search := NumistaTypeSearch{Query: "escudos", Issuer: "portugal", Lang: "pt"}

	got := catalog.Search(ctx, search)

	if got.Status != models.NumistaDataStatusOK || got.Count != 347 || got.Page != 1 || len(got.Types) != 1 {
		t.Fatalf("got %+v", got)
	}
	item := got.Types[0]
	if item.Issuer.Code != "portugal" || *item.MinYear != 1724 || item.ObverseThumbnail == "" || item.ReverseThumbnail != "" {
		t.Fatalf("item = %+v", item)
	}
	if client.lastSearch.Count != NumistaCatalogPageSize || client.lastSearch.Page != 1 {
		t.Fatalf("search sent %+v", client.lastSearch)
	}

	catalog.Search(ctx, NumistaTypeSearch{Query: "  ESCUDOS ", Issuer: "portugal", Lang: "pt"})
	if client.searchCalls != 1 {
		t.Fatalf("same search must be cached, got %d calls", client.searchCalls)
	}
	catalog.Search(ctx, NumistaTypeSearch{Query: "escudos", Issuer: "portugal", Lang: "pt", Page: 2})
	if client.searchCalls != 2 {
		t.Fatalf("another page must be fetched, got %d calls", client.searchCalls)
	}

	client.searchErr = &NumistaError{Kind: NumistaErrorQuotaLimited}
	if failed := catalog.Search(ctx, NumistaTypeSearch{Query: "reis", Lang: "pt"}); failed.Status != models.NumistaDataStatusQuotaLimited || len(failed.Types) != 0 {
		t.Fatalf("quota error: %+v", failed)
	}
}

func TestNumistaCatalog_IssuersRankMatchesAndStoreATrimmedList(t *testing.T) {
	catalog, _, client, store := newCatalogFixture()
	client.issuersBody = `{"count": 5, "issuers": [
		{"code": "portuguese_india", "name": "Índia Portuguesa", "flag": "https://x/flag.gif", "level": 2},
		{"code": "portugal_section", "name": "Portugal", "level": 1},
		{"code": "portugal", "name": "Portugal", "flag": "https://x/flag.gif", "parent": {"code": "portugal_section", "name": "Portugal"}, "level": 2},
		{"code": "sao_tome", "name": "São Tomé e Príncipe", "level": 2},
		{"code": "angola", "name": "Angola", "parent": {"code": "portugal_section", "name": "Portugal"}, "level": 2},
		{"code": "", "name": "No code"}
	]}`
	ctx := context.Background()

	if empty := catalog.Issuers(ctx, "pt", "  "); len(empty.Issuers) != 0 || client.issuerCalls != 0 {
		t.Fatalf("empty query must not call Numista: %+v", empty)
	}

	got := catalog.Issuers(ctx, "pt", "port")
	names := []string{}
	for _, issuer := range got.Issuers {
		names = append(names, issuer.Name)
	}
	if strings.Join(names, "|") != "Portugal|Índia Portuguesa" {
		t.Fatalf("ranking = %v", names)
	}
	if got.Issuers[0].Code != "portugal" || got.Issuers[0].Parent != "" {
		t.Fatalf("a same-name section is dropped in favour of the issuer: %+v", got.Issuers[0])
	}

	if accent := catalog.Issuers(ctx, "pt", "sao tome"); len(accent.Issuers) != 1 || accent.Issuers[0].Code != "sao_tome" {
		t.Fatalf("accent-insensitive match failed: %+v", accent)
	}
	if child := catalog.Issuers(ctx, "pt", "angola"); child.Issuers[0].Parent != "Portugal" {
		t.Fatalf("parent kept for context: %+v", child.Issuers)
	}
	if client.issuerCalls != 1 {
		t.Fatalf("issuer list must be fetched once, got %d", client.issuerCalls)
	}
	if stored := store.entries["issuers:pt"].Payload; strings.Contains(stored, "flag") || strings.Contains(stored, "No code") {
		t.Fatalf("stored issuer list was not trimmed: %s", stored)
	}
}

func TestFoldForSearch(t *testing.T) {
	if got := foldForSearch("  São   Tomé e PRÍNCIPE "); got != "sao tome e principe" {
		t.Fatalf("foldForSearch = %q", got)
	}
}
