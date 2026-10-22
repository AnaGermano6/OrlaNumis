package services

import (
	"context"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/briandenicola/ancient-coins-api/models"
)

// Ready-made type sets: one place for each type of Portugal's circulation
// coins of a monetary system, as Numista lists them year by year, filled by
// the collection's coin linked to that Numista type. Commemorative and
// collector coins are in the catalogue, and patterns are left out.
const (
	// Portugal's escudos, from the Republic's first coins to the euro.
	ChallengeEscudoTypes = "escudo-types"
	// The réis of the decimal system, from 1835 to the Republic.
	ChallengeReisTypes = "reis-types"
)

type portugueseTypeSeries struct {
	first, last int
	// denomination tells a type of the series by its title.
	denomination *regexp.Regexp
	currency     string
}

var portugueseTypeSeriesByChallenge = map[string]portugueseTypeSeries{
	ChallengeEscudoTypes: {first: escudoFirstYear, last: escudoLastYear, denomination: regexp.MustCompile(`(?i)\b(escudos?|centavos?)\b|\d\s*½`), currency: "PTE"},
	ChallengeReisTypes:   {first: 1835, last: 1910, denomination: regexp.MustCompile(`(?i)\b(r[ée]is|real)\b`), currency: "PTR"},
}

// numistaYearsPerView bounds the years of a series asked of Numista in one
// view; the rest come at the next ones, and the list says it is not complete.
const numistaYearsPerView = 6

var (
	typeSeriesValue = regexp.MustCompile(`^\s*(\d*½|\d+(?:[.,]\d+)?)`)
	typeSeriesCents = regexp.MustCompile(`(?i)^\s*[\d.,½]+\s*centavos?\b`)
)

// typeSeriesFaceValue reads a type's face value from its title, in escudos
// for the escudo series (50 Centavos is 0.5) and in réis for the réis.
func typeSeriesFaceValue(title string) float64 {
	match := typeSeriesValue.FindStringSubmatch(title)
	if match == nil {
		return 0
	}
	value := strings.ReplaceAll(match[1], ",", ".")
	half := strings.HasSuffix(value, "½")
	number, _ := strconv.ParseFloat(strings.TrimSuffix(value, "½"), 64)
	if half {
		number += 0.5
	}
	if typeSeriesCents.MatchString(title) {
		number /= 100
	}
	return number
}

// seriesType tells whether a Numista type is one of a series' circulation
// types: of its denomination, not a pattern, and not a commemorative or
// collector coin (Numista's classification says so when cached; otherwise a
// type of one year whose title names what it marks is one).
func seriesType(series portugueseTypeSeries, item models.NumistaCatalogSearchItem, classification string) bool {
	if !series.denomination.MatchString(item.Title) || escudoPattern.MatchString(item.Title) || item.MinYear == nil {
		return false
	}
	if *item.MinYear > series.last || item.MaxYear != nil && *item.MaxYear < series.first {
		return false
	}
	switch {
	case classCollector.MatchString(classification), classCommemorative.MatchString(classification):
		return false
	case classification != "":
		return true
	}
	oneYear := item.MaxYear == nil || *item.MaxYear == *item.MinYear
	return !(oneYear && escudoTopic.MatchString(item.Title))
}

// seriesTypes lists a series' types, oldest first, from the years Numista
// has listed (cached for a year). Years not cached yet are asked for, a few
// per view and only in the background's share of the quota; complete is
// false until every year has answered.
func (s *SetChecklistService) seriesTypes(ctx context.Context, series portugueseTypeSeries) (types []models.NumistaCatalogSearchItem, complete bool) {
	if s.numista == nil {
		return nil, false
	}
	ctx = WithNumistaBackground(ctx)
	complete = true
	asked := 0
	seen := map[int]bool{}
	for year := series.first; year <= series.last; year++ {
		if !s.numista.IssuerYearCached("portugal", year, "pt") {
			if asked >= numistaYearsPerView {
				complete = false
				continue
			}
			asked++
		}
		items, err := s.numista.IssuerYearTypes(ctx, "portugal", year, "pt")
		if err != nil {
			// Numista cannot be asked now (no key, quota): try again later.
			complete = false
			asked = numistaYearsPerView
			continue
		}
		for _, item := range items {
			if seen[item.ID] {
				continue
			}
			seen[item.ID] = true
			facts, _ := s.numista.CachedTypeFacts(item.ID)
			if seriesType(series, item, facts.Classification) {
				types = append(types, item)
			}
		}
	}
	sort.SliceStable(types, func(i, j int) bool {
		if *types[i].MinYear != *types[j].MinYear {
			return *types[i].MinYear < *types[j].MinYear
		}
		return typeSeriesFaceValue(types[i].Title) < typeSeriesFaceValue(types[j].Title)
	})
	return types, complete
}

// typeSeriesSlots lists a series' types, each filled by the collection's coin
// of that Numista type (one with a photo first).
func (s *SetChecklistService) typeSeriesSlots(ctx context.Context, challenge string, coins []models.Coin) ([]ChecklistSlot, bool) {
	series, ok := portugueseTypeSeriesByChallenge[challenge]
	if !ok {
		return []ChecklistSlot{}, true
	}
	types, complete := s.seriesTypes(ctx, series)
	byType := map[int]*models.Coin{}
	for i := range coins {
		id := NumistaTypeIDForCoin(&coins[i])
		if id == 0 {
			continue
		}
		if current := byType[id]; current == nil || coinPhoto(current) == "" && coinPhoto(&coins[i]) != "" {
			byType[id] = &coins[i]
		}
	}
	slots := make([]ChecklistSlot, 0, len(types))
	for _, item := range types {
		slot := ChecklistSlot{
			Key: fmt.Sprintf("type:%d", item.ID), Title: item.Title, Country: "PT", Year: *item.MinYear,
			FaceValue: typeSeriesFaceValue(item.Title), Currency: series.currency,
			ImageURL: item.ObverseThumbnail, PageURL: fmt.Sprintf("https://pt.numista.com/catalogue/pieces%d.html", item.ID),
		}
		if item.MaxYear != nil && *item.MaxYear > *item.MinYear {
			slot.ToYear = *item.MaxYear
		}
		if coin := byType[item.ID]; coin != nil {
			fill(&slot, coin)
		}
		slots = append(slots, slot)
	}
	return slots, complete
}
