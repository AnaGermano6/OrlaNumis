package services

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/briandenicola/ancient-coins-api/models"
	"github.com/briandenicola/ancient-coins-api/repository"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

type fakeMDPShopClient struct {
	sitemaps map[string][]byte
	images   map[string][]byte
	calls    int
	err      error
}

func (f *fakeMDPShopClient) Sitemap(_ context.Context, lang string) ([]byte, error) {
	f.calls++
	if f.err != nil {
		return nil, f.err
	}
	return f.sitemaps[lang], nil
}

func (f *fakeMDPShopClient) Image(_ context.Context, url string) ([]byte, error) {
	f.calls++
	if image, ok := f.images[url]; ok {
		return image, nil
	}
	return nil, errors.New("no such photo")
}

func newTestMdpService(t *testing.T) (*MdpCoinPhotoService, *fakeMDPShopClient) {
	t.Helper()
	client := &fakeMDPShopClient{sitemaps: map[string][]byte{}, images: map[string][]byte{}}
	for _, lang := range []string{"en", "fr"} {
		body, err := os.ReadFile("testdata/mdp_sitemap_" + lang + ".xml")
		if err != nil {
			t.Fatal(err)
		}
		client.sitemaps[lang] = body
	}
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&models.NumistaCacheEntry{}); err != nil {
		t.Fatal(err)
	}
	return NewMdpCoinPhotoService(client, NewNumistaResponseCache(repository.NewNumistaCacheRepository(db), nil, nil)), client
}

const mdpPhotos = "https://www.monnaiedeparis.fr/media/catalog/product/"

func TestParseMdpSitemaps(t *testing.T) {
	en, _ := os.ReadFile("testdata/mdp_sitemap_en.xml")
	fr, _ := os.ReadFile("testdata/mdp_sitemap_fr.xml")
	coins, err := ParseMdpSitemaps(en, fr)
	if err != nil {
		t.Fatal(err)
	}
	byKey := map[string]mintCoin{}
	for _, coin := range coins {
		byKey[coin.DesignKey] = coin
	}
	// The lot of Astérix coins and the museum page are left out.
	if len(coins) != 8 {
		t.Fatalf("coins = %d: %+v", len(coins), byKey)
	}
	aries := byKey["astrological signs aries"]
	if aries.Value != 5 || aries.Year != 2025 || aries.Metal != "common metal" || aries.Title != "Astrological signs - Aries" ||
		aries.Obverse != mdpPhotos+"M/5/M5066_MDP_00066461_BD_c3cb.jpg" || aries.Reverse != "" || aries.ReverseGuess != mdpPhotos+"M/5/M5066_MDP_00066462_BD_b095.jpg" ||
		!strings.Contains(strings.Join(aries.OtherTitles, "|"), "signes astros belier") {
		t.Fatalf("aries = %+v", aries)
	}
	// A photo named "revers" is the reverse.
	baker := byKey["josephine baker"]
	if baker.Value != 10 || !strings.HasSuffix(baker.Reverse, "_25_BAKER_10_AG_revers_BD_2e11.jpg") || baker.ReverseGuess != "" {
		t.Fatalf("baker = %+v", baker)
	}
	if england := byKey["rugby world cup france england"]; england.Value != 0.25 || england.Year != 2023 {
		t.Fatalf("england = %+v", england)
	}
}

func TestMdpCoinPhotoService_Match(t *testing.T) {
	service, client := newTestMdpService(t)
	ctx := context.Background()
	for _, c := range []struct {
		title string
		year  int
		want  string
	}{
		{"5 euros Bélier", 2025, "astrological-signs-aries-5eur"},
		// Numista's type starts the year before.
		{"5 Euros (Aries)", 2024, "astrological-signs-aries-5eur"},
		{"10 Euro (Josephine Baker)", 2025, "josephine-baker-10eur"},
		{"10 Euros (Year of the Snake, silver 999‰)", 2025, "year-of-the-snake-10eur"},
		{"¼ Euro Rugby World Cup 2023 - England", 2023, "1-4eur-coin-england"},
		// The subject is outside the brackets, which only name the metal.
		{"10 euros Josephine Baker (argent 333‰)", 2025, "josephine-baker-10eur"},
	} {
		query, _ := mintQueryForTitle(c.title, &c.year)
		photo, err := service.Match(ctx, query)
		if err != nil || photo == nil || !strings.Contains(photo.PageURL, c.want) {
			t.Errorf("%s: photo = %+v, err = %v", c.title, photo, err)
		}
	}
	for _, c := range []struct {
		title string
		year  int
	}{
		// Sharing "palace" or "mont" with another coin is not enough.
		{"10 Euros (Palace of Versailles)", 2023},
		{"10 Euros (Mont-Saint-Michel)", 2020},
		// The France variant shares all its words with the series.
		{"¼ Euro Rugby World Cup 2023 - France", 2023},
		{"10 Euros (Astérix, partage)", 2022},
	} {
		query, _ := mintQueryForTitle(c.title, &c.year)
		if photo, err := service.Match(ctx, query); photo != nil || err != nil {
			t.Errorf("%s: expected no photo, got %+v (%v)", c.title, photo, err)
		}
	}
	if client.calls != 2 {
		t.Fatalf("the sitemaps should be read once, got %d calls", client.calls)
	}
}

// A letter opener holding a €2 coin is not a coin.
func TestMdpCoinEnd(t *testing.T) {
	for slug, want := range map[string]bool{
		"astrological-signs-aries-5eur-common-metal-coin-circulating-quality-yeardate-2025": true,
		"harry-potter-the-3-wizards-10eu-silver-coin-yeardate-2021-0":                       true,
		"hello-kitty-50eur-1-4-oz-gold-coin-birthday-proof-quality-yeardate-2024-jp":        true,
		"design-letter-opener-with-rodin-commemorativ-2eu-coin-yeardate-2017-steel-letter":  false,
	} {
		if got := mdpCoinEnd.MatchString(slug); got != want {
			t.Errorf("%s: %v, want %v", slug, got, want)
		}
	}
}

func TestMdpCoinPhotoService_ConfirmsTheGuessedReverse(t *testing.T) {
	service, client := newTestMdpService(t)
	ctx := context.Background()
	disc := syntheticPhoto(t, 200, 200, func(x, y int) bool { return (x-100)*(x-100)+(y-100)*(y-100) < 80*80 })
	box := syntheticPhoto(t, 200, 200, func(x, y int) bool { return x > 30 && x < 170 && y > 20 && y < 180 })
	year := 2025
	query, _ := mintQueryForTitle("5 euros Bélier", &year)

	client.images[mdpPhotos+"M/5/M5066_MDP_00066462_BD_b095.jpg"] = disc
	photo, _ := service.Match(ctx, query)
	if photo == nil || photo.Reverse != "" {
		t.Fatalf("an unchecked guess is not a reverse yet: %+v", photo)
	}
	if confirmed := service.Confirm(ctx, photo); confirmed.Reverse != mdpPhotos+"M/5/M5066_MDP_00066462_BD_b095.jpg" {
		t.Fatalf("confirmed = %+v", confirmed)
	}
	calls := client.calls
	service.Confirm(ctx, photo)
	if client.calls != calls {
		t.Fatal("a checked photo is not downloaded again")
	}

	// The packaging is not taken for the reverse.
	client.images[mdpPhotos+"M/5/M5066_MDP_00066462_BD_b095.jpg"] = box
	other, _ := newTestMdpService(t)
	other.client = client
	if confirmed := other.Confirm(ctx, photo); confirmed.Reverse != "" {
		t.Fatalf("a box is not a reverse: %+v", confirmed)
	}
}

func TestMdpCoinPhotoService_ThumbnailAndCountry(t *testing.T) {
	service, _ := newTestMdpService(t)
	if service.Country() != "France" || service.Source() != models.CoinPhotoSourceMDP ||
		service.Thumbnail(mdpPhotos+"a.jpg") != mdpPhotos+"a.jpg?optimize=medium&fit=bounds&height=400&width=400" {
		t.Fatal("unexpected source description")
	}
}
