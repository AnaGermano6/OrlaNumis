package models

// Sources of real coin photos offered when adding or completing a coin.
const (
	CoinPhotoSourceNumista   = "numista"
	CoinPhotoSourceWikimedia = "wikimedia"
)

// CoinPhotoCandidate is a real photo of a coin type found in a public source.
// Photos belong to their authors: Author and License must be shown and kept
// with the photo when it is saved.
type CoinPhotoCandidate struct {
	Source string `json:"source"`
	// Side is "obverse" or "reverse" when the source states it.
	Side      string `json:"side,omitempty"`
	Thumbnail string `json:"thumbnail"`
	// ImageURL is the full-size photo to import into a coin.
	ImageURL string `json:"imageUrl"`
	PageURL  string `json:"pageUrl,omitempty"`
	Title    string `json:"title,omitempty"`
	Author   string `json:"author,omitempty"`
	License  string `json:"license,omitempty"`
}

// CoinPhotoSearchResult lists real photos found for a coin.
type CoinPhotoSearchResult struct {
	Status string               `json:"status"`
	Query  string               `json:"query"`
	Photos []CoinPhotoCandidate `json:"photos"`
}
