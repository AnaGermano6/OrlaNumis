package handlers

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/briandenicola/ancient-coins-api/services"
	"github.com/gin-gonic/gin"
)

// A release names its source, title, face value and month; a coin from
// Numista is its own type.
func TestCoinReleases_NumistaType(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router, _ := setupCoinHandlerRouter(t)
	handler := NewCoinReleasesHandler(services.NewCoinReleasesService(nil), nil, nil)
	group := router.Group("/api/releases-test")
	group.Use(coinTestAuthMiddleware())
	group.GET("/numista-type", handler.NumistaType)

	for _, query := range []string{
		"",
		"source=ecb&title=Rotary&faceValue=2",
		"source=ecb&title=Rotary&faceValue=0&from=2026-07",
		"source=elsewhere&title=Rotary&faceValue=2&from=2026-07",
		"source=ecb&title=&faceValue=2&from=2026-07",
	} {
		if w := coinRequest(router, http.MethodGet, "/api/releases-test/numista-type?"+query, nil); w.Code != http.StatusBadRequest {
			t.Errorf("%q: %d %s", query, w.Code, w.Body.String())
		}
	}

	w := coinRequest(router, http.MethodGet, "/api/releases-test/numista-type?source=numista&country=pt&title=Expo%2098&faceValue=1000&from=1998-01&pageUrl=https%3A%2F%2Fpt.numista.com%2Fcatalogue%2Fpieces9002.html", nil)
	var found services.ReleaseNumistaType
	if err := json.Unmarshal(w.Body.Bytes(), &found); err != nil || w.Code != http.StatusOK || found.TypeID != 9002 || found.Candidates == nil {
		t.Fatalf("a Numista coin: %d %s", w.Code, w.Body.String())
	}
}
