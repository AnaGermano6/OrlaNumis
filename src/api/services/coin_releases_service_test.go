package services

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestReleasePeriod(t *testing.T) {
	for text, want := range map[string][2]string{
		"June 2025":              {"2025-06", "2025-06"},
		"1 July 2022":            {"2022-07", "2022-07"},
		"Fourth quarter of 2025": {"2025-10", "2025-12"},
		"First quarter 2025":     {"2025-01", "2025-03"},
		"Second half of 2025":    {"2025-07", "2025-12"},
		"Autumn 2025":            {"2025-09", "2025-11"},
		"January 2026":           {"2026-01", "2026-01"},
		"":                       {"2025-01", "2025-12"},
		"to be announced":        {"2025-01", "2025-12"},
	} {
		if from, to := releasePeriod(text, 2025); from != want[0] || to != want[1] {
			t.Errorf("releasePeriod(%q) = %s…%s, want %s…%s", text, from, to, want[0], want[1])
		}
	}
}

func TestEcbReleases_JointIssueOnceUpcomingFirst(t *testing.T) {
	photos := ParseEcbCommemoratives(ecb2022Page(t), 2022)
	releases := ecbReleases(photos, 2022, time.Date(2022, 10, 15, 0, 0, 0, 0, time.UTC))

	var joint, portugal []CoinRelease
	for _, release := range releases {
		if strings.Contains(release.Title, "Erasmus") {
			joint = append(joint, release)
		}
		if release.Country == "PT" {
			portugal = append(portugal, release)
		}
	}
	// The Erasmus coin of every euro country is listed once, with each country's image.
	if len(joint) != 1 || joint[0].Country != "" || len(joint[0].Keys) < 10 || joint[0].From != "2022-07" {
		t.Fatalf("joint issue = %+v", joint)
	}
	if len(portugal) != 1 || !strings.Contains(portugal[0].Title, "Gago Coutinho") || portugal[0].From != "2022-03" ||
		portugal[0].Volume != 1000000 || portugal[0].FaceValue != 2 || portugal[0].Upcoming || !strings.HasSuffix(portugal[0].PageURL, "comm_2022.en.html") {
		t.Fatalf("Portugal = %+v", portugal)
	}
	// In October, the coins of the fourth quarter are still to come, and listed first.
	if !releases[0].Upcoming || releases[0].To != "2022-12" {
		t.Fatalf("first release = %+v", releases[0])
	}
	for i := 1; i < len(releases); i++ {
		a, b := releases[i-1], releases[i]
		if !a.Upcoming && b.Upcoming || !a.Upcoming && !b.Upcoming && a.From < b.From {
			t.Fatalf("out of order: %+v before %+v", a, b)
		}
	}
}

func TestReleaseUpcoming(t *testing.T) {
	for _, c := range []struct {
		from, to, current string
		want              bool
	}{
		{"2026-11", "2026-11", "2026-10", true},  // next month
		{"2026-10", "2026-10", "2026-10", false}, // this month: out
		{"2026-10", "2026-12", "2026-11", true},  // a quarter under way
		{"2026-07", "2026-12", "2026-10", false}, // half a year under way
		{"2026-01", "2026-12", "2026-10", false}, // only the year, this year
		{"2027-01", "2027-12", "2026-10", true},  // only the year, next year
	} {
		if got := releaseUpcoming(c.from, c.to, c.current); got != c.want {
			t.Errorf("releaseUpcoming(%s, %s, %s) = %v, want %v", c.from, c.to, c.current, got, c.want)
		}
	}
}

type fakeMint struct {
	country, source string
	coins           []mintCoin
	err             error
}

func (m fakeMint) Coins(context.Context) ([]mintCoin, error) { return m.coins, m.err }
func (m fakeMint) Country() string                           { return m.country }
func (m fakeMint) Source() string                            { return m.source }
func (m fakeMint) Thumbnail(image string) string             { return image + "?width=400" }

func TestMintReleases_GroupsVariantsOfTheLastYear(t *testing.T) {
	coins := []mintCoin{
		{Title: "Rotary (BNC)", Value: 2, Year: 2026, Released: "2026-07", PageURL: "https://loja.incm.pt/products/rotary-bnc", Obverse: "https://cdn/rotary-bnc_A.jpg", Volume: 10000},
		{Title: "Rotary", Value: 2, Year: 2026, Released: "2026-07", PageURL: "https://loja.incm.pt/products/rotary", Obverse: "https://cdn/rotary_A.jpg", Volume: 500000},
		{Title: "Rotary (Proof)", Value: 2, Year: 2026, Released: "2026-07", PageURL: "https://loja.incm.pt/products/rotary-proof", Obverse: "https://cdn/rotary-proof_A.jpg", Volume: 5000},
		{Title: "FIFA World Cup (Ouro Proof)", Value: 5, Metal: "Ouro 999‰", Year: 2026, Released: "2026-05", PageURL: "https://loja.incm.pt/products/fifa-ouro", Obverse: "https://cdn/fifa-ouro_A.jpg"},
		{Title: "FIFA World Cup (Prata Proof)", Value: 7.5, Metal: "Prata 925‰", Year: 2026, Released: "2026-05", PageURL: "https://loja.incm.pt/products/fifa-prata", Obverse: "https://cdn/fifa-prata_A.jpg"},
		{Title: "Natal", Value: 5, Year: 2026, Released: "2026-12", PageURL: "https://loja.incm.pt/products/natal", Obverse: "https://cdn/natal_A.jpg"},
		{Title: "O Conhecimento", Value: 5, Year: 2024, Released: "2024-11", PageURL: "https://loja.incm.pt/products/conhecimento", Obverse: "https://cdn/conhecimento_A.jpg"},
	}
	releases := mintReleases(coins, fakeMint{country: "Portugal", source: "incm"}, time.Date(2026, 10, 13, 0, 0, 0, 0, time.UTC), true)

	titles := []string{}
	for _, release := range releases {
		titles = append(titles, release.Title)
	}
	// What is to come first; then the newest; a coin of 2024 is too old.
	if strings.Join(titles, " | ") != "Natal | Rotary | FIFA World Cup" {
		t.Fatalf("titles = %v", titles)
	}
	natal, rotary := releases[0], releases[1]
	if !natal.Upcoming || natal.Country != "PT" || natal.Source != "incm" || natal.From != "2026-12" {
		t.Fatalf("Natal = %+v", natal)
	}
	// One design of one face value, shown and opened by its plain edition.
	if rotary.Upcoming || strings.Join(rotary.Variants, ",") != "BNC,Proof" || rotary.PageURL != "https://loja.incm.pt/products/rotary" ||
		rotary.ImageURL != "https://cdn/rotary_A.jpg?width=400" || rotary.Volume != 515000 || len(rotary.Keys) != 3 {
		t.Fatalf("Rotary = %+v", rotary)
	}
	if rotary.FaceValue != 2 || rotary.FaceValues != nil {
		t.Fatalf("Rotary face values = %v, %v", rotary.FaceValue, rotary.FaceValues)
	}
	// The gold and silver coins of one design, of two face values, are one release.
	fifa := releases[2]
	if fifa.FaceValue != 5 || len(fifa.FaceValues) != 2 || fifa.FaceValues[1] != 7.5 || strings.Join(fifa.Variants, ",") != "Ouro Proof,Prata Proof" || len(fifa.Keys) != 2 {
		t.Fatalf("FIFA = %+v", fifa)
	}
	// Its metals, once each; the Rotary's mint does not say.
	if strings.Join(fifa.Metals, ",") != "gold,silver" || rotary.Metals != nil {
		t.Fatalf("metals = %v, %v", fifa.Metals, rotary.Metals)
	}
}

func TestReleaseMetal(t *testing.T) {
	for metal, want := range map[string]string{
		"Prata 925‰":       ReleaseMetalSilver,
		"Ouro 999‰":        ReleaseMetalGold,
		"Ouro nórdico":     ReleaseMetalBase,
		"gold":             ReleaseMetalGold,
		"silver":           ReleaseMetalSilver,
		"Or 999":           ReleaseMetalGold,
		"Argent":           ReleaseMetalSilver,
		"common metal":     ReleaseMetalBase,
		"copper":           ReleaseMetalBase,
		"Cuproníquel":      ReleaseMetalBase,
		"Cor-de-rosa":      ReleaseMetalBase,
		"Prata dourada":    ReleaseMetalSilver,
		"Silber vergoldet": ReleaseMetalSilver,
		"":                 "",
	} {
		if got := releaseMetal(metal); got != want {
			t.Errorf("releaseMetal(%q) = %q, want %q", metal, got, want)
		}
	}
}

func TestCoinReleases_GoldAndSpecialFinishCoinsApart(t *testing.T) {
	portugal := fakeMint{country: "Portugal", source: "incm", coins: []mintCoin{
		{Title: "FIFA World Cup (Ouro Proof)", Value: 5, Metal: "Ouro 999‰", Year: 2026, Released: "2026-05", PageURL: "https://loja.incm.pt/products/fifa-ouro", Obverse: "https://cdn/fifa-ouro_A.jpg"},
		{Title: "FIFA World Cup (Prata Proof)", Value: 7.5, Metal: "Prata 925‰", Year: 2026, Released: "2026-05", PageURL: "https://loja.incm.pt/products/fifa-prata", Obverse: "https://cdn/fifa-prata_A.jpg"},
		{Title: "FIFA World Cup (Prata FDC)", Value: 7.5, Metal: "Prata 500‰", Year: 2026, Released: "2026-05", PageURL: "https://loja.incm.pt/products/fifa-prata-fdc", Obverse: "https://cdn/fifa-prata-fdc_A.jpg"},
		{Title: "FIFA World Cup (Prata)", Value: 7.5, Metal: "Prata 500‰", Year: 2026, Released: "2026-05", PageURL: "https://loja.incm.pt/products/fifa-prata-normal", Obverse: "https://cdn/fifa-prata-normal_A.jpg"},
		{Title: "FIFA World Cup", Value: 7.5, Metal: "Cuproníquel", Year: 2026, Released: "2026-05", PageURL: "https://loja.incm.pt/products/fifa", Obverse: "https://cdn/fifa_A.jpg"},
	}}
	releases := NewCoinReleasesService(nil, portugal)
	releases.now = func() time.Time { return time.Date(2026, 10, 13, 0, 0, 0, 0, time.UTC) }

	got := []string{}
	for _, release := range releases.Releases(context.Background())[1].Releases {
		got = append(got, fmt.Sprintf("%s %v %g %v special=%v", release.Title, release.Metals, release.FaceValue, release.Variants, release.SpecialFinish))
	}
	// Its silver and base metal coins together; its proof and FDC coins
	// apart, and its gold one apart too.
	want := "FIFA World Cup [silver base] 7.5 [Prata] special=false | FIFA World Cup [silver] 7.5 [Prata Proof Prata FDC] special=true | FIFA World Cup [gold] 5 [Ouro Proof] special=true"
	if strings.Join(got, " | ") != want {
		t.Fatalf("releases = %s", strings.Join(got, " | "))
	}
}

func TestIsSpecialFinishCoin(t *testing.T) {
	for _, test := range []struct {
		coin mintCoin
		want bool
	}{
		{mintCoin{Title: "Amadeo de Souza-Cardoso (Prata Proof)"}, true},
		{mintCoin{Title: "150 Anos da Caixa Geral de Depósitos (Prata FDC)"}, true},
		{mintCoin{Title: "Multiculturalidade (BNC)"}, false},
		{mintCoin{Title: "Hello Kitty", PageURL: "https://www.monnaiedeparis.fr/en/hello-kitty-10eur-silver-coin-proof-quality-yeardate-2026"}, true},
		{mintCoin{Title: "Astrological signs", PageURL: "https://www.monnaiedeparis.fr/en/astrological-signs-aries-5eur-common-metal-coin-circulating-quality-yeardate-2025"}, false},
		{mintCoin{Title: "Edaphon", PageURL: "https://www.muenzeoesterreich.at/produkte/25-euro-silber-niob-muenze-edaphon-pp"}, true},
		{mintCoin{Title: "Edaphon", PageURL: "https://www.muenzeoesterreich.at/produkte/25-euro-silber-niob-muenze-edaphon-hgh"}, false},
	} {
		if got := isSpecialFinishCoin(test.coin); got != test.want {
			t.Errorf("isSpecialFinishCoin(%q, %q) = %v, want %v", test.coin.Title, test.coin.PageURL, got, test.want)
		}
	}
}

func TestWithMintTwoEuroCoins(t *testing.T) {
	release := func(source, country, title, from string, values ...float64) CoinRelease {
		coin := CoinRelease{Source: source, Country: country, Title: title, From: from, To: from, FaceValue: values[0]}
		if len(values) > 1 {
			coin.FaceValues = values
		}
		return coin
	}
	twoEuro := CoinReleaseGroup{Source: CoinReleaseSourceECB, Year: 2025, Releases: []CoinRelease{
		release("ecb", "PT", "World scouting", "2025-06", 2),
	}}
	mints := CoinReleaseGroup{Source: CoinReleaseGroupMints, Year: 2026, Releases: []CoinRelease{
		release("incm", "PT", "Escutismo mundial", "2025-06", 2),
		release("incm", "PT", "Rotary", "2026-07", 2),
		release("mdp", "FR", "Tour de France", "2026-06", 2),
		release("incm", "PT", "FIFA World Cup", "2026-05", 5, 7.5),
		release("incm", "PT", "Natal", "2026-12", 2, 5),
	}}
	twoEuro, mints = withMintTwoEuroCoins(twoEuro, mints)

	titles := func(group CoinReleaseGroup) string {
		var names []string
		for _, coin := range group.Releases {
			names = append(names, coin.Title)
		}
		return strings.Join(names, " | ")
	}
	// The mints' €2 coins the ECB does not list yet join its own, newest first;
	// Portugal's of 2025 is the ECB's already.
	if got := titles(twoEuro); got != "Rotary | Tour de France | World scouting" {
		t.Errorf("€2 coins = %s", got)
	}
	// A design of €2 and €5 is a collector coin; the collector coins keep their order.
	if got := titles(mints); got != "FIFA World Cup | Natal" {
		t.Errorf("collector coins = %s", got)
	}
}

func TestCoinReleases_EcbListsOfThisYearAndTheLast(t *testing.T) {
	ecb := newTestEcbService(t, ecbPagesByYear{2025: ecb2022Page(t), 2026: ecb2022Page(t)})
	releases := NewCoinReleasesService(ecb)
	releases.now = ecb.now
	both := releases.Releases(context.Background())[0]

	last := NewCoinReleasesService(newTestEcbService(t, ecbPagesByYear{2025: ecb2022Page(t)}))
	last.now = ecb.now
	one := last.Releases(context.Background())[0]
	if both.Year != 2026 || one.Year != 2025 || len(one.Releases) == 0 || len(both.Releases) != 2*len(one.Releases) {
		t.Fatalf("both lists: year %d, %d coins; last year's: year %d, %d coins", both.Year, len(both.Releases), one.Year, len(one.Releases))
	}
	for _, coin := range both.Releases {
		if strings.Join(coin.Metals, ",") != ReleaseMetalBase {
			t.Fatalf("%s metals = %v", coin.Title, coin.Metals)
		}
	}
}

type ecbPagesByYear map[int][]byte

func (p ecbPagesByYear) CommemorativePage(_ context.Context, year int) ([]byte, error) {
	if page, ok := p[year]; ok {
		return page, nil
	}
	return nil, errors.New("ecb page: status 404")
}

func TestMintReleases_ByCountryAndYear(t *testing.T) {
	france := fakeMint{country: "France", source: "mdp", coins: []mintCoin{
		{Title: "Year of the horse", DesignKey: "year horse", Value: 50, Year: 2026, PageURL: "https://www.monnaiedeparis.fr/en/horse-50", Obverse: "https://mdp/horse-50.jpg"},
		{Title: "Year of the horse", DesignKey: "year horse silver", Value: 20, Year: 2026, PageURL: "https://www.monnaiedeparis.fr/en/horse-20", Obverse: "https://mdp/horse-20.jpg"},
		{Title: "Year of the goat", DesignKey: "year goat", Value: 20, Year: 2027, PageURL: "https://www.monnaiedeparis.fr/en/goat-20", Obverse: "https://mdp/goat-20.jpg"},
		{Title: "Year of the snake", DesignKey: "year snake", Value: 20, Year: 2025, PageURL: "https://www.monnaiedeparis.fr/en/snake-20", Obverse: "https://mdp/snake-20.jpg"},
	}}
	portugal := fakeMint{country: "Portugal", source: "incm", coins: []mintCoin{
		{Title: "Rotary", Value: 2, Year: 2026, Released: "2026-07", PageURL: "https://loja.incm.pt/products/rotary", Obverse: "https://cdn/rotary_A.jpg"},
	}}
	austria := fakeMint{country: "Austria", source: "moe", err: errors.New("offline")}
	releases := NewCoinReleasesService(nil, portugal, france, austria)
	releases.now = func() time.Time { return time.Date(2026, 10, 13, 0, 0, 0, 0, time.UTC) }

	groups := releases.Releases(context.Background())
	if len(groups) != 2 || groups[1].Source != CoinReleaseGroupMints || groups[1].Year != 2026 {
		t.Fatalf("groups = %+v", groups)
	}
	// The €2 coin is among the €2 coins, the ECB's list being unread.
	if len(groups[0].Releases) != 1 || groups[0].Releases[0].Title != "Rotary" {
		t.Fatalf("€2 coins = %+v", groups[0].Releases)
	}
	got := []string{}
	for _, release := range groups[1].Releases {
		got = append(got, fmt.Sprintf("%s %s %g %s-%s %v", release.Country, release.Title, release.FaceValue, release.From, release.To, release.Upcoming))
	}
	// Next year's coin is to come; a coin dated by its year only comes after
	// the ones with a month, its face values together; last year's French
	// coin and the unreadable Austrian catalogue add nothing.
	want := []string{
		"FR Year of the goat 20 2027-01-2027-12 true",
		"FR Year of the horse 20 2026-01-2026-12 false",
	}
	if strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Fatalf("mint releases:\n%s\nwant:\n%s", strings.Join(got, "\n"), strings.Join(want, "\n"))
	}
}

func TestCoinReleases_LatestEcbYearAndOwnedCoins(t *testing.T) {
	// The ECB has not opened the current year's page: the last year's list is shown.
	ecb := newTestEcbService(t, ecbPagesByYear{2025: ecb2022Page(t)})
	releases := NewCoinReleasesService(ecb)
	releases.now = ecb.now
	groups := releases.Releases(context.Background())
	if len(groups) != 2 || groups[0].Source != CoinReleaseSourceECB || groups[0].Year != 2025 || len(groups[0].Releases) == 0 {
		t.Fatalf("groups = %+v", groups)
	}
	if groups[1].Source != CoinReleaseGroupMints || len(groups[1].Releases) != 0 {
		t.Fatalf("no mints = %+v", groups[1])
	}

	// A collection coin that is Portugal's coin of that list marks it as owned; a €1 coin is not looked at.
	photos := NewCoinPhotoSearchService(nil, nil, nil).WithECB(ecb)
	releases.MarkOwned(context.Background(), groups, []CoinPhotoQuery{
		{Year: 2025, Country: "Portugal", Denomination: "2 euro", Subject: "Travessia aérea do Atlântico Sul por Gago Coutinho e Sacadura Cabral"},
		{Year: 2025, Country: "Portugal", Denomination: "1 euro"},
	}, photos)
	owned := []string{}
	for _, release := range groups[0].Releases {
		if release.Owned {
			owned = append(owned, release.Title)
		}
	}
	if len(owned) != 1 || !strings.Contains(owned[0], "Gago Coutinho") {
		t.Fatalf("owned = %v", owned)
	}
	// Each call gets its own copy: marking one does not mark the next.
	for _, release := range releases.Releases(context.Background())[0].Releases {
		if release.Owned {
			t.Fatalf("a later call sees %q as owned", release.Title)
		}
	}
}

func TestCoinCatalogue_EveryYearAndEachCountrysJointIssue(t *testing.T) {
	ecb := newTestEcbService(t, ecbPagesByYear{2022: ecb2022Page(t)})
	portugal := fakeMint{country: "Portugal", source: "incm", coins: []mintCoin{
		{Title: "D. Catarina de Bragança", Value: 5, Year: 2016, PageURL: "https://loja.incm.pt/products/catarina", Obverse: "https://cdn/catarina-1.png"},
		{Title: "Rotary (BNC)", Value: 2, Year: 2026, Released: "2026-07", PageURL: "https://loja.incm.pt/products/rotary-bnc", Obverse: "https://cdn/rotary_A.jpg"},
		// The proof version of the ECB's coin of 2022: listed once, by the ECB.
		{Title: "Gago Coutinho e Sacadura Cabral (Proof)", Value: 2, Year: 2022, Released: "2022-03", PageURL: "https://loja.incm.pt/products/gago-proof", Obverse: "https://cdn/gago_A.jpg"},
	}}
	releases := NewCoinReleasesService(ecb, portugal)
	releases.now = ecb.now

	catalogue := releases.Catalogue(context.Background())
	byCountry := map[string][]CoinRelease{}
	for _, coin := range catalogue {
		if coin.Country == "" {
			t.Fatalf("a coin without a country: %+v", coin)
		}
		byCountry[coin.Country] = append(byCountry[coin.Country], coin)
	}
	// Portugal's own €2 commemorative and its coin of the joint Erasmus issue,
	// with the mint's coins of every year, newest first; the mint's €2 coin of
	// a year the ECB lists is not listed again, one of a year it does not is.
	titles := []string{}
	for _, coin := range byCountry["PT"] {
		titles = append(titles, fmt.Sprintf("%s %s", coin.From[:4], coin.Title))
	}
	if len(titles) != 4 || titles[0] != "2026 Rotary" || !strings.Contains(titles[1], "Erasmus") || !strings.Contains(titles[2], "Gago Coutinho") || titles[3] != "2016 D. Catarina de Bragança" {
		t.Fatalf("Portugal = %v", titles)
	}
	if len(byCountry["AT"]) == 0 || len(byCountry["DE"]) == 0 {
		t.Fatalf("Austria %d, Germany %d coins", len(byCountry["AT"]), len(byCountry["DE"]))
	}
	// Each call gets its own copy.
	catalogue[0].Owned = true
	if releases.Catalogue(context.Background())[0].Owned {
		t.Fatal("a later call sees a coin marked by an earlier one")
	}
}

func TestCatalogueCountryNames(t *testing.T) {
	if CatalogueCountryName("pt") != "Portugal" || CatalogueCountryName("VA") != "Vatican City" || CatalogueCountryName("GB") != "" {
		t.Fatal("CatalogueCountryName")
	}
	if CatalogueCountryCode("Alemanha") != "DE" || CatalogueCountryCode("Cidade do Vaticano") != "VA" || CatalogueCountryCode("Reino Unido") != "" {
		t.Fatal("CatalogueCountryCode")
	}
}

func TestJointIssues_EachCountrysVersionOfTheSameCoin(t *testing.T) {
	ecb := newTestEcbService(t, ecbPagesByYear{2022: ecb2022Page(t)})
	releases := NewCoinReleasesService(ecb)
	releases.now = ecb.now

	issues := releases.JointIssues(context.Background())
	if len(issues) != 1 || issues[0].Year != 2022 || issues[0].Title != "35 years of the Erasmus programme" || len(issues[0].Coins) != 19 {
		t.Fatalf("issues = %+v", issues)
	}
	// Germany's image, whose file does not name it, is the one member state missing.
	if germany := issues[0].Coins[3]; germany.Country != "DE" || !strings.Contains(germany.ImageURL, "EU1-erasmus") {
		t.Fatalf("Germany = %+v", germany)
	}
	if releases.JointIssue(context.Background(), 2021) != nil {
		t.Fatal("2021 had no joint issue")
	}

	// The catalogue lists Germany's Erasmus coin too.
	found := false
	for _, coin := range releases.Catalogue(context.Background()) {
		found = found || coin.Country == "DE" && coin.Title == "35 years of the Erasmus programme"
	}
	if !found {
		t.Fatal("Germany's Erasmus coin is not in the catalogue")
	}
}

func TestWithJointIssueCountries(t *testing.T) {
	members := []string{"Austria", "Belgium", "Cyprus", "Estonia", "Finland", "France", "Germany", "Greece", "Ireland", "Italy", "Latvia", "Lithuania", "Luxembourg", "Malta", "Netherlands", "Portugal", "Slovakia", "Slovenia", "Spain"}
	issue := func(year int, without []string, images ...string) []EcbCoinPhoto {
		skip := map[string]bool{}
		for _, country := range without {
			skip[country] = true
		}
		photos := []EcbCoinPhoto{{Year: year, Country: "Portugal", Feature: "Portugal's own", ImageURL: "https://ecb/own.jpg"}}
		for _, country := range members {
			if !skip[country] {
				photos = append(photos, EcbCoinPhoto{Year: year, Country: country, Feature: "Flag", ImageURL: "https://ecb/" + country + ".jpg"})
			}
		}
		for _, image := range images {
			photos = append(photos, EcbCoinPhoto{Year: year, Feature: "Flag", ImageURL: image})
		}
		return photos
	}
	countryOf := func(photos []EcbCoinPhoto, image string) string {
		for _, photo := range photos {
			if photo.ImageURL == image {
				return photo.Country
			}
		}
		return "not listed"
	}

	// Named in the country's own language.
	photos := withJointIssueCountries(issue(2015, []string{"Netherlands"}, "https://ecb/joint_comm_2015_Nederland.jpg"), 2015)
	if got := countryOf(photos, "https://ecb/joint_comm_2015_Nederland.jpg"); got != "Netherlands" {
		t.Fatalf("Nederland = %q", got)
	}
	// Not named, and one member state missing: that one.
	photos = withJointIssueCountries(issue(2022, []string{"Germany"}, "https://ecb/EU1.jpg"), 2022)
	if got := countryOf(photos, "https://ecb/EU1.jpg"); got != "Germany" {
		t.Fatalf("EU1 = %q", got)
	}
	// Two missing, two unnamed: no guess.
	photos = withJointIssueCountries(issue(2022, []string{"Germany", "Spain"}, "https://ecb/EU1.jpg", "https://ecb/EU2.jpg"), 2022)
	if countryOf(photos, "https://ecb/EU1.jpg") != "" || countryOf(photos, "https://ecb/EU2.jpg") != "" {
		t.Fatalf("guessed = %+v", photos)
	}
	// Lithuania was not a member in 2014: Germany is the one missing.
	photos = withJointIssueCountries(issue(2014, []string{"Germany", "Lithuania"}, "https://ecb/EU1.jpg"), 2014)
	if got := countryOf(photos, "https://ecb/EU1.jpg"); got != "Germany" {
		t.Fatalf("2014 EU1 = %q", got)
	}
}

func TestMarkWishlistReleases(t *testing.T) {
	releases := []CoinRelease{
		{Title: "Wanted", Keys: []string{"https://ecb/a.jpg"}},
		{Title: "Had", Keys: []string{"https://ecb/b.jpg"}, Owned: true},
		{Title: "Other", Keys: []string{"numista:7"}},
	}
	MarkWishlistReleases(releases, map[string]uint{"https://ecb/a.jpg": 4})
	if !releases[0].Wishlist || releases[1].Wishlist || releases[2].Wishlist || releases[0].Owned {
		t.Fatalf("releases = %+v", releases)
	}
	// The wishlist coin comes with the mark, to take it off the wishlist.
	if releases[0].WishlistCoinID != 4 || releases[1].WishlistCoinID != 0 {
		t.Fatalf("wishlist coins = %d, %d", releases[0].WishlistCoinID, releases[1].WishlistCoinID)
	}
	// Marking ownership leaves the wishlist mark alone, and the other way round.
	MarkOwnedReleases(releases, map[string]uint{"numista:7": 9})
	if !releases[0].Wishlist || !releases[1].Owned || !releases[2].Owned {
		t.Fatalf("after ownership = %+v", releases)
	}
}
