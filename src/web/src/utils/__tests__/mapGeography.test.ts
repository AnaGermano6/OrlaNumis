import fs from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'
import { describe, expect, it } from 'vitest'
import { EURO_ISSUERS, SMALL_TERRITORIES, continentOf, extinctTerritory } from '@/utils/mapGeography'

const __dirname = path.dirname(fileURLToPath(import.meta.url))
const world = JSON.parse(fs.readFileSync(path.resolve(__dirname, '../../assets/geo/world-countries.json'), 'utf8')) as { features: { properties: { iso: string } }[] }

describe('mapGeography', () => {
  it('knows the continent of every place on the map', () => {
    const places = [...world.features.map((feature) => feature.properties.iso), ...SMALL_TERRITORIES.map((point) => point.iso)]
    expect(places.filter((iso) => !continentOf(iso))).toEqual([])
    expect(continentOf('PT')).toBe('europe')
    expect(continentOf('CV')).toBe('africa')
    expect(continentOf('MO')).toBe('asia')
  })

  it('draws as points the small states the outlines leave out', () => {
    const outlines = new Set(world.features.map((feature) => feature.properties.iso))
    expect(SMALL_TERRITORIES.filter((point) => outlines.has(point.iso)).map((point) => point.iso)).toEqual([])
    // Every euro issuer is somewhere on the map.
    const points = new Set(SMALL_TERRITORIES.map((point) => point.iso))
    expect(EURO_ISSUERS.filter((iso) => !outlines.has(iso) && !points.has(iso))).toEqual([])
    expect(EURO_ISSUERS).toHaveLength(25)
  })

  it('finds the present territory of countries that no longer exist, in Numista\'s forms too', () => {
    expect(extinctTerritory('Jugoslávia')).toEqual(['RS', 'HR', 'SI', 'BA', 'MK', 'ME', 'XK'])
    expect(extinctTerritory('Jugoslávia, República Federal Socialista da')).toContain('HR')
    expect(extinctTerritory('União Soviética (URSS)')).toContain('RU')
    expect(extinctTerritory('Checoslováquia')).toEqual(['CZ', 'SK'])
    expect(extinctTerritory('Império Otomano')).toContain('TR')
    expect(extinctTerritory('Antilhas Holandesas')).toEqual(['CW', 'SX'])
    expect(extinctTerritory('Portugal')).toEqual([])
  })
})
