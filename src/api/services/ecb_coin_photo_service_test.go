package services

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/briandenicola/ancient-coins-api/models"
	"github.com/briandenicola/ancient-coins-api/repository"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

type fakeECBCoinsClient struct {
	page  []byte
	err   error
	calls int
}

func (f *fakeECBCoinsClient) CommemorativePage(context.Context, int) ([]byte, error) {
	f.calls++
	return f.page, f.err
}

func ecb2022Page(t *testing.T) []byte {
	t.Helper()
	page, err := os.ReadFile("testdata/ecb_comm_2022.html")
	if err != nil {
		t.Fatal(err)
	}
	return page
}

func newTestEcbService(t *testing.T, client ECBCoinsClient) *EcbCoinPhotoService {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&models.NumistaCacheEntry{}); err != nil {
		t.Fatal(err)
	}
	svc := NewEcbCoinPhotoService(client, NewNumistaResponseCache(repository.NewNumistaCacheRepository(db), nil, nil))
	svc.now = func() time.Time { return time.Date(2026, 10, 10, 0, 0, 0, 0, time.UTC) }
	return svc
}

func TestParseEcbCommemorativesReadsCoinsAndJointIssues(t *testing.T) {
	photos := ParseEcbCommemoratives(ecb2022Page(t), 2022)
	byCountry := map[string][]EcbCoinPhoto{}
	for _, photo := range photos {
		byCountry[photo.Country] = append(byCountry[photo.Country], photo)
	}

	portugal := byCountry["Portugal"]
	if len(portugal) != 2 {
		t.Fatalf("Portugal 2022 = %+v", portugal)
	}
	if portugal[0].Feature != "35 years of the Erasmus programme" || portugal[0].ImageURL != "https://www.ecb.europa.eu/euro/coins/comm/html/comm_2022/Portugal.jpg" {
		t.Fatalf("joint issue = %+v", portugal[0])
	}
	if portugal[1].ImageURL != "https://www.ecb.europa.eu/euro/coins/comm/html/comm_2022/0421-23r.jpg" || portugal[1].Year != 2022 {
		t.Fatalf("national coin = %+v", portugal[1])
	}
	// "Austria1.jpg" in the joint issue, a heading wrapped in <p>, the German Länder series.
	if len(byCountry["Austria"]) != 1 || len(byCountry["Malta"]) != 3 || byCountry["Germany"][0].Feature != "Bundeslander series – Thuringia" {
		t.Fatalf("Austria %+v, Malta %+v, Germany %+v", byCountry["Austria"], byCountry["Malta"], byCountry["Germany"])
	}
	if len(byCountry["Greece"]) != 2 || byCountry["Greece"][0].Feature != "200 years from the first Greek Constitution" {
		t.Fatalf("Greece = %+v", byCountry["Greece"])
	}
}

func TestEcbCountryName(t *testing.T) {
	for name, want := range map[string]string{
		"Portugal": "Portugal", "Alemanha, República Federal da": "Germany", "Países Baixos": "Netherlands", "the Netherlands": "Netherlands",
		"pays-bas": "Netherlands", "Eslovénia ": "Slovenia", "Vatican": "Vatican City", "Reino Unido": "", "": "",
	} {
		if got := EcbCountryName(name); got != want {
			t.Errorf("EcbCountryName(%q) = %q, want %q", name, got, want)
		}
	}
}

func TestEcbMatchFindsTheCoinByTitleOrSubject(t *testing.T) {
	client := &fakeECBCoinsClient{page: ecb2022Page(t)}
	svc := newTestEcbService(t, client)
	ctx := context.Background()

	// A Numista title in English matches the ECB feature.
	got, err := svc.Match(ctx, EcbPhotoQuery{Year: 2022, Country: "Portugal", Title: "2 Euros (Crossing of the South Atlantic by plane)"})
	if err != nil || len(got) != 1 || got[0].Match != models.CoinPhotoMatchExact || got[0].ImageURL != "https://www.ecb.europa.eu/euro/coins/comm/html/comm_2022/0421-23r.jpg" ||
		got[0].Source != models.CoinPhotoSourceECB || got[0].Side != "obverse" || got[0].Author != "European Central Bank" {
		t.Fatalf("South Atlantic = %+v, %v", got, err)
	}
	got, _ = svc.Match(ctx, EcbPhotoQuery{Year: 2022, Country: "Portugal", Title: "2 Euros (Erasmus Programme)"})
	if len(got) != 1 || got[0].Match != models.CoinPhotoMatchExact || got[0].Title != "35 years of the Erasmus programme" {
		t.Fatalf("Erasmus = %+v", got)
	}
	// A Portuguese subject that shares no word is not guessed: both are offered.
	got, _ = svc.Match(ctx, EcbPhotoQuery{Year: 2022, Country: "Portugal", Subject: "Travessia aérea"})
	if len(got) != 2 || got[0].Match != "" || got[1].Match != "" {
		t.Fatalf("ambiguous = %+v", got)
	}
	got, _ = svc.Match(ctx, EcbPhotoQuery{Year: 2022, Country: "Alemanha", Title: `2 Euros (Bundesländer - "Thüringen")`})
	if len(got) != 1 || got[0].Match != models.CoinPhotoMatchExact {
		t.Fatalf("Germany = %+v", got)
	}
	// Germany's Erasmus coin is in the joint issue under an image that does not
	// name its country: it must not get the image of Germany's other 2022 coin.
	got, _ = svc.Match(ctx, EcbPhotoQuery{Year: 2022, Country: "Alemanha", Subject: "35º Aniversário - Programa Erasmus"})
	if len(got) != 1 || got[0].Match != "" {
		t.Fatalf("German Erasmus = %+v", got)
	}
	// Nor is a subject with no word in common guessed while such an image exists.
	got, _ = svc.Match(ctx, EcbPhotoQuery{Year: 2022, Country: "Alemanha", Subject: "Turíngia, Castelo de Wartburg em Eisenach"})
	if len(got) != 1 || got[0].Match != "" {
		t.Fatalf("Germany by Portuguese subject = %+v", got)
	}
	// Non-euro countries and years without commemoratives return nothing, without a request.
	calls := client.calls
	for _, query := range []EcbPhotoQuery{{Year: 2022, Country: "Reino Unido"}, {Year: 2003, Country: "Portugal"}, {Year: 2031, Country: "Portugal"}} {
		if got, err := svc.Match(ctx, query); err != nil || len(got) != 0 {
			t.Fatalf("%+v = %+v, %v", query, got, err)
		}
	}
	if client.calls != calls {
		t.Fatalf("pages are cached and only fetched when needed: %d calls", client.calls)
	}
	if client.calls != 1 {
		t.Fatalf("the 2022 page should be fetched once, got %d", client.calls)
	}
}

func TestEcbMatchReportsAnUnreachablePage(t *testing.T) {
	svc := newTestEcbService(t, &fakeECBCoinsClient{err: errors.New("offline")})
	if _, err := svc.Match(context.Background(), EcbPhotoQuery{Year: 2022, Country: "Portugal"}); err == nil {
		t.Fatal("expected an error")
	}
}

// Older pages name images "joint_comm_2007_Portugal.jpg" outside a year folder,
// and some write "Feature:&nbsp;".
func TestParseEcbCommemorativesOlderMarkup(t *testing.T) {
	page := []byte(`<div class="box"><img src="comm_2007_Portugal.jpg"><div class="content-box"><h3>Portugal</h3>` +
		`<div><p><strong>Feature:&nbsp;</strong>Portuguese Presidency of the European Union</p></div></div></div>` +
		`<div class="box" id="joint"><img src="joint_comm_2007_Portugal.jpg"><img src="joint_comm_2007_San_Marino.jpg">` +
		`<img src="joint_comm_2007_Portugal_Face.jpg"><img src="EU-joint.jpg">` +
		`<div class="content-box"><h3>Euro area countries</h3><div><p><strong>Feature:</strong> 50th anniversary of the Treaty of Rome</p></div></div></div>`)
	photos := ParseEcbCommemoratives(page, 2007)
	got := map[string]string{}
	for _, photo := range photos {
		got[photo.Country+"|"+photo.Feature] = photo.ImageURL
	}
	base := "https://www.ecb.europa.eu/euro/coins/comm/html/"
	want := map[string]string{
		"Portugal|Portuguese Presidency of the European Union": base + "comm_2007_Portugal.jpg",
		"Portugal|50th anniversary of the Treaty of Rome":      base + "joint_comm_2007_Portugal.jpg",
		"San Marino|50th anniversary of the Treaty of Rome":    base + "joint_comm_2007_San_Marino.jpg",
		"|50th anniversary of the Treaty of Rome":              base + "EU-joint.jpg",
	}
	if len(photos) != len(want) {
		t.Fatalf("photos = %+v", photos)
	}
	for key, url := range want {
		if got[key] != url {
			t.Errorf("%s = %q, want %q", key, got[key], url)
		}
	}

	svc := newTestEcbService(t, &fakeECBCoinsClient{page: page})
	treaty, _ := svc.Match(context.Background(), EcbPhotoQuery{Year: 2007, Country: "Portugal", Title: "2 Euros (50th Anniversary of the Treaty of Rome)"})
	if len(treaty) != 1 || treaty[0].ImageURL != base+"joint_comm_2007_Portugal.jpg" || treaty[0].Match != models.CoinPhotoMatchExact {
		t.Fatalf("Treaty of Rome = %+v", treaty)
	}
}

// The coin's year is no evidence: "Paris 2024 - Torre Eiffel" is not the 2024
// wrestling coin just because both say 2024.
func TestEcbMatchIgnoresTheCoinsYear(t *testing.T) {
	page := []byte(`<div class="box"><img src="eiffel.jpg"><div class="content-box"><h3>France</h3><div><p><strong>Feature:</strong> The Eiffel Tower</p></div></div></div>` +
		`<div class="box"><img src="lutte.jpg"><div class="content-box"><h3>France</h3><div><p><strong>Feature:</strong> Jeux Olympiques 2024 - Lutte</p></div></div></div>`)
	svc := newTestEcbService(t, &fakeECBCoinsClient{page: page})
	got, _ := svc.Match(context.Background(), EcbPhotoQuery{Year: 2024, Country: "França", Subject: "XXXIII Jogos Olímpicos, Paris 2024 - Torre Eiffel"})
	if len(got) != 1 || got[0].Title != "The Eiffel Tower" || got[0].Match != models.CoinPhotoMatchExact {
		t.Fatalf("Eiffel = %+v", got)
	}
}
