import type { Coin } from '@/types'
import { intlLocale, t, type MessageKey } from '@/i18n'
import { formatNumistaYear } from '@/utils/numistaCatalog'

export const COIN_PRESENTATIONS = ['loose', 'coincard', 'capsule', 'flip', 'proof_set', 'blister', 'slab'] as const
export type CoinPresentation = (typeof COIN_PRESENTATIONS)[number]

/** From most to least common. */
export const COIN_RARITIES = ['common', 'uncommon', 'scarce', 'rare', 'very_rare', 'extremely_rare'] as const
export type CoinRarity = (typeof COIN_RARITIES)[number]

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

export type CoinGroupBy = 'none' | 'country' | 'year' | 'rarity' | 'presentation'
export const COIN_GROUPINGS: CoinGroupBy[] = ['none', 'country', 'year', 'rarity', 'presentation']

export interface CoinGroup {
  key: string
  label: string
  coins: Coin[]
}

interface GroupKey {
  key: string
  label: string
  /** Sort position; groups without a value always go last. */
  order: number | string
  missing: boolean
}

function groupKey(coin: Coin, by: Exclude<CoinGroupBy, 'none'>): GroupKey {
  switch (by) {
    case 'country': {
      const country = coin.country?.trim() ?? ''
      return country
        ? { key: country, label: country, order: country, missing: false }
        : { key: '', label: t('collection.groups.noCountry'), order: '', missing: true }
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

/** Splits coins into labelled sections, keeping each section's coins in their given order. */
export function groupCoins(coins: Coin[], by: CoinGroupBy): CoinGroup[] {
  if (by === 'none') return [{ key: '', label: '', coins }]
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
