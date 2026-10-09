package handlers

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/briandenicola/ancient-coins-api/models"
	"github.com/briandenicola/ancient-coins-api/repository"
	"github.com/briandenicola/ancient-coins-api/services"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) { return f(req) }

func stubPhotoSource(t *testing.T, status int, contentType string, body []byte) *[]*http.Request {
	t.Helper()
	requests := &[]*http.Request{}
	old := outboundHTTPClientFactory
	outboundHTTPClientFactory = func() *http.Client {
		return &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			*requests = append(*requests, req)
			return &http.Response{
				StatusCode: status,
				Header:     http.Header{"Content-Type": {contentType}},
				Body:       io.NopCloser(bytes.NewReader(body)),
				Request:    req,
			}, nil
		})}
	}
	t.Cleanup(func() { outboundHTTPClientFactory = old })
	return requests
}

func setupImageImportRouter(t *testing.T) (*gin.Engine, *gorm.DB, uint) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	db := setupClipTestDB(t)
	user := models.User{Username: "importer", Email: "importer@example.test"}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
	coin := models.Coin{UserID: user.ID, Name: "50 Escudos", Notes: "Bought in Lisbon"}
	if err := db.Create(&coin).Error; err != nil {
		t.Fatalf("create coin: %v", err)
	}
	repo := repository.NewImageRepository(db)
	uploadDir := t.TempDir()
	handler := NewImageHandler(uploadDir, repo, services.NewImageService(repo, uploadDir), services.NewLogger(100))
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("userId", user.ID)
		c.Next()
	})
	router.POST("/coins/:id/images/import", handler.ImportRemote)
	return router, db, coin.ID
}

func postImport(router *gin.Engine, coinID uint, payload map[string]any) *httptest.ResponseRecorder {
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/coins/"+itoa(coinID)+"/images/import", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

func TestImportablePhotoURL(t *testing.T) {
	cases := map[string]bool{
		"https://upload.wikimedia.org/wikipedia/commons/5/54/Coin.jpg":               true,
		"https://thumb.wikimedia.org/wikipedia/commons/thumb/5/54/a.jpg/400px-a.jpg": true,
		"https://en.numista.com/catalogue/photos/portugal/3353-original.jpg":         true,
		"http://upload.wikimedia.org/wikipedia/commons/5/54/Coin.jpg":                false,
		"https://upload.wikimedia.org/other/path.jpg":                                false,
		"https://en.numista.com/forum/avatar.jpg":                                    false,
		"https://evil.example/wikipedia/commons/a.jpg":                               false,
		"https://user:pass@upload.wikimedia.org/wikipedia/commons/a.jpg":             false,
		"https://upload.wikimedia.org:8443/wikipedia/commons/a.jpg":                  false,
		"https://upload.wikimedia.org.evil.example/wikipedia/commons/a.jpg":          false,
	}
	for raw, want := range cases {
		if _, got := importablePhotoURL(raw); got != want {
			t.Errorf("importablePhotoURL(%q) = %v, want %v", raw, got, want)
		}
	}
}

func TestImageImport_SavesPhotoAndRecordsCredit(t *testing.T) {
	router, db, coinID := setupImageImportRouter(t)
	requests := stubPhotoSource(t, http.StatusOK, "image/jpeg", makeSyntheticJPEG(64, 64))

	w := postImport(router, coinID, map[string]any{
		"url":       "https://upload.wikimedia.org/wikipedia/commons/5/54/Coin.jpg",
		"imageType": "obverse",
		"isPrimary": true,
		"credit":    "Obverse photo: Jane Doe, CC BY-SA 4.0 (Wikimedia Commons)",
	})

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
	var image models.CoinImage
	if err := json.Unmarshal(w.Body.Bytes(), &image); err != nil || image.ImageType != "obverse" || !image.IsPrimary || !strings.HasSuffix(image.FilePath, ".jpg") {
		t.Fatalf("image = %+v (%v)", image, err)
	}
	if len(*requests) != 1 || (*requests)[0].Header.Get("User-Agent") != services.CommonsUserAgent {
		t.Fatalf("expected one identified request, got %d", len(*requests))
	}
	var coin models.Coin
	db.First(&coin, coinID)
	if coin.Notes != "Bought in Lisbon\nObverse photo: Jane Doe, CC BY-SA 4.0 (Wikimedia Commons)" {
		t.Fatalf("notes = %q", coin.Notes)
	}
}

func TestImageImport_RejectsUnsupportedSourcesAndReportsRefusals(t *testing.T) {
	router, db, coinID := setupImageImportRouter(t)
	requests := stubPhotoSource(t, http.StatusForbidden, "text/html", []byte("<html>challenge</html>"))

	w := postImport(router, coinID, map[string]any{"url": "https://evil.example/photo.jpg", "imageType": "obverse"})
	if w.Code != http.StatusBadRequest || len(*requests) != 0 {
		t.Fatalf("arbitrary host: %d with %d requests", w.Code, len(*requests))
	}
	if w := postImport(router, coinID, map[string]any{"url": "https://upload.wikimedia.org/wikipedia/commons/a.jpg", "imageType": "selfie"}); w.Code != http.StatusBadRequest {
		t.Fatalf("bad image type: %d", w.Code)
	}
	if w := postImport(router, 9999, map[string]any{"url": "https://upload.wikimedia.org/wikipedia/commons/a.jpg", "imageType": "obverse"}); w.Code != http.StatusNotFound {
		t.Fatalf("unknown coin: %d", w.Code)
	}

	w = postImport(router, coinID, map[string]any{
		"url": "https://en.numista.com/catalogue/photos/portugal/3353-original.jpg", "imageType": "obverse", "credit": "should not be recorded",
	})
	if w.Code != http.StatusBadGateway || !strings.Contains(w.Body.String(), `"code":"source_refused"`) {
		t.Fatalf("refused source: %d %s", w.Code, w.Body.String())
	}
	var coin models.Coin
	db.First(&coin, coinID)
	var images int64
	db.Model(&models.CoinImage{}).Where("coin_id = ?", coinID).Count(&images)
	if coin.Notes != "Bought in Lisbon" || images != 0 {
		t.Fatalf("a failed import must change nothing: notes %q, %d images", coin.Notes, images)
	}
}

func TestImageImport_RejectsNonImageContent(t *testing.T) {
	router, _, coinID := setupImageImportRouter(t)
	stubPhotoSource(t, http.StatusOK, "text/html; charset=utf-8", []byte("<html></html>"))

	w := postImport(router, coinID, map[string]any{"url": "https://upload.wikimedia.org/wikipedia/commons/a.jpg", "imageType": "reverse"})
	if w.Code != http.StatusBadGateway || !strings.Contains(w.Body.String(), `"code":"source_failed"`) {
		t.Fatalf("non-image: %d %s", w.Code, w.Body.String())
	}
}

func TestImageImport_RetriesOnceWhenRateLimited(t *testing.T) {
	router, _, coinID := setupImageImportRouter(t)
	oldDelay := photoRetryDefaultDelay
	photoRetryDefaultDelay = time.Millisecond
	t.Cleanup(func() { photoRetryDefaultDelay = oldDelay })

	calls := 0
	old := outboundHTTPClientFactory
	outboundHTTPClientFactory = func() *http.Client {
		return &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			calls++
			status, contentType, body := http.StatusTooManyRequests, "text/plain", []byte("slow down")
			if calls > 1 {
				status, contentType, body = http.StatusOK, "image/jpeg", makeSyntheticJPEG(32, 32)
			}
			return &http.Response{StatusCode: status, Header: http.Header{"Content-Type": {contentType}}, Body: io.NopCloser(bytes.NewReader(body)), Request: req}, nil
		})}
	}
	t.Cleanup(func() { outboundHTTPClientFactory = old })

	w := postImport(router, coinID, map[string]any{"url": "https://upload.wikimedia.org/wikipedia/commons/a/aa/O.jpg", "imageType": "obverse"})
	if w.Code != http.StatusCreated || calls != 2 {
		t.Fatalf("expected a successful retry, got %d after %d calls: %s", w.Code, calls, w.Body.String())
	}
}
