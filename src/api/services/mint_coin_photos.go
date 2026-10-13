package services

import (
	"context"
	"math"
	"regexp"
	"strconv"
	"strings"
	"unicode"

	"github.com/briandenicola/ancient-coins-api/models"
)

// Mints publish official photos of their collector coins: Portugal's (INCM),
// France's (Monnaie de Paris) and Austria's (Münze Österreich). Their
// catalogues are matched to a coin by face value, year, topic (or lettering)
// and designer.

// MintPhotoSource is a mint's catalogue of coin photos.
type MintPhotoSource interface {
	// Country is the ECB name of the country whose coins the mint strikes.
	Country() string
	// Source and Credit name the photos' origin (models.CoinPhotoSource…) and owner.
	Source() string
	Credit() string
	// Thumbnail asks the mint's server for a smaller copy of a photo.
	Thumbnail(image string) string
	Match(ctx context.Context, query MintPhotoQuery) (*MintPhoto, error)
	// Confirm completes a match with photos that need checking first, when
	// the coin is about to be shown on its own or added.
	Confirm(ctx context.Context, photo *MintPhoto) *MintPhoto
}

// mintCoin is a coin of a mint's catalogue, as kept in the cache.
type mintCoin struct {
	Title string `json:"title"`
	// OtherTitles are the coin's names in other languages, or other words
	// that tell it apart from coins of the same series.
	OtherTitles []string `json:"otherTitles,omitempty"`
	// DesignKey groups a design's metal and finish variants; from Title when empty.
	DesignKey string  `json:"designKey,omitempty"`
	PageURL   string  `json:"pageUrl"`
	Value     float64 `json:"value"`
	Metal     string  `json:"metal"`
	Authors   string  `json:"authors,omitempty"`
	Series    string  `json:"series,omitempty"`
	Year      int     `json:"year"`
	// Released is the month the mint puts the coin on sale ("2026-07"), and
	// Volume how many it strikes, when the mint says.
	Released string `json:"released,omitempty"`
	Volume   int64  `json:"volume,omitempty"`
	Text     string `json:"text,omitempty"`
	Obverse  string `json:"obverse"`
	Reverse  string `json:"reverse,omitempty"`
	// ReverseGuess is a photo that probably shows the reverse, to be checked.
	ReverseGuess string `json:"reverseGuess,omitempty"`
}

// MintPhotoQuery describes a euro coin to find in a mint's catalogue.
type MintPhotoQuery struct {
	Year  int
	Value float64
	// Metal is the coin's composition, in any language.
	Metal string
	// Topic is what the coin commemorates (Numista's topic, or its title's subject).
	Topic string
	// Texts are other words about the coin: series, side descriptions, lettering.
	Texts     []string
	Designers []string
	// AltTopics and AltYears are tried in turn when Topic and Year find
	// nothing: the title's subject alone ("Snoopy", without the "75th
	// anniversary of Peanuts" Numista adds) or the coin's lettering, and the
	// year on the coin when Numista's first year differs.
	AltTopics []string
	AltYears  []int
}

// MintPhoto is a mint's photos of a coin; Reverse may be empty.
type MintPhoto struct {
	Title   string
	Obverse string
	Reverse string
	PageURL string
	// reverseGuess is a photo that probably shows the reverse, until Confirm checks it.
	reverseGuess string
}

// mintStopWords are words every coin of the shop shares.
var mintStopWords = map[string]bool{
	"moeda": true, "moedas": true, "coin": true, "portugal": true, "portuguesa": true, "republica": true, "casa": true,
	"prata": true, "ouro": true, "silver": true, "gold": true, "proof": true, "normal": true, "bnc": true, "fdc": true,
	"cuproniquel": true, "cupronickel": true, "nickel": true, "nikel": true, "copper": true, "serie": true, "series": true, "edicao": true,
	"uma": true, "que": true, "nas": true, "nos": true, "pela": true, "pelo": true, "the": true, "and": true, "of": true,
}

func mintWords(text string) []string {
	var words []string
	for _, word := range ecbWords(text) {
		if !mintStopWords[word] {
			words = append(words, word)
		}
	}
	return words
}

// mintSimilarWords allows for small spelling differences between languages
// ("Ulysses" and "Ulisses", "Vincent" and "Vicente").
func mintSimilarWords(a, b string) bool {
	if ecbSimilarWords(a, b) {
		return true
	}
	if len(a) < 5 || len(b) < 5 || unicode.IsDigit(rune(a[0])) || unicode.IsDigit(rune(b[0])) {
		return false
	}
	limit := 1
	if len(a) >= 7 && len(b) >= 7 {
		limit = 2
	}
	return editDistance(a, b) <= limit
}

func editDistance(a, b string) int {
	previous := make([]int, len(b)+1)
	for j := range previous {
		previous[j] = j
	}
	for i := 1; i <= len(a); i++ {
		current := make([]int, len(b)+1)
		current[0] = i
		for j := 1; j <= len(b); j++ {
			cost := 1
			if a[i-1] == b[j-1] {
				cost = 0
			}
			current[j] = min(previous[j]+1, current[j-1]+1, previous[j-1]+cost)
		}
		previous = current
	}
	return previous[len(b)]
}

// mintOverlap counts the words of text found, loosely, among words.
func mintOverlap(words []string, text string) int {
	count := 0
	seen := map[string]bool{}
	for _, word := range mintWords(text) {
		if seen[word] {
			continue
		}
		seen[word] = true
		for _, candidate := range words {
			if mintSimilarWords(word, candidate) {
				count++
				break
			}
		}
	}
	return count
}

// mintMatchedWords returns the words found, loosely, in text.
func mintMatchedWords(words []string, text string) []string {
	textWords := mintWords(text)
	var matched []string
	for _, word := range words {
		for _, candidate := range textWords {
			if mintSimilarWords(word, candidate) {
				matched = append(matched, word)
				break
			}
		}
	}
	return matched
}

// mintDesignerMatch tells whether one of the designers' surnames is among the
// shop's authors.
func mintDesignerMatch(designers []string, authors string) bool {
	authorWords := ecbWords(authors)
	for _, designer := range designers {
		names := ecbWords(designer)
		if len(names) == 0 {
			continue
		}
		surname := names[len(names)-1]
		for _, author := range authorWords {
			if author == surname {
				return true
			}
		}
	}
	return false
}

// mintDesignKey groups a design's metal and finish variants: "O Conhecimento
// (Prata Proof)" and "O Conhecimento" are the same coin.
func mintDesignKey(title string) string {
	if open := strings.Index(title, "("); open > 0 {
		title = title[:open]
	}
	return strings.Join(uniqueStrings(mintWords(title)), " ")
}

// mintSpecialEdition tells a proof, shaped or piedfort edition, whose photos
// look least like the coin of the normal edition.
var mintSpecialEdition = regexp.MustCompile(`(?i)proof|qualite-be|\bbe\b|shape|forme|piedfort|high-relief|haut-relief`)

func mintMetalKind(metal string) string {
	folded := foldASCII(metal)
	switch {
	case strings.Contains(folded, "prata") || strings.Contains(folded, "silver") || strings.Contains(folded, "argent"):
		return "silver"
	case strings.Contains(folded, "ouro") || strings.Contains(folded, "gold"):
		return "gold"
	default:
		return "base"
	}
}

// matchMintCoin returns the coin of a mint's catalogue that is clearly the one
// described, or nil, trying the query's other topics and years when its own
// find nothing.
func matchMintCoin(coins []mintCoin, query MintPhotoQuery) *mintCoin {
	tried := map[string]bool{}
	for _, year := range append([]int{query.Year}, query.AltYears...) {
		for _, topic := range append([]string{query.Topic}, query.AltTopics...) {
			key := strconv.Itoa(year) + "|" + topic
			if tried[key] {
				continue
			}
			tried[key] = true
			variant := query
			variant.Year, variant.Topic = year, topic
			if coin := matchMintDesign(coins, variant); coin != nil {
				return coin
			}
		}
	}
	return nil
}

// matchMintDesign returns the coin of a mint's catalogue that is clearly the
// one described, or nil. A coin must be of the same face value and year (give
// or take one: the release can follow the year on the coin), and its title
// must name the coin's topic with a word no other design has, or its author
// must be the coin's designer.
func matchMintDesign(coins []mintCoin, query MintPhotoQuery) *mintCoin {
	if query.Year == 0 || query.Value <= 0 {
		return nil
	}
	var topicWords []string
	seen := map[string]bool{}
	for _, word := range mintWords(query.Topic) {
		if !seen[word] {
			seen[word] = true
			topicWords = append(topicWords, word)
		}
	}
	textWords := mintWords(strings.Join(query.Texts, " "))
	type design struct {
		coins []mintCoin
		// words name the design; topic holds the topic words its title has.
		words     []string
		topic     map[string]bool
		designer  bool
		text      int
		yearExact bool
		score     int
	}
	designs := map[string]*design{}
	for _, coin := range coins {
		if math.Abs(coin.Value-query.Value) > 0.001 || coin.Year < query.Year-1 || coin.Year > query.Year+1 {
			continue
		}
		// "snoopy snoopy" (the shaped edition's "snoopy-…-snoopy-shape") is "snoopy".
		key := strings.Join(uniqueStrings(strings.Fields(coin.DesignKey)), " ")
		if key == "" {
			key = mintDesignKey(coin.Title)
		}
		d := designs[key]
		if d == nil {
			d = &design{words: mintWords(key), topic: map[string]bool{}}
			designs[key] = d
		}
		d.coins = append(d.coins, coin)
		for _, word := range mintMatchedWords(topicWords, coin.Title+" "+strings.Join(coin.OtherTitles, " ")) {
			d.topic[word] = true
		}
		d.designer = d.designer || mintDesignerMatch(query.Designers, coin.Authors)
		d.text = max(d.text, min(mintOverlap(textWords, coin.Title+" "+coin.Series+" "+coin.Text), 3))
		d.yearExact = d.yearExact || coin.Year == query.Year
	}
	// A word several designs have, such as their series' name, tells none apart.
	shared, sharedNames := map[string]int{}, map[string]int{}
	for _, d := range designs {
		for word := range d.topic {
			shared[word]++
		}
		for _, word := range uniqueStrings(d.words) {
			sharedNames[word]++
		}
	}
	for key, d := range designs {
		distinct := 0
		for word := range d.topic {
			if shared[word] == 1 {
				distinct++
			}
		}
		// The coin's title must have all the topic's words, one at least its
		// own; or, for a coin of the very year, half of them, two its own; or
		// the designer must be its author. A coin sharing a word or two with the
		// topic is another one ("Palace of Versailles" is not "National palace
		// of Sintra").
		// A coin of the year before or after needs half the topic's words its own.
		// Numbers ("85.º aniversário") count when the title has them, but a
		// title without them can still name the whole topic.
		coverage := float64(countWords(d.topic, false)) / float64(max(countWords(seen, false), 1))
		if countWords(seen, false) == 0 {
			coverage = float64(len(d.topic)) / float64(max(len(topicWords), 1))
		}
		named := distinct > 0 && (coverage == 1 || (coverage >= 0.5 && d.yearExact && distinct >= 2)) &&
			(d.yearExact || (2*distinct >= len(topicWords) && namesTopic(d.words, sharedNames, topicWords)))
		if !named && !d.designer {
			delete(designs, key)
			continue
		}
		d.score = 3*distinct + d.text
		if d.designer {
			d.score += 3
		}
		if d.yearExact {
			d.score++
		}
	}
	var best, second *design
	for _, candidate := range designs {
		switch {
		case best == nil || candidate.score > best.score:
			best, second = candidate, best
		case second == nil || candidate.score > second.score:
			second = candidate
		}
	}
	if best == nil || best.score < 3 || (second != nil && best.score-second.score < 2) {
		return nil
	}
	// The version closest to the coin: same metal, else the nearest-looking
	// one, and the normal edition before proofs, shapes and special finishes.
	wanted := mintMetalKind(query.Metal)
	rank := func(coin mintCoin) int {
		order := map[string][]string{"base": {"base", "silver", "gold"}, "silver": {"silver", "base", "gold"}, "gold": {"gold", "silver", "base"}}[wanted]
		r := 0
		for i, kind := range order {
			if kind == mintMetalKind(coin.Metal) {
				r = 4 * i
			}
		}
		if mintSpecialEdition.MatchString(coin.PageURL[strings.LastIndex(coin.PageURL, "/")+1:]) {
			r += 2
		}
		if strings.Contains(coin.Title, "(") {
			r++
		}
		return r
	}
	chosen := best.coins[0]
	for _, coin := range best.coins[1:] {
		if rank(coin) < rank(chosen) {
			chosen = coin
		}
	}
	return &chosen
}

var mintEuroValue = regexp.MustCompile(`(?i)^\s*(\d+)?\s*([.,]\d+|½|¼)?\s*(euros?|eur|€)(\s|\(|$)`)

// MintFaceValue reads a euro face value such as "5 Euros", "7½ euro" or
// "2,50 €"; 0 when the denomination is not in euro.
func MintFaceValue(denomination string) float64 {
	match := mintEuroValue.FindStringSubmatch(denomination)
	if match == nil || (match[1] == "" && match[2] == "") || strings.Contains(foldASCII(denomination), "cent") {
		return 0
	}
	value, _ := strconv.ParseFloat(match[1], 64)
	switch fraction := match[2]; fraction {
	case "":
	case "½":
		value += 0.5
	case "¼":
		value += 0.25
	default:
		part, _ := strconv.ParseFloat("0."+fraction[1:], 64)
		value += part
	}
	return value
}

// MintTopic is the subject of a Numista title: "The knowledge" in "5 Euro
// (The knowledge, cupronickel)", or "Rugby World Cup 2023 - England" in
// "¼ Euro Rugby World Cup 2023 - England".
func MintTopic(title string) string {
	topic := ""
	if open, end := strings.Index(title, "("), strings.LastIndex(title, ")"); open >= 0 && end > open {
		topic = title[open+1 : end]
	} else if match := mintEuroValue.FindStringIndex(title); match != nil {
		topic = title[match[1]:]
	}
	// "silver 999‰" says nothing about the design.
	return strings.TrimSpace(mintFineness.ReplaceAllString(topic, " "))
}

var mintFineness = regexp.MustCompile(`\d+\s*‰`)

// mintVersionWords name a coin's metal or finish, not its design.
var mintVersionWords = map[string]bool{
	"argent": true, "silver": true, "prata": true, "gold": true, "ouro": true, "copper": true, "cobre": true, "nickel": true, "nikel": true,
	"cupronickel": true, "cuproniquel": true, "brass": true, "bronze": true, "bimetallic": true, "bimetalica": true,
	"coloured": true, "colored": true, "colorized": true, "proof": true, "piedfort": true, "shape": true, "forme": true,
}

// MintTitleSubject is a Numista title's subject alone, without its face
// value, metal or finish: "Snoopy" in "10 euros Snoopy (argent 333‰)",
// "Pinocchio" in "10 Euros (Pinocchio, silver 333‰)".
func MintTitleSubject(title string) string {
	if match := mintEuroValue.FindStringIndex(title); match != nil {
		title = title[match[1]:]
	}
	var words []string
	for _, word := range ecbWords(mintFineness.ReplaceAllString(title, " ")) {
		if !mintVersionWords[word] {
			words = append(words, word)
		}
	}
	return strings.Join(words, " ")
}

// mintQueryForType describes a euro coin type for a mint's catalogue from
// its Numista details; ok is false when it has no euro face value.
func mintQueryForType(detail *models.NumistaCatalogType) (MintPhotoQuery, bool) {
	if detail == nil || detail.MinYear == nil {
		return MintPhotoQuery{}, false
	}
	value := MintFaceValue(detail.Denomination)
	if detail.FaceValue != nil && strings.Contains(foldASCII(detail.Currency), "euro") {
		value = *detail.FaceValue
	}
	if value <= 0 {
		value = MintFaceValue(detail.Title)
	}
	query := MintPhotoQuery{
		Year:  *detail.MinYear,
		Value: value,
		Metal: detail.Composition,
		Topic: strings.TrimSpace(detail.CommemoratedTopic + " " + MintTopic(detail.Title)),
		Texts: []string{detail.Series, detail.Obverse.Description, detail.Obverse.Lettering, detail.Reverse.Description, detail.Reverse.Lettering},
	}
	// The title's subject alone, then the lettering of the coin's design (a
	// mint can name the coin in its own language: "Weltmeister"). The
	// obverse's ("REPUBLIK ÖSTERREICH") would name a coin about the republic.
	for _, topic := range []string{MintTitleSubject(detail.Title), detail.Reverse.Lettering} {
		if hasNonNumberWord(mintWords(mintFineness.ReplaceAllString(topic, " "))) {
			query.AltTopics = append(query.AltTopics, topic)
		}
	}
	for _, side := range []models.NumistaCatalogSide{detail.Obverse, detail.Reverse} {
		query.Designers = append(query.Designers, side.Designers...)
		query.Designers = append(query.Designers, side.Engravers...)
	}
	return query, value > 0
}

// mintQueryForTitle describes a euro coin type for a mint's catalogue from a
// search result's title alone.
func mintQueryForTitle(title string, minYear *int) (MintPhotoQuery, bool) {
	if minYear == nil {
		return MintPhotoQuery{}, false
	}
	query := MintPhotoQuery{Year: *minYear, Value: MintFaceValue(title), Metal: title, Topic: MintTopic(title)}
	if subject := MintTitleSubject(title); subject != "" {
		query.AltTopics = []string{subject}
	}
	return query, query.Value > 0
}

// mintSourceFor returns the mint that strikes the coins of the country named
// (an issuer code or name, in English or Portuguese), or nil.
func mintSourceFor(sources []MintPhotoSource, names ...string) MintPhotoSource {
	for _, name := range names {
		if country := EcbCountryName(name); country != "" {
			for _, source := range sources {
				if source.Country() == country {
					return source
				}
			}
		}
	}
	return nil
}

// mintCandidates turns a mint's photos of a coin into photo candidates; the
// reverse is nil when the mint has no photo of it.
func mintCandidates(source MintPhotoSource, photo *MintPhoto) (obverse, reverse *models.CoinPhotoCandidate) {
	if source == nil || photo == nil {
		return nil, nil
	}
	candidate := func(side, image string) *models.CoinPhotoCandidate {
		if image == "" {
			return nil
		}
		return &models.CoinPhotoCandidate{
			Source: source.Source(), Side: side, Thumbnail: source.Thumbnail(image), ImageURL: image, PageURL: photo.PageURL,
			Title: photo.Title, Author: source.Credit(), Match: models.CoinPhotoMatchExact,
		}
	}
	return candidate("obverse", photo.Obverse), candidate("reverse", photo.Reverse)
}

// namesTopic tells whether the topic has at least half the words that set a
// design apart from the others ("aries" in "astrological signs aries"), so a
// coin of another year is not taken for one that merely shares a word with it
// ("Disney" is not "Disney the Lion King 30th").
func namesTopic(words []string, shared map[string]int, topic []string) bool {
	own, named := 0, 0
	for _, word := range uniqueStrings(words) {
		if shared[word] > 1 {
			continue
		}
		own++
		if len(mintMatchedWords([]string{word}, strings.Join(topic, " "))) > 0 {
			named++
		}
	}
	return 2*named >= own
}

func uniqueStrings(values []string) []string {
	seen := map[string]bool{}
	var unique []string
	for _, value := range values {
		if !seen[value] {
			seen[value] = true
			unique = append(unique, value)
		}
	}
	return unique
}

// hasNonNumberWord tells whether a topic names something besides numbers
// (a lettering of just "2021" names no design).
func hasNonNumberWord(words []string) bool {
	for _, word := range words {
		if strings.Trim(word, "0123456789") != "" {
			return true
		}
	}
	return false
}

// countWords counts the words of a set that are numbers (or are not).
func countWords(words map[string]bool, numbers bool) int {
	count := 0
	for word := range words {
		if (strings.Trim(word, "0123456789") == "") == numbers {
			count++
		}
	}
	return count
}
