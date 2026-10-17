// Real coin photos from public sources (ECB, the Portuguese, French and
// Austrian mints, Numista catalogue, Wikimedia Commons), found by search and imported
// into a coin with their credit.
import { api } from '@/api/http'
import type { NumistaDataStatus } from '@/api/endpoints/coins'
import type { CoinImage } from '@/types'

export interface CoinPhotoCandidate {
  source: 'numista' | 'wikimedia' | 'ecb' | 'incm' | 'mdp' | 'moe'
  /** Set when the source says which side the photo shows. */
  side?: 'obverse' | 'reverse'
  thumbnail: string
  /** Full-size photo to import. */
  imageUrl: string
  pageUrl?: string
  title?: string
  author?: string
  license?: string
  /** "exact" when the source identified this very coin (ECB, a mint's shop). */
  match?: 'exact'
}

export interface CoinPhotoSearchResult {
  status: NumistaDataStatus
  query: string
  photos: CoinPhotoCandidate[]
}

export interface CoinPhotoSearchParams {
  q?: string
  numistaTypeId?: number
  lang: string
  /** With a year, a €2 commemorative's ECB image is looked up too. */
  year?: number
  country?: string
  denomination?: string
  subject?: string
}

export const searchCoinPhotos = (params: CoinPhotoSearchParams) =>
  api.get<CoinPhotoSearchResult>('/coin-photos/search', { params })

/**
 * Gives a euro coin the European Central Bank's images of the sides asked for
 * that it has no image of; the credit is appended to the coin's notes.
 */
export const addOfficialCoinPhotos = (coinId: number, sides: Array<'obverse' | 'reverse'>) =>
  api.post<{ images: CoinImage[] }>(`/coins/${coinId}/official-photos`, { sides })

/** Saves a found photo as a coin image; the credit is appended to the coin's notes. */
export const importCoinImage = (
  coinId: number,
  body: { url: string; imageType: 'obverse' | 'reverse'; isPrimary: boolean; credit?: string },
) => api.post<CoinImage>(`/coins/${coinId}/images/import`, body)

/** How many of the user's coins (not banknotes) have no photo at all. */
export const getCoinsWithoutPhotos = () => api.get<{ count: number }>('/coins/official-photos/missing')

/** What a batch of official photos did, and where the next one starts. */
export interface OfficialPhotosBatch {
  checked: number
  withPhotos: number
  images: number
  nextAfterId: number
  done: boolean
}

/** Adds the official images of up to limit coins without photos, after afterId. */
export const addOfficialPhotosBatch = (afterId: number, limit = 10) =>
  api.post<OfficialPhotosBatch>('/coins/official-photos/batch', { afterId, limit })
