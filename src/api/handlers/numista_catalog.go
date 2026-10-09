package handlers

import (
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/briandenicola/ancient-coins-api/models"
	"github.com/briandenicola/ancient-coins-api/services"
	"github.com/gin-gonic/gin"
)

var issuerCodePattern = regexp.MustCompile(`^[a-z0-9_-]{1,64}$`)

// NumistaCatalogHandler lets users browse the Numista catalogue to add a coin
// with its details filled in.
type NumistaCatalogHandler struct {
	catalog     *services.NumistaCatalogService
	marketValue *services.NumistaMarketValueService
}

// NewNumistaCatalogHandler creates a new NumistaCatalogHandler.
func NewNumistaCatalogHandler(catalog *services.NumistaCatalogService, marketValue *services.NumistaMarketValueService) *NumistaCatalogHandler {
	return &NumistaCatalogHandler{catalog: catalog, marketValue: marketValue}
}

// Issuers finds countries and issuing authorities by name.
//
//	@Summary		Search Numista issuers
//	@Description	Returns up to 20 Numista issuers (countries and authorities) whose name matches q, best matches first. The issuer list is cached for 30 days.
//	@Tags			Numista
//	@Produce		json
//	@Param			q		query		string	true	"Part of the issuer name"
//	@Param			lang	query		string	false	"Language of issuer names"	default(en)
//	@Success		200		{object}	models.NumistaCatalogIssuersResult
//	@Failure		400		{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/numista/catalog/issuers [get]
func (h *NumistaCatalogHandler) Issuers(c *gin.Context) {
	lang, ok := numistaQueryLang(c)
	if !ok {
		return
	}
	query := strings.TrimSpace(c.Query("q"))
	if len([]rune(query)) > 100 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "q must be at most 100 characters"})
		return
	}
	c.JSON(http.StatusOK, h.catalog.Issuers(c.Request.Context(), lang, query))
}

// Search finds coin types in the Numista catalogue.
//
//	@Summary		Search the Numista coin catalogue
//	@Description	Returns one page (20 types) of Numista coin types matching the text and/or issuer, optionally limited to a year. Searches are cached for a day.
//	@Tags			Numista
//	@Produce		json
//	@Param			q		query		string	false	"Search text (required unless issuer is set)"
//	@Param			issuer	query		string	false	"Numista issuer code, e.g. portugal"
//	@Param			year	query		int		false	"Year the type was issued (negative for BC)"
//	@Param			page	query		int		false	"Page number"	default(1)
//	@Param			lang	query		string	false	"Language of results"	default(en)
//	@Success		200		{object}	models.NumistaCatalogSearchResult
//	@Failure		400		{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/numista/catalog/search [get]
func (h *NumistaCatalogHandler) Search(c *gin.Context) {
	lang, ok := numistaQueryLang(c)
	if !ok {
		return
	}
	query := strings.TrimSpace(c.Query("q"))
	issuer := strings.TrimSpace(c.Query("issuer"))
	if len([]rune(query)) > models.NumistaMaxQueryLength {
		c.JSON(http.StatusBadRequest, gin.H{"error": "q is too long"})
		return
	}
	if issuer != "" && !issuerCodePattern.MatchString(issuer) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "issuer must be a Numista issuer code"})
		return
	}
	if query == "" && issuer == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "q or issuer is required"})
		return
	}
	year, ok := optionalIntQuery(c, "year", -5000, 3000)
	if !ok {
		return
	}
	page, ok := optionalIntQuery(c, "page", 1, 500)
	if !ok {
		return
	}
	c.JSON(http.StatusOK, h.catalog.Search(c.Request.Context(), services.NumistaTypeSearch{
		Query: query, Issuer: issuer, Year: year, Page: page, Lang: lang,
	}))
}

// Type returns the catalogue details of a coin type, used to prefill a new coin.
//
//	@Summary		Get a Numista coin type
//	@Description	Returns a Numista coin type's catalogue details (denomination, rulers, composition, weight, size, mints, designs, references, reference photos with credit) and its issues. Cached for 30 days.
//	@Tags			Numista
//	@Produce		json
//	@Param			id		path		int		true	"Numista type ID"
//	@Param			lang	query		string	false	"Language of catalogue data"	default(en)
//	@Success		200		{object}	models.NumistaCatalogTypeResult
//	@Failure		400		{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/numista/catalog/types/{id} [get]
func (h *NumistaCatalogHandler) Type(c *gin.Context) {
	typeID, ok := numistaPathID(c)
	if !ok {
		return
	}
	lang, ok := numistaQueryLang(c)
	if !ok {
		return
	}
	c.JSON(http.StatusOK, h.catalog.Type(c.Request.Context(), typeID, lang))
}

// MarketValue returns Numista's estimated price per grade for a type's issue.
//
//	@Summary		Get a Numista type's market value
//	@Description	Returns Numista's estimated price per grade for an issue of a coin type, before a coin exists. Same caching and status rules as the coin market value.
//	@Tags			Numista
//	@Produce		json
//	@Param			id			path		int		true	"Numista type ID"
//	@Param			issueId		query		int		false	"Numista issue ID; defaults to the issue of year, else the first"
//	@Param			year		query		int		false	"Year used to pick the default issue (negative for BC)"
//	@Param			currency	query		string	false	"ISO 4217 currency code"	default(EUR)
//	@Success		200			{object}	models.CoinMarketValue
//	@Failure		400			{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/numista/catalog/types/{id}/market-value [get]
func (h *NumistaCatalogHandler) MarketValue(c *gin.Context) {
	typeID, ok := numistaPathID(c)
	if !ok {
		return
	}
	issueID, ok := numistaQueryID(c, "issueId")
	if !ok {
		return
	}
	year, ok := optionalIntQuery(c, "year", -5000, 3000)
	if !ok {
		return
	}
	currency, ok := numistaQueryCurrency(c)
	if !ok {
		return
	}
	dateHint := ""
	if year < 0 {
		dateHint = strconv.Itoa(-year) + " BC"
	} else if year > 0 {
		dateHint = strconv.Itoa(year)
	}
	c.JSON(http.StatusOK, h.marketValue.ForType(c.Request.Context(), typeID, issueID, dateHint, currency, ""))
}

func numistaPathID(c *gin.Context) (int, bool) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 || id > models.NumistaMaxID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Numista type ID"})
		return 0, false
	}
	return id, true
}

// optionalIntQuery reads an optional integer within [minimum, maximum]; 0 when absent.
func optionalIntQuery(c *gin.Context, name string, minimum, maximum int) (int, bool) {
	raw := strings.TrimSpace(c.Query(name))
	if raw == "" {
		return 0, true
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value < minimum || value > maximum || value == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": name + " is out of range"})
		return 0, false
	}
	return value, true
}
