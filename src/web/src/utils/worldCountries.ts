// Matches the collection's free-text countries (Numista issuer names, in
// English or Portuguese, or typed by hand) to the world map's countries.
// Issuers that are not modern countries ("Corinth", "Roman Empire") stay
// unmatched and are listed beside the map.
import type { Feature, FeatureCollection, MultiPolygon, Polygon } from 'geojson'
import type { CoinCountryCount } from '@/api/client'

export interface WorldCountryProperties {
  /** ISO 3166-1 alpha-2 code (alpha-3 for the few without one). */
  iso: string
  name: string
  namePt: string
  aliases: string[]
}

export type WorldCountryFeature = Feature<Polygon | MultiPolygon, WorldCountryProperties>
export type WorldCountryCollection = FeatureCollection<Polygon | MultiPolygon, WorldCountryProperties>

let pending: Promise<WorldCountryCollection> | null = null

/** Country outlines (Natural Earth, public domain), loaded on first use as a separate chunk. */
export function loadWorldCountries(): Promise<WorldCountryCollection> {
  pending ??= import('@/assets/geo/world-countries.json')
    .then((module) => module.default as unknown as WorldCountryCollection)
    .catch((error: unknown) => {
      pending = null
      throw error
    })
  return pending
}

/** "  República   Checa " → "republica checa": accents, case, punctuation and spacing ignored. */
export function foldCountryName(value: string): string {
  return value
    .toLowerCase()
    .normalize('NFD')
    .replace(/[̀-ͯ]/g, '')
    .replace(/[^a-z0-9]+/g, ' ')
    .trim()
}

/** Folded name → ISO code, from the map's own names plus the browser's names in English and Portuguese. */
export function buildCountryIndex(collection: WorldCountryCollection): Map<string, string> {
  const index = new Map<string, string>()
  const displayNames = ['en', 'pt-PT', 'pt'].flatMap((locale) => {
    try {
      return [new Intl.DisplayNames([locale], { type: 'region' })]
    } catch {
      return []
    }
  })
  for (const feature of collection.features) {
    const { iso, aliases } = feature.properties
    const names = [...aliases]
    if (iso.length === 2) {
      for (const display of displayNames) {
        const name = display.of(iso)
        if (name && name !== iso) names.push(name)
      }
    }
    for (const name of names) {
      const folded = foldCountryName(name)
      if (folded && !index.has(folded)) index.set(folded, iso)
    }
  }
  return index
}

export interface MapCountry {
  iso: string
  count: number
  /** Stored country names that map here, most coins first; the first one is used to filter. */
  names: CoinCountryCount[]
}

export interface CountryDistribution {
  byIso: Map<string, MapCountry>
  /** Countries with no place on the map (historical issuers, typos), most coins first. */
  unmatched: CoinCountryCount[]
}

export function distributeCountries(counts: CoinCountryCount[], index: Map<string, string>): CountryDistribution {
  const byIso = new Map<string, MapCountry>()
  const unmatched: CoinCountryCount[] = []
  for (const entry of counts) {
    const iso = index.get(foldCountryName(entry.country))
    if (!iso) {
      unmatched.push(entry)
      continue
    }
    const country = byIso.get(iso) ?? { iso, count: 0, names: [] }
    country.count += entry.count
    country.names.push(entry)
    byIso.set(iso, country)
  }
  for (const country of byIso.values()) country.names.sort((a, b) => b.count - a.count)
  unmatched.sort((a, b) => b.count - a.count)
  return { byIso, unmatched }
}

/** Colour steps for the map: 0 = no coins, then 1, 2–4, 5–9, 10+. */
export function countryShade(count: number): 0 | 1 | 2 | 3 | 4 {
  if (count <= 0) return 0
  if (count === 1) return 1
  if (count < 5) return 2
  if (count < 10) return 3
  return 4
}
