package services

import (
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/briandenicola/ancient-coins-api/models"
	"github.com/briandenicola/ancient-coins-api/repository"
)

// portugalPeriods are Portugal's periods as uCoin divides them, most recent
// first: a coin belongs to the one in which its type began (the first year
// of its Numista type, else its own year). A boundary year shared by two
// periods (1969, 1986) goes to the later one, whose new coins began then;
// 1910, 1706 and 1495 stay with the earlier one, as the new reign or
// republic struck its coins only later.
var portugalPeriods = []struct {
	name     string
	from, to int
}{
	{"República Portuguesa 1986–2001", 1986, 2001},
	{"República Portuguesa 1969–1986", 1969, 1985},
	{"República Portuguesa 1910–1969", 1911, 1968},
	{"Reino de Portugal 1836–1910", 1836, 1910},
	{"Reino de Portugal 1777–1835", 1777, 1835},
	{"Reino de Portugal 1706–1776", 1707, 1776},
	{"Reino de Portugal 1640–1706", 1640, 1706},
	{"Reino de Portugal 1495–1639", 1496, 1639},
	{"Reino de Portugal 1415–1495", 1415, 1495},
}

const (
	portugalEuroPeriod = "União Europeia (Euro)"
	portugalECUPeriod  = "ECU"
)

// PortugalCoinPeriod is the period of a Portuguese coin whose type began in
// startYear, as uCoin divides Portugal's coins; "" for a year before 1415.
// Euro and ECU coins have periods of their own.
func PortugalCoinPeriod(startYear int, monetarySystem string) string {
	system := strings.ToLower(monetarySystem)
	switch {
	case strings.Contains(system, "ecu"):
		return portugalECUPeriod
	case strings.Contains(system, "euro") || startYear >= 2002:
		return portugalEuroPeriod
	}
	for _, period := range portugalPeriods {
		if startYear >= period.from && startYear <= period.to {
			return period.name
		}
	}
	return ""
}

// ucoinNotePeriod is the period the uCoin import wrote in a coin's notes
// ("Período: República Portuguesa"), or "".
var ucoinNotePeriod = regexp.MustCompile(`(?m)^(?:Período|Period):\s*(.+?)\s*$`)

func periodFromNotes(notes string) string {
	if match := ucoinNotePeriod.FindStringSubmatch(notes); match != nil {
		return match[1]
	}
	return ""
}

// isPortugal tells whether a coin's country is Portugal, in the app's names.
func isPortugal(country string) bool {
	return strings.EqualFold(strings.TrimSpace(country), "Portugal")
}

// CoinPeriodService works out each coin's period, so the gallery can group
// coins by period with nothing to type: a Portuguese coin's from the year its
// Numista type began (its own year without a cached type), another
// country's from the uCoin import's notes. Banknotes have none.
type CoinPeriodService struct {
	coins   *repository.CoinRepository
	catalog *NumistaCatalogService
	logger  *Logger
}

// NewCoinPeriodService creates the service; catalog and logger may be nil.
func NewCoinPeriodService(coins *repository.CoinRepository, catalog *NumistaCatalogService, logger *Logger) *CoinPeriodService {
	return &CoinPeriodService{coins: coins, catalog: catalog, logger: logger}
}

// PeriodFor is the period of a coin as it stands, its references included.
func (s *CoinPeriodService) PeriodFor(coin *models.Coin) string {
	if coin.Kind == models.CoinKindBanknote {
		return ""
	}
	if !isPortugal(coin.Country) {
		return periodFromNotes(coin.Notes)
	}
	system := coin.MonetarySystem
	start := coin.Year
	if typeID := NumistaTypeIDForCoin(coin); typeID != 0 && s.catalog != nil {
		if detail := s.catalog.cachedTypeDetail(typeID); detail != nil {
			if detail.MinYear != nil {
				start = detail.MinYear
			}
			if system == "" {
				system = detail.MonetarySystem
			}
		}
	}
	if start == nil {
		return ""
	}
	return PortugalCoinPeriod(*start, system)
}

// Refresh works out a saved coin's period again and stores it when it changed.
func (s *CoinPeriodService) Refresh(coinID, userID uint) {
	coin, err := s.coins.FindByID(coinID, userID)
	if err != nil {
		return
	}
	if period := s.PeriodFor(coin); period != coin.Period {
		if err := s.coins.UpdateCoinFields(userID, map[uint]map[string]any{coin.ID: {"period": period}}); err != nil {
			s.warn("Could not save the period of coin %d: %v", coin.ID, err)
		}
	}
}

// RefreshAll works out every coin's period, at start-up, and stores those
// that changed; it says how many.
func (s *CoinPeriodService) RefreshAll() int {
	coins, err := s.coins.CoinsForPeriods()
	if err != nil {
		s.warn("Could not list the coins to give periods: %v", err)
		return 0
	}
	changed := 0
	for i := range coins {
		coin := &coins[i]
		if period := s.PeriodFor(coin); period != coin.Period {
			if err := s.coins.UpdateCoinFields(coin.UserID, map[uint]map[string]any{coin.ID: {"period": period}}); err != nil {
				s.warn("Could not save the period of coin %d: %v", coin.ID, err)
				continue
			}
			changed++
		}
	}
	return changed
}

func (s *CoinPeriodService) warn(format string, args ...any) {
	if s.logger != nil {
		s.logger.Warn("coins", format, args...)
	}
}

// periodNameYears reads the years a period's name gives
// ("República Portuguesa 1986–2001" → 1986).
var periodNameYears = regexp.MustCompile(`(\d{3,4})\s*[–-]\s*\d{3,4}\s*$`)

// periodStart is when a period began: the first year its name gives, else
// the earliest of its coins.
func periodStart(count repository.PeriodCount) int {
	if match := periodNameYears.FindStringSubmatch(count.Period); match != nil {
		if year, err := strconv.Atoi(match[1]); err == nil {
			return year
		}
	}
	if count.FromYear != nil {
		return *count.FromYear
	}
	return 0
}

// SortPeriodsMostRecentFirst orders a country's periods as uCoin lists
// them, the most recent first.
func SortPeriodsMostRecentFirst(counts []repository.PeriodCount) {
	sort.SliceStable(counts, func(i, j int) bool {
		if a, b := periodStart(counts[i]), periodStart(counts[j]); a != b {
			return a > b
		}
		return counts[i].Period < counts[j].Period
	})
}
