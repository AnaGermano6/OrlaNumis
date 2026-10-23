package services

import (
	"archive/zip"
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/briandenicola/ancient-coins-api/repository"
)

// The kinds of full backup: made by the schedule, by hand, of the data
// replaced by a restore (to go back), or imported from a file.
const (
	BackupKindAutomatic  = "auto"
	BackupKindManual     = "manual"
	BackupKindPreRestore = "pre-restore"
	BackupKindPreUpdate  = "pre-update"
	BackupKindUploaded   = "uploaded"
)

const (
	// backupFormat names the format in a backup's manifest. It keeps the
	// app's first name, so that every backup, old or new, can be restored.
	backupFormat        = "aurearia-backup"
	backupFormatVersion = 1
	backupManifestName  = "manifest.json"
	backupDatabaseName  = "database/ancientcoins.db"
	backupUploadsPrefix = "uploads/"
	backupTimeLayout    = "2006-01-02-150405"
	defaultBackupKeep   = 7
	maxBackupKeep       = 365
	// restorePendingDir, next to the database, holds a backup to restore at
	// the next start: backup.zip, then request.json once it is complete.
	restorePendingDir   = "restore-pending"
	restoreResultName   = "restore-result.json"
	uploadsRestoringDir = ".restoring"
)

// Backups made before the app was named Orla Numis keep their "aurearia-" name.
var backupNamePattern = regexp.MustCompile(`^(?:orla-numis|aurearia)-(\d{4}-\d{2}-\d{2}-\d{6})-(auto|manual|pre-restore|pre-update|uploaded)\.zip$`)

// ErrBackupNotFound is a backup name that is not one of the folder's backups.
var ErrBackupNotFound = errors.New("backup not found")

// ErrBackupRunning is asked while a backup or a restore is under way.
var ErrBackupRunning = errors.New("a backup is already being made")

// BackupManifest describes a backup, in its manifest.json.
type BackupManifest struct {
	Format        string    `json:"format"`
	FormatVersion int       `json:"formatVersion"`
	AppVersion    string    `json:"appVersion"`
	CreatedAt     time.Time `json:"createdAt"`
	Kind          string    `json:"kind"`
	Users         int64     `json:"users"`
	Coins         int64     `json:"coins"`
	Files         int       `json:"files"`
}

// BackupInfo is a backup of the folder.
type BackupInfo struct {
	Name string `json:"name" example:"orla-numis-2026-10-16-030000-auto.zip"`
	Kind string `json:"kind" enums:"auto,manual,pre-restore,pre-update,uploaded"`
	// CreatedAt is when its data was copied (for an imported backup, when it
	// was made, not imported).
	CreatedAt  time.Time `json:"createdAt"`
	Size       int64     `json:"size"`
	AppVersion string    `json:"appVersion,omitempty"`
	// Users, Coins (coins and banknotes) and Files (photos and other uploads) it holds.
	Users int64 `json:"users"`
	Coins int64 `json:"coins"`
	Files int   `json:"files"`
}

// RestoreResult is how the last restore went, at the start that applied it.
type RestoreResult struct {
	Name  string    `json:"name"`
	At    time.Time `json:"at"`
	OK    bool      `json:"ok"`
	Error string    `json:"error,omitempty"`
}

// BackupService makes full backups (the whole database, a consistent copy
// taken while the app runs, and every file of the uploads folder, in one
// zip), keeps the last ones, and prepares a backup to be restored at the
// next start, when nothing uses the database.
type BackupService struct {
	repo       *repository.BackupRepository
	dbPath     string
	uploadDir  string
	dir        string
	appVersion string
	settings   *SettingsService
	logger     *Logger
	// checkDatabase tells whether a backup's database can replace the app's.
	checkDatabase func(path string) error
	now           func() time.Time

	mu      sync.Mutex
	running atomic.Bool
}

// NewBackupService keeps the backups in dir; checkDatabase vets a backup's
// database before it is restored.
func NewBackupService(repo *repository.BackupRepository, dbPath, uploadDir, dir, appVersion string, settings *SettingsService, logger *Logger, checkDatabase func(path string) error) *BackupService {
	return &BackupService{
		repo: repo, dbPath: dbPath, uploadDir: uploadDir, dir: dir, appVersion: appVersion,
		settings: settings, logger: logger, checkDatabase: checkDatabase, now: time.Now,
	}
}

// Running tells whether a backup is being made.
func (s *BackupService) Running() bool {
	return s.running.Load()
}

// Keep is how many automatic backups are kept.
func (s *BackupService) Keep() int {
	keep, err := strconv.Atoi(strings.TrimSpace(s.settings.GetSetting(SettingBackupKeep)))
	if err != nil || keep < 1 {
		return defaultBackupKeep
	}
	return min(keep, maxBackupKeep)
}

// Start makes a backup in the background; Running is true from now until it
// is made.
func (s *BackupService) Start(kind string) error {
	if !s.mu.TryLock() {
		return ErrBackupRunning
	}
	s.running.Store(true)
	go func() {
		defer s.mu.Unlock()
		_, _ = s.create(kind)
	}()
	return nil
}

// Create makes a backup now.
func (s *BackupService) Create(kind string) (BackupInfo, error) {
	if !s.mu.TryLock() {
		return BackupInfo{}, ErrBackupRunning
	}
	defer s.mu.Unlock()
	return s.create(kind)
}

func (s *BackupService) create(kind string) (info BackupInfo, err error) {
	s.running.Store(true)
	defer s.running.Store(false)
	started := s.now()
	defer func() {
		status := "ok"
		if err != nil {
			status = "failed"
			s.logger.Error("backup", "Backup (%s) failed: %v", kind, err)
		} else {
			s.logger.Info("backup", "Backup %s made in %s: %d users, %d coins, %d files, %d bytes",
				info.Name, s.now().Sub(started).Round(time.Second), info.Users, info.Coins, info.Files, info.Size)
		}
		if setErr := s.settings.SetSetting(SettingBackupStatus, status); setErr != nil {
			s.logger.Warn("backup", "Could not save the backup status: %v", setErr)
		}
	}()

	if err := os.MkdirAll(s.dir, 0o750); err != nil {
		return BackupInfo{}, fmt.Errorf("backup folder: %w", err)
	}
	at := started.UTC().Truncate(time.Second)
	name := backupName(at, kind)
	for exists(filepath.Join(s.dir, name)) {
		at = at.Add(time.Second)
		name = backupName(at, kind)
	}

	// The database first: a copy as of now, outside the backup folder (which
	// may be a folder of the host, where SQLite's locks are not to be relied on).
	tmp, err := os.MkdirTemp("", "orla-numis-backup-")
	if err != nil {
		return BackupInfo{}, err
	}
	defer os.RemoveAll(tmp)
	snapshot := filepath.Join(tmp, "ancientcoins.db")
	if err := s.repo.Snapshot(snapshot); err != nil {
		return BackupInfo{}, fmt.Errorf("copying the database: %w", err)
	}
	manifest := BackupManifest{Format: backupFormat, FormatVersion: backupFormatVersion, AppVersion: s.appVersion, CreatedAt: at, Kind: kind}
	if manifest.Users, manifest.Coins, err = s.repo.Counts(); err != nil {
		return BackupInfo{}, fmt.Errorf("counting: %w", err)
	}

	// Written under another name, so the folder only ever lists whole backups.
	partial := filepath.Join(s.dir, ".partial-"+name)
	defer os.Remove(partial)
	if err := writeBackup(partial, &manifest, snapshot, s.uploadDir); err != nil {
		return BackupInfo{}, err
	}
	final := filepath.Join(s.dir, name)
	if err := os.Rename(partial, final); err != nil {
		return BackupInfo{}, err
	}
	info = backupInfo(name, kind, manifest)
	if stat, err := os.Stat(final); err == nil {
		info.Size = stat.Size()
	}
	if kind == BackupKindAutomatic {
		s.prune()
	}
	return info, nil
}

func backupName(at time.Time, kind string) string {
	return fmt.Sprintf("orla-numis-%s-%s.zip", at.UTC().Format(backupTimeLayout), kind)
}

func backupInfo(name, kind string, manifest BackupManifest) BackupInfo {
	return BackupInfo{
		Name: name, Kind: kind, CreatedAt: manifest.CreatedAt, AppVersion: manifest.AppVersion,
		Users: manifest.Users, Coins: manifest.Coins, Files: manifest.Files,
	}
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// writeBackup writes the zip: the database, every upload (stored as they
// are: photos are compressed already) and, last, the manifest with how many
// files it took.
func writeBackup(target string, manifest *BackupManifest, snapshot, uploadDir string) (err error) {
	file, err := os.OpenFile(target, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o640)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := file.Close(); err == nil {
			err = closeErr
		}
	}()
	buffered := bufio.NewWriterSize(file, 1<<20)
	zw := zip.NewWriter(buffered)
	if err := addFileToZip(zw, backupDatabaseName, snapshot, zip.Deflate, manifest.CreatedAt); err != nil {
		return fmt.Errorf("adding the database: %w", err)
	}
	files, err := uploadFiles(uploadDir)
	if err != nil {
		return fmt.Errorf("listing the uploads: %w", err)
	}
	for _, rel := range files {
		err := addFileToZip(zw, backupUploadsPrefix+rel, filepath.Join(uploadDir, filepath.FromSlash(rel)), zip.Store, time.Time{})
		if errors.Is(err, fs.ErrNotExist) {
			// Deleted since it was listed.
			continue
		}
		if err != nil {
			return fmt.Errorf("adding %s: %w", rel, err)
		}
		manifest.Files++
	}
	w, err := zw.CreateHeader(&zip.FileHeader{Name: backupManifestName, Method: zip.Deflate, Modified: manifest.CreatedAt})
	if err != nil {
		return err
	}
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(manifest); err != nil {
		return err
	}
	if err := zw.Close(); err != nil {
		return err
	}
	if err := buffered.Flush(); err != nil {
		return err
	}
	return file.Sync()
}

func addFileToZip(zw *zip.Writer, name, source string, method uint16, modified time.Time) error {
	in, err := os.Open(source)
	if err != nil {
		return err
	}
	defer in.Close()
	if modified.IsZero() {
		if stat, err := in.Stat(); err == nil {
			modified = stat.ModTime()
		}
	}
	w, err := zw.CreateHeader(&zip.FileHeader{Name: name, Method: method, Modified: modified})
	if err != nil {
		return err
	}
	_, err = io.Copy(w, in)
	return err
}

// uploadFiles lists the uploads folder's files, as slash paths below it,
// leaving out a restore's staging folder.
func uploadFiles(uploadDir string) ([]string, error) {
	files := []string{}
	err := filepath.WalkDir(uploadDir, func(p string, entry fs.DirEntry, err error) error {
		if err != nil {
			if p == uploadDir && errors.Is(err, fs.ErrNotExist) {
				return filepath.SkipDir
			}
			return err
		}
		rel, err := filepath.Rel(uploadDir, p)
		if err != nil {
			return err
		}
		if entry.IsDir() {
			if rel == uploadsRestoringDir {
				return filepath.SkipDir
			}
			return nil
		}
		if entry.Type().IsRegular() {
			files = append(files, filepath.ToSlash(rel))
		}
		return nil
	})
	sort.Strings(files)
	return files, err
}

// List is the folder's backups, newest first.
func (s *BackupService) List() ([]BackupInfo, error) {
	entries, err := os.ReadDir(s.dir)
	if errors.Is(err, fs.ErrNotExist) {
		return []BackupInfo{}, nil
	}
	if err != nil {
		return nil, err
	}
	backups := []BackupInfo{}
	for _, entry := range entries {
		match := backupNamePattern.FindStringSubmatch(entry.Name())
		if entry.IsDir() || match == nil {
			continue
		}
		at, _ := time.ParseInLocation(backupTimeLayout, match[1], time.UTC)
		info := BackupInfo{Name: entry.Name(), Kind: match[2], CreatedAt: at}
		if manifest, err := readManifest(filepath.Join(s.dir, entry.Name())); err == nil {
			info = backupInfo(entry.Name(), match[2], manifest)
		}
		if stat, err := entry.Info(); err == nil {
			info.Size = stat.Size()
		}
		backups = append(backups, info)
	}
	sort.SliceStable(backups, func(i, j int) bool {
		a, b := backups[i], backups[j]
		return a.CreatedAt.After(b.CreatedAt) || a.CreatedAt.Equal(b.CreatedAt) && a.Name > b.Name
	})
	return backups, nil
}

// LastAutomatic is when the newest automatic backup was made, nil when none was.
func (s *BackupService) LastAutomatic() *time.Time {
	entries, err := os.ReadDir(s.dir)
	if err != nil {
		return nil
	}
	var last *time.Time
	for _, entry := range entries {
		match := backupNamePattern.FindStringSubmatch(entry.Name())
		if match == nil || match[2] != BackupKindAutomatic {
			continue
		}
		if at, err := time.ParseInLocation(backupTimeLayout, match[1], time.UTC); err == nil && (last == nil || at.After(*last)) {
			last = &at
		}
	}
	return last
}

// prune deletes the automatic backups beyond the number kept, the oldest
// first. The others are kept until deleted by hand.
func (s *BackupService) prune() {
	backups, err := s.List()
	if err != nil {
		return
	}
	kept := 0
	for _, backup := range backups {
		if backup.Kind != BackupKindAutomatic {
			continue
		}
		if kept++; kept <= s.Keep() {
			continue
		}
		if err := os.Remove(filepath.Join(s.dir, backup.Name)); err != nil {
			s.logger.Warn("backup", "Could not delete the old backup %s: %v", backup.Name, err)
		} else {
			s.logger.Info("backup", "Deleted the old backup %s", backup.Name)
		}
	}
}

// Path is where a backup of the folder is.
func (s *BackupService) Path(name string) (string, error) {
	if !backupNamePattern.MatchString(name) {
		return "", ErrBackupNotFound
	}
	full := filepath.Join(s.dir, name)
	if stat, err := os.Stat(full); err != nil || !stat.Mode().IsRegular() {
		return "", ErrBackupNotFound
	}
	return full, nil
}

// Delete deletes a backup.
func (s *BackupService) Delete(name string) error {
	full, err := s.Path(name)
	if err != nil {
		return err
	}
	return os.Remove(full)
}

// readManifest reads a backup's manifest, checking it is a backup the app
// can restore.
func readManifest(zipPath string) (BackupManifest, error) {
	zr, err := zip.OpenReader(zipPath)
	if err != nil {
		return BackupManifest{}, fmt.Errorf("not a zip file: %w", err)
	}
	defer zr.Close()
	return manifestOf(&zr.Reader)
}

func manifestOf(zr *zip.Reader) (BackupManifest, error) {
	var manifest BackupManifest
	for _, file := range zr.File {
		if file.Name != backupManifestName {
			continue
		}
		r, err := file.Open()
		if err != nil {
			return manifest, err
		}
		defer r.Close()
		if err := json.NewDecoder(io.LimitReader(r, 1<<20)).Decode(&manifest); err != nil {
			return manifest, fmt.Errorf("unreadable manifest: %w", err)
		}
		if manifest.Format != backupFormat {
			return manifest, errors.New("not a backup of the app")
		}
		if manifest.FormatVersion < 1 || manifest.FormatVersion > backupFormatVersion {
			return manifest, fmt.Errorf("a backup of format %d, which this version of the app does not know", manifest.FormatVersion)
		}
		return manifest, nil
	}
	return manifest, errors.New("not a backup of the app: it has no manifest")
}

// uploadEntryPath is where a backup's entry goes below the uploads folder;
// "" for an entry that is not an upload, or one that would land outside it.
func uploadEntryPath(name string) (string, bool) {
	if !strings.HasPrefix(name, backupUploadsPrefix) || strings.HasSuffix(name, "/") {
		return "", false
	}
	rel := strings.TrimPrefix(name, backupUploadsPrefix)
	if rel == "" || strings.Contains(rel, `\`) || path.IsAbs(rel) || path.Clean(rel) != rel || rel == ".." || strings.HasPrefix(rel, "../") {
		return "", false
	}
	if first, _, _ := strings.Cut(rel, "/"); first == uploadsRestoringDir {
		return "", false
	}
	return rel, true
}

// validate checks a backup can be restored: its manifest, its entries (the
// database, uploads within the folder) and its database.
func (s *BackupService) validate(zipPath string) (BackupManifest, error) {
	zr, err := zip.OpenReader(zipPath)
	if err != nil {
		return BackupManifest{}, fmt.Errorf("not a zip file: %w", err)
	}
	defer zr.Close()
	manifest, err := manifestOf(&zr.Reader)
	if err != nil {
		return manifest, err
	}
	var database *zip.File
	for _, file := range zr.File {
		switch {
		case file.Name == backupManifestName:
		case file.Name == backupDatabaseName:
			database = file
		case strings.HasSuffix(file.Name, "/"):
			// A folder entry, as some zip tools add.
		default:
			if _, ok := uploadEntryPath(file.Name); !ok {
				return manifest, fmt.Errorf("unexpected file in the backup: %s", file.Name)
			}
		}
	}
	if database == nil {
		return manifest, errors.New("the backup has no database")
	}
	tmp, err := os.MkdirTemp("", "orla-numis-restore-check-")
	if err != nil {
		return manifest, err
	}
	defer os.RemoveAll(tmp)
	extracted := filepath.Join(tmp, "ancientcoins.db")
	if err := extractZipFile(database, extracted); err != nil {
		return manifest, fmt.Errorf("reading the database: %w", err)
	}
	if s.checkDatabase != nil {
		if err := s.checkDatabase(extracted); err != nil {
			return manifest, err
		}
	}
	return manifest, nil
}

func extractZipFile(file *zip.File, target string) (err error) {
	r, err := file.Open()
	if err != nil {
		return err
	}
	defer r.Close()
	out, err := os.OpenFile(target, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o640)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := out.Close(); err == nil {
			err = closeErr
		}
	}()
	_, err = io.Copy(out, r)
	return err
}

// SaveUpload keeps a backup sent from another place (a download of an
// earlier one), after checking it can be restored.
func (s *BackupService) SaveUpload(body io.Reader) (BackupInfo, error) {
	if err := os.MkdirAll(s.dir, 0o750); err != nil {
		return BackupInfo{}, err
	}
	tmp, err := os.CreateTemp(s.dir, ".upload-*.zip")
	if err != nil {
		return BackupInfo{}, err
	}
	defer os.Remove(tmp.Name())
	_, copyErr := io.Copy(tmp, body)
	closeErr := tmp.Close()
	if copyErr != nil {
		return BackupInfo{}, fmt.Errorf("receiving the file: %w", copyErr)
	}
	if closeErr != nil {
		return BackupInfo{}, closeErr
	}
	manifest, err := s.validate(tmp.Name())
	if err != nil {
		return BackupInfo{}, err
	}
	at := s.now().UTC().Truncate(time.Second)
	name := backupName(at, BackupKindUploaded)
	for exists(filepath.Join(s.dir, name)) {
		at = at.Add(time.Second)
		name = backupName(at, BackupKindUploaded)
	}
	if err := os.Rename(tmp.Name(), filepath.Join(s.dir, name)); err != nil {
		return BackupInfo{}, err
	}
	info := backupInfo(name, BackupKindUploaded, manifest)
	if stat, err := os.Stat(filepath.Join(s.dir, name)); err == nil {
		info.Size = stat.Size()
	}
	s.logger.Info("backup", "Imported the backup %s, made %s", name, manifest.CreatedAt.Format(time.RFC3339))
	return info, nil
}

type restoreRequest struct {
	Name        string    `json:"name"`
	RequestedAt time.Time `json:"requestedAt"`
}

// StageRestore prepares a backup to replace all the data at the next start:
// it checks the backup, backs up the data of now (to go back to it) and puts
// the backup next to the database. The app must then be restarted.
func (s *BackupService) StageRestore(name string) error {
	full, err := s.Path(name)
	if err != nil {
		return err
	}
	if !s.mu.TryLock() {
		return ErrBackupRunning
	}
	defer s.mu.Unlock()
	if _, err := s.validate(full); err != nil {
		return fmt.Errorf("the backup cannot be restored: %w", err)
	}
	if _, err := s.create(BackupKindPreRestore); err != nil {
		return fmt.Errorf("the data of now could not be backed up first: %w", err)
	}
	pending := filepath.Join(filepath.Dir(s.dbPath), restorePendingDir)
	if err := os.RemoveAll(pending); err != nil {
		return err
	}
	if err := os.MkdirAll(pending, 0o750); err != nil {
		return err
	}
	if err := copyFile(full, filepath.Join(pending, "backup.zip")); err != nil {
		os.RemoveAll(pending)
		return fmt.Errorf("preparing the restore: %w", err)
	}
	// The request last: a start finds a whole backup or none.
	request, _ := json.Marshal(restoreRequest{Name: name, RequestedAt: s.now().UTC()})
	if err := os.WriteFile(filepath.Join(pending, "request.json"), request, 0o640); err != nil {
		os.RemoveAll(pending)
		return err
	}
	s.logger.Info("backup", "Restore of %s prepared; it is applied when the app restarts", name)
	return nil
}

func copyFile(source, target string) (err error) {
	in, err := os.Open(source)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.OpenFile(target, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o640)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := out.Close(); err == nil {
			err = closeErr
		}
	}()
	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Sync()
}

// LastRestore is how the last restore went, nil when there was none.
func (s *BackupService) LastRestore() *RestoreResult {
	data, err := os.ReadFile(filepath.Join(filepath.Dir(s.dbPath), restoreResultName))
	if err != nil {
		return nil
	}
	var result RestoreResult
	if json.Unmarshal(data, &result) != nil {
		return nil
	}
	return &result
}

// ApplyPendingRestore restores the backup StageRestore prepared, if any: it
// must run at the start, before the database is opened. The database file
// is replaced, then every file of the uploads folder. The result is kept
// next to the database (see LastRestore) and returned; nil when there was
// nothing to restore. The prepared backup is removed either way, so a
// failed restore is not tried again at every start.
func ApplyPendingRestore(dbPath, uploadDir string) *RestoreResult {
	pending := filepath.Join(filepath.Dir(dbPath), restorePendingDir)
	data, err := os.ReadFile(filepath.Join(pending, "request.json"))
	if err != nil {
		return nil
	}
	var request restoreRequest
	_ = json.Unmarshal(data, &request)
	result := RestoreResult{Name: request.Name, At: time.Now().UTC(), OK: true}
	if err := applyRestore(filepath.Join(pending, "backup.zip"), dbPath, uploadDir); err != nil {
		result.OK, result.Error = false, err.Error()
	}
	if data, err := json.Marshal(result); err == nil {
		_ = os.WriteFile(filepath.Join(filepath.Dir(dbPath), restoreResultName), data, 0o640)
	}
	_ = os.RemoveAll(pending)
	return &result
}

func applyRestore(zipPath, dbPath, uploadDir string) error {
	absUploads, err := filepath.Abs(uploadDir)
	if err != nil || absUploads == string(filepath.Separator) || absUploads == filepath.Dir(absUploads) {
		return fmt.Errorf("refusing to replace the uploads folder %q", uploadDir)
	}
	zr, err := zip.OpenReader(zipPath)
	if err != nil {
		return fmt.Errorf("not a zip file: %w", err)
	}
	defer zr.Close()
	if _, err := manifestOf(&zr.Reader); err != nil {
		return err
	}

	// Everything is unpacked beside what it replaces (the same disk, for the
	// renames) before anything is replaced.
	staging := filepath.Join(uploadDir, uploadsRestoringDir)
	if err := os.RemoveAll(staging); err != nil {
		return err
	}
	if err := os.MkdirAll(staging, 0o750); err != nil {
		return err
	}
	defer os.RemoveAll(staging)
	restoredDB := dbPath + ".restoring"
	defer os.Remove(restoredDB)
	hasDatabase := false
	for _, file := range zr.File {
		switch {
		case file.Name == backupDatabaseName:
			if err := extractZipFile(file, restoredDB); err != nil {
				return fmt.Errorf("unpacking the database: %w", err)
			}
			hasDatabase = true
		case strings.HasPrefix(file.Name, backupUploadsPrefix) && !strings.HasSuffix(file.Name, "/"):
			rel, ok := uploadEntryPath(file.Name)
			if !ok {
				return fmt.Errorf("unexpected file in the backup: %s", file.Name)
			}
			target := filepath.Join(staging, filepath.FromSlash(rel))
			if err := os.MkdirAll(filepath.Dir(target), 0o750); err != nil {
				return err
			}
			if err := extractZipFile(file, target); err != nil {
				return fmt.Errorf("unpacking %s: %w", rel, err)
			}
		}
	}
	if !hasDatabase {
		return errors.New("the backup has no database")
	}
	header := make([]byte, 16)
	if in, err := os.Open(restoredDB); err == nil {
		_, _ = io.ReadFull(in, header)
		in.Close()
	}
	if string(header) != "SQLite format 3\x00" {
		return errors.New("the backup's database is not a SQLite database")
	}

	// The database: its journal files go with the file they belong to.
	for _, suffix := range []string{"-wal", "-shm"} {
		if err := os.Remove(dbPath + suffix); err != nil && !errors.Is(err, fs.ErrNotExist) {
			return err
		}
	}
	if err := os.Rename(restoredDB, dbPath); err != nil {
		return fmt.Errorf("replacing the database: %w", err)
	}
	// Then the uploads: the folder's files make way for the backup's.
	entries, err := os.ReadDir(uploadDir)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if entry.Name() == uploadsRestoringDir {
			continue
		}
		if err := os.RemoveAll(filepath.Join(uploadDir, entry.Name())); err != nil {
			return fmt.Errorf("replacing the uploads: %w", err)
		}
	}
	restored, err := os.ReadDir(staging)
	if err != nil {
		return err
	}
	for _, entry := range restored {
		if err := os.Rename(filepath.Join(staging, entry.Name()), filepath.Join(uploadDir, entry.Name())); err != nil {
			return fmt.Errorf("replacing the uploads: %w", err)
		}
	}
	return nil
}
