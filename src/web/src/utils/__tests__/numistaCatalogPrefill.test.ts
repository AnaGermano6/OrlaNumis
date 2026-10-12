import { afterEach, describe, expect, it } from 'vitest'
import { setLocale } from '@/i18n'
import type { NumistaCatalogType } from '@/api/client'
import type { CatalogRegistry } from '@/types'
import { buildCatalogPrefill, categoryForType, eraForYear, materialFromComposition } from '@/utils/numistaCatalogPrefill'

const CATEGORIES = ['Roman', 'Greek', 'Byzantine', 'Modern', 'Other']
const ERAS = ['ancient', 'medieval', 'modern']
const MATERIALS = ['Gold', 'Silver', 'Bronze', 'Copper', 'Electrum', 'Other']
const REGISTRY: CatalogRegistry[] = [
  { id: 1, catalog: 'KM', displayName: 'Krause-Mishler', era: 'modern', volumeRequired: false },
  { id: 2, catalog: 'Numista', displayName: 'Numista', era: 'modern', volumeRequired: false },
  { id: 3, catalog: 'RIC', displayName: 'Roman Imperial Coinage', era: 'ancient', volumeRequired: true },
]

function modernType(overrides: Partial<NumistaCatalogType> = {}): NumistaCatalogType {
  return {
    id: 2933,
    url: 'https://en.numista.com/catalogue/pieces2933.html',
    title: '2.50 Escudos',
    issuer: { code: 'angola', name: 'Angola' },
    minYear: 1953,
    maxYear: 1974,
    denomination: '2.5 Escudos',
    rulers: ['Província ultramarina de Portugal'],
    composition: 'Cuproníquel',
    weightGrams: 3.5,
    diameterMm: 20,
    thicknessMm: 1.46,
    shape: 'Redondo',
    mints: ['Imprensa Nacional - Casa da Moeda'],
    obverse: { description: 'Brasão de armas', lettering: 'ANGOLA\n2$50' },
    reverse: { description: 'Brasão de Portugal', lettering: 'REPÚBLICA·PORTUGUESA' },
    edge: { description: 'Serrilhado' },
    references: [{ catalogue: 'KM', number: '77' }, { catalogue: 'Schön', number: '17' }],
    photos: [],
    ...overrides,
  }
}

const options = { categories: CATEGORIES, eras: ERAS, materials: MATERIALS, registry: REGISTRY }

describe('buildCatalogPrefill', () => {
  afterEach(() => setLocale('en'))

  it('fills a modern coin from the chosen issue and grade', () => {
    const { fields, references } = buildCatalogPrefill({
      type: modernType(),
      issue: { id: 14184, year: 1968, mintage: 5000000, mintLetter: 'L' },
      grade: 'vf',
      currentValue: 23.808158708367692,
      wishlist: false,
    }, options)

    expect(fields).toMatchObject({
      name: '2.50 Escudos 1968 – Angola',
      category: 'Modern',
      era: 'modern',
      denomination: '2.5 Escudos',
      ruler: 'Província ultramarina de Portugal',
      dateRange: '1968',
      year: 1968,
      country: 'Angola',
      rarity: 'uncommon',
      mint: 'Imprensa Nacional - Casa da Moeda (L)',
      material: 'Other',
      weightGrams: 3.5,
      diameterMm: 20,
      grade: 'VF',
      obverseInscription: 'ANGOLA / 2$50',
      reverseInscription: 'REPÚBLICA·PORTUGUESA',
      obverseDescription: 'Brasão de armas',
      currentValue: 23.81,
      isWishlist: false,
    })
    expect(references).toEqual([
      { catalog: 'Numista', number: '2933', uri: 'https://en.numista.com/catalogue/pieces2933.html' },
      { catalog: 'KM', number: '77' },
    ])
    expect(fields.notes).toContain('From the Numista catalog: 2.50 Escudos (Angola)')
    expect(fields.notes).toContain('Composition: Cuproníquel')
    expect(fields.notes).toContain('Thickness: 1.46 mm')
    expect(fields.notes).toContain('Mintage: 5,000,000')
    expect(fields.notes).toContain('Other references: Schön 17')
  })

  it('fills an ancient provincial coin with BC dates and keeps unregistered references in the notes', () => {
    setLocale('pt-PT')
    const { fields, references } = buildCatalogPrefill({
      type: modernType({
        id: 372576,
        url: 'https://en.numista.com/catalogue/pieces372576.html',
        title: 'Semis - Augustus',
        issuer: { code: 'corinth_rom_prov', name: 'Corinth' },
        minYear: -2,
        maxYear: -1,
        rulers: ['Augustus'],
        composition: 'Bronze (leaded)',
        references: [{ catalogue: 'RIC', number: 'I 207' }, { catalogue: 'RPC Online I', number: '1136' }],
      }),
      wishlist: true,
    }, options)

    expect(fields).toMatchObject({
      name: 'Semis - Augustus – Corinth',
      category: 'Roman',
      era: 'ancient',
      dateRange: '2 a.C.–1 a.C.',
      year: null,
      country: 'Corinth',
      rarity: '',
      material: 'Bronze',
      grade: '',
      currentValue: undefined,
      isWishlist: true,
    })
    expect(references).toHaveLength(1)
    expect(fields.notes).toContain('Outras referências: RIC I 207, RPC Online I 1136')
  })
})

describe('catalog field suggestions', () => {
  it.each([
    ['Silver (.900)', 'Silver'],
    ['Prata', 'Silver'],
    ['Gold', 'Gold'],
    ['Ouro (.917)', 'Gold'],
    ['Nordic gold', 'Other'],
    ['Electrum', 'Electrum'],
    ['Aluminium-bronze', 'Bronze'],
    ['Copper', 'Copper'],
    ['Copper-nickel', 'Other'],
    ['Cobre', 'Copper'],
    ['Zinc', 'Other'],
    [undefined, 'Other'],
  ])('maps composition %s to %s', (composition, material) => {
    expect(materialFromComposition(composition, MATERIALS)).toBe(material)
  })

  it('marks the coin as kept in a coin card when Numista says the type is issued in one', () => {
    const inCard = buildCatalogPrefill({ type: modernType({ issuedInCoincard: true }), wishlist: false }, options)
    expect(inCard.fields.presentation).toBe('coincard')
    // Otherwise the form keeps whatever presentation it has.
    const loose = buildCatalogPrefill({ type: modernType(), wishlist: false }, options)
    expect('presentation' in loose.fields).toBe(false)
  })

  it('falls back to configured categories and eras', () => {
    expect(categoryForType(modernType({ issuer: { code: 'byzantine_empire', name: 'Byzantine Empire' }, minYear: 900, maxYear: 950 }), CATEGORIES)).toBe('Byzantine')
    expect(categoryForType(modernType({ issuer: { code: 'athens', name: 'Ancient Greece – Athens' }, minYear: -450, maxYear: -400 }), CATEGORIES)).toBe('Greek')
    expect(categoryForType(modernType({ minYear: 1200, maxYear: 1250 }), CATEGORIES)).toBe('Other')
    expect(categoryForType(modernType(), ['Imperial', 'Other'])).toBe('Other')
    expect(categoryForType(modernType(), ['Imperial'])).toBe('Imperial')
    expect(eraForYear(1250, ERAS)).toBe('medieval')
    expect(eraForYear(1968, ['ancient'])).toBe('')
    expect(eraForYear(undefined, ERAS)).toBe('')
  })
})
