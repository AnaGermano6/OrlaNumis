import type { CoinPhotoCandidate } from '@/api/client'
import { t } from '@/i18n'
import type { Coin } from '@/types'

export type PhotoSide = 'obverse' | 'reverse'

/** Photos chosen for each side of a coin. */
export interface RealPhotoChoice {
  obverse?: CoinPhotoCandidate
  reverse?: CoinPhotoCandidate
}

export function photoSourceName(photo: CoinPhotoCandidate): string {
  if (photo.source === 'ecb') return t('photos.sources.ecb')
  if (photo.source === 'incm') return t('photos.sources.incm')
  if (photo.source === 'mdp') return 'Monnaie de Paris'
  return photo.source === 'numista' ? 'Numista' : 'Wikimedia Commons'
}

/** What a photo search needs to find a coin's ECB image, besides its Numista type. */
export interface CoinPhotoDetails {
  year?: number
  country?: string
  denomination?: string
  subject?: string
}

/**
 * A coin's year, country, denomination and subject for a photo search. The
 * subject is what follows " – " in the name, unless that is just the country
 * ("2 euro 2026 D – Bremen, Casa do Clima" has one, "… 2022 – Portugal" not).
 */
export function coinPhotoDetails(coin: Pick<Coin, 'name' | 'country' | 'year' | 'denomination'>): CoinPhotoDetails {
  const country = coin.country?.trim() ?? ''
  const separator = coin.name.indexOf(' – ')
  const detail = separator >= 0 ? coin.name.slice(separator + 3).trim() : ''
  return {
    year: coin.year ?? undefined,
    country: country || undefined,
    denomination: coin.denomination?.trim() || undefined,
    subject: detail && detail.toLowerCase() !== country.toLowerCase() ? detail : undefined,
  }
}

/** The source's name, unless the author already says it ("Monnaie de Paris"). */
function sourceUnlessAuthor(photo: CoinPhotoCandidate): string {
  const source = photoSourceName(photo)
  return source === photo.author ? '' : source
}

/** "© Author · CC BY-SA 4.0 · Wikimedia Commons" shown under a photo. */
export function photoCaption(photo: CoinPhotoCandidate): string {
  return [photo.author ? `© ${photo.author}` : '', photo.license ?? '', sourceUnlessAuthor(photo)].filter(Boolean).join(' · ')
}

/** The credit line kept in the coin's notes once a photo is saved. */
export function photoCreditLine(side: PhotoSide, photo: CoinPhotoCandidate): string {
  const credit = [photo.author, photo.license, sourceUnlessAuthor(photo)].filter(Boolean).join(', ')
  const line = t(side === 'obverse' ? 'photos.creditObverse' : 'photos.creditReverse', { credit })
  return photo.pageUrl ? `${line} – ${photo.pageUrl}` : line
}

/** Wikimedia hosts are not in the app's CSP, so their photos go through the image proxy. */
export function photoLoadVia(photo: CoinPhotoCandidate): 'direct' | 'proxy' {
  return photo.source === 'numista' || photo.source === 'incm' || photo.source === 'mdp' ? 'direct' : 'proxy'
}
