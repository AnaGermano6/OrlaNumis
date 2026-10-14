package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/briandenicola/ancient-coins-api/models"
	"github.com/briandenicola/ancient-coins-api/repository"
	"github.com/briandenicola/ancient-coins-api/services"
	"github.com/gin-gonic/gin"
)

// allowedListSortFields is the handler-level allowlist of sort fields for the
// coin list endpoint. Matches the repository's allowedSortFields map.
// "random" uses a seeded deterministic shuffle (see ?seed= query param).
var allowedListSortFields = map[string]bool{
	"created_at":    true,
	"updated_at":    true,
	"current_value": true,
	"purchase_date": true,
	"name":          true,
	"quantity":      true,
	"random":        true,
}

var nullableCoinUpdateScalarFields = map[string]string{
	"purchasePrice":         "PurchasePrice",
	"currentValue":          "CurrentValue",
	"purchaseDate":          "PurchaseDate",
	"soldPrice":             "SoldPrice",
	"soldDate":              "SoldDate",
	"weightGrams":           "WeightGrams",
	"diameterMm":            "DiameterMm",
	"widthMm":               "WidthMm",
	"heightMm":              "HeightMm",
	"romanImperialFigureId": "RomanImperialFigureID",
	"mintLocationId":        "MintLocationID",
	"year":                  "Year",
}

func nullableScalarFieldPresence(raw map[string]json.RawMessage) map[string]bool {
	present := make(map[string]bool, len(nullableCoinUpdateScalarFields))
	for jsonField, modelField := range nullableCoinUpdateScalarFields {
		if _, ok := raw[jsonField]; ok {
			present[modelField] = true
		}
	}
	return present
}

type CoinHandler struct {
	repo        *repository.CoinRepository
	svc         *services.CoinService
	logger      *services.Logger
	settingsSvc *services.SettingsService
	circulation *services.CoinCirculationService
}

// WithCirculation lets the coin list keep only coins still legal tender (or
// only demonetized ones).
func (h *CoinHandler) WithCirculation(circulation *services.CoinCirculationService) *CoinHandler {
	h.circulation = circulation
	return h
}

func NewCoinHandler(repo *repository.CoinRepository, svc *services.CoinService, logger *services.Logger) *CoinHandler {
	return &CoinHandler{repo: repo, svc: svc, logger: logger}
}

// WithSettingsSupport enables MatchCategoryEra to consult the admin-defined
// CoinCategories/CoinEras lists in addition to the built-in defaults.
func (h *CoinHandler) WithSettingsSupport(settingsSvc *services.SettingsService) *CoinHandler {
	h.settingsSvc = settingsSvc
	return h
}

// MatchCategoryEraRequest is the payload for MatchCategoryEra.
type MatchCategoryEraRequest struct {
	Type  string `json:"type" binding:"required,oneof=category era"`
	Value string `json:"value" binding:"required,max=200"`
}

// MatchCategoryEraResponse reports whether a candidate category/era value
// confidently resolves to a known (built-in or admin-defined) value.
type MatchCategoryEraResponse struct {
	Match   string `json:"match"`
	Matched bool   `json:"matched"`
}

// MatchCategoryEra resolves a free-text category/era candidate (e.g. an
// AI-chat suggestion) against the built-in defaults and the admin-defined
// CoinCategories/CoinEras lists, using normalized/fuzzy matching. Intended
// for interactive callers that can fall back to asking the user when no
// confident match is found - it never guesses beyond that.
//
//	@Summary		Match a category/era candidate against known values
//	@Description	Resolves a free-text category or era value against built-in and admin-defined lists
//	@Tags			coins
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		MatchCategoryEraRequest	true	"Candidate value to match"
//	@Success		200		{object}	MatchCategoryEraResponse
//	@Failure		400		{object}	object{error=string}
//	@Router			/coins/match-category-era [post]
func (h *CoinHandler) MatchCategoryEra(c *gin.Context) {
	var req MatchCategoryEraRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var known []string
	switch req.Type {
	case "category":
		known = services.BuiltInCoinCategoryValues()
		if h.settingsSvc != nil {
			known = append(known, services.SplitSettingList(h.settingsSvc.GetSetting(services.SettingCoinCategories))...)
		}
	case "era":
		known = services.BuiltInCoinEraValues()
		if h.settingsSvc != nil {
			known = append(known, services.SplitSettingList(h.settingsSvc.GetSetting(services.SettingCoinEras))...)
		}
	}

	match, matched := services.MatchKnownValue(req.Value, known)
	c.JSON(http.StatusOK, MatchCategoryEraResponse{Match: match, Matched: matched})
}

// CoinOptionLabels holds the display names an admin set for stored category
// and era values in one interface language, keyed by the stored value.
type CoinOptionLabels struct {
	Categories map[string]string `json:"categories"`
	Eras       map[string]string `json:"eras"`
}

// CoinOptionsResponse lists the admin-configured category and era choices
// plus their translated display names, keyed by locale (e.g. "pt-PT").
type CoinOptionsResponse struct {
	Categories []string                    `json:"categories"`
	Eras       []string                    `json:"eras"`
	Labels     map[string]CoinOptionLabels `json:"labels"`
	// Currency is the collection currency coin prices and values are kept in.
	Currency string `json:"currency" example:"EUR"`
}

// GetCoinOptions returns the configured category and era choices with their
// translated display names, so every signed-in user (not only admins, who can
// read /admin/settings) sees the same vocabulary. Stored values are unchanged;
// the labels are display-only.
//
//	@Summary		Get coin category and era options
//	@Description	Returns the admin-configured category and era choices with translated display names, and the collection currency (EUR, USD, GBP…) coin prices and values are kept in
//	@Tags			Coins
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	CoinOptionsResponse
//	@Router			/coin-options [get]
func (h *CoinHandler) GetCoinOptions(c *gin.Context) {
	get := func(key string) string {
		if h.settingsSvc == nil {
			return ""
		}
		return h.settingsSvc.GetSetting(key)
	}
	categories := services.SplitSettingList(get(services.SettingCoinCategories))
	if len(categories) == 0 {
		categories = services.BuiltInCoinCategoryValues()
		sort.Strings(categories)
	}
	eras := services.SplitSettingList(get(services.SettingCoinEras))
	if len(eras) == 0 {
		eras = services.BuiltInCoinEraValues()
		sort.Strings(eras)
	}
	c.JSON(http.StatusOK, CoinOptionsResponse{
		Categories: categories,
		Eras:       eras,
		Currency:   h.settingsSvc.CollectionCurrency(),
		Labels: map[string]CoinOptionLabels{
			"pt-PT": {
				Categories: services.ParseSettingLabelMap(get(services.SettingCoinCategoryLabelsPtPT)),
				Eras:       services.ParseSettingLabelMap(get(services.SettingCoinEraLabelsPtPT)),
			},
		},
	})
}

// PurchaseRequest holds optional details when purchasing a wishlist coin.
type PurchaseRequest struct {
	PurchasePrice    *float64 `json:"purchasePrice"`
	PurchaseDate     string   `json:"purchaseDate"`
	PurchaseLocation string   `json:"purchaseLocation"`
}

// List returns a paginated list of coins for the authenticated user.
//
// The "total" field in the response reflects the total number of coins matching
// the applied filters. For example:
//   - /coins: total = all owned coins (including wishlist & sold)
//   - /coins?wishlist=false&sold=false: total = active collection count
//
// The active collection (owned AND NOT wishlist AND NOT sold) is the canonical
// definition of "collection size" and must match /stats totalCoins and the
// internal collection_summary tool.
//
//	@Summary		List coins
//	@Description	Returns a paginated, filterable list of coins belonging to the authenticated user.
//	@Tags			Coins
//	@Produce		json
//	@Param			kind		query		string	false	"Only coins or only banknotes (every item when omitted)"	Enums(coin, banknote)
//	@Param			category	query		string	false	"Filter by category (Roman, Greek, Byzantine, Modern, Other)"
//	@Param			era			query		string	false	"Filter by era (ancient, medieval, modern)"
//	@Param			country		query		string	false	"Filter by country (exact)"
//	@Param			presentation	query	string	false	"Filter by presentation"	Enums(loose, coincard, capsule, flip, proof_set, blister, slab)
//	@Param			repeated		query	bool	false	"Only coins held more than once (a quantity above one), the copies there to trade"
//	@Param			extinct			query	bool	false	"Only coins of countries that no longer exist (USSR, Yugoslavia, Ottoman Empire…)"
//	@Param			circulation		query	string	false	"Only collection coins issued for circulation and still legal tender (current), or no longer legal tender (demonetized), as Numista says of their types; collector (non-circulating) coins are not current"	Enums(current, demonetized)
//	@Param			commemorative	query	bool	false	"Only collection coins of commemorative types (true) or of regular ones (false), as Numista says"
//	@Param			classification	query	string	false	"Only collection coins of one Numista kind of coin: standard, circulating-commemorative, non-circulating, or another kind's label (see the summary's classifications)"
//	@Param			rarity		query		string	false	"Filter by rarity"	Enums(common, uncommon, scarce, rare, very_rare, extremely_rare)
//	@Param			search		query		string	false	"Search across name, denomination, ruler, era, mint, inscriptions, notes, country and a banknote's serial number, series and signatures"
//	@Param			wishlist	query		string	false	"Filter by wishlist status"	Enums(true, false)
//	@Param			sold		query		string	false	"Filter by sold status"	Enums(true, false)
//	@Param			page		query		int		false	"Page number"	default(1)
//	@Param			limit		query		int		false	"Items per page (max 100)"	default(50)
//	@Param			sort		query		string	false	"Sort field"	Enums(created_at, updated_at, current_value, purchase_date, name, quantity, random)	default(updated_at)
//	@Param			order		query		string	false	"Sort order"	Enums(asc, desc)	default(desc)
//	@Success		200			{object}	CoinListResponse
//	@Failure		400			{object}	ErrorResponse
//	@Failure		401			{object}	ErrorResponse
//	@Failure		500			{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/coins [get]
func (h *CoinHandler) List(c *gin.Context) {
	userID := c.GetUint("userId")

	// Validate page
	pageStr := c.DefaultQuery("page", "1")
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "page must be an integer >= 1"})
		return
	}

	// Validate limit
	limitStr := c.DefaultQuery("limit", "50")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 100 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "limit must be an integer between 1 and 100"})
		return
	}

	// Validate sort field against allowlist (defense-in-depth against SQL injection)
	sortField := c.DefaultQuery("sort", "updated_at")
	if !allowedListSortFields[sortField] {
		allowed := make([]string, 0, len(allowedListSortFields))
		for k := range allowedListSortFields {
			allowed = append(allowed, k)
		}
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("sort must be one of: %s", strings.Join(allowed, ", ")),
		})
		return
	}

	// Validate sort order
	sortOrder := c.DefaultQuery("order", "desc")
	if sortOrder != "asc" && sortOrder != "desc" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "order must be 'asc' or 'desc'"})
		return
	}

	kind, ok := coinKindQuery(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "kind must be 'coin' or 'banknote'"})
		return
	}
	filters := repository.CoinListFilters{
		Kind:         kind,
		Category:     c.Query("category"),
		Era:          c.Query("era"),
		Country:      models.NormalizeCoinCountry(c.Query("country")),
		Presentation: strings.TrimSpace(c.Query("presentation")),
		Repeated:     c.Query("repeated") == "true",
		Rarity:       strings.TrimSpace(c.Query("rarity")),
		Search:       c.Query("search"),
		SortField:    sortField,
		SortOrder:    sortOrder,
		Page:         page,
		Limit:        limit,
	}
	if !models.ValidCoinPresentation(filters.Presentation) || !models.ValidCoinRarity(filters.Rarity) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unknown presentation or rarity"})
		return
	}
	// Parse optional seed for deterministic random sort.
	// Use strconv.Atoi to ensure it's a real integer (defense against SQL injection).
	if sortField == "random" {
		if seedStr := c.Query("seed"); seedStr != "" {
			if seed, err := strconv.Atoi(seedStr); err == nil {
				filters.Seed = &seed
			}
		}
		if filters.Seed == nil {
			// Default seed if none provided — based on current minute to vary slightly.
			s := int(time.Now().Unix() % 1000000)
			filters.Seed = &s
		}
	}
	if w := c.Query("wishlist"); w == "true" {
		v := true
		filters.Wishlist = &v
	} else if w == "false" {
		v := false
		filters.Wishlist = &v
	}
	if s := c.Query("sold"); s == "true" {
		v := true
		filters.Sold = &v
	} else if s == "false" {
		v := false
		filters.Sold = &v
	}
	circulation, commemorative, classification := c.Query("circulation"), c.Query("commemorative"), strings.TrimSpace(c.Query("classification"))
	if (circulation != "" && circulation != "current" && circulation != "demonetized") || (commemorative != "" && commemorative != "true" && commemorative != "false") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "circulation must be 'current' or 'demonetized', commemorative 'true' or 'false'"})
		return
	}
	if (circulation != "" || commemorative != "" || classification != "") && h.circulation != nil {
		sorted, err := h.circulation.Classify(userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch coins"})
			return
		}
		switch circulation {
		case "current":
			filters.IDs = sorted.Current
		case "demonetized":
			filters.IDs = sorted.Demonetized
		}
		switch commemorative {
		case "true":
			filters.IDs = intersectIDs(filters.IDs, sorted.Commemorative)
		case "false":
			filters.IDs = intersectIDs(filters.IDs, sorted.Regular)
		}
		if classification != "" {
			filters.IDs = intersectIDs(filters.IDs, sorted.ClassificationCoins(classification))
		}
	}
	if c.Query("extinct") == "true" {
		countries, err := h.repo.DistinctCountries(userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch coins"})
			return
		}
		filters.Countries = []string{}
		for _, country := range countries {
			if services.IsExtinctCountry(country) {
				filters.Countries = append(filters.Countries, country)
			}
		}
	}
	if t := c.Query("tag"); t != "" {
		if tagID, err := strconv.ParseUint(t, 10, strconv.IntSize); err == nil {
			v := uint(tagID)
			filters.TagID = &v
		}
	}
	if s := c.Query("set"); s != "" {
		if setID, err := strconv.ParseUint(s, 10, strconv.IntSize); err == nil {
			v := uint(setID)
			filters.SetID = &v
		}
	}

	coins, total, err := h.repo.List(userID, filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch coins"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"coins": coins,
		"total": total,
		"page":  filters.Page,
		"limit": filters.Limit,
	})
}

// Get returns a single coin by ID for the authenticated user.
//
//	@Summary		Get a coin
//	@Description	Returns a single coin with its images, owned by the authenticated user.
//	@Tags			Coins
//	@Produce		json
//	@Param			id	path		int	true	"Coin ID"
//	@Success		200	{object}	models.Coin
//	@Failure		400	{object}	ErrorResponse
//	@Failure		401	{object}	ErrorResponse
//	@Failure		404	{object}	ErrorResponse
//	@Failure		500	{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/coins/{id} [get]
func (h *CoinHandler) Get(c *gin.Context) {
	userID := c.GetUint("userId")
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid coin ID"})
		return
	}

	coin, err := h.repo.FindByID(uint(id), userID)
	if err != nil {
		if repository.IsRecordNotFound(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Coin not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch coin"})
		return
	}

	c.JSON(http.StatusOK, coin)
}

// Status tells whether a coin is still legal tender and whether it is
// commemorative.
//
//	@Summary		Get a coin's legal tender and commemorative status
//	@Description	Tells whether one of the user's coins is still legal tender (current), whether it was issued for collectors rather than for circulation (nonCirculating) and whether it is commemorative, from its cached Numista type (never asking Numista); coins of countries that no longer exist are not current. A value is null when nobody knows (the coin is not linked to a cached Numista type).
//	@Tags			Coins
//	@Produce		json
//	@Param			id	path		int	true	"Coin ID"
//	@Success		200	{object}	services.CoinStatus
//	@Failure		400	{object}	ErrorResponse
//	@Failure		401	{object}	ErrorResponse
//	@Failure		404	{object}	ErrorResponse
//	@Failure		500	{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/coins/{id}/status [get]
func (h *CoinHandler) Status(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid coin ID"})
		return
	}
	coin, err := h.repo.FindByID(uint(id), c.GetUint("userId"))
	if err != nil {
		if repository.IsRecordNotFound(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Coin not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch coin"})
		return
	}
	if h.circulation == nil {
		c.JSON(http.StatusOK, services.CoinStatus{})
		return
	}
	c.JSON(http.StatusOK, h.circulation.Status(coin))
}

// CoinStatusGroups lists the collection's coins by legal tender and
// commemorative status; a coin nobody knows about is in neither list of a pair.
type CoinStatusGroups struct {
	// Current are circulation coins still legal tender.
	Current     []uint `json:"current"`
	Demonetized []uint `json:"demonetized"`
	// NonCirculating are collector coins (not demonetized).
	NonCirculating []uint `json:"nonCirculating"`
	Commemorative  []uint `json:"commemorative"`
	Regular        []uint `json:"regular"`
}

// StatusGroups tells, in one answer, which collection coins are current or
// demonetized and which are commemorative or not.
//
//	@Summary		List coins by legal tender and commemorative status
//	@Description	Returns the IDs of the user's collection coins (not wishlist or sold) that are circulation coins still legal tender (current), no longer legal tender (demonetized) or collector coins (nonCirculating), and that are commemorative or not, by the same rules as the summary and the coin list filters: a circulation set on the coin wins, else the cached Numista type says (never asking Numista); coins of extinct countries are not current. A coin nobody knows about is in neither list of a pair.
//	@Tags			Coins
//	@Produce		json
//	@Success		200	{object}	CoinStatusGroups
//	@Failure		401	{object}	ErrorResponse
//	@Failure		500	{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/coins/statuses [get]
func (h *CoinHandler) StatusGroups(c *gin.Context) {
	groups := CoinStatusGroups{Current: []uint{}, Demonetized: []uint{}, NonCirculating: []uint{}, Commemorative: []uint{}, Regular: []uint{}}
	if h.circulation == nil {
		c.JSON(http.StatusOK, groups)
		return
	}
	sorted, err := h.circulation.Classify(c.GetUint("userId"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch coins"})
		return
	}
	groups.Current, groups.Demonetized, groups.NonCirculating = sorted.Current, sorted.Demonetized, sorted.NonCirculating
	groups.Commemorative, groups.Regular = sorted.Commemorative, sorted.Regular
	c.JSON(http.StatusOK, groups)
}

// Create adds a new coin for the authenticated user.
//
//	@Summary		Create a coin
//	@Description	Creates a new coin record for the authenticated user. storageSlot is nullable and one-based; a tray assignment requires an available slot within the tray's persisted capacity.
//	@Tags			Coins
//	@Accept			json
//	@Produce		json
//	@Param			body	body		CoinCreateRequest	true	"Coin data"
//	@Success		201		{object}	models.Coin
//	@Failure		400		{object}	ErrorResponse
//	@Failure		401		{object}	ErrorResponse
//	@Failure		404		{object}	ErrorResponse
//	@Failure		409		{object}	SlotOccupiedErrorResponse
//	@Failure		500		{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/coins [post]
func (h *CoinHandler) Create(c *gin.Context) {
	logger := h.logger
	userID := c.GetUint("userId")

	var req CoinCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "Invalid request payload", err)
		return
	}

	coin := req.toCoin(userID)

	logger.Debug("coins", "Creating coin '%s' for user %d", coin.Name, userID)

	if err := h.svc.CreateCoin(&coin); err != nil {
		if handleCoinMutationError(c, err) {
			return
		}
		logger.Error("coins", "Failed to create coin: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create coin"})
		return
	}

	logger.Info("coins", "Created coin %d '%s' for user %d", coin.ID, coin.Name, userID)
	c.JSON(http.StatusCreated, coin)
}

// Update modifies an existing coin owned by the authenticated user.
//
//	@Summary		Update a coin
//	@Description	Updates an existing coin record. Only the coin owner can update it. storageSlot is nullable and one-based; omitted storage fields preserve the current assignment. Explicit null is accepted only when the resulting location/slot combination is valid, such as clearing the location and slot together.
//	@Tags			Coins
//	@Accept			json
//	@Produce		json
//	@Param			id		path		int			true	"Coin ID"
//	@Param			body	body		CoinUpdateRequest	true	"Updated coin data"
//	@Success		200		{object}	models.Coin
//	@Failure		400		{object}	ErrorResponse
//	@Failure		401		{object}	ErrorResponse
//	@Failure		404		{object}	ErrorResponse
//	@Failure		409		{object}	SlotOccupiedErrorResponse
//	@Failure		500		{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/coins/{id} [put]
func (h *CoinHandler) Update(c *gin.Context) {
	userID := c.GetUint("userId")
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid coin ID"})
		return
	}

	existing, err := h.repo.FindByID(uint(id), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Coin not found"})
		return
	}

	bodyBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		respondError(c, http.StatusBadRequest, "Invalid request payload", err)
		return
	}
	var raw map[string]json.RawMessage
	storageLocationProvided := false
	storageSlotProvided := false
	nullableScalarProvided := map[string]bool{}
	if err := json.Unmarshal(bodyBytes, &raw); err == nil {
		_, storageLocationProvided = raw["storageLocationId"]
		_, storageSlotProvided = raw["storageSlot"]
		nullableScalarProvided = nullableScalarFieldPresence(raw)
		nullableScalarProvided["StorageSlot"] = storageSlotProvided
	}
	c.Request.Body = io.NopCloser(bytes.NewReader(bodyBytes))

	var req CoinUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "Invalid request payload", err)
		return
	}

	updates, updateFields := req.toCoin(existing, storageLocationProvided, nullableScalarProvided)

	source := c.Query("source")
	if err := h.svc.UpdateCoinWithAssignmentFields(existing, &updates, updateFields, userID, source, storageLocationProvided, storageSlotProvided); err != nil {
		if handleCoinMutationError(c, err) {
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update coin"})
		return
	}

	c.JSON(http.StatusOK, existing)
}

// Duplicate creates a copy of an existing coin owned by the authenticated user.
//
//	@Summary		Duplicate a coin
//	@Description	Creates a copy of an existing coin owned by the authenticated user. Images and card-related media/state are not copied.
//	@Tags			Coins
//	@Produce		json
//	@Param			id	path		int	true	"Coin ID"
//	@Success		201	{object}	models.Coin
//	@Failure		400	{object}	ErrorResponse
//	@Failure		401	{object}	ErrorResponse
//	@Failure		404	{object}	ErrorResponse
//	@Failure		500	{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/coins/{id}/duplicate [post]
func (h *CoinHandler) Duplicate(c *gin.Context) {
	userID := c.GetUint("userId")
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid coin ID"})
		return
	}

	duplicate, err := h.svc.DuplicateCoin(uint(id), userID)
	if err != nil {
		if repository.IsRecordNotFound(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Coin not found"})
			return
		}
		h.logger.Error("coins", "Failed to duplicate coin %d for user %d: %v", id, userID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to duplicate coin"})
		return
	}

	c.JSON(http.StatusCreated, duplicate)
}

// Purchase marks a wishlist coin as purchased (moves it to the collection).
//
//	@Summary		Mark coin as purchased
//	@Description	Sets isWishlist to false, moving the coin from wishlist to collection. Optionally accepts purchase details.
//	@Tags			Coins
//	@Accept			json
//	@Produce		json
//	@Param			id		path		int						true	"Coin ID"
//	@Param			body	body		PurchaseRequest			false	"Optional purchase details"
//	@Success		200		{object}	models.Coin
//	@Failure		400		{object}	ErrorResponse
//	@Failure		401		{object}	ErrorResponse
//	@Failure		404		{object}	ErrorResponse
//	@Failure		500		{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/coins/{id}/purchase [post]
func (h *CoinHandler) Purchase(c *gin.Context) {
	userID := c.GetUint("userId")
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid coin ID"})
		return
	}

	coin, err := h.repo.FindByID(uint(id), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Coin not found"})
		return
	}

	if !coin.IsWishlist {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Coin is not on the wishlist"})
		return
	}

	// Apply optional purchase details from request body
	var req PurchaseRequest
	if c.Request.ContentLength > 0 {
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
			return
		}
	}
	if req.PurchasePrice != nil {
		coin.PurchasePrice = req.PurchasePrice
	}
	if req.PurchaseDate != "" {
		if t, err := time.Parse("2006-01-02", req.PurchaseDate); err == nil {
			coin.PurchaseDate = &t
		}
	}
	if req.PurchaseLocation != "" {
		coin.PurchaseLocation = req.PurchaseLocation
	}

	if err := h.svc.PurchaseCoin(coin, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to mark as purchased"})
		return
	}

	c.JSON(http.StatusOK, coin)
}

// Sell marks a collection coin as sold with a sale price.
//
//	@Summary		Mark coin as sold
//	@Description	Sets isSold to true with a sold price and date, moving the coin from collection to sold gallery.
//	@Tags			Coins
//	@Accept			json
//	@Produce		json
//	@Param			id		path		int						true	"Coin ID"
//	@Param			body	body		object{soldPrice float64}	true	"Sale details"
//	@Success		200		{object}	models.Coin
//	@Failure		400		{object}	ErrorResponse
//	@Failure		401		{object}	ErrorResponse
//	@Failure		404		{object}	ErrorResponse
//	@Failure		500		{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/coins/{id}/sell [post]
func (h *CoinHandler) Sell(c *gin.Context) {
	userID := c.GetUint("userId")
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid coin ID"})
		return
	}

	var body struct {
		SoldPrice *float64 `json:"soldPrice"`
		SoldTo    string   `json:"soldTo"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	coin, err := h.repo.FindByID(uint(id), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Coin not found"})
		return
	}

	now := time.Now()
	updates := map[string]interface{}{
		"is_sold":    true,
		"sold_price": body.SoldPrice,
		"sold_date":  &now,
		"sold_to":    body.SoldTo,
	}
	if err := h.svc.SellCoin(coin, updates, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to mark as sold"})
		return
	}

	c.JSON(http.StatusOK, coin)
}

// Delete removes a coin and its associated images.
//
//	@Summary		Delete a coin
//	@Description	Deletes a coin and all associated image records. Only the coin owner can delete it.
//	@Tags			Coins
//	@Produce		json
//	@Param			id	path		int	true	"Coin ID"
//	@Success		200	{object}	CoinDeletedResponse
//	@Failure		400	{object}	ErrorResponse
//	@Failure		401	{object}	ErrorResponse
//	@Failure		404	{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/coins/{id} [delete]
func (h *CoinHandler) Delete(c *gin.Context) {
	userID := c.GetUint("userId")
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid coin ID"})
		return
	}

	rows, err := h.svc.DeleteCoin(uint(id), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete coin"})
		return
	}
	if rows == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Coin not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Coin deleted"})
}

// Stats returns aggregate statistics for the authenticated user's collection.
//
//	@Summary		Get collection statistics
//	@Description	Returns counts by category/material, total/average values, and wishlist count.
//	@Tags			Coins
//	@Produce		json
//	@Success		200	{object}	StatsResponse
//	@Failure		401	{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/stats [get]
func (h *CoinHandler) Stats(c *gin.Context) {
	userID := c.GetUint("userId")

	stats, err := h.repo.GetStats(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch stats"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"totalCoins":    stats.TotalCoins,
		"totalWishlist": stats.TotalWishlist,
		"totalSold":     stats.TotalSold,
		"byCategory":    stats.ByCategory,
		"byMaterial":    stats.ByMaterial,
		"byGrade":       stats.ByGrade,
		"byEra":         stats.ByEra,
		"byRuler":       stats.ByRuler,
		"byPriceRange":  stats.ByPriceRange,
		"values":        stats.Values,
		"soldValues":    stats.SoldValues,
	})
}

// Distribution returns the era × category cross-tabulation for the collection heat map.
//
//	@Summary		Get collection distribution
//	@Description	Returns era × category counts for the heat map visualization.
//	@Tags			Coins
//	@Produce		json
//	@Success		200	{array}		repository.DistributionCell
//	@Failure		401	{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/stats/distribution [get]
func (h *CoinHandler) Distribution(c *gin.Context) {
	userID := c.GetUint("userId")
	cells, err := h.repo.GetDistribution(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch distribution"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"cells": cells})
}

// Suggestions returns distinct values for autocomplete fields.
//
//	@Summary		Get autocomplete suggestions
//	@Description	Returns distinct values for the specified field, optionally filtered by a search query. Limited to 20 results.
//	@Tags			Coins
//	@Produce		json
//	@Param			field	query		string	true	"Field to get suggestions for"	Enums(name, denomination, purchaseLocation)
//	@Param			q		query		string	false	"Search filter"
//	@Success		200		{array}		string
//	@Failure		400		{object}	ErrorResponse
//	@Failure		401		{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/suggestions [get]
func (h *CoinHandler) Suggestions(c *gin.Context) {
	userID := c.GetUint("userId")
	field := c.Query("field")
	q := c.Query("q")

	var column string
	switch field {
	case "name":
		column = "name"
	case "denomination":
		column = "denomination"
	case "ruler":
		column = "ruler"
	case "era":
		column = "era"
	case "purchaseLocation":
		column = "purchase_location"
	case "country":
		column = "country"
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid field"})
		return
	}

	values, err := h.repo.Suggestions(userID, column, q)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch suggestions"})
		return
	}
	c.JSON(http.StatusOK, values)
}

// CoinValueHistory returns value history entries for a specific coin.
//
//	@Summary		Get coin value history
//	@Description	Returns value tracking entries for a coin, ordered oldest first.
//	@Tags			Coins
//	@Produce		json
//	@Param			id	path		int	true	"Coin ID"
//	@Success		200	{array}		models.CoinValueHistory
//	@Failure		401	{object}	ErrorResponse
//	@Failure		404	{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/coins/{id}/value-history [get]
func (h *CoinHandler) CoinValueHistory(c *gin.Context) {
	userID := c.GetUint("userId")
	coinID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid coin ID"})
		return
	}

	exists, err := h.repo.CoinExists(uint(coinID), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify coin"})
		return
	}
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Coin not found"})
		return
	}

	entries, err := h.repo.GetCoinValueHistory(uint(coinID), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch value history"})
		return
	}

	c.JSON(http.StatusOK, entries)
}

func handleCoinMutationError(c *gin.Context, err error) bool {
	switch {
	case errors.Is(err, services.ErrReferenceCatalogRequired),
		errors.Is(err, services.ErrReferenceNumberRequired),
		errors.Is(err, services.ErrReferenceVolumeRequired),
		errors.Is(err, services.ErrReferenceUnknownCatalog),
		errors.Is(err, services.ErrReferenceDuplicate),
		errors.Is(err, services.ErrMintLocationNotFound),
		errors.Is(err, services.ErrCoinInvalidEra),
		errors.Is(err, services.ErrCoinInvalidCategory),
		errors.Is(err, services.ErrCoinInvalidPresentation),
		errors.Is(err, services.ErrCoinInvalidRarity),
		errors.Is(err, services.ErrCoinInvalidCirculation),
		errors.Is(err, services.ErrCoinInvalidKind),
		errors.Is(err, services.ErrCoinInvalidNoteVariant),
		errors.Is(err, services.ErrCoinInvalidQuantity):
		respondError(c, http.StatusBadRequest, err.Error(), err)
		return true
	case errors.Is(err, services.ErrStorageSlotRequired),
		errors.Is(err, services.ErrStorageSlotInvalid):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "message": err.Error(), "code": "validation_error", "field": "storageSlot"})
		return true
	case errors.Is(err, services.ErrStorageLocationNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "Storage location not found"})
		return true
	case errors.Is(err, services.ErrStorageSlotOccupied):
		c.JSON(http.StatusConflict, gin.H{"error": "Storage slot is occupied", "message": "Storage slot is occupied", "code": "slot_occupied"})
		return true
	case errors.Is(err, services.ErrWishlistReferenceDuplicate):
		c.JSON(http.StatusConflict, gin.H{
			"error":   "A wishlist coin already uses this source URL",
			"message": "A wishlist coin already uses this source URL",
			"code":    "wishlist_duplicate",
		})
		return true
	case isUniqueConstraintError(err):
		respondError(c, http.StatusBadRequest, services.ErrReferenceDuplicate.Error(), err)
		return true
	default:
		return false
	}
}

// intersectIDs keeps the IDs in both lists; a nil first list stands for every coin.
// coinKindQuery reads the optional kind filter: "" (every item), "coin" or
// "banknote"; ok is false for anything else.
func coinKindQuery(c *gin.Context) (kind string, ok bool) {
	kind = strings.TrimSpace(c.Query("kind"))
	return kind, kind == "" || models.ValidCoinKind(kind)
}

func intersectIDs(ids, keep []uint) []uint {
	if ids == nil {
		return keep
	}
	allowed := map[uint]bool{}
	for _, id := range keep {
		allowed[id] = true
	}
	both := []uint{}
	for _, id := range ids {
		if allowed[id] {
			both = append(both, id)
		}
	}
	return both
}
