package services

import (
	"html"
	"regexp"
	"strings"
)

// Numista's comments on a type sometimes say its coins are issued in a coin
// card ("Issued in a coincard", "Emitido num cartão de moedas", "Moeda no
// interior do cartão de apresentação oficial"). Other mentions of a card are
// about a special version of the type ("Cartão BU", "Cartão de prova", "Mais
// 10 000 UN em cartão de moedas") and do not count: the ordinary coin is loose.
var (
	commentTags      = regexp.MustCompile(`<[^>]*>`)
	commentSentences = regexp.MustCompile(`[.!?;:]\s+|\n+`)
	issuedInCoincard = []*regexp.Regexp{
		regexp.MustCompile(`\bissued (?:in|on) (?:an? )?(?:official )?(?:coin ?card|coincard|card)\b`),
		regexp.MustCompile(`\bemitid[ao]s? (?:num|numa|em|no|na) (?:um |uma )?cartao\b`),
		regexp.MustCompile(`\bno interior (?:do|de um) cartao\b`),
	}
	// "Also issued in a coincard", "mais 10 000 em cartão": only some of them.
	someOnlyWords = regexp.MustCompile(`\b(?:also|some|part|tambem|mais|alem|algumas|alguns|parte)\b`)
)

// IssuedInCoincard tells whether a type's comments (HTML) say its coins are
// issued in a coin card.
func IssuedInCoincard(comments string) bool {
	text := foldASCII(html.UnescapeString(commentTags.ReplaceAllString(comments, " ")))
	for _, sentence := range commentSentences.Split(text, -1) {
		sentence = strings.Join(strings.Fields(sentence), " ")
		if someOnlyWords.MatchString(sentence) {
			continue
		}
		for _, pattern := range issuedInCoincard {
			if pattern.MatchString(sentence) {
				return true
			}
		}
	}
	return false
}
