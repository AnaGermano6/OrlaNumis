package services

import (
	"strconv"
	"strings"
	"sync"

	"github.com/briandenicola/ancient-coins-api/models"
	"github.com/briandenicola/ancient-coins-api/repository"
)

// CoinCirculation sorts the collection's coins by whether they are still
// legal tender ("current"), no longer are, or nobody knows; by whether they
// are commemorative coins; and by Numista's kind of coin.
type CoinCirculation struct {
	// Current coins are circulation coins still legal tender.
	Current     []uint
	Demonetized []uint
	// NonCirculating coins were issued for collectors (and are not demonetized).
	NonCirculating []uint
	Unknown        []uint

	Commemorative []uint
	Regular       []uint
	// CommemorativeUnknown are coins Numista says nothing of.
	CommemorativeUnknown []uint

	// Classifications holds the coins of each kind, by key (see
	// NumistaClassificationKey); coins Numista says nothing of are in none.
	Classifications map[string]*CoinClassification
}

// CoinCirculationService tells which coins of a collection are current, and
// which are commemorative, from what Numista says of their types (as cached
// when they were looked up); it takes coins of countries that no longer exist
// for demonetized. It never asks Numista.
type CoinCirculationService struct {
	coins   *repository.CoinRepository
	catalog *NumistaCatalogService

	mu    sync.Mutex
	known map[int]NumistaTypeFacts
}

// NewCoinCirculationService creates the service.
func NewCoinCirculationService(coins *repository.CoinRepository, catalog *NumistaCatalogService) *CoinCirculationService {
	return &CoinCirculationService{coins: coins, catalog: catalog, known: map[int]NumistaTypeFacts{}}
}

// Classify sorts the user's collection coins (not wishlist or sold).
func (s *CoinCirculationService) Classify(userID uint) (CoinCirculation, error) {
	refs, err := s.coins.CollectionNumistaTypes(userID)
	if err != nil {
		return CoinCirculation{}, err
	}
	result := CoinCirculation{Current: []uint{}, Demonetized: []uint{}, NonCirculating: []uint{}, Unknown: []uint{}, Commemorative: []uint{}, Regular: []uint{}, CommemorativeUnknown: []uint{}, Classifications: map[string]*CoinClassification{}}
	for _, ref := range refs {
		facts, typeKnown := s.typeFacts(ref.NumistaTypeID)
		known := coinFacts(facts, typeKnown, ref.Country, ref.Circulation)
		switch {
		case !known.commemorativeKnown:
			result.CommemorativeUnknown = append(result.CommemorativeUnknown, ref.CoinID)
		case known.commemorative:
			result.Commemorative = append(result.Commemorative, ref.CoinID)
		default:
			result.Regular = append(result.Regular, ref.CoinID)
		}
		if key := NumistaClassificationKey(facts.Classification); typeKnown && key != "" {
			group := result.Classifications[key]
			if group == nil {
				group = &CoinClassification{Key: key, Label: strings.Join(strings.Fields(facts.Classification), " ")}
				result.Classifications[key] = group
			}
			group.CoinIDs = append(group.CoinIDs, ref.CoinID)
		}
		// Current means a circulation coin still legal tender: a collector
		// coin (non-circulating) is not current even when it is legal tender.
		switch {
		case known.legalTenderKnown && known.demonetized:
			result.Demonetized = append(result.Demonetized, ref.CoinID)
		case known.nonCirculating:
			result.NonCirculating = append(result.NonCirculating, ref.CoinID)
		case !known.legalTenderKnown:
			result.Unknown = append(result.Unknown, ref.CoinID)
		default:
			result.Current = append(result.Current, ref.CoinID)
		}
	}
	return result, nil
}

// knownFacts is what is known of one coin, and what is not.
type knownFacts struct {
	demonetized, legalTenderKnown     bool
	nonCirculating                    bool
	commemorative, commemorativeKnown bool
	// manual is true when the user set them on the coin.
	manual bool
}

// coinFacts tells whether a coin is demonetized and whether it is
// commemorative: as set by hand on the coin, else as Numista says of its type;
// coins of countries that no longer exist are not legal tender.
func coinFacts(facts NumistaTypeFacts, typeKnown bool, country, circulation string) knownFacts {
	if setting, ok := models.CoinCirculationFacts(circulation); ok {
		return knownFacts{
			demonetized: setting.Demonetized, legalTenderKnown: true, nonCirculating: setting.NonCirculating,
			commemorative: setting.Commemorative, commemorativeKnown: true, manual: true,
		}
	}
	known := knownFacts{
		commemorative: facts.Commemorative, commemorativeKnown: typeKnown,
		// Numista's "non-circulating coins": issued for collectors.
		nonCirculating: typeKnown && NumistaClassificationKey(facts.Classification) == ClassificationNonCirculating,
	}
	switch {
	case typeKnown && facts.Demonetized != nil:
		known.demonetized, known.legalTenderKnown = *facts.Demonetized, true
	case IsExtinctCountry(country):
		known.demonetized, known.legalTenderKnown = true, true
	}
	return known
}

// CoinStatus is what is known of one coin: whether it is still legal tender,
// whether it was issued for collectors and whether it is commemorative, by the
// same rules as Classify.
type CoinStatus struct {
	// Current is null when nobody knows.
	Current *bool `json:"current"`
	// DemonetizationDate is when it stopped being legal tender, if Numista says.
	DemonetizationDate string `json:"demonetizationDate,omitempty"`
	// NonCirculating is true for coins issued for collectors rather than for
	// circulation (Numista's "non-circulating coins"); null when nobody knows.
	NonCirculating *bool `json:"nonCirculating"`
	// Commemorative is null when Numista says nothing of the coin.
	Commemorative     *bool  `json:"commemorative"`
	CommemoratedTopic string `json:"commemoratedTopic,omitempty"`
	// Manual is true when the user set them on the coin (its circulation).
	Manual bool `json:"manual"`
}

// Status tells whether one coin is current and whether it is commemorative,
// by the same rules as Classify.
func (s *CoinCirculationService) Status(coin *models.Coin) CoinStatus {
	numistaTypeID := ""
	var firstID uint
	for _, reference := range coin.References {
		if strings.ToLower(reference.Catalog) == "numista" && (numistaTypeID == "" || reference.ID < firstID) {
			numistaTypeID, firstID = strings.TrimSpace(reference.Number), reference.ID
		}
	}
	facts, typeKnown := s.typeFacts(numistaTypeID)
	known := coinFacts(facts, typeKnown, coin.Country, coin.Circulation)
	status := CoinStatus{Manual: known.manual}
	if known.manual || typeKnown {
		nonCirculating := known.nonCirculating
		status.NonCirculating = &nonCirculating
	}
	if known.legalTenderKnown {
		current := !known.demonetized
		status.Current = &current
		if known.demonetized && typeKnown && facts.Demonetized != nil && *facts.Demonetized {
			status.DemonetizationDate = facts.DemonetizationDate
		}
	}
	if known.commemorativeKnown {
		commemorative := known.commemorative
		status.Commemorative = &commemorative
		if commemorative && typeKnown {
			status.CommemoratedTopic = facts.CommemoratedTopic
		}
	}
	return status
}

// typeFacts looks the type up once; what Numista says of it is kept.
func (s *CoinCirculationService) typeFacts(numistaTypeID string) (NumistaTypeFacts, bool) {
	typeID, err := strconv.Atoi(strings.TrimSpace(numistaTypeID))
	if err != nil || typeID <= 0 || s.catalog == nil {
		return NumistaTypeFacts{}, false
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if facts, ok := s.known[typeID]; ok {
		return facts, true
	}
	facts, ok := s.catalog.CachedTypeFacts(typeID)
	if ok {
		s.known[typeID] = facts
	}
	return facts, ok
}
