package services

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/briandenicola/ancient-coins-api/models"
	"github.com/briandenicola/ancient-coins-api/repository"
	"gorm.io/gorm"
)

func TestParseChecklistDefinition(t *testing.T) {
	for _, c := range []struct {
		raw  map[string]interface{}
		want string
	}{
		// The country and currency are read in any case; what a catalogue
		// does not use is dropped.
		{map[string]interface{}{"source": "catalogue", "country": " pt ", "currency": "pte", "year": 2020, "challenge": "annual-series"}, "catalogue PT PTE 0  0"},
		{map[string]interface{}{"source": "catalogue", "country": "FR", "faceValue": 10}, "catalogue FR  10  0"},
		{map[string]interface{}{"source": "challenge", "challenge": "euro-countries", "country": "PT", "faceValue": 2}, "challenge   0 euro-countries 0"},
		{map[string]interface{}{"source": "challenge", "challenge": "joint-issue", "year": 2007, "country": "PT"}, "challenge   0 joint-issue 2007"},
		{map[string]interface{}{"source": "challenge", "challenge": "joint-issue", "year": 2003}, "error"},
		{map[string]interface{}{"source": "challenge", "challenge": "joint-issue"}, "error"},
		{map[string]interface{}{"source": "challenge", "challenge": "commemorative-per-country"}, "error"},
		{map[string]interface{}{"source": "challenge", "challenge": "annual-series", "country": "de", "year": 2002}, "challenge DE  0 annual-series 2002"},
		{map[string]interface{}{"source": "catalogue", "country": "US"}, "error"},
		{map[string]interface{}{"source": "catalogue", "country": "FR", "currency": "PTE"}, "error"},
		{map[string]interface{}{"source": "catalogue", "country": "PT", "currency": "USD"}, "error"},
		{map[string]interface{}{"source": "catalogue", "country": "PT", "faceValue": -2}, "error"},
		{map[string]interface{}{"source": "challenge", "challenge": "annual-series", "country": "PT", "year": 1990}, "error"},
		{map[string]interface{}{"source": "challenge", "challenge": "annual-series", "year": 2002}, "error"},
		{map[string]interface{}{"source": "challenge", "challenge": "every-coin"}, "error"},
		{map[string]interface{}{"source": "wishlist"}, "error"},
		{map[string]interface{}{"source": "catalogue", "country": 7}, "error"},
	} {
		definition, err := ParseChecklistDefinition(c.raw)
		got := "error"
		if err == nil {
			got = fmt.Sprintf("%s %s %s %g %s %d", definition.Source, definition.Country, definition.Currency, definition.FaceValue, definition.Challenge, definition.Year)
		}
		if got != c.want {
			t.Errorf("ParseChecklistDefinition(%v) = %q (%v), want %q", c.raw, got, err, c.want)
		}
	}
	definition, _ := ParseChecklistDefinition(map[string]interface{}{"source": "catalogue", "country": "PT", "faceValue": 2})
	if again, err := ParseChecklistDefinition(map[string]interface{}(definition.JSON())); err != nil || again != definition {
		t.Fatalf("stored definition read back as %+v, %v", again, err)
	}
}

func TestSummarize(t *testing.T) {
	value := 4.0
	slots := []ChecklistSlot{
		{Key: "a", Owned: true, CoinID: 1, CoinValue: &value},
		// One coin filling two slots counts once.
		{Key: "b", Owned: true, CoinID: 1, CoinValue: &value},
		{Key: "c", Owned: true, CoinID: 2},
		{Key: "d"},
	}
	summary := Summarize(slots)
	if summary.Total != 4 || summary.Owned != 3 || summary.TotalValue != 4 || summary.Completion() != 75 {
		t.Fatalf("summary = %+v, %v%%", summary, summary.Completion())
	}
	if (ChecklistSummary{}).Completion() != 0 {
		t.Fatal("an empty checklist is not 0% complete")
	}
}

// growingMint is Portugal's mint, whose shop gains coins.
type growingMint struct{ coins []mintCoin }

func (m *growingMint) Coins(context.Context) ([]mintCoin, error) { return m.coins, nil }
func (m *growingMint) Country() string                           { return "Portugal" }
func (m *growingMint) Source() string                            { return "incm" }
func (m *growingMint) Thumbnail(image string) string             { return image }

// checklistTest has the ECB's coins of 2022 and a mint selling one
// collector coin, a collection, and a clock.
type checklistTest struct {
	db        *gorm.DB
	mint      *growingMint
	releases  *CoinReleasesService
	checklist *SetChecklistService
	clock     time.Time
}

const gagoCoutinhoImage = "https://www.ecb.europa.eu/euro/coins/comm/html/comm_2022/0421-23r.jpg"

func newChecklistTest(t *testing.T) *checklistTest {
	t.Helper()
	db := setupTestDB(t)
	if err := db.AutoMigrate(&models.CoinSetTarget{}, &models.CoinSetChecklistKey{}, &models.SmartCriteriaTemplate{}, &models.Notification{}); err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&models.User{ID: 1, Username: "collector", Email: "collector@example.test", PasswordHash: "x"}).Error; err != nil {
		t.Fatal(err)
	}
	ecb := newTestEcbService(t, ecbPagesByYear{2022: ecb2022Page(t)})
	test := &checklistTest{
		db: db,
		mint: &growingMint{coins: []mintCoin{
			{Title: "D. Catarina de Bragança", Value: 5, Year: 2016, PageURL: "https://loja.incm.pt/products/catarina", Obverse: "https://cdn/catarina.png"},
		}},
		clock: time.Date(2026, 10, 13, 0, 0, 0, 0, time.UTC),
	}
	test.releases = NewCoinReleasesService(ecb, test.mint)
	test.releases.now = func() time.Time { return test.clock }
	photos := NewCoinPhotoSearchService(nil, nil, nil).WithECB(ecb)
	test.checklist = NewSetChecklistService(test.releases, repository.NewCoinRepository(db), photos)
	test.checklist.now = func() time.Time { return test.clock }
	return test
}

// coin adds a coin to user 1's collection, with obverse photos.
func (c *checklistTest) coin(t *testing.T, coin models.Coin, photos ...string) models.Coin {
	t.Helper()
	coin.UserID = 1
	if coin.Name == "" {
		coin.Name = coin.Denomination
	}
	if err := c.db.Create(&coin).Error; err != nil {
		t.Fatal(err)
	}
	for _, path := range photos {
		if err := c.db.Create(&models.CoinImage{CoinID: coin.ID, FilePath: path, ImageType: models.ImageTypeObverse}).Error; err != nil {
			t.Fatal(err)
		}
	}
	return coin
}

// gagoCoutinho is Portugal's €2 commemorative of 2022.
func (c *checklistTest) gagoCoutinho(t *testing.T) models.Coin {
	t.Helper()
	year, value := 2022, 3.5
	return c.coin(t, models.Coin{
		Name: "2 euro – Travessia aérea do Atlântico Sul por Gago Coutinho e Sacadura Cabral", Country: "Portugal", Denomination: "2 euro",
		Year: &year, CurrentValue: &value,
	}, "coins/gago.jpg")
}

func checklistYear(year int) *int { return &year }

func ownedSlots(slots []ChecklistSlot) []string {
	owned := []string{}
	for _, slot := range slots {
		if slot.Owned {
			owned = append(owned, slot.Key)
		}
	}
	return owned
}

func TestChecklistSlots_CatalogueOfACountry(t *testing.T) {
	test := newChecklistTest(t)
	gago := test.gagoCoutinho(t)
	// Wanted, not had.
	test.coin(t, models.Coin{Country: "Portugal", Denomination: "5 euro", Year: checklistYear(2016), IsWishlist: true, Name: "5 euro – D. Catarina de Bragança"})
	ctx := context.Background()

	definition := ChecklistDefinition{Source: ChecklistSourceCatalogue, Country: "PT"}
	slots, err := test.checklist.Slots(ctx, 1, definition)
	if err != nil {
		t.Fatal(err)
	}
	// The joint Erasmus issue and the Gago Coutinho coin of 2022, and the mint's coin of 2016, newest first.
	got := []string{}
	for _, slot := range slots {
		got = append(got, fmt.Sprintf("%d %g %v", slot.Year, slot.FaceValue, slot.Owned))
	}
	if strings.Join(got, ", ") != "2022 2 false, 2022 2 true, 2016 5 false" {
		t.Fatalf("slots = %v", got)
	}
	filled := slots[1]
	if filled.Key != gagoCoutinhoImage || filled.CoinID != gago.ID || filled.CoinName != gago.Name || filled.CoinImage != "coins/gago.jpg" || *filled.CoinValue != 3.5 {
		t.Fatalf("Gago Coutinho = %+v", filled)
	}
	if slots[2].Title != "D. Catarina de Bragança" || slots[2].PageURL != "https://loja.incm.pt/products/catarina" || slots[2].Country != "PT" {
		t.Fatalf("mint coin = %+v", slots[2])
	}
	// The catalogue's slots alone know nothing of the collection.
	if catalogue := test.checklist.CatalogueSlots(ctx, definition); len(catalogue) != 3 || len(ownedSlots(catalogue)) != 0 {
		t.Fatalf("catalogue slots = %+v", catalogue)
	}

	// A face value keeps its coins only.
	slots, _ = test.checklist.Slots(ctx, 1, ChecklistDefinition{Source: ChecklistSourceCatalogue, Country: "PT", FaceValue: 5})
	if len(slots) != 1 || slots[0].Year != 2016 || slots[0].Owned {
		t.Fatalf("5 € slots = %+v", slots)
	}
	if slots, _ := test.checklist.Slots(ctx, 1, ChecklistDefinition{Source: ChecklistSourceCatalogue, Country: "PT", Currency: "PTE"}); len(slots) != 0 {
		t.Fatalf("escudos without Numista = %+v", slots)
	}
	if !test.checklist.Complete() {
		t.Fatal("a catalogue without Numista is complete")
	}
}

func TestChecklistSlots_OneCoinOfEachEuroCountry(t *testing.T) {
	test := newChecklistTest(t)
	test.coin(t, models.Coin{Country: "Portugal", Denomination: "1 euro", Year: checklistYear(2002)})
	withPhoto := test.coin(t, models.Coin{Country: "Portugal", Denomination: "2 euro", Year: checklistYear(2010)}, "coins/pt.jpg")
	germany := test.coin(t, models.Coin{Country: "Alemanha", Denomination: "10 cêntimos", Year: checklistYear(2005)})
	// Not euros, not had, a banknote.
	test.coin(t, models.Coin{Country: "France", Denomination: "1 franc", Year: checklistYear(1995)})
	test.coin(t, models.Coin{Country: "Italy", Denomination: "2 euro", Year: checklistYear(2010), IsSold: true})
	test.coin(t, models.Coin{Country: "Spain", Denomination: "5 euro", Year: checklistYear(2015), Kind: models.CoinKindBanknote})

	slots, err := test.checklist.Slots(context.Background(), 1, ChecklistDefinition{Source: ChecklistSourceChallenge, Challenge: ChallengeEuroCountries})
	if err != nil {
		t.Fatal(err)
	}
	if len(slots) != len(ecbCountryCodes) {
		t.Fatalf("%d slots for %d countries", len(slots), len(ecbCountryCodes))
	}
	byCountry := map[string]ChecklistSlot{}
	for i, slot := range slots {
		if i > 0 && slots[i-1].Country >= slot.Country || slot.Key != "country:"+slot.Country {
			t.Fatalf("slots out of order or badly keyed: %+v", slots)
		}
		byCountry[slot.Country] = slot
	}
	if fmt.Sprint(ownedSlots(slots)) != "[country:DE country:PT]" {
		t.Fatalf("owned = %v", ownedSlots(slots))
	}
	// The coin with a photo fills the country's slot.
	if pt := byCountry["PT"]; pt.CoinID != withPhoto.ID || pt.Year != 2010 || pt.CoinImage != "coins/pt.jpg" || pt.ImageURL == "" {
		t.Fatalf("Portugal = %+v", pt)
	}
	if byCountry["DE"].CoinID != germany.ID || byCountry["FR"].ImageURL == "" {
		t.Fatalf("Germany = %+v, France = %+v", byCountry["DE"], byCountry["FR"])
	}
}

func TestChecklistSlots_AJointIssue(t *testing.T) {
	test := newChecklistTest(t)
	erasmus := test.coin(t, models.Coin{Name: "2 euro – 35 years of the Erasmus programme", Country: "Portugal", Denomination: "2 euro", Year: checklistYear(2022)}, "coins/erasmus.jpg")
	// Portugal's other coin of 2022 is not its Erasmus coin.
	test.gagoCoutinho(t)
	ctx := context.Background()

	slots, err := test.checklist.Slots(ctx, 1, ChecklistDefinition{Source: ChecklistSourceChallenge, Challenge: ChallengeJointIssue, Year: 2022})
	if err != nil {
		t.Fatal(err)
	}
	// The 19 member states of 2022, Germany's image found though its file does not say.
	countries := []string{}
	for _, slot := range slots {
		countries = append(countries, slot.Country)
		if slot.Title != "35 years of the Erasmus programme" || slot.Year != 2022 || slot.FaceValue != 2 || slot.ImageURL == "" || slot.Key != "joint:"+slot.Country ||
			slot.PageURL != "https://www.ecb.europa.eu/euro/coins/comm/html/comm_2022.en.html" {
			t.Fatalf("slot = %+v", slot)
		}
		if slot.Country == "DE" && !strings.Contains(slot.ImageURL, "EU1-erasmus") {
			t.Fatalf("Germany = %+v", slot)
		}
	}
	if strings.Join(countries, " ") != "AT BE CY DE EE ES FI FR GR IE IT LT LU LV MT NL PT SI SK" {
		t.Fatalf("countries = %v", countries)
	}
	if fmt.Sprint(ownedSlots(slots)) != "[joint:PT]" {
		t.Fatalf("owned = %v", ownedSlots(slots))
	}
	for _, slot := range slots {
		if slot.Country == "PT" && (slot.CoinID != erasmus.ID || slot.CoinImage != "coins/erasmus.jpg" || !strings.HasSuffix(slot.ImageURL, "/Portugal.jpg")) {
			t.Fatalf("Portugal = %+v", slot)
		}
	}

	// A year without a joint issue has nothing to collect.
	if slots, _ := test.checklist.Slots(ctx, 1, ChecklistDefinition{Source: ChecklistSourceChallenge, Challenge: ChallengeJointIssue, Year: 2021}); len(slots) != 0 {
		t.Fatalf("2021 = %+v", slots)
	}
}

func TestChecklistSlots_AnnualSeries(t *testing.T) {
	test := newChecklistTest(t)
	euro := test.coin(t, models.Coin{Country: "Portugal", Denomination: "1 euro", Year: checklistYear(2002)}, "coins/1euro.jpg")
	cents := test.coin(t, models.Coin{Country: "Portugal", Denomination: "50 cêntimos", Year: checklistYear(2002)})
	// A commemorative, another year, another country: not the series.
	test.coin(t, models.Coin{Country: "Portugal", Denomination: "2 euro", Year: checklistYear(2002), Name: "2 euro – Presidência"})
	test.coin(t, models.Coin{Country: "Portugal", Denomination: "2 euro", Year: checklistYear(2003)})
	test.coin(t, models.Coin{Country: "Spain", Denomination: "1 cent", Year: checklistYear(2002)})

	slots, err := test.checklist.Slots(context.Background(), 1, ChecklistDefinition{Source: ChecklistSourceChallenge, Challenge: ChallengeAnnualSeries, Country: "PT", Year: 2002})
	if err != nil {
		t.Fatal(err)
	}
	values := []string{}
	for _, slot := range slots {
		values = append(values, fmt.Sprint(slot.FaceValue))
		if slot.ImageURL == "" || slot.Year != 2002 || slot.Country != "PT" {
			t.Fatalf("slot = %+v", slot)
		}
	}
	if strings.Join(values, " ") != "0.01 0.02 0.05 0.1 0.2 0.5 1 2" {
		t.Fatalf("face values = %v", values)
	}
	if fmt.Sprint(ownedSlots(slots)) != "[annual:50cent annual:1euro]" || slots[5].CoinID != cents.ID || slots[6].CoinID != euro.ID {
		t.Fatalf("owned = %v", ownedSlots(slots))
	}
}

func TestSetChecklistWatcher_TellsOfNewCoinsOnly(t *testing.T) {
	test := newChecklistTest(t)
	sets := repository.NewSetRepository(test.db)
	catalogue := ChecklistDefinition{Source: ChecklistSourceCatalogue, Country: "PT"}.JSON()
	challenge := ChecklistDefinition{Source: ChecklistSourceChallenge, Challenge: ChallengeEuroCountries}.JSON()
	set := &models.CoinSet{UserID: 1, Name: "Portugal", SetType: models.CoinSetTypeChecklist, Checklist: &catalogue}
	other := &models.CoinSet{UserID: 1, Name: "Euro", SetType: models.CoinSetTypeChecklist, Checklist: &challenge}
	// A deleted user's set.
	orphan := &models.CoinSet{UserID: 99, Name: "Portugal", SetType: models.CoinSetTypeChecklist, Checklist: &catalogue}
	for _, s := range []*models.CoinSet{set, other, orphan} {
		if err := test.db.Create(s).Error; err != nil {
			t.Fatal(err)
		}
	}
	notifications := repository.NewNotificationRepository(test.db)
	watcher := NewSetChecklistWatcher(sets, test.checklist, notifications, NewLogger(50))
	watcher.now = func() time.Time { return test.clock }
	ctx := context.Background()
	told := func() []models.Notification {
		list := []models.Notification{}
		if err := test.db.Order("id").Find(&list).Error; err != nil {
			t.Fatal(err)
		}
		return list
	}

	// The first look learns the set's coins without a word.
	if err := watcher.Check(ctx); err != nil {
		t.Fatal(err)
	}
	if known, _ := sets.ChecklistKeys(set.ID); len(known) != 3 || len(told()) != 0 {
		t.Fatalf("known %v, told %+v", known, told())
	}
	if known, _ := sets.ChecklistKeys(other.ID); len(known) != 0 {
		t.Fatalf("a challenge has catalogue coins: %v", known)
	}
	if known, _ := sets.ChecklistKeys(orphan.ID); len(known) != 0 {
		t.Fatalf("a deleted user's set was looked at: %v", known)
	}

	// The mint puts this year's coin and an old one on sale.
	test.mint.coins = append(test.mint.coins,
		mintCoin{Title: "Rotary", Value: 2, Year: 2026, Released: "2026-07", PageURL: "https://loja.incm.pt/products/rotary", Obverse: "https://cdn/rotary.png"},
		mintCoin{Title: "Fado", Value: 7.5, Year: 2019, PageURL: "https://loja.incm.pt/products/fado", Obverse: "https://cdn/fado.png"},
	)
	test.clock = test.clock.Add(7 * time.Hour)
	if err := watcher.Check(ctx); err != nil {
		t.Fatal(err)
	}
	list := told()
	if len(list) != 1 {
		t.Fatalf("notifications = %+v", list)
	}
	n := list[0]
	if n.UserID != 1 || n.Type != NotificationTypeSetNewCoins || n.Title != "Portugal" || n.Message != "Rotary (2026)" || n.ReferenceID != set.ID || n.ReferenceURL != fmt.Sprintf("/sets/%d", set.ID) {
		t.Fatalf("notification = %+v", n)
	}
	if known, _ := sets.ChecklistKeys(set.ID); len(known) != 5 {
		t.Fatalf("known = %v", known)
	}

	// Nothing new: nothing said.
	test.clock = test.clock.Add(7 * time.Hour)
	if err := watcher.Check(ctx); err != nil || len(told()) != 1 {
		t.Fatalf("again: %v, %+v", err, told())
	}
}

func TestSetChecklistWatcher_NamesFiveCoins(t *testing.T) {
	test := newChecklistTest(t)
	sets := repository.NewSetRepository(test.db)
	catalogue := ChecklistDefinition{Source: ChecklistSourceCatalogue, Country: "PT", FaceValue: 5}.JSON()
	set := &models.CoinSet{UserID: 1, Name: "5 €", SetType: models.CoinSetTypeChecklist, Checklist: &catalogue}
	if err := test.db.Create(set).Error; err != nil {
		t.Fatal(err)
	}
	watcher := NewSetChecklistWatcher(sets, test.checklist, repository.NewNotificationRepository(test.db), NewLogger(50))
	watcher.now = func() time.Time { return test.clock }
	if err := watcher.Check(context.Background()); err != nil {
		t.Fatal(err)
	}
	for i, title := range []string{"Faróis", "Fado", "Azulejo", "Cortiça", "Galo de Barcelos", "Caravela", "Sardinha"} {
		test.mint.coins = append(test.mint.coins, mintCoin{Title: title, Value: 5, Year: 2026, PageURL: fmt.Sprintf("https://loja.incm.pt/products/coin-%d", i), Obverse: "https://cdn/coin.png"})
	}
	test.clock = test.clock.Add(7 * time.Hour)
	if err := watcher.Check(context.Background()); err != nil {
		t.Fatal(err)
	}
	var n models.Notification
	if err := test.db.First(&n).Error; err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(n.Message, "\n")
	if len(lines) != 6 || lines[5] != "+2" || !strings.HasSuffix(lines[0], "(2026)") {
		t.Fatalf("message = %q", n.Message)
	}
}

func TestSetService_ChecklistSets(t *testing.T) {
	test := newChecklistTest(t)
	test.gagoCoutinho(t)
	setRepo := repository.NewSetRepository(test.db)
	svc := NewSetService(setRepo, repository.NewTagRepository(test.db)).WithChecklist(test.checklist)
	ctx := context.Background()

	set, err := svc.CreateSet(1, map[string]interface{}{"name": "Portugal", "setType": "checklist", "checklist": map[string]interface{}{"source": "catalogue", "country": "pt"}})
	if err != nil {
		t.Fatal(err)
	}
	if set.SetType != models.CoinSetTypeChecklist || (*set.Checklist)["country"] != "PT" {
		t.Fatalf("set = %+v", set)
	}
	// The catalogue's coins of today are known from the start.
	if known, _ := setRepo.ChecklistKeys(set.ID); len(known) != 3 {
		t.Fatalf("known = %v", known)
	}
	for _, input := range []map[string]interface{}{
		{"name": "No list", "setType": "checklist"},
		{"name": "Bad list", "setType": "checklist", "checklist": map[string]interface{}{"source": "catalogue", "country": "US"}},
		{"name": "No joint issue", "setType": "checklist", "checklist": map[string]interface{}{"source": "challenge", "challenge": "joint-issue", "year": 2021}},
	} {
		if _, err := svc.CreateSet(1, input); err == nil {
			t.Fatalf("created %v", input)
		}
	}

	// One of three slots filled, by a coin worth 3.5.
	sets, err := svc.ListSets(1)
	if err != nil || len(sets) != 1 {
		t.Fatalf("sets = %+v, %v", sets, err)
	}
	if sets[0]["coinCount"] != 1 || sets[0]["totalValue"] != 3.5 || fmt.Sprintf("%.1f", sets[0]["completionPercentage"]) != "33.3" {
		t.Fatalf("listed = %+v", sets[0])
	}
	detail, err := svc.GetSetDetail(set.ID, 1)
	if err != nil || detail["coinCount"] != 1 || detail["checklist"] == nil || fmt.Sprintf("%.1f", detail["completionPercentage"]) != "33.3" {
		t.Fatalf("detail = %+v, %v", detail, err)
	}
	completion, err := svc.GetCompletion(set.ID, 1)
	if err != nil || completion["totalTargets"] != 3 || completion["completedTargets"] != 1 {
		t.Fatalf("completion = %+v, %v", completion, err)
	}
	slots, complete, err := svc.ChecklistSlots(ctx, set.ID, 1)
	if err != nil || len(slots) != 3 || !complete {
		t.Fatalf("slots = %+v, %v, %v", slots, complete, err)
	}
	if _, _, err := svc.ChecklistSlots(ctx, set.ID, 2); err == nil {
		t.Fatal("another user's set has slots")
	}

	// Its coins are the catalogue's: none are added, removed or ordered by hand.
	if err := svc.AddCoinToSet(1, set.ID, 1, ""); err == nil {
		t.Fatal("a coin was added to a checklist set")
	}
	if err := svc.RemoveCoinFromSet(1, set.ID, 1); err == nil {
		t.Fatal("a coin was removed from a checklist set")
	}
	if err := svc.ReorderCoinsInSet(set.ID, 1, []uint{1}); err == nil {
		t.Fatal("a checklist set was ordered")
	}
	// It stays a checklist of the same coins; its name can change.
	if _, err := svc.UpdateSet(set.ID, 1, map[string]interface{}{"setType": "standard"}); err == nil {
		t.Fatal("a checklist set became a standard set")
	}
	updated, err := svc.UpdateSet(set.ID, 1, map[string]interface{}{"name": "Portugal inteiro", "checklist": map[string]interface{}{"source": "catalogue", "country": "FR"}})
	if err != nil || updated.Name != "Portugal inteiro" || (*updated.Checklist)["country"] != "PT" {
		t.Fatalf("updated = %+v, %v", updated, err)
	}

	// A joint issue: none of its 19 coins in the collection yet.
	erasmus, err := svc.CreateSet(1, map[string]interface{}{"name": "Erasmus", "setType": "checklist", "checklist": map[string]interface{}{"source": "challenge", "challenge": "joint-issue", "year": 2022}})
	if err != nil {
		t.Fatalf("joint issue of 2022: %v", err)
	}
	if completion, _ := svc.GetCompletion(erasmus.ID, 1); completion["totalTargets"] != 19 || completion["completedTargets"] != 0 {
		t.Fatalf("Erasmus completion = %+v", completion)
	}

	// A standard set has no slots and cannot become a checklist.
	standard, err := svc.CreateSet(1, map[string]interface{}{"name": "Mine"})
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := svc.ChecklistSlots(ctx, standard.ID, 1); err != ErrNotChecklistSet {
		t.Fatalf("standard set slots: %v", err)
	}
	if _, err := svc.UpdateSet(standard.ID, 1, map[string]interface{}{"setType": "checklist"}); err == nil {
		t.Fatal("a standard set became a checklist")
	}

	// Deleting the set forgets the coins it knew.
	if err := svc.DeleteSet(set.ID, 1); err != nil {
		t.Fatal(err)
	}
	if known, _ := setRepo.ChecklistKeys(set.ID); len(known) != 0 {
		t.Fatalf("known after delete = %v", known)
	}
}

func TestSetService_SmartCriteriaCanBeEdited(t *testing.T) {
	test := newChecklistTest(t)
	svc := NewSetService(repository.NewSetRepository(test.db), repository.NewTagRepository(test.db))
	criteria := map[string]interface{}{"operator": "and", "rules": []interface{}{map[string]interface{}{"field": "country", "op": "eq", "value": "Portugal"}}}
	set, err := svc.CreateSet(1, map[string]interface{}{"name": "Portugal", "setType": "smart", "smartCriteria": criteria})
	if err != nil {
		t.Fatal(err)
	}
	changed := map[string]interface{}{"operator": "and", "rules": []interface{}{map[string]interface{}{"field": "year", "op": "gte", "value": 2002}}}
	updated, err := svc.UpdateSet(set.ID, 1, map[string]interface{}{"smartCriteria": changed})
	if err != nil {
		t.Fatal(err)
	}
	rules := (*updated.SmartCriteria)["rules"].([]interface{})
	if len(rules) != 1 || rules[0].(map[string]interface{})["field"] != "year" {
		t.Fatalf("criteria = %+v", *updated.SmartCriteria)
	}
	detail, _ := svc.GetSetDetail(set.ID, 1)
	if detail["smartCriteria"] == nil {
		t.Fatalf("detail = %+v", detail)
	}
	for _, bad := range []map[string]interface{}{
		{"smartCriteria": map[string]interface{}{"operator": "and", "rules": []interface{}{map[string]interface{}{"field": "classification", "op": "eq", "value": "gold"}}}},
		{"smartCriteria": "country = Portugal"},
	} {
		if _, err := svc.UpdateSet(set.ID, 1, bad); err == nil {
			t.Fatalf("updated with %v", bad)
		}
	}
	standard, _ := svc.CreateSet(1, map[string]interface{}{"name": "Mine"})
	if _, err := svc.UpdateSet(standard.ID, 1, map[string]interface{}{"smartCriteria": changed}); err == nil {
		t.Fatal("a standard set got criteria")
	}
}
