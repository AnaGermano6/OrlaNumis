import { describe, expect, it } from 'vitest'
import type { Coin } from '@/types'
import {
  EMPTY_MAP_FILTERS,
  buildMapPlaces,
  continentTotals,
  filterMapCoins,
  groupCoinsByMintPlace,
  groupCoinsByPlace,
  metricShade,
  metricSteps,
  placeKeyOf,
  targetProgress,
  type CoinStatusSets,
} from '@/utils/collectionMap'
import { MODERN_MINTS } from '@/utils/modernMints'
import type { WorldCountryCollection } from '@/utils/worldCountries'

const square = { type: 'Polygon' as const, coordinates: [[[0, 0], [1, 0], [1, 1], [0, 0]]] }
const WORLD: WorldCountryCollection = {
  type: 'FeatureCollection',
  features: [
    { type: 'Feature', properties: { iso: 'PT', name: 'Portugal', namePt: 'Portugal', aliases: ['Portugal'] }, geometry: square },
    { type: 'Feature', properties: { iso: 'DE', name: 'Germany', namePt: 'Alemanha', aliases: ['Germany', 'Alemanha', 'Alemanha, República Federal da'] }, geometry: square },
    { type: 'Feature', properties: { iso: 'RS', name: 'Serbia', namePt: 'Sérvia', aliases: ['Serbia'] }, geometry: square },
  ],
}

let nextId = 0
function coin(fields: Partial<Coin>): Coin {
  nextId += 1
  return { id: nextId, name: `Coin ${nextId}`, country: '', era: 'modern', references: [], images: [], ...fields } as Coin
}

const NO_STATUS: CoinStatusSets = { current: new Set(), demonetized: new Set(), nonCirculating: new Set(), commemorative: new Set(), regular: new Set() }

describe('collectionMap', () => {
  const places = buildMapPlaces(WORLD)
  const coins = [
    coin({ country: 'Portugal', year: 1990, currentValue: 2, references: [{ catalog: 'Numista', number: '7187' }] as Coin['references'] }),
    coin({ country: 'Portugal', year: 2002, currentValue: 3, references: [{ catalog: 'Numista', number: '7187' }] as Coin['references'] }),
    coin({ country: 'Região Autónoma dos Açores', year: 1991, currentValue: 5 }),
    coin({ country: 'Alemanha', year: 2026, mint: 'D' }),
    coin({ country: 'Alemanha, República Federal da', year: 1970, era: 'modern' }),
    coin({ country: 'Jugoslávia', year: 1987 }),
    coin({ country: 'Atlântida', year: 1900 }),
    coin({ country: '' }),
  ]

  it('places countries on outlines, small territories on points and the rest off the map', () => {
    expect(placeKeyOf('Portugal', places)).toBe('PT')
    expect(placeKeyOf('Alemanha, República Federal da', places)).toBe('DE')
    expect(placeKeyOf('Região Autónoma dos Açores', places)).toBe('PT-20')
    expect(placeKeyOf('Mónaco', places)).toBe('MC')
    expect(placeKeyOf('Jugoslávia', places)).toBe('name:Jugoslávia')
    // Long official names and Numista's "Name, form" find the country too.
    expect(placeKeyOf('Estados Unidos da América', buildMapPlaces({ type: 'FeatureCollection', features: [{ type: 'Feature', properties: { iso: 'US', name: 'United States of America', namePt: 'Estados Unidos', aliases: [] }, geometry: square }] }))).toBe('US')
    expect(placeKeyOf('Alemanha, República Federal da (RFA)', places)).toBe('DE')

    const { byKey, withoutCountry } = groupCoinsByPlace(coins, places, 'pt-PT')
    expect(withoutCountry).toBe(1)
    const portugal = byKey.get('PT')!
    expect(portugal).toMatchObject({ kind: 'outline', label: 'Portugal', count: 2, value: 5, types: 1, minYear: 1990, maxYear: 2002, continent: 'europe' })
    expect(byKey.get('PT-20')).toMatchObject({ kind: 'point', label: 'Açores', count: 1 })
    // Both German names count as one country, the commoner name first.
    expect(byKey.get('DE')).toMatchObject({ count: 2, label: 'Alemanha', names: ['Alemanha', 'Alemanha, República Federal da'] })
    // An extinct country stays off the map, with its present territory and continent.
    expect(byKey.get('name:Jugoslávia')).toMatchObject({ kind: 'offMap', continent: 'europe', territory: ['RS', 'HR', 'SI', 'BA', 'MK', 'ME', 'XK'] })
    expect(byKey.get('name:Atlântida')).toMatchObject({ kind: 'offMap', territory: [] })
    expect(byKey.get('name:Atlântida')!.continent).toBeUndefined()
  })

  it('filters by era, year range and status', () => {
    const statuses: CoinStatusSets = { ...NO_STATUS, current: new Set([coins[1]!.id]), commemorative: new Set([coins[2]!.id]) }
    expect(filterMapCoins(coins, { ...EMPTY_MAP_FILTERS, yearFrom: 1990, yearTo: 1999 }, statuses).map((c) => c.country))
      .toEqual(['Portugal', 'Região Autónoma dos Açores'])
    expect(filterMapCoins(coins, { ...EMPTY_MAP_FILTERS, circulation: 'current' }, statuses)).toEqual([coins[1]])
    expect(filterMapCoins(coins, { ...EMPTY_MAP_FILTERS, circulation: 'nonCirculating' }, { ...statuses, nonCirculating: new Set([coins[0]!.id]) })).toEqual([coins[0]])
    expect(filterMapCoins(coins, { ...EMPTY_MAP_FILTERS, commemorative: 'true' }, statuses)).toEqual([coins[2]])
    expect(filterMapCoins(coins, { ...EMPTY_MAP_FILTERS, era: 'ancient' }, statuses)).toEqual([])
    expect(filterMapCoins(coins, EMPTY_MAP_FILTERS, statuses)).toHaveLength(coins.length)
  })

  it('colours counts in fixed steps and values in quarters', () => {
    expect(metricSteps('coins', [])).toEqual([1, 4, 9])
    expect([0, 1, 3, 7, 12].map((count) => metricShade(count, metricSteps('coins', [])))).toEqual([0, 1, 2, 3, 4])
    const steps = metricSteps('value', [0, 1, 2, 3, 4, 100])
    expect(steps).toEqual([2, 3, 4])
    expect([0.5, 2.5, 4, 100].map((value) => metricShade(value, steps))).toEqual([1, 2, 3, 4])
  })

  it('tells the places a target still misses', () => {
    const { byKey } = groupCoinsByPlace(coins, places, 'pt-PT')
    const euro = targetProgress('euro', places, byKey)
    expect(euro.have.sort()).toEqual(['DE', 'PT'])
    expect(euro.missing).toHaveLength(23)
    expect(euro.missing).toContain('MC')
    // The world asks for every country drawn, not for regions such as the Azores.
    const world = targetProgress('world', places, byKey)
    expect(world.missing).toContain('RS')
    expect([...world.have, ...world.missing]).not.toContain('PT-20')
  })

  it('totals places, coins and value by continent', () => {
    const { byKey } = groupCoinsByPlace(coins, places, 'pt-PT')
    expect(continentTotals(byKey)).toEqual([
      { continent: 'europe', places: 4, coins: 6, value: 10 },
      { continent: 'other', places: 1, coins: 1, value: 0 },
    ])
  })

  it('groups coins by the mint found from their text or mark', () => {
    const minted = [coin({ country: 'Portugal', mint: 'Imprensa Nacional - Casa da Moeda (INCM)' }), coin({ country: 'Alemanha', mint: 'D' }), coin({ country: 'Portugal', mint: 'State Mint' }), coin({ country: 'Portugal' })]
    const { groups, unplaced } = groupCoinsByMintPlace(minted, MODERN_MINTS, (c) => placeKeyOf(c.country ?? '', places))
    expect(groups.map((group) => [group.place.key, group.coins.length])).toEqual([['builtin:incm', 1], ['builtin:munich', 1]])
    expect(unplaced).toBe(2)
  })
})
