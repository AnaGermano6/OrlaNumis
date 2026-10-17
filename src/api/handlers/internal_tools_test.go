package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/briandenicola/ancient-coins-api/models"
	"github.com/briandenicola/ancient-coins-api/repository"
	"github.com/briandenicola/ancient-coins-api/services"
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

// fakeDeepNumistaClient is an in-memory stand-in for services.NumistaClient
// used only by these handler tests - no live HTTP is ever reached
// (T054: no-network CI requirement).
type fakeDeepNumistaClient struct {
	searchCalls atomic.Int32
	candidates  []models.NumistaCandidate
	err         error
}

func (f *fakeDeepNumistaClient) Search(_ context.Context, _ string, _ int) ([]models.NumistaCandidate, error) {
	f.searchCalls.Add(1)
	if f.err != nil {
		return nil, f.err
	}
	return f.candidates, nil
}

func (f *fakeDeepNumistaClient) Detail(_ context.Context, id int) (models.NumistaCandidate, error) {
	if f.err != nil {
		return models.NumistaCandidate{}, f.err
	}
	return models.NumistaCandidate{ID: id, Title: "Test Coin"}, nil
}

var deepProviderToolsTestDBCounter int64

func setupDeepProviderToolsTest(t *testing.T, numistaBudget int) (*gin.Engine, *services.InternalTokenService, *fakeDeepNumistaClient) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	dsn := fmt.Sprintf("file:deep_provider_tools_%d_%d?mode=memory&cache=shared", time.Now().UnixNano(), atomic.AddInt64(&deepProviderToolsTestDBCounter, 1))
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := db.AutoMigrate(&models.AppSetting{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	if err := db.Create(&models.AppSetting{
		Key: services.SettingDeepIdentificationNumistaCallBudget, Value: fmt.Sprintf("%d", numistaBudget),
	}).Error; err != nil {
		t.Fatalf("seed numista budget: %v", err)
	}
	settingsSvc := services.NewSettingsService(repository.NewSettingsRepository(db))

	numistaClient := &fakeDeepNumistaClient{candidates: []models.NumistaCandidate{{ID: 1, Title: "Denarius"}}}
	tokenSvc := services.NewInternalTokenService("test-secret")
	budgets := services.NewDeepProviderBudgetTracker()
	handler := NewDeepProviderToolsHandler(numistaClient, settingsSvc, budgets, services.NewLogger(10))

	router := gin.New()
	authed := router.Group("/api/internal/tools")
	authed.Use(func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		tokenString := authHeader
		if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
			tokenString = authHeader[7:]
		}
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization required"})
			return
		}
		userID, jobID, err := tokenSvc.VerifyForJob(tokenString)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			return
		}
		c.Set("userId", userID)
		c.Set("deepJobId", jobID)
		c.Next()
	})
	authed.POST("/numista_search", handler.NumistaSearch)
	authed.POST("/numista_detail", handler.NumistaDetail)

	return router, tokenSvc, numistaClient
}

func TestDeepProviderTools_UnauthenticatedRequestReturns401(t *testing.T) {
	router, _, _ := setupDeepProviderToolsTest(t, 4)

	req := httptest.NewRequest(http.MethodPost, "/api/internal/tools/numista_search", bytes.NewBufferString(`{"query":"denarius"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestDeepProviderTools_ForeignOrTamperedTokenReturns401(t *testing.T) {
	router, tokenSvc, _ := setupDeepProviderToolsTest(t, 4)

	token, err := tokenSvc.MintForJob(1, 5)
	if err != nil {
		t.Fatalf("MintForJob: %v", err)
	}
	tampered := token[:len(token)-1] + "z"

	req := httptest.NewRequest(http.MethodPost, "/api/internal/tools/numista_search", bytes.NewBufferString(`{"query":"denarius"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+tampered)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for tampered/foreign token, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestDeepProviderTools_NumistaSearchHappyPath(t *testing.T) {
	router, tokenSvc, numistaClient := setupDeepProviderToolsTest(t, 4)
	token, _ := tokenSvc.MintForJob(1, 5)

	req := httptest.NewRequest(http.MethodPost, "/api/internal/tools/numista_search", bytes.NewBufferString(`{"query":"denarius","limit":5}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var resp map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if resp["status"] != "ok" {
		t.Fatalf("expected status=ok, got %v", resp["status"])
	}
	if numistaClient.searchCalls.Load() != 1 {
		t.Fatalf("expected exactly one upstream call, got %d", numistaClient.searchCalls.Load())
	}
}

func TestDeepProviderTools_NumistaSearchBudgetExceededReturnsQuotaLimited(t *testing.T) {
	router, tokenSvc, numistaClient := setupDeepProviderToolsTest(t, 2)
	token, _ := tokenSvc.MintForJob(1, 5)

	call := func() (int, map[string]any) {
		req := httptest.NewRequest(http.MethodPost, "/api/internal/tools/numista_search", bytes.NewBufferString(`{"query":"denarius"}`))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		var resp map[string]any
		if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		return rec.Code, resp
	}

	for i := 0; i < 2; i++ {
		code, resp := call()
		if code != http.StatusOK || resp["status"] != "ok" {
			t.Fatalf("call %d: expected 200/ok, got %d/%v", i, code, resp["status"])
		}
	}

	code, resp := call()
	if code != http.StatusOK {
		t.Fatalf("expected 200 (quota_limited is a typed status, not an HTTP error), got %d", code)
	}
	if resp["status"] != "quota_limited" {
		t.Fatalf("expected status=quota_limited after budget exhausted, got %v", resp["status"])
	}
	if numistaClient.searchCalls.Load() != 2 {
		t.Fatalf("expected the upstream client to be called only twice (3rd call short-circuited), got %d", numistaClient.searchCalls.Load())
	}
}

func TestDeepProviderTools_NumistaDetailUnconfiguredMapsToTypedStatus(t *testing.T) {
	router, tokenSvc, numistaClient := setupDeepProviderToolsTest(t, 4)
	numistaClient.err = &services.NumistaError{Kind: services.NumistaErrorUnconfigured}
	token, _ := tokenSvc.MintForJob(1, 5)

	req := httptest.NewRequest(http.MethodPost, "/api/internal/tools/numista_detail", bytes.NewBufferString(`{"id":123}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var resp map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if resp["status"] != "unconfigured" {
		t.Fatalf("expected status=unconfigured, got %v", resp["status"])
	}
}
