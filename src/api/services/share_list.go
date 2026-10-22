package services

import (
	"bytes"
	"encoding/csv"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/briandenicola/ancient-coins-api/models"
	"github.com/briandenicola/ancient-coins-api/repository"
	"golang.org/x/text/collate"
	"golang.org/x/text/language"
)

// The lists a collector shares with others: the duplicates there are to
// trade, the coins and banknotes wanted (the wishlist), or both.
const (
	ShareListTrades   = "trades"
	ShareListWishlist = "wishlist"
	ShareListBoth     = "both"
)

// ShareListRequest says which list to make, in which language, and what it shows.
type ShareListRequest struct {
	List string `json:"list" binding:"required" enums:"trades,wishlist,both"`
	// Format is "pdf" or "csv".
	Format string `json:"format" binding:"required" enums:"pdf,csv"`
	// Lang is the language of the list: "pt" (European Portuguese) or "en".
	Lang string `json:"lang" enums:"pt,en"`
	// IncludeValues adds each coin's current value.
	IncludeValues bool `json:"includeValues"`
	// Contact, when given, is printed at the top (an e-mail, a phone); it is not stored.
	Contact string `json:"contact" binding:"max=120"`
}

// ErrShareListInvalid is returned for a request that names no known list or format.
var ErrShareListInvalid = errors.New("list must be trades, wishlist or both, and format pdf or csv")

// ShareListItem is one coin or banknote of a shared list.
type ShareListItem struct {
	Name         string
	Banknote     bool
	Country      string
	Year         string
	Denomination string
	Grade        string
	// Available is how many there are to trade (the duplicates'); 0 for a wanted item.
	Available  int
	References string
	NumistaURL string
	Value      *float64
	// Photo is the coin's obverse (else its first photo), relative to the upload folder.
	Photo string
}

// ShareListSection is one list: the trades or the wanted.
type ShareListSection struct {
	Kind  string
	Title string
	Intro string
	Items []ShareListItem
}

// ShareList is what a shared file shows.
type ShareList struct {
	Title    string
	Contact  string
	Date     time.Time
	Lang     string
	Currency string
	Values   bool
	Sections []ShareListSection
	Labels   ShareListLabels
}

// ShareListLabels are a shared list's words, in its language.
type ShareListLabels struct {
	List, Kind, Coin, Banknote, Country, Item, Year, Denomination, Grade, Available string
	References, Numista, Value, Contact, Date, Empty, Footer, Page                  string
}

var shareListLabels = map[string]ShareListLabels{
	"pt": {
		List: "Lista", Kind: "Tipo", Coin: "Moeda", Banknote: "Nota", Country: "País", Item: "Peça", Year: "Ano",
		Denomination: "Valor facial", Grade: "Grau", Available: "Disponíveis", References: "Referências", Numista: "Numista",
		Value: "Valor", Contact: "Contacto", Date: "Data", Empty: "Nada nesta lista.", Footer: "Feito com a Orla Numis", Page: "Página {page} de {nb}",
	},
	"en": {
		List: "List", Kind: "Kind", Coin: "Coin", Banknote: "Banknote", Country: "Country", Item: "Item", Year: "Year",
		Denomination: "Face value", Grade: "Grade", Available: "Available", References: "References", Numista: "Numista",
		Value: "Value", Contact: "Contact", Date: "Date", Empty: "Nothing on this list.", Footer: "Made with Orla Numis", Page: "Page {page} of {nb}",
	},
}

var shareListTitles = map[string]map[string][3]string{
	// Per language: the list's title, the line under it, and the file's name.
	"pt": {
		ShareListTrades:   {"Lista de trocas", "Peças repetidas que tenho para trocar.", "lista-de-trocas"},
		ShareListWishlist: {"Procuro", "Peças que procuro.", "procuro"},
		ShareListBoth:     {"Trocas e procuras", "", "trocas-e-procuras"},
	},
	"en": {
		ShareListTrades:   {"Trade list", "Duplicates I have to trade.", "trade-list"},
		ShareListWishlist: {"Wanted", "Coins and banknotes I am looking for.", "wanted"},
		ShareListBoth:     {"Trades and wants", "", "trades-and-wants"},
	},
}

// ShareListLabelsFor is a shared list's words in a language ("pt" or "en").
func ShareListLabelsFor(lang string) ShareListLabels {
	if labels, ok := shareListLabels[lang]; ok {
		return labels
	}
	return shareListLabels["en"]
}

// ShareListService makes the lists a collector shares.
type ShareListService struct {
	coins    *repository.CoinRepository
	settings *SettingsService
	now      func() time.Time
}

// NewShareListService creates the service; settings gives the collection
// currency and may be nil.
func NewShareListService(coins *repository.CoinRepository, settings *SettingsService) *ShareListService {
	return &ShareListService{coins: coins, settings: settings, now: time.Now}
}

// NormalizeShareListRequest checks a request and fills its defaults.
func NormalizeShareListRequest(request ShareListRequest) (ShareListRequest, error) {
	request.List = strings.ToLower(strings.TrimSpace(request.List))
	request.Format = strings.ToLower(strings.TrimSpace(request.Format))
	request.Lang = strings.ToLower(strings.TrimSpace(request.Lang))
	request.Contact = strings.Join(strings.Fields(request.Contact), " ")
	if request.List != ShareListTrades && request.List != ShareListWishlist && request.List != ShareListBoth {
		return request, ErrShareListInvalid
	}
	if request.Format != "pdf" && request.Format != "csv" {
		return request, ErrShareListInvalid
	}
	if strings.HasPrefix(request.Lang, "pt") || request.Lang == "" {
		request.Lang = "pt"
	} else {
		request.Lang = "en"
	}
	return request, nil
}

// FileName is the shared file's name: "lista-de-trocas-2026-10-21.pdf".
func (l ShareList) FileName(list, format string) string {
	return fmt.Sprintf("%s-%s.%s", shareListTitles[l.Lang][list][2], l.Date.Format("2006-01-02"), format)
}

// Build lists a user's trades, wanted coins, or both, each sorted by
// country, year and name.
func (s *ShareListService) Build(userID uint, request ShareListRequest) (ShareList, error) {
	list := ShareList{
		Title: shareListTitles[request.Lang][request.List][0], Contact: request.Contact, Date: s.now(), Lang: request.Lang,
		Values: request.IncludeValues, Labels: shareListLabels[request.Lang], Currency: CurrencyEUR,
	}
	if s.settings != nil {
		list.Currency = s.settings.CollectionCurrency()
	}
	kinds := []string{ShareListTrades, ShareListWishlist}
	if request.List != ShareListBoth {
		kinds = []string{request.List}
	}
	for _, kind := range kinds {
		coins, err := s.coins.ShareListCoins(userID, kind == ShareListWishlist)
		if err != nil {
			return list, err
		}
		section := ShareListSection{Kind: kind, Title: shareListTitles[request.Lang][kind][0], Intro: shareListTitles[request.Lang][kind][1]}
		for i := range coins {
			section.Items = append(section.Items, shareListItem(&coins[i], kind, request.IncludeValues))
		}
		sortShareListItems(section.Items, request.Lang)
		list.Sections = append(list.Sections, section)
	}
	return list, nil
}

func shareListItem(coin *models.Coin, kind string, values bool) ShareListItem {
	item := ShareListItem{
		Name: strings.TrimSpace(coin.Name), Banknote: coin.Kind == models.CoinKindBanknote, Country: strings.TrimSpace(coin.Country),
		Denomination: strings.TrimSpace(coin.Denomination), Grade: strings.TrimSpace(coin.Grade), References: shareListReferences(coin.References),
		Photo: shareListPhoto(coin.Images),
	}
	if item.Grade == "" {
		item.Grade = strings.TrimSpace(coin.CertifiedGrade)
	}
	switch {
	case coin.Year != nil && strings.TrimSpace(coin.DateRange) == "":
		item.Year = strconv.Itoa(*coin.Year)
	default:
		item.Year = strings.TrimSpace(coin.DateRange)
	}
	if kind == ShareListTrades {
		item.Available = max(coin.Quantity-1, 0)
	}
	if id := NumistaTypeIDForCoin(coin); id != 0 {
		item.NumistaURL, _ = models.CanonicalNumistaURL(id)
	}
	if values && coin.CurrentValue != nil && *coin.CurrentValue > 0 {
		value := *coin.CurrentValue
		item.Value = &value
	}
	return item
}

// shareListReferences names a coin's catalogue references ("KM 631 · Numista 2933").
func shareListReferences(references []models.CoinReference) string {
	parts := []string{}
	for _, ref := range references {
		catalog, number := strings.TrimSpace(ref.Catalog), strings.TrimSpace(ref.Number)
		if catalog == "" || number == "" {
			continue
		}
		if volume := strings.TrimSpace(ref.Volume); volume != "" {
			catalog += " " + volume
		}
		parts = append(parts, catalog+" "+number)
	}
	return strings.Join(parts, " · ")
}

// shareListPhoto is the photo a coin is shown by: its obverse, else its first.
func shareListPhoto(images []models.CoinImage) string {
	for _, image := range images {
		if image.ImageType == models.ImageTypeObverse {
			return image.FilePath
		}
	}
	if len(images) > 0 {
		return images[0].FilePath
	}
	return ""
}

func sortShareListItems(items []ShareListItem, lang string) {
	tag := language.English
	if lang == "pt" {
		tag = language.Portuguese
	}
	collator := collate.New(tag, collate.IgnoreCase, collate.IgnoreDiacritics)
	sort.SliceStable(items, func(i, j int) bool {
		if c := collator.CompareString(items[i].Country, items[j].Country); c != 0 {
			return items[i].Country != "" && (items[j].Country == "" || c < 0)
		}
		if items[i].Year != items[j].Year {
			return items[i].Year < items[j].Year
		}
		return collator.CompareString(items[i].Name, items[j].Name) < 0
	})
}

// FormatShareValue writes an amount as the list's language does: "12,50 €"
// in Portuguese, "€12.50" in English.
func FormatShareValue(amount float64, currency, lang string) string {
	symbol := map[string]string{CurrencyEUR: "€", CurrencyUSD: "$", "GBP": "£"}[currency]
	if lang == "pt" {
		number := strings.Replace(strconv.FormatFloat(amount, 'f', 2, 64), ".", ",", 1)
		if symbol == "" {
			symbol = currency
		}
		return number + " " + symbol
	}
	if symbol == "" {
		return fmt.Sprintf("%.2f %s", amount, currency)
	}
	return fmt.Sprintf("%s%.2f", symbol, amount)
}

// csvSafe keeps a cell from opening as a formula in a spreadsheet.
func csvSafe(value string) string {
	if value != "" && strings.ContainsRune("=+@\t\r", rune(value[0])) {
		return "'" + value
	}
	return value
}

// CSV writes the list as a spreadsheet: one row per item, with the list it
// is on; ";" between fields in Portuguese (as Excel there expects), "," in
// English, and a byte order mark so accents open right.
func (l ShareList) CSV() ([]byte, error) {
	var buffer bytes.Buffer
	buffer.WriteString("\xef\xbb\xbf")
	writer := csv.NewWriter(&buffer)
	if l.Lang == "pt" {
		writer.Comma = ';'
	}
	labels := l.Labels
	header := []string{labels.List, labels.Kind, labels.Country, labels.Item, labels.Year, labels.Denomination, labels.Grade, labels.Available, labels.References, labels.Numista}
	if l.Values {
		header = append(header, labels.Value)
	}
	if err := writer.Write(header); err != nil {
		return nil, err
	}
	for _, section := range l.Sections {
		for _, item := range section.Items {
			kind := labels.Coin
			if item.Banknote {
				kind = labels.Banknote
			}
			available := ""
			if section.Kind == ShareListTrades {
				available = strconv.Itoa(item.Available)
			}
			row := []string{section.Title, kind, item.Country, item.Name, item.Year, item.Denomination, item.Grade, available, item.References, item.NumistaURL}
			if l.Values {
				value := ""
				if item.Value != nil {
					value = FormatShareValue(*item.Value, l.Currency, l.Lang)
				}
				row = append(row, value)
			}
			for i := range row {
				row[i] = csvSafe(row[i])
			}
			if err := writer.Write(row); err != nil {
				return nil, err
			}
		}
	}
	writer.Flush()
	return buffer.Bytes(), writer.Error()
}
