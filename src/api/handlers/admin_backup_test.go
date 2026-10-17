package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/briandenicola/ancient-coins-api/models"
	"github.com/briandenicola/ancient-coins-api/repository"
	"github.com/briandenicola/ancient-coins-api/services"
	"github.com/gin-gonic/gin"
)

// An administrator makes a backup, downloads it, imports it back, restores
// it (the app then restarts) and deletes one.
func TestAdminBackups(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router, db := setupCoinHandlerRouter(t)
	createTestUser(t, db, 1, "collector")
	if err := db.Create(&models.Coin{Name: "2 euro 2022", UserID: 1, Category: models.CategoryModern, Material: models.MaterialOther}).Error; err != nil {
		t.Fatal(err)
	}
	root := t.TempDir()
	uploads := filepath.Join(root, "uploads")
	if err := os.MkdirAll(filepath.Join(uploads, "coin-1"), 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(uploads, "coin-1", "obverse.jpg"), []byte("jpg"), 0o640); err != nil {
		t.Fatal(err)
	}
	settings := services.NewSettingsService(repository.NewSettingsRepository(db))
	check := func(path string) error {
		data, err := os.ReadFile(path)
		if err != nil || !bytes.HasPrefix(data, []byte("SQLite format 3\x00")) {
			return errors.New("not a database")
		}
		return nil
	}
	backups := services.NewBackupService(repository.NewBackupRepository(db), filepath.Join(root, "data", "ancientcoins.db"), uploads, filepath.Join(root, "backups"), "4.3.0", settings, services.NewLogger(100), check)
	restarted := 0
	handler := NewAdminBackupHandler(backups, services.NewBackupScheduler(backups, settings, services.NewLogger(100)), settings, "~/Aurearia-backups/test", func() { restarted++ })
	group := router.Group("/api/backups-test")
	group.Use(coinTestAuthMiddleware())
	group.GET("", handler.List)
	group.POST("", handler.Create)
	group.GET("/:name", handler.Download)
	group.DELETE("/:name", handler.Delete)
	group.POST("/:name/restore", handler.Restore)
	group.PUT("/upload", handler.Upload)
	list := func() BackupsResponse {
		t.Helper()
		w := coinRequest(router, http.MethodGet, "/api/backups-test", nil)
		var response BackupsResponse
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil || w.Code != http.StatusOK {
			t.Fatalf("list: %d %s", w.Code, w.Body.String())
		}
		return response
	}

	if response := list(); len(response.Backups) != 0 || !response.Enabled || response.StartTime != "03:00" || response.Keep != 7 || response.Folder != "~/Aurearia-backups/test" {
		t.Fatalf("before = %+v", response)
	}
	if w := coinRequest(router, http.MethodPost, "/api/backups-test", nil); w.Code != http.StatusAccepted {
		t.Fatalf("create: %d %s", w.Code, w.Body.String())
	}
	for deadline := time.Now().Add(5 * time.Second); backups.Running() && time.Now().Before(deadline); {
		time.Sleep(10 * time.Millisecond)
	}
	response := list()
	if len(response.Backups) != 1 || response.Running || response.Backups[0].Kind != services.BackupKindManual || response.Backups[0].Files != 1 {
		t.Fatalf("after = %+v", response)
	}
	name := response.Backups[0].Name

	w := coinRequest(router, http.MethodGet, "/api/backups-test/"+name, nil)
	if w.Code != http.StatusOK || !bytes.HasPrefix(w.Body.Bytes(), []byte("PK")) {
		t.Fatalf("download: %d", w.Code)
	}
	zipBytes := w.Body.Bytes()
	if w := coinRequest(router, http.MethodGet, "/api/backups-test/..%2Fdata%2Fancientcoins.db", nil); w.Code != http.StatusNotFound {
		t.Fatalf("download outside the folder: %d", w.Code)
	}

	upload := func(body []byte) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodPut, "/api/backups-test/upload", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/zip")
		req.Header.Set("Authorization", authHeader(1))
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		return w
	}
	if w := upload([]byte("not a backup")); w.Code != http.StatusBadRequest {
		t.Fatalf("upload of another file: %d", w.Code)
	}
	w = upload(zipBytes)
	var imported services.BackupInfo
	if err := json.Unmarshal(w.Body.Bytes(), &imported); err != nil || w.Code != http.StatusCreated || imported.Kind != services.BackupKindUploaded {
		t.Fatalf("upload: %d %s", w.Code, w.Body.String())
	}

	if w := coinRequest(router, http.MethodPost, "/api/backups-test/aurearia-2020-01-01-000000-auto.zip/restore", nil); w.Code != http.StatusNotFound || restarted != 0 {
		t.Fatalf("restore of no backup: %d, restarted %d", w.Code, restarted)
	}
	if w := coinRequest(router, http.MethodPost, "/api/backups-test/"+imported.Name+"/restore", nil); w.Code != http.StatusAccepted || restarted != 1 {
		t.Fatalf("restore: %d %s, restarted %d", w.Code, w.Body.String(), restarted)
	}
	if _, err := os.Stat(filepath.Join(root, "data", "restore-pending", "request.json")); err != nil {
		t.Fatalf("no restore prepared: %v", err)
	}
	// The data of now was backed up first.
	kinds := map[string]int{}
	for _, backup := range list().Backups {
		kinds[backup.Kind]++
	}
	if len(kinds) != 3 || kinds[services.BackupKindPreRestore] != 1 {
		t.Fatalf("after preparing the restore: %v", kinds)
	}

	if w := coinRequest(router, http.MethodDelete, "/api/backups-test/"+name, nil); w.Code != http.StatusNoContent {
		t.Fatalf("delete: %d", w.Code)
	}
	if w := coinRequest(router, http.MethodDelete, "/api/backups-test/"+name, nil); w.Code != http.StatusNotFound {
		t.Fatalf("delete again: %d", w.Code)
	}
}
