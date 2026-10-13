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
	// A coin search keeps its cache entry; banknotes are another search.
	catalog.Search(ctx, NumistaTypeSearch{Query: "escudos", Issuer: "portugal", Lang: "pt", Category: "coin"})
	if client.searchCalls != 2 {
		t.Fatalf("a coin search is the same search, got %d calls", client.searchCalls)
	}
	catalog.Search(ctx, NumistaTypeSearch{Query: "escudos", Issuer: "portugal", Lang: "pt", Category: "banknote"})
	if client.searchCalls != 3 || client.lastSearch.Category != "banknote" {
		t.Fatalf("banknotes must be searched apart, got %d calls (%+v)", client.searchCalls, client.lastSearch)
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

// Numista's photos cannot be loaded, so a €2 commemorative's search result and
// type details show the ECB's official image instead.
func TestNumistaCatalog_ShowsTheECBImageOfACommemorative(t *testing.T) {
	catalog, _, client, _ := newCatalogFixture()
	catalog.WithECB(newTestEcbService(t, &fakeECBCoinsClient{page: ecb2022Page(t)}))
	client.searchBody = `{"count": 2, "types": [
		{"id": 1, "title": "2 Euros (Bundesländer - \"Thüringen\")", "issuer": {"code": "allemagne", "name": "Alemanha, República Federal da"}, "min_year": 2022, "max_year": 2022, "obverse_thumbnail": "https://en.numista.com/photos/t-180.jpg"},
		{"id": 2, "title": "2 Euros (2nd map)", "issuer": {"code": "allemagne", "name": "Alemanha, República Federal da"}, "min_year": 2008, "max_year": 2026, "obverse_thumbnail": "https://en.numista.com/photos/m-180.jpg"}
	]}`
	got := catalog.Search(context.Background(), NumistaTypeSearch{Issuer: "allemagne", Year: 2022, Lang: "pt"})
	const img = "https://www.ecb.europa.eu/euro/coins/common/shared/img/"
	if len(got.Types) != 2 || got.Types[0].ObverseThumbnail != "https://www.ecb.europa.eu/euro/coins/comm/html/comm_2022/DE-thueringen.jpg" ||
		got.Types[0].ReverseThumbnail != img+"newcommon_2euro_800.jpg" ||
		got.Types[1].ObverseThumbnail != img+"de/Germany_2euro.jpg" || got.Types[1].ReverseThumbnail != img+"newcommon_2euro_800.jpg" {
		t.Fatalf("types = %+v", got.Types)
	}

	client.detailBody = `{"id": 7, "url": "https://en.numista.com/7", "title": "2 Euros (Bundesländer - \"Thüringen\")", "category": "coin",
		"issuer": {"code": "allemagne", "name": "Alemanha"}, "min_year": 2022, "max_year": 2022,
		"obverse": {"picture": "https://en.numista.com/photos/o.jpg", "thumbnail": "https://en.numista.com/photos/o-180.jpg"}}`
	detail := catalog.Type(context.Background(), 7, "pt")
	if detail.Type == nil || len(detail.Type.Photos) != 2 || detail.Type.Photos[0].Picture != "https://www.ecb.europa.eu/euro/coins/comm/html/comm_2022/DE-thueringen.jpg" ||
		detail.Type.Photos[0].Copyright != "European Central Bank" || detail.Type.Photos[1].Side != "reverse" || detail.Type.Photos[1].Picture != img+"newcommon_2euro_800.jpg" {
		t.Fatalf("type photos = %+v", detail.Type)
	}

	// Coins that are not euro coins keep Numista's photos.
	client.searchBody = `{"count": 1, "types": [
		{"id": 3, "title": "50 Escudos", "issuer": {"code": "portugal", "name": "Portugal"}, "min_year": 1968, "max_year": 1968, "obverse_thumbnail": "https://en.numista.com/photos/e-180.jpg"}
	]}`
	got = catalog.Search(context.Background(), NumistaTypeSearch{Query: "escudos", Lang: "pt"})
	if len(got.Types) != 1 || got.Types[0].ObverseThumbnail != "https://en.numista.com/photos/e-180.jpg" || got.Types[0].ReverseThumbnail != "" {
		t.Fatalf("types = %+v", got.Types)
	}
}

func TestNumistaCatalog_ShowsTheINCMPhotosOfAPortugueseCollectorCoin(t *testing.T) {
	catalog, _, client, _ := newCatalogFixture()
	incm, _ := newTestIncmService(t)
	catalog.WithECB(newTestEcbService(t, &fakeECBCoinsClient{page: ecb2022Page(t)})).WithMint(incm)
	client.searchBody = `{"count": 2, "types": [
		{"id": 1, "title": "5 Euro (Ulysses)", "issuer": {"code": "portugal", "name": "Portugal"}, "min_year": 2024, "max_year": 2024, "obverse_thumbnail": "https://en.numista.com/photos/u-180.jpg"},
		{"id": 2, "title": "5 Euro (The knowledge, cupronickel)", "issuer": {"code": "portugal", "name": "Portugal"}, "min_year": 2024, "max_year": 2024, "obverse_thumbnail": "https://en.numista.com/photos/k-180.jpg"}
	]}`
	got := catalog.Search(context.Background(), NumistaTypeSearch{Issuer: "portugal", Year: 2024, Lang: "pt"})
	if len(got.Types) != 2 || !strings.Contains(got.Types[0].ObverseThumbnail, incmFiles+"files/Ulises_CU_A.jpg") || !strings.HasSuffix(got.Types[0].ObverseThumbnail, "&width=400") ||
		!strings.Contains(got.Types[0].ReverseThumbnail, "Ulises_CU_B.jpg") {
		t.Fatalf("types = %+v", got.Types)
	}
	// The shop's English title matches Numista's.
	if !strings.Contains(got.Types[1].ObverseThumbnail, "Conhecimento_CU_A.jpg") {
		t.Fatalf("types = %+v", got.Types)
	}

	client.detailBody = `{"id": 2, "url": "https://en.numista.com/2", "title": "5 Euro (The knowledge, cupronickel)", "category": "coin",
		"issuer": {"code": "portugal", "name": "Portugal"}, "min_year": 2024, "max_year": 2024,
		"value": {"text": "5 Euros", "numeric_value": 5, "currency": {"name": "Euro"}}, "composition": {"text": "Cuproníquel"},
		"series": "Desenhe uma moeda", "commemorated_topic": "O conhecimento",
		"obverse": {"lettering": "PORTUGAL 5 €", "designers": ["Antonio Cerqueira"], "thumbnail": "https://en.numista.com/photos/o-180.jpg"},
		"reverse": {"description": "Um frasco e as palavras Educação e Tecnológica", "designers": ["Antonio Cerqueira"]}}`
	detail := catalog.Type(context.Background(), 2, "pt")
	if detail.Type == nil || detail.Type.CommemoratedTopic != "O conhecimento" || len(detail.Type.Photos) != 2 ||
		!strings.Contains(detail.Type.Photos[0].Picture, "Conhecimento_CU_A.jpg") || detail.Type.Photos[0].Copyright != "INCM" ||
		detail.Type.Photos[1].Side != "reverse" || !strings.Contains(detail.Type.Photos[1].Thumbnail, "Conhecimento_CU_B.jpg") {
		t.Fatalf("type = %+v", detail.Type)
	}
}

// A banknote type gives its height, printers, watermark and issuing bank.
func TestParseNumistaCatalogType_Banknote(t *testing.T) {
	detail, err := parseNumistaCatalogType([]byte(`{"id": 1234, "title": "1000 Escudos (Pedro Álvares Cabral)", "category": "banknote",
		"issuer": {"code": "portugal", "name": "Portugal"}, "issuing_entity": {"id": 9, "name": "Banco de Portugal"},
		"min_year": 1996, "max_year": 2000, "value": {"text": "1000 Escudos"}, "size": 160, "size2": 75,
		"composition": {"text": "Papel"}, "printers": [{"id": 1, "name": "Valora"}, {"name": " "}],
		"watermark": {"description": "Pedro Álvares Cabral"},
		"references": [{"catalogue": {"code": "P"}, "number": "188"}]}`))
	if err != nil {
		t.Fatal(err)
	}
	if detail.Category != "banknote" || detail.IssuingEntity != "Banco de Portugal" || detail.DiameterMm == nil || *detail.DiameterMm != 160 ||
		detail.HeightMm == nil || *detail.HeightMm != 75 || len(detail.Printers) != 1 || detail.Printers[0] != "Valora" ||
		detail.Watermark != "Pedro Álvares Cabral" || len(detail.References) != 1 || detail.References[0].Catalogue != "P" {
		t.Fatalf("detail = %+v", detail)
	}
}
