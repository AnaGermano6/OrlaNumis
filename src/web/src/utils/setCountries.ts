// The checklist sets' slots country by country: for each country, how many
// places its sets have, how many the collection fills and how many are
// missing, set by set. A coin filling two sets' slots counts in each.
import { intlLocale } from '@/i18n'
import type { CoinSetSummary } from '@/types'
import { releaseCountryName } from '@/utils/coinReleases'

export interface CountrySetCompletion {
  id: number
  name: string
  color: string
  owned: number
  total: number
  partial: boolean
}

export interface CountryCompletion {
  /** ISO code. */
  country: string
  owned: number
  total: number
  missing: number
  /** True when a set's list behind it is partly known. */
  partial: boolean
  sets: CountrySetCompletion[]
}

/** The countries of the sets' slots, those with most places first, then by name. */
export function setsByCountry(sets: CoinSetSummary[]): CountryCompletion[] {
  const byCountry = new Map<string, CountryCompletion>()
  for (const set of sets) {
    for (const count of set.countries ?? []) {
      if (!count.country || count.total <= 0) continue
      let entry = byCountry.get(count.country)
      if (!entry) {
        entry = { country: count.country, owned: 0, total: 0, missing: 0, partial: false, sets: [] }
        byCountry.set(count.country, entry)
      }
      entry.owned += count.owned
      entry.total += count.total
      entry.missing += count.total - count.owned
      entry.partial ||= Boolean(set.partial)
      entry.sets.push({ id: set.id, name: set.name, color: set.color, owned: count.owned, total: count.total, partial: Boolean(set.partial) })
    }
  }
  const collator = new Intl.Collator(intlLocale())
  return [...byCountry.values()].sort((a, b) => b.total - a.total || collator.compare(releaseCountryName(a.country), releaseCountryName(b.country)))
}
