package models

import "time"

type Material string

const (
	MaterialGold     Material = "Gold"
	MaterialSilver   Material = "Silver"
	MaterialBronze   Material = "Bronze"
	MaterialCopper   Material = "Copper"
	MaterialElectrum Material = "Electrum"
	MaterialOther    Material = "Other"
)

type Category string

const (
	CategoryRoman     Category = "Roman"
	CategoryGreek     Category = "Greek"
	CategoryByzantine Category = "Byzantine"
	CategoryModern    Category = "Modern"
	CategoryOther     Category = "Other"
)

type ImageType string

const (
	ImageTypeObverse ImageType = "obverse"
	ImageTypeReverse ImageType = "reverse"
	ImageTypeDetail  ImageType = "detail"
	ImageTypeOther   ImageType = "other"
)

type Era string

const (
	EraAncient  Era = "ancient"
	EraMedieval Era = "medieval"
	EraModern   Era = "modern"
)

type Coin struct {
	ID   uint   `gorm:"primaryKey" json:"id"`
	Name string `gorm:"not null" json:"name" binding:"max=200"`
	// Kind is what the item is: a coin or a banknote (see CoinKind…).
	Kind                  string   `gorm:"type:varchar(16);not null;default:'coin';index" json:"kind" enums:"coin,banknote"`
	Category              Category `gorm:"type:varchar(20);not null;default:'Other'" json:"category" binding:"omitempty,max=64"`
	Denomination          string   `json:"denomination" binding:"max=200"`
	Ruler                 string   `json:"ruler" binding:"max=200"`
	RomanImperialFigureID *uint    `gorm:"index" json:"romanImperialFigureId"`
	Era                   Era      `gorm:"type:varchar(64)" json:"era" binding:"omitempty,max=64"`
	DateRange             string   `gorm:"type:varchar(200)" json:"dateRange" binding:"max=200"`
	// Year is derived from DateRange (first year, negative for BC) for sorting and grouping.
	Year *int `gorm:"index" json:"year"`
	// Country is the issuing country or authority (e.g. from the Numista issuer).
	Country string `gorm:"type:varchar(120);index" json:"country" binding:"max=120"`
	// Quantity is how many of this coin or note the collector has (1 or more).
	Quantity int `gorm:"not null;default:1" json:"quantity" minimum:"1" maximum:"9999"`
	// Presentation is how the coin is kept or was issued (see CoinPresentations).
	Presentation string `gorm:"type:varchar(20);default:''" json:"presentation"`
	// Rarity is a fixed scale (see CoinRarities). RarityRating is a legacy free-text field.
	Rarity string `gorm:"type:varchar(20);default:''" json:"rarity"`
	// Circulation, when set by hand (see CoinCirculations), says whether the
	// coin is still legal tender and whether it is commemorative, instead of
	// what Numista says of its type.
	Circulation    string        `gorm:"type:varchar(30);default:''" json:"circulation" enums:",current,current_commemorative,demonetized,demonetized_commemorative,non_circulating,non_circulating_commemorative"`
	Mint           string        `json:"mint" binding:"max=200"`
	MintLocationID *uint         `json:"mintLocationId"`
	MintLocation   *MintLocation `gorm:"foreignKey:MintLocationID;constraint:-" json:"mintLocation"`
	Material       Material      `gorm:"type:varchar(20);default:'Other'" json:"material"`
	WeightGrams    *float64      `json:"weightGrams"`
	DiameterMm     *float64      `json:"diameterMm"`
	Grade          string        `json:"grade" binding:"max=100"`
	// Banknote details (see Kind): the series or issue ("Chapa 13"), the
	// note's own serial number, its signatories, printer, size, watermark,
	// whether it is a replacement or specimen note, and a grading company's
	// grade ("PMG 66 EPQ") with its certificate.
	Series                string     `gorm:"type:varchar(200);default:''" json:"series" binding:"max=200"`
	SerialNumber          string     `gorm:"type:varchar(100);default:''" json:"serialNumber" binding:"max=100"`
	Signatures            string     `gorm:"type:varchar(300);default:''" json:"signatures" binding:"max=300"`
	Printer               string     `gorm:"type:varchar(200);default:''" json:"printer" binding:"max=200"`
	WidthMm               *float64   `json:"widthMm"`
	HeightMm              *float64   `json:"heightMm"`
	Watermark             string     `gorm:"type:varchar(500);default:''" json:"watermark" binding:"max=500"`
	NoteVariant           string     `gorm:"type:varchar(20);default:''" json:"noteVariant" enums:",replacement,specimen"`
	CertifiedGrade        string     `gorm:"type:varchar(60);default:''" json:"certifiedGrade" binding:"max=60"`
	CertificationNumber   string     `gorm:"type:varchar(60);default:''" json:"certificationNumber" binding:"max=60"`
	ObverseInscription    string     `json:"obverseInscription" binding:"max=1000"`
	ReverseInscription    string     `json:"reverseInscription" binding:"max=1000"`
	ObverseDescription    string     `json:"obverseDescription" binding:"max=2000"`
	ReverseDescription    string     `json:"reverseDescription" binding:"max=2000"`
	RarityRating          string     `json:"rarityRating" binding:"max=100"`
	PurchasePrice         *float64   `json:"purchasePrice"`
	CurrentValue          *float64   `json:"currentValue"`
	CurrentValueUpdatedAt *time.Time `json:"currentValueUpdatedAt"`
	PurchaseDate          *time.Time `json:"purchaseDate"`
	PurchaseLocation      string     `json:"purchaseLocation" binding:"max=500"`
	VendorSKU             string     `json:"vendorSku" binding:"max=200"`
	VendorInvoice         string     `json:"vendorInvoice" binding:"max=200"`
	Notes                 string     `gorm:"type:text" json:"notes" binding:"max=5000"`
	AIAnalysis            string     `gorm:"type:text;column:ai_analysis" json:"aiAnalysis"`
	ObverseAnalysis       string     `gorm:"type:text;column:obverse_analysis" json:"obverseAnalysis"`
	ReverseAnalysis       string     `gorm:"type:text;column:reverse_analysis" json:"reverseAnalysis"`
	ReferenceURL          string     `json:"referenceUrl" binding:"max=2000"`
	ReferenceText         string     `json:"referenceText" binding:"max=2000"`
	IsWishlist            bool       `gorm:"default:false" json:"isWishlist"`
	IsSold                bool       `gorm:"default:false" json:"isSold"`
	SoldPrice             *float64   `json:"soldPrice"`
	SoldDate              *time.Time `json:"soldDate"`
	SoldTo                string     `json:"soldTo"`
	ListingStatus         string     `gorm:"type:varchar(20);default:''" json:"listingStatus"`
	ListingCheckedAt      *time.Time `json:"listingCheckedAt"`
	ListingCheckReason    string     `gorm:"type:text" json:"listingCheckReason"`
	StorageLocationID     *uint      `json:"storageLocationId" minimum:"1" extensions:"x-nullable"`
	// StorageSlot is the nullable, one-based row-major position within a Coin Tray.
	StorageSlot            *int             `json:"storageSlot" minimum:"1" maximum:"400" extensions:"x-nullable"`
	StorageLocation        *StorageLocation `gorm:"foreignKey:StorageLocationID;constraint:-" json:"storageLocation"`
	SourceAlertCandidateID *uint            `gorm:"index" json:"sourceAlertCandidateId"`
	IsPrivate              bool             `gorm:"default:false" json:"isPrivate"`
	UserID                 uint             `gorm:"not null" json:"userId"`
	User                   User             `gorm:"foreignKey:UserID" json:"-"`
	Images                 []CoinImage      `gorm:"foreignKey:CoinID" json:"images"`
	References             []CoinReference  `gorm:"foreignKey:CoinID" json:"references"`
	Tags                   []Tag            `gorm:"many2many:coin_tags" json:"tags"`
	Sets                   []CoinSet        `gorm:"many2many:coin_set_memberships;joinForeignKey:CoinID;joinReferences:SetID" json:"sets"`
	CreatedAt              time.Time        `json:"createdAt"`
	UpdatedAt              time.Time        `json:"updatedAt"`
}

type CoinImage struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CoinID    uint      `gorm:"not null" json:"coinId"`
	FilePath  string    `gorm:"not null" json:"filePath"`
	ImageType ImageType `gorm:"type:varchar(20);default:'other'" json:"imageType"`
	IsPrimary bool      `gorm:"default:false" json:"isPrimary"`
	CreatedAt time.Time `json:"createdAt"`
}
