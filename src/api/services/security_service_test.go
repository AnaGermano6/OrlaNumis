package services

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/briandenicola/ancient-coins-api/models"
	"github.com/briandenicola/ancient-coins-api/repository"
	"github.com/glebarez/sqlite"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func setupSecurityServiceTest(t *testing.T) (*gorm.DB, *SecurityService) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	if err := db.AutoMigrate(&models.User{}, &models.SecurityEvent{}, &models.RefreshToken{}); err != nil {
		t.Fatalf("failed to migrate security test db: %v", err)
	}
	return db, NewSecurityService(repository.NewSecurityRepository(db))
}

func createSecurityTestUser(t *testing.T, db *gorm.DB, username string, role models.UserRole) models.User {
	t.Helper()
	hash, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}
	user := models.User{
		Username:     username,
		Email:        username + "@example.com",
		PasswordHash: string(hash),
		Role:         role,
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}
	return user
}

func TestSecurityServiceDistributedFailuresLockAccountAndExpire(t *testing.T) {
	db, securitySvc := setupSecurityServiceTest(t)
	user := createSecurityTestUser(t, db, "victim", models.RoleUser)
	now := time.Now().UTC()
	securitySvc.now = func() time.Time { return now }

	for i, ip := range []string{"203.0.113.1", "203.0.113.2", "203.0.113.3", "203.0.113.4", "203.0.113.5"} {
		securitySvc.RecordPasswordFailure(user.Username, ip, "stuffing-agent")
		var current models.User
		if err := db.First(&current, user.ID).Error; err != nil {
			t.Fatalf("attempt %d failed to reload user: %v", i+1, err)
		}
		if i < AccountFailureLimit-1 && current.LockedUntil != nil {
			t.Fatalf("attempt %d locked account before threshold: %v", i+1, current.LockedUntil)
		}
	}

	var locked models.User
	if err := db.First(&locked, user.ID).Error; err != nil {
		t.Fatalf("failed to reload locked user: %v", err)
	}
	if locked.LockedUntil == nil {
		t.Fatal("expected distributed failures against one username to set LockedUntil")
	}
	if got, want := locked.LockedUntil.Sub(now), LockoutDuration; got != want {
		t.Fatalf("expected lockout duration %s, got %s", want, got)
	}
	if err := securitySvc.CheckAccountAllowed(user.Username); !errors.Is(err, ErrAccountLocked) {
		t.Fatalf("expected locked account to be denied, got %v", err)
	}

	securitySvc.now = func() time.Time { return now.Add(LockoutDuration + time.Second) }
	if err := securitySvc.CheckAccountAllowed(user.Username); err != nil {
		t.Fatalf("expected expired lockout to be ignored, got %v", err)
	}

	var lockoutEvents int64
	if err := db.Model(&models.SecurityEvent{}).Where("type = ?", models.SecurityEventAccountLockout).Count(&lockoutEvents).Error; err != nil {
		t.Fatalf("failed to count lockout events: %v", err)
	}
	if lockoutEvents != 1 {
		t.Fatalf("expected one lockout event, got %d", lockoutEvents)
	}
}

func TestSecurityServiceSuccessfulLoginResetsFailureEscalation(t *testing.T) {
	db, securitySvc := setupSecurityServiceTest(t)
	user := createSecurityTestUser(t, db, "reset-user", models.RoleUser)
	authSvc := NewAuthService(repository.NewAuthRepository(db), "security-test-secret").WithSecurity(securitySvc)

	for i := 0; i < AccountFailureLimit-1; i++ {
		if _, err := authSvc.AuthenticateUserWithRequest(user.Username, "wrong-password", "203.0.113.20", "browser"); !errors.Is(err, ErrInvalidCredentials) {
			t.Fatalf("expected invalid credentials on pre-success failure %d, got %v", i+1, err)
		}
	}
	if _, err := authSvc.AuthenticateUserWithRequest(user.Username, "password123", "203.0.113.20", "browser"); err != nil {
		t.Fatalf("expected successful login before reset assertion, got %v", err)
	}
	if _, err := authSvc.AuthenticateUserWithRequest(user.Username, "wrong-password", "203.0.113.20", "browser"); !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("expected invalid credentials after success reset, got %v", err)
	}

	var current models.User
	if err := db.First(&current, user.ID).Error; err != nil {
		t.Fatalf("failed to reload user: %v", err)
	}
	if current.LockedUntil != nil && current.LockedUntil.After(time.Now()) {
		t.Fatalf("expected successful login to reset prior failure escalation, got lockout until %v", current.LockedUntil)
	}
	var successEvents int64
	if err := db.Model(&models.SecurityEvent{}).Where("type = ? AND user_id = ?", models.SecurityEventPasswordLoginSuccess, user.ID).Count(&successEvents).Error; err != nil {
		t.Fatalf("failed to count success events: %v", err)
	}
	if successEvents != 1 {
		t.Fatalf("expected successful login to record one success event, got %d", successEvents)
	}
}

func TestSecurityServiceDoesNotLockOnlyAdminOutOfAdministration(t *testing.T) {
	db, securitySvc := setupSecurityServiceTest(t)
	admin := createSecurityTestUser(t, db, "only-admin", models.RoleAdmin)

	for i := 0; i < AccountFailureLimit; i++ {
		securitySvc.RecordPasswordFailure(admin.Username, "203.0.113.30", "browser")
	}

	var current models.User
	if err := db.First(&current, admin.ID).Error; err != nil {
		t.Fatalf("failed to reload admin: %v", err)
	}
	if current.LockedUntil != nil && current.LockedUntil.After(time.Now()) {
		t.Fatalf("expected only admin to remain administratively recoverable, got lockout until %v", current.LockedUntil)
	}
}

func TestSecurityServiceOnlyAdminFailuresStillEscalateIPBan(t *testing.T) {
	db, securitySvc := setupSecurityServiceTest(t)
	admin := createSecurityTestUser(t, db, "only-admin-ip-ban", models.RoleAdmin)
	clientIP := "203.0.113.32"

	for i := 0; i < IPFailureLimit; i++ {
		securitySvc.RecordPasswordFailure(admin.Username, clientIP, "browser")
	}

	var current models.User
	if err := db.First(&current, admin.ID).Error; err != nil {
		t.Fatalf("failed to reload admin: %v", err)
	}
	if current.LockedUntil != nil && current.LockedUntil.After(time.Now()) {
		t.Fatalf("expected only admin account lock to be skipped, got lockout until %v", current.LockedUntil)
	}
	if !securitySvc.IsIPDenied(clientIP) {
		t.Fatal("expected repeated only-admin failures to create an IP deny rule")
	}
}

func TestSecurityServiceClearsExistingLockoutForOnlyAdmin(t *testing.T) {
	db, securitySvc := setupSecurityServiceTest(t)
	admin := createSecurityTestUser(t, db, "recoverable-admin", models.RoleAdmin)
	until := time.Now().Add(time.Hour)
	if err := db.Model(&models.User{}).Where("id = ?", admin.ID).Update("locked_until", until).Error; err != nil {
		t.Fatalf("failed to seed admin lockout: %v", err)
	}

	if err := securitySvc.CheckAccountAllowed(admin.Username); err != nil {
		t.Fatalf("expected only admin lockout to be recoverable, got %v", err)
	}

	var current models.User
	if err := db.First(&current, admin.ID).Error; err != nil {
		t.Fatalf("failed to reload admin: %v", err)
	}
	if current.LockedUntil != nil {
		t.Fatalf("expected only admin lockout to be cleared, got %v", current.LockedUntil)
	}
}

func TestSecurityServiceLocksAdminWhenAnotherAdminCanRecover(t *testing.T) {
	db, securitySvc := setupSecurityServiceTest(t)
	admin := createSecurityTestUser(t, db, "lockable-admin", models.RoleAdmin)
	createSecurityTestUser(t, db, "recovery-admin", models.RoleAdmin)

	for i := 0; i < AccountFailureLimit; i++ {
		securitySvc.RecordPasswordFailure(admin.Username, "203.0.113.31", "browser")
	}

	var current models.User
	if err := db.First(&current, admin.ID).Error; err != nil {
		t.Fatalf("failed to reload admin: %v", err)
	}
	if current.LockedUntil == nil || !current.LockedUntil.After(time.Now()) {
		t.Fatalf("expected admin to be lockable when another admin can recover, got %v", current.LockedUntil)
	}
	if err := securitySvc.CheckAccountAllowed(admin.Username); !errors.Is(err, ErrAccountLocked) {
		t.Fatalf("expected locked admin to be denied in multiple-admin deployment, got %v", err)
	}
}

func TestSecurityServiceAdminUnlockClearsLockoutAndAudits(t *testing.T) {
	db, securitySvc := setupSecurityServiceTest(t)
	user := createSecurityTestUser(t, db, "locked-user", models.RoleUser)
	admin := createSecurityTestUser(t, db, "admin", models.RoleAdmin)
	until := time.Now().Add(time.Hour)
	if err := db.Model(&models.User{}).Where("id = ?", user.ID).Update("locked_until", until).Error; err != nil {
		t.Fatalf("failed to lock user: %v", err)
	}

	if err := securitySvc.UnlockUser(user.ID, admin.ID); err != nil {
		t.Fatalf("unlock failed: %v", err)
	}

	var unlocked models.User
	if err := db.First(&unlocked, user.ID).Error; err != nil {
		t.Fatalf("failed to reload user: %v", err)
	}
	if unlocked.LockedUntil != nil {
		t.Fatalf("expected unlock to clear LockedUntil, got %v", unlocked.LockedUntil)
	}

	var event models.SecurityEvent
	if err := db.Where("type = ?", models.SecurityEventAccountUnlock).First(&event).Error; err != nil {
		t.Fatalf("expected account unlock audit event: %v", err)
	}
	if event.UserID == nil || *event.UserID != admin.ID {
		t.Fatalf("expected unlock event to record admin id %d, got %v", admin.ID, event.UserID)
	}
}

func assertSecurityEventMessageRedacted(t *testing.T, message string, forbidden ...string) {
	t.Helper()
	if !strings.Contains(message, "sensitive detail redacted") {
		t.Fatalf("expected sensitive details to be redacted, got %q", message)
	}
	for _, value := range forbidden {
		if strings.Contains(strings.ToLower(message), strings.ToLower(value)) {
			t.Fatalf("expected audit message to omit %q, got %q", value, message)
		}
	}
}
