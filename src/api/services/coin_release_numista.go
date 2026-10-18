package services

import (
	"context"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"

	"github.com/briandenicola/ancient-coins-api/models"
)

// ReleaseNumistaType is the Numista type of a new coin, to fill in its
// details when it is added to the collection: the type when it is sure, and
// the types of its country, year and face value it may be, to choose from.
type ReleaseNumistaType struct {
	Status string `json:"status" enums:"ok,unconfigured,unauthorized,quota_limited,unavailable"`
	// TypeID is the coin's type; absent when none is sure.
	TypeID     int                               `json:"typeId,omitempty"`
	Candidates []models.NumistaCatalogSearchItem `json:"candidates"`
}

// IsCoinReleaseSource tells whether a release can come from source: the
// ECB, a mint (models.CoinPhotoSource…) or Numista.
func IsCoinReleaseSource(source string) bool {
	return source == CoinReleaseSourceECB || source == CoinReleaseSourceNumista || mintReleaseSources[source]
}

// NumistaType finds the Numista type of a release: its own when it comes
// from Numista, the one the catalogue knows it by, else the one of Numista's
// coins of its country, year and face value whose ECB image or mint page is
// the release's, or the only one whose name is the release's. Numista is
// asked once a day at most for a country, year and face value; the rest
// comes from what the app has stored.
func (s *CoinReleasesService) NumistaType(ctx context.Context, release CoinRelease) ReleaseNumistaType {
	result := ReleaseNumistaType{Status: models.NumistaDataStatusOK, Candidates: []models.NumistaCatalogSearchItem{}}
	if release.Source == CoinReleaseSourceNumista {
		if id := numistaPageTypeID(release.PageURL); id > 0 {
			result.TypeID = id
			return result
		}
	}
	keys := s.releaseKeys(ctx, release)
	for _, key := range keys {
		if id, ok := numistaKeyTypeID(key); ok {
			result.TypeID = id
			return result
		}
	}
	if s.numista == nil {
		result.Status = models.NumistaDataStatusUnconfigured
		return result
	}
	issuer := numistaEuroIssuers[release.Country]
	year, _ := strconv.Atoi(release.From[:min(4, len(release.From))])
	if issuer == "" || year == 0 || release.FaceValue <= 0 {
		// A joint issue is of the country the user chooses.
		return result
	}
	search := NumistaTypeSearch{Issuer: issuer, Year: year, Lang: "en", Count: NumistaCatalogMaxPageSize}
	// Numista finds "2 Euros (…)" for "2 euro"; a value such as 2½ is
	// looked for among the year's coins.
	if release.FaceValue == math.Trunc(release.FaceValue) {
		search.Query = fmt.Sprintf("%d euro", int(release.FaceValue))
	}
	found := s.numista.Search(ctx, search)
	if found.Status != models.NumistaDataStatusOK {
		result.Status = found.Status
		return result
	}
	for _, item := range found.Types {
		if MintFaceValue(item.Title) == release.FaceValue && !escudoPattern.MatchString(item.Title) {
			result.Candidates = append(result.Candidates, item)
		}
	}
	result.TypeID = s.releaseCandidate(ctx, release, keys, result.Candidates)
	return result
}

// releaseKeys returns what identifies a release: the keys of the app's own
// copy of it, among the releases and, once read, the catalogue (whose keys
// may name its Numista type), and its page and picture.
func (s *CoinReleasesService) releaseKeys(ctx context.Context, release CoinRelease) []string {
	own := []string{}
	for _, value := range []string{release.PageURL, release.ImageURL} {
		if value != "" {
			own = append(own, value)
		}
	}
	keys := append([]string{}, own...)
	for _, group := range s.Releases(ctx) {
		for _, listed := range group.Releases {
			if sameRelease(listed, release) || sharesKey(listed.Keys, own) {
				keys = append(keys, listed.Keys...)
			}
		}
	}
	// The catalogue is not read for this: only when it is read already.
	s.catalogueMu.Lock()
	catalogue := s.catalogue
	s.catalogueMu.Unlock()
	known := append([]string{}, keys...)
	for _, listed := range catalogue {
		if sameRelease(listed, release) || sharesKey(listed.Keys, known) {
			keys = append(keys, listed.Keys...)
		}
	}
	return uniqueStrings(keys)
}

// sameRelease tells whether two releases are the same coin: of the same
// source, country, title and year.
func sameRelease(a, b CoinRelease) bool {
	return a.Source == b.Source && a.Country == b.Country && a.Title == b.Title &&
		len(a.From) >= 4 && len(b.From) >= 4 && a.From[:4] == b.From[:4]
}

func sharesKey(keys, others []string) bool {
	for _, key := range keys {
		for _, other := range others {
			if key != "" && key == other {
				return true
			}
		}
	}
	return false
}

// releaseCandidate returns the candidate a release is: the one whose ECB
// image or mint page is one of the release's keys, else the only one whose
// name (or, in the mint's language, what it commemorates) shares most of the
// release's distinctive words; 0 when none is sure.
func (s *CoinReleasesService) releaseCandidate(ctx context.Context, release CoinRelease, keys []string, candidates []models.NumistaCatalogSearchItem) int {
	known := map[string]bool{}
	for _, key := range keys {
		known[key] = true
	}
	for _, item := range candidates {
		if image, _ := s.numista.ecbImage(ctx, item.Title, item.Issuer, item.MinYear, item.MaxYear); image != "" && known[image] {
			return item.ID
		}
		query, ok := mintQueryForTitle(item.Title, item.MinYear)
		if detail := s.numista.cachedTypeDetail(item.ID); detail != nil {
			query, ok = mintQueryForType(detail)
		}
		if _, photo := s.numista.mintPhoto(ctx, item.Issuer, query, ok); photo != nil && known[photo.PageURL] {
			return item.ID
		}
	}
	words := topicWords(release.Title)
	if len(words) == 0 {
		return 0
	}
	lang := mintTopicLangs[release.Source]
	best, bestOverlap, tied := 0, 0, false
	for _, item := range candidates {
		names := []string{MintTopic(item.Title)}
		if lang != "" {
			if topic, ok := s.numista.CachedTypeTopic(item.ID, lang); ok {
				names = append(names, topic.Topic, topic.ReverseLettering)
			}
		}
		overlap := 0
		for _, name := range names {
			other := topicWords(name)
			shorter := min(len(words), len(other))
			if count := wordOverlap(words, other, topicSimilarWords); shorter > 0 && count >= (shorter+1)/2 {
				overlap = max(overlap, count)
			}
		}
		switch {
		case overlap == 0:
		case overlap > bestOverlap:
			best, bestOverlap, tied = item.ID, overlap, false
		case overlap == bestOverlap:
			tied = true
		}
	}
	if tied {
		return 0
	}
	return best
}

var numistaPiecesPage = regexp.MustCompile(`^https://(?:[a-z]{2}\.)?numista\.com/catalogue/pieces(\d+)\.html`)

// numistaPageTypeID is the type of a Numista catalogue page
// ("https://pt.numista.com/catalogue/pieces1234.html"), or 0.
func numistaPageTypeID(page string) int {
	match := numistaPiecesPage.FindStringSubmatch(page)
	if match == nil {
		return 0
	}
	id, _ := strconv.Atoi(match[1])
	return id
}

// numistaKeyTypeID is the type a release key names (see numistaReleaseKey).
func numistaKeyTypeID(key string) (int, bool) {
	value, ok := strings.CutPrefix(key, "numista:")
	if !ok {
		return 0, false
	}
	id, err := strconv.Atoi(value)
	return id, err == nil && id > 0
}
