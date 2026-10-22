package services

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/briandenicola/ancient-coins-api/models"
	"github.com/briandenicola/ancient-coins-api/repository"
)

func TestTypeSeriesFaceValue(t *testing.T) {
	cases := map[string]float64{
		"50 Centavos": 0.5, "1 Centavo": 0.01, "20 Centavos (Ø 22 mm)": 0.2, "2.50 Escudos": 2.5, "2½ Escudos": 2.5,
		"1 Escudo": 1, "200 Escudos": 200, "1000 Réis - Carlos I": 1000, "X Réis": 0,
	}
	for title, want := range cases {
		if got := typeSeriesFaceValue(title); got != want {
			t.Errorf("typeSeriesFaceValue(%q) = %v, want %v", title, got, want)
		}
	}
}

func TestSeriesType(t *testing.T) {
	escudos := portugueseTypeSeriesByChallenge[ChallengeEscudoTypes]
	reis := portugueseTypeSeriesByChallenge[ChallengeReisTypes]
	item := func(title string, from, to int) models.NumistaCatalogSearchItem {
		return models.NumistaCatalogSearchItem{ID: 1, Title: title, MinYear: &from, MaxYear: &to}
	}
	cases := []struct {
		series         portugueseTypeSeries
		item           models.NumistaCatalogSearchItem
		classification string
		want           bool
	}{
		{escudos, item("50 Centavos", 1927, 1968), "", true},
		{escudos, item("2 Centavos", 1918, 1918), "", true},
		{escudos, item("20 Centavos (Ø 22 mm)", 1920, 1922), "", true},
		{escudos, item("1 Escudo", 1986, 2001), "", true},
		{escudos, item("50 Centavos (Pattern)", 1927, 1927), "", false},
		{escudos, item("20 Escudos (Infante D. Henrique- Prova)", 1960, 1960), "", false},
		// A coin of one year that names what it marks is a commemorative.
		{escudos, item("200 Escudos (Colombo e Portugal)", 1991, 1991), "", false},
		// Numista's classification, when cached, decides.
		{escudos, item("100 Escudos (Gil Eanes)", 1986, 1987), "Comemorativas circulantes", false},
		{escudos, item("10 Escudos", 1971, 1974), "Moedas circulantes comuns", true},
		{escudos, item("1000 Réis - Carlos I", 1899, 1899), "", false},
		{reis, item("1000 Réis - Carlos I", 1899, 1899), "", true},
		{reis, item("10 Réis - Maria II", 1837, 1853), "", true},
		{reis, item("500 Réis - Carlos I (Discovery of India)", 1898, 1898), "", false},
		{reis, item("1 Escudo", 1915, 1916), "", false},
	}
	for _, tc := range cases {
		if got := seriesType(tc.series, tc.item, tc.classification); got != tc.want {
			t.Errorf("seriesType(%q, %q) = %v, want %v", tc.item.Title, tc.classification, got, tc.want)
		}
	}
}

func TestParseChecklistDefinition_TypeSets(t *testing.T) {
	for _, challenge := range []string{ChallengeEscudoTypes, ChallengeReisTypes} {
		definition, err := ParseChecklistDefinition(map[string]interface{}{"source": "challenge", "challenge": challenge, "country": "PT", "year": 1990})
		if err != nil || definition.Challenge != challenge || definition.Country != "" || definition.Year != 0 {
			t.Fatalf("%s: %+v (%v)", challenge, definition, err)
		}
	}
}

func TestTypeSeriesSlots_FromNumistaYears(t *testing.T) {
	db := setupTestDB(t)
	catalog, _, client, store := newCatalogFixture()
	fetched := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	cacheYear := func(year int, body string) {
		key := fmt.Sprintf("issuer-year:v1:portugal:%d:pt:1", year)
		store.entries[key] = models.NumistaCacheEntry{Key: key, Payload: body, FetchedAt: fetched}
	}
	// Every year of the escudos is cached except 1920, which Numista cannot answer now.
	for year := escudoFirstYear; year <= escudoLastYear; year++ {
		cacheYear(year, `{"count":0,"types":[]}`)
	}
	cacheYear(1927, `{"count":3,"types":[
		{"id":501,"title":"50 Centavos","min_year":1927,"max_year":1968,"obverse_thumbnail":"https://en.numista.com/catalogue/photos/portugal/501-180.jpg"},
		{"id":502,"title":"1 Escudo","min_year":1927,"max_year":1968},
		{"id":503,"title":"50 Centavos (Pattern)","min_year":1927,"max_year":1927}]}`)
	cacheYear(1968, `{"count":2,"types":[
		{"id":501,"title":"50 Centavos","min_year":1927,"max_year":1968},
		{"id":504,"title":"50 Escudos (Pedro Alvares Cabral)","min_year":1968,"max_year":1968}]}`)
	cacheYear(1917, `{"count":1,"types":[{"id":505,"title":"1 Centavo","min_year":1917,"max_year":1922}]}`)
	delete(store.entries, "issuer-year:v1:portugal:1920:pt:1")
	client.searchErr = &NumistaError{Kind: NumistaErrorQuotaLimited}

	coin := &models.Coin{Name: "1 Escudo 1950", UserID: 1, Category: models.CategoryModern}
	if err := db.Create(coin).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&models.CoinReference{CoinID: coin.ID, Catalog: "Numista", Number: "502"}).Error; err != nil {
		t.Fatal(err)
	}
	svc := NewSetChecklistService(nil, repository.NewCoinRepository(db), nil).WithNumista(catalog)

	slots, complete, err := svc.SlotsAndCompleteness(context.Background(), 1, ChecklistDefinition{Source: ChecklistSourceChallenge, Challenge: ChallengeEscudoTypes})
	if err != nil {
		t.Fatal(err)
	}
	if complete {
		t.Fatal("a year Numista did not answer leaves the list incomplete")
	}
	if client.searchCalls != 1 {
		t.Fatalf("only the year not cached is asked for: %d searches", client.searchCalls)
	}
	titles := []string{}
	for _, slot := range slots {
		titles = append(titles, slot.Title)
	}
	// Oldest first, without the pattern or the commemorative, each type once.
	if fmt.Sprint(titles) != "[1 Centavo 50 Centavos 1 Escudo]" {
		t.Fatalf("slots = %v", titles)
	}
	cents := slots[1]
	if cents.Key != "type:501" || cents.Year != 1927 || cents.ToYear != 1968 || cents.FaceValue != 0.5 || cents.Currency != "PTE" ||
		cents.ImageURL != "https://en.numista.com/catalogue/photos/portugal/501-180.jpg" || cents.PageURL != "https://pt.numista.com/catalogue/pieces501.html" || cents.Owned {
		t.Fatalf("50 centavos slot = %+v", cents)
	}
	if escudo := slots[2]; !escudo.Owned || escudo.CoinID != coin.ID || escudo.CoinName != "1 Escudo 1950" {
		t.Fatalf("the collection's coin of type 502 fills its slot: %+v", escudo)
	}

	// Once Numista answers, the list is complete.
	client.searchErr = nil
	client.searchBody = `{"count":0,"types":[]}`
	if _, complete, err := svc.SlotsAndCompleteness(context.Background(), 1, ChecklistDefinition{Source: ChecklistSourceChallenge, Challenge: ChallengeEscudoTypes}); err != nil || !complete {
		t.Fatalf("complete = %v (%v)", complete, err)
	}
}

func TestTypeSeriesSlots_AsksFewYearsPerView(t *testing.T) {
	db := setupTestDB(t)
	catalog, _, client, _ := newCatalogFixture()
	client.searchBody = `{"count":0,"types":[]}`
	svc := NewSetChecklistService(nil, repository.NewCoinRepository(db), nil).WithNumista(catalog)

	_, complete, err := svc.SlotsAndCompleteness(context.Background(), 1, ChecklistDefinition{Source: ChecklistSourceChallenge, Challenge: ChallengeReisTypes})
	if err != nil || complete || client.searchCalls != numistaYearsPerView {
		t.Fatalf("first view: complete %v, %d searches (%v)", complete, client.searchCalls, err)
	}
	_, _, _ = svc.SlotsAndCompleteness(context.Background(), 1, ChecklistDefinition{Source: ChecklistSourceChallenge, Challenge: ChallengeReisTypes})
	if client.searchCalls != 2*numistaYearsPerView {
		t.Fatalf("the next view asks for the next years: %d searches", client.searchCalls)
	}
}
