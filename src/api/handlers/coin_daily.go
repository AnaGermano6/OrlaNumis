package handlers

import (
	"hash/fnv"
	"net/http"
	"strconv"
	"time"

	"github.com/briandenicola/ancient-coins-api/models"
	"github.com/briandenicola/ancient-coins-api/repository"
	"github.com/gin-gonic/gin"
)

// DailyCoinResponse is the coin or banknote of the day; Coin is null when
// the collection has none of that kind.
type DailyCoinResponse struct {
	Date string       `json:"date" example:"2026-10-13"`
	Coin *models.Coin `json:"coin"`
}

// DailyCoinHandler picks a coin of the collection for each day.
type DailyCoinHandler struct {
	coins *repository.CoinRepository
	now   func() time.Time
}

// NewDailyCoinHandler creates the handler.
func NewDailyCoinHandler(coins *repository.CoinRepository) *DailyCoinHandler {
	return &DailyCoinHandler{coins: coins, now: time.Now}
}

// Daily godoc
//
//	@Summary		Coin of the day
//	@Description	One coin (or banknote) of the collection for the day, preferring those with a photo: the same all day, another the next.
//	@Tags			Coins
//	@Produce		json
//	@Param			kind	query		string	false	"A coin (default) or a banknote"	Enums(coin, banknote)
//	@Param			date	query		string	false	"The viewer's day, YYYY-MM-DD (default: today, UTC)"
//	@Success		200		{object}	DailyCoinResponse
//	@Failure		400		{object}	ErrorResponse
//	@Failure		401		{object}	ErrorResponse
//	@Failure		500		{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/coins/daily [get]
func (h *DailyCoinHandler) Daily(c *gin.Context) {
	kind, ok := coinKindQuery(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "kind must be 'coin' or 'banknote'"})
		return
	}
	if kind == "" {
		kind = models.CoinKindCoin
	}
	date := h.now().UTC().Format(time.DateOnly)
	if requested := c.Query("date"); requested != "" {
		day, err := time.Parse(time.DateOnly, requested)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "date must be YYYY-MM-DD"})
			return
		}
		date = day.Format(time.DateOnly)
	}
	userID := c.GetUint("userId")
	seed := fnv.New32a()
	_, _ = seed.Write([]byte(strconv.FormatUint(uint64(userID), 10) + "|" + kind + "|" + date))
	coin, err := h.coins.DailyCoin(userID, kind, seed.Sum32())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to pick the coin of the day"})
		return
	}
	c.JSON(http.StatusOK, DailyCoinResponse{Date: date, Coin: coin})
}
