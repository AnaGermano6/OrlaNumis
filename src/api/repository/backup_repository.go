package repository

import (
	"github.com/briandenicola/ancient-coins-api/models"
	"gorm.io/gorm"
)

// BackupRepository reads the database for a full backup.
type BackupRepository struct {
	db *gorm.DB
}

func NewBackupRepository(db *gorm.DB) *BackupRepository {
	return &BackupRepository{db: db}
}

// Snapshot writes a consistent copy of the whole database to path (which
// must not exist yet), while the app keeps running: SQLite's VACUUM INTO.
func (r *BackupRepository) Snapshot(path string) error {
	return r.db.Exec("VACUUM INTO ?", path).Error
}

// Counts are how many users and coins (and banknotes) the database has.
func (r *BackupRepository) Counts() (users, coins int64, err error) {
	if err = r.db.Model(&models.User{}).Count(&users).Error; err != nil {
		return 0, 0, err
	}
	err = r.db.Model(&models.Coin{}).Count(&coins).Error
	return users, coins, err
}
