import { beforeEach, describe, expect, it } from 'vitest'
import { setLocale } from '@/i18n'
import { buildRomanDenariusCore } from '@/test/fixtures/coins'
import { agenticAlbumSlots, checklistAlbumSlot, checklistDescription, coinAlbumSlot, euroCoinName } from '@/utils/setAlbum'

// Intl writes a no-break space before the euro sign.
const nbsp = (text: string) => text.replace(/\u00a0/g, ' ')

describe('set album', () => {
  beforeEach(() => {
    setLocale('pt-PT')
  })

  it("shows a catalogue's slot by its coin and face values, in the slot's currency", () => {
    const slot = checklistAlbumSlot(
      { key: 'incm:farois', title: 'Faróis', country: 'PT', year: 2025, faceValue: 5, faceValues: [5, 10], owned: true, coinId: 3, coinImage: 'coins/farois.jpg', pageUrl: 'https://incm' },
      { source: 'catalogue', country: 'PT' },
    )
    expect({ ...slot, detail: nbsp(slot.detail ?? '') }).toEqual({
      key: 'incm:farois', title: 'Faróis', detail: '5 € · 10 €', year: 2025, owned: true, coinId: 3, coinImage: 'coins/farois.jpg', imageUrl: undefined, pageUrl: 'https://incm',
    })
    expect(checklistAlbumSlot({ key: 'numista:1', title: 'Expo 98', faceValue: 1000, currency: 'PTE', owned: false }, { source: 'catalogue', country: 'PT', currency: 'PTE' }).detail).toBe('1000 escudos')
  })

  it("shows a challenge's slots by country, or by coin of the year's series", () => {
    const filled = checklistAlbumSlot({ key: 'country:DE', country: 'DE', year: 2005, owned: true, coinId: 8, coinName: '10 cêntimos' }, { source: 'challenge', challenge: 'euro-countries' })
    expect(filled).toEqual(expect.objectContaining({ title: 'Alemanha', flag: '🇩🇪', detail: '10 cêntimos', year: 2005, owned: true }))
    expect(filled.country).toBeUndefined()
    expect(checklistAlbumSlot({ key: 'country:FR', country: 'FR', owned: false, imageUrl: 'https://ecb/fr.jpg' }, { source: 'challenge', challenge: 'euro-countries' }))
      .toEqual(expect.objectContaining({ title: 'França', detail: 'Qualquer moeda do euro', year: null, owned: false, imageUrl: 'https://ecb/fr.jpg' }))

    const joint = { source: 'challenge', challenge: 'joint-issue', year: 2007 } as const
    expect(checklistAlbumSlot({ key: 'joint:PT', country: 'PT', title: '50th anniversary of the Treaty of Rome', year: 2007, owned: true, coinName: 'Tratado de Roma' }, joint))
      .toEqual(expect.objectContaining({ title: 'Portugal', flag: '🇵🇹', detail: 'Tratado de Roma', year: 2007 }))
    expect(checklistAlbumSlot({ key: 'joint:IT', country: 'IT', title: '50th anniversary of the Treaty of Rome', year: 2007, owned: false, imageUrl: 'https://ecb/it.jpg' }, joint))
      .toEqual(expect.objectContaining({ title: 'Itália', detail: undefined, imageUrl: 'https://ecb/it.jpg' }))

    const series = { source: 'challenge', challenge: 'annual-series', country: 'PT', year: 2002 } as const
    const cent = checklistAlbumSlot({ key: 'annual:1cent', country: 'PT', year: 2002, faceValue: 0.01, owned: false }, series)
    expect(cent.title).toBe('1 cêntimo')
    // The series is one country's: its flag is on the set, not on each coin.
    expect(cent.flag).toBeUndefined()
    expect(checklistAlbumSlot({ key: 'annual:50cent', country: 'PT', year: 2002, faceValue: 0.5, owned: true, coinName: 'Meu' }, series)).toEqual(expect.objectContaining({ title: '50 cêntimos', detail: 'Meu' }))
    expect(nbsp(euroCoinName(2))).toBe('2 €')
  })

  it("describes what a checklist set lists", () => {
    expect(nbsp(checklistDescription({ source: 'catalogue', country: 'PT', faceValue: 2 }))).toBe('Catálogo · 🇵🇹 Portugal · 2 €')
    expect(checklistDescription({ source: 'catalogue', country: 'PT', currency: 'PTE', faceValue: 100 })).toBe('Catálogo · 🇵🇹 Portugal · Escudos · 100 escudos')
    expect(checklistDescription({ source: 'challenge', challenge: 'euro-countries' })).toBe('Uma moeda de cada país do euro')
    expect(checklistDescription({ source: 'challenge', challenge: 'joint-issue', year: 2007 })).toBe('Emissão comum de 2007: a mesma moeda, um lugar por país')
    expect(checklistDescription({ source: 'challenge', challenge: 'annual-series', country: 'PT', year: 2002 })).toBe('🇵🇹 Portugal 2002: série anual')
  })

  it("leaves a goal set's wishlist coins to find, and fills an agentic set's targets with their coins", () => {
    const coin = buildRomanDenariusCore({ id: 4, name: 'Denário', isWishlist: true, year: 14, country: 'Roma' })
    expect(coinAlbumSlot(coin, true)).toEqual(expect.objectContaining({ key: 'coin:4', owned: false, wishlist: true, year: 14, country: 'Roma' }))
    expect(coinAlbumSlot(coin, false)).toEqual(expect.objectContaining({ owned: true, wishlist: false }))

    const target = { id: 9, setId: 1, label: 'Augusto', year: 14, sortOrder: 0 }
    expect(agenticAlbumSlots({ totalTargets: 1, completedTargets: 1, completionPercentage: 100, missingTargets: [], targetMatches: [{ target, coin: { ...coin, isWishlist: false } }] }))
      .toEqual([expect.objectContaining({ key: 'target:9', title: 'Denário', detail: 'Augusto', owned: true, coinId: 4, targetId: 9 })])
    expect(agenticAlbumSlots({ totalTargets: 1, completedTargets: 0, completionPercentage: 0, missingTargets: [target] }))
      .toEqual([expect.objectContaining({ key: 'target:9', title: 'Augusto', owned: false, coinId: undefined, targetId: 9 })])
    expect(agenticAlbumSlots(null)).toEqual([])
  })
})
