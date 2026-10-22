package services

import (
	"regexp"
	"strings"

	"github.com/briandenicola/ancient-coins-api/models"
)

// Words and abbreviations collectors write in the free-text grade field,
// checked from the highest grade down so "Good VF" reads as VF and "About
// Unc" as AU; the same as the web's numistaGradeForCoinGrade. Portuguese terms
// follow the usual equivalences (REG≈G, BC≈F, MBC≈VF, Bela≈XF, Soberba/FDC≈UNC).
// Proofs have no price on Numista's grade scale and match none.
var coinGradePatterns = []struct {
	grade   string
	pattern *regexp.Regexp
}{
	{"au", regexp.MustCompile(`\b(ABOUT|ALMOST) UNC|\b(A|CH )?AU\b|\bAUNC\b`)},
	{"unc", regexp.MustCompile(`\b(UNC|UNCIRCULATED|MS|BU|BNC|FDC|SOB|SOBERBA|SOBERBO|FLOR DE CUNHO)\b`)},
	{"xf", regexp.MustCompile(`\bEXTREMELY FINE\b|\b(A|G|CH )?(XF|EF)\b|\bBEL[AO]\b`)},
	{"vf", regexp.MustCompile(`\bVERY FINE\b|\b(A|G|CH )?VF\b|\bMBC\b`)},
	{"vg", regexp.MustCompile(`\bVERY GOOD\b|\b(A|G)?VG\b`)},
	{"f", regexp.MustCompile(`\bFINE\b|\b(A|G|CH )?F\b|\bBC\b`)},
	{"g", regexp.MustCompile(`\bGOOD\b|\b(A)?G\b|\bREG(ULAR)?\b|\bMC\b`)},
}

var gradeNonAlnum = regexp.MustCompile(`[^A-Z0-9]+`)

// NumistaGradeForCoinGrade maps a coin's grade ("UNC", "Ch XF", "MBC",
// "MS 63") to Numista's price scale (g, vg, f, vf, xf, au, unc); false when
// it fits none (no grade, or a proof).
func NumistaGradeForCoinGrade(grade string) (string, bool) {
	normalized := strings.TrimSpace(gradeNonAlnum.ReplaceAllString(strings.ToUpper(foldASCII(grade)), " "))
	if normalized == "" {
		return "", false
	}
	for _, candidate := range coinGradePatterns {
		if candidate.pattern.MatchString(normalized) {
			return candidate.grade, true
		}
	}
	return "", false
}

// numistaPriceForGrade is Numista's price for a grade: the grade's own, else
// the nearest lower grade's, else the nearest higher one's; false when there
// is no price at all.
func numistaPriceForGrade(prices []models.NumistaGradePrice, grade string) (float64, bool) {
	byRank := map[int]float64{}
	for _, price := range prices {
		if rank, ok := numistaGradeOrder[price.Grade]; ok && price.Price > 0 {
			byRank[rank] = price.Price
		}
	}
	want, ok := numistaGradeOrder[grade]
	if !ok || len(byRank) == 0 {
		return 0, false
	}
	for rank := want; rank >= 0; rank-- {
		if price, ok := byRank[rank]; ok {
			return price, true
		}
	}
	for rank := want + 1; rank < len(numistaGradeOrder); rank++ {
		if price, ok := byRank[rank]; ok {
			return price, true
		}
	}
	return 0, false
}
