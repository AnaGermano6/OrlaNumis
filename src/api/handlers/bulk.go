package handlers

import (
	"net/http"

	"github.com/briandenicola/ancient-coins-api/repository"
	"github.com/briandenicola/ancient-coins-api/services"
	"github.com/gin-gonic/gin"
)

// BulkActionRequest is the request body for bulk coin actions.
type BulkActionRequest struct {
	CoinIDs []uint `json:"coinIds" binding:"required"`
	Action  string `json:"action" binding:"required"`
	SetID   *uint  `json:"setId"`
}

// BulkHandler handles bulk operations on coins.
type BulkHandler struct {
	coinRepo    *repository.CoinRepository
	setRepo     *repository.SetRepository
	coinService *services.CoinService
}

func (h *BulkHandler) WithCoinService(coinService *services.CoinService) *BulkHandler {
	h.coinService = coinService
	return h
}

// NewBulkHandler creates a new BulkHandler.
func NewBulkHandler(coinRepo *repository.CoinRepository, setRepo *repository.SetRepository) *BulkHandler {
	return &BulkHandler{coinRepo: coinRepo, setRepo: setRepo}
}

// BulkAction performs a bulk operation on the selected coins.
//
//	@Summary		Bulk coin action
//	@Description	Performs a bulk action (set, delete, export) on selected coins.
//	@Tags			Coins
//	@Accept			json
//	@Produce		json
//	@Param			body	body		BulkActionRequest	true	"Bulk action request"
//	@Success		200		{object}	map[string]interface{}
//	@Failure		400		{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/coins/bulk [post]
func (h *BulkHandler) BulkAction(c *gin.Context) {
	userID := c.GetUint("userId")

	var req BulkActionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "coinIds and action are required"})
		return
	}

	if len(req.CoinIDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No coins selected"})
		return
	}
	if len(req.CoinIDs) > 200 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Maximum 200 coins per bulk operation"})
		return
	}

	switch req.Action {
	case "delete":
		var affected int64
		var err error
		if h.coinService != nil {
			affected, err = h.coinService.BulkDeleteCoins(req.CoinIDs, userID)
		} else {
			affected, err = h.coinRepo.BulkDelete(req.CoinIDs, userID)
		}
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete coins"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Coins deleted", "affected": affected})

	case "set":
		if req.SetID == nil || *req.SetID == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "setId is required for set action"})
			return
		}
		affected, err := h.setRepo.BulkAddCoinsToSet(req.CoinIDs, *req.SetID, userID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Set applied", "affected": affected})

	case "export":
		coins, err := h.coinRepo.GetByIDs(req.CoinIDs, userID)
		if err != nil || len(coins) == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "No matching coins found"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"coins": coins, "total": len(coins)})

	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid action. Must be: set, delete or export"})
	}
}
