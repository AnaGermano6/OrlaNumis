package services

import (
	"testing"
	"time"

	"github.com/briandenicola/ancient-coins-api/models"
	"github.com/briandenicola/ancient-coins-api/repository"
)

func TestPortugalCoinPeriod(t *testing.T) {
	for _, tc := range []struct {
		year   int
		system string
		want   string
	}{
		{2002, "Euro", "União Europeia (Euro)"},
		{2024, "", "União Europeia (Euro)"},
		{1992, "ECU", "ECU"},
		{1986, "Escudo (1911-2001)", "República Portuguesa 1986–2001"},
		{2001, "Escudo (1911-2001)", "República Portuguesa 1986–2001"},
		{1985, "Escudo (1911-2001)", "República Portuguesa 1969–1986"},
		{1969, "Escudo (1911-2001)", "República Portuguesa 1969–1986"},
		{1968, "Escudo (1911-2001)", "República Portuguesa 1910–1969"},
		{1912, "Escudo (1911-2001)", "República Portuguesa 1910–1969"},
		{1910, "Real (decimalizado, 1835-1910)", "Reino de Portugal 1836–1910"},
		{1836, "", "Reino de Portugal 1836–1910"},
		{1835, "", "Reino de Portugal 1777–1835"},
		{1777, "", "Reino de Portugal 1777–1835"},
		{1750, "", "Reino de Portugal 1706–1776"},
		{1706, "", "Reino de Portugal 1640–1706"},
		{1640, "", "Reino de Portugal 1640–1706"},
		{1639, "", "Reino de Portugal 1495–1639"},
		{1495, "", "Reino de Portugal 1415–1495"},
		{1415, "", "Reino de Portugal 1415–1495"},
		{1300, "", ""},
	} {
		if got := PortugalCoinPeriod(tc.year, tc.system); got != tc.want {
			t.Errorf("PortugalCoinPeriod(%d, %q) = %q, want %q", tc.year, tc.system, got, tc.want)
		}
	}
}

func TestPeriodFromNotes(t *testing.T) {
	for notes, want := range map[string]string{
		"Importado do uCoin.\nPeríodo: Rainha Isabel II\nVariante: Escudo": "Rainha Isabel II",
		"Period: Fifth Republic  ":        "Fifth Republic",
		"O período desta moeda é incerto": "",
		"":                                "",
	} {
		if got := periodFromNotes(notes); got != want {
			t.Errorf("periodFromNotes(%q) = %q, want %q", notes, got, want)
		}
	}
}

// A Portuguese coin's period is the one its type began in; another
// country's is the uCoin import's; a banknote has none. They are worked out
// for the coins saved before, and as coins are saved.
func TestCoinPeriodService_WorksOutPeriods(t *testing.T) {
	db := setupTestDB(t)
	catalog, _, _, store := newCatalogFixture()
	cached := func(key, payload string) {
		store.entries[key] = models.NumistaCacheEntry{Key: key, Payload: payload, FetchedAt: time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)}
	}
	// The new escudo series began in 1986; the old one ran on to 1986.
	cached("type:1:pt", `{"id": 1, "title": "1 Escudo", "min_year": 1986, "max_year": 2001}`)
	cached("type:2:pt", `{"id": 2, "title": "1 Escudo", "min_year": 1981, "max_year": 1986}`)
	cached("type:3:pt", `{"id": 3, "title": "2 Euro", "min_year": 2002, "max_year": 2026}`)
	year, old := 1986, 1950
	coins := []*models.Coin{
		{Name: "1 Escudo 1986 new", UserID: 1, Country: "Portugal", Year: &year, MonetarySystem: "Escudo (1911-2001)"},
		{Name: "1 Escudo 1986 old", UserID: 1, Country: "Portugal", Year: &year, MonetarySystem: "Escudo (1911-2001)", Notes: "Período: República Portuguesa"},
		{Name: "2 Euro", UserID: 1, Country: "Portugal", MonetarySystem: "Euro"},
		{Name: "1 Penny", UserID: 2, Country: "Reino Unido", Notes: "Período: Rainha Isabel II"},
		{Name: "Unlinked", UserID: 1, Country: "Portugal", Year: &old},
		{Name: "Banknote", UserID: 1, Country: "Portugal", Year: &old, Kind: models.CoinKindBanknote},
	}
	types := []string{"1", "2", "3", "", "", ""}
	for i, coin := range coins {
		if err := db.Create(coin).Error; err != nil {
			t.Fatal(err)
		}
		if types[i] != "" {
			if err := db.Create(&models.CoinReference{CoinID: coin.ID, Catalog: "Numista", Number: types[i]}).Error; err != nil {
				t.Fatal(err)
			}
		}
	}
	repo := repository.NewCoinRepository(db)
	periods := NewCoinPeriodService(repo, catalog, nil)
	if changed := periods.RefreshAll(); changed != 5 {
		t.Fatalf("changed = %d, want 5", changed)
	}
	want := []string{"República Portuguesa 1986–2001", "República Portuguesa 1969–1986", "União Europeia (Euro)", "Rainha Isabel II", "República Portuguesa 1910–1969", ""}
	period := func(id uint) string {
		var saved models.Coin
		db.First(&saved, id)
		return saved.Period
	}
	for i, coin := range coins {
		if got := period(coin.ID); got != want[i] {
			t.Errorf("%s period = %q, want %q", coin.Name, got, want[i])
		}
	}
	if changed := periods.RefreshAll(); changed != 0 {
		t.Fatalf("a second pass changed %d", changed)
	}

	// A coin saved through the coin service gets its period at once, and
	// again when what it is changes.
	svc := NewCoinService(repo).WithPeriods(periods)
	added := &models.Coin{Name: "50 Centavos 1945", UserID: 1, Country: "Portugal", DateRange: "1945"}
	if err := svc.CreateCoin(added); err != nil {
		t.Fatal(err)
	}
	if got := period(added.ID); got != "República Portuguesa 1910–1969" {
		t.Fatalf("new coin period = %q", got)
	}
	existing, _ := repo.FindByID(added.ID, 1)
	later := 1990
	if err := svc.UpdateCoinWithFields(existing, &models.Coin{Year: &later}, []string{"Year"}, 1, "test"); err != nil {
		t.Fatal(err)
	}
	if got := period(added.ID); got != "República Portuguesa 1986–2001" {
		t.Fatalf("after a new year, period = %q", got)
	}
}
