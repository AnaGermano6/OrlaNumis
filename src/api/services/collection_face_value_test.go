package services

import (
	"math"
	"testing"
	"time"

	"github.com/briandenicola/ancient-coins-api/models"
	"github.com/briandenicola/ancient-coins-api/repository"
)

func TestParseFaceValue(t *testing.T) {
	for denomination, want := range map[string]float64{
		"2 Euros":                       2,
		"50 Cêntimos":                   0.5,
		"10 euro cent":                  0.1,
		"2,5 Escudos":                   2.5,
		"2.50 €":                        2.5,
		"2$50":                          2.5,
		"1000$00":                       1000,
		"1.000 escudos":                 1000,
		"50 Centavos":                   0.5,
		"½ Euro":                        0.5,
		"7½ euros":                      7.5,
		"Centenário":                    0,
		"":                              0,
		"100 Escudos (Fernando Pessoa)": 100,
	} {
		if got := ParseFaceValue(denomination); math.Abs(got-want) > 1e-9 {
			t.Errorf("ParseFaceValue(%q) = %v, want %v", denomination, got, want)
		}
	}
}

func TestCollectionFaceValue_EurosAndEscudosOnly(t *testing.T) {
	db := setupTestDB(t)
	if err := db.AutoMigrate(&models.NumistaCacheEntry{}); err != nil {
		t.Fatal(err)
	}
	// Types described before: a 2 euro coin and a 50 centavos coin of the escudo.
	for key, payload := range map[string]string{
		"type:100:pt": `{"id": 100, "title": "2 Euros", "value": {"text": "2 Euros", "numeric_value": 2, "currency": {"name": "Euro", "full_name": "Euro (2002-date)"}}}`,
		"type:200:pt": `{"id": 200, "title": "50 Centavos", "value": {"text": "50 Centavos", "numeric_value": 0.5, "currency": {"name": "Escudo", "full_name": "Escudo (1911-2001)"}}}`,
	} {
		if err := db.Create(&models.NumistaCacheEntry{Key: key, Payload: payload, FetchedAt: time.Now()}).Error; err != nil {
			t.Fatal(err)
		}
	}
	value := func(v float64) *float64 { return &v }
	coins := []models.Coin{
		// Its type's face value: three of them, worth 3 each.
		{Name: "2 euro", Quantity: 3, CurrentValue: value(3)},
		// Its type's, in escudos: 0.5 escudos.
		{Name: "50 centavos", CurrentValue: value(1)},
		// Not linked: its denomination, in its monetary system; no current value.
		{Name: "1000 escudos", Kind: models.CoinKindBanknote, Denomination: "1000$00", MonetarySystem: "Escudo (1911-2001)"},
		// A denomination in euros needs no monetary system.
		{Name: "Hand-made 5 euro", Denomination: "5 Euros", CurrentValue: value(10)},
		// Other currencies have no face value in euros.
		{Name: "1 dollar", Denomination: "1 Dollar", MonetarySystem: "Dollar (1792-date)", CurrentValue: value(5), Quantity: 2},
		{Name: "Wished 2 euro", Denomination: "2 Euros", IsWishlist: true, CurrentValue: value(9)},
	}
	for i := range coins {
		coins[i].UserID, coins[i].Category, coins[i].Material = 1, models.CategoryModern, models.MaterialSilver
		if coins[i].Kind == "" {
			coins[i].Kind = models.CoinKindCoin
		}
		if err := db.Create(&coins[i]).Error; err != nil {
			t.Fatal(err)
		}
	}
	for coin, number := range map[int]string{0: "100", 1: "200"} {
		if err := db.Create(&models.CoinReference{CoinID: coins[coin].ID, Catalog: "Numista", Number: number}).Error; err != nil {
			t.Fatal(err)
		}
	}
	svc := NewCollectionFaceValueService(repository.NewCoinRepository(db), NewNumistaCatalogService(nil, NewNumistaResponseCache(repository.NewNumistaCacheRepository(db), nil, nil)))

	all, err := svc.Summarise(1, "")
	if err != nil {
		t.Fatal(err)
	}
	escudo := 0.5 / escudosPerEuro
	banknote := 1000 / escudosPerEuro
	if want := 6 + escudo + banknote + 5; math.Abs(all.EUR-want) > 1e-9 || all.Without != 2 {
		t.Errorf("all = %+v, want %v euros and 2 without", all, want)
	}
	// The banknote has no current value: left out of the comparison.
	if want := 6 + escudo + 5; math.Abs(all.ValuedEUR-want) > 1e-9 || all.Value != 9+1+10 {
		t.Errorf("valued = %v of %v, want %v of 20", all.ValuedEUR, all.Value, want)
	}

	banknotes, err := svc.Summarise(1, models.CoinKindBanknote)
	if err != nil {
		t.Fatal(err)
	}
	if math.Abs(banknotes.EUR-banknote) > 1e-9 || banknotes.Without != 0 || banknotes.ValuedEUR != 0 {
		t.Errorf("banknotes = %+v", banknotes)
	}
}
