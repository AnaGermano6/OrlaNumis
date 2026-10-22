package services

import (
	"testing"

	"github.com/briandenicola/ancient-coins-api/models"
)

func TestMaterialFromComposition(t *testing.T) {
	cases := []struct {
		composition string
		want        models.Material
	}{
		// As Numista writes them in Portuguese.
		{"Cuproníquel", models.MaterialCopperNickel},
		{"Cuproníquel (75% copper, 25% nickel)", models.MaterialCopperNickel},
		{"Cuproníquel (61 % Cu, 20 % Zn, 19 % Ni)", models.MaterialCopperNickel},
		{"Cuproníquel (75% Copper &amp; 25% Nickel)", models.MaterialCopperNickel},
		{"Cuproníquel  banhado em  níquel (Cu75%Ni25% Ni-plated)", models.MaterialCopperNickel},
		{"Bimetálica: núcleo de níquel revestido de latão de níquel em anel de cuproníquel", models.MaterialBimetallic},
		{"Bimetálica: núcleo de bronzital em anel de acmonital", models.MaterialBimetallic},
		{"Latão de níquel (61% Copper, 20% Zinc, 19% Nickel)", models.MaterialNickelBrass},
		{"Latão de níquel (Copper-Nickel-Zinc)", models.MaterialNickelBrass},
		{"Latão", models.MaterialBrass},
		{"Ouro nórdico (89% Copper, 5% Aluminium, 5% Zinc, 1% Tin)", models.MaterialNordicGold},
		{"Bronze alumínio (92% copper, 6% aluminium, 2% nickel)", models.MaterialAluminiumBronze},
		{"Alumínio-níquel-bronze (92% Copper, 6% Aluminium, 2% Nickel)", models.MaterialAluminiumBronze},
		{"Bronzital (92% Copper, 6% Aluminium, 2% Nickel)", models.MaterialAluminiumBronze},
		{"Bronze (Cu 950 + Sn 30 + Zn 20)", models.MaterialBronze},
		{"Bronze manganês", models.MaterialBronze},
		{"Cobre", models.MaterialCopper},
		{"Cobre  banhado em  prata", models.MaterialCopper},
		{"Cobre revestido de cuproníquel (91.67% Copper, 8.33% Nickel)", models.MaterialCopper},
		{"Níquel", models.MaterialNickel},
		{"Níquel revestido de cuproníquel", models.MaterialNickel},
		{"Alumínio (97.6% Aluminium, 2% Magnesium, 0.4% Manganese)", models.MaterialAluminium},
		{"Zinco  banhado em  cobre (97.5% Zinc, 2.5% Copper)", models.MaterialZinc},
		{"Aço inoxidável", models.MaterialStainlessSteel},
		{"Acmonital (81.75% Iron, 18.25% Chromium)", models.MaterialStainlessSteel},
		{"Aço  banhado em  cobre (Magnetic)", models.MaterialPlatedSteel},
		{"Aço revestido de níquel", models.MaterialPlatedSteel},
		{"Aço  banhado em  latão (Plating: 72.5% Copper, 27.5% Zinc)", models.MaterialPlatedSteel},
		{"Aço  banhado em  bronze", models.MaterialPlatedSteel},
		{"Bilhão (prata 333‰)", models.MaterialBillon},
		{"Prata 925‰", models.MaterialSilver},
		{"Ouro 916,66‰", models.MaterialGold},
		// And in English.
		{"Copper-nickel", models.MaterialCopperNickel},
		{"Nickel brass", models.MaterialNickelBrass},
		{"Nickel silver", models.MaterialNickelBrass},
		{"Nordic gold", models.MaterialNordicGold},
		{"Aluminium-bronze", models.MaterialAluminiumBronze},
		{"Bimetallic: copper-nickel centre in nickel-brass ring", models.MaterialBimetallic},
		{"Copper plated steel", models.MaterialPlatedSteel},
		{"Nickel-clad steel", models.MaterialPlatedSteel},
		{"Copper plated zinc", models.MaterialZinc},
		{"Silver (.900)", models.MaterialSilver},
		{"Gold (.917)", models.MaterialGold},
		{"Billon (.250 silver)", models.MaterialBillon},
		{"Electrum", models.MaterialElectrum},
		{"Brass", models.MaterialBrass},
		{"Stainless steel", models.MaterialStainlessSteel},
		{"Zinc", models.MaterialZinc},
		// What fits none of the app's materials.
		{"Aço", models.MaterialOther},
		{"Iron", models.MaterialOther},
		{"Plástico", models.MaterialOther},
		{"", models.MaterialOther},
	}
	for _, tc := range cases {
		if got := MaterialFromComposition(tc.composition); got != tc.want {
			t.Errorf("MaterialFromComposition(%q) = %q, want %q", tc.composition, got, tc.want)
		}
	}
}
