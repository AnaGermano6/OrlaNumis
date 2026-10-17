package handlers

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/briandenicola/ancient-coins-api/models"
	"github.com/briandenicola/ancient-coins-api/repository"
	"github.com/briandenicola/ancient-coins-api/services"
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func setupSecurityAdminHandlerTest(t *testing.T) (*gin.Engine, *gorm.DB) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	if err := db.AutoMigrate(&models.User{}, &models.AppSetting{}, &models.SecurityEvent{}); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}
	securitySvc := services.NewSecurityService(repository.NewSecurityRepository(db))
	settingsSvc := services.NewSettingsService(repository.NewSettingsRepository(db))
	handler := NewSecurityAdminHandler(securitySvc, settingsSvc, SecurityExposureConfig{})
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("userId", uint(1))
		c.Next()
	})
	router.GET("/api/admin/security/summary", handler.SecuritySummary)
	router.GET("/api/admin/security/events", handler.SecurityEvents)
	router.GET("/api/admin/security/exposure-check", handler.ExposureCheck)
	return router, db
}

func TestSecurityAdminHandlerEvents(t *testing.T) {
	router, db := setupSecurityAdminHandlerTest(t)
	db.Create(&models.SecurityEvent{Type: models.SecurityEventPasswordLoginFailure, Username: "alice", ClientIP: "198.51.100.7"})

	req := httptest.NewRequest(http.MethodGet, "/api/admin/security/events?type=password_login_failure", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected events 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestSecurityAdminHandlerExposureCheckIncludesRegistrationMode(t *testing.T) {
	router, _ := setupSecurityAdminHandlerTest(t)
	req := httptest.NewRequest(http.MethodGet, "/api/admin/security/exposure-check", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected exposure check 200, got %d: %s", w.Code, w.Body.String())
	}
	if !bytes.Contains(w.Body.Bytes(), []byte(`"registrationMode":"closed"`)) {
		t.Fatalf("expected registration mode in exposure check: %s", w.Body.String())
	}
}
