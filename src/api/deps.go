package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/briandenicola/ancient-coins-api/config"
	"github.com/briandenicola/ancient-coins-api/database"
	"github.com/briandenicola/ancient-coins-api/docs"
	"github.com/briandenicola/ancient-coins-api/handlers"
	"github.com/briandenicola/ancient-coins-api/middleware"
	"github.com/briandenicola/ancient-coins-api/repository"
	"github.com/briandenicola/ancient-coins-api/services"
	"github.com/gin-gonic/gin"
)

// appDeps is the application's dependency container. It exists because the
// composition root previously lived as ~750 lines inside main(), where the
// route groups closed over 40+ locals and no part could be read or moved in
// isolation. Construction happens once in buildDeps; the register*Routes
// functions in routes_*.go read what they need from here.
//
// This is deliberately a plain hand-wired struct rather than a DI framework:
// the wiring order encodes real initialisation dependencies (deepIdentificationSvc
// must have its pipeline runner set before its workers start, availRepo must be
// given its cycle repo before the scheduler reads it), and hand-wiring keeps
// that order explicit and greppable.
type appDeps struct {
	cfg                    *config.Config
	logger                 *services.Logger
	settingsSvc            *services.SettingsService
	securitySvc            *services.SecurityService
	authHandler            *handlers.AuthHandler
	imageSvc               *services.ImageService
	imageHandler           *handlers.ImageHandler
	internalTokenSvc       *services.InternalTokenService
	numistaLookupSvc       *services.NumistaLookupService
	numistaTelemetry       *services.NumistaTelemetry
	numistaUsage           *services.NumistaUsageService
	backupSvc              *services.BackupService
	backupScheduler        *services.BackupScheduler
	numistaValueSvc        *services.NumistaValueService
	numistaValueScheduler  *services.NumistaValueScheduler
	exchangeRates          *services.ExchangeRateService
	authRateLimit          gin.HandlerFunc
	apiRateLimit           gin.HandlerFunc
	writeRateLimit         gin.HandlerFunc
	agentProxy             *services.AgentProxy
	coinRepo               *repository.CoinRepository
	journalRepo            *repository.JournalRepository
	wishlistURLHandler     *handlers.WishlistURLHandler
	collectionSvc          *services.CollectionToolsService
	availRepo              *repository.AvailabilityRepository
	availCycleRepo         *repository.AvailabilityCycleRepository
	availSvc               *services.AvailabilityService
	availScheduler         *services.AvailabilityScheduler
	valRepo                *repository.ValuationRepository
	valSvc                 *services.ValuationService
	aiJobSvc               *services.AIJobService
	setBuilderSvc          *services.SetBuilderService
	coinLookupSvc          *services.CoinLookupService
	deepIdentificationRepo *repository.DeepIdentificationRepository
	deepIdentificationSvc  *services.DeepIdentificationService
	coinCopilotRepo        *repository.CoinCopilotRepository
	coinCopilotSvc         *services.CoinCopilotService
	deepAnalysisHandoffSvc *services.DeepAnalysisHandoffService
	healthSvc              *services.HealthService
	wishlistSearchAlertSvc *services.WishlistSearchAlertService
	schedulerRegistry      *SchedulerRegistry
	timeMachineSvc         *services.TimeMachineService
	numistaClient          *services.HTTPNumistaClient
	deepProviderBudgets    *services.DeepProviderBudgetTracker
}

// buildDeps constructs every repository, service, scheduler, and shared piece
// of middleware the routes need. The returned CancelFunc stops context-aware
// background work and is handed to runServer for graceful shutdown.
//
// database.Connect must have been called before this runs.
func buildDeps(cfg *config.Config) (*appDeps, context.CancelFunc) {
	// Create logger and settings service
	logger := services.NewLogger(1000)
	settingsRepo := repository.NewSettingsRepository(database.DB)
	settingsSvc := services.NewSettingsService(settingsRepo)
	settingsSvc.SyncLogLevel(logger)
	numistaClock := services.NewSystemNumistaClock()
	numistaCache := services.NewNumistaCache(numistaClock, 500, 5000)
	numistaTelemetry := services.NewNumistaTelemetry(500)
	// ECB euro rate for collection values; cached alongside Numista responses.
	exchangeRates := services.NewExchangeRateService(services.NewHTTPECBClient(nil),
		services.NewNumistaResponseCache(repository.NewNumistaCacheRepository(database.DB), services.NewSystemNumistaClock(), logger))
	numistaQueryBuilder := services.NewNumistaQueryBuilder()
	numistaClient, err := services.NewHTTPNumistaClient(services.NumistaClientConfig{
		APIKey:        func() string { return settingsSvc.GetSetting(services.SettingNumistaAPIKey) },
		SearchTimeout: func() time.Duration { return settingsSvc.GetNumistaSettings().SearchTimeout },
		DetailTimeout: func() time.Duration { return settingsSvc.GetNumistaSettings().DetailTimeout },
	})
	if err != nil {
		log.Fatalf("Failed to configure Numista client: %v", err)
	}
	numistaUsage := services.NewNumistaUsageService(repository.NewNumistaUsageRepository(database.DB), settingsSvc, logger)
	numistaClient.WithUsage(numistaUsage)
	numistaLookupSvc := services.NewNumistaLookupService(
		numistaClient, numistaCache, services.NewNumistaV1Scorer(),
		numistaTelemetry, settingsSvc, numistaClock, numistaQueryBuilder,
	)

	// Create internal token service for Python agent callbacks
	internalTokenSvc := services.NewInternalTokenService(cfg.JWTSecret)
	// Deep identification provider-tool boundary (Phase 6, T051-T054):
	// job-scoped call-budget tracking, shared across the numista_search/
	// numista_detail internal tool endpoints.
	deepProviderBudgets := services.NewDeepProviderBudgetTracker()

	logger.Info("startup", "Application starting")
	logger.Info("startup", "Database connected: %s", cfg.DBPath)

	// Ensure upload directory exists
	if err := os.MkdirAll(cfg.UploadDir, 0755); err != nil {
		log.Fatalf("Failed to create upload directory: %v", err)
	}
	logger.Debug("startup", "Upload directory: %s", cfg.UploadDir)

	// Auth routes (public) — rate limited to prevent brute force
	authRepo := repository.NewAuthRepository(database.DB)
	securityRepo := repository.NewSecurityRepository(database.DB)
	securitySvc := services.NewSecurityService(securityRepo)
	authSvc := services.NewAuthService(authRepo, cfg.JWTSecret).WithSettings(settingsSvc).WithSecurity(securitySvc)
	authHandler := handlers.NewAuthHandler(cfg.JWTSecret, authRepo, authSvc)
	imageRepo := repository.NewImageRepository(database.DB)
	imageSvc := services.NewImageService(imageRepo, cfg.UploadDir)
	if err := imageSvc.RetryPendingCleanup(); err != nil {
		logger.Error("startup", "Pending image cleanup requires retry: %v", err)
	}
	imageHandler := handlers.NewImageHandler(cfg.UploadDir, imageRepo, imageSvc, logger)

	authRateLimit := middleware.RateLimit(10, 1*time.Minute)
	apiRateLimit := middleware.AuthenticatedRateLimit(600, 1*time.Minute)  // Authenticated browsing
	writeRateLimit := middleware.AuthenticatedRateLimit(30, 1*time.Minute) // Write operations

	// Protected routes
	agentProxy := services.NewAgentProxy(cfg.AgentServiceURL, cfg.AgentInternalServiceToken, logger)
	availRepo := repository.NewAvailabilityRepository(database.DB)
	availCycleRepo := repository.NewAvailabilityCycleRepository(database.DB)
	availRepo.WithCycleRepo(availCycleRepo)
	coinRepo := repository.NewCoinRepository(database.DB)
	wishlistSearchAlertRepo := repository.NewWishlistSearchAlertRepository(database.DB)
	valRepo := repository.NewValuationRepository(database.DB)
	userRepoForVal := repository.NewUserRepository(database.DB)
	availSvc := services.NewAvailabilityService(coinRepo, availRepo, agentProxy, settingsSvc, logger).WithCycleRepo(availCycleRepo)
	wishlistSearchAlertSvc := services.NewWishlistSearchAlertService(wishlistSearchAlertRepo).WithDiscovery(agentProxy, settingsSvc)
	wishlistSearchAlertSvc.StartWorkers(1)
	wishlistSearchAlertScheduler := services.NewWishlistSearchAlertScheduler(wishlistSearchAlertSvc, wishlistSearchAlertRepo, settingsSvc, logger)
	valSvc := services.NewValuationService(coinRepo, valRepo, agentProxy, userRepoForVal, settingsSvc, logger)
	aiJobRepo := repository.NewAIJobRepository(database.DB)
	aiJobSvc := services.NewAIJobService(aiJobRepo, agentProxy, userRepoForVal, settingsSvc, logger)
	aiJobSvc.StartWorkers(1)
	coinLookupSvc := services.NewCoinLookupService(agentProxy, settingsSvc, logger, numistaQueryBuilder)
	deepIdentificationRepo := repository.NewDeepIdentificationRepository(database.DB)
	deepIdentificationCatalogRegistryRepo := repository.NewCatalogRegistryRepository(database.DB)
	deepIdentificationSvc := services.NewDeepIdentificationService(deepIdentificationRepo, imageRepo, imageSvc, settingsSvc, logger, cfg.UploadDir)
	deepIdentificationSvc.SetProviderBudgetTracker(deepProviderBudgets)
	deepIdentificationSvc.SetInternalTokenService(internalTokenSvc)
	deepIdentificationSvc.SetPipelineRunner(services.NewDeepIdentificationPipelineRunner(
		agentProxy, deepIdentificationRepo, settingsSvc, internalTokenSvc, cfg.AgentInternalCallbackURL, logger, deepIdentificationSvc.Broker(), deepIdentificationCatalogRegistryRepo,
	).WithQuickEvidence(coinLookupSvc))
	// Cancelled on SIGTERM/SIGINT by shutdownRuntime so the deep-identification
	// workers and janitor unwind instead of dying mid-write.
	backgroundCtx, cancelBackground := context.WithCancel(context.Background())
	deepIdentificationSvc.StartWorkers(backgroundCtx)
	deepIdentificationSvc.StartJanitor(backgroundCtx)
	timeMachineRepo := repository.NewTimeMachineRepository(database.DB)
	timeMachineSvc := services.NewTimeMachineService(timeMachineRepo)
	healthRepo := repository.NewHealthRepository(database.DB)
	healthSvc := services.NewHealthService(healthRepo, logger).WithSettings(settingsSvc)

	// Create schedulers before routes so they can be passed to admin handlers
	availScheduler := services.NewAvailabilityScheduler(availSvc, coinRepo, availRepo, settingsSvc, logger).WithCycleRepo(availCycleRepo)
	availScheduler.StartWorkers(1)
	valScheduler := services.NewValuationScheduler(valSvc, coinRepo, valRepo, settingsSvc, logger)
	schedulerRegistry := &SchedulerRegistry{}
	schedulerRegistry.Register(availScheduler)
	schedulerRegistry.Register(valScheduler)
	schedulerRegistry.Register(wishlistSearchAlertScheduler)
	// The full backups: every day, and on demand from the admin pages.
	backupSvc := services.NewBackupService(repository.NewBackupRepository(database.DB), cfg.DBPath, cfg.UploadDir, cfg.BackupDir, docs.SwaggerInfo.Version, settingsSvc, logger, database.CheckBackupDatabase)
	backupScheduler := services.NewBackupScheduler(backupSvc, settingsSvc, logger)
	schedulerRegistry.Register(backupScheduler)
	// The coins' values from Numista's prices, a few a day within the quota.
	numistaValueRepo := repository.NewNumistaValueRepository(database.DB)
	numistaValueMarket := services.NewNumistaMarketValueService(numistaClient,
		services.NewNumistaResponseCache(repository.NewNumistaCacheRepository(database.DB), services.NewSystemNumistaClock(), logger))
	numistaValueSvc := services.NewNumistaValueService(numistaValueRepo, numistaValueMarket, numistaUsage, settingsSvc, logger)
	numistaValueScheduler := services.NewNumistaValueScheduler(numistaValueSvc, numistaValueRepo, settingsSvc, logger)
	schedulerRegistry.Register(numistaValueScheduler)

	// Create shared repositories for cross-group access
	journalRepo := repository.NewJournalRepository(database.DB)
	collectionProposalRepo := repository.NewCollectionUpdateRepository(database.DB)
	wishlistURLHandler := handlers.NewWishlistURLHandler(services.NewWishlistURLService(coinRepo, agentProxy, settingsSvc))
	collectionSvc := services.NewCollectionToolsService(coinRepo, collectionProposalRepo).WithSettingsSupport(settingsSvc)
	coinCopilotRepo := repository.NewCoinCopilotRepository(database.DB)
	coinCopilotSvc := services.NewCoinCopilotService(
		coinCopilotRepo, settingsSvc, agentProxy, internalTokenSvc, logger, cfg.AgentInternalCallbackURL,
	)
	quickCaptureRepo := repository.NewQuickCaptureRepository(database.DB)
	deepAnalysisHandoffSvc := services.NewDeepAnalysisHandoffService(
		coinCopilotRepo, deepIdentificationRepo, coinRepo, quickCaptureRepo,
		deepIdentificationSvc, settingsSvc, cfg.UploadDir,
	)
	coinCopilotSvc.StartWorkers(backgroundCtx)

	return &appDeps{
		cfg:                    cfg,
		logger:                 logger,
		settingsSvc:            settingsSvc,
		securitySvc:            securitySvc,
		authHandler:            authHandler,
		imageSvc:               imageSvc,
		imageHandler:           imageHandler,
		internalTokenSvc:       internalTokenSvc,
		numistaLookupSvc:       numistaLookupSvc,
		numistaTelemetry:       numistaTelemetry,
		numistaUsage:           numistaUsage,
		backupSvc:              backupSvc,
		backupScheduler:        backupScheduler,
		numistaValueSvc:        numistaValueSvc,
		numistaValueScheduler:  numistaValueScheduler,
		exchangeRates:          exchangeRates,
		authRateLimit:          authRateLimit,
		apiRateLimit:           apiRateLimit,
		writeRateLimit:         writeRateLimit,
		agentProxy:             agentProxy,
		coinRepo:               coinRepo,
		journalRepo:            journalRepo,
		wishlistURLHandler:     wishlistURLHandler,
		collectionSvc:          collectionSvc,
		availRepo:              availRepo,
		availCycleRepo:         availCycleRepo,
		availSvc:               availSvc,
		availScheduler:         availScheduler,
		valRepo:                valRepo,
		valSvc:                 valSvc,
		aiJobSvc:               aiJobSvc,
		coinLookupSvc:          coinLookupSvc,
		deepIdentificationRepo: deepIdentificationRepo,
		deepIdentificationSvc:  deepIdentificationSvc,
		coinCopilotRepo:        coinCopilotRepo,
		coinCopilotSvc:         coinCopilotSvc,
		deepAnalysisHandoffSvc: deepAnalysisHandoffSvc,
		healthSvc:              healthSvc,
		wishlistSearchAlertSvc: wishlistSearchAlertSvc,
		schedulerRegistry:      schedulerRegistry,
		timeMachineSvc:         timeMachineSvc,
		numistaClient:          numistaClient,
		deepProviderBudgets:    deepProviderBudgets,
	}, cancelBackground
}
