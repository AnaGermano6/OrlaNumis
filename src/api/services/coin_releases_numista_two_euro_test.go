package services

import (
	"context"
	"errors"
	"sort"
	"testing"
	"time"

	"github.com/briandenicola/ancient-coins-api/models"
)

func TestNumistaTwoEuroRelease(t *testing.T) {
	year := func(value int) *int { return &value }
	for _, c := range []struct {
		title    string
		min, max int
		name     string
	}{
		{"2 Euros (450th Anniversary of the Death of Titian)", 2026, 2026, "450th Anniversary of the Death of Titian"},
		{"2 Euro (Konrad Adenauer)", 2026, 2026, "Konrad Adenauer"},
		// A country's usual €2 coin, struck since before; another face value; a pattern.
		{"2 Euros (2nd map)", 2017, 2026, ""},
		{"2 Euro Cents", 2026, 2026, ""},
		{"2 Euros (Trial strike)", 2026, 2026, ""},
		{"5 Euros (Peregrine Falcon)", 2026, 2026, ""},
	} {
		coin, ok := numistaTwoEuroRelease(models.NumistaCatalogSearchItem{ID: 42, Title: c.title, MinYear: year(c.min), MaxYear: year(c.max)}, "SM", 2026)
		if ok != (c.name != "") || ok && coin.Title != c.name {
			t.Errorf("numistaTwoEuroRelease(%q) = %+v, %v", c.title, coin, ok)
		}
		if ok && (coin.Kind != CoinReleaseKindCommemorative || coin.Source != CoinReleaseSourceNumista || coin.FaceValue != 2 ||
			coin.From != "2026-01" || coin.To != "2026-12" || coin.ImageURL != "" || coin.Keys[0] != "numista:42" ||
			coin.PageURL != "https://pt.numista.com/catalogue/pieces42.html" || coin.Country != "SM") {
			t.Errorf("numistaTwoEuroRelease(%q) = %+v", c.title, coin)
		}
	}
}

func TestNumistaCollectorRelease(t *testing.T) {
	year := func(value int) *int { return &value }
	for _, c := range []struct {
		title string
		min   int
		name  string
		value float64
	}{
		{"5 Euros (Peregrine Falcon)", 2026, "Peregrine Falcon", 5},
		{"3 Euros (Visit of Pope Leo XIV)", 2026, "Visit of Pope Leo XIV", 3},
		{"0.75 Euro (Canticle of the Creatures)", 2026, "Canticle of the Creatures", 0.75},
		{"100 Euros (Saint Marinus)", 2026, "Saint Marinus", 100},
		// For circulation (a country's first euro coins too), struck before, a pattern, not euros.
		{"1 Euro", 2026, "", 0},
		{"2 Euros (450th Anniversary of the Death of Titian)", 2026, "", 0},
		{"50 Euro Cent", 2026, "", 0},
		{"10 Euros (Ships)", 2025, "", 0},
		{"5 Euros (Pattern)", 2026, "", 0},
		{"1 Lira", 2026, "", 0},
	} {
		coin, ok := numistaCollectorRelease(models.NumistaCatalogSearchItem{ID: 77, Title: c.title, MinYear: year(c.min), MaxYear: year(c.min)}, "SM", 2026)
		if ok != (c.name != "") || ok && (coin.Title != c.name || coin.FaceValue != c.value || coin.Kind != CoinReleaseKindCollector ||
			coin.Source != CoinReleaseSourceNumista || coin.From != "2026-01" || coin.Keys[0] != "numista:77" || coin.ImageURL != "") {
			t.Errorf("numistaCollectorRelease(%q) = %+v, %v", c.title, coin, ok)
		}
	}
}

// yearNumista answers the searches of an issuer's coins of the year.
type yearNumista struct {
	escudoNumista
	year     int
	byIssuer map[string]string
	searched []string
	failAt   string
}

func (n *yearNumista) SearchTypes(ctx context.Context, search NumistaTypeSearch) ([]byte, error) {
	if search.Year == n.year && search.Query == "" {
		n.searched = append(n.searched, search.Issuer)
		if search.Issuer == n.failAt {
			return nil, errors.New("quota used up")
		}
		if body, ok := n.byIssuer[search.Issuer]; ok {
			return []byte(body), nil
		}
		return []byte(`{"count": 0, "types": []}`), nil
	}
	return n.escudoNumista.SearchTypes(ctx, search)
}

func newYearReleases(t *testing.T, client *yearNumista, mints ...MintReleaseSource) *CoinReleasesService {
	t.Helper()
	service := newEscudoCatalogue(t, &client.escudoNumista, mints...)
	service.numista = NewNumistaCatalogService(client, service.numista.cache)
	service.now = func() time.Time { return time.Date(client.year, 11, 19, 9, 0, 0, 0, time.UTC) }
	return service
}

func groupTitles(group CoinReleaseGroup) []string {
	list := []string{}
	for _, release := range group.Releases {
		list = append(list, release.Country+" "+release.Title)
	}
	sort.Strings(list)
	return list
}

func TestReleases_TheYearsCoinsFromNumista(t *testing.T) {
	client := &yearNumista{year: 2026, byIssuer: map[string]string{
		"saint-marin": `{"count": 5, "types": [
			{"id": 501, "title": "2 Euros (450th Anniversary of the Death of Titian)", "min_year": 2026, "max_year": 2026},
			{"id": 502, "title": "2 Euros (800th Anniversary of the Death of Saint Francis of Assisi)", "min_year": 2026, "max_year": 2026},
			{"id": 503, "title": "2 Euros (Three towers)", "min_year": 2017, "max_year": 2026},
			{"id": 504, "title": "5 Euros (Peregrine Falcon)", "min_year": 2026, "max_year": 2026},
			{"id": 505, "title": "3 Euros (Visit of Pope Leo XIV)", "min_year": 2026, "max_year": 2026}
		]}`,
		// Portugal's mint sells its coins of the year: Numista's are not listed again.
		"portugal": `{"count": 2, "types": [
			{"id": 601, "title": "2 Euros (Rotary)", "min_year": 2026, "max_year": 2026},
			{"id": 602, "title": "10 Euros (Amadeo de Souza-Cardoso)", "min_year": 2026, "max_year": 2026}
		]}`,
	}}
	portugal := fakeMint{country: "Portugal", source: "incm", coins: []mintCoin{
		{Title: "Rotary", Value: 2, Metal: "Cuproníquel", Year: 2026, Released: "2026-07", PageURL: "https://loja.incm.pt/rotary"},
		{Title: "Amadeo de Souza-Cardoso", Value: 10, Metal: "Prata", Year: 2026, Released: "2026-09", PageURL: "https://loja.incm.pt/amadeo"},
	}}
	service := newYearReleases(t, client, portugal)
	ctx := context.Background()

	// Nobody waits for Numista: the first lists are without its coins.
	first := service.Releases(ctx)
	if got := groupTitles(first[0]); len(got) != 1 || got[0] != "PT Rotary" {
		t.Fatalf("first €2 coins = %v", got)
	}
	if got := groupTitles(first[1]); len(got) != 1 || got[0] != "PT Amadeo de Souza-Cardoso" {
		t.Fatalf("first collector coins = %v", got)
	}
	service.yearCoinsLoads.Wait()
	groups := service.Releases(ctx)
	if got := groupTitles(groups[0]); len(got) != 3 || got[0] != "PT Rotary" || got[1] != "SM 450th Anniversary of the Death of Titian" || got[2] != "SM 800th Anniversary of the Death of Saint Francis of Assisi" {
		t.Fatalf("€2 coins = %v", got)
	}
	if got := groupTitles(groups[1]); len(got) != 3 || got[0] != "PT Amadeo de Souza-Cardoso" || got[1] != "SM Peregrine Falcon" || got[2] != "SM Visit of Pope Leo XIV" {
		t.Fatalf("collector coins = %v", got)
	}
	// Every euro country asked, once.
	if len(client.searched) != len(numistaEuroIssuers) {
		t.Fatalf("searched %d issuers: %v", len(client.searched), client.searched)
	}

	// The catalogue has the €2 coins too.
	catalogue := map[string]bool{}
	for _, coin := range service.Catalogue(ctx) {
		catalogue[coin.Country+" "+coin.Title] = true
	}
	if !catalogue["SM 450th Anniversary of the Death of Titian"] || !catalogue["SM 800th Anniversary of the Death of Saint Francis of Assisi"] {
		t.Fatalf("catalogue = %v", catalogue)
	}

	// Read again a day later, not before.
	service.Releases(ctx)
	service.yearCoinsLoads.Wait()
	if len(client.searched) != len(numistaEuroIssuers) {
		t.Fatalf("asked again within the day: %d", len(client.searched))
	}
}

func TestReleases_TheYearsCoinsPastACountryNumistaDidNotAnswerFor(t *testing.T) {
	client := &yearNumista{year: 2026, failAt: "germany", byIssuer: map[string]string{
		"germany":     `{"count": 1, "types": [{"id": 701, "title": "2 Euros (Konrad Adenauer)", "min_year": 2026, "max_year": 2026}]}`,
		"saint-marin": `{"count": 1, "types": [{"id": 504, "title": "5 Euros (Peregrine Falcon)", "min_year": 2026, "max_year": 2026}]}`,
	}}
	at := time.Date(2026, 11, 19, 9, 0, 0, 0, time.UTC)
	service := newYearReleases(t, client)
	service.now = func() time.Time { return at }

	service.Releases(context.Background())
	service.yearCoinsLoads.Wait()
	// Germany's search failed (too slow): the other countries were asked all the same.
	groups := service.Releases(context.Background())
	if got := groupTitles(groups[1]); len(got) != 1 || got[0] != "SM Peregrine Falcon" {
		t.Fatalf("collector coins = %v", got)
	}
	if len(client.searched) != len(numistaEuroIssuers) {
		t.Fatalf("searched %d issuers", len(client.searched))
	}

	// Asked again a quarter of an hour later, not before.
	at = at.Add(10 * time.Minute)
	service.Releases(context.Background())
	service.yearCoinsLoads.Wait()
	if len(client.searched) != len(numistaEuroIssuers) {
		t.Fatalf("asked again too soon: %d", len(client.searched))
	}
	client.failAt = ""
	at = at.Add(10 * time.Minute)
	service.Releases(context.Background())
	service.yearCoinsLoads.Wait()
	groups = service.Releases(context.Background())
	if got := groupTitles(groups[0]); len(got) != 1 || got[0] != "DE Konrad Adenauer" {
		t.Fatalf("€2 coins = %v", got)
	}
}

// quotaNumista refuses every search from one issuer on, as Numista does once its quota is used up.
type quotaNumista struct {
	yearNumista
	refusing bool
}

func (n *quotaNumista) SearchTypes(ctx context.Context, search NumistaTypeSearch) ([]byte, error) {
	if search.Year == n.year && search.Query == "" && (n.refusing || search.Issuer == n.failAt) {
		n.refusing = true
		n.searched = append(n.searched, search.Issuer)
		return nil, &NumistaError{Kind: NumistaErrorQuotaLimited}
	}
	return n.yearNumista.SearchTypes(ctx, search)
}

func TestReleases_TheYearsCoinsWhenNumistaRefuses(t *testing.T) {
	client := &quotaNumista{yearNumista: yearNumista{year: 2026, failAt: "saint-marin", byIssuer: map[string]string{
		"germany": `{"count": 1, "types": [{"id": 701, "title": "2 Euros (Konrad Adenauer)", "min_year": 2026, "max_year": 2026}]}`,
	}}}
	service := newEscudoCatalogue(t, &client.escudoNumista)
	service.numista = NewNumistaCatalogService(client, service.numista.cache)
	service.now = func() time.Time { return time.Date(2026, 11, 19, 9, 0, 0, 0, time.UTC) }

	service.Releases(context.Background())
	service.yearCoinsLoads.Wait()
	releases := service.Releases(context.Background())[0].Releases
	if len(releases) != 1 || releases[0].Title != "Konrad Adenauer" || releases[0].Country != "DE" {
		t.Fatalf("releases = %+v", releases)
	}
	// Refused at San Marino: the countries after it were not asked.
	if last := client.searched[len(client.searched)-1]; last != "saint-marin" {
		t.Fatalf("searched = %v", client.searched)
	}
}
