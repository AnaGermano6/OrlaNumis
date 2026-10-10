import { afterEach, describe, expect, it } from 'vitest'
import { setLocale } from '@/i18n'
import type { CoinPhotoCandidate } from '@/api/client'
import type { Coin } from '@/types'
import { coinPhotoDetails, photoCaption, photoCreditLine, photoLoadVia, photoSourceName } from '@/utils/coinPhotos'

function coin(fields: Partial<Coin>): Coin {
  return { id: 1, name: 'Coin', ...fields } as Coin
}

describe('coin photos', () => {
  afterEach(() => setLocale('en'))

  it('reads a coin\'s year, country, denomination and subject for a photo search', () => {
    expect(coinPhotoDetails(coin({ name: '2 euro 2026 D – Bremen, Casa do Clima', country: 'Alemanha', denomination: '2 euro', year: 2026 })))
      .toEqual({ year: 2026, country: 'Alemanha', denomination: '2 euro', subject: 'Bremen, Casa do Clima' })
    // A name ending with the country has no subject.
    expect(coinPhotoDetails(coin({ name: '2 Euros (Rotary) 2026 – Portugal', country: 'Portugal', denomination: '2 Euros', year: 2026 })).subject).toBeUndefined()
    expect(coinPhotoDetails(coin({ name: 'Denarius', year: null }))).toEqual({ year: undefined, country: undefined, denomination: undefined, subject: undefined })
  })

  it('names the ECB as the source of its images', () => {
    setLocale('pt-PT')
    const photo = { source: 'ecb', author: 'European Central Bank', thumbnail: '', imageUrl: 'https://www.ecb.europa.eu/a.jpg', pageUrl: 'https://www.ecb.europa.eu/p.html' } as CoinPhotoCandidate
    expect(photoSourceName(photo)).toBe('Banco Central Europeu')
    expect(photoCreditLine('obverse', photo)).toContain('European Central Bank, Banco Central Europeu')
    expect(photoCreditLine('obverse', photo)).toContain('https://www.ecb.europa.eu/p.html')
  })

  it('names the Portuguese mint as the source of its photos and loads them directly', () => {
    setLocale('pt-PT')
    const photo = { source: 'incm', author: 'INCM', thumbnail: '', imageUrl: 'https://cdn.shopify.com/s/files/1/0525/1910/8762/files/A.jpg', pageUrl: 'https://loja.incm.pt/products/o-conhecimento' } as CoinPhotoCandidate
    expect(photoSourceName(photo)).toBe('Imprensa Nacional-Casa da Moeda')
    expect(photoCreditLine('reverse', photo)).toContain('INCM, Imprensa Nacional-Casa da Moeda')
    expect(photoLoadVia(photo)).toBe('direct')
    expect(photoLoadVia({ ...photo, source: 'wikimedia' })).toBe('proxy')
  })

  it('names the Monnaie de Paris as the source of its photos', () => {
    const photo = { source: 'mdp', author: 'Monnaie de Paris', thumbnail: '', imageUrl: 'https://www.monnaiedeparis.fr/media/catalog/product/a.jpg' } as CoinPhotoCandidate
    expect(photoSourceName(photo)).toBe('Monnaie de Paris')
    expect(photoLoadVia(photo)).toBe('direct')
    // The mint is named once, not as author and source.
    expect(photoCaption(photo)).toBe('© Monnaie de Paris')
    expect(photoCreditLine('obverse', photo)).not.toContain('Monnaie de Paris, Monnaie de Paris')
  })
})
