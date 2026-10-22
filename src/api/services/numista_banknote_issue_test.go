package services

import (
	"context"
	"errors"
	"testing"

	"github.com/briandenicola/ancient-coins-api/models"
)

// Shaped like a real /types/{id}/issues response for a banknote; values are made up.
const testBanknoteIssues = `[
	{"id": 1, "is_dated": true, "year": 1978, "gregorian_year": 1978, "month": 9, "day": 13,
		"signatures": [{"signer_name": " José da Silva Lopes ", "title": "Governor", "picture": "https://en.numista.com/g.jpg"}, {"signer_name": "Abel Reis", "title": "Administrator"}, {"signer_name": " "}],
		"mintage": 29645000, "comment": "P#176a; Silva# 68 - Signatures: José da Silva Lopes & Abel Reis"},
	{"id": 2, "is_dated": true, "year": 1960, "gregorian_year": 1960, "month": 3, "day": 1,
		"references": [{"catalogue": {"id": 1186, "code": "AI Notes"}, "number": "C1"}, {"catalogue": {"id": 931, "code": "P"}, "number": "13s"}], "comment": "Specimen"},
	{"id": 3, "is_dated": true, "year": 1986, "gregorian_year": 1986, "comment": "1986-06-16 Replacement note with serial # ZA-ZB-ZC"},
	{"id": 4, "is_dated": true, "year": 1986, "gregorian_year": 1986, "month": 13, "day": 40, "comment": "P# 190b; O Governador: António de Sousa"},
	{"id": 5, "is_dated": true, "year": 1980, "gregorian_year": 1980, "comment": "Color trial. Blue and green on multicolor underprint."}
]`

func TestParseNumistaIssues_BanknoteDetails(t *testing.T) {
	issues, err := parseNumistaIssues([]byte(testBanknoteIssues))
	if err != nil || len(issues) != 5 {
		t.Fatalf("issues = %+v (%v)", issues, err)
	}
	first := issues[0]
	if first.Month != 9 || first.Day != 13 || first.Pick != "176a" || first.Variant != "" {
		t.Errorf("first = %+v", first)
	}
	if len(first.Signatures) != 2 || first.Signatures[0] != (models.NumistaIssueSignature{Name: "José da Silva Lopes", Title: "Governor"}) || first.Signatures[1].Name != "Abel Reis" {
		t.Errorf("signatures = %+v", first.Signatures)
	}
	// The issue's references come before its comment; an "s" Pick is a specimen.
	if issues[1].Pick != "13s" || issues[1].Variant != "specimen" {
		t.Errorf("specimen = %+v", issues[1])
	}
	if issues[2].Pick != "" || issues[2].Variant != "replacement" || issues[2].Month != 0 {
		t.Errorf("replacement = %+v", issues[2])
	}
	// A date Numista got wrong is left out; "P# 190b" is read all the same.
	if issues[3].Month != 0 || issues[3].Day != 0 || issues[3].Pick != "190b" {
		t.Errorf("odd date = %+v", issues[3])
	}
	if issues[4].Variant != "" || issues[4].Pick != "" {
		t.Errorf("a colour trial is no variant the app knows: %+v", issues[4])
	}
}

func TestIssuePickNumber_OnlyNumistasPickReferences(t *testing.T) {
	for comment, want := range map[string]string{
		"P#176a; Silva# 068":       "176a",
		"p # 129b":                 "129b",
		"PMLC# 631; TBB# B425":     "",
		"Printed by De La Rue (P)": "",
		"":                         "",
	} {
		if got := issuePickNumber(nil, comment); got != want {
			t.Errorf("issuePickNumber(%q) = %q, want %q", comment, got, want)
		}
	}
}

func TestNumistaCatalog_IssuesFromTheCacheWithoutPrices(t *testing.T) {
	catalog, _, client, _ := newCatalogFixture()
	client.issues = testBanknoteIssues

	got := catalog.Issues(context.Background(), 202243)
	if got.Status != models.NumistaDataStatusOK || len(got.Issues) != 5 || got.Issues[0].Pick != "176a" {
		t.Fatalf("issues = %+v", got)
	}
	again := catalog.Issues(context.Background(), 202243)
	if len(again.Issues) != 5 || client.issueCalls != 1 || client.priceCalls != 0 {
		t.Fatalf("second read: %d issue calls, %d price calls", client.issueCalls, client.priceCalls)
	}

	client.issues = `[]`
	if got := catalog.Issues(context.Background(), 7); got.Status != models.NumistaDataStatusNoIssues || got.Issues == nil {
		t.Fatalf("no issues = %+v", got)
	}
	client.issuesErr = &NumistaError{Kind: NumistaErrorQuotaLimited}
	if got := catalog.Issues(context.Background(), 8); got.Status != models.NumistaDataStatusQuotaLimited || len(got.Issues) != 0 {
		t.Fatalf("quota = %+v", got)
	}
	client.issuesErr = errors.New("offline")
	if got := catalog.Issues(context.Background(), 9); got.Status == models.NumistaDataStatusOK {
		t.Fatalf("offline = %+v", got)
	}
}

// A coin's own issue is the one valued, whatever its date says.
func TestNumistaMarketValue_UsesTheCoinsOwnIssue(t *testing.T) {
	svc, client, _, _ := newMarketValueFixture()
	coin := linkedCoin("1968")
	coin.NumistaIssueID = 101

	if got := svc.ForCoin(context.Background(), coin, 0, "EUR", ""); got.SelectedIssueID != 101 || client.lastIssueID != 101 {
		t.Fatalf("selected %d, requested %d; want the coin's issue 101", got.SelectedIssueID, client.lastIssueID)
	}
	if got := svc.ForCoin(context.Background(), coin, 102, "EUR", ""); got.SelectedIssueID != 102 {
		t.Fatalf("an issue asked for wins: %d", got.SelectedIssueID)
	}
	// An issue of another type (the link changed) falls back to the date.
	coin.NumistaIssueID = 999
	if got := svc.ForCoin(context.Background(), coin, 0, "EUR", ""); got.SelectedIssueID != 102 {
		t.Fatalf("stale issue: %d", got.SelectedIssueID)
	}
}
