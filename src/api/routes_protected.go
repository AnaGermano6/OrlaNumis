package main

import (
	"github.com/briandenicola/ancient-coins-api/database"
	"github.com/briandenicola/ancient-coins-api/handlers"
	"github.com/briandenicola/ancient-coins-api/middleware"
	"github.com/briandenicola/ancient-coins-api/repository"
	"github.com/briandenicola/ancient-coins-api/services"
	"github.com/gin-gonic/gin"
	"time"
)

// registerProtectedRoutes registers every endpoint that requires a signed-in
// user (a JWT bearer token).
func registerProtectedRoutes(api *gin.RouterGroup, d *appDeps) {
	protected := api.Group("")
	protected.Use(middleware.AuthRequired(d.cfg.JWTSecret))
	protected.Use(d.apiRateLimit)
	{
		coinReferenceRepo := repository.NewCoinReferenceRepository(database.DB)
		mintLocationRepo := repository.NewMintLocationRepository(database.DB)
		catalogRegistryRepo := repository.NewCatalogRegistryRepository(database.DB)
		intakeDraftRepo := repository.NewCoinIntakeDraftRepository(database.DB)
		quickCaptureRepo := repository.NewQuickCaptureRepository(database.DB)
		coinReferenceSvc := services.NewCoinReferenceService(coinReferenceRepo, catalogRegistryRepo)
		referenceMigrationSvc := services.NewReferenceMigrationService(database.DB, coinReferenceRepo, catalogRegistryRepo, d.journalRepo)
		catalogRegistrySvc := services.NewCatalogRegistryService(catalogRegistryRepo)
		catalogRegistryHandler := handlers.NewCatalogRegistryHandler(catalogRegistrySvc)
		coinSvc := services.NewCoinService(d.coinRepo).WithReferenceSupport(coinReferenceRepo, coinReferenceSvc).WithMintLocationSupport(mintLocationRepo).WithCatalogRegistrySupport(catalogRegistryRepo).WithSettingsSupport(d.settingsSvc)
		d.wishlistSearchAlertSvc.WithCoinCreation(coinSvc)
		coinHandler := handlers.NewCoinHandler(d.coinRepo, coinSvc, d.logger).WithSettingsSupport(d.settingsSvc)
		coinReferenceHandler := handlers.NewCoinReferenceHandler(coinReferenceRepo, coinReferenceSvc, referenceMigrationSvc)
		coinIntakeSvc := services.NewCoinIntakeService(intakeDraftRepo, d.coinRepo, d.agentProxy, d.settingsSvc).WithCoinService(coinSvc)
		coinIntakeHandler := handlers.NewCoinIntakeHandler(coinIntakeSvc, d.logger)
		quickCaptureSvc := services.NewQuickCaptureService(quickCaptureRepo, d.cfg.UploadDir).
			WithCoinValidation(coinSvc).
			WithReferenceValidation(coinReferenceSvc).
			WithDeepProposalReferences(d.deepIdentificationRepo)
		quickCaptureHandler := handlers.NewQuickCaptureHandler(quickCaptureSvc, d.logger)
		coinLookupHandler := handlers.NewCoinLookupHandler(d.coinLookupSvc, d.logger)
		protected.GET("/coins", coinHandler.List)
		protected.GET("/coins/:id", coinHandler.Get)
		protected.POST("/coins", coinHandler.Create)
		protected.POST("/coins/:id/duplicate", d.writeRateLimit, coinHandler.Duplicate)
		protected.POST("/coins/match-category-era", coinHandler.MatchCategoryEra)
		protected.GET("/coin-options", coinHandler.GetCoinOptions)
		protected.POST("/coins/intake/draft", d.writeRateLimit, coinIntakeHandler.CreateDraft)
		protected.POST("/coins/intake/commit", d.writeRateLimit, coinIntakeHandler.CommitDraft)
		protected.GET("/quick-capture/drafts", quickCaptureHandler.ListDrafts)
		protected.POST("/quick-capture/drafts", d.writeRateLimit, quickCaptureHandler.CreateDraft)
		protected.GET("/quick-capture/drafts/:id", quickCaptureHandler.GetDraft)
		protected.PUT("/quick-capture/drafts/:id", d.writeRateLimit, quickCaptureHandler.UpdateDraft)
		protected.POST("/quick-capture/drafts/:id/discard", d.writeRateLimit, quickCaptureHandler.DiscardDraft)
		protected.POST("/quick-capture/drafts/:id/promote", d.writeRateLimit, quickCaptureHandler.PromoteDraft)
		protected.POST("/coins/lookup", d.writeRateLimit, coinLookupHandler.Lookup)
		protected.PUT("/coins/:id", coinHandler.Update)
		protected.POST("/coins/:id/references", coinReferenceHandler.Create)
		protected.PUT("/coins/:id/references/:referenceId", coinReferenceHandler.Update)
		protected.DELETE("/coins/:id/references/:referenceId", coinReferenceHandler.Delete)
		protected.POST("/references/migrate-legacy", coinReferenceHandler.MigrateLegacy)
		protected.POST("/coins/:id/purchase", coinHandler.Purchase)
		protected.DELETE("/coins/:id", coinHandler.Delete)
		protected.GET("/catalogs", catalogRegistryHandler.List)

		protected.POST("/wishlist/url-intake/analyze", d.writeRateLimit, d.wishlistURLHandler.Analyze)

		mintLocationSvc := services.NewMintLocationService(mintLocationRepo)
		geocodeSvc := services.NewGeocodeService()
		mintLocationHandler := handlers.NewMintLocationHandler(mintLocationSvc).WithGeocoding(geocodeSvc)
		protected.GET("/mint-locations", mintLocationHandler.List)
		protected.GET("/mint-locations/geocode", mintLocationHandler.Geocode)
		protected.POST("/mint-locations", mintLocationHandler.CreatePrivate)
		protected.PUT("/mint-locations/:id", mintLocationHandler.UpdatePrivate)
		protected.DELETE("/mint-locations/:id", mintLocationHandler.DeletePrivate)

		setRepo := repository.NewSetRepository(database.DB)
		recommendationRepo := repository.NewCoinRecommendationRepository(database.DB)
		recommendationService := services.NewCoinRecommendationService(recommendationRepo, setRepo)
		recommendationHandler := handlers.NewCoinRecommendationHandler(recommendationService)
		bulkHandler := handlers.NewBulkHandler(d.coinRepo, setRepo).WithCoinService(coinSvc)
		protected.POST("/coins/bulk", bulkHandler.BulkAction)

		protected.GET("/coins/:id/recommendations", recommendationHandler.List)
		protected.POST("/coins/:id/recommendations/:recommendationId/accept", recommendationHandler.Accept)
		protected.POST("/coins/:id/recommendations/:recommendationId/reject", recommendationHandler.Reject)

		// Sets - new endpoints for coin sets
		setService := services.NewSetService(setRepo)
		setHandler := handlers.NewSetHandler(setRepo, setService).WithSettingsSupport(d.settingsSvc)
		setBuilderRepo := repository.NewSetBuilderRepository(database.DB)
		setBuilderService := services.NewSetBuilderService(setBuilderRepo).
			WithWorkflow(d.agentProxy, d.settingsSvc, repository.NewAgentRepository(database.DB), d.logger).
			WithSetRepository(setRepo)
		setBuilderService.StartWorkers(1)
		d.setBuilderSvc = setBuilderService
		setBuilderHandler := handlers.NewSetBuilderHandler(setBuilderService)
		setSnapshotScheduler := services.NewSetSnapshotScheduler(setService, d.settingsSvc, d.logger)
		go setSnapshotScheduler.Start()
		protected.GET("/sets", setHandler.List)
		protected.POST("/sets/import-csv", setHandler.CreateFromCSV)
		protected.POST("/sets/preview-smart", setHandler.PreviewSmartSet)
		protected.POST("/sets", setHandler.Create)
		protected.GET("/sets/:id", setHandler.Get)
		protected.PUT("/sets/:id", setHandler.Update)
		protected.DELETE("/sets/:id", setHandler.Delete)
		protected.GET("/sets/:id/coins", setHandler.GetCoins)
		protected.POST("/sets/:id/coins", setHandler.AddCoin)
		protected.PUT("/sets/:id/coins/order", setHandler.ReorderCoins)
		protected.DELETE("/sets/:id/coins/:coinId", setHandler.RemoveCoin)
		protected.GET("/sets/:id/completion", setHandler.GetCompletion)
		protected.GET("/sets/:id/checklist", setHandler.GetChecklist)
		protected.POST("/sets/:id/snapshot", setHandler.CreateSnapshot)
		protected.GET("/sets/:id/trends", setHandler.GetTrends)
		protected.GET("/sets/:id/analytics", setHandler.GetAnalytics)
		protected.POST("/set-builder/runs", setBuilderHandler.CreateRun)
		protected.GET("/set-builder/proposals/:id", setBuilderHandler.GetProposal)
		protected.PUT("/set-builder/proposals/:id", setBuilderHandler.UpdateProposal)
		protected.POST("/set-builder/proposals/:id/approve", setBuilderHandler.ApproveProposal)
		protected.POST("/set-builder/proposals/:id/reject", setBuilderHandler.RejectProposal)
		protected.POST("/set-builder/proposals/:id/regenerate", setBuilderHandler.RegenerateProposal)

		journalHandler := handlers.NewJournalHandler(d.journalRepo)
		protected.GET("/coins/:id/journal", journalHandler.ListEntries)
		protected.POST("/coins/:id/journal", journalHandler.AddEntry)
		protected.DELETE("/coins/:id/journal/:entryId", journalHandler.DeleteEntry)

		protected.GET("/stats", coinHandler.Stats)
		healthHandler := handlers.NewHealthHandler(d.healthSvc, d.logger)
		protected.GET("/stats/health", healthHandler.CollectionSummary)
		protected.GET("/coins/health", healthHandler.ListCoinHealth)
		protected.GET("/coins/:id/health", healthHandler.GetCoinHealth)
		timeMachineHandler := handlers.NewTimeMachineHandler(d.timeMachineSvc)
		protected.GET("/stats/time-machine", timeMachineHandler.GetSnapshot)
		protected.GET("/stats/time-machine/bounds", timeMachineHandler.GetBounds)
		protected.GET("/stats/distribution", coinHandler.Distribution)
		protected.GET("/coins/:id/value-history", coinHandler.CoinValueHistory)
		protected.GET("/suggestions", coinHandler.Suggestions)

		protected.POST("/coins/:id/images", d.writeRateLimit, d.imageHandler.Upload)
		protected.POST("/coins/:id/images/import", d.writeRateLimit, d.imageHandler.ImportRemote)
		protected.DELETE("/coins/:id/images/:imageId", d.imageHandler.Delete)
		protected.POST("/coins/:id/images/swap-sides", d.writeRateLimit, d.imageHandler.SwapSides)
		protected.GET("/uploads/*filepath", d.imageHandler.ServeUpload)
		protected.GET("/proxy-image", d.imageHandler.ProxyImage)
		protected.GET("/scrape-image", d.imageHandler.ScrapeImage)

		analysisRepo := repository.NewAnalysisRepository(database.DB)
		analysisHandler := handlers.NewAnalysisHandler(analysisRepo, d.agentProxy, d.settingsSvc, d.logger)
		aiJobHandler := handlers.NewAIJobHandler(d.aiJobSvc, d.logger)
		protected.POST("/coins/:id/analyze", d.writeRateLimit, aiJobHandler.Analyze)
		protected.POST("/coins/:id/grade", d.writeRateLimit, aiJobHandler.Grade)
		protected.DELETE("/coins/:id/analyze", analysisHandler.DeleteAnalysis)
		protected.GET("/ai-jobs/:id", aiJobHandler.GetJob)
		protected.GET("/coins/:id/ai-jobs", aiJobHandler.ListCoinJobs)
		protected.POST("/extract-text", analysisHandler.ExtractText)
		protected.GET("/ollama-status", analysisHandler.OllamaStatus)
		protected.GET("/ai-status", analysisHandler.AIStatus)

		deepIdentificationProposalSvc := services.NewDeepIdentificationProposalService(d.deepIdentificationRepo, d.coinRepo, coinSvc, quickCaptureSvc, coinReferenceSvc).WithLogger(d.logger)
		deepIdentificationHandler := handlers.NewDeepIdentificationHandler(d.deepIdentificationSvc, d.settingsSvc, d.logger).WithProposalSupport(deepIdentificationProposalSvc)
		protected.GET("/deep-identification/capability", deepIdentificationHandler.Capability)
		protected.POST("/deep-identification/jobs", d.writeRateLimit, deepIdentificationHandler.CreateJob)
		protected.GET("/deep-identification/jobs", deepIdentificationHandler.ListJobs)
		protected.GET("/deep-identification/jobs/:id", deepIdentificationHandler.GetJob)
		protected.DELETE("/deep-identification/jobs/:id", d.writeRateLimit, deepIdentificationHandler.DeleteJob)
		protected.GET("/deep-identification/jobs/:id/events", deepIdentificationHandler.StreamEvents)
		protected.POST("/deep-identification/jobs/:id/cancel", d.writeRateLimit, deepIdentificationHandler.Cancel)
		protected.POST("/deep-identification/jobs/:id/retry", d.writeRateLimit, deepIdentificationHandler.Retry)
		protected.PATCH("/deep-identification/jobs/:id/proposal", d.writeRateLimit, deepIdentificationHandler.UpdateProposal)
		protected.POST("/deep-identification/jobs/:id/apply", d.writeRateLimit, deepIdentificationHandler.ApplyProposal)

		numistaHandler := handlers.NewNumistaHandler(d.numistaLookupSvc)
		protected.POST("/numista/query-proposal", numistaHandler.QueryProposal)
		protected.POST("/numista/lookup", numistaHandler.Lookup)
		protected.POST("/numista/enrich", numistaHandler.Enrich)
		protected.GET("/numista/search", numistaHandler.Search)

		// Numista catalogue browsing and market values share one database cache.
		numistaResponseCache := services.NewNumistaResponseCache(repository.NewNumistaCacheRepository(database.DB), services.NewSystemNumistaClock(), d.logger)
		ecbCoinPhotos := services.NewEcbCoinPhotoService(services.NewHTTPECBCoinsClient(nil), numistaResponseCache)
		incmCoinPhotos := services.NewIncmCoinPhotoService(services.NewHTTPINCMShopClient(nil), numistaResponseCache)
		mdpCoinPhotos := services.NewMdpCoinPhotoService(services.NewHTTPMDPShopClient(nil), numistaResponseCache)
		moeCoinPhotos := services.NewMoeCoinPhotoService(services.NewHTTPMOEShopClient(nil), numistaResponseCache)
		numistaCatalogSvc := services.NewNumistaCatalogService(d.numistaClient, numistaResponseCache).WithECB(ecbCoinPhotos).WithMint(incmCoinPhotos).WithMint(mdpCoinPhotos).WithMint(moeCoinPhotos)
		numistaMarketValueSvc := services.NewNumistaMarketValueService(d.numistaClient, numistaResponseCache).WithCatalog(numistaCatalogSvc)
		coinMarketValueHandler := handlers.NewCoinMarketValueHandler(d.coinRepo, numistaMarketValueSvc)
		protected.GET("/coins/:id/market-value", coinMarketValueHandler.MarketValue)
		numistaCatalogHandler := handlers.NewNumistaCatalogHandler(numistaCatalogSvc, numistaMarketValueSvc)
		protected.GET("/numista/catalog/issuers", numistaCatalogHandler.Issuers)
		protected.GET("/numista/catalog/search", numistaCatalogHandler.Search)
		protected.GET("/numista/catalog/types/:id", numistaCatalogHandler.Type)
		protected.GET("/numista/catalog/types/:id/market-value", numistaCatalogHandler.MarketValue)
		coinPhotoSearch := services.NewCoinPhotoSearchService(numistaCatalogSvc, services.NewHTTPCommonsClient(nil), numistaResponseCache).WithECB(ecbCoinPhotos).WithMint(incmCoinPhotos).WithMint(mdpCoinPhotos).WithMint(moeCoinPhotos)
		protected.GET("/coin-photos/search", handlers.NewCoinPhotoHandler(coinPhotoSearch).Search)
		// Coin subjects in the user's language, translated to compare with Numista's titles.
		protected.GET("/translate", handlers.NewTranslationHandler(services.NewTranslationService(services.NewMyMemoryClient(nil), numistaResponseCache)).Translate)
		// The add-coin page fills the sides the user has no photo of with the
		// coin's official images.
		// The uCoin import asks for every imported coin, so this has its own,
		// higher limit; each request downloads at most two official images.
		officialPhotosRateLimit := middleware.AuthenticatedRateLimit(300, time.Minute)
		officialPhotos := handlers.NewOfficialPhotoHandler(d.coinRepo, d.imageHandler, coinPhotoSearch, d.logger)
		protected.POST("/coins/:id/official-photos", officialPhotosRateLimit, officialPhotos.Add)
		protected.GET("/coins/official-photos/missing", officialPhotos.Missing)
		protected.POST("/coins/official-photos/batch", officialPhotosRateLimit, officialPhotos.Batch)
		coinCountryHandler := handlers.NewCoinCountryHandler(d.coinRepo, services.NewCoinCountryService(d.coinRepo, numistaCatalogSvc))
		protected.GET("/coins/countries", coinCountryHandler.Countries)
		protected.GET("/coins/denominations", coinCountryHandler.Denominations)
		// Whether coins are still legal tender, from their cached Numista types.
		coinCirculation := services.NewCoinCirculationService(d.coinRepo, numistaCatalogSvc)
		coinHandler.WithCirculation(coinCirculation)
		protected.GET("/coins/:id/status", coinHandler.Status)
		protected.GET("/coins/statuses", coinHandler.StatusGroups)
		protected.GET("/coins/summary", handlers.NewCollectionSummaryHandler(d.coinRepo).WithExchangeRates(d.exchangeRates).WithSettings(d.settingsSvc).WithCirculation(coinCirculation).Summary)
		// New and upcoming coins: the ECB's €2 commemoratives and the mints' collector coins.
		coinReleases := services.NewCoinReleasesService(ecbCoinPhotos, incmCoinPhotos, mdpCoinPhotos, moeCoinPhotos).WithNumista(numistaCatalogSvc)
		coinReleasesHandler := handlers.NewCoinReleasesHandler(coinReleases, d.coinRepo, coinPhotoSearch)
		protected.GET("/coin-releases", coinReleasesHandler.Releases)
		// The Numista type of a release, to fill in the coin added from it.
		protected.GET("/coin-releases/numista-type", coinReleasesHandler.NumistaType)
		// The wishlist coins a coin just added to the collection is.
		protected.GET("/coins/:id/wishlist-matches", coinReleasesHandler.WishlistMatches)
		// Checklist sets list the catalogue's coins or a challenge's; smart
		// sets choose coins by their kind too.
		catalogueOverrides := repository.NewCatalogueOverrideRepository(database.DB)
		setChecklist := services.NewSetChecklistService(coinReleases, d.coinRepo, coinPhotoSearch).WithOverrides(catalogueOverrides)
		setService.WithChecklist(setChecklist)
		setRepo.WithCoinClassifier(func(userID uint, kind string) ([]uint, error) {
			sorted, err := coinCirculation.Classify(userID)
			if err != nil {
				return nil, err
			}
			switch kind {
			case "commemorative":
				return sorted.Commemorative, nil
			case "collector":
				return sorted.NonCirculating, nil
			case "circulation":
				collector := map[uint]bool{}
				for _, id := range sorted.NonCirculating {
					collector[id] = true
				}
				ids := []uint{}
				for _, id := range sorted.Regular {
					if !collector[id] {
						ids = append(ids, id)
					}
				}
				return ids, nil
			}
			return nil, nil
		})
		// Every €2 commemorative and the mints' collector coins, by country.
		coinCatalogue := handlers.NewCoinCatalogueHandler(coinReleases, d.coinRepo, coinPhotoSearch).WithOverrides(catalogueOverrides)
		protected.GET("/coin-catalogue/countries", coinCatalogue.Countries)
		protected.GET("/coin-catalogue/joint-issues", coinCatalogue.JointIssues)
		protected.GET("/coin-catalogue", coinCatalogue.Country)
		// Each user's coins hidden from the catalogue, or joined to another.
		protected.PUT("/coin-catalogue/overrides", d.writeRateLimit, coinCatalogue.SetOverride)
		protected.DELETE("/coin-catalogue/overrides", d.writeRateLimit, coinCatalogue.DeleteOverride)
		// ECB rates, to convert amounts into the collection currency.
		protected.GET("/exchange-rates", handlers.NewExchangeRatesHandler(d.exchangeRates).Rates)
		protected.POST("/coins/country-from-numista", d.writeRateLimit, coinCountryHandler.FillFromNumista)

		// Wishlist availability checking
		availHandler := handlers.NewAvailabilityHandler(d.availSvc, d.availScheduler, d.availRepo, d.coinRepo).WithCycleRepo(d.availCycleRepo)
		protected.POST("/wishlist/check-availability", availHandler.CheckAvailability)
		protected.PUT("/coins/:id/listing-status", availHandler.UpdateListingStatus)
		protected.GET("/wishlist/availability-runs", availHandler.ListOwnerRuns)
		protected.GET("/wishlist/availability-runs/:id", availHandler.GetOwnerRunDetail)
		wishlistSearchAlertHandler := handlers.NewWishlistSearchAlertHandler(d.wishlistSearchAlertSvc)
		protected.GET("/wishlist/search-alerts", wishlistSearchAlertHandler.List)
		protected.POST("/wishlist/search-alerts", d.writeRateLimit, wishlistSearchAlertHandler.Create)
		protected.PUT("/wishlist/search-alerts/:alertId", d.writeRateLimit, wishlistSearchAlertHandler.Update)
		protected.DELETE("/wishlist/search-alerts/:alertId", d.writeRateLimit, wishlistSearchAlertHandler.Delete)
		protected.POST("/wishlist/search-alerts/:alertId/run", d.writeRateLimit, wishlistSearchAlertHandler.RunNow)
		protected.GET("/wishlist/search-alerts/:alertId/runs", wishlistSearchAlertHandler.ListRuns)
		protected.GET("/wishlist/search-alerts/:alertId/runs/:runId", wishlistSearchAlertHandler.GetRun)
		protected.GET("/wishlist/search-alerts/:alertId/candidates", wishlistSearchAlertHandler.ListCandidates)
		protected.POST("/wishlist/search-alerts/:alertId/candidates/:candidateId/dismiss", d.writeRateLimit, wishlistSearchAlertHandler.DismissCandidate)
		protected.POST("/wishlist/search-alerts/:alertId/candidates/:candidateId/restore", d.writeRateLimit, wishlistSearchAlertHandler.RestoreCandidate)
		protected.POST("/wishlist/search-alerts/:alertId/candidates/:candidateId/convert", d.writeRateLimit, wishlistSearchAlertHandler.ConvertCandidate)
		protected.POST("/wishlist/search-alerts/:alertId/criteria-adjustments", d.writeRateLimit, wishlistSearchAlertHandler.AdjustCriteria)

		agentRepo := repository.NewAgentRepository(database.DB)
		userRepo := repository.NewUserRepository(database.DB)
		contentGuard := services.NewContentGuard(d.logger)
		agentHandler := handlers.NewAgentHandler(agentRepo, userRepo, d.journalRepo, d.agentProxy, d.collectionSvc, d.settingsSvc, d.internalTokenSvc, contentGuard, d.logger, d.cfg.AgentInternalCallbackURL)
		protected.POST("/agent/chat", d.writeRateLimit, agentHandler.ChatStream)
		coinCopilotHandler := handlers.NewCoinCopilotHandler(d.coinCopilotSvc, d.logger)
		protected.GET("/agent/copilot/capability", coinCopilotHandler.Capability)
		protected.POST("/agent/copilot/runs", d.writeRateLimit, coinCopilotHandler.Start)
		protected.GET("/agent/copilot/threads/:threadId", coinCopilotHandler.GetThread)
		protected.GET("/agent/copilot/runs/:runId", coinCopilotHandler.GetRun)
		protected.GET("/agent/copilot/runs/:runId/events", coinCopilotHandler.StreamEvents)
		protected.POST("/agent/copilot/runs/:runId/cancel", d.writeRateLimit, coinCopilotHandler.Cancel)
		protected.POST("/agent/copilot/runs/:runId/resume", d.writeRateLimit, coinCopilotHandler.Resume)
		protected.POST("/agent/collection/proposals/:proposalId/commit", d.writeRateLimit, agentHandler.CommitCollectionProposal)
		protected.POST("/agent/collection/proposals/:proposalId/cancel", d.writeRateLimit, agentHandler.CancelCollectionProposal)
		protected.POST("/coins/:id/estimate-value", d.writeRateLimit, aiJobHandler.EstimateValue)
		protected.GET("/agent/models", agentHandler.ListModels)
		protected.GET("/agent/coin-search-prompt", agentHandler.GetCoinSearchPrompt)
		protected.GET("/agent/coin-shows-prompt", agentHandler.GetCoinShowsPrompt)
		protected.GET("/agent/valuation-prompt", agentHandler.GetValuationPrompt)
		protected.GET("/agent/portfolio-summary", agentHandler.PortfolioSummary)
		protected.GET("/agent/status", agentHandler.AgentStatus)

		conversationRepo := repository.NewConversationRepository(database.DB)
		convHandler := handlers.NewConversationHandler(conversationRepo)
		protected.GET("/agent/conversations", convHandler.List)
		protected.GET("/agent/conversations/:id", convHandler.Get)
		protected.POST("/agent/conversations", convHandler.Save)
		protected.DELETE("/agent/conversations/:id", convHandler.Delete)

		// User self-service routes
		userHandler := handlers.NewUserHandler(d.cfg.UploadDir, userRepo, d.logger)
		protected.GET("/auth/me", userHandler.GetMe)
		protected.POST("/auth/change-password", userHandler.ChangePassword)
		protected.PUT("/user/profile", userHandler.UpdateProfile)
		protected.POST("/user/avatar", userHandler.UploadAvatar)
		protected.DELETE("/user/avatar", userHandler.DeleteAvatar)
		protected.GET("/user/export", userHandler.ExportCollection)
		protected.GET("/user/export/catalog", userHandler.ExportCatalogPDF)
		protected.POST("/user/import", d.writeRateLimit, userHandler.ImportCollection)
	}
}
