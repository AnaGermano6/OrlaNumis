package repository

import (
	"errors"
	"github.com/briandenicola/ancient-coins-api/models"
	"gorm.io/gorm"
)

// CoinImageMediaAccess says whose coin an uploaded image is, to authorize
// serving it.
type CoinImageMediaAccess struct {
	FilePath   string
	CoinUserID uint
}

type DraftImageMediaAccess struct {
	FilePath string
	UserID   uint
}

// ImageRepository encapsulates all image-related database operations.
type ImageRepository struct {
	db *gorm.DB
}

// NewImageRepository creates a new ImageRepository.
func NewImageRepository(db *gorm.DB) *ImageRepository {
	return &ImageRepository{db: db}
}

// FindCoinByOwner returns a coin if it belongs to the given user.
func (r *ImageRepository) FindCoinByOwner(coinID, userID uint) (*models.Coin, error) {
	var coin models.Coin
	err := r.db.Where("id = ? AND user_id = ?", coinID, userID).First(&coin).Error
	if err != nil {
		return nil, err
	}
	return &coin, nil
}

// CreateImage inserts a new image record.
func (r *ImageRepository) CreateImage(image *models.CoinImage) error {
	return r.db.Create(image).Error
}

// FindImage returns an image by ID and coin ID.
func (r *ImageRepository) FindImage(imageID, coinID uint) (*models.CoinImage, error) {
	var image models.CoinImage
	err := r.db.Where("id = ? AND coin_id = ?", imageID, coinID).First(&image).Error
	if err != nil {
		return nil, err
	}
	return &image, nil
}

// FindCoinImageByType returns the most recently uploaded image of the given
// type for a coin (used by deep identification to resolve an obverse/reverse
// role from a saved coin when no new upload is provided for that role).
func (r *ImageRepository) FindCoinImageByType(coinID uint, imageType models.ImageType) (*models.CoinImage, error) {
	var image models.CoinImage
	err := r.db.Where("coin_id = ? AND image_type = ?", coinID, imageType).
		Order("id DESC").
		First(&image).Error
	if err != nil {
		return nil, err
	}
	return &image, nil
}

// FindCoinImageMediaByPath returns whose coin a stored file path is an image of.
func (r *ImageRepository) FindCoinImageMediaByPath(filePath string) (*CoinImageMediaAccess, error) {
	var media CoinImageMediaAccess
	tx := r.db.Table("coin_images").
		Select(`coin_images.file_path, coins.user_id AS coin_user_id`).
		Joins("JOIN coins ON coins.id = coin_images.coin_id").
		Where("coin_images.file_path = ?", filePath).
		Limit(1).
		Scan(&media)
	if tx.Error != nil {
		return nil, tx.Error
	}
	if tx.RowsAffected == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	return &media, nil
}

// FindAvatarOwnerByPath returns the user whose avatar_path matches filePath.
func (r *ImageRepository) FindAvatarOwnerByPath(filePath string) (*models.User, error) {
	var user models.User
	err := r.db.Where("avatar_path = ?", filePath).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *ImageRepository) FindDraftImageMediaByPath(filePath string) (*DraftImageMediaAccess, error) {
	var media DraftImageMediaAccess
	tx := r.db.Table("quick_capture_draft_images").
		Select("file_path, user_id").
		Where("file_path = ?", filePath).
		Limit(1).
		Scan(&media)
	if tx.Error != nil {
		return nil, tx.Error
	}
	if tx.RowsAffected == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	return &media, nil
}

// DeleteImage removes an image record.
func (r *ImageRepository) DeleteImage(image *models.CoinImage) error {
	return r.db.Delete(image).Error
}

func (r *ImageRepository) DeleteAndQueueCleanup(imageID, coinID, userID uint) (*models.ImageCleanup, error) {
	var cleanup models.ImageCleanup
	err := r.db.Transaction(func(tx *gorm.DB) error {
		err := tx.Where("image_id = ? AND coin_id = ? AND user_id = ?", imageID, coinID, userID).First(&cleanup).Error
		if err == nil {
			return nil
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		var image models.CoinImage
		if err := tx.Where("id = ? AND coin_id = ?", imageID, coinID).First(&image).Error; err != nil {
			return err
		}
		cleanup = models.ImageCleanup{ImageID: image.ID, CoinID: coinID, UserID: userID, FilePath: image.FilePath}
		if err := tx.Create(&cleanup).Error; err != nil {
			return err
		}
		return tx.Delete(&image).Error
	})
	return &cleanup, err
}

func (r *ImageRepository) PendingCleanups(after uint) ([]models.ImageCleanup, error) {
	var cleanups []models.ImageCleanup
	err := r.db.Where("image_id > ?", after).Order("image_id").Limit(100).Find(&cleanups).Error
	return cleanups, err
}

func (r *ImageRepository) CompleteCleanup(imageID uint) error {
	return r.db.Where("image_id = ?", imageID).Delete(&models.ImageCleanup{}).Error
}

// SetPrimaryAndCreate clears the primary flag on existing images and creates
// a new image record in a single transaction.
func (r *ImageRepository) SetPrimaryAndCreate(coinID uint, image *models.CoinImage) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.CoinImage{}).Where("coin_id = ?", coinID).Update("is_primary", false).Error; err != nil {
			return err
		}
		return tx.Create(image).Error
	})
}

// SwapSides turns a coin's obverse images into reverse ones and back, in one
// transaction, and returns the coin's images afterwards.
func (r *ImageRepository) SwapSides(coinID uint) ([]models.CoinImage, error) {
	var images []models.CoinImage
	err := r.db.Transaction(func(tx *gorm.DB) error {
		swap := "CASE image_type WHEN ? THEN ? ELSE ? END"
		if err := tx.Model(&models.CoinImage{}).
			Where("coin_id = ? AND image_type IN ?", coinID, []models.ImageType{models.ImageTypeObverse, models.ImageTypeReverse}).
			Update("image_type", gorm.Expr(swap, models.ImageTypeObverse, models.ImageTypeReverse, models.ImageTypeObverse)).Error; err != nil {
			return err
		}
		return tx.Where("coin_id = ?", coinID).Order("id").Find(&images).Error
	})
	return images, err
}

// UpdateCoinNotes replaces a coin's notes.
func (r *ImageRepository) UpdateCoinNotes(coinID uint, notes string) error {
	return r.db.Model(&models.Coin{}).Where("id = ?", coinID).Update("notes", notes).Error
}
