import { describe, expect, it } from 'vitest'
import worldCountries from '@/assets/geo/world-countries.json'
import { buildCountryIndex, countryShade, distributeCountries, foldCountryName, type WorldCountryCollection } from '@/utils/worldCountries'

const collection = worldCountries as unknown as WorldCountryCollection
const index = buildCountryIndex(collection)

describe('world map country matching', () => {
  it('ships every country with a code, names and an outline', () => {
    expect(collection.features.length).toBeGreaterThan(170)
    for (const feature of collection.features) {
      expect(feature.properties.iso).toMatch(/^[A-Z]{2,3}$/)
      expect(feature.properties.namePt).not.toBe('')
      expect(['Polygon', 'MultiPolygon']).toContain(feature.geometry.type)
    }
  })

  it('folds accents, case, punctuation and spacing', () => {
    expect(foldCountryName('  República   Checa ')).toBe('republica checa')
    expect(foldCountryName("Côte d'Ivoire")).toBe('cote d ivoire')
  })

  it.each([
    ['Portugal', 'PT'],
    ['Angola', 'AO'],
    ['Estados Unidos', 'US'],
    ['United States of America', 'US'],
    ['França', 'FR'],
    ['france', 'FR'],
    ['Noruega', 'NO'],
    ['Alemanha', 'DE'],
    ['Reino Unido', 'GB'],
    ['United Kingdom', 'GB'],
    ['Espanha', 'ES'],
    ['Brasil', 'BR'],
    ['Moçambique', 'MZ'],
  ])('places %s on the map as %s', (name, iso) => {
    expect(index.get(foldCountryName(name))).toBe(iso)
  })

  it('adds up names of the same country and keeps historical issuers apart', () => {
    const { byIso, unmatched } = distributeCountries([
      { country: 'Portugal', count: 3 },
      { country: 'Portuguese Republic', count: 1 },
      { country: 'Angola', count: 2 },
      { country: 'Corinth', count: 1 },
      { country: 'Roman Empire', count: 4 },
    ], index)

    expect(byIso.get('PT')).toEqual({ iso: 'PT', count: 4, names: [{ country: 'Portugal', count: 3 }, { country: 'Portuguese Republic', count: 1 }] })
    expect(byIso.get('AO')?.count).toBe(2)
    expect(unmatched).toEqual([{ country: 'Roman Empire', count: 4 }, { country: 'Corinth', count: 1 }])
  })

  it.each([[0, 0], [1, 1], [2, 2], [4, 2], [5, 3], [9, 3], [10, 4], [250, 4]])('shades %i coins as step %i', (count, shade) => {
    expect(countryShade(count)).toBe(shade)
  })
})
