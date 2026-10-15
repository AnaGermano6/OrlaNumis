import { beforeEach, describe, expect, it } from 'vitest'
import { setLocale } from '@/i18n'
import type { CoinRelease } from '@/api/client'
import { buildReleasePrefill, buildReleaseWishlistCoin, releaseAddRoute, releaseDenomination, releaseFromQuery } from '@/utils/releasePrefill'

const OPTIONS = { categories: ['Roman', 'Modern', 'Other'], materials: ['Silver', 'Gold', 'Other'] }

function release(fields: Partial<CoinRelease>): CoinRelease {
  return {
    source: 'incm', kind: 'collector', country: 'PT', title: 'Faróis de Portugal', faceValue: 5, imageUrl: 'https://cdn/farois.jpg',
    pageUrl: 'https://loja.incm.pt/products/farois', from: '2025-11', to: '2025-11', upcoming: false, owned: false, ...fields,
  }
}

describe('release prefill', () => {
  beforeEach(() => {
    setLocale('pt-PT')
  })

  it('writes face values as the collection does', () => {
    expect(releaseDenomination(2)).toBe('2 euro')
    expect(releaseDenomination(7.5)).toBe('7½ euro')
    expect(releaseDenomination(0.5)).toBe('½ euro')
    expect(releaseDenomination(1.25)).toBe('1,25 euro')
    expect(releaseDenomination(200, 'PTE')).toBe('200 escudos')
    expect(releaseDenomination(1000, 'PTE')).toBe('1000 escudos')
    expect(releaseDenomination(1, 'PTE')).toBe('1 escudo')
    expect(releaseDenomination(2.5, 'PTE')).toBe('2½ escudos')
  })

  it("fills in a mint's coin, its metal only when it comes in one", () => {
    const { fields, references } = buildReleasePrefill(release({ variants: ['Ouro Proof'], volume: 2500 }), OPTIONS)
    expect(fields).toMatchObject({
      name: '5 euro 2025 – Faróis de Portugal',
      category: 'Modern',
      material: 'Gold',
      denomination: '5 euro',
      year: 2025,
      country: 'Portugal',
      circulation: 'non_circulating',
      // From its mintage, as a starting point.
      rarity: 'very_rare',
      referenceUrl: 'https://loja.incm.pt/products/farois',
      referenceText: 'INCM',
      notes: 'Versões: Ouro Proof\nTiragem: 2500',
      isWishlist: false,
    })
    expect(references).toEqual([])
    expect(buildReleasePrefill(release({ variants: ['BNC', 'Prata Proof'] }), OPTIONS).fields.material).toBe('Other')
  })

  it("leaves a joint issue's country to the user, and names the ECB in the app's language", () => {
    const { fields } = buildReleasePrefill(release({ source: 'ecb', kind: 'commemorative', country: 'EU', title: '35 years of the Erasmus programme', faceValue: 2, from: '2022-07' }), OPTIONS)
    expect(fields).toMatchObject({ name: '2 euro 2022 – 35 years of the Erasmus programme', country: '', circulation: 'current_commemorative', referenceText: 'BCE' })
  })

  it("saves a wishlist coin with a shop's page as its source, and another source's in the notes", () => {
    expect(buildReleaseWishlistCoin(release({ variants: ['Prata Proof'] }), OPTIONS)).toEqual({
      name: '5 euro 2025 – Faróis de Portugal',
      category: 'Modern',
      material: 'Silver',
      denomination: '5 euro',
      year: 2025,
      country: 'Portugal',
      circulation: 'non_circulating',
      rarity: undefined,
      referenceUrl: 'https://loja.incm.pt/products/farois',
      referenceText: 'INCM',
      notes: 'Versões: Prata Proof',
      quantity: 1,
      isWishlist: true,
      references: undefined,
    })
    // The ECB's year page lists every coin of the year: in the notes.
    const ecb = buildReleaseWishlistCoin(release({ source: 'ecb', kind: 'commemorative', title: 'Rotary', faceValue: 2, pageUrl: 'https://www.ecb.europa.eu/2026.html', from: '2026-07', volume: 500000 }), OPTIONS)
    expect(ecb).toMatchObject({ name: '2 euro 2026 – Rotary', referenceUrl: undefined, referenceText: undefined, isWishlist: true, circulation: 'current_commemorative' })
    expect(ecb.notes).toBe(`Tiragem: ${new Intl.NumberFormat('pt-PT').format(500000)}\nFonte: BCE – https://www.ecb.europa.eu/2026.html`)
    // A Numista coin keeps its Numista reference.
    const numista = buildReleaseWishlistCoin(release({ source: 'numista', kind: 'commemorative', title: 'Colombo', faceValue: 200, currency: 'PTE', pageUrl: 'https://pt.numista.com/catalogue/pieces100.html', from: '1991-01' }), OPTIONS)
    expect(numista).toMatchObject({ denomination: '200 escudos', referenceUrl: undefined, references: [{ catalog: 'Numista', number: '100', uri: 'https://pt.numista.com/catalogue/pieces100.html' }] })
    expect(numista.notes).toBe('Fonte: Numista – https://pt.numista.com/catalogue/pieces100.html')
  })

  it('goes through the address and back, and reads nothing it does not trust', () => {
    const original = release({ variants: ['BNC'], volume: 10000, upcoming: true })
    const location = releaseAddRoute(original) as { name: string; query: { release: string } }
    expect(location.name).toBe('add-coin')
    expect(releaseFromQuery(location.query.release)).toEqual({ ...original, imageUrl: '', to: original.from, owned: false })

    expect(releaseFromQuery(undefined)).toBeNull()
    expect(releaseFromQuery('{oops')).toBeNull()
    expect(releaseFromQuery(JSON.stringify({ ...original, source: 'shop' }))).toBeNull()
    expect(releaseFromQuery(JSON.stringify({ ...original, faceValue: -1 }))).toBeNull()
    expect(releaseFromQuery(JSON.stringify({ ...original, pageUrl: 'javascript:alert(1)' }))?.pageUrl).toBe('')
  })
})
