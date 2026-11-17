// stats types. Split out of the former single-file src/types/index.ts;
// re-exported from '@/types' so existing imports keep working.

export interface StatsResponse {
  totalCoins: number
  totalWishlist: number
  byCategory: { category: string; count: number }[]
  byMaterial: { material: string; count: number }[]
  byGrade: { grade: string; count: number }[]
  byEra: { era: string; count: number }[]
  /** Coins of each monetary system, most first. */
  byMonetarySystem?: { monetarySystem: string; count: number }[]
  byRuler: { ruler: string; count: number }[]
  byPriceRange: { range: string; count: number }[]
  values: {
    totalPurchasePrice: number
    totalCurrentValue: number
    avgPurchasePrice: number
    avgCurrentValue: number
  }
  /** Coins of each decade of issue (1990 for 1990–1999, negative before Christ), oldest first. */
  byDecade?: { decade: number; count: number }[]
  /** The collection's gold and silver: fine (where the fineness is known) and gross. */
  preciousMetals?: { gold: MetalTotal; silver: MetalTotal }
}

/** How much of a precious metal a collection holds, in grams. */
export interface MetalTotal {
  fineGrams: number
  grossGrams: number
  pieces: number
  /** Pieces whose fineness nobody says: in the gross weight, not the fine. */
  unknownFineness: number
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
  /** The coin's primary (or obverse, or first) image, if any. */
  imagePath?: string
}

export interface CoinHealthListResponse {
  coins: CoinHealthItem[]
  pagination: {
    page: number
    limit: number
    total: number
  }
  /**
   * For the priority list: its coins by country ("" for none) and by missing
   * item's key, each counted with the other filter on.
   */
  facets?: {
    countries: { value: string; count: number }[]
    missing: { value: string; count: number }[]
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
