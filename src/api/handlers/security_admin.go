package handlers

import (
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/briandenicola/ancient-coins-api/repository"
	"github.com/briandenicola/ancient-coins-api/services"
	"github.com/gin-gonic/gin"
)

type SecurityExposureConfig struct {
	PublicAppURL             string
	CORSOrigins              []string
	TrustedProxiesConfigured bool
	AgentInternalTokenSet    bool
	RegistrationMode         string
	BackupStatus             string
}

type SecurityAdminHandler struct {
	securitySvc *services.SecurityService
	settingsSvc *services.SettingsService
	cfg         SecurityExposureConfig
}

func NewSecurityAdminHandler(securitySvc *services.SecurityService, settingsSvc *services.SettingsService, cfg SecurityExposureConfig) *SecurityAdminHandler {
	return &SecurityAdminHandler{securitySvc: securitySvc, settingsSvc: settingsSvc, cfg: cfg}
}

// SecuritySummary returns a 24h auth/security event summary.
//
//	@Summary		Get security summary
//	@Description	Returns aggregate auth/security counters and backup status. Admin only.
//	@Tags			Admin Security
//	@Produce		json
//	@Success		200	{object}	map[string]interface{}
//	@Failure		500	{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/admin/security/summary [get]
func (h *SecurityAdminHandler) SecuritySummary(c *gin.Context) {
	summary, err := h.securitySvc.Summary()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load security summary"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"summary": summary, "backupStatus": h.settingsSvc.GetSetting(services.SettingBackupStatus)})
}

// SecurityEvents returns persisted auth/security events.
//
//	@Summary		List security events
//	@Description	Returns persisted auth/security events with optional type, username, clientIp, and limit filters. Admin only.
//	@Tags			Admin Security
//	@Produce		json
//	@Param			type		query		string	false	"Event type"
//	@Param			username	query		string	false	"Username"
//	@Param			clientIp	query		string	false	"Client IP"
//	@Param			limit		query		int		false	"Maximum events"	default(100)
//	@Success		200			{object}	map[string]interface{}
//	@Failure		500			{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/admin/security/events [get]
func (h *SecurityAdminHandler) SecurityEvents(c *gin.Context) {
	limit := 100
	if raw := c.Query("limit"); raw != "" {
		if n, err := strconv.Atoi(raw); err == nil && n > 0 && n <= 500 {
			limit = n
		}
	}
	events, err := h.securitySvc.ListEvents(repository.SecurityEventFilters{
		Type:     c.Query("type"),
		Username: c.Query("username"),
		ClientIP: c.Query("clientIp"),
		Limit:    limit,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load security events"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"events": events, "count": len(events)})
}

// UnlockUser clears a persisted account lockout.
//
//	@Summary		Unlock user
//	@Description	Clears a user's account lock state. Admin only.
//	@Tags			Admin Security
//	@Produce		json
//	@Param			id	path		int	true	"User ID"
//	@Success		200	{object}	MessageResponse
//	@Failure		400	{object}	ErrorResponse
//	@Failure		500	{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/admin/users/{id}/unlock [post]
func (h *SecurityAdminHandler) UnlockUser(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}
	if err := h.securitySvc.UnlockUser(uint(id), c.GetUint("userId")); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to unlock user"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "User unlocked"})
}

// ExposureCheck returns public-facing configuration warnings.
//
//	@Summary		Get exposure check
//	@Description	Returns release-readiness warnings for public-facing configuration. Admin only.
//	@Tags			Admin Security
//	@Produce		json
//	@Success		200	{object}	map[string]interface{}
//	@Security		BearerAuth
//	@Router			/admin/security/exposure-check [get]
func (h *SecurityAdminHandler) ExposureCheck(c *gin.Context) {
	mode := h.settingsSvc.GetSetting(services.SettingRegistrationMode)
	publicURL := h.settingsSvc.GetSetting(services.SettingPublicAppURL)
	warnings := make([]string, 0)
	if publicURL == "" {
		warnings = append(warnings, "PublicAppURL is blank")
	} else if _, err := url.ParseRequestURI(publicURL); err != nil {
		warnings = append(warnings, "PublicAppURL is invalid")
	}
	if os.Getenv("GIN_MODE") == "release" {
		if len(h.cfg.CORSOrigins) == 0 {
			warnings = append(warnings, "CORS origins are not explicitly configured")
		}
		for _, origin := range h.cfg.CORSOrigins {
			if origin == "*" || strings.Contains(origin, "localhost") {
				warnings = append(warnings, "CORS origin is not release-safe: "+origin)
			}
		}
		if !h.cfg.TrustedProxiesConfigured {
			warnings = append(warnings, "Trusted proxies are not configured")
		}
	}
	if mode != "closed" {
		warnings = append(warnings, "Registration mode is "+mode)
	}
	c.JSON(http.StatusOK, gin.H{
		"warnings": warnings,
		"config": gin.H{
			"publicAppURL":             publicURL,
			"trustedProxiesConfigured": h.cfg.TrustedProxiesConfigured,
			"agentInternalTokenSet":    h.cfg.AgentInternalTokenSet,
			"registrationMode":         mode,
			"backupStatus":             h.settingsSvc.GetSetting(services.SettingBackupStatus),
		},
	})
}
