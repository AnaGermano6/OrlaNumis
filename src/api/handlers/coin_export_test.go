package handlers

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/briandenicola/ancient-coins-api/models"
	"github.com/briandenicola/ancient-coins-api/repository"
	"github.com/briandenicola/ancient-coins-api/services"
	"github.com/gin-gonic/gin"
)

func TestCoinHandler_ExportsTheFilteredCoins(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupCoinHandlerTestDB(t)
	// More than a page of coins, of two countries, and a wished one.
	for i := 0; i < 130; i++ {
		country := "Portugal"
		if i%2 == 1 {
			country = "Espanha"
		}
		db.Create(&models.Coin{Name: fmt.Sprintf("Coin %03d", i), UserID: 1, Country: country, Quantity: 1})
	}
	db.Create(&models.Coin{Name: "Wanted", UserID: 1, Country: "Portugal", IsWishlist: true})
	db.Create(&models.Coin{Name: "Someone else's", UserID: 2, Country: "Portugal"})

	handler := NewCoinHandler(repository.NewCoinRepository(db), services.NewCoinService(repository.NewCoinRepository(db)), services.NewLogger(10))
	router := gin.New()
	api := router.Group("/api")
	api.Use(coinTestAuthMiddleware())
	api.GET("/coins/export", handler.Export)
	get := func(query string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodGet, "/api/coins/export?"+query, nil)
		req.Header.Set("Authorization", authHeader(1))
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		return w
	}
	rows := func(w *httptest.ResponseRecorder, comma rune) [][]string {
		t.Helper()
		reader := csv.NewReader(bytes.NewReader(bytes.TrimPrefix(w.Body.Bytes(), []byte("\xef\xbb\xbf"))))
		reader.Comma = comma
		records, err := reader.ReadAll()
		if err != nil {
			t.Fatalf("csv: %v", err)
		}
		return records
	}

	all := get("format=csv&lang=pt&wishlist=false&sort=name&order=asc")
	if all.Code != http.StatusOK || !strings.Contains(all.Header().Get("Content-Disposition"), `filename="colecao-`) {
		t.Fatalf("all: %d %q", all.Code, all.Header().Get("Content-Disposition"))
	}
	records := rows(all, ';')
	if len(records) != 131 || records[1][2] != "Coin 000" || records[130][2] != "Coin 129" {
		t.Fatalf("every collection coin, in order, past the first page: %d rows, first %v", len(records), records[1])
	}

	filtered := rows(get("format=csv&lang=en&wishlist=false&country=Espanha"), ',')
	if len(filtered) != 66 {
		t.Fatalf("only Spain's: %d rows", len(filtered))
	}
	for _, record := range filtered[1:] {
		if record[3] != "Espanha" {
			t.Fatalf("a coin of another country: %v", record)
		}
	}

	workbook := get("wishlist=false")
	if workbook.Code != http.StatusOK || workbook.Header().Get("Content-Type") != "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet" || !bytes.HasPrefix(workbook.Body.Bytes(), []byte("PK")) {
		t.Fatalf("xlsx by default: %d %q", workbook.Code, workbook.Header().Get("Content-Type"))
	}
	for _, bad := range []string{"format=pdf", "sort=price", "presentation=shoebox"} {
		if w := get(bad); w.Code != http.StatusBadRequest {
			t.Errorf("%s: %d", bad, w.Code)
		}
	}
}
