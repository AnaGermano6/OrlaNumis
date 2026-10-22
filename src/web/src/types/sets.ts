// sets types. Split out of the former single-file src/types/index.ts;
// re-exported from '@/types' so existing imports keep working.
import type { Coin } from '@/types/coin'

export interface CollectionSetOption {
  id: number
  name: string
  color: string
}

export type CoinSetType = 'standard' | 'goal' | 'smart' | 'agentic' | 'checklist'

export type LegacyCoinSetType = 'open' | 'defined' | 'tracker' | 'dynamic'

export type CoinSetTypeResponse = CoinSetType | LegacyCoinSetType

export function normalizeCoinSetType(setType: CoinSetTypeResponse): CoinSetType {
  if (setType === 'open') return 'standard'
  if (setType === 'defined') return 'goal'
  if (setType === 'tracker' || setType === 'dynamic') return 'agentic'
  return setType
}

export interface CoinSet {
  id: number
  userId: number
  name: string
  description?: string
  color: string
  icon?: string
  setType: CoinSetTypeResponse
  parentSetId?: number | null
  targetCompletionDate?: string | null
  createdAt: string
  updatedAt: string
}

export interface CoinSetSummary {
  id: number
  name: string
  color: string
  icon?: string
  setType: CoinSetTypeResponse
  coinCount: number
  totalValue: number
  completionPercentage?: number | null
  valueChangePercent?: number | null
  agenticStatus?: string | null
}

export interface CoinSetDetail extends CoinSetSummary {
  description?: string
  /** A checklist set's definition: the coins it lists. */
  checklist?: ChecklistDefinition | null
  /** A smart set's filters. */
  smartCriteria?: SmartCriteriaGroup | null
  parentSetId?: number | null
  targetCompletionDate?: string | null
  totalInvested: number
  avgValuePerCoin?: number | null
  highestValueCoinId?: number | null
  agenticPrompt?: string | null
  agenticStatus?: string | null
}

export type CoinRecommendationTargetType = 'set'

export type CoinRecommendationStatus = 'pending' | 'accepted' | 'rejected' | 'dismissed'

export interface CoinRecommendation {
  id: number
  targetType: CoinRecommendationTargetType
  targetId: number
  targetName: string
  score: number
  confidence: 'high' | 'medium' | 'low'
  reasons: string[]
  status: CoinRecommendationStatus
}

export interface CreateCoinSetRequest {
  name: string
  description?: string
  color?: string
  icon?: string
  setType: CoinSetType
  parentSetId?: number | null
  targetCompletionDate?: string | null
  smartCriteria?: Record<string, unknown> | null
  templateId?: string | null
  checklist?: ChecklistDefinition
}

/** The challenges a checklist set can be. */
export type ChecklistChallenge = 'euro-countries' | 'joint-issue' | 'annual-series' | 'escudo-types' | 'reis-types'

/**
 * What a checklist set lists: a country's coins in the catalogue (all, or
 * of one face value), or a challenge. Its slots are worked out on the
 * server, and fill by themselves as the collection grows.
 */
export interface ChecklistDefinition {
  source: 'catalogue' | 'challenge'
  /** ISO code: the catalogue's country, or the annual series'. */
  country?: string
  /** "PTE" for Portugal's escudos; the euro otherwise. */
  currency?: string
  faceValue?: number
  challenge?: ChecklistChallenge
  /** The annual series' or the joint issue's year. */
  year?: number
}

/** A coin a checklist set lists, filled by a coin of the collection or not. */
export interface ChecklistSlot {
  key: string
  /** The coin's name; a challenge's country slots leave it empty. */
  title?: string
  /** ISO code. */
  country?: string
  year?: number
  /** A type's last year, when it was struck over several. */
  toYear?: number
  faceValue?: number
  faceValues?: number[]
  currency?: string
  /** The catalogue's picture and page. */
  imageUrl?: string
  pageUrl?: string
  owned: boolean
  /** The collection's coin filling the slot. */
  coinId?: number
  coinName?: string
  coinImage?: string
  coinValue?: number
}

export interface ChecklistResponse {
  slots: ChecklistSlot[]
  /** False while the catalogue behind the slots is partly listed. */
  complete: boolean
}

export interface CreateCoinSetFromCsvRequest extends CreateCoinSetRequest {
  csv: string
}

export type UpdateCoinSetRequest = Partial<CreateCoinSetRequest>

export interface AddCoinToSetRequest {
  coinId: number
  notes?: string
  targetId?: number
}

export interface ReorderSetCoinsRequest {
  coinIds: number[]
}

// US2: Defined/Goal Sets and Completion
export interface CoinSetTarget {
  id: number
  setId: number
  label: string
  year?: number | null
  mintMark?: string | null
  denomination?: string | null
  country?: string | null
  material?: string | null
  matchRules?: Record<string, unknown> | null
  sortOrder: number
  createdAt?: string
}

export interface CoinSetCompletion {
  totalTargets: number
  completedTargets: number
  completionPercentage: number
  missingTargets: CoinSetTarget[]
  targets?: CoinSetTarget[]
  targetMatches?: Array<{
    target: CoinSetTarget
    coin?: Coin | null
  }>
  collectionItems?: number
  wishlistItems?: number
}

export interface CoinSetTemplateTarget {
  label: string
  year?: number | null
  mintMark?: string | null
  denomination?: string | null
  country?: string | null
  material?: string | null
  sortOrder: number
}

// US3: Snapshots and Trends
export interface CoinSetSnapshot {
  id?: number
  setId?: number
  userId?: number
  snapshotDate: string
  totalValue: number
  totalInvested: number
  coinCount: number
  completionPercentage?: number | null
  avgValuePerCoin?: number | null
  highestValueCoinId?: number | null
}

export interface CoinSetAnalytics {
  roiPercent?: number | null
  bestPerformerCoinId?: number | null
  worstPerformerCoinId?: number | null
  acquisitionRatePerMonth?: number | null
  projectedCompletionDate?: string | null
}

export type SmartCriteriaOperator = 'and' | 'or'

export type SmartCriteriaRuleOp = 'eq' | 'neq' | 'contains' | 'startsWith' | 'in' | 'between' | 'gte' | 'lte' | 'isNull' | 'isNotNull'

export interface SmartCriteriaRule {
  field: string
  op: SmartCriteriaRuleOp
  value?: unknown
}

export interface SmartCriteriaGroup {
  operator: SmartCriteriaOperator
  rules: Array<SmartCriteriaRule | SmartCriteriaGroup>
}

export interface SmartSetPreview {
  coinIds: number[]
  coinCount: number
  totalValue: number
}

