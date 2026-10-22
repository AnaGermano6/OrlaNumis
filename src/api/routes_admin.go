package main

import (
	"os"
	"time"

	"github.com/briandenicola/ancient-coins-api/database"
	"github.com/briandenicola/ancient-coins-api/handlers"
	"github.com/briandenicola/ancient-coins-api/middleware"
	"github.com/briandenicola/ancient-coins-api/repository"
	"github.com/briandenicola/ancient-coins-api/services"
	"github.com/gin-gonic/gin"
)

// registerAdminRoutes registers the /api/admin endpoints, which require the
// admin role.
func registerAdminRoutes(api *gin.RouterGroup, d *appDeps) {
	admin := api.Group("/admin")
	admin.Use(middleware.AuthRequired(d.cfg.JWTSecret))
	admin.Use(handlers.AdminRequired())
	{
		adminRepo := repository.NewAdminRepository(database.DB)
		adminRecoverySvc := services.NewAdminRecoveryService(adminRepo, d.securitySvc)
		adminHandler := handlers.NewAdminHandler(d.cfg.UploadDir, adminRepo, adminRecoverySvc, d.agentProxy, d.settingsSvc, d.logger)
		admin.GET("/users", adminHandler.ListUsers)
		admin.DELETE("/users/:id", adminHandler.DeleteUser)
		admin.POST("/users/:id/reset-password", adminHandler.ResetPassword)
		admin.PUT("/users/:id/role", adminHandler.UpdateUserRole)
		admin.GET("/settings", adminHandler.GetSettings)
		// Full backups: the whole database and every upload.
		backups := newAdminBackupHandler(d)
		admin.GET("/backups", backups.List)
		admin.POST("/backups", backups.Create)
		admin.GET("/backups/:name", backups.Download)
		admin.DELETE("/backups/:name", backups.Delete)
		admin.POST("/backups/:name/restore", backups.Restore)
		admin.GET("/settings/defaults", adminHandler.GetSettingDefaults)
		admin.PUT("/settings", adminHandler.UpdateSettings)
		admin.GET("/logs", adminHandler.GetLogs)
		collectionCurrencyHandler := handlers.NewCollectionCurrencyHandler(services.NewCollectionCurrencyService(repository.NewCollectionCurrencyRepository(database.DB), d.settingsSvc, d.exchangeRates))
		admin.POST("/collection-currency", collectionCurrencyHandler.Change)
		adminNumistaHandler := handlers.NewAdminNumistaHandler(d.numistaTelemetry, d.settingsSvc).WithUsage(d.numistaUsage)
		admin.GET("/numista/health", adminNumistaHandler.Health)
		admin.GET("/test-anthropic", adminHandler.TestAnthropicConnection)
		admin.GET("/test-searxng", adminHandler.TestSearXNGConnection)

		securityAdminHandler := handlers.NewSecurityAdminHandler(d.securitySvc, d.settingsSvc, handlers.SecurityExposureConfig{
			PublicAppURL:             d.settingsSvc.GetSetting(services.SettingPublicAppURL),
			CORSOrigins:              d.cfg.AllowedOrigins(),
			TrustedProxiesConfigured: d.cfg.TrustedProxies != "",
			AgentInternalTokenSet:    d.cfg.AgentInternalServiceToken != "",
			RegistrationMode:         d.settingsSvc.GetSetting(services.SettingRegistrationMode),
			BackupStatus:             d.settingsSvc.GetSetting(services.SettingBackupStatus),
		})
		admin.GET("/security/summary", securityAdminHandler.SecuritySummary)
		admin.GET("/security/events", securityAdminHandler.SecurityEvents)
		admin.POST("/users/:id/unlock", securityAdminHandler.UnlockUser)
		admin.GET("/security/exposure-check", securityAdminHandler.ExposureCheck)

		// Catalog registry management (shared handler from protected scope)
		catalogRegistryRepo := repository.NewCatalogRegistryRepository(database.DB)
		catalogRegistrySvc := services.NewCatalogRegistryService(catalogRegistryRepo)
		catalogRegistryHandler := handlers.NewCatalogRegistryHandler(catalogRegistrySvc)
		admin.POST("/catalogs", catalogRegistryHandler.Create)
		admin.PUT("/catalogs/:id", catalogRegistryHandler.Update)
		admin.DELETE("/catalogs/:id", catalogRegistryHandler.Delete)

		// Mint location management
		mintLocationRepo := repository.NewMintLocationRepository(database.DB)
		mintLocationSvc := services.NewMintLocationService(mintLocationRepo)
		mintLocationHandler := handlers.NewMintLocationHandler(mintLocationSvc)
		admin.POST("/mint-locations", mintLocationHandler.Create)
		admin.PUT("/mint-locations/:id", mintLocationHandler.Update)
		admin.DELETE("/mint-locations/:id", mintLocationHandler.Delete)

		// Availability check run history and manual trigger (reuse outer scope services)
		adminAvailHandler := handlers.NewAvailabilityHandler(nil, d.availScheduler, d.availRepo, nil).WithCycleRepo(d.availCycleRepo)
		admin.GET("/availability-runs", adminAvailHandler.ListRuns)
		admin.GET("/availability-runs/:id", adminAvailHandler.GetRunDetail)
		admin.POST("/availability/run", adminAvailHandler.TriggerRun)
		admin.GET("/availability-cycles", adminAvailHandler.ListCycles)
		admin.GET("/availability-cycles/:id", adminAvailHandler.GetCycleDetail)
		wishlistSearchAlertAdminHandler := handlers.NewWishlistSearchAlertHandler(d.wishlistSearchAlertSvc)
		admin.GET("/wishlist-search-alert-runs", wishlistSearchAlertAdminHandler.ListAdminRuns)

		// Valuation run history and manual trigger
		valAdminHandler := handlers.NewValuationAdminHandler(d.valRepo, d.valSvc, d.logger)
		admin.GET("/valuation-runs", valAdminHandler.ListRuns)
		admin.GET("/valuation-runs/:id", valAdminHandler.GetRunDetail)
		admin.POST("/valuation-runs/trigger", valAdminHandler.TriggerValuation)
		admin.POST("/valuation-runs/:id/cancel", valAdminHandler.CancelValuation)

		// The daily batches of coin values from Numista's prices
		numistaValuesHandler := handlers.NewAdminNumistaValuesHandler(d.numistaValueSvc, d.numistaValueScheduler, d.logger)
		admin.GET("/numista-values", numistaValuesHandler.Status)
		admin.POST("/numista-values/run", numistaValuesHandler.Run)

		// Aggregate health metrics
		adminHealthHandler := handlers.NewAdminHealthHandler(d.healthSvc, d.logger)
		admin.GET("/health/summary", adminHealthHandler.Summary)

		// Deep Identification aggregate operational metrics
		adminDeepIdentificationHandler := handlers.NewAdminDeepIdentificationHandler(d.deepIdentificationSvc)
		admin.GET("/deep-identification/observability", adminDeepIdentificationHandler.Observability)
	}
}

// backupUploadLimit bounds an imported backup: the database and every photo.
const backupUploadLimit int64 = 20 << 30

func newAdminBackupHandler(d *appDeps) *handlers.AdminBackupHandler {
	return handlers.NewAdminBackupHandler(d.backupSvc, d.backupScheduler, d.settingsSvc, d.cfg.BackupDirLabel, restartApp(d.logger))
}

// registerBackupUploadRoute takes an imported backup, apart from the /api
// group, whose body limit is far below a backup's size.
func registerBackupUploadRoute(r *gin.Engine, d *appDeps) {
	r.PUT("/api/admin/backups/upload",
		middleware.AuthRequired(d.cfg.JWTSecret),
		handlers.AdminRequired(),
		middleware.RequestBodyLimit(backupUploadLimit),
		newAdminBackupHandler(d).Upload)
}

// restartApp stops the app as SIGINT would, a second after it is called (for
// the response to go out): its container starts it again, and a restore
// prepared by then is applied before the database is opened.
func restartApp(logger *services.Logger) func() {
	return func() {
		go func() {
			time.Sleep(time.Second)
			logger.Info("backup", "Restarting to restore a backup")
			if process, err := os.FindProcess(os.Getpid()); err == nil {
				_ = process.Signal(os.Interrupt)
			}
		}()
	}
}
