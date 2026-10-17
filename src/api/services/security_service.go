package services

import (
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/briandenicola/ancient-coins-api/models"
	"github.com/briandenicola/ancient-coins-api/repository"
)

const (
	AccountFailureLimit = 5
	IPFailureLimit      = 20
	FailureWindow       = 15 * time.Minute
	LockoutDuration     = 15 * time.Minute
)

var ErrAccountLocked = errors.New("account locked")

type SecuritySummary struct {
	Since           time.Time `json:"since"`
	LoginFailures   int64     `json:"loginFailures"`
	LoginSuccesses  int64     `json:"loginSuccesses"`
	RefreshFailures int64     `json:"refreshFailures"`
}

type SecurityService struct {
	repo *repository.SecurityRepository
	now  func() time.Time
}

func NewSecurityService(repo *repository.SecurityRepository) *SecurityService {
	return &SecurityService{repo: repo, now: time.Now}
}

func (s *SecurityService) RecordEvent(eventType models.SecurityEventType, userID *uint, username, clientIP, userAgent, message string) {
	_ = s.repo.CreateEvent(&models.SecurityEvent{
		Type:      eventType,
		UserID:    userID,
		Username:  username,
		ClientIP:  clientIP,
		UserAgent: userAgent,
		Message:   message,
		CreatedAt: s.now(),
	})
}

func (s *SecurityService) CheckAccountAllowed(username string) error {
	user, err := s.repo.FindUserByUsername(username)
	if err != nil {
		return nil
	}
	if user.LockedUntil != nil && user.LockedUntil.After(s.now()) {
		if s.isOnlyAdmin(user) {
			_ = s.repo.UnlockUser(user.ID)
			return nil
		}
		return ErrAccountLocked
	}
	return nil
}

func (s *SecurityService) RecordPasswordSuccess(user models.User, clientIP, userAgent string) {
	userID := user.ID
	s.RecordEvent(models.SecurityEventPasswordLoginSuccess, &userID, user.Username, clientIP, userAgent, "")
	if user.LockedUntil != nil {
		_ = s.repo.UnlockUser(user.ID)
	}
}

func (s *SecurityService) RecordFinalLocalAdminBlocked(userID *uint, username string, adminID uint, operation string) {
	message := fmt.Sprintf("%s blocked for user %s by admin_id=%d", sanitizeSecurityEventDetail(operation), username, adminID)
	s.RecordEvent(models.SecurityEventFinalAdminBlocked, userID, username, "", "", message)
}

func sanitizeSecurityEventDetail(value string) string {
	clean := strings.TrimSpace(value)
	lower := strings.ToLower(clean)
	sensitiveMarkers := []string{
		"client_secret", "client secret", "authorization code", "code=", "id_token", "access_token",
		"refresh_token", "refresh token", "pkce", "verifier", "raw claims", "claims payload",
	}
	for _, marker := range sensitiveMarkers {
		if strings.Contains(lower, marker) {
			return "sensitive detail redacted"
		}
	}
	if len(clean) > 300 {
		return clean[:300]
	}
	return clean
}

func (s *SecurityService) RecordPasswordFailure(username, clientIP, userAgent string) {
	var userID *uint
	user, err := s.repo.FindUserByUsername(username)
	if err == nil {
		id := user.ID
		userID = &id
	}
	s.RecordEvent(models.SecurityEventPasswordLoginFailure, userID, username, clientIP, userAgent, "")

	if userID != nil {
		count, _ := s.repo.CountFailuresForUser(*userID, s.now().Add(-FailureWindow))
		if count >= AccountFailureLimit {
			if s.canLockUser(user) {
				until := s.now().Add(LockoutDuration)
				_ = s.repo.LockUser(*userID, until)
				s.RecordEvent(models.SecurityEventAccountLockout, userID, username, clientIP, userAgent, "too many failed password attempts")
			}
		}
	}
}

// IsIPDenied tells whether an address made IPFailureLimit failed password
// attempts in the last FailureWindow: it is refused until they age out.
func (s *SecurityService) IsIPDenied(clientIP string) bool {
	if net.ParseIP(clientIP) == nil {
		return false
	}
	count, err := s.repo.CountFailuresForIP(clientIP, s.now().Add(-FailureWindow))
	return err == nil && count >= IPFailureLimit
}

func (s *SecurityService) canLockUser(user *models.User) bool {
	return user != nil && !s.isOnlyAdmin(user)
}

func (s *SecurityService) isOnlyAdmin(user *models.User) bool {
	if user == nil || user.Role != models.RoleAdmin {
		return false
	}
	admins, err := s.repo.CountAdmins()
	return err == nil && admins <= 1
}

func (s *SecurityService) ListEvents(filters repository.SecurityEventFilters) ([]models.SecurityEvent, error) {
	return s.repo.ListEvents(filters)
}

func (s *SecurityService) Summary() (SecuritySummary, error) {
	since := s.now().Add(-24 * time.Hour)
	failures, err := s.repo.CountEvents(models.SecurityEventPasswordLoginFailure, since)
	if err != nil {
		return SecuritySummary{}, err
	}
	successes, err := s.repo.CountEvents(models.SecurityEventPasswordLoginSuccess, since)
	if err != nil {
		return SecuritySummary{}, err
	}
	refreshFailures, err := s.repo.CountEvents(models.SecurityEventRefreshFailure, since)
	if err != nil {
		return SecuritySummary{}, err
	}
	return SecuritySummary{Since: since, LoginFailures: failures, LoginSuccesses: successes, RefreshFailures: refreshFailures}, nil
}

func (s *SecurityService) UnlockUser(userID uint, adminID uint) error {
	if err := s.repo.UnlockUser(userID); err != nil {
		return err
	}
	s.RecordEvent(models.SecurityEventAccountUnlock, &adminID, "", "", "", "")
	return nil
}
