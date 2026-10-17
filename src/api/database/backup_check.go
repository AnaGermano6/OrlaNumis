package database

import (
	"errors"
	"fmt"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// CheckBackupDatabase tells whether the file at path is a sound database of
// the app, one that can replace the running one: SQLite's quick check passes
// and it has users.
func CheckBackupDatabase(path string) error {
	db, err := gorm.Open(sqlite.Open(path), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		return fmt.Errorf("not a database: %w", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	defer sqlDB.Close()
	var check string
	if err := db.Raw("PRAGMA quick_check").Scan(&check).Error; err != nil {
		return fmt.Errorf("not a database: %w", err)
	}
	if check != "ok" {
		return fmt.Errorf("the database is damaged: %s", check)
	}
	if !db.Migrator().HasTable("users") {
		return errors.New("not a database of the app: it has no users")
	}
	var users int64
	if err := db.Table("users").Count(&users).Error; err != nil {
		return err
	}
	if users == 0 {
		return errors.New("the database has no users: nobody could sign in")
	}
	return nil
}
