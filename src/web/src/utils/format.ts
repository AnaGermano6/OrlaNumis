import { intlLocale, t, type MessageKey } from '@/i18n'
import { collectionCurrency } from '@/utils/collectionCurrency'

/** Formats an amount; without a currency, in the collection currency. */
export function formatCurrency(value: number, currency?: string): string {
  return new Intl.NumberFormat(intlLocale(), { style: 'currency', currency: currency || collectionCurrency.value }).format(value)
}

/**
 * Translates an AI confidence level ("high" | "medium" | "low") for display.
 * Unknown values are shown as-is.
 */
export function confidenceLabel(level: string | null | undefined): string {
  switch ((level ?? '').toLowerCase()) {
    case 'high': return t('common.confidence.high')
    case 'medium': return t('common.confidence.medium')
    case 'low': return t('common.confidence.low')
    default: return level ?? ''
  }
}

/** Wishlist search alert run status, lower-case as it appears mid-sentence ("rate limited"). */
export function alertRunStatusLabel(status: string): string {
  switch (status) {
    case 'queued': return t('wishlistAlerts.runStatus.queued')
    case 'running': return t('wishlistAlerts.runStatus.running')
    case 'completed': return t('wishlistAlerts.runStatus.completed')
    case 'failed': return t('wishlistAlerts.runStatus.failed')
    case 'partial': return t('wishlistAlerts.runStatus.partial')
    case 'rate_limited': return t('wishlistAlerts.runStatus.rateLimited')
    case 'cancelled': return t('wishlistAlerts.runStatus.cancelled')
    default: return status.replace(/_/g, ' ')
  }
}

/** Wishlist search alert candidate lifecycle state ("Needs review"). */
export function candidateStateLabel(state: string): string {
  switch (state) {
    case 'active': return t('wishlistAlerts.candidateState.active')
    case 'needs_review': return t('wishlistAlerts.candidateState.needsReview')
    case 'dismissed': return t('wishlistAlerts.candidateState.dismissed')
    case 'converted': return t('wishlistAlerts.candidateState.converted')
    case 'suppressed': return t('wishlistAlerts.candidateState.suppressed')
    default: return state.replace(/_/g, ' ')
  }
}

/** Wishlist search alert candidate provenance status ("Verified"). */
export function provenanceStatusLabel(status: string): string {
  switch (status) {
    case 'verified': return t('wishlistAlerts.provenance.verified')
    case 'partial': return t('wishlistAlerts.provenance.partial')
    case 'unverified': return t('wishlistAlerts.provenance.unverified')
    default: return status.replace(/_/g, ' ')
  }
}

/** Wishlist search alert cadence ("Weekly"). */
export function alertCadenceLabel(cadence: string): string {
  switch (cadence) {
    case 'manual': return t('wishlistAlerts.cadence.manual')
    case 'daily': return t('wishlistAlerts.cadence.daily')
    case 'weekly': return t('wishlistAlerts.cadence.weekly')
    case 'monthly': return t('wishlistAlerts.cadence.monthly')
    default: return cadence
  }
}

/** Display label for a scheduler run trigger (manual/scheduled). */
export function runTriggerLabel(trigger: string): string {
  switch (trigger) {
    case 'manual': return t('admin.schedules.runTriggers.manual')
    case 'scheduled': return t('admin.schedules.runTriggers.scheduled')
    default: return trigger
  }
}

/** Display label for a scheduler run status; unknown values are shown as stored. */
export function runStatusLabel(status: string): string {
  switch (status) {
    case 'success': return t('admin.schedules.runStatuses.success')
    case 'error': return t('admin.schedules.runStatuses.error')
    case 'running': return t('admin.schedules.runStatuses.running')
    case 'queued': return t('admin.schedules.runStatuses.queued')
    case 'completed': return t('admin.schedules.runStatuses.completed')
    case 'failed': return t('admin.schedules.runStatuses.failed')
    case 'cancelled': return t('admin.schedules.runStatuses.cancelled')
    case 'skipped': return t('admin.schedules.runStatuses.skipped')
    case 'done': return t('admin.schedules.runStatuses.done')
    case 'partial_failure': return t('admin.schedules.runStatuses.partialFailure')
    default: return status
  }
}

/** Display label for a Deep Analysis job status; unknown values are shown as stored. */
export function deepJobStatusLabel(status: string): string {
  switch (status) {
    case 'queued': return t('deepAnalysis.jobStatus.queued')
    case 'running': return t('deepAnalysis.jobStatus.running')
    case 'partial': return t('deepAnalysis.jobStatus.partial')
    case 'completed': return t('deepAnalysis.jobStatus.completed')
    case 'failed': return t('deepAnalysis.jobStatus.failed')
    case 'cancelled': return t('deepAnalysis.jobStatus.cancelled')
    default: return status
  }
}

/** Display label for a Deep Analysis provider status; unknown values are title-cased. */
export function deepProviderStatusLabel(status: string): string {
  switch (status) {
    case 'pending': return t('deepAnalysis.providerStatus.pending')
    case 'running': return t('deepAnalysis.providerStatus.running')
    case 'contributed': return t('deepAnalysis.providerStatus.contributed')
    case 'no_match': return t('deepAnalysis.providerStatus.noMatch')
    case 'failed': return t('deepAnalysis.providerStatus.failed')
    case 'timed_out': return t('deepAnalysis.providerStatus.timedOut')
    case 'skipped': return t('deepAnalysis.providerStatus.skipped')
    case 'not_automated': return t('deepAnalysis.providerStatus.notAutomated')
    case 'unavailable': return t('deepAnalysis.providerStatus.unavailable')
    default:
      return status
        .split('_')
        .map((part) => (part ? part.charAt(0).toUpperCase() + part.slice(1) : part))
        .join(' ')
  }
}

/** Display label for a coin material value (stored values stay in English). */
export function materialLabel(material: string | null | undefined): string {
  switch (material) {
    case 'Gold': return t('materials.gold')
    case 'Silver': return t('materials.silver')
    case 'Bronze': return t('materials.bronze')
    case 'Copper': return t('materials.copper')
    case 'Electrum': return t('materials.electrum')
    // What banknotes are printed on.
    case 'Paper': return t('banknoteForm.materials.paper')
    case 'Polymer': return t('banknoteForm.materials.polymer')
    case 'Hybrid': return t('banknoteForm.materials.hybrid')
    case 'Other': return t('materials.other')
    default: return material ?? ''
  }
}

/** Display label for a coin image type (`obverse`, `reverse`, `detail`, `other`). */
export function imageTypeLabel(type: string): string {
  switch (type) {
    case 'obverse': return t('common.obverse')
    case 'reverse': return t('common.reverse')
    case 'detail': return t('imageTypes.detail')
    case 'other': return t('imageTypes.other')
    default: return type ? type.charAt(0).toUpperCase() + type.slice(1) : ''
  }
}

const FIELD_NAME_KEYS: Record<string, MessageKey> = {
  name: 'coinFields.name',
  category: 'coinFields.category',
  material: 'coinFields.material',
  denomination: 'coinFields.denomination',
  ruler: 'coinFields.ruler',
  era: 'coinFields.era',
  mint: 'coinFields.mint',
  dateRange: 'coinFields.dateRange',
  obverseInscription: 'coinFields.obverseInscription',
  reverseInscription: 'coinFields.reverseInscription',
  obverseDescription: 'coinFields.obverseDescription',
  reverseDescription: 'coinFields.reverseDescription',
  diameterMm: 'coinFields.diameterMm',
  weightGrams: 'coinFields.weightG',
  grade: 'coinFields.grade',
  notes: 'coinFields.notes',
  coin_type: 'coinFields.coinType',
  rarityRating: 'coinFields.rarityRating',
}

/**
 * Turns a coin-field vocabulary key (e.g. `dateRange`, `obverseInscription`,
 * `coin_type`) into a human-readable label (`Date Range`, `Obverse
 * Inscription`, `Coin Type`). Shared by every Deep Analysis surface that
 * renders proposed/hypothesis field names so the formatting stays identical
 * across the proposal editor and the report panel. Known fields use the
 * coinFields catalog so they follow the interface language.
 */
export function formatFieldName(name: string): string {
  const key = FIELD_NAME_KEYS[name]
  if (key) return t(key)
  return name
    .replace(/_/g, ' ')
    .replace(/([A-Z])/g, ' $1')
    .trim()
    .replace(/^./, (c) => c.toUpperCase())
}
