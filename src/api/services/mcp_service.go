package services

import (
	"errors"

	"github.com/briandenicola/ancient-coins-api/models"
	"github.com/briandenicola/ancient-coins-api/repository"
)

var ErrMCPInvalidRequest = errors.New("invalid MCP request")

type MCPService struct {
	collectionSvc *CollectionToolsService
	coinRepo      *repository.CoinRepository
	copilotSvc    *CoinCopilotService
}

func NewMCPService(
	collectionSvc *CollectionToolsService,
	coinRepo *repository.CoinRepository,
	copilotSvc *CoinCopilotService,
) *MCPService {
	return &MCPService{
		collectionSvc: collectionSvc,
		coinRepo:      coinRepo,
		copilotSvc:    copilotSvc,
	}
}

func (s *MCPService) SearchCollection(userID uint, query string, limit *int) ([]CollectionCoinSummary, error) {
	if err := validateMCPQueryLimit(query, limit, 20); err != nil {
		return nil, err
	}
	return s.collectionSvc.SearchActiveCollection(userID, query, limit)
}

func (s *MCPService) GetCoin(userID, coinID uint) (*CollectionCoinSummary, error) {
	return s.collectionSvc.GetCoin(userID, coinID)
}

func (s *MCPService) ListWishlist(userID uint, query string, limit *int) ([]CollectionCoinSummary, error) {
	if err := validateMCPQueryLimit(query, limit, 50); err != nil {
		return nil, err
	}
	return s.collectionSvc.SearchWishlist(userID, query, limit)
}

func (s *MCPService) CollectionStats(userID uint) (*repository.CollectionStats, error) {
	return s.coinRepo.GetStats(userID)
}

func (s *MCPService) TopCoinsByValue(userID uint, limit *int) ([]CollectionCoinSummary, error) {
	if limit != nil && (*limit < 1 || *limit > 10) {
		return nil, ErrMCPInvalidRequest
	}
	return s.collectionSvc.TopCoinsByValue(userID, limit)
}

func (s *MCPService) StartCopilot(userID uint, input CoinCopilotStartInput) (*models.CoinCopilotRun, bool, error) {
	return s.copilotSvc.Start(userID, input)
}

func (s *MCPService) GetCopilotRun(userID uint, runID string) (*models.CoinCopilotRun, error) {
	return s.copilotSvc.GetRun(userID, runID)
}

func (s *MCPService) ResumeCopilot(userID uint, runID string, input CoinCopilotResumeInput) (*models.CoinCopilotRun, bool, error) {
	return s.copilotSvc.Resume(userID, runID, input)
}

func (s *MCPService) CancelCopilot(userID uint, runID string) (*models.CoinCopilotRun, bool, error) {
	return s.copilotSvc.Cancel(userID, runID)
}

func validateMCPQueryLimit(query string, limit *int, max int) error {
	if len(query) > 200 || (limit != nil && (*limit < 1 || *limit > max)) {
		return ErrMCPInvalidRequest
	}
	return nil
}
