package services

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Short texts (a coin's subject, written in the user's language) are
// translated with MyMemory, a free service that needs no key, so they can be
// compared with Numista's English titles. Only the text is sent.
const (
	myMemoryURL          = "https://api.mymemory.translated.net/get"
	translationTimeout   = 15 * time.Second
	translationCacheTTL  = 365 * 24 * time.Hour
	translationBodyLimit = 1 << 20
	// MyMemory takes up to 500 bytes per request.
	TranslationMaxBytes = 500
)

// TranslationLanguages are the languages a text can be translated from or to.
var TranslationLanguages = map[string]bool{"pt": true, "en": true, "fr": true, "es": true, "de": true, "it": true, "nl": true}

// TranslationClient asks a translation service for one text.
type TranslationClient interface {
	Translate(ctx context.Context, text, from, to string) (string, error)
}

// MyMemoryClient translates with the MyMemory API.
type MyMemoryClient struct {
	client  *http.Client
	baseURL string
}

// NewMyMemoryClient creates the client; a nil client gets a default one.
func NewMyMemoryClient(client *http.Client) *MyMemoryClient {
	if client == nil {
		client = &http.Client{Timeout: translationTimeout}
	}
	return &MyMemoryClient{client: client, baseURL: myMemoryURL}
}

// Translate returns the text in the target language.
func (c *MyMemoryClient) Translate(ctx context.Context, text, from, to string) (string, error) {
	query := url.Values{"q": {text}, "langpair": {from + "|" + to}}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"?"+query.Encode(), nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", CommonsUserAgent)
	resp, err := c.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, translationBodyLimit))
	if err != nil {
		return "", err
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("mymemory: status %d", resp.StatusCode)
	}
	return ParseMyMemoryResponse(body)
}

// ParseMyMemoryResponse reads MyMemory's answer; its errors (such as the
// daily quota) come back as JSON with another responseStatus.
func ParseMyMemoryResponse(body []byte) (string, error) {
	var response struct {
		ResponseData struct {
			TranslatedText string `json:"translatedText"`
		} `json:"responseData"`
		ResponseStatus json.RawMessage `json:"responseStatus"`
		QuotaFinished  bool            `json:"quotaFinished"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		return "", err
	}
	status := strings.Trim(string(response.ResponseStatus), `"`)
	text := strings.TrimSpace(response.ResponseData.TranslatedText)
	if status != "200" || response.QuotaFinished || text == "" {
		return "", fmt.Errorf("mymemory: status %s", status)
	}
	return text, nil
}

// TranslationService translates short texts, keeping each answer.
type TranslationService struct {
	client TranslationClient
	cache  *NumistaResponseCache
}

// NewTranslationService creates the service; translations are kept in the
// shared response cache for a year.
func NewTranslationService(client TranslationClient, cache *NumistaResponseCache) *TranslationService {
	return &TranslationService{client: client, cache: cache}
}

// Translate returns the text translated, or an error when the service could
// not be reached or refused it (its daily quota is used up, for instance).
func (s *TranslationService) Translate(ctx context.Context, text, from, to string) (string, error) {
	text = strings.Join(strings.Fields(text), " ")
	if from == to {
		return text, nil
	}
	digest := sha256.Sum256([]byte(text))
	key := "translate:v1:" + from + ":" + to + ":" + hex.EncodeToString(digest[:])
	body, _, _, err := s.cache.Get(ctx, key, translationCacheTTL, func(ctx context.Context) ([]byte, error) {
		translated, err := s.client.Translate(ctx, text, from, to)
		if err != nil {
			return nil, err
		}
		return json.Marshal(translated)
	})
	if err != nil {
		return "", err
	}
	var translated string
	if err := json.Unmarshal(body, &translated); err != nil {
		return "", err
	}
	return translated, nil
}
