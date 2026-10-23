import { afterEach, describe, expect, it } from 'vitest'
import { setLocale } from '@/i18n'
import type { Coin } from '@/types'
import { groupCoins } from '@/utils/coinAttributes'
import { periodYears } from '@/utils/coinPeriods'

const coin = (id: number, country: string, period: string, year?: number) =>
  ({ id, name: `#${id}`, country, period, year: year ?? null }) as unknown as Coin

describe('coin periods', () => {
  afterEach(() => setLocale('en'))

  it('reads the years a period names', () => {
    expect(periodYears('Reino de Portugal 1777–1835')).toEqual({ from: 1777, to: 1835 })
    expect(periodYears('Rainha Isabel II')).toBeUndefined()
  })

  it('groups by country, then period, most recent first, with the years of periods that do not name them', () => {
    setLocale('pt-PT')
    const groups = groupCoins([
      coin(1, 'Portugal', 'República Portuguesa 1910–1969', 1950),
      coin(2, 'Portugal', 'União Europeia (Euro)', 2002),
      coin(3, 'Reino Unido', 'Rainha Isabel II', 1971),
      coin(4, 'Portugal', 'República Portuguesa 1986–2001', 1990),
      coin(5, 'Portugal', 'União Europeia (Euro)', 2024),
      coin(6, 'Portugal', '', 1999),
      coin(7, 'Reino Unido', 'Rainha Isabel II', 2008),
    ], 'period')

    expect(groups.map((group) => `${group.label} (${group.coins.length})`)).toEqual([
      'Portugal · União Europeia (Euro) 2002–2024 (2)',
      'Portugal · República Portuguesa 1986–2001 (1)',
      'Portugal · República Portuguesa 1910–1969 (1)',
      'Portugal · Sem período (1)',
      'Reino Unido · Rainha Isabel II 1971–2008 (2)',
    ])
    expect(groups[0]!.country).toBe('Portugal')
  })
})
