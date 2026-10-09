import { t, type MessageKey } from '@/i18n'

// Friendly Coin Copilot helper names. Must match _TOOL_LABELS in
// src/agent/app/teams/coin_copilot.py so the checklist and result lines agree;
// copilotToolLabels.test.ts fails if the two drift.
export const COPILOT_TOOL_LABELS: Readonly<Record<string, string>> = {
  search_my_collection: 'Search owned collection',
  get_coin: 'Read coin details',
  collection_summary: 'Summarize collection',
  top_coins_by_value: 'Review top recorded values',
  portfolio_review: 'Review collection portfolio',
  gap_analysis: 'Analyze collection gaps',
  market_search: 'Search dealer listings',
  auction_search: 'Search auction lots',
  price_trends: 'Analyze completed-sale price trends',
  similar_lots: 'Find similar auction lots',
  deep_analysis_handoff: 'Use existing Deep Analysis',
}

const COPILOT_TOOL_LABEL_KEYS: Readonly<Record<string, MessageKey>> = {
  search_my_collection: 'copilot.tools.searchMyCollection',
  get_coin: 'copilot.tools.getCoin',
  collection_summary: 'copilot.tools.collectionSummary',
  top_coins_by_value: 'copilot.tools.topCoinsByValue',
  portfolio_review: 'copilot.tools.portfolioReview',
  gap_analysis: 'copilot.tools.gapAnalysis',
  market_search: 'copilot.tools.marketSearch',
  auction_search: 'copilot.tools.auctionSearch',
  price_trends: 'copilot.tools.priceTrends',
  similar_lots: 'copilot.tools.similarLots',
  deep_analysis_handoff: 'copilot.tools.deepAnalysisHandoff',
}

// Unknown helpers get a generic name rather than an internal tool name. The
// English catalog mirrors COPILOT_TOOL_LABELS; other locales translate it.
export function copilotToolLabel(name: string): string {
  const key = COPILOT_TOOL_LABEL_KEYS[name]
  return key ? t(key) : t('copilot.tools.fallback')
}
