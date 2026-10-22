package handlers

import (
	"context"
	"net/http"

	"github.com/briandenicola/ancient-coins-api/services"
	"github.com/gin-gonic/gin"
)

// AdminNumistaValuesHandler shows and starts the batches that update coin
// values from Numista's prices.
type AdminNumistaValuesHandler struct {
	svc       *services.NumistaValueService
	scheduler *services.NumistaValueScheduler
	logger    *services.Logger
}

// NewAdminNumistaValuesHandler creates the handler.
func NewAdminNumistaValuesHandler(svc *services.NumistaValueService, scheduler *services.NumistaValueScheduler, logger *services.Logger) *AdminNumistaValuesHandler {
	return &AdminNumistaValuesHandler{svc: svc, scheduler: scheduler, logger: logger}
}

// Status tells how the daily update of coin values from Numista stands.
//
//	@Summary		Numista values batch status
//	@Description	The month's Numista requests and whether they run low (from 75% of the quota) or are spent for background requests, the next batch's budget, the coins it covers (linked to Numista with a grade Numista prices) and those due a check (none in 30 days), and the latest batches. Admin only.
//	@Tags			Admin
//	@Produce		json
//	@Success		200	{object}	models.NumistaValueStatus
//	@Failure		401	{object}	ErrorResponse
//	@Failure		403	{object}	ErrorResponse
//	@Failure		500	{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/admin/numista-values [get]
func (h *AdminNumistaValuesHandler) Status(c *gin.Context) {
	status, err := h.svc.Status(h.scheduler.GetStatus().NextRunIn)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read the Numista values status"})
		return
	}
	c.JSON(http.StatusOK, status)
}

// Run starts a batch now, whether the daily batch is on or off.
//
//	@Summary		Run a Numista values batch now
//	@Description	Starts one batch in the background: it updates the current value of the coins due from Numista's price for their grade, within today's share of the monthly quota. The previous value stays in each coin's value history. Admin only.
//	@Tags			Admin
//	@Produce		json
//	@Success		202	{object}	MessageResponse
//	@Failure		401	{object}	ErrorResponse
//	@Failure		403	{object}	ErrorResponse
//	@Failure		409	{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/admin/numista-values/run [post]
func (h *AdminNumistaValuesHandler) Run(c *gin.Context) {
	if h.svc.IsRunning() {
		c.JSON(http.StatusConflict, gin.H{"error": "A batch is already running"})
		return
	}
	go func() {
		if _, err := h.svc.Run(context.Background(), "manual"); err != nil {
			h.logger.Warn("numista-values", "Manual Numista values batch did not run: %v", err)
		}
	}()
	c.JSON(http.StatusAccepted, gin.H{"message": "Batch started"})
}
