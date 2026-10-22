package services

import (
	"archive/zip"
	"bytes"
	"encoding/csv"
	"encoding/xml"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/briandenicola/ancient-coins-api/models"
)

func exportCoins() []models.Coin {
	weight, value, price := 7.5, 12.5, 4.0
	year := 1986
	bought := time.Date(2024, 3, 9, 15, 0, 0, 0, time.UTC)
	return []models.Coin{
		{
			ID: 7, Name: "1 Escudo 1986 & Açores", Country: "Portugal", MonetarySystem: "Escudo (1911-2001)", Denomination: "1 Escudo",
			Year: &year, Material: models.MaterialNickelBrass, WeightGrams: &weight, Grade: "UNC", Presentation: "coincard",
			Rarity: "scarce", Quantity: 2, PurchasePrice: &price, PurchaseDate: &bought, CurrentValue: &value,
			References: []models.CoinReference{{Catalog: "KM", Number: "631"}, {Catalog: "Numista", Number: "1234"}},
			Sets:       []models.CoinSet{{Name: "Escudos"}, {Name: "Açores"}},
			Notes:      "=SUM(A1)",
		},
		{ID: 8, Name: "100 Escudos", Kind: models.CoinKindBanknote, Country: "Angola", Material: "Paper", SerialNumber: "AB123", Quantity: 1},
	}
}

func TestBuildCollectionExport_Columns(t *testing.T) {
	export := BuildCollectionExport(exportCoins(), "pt", "EUR", time.Date(2026, 10, 21, 0, 0, 0, 0, time.UTC))
	header := strings.Join(export.Columns, "|")
	for _, want := range []string{"Sistema monetário", "Preço de compra (EUR)", "Valor atual (EUR)", "Número de série", "Notas"} {
		if !strings.Contains(header, want) {
			t.Errorf("no %q column in %s", want, header)
		}
	}
	if strings.Contains(header, "Lista de desejos") {
		t.Error("the wishlist column is there only when a wished coin is")
	}
	if export.FileName("xlsx") != "colecao-2026-10-21.xlsx" || export.Sheet != "Coleção" {
		t.Fatalf("name = %s, sheet = %s", export.FileName("xlsx"), export.Sheet)
	}
	at := func(row int, column string) ExportCell {
		for i, name := range export.Columns {
			if name == column {
				return export.Rows[row][i]
			}
		}
		t.Fatalf("no column %s", column)
		return ExportCell{}
	}
	if got := at(0, "Material"); got.text != "Latão de níquel" {
		t.Errorf("material = %+v", got)
	}
	if at(0, "Apresentação").text != "Coincard" || at(0, "Raridade").text != "Escassa" || at(0, "Conjuntos").text != "Escudos, Açores" {
		t.Errorf("labels: %+v %+v %+v", at(0, "Apresentação"), at(0, "Raridade"), at(0, "Conjuntos"))
	}
	if got := at(0, "Peso (g)"); got.kind != exportNumber || got.number != 7.5 {
		t.Errorf("weight = %+v", got)
	}
	if got := at(0, "Data de compra"); got.kind != exportDate || got.date.Day() != 9 {
		t.Errorf("purchase date = %+v", got)
	}
	if at(0, "Numista").text != "https://en.numista.com/catalogue/pieces1234.html" || at(0, "Referências").text != "KM 631 · Numista 1234" {
		t.Errorf("references: %+v %+v", at(0, "Numista"), at(0, "Referências"))
	}
	if at(1, "Tipo").text != "Nota" || at(1, "Material").text != "Papel" || at(1, "Número de série").text != "AB123" || at(1, "Valor atual (EUR)").kind != exportEmpty {
		t.Errorf("banknote row: %+v", export.Rows[1])
	}

	coinsOnly := BuildCollectionExport(exportCoins()[:1], "en", "USD", time.Now())
	if header := strings.Join(coinsOnly.Columns, "|"); strings.Contains(header, "Serial number") || !strings.Contains(header, "Current value (USD)") {
		t.Errorf("coins only, in English: %s", header)
	}
	wished := exportCoins()[:1]
	wished[0].IsWishlist = true
	if header := strings.Join(BuildCollectionExport(wished, "en", "USD", time.Now()).Columns, "|"); !strings.Contains(header, "Wishlist") {
		t.Errorf("a wished coin brings the wishlist column: %s", header)
	}
}

func TestCollectionExport_CSV(t *testing.T) {
	export := BuildCollectionExport(exportCoins(), "pt", "EUR", time.Now())
	body, err := export.CSV()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.HasPrefix(body, []byte("\xef\xbb\xbf")) {
		t.Fatal("no byte order mark")
	}
	reader := csv.NewReader(bytes.NewReader(bytes.TrimPrefix(body, []byte("\xef\xbb\xbf"))))
	reader.Comma = ';'
	rows, err := reader.ReadAll()
	if err != nil || len(rows) != 3 {
		t.Fatalf("rows = %v (%v)", rows, err)
	}
	line := strings.Join(rows[1], "|")
	for _, want := range []string{"|7,5|", "|12,50|", "|4,00|", "|2024-03-09|", "|1 Escudo 1986 & Açores|", "|'=SUM(A1)"} {
		if !strings.Contains(line+"|", want) {
			t.Errorf("no %q in %s", want, line)
		}
	}
	english, _ := BuildCollectionExport(exportCoins(), "en", "EUR", time.Now()).CSV()
	if !strings.Contains(string(english), ",7.5,") {
		t.Errorf("english csv uses commas and points: %s", english)
	}
}

func TestCollectionExport_XLSX(t *testing.T) {
	export := BuildCollectionExport(exportCoins(), "pt", "EUR", time.Now())
	body, err := export.XLSX()
	if err != nil {
		t.Fatal(err)
	}
	archive, err := zip.NewReader(bytes.NewReader(body), int64(len(body)))
	if err != nil {
		t.Fatal(err)
	}
	files := map[string]string{}
	for _, file := range archive.File {
		reader, _ := file.Open()
		content, _ := io.ReadAll(reader)
		reader.Close()
		files[file.Name] = string(content)
	}
	for _, name := range []string{"[Content_Types].xml", "_rels/.rels", "xl/workbook.xml", "xl/_rels/workbook.xml.rels", "xl/styles.xml", "xl/worksheets/sheet1.xml"} {
		content, ok := files[name]
		if !ok {
			t.Fatalf("no %s", name)
		}
		decoder := xml.NewDecoder(strings.NewReader(content))
		for {
			if _, err := decoder.Token(); err == io.EOF {
				break
			} else if err != nil {
				t.Fatalf("%s is not well-formed XML: %v", name, err)
			}
		}
	}
	sheet := files["xl/worksheets/sheet1.xml"]
	for _, want := range []string{
		`<c r="A1" t="inlineStr" s="1"><is><t>ID</t></is></c>`,
		`1 Escudo 1986 &amp; Açores`,
		`<c r="A2" s="3"><v>7</v></c>`,
		`state="frozen"`,
		`<autoFilter ref="A1:`,
	} {
		if !strings.Contains(sheet, want) {
			t.Errorf("no %q in the sheet", want)
		}
	}
	if !strings.Contains(files["xl/workbook.xml"], `<sheet name="Coleção"`) || !strings.Contains(files["xl/styles.xml"], `formatCode="dd/mm/yyyy"`) {
		t.Error("the Portuguese sheet name and date format")
	}
}

func TestSpreadsheetHelpers(t *testing.T) {
	for index, want := range map[int]string{1: "A", 26: "Z", 27: "AA", 52: "AZ", 53: "BA", 703: "AAA"} {
		if got := columnName(index); got != want {
			t.Errorf("columnName(%d) = %s, want %s", index, got, want)
		}
	}
	if got := excelSerial(time.Date(2000, 1, 1, 18, 0, 0, 0, time.UTC)); got != 36526 {
		t.Errorf("excelSerial(2000-01-01) = %v", got)
	}
	if got := xmlText("a\x00b<c>\"d\"&"); got != "ab&lt;c&gt;&quot;d&quot;&amp;" {
		t.Errorf("xmlText = %q", got)
	}
	if ExportLang("pt-PT") != "pt" || ExportLang("en-GB") != "en" || ExportLang("") != "pt" {
		t.Error("ExportLang")
	}
}
