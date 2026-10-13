package handlers

import (
	"time"

	"github.com/briandenicola/ancient-coins-api/models"
)

type CoinReferenceRequest struct {
	Catalog string `json:"catalog" binding:"max=32"`
	Volume  string `json:"volume" binding:"max=64"`
	Number  string `json:"number" binding:"max=128"`
	URI     string `json:"uri" binding:"max=2000"`
}

type CoinCreateRequest struct {
	Name string `json:"name" binding:"max=200"`
	// Kind is a coin (default) or a banknote.
	Kind                  string          `json:"kind" binding:"max=16" enums:"coin,banknote"`
	Category              models.Category `json:"category" binding:"omitempty,max=64"`
	Denomination          string          `json:"denomination" binding:"max=200"`
	Ruler                 string          `json:"ruler" binding:"max=200"`
	RomanImperialFigureID *uint           `json:"romanImperialFigureId"`
	Era                   models.Era      `json:"era" binding:"omitempty,max=64"`
	DateRange             string          `json:"dateRange" binding:"max=200"`
	// Year is the coin's year (negative for BC); when omitted it is read from dateRange.
	Year    *int   `json:"year" binding:"omitempty,min=-5000,max=3000"`
	Country string `json:"country" binding:"max=120"`
	// Quantity is how many the collector has; 1 when omitted.
	Quantity     int    `json:"quantity" binding:"omitempty,min=1,max=9999" minimum:"1" maximum:"9999"`
	Presentation string `json:"presentation" binding:"max=20" enums:"loose,coincard,capsule,flip,proof_set,blister,slab"`
	Rarity       string `json:"rarity" binding:"max=20" enums:"common,uncommon,scarce,rare,very_rare,extremely_rare"`
	// Circulation set by hand; empty means "as Numista says".
	Circulation    string          `json:"circulation" binding:"max=30" enums:",current,current_commemorative,demonetized,demonetized_commemorative,non_circulating,non_circulating_commemorative"`
	Mint           string          `json:"mint" binding:"max=200"`
	MintLocationID *uint           `json:"mintLocationId"`
	Material       models.Material `json:"material"`
	WeightGrams    *float64        `json:"weightGrams"`
	DiameterMm     *float64        `json:"diameterMm"`
	Grade          string          `json:"grade" binding:"max=100"`
	// Banknote details: series, serial number, signatories, printer, size,
	// watermark, replacement or specimen, and a grading company's grade.
	Series              string     `json:"series" binding:"max=200"`
	SerialNumber        string     `json:"serialNumber" binding:"max=100"`
	Signatures          string     `json:"signatures" binding:"max=300"`
	Printer             string     `json:"printer" binding:"max=200"`
	WidthMm             *float64   `json:"widthMm"`
	HeightMm            *float64   `json:"heightMm"`
	Watermark           string     `json:"watermark" binding:"max=500"`
	NoteVariant         string     `json:"noteVariant" binding:"max=20" enums:",replacement,specimen"`
	CertifiedGrade      string     `json:"certifiedGrade" binding:"max=60"`
	CertificationNumber string     `json:"certificationNumber" binding:"max=60"`
	ObverseInscription  string     `json:"obverseInscription" binding:"max=1000"`
	ReverseInscription  string     `json:"reverseInscription" binding:"max=1000"`
	ObverseDescription  string     `json:"obverseDescription" binding:"max=2000"`
	ReverseDescription  string     `json:"reverseDescription" binding:"max=2000"`
	RarityRating        string     `json:"rarityRating" binding:"max=100"`
	PurchasePrice       *float64   `json:"purchasePrice"`
	CurrentValue        *float64   `json:"currentValue"`
	PurchaseDate        *time.Time `json:"purchaseDate"`
	PurchaseLocation    string     `json:"purchaseLocation" binding:"max=500"`
	VendorSKU           string     `json:"vendorSku" binding:"max=200"`
	VendorInvoice       string     `json:"vendorInvoice" binding:"max=200"`
	Notes               string     `json:"notes" binding:"max=5000"`
	ReferenceURL        string     `json:"referenceUrl" binding:"max=2000"`
	ReferenceText       string     `json:"referenceText" binding:"max=2000"`
	IsWishlist          bool       `json:"isWishlist"`
	IsSold              bool       `json:"isSold"`
	SoldPrice           *float64   `json:"soldPrice"`
	SoldDate            *time.Time `json:"soldDate"`
	SoldTo              string     `json:"soldTo"`
	StorageLocationID   *uint      `json:"storageLocationId" minimum:"1" extensions:"x-nullable"`
	// StorageSlot is a nullable, one-based row-major tray position.
	StorageSlot *int                   `json:"storageSlot" minimum:"1" maximum:"400" extensions:"x-nullable"`
	IsPrivate   bool                   `json:"isPrivate"`
	References  []CoinReferenceRequest `json:"references"`
}

type CoinUpdateRequest struct {
	Name *string `json:"name" binding:"omitempty,max=200"`
	// Kind is a coin or a banknote.
	Kind                  *string          `json:"kind" binding:"omitempty,max=16" enums:"coin,banknote"`
	Category              *models.Category `json:"category" binding:"omitempty,max=64"`
	Denomination          *string          `json:"denomination" binding:"omitempty,max=200"`
	Ruler                 *string          `json:"ruler" binding:"omitempty,max=200"`
	RomanImperialFigureID *uint            `json:"romanImperialFigureId"`
	Era                   *models.Era      `json:"era" binding:"omitempty,max=64"`
	DateRange             *string          `json:"dateRange" binding:"omitempty,max=200"`
	// Year (negative for BC); null means "read it from dateRange".
	Year    *int    `json:"year" binding:"omitempty,min=-5000,max=3000"`
	Country *string `json:"country" binding:"omitempty,max=120"`
	// Quantity is how many the collector has (1 or more).
	Quantity     *int    `json:"quantity" binding:"omitempty,min=1,max=9999" minimum:"1" maximum:"9999"`
	Presentation *string `json:"presentation" binding:"omitempty,max=20" enums:"loose,coincard,capsule,flip,proof_set,blister,slab"`
	Rarity       *string `json:"rarity" binding:"omitempty,max=20" enums:"common,uncommon,scarce,rare,very_rare,extremely_rare"`
	// Circulation set by hand; empty means "as Numista says".
	Circulation    *string          `json:"circulation" binding:"omitempty,max=30" enums:",current,current_commemorative,demonetized,demonetized_commemorative,non_circulating,non_circulating_commemorative"`
	Mint           *string          `json:"mint" binding:"omitempty,max=200"`
	MintLocationID *uint            `json:"mintLocationId"`
	Material       *models.Material `json:"material"`
	WeightGrams    *float64         `json:"weightGrams"`
	DiameterMm     *float64         `json:"diameterMm"`
	Grade          *string          `json:"grade" binding:"omitempty,max=100"`
	// Banknote details.
	Series              *string    `json:"series" binding:"omitempty,max=200"`
	SerialNumber        *string    `json:"serialNumber" binding:"omitempty,max=100"`
	Signatures          *string    `json:"signatures" binding:"omitempty,max=300"`
	Printer             *string    `json:"printer" binding:"omitempty,max=200"`
	WidthMm             *float64   `json:"widthMm"`
	HeightMm            *float64   `json:"heightMm"`
	Watermark           *string    `json:"watermark" binding:"omitempty,max=500"`
	NoteVariant         *string    `json:"noteVariant" binding:"omitempty,max=20" enums:",replacement,specimen"`
	CertifiedGrade      *string    `json:"certifiedGrade" binding:"omitempty,max=60"`
	CertificationNumber *string    `json:"certificationNumber" binding:"omitempty,max=60"`
	ObverseInscription  *string    `json:"obverseInscription" binding:"omitempty,max=1000"`
	ReverseInscription  *string    `json:"reverseInscription" binding:"omitempty,max=1000"`
	ObverseDescription  *string    `json:"obverseDescription" binding:"omitempty,max=2000"`
	ReverseDescription  *string    `json:"reverseDescription" binding:"omitempty,max=2000"`
	RarityRating        *string    `json:"rarityRating" binding:"omitempty,max=100"`
	PurchasePrice       *float64   `json:"purchasePrice"`
	CurrentValue        *float64   `json:"currentValue"`
	PurchaseDate        *time.Time `json:"purchaseDate"`
	PurchaseLocation    *string    `json:"purchaseLocation" binding:"omitempty,max=500"`
	VendorSKU           *string    `json:"vendorSku" binding:"omitempty,max=200"`
	VendorInvoice       *string    `json:"vendorInvoice" binding:"omitempty,max=200"`
	Notes               *string    `json:"notes" binding:"omitempty,max=5000"`
	ReferenceURL        *string    `json:"referenceUrl" binding:"omitempty,max=2000"`
	ReferenceText       *string    `json:"referenceText" binding:"omitempty,max=2000"`
	IsWishlist          *bool      `json:"isWishlist"`
	IsSold              *bool      `json:"isSold"`
	SoldPrice           *float64   `json:"soldPrice"`
	SoldDate            *time.Time `json:"soldDate"`
	SoldTo              *string    `json:"soldTo"`
	StorageLocationID   *uint      `json:"storageLocationId" minimum:"1" extensions:"x-nullable"`
	// StorageSlot is a nullable, one-based row-major tray position.
	StorageSlot *int                   `json:"storageSlot" minimum:"1" maximum:"400" extensions:"x-nullable"`
	IsPrivate   *bool                  `json:"isPrivate"`
	References  []CoinReferenceRequest `json:"references"`
}

func (r CoinCreateRequest) toCoin(userID uint) models.Coin {
	romanImperialFigureID := r.RomanImperialFigureID
	if r.Category != models.CategoryRoman {
		romanImperialFigureID = nil
	}
	return models.Coin{
		Name:                  r.Name,
		Kind:                  r.Kind,
		Category:              r.Category,
		Denomination:          r.Denomination,
		Ruler:                 r.Ruler,
		RomanImperialFigureID: romanImperialFigureID,
		Era:                   r.Era,
		DateRange:             r.DateRange,
		Year:                  r.Year,
		Country:               r.Country,
		Quantity:              r.Quantity,
		Presentation:          r.Presentation,
		Rarity:                r.Rarity,
		Circulation:           r.Circulation,
		Mint:                  r.Mint,
		MintLocationID:        r.MintLocationID,
		Material:              r.Material,
		WeightGrams:           r.WeightGrams,
		DiameterMm:            r.DiameterMm,
		Grade:                 r.Grade,
		Series:                r.Series,
		SerialNumber:          r.SerialNumber,
		Signatures:            r.Signatures,
		Printer:               r.Printer,
		WidthMm:               r.WidthMm,
		HeightMm:              r.HeightMm,
		Watermark:             r.Watermark,
		NoteVariant:           r.NoteVariant,
		CertifiedGrade:        r.CertifiedGrade,
		CertificationNumber:   r.CertificationNumber,
		ObverseInscription:    r.ObverseInscription,
		ReverseInscription:    r.ReverseInscription,
		ObverseDescription:    r.ObverseDescription,
		ReverseDescription:    r.ReverseDescription,
		RarityRating:          r.RarityRating,
		PurchasePrice:         r.PurchasePrice,
		CurrentValue:          r.CurrentValue,
		PurchaseDate:          r.PurchaseDate,
		PurchaseLocation:      r.PurchaseLocation,
		VendorSKU:             r.VendorSKU,
		VendorInvoice:         r.VendorInvoice,
		Notes:                 r.Notes,
		ReferenceURL:          r.ReferenceURL,
		ReferenceText:         r.ReferenceText,
		IsWishlist:            r.IsWishlist,
		IsSold:                r.IsSold,
		SoldPrice:             r.SoldPrice,
		SoldDate:              r.SoldDate,
		SoldTo:                r.SoldTo,
		StorageLocationID:     r.StorageLocationID,
		StorageSlot:           r.StorageSlot,
		IsPrivate:             r.IsPrivate,
		UserID:                userID,
		References:            mapCoinReferenceRequests(r.References),
	}
}

func appendUpdateField(fields []string, field string) []string {
	for _, existing := range fields {
		if existing == field {
			return fields
		}
	}
	return append(fields, field)
}

func (r CoinUpdateRequest) toCoin(existing *models.Coin, storageLocationProvided bool, nullableScalarProvided map[string]bool) (models.Coin, []string) {
	updates := models.Coin{ID: existing.ID, UserID: existing.UserID}
	updateFields := make([]string, 0, 32)
	targetCategory := existing.Category
	if r.Name != nil {
		updates.Name = *r.Name
		updateFields = append(updateFields, "Name")
	}
	if r.Category != nil {
		updates.Category = *r.Category
		targetCategory = *r.Category
		updateFields = append(updateFields, "Category")
	}
	if r.Denomination != nil {
		updates.Denomination = *r.Denomination
		updateFields = append(updateFields, "Denomination")
	}
	if r.Ruler != nil {
		updates.Ruler = *r.Ruler
		updateFields = append(updateFields, "Ruler")
	}
	if r.RomanImperialFigureID != nil || nullableScalarProvided["RomanImperialFigureID"] {
		updates.RomanImperialFigureID = r.RomanImperialFigureID
		updateFields = appendUpdateField(updateFields, "RomanImperialFigureID")
	}
	if r.Era != nil {
		updates.Era = *r.Era
		updateFields = append(updateFields, "Era")
	}
	if r.DateRange != nil {
		updates.DateRange = *r.DateRange
		updateFields = append(updateFields, "DateRange")
	}
	if r.Mint != nil {
		updates.Mint = *r.Mint
		updateFields = append(updateFields, "Mint")
	}
	if r.MintLocationID != nil || nullableScalarProvided["MintLocationID"] {
		updates.MintLocationID = r.MintLocationID
		updateFields = appendUpdateField(updateFields, "MintLocationID")
	}
	if r.Material != nil {
		updates.Material = *r.Material
		updateFields = append(updateFields, "Material")
	}
	if r.WeightGrams != nil || nullableScalarProvided["WeightGrams"] {
		updates.WeightGrams = r.WeightGrams
		updateFields = append(updateFields, "WeightGrams")
	}
	if r.DiameterMm != nil || nullableScalarProvided["DiameterMm"] {
		updates.DiameterMm = r.DiameterMm
		updateFields = append(updateFields, "DiameterMm")
	}
	if r.Grade != nil {
		updates.Grade = *r.Grade
		updateFields = append(updateFields, "Grade")
	}
	for _, field := range []struct {
		value *string
		into  *string
		name  string
	}{
		{r.Series, &updates.Series, "Series"},
		{r.SerialNumber, &updates.SerialNumber, "SerialNumber"},
		{r.Signatures, &updates.Signatures, "Signatures"},
		{r.Printer, &updates.Printer, "Printer"},
		{r.Watermark, &updates.Watermark, "Watermark"},
		{r.NoteVariant, &updates.NoteVariant, "NoteVariant"},
		{r.CertifiedGrade, &updates.CertifiedGrade, "CertifiedGrade"},
		{r.CertificationNumber, &updates.CertificationNumber, "CertificationNumber"},
	} {
		if field.value != nil {
			*field.into = *field.value
			updateFields = append(updateFields, field.name)
		}
	}
	if r.WidthMm != nil || nullableScalarProvided["WidthMm"] {
		updates.WidthMm = r.WidthMm
		updateFields = append(updateFields, "WidthMm")
	}
	if r.HeightMm != nil || nullableScalarProvided["HeightMm"] {
		updates.HeightMm = r.HeightMm
		updateFields = append(updateFields, "HeightMm")
	}
	if r.ObverseInscription != nil {
		updates.ObverseInscription = *r.ObverseInscription
		updateFields = append(updateFields, "ObverseInscription")
	}
	if r.ReverseInscription != nil {
		updates.ReverseInscription = *r.ReverseInscription
		updateFields = append(updateFields, "ReverseInscription")
	}
	if r.ObverseDescription != nil {
		updates.ObverseDescription = *r.ObverseDescription
		updateFields = append(updateFields, "ObverseDescription")
	}
	if r.ReverseDescription != nil {
		updates.ReverseDescription = *r.ReverseDescription
		updateFields = append(updateFields, "ReverseDescription")
	}
	if r.RarityRating != nil {
		updates.RarityRating = *r.RarityRating
		updateFields = append(updateFields, "RarityRating")
	}
	if r.Year != nil || nullableScalarProvided["Year"] {
		updates.Year = r.Year
		updateFields = append(updateFields, "Year")
	}
	if r.Country != nil {
		updates.Country = *r.Country
		updateFields = append(updateFields, "Country")
	}
	if r.Quantity != nil {
		updates.Quantity = *r.Quantity
		updateFields = append(updateFields, "Quantity")
	}
	if r.Presentation != nil {
		updates.Presentation = *r.Presentation
		updateFields = append(updateFields, "Presentation")
	}
	if r.Circulation != nil {
		updates.Circulation = *r.Circulation
		updateFields = append(updateFields, "Circulation")
	}
	if r.Kind != nil {
		updates.Kind = *r.Kind
		updateFields = append(updateFields, "Kind")
	}
	if r.Rarity != nil {
		updates.Rarity = *r.Rarity
		updateFields = append(updateFields, "Rarity")
	}
	if r.PurchasePrice != nil || nullableScalarProvided["PurchasePrice"] {
		updates.PurchasePrice = r.PurchasePrice
		updateFields = append(updateFields, "PurchasePrice")
	}
	if r.CurrentValue != nil || nullableScalarProvided["CurrentValue"] {
		updates.CurrentValue = r.CurrentValue
		updateFields = append(updateFields, "CurrentValue")
	}
	if r.PurchaseDate != nil || nullableScalarProvided["PurchaseDate"] {
		updates.PurchaseDate = r.PurchaseDate
		updateFields = append(updateFields, "PurchaseDate")
	}
	if r.PurchaseLocation != nil {
		updates.PurchaseLocation = *r.PurchaseLocation
		updateFields = append(updateFields, "PurchaseLocation")
	}
	if r.VendorSKU != nil {
		updates.VendorSKU = *r.VendorSKU
		updateFields = append(updateFields, "VendorSKU")
	}
	if r.VendorInvoice != nil {
		updates.VendorInvoice = *r.VendorInvoice
		updateFields = append(updateFields, "VendorInvoice")
	}
	if r.Notes != nil {
		updates.Notes = *r.Notes
		updateFields = append(updateFields, "Notes")
	}
	if r.ReferenceURL != nil {
		updates.ReferenceURL = *r.ReferenceURL
		updateFields = append(updateFields, "ReferenceURL")
	}
	if r.ReferenceText != nil {
		updates.ReferenceText = *r.ReferenceText
		updateFields = append(updateFields, "ReferenceText")
	}
	if r.IsWishlist != nil {
		updates.IsWishlist = *r.IsWishlist
		updateFields = append(updateFields, "IsWishlist")
	}
	if r.IsSold != nil {
		updates.IsSold = *r.IsSold
		updateFields = append(updateFields, "IsSold")
	}
	if r.SoldPrice != nil || nullableScalarProvided["SoldPrice"] {
		updates.SoldPrice = r.SoldPrice
		updateFields = append(updateFields, "SoldPrice")
	}
	if r.SoldDate != nil || nullableScalarProvided["SoldDate"] {
		updates.SoldDate = r.SoldDate
		updateFields = append(updateFields, "SoldDate")
	}
	if r.SoldTo != nil {
		updates.SoldTo = *r.SoldTo
		updateFields = append(updateFields, "SoldTo")
	}
	if storageLocationProvided {
		updates.StorageLocationID = r.StorageLocationID
	}
	if nullableScalarProvided["StorageSlot"] {
		updates.StorageSlot = r.StorageSlot
	}
	if r.IsPrivate != nil {
		updates.IsPrivate = *r.IsPrivate
		updateFields = append(updateFields, "IsPrivate")
	}
	if r.References != nil {
		updates.References = mapCoinReferenceRequests(r.References)
	}
	if targetCategory != models.CategoryRoman {
		updates.RomanImperialFigureID = nil
		updateFields = appendUpdateField(updateFields, "RomanImperialFigureID")
	}
	return updates, updateFields
}

func mapCoinReferenceRequests(requests []CoinReferenceRequest) []models.CoinReference {
	if requests == nil {
		return nil
	}
	refs := make([]models.CoinReference, 0, len(requests))
	for _, ref := range requests {
		refs = append(refs, models.CoinReference{
			Catalog: ref.Catalog,
			Volume:  ref.Volume,
			Number:  ref.Number,
			URI:     ref.URI,
		})
	}
	return refs
}
