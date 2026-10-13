package services

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/briandenicola/ancient-coins-api/models"
	"github.com/briandenicola/ancient-coins-api/repository"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

type fakeMOEShopClient struct {
	pages map[string][]byte
	calls int
}

func (f *fakeMOEShopClient) Page(_ context.Context, path string) ([]byte, error) {
	f.calls++
	if page, ok := f.pages[path]; ok {
		return page, nil
	}
	return nil, errors.New("no such page")
}

func newTestMoeService(t *testing.T) (*MoeCoinPhotoService, *fakeMOEShopClient) {
	t.Helper()
	client := &fakeMOEShopClient{pages: map[string][]byte{}}
	for path, file := range map[string]string{
		moeSitemapPath:                                       "moe_sitemap.xml",
		fmt.Sprintf(moeCataloguePath, "5"):                   "moe_catalogue_5.html",
		fmt.Sprintf(moeCataloguePath, "10"):                  "moe_catalogue_10.html",
		"/produkte/5-euro-kupfermuenze-spring-ins-glueck-np": "moe_product.html",
	} {
		body, err := os.ReadFile("testdata/" + file)
		if err != nil {
			t.Fatal(err)
		}
		client.pages[path] = body
	}
	for _, value := range moeCatalogueValues {
		if _, ok := client.pages[fmt.Sprintf(moeCataloguePath, value)]; !ok {
			client.pages[fmt.Sprintf(moeCataloguePath, value)] = []byte("<html></html>")
		}
	}
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&models.NumistaCacheEntry{}); err != nil {
		t.Fatal(err)
	}
	return NewMoeCoinPhotoService(client, NewNumistaResponseCache(repository.NewNumistaCacheRepository(db), nil, nil)), client
}

func TestParseMoeCatalogue(t *testing.T) {
	service, _ := newTestMoeService(t)
	coins, err := service.Coins(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	pages := map[string]mintCoin{}
	for _, coin := range coins {
		pages[coin.PageURL] = coin
	}
	// Each version the shop sells, found by the entry's name ("‘s Glück is a
	// Vogerl" has no link of its own) or its link.
	for page, want := range map[string]struct {
		title, metal string
		year         int
		value        float64
	}{
		moeShopURL + "/produkte/5-euro-kupfermuenze-spring-ins-glueck-np":    {"Spring ins Glück", "copper", 2026, 5},
		moeShopURL + "/produkte/5-euro-silbermuenze-spring-ins-glueck-hgh":   {"Spring ins Glück", "silver", 2026, 5},
		moeShopURL + "/produkte/5-euro-kupfermuenze-weltmeister-np":          {"Weltmeister", "copper", 2024, 5},
		moeShopURL + "/produkte/5-euro-kupfermuenze-s-glueck-is-a-vogerl-np": {"‘s Glück is a Vogerl", "copper", 2022, 5},
		moeShopURL + "/produkte/5-euro-kupfermuenze-osterkueken-np":          {"Osterküken", "copper", 2021, 5},
		moeShopURL + "/produkte/10-euro-kupfermuenze-aufdirndln-leinen-np":   {"Aufdirndln - Leinen", "copper", 2025, 10},
	} {
		coin, ok := pages[page]
		if !ok || coin.Title != want.title || coin.Metal != want.metal || coin.Year != want.year || coin.Value != want.value || coin.Obverse != "" {
			t.Errorf("%s: coin = %+v (found %v)", page, coin, ok)
		}
	}
	// A coin no longer sold has no photo: the catalogue's picture shows both
	// sides in two metals at once.
	salzburg, ok := pages[moeShopURL+fmt.Sprintf(moeCataloguePath, "10")]
	if !ok || salzburg.Title != "Salzburg" || salzburg.Year != 2014 || salzburg.Series != "Bundesländer-Serie" || salzburg.Obverse != "" {
		t.Errorf("salzburg = %+v", salzburg)
	}
	if len(coins) != 10 {
		t.Errorf("got %d coins, want 10 (sets are not coins)", len(coins))
	}
}

func TestMoeCoinPhotoService_MatchAndConfirm(t *testing.T) {
	service, client := newTestMoeService(t)
	ctx := context.Background()
	year := func(y int) *int { return &y }
	five, ten := 5.0, 10.0
	for _, c := range []struct {
		detail models.NumistaCatalogType
		want   string
	}{
		// By the lettering of the design: Numista names the coin in English.
		{models.NumistaCatalogType{Title: "5 Euros (Jump for Joy Coin, copper)", MinYear: year(2025), FaceValue: &five, Currency: "Euro", Composition: "Cobre",
			Reverse: models.NumistaCatalogSide{Lettering: "SPRING\r\nINS\r\nGLÜCK\r\n2026"}}, "5-euro-kupfermuenze-spring-ins-glueck-np"},
		{models.NumistaCatalogType{Title: "5 Euro (Alpine World Ski Championships, copper)", MinYear: year(2024), FaceValue: &five, Currency: "Euro", Composition: "Cobre",
			Reverse: models.NumistaCatalogSide{Lettering: "1\r\nWELTMEISTER"}}, "5-euro-kupfermuenze-weltmeister-np"},
		{models.NumistaCatalogType{Title: "5 Euros (New Year)", MinYear: year(2022), FaceValue: &five, Currency: "Euro", Composition: "Cobre",
			Reverse: models.NumistaCatalogSide{Lettering: "2022\r\n'S GLÜCK\r\nIS A \r\nVOGERL"}}, "s-glueck-is-a-vogerl"},
		// By the title's subject ("Linen" is "Leinen").
		{models.NumistaCatalogType{Title: "10 Euros (Aufdirndln – Linen, copper)", MinYear: year(2025), FaceValue: &ten, Currency: "Euro", Composition: "Cobre"}, "10-euro-kupfermuenze-aufdirndln-leinen-np"},
		{models.NumistaCatalogType{Title: "10 Euros (Salzburg)", MinYear: year(2014), FaceValue: &ten, Currency: "Euro", Reverse: models.NumistaCatalogSide{Lettering: "SALZBURG"}}, "muenzkatalog/10-euro-muenzen"},
	} {
		query, _ := mintQueryForType(&c.detail)
		photo, err := service.Match(ctx, query)
		if err != nil || photo == nil || !strings.Contains(photo.PageURL, c.want) {
			t.Errorf("%s: photo = %+v, err = %v", c.detail.Title, photo, err)
		}
	}
	// Easter's lettering is only its year: nothing names the chick.
	easter := models.NumistaCatalogType{Title: "5 Euros (Easter coin 2021)", MinYear: year(2021), FaceValue: &five, Currency: "Euro", Reverse: models.NumistaCatalogSide{Lettering: "2021"}}
	query, _ := mintQueryForType(&easter)
	if photo, _ := service.Match(ctx, query); photo != nil {
		t.Errorf("easter: expected no photo, got %+v", photo)
	}

	// The shop's page has a photo of each side: VS is the value side.
	photo, _ := service.Match(ctx, MintPhotoQuery{Year: 2026, Value: 5, Topic: "Spring ins Glück", Metal: "copper"})
	confirmed := service.Confirm(ctx, photo)
	if !strings.HasSuffix(confirmed.Obverse, "Spring_ins_Glueck_Cu_NP_VS_2D.png.webp") || !strings.HasSuffix(confirmed.Reverse, "Spring_ins_Glueck_Cu_NP_RS_2D.png.webp") {
		t.Fatalf("confirmed = %+v", confirmed)
	}
	calls := client.calls
	service.Confirm(ctx, photo)
	if client.calls != calls {
		t.Fatal("a product page is read once")
	}
	// Without the shop's page there is no photo.
	weltmeister, _ := service.Match(ctx, MintPhotoQuery{Year: 2024, Value: 5, Topic: "Weltmeister", Metal: "copper"})
	if kept := service.Confirm(ctx, weltmeister); kept == nil || kept.Obverse != "" || kept.Reverse != "" {
		t.Fatalf("kept = %+v", kept)
	}
	obverse, reverse := mintCandidates(service, weltmeister)
	if obverse != nil || reverse != nil {
		t.Fatalf("no photo, no candidates: %+v %+v", obverse, reverse)
	}
	if service.Country() != "Austria" || service.Source() != models.CoinPhotoSourceMOE ||
		service.Thumbnail(confirmed.Obverse) != strings.Replace(confirmed.Obverse, "/product_full/", "/product_preview/", 1) {
		t.Fatal("unexpected source description")
	}
}

func TestMoeSlugAndProductPhotos(t *testing.T) {
	for name, want := range map[string]string{
		"‘s Glück is a Vogerl": "s-glueck-is-a-vogerl",
		"365 + 1 Tag Glück":    "365-1-tag-glueck",
		"Aufdirndln - Leinen":  "aufdirndln-leinen",
		"Osterküken":           "osterkueken",
	} {
		if got := moeSlug(name); got != want {
			t.Errorf("moeSlug(%q) = %q, want %q", name, got, want)
		}
	}
	page := []byte(`<img src="/var/storage/images/_aliases/product_full/1/a-2024_5E_X_Ag_HGH_Blister.png.webp">
		<img src="/var/storage/images/_aliases/product_full/2/b-2024_5E_X_VS_RS_2D.png.webp">
		<img src="/var/storage/images/_aliases/product_full/3/5e_2021_ostermuenze_cu_rs.png.webp">
		<img src="/var/storage/images/_aliases/product_full/4/5e_2021_ostermuenze_cu_vs.png.webp">`)
	obverse, reverse := ParseMoeProductPhotos(page)
	if obverse != moeShopURL+"/var/storage/images/_aliases/product_full/4/5e_2021_ostermuenze_cu_vs.png.webp" ||
		reverse != moeShopURL+"/var/storage/images/_aliases/product_full/3/5e_2021_ostermuenze_cu_rs.png.webp" {
		t.Fatalf("obverse = %q, reverse = %q", obverse, reverse)
	}
}
