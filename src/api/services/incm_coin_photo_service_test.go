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
}
