package services

import (
	"fmt"
	"strings"
	"testing"

	"github.com/briandenicola/ancient-coins-api/models"
	"github.com/briandenicola/ancient-coins-api/repository"
)

func TestChecklistCovers(t *testing.T) {
	slots := []ChecklistSlot{
		{Country: "AT", ImageURL: "at.jpg"}, {Country: "BE", ImageURL: "be.jpg"}, {Country: "BE", ImageURL: "be.jpg"},
		{Country: "PT", ImageURL: "pt.jpg", CoinImage: "coins/pt.jpg"}, {Country: "SK", ImageURL: "sk.jpg"},
	}
	for name, test := range map[string]struct {
		slots           []ChecklistSlot
		oneCoin         bool
		urls, coinPhoto string
	}{
		"three different pictures, Portugal's first": {slots: slots, urls: "[pt.jpg at.jpg be.jpg]", coinPhoto: "[]"},
		"a joint issue's one, Portugal's":            {slots: slots, oneCoin: true, urls: "[pt.jpg]", coinPhoto: "[]"},
		"else the collection's photos":               {slots: []ChecklistSlot{{Country: "AT"}, {Country: "BE", CoinImage: "coins/be.jpg"}}, urls: "[]", coinPhoto: "[coins/be.jpg]"},
		"else none":                                  {slots: []ChecklistSlot{{Country: "AT"}}, urls: "[]", coinPhoto: "[]"},
	} {
		urls, coinImages := ChecklistCovers(test.slots, test.oneCoin)
		if fmt.Sprint(urls) != test.urls || fmt.Sprint(coinImages) != test.coinPhoto {
			t.Fatalf("%s: %v, %v", name, urls, coinImages)
		}
	}
}

// A set's covers are photos of up to three of its coins, one each (a
// primary one first); an empty set and a smart set have none.
func TestListSets_CoverImages(t *testing.T) {
	service, db := setupSetServiceTestWithDB(t)
	coins := []models.Coin{{Name: "Sem foto", UserID: 1}, {Name: "Com fotos", UserID: 1}, {Name: "Terceira", UserID: 1}, {Name: "Quarta", UserID: 1}, {Name: "Quinta", UserID: 1}}
	if err := db.Create(&coins).Error; err != nil {
		t.Fatal(err)
	}
	images := []models.CoinImage{
		{CoinID: coins[1].ID, FilePath: "coins/reverse.jpg", ImageType: models.ImageTypeReverse},
		{CoinID: coins[1].ID, FilePath: "coins/obverse.jpg", ImageType: models.ImageTypeObverse, IsPrimary: true},
	}
	for _, coin := range coins[2:] {
		images = append(images, models.CoinImage{CoinID: coin.ID, FilePath: "coins/" + coin.Name + ".jpg"})
	}
	for _, image := range images {
		if err := db.Create(&image).Error; err != nil {
			t.Fatal(err)
		}
	}
	sets := []models.CoinSet{
		{UserID: 1, Name: "Com moedas", SetType: models.CoinSetTypeStandard},
		{UserID: 1, Name: "Vazio", SetType: models.CoinSetTypeStandard},
		{UserID: 1, Name: "Inteligente", SetType: models.CoinSetTypeSmart},
	}
	if err := db.Create(&sets).Error; err != nil {
		t.Fatal(err)
	}
	for _, coin := range coins {
		if err := db.Create(&models.CoinSetMembership{SetID: sets[0].ID, CoinID: coin.ID}).Error; err != nil {
			t.Fatal(err)
		}
	}

	listed, err := service.ListSets(1)
	if err != nil {
		t.Fatal(err)
	}
	covers := map[string]string{}
	for _, set := range listed {
		covers[set["name"].(string)] = fmt.Sprint(set["coverImages"])
		if set["coverImageUrls"] != nil || set["year"] != nil || set["total"] != nil || set["oneCoin"] != nil {
			t.Fatalf("set = %+v", set)
		}
	}
	if covers["Com moedas"] != "[coins/obverse.jpg coins/Terceira.jpg coins/Quarta.jpg]" || covers["Vazio"] != "<nil>" || covers["Inteligente"] != "<nil>" {
		t.Fatalf("covers = %+v", covers)
	}
}

// A joint issue's set has its year, its slots in all, and the catalogue's picture of Portugal's coin alone: it is one coin.
func TestListSets_AJointIssueYearTotalAndCover(t *testing.T) {
	test := newChecklistTest(t)
	svc := NewSetService(repository.NewSetRepository(test.db)).WithChecklist(test.checklist)
	if _, err := svc.CreateSet(1, map[string]interface{}{"name": "Emissão comum de 2022: 35 years of the Erasmus programme", "setType": "checklist", "checklist": map[string]interface{}{"source": "challenge", "challenge": "joint-issue", "year": 2022}}); err != nil {
		t.Fatal(err)
	}

	listed, err := svc.ListSets(1)
	if err != nil || len(listed) != 1 {
		t.Fatalf("sets = %+v, %v", listed, err)
	}
	set := listed[0]
	if set["year"] != 2022 || set["total"] != 19 || set["coinCount"] != 0 || set["oneCoin"] != true || set["coverImages"] != nil {
		t.Fatalf("set = %+v", set)
	}
	if urls, _ := set["coverImageUrls"].([]string); len(urls) != 1 || !strings.HasSuffix(urls[0], "/Portugal.jpg") {
		t.Fatalf("covers = %v", set["coverImageUrls"])
	}
}
