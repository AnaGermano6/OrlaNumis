package services

import (
	"html"
	"regexp"
	"strings"

	"github.com/briandenicola/ancient-coins-api/models"
)

// Numista writes a type's composition in the language asked for, with the
// alloy in brackets: "Cuproníquel (75% copper, 25% nickel)", "Bimetálica:
// núcleo de níquel revestido de latão de níquel em anel de cuproníquel",
// "Aço banhado em cobre", "Copper plated steel", "Silver (.925)".
var (
	// A coin of two or three metals counts as bimetallic, whatever they are.
	compositionMultiMetal = regexp.MustCompile(`\b(?:bi|tri)-?metal`)
	// "Aço banhado em cobre", "Cobre revestido de cuproníquel": the core comes first.
	compositionPlatedPt = regexp.MustCompile(`^(.+?) (?:banhad[oa]s?|revestid[oa]s?|chapead[oa]s?)\b`)
	// "Copper plated steel", "Nickel-clad steel": the core comes last.
	compositionPlatedEn = regexp.MustCompile(`^.+?[- ](?:plated|clad|coated) (.+)$`)
	compositionSteel    = regexp.MustCompile(`\baco\b|\bsteel\b|\bferro\b|\biron\b`)
)

// compositionMaterials are tried in order: alloys before the metals they
// are named after ("ouro nórdico" is not gold, "latão de níquel" is not
// nickel, "nickel silver" is not silver).
var compositionMaterials = []struct {
	material models.Material
	pattern  *regexp.Regexp
}{
	{models.MaterialNordicGold, regexp.MustCompile(`ouro nordico|nordic gold`)},
	{models.MaterialAluminiumBronze, regexp.MustCompile(`alumini?(?:o|um)[- ](?:(?:niquel|nickel)[- ])?bronze|bronze[- ]alumini?(?:o|um)|aluminum[- ]bronze|bronzital`)},
	{models.MaterialStainlessSteel, regexp.MustCompile(`inoxidavel|stainless|acmonital`)},
	{models.MaterialNickelBrass, regexp.MustCompile(`latao de niquel|nickel[- ]brass|nickel silver|german silver|alpaca|maillechort|argentan`)},
	{models.MaterialCopperNickel, regexp.MustCompile(`cupro[- ]?(?:niquel|nickel)|(?:copper|cobre)[- ](?:nickel|niquel)`)},
	{models.MaterialBillon, regexp.MustCompile(`\bbilhao\b|\bbillon\b`)},
	{models.MaterialGold, regexp.MustCompile(`\bouro\b|\bgold\b`)},
	{models.MaterialSilver, regexp.MustCompile(`\bprata\b|\bsilver\b`)},
	{models.MaterialElectrum, regexp.MustCompile(`electrum|\beletro\b|\belectro\b`)},
	{models.MaterialBronze, regexp.MustCompile(`bronze`)},
	{models.MaterialBrass, regexp.MustCompile(`\blatao\b|\bbrass\b|\btombac`)},
	{models.MaterialNickel, regexp.MustCompile(`\bniquel\b|\bnickel\b`)},
	{models.MaterialAluminium, regexp.MustCompile(`\balumini?(?:o|um)\b|\baluminum\b`)},
	{models.MaterialZinc, regexp.MustCompile(`\bzinco?\b`)},
	{models.MaterialCopper, regexp.MustCompile(`\bcobre\b|\bcopper\b`)},
}

// MaterialFromComposition maps a Numista composition, in Portuguese or
// English, to the app's material; Other when it fits none. A plated or clad
// coin counts by its core, a plated steel one as plated steel.
func MaterialFromComposition(composition string) models.Material {
	text := strings.Join(strings.Fields(foldASCII(html.UnescapeString(composition))), " ")
	// The alloy in brackets names other metals ("Cuproníquel (75% copper").
	if cut, _, found := strings.Cut(text, "("); found && strings.TrimSpace(cut) != "" {
		text = strings.TrimSpace(cut)
	}
	if text == "" {
		return models.MaterialOther
	}
	if compositionMultiMetal.MatchString(text) {
		return models.MaterialBimetallic
	}
	core := ""
	if match := compositionPlatedPt.FindStringSubmatch(text); match != nil {
		core = match[1]
	} else if match := compositionPlatedEn.FindStringSubmatch(text); match != nil {
		core = match[1]
	}
	if core != "" {
		if compositionSteel.MatchString(core) && !strings.Contains(core, "inox") && !strings.Contains(core, "stainless") {
			return models.MaterialPlatedSteel
		}
		text = core
	}
	for _, candidate := range compositionMaterials {
		if candidate.pattern.MatchString(text) {
			return candidate.material
		}
	}
	return models.MaterialOther
}
