package handlers

import (
	"bytes"
	"encoding/json"
	"image"
	"image/color"
	"image/jpeg"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/briandenicola/ancient-coins-api/models"
	"github.com/briandenicola/ancient-coins-api/repository"
	"github.com/briandenicola/ancient-coins-api/services"
	"github.com/gin-gonic/gin"
)

func TestShareListHandler_DownloadsPDFAndCSV(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupCoinHandlerTestDB(t)
	uploads := t.TempDir()

	// A photo with its small copy, and one whose file is gone.
	if err := os.MkdirAll(filepath.Join(uploads, "coins"), 0o755); err != nil {
		t.Fatal(err)
	}
	picture := image.NewRGBA(image.Rect(0, 0, 40, 40))
	for x := 0; x < 40; x++ {
		picture.Set(x, x, color.RGBA{R: 200, G: 150, B: 40, A: 255})
	}
	file, err := os.Create(filepath.Join(uploads, "coins", "escudo_thumb.jpg"))
	if err != nil {
		t.Fatal(err)
	}
	if err := jpeg.Encode(file, picture, nil); err != nil {
		t.Fatal(err)
	}
	file.Close()

	coins := []*models.Coin{
		{Name: "1 Escudo 1986 — Açores", UserID: 1, Country: "Portugal", Denomination: "1 Escudo", Grade: "UNC", Quantity: 4},
		{Name: "2 Euro 2007", UserID: 1, Country: "Áustria", Quantity: 2},
		{Name: "Coroa de D. João V", UserID: 1, Country: "Portugal", IsWishlist: true},
	}
	for _, coin := range coins {
		if err := db.Create(coin).Error; err != nil {
			t.Fatal(err)
		}
	}
	db.Create(&models.CoinImage{CoinID: coins[0].ID, FilePath: "coins/escudo.jpg", ImageType: models.ImageTypeObverse})
	db.Create(&models.CoinImage{CoinID: coins[1].ID, FilePath: "coins/missing.jpg", ImageType: models.ImageTypeObverse})
	db.Create(&models.CoinImage{CoinID: coins[2].ID, FilePath: "../../etc/passwd", ImageType: models.ImageTypeObverse})

	handler := NewShareListHandler(services.NewShareListService(repository.NewCoinRepository(db), nil), uploads, services.NewLogger(10))
	router := gin.New()
	api := router.Group("/api")
	api.Use(coinTestAuthMiddleware())
	api.POST("/coins/share-list", handler.Download)
	post := func(body any) *httptest.ResponseRecorder {
		payload, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPost, "/api/coins/share-list", bytes.NewReader(payload))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", authHeader(1))
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		return w
	}

	w := post(map[string]any{"list": "both", "format": "pdf", "lang": "pt", "includeValues": true, "contact": "ana@exemplo.pt"})
	if w.Code != http.StatusOK || w.Header().Get("Content-Type") != "application/pdf" || !strings.HasPrefix(w.Body.String(), "%PDF-") {
		t.Fatalf("pdf: %d %q %.40q", w.Code, w.Header().Get("Content-Type"), w.Body.String())
	}
	if !strings.Contains(w.Header().Get("Content-Disposition"), `filename="trocas-e-procuras-`) {
		t.Fatalf("pdf name: %q", w.Header().Get("Content-Disposition"))
	}
	// The small photo is in; the missing one and the one outside the folder are not.
	if !bytes.Contains(w.Body.Bytes(), []byte("/Subtype /Image")) {
		t.Fatal("the PDF should hold the coin's photo")
	}

	w = post(map[string]any{"list": "trades", "format": "csv", "lang": "en"})
	if w.Code != http.StatusOK || !strings.HasPrefix(w.Header().Get("Content-Type"), "text/csv") || !strings.Contains(w.Body.String(), "1 Escudo 1986 — Açores") {
		t.Fatalf("csv: %d %s", w.Code, w.Body.String())
	}
	if strings.Contains(w.Body.String(), "Coroa") {
		t.Fatal("the trade list has no wanted coin")
	}

	for _, bad := range []map[string]any{{"list": "all", "format": "pdf"}, {"list": "trades", "format": "doc"}, {"format": "pdf"}} {
		if w := post(bad); w.Code != http.StatusBadRequest {
			t.Errorf("%v: %d", bad, w.Code)
		}
	}
}

func TestShareListPhotoPath_StaysInTheUploadFolder(t *testing.T) {
	root := t.TempDir()
	if got := shareListPhotoPath(root, "../outside.jpg"); got != "" {
		t.Fatalf("outside the folder: %q", got)
	}
	if got := shareListPhotoPath(root, ""); got != "" {
		t.Fatalf("no photo: %q", got)
	}
	path := filepath.Join(root, "a.png")
	if err := os.WriteFile(path, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := shareListPhotoPath(root, "a.png"); got != path {
		t.Fatalf("without a small copy, the photo itself: %q", got)
	}
}
