package database

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/briandenicola/ancient-coins-api/models"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func Connect(dbPath string) {
	var err error
	// Root cause of "database is locked (SQLITE_BUSY)" on job claim: every
	// GORM db.Transaction() begins a *deferred* SQLite transaction, which
	// takes no lock until its first statement, and only a *read* lock on
	// the initial SELECT. When the later UPDATE in the same transaction
	// tries to upgrade that read lock to a write lock, and a concurrent
	// writer has taken the write lock (or already committed a change to
	// the WAL) in the meantime, SQLite has no choice but to fail the
	// upgrade immediately (SQLITE_BUSY / SQLITE_BUSY_SNAPSHOT) - a
	// deferred transaction cannot "wait" its way out of a lock upgrade
	// conflict, because waiting could deadlock two transactions that each
	// hold the other's needed read lock. This is exactly the
	// SELECT-then-UPDATE shape of ClaimNextQueuedJob, and is why it
	// reproduces reliably under concurrent claims (verified with a
	// standalone probe: 30 goroutines racing 300 claims against a real
	// on-disk WAL db hit ~29 SQLITE_BUSY/SQLITE_BUSY_SNAPSHOT errors).
	//
	// busy_timeout alone (a PRAGMA that only makes an *outright lock
	// acquisition* wait, not a lock-upgrade conflict) measurably did NOT
	// fix this in the same probe (still ~29 errors) - it is necessary but
	// not sufficient here, confirming the deferred-transaction upgrade
	// race, not a slow writer, is the true cause.
	//
	// The fix is `_txlock=immediate`: every transaction acquires SQLite's
	// write lock (RESERVED) at BEGIN time instead of deferring and later
	// trying to upgrade, so competing writers simply queue for the lock
	// instead of racing to upgrade an existing read lock. Combined with
	// busy_timeout so that queuing waits (default 0ms) rather than
	// failing immediately, this eliminated all errors in the same probe
	// (0/300 failures, all claims eventually succeeded).
	//
	// Both settings are per-connection SQLite driver options, not
	// something a one-off PRAGMA exec after Open can guarantee across a
	// pooled *sql.DB (each pooled connection needs it applied as it is
	// opened), so they are encoded in the DSN rather than exec'd once.
	//
	// This is a global change: every scheduler in this process (deep
	// identification, valuation, health, coin-of-day, wishlist, ...) shares this one *gorm.DB/*sql.DB. Making every write
	// transaction acquire its lock immediately plus wait up to 5s under
	// contention is strictly safer for all of them than the previous
	// fail-fast default - none of those schedulers currently retry a bare
	// write failure themselves, and none holds a transaction open long
	// enough for a 5s busy_timeout to introduce a meaningful stall.
	//
	// models.SQLiteConcurrencyDSNParams is the single source of truth for
	// these params - repository/deep_identification_repository_test.go's
	// concurrency regression test derives its DSN from the same constant
	// (it cannot import this package; see TestNoDirectDatabaseImports in
	// architecture_test.go), so the two can never silently drift apart.
	dsn := dbPath + "?" + models.SQLiteConcurrencyDSNParams
	DB, err = gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Enable WAL mode for better concurrent performance
	for _, statement := range []string{"PRAGMA journal_mode=WAL", "PRAGMA foreign_keys=ON"} {
		if err := DB.Exec(statement).Error; err != nil {
			log.Fatalf("Failed to configure database: %v", err)
		}
	}

	// Migrate certainty → invoice_number column in coin_references (idempotent)
	if err := migrateCoinReferenceCertaintyColumn(DB); err != nil {
		log.Fatalf("Failed to migrate coin_references column: %v", err)
	}

	err = DB.AutoMigrate(&models.User{}, &models.MintLocation{}, &models.Coin{}, &models.CoinImage{}, &models.CoinReference{}, &models.CatalogRegistry{}, &models.AppSetting{}, &models.RefreshToken{}, &models.SecurityEvent{}, &models.CoinJournal{}, &models.CoinIntakeDraft{}, &models.QuickCaptureDraft{}, &models.QuickCaptureDraftImage{}, &models.QuickCaptureDraftReference{}, &models.DraftLifecycleEvent{}, &models.AgentConversation{}, &models.CollectionUpdateProposal{}, &models.SetBuilderRun{}, &models.SetProposal{}, &models.ProposalSlot{}, &models.CoinValueHistory{}, &models.AvailabilityCycle{}, &models.AvailabilityRun{}, &models.AvailabilityResult{}, &models.WishlistSearchAlert{}, &models.AlertRun{}, &models.AlertCandidate{}, &models.CandidateProvenance{}, &models.CandidateReviewAction{}, &models.AIJob{}, &models.CoinSet{}, &models.CoinSetMembership{}, &models.CoinSetTarget{}, &models.CoinSetValuationSnapshot{}, &models.CoinRecommendation{}, &models.RecommendationFeedback{}, &models.ValuationRun{}, &models.ValuationResult{}, &models.DeepIdentificationEvent{}, &models.DeepIdentificationProviderRun{}, &models.DeepIdentificationArtifact{}, &models.CoinCopilotThread{}, &models.CoinCopilotRun{}, &models.CoinCopilotCheckpoint{}, &models.CoinCopilotEvent{}, &models.CoinCopilotResumeRequest{})
	if err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}
	// Feature 362 Release B remains a separately named, additive migration so
	// its order after the deployed compatibility interlock is executable and
	// regression-testable. AutoMigrate adds only nullable columns/new tables;
	// operational rollback deliberately preserves both.
	if err := DB.AutoMigrate(&models.ImageCleanup{}); err != nil {
		log.Fatalf("Failed to migrate image cleanup records: %v", err)
	}
	if err := backfillCoinYears(DB); err != nil {
		log.Fatalf("Failed to backfill coin years: %v", err)
	}

	// Numista market-value cache: a new standalone table, purely additive.
	if err := DB.AutoMigrate(&models.NumistaCacheEntry{}); err != nil {
		log.Fatalf("Failed to migrate Numista cache: %v", err)
	}
	// Numista requests by month: a new standalone table.
	if err := DB.AutoMigrate(&models.NumistaUsage{}); err != nil {
		log.Fatalf("Failed to migrate Numista usage: %v", err)
	}
	// The daily batches of coin values from Numista's prices: new standalone tables.
	if err := DB.AutoMigrate(&models.NumistaValueRun{}, &models.NumistaValueCheck{}); err != nil {
		log.Fatalf("Failed to migrate Numista value batches: %v", err)
	}
	if err := seedNumistaUsage(DB, time.Now()); err != nil {
		log.Printf("database: could not estimate this month's Numista requests: %v", err)
	}
	// The coins each user hid in the catalogue, or joined to another: a new standalone table.
	if err := DB.AutoMigrate(&models.CatalogueOverride{}); err != nil {
		log.Fatalf("Failed to migrate catalogue overrides: %v", err)
	}
	if err := migrateFeature362(DB); err != nil {
		log.Fatalf("Failed to migrate Feature 362 schema: %v", err)
	}
	if err := dropRetiredTables(DB); err != nil {
		log.Fatalf("Failed to drop retired tables: %v", err)
	}
	if err := dropRetiredUserColumns(DB); err != nil {
		log.Fatalf("Failed to drop retired user columns: %v", err)
	}
	if err := dropRetiredCoinColumns(DB); err != nil {
		log.Fatalf("Failed to drop retired coin columns: %v", err)
	}
	if err := dropRetiredSetColumns(DB); err != nil {
		log.Fatalf("Failed to drop retired set columns: %v", err)
	}
	if err := dropRetiredMintLocationColumns(DB); err != nil {
		log.Fatalf("Failed to drop retired mint location columns: %v", err)
	}
	if err := dropRetiredCriteriaRules(DB); err != nil {
		log.Fatalf("Failed to drop retired smart set rules: %v", err)
	}
	if err := dropRetiredSettings(DB); err != nil {
		log.Fatalf("Failed to drop retired settings: %v", err)
	}
	if err := dropRetiredCategories(DB); err != nil {
		log.Fatalf("Failed to drop retired categories: %v", err)
	}
	if err := retireOldSearchSources(DB); err != nil {
		log.Printf("database: could not retire the old search sources: %v", err)
	}
	if err := retireOldDefaultPrompts(DB); err != nil {
		log.Fatalf("Failed to retire the old default prompts: %v", err)
	}

	// 353-wishlist-availability-run-observability: AvailabilityCycle is a brand-new table and
	// AvailabilityRun.CycleID is a brand-new nullable column. This AutoMigrate call is the
	// entire "migration" — purely additive DDL, no backfill/reparenting/retagging of any
	// pre-existing AvailabilityRun row (legacy UserID=0 admin rows keep their historical
	// TriggerType and are never attached to a synthesized cycle).
	logLegacyAvailabilityRunCount(DB)
	// 344-deep-agentic-coin-identification: partial unique index enforcing
	// at most one obverse and one reverse artifact per job. Hint artifacts
	// (role='hint') are explicitly excluded - up to 3 are allowed per job,
	// enforced in service-layer validation instead (data-model.md §5).
	if err := DB.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS uix_deep_artifact_job_role ON deep_identification_artifacts(job_id, role) WHERE role <> 'hint'`).Error; err != nil {
		log.Fatalf("Failed to create deep identification artifact role index: %v", err)
	}
	if err := migrateCoinSetTypes(DB); err != nil {
		log.Fatalf("Failed to migrate coin set types: %v", err)
	}
	if err := backfillVendorInvoiceFromCoinReferences(DB); err != nil {
		log.Fatalf("Failed to backfill vendor invoice values: %v", err)
	}

	// Note: CurrentValueUpdatedAt is a new nullable time.Time column.
	// SQLite AutoMigrate adds it as a plain NULL column without FK constraints — safe additive change.

	if err := backfillRequiredDefaults(DB); err != nil {
		log.Fatalf("Failed required database backfill: %v", err)
	}

	// D2 source backfill + D4 cleanup: see backfillCoinValueHistorySources below.
	// D4 runs only after a successful backfill -- data-integrity gate (B2 fix).
	if err := backfillCoinValueHistorySources(DB); err != nil {
		log.Fatalf("database: D2 source backfill failed: %v", err)
	}
	// D4: one-time idempotent cleanup of legacy scheduled-valuation journal entries.
	// Safe: the scheduled valuation writer always writes a coin_value_history row first,
	// so no information is lost by removing the duplicate journal lines.
	// On-demand entries intentionally preserved.
	if result := DB.Exec("DELETE FROM coin_journals WHERE entry LIKE 'Scheduled AI Value Estimate: $%'"); result.Error != nil {
		log.Printf("database: D4 cleanup warning -- could not remove legacy scheduled valuation journal entries: %v", result.Error)
	} else if result.RowsAffected > 0 {
		log.Printf("database: removed %d legacy scheduled valuation journal entries (D4 cleanup)", result.RowsAffected)
	}

	if err := seedCatalogRegistry(DB); err != nil {
		log.Fatalf("Failed to seed catalog registry: %v", err)
	}
	if err := retireOldCatalogs(DB); err != nil {
		log.Fatalf("Failed to retire the old catalogs: %v", err)
	}
	if err := seedMintLocations(DB); err != nil {
		log.Fatalf("Failed to seed mint locations: %v", err)
	}
	if err := backfillCoinMintLocations(DB); err != nil {
		log.Fatalf("Failed to backfill coin mint locations: %v", err)
	}

	log.Println("Database connected and migrated")
}

func backfillRequiredDefaults(db *gorm.DB) error {
	for _, step := range []struct {
		name string
		sql  string
		args []interface{}
	}{
		{"completed identification retention", "UPDATE deep_identification_jobs SET expires_at=? WHERE status IN (?, ?)", []interface{}{models.DeepIdentificationNoExpirySentinel, models.DeepJobStatusCompleted, models.DeepJobStatusPartial}},
	} {
		if err := db.Exec(step.sql, step.args...).Error; err != nil {
			return fmt.Errorf("%s: %w", step.name, err)
		}
	}
	return nil
}

// numistaAnswerKeys start the cache entries that are answers of Numista's
// API; the same cache keeps the ECB's, the mints' and other pages.
var numistaAnswerKeys = []string{"search:", "type:", "issues:", "prices:", "issuer-euro:", "issuer-year:", "issuers:"}

// seedNumistaUsage starts this month's Numista request count, when it has
// none yet, from the answers cached since the month began: the requests
// made before the count began, at least.
func seedNumistaUsage(db *gorm.DB, now time.Time) error {
	now = now.UTC()
	month := now.Format("2006-01")
	var counted int64
	if err := db.Model(&models.NumistaUsage{}).Where("month = ?", month).Count(&counted).Error; err != nil || counted > 0 {
		return err
	}
	conditions := make([]string, 0, len(numistaAnswerKeys))
	args := make([]any, 0, len(numistaAnswerKeys))
	for _, prefix := range numistaAnswerKeys {
		conditions = append(conditions, "key LIKE ?")
		args = append(args, prefix+"%")
	}
	var cached int64
	start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	if err := db.Model(&models.NumistaCacheEntry{}).Where("fetched_at >= ?", start).
		Where(strings.Join(conditions, " OR "), args...).Count(&cached).Error; err != nil || cached == 0 {
		return err
	}
	return db.Create(&models.NumistaUsage{Month: month, Requests: int(cached), UpdatedAt: now}).Error
}

// retiredTables are the tables of features removed in 2026-10: value
// history, notes, auction tracking (lots, the calendar, price alerts, bid
// reminders and their run logs), shipments, followers (with the coins'
// comments and ratings), showcases, storage locations, saved smart-set
// criteria; then single sign-on (OIDC), passkeys, API keys and IP rules;
// notifications, with what only made them (purchase reminders, set
// milestone alerts, the coins a set had announced); the Coin of the Day,
// collection health snapshots, Quick Access pins, the collector profile, the
// emperor tracker and tags. A table comes before the retired tables it
// points at, so each drop passes the foreign key checks.
var retiredTables = []string{
	"price_alerts", "bid_reminders", "auction_lots", "auction_events",
	"auction_alert_runs", "auction_ending_runs", "auction_watch_bid_digest_runs",
	"shipment_events", "shipments",
	"notes", "value_snapshots",
	"follows", "coin_comments", "showcase_coins", "showcases", "storage_locations",
	"smart_criteria_templates",
	"external_identities", "o_id_c_auth_states", "o_id_c_providers", "web_authn_credentials",
	"api_keys", "ip_rules",
	"notifications", "purchase_reminders", "coin_set_milestone_alerts", "coin_set_checklist_keys",
	"featured_coins", "coin_of_day_runs", "collection_health_snapshots", "collection_health_snapshot_runs",
	"quick_access_pins", "collector_profiles", "roman_imperial_figure_highlights", "roman_imperial_figures",
	"coin_tags", "tags",
}

// dropRetiredTables drops the retired tables, all or none.
func dropRetiredTables(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		for _, table := range retiredTables {
			if err := tx.Exec("DROP TABLE IF EXISTS " + table).Error; err != nil {
				return fmt.Errorf("drop %s: %w", table, err)
			}
		}
		return nil
	})
}

// retiredUserColumns held the accounts and keys of the integrations removed
// in 2026-10 (NumisBids, CNG, ParcelApp, Pushover), whether the collection
// was public to followers, and the Coin of the Day's and the emperor
// tracker's preferences, removed with them.
var retiredUserColumns = []string{
	"numis_bids_username", "numis_bids_password", "cng_username", "cng_password",
	"parcel_app_api_key", "pushover_user_key", "pushover_enabled",
	"is_public",
	"coin_of_day_enabled", "coin_of_day_include_wishlist",
	"emperor_tracker_enabled", "emperor_tracker_show_usurpers", "emperor_tracker_show_empresses", "emperor_tracker_show_other_figures",
}

// dropRetiredUserColumns drops the retired columns of users, all or none.
func dropRetiredUserColumns(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		return dropColumns(tx, "users", retiredUserColumns)
	})
}

// retiredCoinColumns held a coin's storage location (and its tray slot), its
// sale, whether followers could see it and its emperor, removed in 2026-10;
// retiredCoinIndexes are their indexes, dropped first, as SQLite drops no
// indexed column.
var (
	retiredCoinColumns = []string{"storage_location_id", "storage_slot", "is_sold", "sold_price", "sold_date", "sold_to", "is_private", "roman_imperial_figure_id"}
	retiredCoinIndexes = []string{"idx_coins_storage_location_slot_unique", "idx_coins_roman_imperial_figure_id"}
)

// dropRetiredCoinColumns drops the retired columns of coins, all or none.
func dropRetiredCoinColumns(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		for _, index := range retiredCoinIndexes {
			if err := tx.Exec("DROP INDEX IF EXISTS " + index).Error; err != nil {
				return fmt.Errorf("drop index %s: %w", index, err)
			}
		}
		return dropColumns(tx, "coins", retiredCoinColumns)
	})
}

// retiredSetColumns held whether a set was public, removed with the
// showcases in 2026-10, and when it was pinned to Quick Access;
// retiredSetIndexes are their indexes, dropped first.
var (
	retiredSetColumns = []string{"is_public", "pinned_at"}
	retiredSetIndexes = []string{"idx_coin_sets_pinned_at"}
)

// dropRetiredSetColumns drops the retired columns of sets, all or none.
func dropRetiredSetColumns(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		for _, index := range retiredSetIndexes {
			if err := tx.Exec("DROP INDEX IF EXISTS " + index).Error; err != nil {
				return fmt.Errorf("drop index %s: %w", index, err)
			}
		}
		return dropColumns(tx, "coin_sets", retiredSetColumns)
	})
}

// retiredMintLocationColumns held a mint's link to Nomisma.org, removed in
// 2026-10.
var retiredMintLocationColumns = []string{"nomisma_uri", "nomisma_label", "nomisma_linked_at"}

// dropRetiredMintLocationColumns drops the retired columns of mint
// locations, all or none.
func dropRetiredMintLocationColumns(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		return dropColumns(tx, "mint_locations", retiredMintLocationColumns)
	})
}

// retiredCriteriaFields are smart-set criteria fields of the coin columns
// removed in 2026-10: no coin is sold, stored in a location or private any
// more, so their rules ("not sold", "in this location", "private") are left
// out of the sets.
var retiredCriteriaFields = map[string]bool{"isSold": true, "storageLocationId": true, "isPrivate": true}

// dropRetiredCriteriaRules takes the retired fields' rules out of the smart
// sets' criteria, which would otherwise no longer be valid.
func dropRetiredCriteriaRules(db *gorm.DB) error {
	var sets []models.CoinSet
	if err := db.Select("id", "smart_criteria").Where("smart_criteria IS NOT NULL").Find(&sets).Error; err != nil {
		return err
	}
	for _, set := range sets {
		if set.SmartCriteria == nil {
			continue
		}
		before, _ := json.Marshal(*set.SmartCriteria)
		cleaned := models.JSONObject(withoutRetiredRules(*set.SmartCriteria))
		after, _ := json.Marshal(cleaned)
		if string(before) == string(after) {
			continue
		}
		if err := db.Model(&models.CoinSet{}).Where("id = ?", set.ID).Update("smart_criteria", cleaned).Error; err != nil {
			return fmt.Errorf("set %d: %w", set.ID, err)
		}
	}
	return nil
}

// withoutRetiredRules is a criteria group without the retired fields' rules,
// nor the groups left empty by it.
func withoutRetiredRules(group map[string]interface{}) map[string]interface{} {
	rules, ok := group["rules"].([]interface{})
	if !ok {
		return group
	}
	kept := make([]interface{}, 0, len(rules))
	for _, raw := range rules {
		rule, ok := raw.(map[string]interface{})
		if !ok {
			kept = append(kept, raw)
			continue
		}
		if field, _ := rule["field"].(string); retiredCriteriaFields[field] {
			continue
		}
		if _, isGroup := rule["rules"]; isGroup {
			rule = withoutRetiredRules(rule)
			if inner, _ := rule["rules"].([]interface{}); len(inner) == 0 {
				continue
			}
		}
		kept = append(kept, rule)
	}
	group["rules"] = kept
	return group
}

// retiredSettings are the stored admin settings of features removed in
// 2026-10 (auction alerts, shipment tracking and its carriers' credentials,
// Pushover, API key rotation, the Coin of the Day, collection health
// snapshots, OCRE, the external tool server, purchase reminders), and the
// emperor list's seed version.
var retiredSettings = []string{
	"AuctionAlertsCheckEnabled", "AuctionAlertsCheckInterval", "AuctionAlertsCheckStartTime",
	"AuctionEndingCheckEnabled", "AuctionEndingCheckInterval", "AuctionEndingCheckStartTime",
	"AuctionWatchBidDigestEnabled", "AuctionWatchBidDigestInterval", "AuctionWatchBidDigestStartTime",
	"ShipmentSyncBatchSize", "ShipmentSyncEnabled", "ShipmentSyncInterval", "ShipmentSyncStartTime",
	"FedExAPIBaseURL", "FedExClientID", "FedExClientSecret", "FedExScope", "FedExTokenURL",
	"UPSAPIBaseURL", "UPSClientID", "UPSClientSecret", "UPSScope", "UPSTokenURL",
	"USPSAPIBaseURL", "USPSAPIKey", "USPSAPIKeyHeader",
	"ParcelAppEnabled", "PushoverAppToken",
	"APIKeyRotationCutoffAt",
	"CoinOfDayEnabled", "CoinOfDayStartTime", "CoinOfDayTimezone",
	"CollectionHealthSnapshotsEnabled", "CollectionHealthSnapshotsStartTime",
	"DeepIdentificationOCREEnabled", "DeepIdentificationOCRECallBudget",
	"ExternalToolServerEnabled",
	"ReminderCheckEnabled", "ReminderCheckStartTime",
	"RomanImperialFigureSeedVersion",
}

// dropRetiredSettings deletes the retired settings' stored values.
func dropRetiredSettings(db *gorm.DB) error {
	return db.Where("key IN ?", retiredSettings).Delete(&models.AppSetting{}).Error
}

// retiredCategories are the ancient coin categories removed in 2026-10: they
// leave the admin's category list and its labels, and a coin of one becomes
// "Other".
var retiredCategories = []string{"Roman", "Greek", "Byzantine"}

// dropRetiredCategories takes the retired categories out of the category
// list, its Portuguese labels and the coins, all or none.
func dropRetiredCategories(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		retired := map[string]bool{}
		for _, category := range retiredCategories {
			retired[category] = true
		}
		for key, name := range map[string]func(line string) string{
			"CoinCategories":         func(line string) string { return strings.TrimSpace(line) },
			"CoinCategoryLabelsPtPT": func(line string) string { return strings.TrimSpace(strings.SplitN(line, "=", 2)[0]) },
		} {
			var setting models.AppSetting
			if err := tx.Where("key = ?", key).Limit(1).Find(&setting).Error; err != nil {
				return err
			}
			if setting.Key == "" {
				continue
			}
			kept := []string{}
			for _, line := range strings.Split(setting.Value, "\n") {
				if !retired[name(line)] {
					kept = append(kept, line)
				}
			}
			if value := strings.Join(kept, "\n"); value != setting.Value {
				if err := tx.Model(&models.AppSetting{}).Where("key = ?", key).Update("value", value).Error; err != nil {
					return err
				}
			}
		}
		return tx.Model(&models.Coin{}).Where("category IN ?", retiredCategories).Update("category", string(models.CategoryOther)).Error
	})
}

// retiredDefaultPrompts are the SHA-256 digests of the AI prompts the app
// shipped with until 2026-10 (trimmed, with plain line ends), written for
// ancient Greek and Roman coins. Saving the AI settings stored them as they
// were; a prompt still exactly so is cleared, so that today's default (this
// collection's coins and banknotes, in Portuguese) applies. A prompt the
// admin changed is kept.
var retiredDefaultPrompts = map[string]string{
	"CoinSearchPrompt":     "e759e7a2a9cfcd0c4ae05e6edee4c58c430045c21ace16b27d59f795fabde335",
	"CoinShowsPrompt":      "8502061c93fe969609cc5ec45fbeca70942573fb3a1db6773affff7c61738445",
	"ValuationPrompt":      "7238a53f1ccac07a7d195cbe86c3b937b7c7395edd97c999d87c84c67428e26b",
	"ObversePrompt":        "0cc12f52694e457fe10893bf8a00aadc482ff8ddfe03454ce2896b43de3013c6",
	"ReversePrompt":        "563556bec5379317e31567ef9200de5c62c587124f8d780e20ec31249fb5c2ff",
	"TextExtractionPrompt": "8fda1b2b72a3ae3e7bb760e175d08b24149b66a37eaba2e21ac79d935bf4e74a",
}

// retireOldDefaultPrompts clears the stored prompts that are the old defaults.
func retireOldDefaultPrompts(db *gorm.DB) error {
	keys := make([]string, 0, len(retiredDefaultPrompts))
	for key := range retiredDefaultPrompts {
		keys = append(keys, key)
	}
	var settings []models.AppSetting
	if err := db.Where("key IN ?", keys).Find(&settings).Error; err != nil {
		return err
	}
	for _, setting := range settings {
		digest := sha256.Sum256([]byte(strings.TrimSpace(strings.ReplaceAll(setting.Value, "\r\n", "\n"))))
		if hex.EncodeToString(digest[:]) != retiredDefaultPrompts[setting.Key] {
			continue
		}
		if err := db.Model(&models.AppSetting{}).Where("key = ?", setting.Key).Update("value", "").Error; err != nil {
			return err
		}
	}
	return nil
}

// retiredSearchSources are the dealer and auction lists the app shipped with
// until 2026-10, for ancient coins (Forum Ancient Coins, Harlan J. Berk, CNG).
// Saving the admin settings stored them as they were; a list still exactly
// so is cleared, so that today's default (European sites of modern coins and
// banknotes) applies. A list the admin changed is kept.
var retiredSearchSources = map[string]string{
	"DealerSearchSources":  "vcoins.com\nma-shops.com\nforumancientcoins.com\ncatawiki.com\nhjbltd.com",
	"AuctionSearchSources": "numisbids.com\ncngcoins.com\nbiddr.com",
}

// retireOldSearchSources clears the stored search source lists that are the old defaults.
func retireOldSearchSources(db *gorm.DB) error {
	for key, retired := range retiredSearchSources {
		var setting models.AppSetting
		if err := db.Where("key = ?", key).Limit(1).Find(&setting).Error; err != nil {
			return err
		}
		if setting.Key == "" || searchSourceList(setting.Value) != retired {
			continue
		}
		if err := db.Model(&models.AppSetting{}).Where("key = ?", key).Update("value", "").Error; err != nil {
			return err
		}
	}
	return nil
}

// searchSourceList is a stored list of hostnames one per line, trimmed and
// in lower case, as the settings keep it.
func searchSourceList(value string) string {
	hosts := []string{}
	for _, line := range strings.Split(strings.ReplaceAll(value, "\r\n", "\n"), "\n") {
		if host := strings.ToLower(strings.TrimSuffix(strings.TrimSpace(line), ".")); host != "" {
			hosts = append(hosts, host)
		}
	}
	return strings.Join(hosts, "\n")
}

// dropColumns drops those of the columns the table still has.
func dropColumns(tx *gorm.DB, table string, columns []string) error {
	for _, column := range columns {
		if !tx.Migrator().HasColumn(table, column) {
			continue
		}
		if err := tx.Exec("ALTER TABLE " + table + " DROP COLUMN " + column).Error; err != nil {
			return fmt.Errorf("drop %s.%s: %w", table, column, err)
		}
	}
	return nil
}

func migrateFeature362(db *gorm.DB) error {
	return db.AutoMigrate(
		&models.DeepIdentificationJob{},
		&models.CoinCopilotDeepHandoff{},
	)
}

// logLegacyAvailabilityRunCount logs an informational (non-fatal) count of pre-existing legacy
// admin availability_runs rows (UserID = 0) for operator visibility only. This is diagnostic,
// not an invariant check — legacy UserID=0 rows are expected and valid; the UserID > 0
// invariant for *new* child runs is enforced at creation time by
// AvailabilityRepository.CreateChildRun, not by this boot-time scan of historical data.
func logLegacyAvailabilityRunCount(db *gorm.DB) {
	var count int64
	if err := db.Model(&models.AvailabilityRun{}).Where("user_id = ?", 0).Count(&count).Error; err != nil {
		log.Printf("Unable to count legacy availability runs: %v", err)
		return
	}
	log.Printf("Availability runs: %d legacy admin row(s) (user_id = 0) present", count)
}

// backfillCoinValueHistorySources repairs the source field on legacy coin_value_history rows.
//
// SQLite applies the column DEFAULT ('manual') to every pre-existing row at ALTER TABLE time,
// so a NULL/empty WHERE predicate matches zero legacy rows (B1 root cause). The AI-tier
// backfill therefore keys on source='manual' AND confidence IN ('high','medium','low'):
//   - correct: only the scheduled valuation writer produced rows with those confidence tiers
//   - safe: true manual edits carry confidence='manual' and will never match the IN list
//   - idempotent: post-deploy AI rows are written with source!='manual' and never re-matched
//
// Returns a non-nil error if either SQL statement fails. Connect() gates the D4 destructive
// journal cleanup on a nil return so no journal rows are deleted if the backfill fails (B2).
func backfillCoinValueHistorySources(db *gorm.DB) error {
	// AI-confidence rows came from the scheduled valuation writer; the ALTER TABLE default
	// stamped them 'manual' -- flip them to the correct attribution.
	if result := db.Exec("UPDATE coin_value_histories SET source='ai_scheduled' WHERE source='manual' AND confidence IN ('high','medium','low')"); result.Error != nil {
		return fmt.Errorf("ai_scheduled backfill: %w", result.Error)
	}

	// Defensive: stamp any row still holding NULL or empty source as manual.
	if result := db.Exec("UPDATE coin_value_histories SET source='manual' WHERE source IS NULL OR source=''"); result.Error != nil {
		return fmt.Errorf("manual default backfill: %w", result.Error)
	}
	return nil
}

const mintLocationSeedVersionKey = "MintLocationSeedVersion"
const currentMintLocationSeedVersion = "1"

func seedMintLocations(db *gorm.DB) error {
	var existingSetting models.AppSetting
	err := db.First(&existingSetting, "key = ?", mintLocationSeedVersionKey).Error
	if err == nil && existingSetting.Value == currentMintLocationSeedVersion {
		return nil
	}
	if err != nil && err != gorm.ErrRecordNotFound {
		return err
	}

	seed := []models.MintLocation{
		{DisplayName: "Rome", Lat: 41.9028, Lng: 12.4964, Region: "Italy", Aliases: models.StringList{"Roma", "Rome mint"}},
		{DisplayName: "Athens", Lat: 37.9838, Lng: 23.7275, Region: "Greece", Aliases: models.StringList{"Athenai", "Athenae", "Athens mint"}},
		{DisplayName: "Constantinople", Lat: 41.0082, Lng: 28.9784, Region: "Thrace", Aliases: models.StringList{"Byzantium", "Istanbul", "Constantinopolis", "Constantinople mint"}},
		{DisplayName: "Alexandria", Lat: 31.2001, Lng: 29.9187, Region: "Egypt", Aliases: models.StringList{"Alexandria Egypt", "Alexandria ad Aegyptum"}},
		{DisplayName: "Antioch", Lat: 36.2021, Lng: 36.1603, Region: "Syria", Aliases: models.StringList{"Antiochia", "Antioch on the Orontes", "Antioch mint"}},
		{DisplayName: "Syracuse", Lat: 37.0755, Lng: 15.2866, Region: "Sicily", Aliases: models.StringList{"Syracusa", "Syracuse Sicily"}},
		{DisplayName: "Trier", Lat: 49.7499, Lng: 6.6371, Region: "Gaul", Aliases: models.StringList{"Treveri", "Treves", "Augusta Treverorum"}},
		{DisplayName: "Lugdunum", Lat: 45.764, Lng: 4.8357, Region: "Gaul", Aliases: models.StringList{"Lyon", "Lyons", "Lugdunum Lyon"}},
		{DisplayName: "Siscia", Lat: 45.4872, Lng: 16.376, Region: "Pannonia", Aliases: models.StringList{"Sisak", "Siscia mint"}},
		{DisplayName: "Nicomedia", Lat: 40.7654, Lng: 29.9408, Region: "Bithynia", Aliases: models.StringList{"Nikomedia", "Izmit"}},
		{DisplayName: "Cyzicus", Lat: 40.3991, Lng: 27.7936, Region: "Mysia", Aliases: models.StringList{"Kyzikos", "Cyzicus mint"}},
		{DisplayName: "Carthage", Lat: 36.8528, Lng: 10.3233, Region: "Africa", Aliases: models.StringList{"Carthago", "Qart Hadasht"}},
		{DisplayName: "Thessalonica", Lat: 40.6401, Lng: 22.9444, Region: "Macedonia", Aliases: models.StringList{"Thessalonika", "Thessaloniki"}},
		{DisplayName: "Heraclea", Lat: 41.2797, Lng: 27.9553, Region: "Thrace", Aliases: models.StringList{"Heraclea Thraciae", "Herakleia"}},
		{DisplayName: "Aquileia", Lat: 45.7686, Lng: 13.3678, Region: "Italy", Aliases: models.StringList{"Aquileia mint"}},
		{DisplayName: "Arelate", Lat: 43.6766, Lng: 4.6278, Region: "Gaul", Aliases: models.StringList{"Arles", "Arelate Arles", "Constantina"}},
		{DisplayName: "Ephesus", Lat: 37.9393, Lng: 27.3416, Region: "Ionia", Aliases: models.StringList{"Ephesos", "Ephesus mint"}},
	}

	return db.Transaction(func(tx *gorm.DB) error {
		for _, entry := range seed {
			entry.NormalizedName = models.NormalizeMintLocationName(entry.DisplayName)
			var existing models.MintLocation
			err := tx.Where("normalized_name = ?", entry.NormalizedName).First(&existing).Error
			if err == nil {
				updates := map[string]interface{}{
					"display_name": entry.DisplayName,
					"lat":          entry.Lat,
					"lng":          entry.Lng,
					"region":       entry.Region,
					"aliases":      entry.Aliases,
				}
				if err := tx.Model(&existing).Updates(updates).Error; err != nil {
					return err
				}
				continue
			}
			if err != gorm.ErrRecordNotFound {
				return err
			}
			if err := tx.Create(&entry).Error; err != nil {
				return err
			}
		}
		return tx.Save(&models.AppSetting{Key: mintLocationSeedVersionKey, Value: currentMintLocationSeedVersion}).Error
	})
}

const coinMintLocationBackfillVersionKey = "CoinMintLocationBackfillVersion"
const currentCoinMintLocationBackfillVersion = "1"

// backfillCoinMintLocations links existing coins' free-text Mint values to
// a MintLocation (global or the coin owner's own private one) whenever the
// normalized text exactly matches a display name or alias. Coins with no
// match are left untouched (never auto-creates a mint location) - the user
// gets a chance to link or create one via the unlinked-mint nudge instead.
// Idempotent and versioned, run once per version bump like seedMintLocations.
func backfillCoinMintLocations(db *gorm.DB) error {
	var existingSetting models.AppSetting
	err := db.First(&existingSetting, "key = ?", coinMintLocationBackfillVersionKey).Error
	if err == nil && existingSetting.Value == currentCoinMintLocationBackfillVersion {
		return nil
	}
	if err != nil && err != gorm.ErrRecordNotFound {
		return err
	}

	var locations []models.MintLocation
	if err := db.Find(&locations).Error; err != nil {
		return err
	}

	type lookupKey struct {
		ownerUserID uint // 0 means global (visible to everyone)
		normalized  string
	}
	lookup := make(map[lookupKey]uint)
	for _, loc := range locations {
		var owner uint
		if loc.UserID != nil {
			owner = *loc.UserID
		}
		for key := range mintLocationBackfillKeys(loc) {
			lookup[lookupKey{ownerUserID: owner, normalized: key}] = loc.ID
		}
	}

	var coins []models.Coin
	if err := db.Where("mint_location_id IS NULL AND mint <> ''").Find(&coins).Error; err != nil {
		return err
	}

	return db.Transaction(func(tx *gorm.DB) error {
		for _, coin := range coins {
			normalized := models.NormalizeMintLocationName(coin.Mint)
			if normalized == "" {
				continue
			}
			matchID, ok := lookup[lookupKey{ownerUserID: 0, normalized: normalized}]
			if !ok {
				matchID, ok = lookup[lookupKey{ownerUserID: coin.UserID, normalized: normalized}]
			}
			if !ok {
				continue
			}
			if err := tx.Model(&models.Coin{}).Where("id = ?", coin.ID).Update("mint_location_id", matchID).Error; err != nil {
				return err
			}
		}
		return tx.Save(&models.AppSetting{Key: coinMintLocationBackfillVersionKey, Value: currentCoinMintLocationBackfillVersion}).Error
	})
}

func mintLocationBackfillKeys(loc models.MintLocation) map[string]bool {
	keys := make(map[string]bool, len(loc.Aliases)+1)
	name := loc.NormalizedName
	if name == "" {
		name = models.NormalizeMintLocationName(loc.DisplayName)
	}
	if name != "" {
		keys[name] = true
	}
	for _, alias := range loc.Aliases {
		if n := models.NormalizeMintLocationName(alias); n != "" {
			keys[n] = true
		}
	}
	return keys
}

func migrateCoinSetTypes(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec("UPDATE coin_sets SET set_type='goal' WHERE LOWER(set_type)='defined'").Error; err != nil {
			return err
		}
		if err := tx.Exec("UPDATE coin_sets SET set_type='standard' WHERE LOWER(set_type)='open'").Error; err != nil {
			return err
		}
		if err := tx.Exec("UPDATE coin_sets SET set_type='agentic', creation_mode='dynamic' WHERE LOWER(set_type) IN ('dynamic', 'tracker')").Error; err != nil {
			return err
		}
		if err := tx.Exec("UPDATE coin_sets SET creation_mode='manual' WHERE creation_mode IS NULL OR creation_mode=''").Error; err != nil {
			return err
		}
		return nil
	})
}

// migrateCoinReferenceCertaintyColumn renames certainty → invoice_number if needed (idempotent).
func migrateCoinReferenceCertaintyColumn(db *gorm.DB) error {
	var columns []struct {
		Name string
	}
	if err := db.Raw("PRAGMA table_info(coin_references)").Scan(&columns).Error; err != nil {
		// Table doesn't exist yet — fresh install, nothing to migrate
		return nil
	}

	hasCertainty := false
	hasInvoiceNumber := false
	for _, col := range columns {
		if col.Name == "certainty" {
			hasCertainty = true
		}
		if col.Name == "invoice_number" {
			hasInvoiceNumber = true
		}
	}

	// Rename only if old column exists and new one doesn't
	if hasCertainty && !hasInvoiceNumber {
		if err := db.Exec("ALTER TABLE coin_references RENAME COLUMN certainty TO invoice_number").Error; err != nil {
			return err
		}
		log.Println("Migrated coin_references.certainty → invoice_number")
	}

	return nil
}

const vendorInvoiceBackfillVersionKey = "VendorInvoiceBackfillVersion"
const currentVendorInvoiceBackfillVersion = "1"

// backfillVendorInvoiceFromCoinReferences copies existing invoice_number values
// from structured coin references into the new top-level coins.vendor_invoice
// field when vendor_invoice is blank. It is idempotent and versioned.
func backfillVendorInvoiceFromCoinReferences(db *gorm.DB) error {
	var existingSetting models.AppSetting
	err := db.First(&existingSetting, "key = ?", vendorInvoiceBackfillVersionKey).Error
	if err == nil && existingSetting.Value == currentVendorInvoiceBackfillVersion {
		return nil
	}
	if err != nil && err != gorm.ErrRecordNotFound {
		return err
	}

	// Fresh installs created from the new model no longer include invoice_number
	// on coin_references, so there is nothing to backfill.
	var columns []struct {
		Name string
	}
	if err := db.Raw("PRAGMA table_info(coin_references)").Scan(&columns).Error; err != nil {
		return err
	}
	hasInvoiceNumber := false
	for _, col := range columns {
		if col.Name == "invoice_number" {
			hasInvoiceNumber = true
			break
		}
	}

	return db.Transaction(func(tx *gorm.DB) error {
		if hasInvoiceNumber {
			if err := tx.Exec(`
				UPDATE coins
				SET vendor_invoice = (
					SELECT cr.invoice_number
					FROM coin_references cr
					WHERE cr.coin_id = coins.id
						AND cr.invoice_number IS NOT NULL
						AND TRIM(cr.invoice_number) <> ''
					ORDER BY cr.updated_at DESC, cr.id DESC
					LIMIT 1
				)
				WHERE (vendor_invoice IS NULL OR TRIM(vendor_invoice) = '')
					AND EXISTS (
						SELECT 1
						FROM coin_references cr
						WHERE cr.coin_id = coins.id
							AND cr.invoice_number IS NOT NULL
							AND TRIM(cr.invoice_number) <> ''
					)
			`).Error; err != nil {
				return err
			}
		}
		return tx.Save(&models.AppSetting{
			Key:   vendorInvoiceBackfillVersionKey,
			Value: currentVendorInvoiceBackfillVersion,
		}).Error
	})
}

// seedCatalogRegistry registers, at every start, the catalogues of modern
// coins and banknotes the app relies on (see retiredCatalogs).
func seedCatalogRegistry(db *gorm.DB) error {
	seed := []models.CatalogRegistry{
		{Catalog: "KM", DisplayName: "Krause-Mishler", Era: models.EraModern, VolumeRequired: false},
		{Catalog: "Y", DisplayName: "Y Number", Era: models.EraModern, VolumeRequired: false},
		{Catalog: "REDBOOK", DisplayName: "Red Book", Era: models.EraModern, VolumeRequired: false},
		{Catalog: "NGC", DisplayName: "NGC Certification", Era: models.EraModern, VolumeRequired: false},
		{Catalog: "Pick", DisplayName: "Standard Catalog of World Paper Money (Pick)", Era: models.EraModern, VolumeRequired: false},
		{Catalog: "Numista", DisplayName: "Numista", Era: models.EraModern, VolumeRequired: false},
	}

	for _, entry := range seed {
		var existing models.CatalogRegistry
		err := db.Where("catalog = ?", entry.Catalog).First(&existing).Error
		if err == nil {
			existing.DisplayName = entry.DisplayName
			existing.Era = entry.Era
			existing.VolumeRequired = entry.VolumeRequired
			if err := db.Save(&existing).Error; err != nil {
				return err
			}
			continue
		}
		if err != nil && err != gorm.ErrRecordNotFound {
			return err
		}
		if err := db.Create(&entry).Error; err != nil {
			return err
		}
	}

	return nil
}

// retiredCatalogs are the catalogues of ancient and medieval coins the app
// registered at every start until 2026-10; it now registers modern ones only.
var retiredCatalogs = []string{"RIC", "RPC", "SEAR", "CRAWFORD", "SNG", "SPINK", "DUPLESSY", "CNI", "CRAIG", "PRICE", "BM", "VENÈRA"}

const retiredCatalogsVersionKey = "RetiredCatalogsVersion"
const currentRetiredCatalogsVersion = "1"

// retireOldCatalogs removes the retired catalogues no reference uses (as
// deleting one in Admin would), once: one the admin registers again
// afterwards is kept.
func retireOldCatalogs(db *gorm.DB) error {
	var marker models.AppSetting
	err := db.First(&marker, "key = ?", retiredCatalogsVersionKey).Error
	if err == nil && marker.Value == currentRetiredCatalogsVersion {
		return nil
	}
	if err != nil && err != gorm.ErrRecordNotFound {
		return err
	}
	return db.Transaction(func(tx *gorm.DB) error {
		unused := "NOT EXISTS (SELECT 1 FROM coin_references WHERE LOWER(coin_references.catalog) = LOWER(catalog_registries.catalog))"
		if err := tx.Where("catalog IN ?", retiredCatalogs).Where(unused).Delete(&models.CatalogRegistry{}).Error; err != nil {
			return err
		}
		return tx.Save(&models.AppSetting{Key: retiredCatalogsVersionKey, Value: currentRetiredCatalogsVersion}).Error
	})
}

// backfillCoinYears fills Coin.Year from DateRange for coins saved before the
// column existed (or whose date could not be read before). Idempotent: rows
// whose date has no readable year are simply left NULL.
func backfillCoinYears(db *gorm.DB) error {
	var coins []models.Coin
	if err := db.Select("id", "date_range").Where("year IS NULL AND date_range <> ''").Find(&coins).Error; err != nil {
		return err
	}
	for _, coin := range coins {
		if year := models.DeriveCoinYear(coin.DateRange); year != nil {
			if err := db.Model(&models.Coin{}).Where("id = ?", coin.ID).Update("year", *year).Error; err != nil {
				return err
			}
		}
	}
	return nil
}
