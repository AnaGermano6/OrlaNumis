// Simple filters for a set: country, face value, years, kind of coin and
// presentation. They are kept as the smart criteria the server matches
// coins with, so a set made of them fills by itself.
import type { SmartCriteriaGroup, SmartCriteriaRule } from '@/types'
import { t } from '@/i18n'
import { presentationLabel } from '@/utils/coinAttributes'

/** The kinds of coin the server tells apart (see its coin classifier). */
export const COIN_CLASSIFICATIONS = ['circulation', 'commemorative', 'collector'] as const
export type CoinClassification = (typeof COIN_CLASSIFICATIONS)[number]

export interface SimpleSetFilters {
  /** As the collection writes it ("Portugal"); '' for any. */
  country: string
  /** As the collection writes it ("2 euro"); '' for any. */
  denomination: string
  yearFrom: number | null
  yearTo: number | null
  classification: CoinClassification | ''
  /** One of COIN_PRESENTATIONS; '' for any. */
  presentation: string
}

export function emptyFilters(): SimpleSetFilters {
  return { country: '', denomination: '', yearFrom: null, yearTo: null, classification: '', presentation: '' }
}

// Only coins of the collection: not wanted, not sold.
const COLLECTION_RULES: SmartCriteriaRule[] = [
  { field: 'isWishlist', op: 'eq', value: false },
  { field: 'isSold', op: 'eq', value: false },
]

/** The smart criteria of the filters: all of them, and the collection's coins only. */
export function filtersToCriteria(filters: SimpleSetFilters): SmartCriteriaGroup {
  const rules: SmartCriteriaRule[] = []
  if (filters.country) rules.push({ field: 'country', op: 'eq', value: filters.country })
  if (filters.denomination) rules.push({ field: 'denomination', op: 'eq', value: filters.denomination })
  const from = filters.yearFrom || null
  const to = filters.yearTo || null
  if (from && to) rules.push(from === to ? { field: 'year', op: 'eq', value: from } : { field: 'year', op: 'between', value: [Math.min(from, to), Math.max(from, to)] })
  else if (from) rules.push({ field: 'year', op: 'gte', value: from })
  else if (to) rules.push({ field: 'year', op: 'lte', value: to })
  if (filters.classification) rules.push({ field: 'classification', op: 'eq', value: filters.classification })
  if (filters.presentation) rules.push({ field: 'presentation', op: 'eq', value: filters.presentation })
  return { operator: 'and', rules: [...rules, ...COLLECTION_RULES] }
}

/** The filters criteria were made of; null when they are more than simple filters can say. */
export function criteriaToFilters(criteria: SmartCriteriaGroup | null | undefined): SimpleSetFilters | null {
  if (!criteria || criteria.operator !== 'and' || !Array.isArray(criteria.rules)) return null
  const filters = emptyFilters()
  for (const node of criteria.rules) {
    if (!('field' in node)) return null
    const { field, op, value } = node
    if ((field === 'isWishlist' || field === 'isSold') && op === 'eq' && value === false) continue
    if (field === 'country' && op === 'eq' && typeof value === 'string' && !filters.country) filters.country = value
    else if (field === 'denomination' && op === 'eq' && typeof value === 'string' && !filters.denomination) filters.denomination = value
    else if (field === 'year' && op === 'eq' && typeof value === 'number') filters.yearFrom = filters.yearTo = value
    else if (field === 'year' && op === 'gte' && typeof value === 'number') filters.yearFrom = value
    else if (field === 'year' && op === 'lte' && typeof value === 'number') filters.yearTo = value
    else if (field === 'year' && op === 'between' && Array.isArray(value) && value.length === 2) {
      filters.yearFrom = Number(value[0])
      filters.yearTo = Number(value[1])
    }
    else if (field === 'classification' && op === 'eq' && COIN_CLASSIFICATIONS.includes(value as CoinClassification)) filters.classification = value as CoinClassification
    else if (field === 'presentation' && op === 'eq' && typeof value === 'string') filters.presentation = value
    else return null
  }
  return filters
}

export function classificationLabel(value: CoinClassification): string {
  return t(`sets.filters.classifications.${value}`)
}

function years(filters: SimpleSetFilters): string {
  const { yearFrom: from, yearTo: to } = filters
  if (from && to) return from === to ? String(from) : `${Math.min(from, to)}–${Math.max(from, to)}`
  if (from) return t('sets.filters.since', { year: from })
  if (to) return t('sets.filters.until', { year: to })
  return ''
}

/** A name for a set of the filters: "Portugal · 2 euro · 2002–2010 · Comemorativas". */
export function filtersName(filters: SimpleSetFilters): string {
  const parts = [
    filters.country,
    filters.denomination,
    years(filters),
    filters.classification ? classificationLabel(filters.classification) : '',
    filters.presentation ? presentationLabel(filters.presentation) : '',
  ].filter(Boolean)
  return parts.length ? parts.join(' · ') : t('sets.filters.allCoins')
}

/** The countries of the euro, as the catalogue names them (ISO codes). */
export const EURO_COUNTRIES = [
  'AD', 'AT', 'BE', 'BG', 'CY', 'DE', 'EE', 'ES', 'FI', 'FR', 'GR', 'HR', 'IE',
  'IT', 'LT', 'LU', 'LV', 'MC', 'MT', 'NL', 'PT', 'SI', 'SK', 'SM', 'VA',
] as const

