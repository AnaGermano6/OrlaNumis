import { describe, expect, it } from 'vitest'
import worldCountries from '@/assets/geo/world-countries.json'
import { buildMapPlaces, placeKeyOf } from '@/utils/collectionMap'
import { foldCountryName, type WorldCountryCollection } from '@/utils/worldCountries'

const collection = worldCountries as unknown as WorldCountryCollection
const places = buildMapPlaces(collection)

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
    ['Estados Unidos da América', 'US'],
    ['Alemanha, República Federal da', 'DE'],
    ['Mónaco', 'MC'],
    ['Cabo Verde', 'CV'],
  ])('places %s on the map as %s', (name, iso) => {
    expect(placeKeyOf(name, places)).toBe(iso)
  })

  it('keeps historical issuers off the map', () => {
    expect(placeKeyOf('Roman Empire', places)).toBe('name:Roman Empire')
    expect(placeKeyOf('Corinth', places)).toBe('name:Corinth')
  })
})
