package services

import "github.com/briandenicola/ancient-coins-api/models"

// EuroCirculationYear is a euro country's year of circulation coins: the
// eight coins of its annual series, and how many of them the collection has.
type EuroCirculationYear struct {
	Year  int `json:"year" example:"2024"`
	Total int `json:"total" example:"8"`
	Owned int `json:"owned" example:"3"`
}

// EuroCirculationYears lists, for each euro country (by its ISO code), every
// year from its first euro coins to thisYear, each with the eight coins of the
// year (1 cent to 2 euro) and how many of them the collection has, as the
// annual series challenge counts them: a coin of that value, country and year
// that is not a commemorative.
func EuroCirculationYears(coins []models.Coin, thisYear int) map[string][]EuroCirculationYear {
	owned := map[string]map[int]map[string]bool{}
	for i := range coins {
		coin := &coins[i]
		if coin.IsWishlist || coin.Year == nil || CoinPhotoQueryFor(coin).Subject != "" {
			continue
		}
		ecbCountry, key, ok := euroCoin(coin.Denomination, *coin.Year, coin.Country)
		code := ecbCountryCodes[ecbCountry]
		if !ok || code == "" {
			continue
		}
		if owned[code] == nil {
			owned[code] = map[int]map[string]bool{}
		}
		if owned[code][*coin.Year] == nil {
			owned[code][*coin.Year] = map[string]bool{}
		}
		owned[code][*coin.Year][key] = true
	}
	result := map[string][]EuroCirculationYear{}
	for country, first := range euroCoinFirstYear {
		code := ecbCountryCodes[country]
		if code == "" || first > thisYear {
			continue
		}
		years := make([]EuroCirculationYear, 0, thisYear-first+1)
		for year := first; year <= thisYear; year++ {
			years = append(years, EuroCirculationYear{Year: year, Total: len(annualSeriesCoins), Owned: len(owned[code][year])})
		}
		result[code] = years
	}
	return result
}
