package handlers

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/briandenicola/ancient-coins-api/models"
	"github.com/briandenicola/ancient-coins-api/services"
	"github.com/gin-gonic/gin"
)

// maxExportCoins bounds a spreadsheet export.
const maxExportCoins = 10000

// Export downloads the coins as a spreadsheet: all of them, or those a
// gallery's filters keep, in its order.
//
//	@Summary		Export coins as a spreadsheet
//	@Description	Downloads, as an Excel workbook (xlsx) or a CSV, the user's coins and banknotes that the filters keep (the same filters and order as GET /coins, without pages; up to 10,000), one row each with its details, purchase, value, references and sets. Columns are named in the language asked for; a CSV in Portuguese uses ";" and a decimal comma, as Excel there expects.
//	@Tags			Coins
//	@Produce		application/vnd.openxmlformats-officedocument.spreadsheetml.sheet
//	@Produce		text/csv
//	@Param			format			query	string	false	"Spreadsheet format"	Enums(xlsx, csv)	default(xlsx)
//	@Param			lang			query	string	false	"Language of the columns"	Enums(pt, en)	default(pt)
//	@Param			kind			query	string	false	"Only coins or only banknotes (every item when omitted)"	Enums(coin, banknote)
//	@Param			wishlist		query	string	false	"Filter by wishlist status"	Enums(true, false)
//	@Param			country			query	string	false	"Filter by country (exact)"
//	@Param			denomination	query	string	false	"Filter by face value, as written"
//	@Param			presentation	query	string	false	"Filter by presentation"	Enums(loose, coincard, capsule, flip, proof_set, blister, slab)
//	@Param			rarity			query	string	false	"Filter by rarity"	Enums(common, uncommon, scarce, rare, very_rare, extremely_rare)
//	@Param			repeated		query	bool	false	"Only coins held more than once"
//	@Param			extinct			query	bool	false	"Only coins of countries that no longer exist"
//	@Param			circulation		query	string	false	"Only current or demonetized coins, as Numista says"	Enums(current, demonetized)
//	@Param			commemorative	query	bool	false	"Only commemorative (true) or regular (false) coins, as Numista says"
//	@Param			classification	query	string	false	"Only one Numista kind of coin"
//	@Param			set				query	int		false	"Only the coins of a set"
//	@Param			search			query	string	false	"Search, as in the coin list"
//	@Param			sort			query	string	false	"Sort field"	Enums(created_at, updated_at, current_value, purchase_date, name, quantity, random)	default(updated_at)
//	@Param			order			query	string	false	"Sort order"	Enums(asc, desc)	default(desc)
//	@Success		200				"The spreadsheet"
//	@Failure		400				{object}	ErrorResponse
//	@Failure		500				{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/coins/export [get]
func (h *CoinHandler) Export(c *gin.Context) {
	format := strings.ToLower(strings.TrimSpace(c.DefaultQuery("format", "xlsx")))
	if format != "xlsx" && format != "csv" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "format must be 'xlsx' or 'csv'"})
		return
	}
	filters, ok := h.listFilters(c)
	if !ok {
		return
	}
	userID := c.GetUint("userId")
	coins := []models.Coin{}
	filters.Limit = 100
	for page := 1; len(coins) < maxExportCoins; page++ {
		filters.Page = page
		batch, total, err := h.repo.List(userID, filters)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch coins"})
			return
		}
		coins = append(coins, batch...)
		if len(batch) == 0 || int64(len(coins)) >= total {
			break
		}
	}
	if len(coins) > maxExportCoins {
		coins = coins[:maxExportCoins]
	}

	currency := services.CurrencyUSD
	if h.settingsSvc != nil {
		currency = h.settingsSvc.CollectionCurrency()
	}
	export := services.BuildCollectionExport(coins, services.ExportLang(c.Query("lang")), currency, time.Now())
	var body []byte
	var err error
	contentType := "text/csv; charset=utf-8"
	if format == "csv" {
		body, err = export.CSV()
	} else {
		body, err = export.XLSX()
		contentType = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to make the spreadsheet"})
		return
	}
	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, export.FileName(format)))
	c.Data(http.StatusOK, contentType, body)
}
