package services

import (
	"testing"

	"github.com/briandenicola/ancient-coins-api/models"
)

func TestEuroCirculationYears(t *testing.T) {
	year := func(v int) *int { return &v }
	coins := []models.Coin{
		{Name: "2 euro", Denomination: "2 euro", Country: "Portugal", Year: year(2002)},
		{Name: "1 euro 2002 – Portugal", Denomination: "1 euro", Country: "Portugal", Year: year(2002)},
		// The same coin twice counts once.
		{Name: "2 euro", Denomination: "2 euro", Country: "Portugal", Year: year(2002)},
		// A commemorative is not one of the year's eight.
		{Name: "2 euro 2007 – Tratado de Roma", Denomination: "2 euro", Country: "Portugal", Year: year(2007)},
		// Wanted, not had.
		{Name: "50 cent", Denomination: "50 cêntimos", Country: "Portugal", Year: year(2003), IsWishlist: true},
		{Name: "10 cent", Denomination: "10 cent", Country: "Croácia", Year: year(2023)},
	}
	years := EuroCirculationYears(coins, 2025)

	portugal := years["PT"]
	if len(portugal) != 24 || portugal[0] != (EuroCirculationYear{Year: 2002, Total: 8, Owned: 2}) || portugal[1].Owned != 0 || portugal[5].Owned != 0 {
		t.Fatalf("Portugal = %+v", portugal)
	}
	if croatia := years["HR"]; len(croatia) != 3 || croatia[0] != (EuroCirculationYear{Year: 2023, Total: 8, Owned: 1}) {
		t.Fatalf("Croatia = %+v", croatia)
	}
	// Countries not yet in the euro that year have none; those since 1999 start then.
	if _, ok := years["BG"]; ok {
		t.Fatalf("Bulgaria before 2026 = %+v", years["BG"])
	}
	if belgium := years["BE"]; len(belgium) != 27 || belgium[0].Year != 1999 {
		t.Fatalf("Belgium = %+v", belgium)
	}
}
