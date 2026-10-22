package services

import (
	"bytes"
	"encoding/csv"
	"strings"
	"testing"
	"time"

	"github.com/briandenicola/ancient-coins-api/models"
	"github.com/briandenicola/ancient-coins-api/repository"
)

func TestNormalizeShareListRequest(t *testing.T) {
	got, err := NormalizeShareListRequest(ShareListRequest{List: " Trades ", Format: "PDF", Lang: "pt-PT", Contact: "  ana@exemplo.pt \n 912 000 000 "})
	if err != nil || got.List != ShareListTrades || got.Format != "pdf" || got.Lang != "pt" || got.Contact != "ana@exemplo.pt 912 000 000" {
		t.Fatalf("normalized = %+v (%v)", got, err)
	}
	if got, _ := NormalizeShareListRequest(ShareListRequest{List: "both", Format: "csv", Lang: "en"}); got.Lang != "en" {
		t.Fatalf("english = %+v", got)
	}
	if got, _ := NormalizeShareListRequest(ShareListRequest{List: "both", Format: "csv"}); got.Lang != "pt" {
		t.Fatalf("no language is Portuguese: %+v", got)
	}
	for _, bad := range []ShareListRequest{{List: "everything", Format: "pdf"}, {List: "trades", Format: "xlsx"}} {
		if _, err := NormalizeShareListRequest(bad); err != ErrShareListInvalid {
			t.Errorf("%+v: err = %v", bad, err)
		}
	}
}

func TestShareListService_BuildsTradesAndWanted(t *testing.T) {
	db := setupTestDB(t)
	value := 12.5
	year := func(y int) *int { return &y }
	coins := []*models.Coin{
		{Name: "50 Centavos 1968", UserID: 1, Country: "Portugal", Year: year(1968), Denomination: "50 Centavos", Grade: "VF", Quantity: 3, CurrentValue: &value},
		{Name: "2 Euro 2007", UserID: 1, Country: "Áustria", Year: year(2007), Denomination: "2 Euro", Grade: "UNC", Quantity: 2},
		{Name: "1 Escudo 1986", UserID: 1, Country: "Portugal", Year: year(1986), Quantity: 1},
		{Name: "100 Escudos 1972", UserID: 1, Kind: models.CoinKindBanknote, Country: "Angola", Year: year(1972), Quantity: 2, CertifiedGrade: "PMG 64"},
		{Name: "200 Escudos 1991", UserID: 1, Country: "Portugal", DateRange: "1991-2001", IsWishlist: true, Quantity: 1},
		{Name: "Other user's", UserID: 2, Country: "Portugal", Quantity: 5},
	}
	for _, coin := range coins {
		if err := db.Create(coin).Error; err != nil {
			t.Fatal(err)
		}
	}
	db.Create(&models.CoinReference{CoinID: coins[0].ID, Catalog: "KM", Number: "596"})
	db.Create(&models.CoinReference{CoinID: coins[0].ID, Catalog: "Numista", Number: "1234"})
	db.Create(&models.CoinImage{CoinID: coins[0].ID, FilePath: "coins/1/reverse.jpg", ImageType: models.ImageTypeReverse})
	db.Create(&models.CoinImage{CoinID: coins[0].ID, FilePath: "coins/1/obverse.jpg", ImageType: models.ImageTypeObverse})
	settings := NewSettingsService(repository.NewSettingsRepository(db))
	if err := settings.SetSetting(SettingCollectionCurrency, CurrencyEUR); err != nil {
		t.Fatal(err)
	}
	svc := NewShareListService(repository.NewCoinRepository(db), settings)
	svc.now = func() time.Time { return time.Date(2026, 10, 21, 9, 0, 0, 0, time.UTC) }

	list, err := svc.Build(1, ShareListRequest{List: ShareListBoth, Format: "csv", Lang: "pt", IncludeValues: true, Contact: "ana@exemplo.pt"})
	if err != nil {
		t.Fatal(err)
	}
	if list.Title != "Trocas e procuras" || len(list.Sections) != 2 || list.FileName(ShareListBoth, "csv") != "trocas-e-procuras-2026-10-21.csv" {
		t.Fatalf("list = %+v", list)
	}
	trades, wanted := list.Sections[0], list.Sections[1]
	names := []string{}
	for _, item := range trades.Items {
		names = append(names, item.Name)
	}
	// Held more than once, by country (accents ignored), year and name.
	if strings.Join(names, " | ") != "100 Escudos 1972 | 2 Euro 2007 | 50 Centavos 1968" {
		t.Fatalf("trades = %v", names)
	}
	centavos := trades.Items[2]
	if centavos.Available != 2 || centavos.References != "KM 596 · Numista 1234" || centavos.NumistaURL != "https://en.numista.com/catalogue/pieces1234.html" ||
		centavos.Photo != "coins/1/obverse.jpg" || centavos.Value == nil || *centavos.Value != 12.5 || centavos.Year != "1968" {
		t.Fatalf("50 centavos = %+v", centavos)
	}
	if banknote := trades.Items[0]; !banknote.Banknote || banknote.Grade != "PMG 64" || banknote.Available != 1 {
		t.Fatalf("banknote = %+v", banknote)
	}
	if len(wanted.Items) != 1 || wanted.Items[0].Year != "1991-2001" || wanted.Items[0].Available != 0 || wanted.Title != "Procuro" {
		t.Fatalf("wanted = %+v", wanted)
	}

	body, err := list.CSV()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.HasPrefix(body, []byte("\xef\xbb\xbf")) {
		t.Fatal("the CSV starts with a byte order mark")
	}
	reader := csv.NewReader(bytes.NewReader(bytes.TrimPrefix(body, []byte("\xef\xbb\xbf"))))
	reader.Comma = ';'
	rows, err := reader.ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	if strings.Join(rows[0], "|") != "Lista|Tipo|País|Peça|Ano|Valor facial|Grau|Disponíveis|Referências|Numista|Valor" || len(rows) != 5 {
		t.Fatalf("rows = %q", rows)
	}
	if got := strings.Join(rows[3], "|"); got != "Lista de trocas|Moeda|Portugal|50 Centavos 1968|1968|50 Centavos|VF|2|KM 596 · Numista 1234|https://en.numista.com/catalogue/pieces1234.html|12,50 €" {
		t.Fatalf("row = %s", got)
	}
	if got := strings.Join(rows[4], "|"); got != "Procuro|Moeda|Portugal|200 Escudos 1991|1991-2001||||||" {
		t.Fatalf("wanted row = %s", got)
	}

	// Without values, in English, separated by commas.
	english, _ := svc.Build(1, ShareListRequest{List: ShareListWishlist, Format: "csv", Lang: "en"})
	body, _ = english.CSV()
	if !strings.Contains(string(body), "List,Kind,Country,Item,Year,Face value,Grade,Available,References,Numista\n") ||
		strings.Contains(string(body), "Value") || english.FileName(ShareListWishlist, "pdf") != "wanted-2026-10-21.pdf" {
		t.Fatalf("english csv = %s", body)
	}
}

func TestShareListCSV_KeepsCellsFromBeingFormulas(t *testing.T) {
	list := ShareList{Lang: "en", Labels: shareListLabels["en"], Sections: []ShareListSection{{Kind: ShareListWishlist, Title: "Wanted", Items: []ShareListItem{{Name: "=HYPERLINK(\"x\")", Country: "+Portugal"}}}}}
	body, err := list.CSV()
	if err != nil || !strings.Contains(string(body), `"'=HYPERLINK(""x"")"`) || !strings.Contains(string(body), "'+Portugal") {
		t.Fatalf("csv = %s (%v)", body, err)
	}
}

func TestFormatShareValue(t *testing.T) {
	cases := []struct {
		amount         float64
		currency, lang string
		want           string
	}{
		{12.5, "EUR", "pt", "12,50 €"}, {12.5, "EUR", "en", "€12.50"}, {3, "USD", "en", "$3.00"}, {3, "CHF", "pt", "3,00 CHF"},
	}
	for _, tc := range cases {
		if got := FormatShareValue(tc.amount, tc.currency, tc.lang); got != tc.want {
			t.Errorf("FormatShareValue(%v, %s, %s) = %q, want %q", tc.amount, tc.currency, tc.lang, got, tc.want)
		}
	}
}
