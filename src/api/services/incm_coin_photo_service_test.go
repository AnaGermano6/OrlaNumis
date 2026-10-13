package services

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/briandenicola/ancient-coins-api/models"
	"github.com/briandenicola/ancient-coins-api/repository"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

type fakeINCMShopClient struct {
	pages map[string][]byte
	calls int
	err   error
}

func (f *fakeINCMShopClient) ProductsPage(_ context.Context, lang string, page int) ([]byte, error) {
	f.calls++
	if f.err != nil {
		return nil, f.err
	}
	if body, ok := f.pages[lang]; ok && page == 1 {
		return body, nil
	}
	return []byte(`{"products": []}`), nil
}

// incmEnglishTitles is the English feed: the same products, English titles.
func incmEnglishTitles(t *testing.T, body []byte) []byte {
	t.Helper()
	var feed map[string][]map[string]any
	if err := json.Unmarshal(body, &feed); err != nil {
		t.Fatal(err)
	}
	english := map[string]string{"O Conhecimento": "The knowledge", "Ulisses — Heróis e Criaturas da Mitologia": "Ulysses — Heroes and Creatures of Mythology"}
	for _, product := range feed["products"] {
		if title, ok := english[product["title"].(string)]; ok {
			product["title"] = title
		}
	}
	translated, err := json.Marshal(feed)
	if err != nil {
		t.Fatal(err)
	}
	return translated
}

func newTestIncmService(t *testing.T) (*IncmCoinPhotoService, *fakeINCMShopClient) {
	t.Helper()
	body, err := os.ReadFile("testdata/incm_products.json")
	if err != nil {
		t.Fatal(err)
	}
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&models.NumistaCacheEntry{}); err != nil {
		t.Fatal(err)
	}
	client := &fakeINCMShopClient{pages: map[string][]byte{"": body, "en": incmEnglishTitles(t, body)}}
	return NewIncmCoinPhotoService(client, NewNumistaResponseCache(repository.NewNumistaCacheRepository(db), nil, nil)), client
}

const incmFiles = "https://cdn.shopify.com/s/files/1/0525/1910/8762/"

func TestParseIncmProducts(t *testing.T) {
	body, err := os.ReadFile("testdata/incm_products.json")
	if err != nil {
		t.Fatal(err)
	}
	page, err := ParseIncmProducts(body)
	if err != nil {
		t.Fatal(err)
	}
	coins := map[string]mintCoin{}
	for _, coin := range page.coins {
		coins[coin.Title] = coin
	}
	// The book and the yearly set of coins are left out.
	if page.count != 13 || len(page.coins) != 11 {
		t.Fatalf("count = %d, coins = %d", page.count, len(page.coins))
	}
	knowledge := coins["O Conhecimento"]
	if knowledge.Value != 5 || knowledge.Year != 2024 || knowledge.Authors != "António Cerqueira" ||
		!strings.Contains(knowledge.Obverse, "Conhecimento_CU_A.jpg") || !strings.Contains(knowledge.Reverse, "Conhecimento_CU_B.jpg") {
		t.Fatalf("O Conhecimento = %+v", knowledge)
	}
	// When it went on sale and how many were struck.
	if knowledge.Released != "2024-11" || knowledge.Volume != 20000 {
		t.Fatalf("O Conhecimento release = %q, volume = %d", knowledge.Released, knowledge.Volume)
	}
	// Numbered photos, and packaging photos that are numbered too.
	catarina := coins["D. Catarina de Bragança"]
	if catarina.Year != 2016 || !strings.HasSuffix(strings.Split(catarina.Obverse, "?")[0], "-1.png") || !strings.HasSuffix(strings.Split(catarina.Reverse, "?")[0], "-2.png") {
		t.Fatalf("D. Catarina = %+v", catarina)
	}
	lighthouse := coins["Cabo de São Vicente — Faróis de Portugal (Prata Proof)"]
	if !strings.Contains(lighthouse.Obverse, "FarolPrata_a.jpg") || !strings.Contains(lighthouse.Reverse, "FarolPrata_b.jpg") {
		t.Fatalf("Cabo de São Vicente = %+v", lighthouse)
	}
	if coins["900 anos do foral de Ponte de Lima"].Value != 7.5 {
		t.Fatalf("Ponte de Lima = %+v", coins["900 anos do foral de Ponte de Lima"])
	}
}

func TestIncmCoinPhotoService_Match(t *testing.T) {
	service, client := newTestIncmService(t)
	ctx := context.Background()
	cases := []struct {
		name  string
		query MintPhotoQuery
		want  string
	}{
		{"Numista's Portuguese topic and designer", MintPhotoQuery{Year: 2024, Value: 5, Metal: "Cuproníquel", Topic: "O conhecimento", Designers: []string{"Antonio Cerqueira"},
			Texts: []string{"Desenhe uma moeda", "À direita, a imagem de um frasco, com as palavras \"Educação\" e \"Tecnológica\""}}, "Conhecimento_CU_A.jpg"},
		{"the silver version", MintPhotoQuery{Year: 2024, Value: 5, Metal: "Prata 925‰", Topic: "O conhecimento"}, "Conhecimento_AG_A.jpg"},
		{"the normal edition when the metal is unknown", MintPhotoQuery{Year: 2024, Value: 5, Topic: "O conhecimento"}, "Conhecimento_CU_A.jpg"},
		{"the base-metal version of a coin also sold in silver", MintPhotoQuery{Year: 2025, Value: 5, Metal: "Cuproníquel", Topic: "Cape St. Vincent", Designers: []string{"José João de Brito"}}, "FAROL_CU_A.jpg"},
		{"an English title", MintPhotoQuery{Year: 2024, Value: 5, Topic: "Ulysses"}, "Ulises_CU_A.jpg"},
		{"the shop's English title", MintPhotoQuery{Year: 2024, Value: 5, Topic: "The knowledge, cupronickel"}, "Conhecimento_CU_A.jpg"},
		{"a seven and a half euro coin", MintPhotoQuery{Year: 2025, Value: 7.5, Topic: "900 years of the Charter of Ponte de Lima"}, "foral_CU_A.jpg"},
		{"released the year after", MintPhotoQuery{Year: 2025, Value: 5, Metal: "Silver", Topic: "Cape St. Vincent Lighthouse", Designers: []string{"José João de Brito"}}, "FarolPrata_a.jpg"},
		{"an older coin still sold", MintPhotoQuery{Year: 2016, Value: 5, Topic: "Catarina de Bragança"}, "catarina-de-braganca-normal-12500352-1"},
	}
	for _, c := range cases {
		photo, err := service.Match(ctx, c.query)
		if err != nil || photo == nil || !strings.Contains(photo.Obverse, c.want) || !strings.HasPrefix(photo.PageURL, "https://loja.incm.pt/products/") {
			t.Errorf("%s: photo = %+v, err = %v", c.name, photo, err)
		}
	}

	for _, query := range []MintPhotoQuery{
		// A topic with no word in common, and no designer.
		{Year: 2024, Value: 5, Topic: "Skating Federation"},
		// Two coins of the series fit equally well.
		{Year: 2025, Value: 5, Topic: "Heroes and Mythological Creatures"},
		// Another value or year.
		{Year: 2024, Value: 10, Topic: "O conhecimento"},
		{Year: 2021, Value: 5, Topic: "O conhecimento"},
		{Year: 2021, Value: 5, Topic: "Aristides de Sousa Mendes"},
	} {
		if photo, err := service.Match(ctx, query); err != nil || photo != nil {
			t.Errorf("%+v: expected no photo, got %+v (%v)", query, photo, err)
		}
	}
	if client.calls != 4 {
		t.Fatalf("the shop should be read once (two pages in each language), got %d calls", client.calls)
	}
}

func TestMatchMintCoin_IgnoresAnniversaryNumbersTheTitleLacks(t *testing.T) {
	coins := []mintCoin{
		{Title: "Disney animation classics - Fantasia", OtherTitles: []string{"classiques de l animation disney fantasia"}, Value: 10, Year: 2025, Obverse: "f.jpg", PageURL: "f"},
		{Title: "Disney animation classics - Pinocchio", OtherTitles: []string{"classiques de l animation disney pinocchio"}, Value: 10, Year: 2025, Obverse: "p.jpg", PageURL: "p"},
	}
	coin := matchMintCoin(coins, MintPhotoQuery{Year: 2025, Value: 10, Topic: "Clássicos da Animação Disney - 85.º aniversário de Fantasia"})
	if coin == nil || coin.PageURL != "f" {
		t.Fatalf("coin = %+v", coin)
	}
	if coin := matchMintCoin(coins, MintPhotoQuery{Year: 2025, Value: 10, Topic: "Clássicos da Animação Disney - Bambi"}); coin != nil {
		t.Fatalf("Bambi is neither coin: %+v", coin)
	}
}

// The Numista topic can add words the shop's title lacks ("75th anniversary
// of Peanuts", a translation), and its first year can be the year before the
// coin's: the title's subject alone, and the coin's year, are tried then.
func TestMatchMintCoin_TriesTheTitleSubjectAndTheCoinsYear(t *testing.T) {
	shop := "https://www.monnaiedeparis.fr/en/"
	coins := []mintCoin{
		{Title: "Snoopy", DesignKey: "snoopy", Value: 10, Metal: "silver", Year: 2025, Obverse: "s.jpg", PageURL: shop + "snoopy-10eur-silver-coin-proof-quality-yeardate-2025"},
		{Title: "Snoopy", DesignKey: "snoopy snoopy", Value: 10, Metal: "silver", Year: 2025, Obverse: "s.jpg", PageURL: shop + "snoopy-10eur-silver-coin-snoopy-shape-proof-quality-yeardate-2025"},
		{Title: "Snoopy", DesignKey: "snoopy", Value: 10, Metal: "silver", Year: 2025, Obverse: "s.jpg", PageURL: shop + "snoopy-10eur-silver-coin-circulating-quality-yeardate-2025"},
		{Title: "Disney animation classics - Pinocchio", DesignKey: "disney animation classics pinocchio", Value: 10, Metal: "silver", Year: 2025, Obverse: "p.jpg", PageURL: shop + "disney-animation-classics-pinocchio-10eur-silver-coin-circulating-quality-yeardate-2025"},
		{Title: "Astrological signs - Leo", DesignKey: "astrological signs leo", Value: 5, Metal: "common metal", Year: 2025, Obverse: "l.jpg", PageURL: shop + "astrological-signs-leo-5eur-common-metal-coin-circulating-quality-yeardate-2025"},
		{Title: "Astrological signs - Aries", DesignKey: "astrological signs aries", Value: 5, Metal: "common metal", Year: 2025, Obverse: "a.jpg", PageURL: shop + "astrological-signs-aries-5eur-common-metal-coin-circulating-quality-yeardate-2025"},
	}
	year2024, year2025, ten, five := 2024, 2025, 10.0, 5.0
	for _, c := range []struct {
		detail models.NumistaCatalogType
		want   string
	}{
		// The circulating edition, not the proof or the shaped one.
		{models.NumistaCatalogType{Title: "10 euros Snoopy (argent 333‰)", CommemoratedTopic: "75º aniversário dos Peanuts", MinYear: &year2025, FaceValue: &ten, Currency: "Euro", Composition: "Bilhão (prata 333‰)"}, "snoopy-10eur-silver-coin-circulating"},
		{models.NumistaCatalogType{Title: "10 Euros (Pinocchio, silver 333‰)", CommemoratedTopic: "Pinóquio.", MinYear: &year2025, FaceValue: &ten, Currency: "Euro", Composition: "Bilhão (prata 333‰)"}, "pinocchio-10eur"},
		{models.NumistaCatalogType{Title: "5 Euros (Leo)", CommemoratedTopic: "Leão", MinYear: &year2024, FaceValue: &five, Currency: "Euro"}, "astrological-signs-leo"},
	} {
		query, ok := mintQueryForType(&c.detail)
		coin := matchMintCoin(coins, query)
		if !ok || coin == nil || !strings.Contains(coin.PageURL, c.want) {
			t.Errorf("%s: coin = %+v", c.detail.Title, coin)
		}
	}
	// The coin's own year, when Numista's is further off, is tried as well.
	query := MintPhotoQuery{Year: 2023, Value: 5, Topic: "Leo"}
	if coin := matchMintCoin(coins, query); coin != nil {
		t.Errorf("no Leo of 2022 to 2024: coin = %+v", coin)
	}
	query.AltYears = []int{2025}
	if coin := matchMintCoin(coins, query); coin == nil || !strings.Contains(coin.PageURL, "leo") {
		t.Errorf("Leo of 2025: coin = %+v", coin)
	}
	// A subject no design has still finds nothing.
	if coin := matchMintCoin(coins, MintPhotoQuery{Year: 2025, Value: 5, Topic: "Leão", AltTopics: []string{"Virgo"}}); coin != nil {
		t.Errorf("Virgo is no coin of the shop: %+v", coin)
	}
}

func TestIncmCoinPhotoService_ShopUnavailable(t *testing.T) {
	service, client := newTestIncmService(t)
	client.err = errors.New("offline")
	query := MintPhotoQuery{Year: 2024, Value: 5, Topic: "O conhecimento"}
	if photo, err := service.Match(context.Background(), query); photo != nil || err == nil {
		t.Fatalf("photo = %+v, err = %v", photo, err)
	}
	for range 2 {
		if photo, _ := service.Match(context.Background(), query); photo != nil {
			t.Fatalf("photo = %+v", photo)
		}
	}
	if client.calls != 1 {
		t.Fatalf("an unavailable shop should not be asked on every search, got %d calls", client.calls)
	}
}

func TestMintFaceValueAndTopic(t *testing.T) {
	for denomination, want := range map[string]float64{
		"5 Euros": 5, "5 Euro (The knowledge, cupronickel)": 5, "7½ euro": 7.5, "2½ Euro": 2.5, "2,50 €": 2.5,
		"10 Euro Cents": 0, "1 cêntimo": 0, "200 Escudos": 0, "": 0,
	} {
		if got := MintFaceValue(denomination); got != want {
			t.Errorf("MintFaceValue(%q) = %v, want %v", denomination, got, want)
		}
	}
	if got := MintTopic("5 Euro (The knowledge, cupronickel)"); got != "The knowledge, cupronickel" {
		t.Errorf("MintTopic = %q", got)
	}
	for title, want := range map[string]string{
		"10 euros Snoopy (argent 333‰)":       "snoopy",
		"10 Euros (Pinocchio, silver 333‰)":   "pinocchio",
		"5 Euro (The knowledge, cupronickel)": "knowledge",
		"¼ Euro (Year of the Snake)":          "snake",
		"2 Euros (Olympic Team; Coloured)":    "olympic team",
	} {
		if got := MintTitleSubject(title); got != want {
			t.Errorf("MintTitleSubject(%q) = %q, want %q", title, got, want)
		}
	}
}
