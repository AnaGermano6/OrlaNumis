package services

import (
	"testing"

	"github.com/briandenicola/ancient-coins-api/models"
)

func TestMentionedMaterial(t *testing.T) {
	cases := map[string]models.Material{
		"my gold coins":                    models.MaterialGold,
		"moedas de ouro":                   models.MaterialGold,
		"nordic gold 10 cent":              models.MaterialNordicGold,
		"moedas de ouro nórdico":           models.MaterialNordicGold,
		"silver denarius":                  models.MaterialSilver,
		"copper-nickel escudos":            models.MaterialCopperNickel,
		"escudos em cuproníquel":           models.MaterialCopperNickel,
		"bimetallic 2 euro":                models.MaterialBimetallic,
		"moedas bimetálicas":               models.MaterialBimetallic,
		"latão de níquel":                  models.MaterialNickelBrass,
		"aço revestido a cobre":            models.MaterialPlatedSteel,
		"stainless steel coins":            models.MaterialStainlessSteel,
		"moedas de alumínio":               models.MaterialAluminium,
		"bronze coins":                     models.MaterialBronze,
		"moedas de cobre":                  models.MaterialCopper,
		"the treasure of 1940 (tesouro)":   "",
		"coins from portugal without year": "",
	}
	for query, want := range cases {
		got, ok := mentionedMaterial(query)
		if got != want || ok != (want != "") {
			t.Errorf("mentionedMaterial(%q) = %q, %v; want %q", query, got, ok, want)
		}
	}
}
