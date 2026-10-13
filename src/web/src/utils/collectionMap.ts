// What the collection map shows, worked out from the collection's coins:
// which coins pass the map's filters, the coins of each place (a country
// outline, a small territory drawn as a point, or a name off the map), the
// colour of each place by the chosen measure, the places still missing from a
// target (the euro, Europe, the world), totals by continent, the territory of
// extinct countries and the coins of each mint.
import type { Coin } from '@/types'
import {
  CONTINENTS,
  EURO_ISSUERS,
  SMALL_TERRITORIES,
  continentOf,
  continentPlaces,
  extinctTerritory,
  type Continent,
  type MapPoint,
} from '@/utils/mapGeography'
import { findMintPlace, type MintPlace } from '@/utils/modernMints'
import { EXTRA_COUNTRY_ALIASES, foldCountryName, lookupCountry, type WorldCountryCollection } from '@/utils/worldCountries'

export type MapMetric = 'coins' | 'value' | 'types'
export type MapTarget = '' | 'euro' | 'europe' | 'world'

export interface MapFilters {
  era: string
  yearFrom: number | null
  yearTo: number | null
  circulation: '' | 'current' | 'demonetized' | 'nonCirculating'
  commemorative: '' | 'true' | 'false'
}

export const EMPTY_MAP_FILTERS: MapFilters = { era: '', yearFrom: null, yearTo: null, circulation: '', commemorative: '' }

/** Coin IDs by legal tender and commemorative status (see GET /coins/statuses). */
export interface CoinStatusSets {
  current: Set<number>
  demonetized: Set<number>
  nonCirculating: Set<number>
  commemorative: Set<number>
  regular: Set<number>
}

export function hasMapFilters(filters: MapFilters): boolean {
  return Boolean(filters.era || filters.yearFrom !== null || filters.yearTo !== null || filters.circulation || filters.commemorative)
}

/** The coins that pass the map's filters. A coin with no year fails a year range. */
export function filterMapCoins(coins: Coin[], filters: MapFilters, statuses: CoinStatusSets): Coin[] {
  return coins.filter((coin) => {
    if (filters.era && coin.era !== filters.era) return false
    if (filters.yearFrom !== null && (coin.year == null || coin.year < filters.yearFrom)) return false
    if (filters.yearTo !== null && (coin.year == null || coin.year > filters.yearTo)) return false
    if (filters.circulation && !statuses[filters.circulation].has(coin.id)) return false
    if (filters.commemorative === 'true' && !statuses.commemorative.has(coin.id)) return false
    if (filters.commemorative === 'false' && !statuses.regular.has(coin.id)) return false
    return true
  })
}

// ---- Places ----------------------------------------------------------------

export type PlaceKind = 'outline' | 'point' | 'offMap'

/** Where a key is on the map, and its name in each language. */
export interface MapPlaceInfo {
  key: string
  kind: PlaceKind
  name: string
  namePt: string
}

export interface MapPlaces {
  /** Folded country name → place key. */
  index: Map<string, string>
  byKey: Map<string, MapPlaceInfo>
}

/** Every place drawn on the map (outlines and points), and an index from the names the collection may use. */
export function buildMapPlaces(collection: WorldCountryCollection | null, points: MapPoint[] = SMALL_TERRITORIES): MapPlaces {
  const index = new Map<string, string>()
  const byKey = new Map<string, MapPlaceInfo>()
  const displayNames = ['en', 'pt-PT', 'pt'].flatMap((language) => {
    try {
      return [new Intl.DisplayNames([language], { type: 'region' })]
    } catch {
      return []
    }
  })
  function add(key: string, kind: PlaceKind, name: string, namePt: string, aliases: string[]) {
    byKey.set(key, { key, kind, name, namePt })
    const names = [...aliases]
    if (/^[A-Z]{2}$/.test(key)) {
      for (const display of displayNames) {
        const shown = display.of(key)
        if (shown && shown !== key) names.push(shown)
      }
    }
    for (const entry of names) {
      const folded = foldCountryName(entry)
      if (folded && !index.has(folded)) index.set(folded, key)
    }
  }
  // Points first: "Madeira" is the islands, not a name of Portugal.
  for (const point of points) add(point.iso, 'point', point.name, point.namePt, point.aliases)
  for (const feature of collection?.features ?? []) {
    const { iso, name, namePt, aliases } = feature.properties
    if (!byKey.has(iso)) add(iso, 'outline', name, namePt, [name, namePt, ...aliases, ...(EXTRA_COUNTRY_ALIASES[iso] ?? [])])
  }
  return { index, byKey }
}

/** A place's name in the app's language. */
export function placeLabel(place: Pick<MapPlaceInfo, 'name' | 'namePt'>, language: string): string {
  return language === 'pt-PT' ? place.namePt : place.name
}

export interface PlaceStats {
  /** The place's ISO (or region) code, or "name:" and the stored name of one off the map. */
  key: string
  kind: PlaceKind
  label: string
  /** Stored country names counted here, most coins first. */
  names: string[]
  coins: Coin[]
  count: number
  /** Sum of the coins' current values, in the collection currency (0 when none has one). */
  value: number
  /** Distinct types: the same Numista type, or the same name when not linked, counts once. */
  types: number
  minYear?: number
  maxYear?: number
  continent?: Continent
  /** Present-day places of a country that no longer exists. */
  territory: string[]
}

function numistaType(coin: Coin): string | undefined {
  return coin.references?.find((reference) => reference.catalog.trim().toLowerCase() === 'numista')?.number.trim() || undefined
}

/** The place key of a stored country name: its place on the map, or "name:" and the name. */
export function placeKeyOf(country: string, places: MapPlaces): string {
  return lookupCountry(places.index, country) ?? `name:${country.trim()}`
}

/** The collection's coins by place, and how many have no country. */
export function groupCoinsByPlace(coins: Coin[], places: MapPlaces, language: string): { byKey: Map<string, PlaceStats>; withoutCountry: number } {
  const byKey = new Map<string, PlaceStats>()
  const nameCounts = new Map<string, Map<string, number>>()
  let withoutCountry = 0
  for (const coin of coins) {
    const country = coin.country?.trim() ?? ''
    if (!country) {
      withoutCountry += 1
      continue
    }
    const key = placeKeyOf(country, places)
    let stats = byKey.get(key)
    if (!stats) {
      const info = places.byKey.get(key)
      const territory = info ? [] : extinctTerritory(country)
      stats = {
        key,
        kind: info?.kind ?? 'offMap',
        label: info ? placeLabel(info, language) : country,
        names: [],
        coins: [],
        count: 0,
        value: 0,
        types: 0,
        continent: continentOf(info ? key : territory[0] ?? ''),
        territory,
      }
      byKey.set(key, stats)
    }
    stats.coins.push(coin)
    stats.count += 1
    stats.value += coin.currentValue ?? 0
    if (typeof coin.year === 'number') {
      stats.minYear = stats.minYear === undefined ? coin.year : Math.min(stats.minYear, coin.year)
      stats.maxYear = stats.maxYear === undefined ? coin.year : Math.max(stats.maxYear, coin.year)
    }
    const names = nameCounts.get(key) ?? new Map<string, number>()
    names.set(country, (names.get(country) ?? 0) + 1)
    nameCounts.set(key, names)
  }
  for (const stats of byKey.values()) {
    stats.types = new Set(stats.coins.map((coin) => numistaType(coin) ?? `name:${foldCountryName(coin.name)}`)).size
    stats.names = [...(nameCounts.get(stats.key) ?? new Map()).entries()].sort((a, b) => b[1] - a[1]).map(([name]) => name)
  }
  return { byKey, withoutCountry }
}

// ---- Colours ---------------------------------------------------------------

export function metricValue(stats: PlaceStats | undefined, metric: MapMetric): number {
  if (!stats) return 0
  return metric === 'value' ? stats.value : metric === 'types' ? stats.types : stats.count
}

/**
 * Upper bounds of colour steps 1–3 (step 4 is above the last). Counts use
 * fixed steps (1, 2–4, 5–9, 10+); values split the places with a value in
 * quarters.
 */
export function metricSteps(metric: MapMetric, values: number[]): [number, number, number] {
  if (metric !== 'value') return [1, 4, 9]
  const sorted = values.filter((value) => value > 0).sort((a, b) => a - b)
  if (!sorted.length) return [0, 0, 0]
  const at = (fraction: number) => sorted[Math.min(sorted.length - 1, Math.floor(fraction * (sorted.length - 1)))]!
  const round = (value: number) => (value >= 10 ? Math.round(value) : Math.round(value * 100) / 100)
  return [round(at(0.25)), round(at(0.5)), round(at(0.75))]
}

/** Colour step of a value: 0 = nothing, then 1–4. */
export function metricShade(value: number, steps: [number, number, number]): 0 | 1 | 2 | 3 | 4 {
  if (value <= 0) return 0
  if (value <= steps[0]) return 1
  if (value <= steps[1]) return 2
  if (value <= steps[2]) return 3
  return 4
}

// ---- Targets ---------------------------------------------------------------

// Not countries that issue coins of their own, or parts of one.
const NOT_TARGETS = new Set(['AQ', 'TF', 'CYN', 'EH', 'PR', 'GL', 'PT-20', 'PT-30'])

/** The places a target asks for. */
export function targetPlaces(target: MapTarget, places: MapPlaces): string[] {
  if (!target) return []
  if (target === 'euro') return EURO_ISSUERS
  const all = target === 'europe' ? continentPlaces('europe') : [...places.byKey.keys()]
  return all.filter((key) => places.byKey.has(key) && !NOT_TARGETS.has(key))
}

export interface TargetProgress {
  have: string[]
  missing: string[]
}

export function targetProgress(target: MapTarget, places: MapPlaces, byKey: Map<string, PlaceStats>): TargetProgress {
  const wanted = targetPlaces(target, places)
  return {
    have: wanted.filter((key) => (byKey.get(key)?.count ?? 0) > 0),
    missing: wanted.filter((key) => (byKey.get(key)?.count ?? 0) === 0),
  }
}

// ---- Continents ------------------------------------------------------------

export interface ContinentTotals {
  continent: Continent | 'other'
  places: number
  coins: number
  value: number
}

/** Places, coins and value by continent, the busiest first; places with none fall under "other". */
export function continentTotals(byKey: Map<string, PlaceStats>): ContinentTotals[] {
  const totals = new Map<Continent | 'other', ContinentTotals>()
  for (const stats of byKey.values()) {
    const continent = stats.continent ?? 'other'
    const entry = totals.get(continent) ?? { continent, places: 0, coins: 0, value: 0 }
    entry.places += 1
    entry.coins += stats.count
    entry.value += stats.value
    totals.set(continent, entry)
  }
  const order = [...CONTINENTS, 'other']
  return [...totals.values()].sort((a, b) => b.coins - a.coins || order.indexOf(a.continent) - order.indexOf(b.continent))
}

// ---- Mints -----------------------------------------------------------------

export interface MintGroupOnMap {
  place: MintPlace
  coins: Coin[]
}

/** The coins of each mint found on the map, and how many coins have no mint the map knows. */
export function groupCoinsByMintPlace(
  coins: Coin[],
  mints: MintPlace[],
  countryIso: (coin: Coin) => string | undefined,
): { groups: MintGroupOnMap[]; unplaced: number } {
  const byKey = new Map<string, MintGroupOnMap>()
  let unplaced = 0
  for (const coin of coins) {
    const linked = coin.mintLocationId != null ? mints.find((place) => place.key === `location:${coin.mintLocationId}`) : undefined
    const place = linked ?? findMintPlace(coin.mint ?? '', countryIso(coin), mints)
    if (!place) {
      unplaced += 1
      continue
    }
    const group = byKey.get(place.key) ?? { place, coins: [] }
    group.coins.push(coin)
    byKey.set(place.key, group)
  }
  return { groups: [...byKey.values()].sort((a, b) => b.coins.length - a.coins.length), unplaced }
}
