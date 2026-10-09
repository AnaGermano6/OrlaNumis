import type { CoinPhotoCandidate } from '@/api/client'
import { t } from '@/i18n'

export type PhotoSide = 'obverse' | 'reverse'

/** Photos chosen for each side of a coin. */
export interface RealPhotoChoice {
  obverse?: CoinPhotoCandidate
  reverse?: CoinPhotoCandidate
}

export function photoSourceName(photo: CoinPhotoCandidate): string {
  return photo.source === 'numista' ? 'Numista' : 'Wikimedia Commons'
}

/** "© Author · CC BY-SA 4.0 · Wikimedia Commons" shown under a photo. */
export function photoCaption(photo: CoinPhotoCandidate): string {
  return [photo.author ? `© ${photo.author}` : '', photo.license ?? '', photoSourceName(photo)].filter(Boolean).join(' · ')
}

/** The credit line kept in the coin's notes once a photo is saved. */
export function photoCreditLine(side: PhotoSide, photo: CoinPhotoCandidate): string {
  const credit = [photo.author, photo.license, photoSourceName(photo)].filter(Boolean).join(', ')
  const line = t(side === 'obverse' ? 'photos.creditObverse' : 'photos.creditReverse', { credit })
  return photo.pageUrl ? `${line} – ${photo.pageUrl}` : line
}

/** Wikimedia hosts are not in the app's CSP, so their photos go through the image proxy. */
export function photoLoadVia(photo: CoinPhotoCandidate): 'direct' | 'proxy' {
  return photo.source === 'numista' ? 'direct' : 'proxy'
}
