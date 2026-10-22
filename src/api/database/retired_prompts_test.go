package database

import (
	"crypto/sha256"
	"encoding/hex"
	"path/filepath"
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
