import type { Coin } from '@/types'
import { intlLocale, t, type MessageKey } from '@/i18n'
import { periodYears } from '@/utils/coinPeriods'
import { formatNumistaYear } from '@/utils/numistaCatalog'

export const COIN_PRESENTATIONS = ['loose', 'coincard', 'capsule', 'flip', 'proof_set', 'blister', 'slab'] as const
export type CoinPresentation = (typeof COIN_PRESENTATIONS)[number]

/** From most to least common. */
export const COIN_RARITIES = ['common', 'uncommon', 'scarce', 'rare', 'very_rare', 'extremely_rare'] as const
export type CoinRarity = (typeof COIN_RARITIES)[number]

/** How many of a coin or note the collector has; the API accepts 1 to 9999. */
export const QUANTITY_MIN = 1
export const QUANTITY_MAX = 9999

/** The quantity to save: a whole number in range, else 1. */
export function validQuantity(value: number | null | undefined): number {
  return typeof value === 'number' && Number.isInteger(value) && value >= QUANTITY_MIN && value <= QUANTITY_MAX ? value : QUANTITY_MIN
}

/** Set by hand: whether the coin is still legal tender and whether it is commemorative. */
export const COIN_CIRCULATIONS = [
  'current', 'current_commemorative',
  'demonetized', 'demonetized_commemorative',
  'non_circulating', 'non_circulating_commemorative',
] as const
export type CoinCirculation = (typeof COIN_CIRCULATIONS)[number]

const CIRCULATION_KEYS: Record<CoinCirculation, MessageKey> = {
  current: 'coinAttributes.circulation.current',
  current_commemorative: 'coinAttributes.circulation.current_commemorative',
  demonetized: 'coinAttributes.circulation.demonetized',
  demonetized_commemorative: 'coinAttributes.circulation.demonetized_commemorative',
  non_circulating: 'coinAttributes.circulation.non_circulating',
  non_circulating_commemorative: 'coinAttributes.circulation.non_circulating_commemorative',
}

export function circulationLabel(value: string | null | undefined): string {
  if (!value) return ''
  return value in CIRCULATION_KEYS ? t(CIRCULATION_KEYS[value as CoinCirculation]) : value
}

const PRESENTATION_KEYS: Record<CoinPresentation, MessageKey> = {
  loose: 'coinAttributes.presentation.loose',
  coincard: 'coinAttributes.presentation.coincard',
  capsule: 'coinAttributes.presentation.capsule',
  flip: 'coinAttributes.presentation.flip',
  proof_set: 'coinAttributes.presentation.proof_set',
  blister: 'coinAttributes.presentation.blister',
  slab: 'coinAttributes.presentation.slab',
}

const RARITY_KEYS: Record<CoinRarity, MessageKey> = {
  common: 'coinAttributes.rarity.common',
  uncommon: 'coinAttributes.rarity.uncommon',
  scarce: 'coinAttributes.rarity.scarce',
  rare: 'coinAttributes.rarity.rare',
  very_rare: 'coinAttributes.rarity.very_rare',
  extremely_rare: 'coinAttributes.rarity.extremely_rare',
}

export function presentationLabel(value: string | null | undefined): string {
  if (!value) return ''
  return value in PRESENTATION_KEYS ? t(PRESENTATION_KEYS[value as CoinPresentation]) : value
}

/** A coin's conservation grades, the best first, as Numista and dealers write them. */
export const COIN_GRADES = ['UNC', 'AU', 'XF', 'VF', 'F', 'VG', 'G'] as const
/** Collector finishes, often written where the grade goes. */
export const COIN_FINISHES = ['BNC', 'PRF'] as const

/** "VF — Muito bem conservada (MBC)", or a grade the app does not know as written. */
export function coinGradeLabel(grade: string): string {
  const known = [...COIN_GRADES, ...COIN_FINISHES].find((option) => option === grade)
  return known ? `${known} — ${t(`coinForm.grades.${known}`)}` : grade
}

// The usual kinds of coin or banknote are named in the app's language;
// others as Numista names them.
const CLASSIFICATION_KEYS: Record<string, MessageKey> = {
  standard: 'collection.view.classificationStandard',
  'circulating-commemorative': 'collection.view.classificationCirculatingCommemorative',
  'non-circulating': 'collection.view.classificationNonCirculating',
}

/** "Circulantes comuns", or Numista's name of a kind the app does not know. */
export function collectionClassificationLabel(option: { key: string; label: string }): string {
  const key = CLASSIFICATION_KEYS[option.key]
  return key ? t(key) : option.label
}

export function rarityLabel(value: string | null | undefined): string {
  if (!value) return ''
  return value in RARITY_KEYS ? t(RARITY_KEYS[value as CoinRarity]) : value
}

/**
 * Suggests a rarity from a mintage. Only a starting point: survival, demand
 * and varieties matter too, so the user can always change it.
 */
export function rarityFromMintage(mintage: number | null | undefined): CoinRarity | undefined {
  if (!mintage || mintage <= 0) return undefined
  if (mintage >= 10_000_000) return 'common'
  if (mintage >= 1_000_000) return 'uncommon'
  if (mintage >= 100_000) return 'scarce'
  if (mintage >= 10_000) return 'rare'
  if (mintage >= 1_000) return 'very_rare'
  return 'extremely_rare'
}

/** A monetary system without its dates ("Escudo (1911-2001)" → "Escudo"), for short labels. */
export function monetarySystemShortName(system: string | null | undefined): string {
  const value = (system ?? '').trim()
  const cut = value.indexOf(' (')
  return cut > 0 ? value.slice(0, cut) : value
}

export type CoinGroupBy = 'none' | 'country' | 'period' | 'monetarySystem' | 'year' | 'rarity' | 'presentation'
export const COIN_GROUPINGS: CoinGroupBy[] = ['none', 'country', 'period', 'monetarySystem', 'year', 'rarity', 'presentation']

export interface CoinGroup {
  key: string
  label: string
  coins: Coin[]
  /** The country a group by period belongs to, for its flag. */
  country?: string
}

interface GroupKey {
  key: string
  label: string
  /** Sort position; groups without a value always go last. */
  order: number | string
  missing: boolean
}

function groupKey(coin: Coin, by: Exclude<CoinGroupBy, 'none' | 'period'>): GroupKey {
  switch (by) {
    case 'country': {
      const country = coin.country?.trim() ?? ''
      return country
        ? { key: country, label: country, order: country, missing: false }
        : { key: '', label: t('collection.groups.noCountry'), order: '', missing: true }
    }
    case 'monetarySystem': {
      const system = coin.monetarySystem?.trim() ?? ''
      return system
        ? { key: system, label: system, order: system, missing: false }
        : { key: '', label: t('collection.groups.noMonetarySystem'), order: '', missing: true }
    }
    case 'year': {
      const year = coin.year
      return year !== null && year !== undefined
        ? { key: String(year), label: formatNumistaYear(year), order: year, missing: false }
        : { key: '', label: t('collection.groups.noYear'), order: 0, missing: true }
    }
    case 'rarity': {
      const index = COIN_RARITIES.indexOf(coin.rarity as CoinRarity)
      return index >= 0
        ? { key: coin.rarity!, label: rarityLabel(coin.rarity), order: index, missing: false }
        : { key: '', label: t('collection.groups.noRarity'), order: 0, missing: true }
    }
    case 'presentation': {
      const index = COIN_PRESENTATIONS.indexOf(coin.presentation as CoinPresentation)
      return index >= 0
        ? { key: coin.presentation!, label: presentationLabel(coin.presentation), order: index, missing: false }
        : { key: '', label: t('collection.groups.noPresentation'), order: 0, missing: true }
    }
  }
}

/**
 * Coins by country and, within it, by period, most recent first, as uCoin
 * lists a country's periods: "Portugal · República Portuguesa 1986–2001". A
 * period whose name gives no years shows those of its coins; coins without
 * a period come last in their country.
 */
function groupCoinsByPeriod(coins: Coin[]): CoinGroup[] {
  const groups = new Map<string, { country: string; period: string; coins: Coin[] }>()
  for (const coin of coins) {
    const country = coin.country?.trim() ?? ''
    const period = coin.period?.trim() ?? ''
    const key = `${country}\n${period}`
    const group = groups.get(key) ?? { country, period, coins: [] }
    group.coins.push(coin)
    groups.set(key, group)
  }
  const collator = new Intl.Collator(intlLocale(), { sensitivity: 'base', numeric: true })
  const sections = [...groups.entries()].map(([key, group]) => {
    const coinYears = group.coins.map((coin) => coin.year).filter((year): year is number => typeof year === 'number')
    const named = periodYears(group.period)
    const from = named?.from ?? (coinYears.length ? Math.min(...coinYears) : undefined)
    const to = named?.to ?? (coinYears.length ? Math.max(...coinYears) : undefined)
    const years = named || from === undefined ? '' : from === to ? ` ${formatNumistaYear(from)}` : ` ${formatNumistaYear(from)}–${formatNumistaYear(to!)}`
    const period = group.period ? `${group.period}${years}` : t('collection.groups.noPeriod')
    const label = group.country ? `${group.country} · ${period}` : period
    return { key, label, country: group.country, coins: group.coins, missing: !group.period, start: from ?? 0 }
  })
  return sections
    .sort((a, b) => {
      if (!a.country !== !b.country) return a.country ? -1 : 1
      const byCountry = collator.compare(a.country, b.country)
      if (byCountry) return byCountry
      if (a.missing !== b.missing) return a.missing ? 1 : -1
      return b.start - a.start || collator.compare(a.label, b.label)
    })
    .map(({ key, label, country, coins: sectionCoins }) => ({ key, label, coins: sectionCoins, ...(country ? { country } : {}) }))
}

/** Splits coins into labelled sections, keeping each section's coins in their given order. */
export function groupCoins(coins: Coin[], by: CoinGroupBy): CoinGroup[] {
  if (by === 'none') return [{ key: '', label: '', coins }]
  if (by === 'period') return groupCoinsByPeriod(coins)
  const groups = new Map<string, GroupKey & { coins: Coin[] }>()
  for (const coin of coins) {
    const key = groupKey(coin, by)
    const group = groups.get(key.key) ?? { ...key, coins: [] }
    group.coins.push(coin)
    groups.set(key.key, group)
  }
  const collator = new Intl.Collator(intlLocale(), { sensitivity: 'base', numeric: true })
  return [...groups.values()]
    .sort((a, b) => {
      if (a.missing !== b.missing) return a.missing ? 1 : -1
      if (typeof a.order === 'number' && typeof b.order === 'number') return a.order - b.order
      return collator.compare(String(a.order), String(b.order))
    })
    .map(({ key, label, coins: groupCoinsList }) => ({ key, label, coins: groupCoinsList }))
}

const YEAR_NUMBER = /\d{1,4}/
const YEAR_BC = /\b(B\.?\s?C\.?(E\.?)?|a\.\s?C\.?)(\s|$|[-–—,;)])/i
const YEAR_AD = /\b(A\.?\s?D\.?|C\.?E\.?|d\.\s?C\.?)(\s|$|[-–—,;)])|\bAD\b/i
const YEAR_CENTURY = /century|séc|sec\.|século|seculo|\d(st|nd|rd|th)\b/i

/**
 * The year the server will read from a date range when no year is typed
 * (same rules as models.DeriveCoinYear): the first number, negative when a BC
 * marker follows it with no AD marker in between. Centuries give undefined.
 */
export function deriveCoinYear(dateRange: string | null | undefined): number | undefined {
  const text = dateRange ?? ''
  if (YEAR_CENTURY.test(text)) return undefined
  const match = YEAR_NUMBER.exec(text)
  if (!match) return undefined
  const year = Number(match[0])
  if (!year) return undefined
  const rest = text.slice(match.index + match[0].length)
  const bc = YEAR_BC.exec(rest)
  return bc && !YEAR_AD.test(rest.slice(0, bc.index)) ? -year : year
}
