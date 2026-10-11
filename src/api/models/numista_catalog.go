package models

// NumistaCatalogIssuer is a country or issuing authority in the Numista catalogue.
type NumistaCatalogIssuer struct {
	Code   string `json:"code"`
	Name   string `json:"name"`
	Parent string `json:"parent,omitempty"`
}

// NumistaCatalogIssuersResult lists issuers matching a name search.
type NumistaCatalogIssuersResult struct {
	Status  string                 `json:"status"`
	Issuers []NumistaCatalogIssuer `json:"issuers"`
}

// NumistaCatalogSearchItem is one coin type in a catalogue search.
type NumistaCatalogSearchItem struct {
	ID               int                  `json:"id"`
	Title            string               `json:"title"`
	Issuer           NumistaCatalogIssuer `json:"issuer"`
	MinYear          *int                 `json:"minYear,omitempty"`
	MaxYear          *int                 `json:"maxYear,omitempty"`
	ObverseThumbnail string               `json:"obverseThumbnail,omitempty"`
	ReverseThumbnail string               `json:"reverseThumbnail,omitempty"`
}

// NumistaCatalogSearchResult is one page of catalogue search results.
type NumistaCatalogSearchResult struct {
	Status string                     `json:"status"`
	Count  int                        `json:"count"`
	Page   int                        `json:"page"`
	Types  []NumistaCatalogSearchItem `json:"types"`
}

// NumistaCatalogPhoto is a reference photo from the Numista catalogue. The
// photos belong to their authors; Copyright and License must be shown with them.
type NumistaCatalogPhoto struct {
	Side      string `json:"side"`
	Thumbnail string `json:"thumbnail"`
	Picture   string `json:"picture,omitempty"`
	Copyright string `json:"copyright,omitempty"`
	License   string `json:"license,omitempty"`
}

// NumistaCatalogSide describes the obverse, reverse or edge of a coin type.
type NumistaCatalogSide struct {
	Description string   `json:"description,omitempty"`
	Lettering   string   `json:"lettering,omitempty"`
	Designers   []string `json:"designers,omitempty"`
	Engravers   []string `json:"engravers,omitempty"`
}

// NumistaCatalogReference is a catalogue number for a type (e.g. KM 77).
type NumistaCatalogReference struct {
	Catalogue string `json:"catalogue"`
	Number    string `json:"number"`
}

// NumistaCatalogType holds the catalogue details used to prefill a new coin.
type NumistaCatalogType struct {
	ID           int                  `json:"id"`
	URL          string               `json:"url"`
	Title        string               `json:"title"`
	Issuer       NumistaCatalogIssuer `json:"issuer"`
	MinYear      *int                 `json:"minYear,omitempty"`
	MaxYear      *int                 `json:"maxYear,omitempty"`
	Denomination string               `json:"denomination,omitempty"`
	// FaceValue is the denomination as a number (5 for "5 Euros").
	FaceValue *float64 `json:"faceValue,omitempty"`
	Currency  string   `json:"currency,omitempty"`
	// CommemoratedTopic is what a commemorative coin celebrates, in the
	// language asked for when Numista has it (e.g. "O conhecimento").
	CommemoratedTopic string `json:"commemoratedTopic,omitempty"`
	Series            string `json:"series,omitempty"`
	// Classification is Numista's kind of coin (e.g. "Moedas circulantes
	// comuns"), in the language asked for.
	Classification string `json:"classification,omitempty"`
	// Demonetized tells whether the type is no longer legal tender (nil
	// when Numista does not say); DemonetizationDate is when, if known.
	Demonetized        *bool                     `json:"demonetized,omitempty"`
	DemonetizationDate string                    `json:"demonetizationDate,omitempty"`
	Rulers             []string                  `json:"rulers"`
	Composition        string                    `json:"composition,omitempty"`
	Technique          string                    `json:"technique,omitempty"`
	WeightGrams        *float64                  `json:"weightGrams,omitempty"`
	DiameterMm         *float64                  `json:"diameterMm,omitempty"`
	ThicknessMm        *float64                  `json:"thicknessMm,omitempty"`
	Shape              string                    `json:"shape,omitempty"`
	Mints              []string                  `json:"mints"`
	Obverse            NumistaCatalogSide        `json:"obverse"`
	Reverse            NumistaCatalogSide        `json:"reverse"`
	Edge               NumistaCatalogSide        `json:"edge"`
	References         []NumistaCatalogReference `json:"references"`
	Photos             []NumistaCatalogPhoto     `json:"photos"`
}

// NumistaCatalogTypeResult is a catalogue type with its issues.
type NumistaCatalogTypeResult struct {
	Status string                `json:"status"`
	Type   *NumistaCatalogType   `json:"type,omitempty"`
	Issues []NumistaIssueSummary `json:"issues"`
}
