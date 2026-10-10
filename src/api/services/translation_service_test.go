package services

import (
	"context"
	"errors"
	"testing"

	"github.com/briandenicola/ancient-coins-api/models"
	"github.com/briandenicola/ancient-coins-api/repository"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

type fakeTranslationClient struct {
	answers map[string]string
	calls   int
	err     error
}

func (f *fakeTranslationClient) Translate(_ context.Context, text, from, to string) (string, error) {
	f.calls++
	if f.err != nil {
		return "", f.err
	}
	return f.answers[from+"|"+to+"|"+text], nil
}

func newTestTranslationService(t *testing.T, client TranslationClient) *TranslationService {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&models.NumistaCacheEntry{}); err != nil {
		t.Fatal(err)
	}
	return NewTranslationService(client, NewNumistaResponseCache(repository.NewNumistaCacheRepository(db), nil, nil))
}

func TestTranslationService_TranslatesOnceAndKeepsTheAnswer(t *testing.T) {
	client := &fakeTranslationClient{answers: map[string]string{"pt|en|Juntos com Ucrânia": "Together with Ukraine"}}
	svc := newTestTranslationService(t, client)
	for range 2 {
		got, err := svc.Translate(context.Background(), "  Juntos com   Ucrânia ", "pt", "en")
		if err != nil || got != "Together with Ukraine" {
			t.Fatalf("got %q, %v", got, err)
		}
	}
	if client.calls != 1 {
		t.Fatalf("calls = %d, want 1", client.calls)
	}
	if got, _ := svc.Translate(context.Background(), "Leão", "pt", "pt"); got != "Leão" || client.calls != 1 {
		t.Fatalf("same language: %q (%d calls)", got, client.calls)
	}
}

func TestTranslationService_ReportsAnUnavailableService(t *testing.T) {
	svc := newTestTranslationService(t, &fakeTranslationClient{err: errors.New("offline")})
	if _, err := svc.Translate(context.Background(), "Paz Entre as Nações", "pt", "en"); err == nil {
		t.Fatal("expected an error")
	}
}

func TestParseMyMemoryResponse(t *testing.T) {
	got, err := ParseMyMemoryResponse([]byte(`{"responseData":{"translatedText":"Introduction of the euro"},"responseStatus":200,"quotaFinished":false}`))
	if err != nil || got != "Introduction of the euro" {
		t.Fatalf("got %q, %v", got, err)
	}
	for _, body := range []string{
		`{"responseData":{"translatedText":"MYMEMORY WARNING: YOU USED ALL AVAILABLE FREE TRANSLATIONS FOR TODAY"},"responseStatus":429,"quotaFinished":true}`,
		`{"responseData":{"translatedText":"INVALID LANGUAGE PAIR"},"responseStatus":"403"}`,
		`not json`,
	} {
		if text, err := ParseMyMemoryResponse([]byte(body)); err == nil {
			t.Errorf("%s: expected an error, got %q", body, text)
		}
	}
}
