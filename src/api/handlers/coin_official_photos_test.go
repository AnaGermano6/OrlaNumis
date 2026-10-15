package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/briandenicola/ancient-coins-api/models"
	"github.com/briandenicola/ancient-coins-api/repository"
	"github.com/briandenicola/ancient-coins-api/services"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

var bothSides = []models.ImageType{models.ImageTypeObverse, models.ImageTypeReverse}

func setupOfficialPhotos(t *testing.T) (*OfficialPhotoHandler, *gorm.DB) {
	t.Helper()
	db := setupCoinHandlerTestDB(t)
	if err := db.AutoMigrate(&models.NumistaCacheEntry{}); err != nil {
		t.Fatal(err)
	}
	createTestUser(t, db, 1, "collector")
	cache := services.NewNumistaResponseCache(repository.NewNumistaCacheRepository(db), nil, nil)
	photos := services.NewCoinPhotoSearchService(nil, stubCommonsClient{}, cache).WithECB(services.NewEcbCoinPhotoService(stubECBCoinsClient{}, cache))
	imageRepo := repository.NewImageRepository(db)
	uploadDir := t.TempDir()
	images := NewImageHandler(uploadDir, imageRepo, services.NewImageService(imageRepo, uploadDir), services.NewLogger(100))
	return NewOfficialPhotoHandler(repository.NewCoinRepository(db), images, photos, services.NewLogger(100)), db
}

func coinImages(t *testing.T, db *gorm.DB, coinID uint) map[models.ImageType]models.CoinImage {
	t.Helper()
	var images []models.CoinImage
	db.Where("coin_id = ?", coinID).Find(&images)
	sides := map[models.ImageType]models.CoinImage{}
	for _, image := range images {
		sides[image.ImageType] = image
	}
	if len(sides) != len(images) {
		t.Fatalf("coin %d has the same side twice: %+v", coinID, images)
	}
	return sides
}

func downloadedURLs(requests *[]*http.Request) string {
	var urls []string
	for _, request := range *requests {
		urls = append(urls, request.URL.String())
	}
	return strings.Join(urls, " ")
}

func TestOfficialPhotos_AddsACommemorativesECBImages(t *testing.T) {
	handler, db := setupOfficialPhotos(t)
	requests := stubPhotoSource(t, http.StatusOK, "image/jpeg", makeSyntheticJPEG(64, 64))
	year := 2022
	thuringia := models.Coin{Name: "2 euro 2022 D – Bundesländer, Thuringia", Country: "Alemanha", Denomination: "2 euro", Year: &year, UserID: 1, Notes: "Bought in Erfurt."}
	fiveEuro := models.Coin{Name: "5 euro 2022 – Thuringia", Country: "Alemanha", Denomination: "5 euro", Year: &year, UserID: 1}
	for _, coin := range []*models.Coin{&thuringia, &fiveEuro} {
		if err := db.Create(coin).Error; err != nil {
			t.Fatal(err)
		}
	}

	added, found := handler.Attach(context.Background(), thuringia.ID, 1, bothSides)
	if !found || len(added) != 2 {
		t.Fatalf("added = %+v, found = %v", added, found)
	}
	sides := coinImages(t, db, thuringia.ID)
	if len(sides) != 2 || !sides[models.ImageTypeObverse].IsPrimary || sides[models.ImageTypeReverse].IsPrimary {
		t.Fatalf("images = %+v", sides)
	}
	if got := downloadedURLs(requests); got != "https://www.ecb.europa.eu/euro/coins/comm/html/comm_2022/DE-thueringen.jpg https://www.ecb.europa.eu/euro/coins/common/shared/img/newcommon_2euro_800.jpg" {
		t.Fatalf("downloads = %s", got)
	}
	var coin models.Coin
	db.First(&coin, thuringia.ID)
	if !strings.HasPrefix(coin.Notes, "Bought in Erfurt.\n© European Central Bank – https://www.ecb.europa.eu/euro/coins/comm/") ||
		!strings.HasSuffix(coin.Notes, "\n© European Central Bank – https://www.ecb.europa.eu/euro/coins/common/html/index.en.html") {
		t.Fatalf("notes = %q", coin.Notes)
	}

	// Sides that have an image, and collector coins, get nothing.
	for _, id := range []uint{thuringia.ID, fiveEuro.ID} {
		if added, _ := handler.Attach(context.Background(), id, 1, bothSides); len(added) != 0 {
			t.Errorf("coin %d should get no image, got %+v", id, added)
		}
	}
	if _, found := handler.Attach(context.Background(), thuringia.ID, 2, bothSides); found {
		t.Error("another user's coin must not be found")
	}
	if len(*requests) != 2 {
		t.Fatalf("no further downloads expected, got %d", len(*requests))
	}
}

func TestOfficialPhotos_FillsOnlyTheSidesAskedFor(t *testing.T) {
	handler, db := setupOfficialPhotos(t)
	requests := stubPhotoSource(t, http.StatusOK, "image/jpeg", makeSyntheticJPEG(64, 64))
	year, old := 2022, 2002
	regular := models.Coin{Name: "2 euro 2022 – Alemanha", Country: "Alemanha", Denomination: "2 euro", Year: &year, UserID: 1}
	cent := models.Coin{Name: "10 cêntimos 2002 – Portugal", Country: "Portugal", Denomination: "10 cêntimos", Year: &old, UserID: 1}
	another := models.Coin{Name: "2 euro 2022 – França", Country: "França", Denomination: "2 euro", Year: &year, UserID: 1}
	for _, coin := range []*models.Coin{&regular, &cent, &another} {
		if err := db.Create(coin).Error; err != nil {
			t.Fatal(err)
		}
	}

	// A regular coin gets its country's design and the common side.
	if added, _ := handler.Attach(context.Background(), regular.ID, 1, bothSides); len(added) != 2 {
		t.Fatalf("regular coin images = %+v", added)
	}
	if got := downloadedURLs(requests); got != "https://www.ecb.europa.eu/euro/coins/common/shared/img/de/Germany_2euro.jpg https://www.ecb.europa.eu/euro/coins/common/shared/img/newcommon_2euro_800.jpg" {
		t.Fatalf("downloads = %s", got)
	}

	// The user photographed the obverse: only the reverse is asked for.
	if err := db.Create(&models.CoinImage{CoinID: cent.ID, FilePath: "mine.jpg", ImageType: models.ImageTypeObverse, IsPrimary: true}).Error; err != nil {
		t.Fatal(err)
	}
	added, _ := handler.Attach(context.Background(), cent.ID, 1, []models.ImageType{models.ImageTypeReverse})
	if len(added) != 1 || added[0].ImageType != models.ImageTypeReverse || added[0].IsPrimary {
		t.Fatalf("cent images = %+v", added)
	}
	if sides := coinImages(t, db, cent.ID); sides[models.ImageTypeObverse].FilePath != "mine.jpg" {
		t.Fatalf("the user's photo must stay: %+v", sides)
	}

	// Only the obverse asked for; the common side is downloaded once anyway.
	before := len(*requests)
	if added, _ := handler.Attach(context.Background(), another.ID, 1, []models.ImageType{models.ImageTypeObverse}); len(added) != 1 || !added[0].IsPrimary {
		t.Fatalf("obverse only = %+v", added)
	}
	if added, _ := handler.Attach(context.Background(), another.ID, 1, []models.ImageType{models.ImageTypeReverse}); len(added) != 1 || len(*requests) != before+1 {
		t.Fatalf("the cached common side should be reused: %+v, downloads %d", added, len(*requests)-before)
	}
}

func TestOfficialPhotos_Endpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler, db := setupOfficialPhotos(t)
	stubPhotoSource(t, http.StatusOK, "image/jpeg", makeSyntheticJPEG(64, 64))
	year := 2022
	coin := models.Coin{Name: "2 euro 2022 – Alemanha", Country: "Alemanha", Denomination: "2 euro", Year: &year, UserID: 1}
	db.Create(&coin)
	router := gin.New()
	group := router.Group("/api")
	group.Use(coinTestAuthMiddleware())
	group.POST("/coins/:id/official-photos", handler.Add)
	path := fmt.Sprintf("/api/coins/%d/official-photos", coin.ID)

	for _, body := range []any{nil, map[string]any{"sides": []string{}}, map[string]any{"sides": []string{"edge"}}} {
		if w := coinRequest(router, http.MethodPost, path, body); w.Code != http.StatusBadRequest {
			t.Errorf("%v: expected 400, got %d", body, w.Code)
		}
	}
	if w := coinRequest(router, http.MethodPost, "/api/coins/999/official-photos", map[string]any{"sides": []string{"obverse"}}); w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
	w := coinRequest(router, http.MethodPost, path, map[string]any{"sides": []string{"obverse", "reverse"}})
	var got OfficialPhotosResponse
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil || w.Code != http.StatusOK || len(got.Images) != 2 {
		t.Fatalf("add: %d %s", w.Code, w.Body.String())
	}
}

// The coins without any photo get their official images a batch at a time:
// those that have one or are sold, banknotes and another user's coins are
// left alone, and a coin with no official image is passed over.
func TestOfficialPhotos_Batch(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler, db := setupOfficialPhotos(t)
	stubPhotoSource(t, http.StatusOK, "image/jpeg", makeSyntheticJPEG(64, 64))
	year := 2022
	newCoin := func(coin models.Coin) models.Coin {
		coin.Year = &year
		if coin.UserID == 0 {
			coin.UserID = 1
		}
		if err := db.Create(&coin).Error; err != nil {
			t.Fatal(err)
		}
		return coin
	}
	germany := newCoin(models.Coin{Name: "2 euro 2022 – Alemanha", Country: "Alemanha", Denomination: "2 euro"})
	collector := newCoin(models.Coin{Name: "5 euro 2022 – Thuringia", Country: "Alemanha", Denomination: "5 euro"})
	france := newCoin(models.Coin{Name: "2 euro 2022 – França", Country: "França", Denomination: "2 euro"})
	photographed := newCoin(models.Coin{Name: "2 euro 2022 – Itália", Country: "Itália", Denomination: "2 euro"})
	db.Create(&models.CoinImage{CoinID: photographed.ID, FilePath: "mine.jpg", ImageType: models.ImageTypeObverse, IsPrimary: true})
	newCoin(models.Coin{Name: "2 euro 2022 – Bélgica", Country: "Bélgica", Denomination: "2 euro", IsSold: true})
	newCoin(models.Coin{Name: "5 euro note", Country: "Portugal", Denomination: "5 euro", Kind: models.CoinKindBanknote})
	newCoin(models.Coin{Name: "2 euro 2022 – Áustria", Country: "Áustria", Denomination: "2 euro", UserID: 2})

	router := gin.New()
	group := router.Group("/api")
	group.Use(coinTestAuthMiddleware())
	group.GET("/coins/official-photos/missing", handler.Missing)
	group.POST("/coins/official-photos/batch", handler.Batch)
	missing := func() int64 {
		var got MissingPhotosResponse
		w := coinRequest(router, http.MethodGet, "/api/coins/official-photos/missing", nil)
		if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil || w.Code != http.StatusOK {
			t.Fatalf("missing: %d %s", w.Code, w.Body.String())
		}
		return got.Count
	}
	if count := missing(); count != 3 {
		t.Fatalf("coins without photos = %d, want 3", count)
	}
	if w := coinRequest(router, http.MethodPost, "/api/coins/official-photos/batch", map[string]any{"limit": 21}); w.Code != http.StatusBadRequest {
		t.Fatalf("limit 21: %d", w.Code)
	}

	batch := func(afterID uint) OfficialPhotosBatchResponse {
		var got OfficialPhotosBatchResponse
		w := coinRequest(router, http.MethodPost, "/api/coins/official-photos/batch", map[string]any{"afterId": afterID, "limit": 2})
		if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil || w.Code != http.StatusOK {
			t.Fatalf("batch: %d %s", w.Code, w.Body.String())
		}
		return got
	}
	first := batch(0)
	if first.Checked != 2 || first.WithPhotos != 1 || first.Images != 2 || first.NextAfterID != collector.ID || first.Done {
		t.Fatalf("first batch = %+v", first)
	}
	second := batch(first.NextAfterID)
	if second.Checked != 1 || second.WithPhotos != 1 || second.NextAfterID != france.ID || !second.Done {
		t.Fatalf("second batch = %+v", second)
	}
	if len(coinImages(t, db, germany.ID)) != 2 || len(coinImages(t, db, collector.ID)) != 0 {
		t.Fatal("the euro coins get both sides, the collector coin none")
	}
	// The collector coin, which has no official image, is the one left.
	if count := missing(); count != 1 {
		t.Fatalf("coins still without photos = %d, want 1", count)
	}
}

type fixtureINCMShop struct{ body []byte }

func (f fixtureINCMShop) ProductsPage(_ context.Context, _ string, page int) ([]byte, error) {
	if page > 1 {
		return []byte(`{"products": []}`), nil
	}
	return f.body, nil
}

func TestOfficialPhotos_AddsTheINCMPhotosOfAPortugueseCollectorCoin(t *testing.T) {
	handler, db := setupOfficialPhotos(t)
	body, err := os.ReadFile("../services/testdata/incm_products.json")
	if err != nil {
		t.Fatal(err)
	}
	cache := services.NewNumistaResponseCache(repository.NewNumistaCacheRepository(db), nil, nil)
	handler.photos.WithMint(services.NewIncmCoinPhotoService(fixtureINCMShop{body: body}, cache))
	requests := stubPhotoSource(t, http.StatusOK, "image/jpeg", makeSyntheticJPEG(64, 64))
	year := 2025
	coin := models.Coin{Name: "5 euro 2025 – Fénix", Country: "Portugal", Denomination: "5 euro", Year: &year, Material: models.MaterialOther, UserID: 1}
	if err := db.Create(&coin).Error; err != nil {
		t.Fatal(err)
	}

	added, _ := handler.Attach(context.Background(), coin.ID, 1, bothSides)
	if len(added) != 2 || !added[0].IsPrimary {
		t.Fatalf("added = %+v", added)
	}
	got := downloadedURLs(requests)
	if !strings.Contains(got, "https://cdn.shopify.com/s/files/1/0525/1910/8762/files/FENIX_CU_A.jpg") || !strings.Contains(got, "FENIX_CU_b.jpg") {
		t.Fatalf("downloads = %s", got)
	}
	var saved models.Coin
	db.First(&saved, coin.ID)
	if !strings.Contains(saved.Notes, "© INCM – https://loja.incm.pt/products/fenix-herois-e-criaturas-da-mitologia-12500800") {
		t.Fatalf("notes = %q", saved.Notes)
	}
}

func TestImportablePhotoURL_OnlyTheINCMShopOnShopify(t *testing.T) {
	for raw, want := range map[string]bool{
		"https://cdn.shopify.com/s/files/1/0525/1910/8762/files/FENIX_CU_A.jpg?v=1":                                                                                 true,
		"https://cdn.shopify.com/s/files/1/0000/0000/0000/files/other.jpg":                                                                                          false,
		"https://cdn.shopify.com/s/files/1/0525/1910/8762/files/notes.pdf":                                                                                          false,
		"https://www.monnaiedeparis.fr/media/catalog/product/M/5/M5066_MDP_1_BD.jpg":                                                                                true,
		"https://www.monnaiedeparis.fr/en/shop":                                                                                                                     false,
		"https://www.muenzeoesterreich.at/var/storage/images/_aliases/product_full/8/6/3/7/9527368-3-ger-DE/1729ea87f981-2024_5E_Weltmeister_Ag_HGH_VS_2D.png.webp": true,
		"https://www.muenzeoesterreich.at/produkte/5-euro-kupfermuenze-weltmeister-np":                                                                              false,
	} {
		if _, ok := importablePhotoURL(raw); ok != want {
			t.Errorf("importablePhotoURL(%q) = %v, want %v", raw, ok, want)
		}
	}
}

func TestImageHandler_SwapSides(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler, db := setupOfficialPhotos(t)
	coin := models.Coin{Name: "5 euro 2024 – O Conhecimento", UserID: 1}
	db.Create(&coin)
	other := models.Coin{Name: "Other", UserID: 2}
	db.Create(&other)
	for _, image := range []models.CoinImage{
		{CoinID: coin.ID, FilePath: "a.jpg", ImageType: models.ImageTypeObverse, IsPrimary: true},
		{CoinID: coin.ID, FilePath: "b.jpg", ImageType: models.ImageTypeReverse},
		{CoinID: coin.ID, FilePath: "c.jpg", ImageType: models.ImageTypeDetail},
	} {
		db.Create(&image)
	}
	router := gin.New()
	group := router.Group("/api")
	group.Use(coinTestAuthMiddleware())
	group.POST("/coins/:id/images/swap-sides", handler.images.SwapSides)

	if w := coinRequest(router, http.MethodPost, fmt.Sprintf("/api/coins/%d/images/swap-sides", other.ID), nil); w.Code != http.StatusNotFound {
		t.Fatalf("another user's coin: %d", w.Code)
	}
	w := coinRequest(router, http.MethodPost, fmt.Sprintf("/api/coins/%d/images/swap-sides", coin.ID), nil)
	var images []models.CoinImage
	if err := json.Unmarshal(w.Body.Bytes(), &images); err != nil || w.Code != http.StatusOK || len(images) != 3 {
		t.Fatalf("swap: %d %s", w.Code, w.Body.String())
	}
	sides := map[string]models.CoinImage{}
	for _, image := range images {
		sides[image.FilePath] = image
	}
	if sides["a.jpg"].ImageType != models.ImageTypeReverse || !sides["a.jpg"].IsPrimary || sides["b.jpg"].ImageType != models.ImageTypeObverse || sides["c.jpg"].ImageType != models.ImageTypeDetail {
		t.Fatalf("images = %+v", images)
	}
}
