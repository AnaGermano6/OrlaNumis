package database

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/briandenicola/ancient-coins-api/models"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

// A backup's database replaces the app's only when it is a sound database
// of the app with users.
func TestCheckBackupDatabase(t *testing.T) {
	dir := t.TempDir()
	newDB := func(name string, users int) string {
		path := filepath.Join(dir, name)
		db, err := gorm.Open(sqlite.Open(path), &gorm.Config{})
		if err != nil {
			t.Fatal(err)
		}
		sqlDB, _ := db.DB()
		defer sqlDB.Close()
		if err := db.AutoMigrate(&models.User{}); err != nil {
			t.Fatal(err)
		}
		for i := range users {
			if err := db.Create(&models.User{Username: name + string(rune('a'+i)), Email: name + string(rune('a'+i)) + "@example.test"}).Error; err != nil {
				t.Fatal(err)
			}
		}
		return path
	}

	if err := CheckBackupDatabase(newDB("app.db", 1)); err != nil {
		t.Fatalf("a database with a user: %v", err)
	}
	if err := CheckBackupDatabase(newDB("empty.db", 0)); err == nil {
		t.Fatal("a database without users passed")
	}
	other := filepath.Join(dir, "other.db")
	db, _ := gorm.Open(sqlite.Open(other), &gorm.Config{})
	db.Exec("CREATE TABLE notes (id INTEGER PRIMARY KEY)")
	sqlDB, _ := db.DB()
	sqlDB.Close()
	if err := CheckBackupDatabase(other); err == nil {
		t.Fatal("another app's database passed")
	}
	text := filepath.Join(dir, "notes.txt")
	if err := os.WriteFile(text, []byte("not a database at all, but long enough to have a header of sorts"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := CheckBackupDatabase(text); err == nil {
		t.Fatal("a text file passed")
	}
}
