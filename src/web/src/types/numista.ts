// numista types. Split out of the former single-file src/types/index.ts;
// re-exported from '@/types' so existing imports keep working.

export interface NumistaType {
  id: number
  title: string
  category: string
  issuer?: { name: string }
  min_year?: number
  max_year?: number
  obverse_thumbnail?: string
  reverse_thumbnail?: string
}

export interface NumistaSearchResponse {
  count: number
  types: NumistaType[]
}

export type NumistaLookupPath = 'direct' | 'photo'

export type NumistaLookupStatus = 'success' | 'empty' | 'unconfigured' | 'quota-limited' | 'timeout' | 'unavailable'

export type NumistaQuerySource = 'generated' | 'user-edited' | 'manual'

export type NumistaSearchAttempt = 'primary' | 'relaxed'

export type NumistaQueryGenerationVersion = 'numista-query-v2'

export type NumistaEnrichmentState = 'not_requested' | 'enriched' | 'cached' | 'failed'

export type NumistaRelevanceField = 'exact_id' | 'title' | 'issuer' | 'denomination' | 'mint' | 'date' | 'material' | 'inscription'

export type NumistaRelevanceKind = 'match' | 'conflict' | 'unavailable'

export type NumistaRelevanceBand = 'strong' | 'possible' | 'weak'

export interface NumistaEvidence {
  title?: string
  issuer?: string
  denomination?: string
  mint?: string
  dateText?: string
  material?: string
  obverseInscription?: string
  reverseInscription?: string
  reverseType?: string
  visibleText?: string
  exactNumistaId?: number
}

export interface NumistaQueryProposalRequest {
  path: NumistaLookupPath
  evidence: NumistaEvidence
}

export interface NumistaQueryProposal {
  query: string
  querySource: 'generated'
  generationVersion: NumistaQueryGenerationVersion
}

export interface NumistaLookupRequest {
  query: string
  path: NumistaLookupPath
  evidence: NumistaEvidence
  querySource: NumistaQuerySource
  generationVersion?: NumistaQueryGenerationVersion
}

export interface NumistaEnrichmentRequest extends NumistaLookupRequest {
  candidates: NumistaCandidate[]
}

export interface NumistaRelevanceReason {
  field: NumistaRelevanceField
  kind: NumistaRelevanceKind
  code: string
  label: string
}

export interface NumistaRelevanceAssessment {
  scoringVersion: 'numista-v1'
  score: number
  band: NumistaRelevanceBand
  reasons: NumistaRelevanceReason[]
}

export interface NumistaCandidate {
  id: number
  canonicalUrl: string
  title: string
  issuer?: string
  denomination?: string
  mint?: string
  minYear?: number
  maxYear?: number
  yearDisplay?: string
  material?: string
  obverseInscription?: string
  reverseInscription?: string
  obverseThumbnail?: string
  reverseThumbnail?: string
  providerPosition: number
  enrichmentState: NumistaEnrichmentState
  assessment: NumistaRelevanceAssessment
}

export interface NumistaCacheMetadata {
  hit: boolean
  coalesced: boolean
  createdAt: string
  expiresAt: string
  ageSeconds: number
}

export interface NumistaLookupOutcome {
  status: NumistaLookupStatus
  effectiveQuery: string
  candidates: NumistaCandidate[]
  guidanceCode?: string
  retryAfterSeconds?: number
  cache?: NumistaCacheMetadata
  stage: 'broad' | 'enriched'
  querySource: NumistaQuerySource
  searchAttempt: NumistaSearchAttempt
  searchAttemptCount: 1 | 2
}

export interface NumistaSettings {
  NumistaSearchTTLHours: string
  NumistaDetailTTLHours: string
  NumistaEnrichmentLimit: string
  NumistaSearchResultLimit: string
  NumistaSearchTimeoutSeconds: string
  NumistaDetailTimeoutSeconds: string
}

export type NumistaStatusCounts = Partial<Record<NumistaLookupStatus, number>>

export interface NumistaHealthSummary {
  configured: boolean
  configurationValid: boolean
  lastOutcome?: NumistaLookupStatus | null
  lastCheckedAt?: string | null
  statusCounts: NumistaStatusCounts
  broadRequestCount: number
  detailRequestCount: number
  freshCacheHitCount: number
  coalescedRequestCount: number
  providerLoadCount: number
  providerFailureCount: number
  cancelledRequestCount: number
  freshCacheHitRate: number
  p50ElapsedMs: number
  p95ElapsedMs: number
  enrichmentAttempted: number
  enrichmentSucceeded: number
  enrichmentFailed: number
  lastQuotaLimitedAt?: string | null
  lastRetryAfterSeconds?: number | null
  /** This month's request count against the quota. */
  usage?: NumistaUsageSummary
}

/** The month's Numista requests: above pauseAt, background requests wait for the next month. */
export interface NumistaUsageSummary {
  month: string
  requests: number
  limit: number
  pauseAt: number
  backgroundPaused: boolean
}

/** One batch of coin values updated from Numista's prices. */
export interface NumistaValueRun {
  id: number
  triggerType: 'scheduled' | 'manual' | string
  status: 'running' | 'completed' | 'budget_spent' | 'paused' | 'unconfigured' | 'failed' | string
  startedAt: string
  completedAt?: string | null
  /** Numista requests the batch could make, and made. */
  budget: number
  requests: number
  checked: number
  updated: number
  unchanged: number
  /** Coins whose issue has no price on Numista. */
  noPrice: number
  /** Coins Numista did not answer for; tried again next time. */
  failed: number
}

/** How the daily update of coin values from Numista stands. */
export interface NumistaValueStatus {
  enabled: boolean
  isRunning: boolean
  /** Go duration ("5h0m0s") until the next scheduled batch; empty when off. */
  nextRunIn: string
  usage: NumistaUsageSummary
  /** ok, low (from 75% of the quota) or paused (no background requests until resumesOn). */
  quotaLevel: 'ok' | 'low' | 'paused'
  /** First day of next month (UTC), YYYY-MM-DD. */
  resumesOn: string
  /** Requests the next batch may make. */
  dailyBudget: number
  /** Coins linked to Numista with a grade Numista prices; due: not checked in 30 days. */
  eligible: number
  due: number
  /** Linked coins without a grade Numista prices (none, or a proof). */
  withoutGrade: number
  currency: string
  runs: NumistaValueRun[]
}

export interface SelectedNumistaReference {
  catalog: 'Numista'
  number: string
  uri: string
}
