package handlers

import (
	"errors"
	"net/http"

	"github.com/briandenicola/ancient-coins-api/services"
	"github.com/gin-gonic/gin"
)

// AdminUpdatesHandler tells the admin about new versions and starts an
// update through Watchtower.
type AdminUpdatesHandler struct {
	updates *services.UpdateService
}

// NewAdminUpdatesHandler creates the handler.
func NewAdminUpdatesHandler(updates *services.UpdateService) *AdminUpdatesHandler {
	return &AdminUpdatesHandler{updates: updates}
}

// Status tells which version runs and whether a newer one is out.
//
//	@Summary		Get the update status
//	@Description	Returns the running version, the latest released version (from the project's GitHub releases, checked every few hours) with its notes, and whether Watchtower is set up to update the container. Admin only.
//	@Tags			Admin
//	@Produce		json
//	@Success		200	{object}	services.UpdateStatus
//	@Failure		401	{object}	ErrorResponse
//	@Failure		403	{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/admin/updates [get]
func (h *AdminUpdatesHandler) Status(c *gin.Context) {
	c.JSON(http.StatusOK, h.updates.Status(c.Request.Context(), false))
}

// Check looks for a new version now.
//
//	@Summary		Check for a new version now
//	@Description	Asks GitHub for the latest release now, instead of using the last check. Admin only.
//	@Tags			Admin
//	@Produce		json
//	@Success		200	{object}	services.UpdateStatus
//	@Failure		401	{object}	ErrorResponse
//	@Failure		403	{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/admin/updates/check [post]
func (h *AdminUpdatesHandler) Check(c *gin.Context) {
	c.JSON(http.StatusOK, h.updates.Status(c.Request.Context(), true))
}

// Apply backs up and asks Watchtower to update the container.
//
//	@Summary		Update the app now
//	@Description	Makes a backup (kind pre-update), then asks Watchtower to pull the latest image and recreate the container; the app restarts on the new version a moment later and migrates its data. Needs Watchtower's address and token in the settings. Admin only.
//	@Tags			Admin
//	@Produce		json
//	@Success		202	{object}	MessageResponse
//	@Failure		400	{object}	ErrorResponse
//	@Failure		401	{object}	ErrorResponse
//	@Failure		403	{object}	ErrorResponse
//	@Failure		409	{object}	ErrorResponse
//	@Failure		502	{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/admin/updates/apply [post]
func (h *AdminUpdatesHandler) Apply(c *gin.Context) {
	err := h.updates.Apply(c.Request.Context())
	switch {
	case err == nil:
		c.JSON(http.StatusAccepted, gin.H{"message": "Update started"})
	case errors.Is(err, services.ErrUpdateNotConfigured):
		c.JSON(http.StatusBadRequest, gin.H{"error": "Watchtower is not configured"})
	case errors.Is(err, services.ErrBackupRunning):
		c.JSON(http.StatusConflict, gin.H{"error": "A backup is already running; try again when it ends"})
	default:
		respondError(c, http.StatusBadGateway, "The update could not be started", err)
	}
}
