package repository

import (
	"fmt"
	"testing"
	"time"

	"github.com/briandenicola/ancient-coins-api/models"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func TestFindDraftImageMediaByPathReturnsDraftOwner(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(fmt.Sprintf("file:image_repository_draft_%d?mode=memory&cache=shared", time.Now().UnixNano())), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	if err := db.AutoMigrate(&models.User{}, &models.QuickCaptureDraft{}, &models.QuickCaptureDraftImage{}); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}
	owner := models.User{Username: "draft-owner", Email: "draft-owner@example.com", PasswordHash: "x"}
	if err := db.Create(&owner).Error; err != nil {
		t.Fatal(err)
	}
	draft := models.QuickCaptureDraft{UserID: owner.ID, WorkingTitle: "Draft", Status: models.QuickCaptureDraftStatusActive}
	if err := db.Create(&draft).Error; err != nil {
		t.Fatal(err)
	}
	image := models.QuickCaptureDraftImage{DraftID: draft.ID, UserID: owner.ID, FilePath: "quick-capture-draft-1/obverse.png", ImageType: models.ImageTypeObverse}
	if err := db.Create(&image).Error; err != nil {
		t.Fatal(err)
	}

	media, err := NewImageRepository(db).FindDraftImageMediaByPath(image.FilePath)
	if err != nil {
		t.Fatalf("expected draft image media: %v", err)
	}
	if media.UserID != owner.ID {
		t.Fatalf("expected owner %d, got %d", owner.ID, media.UserID)
	}
}
