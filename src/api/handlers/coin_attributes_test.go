package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/gin-gonic/gin"
)

func coinRequest(router *gin.Engine, method, path string, body any) *httptest.ResponseRecorder {
	var payload []byte
	if body != nil {
		payload, _ = json.Marshal(body)
	}
	req := httptest.NewRequest(method, path, bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", authHeader(1))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

func TestCoinHandler_CountryPresentationRarityAndYear(t *testing.T) {
	router, db := setupCoinHandlerRouter(t)
	createTestUser(t, db, 1, "collector")

	w := coinRequest(router, http.MethodPost, "/api/coins", map[string]any{
		"name": "50 Escudos", "category": "Modern", "material": "Silver",
		"dateRange": "1968", "country": " Portugal ", "presentation": "coincard", "rarity": "scarce",
	})
	if w.Code != http.StatusCreated {
		t.Fatalf("create: %d %s", w.Code, w.Body.String())
	}
	var created struct {
		ID           uint   `json:"id"`
		Year         *int   `json:"year"`
		Country      string `json:"country"`
		Presentation string `json:"presentation"`
		Rarity       string `json:"rarity"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &created); err != nil || created.Year == nil || *created.Year != 1968 || created.Country != "Portugal" || created.Presentation != "coincard" || created.Rarity != "scarce" {
		t.Fatalf("created = %+v (%v)", created, err)
	}
	coinRequest(router, http.MethodPost, "/api/coins", map[string]any{"name": "Loose one", "category": "Modern", "material": "Silver", "presentation": "loose"})

	w = coinRequest(router, http.MethodGet, "/api/coins?presentation=coincard", nil)
	var list struct {
		Coins []struct {
			Name string `json:"name"`
		} `json:"coins"`
		Total int `json:"total"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &list); err != nil || w.Code != http.StatusOK || list.Total != 1 || list.Coins[0].Name != "50 Escudos" {
		t.Fatalf("coincard filter: %d %s", w.Code, w.Body.String())
	}

	for _, path := range []string{"/api/coins?presentation=shoebox", "/api/coins?rarity=R2"} {
		if w := coinRequest(router, http.MethodGet, path, nil); w.Code != http.StatusBadRequest {
			t.Errorf("%s: expected 400, got %d", path, w.Code)
		}
	}
	if w := coinRequest(router, http.MethodPost, "/api/coins", map[string]any{"name": "Bad", "category": "Modern", "material": "Silver", "rarity": "legendary"}); w.Code != http.StatusBadRequest {
		t.Errorf("invalid rarity on create: expected 400, got %d", w.Code)
	}

	w = coinRequest(router, http.MethodPut, "/api/coins/"+itoa(created.ID), map[string]any{"dateRange": "44 BC", "presentation": "slab"})
	if w.Code != http.StatusOK {
		t.Fatalf("update: %d %s", w.Code, w.Body.String())
	}
	if err := json.Unmarshal(w.Body.Bytes(), &created); err != nil || created.Year == nil || *created.Year != -44 || created.Presentation != "slab" || created.Country != "Portugal" {
		t.Fatalf("updated = %+v", created)
	}
}

func TestCoinHandler_YearCanBeTypedOrLeftToTheDate(t *testing.T) {
	router, db := setupCoinHandlerRouter(t)
	createTestUser(t, db, 1, "collector")
	type yearResponse struct {
		ID   uint `json:"id"`
		Year *int `json:"year"`
	}

	w := coinRequest(router, http.MethodPost, "/api/coins", map[string]any{"name": "Denarius", "category": "Modern", "material": "Silver", "dateRange": "c. 46-44 BC", "year": -44})
	var created yearResponse
	if err := json.Unmarshal(w.Body.Bytes(), &created); err != nil || w.Code != http.StatusCreated || created.Year == nil || *created.Year != -44 {
		t.Fatalf("create with year: %d %s", w.Code, w.Body.String())
	}

	w = coinRequest(router, http.MethodPut, "/api/coins/"+itoa(created.ID), map[string]any{"year": nil})
	var updated yearResponse
	if err := json.Unmarshal(w.Body.Bytes(), &updated); err != nil || w.Code != http.StatusOK || updated.Year == nil || *updated.Year != -46 {
		t.Fatalf("clearing the year should read it from the date: %d %s", w.Code, w.Body.String())
	}

	if w := coinRequest(router, http.MethodPost, "/api/coins", map[string]any{"name": "Bad", "category": "Modern", "material": "Silver", "year": 99999}); w.Code != http.StatusBadRequest {
		t.Fatalf("out-of-range year: expected 400, got %d", w.Code)
	}
}

func itoa(value uint) string {
	return strconv.FormatUint(uint64(value), 10)
}
