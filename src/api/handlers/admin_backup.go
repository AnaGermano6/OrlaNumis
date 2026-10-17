package handlers

import (
	"errors"
	"net/http"

	"github.com/briandenicola/ancient-coins-api/services"
	"github.com/gin-gonic/gin"
)

// AdminBackupHandler serves the full backups (the whole database and every
// upload) to administrators: the list, a backup now, a download, an import,
// and a restore.
type AdminBackupHandler struct {
	backups   *services.BackupService
	scheduler *services.BackupScheduler
	settings  *services.SettingsService
	folder    string
	// restart stops the app, for its container to start it again and the
	// prepared restore to be applied.
	restart func()
}

func NewAdminBackupHandler(backups *services.BackupService, scheduler *services.BackupScheduler, settings *services.SettingsService, folder string, restart func()) *AdminBackupHandler {
	return &AdminBackupHandler{backups: backups, scheduler: scheduler, settings: settings, folder: folder, restart: restart}
}

// BackupsResponse lists the full backups, newest first, with the automatic
// backup's settings and how the last restore went.
type BackupsResponse struct {
	Backups []services.BackupInfo `json:"backups"`
	// Running is true while a backup is being made.
	Running bool `json:"running"`
	// Enabled, StartTime and Keep are the automatic backup's settings: on, its
	// daily time, how many automatic backups are kept.
	Enabled   bool   `json:"enabled"`
	StartTime string `json:"startTime" example:"03:00"`
	Keep      int    `json:"keep" example:"7"`
	// NextRunIn is the time to the next automatic backup, in seconds.
	NextRunIn int64 `json:"nextRunIn" example:"3600"`
	// Folder is where the backups are kept.
	Folder      string                  `json:"folder" example:"~/Aurearia-backups/aurearia-dev"`
	LastRestore *services.RestoreResult `json:"lastRestore,omitempty"`
}

// List godoc
//
//	@Summary		Full backups
//	@Description	The full backups of the backup folder, newest first: the automatic ones (made every day, the last ones kept), those made by hand, those made of the data a restore replaced, and the imported ones. Each says when its data was copied, its size and what it holds. Admin only.
//	@Tags			Admin
//	@Produce		json
//	@Success		200	{object}	BackupsResponse
//	@Failure		401	{object}	ErrorResponse
//	@Failure		403	{object}	ErrorResponse
//	@Failure		500	{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/admin/backups [get]
func (h *AdminBackupHandler) List(c *gin.Context) {
	backups, err := h.backups.List()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read the backup folder"})
		return
	}
	response := BackupsResponse{
		Backups:     backups,
		Running:     h.backups.Running(),
		Enabled:     h.settings.GetSetting(services.SettingBackupEnabled) == "true",
		StartTime:   h.settings.GetSetting(services.SettingBackupStartTime),
		Keep:        h.backups.Keep(),
		Folder:      h.folder,
		LastRestore: h.backups.LastRestore(),
	}
	if h.scheduler != nil {
		response.NextRunIn = int64(h.scheduler.GetStatus().NextRunIn.Seconds())
	}
	c.JSON(http.StatusOK, response)
}

// Create godoc
//
//	@Summary		Make a full backup now
//	@Description	Starts a full backup (the whole database and every upload), made in the background: the list says when it is running, and shows it once made. Admin only.
//	@Tags			Admin
//	@Success		202
//	@Failure		401	{object}	ErrorResponse
//	@Failure		403	{object}	ErrorResponse
//	@Failure		409	{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/admin/backups [post]
func (h *AdminBackupHandler) Create(c *gin.Context) {
	if err := h.backups.Start(services.BackupKindManual); err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "A backup is already being made"})
		return
	}
	c.Status(http.StatusAccepted)
}

// Download godoc
//
//	@Summary		Download a full backup
//	@Description	The backup's zip: manifest.json, the database (database/ancientcoins.db) and every upload (uploads/…). It holds every user's data, password hashes included: keep it safe. Admin only.
//	@Tags			Admin
//	@Produce		application/zip
//	@Param			name	path	string	true	"The backup's name"
//	@Success		200		"The zip"
//	@Failure		401		{object}	ErrorResponse
//	@Failure		403		{object}	ErrorResponse
//	@Failure		404		{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/admin/backups/{name} [get]
func (h *AdminBackupHandler) Download(c *gin.Context) {
	path, err := h.backups.Path(c.Param("name"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Backup not found"})
		return
	}
	c.FileAttachment(path, c.Param("name"))
}

// Delete godoc
//
//	@Summary		Delete a full backup
//	@Description	Deletes a backup of the backup folder. Admin only.
//	@Tags			Admin
//	@Param			name	path	string	true	"The backup's name"
//	@Success		204
//	@Failure		401	{object}	ErrorResponse
//	@Failure		403	{object}	ErrorResponse
//	@Failure		404	{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/admin/backups/{name} [delete]
func (h *AdminBackupHandler) Delete(c *gin.Context) {
	if err := h.backups.Delete(c.Param("name")); err != nil {
		if errors.Is(err, services.ErrBackupNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Backup not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete the backup"})
		return
	}
	c.Status(http.StatusNoContent)
}

// Restore godoc
//
//	@Summary		Restore a full backup
//	@Description	Replaces all the data (every user's, the settings, the uploads) with the backup's. The backup is checked, the data of now is backed up first (a "pre-restore" backup, to go back to it), then the app restarts and applies the backup as it starts: in Docker its container starts it again by itself. Signed-in users may have to sign in again. Admin only.
//	@Tags			Admin
//	@Produce		json
//	@Param			name	path		string	true	"The backup's name"
//	@Success		202		{object}	map[string]bool
//	@Failure		400		{object}	ErrorResponse
//	@Failure		401		{object}	ErrorResponse
//	@Failure		403		{object}	ErrorResponse
//	@Failure		404		{object}	ErrorResponse
//	@Failure		409		{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/admin/backups/{name}/restore [post]
func (h *AdminBackupHandler) Restore(c *gin.Context) {
	err := h.backups.StageRestore(c.Param("name"))
	switch {
	case errors.Is(err, services.ErrBackupNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "Backup not found"})
		return
	case errors.Is(err, services.ErrBackupRunning):
		c.JSON(http.StatusConflict, gin.H{"error": "A backup is being made: try again when it is done"})
		return
	case err != nil:
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusAccepted, gin.H{"restarting": true})
	if h.restart != nil {
		h.restart()
	}
}

// Upload godoc
//
//	@Summary		Import a full backup
//	@Description	Keeps a backup's zip (a download of an earlier backup, of this app or another) in the backup folder, after checking it can be restored; restoring it is a second step. The body is the zip itself. Admin only.
//	@Tags			Admin
//	@Accept			application/zip
//	@Produce		json
//	@Param			backup	body		string	true	"The backup's zip"
//	@Success		201		{object}	services.BackupInfo
//	@Failure		400		{object}	ErrorResponse
//	@Failure		401		{object}	ErrorResponse
//	@Failure		403		{object}	ErrorResponse
//	@Failure		413		{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/admin/backups/upload [put]
func (h *AdminBackupHandler) Upload(c *gin.Context) {
	info, err := h.backups.SaveUpload(c.Request.Body)
	if err != nil {
		var tooLarge *http.MaxBytesError
		if errors.As(err, &tooLarge) {
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "The file is too large"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, info)
}
