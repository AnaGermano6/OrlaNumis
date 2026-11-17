package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/briandenicola/ancient-coins-api/models"
)

func TestHealthHandler_ListCoinHealth_Priority(t *testing.T) {
	router, db, userID := setupHealthHandlerRouter(t)
	value, now := 10.0, time.Now()
	complete := func(name string) models.Coin {
		return models.Coin{
			Name: name, UserID: userID, Category: models.CategoryModern, Material: models.MaterialSilver,
			Denomination: "1 escudo", Country: "Portugal", MonetarySystem: "Escudo", Mint: "Lisboa", Grade: "VF",
			CurrentValue: &value, CurrentValueUpdatedAt: &now,
			// Both sides analysed, in case an AI provider is set up.
			AIAnalysis: "Done", ObverseAnalysis: "Obverse", ReverseAnalysis: "Reverse",
		}
	}
	whole := complete("Complete")
	ungraded := complete("Ungraded")
	ungraded.Grade = ""
	unvalued := complete("Unvalued")
	unvalued.CurrentValue, unvalued.CurrentValueUpdatedAt = nil, nil
	bare := models.Coin{Name: "Bare", UserID: userID, Category: models.CategoryModern}
	for _, coin := range []*models.Coin{&whole, &ungraded, &unvalued, &bare} {
		if err := db.Create(coin).Error; err != nil {
			t.Fatal(err)
		}
	}
	// Both sides of each but the bare coin.
	for _, coin := range []*models.Coin{&whole, &ungraded, &unvalued} {
		for _, side := range []models.ImageType{models.ImageTypeObverse, models.ImageTypeReverse} {
			if err := db.Create(&models.CoinImage{CoinID: coin.ID, FilePath: "coins/" + coin.Name + "-" + string(side) + ".jpg", ImageType: side, IsPrimary: side == models.ImageTypeObverse}).Error; err != nil {
				t.Fatal(err)
			}
		}
	}

	list := func(query string) []struct {
		Title        string `json:"title"`
		ImagePath    string `json:"imagePath"`
		MissingItems []struct {
			Key string `json:"key"`
		} `json:"missingItems"`
	} {
		req := httptest.NewRequest(http.MethodGet, "/api/coins/health?"+query, nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("%s: %d %s", query, w.Code, w.Body.String())
		}
		var resp struct {
			Coins []struct {
				Title        string `json:"title"`
				ImagePath    string `json:"imagePath"`
				MissingItems []struct {
					Key string `json:"key"`
				} `json:"missingItems"`
			} `json:"coins"`
		}
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatal(err)
		}
		return resp.Coins
	}

	// Only the coins that miss something, the most urgent first: the bare coin
	// and the unvalued one miss something urgent, the bare one scores lower;
	// the grade is less urgent.
	coins := list("scope=priority")
	titles := []string{}
	for _, coin := range coins {
		titles = append(titles, coin.Title)
	}
	if len(titles) != 3 || titles[0] != "Bare" || titles[1] != "Unvalued" || titles[2] != "Ungraded" {
		t.Fatalf("priority = %v", titles)
	}
	if coins[2].MissingItems[0].Key != "metadata.grade" || coins[2].ImagePath != "coins/Ungraded-obverse.jpg" || coins[0].ImagePath != "" {
		t.Fatalf("ungraded = %+v, bare = %+v", coins[2], coins[0])
	}

	// Of one dimension: only the coins that miss something of it, that first.
	images := list("scope=priority&dimension=images")
	if len(images) != 1 || images[0].Title != "Bare" || images[0].MissingItems[0].Key != "images.any" {
		t.Fatalf("images = %+v", images)
	}

	// Of one country, and of one missing item (that item first): the bare coin has no country.
	portugal := list("scope=priority&country=portugal")
	if len(portugal) != 2 || portugal[0].Title != "Unvalued" || portugal[1].Title != "Ungraded" {
		t.Fatalf("Portugal = %+v", portugal)
	}
	grade := list("scope=priority&missing=metadata.grade")
	if len(grade) != 2 || grade[0].Title != "Bare" || grade[0].MissingItems[0].Key != "metadata.grade" || grade[1].Title != "Ungraded" {
		t.Fatalf("missing grade = %+v", grade)
	}
	if both := list("scope=priority&country=Portugal&missing=metadata.grade"); len(both) != 1 || both[0].Title != "Ungraded" {
		t.Fatalf("both = %+v", both)
	}

	// The facets: by country with the missing item on, by missing item with the country on.
	req := httptest.NewRequest(http.MethodGet, "/api/coins/health?scope=priority&missing=metadata.grade", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	var faceted struct {
		Facets struct {
			Countries []struct {
				Value string `json:"value"`
				Count int    `json:"count"`
			} `json:"countries"`
			Missing []struct {
				Value string `json:"value"`
				Count int    `json:"count"`
			} `json:"missing"`
		} `json:"facets"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &faceted); err != nil {
		t.Fatal(err)
	}
	if len(faceted.Facets.Countries) != 2 || faceted.Facets.Countries[0].Value != "" || faceted.Facets.Countries[1] != (struct {
		Value string `json:"value"`
		Count int    `json:"count"`
	}{Value: "Portugal", Count: 1}) {
		t.Fatalf("countries = %+v", faceted.Facets.Countries)
	}
	counts := map[string]int{}
	for _, facet := range faceted.Facets.Missing {
		counts[facet.Value] = facet.Count
	}
	if counts["metadata.grade"] != 2 || counts["valuation.value"] != 2 || counts["images.any"] != 1 {
		t.Fatalf("missing = %+v", faceted.Facets.Missing)
	}

	// A dimension needs the priority scope.
	req = httptest.NewRequest(http.MethodGet, "/api/coins/health?scope=all&dimension=images", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("dimension without priority: %d", w.Code)
	}
}
