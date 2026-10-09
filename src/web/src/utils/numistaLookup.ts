import type {
  Coin,
  NumistaCandidate,
  NumistaCacheMetadata,
  NumistaEvidence,
  NumistaLookupStatus,
  SelectedNumistaReference,
} from '@/types'
import { t } from '@/i18n'

export type NumistaPanelState = 'idle' | 'loading' | NumistaLookupStatus

type DirectCoinEvidence = Pick<
  Coin,
  'name' | 'ruler' | 'denomination' | 'mint' | 'dateRange' | 'material' | 'obverseInscription' | 'reverseInscription'
>

export interface NumistaStatusGuidance {
  state: NumistaPanelState
  label: string
  title: string
  message: string
  canRetry: boolean
  settingsHref?: string
}

function bounded(value: string | null | undefined, limit: number): string | undefined {
  const trimmed = value?.trim()
  return trimmed ? trimmed.slice(0, limit) : undefined
}

export function buildDirectNumistaEvidence(coin: DirectCoinEvidence): NumistaEvidence {
  return {
    title: bounded(coin.name, 200),
    issuer: bounded(coin.ruler, 200),
    denomination: bounded(coin.denomination, 100),
    mint: bounded(coin.mint, 200),
    dateText: bounded(coin.dateRange, 100),
    material: bounded(coin.material, 100),
    obverseInscription: bounded(coin.obverseInscription, 500),
    reverseInscription: bounded(coin.reverseInscription, 500),
  }
}

export function numistaCandidateIdentity(candidate: Pick<NumistaCandidate, 'id'>): string {
  return `numista:${candidate.id}`
}

export function selectedNumistaReferenceFromCandidate(
  candidate: NumistaCandidate,
): SelectedNumistaReference {
  return {
    catalog: 'Numista',
    number: String(candidate.id),
    uri: candidate.canonicalUrl,
  }
}

export function numistaCandidateFromReference(
  reference: SelectedNumistaReference | null | undefined,
): NumistaCandidate | null {
  if (!reference) return null
  const id = Number(reference.number)
  if (!Number.isSafeInteger(id) || id <= 0) return null

  return {
    id,
    canonicalUrl: reference.uri,
    title: `Numista #${reference.number}`,
    providerPosition: 0,
    enrichmentState: 'not_requested',
    assessment: {
      scoringVersion: 'numista-v1',
      score: 50,
      band: 'weak',
      reasons: [],
    },
  }
}

export function retainNumistaSelection(
  selected: NumistaCandidate | null,
  candidates: NumistaCandidate[],
): NumistaCandidate | null {
  if (!selected) return null
  return candidates.find(candidate => candidate.id === selected.id) ?? selected
}

export function isSelectionOutsideResults(
  selected: NumistaCandidate | null,
  candidates: NumistaCandidate[],
): boolean {
  return selected !== null && !candidates.some(candidate => candidate.id === selected.id)
}

export function getNumistaStatusGuidance(
  state: NumistaPanelState,
  isAdmin: boolean,
  retryAfterSeconds?: number,
): NumistaStatusGuidance {
  switch (state) {
    case 'idle':
      return {
        state,
        label: t('numista.status.idle.label'),
        title: t('numista.status.idle.title'),
        message: t('numista.status.idle.message'),
        canRetry: false,
      }
    case 'loading':
      return {
        state,
        label: t('numista.status.loading.label'),
        title: t('numista.status.loading.title'),
        message: t('numista.status.loading.message'),
        canRetry: false,
      }
    case 'success':
      return {
        state,
        label: t('numista.status.success.label'),
        title: t('numista.status.success.title'),
        message: t('numista.status.success.message'),
        canRetry: false,
      }
    case 'empty':
      return {
        state,
        label: t('numista.status.empty.label'),
        title: t('numista.status.empty.title'),
        message: t('numista.status.empty.message'),
        canRetry: true,
      }
    case 'unconfigured':
      return {
        state,
        label: t('numista.status.unconfigured.label'),
        title: t('numista.status.unconfigured.title'),
        message: isAdmin
          ? t('numista.status.unconfigured.adminMessage')
          : t('numista.status.unconfigured.userMessage'),
        canRetry: isAdmin,
        settingsHref: isAdmin ? '/admin?tab=system' : undefined,
      }
    case 'quota-limited':
      return {
        state,
        label: t('numista.status.quota.label'),
        title: t('numista.status.quota.title'),
        message: retryAfterSeconds
          ? t('numista.status.quota.wait', { duration: formatDuration(retryAfterSeconds) })
          : t('numista.status.quota.waitGeneric'),
        canRetry: true,
      }
    case 'timeout':
      return {
        state,
        label: t('numista.status.timeout.label'),
        title: t('numista.status.timeout.title'),
        message: t('numista.status.timeout.message'),
        canRetry: true,
      }
    case 'unavailable':
      return {
        state,
        label: t('numista.status.unavailable.label'),
        title: t('numista.status.unavailable.title'),
        message: t('numista.status.unavailable.message'),
        canRetry: true,
      }
  }
}

export function getNumistaCacheFreshnessText(
  status: NumistaLookupStatus | null,
  cache: NumistaCacheMetadata | null | undefined,
): string | null {
  if (!cache || (status !== 'success' && status !== 'empty')) return null
  if (cache.coalesced) return t('numista.cache.shared')
  if (!cache.hit) return t('numista.cache.fresh')
  return t('numista.cache.cached', { age: formatAge(cache.ageSeconds) })
}

function formatAge(seconds: number): string {
  if (seconds < 60) return t('numista.age.seconds', { count: seconds })
  if (seconds < 3600) return t('numista.age.minutes', { count: Math.floor(seconds / 60) })
  if (seconds < 86400) return t('numista.age.hours', { count: Math.floor(seconds / 3600) })
  return t('numista.age.days', { count: Math.floor(seconds / 86400) })
}

function formatDuration(seconds: number): string {
  if (seconds < 60) return t('numista.duration.seconds', { count: seconds })
  if (seconds % 60 === 0 && seconds < 3600) {
    return t('numista.duration.minutes', { count: seconds / 60 })
  }
  return t('numista.duration.seconds', { count: seconds })
}
