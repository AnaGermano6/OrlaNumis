package handlers

import (
	"net/http"

	"github.com/briandenicola/ancient-coins-api/services"
	"github.com/gin-gonic/gin"
)

// AdminHealthHandler serves aggregate health metrics for admin users.
type AdminHealthHandler struct {
	svc    *services.HealthService
	logger *services.Logger
}

// NewAdminHealthHandler creates a new AdminHealthHandler.
func NewAdminHealthHandler(svc *services.HealthService, logger *services.Logger) *AdminHealthHandler {
	return &AdminHealthHandler{svc: svc, logger: logger}
}

// Summary returns aggregate health metrics across users.
//
//	@Summary		Get admin collection health summary
//	@Description	Returns aggregate collection health metrics across all users. Admin only.
//	@Tags			Admin
//	@Produce		json
//	@Success		200	{object}	services.AdminHealthSummary
//	@Failure		401	{object}	ErrorResponse
//	@Failure		403	{object}	ErrorResponse
//	@Failure		500	{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/admin/health/summary [get]
func (h *AdminHealthHandler) Summary(c *gin.Context) {
	summary, err := h.svc.GetAdminHealthSummary()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch admin health summary"})
		return
	}
	c.JSON(http.StatusOK, summary)
}
