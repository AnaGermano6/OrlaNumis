package handlers

import (
	"encoding/json"
	"net/http"
	"sort"
	"testing"

	"github.com/briandenicola/ancient-coins-api/models"
	"github.com/gin-gonic/gin"
)

func TestCoinList_Ungraded(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router, db := setupCoinHandlerRouter(t)
	createTestUser(t, db, 1, "collector")
	for _, coin := range []models.Coin{
		{Name: "Graded", Grade: "UNC"},
		{Name: "No grade"},
		{Name: "Blank grade", Grade: "  "},
		{Name: "Wanted", IsWishlist: true},
	} {
		coin.UserID, coin.Category, coin.Material = 1, models.CategoryModern, models.MaterialOther
		if err := db.Create(&coin).Error; err != nil {
			t.Fatal(err)
		}
	}

	w := coinRequest(router, http.MethodGet, "/api/coins?ungraded=true&wishlist=false", nil)
	var got struct {
		Coins []models.Coin `json:"coins"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil || w.Code != http.StatusOK {
		t.Fatalf("list: %d %s", w.Code, w.Body.String())
	}
	names := []string{}
	for _, coin := range got.Coins {
		names = append(names, coin.Name)
	}
	sort.Strings(names)
	// Only the collection's coins without a grade, a blank one included.
	if len(names) != 2 || names[0] != "Blank grade" || names[1] != "No grade" {
		t.Fatalf("ungraded = %v", names)
	}
}
