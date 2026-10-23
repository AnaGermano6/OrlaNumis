package services

import (
	"fmt"
	"time"

	"github.com/briandenicola/ancient-coins-api/models"
	"github.com/briandenicola/ancient-coins-api/repository"
)

// HealthService orchestrates collection and coin health scoring workflows.
type HealthService struct {
	repo     *repository.HealthRepository
	logger   *Logger
	settings *SettingsService
}

// NewHealthService creates a new HealthService.
func NewHealthService(repo *repository.HealthRepository, logger *Logger) *HealthService {
	return &HealthService{repo: repo, logger: logger}
}

// WithSettings lets the scores leave out the AI analyses while no AI
// provider is configured.
func (s *HealthService) WithSettings(settings *SettingsService) *HealthService {
	s.settings = settings
	return s
}

// aiConfigured tells whether AI analyses can be run; without settings they
// are counted, as before.
func (s *HealthService) aiConfigured() bool {
	if s.settings == nil {
		return true
	}
	_, err := s.settings.ResolveLLMConfig()
	return err == nil
}

// weights are the score's weights: without an AI provider no item can have
// an AI analysis, so that part is left out and the others share its weight.
func (s *HealthService) weights() HealthWeights {
	if s.aiConfigured() {
		return DefaultHealthWeights()
	}
	return HealthWeights{
		Metadata:           HealthWeightMetadataWithoutAI,
		ImageCoverage:      HealthWeightImageCoverageWithoutAI,
		ValuationFreshness: HealthWeightValuationFreshnessWithoutAI,
	}
}

// GetCollectionHealthSummary returns collection-level scorecard data.
func (s *HealthService) GetCollectionHealthSummary(userID uint) (*CollectionHealthSummary, error) {
	coins, err := s.repo.ListEligibleCoins(userID)
	if err != nil {
		return nil, err
	}

	if len(coins) == 0 {
		return &CollectionHealthSummary{
			Score:             0,
			Grade:             HealthGradeF,
			EligibleCoinCount: 0,
			Weights:           s.weights(),
			Dimensions:        HealthDimensionScores{},
		}, nil
	}

	totalMetadata := 0
	totalImage := 0
	totalValuation := 0
	totalAI := 0

	for _, coin := range coins {
		totalMetadata += s.scoreCoinMetadata(&coin)
		totalImage += s.scoreCoinImages(&coin)
		totalValuation += s.scoreCoinValuationFreshness(&coin)
		totalAI += s.scoreCoinAICoverage(&coin)
	}

	count := len(coins)
	avgMetadata := totalMetadata / count
	avgImage := totalImage / count
	avgValuation := totalValuation / count
	avgAI := totalAI / count

	collectionScore := s.computeWeightedScore(avgMetadata, avgImage, avgValuation, avgAI)

	return &CollectionHealthSummary{
		Score:             collectionScore,
		Grade:             HealthGradeFromScore(collectionScore),
		EligibleCoinCount: int64(count),
		Weights:           s.weights(),
		Dimensions: HealthDimensionScores{
			Metadata:           avgMetadata,
			ImageCoverage:      avgImage,
			ValuationFreshness: avgValuation,
			AICoverage:         avgAI,
		},
	}, nil
}

// ListCoinHealth returns a paginated coin-health list.
func (s *HealthService) ListCoinHealth(userID uint, page, limit int, scope string) (*CoinHealthListResponse, error) {
	rows, total, err := s.repo.ListEligibleCoinsPaged(userID, page, limit, scope)
	if err != nil {
		return nil, err
	}

	items := make([]CoinHealthItem, 0, len(rows))
	for _, row := range rows {
		metadataScore := s.scoreCoinMetadata(&row)
		imageScore := s.scoreCoinImages(&row)
		valuationScore := s.scoreCoinValuationFreshness(&row)
		aiScore := s.scoreCoinAICoverage(&row)
		totalScore := s.computeWeightedScore(metadataScore, imageScore, valuationScore, aiScore)

		checklist := s.generateCoinChecklist(&row)
		quickActions := s.extractQuickActions(checklist)

		items = append(items, CoinHealthItem{
			CoinID: row.CoinID,
			Title:  row.Title,
			Score:  totalScore,
			Grade:  HealthGradeFromScore(totalScore),
			Dimensions: HealthDimensionScores{
				Metadata:           metadataScore,
				ImageCoverage:      imageScore,
				ValuationFreshness: valuationScore,
				AICoverage:         aiScore,
			},
			MissingItems: checklist,
			QuickActions: quickActions,
		})
	}

	return &CoinHealthListResponse{
		Coins: items,
		Pagination: HealthPagination{
			Page:  page,
			Limit: limit,
			Total: total,
		},
	}, nil
}

// GetCoinHealth returns health data for a single coin.
func (s *HealthService) GetCoinHealth(coinID, userID uint) (*CoinHealthItem, error) {
	row, err := s.repo.GetSingleEligibleCoin(coinID, userID)
	if err != nil {
		return nil, err
	}

	metadataScore := s.scoreCoinMetadata(row)
	imageScore := s.scoreCoinImages(row)
	valuationScore := s.scoreCoinValuationFreshness(row)
	aiScore := s.scoreCoinAICoverage(row)
	totalScore := s.computeWeightedScore(metadataScore, imageScore, valuationScore, aiScore)

	checklist := s.generateCoinChecklist(row)
	quickActions := s.extractQuickActions(checklist)

	return &CoinHealthItem{
		CoinID: row.CoinID,
		Title:  row.Title,
		Score:  totalScore,
		Grade:  HealthGradeFromScore(totalScore),
		Dimensions: HealthDimensionScores{
			Metadata:           metadataScore,
			ImageCoverage:      imageScore,
			ValuationFreshness: valuationScore,
			AICoverage:         aiScore,
		},
		MissingItems: checklist,
		QuickActions: quickActions,
	}, nil
}

// GetAdminHealthSummary returns aggregate health metrics for admin views.
func (s *HealthService) GetAdminHealthSummary() (*AdminHealthSummary, error) {
	allCoins, err := s.repo.ListAllEligibleCoins()
	if err != nil {
		return nil, err
	}

	if len(allCoins) == 0 {
		return &AdminHealthSummary{
			MedianScore:        0,
			LowScorePercentage: 0,
			LowScoreThreshold:  HealthLowScoreThreshold,
			EligibleCoinCount:  0,
			TopMissingFields:   []MissingFieldStat{},
		}, nil
	}

	scores := make([]int, 0, len(allCoins))
	lowScoreCount := 0
	missingFieldCounts := make(map[string]int)

	for _, coin := range allCoins {
		metadataScore := s.scoreCoinMetadata(&coin)
		imageScore := s.scoreCoinImages(&coin)
		valuationScore := s.scoreCoinValuationFreshness(&coin)
		aiScore := s.scoreCoinAICoverage(&coin)
		totalScore := s.computeWeightedScore(metadataScore, imageScore, valuationScore, aiScore)

		scores = append(scores, totalScore)

		if totalScore < HealthLowScoreThreshold {
			lowScoreCount++
			checklist := s.generateCoinChecklist(&coin)
			for _, item := range checklist {
				missingFieldCounts[item.Key]++
			}
		}
	}

	medianScore := s.computeMedian(scores)
	lowScorePercentage := (float64(lowScoreCount) / float64(len(allCoins))) * 100

	topMissing := s.computeTopMissingFields(missingFieldCounts, len(allCoins))

	return &AdminHealthSummary{
		MedianScore:        medianScore,
		LowScorePercentage: lowScorePercentage,
		LowScoreThreshold:  HealthLowScoreThreshold,
		EligibleCoinCount:  int64(len(allCoins)),
		TopMissingFields:   topMissing,
	}, nil
}

// computeMedian calculates the median score from a list of scores.
func (s *HealthService) computeMedian(scores []int) int {
	if len(scores) == 0 {
		return 0
	}

	sorted := make([]int, len(scores))
	copy(sorted, scores)

	for i := 0; i < len(sorted); i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[i] > sorted[j] {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	mid := len(sorted) / 2
	if len(sorted)%2 == 0 {
		return (sorted[mid-1] + sorted[mid]) / 2
	}
	return sorted[mid]
}

// computeTopMissingFields returns the top 5 missing fields by count.
func (s *HealthService) computeTopMissingFields(counts map[string]int, totalCoins int) []MissingFieldStat {
	type kv struct {
		key   string
		count int
	}

	sorted := make([]kv, 0, len(counts))
	for k, v := range counts {
		sorted = append(sorted, kv{k, v})
	}

	for i := 0; i < len(sorted); i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[i].count < sorted[j].count {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	limit := 5
	if len(sorted) < limit {
		limit = len(sorted)
	}

	topFields := make([]MissingFieldStat, 0, limit)
	for i := 0; i < limit; i++ {
		percentage := (float64(sorted[i].count) / float64(totalCoins)) * 100
		topFields = append(topFields, MissingFieldStat{
			Key:        sorted[i].key,
			Count:      int64(sorted[i].count),
			Percentage: percentage,
		})
	}

	return topFields
}

// scoreCoinMetadata computes metadata completeness (0-100) from the fields
// a modern coin or banknote has: face value, country, monetary system, mint
// (coins only), category, material and grade.
func (s *HealthService) scoreCoinMetadata(coin *repository.EligibleCoinRow) int {
	fields := []string{
		coin.Denomination,
		coin.Country,
		coin.MonetarySystem,
		coin.Category,
		coin.Material,
		coin.Grade,
	}
	if coin.Kind != models.CoinKindBanknote {
		fields = append(fields, coin.Mint)
	}

	present := 0
	for _, field := range fields {
		if field != "" {
			present++
		}
	}

	if len(fields) == 0 {
		return 0
	}
	return (present * 100) / len(fields)
}

// scoreCoinImages computes image coverage (0-100).
func (s *HealthService) scoreCoinImages(coin *repository.EligibleCoinRow) int {
	if coin.ImageCount >= 2 {
		return 100
	}
	if coin.ImageCount == 1 {
		return 50
	}
	return 0
}

// healthUndatedValueScore scores a value of unknown age (neither when it was
// set nor a purchase date is known) as a value over six months old.
const healthUndatedValueScore = 35

// scoreCoinValuationFreshness computes valuation freshness (0-100).
// Measures age from CurrentValueUpdatedAt when available; falls back to PurchaseDate for legacy coins.
func (s *HealthService) scoreCoinValuationFreshness(coin *repository.EligibleCoinRow) int {
	if coin.CurrentValue == nil {
		return 0
	}

	now := time.Now()

	// Use CurrentValueUpdatedAt if available (new behavior)
	var referenceTime *time.Time
	if coin.CurrentValueUpdatedAt != nil {
		referenceTime = coin.CurrentValueUpdatedAt
	} else if coin.PurchaseDate != nil {
		// Fallback: for legacy coins valued before CurrentValueUpdatedAt existed, use PurchaseDate
		referenceTime = coin.PurchaseDate
	} else {
		return healthUndatedValueScore
	}

	age := now.Sub(*referenceTime)
	days := int(age.Hours() / 24)

	if days <= 30 {
		return 100
	} else if days <= 90 {
		return 80
	} else if days <= 180 {
		return 60
	} else if days <= 365 {
		return 35
	}
	return 0
}

// scoreCoinAICoverage computes AI analysis coverage (0-100).
// Coverage is measured solely by per-side analysis: obverse and reverse.
// Both sides analyzed = 100, one side = 50, neither = 0. The legacy combined
// AIAnalysis field is intentionally not counted.
func (s *HealthService) scoreCoinAICoverage(coin *repository.EligibleCoinRow) int {
	hasObverse := coin.ObverseAnalysis != ""
	hasReverse := coin.ReverseAnalysis != ""

	if hasObverse && hasReverse {
		return 100
	}
	if hasObverse || hasReverse {
		return 50
	}
	return 0
}

// computeWeightedScore combines dimension scores using the weights.
func (s *HealthService) computeWeightedScore(metadata, image, valuation, ai int) int {
	weights := s.weights()
	weighted := (metadata * weights.Metadata) +
		(image * weights.ImageCoverage) +
		(valuation * weights.ValuationFreshness) +
		(ai * weights.AICoverage)
	score := weighted / 100
	return ClampHealthScore(score)
}

// generateCoinChecklist identifies missing items for a coin.
func (s *HealthService) generateCoinChecklist(coin *repository.EligibleCoinRow) []MissingChecklistItem {
	items := []MissingChecklistItem{}

	if coin.Denomination == "" {
		items = append(items, MissingChecklistItem{
			Key:        "metadata.denomination",
			Dimension:  ChecklistDimensionMetadata,
			Label:      "Denomination",
			Severity:   ChecklistSeverityHigh,
			ActionHint: HealthActionEditMetadata,
		})
	}
	if coin.Country == "" {
		items = append(items, MissingChecklistItem{
			Key:        "metadata.country",
			Dimension:  ChecklistDimensionMetadata,
			Label:      "Country",
			Severity:   ChecklistSeverityHigh,
			ActionHint: HealthActionEditMetadata,
		})
	}
	if coin.MonetarySystem == "" {
		items = append(items, MissingChecklistItem{
			Key:        "metadata.monetarySystem",
			Dimension:  ChecklistDimensionMetadata,
			Label:      "Monetary system",
			Severity:   ChecklistSeverityMedium,
			ActionHint: HealthActionEditMetadata,
		})
	}
	// A banknote has a printer, not a mint.
	if coin.Mint == "" && coin.Kind != models.CoinKindBanknote {
		items = append(items, MissingChecklistItem{
			Key:        "metadata.mint",
			Dimension:  ChecklistDimensionMetadata,
			Label:      "Mint",
			Severity:   ChecklistSeverityMedium,
			ActionHint: HealthActionEditMetadata,
		})
	}
	if coin.Category == "" {
		items = append(items, MissingChecklistItem{
			Key:        "metadata.category",
			Dimension:  ChecklistDimensionMetadata,
			Label:      "Category",
			Severity:   ChecklistSeverityMedium,
			ActionHint: HealthActionEditMetadata,
		})
	}
	if coin.Material == "" {
		items = append(items, MissingChecklistItem{
			Key:        "metadata.material",
			Dimension:  ChecklistDimensionMetadata,
			Label:      "Material",
			Severity:   ChecklistSeverityLow,
			ActionHint: HealthActionEditMetadata,
		})
	}
	if coin.Grade == "" {
		items = append(items, MissingChecklistItem{
			Key:        "metadata.grade",
			Dimension:  ChecklistDimensionMetadata,
			Label:      "Grade",
			Severity:   ChecklistSeverityLow,
			ActionHint: HealthActionEditMetadata,
		})
	}

	if coin.ImageCount == 0 {
		items = append(items, MissingChecklistItem{
			Key:        "images.any",
			Dimension:  ChecklistDimensionImages,
			Label:      "At least one image",
			Severity:   ChecklistSeverityHigh,
			ActionHint: HealthActionUploadImages,
		})
	} else if coin.ImageCount == 1 {
		items = append(items, MissingChecklistItem{
			Key:        "images.coverage",
			Dimension:  ChecklistDimensionImages,
			Label:      "Both obverse and reverse images",
			Severity:   ChecklistSeverityMedium,
			ActionHint: HealthActionUploadImages,
		})
	}

	// A value counts whatever is known of the purchase.
	if coin.CurrentValue == nil {
		items = append(items, MissingChecklistItem{
			Key:        "valuation.value",
			Dimension:  ChecklistDimensionValuation,
			Label:      "Current valuation",
			Severity:   ChecklistSeverityHigh,
			ActionHint: HealthActionRunValuation,
		})
	} else {
		// Check valuation freshness from CurrentValueUpdatedAt (or fallback to PurchaseDate for legacy coins)
		now := time.Now()
		var referenceTime *time.Time
		if coin.CurrentValueUpdatedAt != nil {
			referenceTime = coin.CurrentValueUpdatedAt
		} else {
			referenceTime = coin.PurchaseDate
		}

		// A value of unknown age is not known to be recent either.
		if referenceTime == nil || int(now.Sub(*referenceTime).Hours()/24) > 180 {
			items = append(items, MissingChecklistItem{
				Key:        "valuation.freshness",
				Dimension:  ChecklistDimensionValuation,
				Label:      "Recent valuation (>180 days old)",
				Severity:   ChecklistSeverityMedium,
				ActionHint: HealthActionRunValuation,
			})
		}
	}

	// Without an AI provider no analysis can be run, so none is asked for.
	if !s.aiConfigured() {
		return items
	}
	hasObverse := coin.ObverseAnalysis != ""
	hasReverse := coin.ReverseAnalysis != ""

	if !hasObverse && !hasReverse {
		// No per-side AI analysis at all
		items = append(items, MissingChecklistItem{
			Key:        "ai.analysis",
			Dimension:  ChecklistDimensionAI,
			Label:      "Run AI analysis on the obverse and reverse",
			Severity:   ChecklistSeverityMedium,
			ActionHint: HealthActionRunAIAnalysis,
		})
	} else if !hasObverse || !hasReverse {
		// One side is analyzed; name the side that still needs analysis
		missingSide, doneSide := "reverse", "obverse"
		if !hasObverse {
			missingSide, doneSide = "obverse", "reverse"
		}
		items = append(items, MissingChecklistItem{
			Key:        "ai.coverage",
			Dimension:  ChecklistDimensionAI,
			Label:      fmt.Sprintf("Run AI analysis on the %s (%s already done)", missingSide, doneSide),
			Severity:   ChecklistSeverityLow,
			ActionHint: HealthActionRunAIAnalysis,
		})
	}

	return items
}

// extractQuickActions derives action hints from the checklist.
func (s *HealthService) extractQuickActions(checklist []MissingChecklistItem) []HealthActionHint {
	seen := make(map[HealthActionHint]bool)
	actions := []HealthActionHint{}

	for _, item := range checklist {
		if !seen[item.ActionHint] {
			seen[item.ActionHint] = true
			actions = append(actions, item.ActionHint)
		}
	}

	return actions
}
