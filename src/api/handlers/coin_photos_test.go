package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/briandenicola/ancient-coins-api/models"
	"github.com/briandenicola/ancient-coins-api/repository"
	"github.com/briandenicola/ancient-coins-api/services"
	"github.com/gin-gonic/gin"
)

type stubCommonsClient struct{}

func (stubCommonsClient) SearchFiles(context.Context, string, int) ([]byte, error) {
	return []byte(`{"batchcomplete": true}`), nil
}

type stubECBCoinsClient struct{}

func (stubECBCoinsClient) CommemorativePage(context.Context, int) ([]byte, error) {
	return []byte(`<div class="box"><img src="comm_2022/DE-thueringen.jpg" loading="lazy"><div class="content-box"><h3>Germany</h3>` +
		`<div><p><strong>Feature:</strong> Bundeslander series – Thuringia</p></div></div></div>`), nil
}

func setupCoinPhotoRouter(t *testing.T) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	db := setupCoinHandlerTestDB(t)
	if err := db.AutoMigrate(&models.NumistaCacheEntry{}); err != nil {
		t.Fatal(err)
	}
	cache := services.NewNumistaResponseCache(repository.NewNumistaCacheRepository(db), nil, nil)
	svc := services.NewCoinPhotoSearchService(nil, stubCommonsClient{}, cache).WithECB(services.NewEcbCoinPhotoService(stubECBCoinsClient{}, cache))
	router := gin.New()
	group := router.Group("/api")
	group.Use(coinTestAuthMiddleware())
	group.GET("/coin-photos/search", NewCoinPhotoHandler(svc).Search)
	return router
}

func TestCoinPhotoHandler_SearchesTheECBByYearAndCountry(t *testing.T) {
	router := setupCoinPhotoRouter(t)
	for _, path := range []string{
		"/api/coin-photos/search",
		"/api/coin-photos/search?year=2022",
		"/api/coin-photos/search?country=Alemanha",
		"/api/coin-photos/search?year=1700&country=Alemanha",
	} {
		if w := coinRequest(router, http.MethodGet, path, nil); w.Code != http.StatusBadRequest {
			t.Errorf("%s: expected 400, got %d", path, w.Code)
		}
	}

	// A coin not linked to Numista is found by its year, country, denomination and subject.
	w := coinRequest(router, http.MethodGet, "/api/coin-photos/search?lang=pt&year=2022&country=Alemanha&denomination=2+euro&subject=Tur%C3%ADngia", nil)
	var got models.CoinPhotoSearchResult
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil || w.Code != http.StatusOK {
		t.Fatalf("search: %d %s", w.Code, w.Body.String())
	}
	if len(got.Photos) != 1 || got.Photos[0].Source != models.CoinPhotoSourceECB || got.Photos[0].Match != models.CoinPhotoMatchExact {
		t.Fatalf("photos = %+v", got.Photos)
	}
}
