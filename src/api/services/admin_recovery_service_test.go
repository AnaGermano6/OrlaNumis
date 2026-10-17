package services

import (
	"strings"
	"testing"

	"github.com/briandenicola/ancient-coins-api/models"
	"github.com/briandenicola/ancient-coins-api/repository"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func setupAdminRecoveryServiceTest(t *testing.T) (*gorm.DB, *AdminRecoveryService) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	if err := db.AutoMigrate(&models.User{}, &models.SecurityEvent{}); err != nil {
		t.Fatalf("failed to migrate test db: %v", err)
	}
	securitySvc := NewSecurityService(repository.NewSecurityRepository(db))
	return db, NewAdminRecoveryService(repository.NewAdminRepository(db), securitySvc)
}

func createAdminRecoveryTestUser(t *testing.T, db *gorm.DB, username string, role models.UserRole, passwordHash string) models.User {
	t.Helper()
	user := models.User{
		Username:     username,
		Email:        username + "@example.com",
		PasswordHash: passwordHash,
		Role:         role,
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}
	return user
}

func TestAdminRecoveryServiceFinalLocalAdminBlockedEventRedactsSensitiveOperationDetails(t *testing.T) {
	db, svc := setupAdminRecoveryServiceTest(t)
	admin := createAdminRecoveryTestUser(t, db, "only-admin-sensitive", models.RoleAdmin, "local-password-hash")
	actorID := admin.ID

	svc.recordBlocked(&admin, &actorID, "convert to OIDC-only with client_secret=super-secret pkce verifier")

	var event models.SecurityEvent
	if err := db.Where("type = ?", models.SecurityEventFinalAdminBlocked).First(&event).Error; err != nil {
		t.Fatalf("failed to load security event: %v", err)
	}
	if strings.Contains(strings.ToLower(event.Message), "super-secret") ||
		strings.Contains(strings.ToLower(event.Message), "client_secret") ||
		strings.Contains(strings.ToLower(event.Message), "pkce") ||
		strings.Contains(strings.ToLower(event.Message), "verifier") {
		t.Fatalf("expected sensitive operation detail redacted from final-local-admin event, got %q", event.Message)
	}
	if !strings.Contains(event.Message, "sensitive detail redacted") {
		t.Fatalf("expected redaction marker in final-local-admin event, got %q", event.Message)
	}
}
