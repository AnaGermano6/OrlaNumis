package services

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/briandenicola/ancient-coins-api/models"
	"github.com/briandenicola/ancient-coins-api/repository"
)

// NotificationTypeSetNewCoins tells a user that coins came out for one of
// their catalogue checklist sets. Its title is the set's name and its
// message the coins, one per line.
const NotificationTypeSetNewCoins = "set_new_coins"

const (
	// The catalogue is read again every six hours.
	setChecklistWatchInterval = 6 * time.Hour
	// A coin is announced when it is of this year or last year: an older
	// one is an addition to the catalogue (a source answering again), not a
	// release.
	setChecklistNewYears = 1
	// At most this many coins are named in a notification.
	setChecklistNotifiedCoins = 5
)

// SetChecklistWatcher announces the coins that come out for catalogue
// checklist sets ("2 € comemorativas de Portugal": a new 2 € coin of
// Portugal).
type SetChecklistWatcher struct {
	sets          *repository.SetRepository
	checklist     *SetChecklistService
	notifications *repository.NotificationRepository
	logger        *Logger
	now           func() time.Time
	stopCh        chan struct{}
	once          sync.Once
	lastRun       time.Time
}

// NewSetChecklistWatcher creates the watcher; notifications may be nil.
func NewSetChecklistWatcher(sets *repository.SetRepository, checklist *SetChecklistService, notifications *repository.NotificationRepository, logger *Logger) *SetChecklistWatcher {
	return &SetChecklistWatcher{sets: sets, checklist: checklist, notifications: notifications, logger: logger, now: time.Now, stopCh: make(chan struct{})}
}

// Start checks a few minutes after startup, then every six hours.
func (w *SetChecklistWatcher) Start() {
	scheduleLoop{
		category:     "set-checklist-watcher",
		name:         "new coins for checklist sets",
		logger:       w.logger,
		stopCh:       w.stopCh,
		initialDelay: 2 * time.Minute,
		next: func(now time.Time, catchUp bool) time.Time {
			if catchUp || w.lastRun.IsZero() {
				return now
			}
			return w.lastRun.Add(setChecklistWatchInterval)
		},
		run: func() {
			w.lastRun = time.Now()
			if err := w.Check(context.Background()); err != nil {
				w.logger.Error("set-checklist-watcher", "New coins check failed: %v", err)
			}
		},
	}.Run()
}

// Stop ends the checks.
func (w *SetChecklistWatcher) Stop() {
	w.once.Do(func() { close(w.stopCh) })
}

// Check compares each catalogue checklist set with the catalogue: the coins
// it lists now and did not before are remembered, and those just out
// (this year or last) are announced to the set's owner.
func (w *SetChecklistWatcher) Check(ctx context.Context) error {
	sets, err := w.sets.ListChecklistSets()
	if err != nil {
		return err
	}
	for _, set := range sets {
		if set.Checklist == nil {
			continue
		}
		definition, err := ParseChecklistDefinition(map[string]interface{}(*set.Checklist))
		if err != nil || definition.Source != ChecklistSourceCatalogue {
			continue
		}
		known, err := w.sets.ChecklistKeys(set.ID)
		if err != nil {
			return err
		}
		fresh := []ChecklistSlot{}
		keys := []string{}
		for _, slot := range w.checklist.CatalogueSlots(ctx, definition) {
			if !known[slot.Key] {
				fresh = append(fresh, slot)
				keys = append(keys, slot.Key)
			}
		}
		if len(fresh) == 0 {
			continue
		}
		if err := w.sets.AddChecklistKeys(set.ID, keys); err != nil {
			return err
		}
		// A set looked at for the first time has nothing new to it.
		if len(known) == 0 {
			continue
		}
		recent := []ChecklistSlot{}
		for _, slot := range fresh {
			if slot.Year >= w.now().Year()-setChecklistNewYears {
				recent = append(recent, slot)
			}
		}
		if len(recent) > 0 {
			w.notify(set, recent)
		}
	}
	return nil
}

func (w *SetChecklistWatcher) notify(set models.CoinSet, coins []ChecklistSlot) {
	if w.notifications == nil {
		return
	}
	lines := []string{}
	for i, coin := range coins {
		if i == setChecklistNotifiedCoins {
			lines = append(lines, fmt.Sprintf("+%d", len(coins)-setChecklistNotifiedCoins))
			break
		}
		lines = append(lines, fmt.Sprintf("%s (%d)", coin.Title, coin.Year))
	}
	notification := &models.Notification{
		UserID:       set.UserID,
		Type:         NotificationTypeSetNewCoins,
		Title:        set.Name,
		Message:      strings.Join(lines, "\n"),
		ReferenceID:  set.ID,
		ReferenceURL: fmt.Sprintf("/sets/%d", set.ID),
	}
	if err := w.notifications.Create(notification); err != nil {
		w.logger.Error("set-checklist-watcher", "Failed to tell user %d of new coins for set %d: %v", set.UserID, set.ID, err)
		return
	}
	w.logger.Info("set-checklist-watcher", "Told user %d of %d new coin(s) for set %d", set.UserID, len(coins), set.ID)
}
