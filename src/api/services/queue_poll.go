package services

import (
	"context"
	"time"

	"github.com/briandenicola/ancient-coins-api/repository"
)

// queuePollInterval is how often an idle background worker looks for queued
// work. New work wakes a worker at once; the poll only finds what no wake
// reached (work recovered at the start, or queued while the worker was
// busy), so an idle app does not query the database several times a second.
// Tests shorten it.
var queuePollInterval = 5 * time.Second

// queueNextPageDelay is the pause before a worker reads its queue again after
// a full page: more work may be waiting.
const queueNextPageDelay = time.Second

// fullQueuePage tells whether a worker read a full page of queued IDs, and
// waits queueNextPageDelay if so; false when it did not, or must stop.
func fullQueuePage(ctx context.Context, ids []uint) bool {
	if len(ids) < repository.QueuedIDsPageSize {
		return false
	}
	select {
	case <-ctx.Done():
		return false
	case <-time.After(queueNextPageDelay):
		return true
	}
}
