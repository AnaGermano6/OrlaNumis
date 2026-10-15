package services

import (
	"context"
	"fmt"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/briandenicola/ancient-coins-api/models"
)

// New and upcoming coins, from the lists the app already reads for official
// photos: the ECB's €2 commemorative coins of the year and the collector
// coins the mints of Portugal (INCM), France (Monnaie de Paris) and Austria
// (Münze Österreich) sell.
const (
	CoinReleaseSourceECB = "ecb"
	// CoinReleaseSourceNumista is Numista's catalogue: Portugal's coins in escudos.
	CoinReleaseSourceNumista = "numista"
	// A coin is a commemorative (circulating) one, or a collector
	// (non-circulating) one.
	CoinReleaseKindCommemorative = "commemorative"
	CoinReleaseKindCollector     = "collector"
	// CoinReleaseGroupMints is the group of the mints' collector coins; each
	// release names its mint (models.CoinPhotoSource…).
	CoinReleaseGroupMints = "mints"
	// The lists change a few times a month; they are read again after an hour.
	coinReleasesLifetime = time.Hour
	// A mint's coins of the last twelve months, and those to come.
	mintReleaseWindowMonths = 12
	// The catalogue of every coin is read again after six hours.
	coinCatalogueLifetime = 6 * time.Hour
)

// MintReleaseSource is a mint's catalogue of the coins it sells.
type MintReleaseSource interface {
	Coins(ctx context.Context) ([]mintCoin, error)
	// Country is the ECB name of the mint's country; Source names the mint.
	Country() string
	Source() string
	Thumbnail(image string) string
}

// CoinRelease is a coin the ECB lists or a mint sells, issued lately or to come.
type CoinRelease struct {
	Source string `json:"source" enums:"ecb,incm,mdp,moe,numista"`
	Kind   string `json:"kind" enums:"commemorative,collector"`
	// Country is the issuing country's ISO code ("PT"); "" for a joint issue
	// of the euro area countries.
	Country string `json:"country"`
	// Title is what the coin shows: the ECB's feature (in English) or the
	// mint's name for the design.
	Title string `json:"title"`
	// Variants are the finishes and metals a mint sells the design in ("BNC", "Prata Proof").
	Variants []string `json:"variants,omitempty"`
	// FaceValue is the lowest of the design's face values; FaceValues all of
	// them, from the lowest, when there are several.
	FaceValue  float64   `json:"faceValue"`
	FaceValues []float64 `json:"faceValues,omitempty"`
	// Currency is the face value's, when not euros: "PTE" for escudos.
	Currency string `json:"currency,omitempty"`
	ImageURL string `json:"imageUrl"`
	PageURL  string `json:"pageUrl"`
	// From and To are the first and last month of the issue period ("2025-10"
	// and "2025-12" for "Fourth quarter of 2025"); the same month when it is
	// known; January to December when only the year is.
	From string `json:"from" example:"2025-10"`
	To   string `json:"to" example:"2025-12"`
	// IssueText is the period as the ECB writes it, in English.
	IssueText string `json:"issueText,omitempty"`
	// Volume is how many are struck, when the source says.
	Volume int64 `json:"volume,omitempty"`
	// Upcoming is true before the issue period starts, and during a month or
	// a quarter of issue until it ends.
	Upcoming bool `json:"upcoming"`
	// Owned is true when a coin of the collection is this one.
	Owned bool `json:"owned"`
	// Wishlist is true when a coin of the wishlist is this one.
	Wishlist bool `json:"wishlist"`
	// Keys identify the coin in the official photos found for a collection
	// coin: the ECB images of a €2 commemorative, the shop pages of a mint's design.
	Keys []string `json:"-"`
}

// CoinReleaseGroup is the ECB's releases, or the mints'.
type CoinReleaseGroup struct {
	Source string `json:"source" enums:"ecb,mints"`
	// Year is the year of the ECB's list (the latest it has published). The
	// mints' releases are those of the last twelve months (of this year for a
	// mint that gives only the year) and to come.
	Year     int           `json:"year"`
	Releases []CoinRelease `json:"releases"`
}

// CoinReleasesService lists new and upcoming coins.
type CoinReleasesService struct {
	ecb     *EcbCoinPhotoService
	mints   []MintReleaseSource
	numista *NumistaCatalogService
	now     func() time.Time

	mu       sync.Mutex
	groups   []CoinReleaseGroup
	loadedAt time.Time

	catalogueMu       sync.Mutex
	catalogue         []CoinRelease
	catalogueLoadedAt time.Time
	// numistaComplete is true once Numista has listed everything the
	// catalogue asks it for (Portugal's escudos, every country's €5, €10 and
	// €20) and described the coins a mint may name in its own language.
	numistaComplete bool

	// topicsRunning is true while topics are asked for in the background;
	// topics waits for them.
	topicsMu      sync.Mutex
	topicsRunning bool
	topics        sync.WaitGroup
}

// WithNumista adds Portugal's coins in escudos, from Numista's catalogue, to
// the catalogue.
func (s *CoinReleasesService) WithNumista(numista *NumistaCatalogService) *CoinReleasesService {
	s.numista = numista
	return s
}

// NewCoinReleasesService creates the service; ecb may be nil.
func NewCoinReleasesService(ecb *EcbCoinPhotoService, mints ...MintReleaseSource) *CoinReleasesService {
	return &CoinReleasesService{ecb: ecb, mints: mints, now: time.Now}
}

// Releases returns the ECB's latest €2 commemoratives and the mints'
// collector coins, those to come first (soonest first), then the newest. A
// source that cannot be read adds nothing.
func (s *CoinReleasesService) Releases(ctx context.Context) []CoinReleaseGroup {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.groups == nil || s.now().Sub(s.loadedAt) >= coinReleasesLifetime {
		s.groups, s.loadedAt = []CoinReleaseGroup{s.ecbReleases(ctx), s.mintReleases(ctx)}, s.now()
	}
	// Callers mark the coins they own: each gets its own copy.
	groups := make([]CoinReleaseGroup, len(s.groups))
	for i, group := range s.groups {
		groups[i] = group
		groups[i].Releases = append([]CoinRelease{}, group.Releases...)
	}
	return groups
}

// MarkOwned flags the releases a collection coin is: the ECB image of a €2
// commemorative, or the mint's page of a collector coin, that the official
// photo search finds for it.
func (s *CoinReleasesService) MarkOwned(ctx context.Context, groups []CoinReleaseGroup, coins []CoinPhotoQuery, photos *CoinPhotoSearchService) {
	owned := s.OwnedKeys(ctx, coins, photos)
	for g := range groups {
		MarkOwnedReleases(groups[g].Releases, owned)
	}
}

// MarkOwnedReleases flags the releases one of the owned keys (see OwnedKeys) identifies.
func MarkOwnedReleases(releases []CoinRelease, owned map[string]uint) {
	for r := range releases {
		releases[r].Owned = releases[r].Owned || identified(releases[r], owned)
	}
}

// MarkWishlistReleases flags the releases one of the wishlist's keys (see
// OwnedKeys, given the wishlist's coins) identifies.
func MarkWishlistReleases(releases []CoinRelease, wishlist map[string]uint) {
	for r := range releases {
		releases[r].Wishlist = releases[r].Wishlist || identified(releases[r], wishlist)
	}
}

func identified(release CoinRelease, keys map[string]uint) bool {
	for _, key := range release.Keys {
		if _, ok := keys[key]; ok {
			return true
		}
	}
	return false
}

// MarkWishlist flags the releases a coin of the wishlist is.
func (s *CoinReleasesService) MarkWishlist(ctx context.Context, groups []CoinReleaseGroup, coins []CoinPhotoQuery, photos *CoinPhotoSearchService) {
	wishlist := s.OwnedKeys(ctx, coins, photos)
	for g := range groups {
		MarkWishlistReleases(groups[g].Releases, wishlist)
	}
}

// OwnedKeys finds what identifies each collection coin among the releases:
// its Numista type, and the ECB image of a €2 commemorative or the mint's
// page of a collector coin that the official photo search (when given) finds
// for it. Each key maps to the coin (its CoinID).
func (s *CoinReleasesService) OwnedKeys(ctx context.Context, coins []CoinPhotoQuery, photos *CoinPhotoSearchService) map[string]uint {
	owned := map[string]uint{}
	mintCountries := map[string]bool{}
	for _, mint := range s.mints {
		mintCountries[mint.Country()] = true
	}
	for _, coin := range coins {
		if coin.NumistaTypeID > 0 {
			owned[numistaReleaseKey(coin.NumistaTypeID)] = coin.CoinID
		}
		// Only €2 coins and coins of a mint's country can be one of the releases.
		if photos == nil || !twoEuroDenomination.MatchString(coin.Denomination) && !mintCountries[EcbCountryName(coin.Country)] {
			continue
		}
		obverse, _ := photos.OfficialPhotos(ctx, coin)
		if obverse == nil {
			continue
		}
		switch obverse.Source {
		case models.CoinPhotoSourceECB:
			owned[obverse.ImageURL] = coin.CoinID
		case models.CoinPhotoSourceINCM, models.CoinPhotoSourceMDP, models.CoinPhotoSourceMOE:
			owned[obverse.PageURL] = coin.CoinID
		}
	}
	return owned
}

// Catalogue lists every €2 commemorative coin the ECB has published since
// 2004, each country's own (a joint issue once for each country), and every
// collector coin the mints sell, newest first. It is read at most every six
// hours; each call gets its own copy.
func (s *CoinReleasesService) Catalogue(ctx context.Context) []CoinRelease {
	s.catalogueMu.Lock()
	defer s.catalogueMu.Unlock()
	if s.catalogue == nil || s.now().Sub(s.catalogueLoadedAt) >= coinCatalogueLifetime {
		s.catalogue, s.catalogueLoadedAt = s.readCatalogue(ctx), s.now()
	}
	return append([]CoinRelease{}, s.catalogue...)
}

// CatalogueComplete tells whether the catalogue lists all its coins, once
// each: those from Numista (Portugal's escudos, the €5, €10 and €20 coins)
// are partly listed until Numista has answered every search, and some may
// be listed twice (by Numista and by a mint, in its language) until Numista
// has described them.
func (s *CoinReleasesService) CatalogueComplete() bool {
	if s.numista == nil {
		return true
	}
	s.catalogueMu.Lock()
	defer s.catalogueMu.Unlock()
	return s.numistaComplete
}

func (s *CoinReleasesService) readCatalogue(ctx context.Context) []CoinRelease {
	now := s.now()
	coins := []CoinRelease{}
	if s.ecb != nil {
		for year := ecbFirstCommemorativeYear; year <= now.Year(); year++ {
			photos, err := s.ecb.Commemoratives(ctx, year)
			if err != nil {
				continue
			}
			coins = append(coins, ecbCatalogue(photos, year, now)...)
		}
	}
	// A mint's €2 coin is a commemorative's collector version (BNC, proof):
	// the ECB's list has it already, when it has the country's year.
	listed := map[string]bool{}
	for _, coin := range coins {
		listed[coin.Country+coin.From[:4]] = true
	}
	for _, mint := range s.mints {
		mintCoins, err := mint.Coins(ctx)
		if err != nil {
			continue
		}
		for _, coin := range mintReleases(mintCoins, mint, now, false) {
			if coin.FaceValue == 2 && coin.FaceValues == nil && listed[coin.Country+coin.From[:4]] {
				continue
			}
			coins = append(coins, coin)
		}
	}
	escudos, complete := s.escudoCoins(ctx)
	coins = append(coins, escudos...)
	if complete {
		// Numista answers: ask it for the euro collector coins too.
		coins, complete = s.euroCollectorCoins(ctx, coins)
	}
	s.numistaComplete = complete
	sortReleases(coins)
	return coins
}

// The face values of the collector coins listed for every euro country.
var catalogueEuroValues = []int{5, 10, 20}

// catalogueLeftOutSeries are Numista series a country's catalogue does not
// list, by the start of their names: Portugal's square, coloured "Portuguese
// tiles" coins (one per century), which the owner does not collect.
var catalogueLeftOutSeries = map[string][]string{"PT": {"portuguese tiles"}}

// catalogueLeavesOut tells whether a coin is of a series the catalogue leaves out.
func catalogueLeavesOut(coin CoinRelease) bool {
	name := strings.ToLower(strings.TrimSpace(coin.Title))
	for _, series := range catalogueLeftOutSeries[coin.Country] {
		if strings.HasPrefix(name, series) {
			return true
		}
	}
	return false
}

// numistaEuroIssuers are Numista's codes of the euro countries, by ISO code.
var numistaEuroIssuers = map[string]string{
	"AD": "andorre", "AT": "autriche", "BE": "belgique", "BG": "bulgarie", "HR": "croatie", "CY": "chypre",
	"EE": "estonie", "FI": "finlande", "FR": "france", "DE": "germany", "GR": "grece", "IE": "irlande",
	"IT": "italie", "LV": "lettonie", "LT": "lituanie", "LU": "luxembourg", "MT": "malte", "MC": "monaco",
	"NL": "pays-bas", "PT": "portugal", "SM": "saint-marin", "SK": "slovaquie", "SI": "slovenie",
	"ES": "espagne", "VA": "vatican",
}

// euroCollectorCoins adds every euro country's €5, €10 and €20 coins from
// Numista to coins. A coin its mint's catalogue lists already (Portugal,
// France, Austria) is not listed again: the mint's, with its pictures, is
// known by its Numista type too. complete is false when Numista stopped
// answering.
func (s *CoinReleasesService) euroCollectorCoins(ctx context.Context, coins []CoinRelease) ([]CoinRelease, bool) {
	countries := make([]string, 0, len(numistaEuroIssuers))
	for country := range numistaEuroIssuers {
		countries = append(countries, country)
	}
	sort.Strings(countries)
	seen := map[int]bool{}
	var unmatched []numistaEuroCoin
	for _, country := range countries {
		for _, value := range catalogueEuroValues {
			types, err := s.numista.IssuerEuroTypes(ctx, numistaEuroIssuers[country], value, "en")
			if err != nil {
				return coins, false
			}
			for _, item := range types {
				coin, ok := euroCollectorRelease(item, country, value)
				if !ok || seen[item.ID] || catalogueLeavesOut(coin) {
					continue
				}
				seen[item.ID] = true
				if knownToMint(coins, coin) {
					continue
				}
				unmatched = append(unmatched, numistaEuroCoin{coin: coin, typeID: item.ID})
			}
		}
	}
	unmatched, complete := s.knownToMintByTopic(coins, unmatched)
	for _, numista := range unmatched {
		coins = append(coins, numista.coin)
	}
	return coins, complete
}

// numistaEuroCoin is a collector coin from Numista and its type.
type numistaEuroCoin struct {
	coin   CoinRelease
	typeID int
}

// mintTopicLangs are the languages of the mints that do not name their coins
// in English, as Numista's searches do: the INCM names "The knowledge" "O
// Conhecimento", Münze Österreich "Waterpower" "Wasserkraft". The Monnaie de
// Paris names its coins in English already.
var mintTopicLangs = map[string]string{models.CoinPhotoSourceINCM: "pt", models.CoinPhotoSourceMOE: "de"}

// topicLookup is a type whose topic the catalogue has not read yet, and the
// language to read it in.
type topicLookup struct {
	typeID int
	lang   string
}

// knownToMintByTopic leaves out the coins a mint lists under a name in its
// own language. What Numista says a coin commemorates, in that language, or
// its reverse's lettering, shares the words of the mint's name ("Heróis e
// criaturas da mitologia: Fénix" and "Fénix — Heróis e Criaturas da
// Mitologia"; "EINSTEINRING" and "Einsteinring"); or the coin is a
// version of a type the mint's coin is known by already (Numista's related
// types: "Freedom, freedom! Proof issue" of "Liberdade, Liberdade!"). Only a
// coin of a year and face value its mint has is looked up. Topics come from
// the cache; those not read yet are asked for in the background, and
// complete is false until they are.
func (s *CoinReleasesService) knownToMintByTopic(coins []CoinRelease, unmatched []numistaEuroCoin) ([]numistaEuroCoin, bool) {
	type lookedUp struct {
		numistaEuroCoin
		related []int
	}
	var kept []numistaEuroCoin
	var versions []lookedUp
	var missing []topicLookup
	for _, numista := range unmatched {
		lang := mintTopicLang(coins, numista.coin)
		if lang == "" {
			kept = append(kept, numista)
			continue
		}
		topic, ok := s.numista.CachedTypeTopic(numista.typeID, lang)
		if !ok {
			missing = append(missing, topicLookup{numista.typeID, lang})
			kept = append(kept, numista)
			continue
		}
		if knownToMintByName(coins, numista.coin, topic.Topic, topic.ReverseLettering) {
			continue
		}
		versions = append(versions, lookedUp{numista, topic.Related})
	}
	// A version is known once the type it relates to is; that one may be a
	// version found here too.
	for found := true; found; {
		found = false
		left := versions[:0]
		for _, version := range versions {
			if mint := mintKnownByType(coins, version.coin, version.related); mint != nil {
				mint.Keys = append(mint.Keys, version.coin.Keys...)
				found = true
				continue
			}
			left = append(left, version)
		}
		versions = left
	}
	for _, version := range versions {
		kept = append(kept, version.numistaEuroCoin)
	}
	if len(missing) > 0 {
		s.lookUpTopics(missing)
		return kept, false
	}
	return kept, true
}

// lookUpTopics asks Numista, in the background, for the topics the catalogue
// has not read yet: a few hundred types the first time, too many to wait
// for. It stops at the first refusal (out of quota) and tries again at the
// next reading; the catalogue is read again once some topics are known.
func (s *CoinReleasesService) lookUpTopics(missing []topicLookup) {
	s.topicsMu.Lock()
	defer s.topicsMu.Unlock()
	if s.topicsRunning {
		return
	}
	s.topicsRunning = true
	s.topics.Add(1)
	go func() {
		defer s.topics.Done()
		found := 0
		for _, lookup := range missing {
			if _, err := s.numista.TypeTopic(context.Background(), lookup.typeID, lookup.lang); err != nil {
				break
			}
			found++
		}
		s.topicsMu.Lock()
		s.topicsRunning = false
		s.topicsMu.Unlock()
		if found > 0 {
			s.catalogueMu.Lock()
			s.catalogue = nil
			s.catalogueMu.Unlock()
		}
	}()
}

// knownToMintByName tells whether the mint lists the coin under one of the
// names Numista gives it in the mint's language: what it commemorates, or the
// lettering of its reverse ("EINSTEINRING"). Its obverse often lists every
// coin of a series (the other signs of the zodiac), and is left out.
func knownToMintByName(coins []CoinRelease, coin CoinRelease, names ...string) bool {
	for _, name := range names {
		if name == "" {
			continue
		}
		named := coin
		named.Title = name
		if knownToMintBy(coins, named, topicWords, topicSimilarWords) {
			return true
		}
	}
	return false
}

// mintTopicLang is the language to read a coin's topic in: its mint's, when
// the mint has a coin of the same country, year and face value and does not
// name its coins in English; "" otherwise.
func mintTopicLang(coins []CoinRelease, coin CoinRelease) string {
	for _, mint := range coins {
		if lang := mintTopicLangs[mint.Source]; lang != "" && sameMintIssue(mint, coin) {
			return lang
		}
	}
	return ""
}

// mintKnownByType returns the mint's coin of the same issue known by one of
// the types, or nil.
func mintKnownByType(coins []CoinRelease, coin CoinRelease, typeIDs []int) *CoinRelease {
	for i := range coins {
		mint := &coins[i]
		if !sameMintIssue(*mint, coin) {
			continue
		}
		for _, typeID := range typeIDs {
			if slices.Contains(mint.Keys, numistaReleaseKey(typeID)) {
				return mint
			}
		}
	}
	return nil
}

// sameMintIssue tells whether mint is a mint's coin of the coin's country,
// year and face value.
func sameMintIssue(mint, coin CoinRelease) bool {
	return mintReleaseSources[mint.Source] && mint.Country == coin.Country && mint.From[:4] == coin.From[:4] && releaseHasFaceValue(mint, coin.FaceValue)
}

// knownToMint tells whether the country's mint lists the coin already: a
// coin of its mint of the same year and face value whose name shares most
// of the coin's distinctive words ("Notre-Dame de Paris" and "French
// Excellence - Notre-Dame de Paris"; "10 euros Asterix aux jeux Olympiques"
// and "Asterix and the Olympic Games"). The mint's coin is then known by the
// coin's Numista type as well.
func knownToMint(coins []CoinRelease, coin CoinRelease) bool {
	return knownToMintBy(coins, coin, func(name string) []string { return uniqueStrings(mintWords(name)) }, mintSimilarWords)
}

// knownToMintBy is knownToMint with the words each name is read as, and
// when two words are alike.
func knownToMintBy(coins []CoinRelease, coin CoinRelease, nameWords func(string) []string, similar func(a, b string) bool) bool {
	words := nameWords(coin.Title)
	if len(words) == 0 {
		return false
	}
	for i := range coins {
		mint := &coins[i]
		if !sameMintIssue(*mint, coin) {
			continue
		}
		mintNameWords := nameWords(mint.Title)
		shorter := min(len(words), len(mintNameWords))
		if shorter > 0 && wordOverlap(words, mintNameWords, similar) >= (shorter+1)/2 {
			mint.Keys = append(mint.Keys, coin.Keys...)
			return true
		}
	}
	return false
}

// wordOverlap counts the words of other found among words.
func wordOverlap(words, other []string, similar func(a, b string) bool) int {
	count := 0
	for _, word := range other {
		for _, candidate := range words {
			if similar(word, candidate) {
				count++
				break
			}
		}
	}
	return count
}

// topicSimilarWords allows, besides small spelling differences, for German
// compounds: "Salzbergbau" is "Bergbau", "Bienentanz" "Tanz der Bienen".
func topicSimilarWords(a, b string) bool {
	if mintSimilarWords(a, b) {
		return true
	}
	if len(a) < len(b) {
		a, b = b, a
	}
	return len(b) >= 5 && (strings.HasPrefix(a, b) || strings.HasSuffix(a, b))
}

// topicWords are the distinctive words of a name without its years: the
// coins compared are of the same year, and a topic says it in words ("50
// anos de eleições livres" for "50 anos de Eleições em Liberdade –
// 1975.2025"). Names, unlike topics, keep them: "FIFA World Cup 2026"
// against "Coupe du Monde FIFA 2026".
func topicWords(name string) []string {
	var words []string
	for _, word := range uniqueStrings(mintWords(name)) {
		if year, err := strconv.Atoi(word); err == nil && year >= 1900 && year <= 2100 {
			continue
		}
		words = append(words, word)
	}
	return words
}

// mintReleaseSources are the mints whose catalogues the releases come from.
var mintReleaseSources = map[string]bool{models.CoinPhotoSourceINCM: true, models.CoinPhotoSourceMDP: true, models.CoinPhotoSourceMOE: true}

func releaseHasFaceValue(release CoinRelease, value float64) bool {
	if release.FaceValue == value {
		return true
	}
	for _, other := range release.FaceValues {
		if other == value {
			return true
		}
	}
	return false
}

var (
	collectorEuroValue = regexp.MustCompile(`(?i)^\s*(\d+)\s+euros?\b`)
	euroCents          = regexp.MustCompile(`(?i)\bcents?\b|\bcentimes?\b`)
)

// euroCollectorRelease describes a Numista type of a euro country as one of
// its collector coins, when it is one of the face value searched for ("10
// Euros (Albert Einstein)"), not a cent coin, a pattern or a coin of
// before the euro.
func euroCollectorRelease(item models.NumistaCatalogSearchItem, country string, value int) (CoinRelease, bool) {
	denomination := collectorEuroValue.FindStringSubmatch(item.Title)
	if denomination == nil || denomination[1] != strconv.Itoa(value) || euroCents.MatchString(item.Title) ||
		escudoPattern.MatchString(item.Title) || item.MinYear == nil || *item.MinYear < 1999 {
		return CoinRelease{}, false
	}
	name := item.Title
	if topic := escudoTopic.FindStringSubmatch(item.Title); topic != nil {
		name = strings.TrimSpace(topic[1])
	}
	return CoinRelease{
		Source:    CoinReleaseSourceNumista,
		Kind:      CoinReleaseKindCollector,
		Country:   country,
		Title:     name,
		FaceValue: float64(value),
		PageURL:   fmt.Sprintf("https://pt.numista.com/catalogue/pieces%d.html", item.ID),
		From:      fmt.Sprintf("%04d-01", *item.MinYear),
		To:        fmt.Sprintf("%04d-12", *item.MinYear),
		Keys:      []string{numistaReleaseKey(item.ID)},
	}, true
}

// Portugal's coins in escudos: from the Republic's first coins to the euro.
const (
	escudoFirstYear = 1911
	escudoLastYear  = 2001
)

// escudoCoins lists Portugal's commemorative and collector coins in escudos:
// those Numista lists year by year, and those whose details the app has
// cached (the collection's own among them). complete is false while Numista
// has not answered for every year (its monthly quota used up).
func (s *CoinReleasesService) escudoCoins(ctx context.Context) (coins []CoinRelease, complete bool) {
	if s.numista == nil {
		return nil, false
	}
	byID := map[int]CoinRelease{}
	complete = true
	for year := escudoFirstYear; year <= escudoLastYear; year++ {
		types, err := s.numista.IssuerYearTypes(ctx, "portugal", year, "pt")
		if err != nil {
			// Try again at the next reading, not for every year now.
			complete = false
			break
		}
		for _, item := range types {
			if _, seen := byID[item.ID]; seen {
				continue
			}
			facts, _ := s.numista.CachedTypeFacts(item.ID)
			if coin, ok := escudoRelease(item.ID, item.Title, item.MinYear, item.MaxYear, facts.Classification); ok {
				byID[item.ID] = coin
			}
		}
	}
	for _, detail := range s.numista.CachedIssuerTypes("portugal") {
		if _, seen := byID[detail.ID]; seen {
			continue
		}
		if coin, ok := escudoRelease(detail.ID, detail.Title, detail.MinYear, detail.MaxYear, detail.Classification); ok {
			byID[detail.ID] = coin
		}
	}
	for _, coin := range byID {
		coins = append(coins, coin)
	}
	return coins, complete
}

var (
	escudoDenomination = regexp.MustCompile(`(?i)^\s*(\d+(?:[.,]\d+)?|\d*½)\s*escudos?\b`)
	escudoTopic        = regexp.MustCompile(`\(([^)]+)\)\s*$`)
	// Patterns and trial strikes are not coins one can collect as issued.
	escudoPattern = regexp.MustCompile(`(?i)\b(pattern|ensaio|prova|essai|trial)\b`)
	// A version in a precious metal, or struck as a proof, is a collector coin.
	escudoPrecious     = regexp.MustCompile(`(?i)silver|prata|gold|ouro|proof|piedfort|platin`)
	classCollector     = regexp.MustCompile(`(?i)n[ãa]o circulante|non.?circulating`)
	classCommemorative = regexp.MustCompile(`(?i)comemorativ|commemorativ`)
)

// escudoRelease describes a Portuguese type in escudos as a coin of the
// catalogue, when it is a commemorative or collector coin. Numista's
// classification, when known, says which; otherwise a type of a single year
// whose title names what it marks ("200 Escudos (Colombo e Portugal)") is
// one, a collector coin when struck in a precious metal or as a proof.
func escudoRelease(id int, title string, minYear, maxYear *int, classification string) (CoinRelease, bool) {
	denomination := escudoDenomination.FindStringSubmatch(title)
	if denomination == nil || escudoPattern.MatchString(title) || minYear == nil || *minYear > escudoLastYear {
		return CoinRelease{}, false
	}
	topic := escudoTopic.FindStringSubmatch(title)
	kind := ""
	switch {
	case classCollector.MatchString(classification):
		kind = CoinReleaseKindCollector
	case classCommemorative.MatchString(classification):
		kind = CoinReleaseKindCommemorative
	case classification != "":
		// A regular circulation coin.
		return CoinRelease{}, false
	case topic == nil || maxYear == nil || *maxYear != *minYear:
		return CoinRelease{}, false
	case escudoPrecious.MatchString(title):
		kind = CoinReleaseKindCollector
	default:
		kind = CoinReleaseKindCommemorative
	}
	value := strings.ReplaceAll(denomination[1], ",", ".")
	half := strings.HasSuffix(value, "½")
	value = strings.TrimSuffix(value, "½")
	faceValue, _ := strconv.ParseFloat(value, 64)
	if half {
		faceValue += 0.5
	}
	name := title
	if topic != nil {
		name = strings.TrimSpace(topic[1])
	}
	return CoinRelease{
		Source:    CoinReleaseSourceNumista,
		Kind:      kind,
		Country:   "PT",
		Title:     name,
		FaceValue: faceValue,
		Currency:  "PTE",
		PageURL:   fmt.Sprintf("https://pt.numista.com/catalogue/pieces%d.html", id),
		From:      fmt.Sprintf("%04d-01", *minYear),
		To:        fmt.Sprintf("%04d-12", *minYear),
		Keys:      []string{numistaReleaseKey(id)},
	}, true
}

// numistaReleaseKey identifies a coin of the catalogue by its Numista type.
func numistaReleaseKey(typeID int) string {
	return fmt.Sprintf("numista:%d", typeID)
}

// ecbCatalogue lists a year's €2 commemoratives as each country issues them:
// a joint issue once for every country whose image the ECB shows.
func ecbCatalogue(photos []EcbCoinPhoto, year int, now time.Time) []CoinRelease {
	page := ecbBaseURL + fmt.Sprintf(ecbCommemorativePagePath, year)
	current := now.Format("2006-01")
	coins := []CoinRelease{}
	for _, photo := range withJointIssueCountries(photos, year) {
		country := ecbCountryCodes[photo.Country]
		if country == "" {
			continue
		}
		from, to := releasePeriod(photo.IssueDate, year)
		coins = append(coins, CoinRelease{
			Source:    CoinReleaseSourceECB,
			Kind:      CoinReleaseKindCommemorative,
			Country:   country,
			Title:     photo.Feature,
			FaceValue: 2,
			ImageURL:  photo.ImageURL,
			PageURL:   page,
			From:      from,
			To:        to,
			IssueText: photo.IssueDate,
			Volume:    photo.Volume,
			Upcoming:  releaseUpcoming(from, to, current),
			Keys:      []string{photo.ImageURL},
		})
	}
	return coins
}

// jointIssueMinCountries: a joint issue of the euro area is one design many
// countries issue together; a theme two or four countries share (Louis
// Braille, the Élysée Treaty, human rights) is not one.
const jointIssueMinCountries = 5

// euroMicrostates issue the euro by agreement, outside the European Union.
var euroMicrostates = map[string]bool{"Andorra": true, "Monaco": true, "San Marino": true, "Vatican City": true}

// JointIssue is a €2 coin the euro area's countries issued together (the
// Treaty of Rome in 2007, the EU flag in 2015): one design, each country's
// own version.
type JointIssue struct {
	Year int `json:"year" example:"2007"`
	// Title is the ECB's feature, in English.
	Title   string           `json:"title" example:"50th anniversary of the Treaty of Rome"`
	PageURL string           `json:"pageUrl"`
	Coins   []JointIssueCoin `json:"coins"`
}

// JointIssueCoin is a country's version of a joint issue.
type JointIssueCoin struct {
	// Country is an ISO code.
	Country  string `json:"country" example:"PT"`
	ImageURL string `json:"imageUrl"`
}

// jointIssueGroups groups a year's coins by feature, keeping the groups of a
// joint issue.
func jointIssueGroups(photos []EcbCoinPhoto) [][]int {
	byFeature := map[string][]int{}
	order := []string{}
	for i, photo := range photos {
		if _, seen := byFeature[photo.Feature]; !seen {
			order = append(order, photo.Feature)
		}
		byFeature[photo.Feature] = append(byFeature[photo.Feature], i)
	}
	groups := [][]int{}
	for _, feature := range order {
		if len(byFeature[feature]) >= jointIssueMinCountries {
			groups = append(groups, byFeature[feature])
		}
	}
	return groups
}

// withJointIssueCountries names the country of the joint-issue images the
// page did not: from the file name when it says it in another language, or,
// when one image is left and one member state of that year is missing, that
// member state (2022's "EU1-erasmus" is Germany's).
func withJointIssueCountries(photos []EcbCoinPhoto, year int) []EcbCoinPhoto {
	resolved := append([]EcbCoinPhoto{}, photos...)
	for _, group := range jointIssueGroups(resolved) {
		known := map[string]bool{}
		unknown := []int{}
		for _, i := range group {
			if resolved[i].Country == "" {
				resolved[i].Country = ecbCountryFromFile(resolved[i].ImageURL)
			}
			if resolved[i].Country == "" {
				unknown = append(unknown, i)
			}
			known[resolved[i].Country] = true
		}
		if len(unknown) != 1 {
			continue
		}
		missing := []string{}
		for country, first := range euroCoinFirstYear {
			if !euroMicrostates[country] && first <= year && !known[country] {
				missing = append(missing, country)
			}
		}
		if len(missing) == 1 {
			resolved[unknown[0]].Country = missing[0]
		}
	}
	return resolved
}

// JointIssue is the euro area's joint issue of a year; nil when it had none.
func (s *CoinReleasesService) JointIssue(ctx context.Context, year int) *JointIssue {
	if s.ecb == nil {
		return nil
	}
	photos, err := s.ecb.Commemoratives(ctx, year)
	if err != nil {
		return nil
	}
	photos = withJointIssueCountries(photos, year)
	var issue *JointIssue
	for _, group := range jointIssueGroups(photos) {
		candidate := &JointIssue{Year: year, Title: photos[group[0]].Feature, PageURL: ecbBaseURL + fmt.Sprintf(ecbCommemorativePagePath, year), Coins: []JointIssueCoin{}}
		seen := map[string]bool{}
		for _, i := range group {
			code := ecbCountryCodes[photos[i].Country]
			if code == "" || seen[code] {
				continue
			}
			seen[code] = true
			candidate.Coins = append(candidate.Coins, JointIssueCoin{Country: code, ImageURL: photos[i].ImageURL})
		}
		sort.Slice(candidate.Coins, func(a, b int) bool { return candidate.Coins[a].Country < candidate.Coins[b].Country })
		if issue == nil || len(candidate.Coins) > len(issue.Coins) {
			issue = candidate
		}
	}
	return issue
}

// JointIssues lists the euro area's joint issues, newest first.
func (s *CoinReleasesService) JointIssues(ctx context.Context) []JointIssue {
	issues := []JointIssue{}
	if s.ecb == nil {
		return issues
	}
	for year := s.now().Year(); year >= ecbFirstCommemorativeYear; year-- {
		if issue := s.JointIssue(ctx, year); issue != nil {
			issues = append(issues, *issue)
		}
	}
	return issues
}

// CatalogueCountryName is the ECB name of a country of the catalogue (an ISO
// code, "PT"); "" for another.
func CatalogueCountryName(code string) string {
	for name, iso := range ecbCountryCodes {
		if strings.EqualFold(iso, code) {
			return name
		}
	}
	return ""
}

// CatalogueCountryCode is the ISO code of a country written in English or
// Portuguese, when the catalogue has it ("Alemanha" → "DE"); "" otherwise.
func CatalogueCountryCode(name string) string {
	return ecbCountryCodes[EcbCountryName(name)]
}

func (s *CoinReleasesService) ecbReleases(ctx context.Context) CoinReleaseGroup {
	group := CoinReleaseGroup{Source: CoinReleaseSourceECB, Releases: []CoinRelease{}}
	if s.ecb == nil {
		return group
	}
	now := s.now()
	// The ECB opens the page of a year some time into it.
	for year := now.Year(); year >= now.Year()-1; year-- {
		photos, err := s.ecb.Commemoratives(ctx, year)
		if err != nil || len(photos) == 0 {
			continue
		}
		group.Year = year
		group.Releases = ecbReleases(photos, year, now)
		break
	}
	return group
}

// ecbJointIssueCountries: a feature shared by more countries than this is a
// joint issue of the euro area, listed once.
const ecbJointIssueCountries = 3

func ecbReleases(photos []EcbCoinPhoto, year int, now time.Time) []CoinRelease {
	countries := map[string]int{}
	for _, photo := range photos {
		countries[photo.Feature]++
	}
	page := ecbBaseURL + fmt.Sprintf(ecbCommemorativePagePath, year)
	current := now.Format("2006-01")
	releases := []CoinRelease{}
	joint := map[string]int{}
	for _, photo := range photos {
		if countries[photo.Feature] > ecbJointIssueCountries {
			if at, seen := joint[photo.Feature]; seen {
				releases[at].Keys = append(releases[at].Keys, photo.ImageURL)
				continue
			}
			joint[photo.Feature] = len(releases)
			photo.Country = ""
		}
		from, to := releasePeriod(photo.IssueDate, year)
		releases = append(releases, CoinRelease{
			Source:    CoinReleaseSourceECB,
			Kind:      CoinReleaseKindCommemorative,
			Country:   ecbCountryCodes[photo.Country],
			Title:     photo.Feature,
			FaceValue: 2,
			ImageURL:  photo.ImageURL,
			PageURL:   page,
			From:      from,
			To:        to,
			IssueText: photo.IssueDate,
			Volume:    photo.Volume,
			Upcoming:  releaseUpcoming(from, to, current),
			Keys:      []string{photo.ImageURL},
		})
	}
	sortReleases(releases)
	return releases
}

func (s *CoinReleasesService) mintReleases(ctx context.Context) CoinReleaseGroup {
	now := s.now()
	group := CoinReleaseGroup{Source: CoinReleaseGroupMints, Year: now.Year(), Releases: []CoinRelease{}}
	for _, mint := range s.mints {
		coins, err := mint.Coins(ctx)
		if err != nil {
			continue
		}
		group.Releases = append(group.Releases, mintReleases(coins, mint, now, true)...)
	}
	sortReleases(group.Releases)
	return group
}

var mintVariant = regexp.MustCompile(`\(([^)]*)\)\s*$`)

// mintReleases groups a design's finishes, metals and face values ("Rotary",
// "Rotary (BNC)", "Rotary (Proof)"; the Year of the Goat in gold and silver)
// into one release, of the month its first one went on sale. Recent only
// lists the coins of the last twelve months and to come; a coin the mint
// dates by its year only, from that year on.
func mintReleases(coins []mintCoin, mint MintReleaseSource, now time.Time, recent bool) []CoinRelease {
	since := now.AddDate(0, -mintReleaseWindowMonths, 0).Format("2006-01")
	current := now.Format("2006-01")
	releases := []CoinRelease{}
	byDesign := map[string]int{}
	for _, coin := range coins {
		from, to := coin.Released, coin.Released
		if from == "" {
			if recent && coin.Year < now.Year() {
				continue
			}
			from, to = fmt.Sprintf("%04d-01", coin.Year), fmt.Sprintf("%04d-12", coin.Year)
		}
		if recent && to < since {
			continue
		}
		title, variant := coin.Title, ""
		if match := mintVariant.FindStringSubmatchIndex(title); match != nil {
			variant = strings.TrimSpace(title[match[2]:match[3]])
			title = strings.TrimSpace(title[:match[0]])
		}
		// The design by its name and year: a mint can file its gold and silver
		// versions apart, under the same name.
		design := mintDesignKey(title)
		if design == "" {
			design = coin.DesignKey
		}
		key := fmt.Sprintf("%s|%d", design, coin.Year)
		at, seen := byDesign[key]
		if !seen {
			at = len(releases)
			byDesign[key] = at
			releases = append(releases, CoinRelease{
				Source:    mint.Source(),
				Kind:      CoinReleaseKindCollector,
				Country:   ecbCountryCodes[mint.Country()],
				Title:     title,
				FaceValue: coin.Value,
				ImageURL:  mint.Thumbnail(coin.Obverse),
				PageURL:   coin.PageURL,
				From:      from,
				To:        to,
			})
		}
		release := &releases[at]
		if variant != "" {
			release.Variants = append(release.Variants, variant)
		} else {
			// The plain edition is the one to show and open.
			release.Title, release.ImageURL, release.PageURL = title, mint.Thumbnail(coin.Obverse), coin.PageURL
		}
		if from < release.From {
			release.From, release.To = from, to
		}
		release.Volume += coin.Volume
		release.Keys = append(release.Keys, coin.PageURL)
		release.FaceValues = append(release.FaceValues, coin.Value)
	}
	for i := range releases {
		releases[i].Upcoming = releaseUpcoming(releases[i].From, releases[i].To, current)
		releases[i].Variants = uniqueStrings(releases[i].Variants)
		releases[i].FaceValues = sortedFaceValues(releases[i].FaceValues)
		releases[i].FaceValue = releases[i].FaceValues[0]
		if len(releases[i].FaceValues) == 1 {
			releases[i].FaceValues = nil
		}
	}
	sortReleases(releases)
	return releases
}

// sortedFaceValues returns the face values once each, from the lowest.
func sortedFaceValues(values []float64) []float64 {
	sort.Float64s(values)
	unique := values[:0]
	for i, value := range values {
		if i == 0 || value != values[i-1] {
			unique = append(unique, value)
		}
	}
	return unique
}

// releaseUpcoming tells whether a coin is still to come in the month current
// (all "YYYY-MM"): its period has not started, or, a month or a quarter long,
// has not ended. A coin known by a longer period, such as its year, is out
// once that period starts.
func releaseUpcoming(from, to, current string) bool {
	if current < from {
		return true
	}
	return monthsBetween(from, to) < 3 && current < to
}

// monthsBetween counts the months from one "YYYY-MM" to another.
func monthsBetween(from, to string) int {
	month := func(value string) int {
		year, _ := strconv.Atoi(value[:min(4, len(value))])
		number := 0
		if len(value) >= 7 {
			number, _ = strconv.Atoi(value[5:7])
		}
		return year*12 + number
	}
	return month(to) - month(from)
}

// sortReleases puts the coins to come first, soonest first, then the newest,
// a known month before a mere year, and those of one period by title.
func sortReleases(releases []CoinRelease) {
	sort.SliceStable(releases, func(i, j int) bool {
		a, b := releases[i], releases[j]
		switch {
		case a.Upcoming != b.Upcoming:
			return a.Upcoming
		case a.From != b.From:
			if a.Upcoming {
				return a.From < b.From
			}
			return a.From > b.From
		case a.To != b.To:
			return a.To < b.To
		}
		// The same period: by title, a design's versions together.
		return strings.ToLower(a.Title) < strings.ToLower(b.Title)
	})
}

var (
	releaseYear   = regexp.MustCompile(`\b(20\d{2})\b`)
	releaseMonths = map[string]int{
		"january": 1, "february": 2, "march": 3, "april": 4, "may": 5, "june": 6,
		"july": 7, "august": 8, "september": 9, "october": 10, "november": 11, "december": 12,
	}
	releaseOrdinals = map[string]int{"first": 1, "second": 2, "third": 3, "fourth": 4}
	// Seasons, as the ECB writes them for the northern hemisphere.
	releaseSeasons = map[string][2]int{"spring": {3, 5}, "summer": {6, 8}, "autumn": {9, 11}, "fall": {9, 11}}
)

// releasePeriod reads when the ECB says a coin is issued ("June 2025", "1 July
// 2022", "Fourth quarter of 2025", "Second half of 2025", "Autumn 2025") as
// the first and last month of that period; the whole year when it says no more.
func releasePeriod(text string, year int) (from, to string) {
	lower := strings.ToLower(text)
	if match := releaseYear.FindString(lower); match != "" {
		year, _ = strconv.Atoi(match)
	}
	first, last := 1, 12
	words := strings.FieldsFunc(lower, func(r rune) bool { return !unicode.IsLetter(r) })
words:
	for i, word := range words {
		if month, ok := releaseMonths[word]; ok {
			first, last = month, month
			break
		}
		if season, ok := releaseSeasons[word]; ok {
			first, last = season[0], season[1]
			break
		}
		if i == 0 {
			continue
		}
		switch n := releaseOrdinals[words[i-1]]; {
		case word == "quarter" && n > 0:
			first, last = 3*n-2, 3*n
			break words
		case word == "half" && (n == 1 || n == 2):
			first, last = 6*n-5, 6*n
			break words
		}
	}
	return fmt.Sprintf("%04d-%02d", year, first), fmt.Sprintf("%04d-%02d", year, last)
}

// ecbCountryCodes maps the ECB's country names to ISO codes.
var ecbCountryCodes = map[string]string{
	"Andorra": "AD", "Austria": "AT", "Belgium": "BE", "Bulgaria": "BG", "Croatia": "HR", "Cyprus": "CY",
	"Estonia": "EE", "Finland": "FI", "France": "FR", "Germany": "DE", "Greece": "GR", "Ireland": "IE",
	"Italy": "IT", "Latvia": "LV", "Lithuania": "LT", "Luxembourg": "LU", "Malta": "MT", "Monaco": "MC",
	"Netherlands": "NL", "Portugal": "PT", "San Marino": "SM", "Slovakia": "SK", "Slovenia": "SI",
	"Spain": "ES", "Vatican City": "VA",
}
