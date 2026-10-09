// Real coin photos from public sources (Numista catalogue, Wikimedia Commons),
// found by search and imported into a coin with their credit.
import { api } from '@/api/http'
import type { NumistaDataStatus } from '@/api/endpoints/coins'
import type { CoinImage } from '@/types'

export interface CoinPhotoCandidate {
  source: 'numista' | 'wikimedia'
  /** Set when the source says which side the photo shows. */
  side?: 'obverse' | 'reverse'
  thumbnail: string
  /** Full-size photo to import. */
  imageUrl: string
  pageUrl?: string
  title?: string
  author?: string
  license?: string
}

export interface CoinPhotoSearchResult {
  status: NumistaDataStatus
  query: string
  photos: CoinPhotoCandidate[]
}

export const searchCoinPhotos = (params: { q?: string; numistaTypeId?: number; lang: string }) =>
  api.get<CoinPhotoSearchResult>('/coin-photos/search', { params })

/** Saves a found photo as a coin image; the credit is appended to the coin's notes. */
export const importCoinImage = (
  coinId: number,
  body: { url: string; imageType: 'obverse' | 'reverse'; isPrimary: boolean; credit?: string },
) => api.post<CoinImage>(`/coins/${coinId}/images/import`, body)
