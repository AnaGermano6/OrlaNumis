package services

import (
	"fmt"
	"sort"
	"strings"

	"github.com/briandenicola/ancient-coins-api/models"
	"github.com/briandenicola/ancient-coins-api/repository"
)

// NotificationService handles creating and managing notifications.
type NotificationService struct {
	notifRepo  *repository.NotificationRepository
	socialRepo *repository.SocialRepository
	userRepo   *repository.UserRepository
	logger     *Logger
}

const NotificationTypeFollowRequest = "follow_request"

const (
	// NotificationTypeAvailabilityRun is the terminal-outcome notification created for every
	// terminal child AvailabilityRun (owner/scheduled/admin-triggered), in addition to (never
	// instead of) any per-coin wishlist_unavailable notifications fired during the same run (D6).
	NotificationTypeAvailabilityRun = "wishlist_availability_run"
	// NotificationTypePurchaseReminder is fired by the daily reminder scheduler when a
	// pending reminder's remind_date has arrived. ReferenceID = reminder.ID, ReferenceURL = /coin/{coinID}.
	NotificationTypePurchaseReminder = "purchase_reminder"
)

// NewNotificationService creates a new NotificationService.
func NewNotificationService(
	notifRepo *repository.NotificationRepository,
	socialRepo *repository.SocialRepository,
	userRepo *repository.UserRepository,
	logger *Logger,
) *NotificationService {
	return &NotificationService{
		notifRepo:  notifRepo,
		socialRepo: socialRepo,
		userRepo:   userRepo,
		logger:     logger,
	}
}

// NotifyWishlistUnavailable creates a notification when a wishlist coin
// is detected as no longer available.
func (s *NotificationService) NotifyWishlistUnavailable(userID uint, coin models.Coin, reason string) {
	coinName := coin.Name
	if coinName == "" {
		coinName = "Unnamed coin"
	}

	title := "Wishlist item unavailable"
	message := fmt.Sprintf("%s appears to no longer be available. %s", coinName, reason)

	n := &models.Notification{
		UserID:       userID,
		Type:         "wishlist_unavailable",
		Title:        title,
		Message:      message,
		ReferenceID:  coin.ID,
		ReferenceURL: coin.ReferenceURL,
	}

	if err := s.notifRepo.Create(n); err != nil {
		s.logger.Error("notifications", "Failed to create wishlist notification for user %d, coin %d: %v", userID, coin.ID, err)
	}

}

// NotifyAvailabilityRunTerminal creates the owner-facing terminal-outcome notification for a
// single completed or failed child AvailabilityRun. Exactly one is created per terminal child
// run, regardless of how many (if any) per-coin wishlist_unavailable notifications also fired
// for that run — this is purely additive and never gates or replaces the per-coin call (D6).
// The message is always generic: no URLs, no query text, no internal error details (FR-015);
// it may mention up to 3 newly-unavailable coin names (plus "and N more") for readability.
func (s *NotificationService) NotifyAvailabilityRunTerminal(userID uint, run *models.AvailabilityRun, newlyUnavailableCoinNames []string) {
	title := "Wishlist availability check complete"
	message := fmt.Sprintf(
		"Checked %d item(s): %d available, %d unavailable, %d unknown.",
		run.CoinsChecked, run.Available, run.Unavailable, run.Unknown,
	)

	if run.Status == models.AvailabilityRunStatusFailed {
		title = "Wishlist availability check failed"
		message = models.GenericAvailabilityFailureMessage
	} else if len(newlyUnavailableCoinNames) > 0 {
		message = fmt.Sprintf("%s %s", message, summarizeUnavailableCoinNames(newlyUnavailableCoinNames))
	}

	refURL := fmt.Sprintf("/wishlist/availability-runs/%d", run.ID)
	n := &models.Notification{
		UserID:       userID,
		Type:         NotificationTypeAvailabilityRun,
		Title:        title,
		Message:      message,
		ReferenceID:  run.ID,
		ReferenceURL: refURL,
	}
	if err := s.notifRepo.Create(n); err != nil {
		s.logger.Error("notifications", "Failed to create availability run notification for user %d, run %d: %v", userID, run.ID, err)
	}

}

// NotifyAdminCycleChildFailure notifies the admin who triggered a cycle that one of its
// per-owner child runs failed. The message is generic (owner username + cycle ID only) — no
// URLs, no query text, no internal error details (FR-012, FR-015).
func (s *NotificationService) NotifyAdminCycleChildFailure(adminID uint, ownerUsername string, cycleID uint) {
	if adminID == 0 {
		return
	}
	if ownerUsername == "" {
		ownerUsername = "a user"
	}

	title := "Wishlist availability check failed"
	message := fmt.Sprintf("The wishlist availability check for %s failed (cycle #%d). %s",
		ownerUsername, cycleID, models.GenericAvailabilityFailureMessage)
	refURL := fmt.Sprintf("/admin/availability-cycles/%d", cycleID)

	n := &models.Notification{
		UserID:       adminID,
		Type:         NotificationTypeAvailabilityRun,
		Title:        title,
		Message:      message,
		ReferenceID:  cycleID,
		ReferenceURL: refURL,
	}
	if err := s.notifRepo.Create(n); err != nil {
		s.logger.Error("notifications", "Failed to create admin cycle-child-failure notification for admin %d, cycle %d: %v", adminID, cycleID, err)
	}

}

// summarizeUnavailableCoinNames formats up to 3 newly-unavailable coin names plus an
// "and N more" suffix for the remainder, for use in the generic availability-run summary.
func summarizeUnavailableCoinNames(names []string) string {
	const maxShown = 3
	if len(names) == 0 {
		return ""
	}
	shown := names
	suffix := ""
	if len(names) > maxShown {
		shown = names[:maxShown]
		suffix = fmt.Sprintf(", and %d more", len(names)-maxShown)
	}
	return fmt.Sprintf("Newly unavailable: %s%s.", strings.Join(shown, ", "), suffix)
}

// NotifyNewCoin creates notifications for all accepted followers when a user
// adds a new coin to their collection (non-wishlist only).
func (s *NotificationService) NotifyNewCoin(ownerID uint, coin models.Coin) {
	if coin.IsWishlist {
		return
	}

	followers, err := s.socialRepo.GetAcceptedFollowerIDs(ownerID)
	if err != nil {
		s.logger.Error("notifications", "Failed to get followers for user %d: %v", ownerID, err)
		return
	}

	if len(followers) == 0 {
		return
	}

	// Look up the owner's username for the message
	ownerName := fmt.Sprintf("User #%d", ownerID)
	if user, err := s.socialRepo.GetUserByID(ownerID); err == nil && user != nil {
		ownerName = user.Username
	}

	coinName := coin.Name
	if coinName == "" {
		coinName = "a new coin"
	}

	for _, followerID := range followers {
		n := &models.Notification{
			UserID:      followerID,
			Type:        "friend_new_coin",
			Title:       "New coin added",
			Message:     fmt.Sprintf("%s added %s to their collection.", ownerName, coinName),
			ReferenceID: coin.ID,
		}
		if err := s.notifRepo.Create(n); err != nil {
			s.logger.Error("notifications", "Failed to notify follower %d about coin %d: %v", followerID, coin.ID, err)
		}
	}

	s.logger.Debug("notifications", "Notified %d followers about new coin %d from user %d", len(followers), coin.ID, ownerID)
}

// NotifyFollowRequest creates a notification for a user who received a new
// follower request.
func (s *NotificationService) NotifyFollowRequest(followerID, targetID uint) {
	if followerID == 0 || targetID == 0 || followerID == targetID {
		return
	}

	followerName := fmt.Sprintf("User #%d", followerID)
	if user, err := s.socialRepo.GetUserByID(followerID); err == nil && user != nil {
		followerName = user.Username
	}

	title := "New follower request"
	message := fmt.Sprintf("%s requested to follow you.", followerName)

	n := &models.Notification{
		UserID:       targetID,
		Type:         NotificationTypeFollowRequest,
		Title:        title,
		Message:      message,
		ReferenceID:  followerID,
		ReferenceURL: "/followers",
	}

	if err := s.notifRepo.Create(n); err != nil {
		s.logger.Error("notifications", "Failed to create follow-request notification for user %d from follower %d: %v", targetID, followerID, err)
		return
	}

}

// NotifyCoinOfDay creates an in-app notification for the user's daily
// featured coin. The ReferenceID points to the FeaturedCoin record
// so the frontend can open the dedicated modal.
func (s *NotificationService) NotifyCoinOfDay(userID uint, featuredCoinID, coinID uint, coinName, summary string) {
	if coinName == "" {
		coinName = "Today's coin"
	}

	title := "Coin of the Day"
	message := coinName
	if summary != "" {
		// Keep notification message short — the modal shows the full summary.
		preview := summary
		if len(preview) > 140 {
			preview = preview[:137] + "..."
		}
		message = fmt.Sprintf("%s — %s", coinName, preview)
	}

	n := &models.Notification{
		UserID:      userID,
		Type:        "coin_of_day",
		Title:       title,
		Message:     message,
		ReferenceID: featuredCoinID,
	}

	if err := s.notifRepo.Create(n); err != nil {
		s.logger.Error("notifications", "Failed to create coin-of-day notification for user %d: %v", userID, err)
	}

}

// NotifyPurchaseReminder creates an in-app notification when a purchase reminder's
// remind_date arrives. ReferenceID is the reminder ID; ReferenceURL deep-links to
// the coin detail page.
func (s *NotificationService) NotifyPurchaseReminder(userID, reminderID, coinID uint, coinName string) {
	if coinName == "" {
		coinName = "Unnamed coin"
	}
	title := "Purchase Reminder"
	message := fmt.Sprintf("%s is on your wishlist — time to buy!", coinName)
	refURL := fmt.Sprintf("/coin/%d", coinID)
	n := &models.Notification{
		UserID:       userID,
		Type:         NotificationTypePurchaseReminder,
		Title:        title,
		Message:      message,
		ReferenceID:  reminderID,
		ReferenceURL: refURL,
	}
	if err := s.notifRepo.Create(n); err != nil {
		s.logger.Error("notifications", "Failed to create purchase reminder notification for user %d, reminder %d: %v", userID, reminderID, err)
	}
}

// NotifyAIJobCompleted creates a notification when an asynchronous AI job completes.
func (s *NotificationService) NotifyAIJobCompleted(userID, jobID, coinID uint, coinName, jobType string) {
	if coinName == "" {
		coinName = "coin"
	}
	label := formatAIJobType(jobType)
	title := fmt.Sprintf("AI %s complete", label)
	message := fmt.Sprintf("%s is ready.", coinName)
	refURL := fmt.Sprintf("/coin/%d", coinID)
	n := &models.Notification{
		UserID:       userID,
		Type:         "ai_job_completed",
		Title:        title,
		Message:      message,
		ReferenceID:  jobID,
		ReferenceURL: refURL,
	}
	if err := s.notifRepo.Create(n); err != nil {
		s.logger.Error("notifications", "Failed to create AI job completion notification for user %d, job %d: %v", userID, jobID, err)
	}
}

// NotifyAIJobFailed creates a notification when an asynchronous AI job fails.
func (s *NotificationService) NotifyAIJobFailed(userID, jobID, coinID uint, jobType, reason string) {
	label := formatAIJobType(jobType)
	title := fmt.Sprintf("AI %s failed", label)
	message := fmt.Sprintf("AI %s could not be completed.", label)
	if reason != "" {
		message = fmt.Sprintf("%s Please check AI provider configuration and try again.", message)
	}
	refURL := fmt.Sprintf("/coin/%d", coinID)
	n := &models.Notification{
		UserID:       userID,
		Type:         "ai_job_failed",
		Title:        title,
		Message:      message,
		ReferenceID:  jobID,
		ReferenceURL: refURL,
	}
	if err := s.notifRepo.Create(n); err != nil {
		s.logger.Error("notifications", "Failed to create AI job failure notification for user %d, job %d: %v", userID, jobID, err)
	}
}

// NotifyValuationRunComplete creates an in-app notification when a background valuation run completes.
func (s *NotificationService) NotifyValuationRunComplete(userID, runID uint, checked, updated, skipped, errors int) {
	title := "Valuation complete"
	message := fmt.Sprintf("Checked: %d | Updated: %d | Skipped: %d | Errors: %d", checked, updated, skipped, errors)
	n := &models.Notification{
		UserID:       userID,
		Type:         "valuation_complete",
		Title:        title,
		Message:      message,
		ReferenceID:  runID,
		ReferenceURL: "/stats/health",
	}
	if err := s.notifRepo.Create(n); err != nil {
		s.logger.Error("notifications", "Failed to create valuation completion notification for user %d, run %d: %v", userID, runID, err)
	}
}

// NotifyAPIKeyRotationRequired creates a single actionable notification that lists
// active API key names that must be recreated.
func (s *NotificationService) NotifyAPIKeyRotationRequired(userID uint, keyNames []string) error {
	if len(keyNames) == 0 {
		return nil
	}

	normalized := make([]string, 0, len(keyNames))
	seen := make(map[string]struct{}, len(keyNames))
	for _, keyName := range keyNames {
		name := strings.TrimSpace(keyName)
		if name == "" {
			continue
		}
		if _, exists := seen[name]; exists {
			continue
		}
		seen[name] = struct{}{}
		normalized = append(normalized, name)
	}
	if len(normalized) == 0 {
		return nil
	}
	sort.Strings(normalized)

	n := &models.Notification{
		UserID:       userID,
		Type:         NotificationTypeAPIKeyRotationRequired,
		Title:        "Action required: Recreate API keys",
		Message:      fmt.Sprintf("Recreate these API keys in Settings: %s", strings.Join(normalized, ", ")),
		ReferenceURL: "/settings",
	}
	if err := s.notifRepo.ReplaceByUserAndType(n); err != nil {
		s.logger.Error("notifications", "Failed to create API key rotation notification for user %d: %v", userID, err)
		return err
	}
	return nil
}

func formatAIJobType(jobType string) string {
	switch jobType {
	case "analysis":
		return "analysis"
	case "value_estimate":
		return "value estimate"
	default:
		return "AI job"
	}
}
