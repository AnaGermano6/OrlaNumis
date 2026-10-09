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

type stubNumistaCatalogClient struct {
	stubNumistaPriceClient
	lastSearch services.NumistaTypeSearch
}

func (s *stubNumistaCatalogClient) SearchTypes(_ context.Context, search services.NumistaTypeSearch) ([]byte, error) {
	s.calls++
	s.lastSearch = search
	return []byte(`{"count": 1, "types": [{"id": 5, "title": "1 Escudo", "issuer": {"code": "portugal", "name": "Portugal"}}]}`), nil
}

func (s *stubNumistaCatalogClient) TypeDetail(context.Context, int, string) ([]byte, error) {
	s.calls++
	return []byte(`{"id": 5, "title": "1 Escudo", "issuer": {"code": "portugal", "name": "Portugal"}, "weight": 8}`), nil
}

func (s *stubNumistaCatalogClient) Issuers(context.Context, string) ([]byte, error) {
	s.calls++
	return []byte(`{"issuers": [{"code": "portugal", "name": "Portugal"}]}`), nil
}

func setupNumistaCatalogRouter(t *testing.T) (*gin.Engine, *stubNumistaCatalogClient) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	db := setupCoinHandlerTestDB(t)
	if err := db.AutoMigrate(&models.NumistaCacheEntry{}); err != nil {
		t.Fatalf("migrate cache: %v", err)
	}
	client := &stubNumistaCatalogClient{}
	cache := services.NewNumistaResponseCache(repository.NewNumistaCacheRepository(db), nil, nil)
	catalog := services.NewNumistaCatalogService(client, cache)
	handler := NewNumistaCatalogHandler(catalog, services.NewNumistaMarketValueService(client, cache).WithCatalog(catalog))

	r := gin.New()
	protected := r.Group("/api")
	protected.Use(coinTestAuthMiddleware())
	protected.GET("/numista/catalog/issuers", handler.Issuers)
	protected.GET("/numista/catalog/search", handler.Search)
	protected.GET("/numista/catalog/types/:id", handler.Type)
	protected.GET("/numista/catalog/types/:id/market-value", handler.MarketValue)
	return r, client
}

func TestNumistaCatalogHandler_ValidatesInput(t *testing.T) {
	router, client := setupNumistaCatalogRouter(t)
	for _, path := range []string{
		"/api/numista/catalog/search",
		"/api/numista/catalog/search?issuer=Portugal!",
		"/api/numista/catalog/search?q=escudo&lang=portuguese",
		"/api/numista/catalog/search?q=escudo&year=abc",
		"/api/numista/catalog/search?q=escudo&page=0",
		"/api/numista/catalog/types/0",
		"/api/numista/catalog/types/5/market-value?currency=euros",
		"/api/numista/catalog/types/5/market-value?issueId=-1",
	} {
		if w := getMarketValue(router, path, 1); w.Code != http.StatusBadRequest {
			t.Errorf("%s: expected 400, got %d", path, w.Code)
		}
	}
	if client.calls != 0 {
		t.Fatalf("invalid requests must not reach Numista, got %d calls", client.calls)
	}
}

func TestNumistaCatalogHandler_ServesSearchTypeAndIssuers(t *testing.T) {
	router, client := setupNumistaCatalogRouter(t)

	w := getMarketValue(router, "/api/numista/catalog/search?issuer=portugal&year=-44&page=2&lang=pt", 1)
	var search models.NumistaCatalogSearchResult
	if err := json.Unmarshal(w.Body.Bytes(), &search); err != nil || w.Code != http.StatusOK {
		t.Fatalf("search: %d %v %s", w.Code, err, w.Body.String())
	}
	if len(search.Types) != 1 || search.Page != 2 || client.lastSearch.Year != -44 || client.lastSearch.Lang != "pt" || client.lastSearch.Issuer != "portugal" {
		t.Fatalf("search = %+v sent %+v", search, client.lastSearch)
	}

	w = getMarketValue(router, "/api/numista/catalog/types/5?lang=pt", 1)
	var typeResult models.NumistaCatalogTypeResult
	if err := json.Unmarshal(w.Body.Bytes(), &typeResult); err != nil || typeResult.Type == nil || *typeResult.Type.WeightGrams != 8 || len(typeResult.Issues) != 1 {
		t.Fatalf("type: %d %s", w.Code, w.Body.String())
	}

	w = getMarketValue(router, "/api/numista/catalog/issuers?q=portu", 1)
	var issuers models.NumistaCatalogIssuersResult
	if err := json.Unmarshal(w.Body.Bytes(), &issuers); err != nil || len(issuers.Issuers) != 1 || issuers.Issuers[0].Code != "portugal" {
		t.Fatalf("issuers: %d %s", w.Code, w.Body.String())
	}

	w = getMarketValue(router, "/api/numista/catalog/types/5/market-value?currency=usd", 1)
	var value models.CoinMarketValue
	if err := json.Unmarshal(w.Body.Bytes(), &value); err != nil || value.Status != models.NumistaDataStatusOK || value.Currency != "USD" {
		t.Fatalf("market value: %d %s", w.Code, w.Body.String())
	}
}
