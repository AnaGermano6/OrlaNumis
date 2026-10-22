import { describe, expect, it } from 'vitest'
import type { CoinSetSummary } from '@/types'
import { setsByCountry } from '@/utils/setCountries'

function set(id: number, name: string, countries: CoinSetSummary['countries'], partial = false): CoinSetSummary {
  return { id, name, color: '#c9a84c', setType: 'checklist', coinCount: 0, totalValue: 0, countries, partial }
}

describe('setsByCountry', () => {
  it('adds up the places of every set in each country, most places first', () => {
    const result = setsByCountry([
      set(1, 'Portugal · 2 €', [{ country: 'PT', owned: 20, total: 60 }]),
      set(2, 'Erasmus', [{ country: 'DE', owned: 0, total: 1 }, { country: 'AT', owned: 1, total: 1 }, { country: 'PT', owned: 1, total: 1 }]),
      set(3, 'Réis types', [{ country: 'PT', owned: 3, total: 40 }], true),
      { id: 4, name: 'Mine', color: '#fff', setType: 'standard', coinCount: 5, totalValue: 0 },
    ])

    expect(result.map((entry) => entry.country)).toEqual(['PT', 'AT', 'DE'])
    expect(result[0]).toMatchObject({ owned: 24, total: 101, missing: 77, partial: true })
    expect(result[0]!.sets.map((item) => `${item.name} ${item.owned}/${item.total}`)).toEqual(['Portugal · 2 € 20/60', 'Erasmus 1/1', 'Réis types 3/40'])
    expect(result[1]).toMatchObject({ country: 'AT', missing: 0, partial: false })
  })

  it('has nothing for sets without places', () => {
    expect(setsByCountry([set(1, 'Empty', []), set(2, 'Odd', [{ country: '', owned: 0, total: 2 }, { country: 'FR', owned: 0, total: 0 }])])).toEqual([])
  })
})
