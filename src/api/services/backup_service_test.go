package services

import (
	"archive/zip"
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/briandenicola/ancient-coins-api/models"
	"github.com/briandenicola/ancient-coins-api/repository"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// backupTest has an app's database file (a user, a coin), its uploads and a
// backup folder.
type backupTest struct {
	db        *gorm.DB
	dbPath    string
	uploadDir string
	dir       string
	settings  *SettingsService
	backups   *BackupService
}

func openBackupTestDB(t *testing.T, path string) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(path), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, _ := db.DB()
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { sqlDB.Close() })
	return db
}

// sqliteHeaderCheck stands for the database package's check, which this
// package may not import.
func sqliteHeaderCheck(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if !bytes.HasPrefix(data, []byte("SQLite format 3\x00")) {
		return errors.New("not a database")
	}
	return nil
}

func writeUpload(t *testing.T, uploadDir, rel, content string) {
	t.Helper()
	target := filepath.Join(uploadDir, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(target), 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(target, []byte(content), 0o640); err != nil {
		t.Fatal(err)
	}
}

func newBackupTest(t *testing.T) *backupTest {
	t.Helper()
	root := t.TempDir()
	test := &backupTest{
		dbPath:    filepath.Join(root, "data", "ancientcoins.db"),
		uploadDir: filepath.Join(root, "uploads"),
		dir:       filepath.Join(root, "backups"),
	}
	if err := os.MkdirAll(filepath.Dir(test.dbPath), 0o750); err != nil {
		t.Fatal(err)
	}
	test.db = openBackupTestDB(t, test.dbPath)
	if err := test.db.AutoMigrate(&models.User{}, &models.Coin{}, &models.AppSetting{}); err != nil {
		t.Fatal(err)
	}
	if err := test.db.Create(&models.User{Username: "collector", Email: "collector@example.test", PasswordHash: "x"}).Error; err != nil {
		t.Fatal(err)
	}
	if err := test.db.Create(&models.Coin{Name: "2 euro 2022", UserID: 1}).Error; err != nil {
		t.Fatal(err)
	}
	writeUpload(t, test.uploadDir, "coin-1/obverse.jpg", "obverse")
	writeUpload(t, test.uploadDir, "avatars/1.png", "avatar")
	test.settings = NewSettingsService(repository.NewSettingsRepository(test.db))
	test.backups = NewBackupService(repository.NewBackupRepository(test.db), test.dbPath, test.uploadDir, test.dir, "4.3.0", test.settings, NewLogger(100), sqliteHeaderCheck)
	// A minute between backups.
	clock := time.Date(2026, 10, 16, 3, 0, 0, 0, time.UTC)
	test.backups.now = func() time.Time {
		clock = clock.Add(time.Minute)
		return clock
	}
	return test
}

func zipNames(t *testing.T, path string) []string {
	t.Helper()
	zr, err := zip.OpenReader(path)
	if err != nil {
		t.Fatal(err)
	}
	defer zr.Close()
	names := []string{}
	for _, file := range zr.File {
		names = append(names, file.Name)
	}
	sort.Strings(names)
	return names
}

func TestBackupService_MakesListsAndPrunes(t *testing.T) {
	test := newBackupTest(t)

	info, err := test.backups.Create(BackupKindManual)
	if err != nil {
		t.Fatal(err)
	}
	if info.Kind != BackupKindManual || info.Users != 1 || info.Coins != 1 || info.Files != 2 || info.Size == 0 || info.AppVersion != "4.3.0" {
		t.Fatalf("backup = %+v", info)
	}
	path, err := test.backups.Path(info.Name)
	if err != nil {
		t.Fatal(err)
	}
	if got := strings.Join(zipNames(t, path), ", "); got != "database/ancientcoins.db, manifest.json, uploads/avatars/1.png, uploads/coin-1/obverse.jpg" {
		t.Fatalf("zip = %s", got)
	}
	if test.settings.GetSetting(SettingBackupStatus) != "ok" {
		t.Fatalf("status = %s", test.settings.GetSetting(SettingBackupStatus))
	}

	// Two automatic backups kept: the oldest of three goes, the manual one stays.
	if err := test.settings.SetSetting(SettingBackupKeep, "2"); err != nil {
		t.Fatal(err)
	}
	if test.backups.LastAutomatic() != nil {
		t.Fatal("no automatic backup yet")
	}
	automatic := []string{}
	for range 3 {
		info, err := test.backups.Create(BackupKindAutomatic)
		if err != nil {
			t.Fatal(err)
		}
		automatic = append(automatic, info.Name)
	}
	backups, err := test.backups.List()
	if err != nil {
		t.Fatal(err)
	}
	names := []string{}
	for _, backup := range backups {
		names = append(names, backup.Name)
	}
	if len(backups) != 3 || names[0] != automatic[2] || names[1] != automatic[1] || !strings.HasSuffix(names[2], "-manual.zip") {
		t.Fatalf("backups = %v (made %v)", names, automatic)
	}
	if backups[2].Coins != 1 || backups[2].Files != 2 {
		t.Fatalf("listed manual backup = %+v", backups[2])
	}
	if last := test.backups.LastAutomatic(); last == nil || !last.Equal(backups[0].CreatedAt) {
		t.Fatalf("last automatic = %v, want %v", last, backups[0].CreatedAt)
	}

	// Only the folder's backups, by name.
	for _, name := range []string{"../data/ancientcoins.db", "aurearia-2026-10-16-030000-auto.zip", "notes.txt"} {
		if _, err := test.backups.Path(name); !errors.Is(err, ErrBackupNotFound) {
			t.Fatalf("Path(%q) = %v", name, err)
		}
	}
	if err := test.backups.Delete(names[2]); err != nil {
		t.Fatal(err)
	}
	if backups, _ := test.backups.List(); len(backups) != 2 {
		t.Fatalf("after deleting: %d backups", len(backups))
	}
}

// A restore is prepared (the data of now backed up first), then applied at
// the next start: the database and the uploads are the backup's again.
func TestBackupService_RestoresAtTheNextStart(t *testing.T) {
	test := newBackupTest(t)
	backup, err := test.backups.Create(BackupKindManual)
	if err != nil {
		t.Fatal(err)
	}
	// Changed since.
	if err := test.db.Create(&models.Coin{Name: "5 euro 2023", UserID: 1}).Error; err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(filepath.Join(test.uploadDir, "coin-1", "obverse.jpg")); err != nil {
		t.Fatal(err)
	}
	writeUpload(t, test.uploadDir, "coin-2/obverse.jpg", "new")

	if err := test.backups.StageRestore(backup.Name); err != nil {
		t.Fatal(err)
	}
	backups, _ := test.backups.List()
	if len(backups) != 2 || backups[0].Kind != BackupKindPreRestore || backups[0].Coins != 2 {
		t.Fatalf("backups after preparing = %+v", backups)
	}
	if test.backups.LastRestore() != nil {
		t.Fatal("nothing restored yet")
	}

	// The app stops; at its start the backup is applied.
	sqlDB, _ := test.db.DB()
	sqlDB.Close()
	result := ApplyPendingRestore(test.dbPath, test.uploadDir)
	if result == nil || !result.OK || result.Name != backup.Name {
		t.Fatalf("result = %+v", result)
	}
	db := openBackupTestDB(t, test.dbPath)
	var coins int64
	db.Model(&models.Coin{}).Count(&coins)
	if coins != 1 {
		t.Fatalf("coins after restoring = %d, want 1", coins)
	}
	files, err := uploadFiles(test.uploadDir)
	if err != nil {
		t.Fatal(err)
	}
	if got := strings.Join(files, ", "); got != "avatars/1.png, coin-1/obverse.jpg" {
		t.Fatalf("uploads after restoring = %s", got)
	}
	if _, err := os.Stat(filepath.Join(filepath.Dir(test.dbPath), restorePendingDir)); !os.IsNotExist(err) {
		t.Fatalf("the prepared restore is still there: %v", err)
	}
	if last := test.backups.LastRestore(); last == nil || !last.OK || last.Name != backup.Name {
		t.Fatalf("last restore = %+v", last)
	}
	// Nothing more at the next start.
	if result := ApplyPendingRestore(test.dbPath, test.uploadDir); result != nil {
		t.Fatalf("a second start restored again: %+v", result)
	}
}

// A file that is not one of the app's backups is not kept, nor restored.
func TestBackupService_ImportsOnlyItsBackups(t *testing.T) {
	test := newBackupTest(t)
	backup, err := test.backups.Create(BackupKindManual)
	if err != nil {
		t.Fatal(err)
	}
	path, _ := test.backups.Path(backup.Name)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	imported, err := test.backups.SaveUpload(bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	if imported.Kind != BackupKindUploaded || !imported.CreatedAt.Equal(backup.CreatedAt) || imported.Coins != 1 {
		t.Fatalf("imported = %+v", imported)
	}

	zipOf := func(files map[string]string) []byte {
		var buffer bytes.Buffer
		zw := zip.NewWriter(&buffer)
		for name, content := range files {
			w, _ := zw.Create(name)
			_, _ = w.Write([]byte(content))
		}
		zw.Close()
		return buffer.Bytes()
	}
	manifest := `{"format":"aurearia-backup","formatVersion":1}`
	for label, body := range map[string][]byte{
		"not a zip":        []byte("hello"),
		"no manifest":      zipOf(map[string]string{"coins.json": "[]"}),
		"another format":   zipOf(map[string]string{"manifest.json": `{"format":"other","formatVersion":1}`}),
		"a newer format":   zipOf(map[string]string{"manifest.json": `{"format":"aurearia-backup","formatVersion":9}`}),
		"no database":      zipOf(map[string]string{"manifest.json": manifest}),
		"not a database":   zipOf(map[string]string{"manifest.json": manifest, "database/ancientcoins.db": "text"}),
		"outside uploads":  zipOf(map[string]string{"manifest.json": manifest, "database/ancientcoins.db": "SQLite format 3\x00", "uploads/../../etc/passwd": "x"}),
		"another top file": zipOf(map[string]string{"manifest.json": manifest, "database/ancientcoins.db": "SQLite format 3\x00", "run.sh": "x"}),
	} {
		if _, err := test.backups.SaveUpload(bytes.NewReader(body)); err == nil {
			t.Fatalf("%s: imported", label)
		}
	}
	backups, _ := test.backups.List()
	if len(backups) != 2 {
		t.Fatalf("backups = %+v", backups)
	}
	if entries, _ := os.ReadDir(test.dir); len(entries) != 2 {
		t.Fatalf("the backup folder has %d files, want the 2 backups only", len(entries))
	}
}

func TestUploadEntryPath(t *testing.T) {
	for name, want := range map[string]string{
		"uploads/coin-1/obverse.jpg": "coin-1/obverse.jpg",
		"uploads/a.png":              "a.png",
		"uploads/../x":               "",
		"uploads/coin-1/../../x":     "",
		"uploads//etc/passwd":        "",
		"uploads/./a.png":            "",
		`uploads/a\..\b`:             "",
		"uploads/.restoring/a.png":   "",
		"uploads/":                   "",
		"database/ancientcoins.db":   "",
		"other/a.png":                "",
	} {
		got, ok := uploadEntryPath(name)
		if got != want || ok != (want != "") {
			t.Errorf("uploadEntryPath(%q) = %q, %v; want %q", name, got, ok, want)
		}
	}
}

func TestBackupScheduler_RunsDailyAndCatchesUp(t *testing.T) {
	test := newBackupTest(t)
	scheduler := NewBackupScheduler(test.backups, test.settings, NewLogger(100))
	if err := test.settings.SetSetting(SettingScheduleTimezone, "UTC"); err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 10, 16, 12, 0, 0, 0, time.UTC)

	// None yet: now.
	if next := scheduler.nextRun(now); !next.Equal(now) {
		t.Fatalf("first backup at %v", next)
	}
	// Made today at 03:00: tomorrow at 03:00.
	writeUpload(t, test.dir, "aurearia-2026-10-16-030000-auto.zip", "")
	if next := scheduler.nextRun(now); !next.Equal(time.Date(2026, 10, 17, 3, 0, 0, 0, time.UTC)) {
		t.Fatalf("next = %v", next)
	}
	// Last made two days ago (the computer was asleep): now, but an hour after a failed try.
	if err := os.Remove(filepath.Join(test.dir, "aurearia-2026-10-16-030000-auto.zip")); err != nil {
		t.Fatal(err)
	}
	writeUpload(t, test.dir, "aurearia-2026-10-14-030000-auto.zip", "")
	if next := scheduler.nextRun(now); !next.Equal(now) {
		t.Fatalf("catch-up at %v", next)
	}
	scheduler.lastAttempt = now.Add(-10 * time.Minute)
	if next := scheduler.nextRun(now); !next.Equal(now.Add(50 * time.Minute)) {
		t.Fatalf("retry at %v", next)
	}
	// Off: only the daily time, where the run does nothing.
	if err := test.settings.SetSetting(SettingBackupEnabled, "false"); err != nil {
		t.Fatal(err)
	}
	if next := scheduler.nextRun(now); !next.Equal(time.Date(2026, 10, 17, 3, 0, 0, 0, time.UTC)) {
		t.Fatalf("next when off = %v", next)
	}
	if status := scheduler.GetStatus(); status.Enabled || status.Name != "Automatic Backup" {
		t.Fatalf("status = %+v", status)
	}
}
