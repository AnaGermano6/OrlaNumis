// The sets a coin or a banknote could start, from what it is: an album of
// the catalogue (its country's €2 commemoratives, its year's eight euro
// coins, Portugal's escudos or réis) and sets that fill themselves with the
// collection's coins (or banknotes) like it (its country's commemoratives,
// its year, its country). They come into each by themselves, this one and
// those to come.
import type { CoinStatus } from '@/api/client'
import type { Coin, CreateCoinSetRequest } from '@/types'
import { t } from '@/i18n'
import { formatFaceValue, releaseCountryName } from '@/utils/coinReleases'
import { countryFlagCode } from '@/utils/countryFlags'
import { emptyFilters, EURO_COUNTRIES, filtersToCriteria, type SimpleSetFilters } from '@/utils/setFilters'
import { parseDenomination } from '@/utils/ucoinNumista'

export interface CoinSetProposal {
  /** Tells the proposals apart: "album", "commemorative", "year", "country". */
  key: string
  /** An album of the catalogue (owned and missing), or a set of the collection's coins. */
  kind: 'album' | 'smart'
  name: string
  /** The set to create, but for its colour. */
  request: Omit<CreateCoinSetRequest, 'color'>
}

/** The euro area's countries: their coins are in the catalogue. */
const EURO_AREA = new Set<string>(EURO_COUNTRIES)
/** The eight coins of a euro country's year. */
const EURO_CIRCULATION_VALUES = [0.01, 0.02, 0.05, 0.1, 0.2, 0.5, 1, 2]
/** At most this many proposals, the album first. */
const MAX_PROPOSALS = 3

/** A denomination's face value in euros ("50 cêntimos" → 0.5); undefined when it says none. */
function euroFaceValue(denomination: string | undefined): number | undefined {
  const parsed = parseDenomination(denomination ?? '')
  if (parsed.value === undefined || parsed.value <= 0) return undefined
  return parsed.subunit ? Math.round(parsed.value) / 100 : parsed.value
}

// The catalogue's album the coin is in, or none.
function albumFor(coin: Coin, status: CoinStatus | null | undefined): CoinSetProposal | null {
  const code = countryFlagCode(coin.country)
  const system = (coin.monetarySystem ?? '').trim().toLowerCase()
  const collector = Boolean(status?.commemorative || status?.nonCirculating)
  if (code && EURO_AREA.has(code) && (system === 'euro' || /\beuros?\b|€/i.test(coin.denomination ?? ''))) {
    const value = euroFaceValue(coin.denomination)
    if (!value) return null
    // A €2 commemorative, or a collector coin: the country's coins of that face value.
    if (collector && (value === 2 || value >= 5 || status?.nonCirculating)) {
      return {
        key: 'album', kind: 'album', name: `${releaseCountryName(code)} · ${formatFaceValue(value)}`,
        request: { name: '', setType: 'checklist', checklist: { source: 'catalogue', country: code, faceValue: value } },
      }
    }
    // A coin of the year's eight: that year's series.
    if (!collector && coin.year && EURO_CIRCULATION_VALUES.includes(value)) {
      return {
        key: 'album', kind: 'album', name: t('sets.challenges.annualSeries.named', { country: releaseCountryName(code), year: coin.year }),
        request: { name: '', setType: 'checklist', checklist: { source: 'challenge', challenge: 'annual-series', country: code, year: coin.year } },
      }
    }
    return null
  }
  if (code !== 'PT') return null
  // Portugal's escudos: the commemoratives' catalogue, or the types in circulation.
  if (system.startsWith('escudo (1911')) {
    return collector
      ? { key: 'album', kind: 'album', name: `${releaseCountryName('PT')} · ${t('sets.catalogue.currencies.PTE')}`, request: { name: '', setType: 'checklist', checklist: { source: 'catalogue', country: 'PT', currency: 'PTE' } } }
      : { key: 'album', kind: 'album', name: t('sets.challenges.escudoTypes.title'), request: { name: '', setType: 'checklist', checklist: { source: 'challenge', challenge: 'escudo-types' } } }
  }
  if (system.startsWith('real')) {
    return { key: 'album', kind: 'album', name: t('sets.challenges.reisTypes.title'), request: { name: '', setType: 'checklist', checklist: { source: 'challenge', challenge: 'reis-types' } } }
  }
  return null
}

// The same smart criteria the sets' own filters make, so the set can be changed there.
function smart(key: string, name: string, filters: Partial<SimpleSetFilters>): CoinSetProposal {
  const criteria = filtersToCriteria({ ...emptyFilters(), ...filters }) as unknown as Record<string, unknown>
  return { key, kind: 'smart', name, request: { name: '', setType: 'smart', smartCriteria: criteria } }
}

/**
 * The sets a coin or banknote of the collection could start, the album first
 * (a coin's); none for one of the wishlist (the sets fill with the
 * collection's). Coins and banknotes each have sets of their own.
 */
export function coinSetProposals(coin: Coin, status?: CoinStatus | null): CoinSetProposal[] {
  if (coin.isWishlist) return []
  const kind = coin.kind === 'banknote' ? 'banknote' : 'coin'
  const proposals: CoinSetProposal[] = []
  const album = kind === 'coin' ? albumFor(coin, status) : null
  if (album) proposals.push(album)
  const country = coin.country?.trim()
  if (kind === 'coin' && country && status?.commemorative) {
    proposals.push(smart('commemorative', t('coinDetail.sets.proposals.commemorativeOf', { country }), { country, classification: 'commemorative', kind }))
  }
  if (coin.year) {
    const name = kind === 'coin' ? t('coinDetail.sets.proposals.coinsOfYear', { year: coin.year }) : t('coinDetail.sets.proposals.banknotesOfYear', { year: coin.year })
    proposals.push(smart('year', name, { yearFrom: coin.year, yearTo: coin.year, kind }))
  }
  if (country) {
    const name = kind === 'coin' ? t('coinDetail.sets.proposals.coinsOf', { country }) : t('coinDetail.sets.proposals.banknotesOf', { country })
    proposals.push(smart('country', name, { country, kind }))
  }
  return proposals.slice(0, MAX_PROPOSALS).map((proposal) => ({ ...proposal, request: { ...proposal.request, name: proposal.name } }))
}
