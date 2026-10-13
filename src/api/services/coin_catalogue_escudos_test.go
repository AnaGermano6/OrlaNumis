package services

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/briandenicola/ancient-coins-api/models"
	"github.com/briandenicola/ancient-coins-api/repository"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func TestEscudoRelease(t *testing.T) {
	year := func(value int) *int { return &value }
	for _, c := range []struct {
		title, classification string
		min, max              int
		kind, name            string
		value                 float64
	}{
		{"200 Escudos (Colombo e Portugal)", "", 1991, 1991, CoinReleaseKindCommemorative, "Colombo e Portugal", 200},
		{"100 Escudos (Fernando Pessoa; Prata)", "", 1985, 1985, CoinReleaseKindCollector, "Fernando Pessoa; Prata", 100},
		{"1000 Escudos (Expo 98)", "Moedas não circulantes", 1998, 1998, CoinReleaseKindCollector, "Expo 98", 1000},
		{"2½ Escudos (Alexandre Herculano)", "", 1977, 1977, CoinReleaseKindCommemorative, "Alexandre Herculano", 2.5},
		{"2.50 Escudos (Alexandre Herculano)", "Circulating commemorative coins", 1977, 1977, CoinReleaseKindCommemorative, "Alexandre Herculano", 2.5},
		// Not commemorative or collector coins of the escudo:
		{"100 Escudos", "", 1989, 2001, "", "", 0},
		{"100 Escudos (small type)", "", 1989, 1990, "", "", 0},
		{"10 Escudos (Pattern)", "", 1990, 1990, "", "", 0},
		{"50 Centavos (Republic)", "", 1927, 1927, "", "", 0},
		{"25 Escudos (Revolução)", "Moedas circulantes comuns", 1974, 1974, "", "", 0},
		{"2 Euros (Rotary)", "", 2026, 2026, "", "", 0},
	} {
		coin, ok := escudoRelease(7, c.title, year(c.min), year(c.max), c.classification)
		if ok != (c.kind != "") || ok && (coin.Kind != c.kind || coin.Title != c.name || coin.FaceValue != c.value) {
			t.Errorf("escudoRelease(%q, %q) = %+v, %v", c.title, c.classification, coin, ok)
		}
		if ok && (coin.Currency != "PTE" || coin.Country != "PT" || coin.From != fmt.Sprint(c.min)+"-01" || coin.Keys[0] != "numista:7" || coin.PageURL != "https://pt.numista.com/catalogue/pieces7.html") {
			t.Errorf("escudoRelease(%q) = %+v", c.title, coin)
		}
	}
}

// escudoNumista answers searches of Portugal's types by year, until a year
// it is out of quota for, and searches of an issuer's euro face value.
type escudoNumista struct {
	years    map[int]string
	values   map[string]string
	failFrom int
	calls    int
}

func (n *escudoNumista) SearchTypes(_ context.Context, search NumistaTypeSearch) ([]byte, error) {
	n.calls++
	if n.failFrom != 0 && search.Year >= n.failFrom {
		return nil, errors.New("quota used up")
	}
	if body, ok := n.years[search.Year]; ok && search.Year != 0 {
		return []byte(body), nil
	}
	if body, ok := n.values[search.Issuer+"|"+search.Query]; ok {
		return []byte(body), nil
	}
	return []byte(`{"count": 0, "types": []}`), nil
}

func (n *escudoNumista) TypeDetail(context.Context, int, string) ([]byte, error) {
	return nil, errors.New("not asked")
}
func (n *escudoNumista) Issuers(context.Context, string) ([]byte, error) { return []byte(`[]`), nil }
func (n *escudoNumista) Issues(context.Context, int) ([]byte, error) {
	return nil, errors.New("not asked")
}

func newEscudoCatalogue(t *testing.T, client *escudoNumista, mints ...MintReleaseSource) *CoinReleasesService {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&models.NumistaCacheEntry{}); err != nil {
		t.Fatal(err)
	}
	// Types described before: the last escudo (the collection's) and the Expo 98 coin.
	for key, payload := range map[string]string{
		"type:9001:pt": `{"id": 9001, "title": "1 Escudo (O Último Escudo)", "category": "coin", "issuer": {"code": "portugal", "name": "Portugal"}, "min_year": 2001, "max_year": 2001, "type": "Moedas não circulantes"}`,
		"type:9002:pt": `{"id": 9002, "title": "1000 Escudos (Expo 98)", "category": "coin", "issuer": {"code": "portugal", "name": "Portugal"}, "min_year": 1998, "max_year": 1998, "type": "Moedas não circulantes"}`,
		"type:9003:pt": `{"id": 9003, "title": "2.50 Escudos", "category": "coin", "issuer": {"code": "angola", "name": "Angola"}, "min_year": 1953, "max_year": 1953, "type": "Moedas circulantes comemorativas"}`,
	} {
		if err := db.Create(&models.NumistaCacheEntry{Key: key, Payload: payload, FetchedAt: time.Now()}).Error; err != nil {
			t.Fatal(err)
		}
	}
	cache := NewNumistaResponseCache(repository.NewNumistaCacheRepository(db), nil, nil)
	return NewCoinReleasesService(nil, mints...).WithNumista(NewNumistaCatalogService(client, cache))
}

func escudoIDs(coins []CoinRelease) map[string]string {
	ids := map[string]string{}
	for _, coin := range coins {
		if coin.Source == CoinReleaseSourceNumista {
			ids[coin.Keys[0]] = coin.Kind
		}
	}
	return ids
}

func TestCoinCatalogue_PortugalEscudosFromNumista(t *testing.T) {
	client := &escudoNumista{years: map[int]string{
		1991: `{"count": 2, "types": [
			{"id": 100, "title": "200 Escudos (Colombo e Portugal)", "issuer": {"code": "portugal"}, "min_year": 1991, "max_year": 1991},
			{"id": 101, "title": "100 Escudos", "issuer": {"code": "portugal"}, "min_year": 1989, "max_year": 2001}]}`,
		1985: `{"count": 1, "types": [{"id": 102, "title": "100 Escudos (Fernando Pessoa; Prata)", "issuer": {"code": "portugal"}, "min_year": 1985, "max_year": 1985}]}`,
		1998: `{"count": 1, "types": [{"id": 9002, "title": "1000 Escudos (Expo 98)", "issuer": {"code": "portugal"}, "min_year": 1998, "max_year": 1998}]}`,
		1990: `{"count": 1, "types": [{"id": 103, "title": "10 Escudos (Pattern)", "issuer": {"code": "portugal"}, "min_year": 1990, "max_year": 1990}]}`,
	}}
	catalogue := newEscudoCatalogue(t, client)

	coins := catalogue.Catalogue(context.Background())
	// Every year asked once; the commemorative and collector coins kept,
	// Numista's own word on the Expo 98 coin followed.
	want := map[string]string{"numista:100": "commemorative", "numista:102": "collector", "numista:9002": "collector", "numista:9001": "collector"}
	// Every year of escudos, then each euro country's three face values.
	searches := escudoLastYear - escudoFirstYear + 1 + len(numistaEuroIssuers)*len(catalogueEuroValues)
	if got := escudoIDs(coins); fmt.Sprint(got) != fmt.Sprint(want) || client.calls != searches || !catalogue.CatalogueComplete() {
		t.Fatalf("escudos = %v after %d searches, complete %v", got, client.calls, catalogue.CatalogueComplete())
	}

	// A coin of the collection linked to a Numista type is that coin.
	owned := catalogue.OwnedKeys(context.Background(), []CoinPhotoQuery{{NumistaTypeID: 100}}, nil)
	MarkOwnedReleases(coins, owned)
	for _, coin := range coins {
		if coin.Owned != (coin.Keys[0] == "numista:100") {
			t.Fatalf("owned = %+v", coin)
		}
	}
}

func TestCoinCatalogue_EscudosWhileNumistaIsOutOfQuota(t *testing.T) {
	client := &escudoNumista{failFrom: escudoFirstYear}
	catalogue := newEscudoCatalogue(t, client)

	coins := catalogue.Catalogue(context.Background())
	// One try, not one per year; what was cached is listed meanwhile.
	want := map[string]string{"numista:9001": "collector", "numista:9002": "collector"}
	if got := escudoIDs(coins); fmt.Sprint(got) != fmt.Sprint(want) || client.calls != 1 || catalogue.CatalogueComplete() {
		t.Fatalf("escudos = %v after %d searches, complete %v", got, client.calls, catalogue.CatalogueComplete())
	}
}

func TestEuroCollectorRelease(t *testing.T) {
	year := func(value int) *int { return &value }
	for _, c := range []struct {
		title string
		min   int
		value int
		name  string
	}{
		{"10 Euros (Albert Einstein)", 2005, 10, "Albert Einstein"},
		{"10 Euro (Bauhaus Dessau)", 2004, 10, "Bauhaus Dessau"},
		{"20 Euros (Wren)", 2016, 20, "Wren"},
		{"5 Euros (Planet Earth)", 2016, 5, "Planet Earth"},
		// Not one of the value's collector coins:
		{"10 Euro Cents (2nd map)", 2007, 10, ""},
		{"1 Euro Cent", 2002, 10, ""},
		{"2 Euros (10 Years of Euro Cash)", 2012, 10, ""},
		{"10 Euros (Pattern)", 2002, 10, ""},
		{"20 Euros (Silver)", 1998, 20, ""},
		{"10 Euros (Albert Einstein)", 2005, 5, ""},
	} {
		coin, ok := euroCollectorRelease(models.NumistaCatalogSearchItem{ID: 42, Title: c.title, MinYear: year(c.min)}, "DE", c.value)
		if ok != (c.name != "") || ok && (coin.Title != c.name || coin.FaceValue != float64(c.value) || coin.Kind != CoinReleaseKindCollector || coin.Country != "DE" || coin.Currency != "" || coin.Keys[0] != "numista:42") {
			t.Errorf("euroCollectorRelease(%q, %d) = %+v, %v", c.title, c.value, coin, ok)
		}
	}
}

// Every euro country's €5, €10 and €20 coins come from Numista; one its mint
// lists already is listed once, by the mint, and known by its Numista type.
func TestCoinCatalogue_EuroCollectorCoins(t *testing.T) {
	client := &escudoNumista{values: map[string]string{
		"germany|10 euro": `{"count": 3, "types": [
			{"id": 501, "title": "10 Euro Cents (1st map)", "issuer": {"code": "germany"}, "min_year": 2002, "max_year": 2006},
			{"id": 502, "title": "10 Euros (Albert Einstein)", "issuer": {"code": "germany"}, "min_year": 2005, "max_year": 2005},
			{"id": 503, "title": "10 Euros (Franz Kafka)", "issuer": {"code": "germany"}, "min_year": 2008, "max_year": 2008}]}`,
		"france|10 euro": `{"count": 2, "types": [
			{"id": 701, "title": "10 Euros (10 euros Asterix aux jeux Olympiques)", "issuer": {"code": "france"}, "min_year": 2024, "max_year": 2024},
			{"id": 702, "title": "10 Euros (JO 2024 Paris - Eiffel Tower)", "issuer": {"code": "france"}, "min_year": 2024, "max_year": 2024}]}`,
		"portugal|5 euro": `{"count": 2, "types": [
			{"id": 601, "title": "5 Euros (Faróis de Portugal - Cabo de São Vicente)", "issuer": {"code": "portugal"}, "min_year": 2025, "max_year": 2025},
			{"id": 602, "title": "5 Euros (Mosteiro dos Jerónimos)", "issuer": {"code": "portugal"}, "min_year": 2004, "max_year": 2004}]}`,
	}}
	portugal := fakeMint{country: "Portugal", source: "incm", coins: []mintCoin{
		{Title: "Faróis de Portugal - Cabo de São Vicente (Prata Proof)", Value: 5, Year: 2025, Released: "2025-11", PageURL: "https://loja.incm.pt/products/farol", Obverse: "https://cdn/farol_A.jpg"},
	}}
	france := fakeMint{country: "France", source: "mdp", coins: []mintCoin{
		{Title: "Asterix and the Olympic Games", DesignKey: "asterix olympic games", Value: 10, Year: 2024, PageURL: "https://www.monnaiedeparis.fr/en/asterix-10", Obverse: "https://mdp/asterix-10.jpg"},
		{Title: "Asterix and the Olympic Games", DesignKey: "asterix olympic games gold", Value: 200, Year: 2024, PageURL: "https://www.monnaiedeparis.fr/en/asterix-200", Obverse: "https://mdp/asterix-200.jpg"},
		{Title: "French Excellence - Notre-Dame de Paris", Value: 10, Year: 2024, PageURL: "https://www.monnaiedeparis.fr/en/notre-dame-10", Obverse: "https://mdp/notre-dame.jpg"},
	}}
	catalogue := newEscudoCatalogue(t, client, portugal, france)

	coins := catalogue.Catalogue(context.Background())
	byKey := map[string]CoinRelease{}
	for _, coin := range coins {
		for _, key := range coin.Keys {
			byKey[key] = coin
		}
	}
	if byKey["numista:502"].Country != "DE" || byKey["numista:502"].Title != "Albert Einstein" || byKey["numista:503"].From != "2008-01" {
		t.Fatalf("German €10 coins = %+v, %+v", byKey["numista:502"], byKey["numista:503"])
	}
	if _, cent := byKey["numista:501"]; cent {
		t.Fatal("a cent coin is listed")
	}
	// The lighthouse coin is the INCM's, with its picture; the Jerónimos coin of 2004, which it no longer sells, is Numista's.
	if farol := byKey["numista:601"]; farol.Source != "incm" || farol.ImageURL == "" {
		t.Fatalf("the INCM's coin = %+v", farol)
	}
	if byKey["numista:602"].Source != CoinReleaseSourceNumista {
		t.Fatalf("an older coin = %+v", byKey["numista:602"])
	}
	// Asterix is the Monnaie de Paris's coin, though Numista names it in
	// French; the Eiffel Tower coin shares only "Paris" with Notre-Dame, and stays.
	if asterix := byKey["numista:701"]; asterix.Source != "mdp" || len(asterix.FaceValues) != 2 {
		t.Fatalf("Asterix = %+v", asterix)
	}
	if eiffel := byKey["numista:702"]; eiffel.Source != CoinReleaseSourceNumista || eiffel.Title != "JO 2024 Paris - Eiffel Tower" {
		t.Fatalf("Eiffel Tower = %+v", eiffel)
	}
	if !catalogue.CatalogueComplete() {
		t.Fatal("the catalogue is complete")
	}
}
