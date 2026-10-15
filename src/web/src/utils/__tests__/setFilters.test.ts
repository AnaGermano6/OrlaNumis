import { beforeEach, describe, expect, it } from 'vitest'
import { setLocale } from '@/i18n'
import { criteriaToFilters, emptyFilters, filtersName, filtersToCriteria } from '@/utils/setFilters'

const COLLECTION_RULES = [
  { field: 'isWishlist', op: 'eq', value: false },
  { field: 'isSold', op: 'eq', value: false },
]

describe('set filters', () => {
  beforeEach(() => {
    setLocale('pt-PT')
  })

  it('makes criteria of the filters, for the collection only', () => {
    expect(filtersToCriteria(emptyFilters())).toEqual({ operator: 'and', rules: COLLECTION_RULES })
    expect(filtersToCriteria({ country: 'Portugal', denomination: '2 euro', yearFrom: 2010, yearTo: 2002, classification: 'collector', presentation: 'proof_set' })).toEqual({
      operator: 'and',
      rules: [
        { field: 'country', op: 'eq', value: 'Portugal' },
        { field: 'denomination', op: 'eq', value: '2 euro' },
        // The years in order, whichever way they were typed.
        { field: 'year', op: 'between', value: [2002, 2010] },
        { field: 'classification', op: 'eq', value: 'collector' },
        { field: 'presentation', op: 'eq', value: 'proof_set' },
        ...COLLECTION_RULES,
      ],
    })
    expect(filtersToCriteria({ ...emptyFilters(), yearFrom: 1990, yearTo: 1990 }).rules[0]).toEqual({ field: 'year', op: 'eq', value: 1990 })
    expect(filtersToCriteria({ ...emptyFilters(), yearTo: 2001 }).rules[0]).toEqual({ field: 'year', op: 'lte', value: 2001 })
  })

  it('reads the filters back from their criteria, and nothing from criteria they cannot say', () => {
    for (const filters of [
      emptyFilters(),
      { country: 'Portugal', denomination: '2 euro', yearFrom: 2002, yearTo: 2010, classification: 'commemorative' as const, presentation: 'coincard' },
      { ...emptyFilters(), yearFrom: 2002 },
      { ...emptyFilters(), yearFrom: 1990, yearTo: 1990 },
    ]) {
      expect(criteriaToFilters(filtersToCriteria(filters))).toEqual(filters)
    }
    expect(criteriaToFilters(null)).toBeNull()
    expect(criteriaToFilters({ operator: 'or', rules: [{ field: 'country', op: 'eq', value: 'Portugal' }] })).toBeNull()
    expect(criteriaToFilters({ operator: 'and', rules: [{ field: 'material', op: 'eq', value: 'Silver' }] })).toBeNull()
    expect(criteriaToFilters({ operator: 'and', rules: [{ operator: 'or', rules: [] }] })).toBeNull()
    expect(criteriaToFilters({ operator: 'and', rules: [{ field: 'country', op: 'contains', value: 'Port' }] })).toBeNull()
  })

  it('names a set after its filters', () => {
    expect(filtersName(emptyFilters())).toBe('Todas as minhas moedas')
    expect(filtersName({ country: 'Portugal', denomination: '2 euro', yearFrom: 2002, yearTo: 2010, classification: 'commemorative', presentation: 'coincard' }))
      .toBe('Portugal · 2 euro · 2002–2010 · Comemorativas · Coincard')
    expect(filtersName({ ...emptyFilters(), yearFrom: 2002 })).toBe('desde 2002')
    expect(filtersName({ ...emptyFilters(), yearTo: 2001 })).toBe('até 2001')
  })
})
