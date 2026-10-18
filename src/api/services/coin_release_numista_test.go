package services

import (
	"context"
	"testing"
	"time"

	"github.com/briandenicola/ancient-coins-api/models"
)

func candidateIDs(items []models.NumistaCatalogSearchItem) []int {
	ids := []int{}
	for _, item := range items {
		ids = append(ids, item.ID)
	}
	return ids
}

// A coin from Numista is its own type; a mint's coin the catalogue has
// matched to Numista is known by that type, without asking Numista again.
func TestReleaseNumistaType_KnownWithoutAskingNumista(t *testing.T) {
	client := &escudoNumista{values: map[string]string{
		"portugal|5 euro": `{"count": 2, "types": [
			{"id": 601, "title": "5 Euros (Faróis de Portugal - Cabo de São Vicente)", "issuer": {"code": "portugal"}, "min_year": 2025, "max_year": 2025},
			{"id": 602, "title": "5 Euros (Mosteiro dos Jerónimos)", "issuer": {"code": "portugal"}, "min_year": 2004, "max_year": 2004}]}`,
	}}
	portugal := fakeMint{country: "Portugal", source: "incm", coins: []mintCoin{
		{Title: "Faróis de Portugal - Cabo de São Vicente (Prata Proof)", Value: 5, Year: 2025, Released: "2025-11", PageURL: "https://loja.incm.pt/products/farol", Obverse: "https://cdn/farol_A.jpg"},
	}}
	releases := newEscudoCatalogue(t, client, portugal)
	releases.now = func() time.Time { return time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC) }
	ctx := context.Background()

	escudo := releases.NumistaType(ctx, CoinRelease{Source: CoinReleaseSourceNumista, Country: "PT", Title: "Expo 98", FaceValue: 1000, Currency: "PTE", From: "1998-01", PageURL: "https://pt.numista.com/catalogue/pieces9002.html"})
	if escudo.Status != models.NumistaDataStatusOK || escudo.TypeID != 9002 || client.calls != 0 {
		t.Fatalf("a Numista coin = %+v after %d searches", escudo, client.calls)
	}

	releases.Catalogue(ctx)
	asked := client.calls
	farol := releases.NumistaType(ctx, CoinRelease{Source: "incm", Country: "PT", Title: "Faróis de Portugal - Cabo de São Vicente", FaceValue: 5, From: "2025-11", PageURL: "https://loja.incm.pt/products/farol"})
	if farol.TypeID != 601 || client.calls != asked {
		t.Fatalf("the INCM's lighthouse = %+v after %d more searches", farol, client.calls-asked)
	}
}

// Otherwise Numista's coins of the country, year and face value are the
// candidates, and the coin is the only one whose name it shares.
func TestReleaseNumistaType_ByNameAmongTheYearsCoins(t *testing.T) {
	client := &escudoNumista{values: map[string]string{
		"germany|2 euro": `{"count": 4, "types": [
			{"id": 901, "title": "2 Euros (Treaty of Rome)", "issuer": {"code": "germany"}, "min_year": 2007, "max_year": 2007},
			{"id": 902, "title": "2 Euros (Mecklenburg-Vorpommern)", "issuer": {"code": "germany"}, "min_year": 2007, "max_year": 2007},
			{"id": 903, "title": "2 Euros (2nd map)", "issuer": {"code": "germany"}, "min_year": 2007, "max_year": 2025},
			{"id": 904, "title": "2 Euro Cents", "issuer": {"code": "germany"}, "min_year": 2002, "max_year": 2025}]}`,
		"portugal|5 euro": `{"count": 2, "types": [
			{"id": 801, "title": "5 Euro (The knowledge, cupronickel)", "issuer": {"code": "portugal"}, "min_year": 2024, "max_year": 2024},
			{"id": 802, "title": "5 Euro (The knowledge, Silver)", "issuer": {"code": "portugal"}, "min_year": 2024, "max_year": 2024}]}`,
	}}
	releases := newEscudoCatalogue(t, client)
	ctx := context.Background()

	rome := releases.NumistaType(ctx, CoinRelease{Source: CoinReleaseSourceECB, Country: "DE", Title: "50th anniversary of the Treaty of Rome", FaceValue: 2, From: "2007-03"})
	if rome.Status != models.NumistaDataStatusOK || rome.TypeID != 901 {
		t.Fatalf("Treaty of Rome = %+v", rome)
	}
	if ids := candidateIDs(rome.Candidates); len(ids) != 3 || ids[0] != 901 || ids[2] != 903 {
		t.Fatalf("candidates = %v; the cent coin is not one", ids)
	}

	// Two versions of the coin share its name: the user chooses.
	knowledge := releases.NumistaType(ctx, CoinRelease{Source: "incm", Country: "PT", Title: "The knowledge", FaceValue: 5, From: "2024-11"})
	if knowledge.TypeID != 0 || len(knowledge.Candidates) != 2 {
		t.Fatalf("The knowledge = %+v", knowledge)
	}
	// A name the coins do not share: candidates only.
	other := releases.NumistaType(ctx, CoinRelease{Source: CoinReleaseSourceECB, Country: "DE", Title: "Hamburg", FaceValue: 2, From: "2007-01"})
	if other.TypeID != 0 || len(other.Candidates) != 3 {
		t.Fatalf("an unknown coin = %+v", other)
	}
}

func TestReleaseNumistaType_WithoutACountryOrNumista(t *testing.T) {
	client := &escudoNumista{failFrom: 2000}
	releases := newEscudoCatalogue(t, client)
	ctx := context.Background()

	joint := releases.NumistaType(ctx, CoinRelease{Source: CoinReleaseSourceECB, Title: "Ten years of the euro", FaceValue: 2, From: "2012-01"})
	if joint.Status != models.NumistaDataStatusOK || joint.TypeID != 0 || len(joint.Candidates) != 0 || client.calls != 0 {
		t.Fatalf("a joint issue = %+v after %d searches", joint, client.calls)
	}
	refused := releases.NumistaType(ctx, CoinRelease{Source: CoinReleaseSourceECB, Country: "PT", Title: "Rotary", FaceValue: 2, From: "2026-07"})
	if refused.Status == models.NumistaDataStatusOK || refused.TypeID != 0 || len(refused.Candidates) != 0 {
		t.Fatalf("Numista out of quota = %+v", refused)
	}
	if releases := NewCoinReleasesService(nil).NumistaType(ctx, CoinRelease{Source: CoinReleaseSourceECB, Country: "PT", Title: "Rotary", FaceValue: 2, From: "2026-07"}); releases.Status != models.NumistaDataStatusUnconfigured {
		t.Fatalf("without Numista = %+v", releases)
	}
}

func TestNumistaPageTypeID(t *testing.T) {
	for page, want := range map[string]int{
		"https://pt.numista.com/catalogue/pieces1234.html":        1234,
		"https://en.numista.com/catalogue/pieces7.html":           7,
		"https://numista.com/catalogue/pieces8.html":              8,
		"https://evil.example/numista.com/catalogue/pieces9.html": 0,
		"https://loja.incm.pt/products/farol":                     0,
	} {
		if got := numistaPageTypeID(page); got != want {
			t.Errorf("numistaPageTypeID(%q) = %d, want %d", page, got, want)
		}
	}
}
