package services

import (
	"strings"
	"testing"

	"github.com/briandenicola/ancient-coins-api/models"
	"github.com/briandenicola/ancient-coins-api/repository"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func setupNotificationServiceDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("failed to get sql db: %v", err)
	}
	// Pin the pool to one connection so no query lands on a second, unmigrated ":memory:" instance.
	sqlDB.SetMaxOpenConns(1)
	if err := db.AutoMigrate(&models.User{}, &models.Notification{}, &models.AppSetting{}); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}
	return db
}

func newTestNotificationService(db *gorm.DB) *NotificationService {
	logger := NewLogger(100)
	return NewNotificationService(
		repository.NewNotificationRepository(db),
		nil,
		repository.NewUserRepository(db),
		logger,
	)
}

// --- Feature 353 (T023): NotifyAvailabilityRunTerminal / NotifyAdminCycleChildFailure ---

func TestNotifyAvailabilityRunTerminal_CreatesOneNotificationWithGenericSuccessMessage(t *testing.T) {
	db := setupNotificationServiceDB(t)
	svc := newTestNotificationService(db)

	user := models.User{Username: "owner1", Email: "owner1@example.com", PasswordHash: "hash"}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	run := &models.AvailabilityRun{
		ID:           42,
		UserID:       user.ID,
		Status:       models.AvailabilityRunStatusCompleted,
		CoinsChecked: 2,
		Available:    2,
		Unavailable:  0,
		Unknown:      0,
	}

	svc.NotifyAvailabilityRunTerminal(user.ID, run, nil)

	var notifications []models.Notification
	if err := db.Where("user_id = ?", user.ID).Find(&notifications).Error; err != nil {
		t.Fatalf("failed to query notifications: %v", err)
	}
	if len(notifications) != 1 {
		t.Fatalf("expected exactly 1 notification, got %d", len(notifications))
	}
	n := notifications[0]
	if n.Type != NotificationTypeAvailabilityRun {
		t.Errorf("type = %q, want %q", n.Type, NotificationTypeAvailabilityRun)
	}
	if n.ReferenceID != run.ID {
		t.Errorf("referenceId = %d, want %d", n.ReferenceID, run.ID)
	}
	if strings.Contains(n.Message, "http://") || strings.Contains(n.Message, "https://") {
		t.Errorf("message must not contain a URL, got %q", n.Message)
	}
	if strings.Contains(n.Title, "http://") || strings.Contains(n.Title, "https://") {
		t.Errorf("title must not contain a URL, got %q", n.Title)
	}
}

func TestNotifyAvailabilityRunTerminal_FailureMessageIsGenericNoURLOrErrorText(t *testing.T) {
	db := setupNotificationServiceDB(t)
	svc := newTestNotificationService(db)

	user := models.User{Username: "owner2", Email: "owner2@example.com", PasswordHash: "hash"}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	run := &models.AvailabilityRun{
		ID:          7,
		UserID:      user.ID,
		Status:      models.AvailabilityRunStatusFailed,
		FailMessage: models.GenericAvailabilityFailureMessage,
	}

	svc.NotifyAvailabilityRunTerminal(user.ID, run, nil)

	var n models.Notification
	if err := db.Where("user_id = ? AND type = ?", user.ID, NotificationTypeAvailabilityRun).First(&n).Error; err != nil {
		t.Fatalf("failed to query notification: %v", err)
	}
	if n.Message != models.GenericAvailabilityFailureMessage {
		t.Errorf("message = %q, want the single generic failure message %q", n.Message, models.GenericAvailabilityFailureMessage)
	}
	for _, forbidden := range []string{"http://", "https://", "SELECT", "panic", "?"} {
		if strings.Contains(n.Message, forbidden) {
			t.Errorf("failure message leaked internal detail (%q found): %q", forbidden, n.Message)
		}
	}
}

func TestNotifyAvailabilityRunTerminal_ZeroURLsStillNotifiesOnce(t *testing.T) {
	db := setupNotificationServiceDB(t)
	svc := newTestNotificationService(db)

	user := models.User{Username: "owner3", Email: "owner3@example.com", PasswordHash: "hash"}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	run := &models.AvailabilityRun{
		ID:           9,
		UserID:       user.ID,
		Status:       models.AvailabilityRunStatusCompleted,
		CoinsChecked: 0,
	}

	svc.NotifyAvailabilityRunTerminal(user.ID, run, nil)

	var count int64
	db.Model(&models.Notification{}).Where("user_id = ? AND type = ?", user.ID, NotificationTypeAvailabilityRun).Count(&count)
	if count != 1 {
		t.Fatalf("expected exactly 1 notification for a zero-URL run, got %d", count)
	}
}

func TestNotifyAvailabilityRunTerminal_NoDedupeAcrossMultipleRuns(t *testing.T) {
	db := setupNotificationServiceDB(t)
	svc := newTestNotificationService(db)

	user := models.User{Username: "owner4", Email: "owner4@example.com", PasswordHash: "hash"}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	for i := 0; i < 3; i++ {
		run := &models.AvailabilityRun{
			ID:           uint(100 + i),
			UserID:       user.ID,
			Status:       models.AvailabilityRunStatusCompleted,
			CoinsChecked: 1,
			Available:    1,
		}
		svc.NotifyAvailabilityRunTerminal(user.ID, run, nil)
	}

	var count int64
	db.Model(&models.Notification{}).Where("user_id = ? AND type = ?", user.ID, NotificationTypeAvailabilityRun).Count(&count)
	if count != 3 {
		t.Fatalf("expected 3 separate notifications (no daily/content dedupe), got %d", count)
	}
}

func TestNotifyAdminCycleChildFailure_MessageContainsOwnerAndCycleButNoURL(t *testing.T) {
	db := setupNotificationServiceDB(t)
	svc := newTestNotificationService(db)

	admin := models.User{Username: "adminuser", Email: "admin@example.com", PasswordHash: "hash"}
	if err := db.Create(&admin).Error; err != nil {
		t.Fatalf("failed to create admin: %v", err)
	}

	svc.NotifyAdminCycleChildFailure(admin.ID, "affected-owner", 55)

	var n models.Notification
	if err := db.Where("user_id = ? AND type = ?", admin.ID, NotificationTypeAvailabilityRun).First(&n).Error; err != nil {
		t.Fatalf("failed to query admin failure notification: %v", err)
	}
	if !strings.Contains(n.Message, "affected-owner") {
		t.Errorf("message should mention the affected owner's username, got %q", n.Message)
	}
	if !strings.Contains(n.Message, "55") {
		t.Errorf("message should mention the cycle ID, got %q", n.Message)
	}
	if strings.Contains(n.Message, "http://") || strings.Contains(n.Message, "https://") {
		t.Errorf("admin failure message must not contain a URL, got %q", n.Message)
	}
}

func TestNotifyAdminCycleChildFailure_NoOpWhenAdminIDZero(t *testing.T) {
	db := setupNotificationServiceDB(t)
	svc := newTestNotificationService(db)

	svc.NotifyAdminCycleChildFailure(0, "someone", 1)

	var count int64
	db.Model(&models.Notification{}).Where("type = ?", NotificationTypeAvailabilityRun).Count(&count)
	if count != 0 {
		t.Fatalf("expected no notification created for adminID == 0, got %d", count)
	}
}
