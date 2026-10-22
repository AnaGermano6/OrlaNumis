package services

import (
	"context"
	"errors"
	"fmt"
	"hash/fnv"
	"strings"
	"time"

	"github.com/briandenicola/ancient-coins-api/models"
	"github.com/briandenicola/ancient-coins-api/repository"
)

const maxSetsPerUser = 100
const maxSetNameLength = 80

var (
	ErrInvalidSetOrder = errors.New("ordered coin IDs must exactly match current set members")
	ErrSmartSetOrder   = errors.New("cannot manually reorder smart sets")
)

type SetValidationError struct {
	Message string
}

func (e *SetValidationError) Error() string {
	return e.Message
}

func newSetValidationError(message string) error {
	return &SetValidationError{Message: message}
}

// SetService handles business logic for coin sets.
type SetService struct {
	repo      *repository.SetRepository
	checklist *SetChecklistService
}

// WithChecklist lets the service work out checklist sets' slots.
func (s *SetService) WithChecklist(checklist *SetChecklistService) *SetService {
	s.checklist = checklist
	return s
}

// ErrNotChecklistSet is returned for the slots of a set that is not a checklist.
var ErrNotChecklistSet = errors.New("the set is not a checklist")

// ChecklistSlots lists a checklist set's slots; complete is false while the
// catalogue behind them is partly listed.
func (s *SetService) ChecklistSlots(ctx context.Context, setID, userID uint) (slots []ChecklistSlot, complete bool, err error) {
	set, err := s.repo.GetByID(setID, userID)
	if err != nil {
		return nil, false, err
	}
	if set.SetType != models.CoinSetTypeChecklist || set.Checklist == nil || s.checklist == nil {
		return nil, false, ErrNotChecklistSet
	}
	definition, err := ParseChecklistDefinition(map[string]interface{}(*set.Checklist))
	if err != nil {
		return nil, false, err
	}
	return s.checklist.SlotsAndCompleteness(ctx, userID, definition)
}

// checklistSummary counts a checklist set's slots; nil when it cannot.
func (s *SetService) checklistSummary(set *models.CoinSet) *ChecklistSummary {
	if overview := s.checklistOverview(set); overview != nil {
		return &overview.summary
	}
	return nil
}

// checklistOverview is a checklist set's slots counted, in all and country
// by country; complete is false while its list is partly known.
type checklistOverview struct {
	summary   ChecklistSummary
	countries []ChecklistCountry
	complete  bool
}

func (s *SetService) checklistOverview(set *models.CoinSet) *checklistOverview {
	slots, complete, err := s.ChecklistSlots(context.Background(), set.ID, set.UserID)
	if err != nil {
		return nil
	}
	return &checklistOverview{summary: Summarize(slots), countries: SummarizeByCountry(slots), complete: complete}
}

// NewSetService creates a new SetService.
func NewSetService(repo *repository.SetRepository) *SetService {
	return &SetService{repo: repo}
}

// ListSets returns all sets for a user with summary data.
func (s *SetService) ListSets(userID uint) ([]map[string]interface{}, error) {
	sets, err := s.repo.List(userID)
	if err != nil {
		return nil, err
	}

	// Summaries and goal completions are fetched in bulk rather than per set.
	// Only standard and goal sets qualify: smart sets need their criteria
	// evaluated individually, and agentic sets summarize in Go with different
	// averaging semantics (see SetRepository.GetSetSummaries). Those two types
	// fall back to the per-set path below.
	batchableIDs := make([]uint, 0, len(sets))
	goalIDs := make([]uint, 0, len(sets))
	for _, set := range sets {
		switch set.SetType {
		case models.CoinSetTypeStandard:
			batchableIDs = append(batchableIDs, set.ID)
		case models.CoinSetTypeGoal:
			batchableIDs = append(batchableIDs, set.ID)
			goalIDs = append(goalIDs, set.ID)
		}
	}

	batchedSummaries, err := s.repo.GetSetSummaries(batchableIDs, userID)
	if err != nil {
		return nil, err
	}
	batchedCompletions, err := s.repo.GetGoalSetCompletions(goalIDs, userID)
	if err != nil {
		return nil, err
	}
	batchable := make(map[uint]bool, len(batchableIDs))
	for _, id := range batchableIDs {
		batchable[id] = true
	}

	result := make([]map[string]interface{}, 0, len(sets))
	for _, set := range sets {
		var summary map[string]interface{}
		if batchable[set.ID] {
			summary = batchedSummaries[set.ID]
			if summary == nil {
				// No members, so the grouped query produced no row.
				summary = repository.EmptySetSummary()
			}
		} else {
			summary, err = s.repo.GetSetSummary(set.ID, userID)
			if err != nil {
				return nil, err
			}
		}

		var completion interface{}
		var checklist *checklistOverview
		switch {
		case set.SetType == models.CoinSetTypeChecklist:
			if checklist = s.checklistOverview(&set); checklist != nil {
				summary = map[string]interface{}{"coinCount": checklist.summary.Owned, "totalValue": checklist.summary.TotalValue}
				completion = checklist.summary.Completion()
			}
		case set.SetType == models.CoinSetTypeGoal:
			if pct, ok := batchedCompletions[set.ID]; ok {
				completion = pct
			} else {
				completion = 0.0
			}
		case set.SetType == models.CoinSetTypeAgentic:
			if c, err := s.repo.GetSetCompletion(set.ID, userID); err == nil {
				completion = c["completionPercentage"]
			}
		}

		setData := map[string]interface{}{
			"id":                   set.ID,
			"name":                 set.Name,
			"color":                set.Color,
			"icon":                 set.Icon,
			"setType":              set.SetType,
			"coinCount":            summary["coinCount"],
			"totalValue":           summary["totalValue"],
			"completionPercentage": completion,
			"valueChangePercent":   nil, // Will be populated in US3
			"agenticStatus":        set.AgenticStatus,
		}
		if checklist != nil {
			// Its slots country by country, for the sets' completion by country.
			setData["countries"] = checklist.countries
			setData["partial"] = !checklist.complete
		}
		result = append(result, setData)
	}

	return result, nil
}

// GetSetDetail returns detailed set information with aggregates.
func (s *SetService) GetSetDetail(setID, userID uint) (map[string]interface{}, error) {
	set, err := s.repo.GetByID(setID, userID)
	if err != nil {
		return nil, err
	}

	summary, err := s.repo.GetSetSummary(setID, userID)
	if err != nil {
		return nil, err
	}

	result := map[string]interface{}{
		"id":                   set.ID,
		"name":                 set.Name,
		"description":          set.Description,
		"color":                set.Color,
		"icon":                 set.Icon,
		"setType":              set.SetType,
		"parentSetId":          set.ParentSetID,
		"targetCompletionDate": set.TargetCompletionDate,
		"coinCount":            summary["coinCount"],
		"totalValue":           summary["totalValue"],
		"totalInvested":        summary["totalInvested"],
		"avgValuePerCoin":      summary["avgValuePerCoin"],
		"highestValueCoinId":   summary["highestValueCoinId"],
		"completionPercentage": nil,
		"agenticPrompt":        set.AgenticPrompt,
		"agenticStatus":        set.AgenticStatus,
	}
	if set.SetType == models.CoinSetTypeGoal || set.SetType == models.CoinSetTypeAgentic {
		if c, err := s.repo.GetSetCompletion(set.ID, userID); err == nil {
			result["completionPercentage"] = c["completionPercentage"]
		}
	}
	if set.SetType == models.CoinSetTypeChecklist {
		result["checklist"] = set.Checklist
		if checklist := s.checklistSummary(set); checklist != nil {
			result["coinCount"], result["totalValue"] = checklist.Owned, checklist.TotalValue
			result["completionPercentage"] = checklist.Completion()
		}
	}
	if set.SetType == models.CoinSetTypeSmart {
		result["smartCriteria"] = set.SmartCriteria
	}

	return result, nil
}

// CreateSet creates a new set with validation.
func (s *SetService) CreateSet(userID uint, input map[string]interface{}) (*models.CoinSet, error) {
	// Validate name
	name, ok := input["name"].(string)
	if !ok || strings.TrimSpace(name) == "" {
		return nil, fmt.Errorf("name is required")
	}
	name = strings.TrimSpace(name)
	if len(name) > maxSetNameLength {
		return nil, fmt.Errorf("name must be %d characters or less", maxSetNameLength)
	}

	// Check max sets limit
	count, err := s.repo.CountByUser(userID)
	if err != nil {
		return nil, err
	}
	if count >= maxSetsPerUser {
		return nil, fmt.Errorf("maximum of %d sets allowed", maxSetsPerUser)
	}

	// Check case-insensitive uniqueness
	exists, err := s.repo.ExistsByName(userID, name)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, fmt.Errorf("a set with this name already exists")
	}

	// Validate set type
	setType, ok := input["setType"].(string)
	if !ok || setType == "" {
		setType = string(models.CoinSetTypeStandard)
	}
	setType, err = normalizeSetType(setType)
	if err != nil {
		return nil, fmt.Errorf("invalid set type")
	}
	if setType == string(models.CoinSetTypeAgentic) {
		return nil, fmt.Errorf("agentic set creation requires the proposal review workflow; direct set creation is disabled")
	}
	creationMode := models.CoinSetCreationModeManual
	if rawMode, ok := input["creationMode"].(string); ok && strings.TrimSpace(rawMode) != "" {
		mode, err := normalizeCreationMode(rawMode)
		if err != nil {
			return nil, fmt.Errorf("invalid creation mode")
		}
		creationMode = mode
	}
	if setType == string(models.CoinSetTypeAgentic) {
		creationMode = models.CoinSetCreationModeDynamic
	}
	if creationMode == models.CoinSetCreationModeDynamic && setType != string(models.CoinSetTypeAgentic) {
		return nil, fmt.Errorf("dynamic creation mode is only valid for agentic sets")
	}

	var smartCriteria *models.JSONObject
	if raw, ok := input["smartCriteria"].(map[string]interface{}); ok && raw != nil {
		if err := ValidateSmartCriteria(raw); err != nil {
			return nil, err
		}
		criteria := models.JSONObject(raw)
		smartCriteria = &criteria
	}
	if setType == string(models.CoinSetTypeSmart) && smartCriteria == nil {
		return nil, fmt.Errorf("smart criteria is required for smart sets")
	}
	var checklist *ChecklistDefinition
	if setType == string(models.CoinSetTypeChecklist) {
		raw, ok := input["checklist"].(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("checklist is required for checklist sets")
		}
		definition, err := ParseChecklistDefinition(raw)
		if err != nil {
			return nil, err
		}
		if definition.Challenge == ChallengeJointIssue && s.checklist != nil && s.checklist.JointIssue(context.Background(), definition.Year) == nil {
			return nil, fmt.Errorf("the euro area had no joint issue in %d", definition.Year)
		}
		checklist = &definition
	}
	// Validate parent set ID if provided
	if parentSetID, ok := input["parentSetId"].(float64); ok && parentSetID > 0 {
		parentID := uint(parentSetID)
		parent, err := s.repo.GetByID(parentID, userID)
		if err != nil {
			if repository.IsRecordNotFound(err) {
				return nil, fmt.Errorf("parent set not found")
			}
			return nil, err
		}
		// Check for cycle (simplified - just check direct parent)
		if parent.ParentSetID != nil && *parent.ParentSetID == parentID {
			return nil, fmt.Errorf("parent set cycle detected")
		}
	}

	// Create the set
	color := getStringValue(input, "color")
	if color == "" {
		color = DefaultSetColor(userID, name, count)
	}
	set := &models.CoinSet{
		UserID:        userID,
		Name:          name,
		Description:   getStringValue(input, "description"),
		Color:         color,
		Icon:          getStringValue(input, "icon"),
		SetType:       models.CoinSetType(setType),
		CreationMode:  creationMode,
		SmartCriteria: smartCriteria,
		AgenticStatus: agenticInitialStatus(models.CoinSetType(setType)),
	}
	if checklist != nil {
		definition := checklist.JSON()
		set.Checklist = &definition
	}
	if rawDate := getStringValue(input, "targetCompletionDate"); rawDate != "" {
		t, err := time.Parse("2006-01-02", rawDate)
		if err != nil {
			return nil, fmt.Errorf("target completion date must be YYYY-MM-DD")
		}
		set.TargetCompletionDate = &t
	}

	if parentSetID, ok := input["parentSetId"].(float64); ok && parentSetID > 0 {
		pid := uint(parentSetID)
		set.ParentSetID = &pid
	}

	if err := s.repo.Create(set); err != nil {
		return nil, err
	}
	if templateID := getStringValue(input, "templateId"); templateID != "" {
		template := GetTemplateByID(templateID)
		if template == nil {
			return nil, fmt.Errorf("template not found")
		}
		if err := s.repo.CreateTargetsForSet(set.ID, CopyTemplateToCoinSetTargets(template, set.ID)); err != nil {
			return nil, err
		}
	}

	return set, nil
}

// UpdateSet updates a set with validation.
func (s *SetService) UpdateSet(setID, userID uint, updates map[string]interface{}) (*models.CoinSet, error) {
	set, err := s.repo.GetByID(setID, userID)
	if err != nil {
		return nil, err
	}

	// Validate name if provided
	if name, ok := updates["name"].(string); ok {
		name = strings.TrimSpace(name)
		if name == "" {
			return nil, newSetValidationError("name cannot be empty")
		}
		if len(name) > maxSetNameLength {
			return nil, newSetValidationError(fmt.Sprintf("name must be %d characters or less", maxSetNameLength))
		}
		// Check uniqueness (excluding current set)
		exists, err := s.repo.ExistsByName(userID, name)
		if err != nil {
			return nil, err
		}
		if exists && strings.ToLower(name) != strings.ToLower(set.Name) {
			return nil, newSetValidationError("a set with this name already exists")
		}
		updates["name"] = name
	}
	// A checklist is defined when created; its definition does not change.
	delete(updates, "checklist")
	if raw, ok := updates["smartCriteria"]; ok {
		delete(updates, "smartCriteria")
		criteria, isMap := raw.(map[string]interface{})
		if set.SetType != models.CoinSetTypeSmart || !isMap {
			return nil, newSetValidationError("only a smart set has criteria")
		}
		if err := ValidateSmartCriteria(criteria); err != nil {
			return nil, newSetValidationError(err.Error())
		}
		updates["smart_criteria"] = models.JSONObject(criteria)
	}
	if setTypeRaw, ok := updates["setType"].(string); ok {
		normalizedType, err := normalizeSetType(setTypeRaw)
		if err != nil {
			return nil, newSetValidationError("invalid set type")
		}
		if (normalizedType == string(models.CoinSetTypeChecklist)) != (set.SetType == models.CoinSetTypeChecklist) {
			return nil, newSetValidationError("a checklist set stays a checklist")
		}
		updates["setType"] = normalizedType
		if normalizedType != string(models.CoinSetTypeAgentic) && set.CreationMode == models.CoinSetCreationModeDynamic {
			updates["creationMode"] = string(models.CoinSetCreationModeManual)
		}
	}
	if modeRaw, ok := updates["creationMode"].(string); ok {
		mode, err := normalizeCreationMode(modeRaw)
		if err != nil {
			return nil, newSetValidationError("invalid creation mode")
		}
		targetType := set.SetType
		if updatedType, ok := updates["setType"].(string); ok {
			targetType = models.CoinSetType(updatedType)
		}
		if mode == models.CoinSetCreationModeDynamic && targetType != models.CoinSetTypeAgentic {
			return nil, newSetValidationError("dynamic creation mode is only valid for agentic sets")
		}
		updates["creationMode"] = string(mode)
	}

	if err := s.repo.Update(set, updates); err != nil {
		return nil, err
	}

	return s.repo.GetByID(setID, userID)
}

// DeleteSet deletes a set and its memberships.
func (s *SetService) DeleteSet(setID, userID uint) error {
	return s.repo.Delete(setID, userID)
}

// AddCoinToSet adds a coin to a set with validation.
func (s *SetService) AddCoinToSet(coinID, setID, userID uint, notes string, targetID ...uint) error {
	// Verify the set supports manual membership
	set, err := s.repo.GetByID(setID, userID)
	if err != nil {
		return err
	}
	if set.SetType == models.CoinSetTypeSmart {
		return fmt.Errorf("cannot manually add coins to smart sets")
	}
	if set.SetType == models.CoinSetTypeChecklist {
		return fmt.Errorf("cannot manually add coins to checklist sets")
	}
	if set.SetType == models.CoinSetTypeAgentic {
		if len(targetID) == 0 {
			return fmt.Errorf("targetId is required for agentic sets")
		}
		return s.repo.AddCoinToSet(coinID, setID, userID, notes, targetID[0])
	}

	return s.repo.AddCoinToSet(coinID, setID, userID, notes)
}

func (s *SetService) GetCompletion(setID, userID uint) (map[string]interface{}, error) {
	set, err := s.repo.GetByID(setID, userID)
	if err != nil {
		return nil, err
	}
	if set.SetType == models.CoinSetTypeChecklist {
		summary := s.checklistSummary(set)
		if summary == nil {
			summary = &ChecklistSummary{}
		}
		return map[string]interface{}{
			"totalTargets":         summary.Total,
			"completedTargets":     summary.Owned,
			"completionPercentage": summary.Completion(),
			"missingTargets":       []models.CoinSetTarget{},
		}, nil
	}
	return s.repo.GetSetCompletion(setID, userID)
}

func (s *SetService) CreateSnapshot(setID, userID uint) (*models.CoinSetValuationSnapshot, error) {
	var completion *float64
	set, err := s.repo.GetByID(setID, userID)
	if err != nil {
		return nil, err
	}
	if set.SetType == models.CoinSetTypeGoal || set.SetType == models.CoinSetTypeAgentic {
		c, err := s.repo.GetSetCompletion(setID, userID)
		if err == nil {
			if pct, ok := c["completionPercentage"].(float64); ok {
				completion = &pct
			}
		}
	}
	snapshot, err := s.repo.CreateSnapshot(setID, userID, completion)
	if err != nil {
		return nil, err
	}
	return snapshot, nil
}

func (s *SetService) CreateSnapshotsForAllUsers() error {
	users, err := s.repo.ListUsersWithSets()
	if err != nil {
		return err
	}
	for _, userID := range users {
		ids, err := s.repo.GetAllSetIDsForUser(userID)
		if err != nil {
			return err
		}
		for _, id := range ids {
			if _, err := s.CreateSnapshot(id, userID); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *SetService) GetTrends(setID, userID uint, rangeKey string) ([]models.CoinSetValuationSnapshot, error) {
	since := snapshotSince(rangeKey)
	return s.repo.GetSnapshots(setID, userID, since)
}

func (s *SetService) GetAnalytics(setID, userID uint) (map[string]interface{}, error) {
	detail, err := s.GetSetDetail(setID, userID)
	if err != nil {
		return nil, err
	}
	totalValue, _ := detail["totalValue"].(float64)
	totalInvested, _ := detail["totalInvested"].(float64)
	var roi *float64
	if totalInvested > 0 {
		v := ((totalValue - totalInvested) / totalInvested) * 100
		roi = &v
	}
	snapshots, err := s.repo.GetSnapshots(setID, userID, nil)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"roiPercent":              roi,
		"bestPerformerCoinId":     detail["highestValueCoinId"],
		"worstPerformerCoinId":    nil,
		"acquisitionRatePerMonth": acquisitionRate(snapshots),
		"projectedCompletionDate": nil,
	}, nil
}

func (s *SetService) PreviewSmartSet(userID uint, criteria map[string]interface{}) (map[string]interface{}, error) {
	if err := ValidateSmartCriteria(criteria); err != nil {
		return nil, err
	}
	coins, err := s.repo.GetCoinsMatchingCriteria(userID, criteria)
	if err != nil {
		return nil, err
	}
	ids := make([]uint, 0, len(coins))
	var totalValue float64
	for _, coin := range coins {
		ids = append(ids, coin.ID)
		if coin.CurrentValue != nil {
			totalValue += *coin.CurrentValue
		}
	}
	return map[string]interface{}{
		"coinIds":    ids,
		"coinCount":  len(coins),
		"totalValue": totalValue,
	}, nil
}

func snapshotSince(rangeKey string) *time.Time {
	now := time.Now()
	var since time.Time
	switch rangeKey {
	case "1m":
		since = now.AddDate(0, -1, 0)
	case "3m":
		since = now.AddDate(0, -3, 0)
	case "1y", "":
		since = now.AddDate(-1, 0, 0)
	case "all":
		return nil
	default:
		since = now.AddDate(-1, 0, 0)
	}
	return &since
}

func acquisitionRate(snapshots []models.CoinSetValuationSnapshot) *float64 {
	if len(snapshots) < 2 {
		return nil
	}
	days := snapshots[len(snapshots)-1].SnapshotDate.Sub(snapshots[0].SnapshotDate).Hours() / 24
	if days <= 0 {
		return nil
	}
	v := float64(snapshots[len(snapshots)-1].CoinCount-snapshots[0].CoinCount) / (days / 30)
	return &v
}

// RemoveCoinFromSet removes a coin from a manual set with validation.
func (s *SetService) RemoveCoinFromSet(coinID, setID, userID uint) error {
	// Verify the set supports manual membership
	set, err := s.repo.GetByID(setID, userID)
	if err != nil {
		return err
	}
	if set.SetType == models.CoinSetTypeSmart {
		return fmt.Errorf("cannot manually remove coins from smart sets")
	}
	if set.SetType == models.CoinSetTypeChecklist {
		return fmt.Errorf("cannot manually remove coins from checklist sets")
	}
	if set.SetType == models.CoinSetTypeAgentic {
		return fmt.Errorf("cannot manually remove coins from agentic sets")
	}

	return s.repo.RemoveCoinFromSet(coinID, setID, userID)
}

// GetCoinsInSet returns all coins in a set.
func (s *SetService) GetCoinsInSet(setID, userID uint) ([]models.Coin, error) {
	return s.repo.GetCoinsInSet(setID, userID)
}

// ReorderCoinsInSet saves the manual order for every current member of a non-smart set.
func (s *SetService) ReorderCoinsInSet(setID, userID uint, coinIDs []uint) error {
	set, err := s.repo.GetByID(setID, userID)
	if err != nil {
		return err
	}
	if set.SetType == models.CoinSetTypeSmart || set.SetType == models.CoinSetTypeChecklist {
		return ErrSmartSetOrder
	}

	seen := make(map[uint]struct{}, len(coinIDs))
	for _, coinID := range coinIDs {
		if coinID == 0 {
			return ErrInvalidSetOrder
		}
		if _, exists := seen[coinID]; exists {
			return ErrInvalidSetOrder
		}
		seen[coinID] = struct{}{}
	}

	if err := s.repo.ReorderCoinsInSet(setID, userID, coinIDs); err != nil {
		if errors.Is(err, repository.ErrInvalidSetOrder) {
			return ErrInvalidSetOrder
		}
		return err
	}
	return nil
}

// Helper functions
func getStringValue(input map[string]interface{}, key string) string {
	if val, ok := input[key].(string); ok {
		return strings.TrimSpace(val)
	}
	return ""
}

var setColorPalette = []string{
	"#c9a84c",
	"#b08d57",
	"#9b59b6",
	"#6b8e23",
	"#c0392b",
	"#4682b4",
	"#10b981",
	"#8b5cf6",
}

func DefaultSetColor(userID uint, name string, existingSetCount int64) string {
	h := fnv.New32a()
	_, _ = h.Write([]byte(fmt.Sprintf("%d:%s:%d", userID, strings.ToLower(strings.TrimSpace(name)), existingSetCount)))
	return setColorPalette[int(h.Sum32())%len(setColorPalette)]
}

func agenticInitialStatus(setType models.CoinSetType) string {
	if setType == models.CoinSetTypeAgentic {
		return "generating"
	}
	return "ready"
}

// CreateSetFromTemplate creates a goal set from a template.
func (s *SetService) CreateSetFromTemplate(userID uint, input map[string]interface{}) (*models.CoinSet, error) {
	// Get template ID
	templateID, ok := input["templateId"].(string)
	if !ok || templateID == "" {
		return nil, fmt.Errorf("templateId is required")
	}

	// Get template
	template := GetTemplateByID(templateID)
	if template == nil {
		return nil, fmt.Errorf("template not found")
	}

	setType := string(models.CoinSetTypeGoal)
	if st, ok := input["setType"].(string); ok && strings.TrimSpace(st) != "" {
		normalizedType, err := normalizeSetType(st)
		if err != nil || normalizedType != string(models.CoinSetTypeGoal) {
			return nil, fmt.Errorf("template import only supports goal sets")
		}
		setType = normalizedType
	}

	// Create the set
	setInput := map[string]interface{}{
		"name":        input["name"],
		"description": fmt.Sprintf("%s (from template: %s)", template.Description, template.Name),
		"setType":     setType,
	}
	if color, ok := input["color"]; ok {
		setInput["color"] = color
	}
	if icon, ok := input["icon"]; ok {
		setInput["icon"] = icon
	}
	if targetDate, ok := input["targetCompletionDate"]; ok {
		setInput["targetCompletionDate"] = targetDate
	}

	set, err := s.CreateSet(userID, setInput)
	if err != nil {
		return nil, err
	}

	// Copy template targets to set targets
	targets := CopyTemplateToCoinSetTargets(template, set.ID)
	if err := s.repo.CreateTargetsForSet(set.ID, targets); err != nil {
		// Rollback - delete the set
		s.repo.Delete(set.ID, userID)
		return nil, fmt.Errorf("failed to create targets: %w", err)
	}

	return set, nil
}

// CreateSetFromCSV creates a goal set from CSV target import.
func (s *SetService) CreateSetFromCSV(userID uint, input map[string]interface{}, csvContent string) (*models.CoinSet, error) {
	// Parse CSV
	importer := NewSetTargetImport()
	targets, err := importer.ParseCSV(strings.NewReader(csvContent))
	if err != nil {
		return nil, err
	}

	// Validate targets
	if err := importer.ValidateTargets(targets); err != nil {
		return nil, err
	}

	setType := string(models.CoinSetTypeGoal)
	if st, ok := input["setType"].(string); ok && strings.TrimSpace(st) != "" {
		normalizedType, err := normalizeSetType(st)
		if err != nil || normalizedType != string(models.CoinSetTypeGoal) {
			return nil, fmt.Errorf("CSV import only supports goal sets")
		}
		setType = normalizedType
	}

	// Create the set
	setInput := map[string]interface{}{
		"name":        input["name"],
		"description": input["description"],
		"setType":     setType,
	}
	if color, ok := input["color"]; ok {
		setInput["color"] = color
	}
	if icon, ok := input["icon"]; ok {
		setInput["icon"] = icon
	}

	set, err := s.CreateSet(userID, setInput)
	if err != nil {
		return nil, err
	}

	// Create targets
	if err := s.repo.CreateTargetsForSet(set.ID, targets); err != nil {
		// Rollback - delete the set
		s.repo.Delete(set.ID, userID)
		return nil, fmt.Errorf("failed to create targets: %w", err)
	}

	return set, nil
}

func normalizeSetType(raw string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "open", string(models.CoinSetTypeStandard):
		return string(models.CoinSetTypeStandard), nil
	case "defined", string(models.CoinSetTypeGoal):
		return string(models.CoinSetTypeGoal), nil
	case string(models.CoinSetTypeSmart):
		return string(models.CoinSetTypeSmart), nil
	case "tracker", string(models.CoinSetTypeAgentic):
		return string(models.CoinSetTypeAgentic), nil
	case string(models.CoinSetTypeChecklist):
		return string(models.CoinSetTypeChecklist), nil
	case "dynamic":
		return "", fmt.Errorf("dynamic is a workflow, not a set type")
	default:
		return "", fmt.Errorf("invalid set type")
	}
}

func normalizeCreationMode(raw string) (models.CoinSetCreationMode, error) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case string(models.CoinSetCreationModeManual):
		return models.CoinSetCreationModeManual, nil
	case string(models.CoinSetCreationModeDynamic):
		return models.CoinSetCreationModeDynamic, nil
	default:
		return "", fmt.Errorf("invalid creation mode")
	}
}
