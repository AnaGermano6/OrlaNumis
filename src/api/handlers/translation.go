package handlers

import (
	"net/http"
	"strings"

	"github.com/briandenicola/ancient-coins-api/services"
	"github.com/gin-gonic/gin"
)

// TranslationHandler translates short texts for the client, such as a coin's
// subject written in Portuguese to compare with Numista's English titles.
type TranslationHandler struct {
	svc *services.TranslationService
}

// NewTranslationHandler creates the handler.
func NewTranslationHandler(svc *services.TranslationService) *TranslationHandler {
	return &TranslationHandler{svc: svc}
}

// TranslationResult is a translated text; status is "unavailable" when the
// translation service could not answer.
type TranslationResult struct {
	Status string `json:"status" enums:"ok,unavailable"`
	Text   string `json:"text,omitempty"`
}

// Translate translates a short text.
//
//	@Summary		Translate a short text
//	@Description	Translates a short text (up to 500 bytes) with MyMemory, a free translation service; only the text is sent. Answers are cached for a year. Used to compare coin subjects written in the user's language with Numista's English titles.
//	@Tags			Numista
//	@Produce		json
//	@Param			q		query		string	true	"Text to translate"
//	@Param			from	query		string	true	"Source language"	Enums(pt, en, fr, es, de, it, nl)
//	@Param			to		query		string	false	"Target language"	Enums(pt, en, fr, es, de, it, nl)	default(en)
//	@Success		200		{object}	TranslationResult
//	@Failure		400		{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/translate [get]
func (h *TranslationHandler) Translate(c *gin.Context) {
	text := strings.TrimSpace(c.Query("q"))
	from, to := strings.ToLower(c.Query("from")), strings.ToLower(c.DefaultQuery("to", "en"))
	if text == "" || len(text) > services.TranslationMaxBytes {
		c.JSON(http.StatusBadRequest, gin.H{"error": "q must have 1 to 500 bytes"})
		return
	}
	if !services.TranslationLanguages[from] || !services.TranslationLanguages[to] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "from and to must be one of: pt, en, fr, es, de, it, nl"})
		return
	}
	translated, err := h.svc.Translate(c.Request.Context(), text, from, to)
	if err != nil {
		c.JSON(http.StatusOK, TranslationResult{Status: "unavailable"})
		return
	}
	c.JSON(http.StatusOK, TranslationResult{Status: "ok", Text: translated})
}
