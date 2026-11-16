package services

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/briandenicola/ancient-coins-api/models"
	"github.com/briandenicola/ancient-coins-api/repository"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

type fakeMDPShopClient struct {
	sitemaps map[string][]byte
	products map[string][]byte
	images   map[string][]byte
	calls    int
	// read are the product pages asked for.
	read []string
	err  error
}

func (f *fakeMDPShopClient) Product(_ context.Context, url string) ([]byte, error) {
	f.calls++
	f.read = append(f.read, url)
	if page, ok := f.products[url]; ok {
		return page, nil
	}
	return nil, errors.New("no such product")
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
	client := &fakeMDPShopClient{sitemaps: map[string][]byte{}, products: map[string][]byte{}, images: map[string][]byte{}}
	// The sitemaps end in .xml.txt: Unraid's Community Applications reads
	// every .xml file of the repository as an app template.
	for _, lang := range []string{"en", "fr"} {
		body, err := os.ReadFile("testdata/mdp_sitemap_" + lang + ".xml.txt")
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
	en, _ := os.ReadFile("testdata/mdp_sitemap_en.xml.txt")
	fr, _ := os.ReadFile("testdata/mdp_sitemap_fr.xml.txt")
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
	// Each version's metal and finish, from its name.
	variants := map[string]string{}
	for _, coin := range coins {
		variants[mdpSlug(coin.PageURL)] = coin.Variant
	}
	for slug, want := range map[string]string{
		"astrological-signs-aries-5eur-common-metal-coin-circulating-quality-yeardate-2025": "",
		"josephine-baker-10eur-silver-coin-historic-quality-yeardate-2025":                  "Prata",
		"year-of-the-snake-10eur-silver-coin-proof-quality-yeardate-2025":                   "Prata Proof",
	} {
		if got, ok := variants[slug]; !ok || got != want {
			t.Errorf("%s variant = %q (found %v), want %q", slug, got, ok, want)
		}
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

// A product page as the shop writes it: its own product among others in
// the scripts, and its details table.
const mdpTennisPage = `<script>{"handles":["catalog_product_view","catalog_product_view_sku_10041403240000"]}</script>
<script type="text/x-magento-init">{"*":{"relatedProducts":[{"date_lancement":"2024-01-15 00:00:00","ean128":"10041400000000","tirage":"500"},
{"date_lancement":"2026-02-24 00:00:00","designation_bou":"26 L'EQUIPE TENNIS 10\u20ac AG BE","ean128":"10041403240000","tirage":"3000"}]}}</script>
<table><tr><th class="col label" scope="row">Metal</th><td class="col data" data-th="Metal">SILVER 999‰</td></tr>
<tr><th class="col label" scope="row">Mintage</th><td class="col data" data-th="Mintage">3000</td></tr>
<tr><th class="col label" scope="row">Millésime</th><td class="col data" data-th="Millésime">2026</td></tr></table>`

// A coin still to come: its date in words, its year where its mintage goes.
const mdpAquariusPage = `<script>{"handles":["catalog_product_view_sku_10041405320000"]}</script>
<p class="available">Available on the 02/12/2026</p>
<table><tr><th class="col label" scope="row">Mintage</th><td class="col data" data-th="Mintage">2027</td></tr>
<tr><th class="col label" scope="row">Millésime</th><td class="col data" data-th="Millésime">2027</td></tr></table>`

func TestParseMdpProduct(t *testing.T) {
	for name, test := range map[string]struct {
		page string
		want mdpProductDetails
	}{
		"launched":   {mdpTennisPage, mdpProductDetails{Released: "2026-02", Volume: 3000}},
		"to come":    {mdpAquariusPage, mdpProductDetails{Released: "2026-12"}},
		"in French":  {`<p>Disponible le 15/03/2027</p><table><tr><th class="col label" scope="row">Tirage</th><td class="col data">1 000</td></tr></table>`, mdpProductDetails{Released: "2027-03", Volume: 1000}},
		"not a page": {"", mdpProductDetails{}},
	} {
		if got := ParseMdpProduct([]byte(test.page)); got != test.want {
			t.Errorf("%s: %+v, want %+v", name, got, test.want)
		}
	}
}

func TestMdpCoinPhotoService_ReadsProductPagesInTheBackground(t *testing.T) {
	service, client := newTestMdpService(t)
	service.now = func() time.Time { return time.Date(2025, 10, 1, 0, 0, 0, 0, time.UTC) }
	service.productDelay = 0
	snake := mdpShopURL + "/en/year-of-the-snake-10eur-silver-coin-proof-quality-yeardate-2025"
	client.products[snake] = []byte(mdpTennisPage)

	coins, err := service.Coins(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	for _, coin := range coins {
		if coin.Released != "" || coin.Volume != 0 {
			t.Fatalf("%s read before its page: %+v", coin.Title, coin)
		}
	}
	service.reads.Wait()
	// Only this year's coins' pages are read.
	if len(client.read) != 4 {
		t.Fatalf("pages read = %v", client.read)
	}

	coins, err = service.Coins(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	for _, coin := range coins {
		want := mdpProductDetails{}
		if coin.PageURL == snake {
			want = mdpProductDetails{Released: "2026-02", Volume: 3000}
		}
		if got := (mdpProductDetails{Released: coin.Released, Volume: coin.Volume}); got != want {
			t.Errorf("%s: %+v, want %+v", coin.PageURL, got, want)
		}
	}
	// What the pages said is kept: they are not read again.
	if service.reads.Wait(); len(client.read) != 4 {
		t.Fatalf("pages read again: %v", client.read)
	}
}
