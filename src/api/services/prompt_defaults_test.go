package services

import (
	"strings"
	"testing"

	"github.com/briandenicola/ancient-coins-api/models"
)

// A prompt saved as it comes (the admin page saves every field it shows) is
// stored empty, so that its default, even a later one, applies; an edited
// one is kept.
func TestSetSettingStoresADefaultPromptEmpty(t *testing.T) {
	svc, _ := newTestSettingsService(t)

	for key, prompt := range defaultPrompts() {
		if err := svc.SetSetting(key, "\n  "+strings.ReplaceAll(prompt, "\n", "\r\n")+"  \n"); err != nil {
			t.Fatal(err)
		}
		setting, err := svc.repo.FindByKey(key)
		if err != nil {
			t.Fatal(err)
		}
		if setting.Value != "" {
			t.Fatalf("%s stored as %q", key, setting.Value[:min(40, len(setting.Value))])
		}
	}
	if svc.GetCoinSearchPrompt() != DefaultCoinSearchPrompt || svc.GetSetting(SettingObversePrompt) != DefaultObversePrompt {
		t.Fatal("an empty prompt uses its default")
	}

	if err := svc.SetSetting(SettingCoinSearchPrompt, "Procura só moedas de Portugal."); err != nil {
		t.Fatal(err)
	}
	if svc.GetCoinSearchPrompt() != "Procura só moedas de Portugal." {
		t.Fatalf("an edited prompt is kept: %q", svc.GetCoinSearchPrompt())
	}
}

// The default prompts speak of this collection: modern coins and banknotes,
// in European Portuguese; the coin shows' has no braces, as the agent fills
// in its location with Python's format.
func TestDefaultPromptsFitTheCollection(t *testing.T) {
	for key, prompt := range defaultPrompts() {
		if strings.Contains(strings.ToLower(prompt), "roman") || strings.Contains(strings.ToLower(prompt), "greek") {
			t.Errorf("%s still speaks of ancient coins", key)
		}
	}
	for _, prompt := range []string{DefaultCoinSearchPrompt, DefaultCoinShowsPrompt, DefaultObversePrompt, DefaultReversePrompt} {
		if !strings.Contains(prompt, "português europeu") {
			t.Errorf("not asked in European Portuguese: %.60s", prompt)
		}
	}
	if strings.ContainsAny(DefaultCoinShowsPrompt, "{}") {
		t.Error("the coin shows prompt has braces")
	}
	// The valuation's answer keeps the fields the parser reads.
	for _, field := range []string{"estimatedValue", "confidence", "reasoning", "changeExplanation", "comparables", "```json"} {
		if !strings.Contains(DefaultValuationPrompt, field) {
			t.Errorf("the valuation prompt lost %s", field)
		}
	}
	if estimate := ParseValueEstimate(DefaultValuationPrompt[strings.Index(DefaultValuationPrompt, "```json\n{"):]); estimate.EstimatedValue != 12.5 || estimate.Confidence != "high" {
		t.Errorf("the valuation prompt's example does not parse: %+v", estimate)
	}
}

// The valuation knows what the item is, where and when it is from, and how it is kept.
func TestBuildCoinDescriptionSaysTheCountryYearAndPresentation(t *testing.T) {
	year := 1980
	description := BuildCoinDescription(&models.Coin{
		Kind: models.CoinKindBanknote, Name: "50 Escudos (9th. print) 1980 – Portugal", Country: "Portugal", Year: &year,
		Denomination: "50 Escudos", Grade: "XF", Presentation: "loose",
	}, CurrencyEUR)
	for _, line := range []string{"Item: banknote", "Country: Portugal", "Year: 1980", "Grade/Condition: XF", "Presentation: loose"} {
		if !strings.Contains(description, line) {
			t.Errorf("no %q in:\n%s", line, description)
		}
	}
	if strings.Contains(BuildCoinDescription(&models.Coin{Kind: models.CoinKindCoin, Name: "1 escudo", DateRange: "c. 1930"}, CurrencyEUR), "Item:") {
		t.Error("a coin is not called an item")
	}
	if !strings.Contains(BuildCoinDescription(&models.Coin{Name: "1 escudo", DateRange: "c. 1930"}, CurrencyEUR), "Date: c. 1930") {
		t.Error("a date range stands in for the year")
	}
}
