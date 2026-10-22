import { afterEach, describe, expect, it } from 'vitest'
import { setLocale } from '@/i18n'
import type { Coin } from '@/types'
import { deriveCoinYear, groupCoins, monetarySystemShortName, presentationLabel, rarityFromMintage, rarityLabel } from '@/utils/coinAttributes'

function coin(id: number, fields: Partial<Coin>): Coin {
  return { id, name: `Coin ${id}`, ...fields } as Coin
}

describe('coin attributes', () => {
  afterEach(() => setLocale('en'))

  it.each([
    [25_000_000, 'common'],
    [10_000_000, 'common'],
    [5_000_000, 'uncommon'],
    [250_000, 'scarce'],
    [40_000, 'rare'],
    [2_500, 'very_rare'],
    [400, 'extremely_rare'],
    [0, undefined],
    [undefined, undefined],
  ])('suggests rarity for a mintage of %s', (mintage, rarity) => {
    expect(rarityFromMintage(mintage)).toBe(rarity)
  })

  it('translates presentations and rarities and keeps unknown values', () => {
    setLocale('pt-PT')
    expect(presentationLabel('coincard')).toBe('Coincard')
    expect(presentationLabel('flip')).toBe('Carteira / 2x2')
    expect(rarityLabel('very_rare')).toBe('Muito rara')
    expect(rarityLabel('R2')).toBe('R2')
    expect(rarityLabel('')).toBe('')
  })
})

describe('groupCoins', () => {
  const coins = [
    coin(1, { country: 'Portugal', year: 1968, rarity: 'rare', presentation: 'coincard', monetarySystem: 'Escudo (1911-2001)' }),
    coin(2, { country: 'Angola', year: -27, rarity: 'common', monetarySystem: 'Escudo (1958-1977)' }),
    coin(3, { year: 1953, presentation: 'loose', monetarySystem: 'Euro' }),
    coin(4, { country: 'Éire', year: 1968, rarity: 'rare', presentation: 'coincard', monetarySystem: 'Euro' }),
    coin(5, { country: 'Portugal' }),
  ]
  const summary = (by: Parameters<typeof groupCoins>[1]) =>
    groupCoins(coins, by).map((group) => `${group.label}:${group.coins.map((item) => item.id).join(',')}`)

  it('keeps everything in one section when not grouping', () => {
    expect(groupCoins(coins, 'none')).toEqual([{ key: '', label: '', coins }])
  })

  it('groups by country alphabetically, ignoring accents, with no country last', () => {
    expect(summary('country')).toEqual(['Angola:2', 'Éire:4', 'Portugal:1,5', 'No country:3'])
  })

  it('groups by monetary system alphabetically, with none last', () => {
    expect(summary('monetarySystem')).toEqual(['Escudo (1911-2001):1', 'Escudo (1958-1977):2', 'Euro:3,4', 'No monetary system:5'])
  })

  it('names a monetary system without its dates for short labels', () => {
    expect(monetarySystemShortName('Escudo (1911-2001)')).toBe('Escudo')
    expect(monetarySystemShortName('Libra esterlina (decimalizada, 1971-hoje)')).toBe('Libra esterlina')
    expect(monetarySystemShortName('Euro')).toBe('Euro')
    expect(monetarySystemShortName(undefined)).toBe('')
  })

  it('groups by year chronologically, BC first, with no year last', () => {
    expect(summary('year')).toEqual(['27 BC:2', '1953:3', '1968:1,4', 'No year:5'])
  })

  it('groups by rarity and presentation in scale order', () => {
    expect(summary('rarity')).toEqual(['Common:2', 'Rare:1,4', 'No rarity:3,5'])
    expect(summary('presentation')).toEqual(['Loose:3', 'Coincard:1,4', 'No presentation:2,5'])
  })
})

describe('deriveCoinYear', () => {
  it.each([
    ['1968', 1968],
    ['27 BC – AD 14', -27],
    ['c. 46-44 BC', -46],
    ['46–44 a.C.', -46],
    ['100 BC – AD 50', -100],
    ['AD 14 – 37', 14],
    ['c. 300 a.C.', -300],
    ['1953–1974', 1953],
    ['1st century BC', undefined],
    ['séc. XV', undefined],
    ['', undefined],
    [undefined, undefined],
  ])('reads %s as %s, like the server', (dateRange, year) => {
    expect(deriveCoinYear(dateRange)).toBe(year)
  })
})
