package services

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"time"

	"github.com/briandenicola/ancient-coins-api/models"
)

// NumistaCacheStore persists raw Numista responses.
type NumistaCacheStore interface {
	Find(key string) (*models.NumistaCacheEntry, error)
	Save(entry *models.NumistaCacheEntry) error
}

// NumistaResponseCache keeps raw Numista responses in the database so repeat
// lookups and restarts do not spend the free plan's 2,000 monthly requests.
// When a refresh fails, the last stored copy is served and marked stale.
type NumistaResponseCache struct {
	store  NumistaCacheStore
	clock  NumistaClock
	logger *Logger
	// fetchMu serialises provider calls so concurrent requests for the same
	// data trigger one call; the cache is re-checked after acquiring it.
	fetchMu sync.Mutex
}

// NewNumistaResponseCache creates the cache. clock and logger may be nil.
func NewNumistaResponseCache(store NumistaCacheStore, clock NumistaClock, logger *Logger) *NumistaResponseCache {
	if clock == nil {
		clock = NewSystemNumistaClock()
	}
	return &NumistaResponseCache{store: store, clock: clock, logger: logger}
}

// Get returns the stored payload for key while it is younger than ttl,
// otherwise loads and stores a fresh one. If loading fails and an older copy
// exists, that copy is returned with stale=true instead of the error.
func (c *NumistaResponseCache) Get(ctx context.Context, key string, ttl time.Duration, load func(context.Context) ([]byte, error)) ([]byte, time.Time, bool, error) {
	entry := c.find(key)
	if entry != nil && c.clock.Now().Sub(entry.FetchedAt) < ttl {
		return []byte(entry.Payload), entry.FetchedAt, false, nil
	}

	c.fetchMu.Lock()
	defer c.fetchMu.Unlock()
	if fresh := c.find(key); fresh != nil && c.clock.Now().Sub(fresh.FetchedAt) < ttl {
		return []byte(fresh.Payload), fresh.FetchedAt, false, nil
	}

	body, err := load(ctx)
	if err == nil && !json.Valid(body) {
		err = &NumistaError{Kind: NumistaErrorMalformedResponse}
	}
	if err != nil {
		if entry != nil {
			return []byte(entry.Payload), entry.FetchedAt, true, nil
		}
		return nil, time.Time{}, false, err
	}
	now := c.clock.Now()
	if saveErr := c.store.Save(&models.NumistaCacheEntry{Key: key, Payload: string(body), FetchedAt: now}); saveErr != nil && c.logger != nil {
		c.logger.Warn("numista", "Failed to cache %s: %v", key, saveErr)
	}
	return body, now, false, nil
}

func (c *NumistaResponseCache) find(key string) *models.NumistaCacheEntry {
	entry, err := c.store.Find(key)
	if err != nil {
		if c.logger != nil {
			c.logger.Warn("numista", "Failed to read cached %s: %v", key, err)
		}
		return nil
	}
	return entry
}

// numistaStatusForError maps a provider error to a lookup status.
func numistaStatusForError(err error) string {
	var numistaErr *NumistaError
	if !errors.As(err, &numistaErr) {
		return models.NumistaDataStatusUnavailable
	}
	switch numistaErr.Kind {
	case NumistaErrorUnconfigured:
		return models.NumistaDataStatusUnconfigured
	case NumistaErrorUnauthorized:
		return models.NumistaDataStatusUnauthorized
	case NumistaErrorQuotaLimited:
		return models.NumistaDataStatusQuotaLimited
	default:
		return models.NumistaDataStatusUnavailable
	}
}
