package database

import (
	"crypto/sha256"
	"encoding/hex"
	"path/filepath"
	"strings"
	"testing"

	"github.com/briandenicola/ancient-coins-api/models"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

// A stored prompt that is still an old default (whatever its spaces and line
// ends) is cleared, so the new default applies; one the admin edited, and
// other settings, are kept.
func TestRetireOldDefaultPromptsClearsOnlyUnchangedCopies(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "prompts.db")), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, _ := db.DB()
	t.Cleanup(func() { sqlDB.Close() })
	if err := db.AutoMigrate(&models.AppSetting{}); err != nil {
		t.Fatal(err)
	}
	digest := func(text string) string {
		sum := sha256.Sum256([]byte(text))
		return hex.EncodeToString(sum[:])
	}
	saved := retiredDefaultPrompts
	t.Cleanup(func() { retiredDefaultPrompts = saved })
	retiredDefaultPrompts = map[string]string{
		"CoinSearchPrompt": digest("Search Roman coins.\nOnly for sale."),
		"ObversePrompt":    digest("Analyze the obverse."),
	}
	for key, value := range map[string]string{
		"CoinSearchPrompt": "  Search Roman coins.\r\nOnly for sale.\n",
		"ObversePrompt":    "Analyze the obverse. Also say the grade.",
		"AIProvider":       "anthropic",
	} {
		if err := db.Create(&models.AppSetting{Key: key, Value: value}).Error; err != nil {
			t.Fatal(err)
		}
	}

	if err := retireOldDefaultPrompts(db); err != nil {
		t.Fatal(err)
	}
	values := map[string]string{}
	var settings []models.AppSetting
	if err := db.Find(&settings).Error; err != nil {
		t.Fatal(err)
	}
	for _, setting := range settings {
		values[setting.Key] = setting.Value
	}
	if values["CoinSearchPrompt"] != "" || values["ObversePrompt"] != "Analyze the obverse. Also say the grade." || values["AIProvider"] != "anthropic" {
		t.Fatalf("settings = %v", values)
	}
}

// The digests are of the six prompts.
func TestRetiredDefaultPromptsAreTheSixPrompts(t *testing.T) {
	for _, key := range []string{"CoinSearchPrompt", "CoinShowsPrompt", "ValuationPrompt", "ObversePrompt", "ReversePrompt", "TextExtractionPrompt"} {
		if len(retiredDefaultPrompts[key]) != 64 {
			t.Errorf("no digest for %s", key)
		}
	}
}

// A stored source list that is still an old default (whatever its case and
// spaces) is cleared, so today's European default applies; an edited one is kept.
func TestRetireOldSearchSourcesClearsOnlyUnchangedLists(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "sources.db")), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, _ := db.DB()
	t.Cleanup(func() { sqlDB.Close() })
	if err := db.AutoMigrate(&models.AppSetting{}); err != nil {
		t.Fatal(err)
	}
	for key, value := range map[string]string{
		"DealerSearchSources":  "VCoins.com\r\nma-shops.com\n forumancientcoins.com \ncatawiki.com\nhjbltd.com\n",
		"AuctionSearchSources": "numisbids.com\ncngcoins.com",
	} {
		if err := db.Create(&models.AppSetting{Key: key, Value: value}).Error; err != nil {
			t.Fatal(err)
		}
	}

	if err := retireOldSearchSources(db); err != nil {
		t.Fatal(err)
	}
	values := map[string]string{}
	var settings []models.AppSetting
	if err := db.Find(&settings).Error; err != nil {
		t.Fatal(err)
	}
	for _, setting := range settings {
		values[setting.Key] = setting.Value
	}
	if values["DealerSearchSources"] != "" || values["AuctionSearchSources"] != "numisbids.com\ncngcoins.com" {
		t.Fatalf("settings = %v", values)
	}
}

// The app registers modern catalogues only, and removes the ancient and
// medieval ones it used to register, once and only when unused: one a coin
// cites stays, and one the admin registers again afterwards is kept.
func TestRetireOldCatalogsRemovesUnusedOnesOnce(t *testing.T) {
	Connect(filepath.Join(t.TempDir(), "catalogs.db"))
	sqlDB, _ := DB.DB()
	t.Cleanup(func() { sqlDB.Close() })
	codes := func() []string {
		var entries []models.CatalogRegistry
		if err := DB.Order("catalog").Find(&entries).Error; err != nil {
			t.Fatal(err)
		}
		out := []string{}
		for _, entry := range entries {
			out = append(out, entry.Catalog)
		}
		return out
	}
	if got := strings.Join(codes(), " "); got != "KM NGC Numista Pick REDBOOK Y" {
		t.Fatalf("registered at start = %s", got)
	}

	// A database from before: the old catalogues, one of them cited.
	user := models.User{Username: "collector", Email: "collector@example.test"}
	if err := DB.Create(&user).Error; err != nil {
		t.Fatal(err)
	}
	coin := models.Coin{Name: "Denarius", UserID: user.ID}
	if err := DB.Create(&coin).Error; err != nil {
		t.Fatal(err)
	}
	for _, code := range []string{"RIC", "SNG"} {
		if err := DB.Create(&models.CatalogRegistry{Catalog: code, DisplayName: code, Era: models.EraAncient}).Error; err != nil {
			t.Fatal(err)
		}
	}
	if err := DB.Create(&models.CoinReference{CoinID: coin.ID, Catalog: "RIC", Volume: "II", Number: "207"}).Error; err != nil {
		t.Fatal(err)
	}
	if err := DB.Where("key = ?", retiredCatalogsVersionKey).Delete(&models.AppSetting{}).Error; err != nil {
		t.Fatal(err)
	}

	if err := retireOldCatalogs(DB); err != nil {
		t.Fatal(err)
	}
	if got := strings.Join(codes(), " "); got != "KM NGC Numista Pick REDBOOK RIC Y" {
		t.Fatalf("after retiring = %s", got)
	}
	if err := DB.Create(&models.CatalogRegistry{Catalog: "SNG", DisplayName: "SNG", Era: models.EraAncient}).Error; err != nil {
		t.Fatal(err)
	}
	if err := retireOldCatalogs(DB); err != nil {
		t.Fatal(err)
	}
	if got := strings.Join(codes(), " "); got != "KM NGC Numista Pick REDBOOK RIC SNG Y" {
		t.Fatalf("registered again = %s", got)
	}
}
