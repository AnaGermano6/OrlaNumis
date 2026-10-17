// stats types. Split out of the former single-file src/types/index.ts;
// re-exported from '@/types' so existing imports keep working.

export interface StatsResponse {
  totalCoins: number
  totalWishlist: number
  byCategory: { category: string; count: number }[]
  byMaterial: { material: string; count: number }[]
  byGrade: { grade: string; count: number }[]
  byEra: { era: string; count: number }[]
  byRuler: { ruler: string; count: number }[]
  byPriceRange: { range: string; count: number }[]
  values: {
    totalPurchasePrice: number
    totalCurrentValue: number
    avgPurchasePrice: number
    avgCurrentValue: number
  }
}

export type HealthGrade = 'A' | 'B' | 'C' | 'D' | 'F'

export type HealthChecklistDimension = 'metadata' | 'images' | 'valuation' | 'ai'

export type HealthChecklistSeverity = 'high' | 'medium' | 'low'

export type HealthQuickAction = 'edit_metadata' | 'upload_images' | 'run_valuation' | 'run_ai_analysis'

export interface HealthWeights {
  metadata: number
  imageCoverage: number
  valuationFreshness: number
  aiCoverage: number
}

export interface HealthDimensions {
  metadata: number
  imageCoverage: number
  valuationFreshness: number
  aiCoverage: number
}

export interface CollectionHealthSummary {
  score: number
  grade: HealthGrade
  eligibleCoinCount: number
  weights: HealthWeights
  dimensions: HealthDimensions
}

export interface MissingChecklistItem {
  key: string
  dimension: HealthChecklistDimension
  label: string
  severity: HealthChecklistSeverity
  actionHint: HealthQuickAction
}

export interface CoinHealthItem {
  coinId: number
  title: string
  score: number
  grade: HealthGrade
  dimensions: HealthDimensions
  missingItems: MissingChecklistItem[]
  quickActions: HealthQuickAction[]
}

export interface CoinHealthListResponse {
  coins: CoinHealthItem[]
  pagination: {
    page: number
    limit: number
    total: number
  }
}

export interface MissingFieldStat {
  key: string
  count: number
  percentage: number
}

export interface AdminHealthSummaryResponse {
  medianScore: number
  lowScorePercentage: number
  lowScoreThreshold: number
  eligibleCoinCount: number
  topMissingFields: MissingFieldStat[]
}

export interface PortfolioSummary {
  totalCoins: number
  totalValue: number
  totalInvested: number
  categories: { category: string; count: number }[]
  materials: { material: string; count: number }[]
  eras: { era: string; count: number }[]
  rulers: { ruler: string; count: number }[]
  topCoins: { name: string; category: string; currentValue: number | null; ruler: string; era: string }[]
  missingFields?: Record<string, number>
}

// Collection Time Machine — the collection reconstructed as it stood on a past
// date. Mirrors services.TimeMachineSnapshot / services.TimeMachineBounds in
// the Go API.

export interface TimeMachineBreakdownEntry {
  label: string
  count: number
  value: number
}

export interface TimeMachineTopCoin {
  id: number
  name: string
  value: number
  /** false means the figure is the purchase price, not a recorded valuation. */
  valueFromHistory: boolean
}

export interface TimeMachineValueBasis {
  fromValuationHistory: number
  fromPurchasePrice: number
}

export interface TimeMachineSnapshot {
  asOfDate: string
  coinCount: number
  totalValue: number
  totalInvested: number
  unrealizedGain: number
  byCategory: TimeMachineBreakdownEntry[]
  byMaterial: TimeMachineBreakdownEntry[]
  byEra: TimeMachineBreakdownEntry[]
  topCoins: TimeMachineTopCoin[]
  acquiredInYear: number
  valueBasis: TimeMachineValueBasis
  /** Owned coins with no purchase date, which appear at no point on the timeline. */
  undatedCoinCount: number
}

export interface TimeMachineBounds {
  earliestDate: string
  latestDate: string
  hasData: boolean
}
