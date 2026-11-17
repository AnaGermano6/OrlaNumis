package services

import (
	"math"
	"testing"
)

func TestFineness(t *testing.T) {
	for composition, want := range map[string]float64{
		"Prata 925‰":          0.925,
		"Ouro 916,66‰":        0.91666,
		"Bilhão (prata 333‰)": 0.333,
		"Silver (.900)":       0.9,
		"Gold (.9999)":        0.9999,
		"Silver 0.835":        0.835,
	} {
		if got, ok := Fineness(composition); !ok || math.Abs(got-want) > 1e-9 {
			t.Errorf("Fineness(%q) = %v, %v; want %v", composition, got, ok, want)
		}
	}
	for _, composition := range []string{"Prata", "Cobre  banhado em  prata", "Ouro nórdico (89% Copper, 5% Aluminium, 5% Zinc, 1% Tin)", ""} {
		if got, ok := Fineness(composition); ok {
			t.Errorf("Fineness(%q) = %v; want none", composition, got)
		}
	}
}
