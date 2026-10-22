package services

import (
	"archive/zip"
	"bytes"
	"encoding/csv"
	"fmt"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/briandenicola/ancient-coins-api/models"
)

// A collection exported as a spreadsheet: a row per coin or banknote, the
// columns named in the user's language, numbers and dates as such.

type exportCellKind int

const (
	exportEmpty exportCellKind = iota
	exportText
	exportNumber
	// exportMoney is a number shown with two decimals.
	exportMoney
	exportDate
)

// ExportCell is one value of the spreadsheet.
type ExportCell struct {
	kind   exportCellKind
	text   string
	number float64
	date   time.Time
}

func textCell(value string) ExportCell {
	value = strings.TrimSpace(value)
	if value == "" {
		return ExportCell{}
	}
	return ExportCell{kind: exportText, text: value}
}

func numberCell(value *float64) ExportCell {
	if value == nil {
		return ExportCell{}
	}
	return ExportCell{kind: exportNumber, number: *value}
}

func moneyCell(value *float64) ExportCell {
	if value == nil {
		return ExportCell{}
	}
	return ExportCell{kind: exportMoney, number: *value}
}

func intCell(value int) ExportCell {
	return ExportCell{kind: exportNumber, number: float64(value)}
}

func dateCell(value *time.Time) ExportCell {
	if value == nil || value.IsZero() {
		return ExportCell{}
	}
	return ExportCell{kind: exportDate, date: *value}
}

// CollectionExport is the spreadsheet: its columns and rows.
type CollectionExport struct {
	Lang    string
	Sheet   string
	Columns []string
	// Widths are the columns' widths, in characters.
	Widths []float64
	Rows   [][]ExportCell
	Date   time.Time
}

type exportLabels struct {
	sheet, file                                                                string
	id, kind, coin, banknote, name, country, system, denomination, year, dates string
	mint, material, weight, diameter, grade, presentation, rarity, quantity    string
	purchasePrice, purchaseDate, purchasedFrom, value, valueDate, references   string
	numista, sets, notes, wishlist, yes                                        string
	series, serial, signatures, printer, width, height                         string
	materials, presentations, rarities                                         map[string]string
}

var exportLabelsByLang = map[string]exportLabels{
	"pt": {
		sheet: "Coleção", file: "colecao",
		id: "ID", kind: "Tipo", coin: "Moeda", banknote: "Nota", name: "Nome", country: "País", system: "Sistema monetário",
		denomination: "Valor facial", year: "Ano", dates: "Datas", mint: "Casa da moeda", material: "Material", weight: "Peso (g)",
		diameter: "Diâmetro (mm)", grade: "Grau", presentation: "Apresentação", rarity: "Raridade", quantity: "Quantidade",
		purchasePrice: "Preço de compra (%s)", purchaseDate: "Data de compra", purchasedFrom: "Local de compra",
		value: "Valor atual (%s)", valueDate: "Valor atualizado em", references: "Referências", numista: "Numista",
		sets: "Conjuntos", notes: "Notas", wishlist: "Lista de desejos", yes: "Sim",
		series: "Série", serial: "Número de série", signatures: "Assinaturas", printer: "Impressor", width: "Largura (mm)", height: "Altura (mm)",
		materials: map[string]string{
			"Gold": "Ouro", "Silver": "Prata", "Bronze": "Bronze", "Copper": "Cobre", "Electrum": "Eletro", "Billon": "Bilhão",
			"Aluminium bronze": "Bronze-alumínio", "Brass": "Latão", "Nickel brass": "Latão de níquel", "Nordic gold": "Ouro nórdico",
			"Copper-nickel": "Cuproníquel", "Nickel": "Níquel", "Bimetallic": "Bimetálica", "Aluminium": "Alumínio", "Zinc": "Zinco",
			"Stainless steel": "Aço inoxidável", "Plated steel": "Aço revestido", "Other": "Outro",
			"Paper": "Papel", "Polymer": "Polímero", "Hybrid": "Híbrido (papel e polímero)",
		},
		presentations: map[string]string{
			"loose": "Solta", "coincard": "Coincard", "capsule": "Cápsula", "flip": "Carteira / 2x2", "proof_set": "Estojo / proof",
			"blister": "Blister", "slab": "Certificada (NGC/PCGS)",
		},
		rarities: map[string]string{
			"common": "Comum", "uncommon": "Pouco comum", "scarce": "Escassa", "rare": "Rara", "very_rare": "Muito rara", "extremely_rare": "Extremamente rara",
		},
	},
	"en": {
		sheet: "Collection", file: "collection",
		id: "ID", kind: "Kind", coin: "Coin", banknote: "Banknote", name: "Name", country: "Country", system: "Monetary system",
		denomination: "Face value", year: "Year", dates: "Dates", mint: "Mint", material: "Material", weight: "Weight (g)",
		diameter: "Diameter (mm)", grade: "Grade", presentation: "Presentation", rarity: "Rarity", quantity: "Quantity",
		purchasePrice: "Purchase price (%s)", purchaseDate: "Purchase date", purchasedFrom: "Purchased from",
		value: "Current value (%s)", valueDate: "Value updated", references: "References", numista: "Numista",
		sets: "Sets", notes: "Notes", wishlist: "Wishlist", yes: "Yes",
		series: "Series", serial: "Serial number", signatures: "Signatures", printer: "Printer", width: "Width (mm)", height: "Height (mm)",
		materials: map[string]string{"Paper": "Paper", "Polymer": "Polymer", "Hybrid": "Hybrid (paper and polymer)"},
		presentations: map[string]string{
			"loose": "Loose", "coincard": "Coincard", "capsule": "Capsule", "flip": "Flip / 2x2 holder", "proof_set": "Case / proof set",
			"blister": "Blister", "slab": "Certified slab (NGC/PCGS)",
		},
		rarities: map[string]string{
			"common": "Common", "uncommon": "Uncommon", "scarce": "Scarce", "rare": "Rare", "very_rare": "Very rare", "extremely_rare": "Extremely rare",
		},
	},
}

func exportLabelsFor(lang string) exportLabels {
	if labels, ok := exportLabelsByLang[lang]; ok {
		return labels
	}
	return exportLabelsByLang["en"]
}

// ExportLang is the export's language: "pt" (European Portuguese) or "en".
func ExportLang(value string) string {
	if value == "" || strings.HasPrefix(strings.ToLower(strings.TrimSpace(value)), "pt") {
		return "pt"
	}
	return "en"
}

// ExportFileName is the spreadsheet's name: "colecao-2026-10-21.xlsx".
func (e CollectionExport) FileName(format string) string {
	return fmt.Sprintf("%s-%s.%s", exportLabelsFor(e.Lang).file, e.Date.Format("2006-01-02"), format)
}

// BuildCollectionExport lays coins out as a spreadsheet, in the order given.
// The banknotes' own columns (series, serial number…) are there only when
// a banknote is; the wishlist column only when a wished coin is.
func BuildCollectionExport(coins []models.Coin, lang, currency string, now time.Time) CollectionExport {
	labels := exportLabelsFor(lang)
	banknotes, wished := false, false
	for i := range coins {
		banknotes = banknotes || coins[i].Kind == models.CoinKindBanknote
		wished = wished || coins[i].IsWishlist
	}
	type column struct {
		name  string
		width float64
		value func(coin *models.Coin) ExportCell
	}
	columns := []column{
		{labels.id, 7, func(coin *models.Coin) ExportCell { return intCell(int(coin.ID)) }},
		{labels.kind, 8, func(coin *models.Coin) ExportCell {
			if coin.Kind == models.CoinKindBanknote {
				return textCell(labels.banknote)
			}
			return textCell(labels.coin)
		}},
		{labels.name, 36, func(coin *models.Coin) ExportCell { return textCell(coin.Name) }},
		{labels.country, 18, func(coin *models.Coin) ExportCell { return textCell(coin.Country) }},
		{labels.system, 24, func(coin *models.Coin) ExportCell { return textCell(coin.MonetarySystem) }},
		{labels.denomination, 14, func(coin *models.Coin) ExportCell { return textCell(coin.Denomination) }},
		{labels.year, 7, func(coin *models.Coin) ExportCell {
			if coin.Year == nil {
				return ExportCell{}
			}
			return intCell(*coin.Year)
		}},
		{labels.dates, 12, func(coin *models.Coin) ExportCell { return textCell(coin.DateRange) }},
		{labels.mint, 18, func(coin *models.Coin) ExportCell { return textCell(coin.Mint) }},
		{labels.material, 16, func(coin *models.Coin) ExportCell {
			return textCell(exportLabel(labels.materials, string(coin.Material)))
		}},
		{labels.weight, 10, func(coin *models.Coin) ExportCell { return numberCell(coin.WeightGrams) }},
		{labels.diameter, 12, func(coin *models.Coin) ExportCell { return numberCell(coin.DiameterMm) }},
		{labels.grade, 9, func(coin *models.Coin) ExportCell {
			if strings.TrimSpace(coin.Grade) == "" {
				return textCell(coin.CertifiedGrade)
			}
			return textCell(coin.Grade)
		}},
		{labels.presentation, 14, func(coin *models.Coin) ExportCell {
			return textCell(exportLabel(labels.presentations, coin.Presentation))
		}},
		{labels.rarity, 14, func(coin *models.Coin) ExportCell { return textCell(exportLabel(labels.rarities, coin.Rarity)) }},
		{labels.quantity, 10, func(coin *models.Coin) ExportCell { return intCell(max(coin.Quantity, 1)) }},
		{fmt.Sprintf(labels.purchasePrice, currency), 14, func(coin *models.Coin) ExportCell { return moneyCell(coin.PurchasePrice) }},
		{labels.purchaseDate, 12, func(coin *models.Coin) ExportCell { return dateCell(coin.PurchaseDate) }},
		{labels.purchasedFrom, 20, func(coin *models.Coin) ExportCell { return textCell(coin.PurchaseLocation) }},
		{fmt.Sprintf(labels.value, currency), 14, func(coin *models.Coin) ExportCell { return moneyCell(coin.CurrentValue) }},
		{labels.valueDate, 12, func(coin *models.Coin) ExportCell { return dateCell(coin.CurrentValueUpdatedAt) }},
		{labels.references, 24, func(coin *models.Coin) ExportCell { return textCell(shareListReferences(coin.References)) }},
		{labels.numista, 44, func(coin *models.Coin) ExportCell {
			if id := NumistaTypeIDForCoin(coin); id != 0 {
				url, _ := models.CanonicalNumistaURL(id)
				return textCell(url)
			}
			return ExportCell{}
		}},
		{labels.sets, 20, func(coin *models.Coin) ExportCell {
			names := []string{}
			for _, set := range coin.Sets {
				names = append(names, set.Name)
			}
			return textCell(strings.Join(names, ", "))
		}},
	}
	if wished {
		columns = append(columns, column{labels.wishlist, 10, func(coin *models.Coin) ExportCell {
			if coin.IsWishlist {
				return textCell(labels.yes)
			}
			return ExportCell{}
		}})
	}
	if banknotes {
		columns = append(columns,
			column{labels.series, 14, func(coin *models.Coin) ExportCell { return textCell(coin.Series) }},
			column{labels.serial, 14, func(coin *models.Coin) ExportCell { return textCell(coin.SerialNumber) }},
			column{labels.signatures, 20, func(coin *models.Coin) ExportCell { return textCell(coin.Signatures) }},
			column{labels.printer, 16, func(coin *models.Coin) ExportCell { return textCell(coin.Printer) }},
			column{labels.width, 11, func(coin *models.Coin) ExportCell { return numberCell(coin.WidthMm) }},
			column{labels.height, 11, func(coin *models.Coin) ExportCell { return numberCell(coin.HeightMm) }},
		)
	}
	columns = append(columns, column{labels.notes, 40, func(coin *models.Coin) ExportCell { return textCell(coin.Notes) }})

	export := CollectionExport{Lang: lang, Sheet: labels.sheet, Date: now}
	for _, col := range columns {
		export.Columns = append(export.Columns, col.name)
		export.Widths = append(export.Widths, col.width)
	}
	for i := range coins {
		row := make([]ExportCell, len(columns))
		for j, col := range columns {
			row[j] = col.value(&coins[i])
		}
		export.Rows = append(export.Rows, row)
	}
	return export
}

// exportLabel names a stored value in the export's language, or keeps it.
func exportLabel(names map[string]string, value string) string {
	if name, ok := names[value]; ok {
		return name
	}
	return value
}

// CSV writes the spreadsheet as Excel opens it in the language's countries:
// in Portuguese ";" between fields and a decimal comma, in English "," and a
// point; a byte order mark so accents open right; dates as 2026-10-21.
func (e CollectionExport) CSV() ([]byte, error) {
	var buffer bytes.Buffer
	buffer.WriteString("\xef\xbb\xbf")
	writer := csv.NewWriter(&buffer)
	if e.Lang == "pt" {
		writer.Comma = ';'
	}
	if err := writer.Write(e.Columns); err != nil {
		return nil, err
	}
	for _, row := range e.Rows {
		record := make([]string, len(row))
		for i, cell := range row {
			switch cell.kind {
			case exportText:
				record[i] = csvSafe(cell.text)
			case exportNumber, exportMoney:
				precision := -1
				if cell.kind == exportMoney {
					precision = 2
				}
				record[i] = strconv.FormatFloat(cell.number, 'f', precision, 64)
				if e.Lang == "pt" {
					record[i] = strings.Replace(record[i], ".", ",", 1)
				}
			case exportDate:
				record[i] = cell.date.Format("2006-01-02")
			}
		}
		if err := writer.Write(record); err != nil {
			return nil, err
		}
	}
	writer.Flush()
	return buffer.Bytes(), writer.Error()
}

// XLSX writes the spreadsheet as an Excel workbook: one sheet, its header
// bold and frozen with filters, numbers and dates as Excel's own.
func (e CollectionExport) XLSX() ([]byte, error) {
	var buffer bytes.Buffer
	archive := zip.NewWriter(&buffer)
	dateFormat := "dd/mm/yyyy"
	if e.Lang != "pt" {
		dateFormat = "yyyy-mm-dd"
	}
	files := []struct{ name, body string }{
		{"[Content_Types].xml", `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>` +
			`<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">` +
			`<Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>` +
			`<Default Extension="xml" ContentType="application/xml"/>` +
			`<Override PartName="/xl/workbook.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.sheet.main+xml"/>` +
			`<Override PartName="/xl/worksheets/sheet1.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.worksheet+xml"/>` +
			`<Override PartName="/xl/styles.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.styles+xml"/>` +
			`</Types>`},
		{"_rels/.rels", `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>` +
			`<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">` +
			`<Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="xl/workbook.xml"/>` +
			`</Relationships>`},
		{"xl/workbook.xml", `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>` +
			`<workbook xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships">` +
			`<sheets><sheet name="` + xmlText(e.Sheet) + `" sheetId="1" r:id="rId1"/></sheets>` +
			`<definedNames><definedName name="_xlnm._FilterDatabase" localSheetId="0" hidden="1">` + xmlText(quoteSheet(e.Sheet)) + `!$A$1:$` + columnName(len(e.Columns)) + `$` + strconv.Itoa(len(e.Rows)+1) + `</definedName></definedNames>` +
			`</workbook>`},
		{"xl/_rels/workbook.xml.rels", `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>` +
			`<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">` +
			`<Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/worksheet" Target="worksheets/sheet1.xml"/>` +
			`<Relationship Id="rId2" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/styles" Target="styles.xml"/>` +
			`</Relationships>`},
		// Styles: 0 plain, 1 the header (bold, filled), 2 a date, 3 a number with
		// decimals as needed, 4 an amount with two decimals.
		{"xl/styles.xml", `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>` +
			`<styleSheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main">` +
			`<numFmts count="1"><numFmt numFmtId="164" formatCode="` + dateFormat + `"/></numFmts>` +
			`<fonts count="2"><font><sz val="11"/><name val="Calibri"/></font><font><b/><sz val="11"/><name val="Calibri"/></font></fonts>` +
			`<fills count="3"><fill><patternFill patternType="none"/></fill><fill><patternFill patternType="gray125"/></fill>` +
			`<fill><patternFill patternType="solid"><fgColor rgb="FFF4EBD5"/><bgColor indexed="64"/></patternFill></fill></fills>` +
			`<borders count="1"><border><left/><right/><top/><bottom/><diagonal/></border></borders>` +
			`<cellStyleXfs count="1"><xf numFmtId="0" fontId="0" fillId="0" borderId="0"/></cellStyleXfs>` +
			`<cellXfs count="5"><xf numFmtId="0" fontId="0" fillId="0" borderId="0" xfId="0"/>` +
			`<xf numFmtId="0" fontId="1" fillId="2" borderId="0" xfId="0" applyFont="1" applyFill="1"/>` +
			`<xf numFmtId="164" fontId="0" fillId="0" borderId="0" xfId="0" applyNumberFormat="1"/>` +
			`<xf numFmtId="0" fontId="0" fillId="0" borderId="0" xfId="0"/>` +
			`<xf numFmtId="4" fontId="0" fillId="0" borderId="0" xfId="0" applyNumberFormat="1"/></cellXfs>` +
			`<cellStyles count="1"><cellStyle name="Normal" xfId="0" builtinId="0"/></cellStyles>` +
			`</styleSheet>`},
		{"xl/worksheets/sheet1.xml", e.sheetXML()},
	}
	for _, file := range files {
		writer, err := archive.Create(file.name)
		if err != nil {
			return nil, err
		}
		if _, err := writer.Write([]byte(file.body)); err != nil {
			return nil, err
		}
	}
	if err := archive.Close(); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

func (e CollectionExport) sheetXML() string {
	var sheet strings.Builder
	last := columnName(len(e.Columns))
	sheet.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>`)
	sheet.WriteString(`<worksheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main">`)
	sheet.WriteString(`<sheetViews><sheetView workbookViewId="0"><pane ySplit="1" topLeftCell="A2" activePane="bottomLeft" state="frozen"/></sheetView></sheetViews>`)
	sheet.WriteString(`<cols>`)
	for i, width := range e.Widths {
		fmt.Fprintf(&sheet, `<col min="%d" max="%d" width="%g" customWidth="1"/>`, i+1, i+1, width)
	}
	sheet.WriteString(`</cols><sheetData>`)
	sheet.WriteString(`<row r="1">`)
	for i, name := range e.Columns {
		fmt.Fprintf(&sheet, `<c r="%s1" t="inlineStr" s="1"><is><t>%s</t></is></c>`, columnName(i+1), xmlText(name))
	}
	sheet.WriteString(`</row>`)
	for r, row := range e.Rows {
		fmt.Fprintf(&sheet, `<row r="%d">`, r+2)
		for i, cell := range row {
			ref := columnName(i+1) + strconv.Itoa(r+2)
			switch cell.kind {
			case exportText:
				fmt.Fprintf(&sheet, `<c r="%s" t="inlineStr"><is><t xml:space="preserve">%s</t></is></c>`, ref, xmlText(cell.text))
			case exportNumber:
				fmt.Fprintf(&sheet, `<c r="%s" s="3"><v>%s</v></c>`, ref, strconv.FormatFloat(cell.number, 'f', -1, 64))
			case exportMoney:
				fmt.Fprintf(&sheet, `<c r="%s" s="4"><v>%s</v></c>`, ref, strconv.FormatFloat(cell.number, 'f', -1, 64))
			case exportDate:
				fmt.Fprintf(&sheet, `<c r="%s" s="2"><v>%s</v></c>`, ref, strconv.FormatFloat(excelSerial(cell.date), 'f', -1, 64))
			}
		}
		sheet.WriteString(`</row>`)
	}
	sheet.WriteString(`</sheetData>`)
	fmt.Fprintf(&sheet, `<autoFilter ref="A1:%s%d"/>`, last, len(e.Rows)+1)
	sheet.WriteString(`</worksheet>`)
	return sheet.String()
}

// columnName is a spreadsheet column's letters: 1 is A, 27 is AA.
func columnName(index int) string {
	name := ""
	for index > 0 {
		index--
		name = string(rune('A'+index%26)) + name
		index /= 26
	}
	return name
}

// excelSerial is a date as Excel counts days (from 30 December 1899).
func excelSerial(date time.Time) float64 {
	day := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)
	return day.Sub(time.Date(1899, 12, 30, 0, 0, 0, 0, time.UTC)).Hours() / 24
}

func quoteSheet(name string) string {
	return "'" + strings.ReplaceAll(name, "'", "''") + "'"
}

// xmlText escapes text for XML, leaving out the characters XML cannot hold.
func xmlText(value string) string {
	var out strings.Builder
	for _, r := range value {
		switch {
		case r == utf8.RuneError:
			continue
		case r == '&':
			out.WriteString("&amp;")
		case r == '<':
			out.WriteString("&lt;")
		case r == '>':
			out.WriteString("&gt;")
		case r == '"':
			out.WriteString("&quot;")
		case r == '\t' || r == '\n' || r == '\r' || r >= 0x20 && r != 0xFFFE && r != 0xFFFF:
			out.WriteRune(r)
		}
	}
	return out.String()
}
