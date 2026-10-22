package services

import (
	"regexp"
	"strings"
)

// providerReference is a catalogue number as Numista gives it, for a type or
// for one of its issues.
type providerReference struct {
	Catalogue struct {
		Code string `json:"code"`
	} `json:"catalogue"`
	Number string `json:"number"`
}

var (
	// Numista writes a banknote issue's Pick number in its references or at
	// the start of its comment: "P#176a; Silva# 68 - Signatures: …".
	issueCommentPick = regexp.MustCompile(`(?i)\bP\s*#\s*(\d{1,4}[a-z]{0,2}\d?)\b`)
	// A specimen's Pick number ends in "s" ("13s", "176as").
	specimenPick     = regexp.MustCompile(`(?i)^\d+[a-z]?s\d?$`)
	specimenWords    = regexp.MustCompile(`(?i)\b(specimen|esp[eé]cime|amostra)`)
	replacementWords = regexp.MustCompile(`(?i)\b(replacement|star note|substitui[çc][ãa]o|reposi[çc][ãa]o)`)
)

// issuePickNumber is a banknote issue's own Pick number ("176a"), or "".
func issuePickNumber(references []providerReference, comment string) string {
	for _, reference := range references {
		if strings.EqualFold(strings.TrimSpace(reference.Catalogue.Code), "P") {
			if number := strings.TrimSpace(reference.Number); number != "" {
				return number
			}
		}
	}
	if match := issueCommentPick.FindStringSubmatch(comment); match != nil {
		return match[1]
	}
	return ""
}

// issueNoteVariant tells whether an issue's notes are specimen or replacement
// notes (see models.NoteVariants), or "" for ordinary notes.
func issueNoteVariant(comment, pick string) string {
	switch {
	case specimenWords.MatchString(comment) || specimenPick.MatchString(pick):
		return "specimen"
	case replacementWords.MatchString(comment):
		return "replacement"
	}
	return ""
}
